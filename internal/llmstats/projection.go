package llmstats

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/quailyquaily/mistermorph/internal/fsstore"
)

type ProjectionStore struct {
	journalDir string
	path       string
	now        func() time.Time
}

type aggregateState struct {
	summary Totals
	skipped int64
	byHost  map[string]*apiHostState
	byModel map[string]*Totals
}

type apiHostState struct {
	totals  Totals
	byModel map[string]*Totals
}

func NewProjectionStore(journalDir, path string) *ProjectionStore {
	return &ProjectionStore{
		journalDir: strings.TrimSpace(journalDir),
		path:       strings.TrimSpace(path),
		now:        time.Now,
	}
}

func (s *ProjectionStore) Refresh() (Projection, error) {
	proj, ok, err := loadProjection(s.path)
	if err != nil || !ok {
		proj = Projection{}
	}

	segments, err := listSegmentFiles(s.journalDir)
	if err != nil {
		return Projection{}, err
	}
	if len(segments) == 0 {
		zero := Projection{UpdatedAt: s.now().UTC().Format(time.RFC3339)}
		if err := saveProjection(s.path, zero); err != nil {
			return Projection{}, err
		}
		return zero, nil
	}

	start := proj.ProjectedOffset
	if !offsetValidForSegments(s.journalDir, segments, start) {
		proj = Projection{}
		start = Offset{}
	}

	state := aggregateStateFromProjection(proj)
	nextOffset, skipped, err := scanJournalFrom(s.journalDir, segments, start, func(rec RequestRecord, _ Offset) error {
		state.add(rec)
		return nil
	})
	if err != nil {
		return Projection{}, err
	}
	state.skipped += skipped

	out := state.toProjection()
	out.UpdatedAt = s.now().UTC().Format(time.RFC3339)
	out.ProjectedOffset = nextOffset
	out.ProjectedRecords = out.Summary.Requests
	if err := saveProjection(s.path, out); err != nil {
		return Projection{}, err
	}
	return out, nil
}

func aggregateStateFromProjection(p Projection) *aggregateState {
	st := &aggregateState{
		summary: p.Summary,
		skipped: p.SkippedRecords,
		byHost:  map[string]*apiHostState{},
		byModel: map[string]*Totals{},
	}
	for _, host := range p.APIHosts {
		hs := &apiHostState{totals: host.Totals, byModel: map[string]*Totals{}}
		for _, model := range host.Models {
			m := model.Totals
			hs.byModel[model.Model] = &m
		}
		st.byHost[host.APIHost] = hs
	}
	for _, model := range p.Models {
		m := model.Totals
		st.byModel[model.Model] = &m
	}
	return st
}

func (s *aggregateState) add(rec RequestRecord) {
	if s == nil {
		return
	}
	rec = normalizeRequestRecord(rec)
	s.summary.AddRecord(rec)
	if s.byModel == nil {
		s.byModel = map[string]*Totals{}
	}
	if s.byHost == nil {
		s.byHost = map[string]*apiHostState{}
	}

	modelTotals, ok := s.byModel[rec.Model]
	if !ok {
		modelTotals = &Totals{}
		s.byModel[rec.Model] = modelTotals
	}
	modelTotals.AddRecord(rec)

	hostState, ok := s.byHost[rec.APIHost]
	if !ok {
		hostState = &apiHostState{byModel: map[string]*Totals{}}
		s.byHost[rec.APIHost] = hostState
	}
	hostState.totals.AddRecord(rec)

	hostModelTotals, ok := hostState.byModel[rec.Model]
	if !ok {
		hostModelTotals = &Totals{}
		hostState.byModel[rec.Model] = hostModelTotals
	}
	hostModelTotals.AddRecord(rec)
}

func (s *aggregateState) toProjection() Projection {
	if s == nil {
		return Projection{}
	}
	models := make([]ModelSummary, 0, len(s.byModel))
	for model, totals := range s.byModel {
		models = append(models, ModelSummary{Model: model, Totals: *totals})
	}
	sortModelSummaries(models)

	hosts := make([]APIHostSummary, 0, len(s.byHost))
	for host, hostState := range s.byHost {
		hostSummary := APIHostSummary{APIHost: host, Totals: hostState.totals}
		hostSummary.Models = make([]ModelSummary, 0, len(hostState.byModel))
		for model, totals := range hostState.byModel {
			hostSummary.Models = append(hostSummary.Models, ModelSummary{Model: model, Totals: *totals})
		}
		hosts = append(hosts, hostSummary)
	}
	sortAPIHostSummaries(hosts)

	return Projection{
		Summary:        s.summary,
		APIHosts:       hosts,
		Models:         models,
		SkippedRecords: s.skipped,
	}
}

func loadProjection(path string) (Projection, bool, error) {
	if strings.TrimSpace(path) == "" {
		return Projection{}, false, nil
	}
	var proj Projection
	ok, err := fsstore.ReadJSON(path, &proj)
	if err != nil {
		return Projection{}, false, err
	}
	if !ok {
		return Projection{}, false, nil
	}
	return proj, true, nil
}

func saveProjection(path string, proj Projection) error {
	if strings.TrimSpace(path) == "" {
		return nil
	}
	return fsstore.WriteJSONAtomic(path, proj, fsstore.FileOptions{})
}

func offsetValidForSegments(dir string, segments []journalSegmentFile, off Offset) bool {
	off.File = strings.TrimSpace(off.File)
	if off.File == "" {
		return off.Line == 0
	}
	if off.Line < 0 {
		return false
	}
	var target *journalSegmentFile
	for i := range segments {
		if segments[i].Key == off.File {
			target = &segments[i]
			break
		}
	}
	if target == nil {
		return false
	}
	lines, err := countLines(filepath.Join(dir, target.ActualName))
	if err != nil {
		return false
	}
	return off.Line <= lines
}

func scanJournalFrom(dir string, segments []journalSegmentFile, from Offset, fn func(RequestRecord, Offset) error) (Offset, int64, error) {
	if fn == nil {
		return from, 0, fmt.Errorf("scan callback is required")
	}
	next := from
	var skipped int64
	for _, seg := range segments {
		if from.File != "" && seg.Key < from.File {
			continue
		}
		path := filepath.Join(dir, seg.ActualName)
		file, err := os.Open(path)
		if err != nil {
			return next, skipped, err
		}
		reader := bufio.NewReader(file)
		var lineNo int64
		for {
			line, readErr := reader.ReadBytes('\n')
			if len(line) > 0 {
				if line[len(line)-1] != '\n' && readErr == io.EOF {
					break
				}
				lineNo++
				current := Offset{File: seg.Key, Line: lineNo}
				if from.File == seg.Key && lineNo <= from.Line {
					if readErr == io.EOF {
						break
					}
					if readErr != nil {
						_ = file.Close()
						return next, skipped, readErr
					}
					continue
				}

				raw := bytes.TrimSpace(line)
				if len(raw) == 0 {
					next = current
					skipped++
				} else {
					var rec RequestRecord
					if err := json.Unmarshal(raw, &rec); err != nil {
						next = current
						skipped++
					} else {
						rec = normalizeRequestRecord(rec)
						if err := validateRequestRecord(rec); err != nil {
							next = current
							skipped++
						} else {
							if err := fn(rec, current); err != nil {
								_ = file.Close()
								return next, skipped, err
							}
							next = current
						}
					}
				}
			}
			if readErr != nil {
				if readErr == io.EOF {
					break
				}
				_ = file.Close()
				return next, skipped, readErr
			}
		}
		if err := file.Close(); err != nil {
			return next, skipped, err
		}
	}
	return next, skipped, nil
}
