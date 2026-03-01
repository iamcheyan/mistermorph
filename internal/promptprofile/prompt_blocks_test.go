package promptprofile

import (
	"strings"
	"testing"

	"github.com/quailyquaily/mistermorph/agent"
)

func TestAppendSlackRuntimeBlocks_Group(t *testing.T) {
	spec := agent.PromptSpec{}
	mentions := []string{"U111", "U222"}

	AppendSlackRuntimeBlocks(&spec, true, mentions)

	if len(spec.Blocks) != 2 {
		t.Fatalf("blocks len = %d, want 2", len(spec.Blocks))
	}
	if !strings.Contains(spec.Blocks[0].Content, "### Slack Policies") {
		t.Fatalf("slack policy heading missing: %q", spec.Blocks[0].Content)
	}
	if !strings.Contains(spec.Blocks[0].Content, "<Slack Group Policies>") {
		t.Fatalf("group policy block missing marker: %q", spec.Blocks[0].Content)
	}
	if !strings.Contains(spec.Blocks[1].Content, "### Slack Mention Users") {
		t.Fatalf("mention heading missing: %q", spec.Blocks[1].Content)
	}
	if spec.Blocks[1].Content != "### Slack Mention Users\nU111\nU222" {
		t.Fatalf("mention block content = %q, want %q", spec.Blocks[1].Content, "### Slack Mention Users\nU111\nU222")
	}
}

func TestAppendSlackRuntimeBlocks_DM(t *testing.T) {
	spec := agent.PromptSpec{}

	AppendSlackRuntimeBlocks(&spec, false, []string{"U111"})

	if len(spec.Blocks) != 1 {
		t.Fatalf("blocks len = %d, want 1", len(spec.Blocks))
	}
	if !strings.Contains(spec.Blocks[0].Content, "### Slack Policies") {
		t.Fatalf("slack policy heading missing: %q", spec.Blocks[0].Content)
	}
	if !strings.Contains(spec.Blocks[0].Content, "<Slack DM Policies>") {
		t.Fatalf("dm policy block missing marker: %q", spec.Blocks[0].Content)
	}
}
