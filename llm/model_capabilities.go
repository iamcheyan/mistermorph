package llm

import "strings"

func ModelSupportsImageParts(model string) bool {
	model = strings.ToLower(strings.TrimSpace(model))
	if model == "" {
		return false
	}
	if strings.HasPrefix(model, "gpt-") || strings.Contains(model, "/gpt-") {
		return true
	}
	if strings.HasPrefix(model, "gemini") || strings.Contains(model, "/gemini") {
		return true
	}
	return false
}
