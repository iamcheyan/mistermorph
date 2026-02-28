package memory

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultJournalMaxFileBytes = 1 * 1024 * 1024
	journalDateLayout          = "2006-01-02"
	journalKeyPattern          = "^since-([0-9]{4}-[0-9]{2}-[0-9]{2})-([0-9]{4})\\.jsonl$"
	journalFilePattern         = "^since-([0-9]{4}-[0-9]{2}-[0-9]{2})-([0-9]{4})\\.jsonl(?:\\.gz)?$"
	journalCheckpointFilename  = "checkpoint.json"
)

var (
	journalKeyRe  = regexp.MustCompile(journalKeyPattern)
	journalFileRe = regexp.MustCompile(journalFilePattern)
)

type JournalOptions struct {
	MaxFileBytes           int64
	CompressClosedSegments bool
}

type Journal struct {
	dir  string
	opts JournalOptions

	now func() time.Time

	mu       sync.Mutex
	file     *os.File
	fileName string
	fileSize int64
	fileLine int64
}

type JournalOffset struct {
	File string `json:"file"`
	Line int64  `json:"line"`
}

type JournalRecord struct {
	Offset JournalOffset `json:"offset"`
	Event  MemoryEvent   `json:"event"`
}

type journalSegmentFile struct {
	Key        string
	ActualName string
	Date       string
	Index      int
	Compressed bool
}

