package clifmt

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"
)

// RenderMarkdown renders markdown text for terminal display using Glamour,
// while preserving our custom boxed code blocks.
func RenderMarkdown(text string) string {
	if !useColor() {
		return text
	}

	// 1. Extract and process explicit markdown code blocks
	var codeBlocks []string
	placeholderFormat := "MM_CODE_BLOCK_%d_MM"

	processedText := codeBlockRe.ReplaceAllStringFunc(text, func(block string) string {
		highlighted := HighlightCodeBlocks(block)
		codeBlocks = append(codeBlocks, highlighted)
		return fmt.Sprintf(placeholderFormat, len(codeBlocks)-1)
	})

	// 2. Detect and extract plain code blocks (no markdown fences)
	// Split by double newlines to find potential code paragraphs
	paragraphs := strings.Split(processedText, "\n\n")
	for i, para := range paragraphs {
		trimmed := strings.TrimSpace(para)
		if trimmed == "" || strings.Contains(para, fmt.Sprintf(placeholderFormat, 0)) {
			continue // skip empty or already-processed paragraphs
		}
		if looksLikeCodeBlock(trimmed) {
			highlighted := HighlightCodeBlocks(trimmed)
			codeBlocks = append(codeBlocks, highlighted)
			paragraphs[i] = fmt.Sprintf(placeholderFormat, len(codeBlocks)-1)
		}
	}
	processedText = strings.Join(paragraphs, "\n\n")

	// 3. Render the remaining markdown with Glamour
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(0), // Handled by terminal or LLM output
	)
	if err != nil {
		// Fallback to our custom high-lighter if Glamour fails
		return HighlightCodeBlocks(text)
	}

	rendered, err := r.Render(processedText)
	if err != nil {
		return HighlightCodeBlocks(text)
	}

	// 4. Re-insert the boxed code blocks into the rendered output
	for i, block := range codeBlocks {
		placeholder := fmt.Sprintf(placeholderFormat, i)
		rendered = strings.Replace(rendered, placeholder, block, 1)
	}

	return rendered
}
