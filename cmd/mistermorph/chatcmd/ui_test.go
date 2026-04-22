package chatcmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestNormalizeThinkingMessage(t *testing.T) {
	got := normalizeThinkingMessage(" \n  plan:\tCreate VPC  \n with subnet \r\n")
	want := "plan: Create VPC with subnet"
	if got != want {
		t.Fatalf("normalizeThinkingMessage() = %q, want %q", got, want)
	}
}

func TestNormalizeThinkingMessageEmptyFallback(t *testing.T) {
	got := normalizeThinkingMessage(" \t\r\n ")
	want := "assistant is thinking..."
	if got != want {
		t.Fatalf("normalizeThinkingMessage() = %q, want %q", got, want)
	}
}

func TestTruncateDisplayWidth(t *testing.T) {
	got := truncateDisplayWidth("plan: Create VPC with public and private subnets", 20)
	if stringDisplayWidth(got) > 20 {
		t.Fatalf("truncateDisplayWidth() width = %d, want <= 20", stringDisplayWidth(got))
	}
	if !strings.HasSuffix(got, "...") {
		t.Fatalf("truncateDisplayWidth() = %q, want ellipsis suffix", got)
	}
}

func TestIsTerminalWriter(t *testing.T) {
	if isTerminalWriter(&bytes.Buffer{}) {
		t.Fatal("isTerminalWriter(bytes.Buffer) = true, want false")
	}
}

func TestPrintChatSessionHeaderCompactModeSuppressesOutput(t *testing.T) {
	var buf bytes.Buffer
	printChatSessionHeader(&buf, true, "gpt-test", "/tmp/project")
	got := buf.String()
	for _, want := range []string{
		"model=gpt-test",
		"file_cache_dir=/tmp/project",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("printChatSessionHeader() = %q, want substring %q", got, want)
		}
	}
	for _, unwanted := range []string{
		"▄▄   ▄▄",
		"Interactive chat started. Press Ctrl+C or type /exit to quit.",
	} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("printChatSessionHeader() = %q, want no substring %q", got, unwanted)
		}
	}
}

func TestPrintChatSessionHeaderDefaultModeIncludesBannerAndContext(t *testing.T) {
	var buf bytes.Buffer
	printChatSessionHeader(&buf, false, "gpt-test", "/tmp/project")
	got := buf.String()
	for _, want := range []string{
		"▄▄   ▄▄",
		"model=gpt-test",
		"file_cache_dir=/tmp/project",
		"Interactive chat started. Press Ctrl+C or type /exit to quit.",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("printChatSessionHeader() = %q, want substring %q", got, want)
		}
	}
}

func TestPrintChatSessionHeaderDisplaysBedrockModelNameFromARN(t *testing.T) {
	var buf bytes.Buffer
	printChatSessionHeader(&buf, true, "arn:aws:bedrock:ap-northeast-1::foundation-model/moonshotai.kimi-k2.5", "/tmp/project")
	got := buf.String()
	if !strings.Contains(got, "model=moonshotai.kimi-k2.5") {
		t.Fatalf("printChatSessionHeader() = %q, want bedrock display model name", got)
	}
}
