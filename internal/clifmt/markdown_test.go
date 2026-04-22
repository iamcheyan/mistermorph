package clifmt

import (
	"strings"
	"testing"
)

func TestRenderMarkdownReinsertsMultipleCodeBlocks(t *testing.T) {
	input := "Here:\n\n```sh\necho hi\n```\n\nAnd:\n\n```go\nfmt.Println(1)\n```"

	got := renderMarkdown(input, true)

	if strings.Contains(got, "MM_CODE_BLOCK") || strings.Contains(got, "§§MMCODEBLOCK") {
		t.Fatalf("renderMarkdown() leaked placeholder: %q", got)
	}
	if !strings.Contains(got, "echo hi") {
		t.Fatalf("renderMarkdown() missing first code block: %q", got)
	}
	if !strings.Contains(got, "fmt.Println") {
		t.Fatalf("renderMarkdown() missing second code block: %q", got)
	}
}
