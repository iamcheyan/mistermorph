package guard

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestJSONLAuditSink_WritesDecisionMirrors(t *testing.T) {
	dir := t.TempDir()
	basePath := filepath.Join(dir, "guard_audit.jsonl")

	sink, err := NewJSONLAuditSink(basePath, 0, "")
	if err != nil {
		t.Fatalf("NewJSONLAuditSink() error = %v", err)
	}
	defer func() {
		_ = sink.Close()
	}()

	events := []AuditEvent{
		{
			EventID:    "evt-allow",
			RunID:      "run-1",
			Timestamp:  time.Unix(1, 0).UTC(),
			Step:       0,
			ActionType: ActionToolCallPre,
			Decision:   DecisionAllow,
			RiskLevel:  RiskLow,
		},
		{
			EventID:    "evt-redact",
			RunID:      "run-1",
			Timestamp:  time.Unix(2, 0).UTC(),
			Step:       1,
			ActionType: ActionToolCallPost,
			Decision:   DecisionAllowWithRedact,
			RiskLevel:  RiskHigh,
			Reasons:    []string{"redacted_secret_value"},
		},
		{
			EventID:    "evt-deny",
			RunID:      "run-2",
			Timestamp:  time.Unix(3, 0).UTC(),
			Step:       0,
			ActionType: ActionToolCallPre,
			Decision:   DecisionDeny,
			RiskLevel:  RiskHigh,
			Reasons:    []string{"private_ip"},
		},
	}

	for _, event := range events {
		if err := sink.Emit(context.Background(), event); err != nil {
			t.Fatalf("Emit(%s) error = %v", event.EventID, err)
		}
	}
	if err := sink.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	mainRaw, err := os.ReadFile(basePath)
	if err != nil {
		t.Fatalf("ReadFile(main) error = %v", err)
	}
	mainText := string(mainRaw)
	for _, want := range []string{"evt-allow", "evt-redact", "evt-deny"} {
		if !strings.Contains(mainText, want) {
			t.Fatalf("main audit log missing %q: %s", want, mainText)
		}
	}

	redactPath := auditDecisionMirrorPath(basePath, DecisionAllowWithRedact)
	redactRaw, err := os.ReadFile(redactPath)
	if err != nil {
		t.Fatalf("ReadFile(allow_with_redaction) error = %v", err)
	}
	redactText := string(redactRaw)
	if !strings.Contains(redactText, "evt-redact") || strings.Contains(redactText, "evt-deny") || strings.Contains(redactText, "evt-allow") {
		t.Fatalf("allow_with_redaction mirror has unexpected content: %s", redactText)
	}

	denyPath := auditDecisionMirrorPath(basePath, DecisionDeny)
	denyRaw, err := os.ReadFile(denyPath)
	if err != nil {
		t.Fatalf("ReadFile(deny) error = %v", err)
	}
	denyText := string(denyRaw)
	if !strings.Contains(denyText, "evt-deny") || strings.Contains(denyText, "evt-redact") || strings.Contains(denyText, "evt-allow") {
		t.Fatalf("deny mirror has unexpected content: %s", denyText)
	}

	requireApprovalPath := auditDecisionMirrorPath(basePath, DecisionRequireApproval)
	if _, err := os.Stat(requireApprovalPath); !os.IsNotExist(err) {
		t.Fatalf("require_approval mirror should not exist before any matching events, stat err = %v", err)
	}
}
