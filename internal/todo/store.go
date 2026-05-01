package todo

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/quailyquaily/mistermorph/internal/fsstore"
	"github.com/quailyquaily/mistermorph/internal/pathutil"
)

func NewStore(wipPath string, donePath string) *Store {
	wipPath = pathutil.ExpandHomePath(strings.TrimSpace(wipPath))
	donePath = pathutil.ExpandHomePath(strings.TrimSpace(donePath))
	recurringPath := ""
	if wipPath != "" {
		recurringPath = filepath.Join(filepath.Dir(wipPath), DefaultRECURFilename)
	}
	return &Store{
		WIPPath:       wipPath,
		DONEPath:      donePath,
		RecurringPath: recurringPath,
		Now:           time.Now,
	}
}

func (s *Store) readFiles() (WIPFile, DONEFile, error) {
	now := s.nowUTC()
	wip, err := s.readWIP(now)
	if err != nil {
		return WIPFile{}, DONEFile{}, err
	}
	done, err := s.readDONE(now)
	if err != nil {
		return WIPFile{}, DONEFile{}, err
	}
	return wip, done, nil
}

func (s *Store) writeFiles(wip WIPFile, done DONEFile) error {
	now := s.nowUTC().Format(time.RFC3339)
	if strings.TrimSpace(wip.CreatedAt) == "" {
		wip.CreatedAt = now
	}
	if strings.TrimSpace(done.CreatedAt) == "" {
		done.CreatedAt = now
	}
	wip.UpdatedAt = now
	done.UpdatedAt = now
	wip.OpenCount = len(wip.Entries)
	done.DoneCount = len(done.Entries)
	if err := fsstore.WriteTextAtomic(s.WIPPath, RenderWIP(wip), fsstore.FileOptions{DirPerm: 0o700, FilePerm: 0o600}); err != nil {
		return err
	}
	if err := fsstore.WriteTextAtomic(s.DONEPath, RenderDONE(done), fsstore.FileOptions{DirPerm: 0o700, FilePerm: 0o600}); err != nil {
		return err
	}
	return nil
}

func (s *Store) readWIP(now time.Time) (WIPFile, error) {
	nowRFC3339 := now.UTC().Format(time.RFC3339)
	text, exists, err := fsstore.ReadText(s.WIPPath)
	if err != nil {
		return WIPFile{}, err
	}
	if !exists || strings.TrimSpace(text) == "" {
		return WIPFile{
			CreatedAt: nowRFC3339,
			UpdatedAt: nowRFC3339,
			OpenCount: 0,
			Entries:   nil,
		}, nil
	}
	wip, err := ParseWIP(text)
	if err != nil {
		return WIPFile{}, err
	}
	if strings.TrimSpace(wip.CreatedAt) == "" {
		wip.CreatedAt = nowRFC3339
	}
	if strings.TrimSpace(wip.UpdatedAt) == "" {
		wip.UpdatedAt = nowRFC3339
	}
	wip.OpenCount = len(wip.Entries)
	return wip, nil
}

func (s *Store) readDONE(now time.Time) (DONEFile, error) {
	nowRFC3339 := now.UTC().Format(time.RFC3339)
	text, exists, err := fsstore.ReadText(s.DONEPath)
	if err != nil {
		return DONEFile{}, err
	}
	if !exists || strings.TrimSpace(text) == "" {
		return DONEFile{
			CreatedAt: nowRFC3339,
			UpdatedAt: nowRFC3339,
			DoneCount: 0,
			Entries:   nil,
		}, nil
	}
	done, err := ParseDONE(text)
	if err != nil {
		return DONEFile{}, err
	}
	if strings.TrimSpace(done.CreatedAt) == "" {
		done.CreatedAt = nowRFC3339
	}
	if strings.TrimSpace(done.UpdatedAt) == "" {
		done.UpdatedAt = nowRFC3339
	}
	done.DoneCount = len(done.Entries)
	return done, nil
}

func (s *Store) readRECUR(now time.Time) (RECURFile, bool, error) {
	nowRFC3339 := now.UTC().Format(time.RFC3339)
	path := strings.TrimSpace(s.RecurringPath)
	if path == "" {
		return RECURFile{
			CreatedAt:      nowRFC3339,
			UpdatedAt:      nowRFC3339,
			RecurringCount: 0,
			Entries:        nil,
		}, false, nil
	}
	text, exists, err := fsstore.ReadText(path)
	if err != nil {
		return RECURFile{}, false, err
	}
	if !exists || strings.TrimSpace(text) == "" {
		return RECURFile{
			CreatedAt:      nowRFC3339,
			UpdatedAt:      nowRFC3339,
			RecurringCount: 0,
			Entries:        nil,
		}, exists, nil
	}
	recur, err := ParseRECUR(text)
	if err != nil {
		return RECURFile{}, exists, err
	}
	if strings.TrimSpace(recur.CreatedAt) == "" {
		recur.CreatedAt = nowRFC3339
	}
	if strings.TrimSpace(recur.UpdatedAt) == "" {
		recur.UpdatedAt = nowRFC3339
	}
	recur.RecurringCount = len(recur.Entries)
	return recur, exists, nil
}

func (s *Store) writeRECUR(file RECURFile) error {
	path := strings.TrimSpace(s.RecurringPath)
	if path == "" {
		return nil
	}
	now := s.nowUTC().Format(time.RFC3339)
	if strings.TrimSpace(file.CreatedAt) == "" {
		file.CreatedAt = now
	}
	file.UpdatedAt = now
	file.RecurringCount = len(file.Entries)
	return fsstore.WriteTextAtomic(path, RenderRECUR(file), fsstore.FileOptions{DirPerm: 0o700, FilePerm: 0o600})
}

func (s *Store) nowUTC() time.Time {
	if s != nil && s.Now != nil {
		return s.Now().UTC()
	}
	return time.Now().UTC()
}
