package entryutil

import (
	"context"
	"log/slog"
	"sort"
	"strings"
	"time"
)

const TimestampLayout = "2006-01-02 15:04"

type SemanticItem struct {
	CreatedAt string
	Content   string
}

type SemanticResolver interface {
	SelectDedupKeepIndices(ctx context.Context, items []SemanticItem) ([]int, error)
}

func IsValidTimestamp(v string) bool {
	v = strings.TrimSpace(v)
	if v == "" {
		return false
	}
	_, err := time.Parse(TimestampLayout, v)
	return err == nil
}

func ResolveKeepIndices(ctx context.Context, items []SemanticItem, resolver SemanticResolver) ([]int, error) {
	if len(items) == 0 {
		return nil, nil
	}
	if len(items) == 1 {
		return []int{0}, nil
	}
	if resolver == nil {
		slog.Error("failed to use resolver, it's nil")
		return []int{0}, nil
	}

	keepIndices, err := resolver.SelectDedupKeepIndices(ctx, items)
	if err != nil {
		return nil, err
	}

	seen := make(map[int]bool, len(keepIndices))
	out := make([]int, 0, len(keepIndices))
	for _, idx := range keepIndices {
		if idx < 0 || idx >= len(items) {
			continue
		}
		if seen[idx] {
			continue
		}
		seen[idx] = true
		out = append(out, idx)
	}
	if !seen[0] {
		out = append(out, 0)
	}
	sort.Ints(out)
	return out, nil
}
