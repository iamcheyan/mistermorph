package guard

import (
	"context"
	"encoding/json"
	"testing"
)

type captureAuditSink struct {
	events []AuditEvent
}

func (s *captureAuditSink) Emit(_ context.Context, e AuditEvent) error {
	s.events = append(s.events, e)
	return nil
}

func (s *captureAuditSink) Close() error {
	return nil
}

func TestAuditEvent_OutputPublishMarksBodyOmitted(t *testing.T) {
	sink := &captureAuditSink{}
	g := New(Config{
		Enabled: true,
		Redaction: RedactionConfig{
			Enabled: true,
		},
	}, sink, nil)

	_, err := g.Evaluate(context.Background(), Meta{
		RunID: "run-output-publish",
		Step:  1,
	}, Action{
		Type:    ActionOutputPublish,
		Content: "plain final output",
	})
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if len(sink.events) != 1 {
		t.Fatalf("captured %d events, want 1", len(sink.events))
	}
	if !sink.events[0].BodyOmittedFromAudit {
		t.Fatalf("BodyOmittedFromAudit = false, want true")
	}

	raw, err := json.Marshal(sink.events[0])
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if got["body_omitted_from_audit"] != true {
		t.Fatalf("body_omitted_from_audit = %v, want true", got["body_omitted_from_audit"])
	}
}

func TestAuditEvent_ToolCallDoesNotMarkBodyOmitted(t *testing.T) {
	sink := &captureAuditSink{}
	g := New(Config{
		Enabled: true,
	}, sink, nil)

	_, err := g.Evaluate(context.Background(), Meta{
		RunID: "run-tool-call",
		Step:  0,
	}, Action{
		Type:     ActionToolCallPre,
		ToolName: "read_file",
		ToolParams: map[string]any{
			"path": "/tmp/demo.txt",
		},
	})
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if len(sink.events) != 1 {
		t.Fatalf("captured %d events, want 1", len(sink.events))
	}
	if sink.events[0].BodyOmittedFromAudit {
		t.Fatalf("BodyOmittedFromAudit = true, want false")
	}

	raw, err := json.Marshal(sink.events[0])
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if _, ok := got["body_omitted_from_audit"]; ok {
		t.Fatalf("body_omitted_from_audit should be omitted for tool calls")
	}
}
