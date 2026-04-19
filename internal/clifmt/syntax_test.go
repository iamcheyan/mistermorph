package clifmt

import (
	"strings"
	"testing"
)

func TestHighlightCodeBlocksWithColor(t *testing.T) {
	// We can't easily mock useColor() since it reads env directly,
	// so we test the highlightCode function directly.

	src := `print("Hello, World!")`
	result, err := highlightCode(src, "python")
	if err != nil {
		t.Fatalf("highlightCode failed: %v", err)
	}
	t.Logf("Highlighted result:\n%s", result)

	// Should contain ANSI escape sequences
	if !strings.Contains(result, "\x1b[") {
		t.Error("Expected ANSI escape codes in highlighted output")
	}
}

func TestHighlightCodeBlocksNoColor(t *testing.T) {
	// When useColor() is false, text should pass through unchanged
	text := "Here is some code:\n\n```python\nprint(\"Hello, World!\")\n```\n\nAnd some more text."

	// Since we're in a non-terminal test environment, useColor() returns false
	result := HighlightCodeBlocks(text)

	// Should be unchanged
	if result != text {
		t.Errorf("Expected unchanged text in non-terminal, got:\n%s", result)
	}
}

func TestLooksLikeCode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name: "plain python code",
			input: `def hello():
    print("Hello, World!")
    return 42`,
			expected: true,
		},
		{
			name: "plain go code",
			input: `func main() {
    fmt.Println("hello")
}`,
			expected: true,
		},
		{
			name: "plain text paragraph",
			input: `This is just a regular paragraph of text.
It has multiple lines but no code indicators.
Nothing special here.`,
			expected: false,
		},
		{
			name: "markdown headers",
			input: `# Title
## Subtitle
Some text here.`,
			expected: false,
		},
		{
			name: "single line",
			input: `fmt.Println("hello")`,
			expected: false,
		},
		{
			name: "javascript with const",
			input: `const x = 10;
function foo() {
    return x + 1;
}`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := looksLikeCode(tt.input)
			if result != tt.expected {
				t.Errorf("looksLikeCode() = %v, want %v for input:\n%s", result, tt.expected, tt.input)
			}
		})
	}
}
