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
	"github.com/quailyquaily/mistermorph/internal/toolsutil"
	"github.com/quailyquaily/mistermorph/llm"
	"github.com/quailyquaily/mistermorph/tools"
)

type PromptSpecFunc func(ctx context.Context, logger *slog.Logger, logOpts agent.LogOptions, task string, client llm.Client, model string, stickySkills []string) (agent.PromptSpec, []string, []string, error)

type CommonDependencies struct {
	Logger                 func() (*slog.Logger, error)
	LogOptions             func() agent.LogOptions
	CreateLLMClient        func(provider, endpoint, apiKey, model string, timeout time.Duration) (llm.Client, error)
	LLMProvider            func() string
	LLMEndpointForProvider func(provider string) string
	LLMAPIKeyForProvider   func(provider string) string
	LLMModelForProvider    func(provider string) string
	Registry               func() *tools.Registry
	RuntimeToolsConfig     toolsutil.RuntimeToolsRegisterConfig
	Guard                  func(logger *slog.Logger) *guard.Guard
	PromptSpec             PromptSpecFunc
}

type HeartbeatDependencies struct {
	Logger                 func() (*slog.Logger, error)
	LogOptions             func() agent.LogOptions
	CreateLLMClient        func(provider, endpoint, apiKey, model string, timeout time.Duration) (llm.Client, error)
	LLMProvider            func() string
	LLMEndpointForProvider func(provider string) string
	LLMAPIKeyForProvider   func(provider string) string
	LLMModelForProvider    func(provider string) string
	Registry               func() *tools.Registry
	RuntimeToolsConfig     toolsutil.RuntimeToolsRegisterConfig
	Guard                  func(logger *slog.Logger) *guard.Guard
	PromptSpec             PromptSpecFunc
	BuildHeartbeatTask     func(checklistPath string) (string, bool, error)
	BuildHeartbeatMeta     func(source string, interval time.Duration, checklistPath string, checklistEmpty bool, extra map[string]any) map[string]any
}

func CommonFromHeartbeat(d HeartbeatDependencies) CommonDependencies {
	return CommonDependencies{
		Logger:                 d.Logger,
		LogOptions:             d.LogOptions,
		CreateLLMClient:        d.CreateLLMClient,
		LLMProvider:            d.LLMProvider,
		LLMEndpointForProvider: d.LLMEndpointForProvider,
		LLMAPIKeyForProvider:   d.LLMAPIKeyForProvider,
		LLMModelForProvider:    d.LLMModelForProvider,
		Registry:               d.Registry,
		RuntimeToolsConfig:     d.RuntimeToolsConfig,
		Guard:                  d.Guard,
		PromptSpec:             d.PromptSpec,
	}
}

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

func LoggerFromCommon(d CommonDependencies) (*slog.Logger, error) {
	return Logger(d.Logger)
}

func LogOptionsFromCommon(d CommonDependencies) agent.LogOptions {
	return LogOptions(d.LogOptions)
}

func ProviderFromCommon(d CommonDependencies) string {
	return Provider(d.LLMProvider)
}

func EndpointForProviderFromCommon(d CommonDependencies, provider string) string {
	return ProviderField(d.LLMEndpointForProvider, provider)
}

func APIKeyForProviderFromCommon(d CommonDependencies, provider string) string {
	return ProviderField(d.LLMAPIKeyForProvider, provider)
}

func ModelForProviderFromCommon(d CommonDependencies, provider string) string {
	return ProviderField(d.LLMModelForProvider, provider)
}

func EndpointFromCommon(d CommonDependencies) string {
	return EndpointForProviderFromCommon(d, ProviderFromCommon(d))
}

func APIKeyFromCommon(d CommonDependencies) string {
	return APIKeyForProviderFromCommon(d, ProviderFromCommon(d))
}

func ModelFromCommon(d CommonDependencies) string {
	return ModelForProviderFromCommon(d, ProviderFromCommon(d))
}

func CreateClientFromCommon(d CommonDependencies, cfg llmconfig.ClientConfig) (llm.Client, error) {
	return CreateClient(d.CreateLLMClient, cfg)
}

func RegistryFromCommon(d CommonDependencies) *tools.Registry {
	return Registry(d.Registry)
}

func GuardFromCommon(d CommonDependencies, logger *slog.Logger) *guard.Guard {
	return Guard(d.Guard, logger)
}

func PromptSpecFromCommon(d CommonDependencies, ctx context.Context, logger *slog.Logger, logOpts agent.LogOptions, task string, client llm.Client, model string, stickySkills []string) (agent.PromptSpec, []string, []string, error) {
	return PromptSpec(d.PromptSpec, ctx, logger, logOpts, task, client, model, stickySkills)
}

func BuildHeartbeatTaskFromDeps(d HeartbeatDependencies, checklistPath string) (string, bool, error) {
	return BuildHeartbeatTask(d.BuildHeartbeatTask, checklistPath)
}

func BuildHeartbeatMetaFromDeps(d HeartbeatDependencies, source string, interval time.Duration, checklistPath string, checklistEmpty bool, extra map[string]any) map[string]any {
	return BuildHeartbeatMeta(d.BuildHeartbeatMeta, source, interval, checklistPath, checklistEmpty, extra)
}