type JournalCheckpoint struct {
	File      string `json:"file,omitempty"`
	Line      int64  `json:"line,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

func NewJournal(root string, opts JournalOptions) *Journal {
	root = strings.TrimSpace(root)
	if root == "" {
		root = "memory"
	}
	opts = normalizeJournalOptions(opts)
	return &Journal{
		dir:  filepath.Join(root, "log"),
		opts: opts,
		now:  time.Now,
	}
}

func (m *Manager) NewJournal(opts JournalOptions) *Journal {
	if m == nil {
		return NewJournal("", opts)
	}
	return NewJournal(m.memoryRoot(), opts)
}

func normalizeJournalOptions(opts JournalOptions) JournalOptions {
	if opts.MaxFileBytes <= 0 {
		opts.MaxFileBytes = defaultJournalMaxFileBytes
	}
	return opts
}

func (j *Journal) Append(event MemoryEvent) (JournalOffset, error) {
	if err := ValidateMemoryEventForAppend(event); err != nil {
		return JournalOffset{}, err
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return JournalOffset{}, fmt.Errorf("journal append: encode event: %w", err)
	}
	payload = append(payload, '\n')

	j.mu.Lock()
	defer j.mu.Unlock()

	if err := j.ensureWritableFileLocked(int64(len(payload))); err != nil {
		return JournalOffset{}, err
	}
	if j.file == nil {
		return JournalOffset{}, fmt.Errorf("journal append: no writable file")
	}

	prevLine := j.fileLine

	n, err := j.file.Write(payload)
	if err != nil {
		return JournalOffset{}, err
	}
	if int64(n) != int64(len(payload)) {
		return JournalOffset{}, fmt.Errorf("journal append: short write")
	}
	if err := j.file.Sync(); err != nil {
		return JournalOffset{}, err
	}
	j.fileSize += int64(n)
	j.fileLine++

	return JournalOffset{
		File: j.fileName,
		Line: prevLine + 1,
	}, nil
}

func (j *Journal) ReplayFrom(offset JournalOffset, limit int, fn func(JournalRecord) error) (JournalOffset, bool, error) {
	if fn == nil {
		return JournalOffset{}, false, fmt.Errorf("replay callback is required")
	}
	if limit <= 0 {
		return JournalOffset{}, false, fmt.Errorf("limit must be > 0")
	}
	if strings.TrimSpace(offset.File) != offset.File {
		return JournalOffset{}, false, fmt.Errorf("offset.file must not contain leading/trailing spaces")
	}
	if offset.File != "" {
		if _, _, ok := parseJournalKey(offset.File); !ok {
			return JournalOffset{}, false, fmt.Errorf("offset.file is invalid")
		}
	}
	if offset.Line < 0 {
		return JournalOffset{}, false, fmt.Errorf("offset.line must be >= 0")
	}

	segments, err := j.listSegmentFiles()
	if err != nil {
		return JournalOffset{}, false, err
	}

	next := offset
	deliveredTotal := 0
	for _, seg := range segments {
		if offset.File != "" && seg.Key < offset.File {
			continue
		}
		remaining := limit - deliveredTotal
		if remaining <= 0 {
			return next, false, nil
		}
		delivered, last, err := j.replayFileLimit(seg, offset, remaining, fn)
		if err != nil {
			return next, false, err
		}
		if delivered > 0 {
			next = last
			deliveredTotal += delivered
		}
		if deliveredTotal >= limit {
			return next, false, nil
		}
	}
	return next, true, nil
}

func (j *Journal) LoadCheckpoint() (JournalCheckpoint, bool, error) {
	path := j.checkpointPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return JournalCheckpoint{}, false, nil
		}
		return JournalCheckpoint{}, false, err
	}
	var cp JournalCheckpoint
	if err := json.Unmarshal(data, &cp); err != nil {
		return JournalCheckpoint{}, false, err
	}
	if err := validateCheckpoint(cp); err != nil {
		return JournalCheckpoint{}, false, err
	}
	return cp, true, nil
}

func (j *Journal) SaveCheckpoint(cp JournalCheckpoint) error {
	if strings.TrimSpace(cp.UpdatedAt) == "" {
		cp.UpdatedAt = j.now().UTC().Format(time.RFC3339)
	}
	if err := validateCheckpoint(cp); err != nil {
		return err
	}
	if err := os.MkdirAll(j.dir, 0o700); err != nil {
		return err
	}
	path := j.checkpointPath()
	data, err := json.Marshal(cp)
	if err != nil {
		return err
	}
	data = append(data, '\n')

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func (j *Journal) Close() error {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.file == nil {
		return nil
	}
	err := j.file.Close()
	j.file = nil
	j.fileName = ""
	j.fileSize = 0
	j.fileLine = 0
	return err
}

func (j *Journal) ensureWritableFileLocked(incomingBytes int64) error {
	if j.file == nil {
		if err := j.reopenLatestLocked(); err != nil {
			return err
		}
	}
	if j.file == nil {
		return fmt.Errorf("journal writable file is nil")
	}

	// Rotate only by size. The date in filename is the first record date of the segment.
	if j.opts.MaxFileBytes > 0 && j.fileSize > 0 && j.fileSize+incomingBytes > j.opts.MaxFileBytes {
		startDate := j.now().UTC().Format(journalDateLayout)
		if err := j.rotateToNewSegmentLocked(startDate); err != nil {
			return err
		}
	}
	return nil
}

func (j *Journal) reopenLatestLocked() error {
	if err := j.closeFileLocked(); err != nil {
		return err
	}
	if err := os.MkdirAll(j.dir, 0o700); err != nil {
		return err
	}

	segments, err := j.listSegmentFiles()
	if err != nil {
		return err
	}
	if len(segments) == 0 {
		date := j.now().UTC().Format(journalDateLayout)
		return j.openNewFileForDateAndIndexLocked(date, 1)
	}

	latestWritable := journalSegmentFile{}
	foundWritable := false
	for i := len(segments) - 1; i >= 0; i-- {
		if segments[i].Compressed {
			continue
		}
		latestWritable = segments[i]
		foundWritable = true
		break
	}
	if foundWritable {
		return j.openExistingFileForDateAndIndexLocked(latestWritable.Date, latestWritable.Index)
	}

	date := j.now().UTC().Format(journalDateLayout)
	maxIdx, err := j.maxIndexForDate(date)
	if err != nil {
		return err
	}
	return j.openNewFileForDateAndIndexLocked(date, maxIdx+1)
}

func (j *Journal) rotateToNewSegmentLocked(startDate string) error {
	maxIdx, err := j.maxIndexForDate(startDate)
	if err != nil {
		return err
	}
	next := maxIdx + 1
	closedName := j.fileName
	if err := j.closeFileLocked(); err != nil {
		return err
	}
	if j.opts.CompressClosedSegments && strings.TrimSpace(closedName) != "" {
		if err := j.compressClosedSegment(closedName); err != nil {
			return err
		}
	}
	return j.openNewFileForDateAndIndexLocked(startDate, next)
}

func (j *Journal) openExistingFileForDateAndIndexLocked(date string, idx int) error {
	if err := os.MkdirAll(j.dir, 0o700); err != nil {
		return err
	}
	name := formatJournalFileName(date, idx)
	path := filepath.Join(j.dir, name)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return err
	}
	lineCount, err := countJournalLines(path)
	if err != nil {
		_ = file.Close()
		return err
	}
	j.file = file
	j.fileName = name
	j.fileSize = info.Size()
	j.fileLine = lineCount
	return nil
}

func (j *Journal) openNewFileForDateAndIndexLocked(date string, idx int) error {
	if err := os.MkdirAll(j.dir, 0o700); err != nil {
		return err
	}
	name := formatJournalFileName(date, idx)
	path := filepath.Join(j.dir, name)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	j.file = file
	j.fileName = name
	j.fileSize = 0
	j.fileLine = 0
	return nil
}

func (j *Journal) closeFileLocked() error {
	if j.file == nil {
		return nil
	}
	err := j.file.Close()
	j.file = nil
	j.fileName = ""
	j.fileSize = 0
	j.fileLine = 0
	return err
}

func (j *Journal) replayFileLimit(seg journalSegmentFile, from JournalOffset, maxDeliver int, fn func(JournalRecord) error) (int, JournalOffset, error) {
	path := filepath.Join(j.dir, seg.ActualName)
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, JournalOffset{}, nil
		}
		return 0, JournalOffset{}, err
	}
	defer file.Close()

	var reader *bufio.Reader
	if seg.Compressed {
		gz, err := gzip.NewReader(file)
		if err != nil {
			return 0, JournalOffset{}, err
		}
		defer gz.Close()
		reader = bufio.NewReader(gz)
	} else {
		reader = bufio.NewReader(file)
	}

	var lineNo int64
	delivered := 0
	last := JournalOffset{}
	for {
		line, readErr := reader.ReadBytes('\n')
		if len(line) > 0 {
			lineNo++

			trimmed := bytes.TrimSpace(line)
			if len(trimmed) > 0 {
				if from.File == seg.Key && from.Line > 0 && lineNo <= from.Line {
					// Skip already applied offsets.
				} else {
					var event MemoryEvent
					if err := json.Unmarshal(trimmed, &event); err != nil {
						return delivered, last, fmt.Errorf("journal replay decode %s:%d: %w", seg.ActualName, lineNo, err)
					}
					if err := ValidateMemoryEventForAppend(event); err != nil {
						return delivered, last, fmt.Errorf("journal replay invalid event %s:%d: %w", seg.ActualName, lineNo, err)
					}
					rec := JournalRecord{
						Offset: JournalOffset{
							File: seg.Key,
							Line: lineNo,
						},
						Event: event,
					}
					if err := fn(rec); err != nil {
						return delivered, last, err
					}
					last = rec.Offset
					delivered++
					if delivered >= maxDeliver {
						return delivered, last, nil
					}
				}
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				return delivered, last, nil
			}
			return delivered, last, readErr
		}
	}
}

func (j *Journal) listSegmentFiles() ([]journalSegmentFile, error) {
	if err := os.MkdirAll(j.dir, 0o700); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(j.dir)
	if err != nil {
		return nil, err
	}
	byKey := make(map[string]journalSegmentFile, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if name == journalCheckpointFilename {
			continue
		}
		seg, ok := parseJournalSegmentFile(name)
		if !ok {
			continue
		}
		if _, exists := byKey[seg.Key]; exists {
			return nil, fmt.Errorf("duplicate segment key found: %s", seg.Key)
		}
		byKey[seg.Key] = seg
	}
	out := make([]journalSegmentFile, 0, len(byKey))
	for _, seg := range byKey {
		out = append(out, seg)
	}
	sort.Slice(out, func(i, k int) bool {
		return out[i].Key < out[k].Key
	})
	return out, nil
}

func (j *Journal) maxIndexForDate(date string) (int, error) {
	segments, err := j.listSegmentFiles()
	if err != nil {
		return 0, err
	}
	maxIdx := 0
	for _, seg := range segments {
		if seg.Date != date {
			continue
		}
		if seg.Index > maxIdx {
			maxIdx = seg.Index
		}
	}
	return maxIdx, nil
}

func (j *Journal) checkpointPath() string {
	return filepath.Join(j.dir, journalCheckpointFilename)
}

func formatJournalFileName(date string, idx int) string {
	if idx <= 0 {
		idx = 1
	}
	return fmt.Sprintf("since-%s-%04d.jsonl", date, idx)
}

func parseJournalKey(name string) (date string, idx int, ok bool) {
	m := journalKeyRe.FindStringSubmatch(name)
	if len(m) != 3 {
		return "", 0, false
	}
	n, err := strconv.Atoi(m[2])
	if err != nil || n <= 0 {
		return "", 0, false
	}
	return m[1], n, true
}

func parseJournalSegmentFile(name string) (journalSegmentFile, bool) {
	if !journalFileRe.MatchString(name) {
		return journalSegmentFile{}, false
	}
	key := name
	if strings.HasSuffix(name, ".gz") {
		key = strings.TrimSuffix(name, ".gz")
	}
	date, idx, ok := parseJournalKey(key)
	if !ok {
		return journalSegmentFile{}, false
	}
	return journalSegmentFile{
		Key:        key,
		ActualName: name,
		Date:       date,
		Index:      idx,
		Compressed: strings.HasSuffix(name, ".gz"),
	}, true
}

func (j *Journal) compressClosedSegment(logicalName string) error {
	logicalName = strings.TrimSpace(logicalName)
	if _, _, ok := parseJournalKey(logicalName); !ok {
		return fmt.Errorf("invalid segment key: %s", logicalName)
	}
	srcPath := filepath.Join(j.dir, logicalName)
	dstPath := srcPath + ".gz"
	if _, err := os.Stat(srcPath); err != nil {
		return err
	}
	if _, err := os.Stat(dstPath); err == nil {
		return fmt.Errorf("compressed segment already exists: %s", dstPath)
	} else if !os.IsNotExist(err) {
		return err
	}

	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	tmpPath := dstPath + ".tmp"
	dst, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}

	gzw := gzip.NewWriter(dst)
	if _, err := io.Copy(gzw, src); err != nil {
		_ = gzw.Close()
		_ = dst.Close()
		_ = os.Remove(tmpPath)
		return err
	}
	if err := gzw.Close(); err != nil {
		_ = dst.Close()
		_ = os.Remove(tmpPath)
		return err
	}
	if err := dst.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	if err := os.Rename(tmpPath, dstPath); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return os.Remove(srcPath)
}

func countJournalLines(path string) (int64, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	var lines int64
	var sawAnyByte bool
	var lastByte byte
	for {
		chunk, readErr := reader.ReadBytes('\n')
		if len(chunk) > 0 {
			sawAnyByte = true
			lastByte = chunk[len(chunk)-1]
			if lastByte == '\n' {
				lines++
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return 0, readErr
		}
	}
	if sawAnyByte && lastByte != '\n' {
		lines++
	}
	return lines, nil
}

func validateCheckpoint(cp JournalCheckpoint) error {
	if strings.TrimSpace(cp.File) != cp.File {
		return fmt.Errorf("checkpoint.file must not contain leading/trailing spaces")
	}
	if cp.File != "" {
		if _, _, ok := parseJournalKey(cp.File); !ok {
			return fmt.Errorf("checkpoint.file is invalid")
		}
	}
	if cp.Line < 0 {
		return fmt.Errorf("checkpoint.line must be >= 0")
	}
	if strings.TrimSpace(cp.UpdatedAt) != cp.UpdatedAt {
		return fmt.Errorf("checkpoint.updated_at must not contain leading/trailing spaces")
	}
	if cp.UpdatedAt != "" {
		if _, err := time.Parse(time.RFC3339, cp.UpdatedAt); err != nil {
			return fmt.Errorf("checkpoint.updated_at must be RFC3339")
		}
	}
	return nil
}
