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
	if spec.Blocks[0].Title != SlackRuntimePromptBlockTitle {
		t.Fatalf("first block title = %q, want %q", spec.Blocks[0].Title, SlackRuntimePromptBlockTitle)
	}
	if !strings.Contains(spec.Blocks[0].Content, "<Slack Group Policies>") {
		t.Fatalf("group policy block missing marker: %q", spec.Blocks[0].Content)
	}
	if spec.Blocks[1].Title != slackMentionsPromptBlockTitle {
		t.Fatalf("second block title = %q, want %q", spec.Blocks[1].Title, slackMentionsPromptBlockTitle)
	}
	if spec.Blocks[1].Content != "U111\nU222" {
		t.Fatalf("mention block content = %q, want %q", spec.Blocks[1].Content, "U111\nU222")
	}
}

func TestAppendSlackRuntimeBlocks_DM(t *testing.T) {
	spec := agent.PromptSpec{}

	AppendSlackRuntimeBlocks(&spec, false, []string{"U111"})

	if len(spec.Blocks) != 1 {
		t.Fatalf("blocks len = %d, want 1", len(spec.Blocks))
	}
	if spec.Blocks[0].Title != SlackRuntimePromptBlockTitle {
		t.Fatalf("block title = %q, want %q", spec.Blocks[0].Title, SlackRuntimePromptBlockTitle)
	}
	if !strings.Contains(spec.Blocks[0].Content, "<Slack DM Policies>") {
		t.Fatalf("dm policy block missing marker: %q", spec.Blocks[0].Content)
	}
}
