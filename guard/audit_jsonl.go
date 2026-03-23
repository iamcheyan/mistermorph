package guard

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/quailyquaily/mistermorph/internal/fsstore"
)

type JSONLAuditSink struct {
	path            string
	lockPath        string
	rotateMaxBytes  int64
	writer          *fsstore.JSONLWriter
	decisionWriters map[Decision]*fsstore.JSONLWriter
	mirrorPaths     map[Decision]string

	mu sync.Mutex
}

var mirroredAuditDecisions = []Decision{
	DecisionAllowWithRedact,
	DecisionRequireApproval,
	DecisionDeny,
}

func NewJSONLAuditSink(path string, rotateMaxBytes int64, lockRoot string) (*JSONLAuditSink, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, fmt.Errorf("missing jsonl path")
	}
	if strings.TrimSpace(lockRoot) == "" {
		lockRoot = filepath.Join(filepath.Dir(path), ".fslocks")
	}
	lockPath, err := fsstore.BuildLockPath(lockRoot, "audit.guard_audit_jsonl")
	if err != nil {
		return nil, err
	}
	writer, err := fsstore.NewJSONLWriter(path, fsstore.JSONLOptions{
		RotateMaxBytes: rotateMaxBytes,
		FlushEachWrite: true,
	})
	if err != nil {
		return nil, err
	}
	return &JSONLAuditSink{
		path:            path,
		lockPath:        lockPath,
		rotateMaxBytes:  rotateMaxBytes,
		writer:          writer,
		decisionWriters: map[Decision]*fsstore.JSONLWriter{},
		mirrorPaths:     buildAuditMirrorPaths(path),
	}, nil
}

func (s *JSONLAuditSink) Emit(ctx context.Context, e AuditEvent) error {
	if s == nil || s.writer == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	return fsstore.WithLock(ctx, s.lockPath, func() error {
		if err := s.writer.AppendJSON(e); err != nil {
			return err
		}
		writer, err := s.writerForDecisionLocked(e.Decision)
		if err != nil || writer == nil {
			return err
		}
		return writer.AppendJSON(e)
	})
}

func (s *JSONLAuditSink) Close() error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	var firstErr error
	if s.writer != nil {
		if err := s.writer.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		s.writer = nil
	}
	for decision, writer := range s.decisionWriters {
		if writer == nil {
			continue
		}
		if err := writer.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		s.decisionWriters[decision] = nil
	}
	return firstErr
}

func (s *JSONLAuditSink) writerForDecisionLocked(decision Decision) (*fsstore.JSONLWriter, error) {
	if s == nil {
		return nil, nil
	}
	if writer := s.decisionWriters[decision]; writer != nil {
		return writer, nil
	}
	path := strings.TrimSpace(s.mirrorPaths[decision])
	if path == "" {
		return nil, nil
	}
	writer, err := fsstore.NewJSONLWriter(path, fsstore.JSONLOptions{
		RotateMaxBytes: s.rotateMaxBytes,
		FlushEachWrite: true,
	})
	if err != nil {
		return nil, err
	}
	s.decisionWriters[decision] = writer
	return writer, nil
}

func buildAuditMirrorPaths(path string) map[Decision]string {
	out := make(map[Decision]string, len(mirroredAuditDecisions))
	for _, decision := range mirroredAuditDecisions {
		out[decision] = auditDecisionMirrorPath(path, decision)
	}
	return out
}

func auditDecisionMirrorPath(path string, decision Decision) string {
	path = strings.TrimSpace(path)
	suffix := strings.TrimSpace(string(decision))
	if path == "" || suffix == "" {
		return path
	}
	dir := filepath.Dir(path)
	name := filepath.Base(path)
	ext := filepath.Ext(name)
	stem := strings.TrimSuffix(name, ext)
	if ext == "" {
		return filepath.Join(dir, stem+"."+suffix)
	}
	return filepath.Join(dir, stem+"."+suffix+ext)
}
