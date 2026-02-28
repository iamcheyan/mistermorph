package depsutil

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/quailyquaily/mistermorph/agent"
	"github.com/quailyquaily/mistermorph/guard"
	"github.com/quailyquaily/mistermorph/internal/llmconfig"
	"github.com/quailyquaily/mistermorph/internal/outputfmt"
	"github.com/quailyquaily/mistermorph/llm"
	"github.com/quailyquaily/mistermorph/tools"
)

type PromptSpecFunc func(ctx context.Context, logger *slog.Logger, logOpts agent.LogOptions, task string, client llm.Client, model string, stickySkills []string) (agent.PromptSpec, []string, []string, error)

func Logger(fn func() (*slog.Logger, error)) (*slog.Logger, error) {
	if fn == nil {
		return nil, fmt.Errorf("Logger dependency missing")
	}
	return fn()
}

func LogOptions(fn func() agent.LogOptions) agent.LogOptions {
	if fn == nil {
		return agent.LogOptions{}
	}
	return fn()
}

func Provider(fn func() string) string {
	if fn == nil {
		return ""
	}
	return fn()
}

func ProviderField(fn func(provider string) string, provider string) string {
	if fn == nil {
		return ""
	}
	return fn(provider)
}

func CreateClient(fn func(provider, endpoint, apiKey, model string, timeout time.Duration) (llm.Client, error), cfg llmconfig.ClientConfig) (llm.Client, error) {
	if fn == nil {
		return nil, fmt.Errorf("CreateLLMClient dependency missing")
	}
	return fn(cfg.Provider, cfg.Endpoint, cfg.APIKey, cfg.Model, cfg.RequestTimeout)
}

func Registry(fn func() *tools.Registry) *tools.Registry {
	if fn == nil {
		return nil
	}
	return fn()
}

func Guard(gf func(logger *slog.Logger) *guard.Guard, logger *slog.Logger) *guard.Guard {
	if gf == nil {
		return nil
	}
	return gf(logger)
}

func PromptSpec(fn PromptSpecFunc, ctx context.Context, logger *slog.Logger, logOpts agent.LogOptions, task string, client llm.Client, model string, stickySkills []string) (agent.PromptSpec, []string, []string, error) {
	if fn == nil {
		return agent.PromptSpec{}, nil, nil, fmt.Errorf("PromptSpec dependency missing")
	}
	return fn(ctx, logger, logOpts, task, client, model, stickySkills)
}

func BuildHeartbeatTask(fn func(checklistPath string) (string, bool, error), checklistPath string) (string, bool, error) {
	if fn == nil {
		return "", true, fmt.Errorf("BuildHeartbeatTask dependency missing")
	}
	return fn(checklistPath)
}

func BuildHeartbeatMeta(fn func(source string, interval time.Duration, checklistPath string, checklistEmpty bool, extra map[string]any) map[string]any, source string, interval time.Duration, checklistPath string, checklistEmpty bool, extra map[string]any) map[string]any {
	if fn == nil {
		return map[string]any{
			"trigger":   "heartbeat",
			"heartbeat": map[string]any{"source": source, "interval": interval.String()},
		}
	}
	return fn(source, interval, checklistPath, checklistEmpty, extra)
}

func FormatFinalOutput(final *agent.Final) string {
	return outputfmt.FormatFinalOutput(final)
}

func FormatRuntimeError(err error) string {
	s := strings.TrimSpace(outputfmt.FormatErrorForDisplay(err))
	if s != "" {
		return s
	}
	if err == nil {
		return "unknown error"
	}
	raw := strings.TrimSpace(err.Error())
	if raw == "" {
		return "unknown error"
	}
	return raw
}
