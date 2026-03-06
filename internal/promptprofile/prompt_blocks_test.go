package promptprofile

import (
	"strings"
	"testing"

	"github.com/quailyquaily/mistermorph/agent"
)

func TestAppendSlackRuntimeBlocks_Group(t *testing.T) {
	spec := agent.PromptSpec{}
	mentions := []string{"U111", "U222"}
	emojiList := "party_parrot,thumbsup,wave"

	AppendSlackRuntimeBlocks(&spec, true, mentions, emojiList)

	if len(spec.Blocks) != 2 {
		t.Fatalf("blocks len = %d, want 2", len(spec.Blocks))
	}
	if !strings.Contains(spec.Blocks[0].Content, "[[ Slack Policies ]]") {
		t.Fatalf("slack policy heading missing: %q", spec.Blocks[0].Content)
	}
	if !strings.Contains(spec.Blocks[0].Content, "[[ Slack Group Policies ]]") {
		t.Fatalf("group policy block missing marker: %q", spec.Blocks[0].Content)
	}
	if !strings.Contains(spec.Blocks[0].Content, "Use only these emoji names for `message_react`: party_parrot,thumbsup,wave") {
		t.Fatalf("emoji names csv line missing expected values: %q", spec.Blocks[0].Content)
	}
	if !strings.Contains(spec.Blocks[1].Content, "[[ Slack Mention Users ]]") {
		t.Fatalf("mention heading missing: %q", spec.Blocks[1].Content)
	}
	if spec.Blocks[1].Content != "[[ Slack Mention Users ]]\nU111\nU222" {
		t.Fatalf("mention block content = %q, want %q", spec.Blocks[1].Content, "[[ Slack Mention Users ]]\nU111\nU222")
	}
}

func TestAppendSlackRuntimeBlocks_DM(t *testing.T) {
	spec := agent.PromptSpec{}

	AppendSlackRuntimeBlocks(&spec, false, []string{"U111"}, "")

	if len(spec.Blocks) != 1 {
		t.Fatalf("blocks len = %d, want 1", len(spec.Blocks))
	}
	if !strings.Contains(spec.Blocks[0].Content, "[[ Slack Policies ]]") {
		t.Fatalf("slack policy heading missing: %q", spec.Blocks[0].Content)
	}
	if !strings.Contains(spec.Blocks[0].Content, "[[ Slack DM Policies ]]") {
		t.Fatalf("dm policy block missing marker: %q", spec.Blocks[0].Content)
	}
	if strings.Contains(spec.Blocks[0].Content, "Use only these emoji names for `message_react`:") {
		t.Fatalf("emoji list line should be omitted when list is empty: %q", spec.Blocks[0].Content)
	}
}

func TestAppendLineRuntimeBlocks_Group(t *testing.T) {
	spec := agent.PromptSpec{}

	AppendLineRuntimeBlocks(&spec, true)

	if len(spec.Blocks) != 1 {
		t.Fatalf("blocks len = %d, want 1", len(spec.Blocks))
	}
	if !strings.Contains(spec.Blocks[0].Content, "[[ LINE Policies ]]") {
		t.Fatalf("line policy heading missing: %q", spec.Blocks[0].Content)
	}
	if !strings.Contains(spec.Blocks[0].Content, "[[ LINE Group Policies ]]") {
		t.Fatalf("group policy block missing marker: %q", spec.Blocks[0].Content)
	}
}

func TestAppendLineRuntimeBlocks_Private(t *testing.T) {
	spec := agent.PromptSpec{}

	AppendLineRuntimeBlocks(&spec, false)

	if len(spec.Blocks) != 1 {
		t.Fatalf("blocks len = %d, want 1", len(spec.Blocks))
	}
	if !strings.Contains(spec.Blocks[0].Content, "[[ LINE Policies ]]") {
		t.Fatalf("line policy heading missing: %q", spec.Blocks[0].Content)
	}
	if !strings.Contains(spec.Blocks[0].Content, "[[ LINE Private Policies ]]") {
		t.Fatalf("private policy block missing marker: %q", spec.Blocks[0].Content)
	}
}

func TestAppendLarkRuntimeBlocks_Group(t *testing.T) {
	spec := agent.PromptSpec{}

	AppendLarkRuntimeBlocks(&spec, true)

	if len(spec.Blocks) != 1 {
		t.Fatalf("blocks len = %d, want 1", len(spec.Blocks))
	}
	if !strings.Contains(spec.Blocks[0].Content, "[[ Lark Policies ]]") {
		t.Fatalf("lark policy heading missing: %q", spec.Blocks[0].Content)
	}
	if !strings.Contains(spec.Blocks[0].Content, "[[ Lark Group Policies ]]") {
		t.Fatalf("group policy block missing marker: %q", spec.Blocks[0].Content)
	}
}

func TestAppendLarkRuntimeBlocks_Private(t *testing.T) {
	spec := agent.PromptSpec{}

	AppendLarkRuntimeBlocks(&spec, false)

	if len(spec.Blocks) != 1 {
		t.Fatalf("blocks len = %d, want 1", len(spec.Blocks))
	}
	if !strings.Contains(spec.Blocks[0].Content, "[[ Lark Policies ]]") {
		t.Fatalf("lark policy heading missing: %q", spec.Blocks[0].Content)
	}
	if !strings.Contains(spec.Blocks[0].Content, "[[ Lark Private Policies ]]") {
		t.Fatalf("private policy block missing marker: %q", spec.Blocks[0].Content)
	}
}
