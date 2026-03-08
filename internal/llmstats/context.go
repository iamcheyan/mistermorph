package llmstats

import (
	"context"
	"strings"

	"github.com/quailyquaily/mistermorph/internal/llminspect"
)

type runIDContextKey struct{}
type originEventIDContextKey struct{}

func WithRunID(ctx context.Context, runID string) context.Context {
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return ctx
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, runIDContextKey{}, runID)
}

func RunIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v := ctx.Value(runIDContextKey{}); v != nil {
		if s, ok := v.(string); ok {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

func WithOriginEventID(ctx context.Context, eventID string) context.Context {
	eventID = strings.TrimSpace(eventID)
	if eventID == "" {
		return ctx
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, originEventIDContextKey{}, eventID)
}

func OriginEventIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v := ctx.Value(originEventIDContextKey{}); v != nil {
		if s, ok := v.(string); ok {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

func WithMetadata(ctx context.Context, runID string, originEventID string) context.Context {
	ctx = WithRunID(ctx, runID)
	ctx = WithOriginEventID(ctx, originEventID)
	return ctx
}

func WithScene(ctx context.Context, scene string) context.Context {
	return llminspect.WithModelScene(ctx, scene)
}
