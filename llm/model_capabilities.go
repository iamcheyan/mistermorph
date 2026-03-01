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
	if claude3OrAbove(model) {
		return true
	}
	if grok4OrAbove(model) {
		return true
	}
	return false
}

func claude3OrAbove(model string) bool {
	if major, ok := parseFamilyMajor(model, "claude"); ok && major >= 3 {
		return true
	}
	if idx := strings.Index(model, "/claude"); idx >= 0 {
		if major, ok := parseFamilyMajor(model[idx+1:], "claude"); ok && major >= 3 {
			return true
		}
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

func parseFamilyMajor(model string, family string) (int, bool) {
	model = strings.TrimSpace(model)
	family = strings.TrimSpace(family)
	if family == "" || !strings.HasPrefix(model, family) {
		return 0, false
	}
	rest := model[len(family):]
	for i := 0; i < len(rest); i++ {
		if rest[i] < '0' || rest[i] > '9' {
			continue
		}
		j := i
		for j < len(rest) && rest[j] >= '0' && rest[j] <= '9' {
			j++
		}
		v, err := strconv.Atoi(rest[i:j])
		if err != nil {
			return 0, false
		}
		return v, true
	}
	return 0, false
}

func parseGrokMajor(model string) (int, bool) {
	return parseFamilyMajor(strings.TrimSpace(model), "grok-")
}
