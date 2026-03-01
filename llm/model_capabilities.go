package llm

import (
	"strconv"
	"strings"
)

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
	if grok4OrAbove(model) {
		return true
	}
	return false
}

func grok4OrAbove(model string) bool {
	if major, ok := parseGrokMajor(model); ok && major >= 4 {
		return true
	}
	if idx := strings.Index(model, "/grok-"); idx >= 0 {
		if major, ok := parseGrokMajor(model[idx+1:]); ok && major >= 4 {
			return true
		}
	}
	return false
}

func parseGrokMajor(model string) (int, bool) {
	model = strings.TrimSpace(model)
	if !strings.HasPrefix(model, "grok-") {
		return 0, false
	}
	rest := model[len("grok-"):]
	if rest == "" {
		return 0, false
	}
	end := 0
	for end < len(rest) && rest[end] >= '0' && rest[end] <= '9' {
		end++
	}
	if end == 0 {
		return 0, false
	}
	v, err := strconv.Atoi(rest[:end])
	if err != nil {
		return 0, false
	}
	return v, true
}
