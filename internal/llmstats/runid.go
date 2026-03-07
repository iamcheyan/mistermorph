package llmstats

import (
	"strings"

	"github.com/google/uuid"
)

func NewSyntheticRunID(prefix string) string {
	prefix = strings.ToLower(strings.TrimSpace(prefix))
	if prefix == "" {
		prefix = "run"
	}
	return prefix + "_" + uuid.NewString()
}
