package telegram

import (
	"context"
	"log/slog"
	"time"

	"github.com/quailyquaily/mistermorph/agent"
	"github.com/quailyquaily/mistermorph/guard"
	"github.com/quailyquaily/mistermorph/internal/channelruntime/depsutil"
	"github.com/quailyquaily/mistermorph/internal/llmconfig"
	"github.com/quailyquaily/mistermorph/internal/toolsutil"
	"github.com/quailyquaily/mistermorph/llm"
	"github.com/quailyquaily/mistermorph/tools"
)

type Dependencies struct {
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
	PromptSpec             func(ctx context.Context, logger *slog.Logger, logOpts agent.LogOptions, task string, client llm.Client, model string, stickySkills []string) (agent.PromptSpec, []string, []string, error)
	BuildHeartbeatTask     func(checklistPath string) (string, bool, error)
	BuildHeartbeatMeta     func(source string, interval time.Duration, checklistPath string, checklistEmpty bool, extra map[string]any) map[string]any
}

func loggerFromDeps(d Dependencies) (*slog.Logger, error) {
	return depsutil.Logger(d.Logger)
}

func logOptionsFromDeps(d Dependencies) agent.LogOptions {
	return depsutil.LogOptions(d.LogOptions)
}

func llmProviderFromDeps(d Dependencies) string {
	return depsutil.Provider(d.LLMProvider)
}

func llmEndpointForProvider(d Dependencies, provider string) string {
	return depsutil.ProviderField(d.LLMEndpointForProvider, provider)
}

func llmAPIKeyForProvider(d Dependencies, provider string) string {
	return depsutil.ProviderField(d.LLMAPIKeyForProvider, provider)
}

func llmModelForProvider(d Dependencies, provider string) string {
	return depsutil.ProviderField(d.LLMModelForProvider, provider)
}

func llmEndpointFromDeps(d Dependencies) string {
	return llmEndpointForProvider(d, llmProviderFromDeps(d))
}

func llmAPIKeyFromDeps(d Dependencies) string {
	return llmAPIKeyForProvider(d, llmProviderFromDeps(d))
}

func llmModelFromDeps(d Dependencies) string {
	return llmModelForProvider(d, llmProviderFromDeps(d))
}

func llmClientFromConfig(d Dependencies, cfg llmconfig.ClientConfig) (llm.Client, error) {
	return depsutil.CreateClient(d.CreateLLMClient, cfg)
}

func registryFromDeps(d Dependencies) *tools.Registry {
	return depsutil.Registry(d.Registry)
}

func guardFromDeps(d Dependencies, log *slog.Logger) *guard.Guard {
	return depsutil.Guard(d.Guard, log)
}

func promptSpecForTelegram(d Dependencies, ctx context.Context, logger *slog.Logger, logOpts agent.LogOptions, task string, client llm.Client, model string, stickySkills []string) (agent.PromptSpec, []string, []string, error) {
	return depsutil.PromptSpec(d.PromptSpec, ctx, logger, logOpts, task, client, model, stickySkills)
}

func formatFinalOutput(final *agent.Final) string {
	return depsutil.FormatFinalOutput(final)
}

func formatRuntimeError(err error) string {
	return depsutil.FormatRuntimeError(err)
}

func shouldPublishTelegramText(final *agent.Final) bool {
	if final == nil {
		return true
	}
	return !final.IsLightweight
}

func buildHeartbeatTask(d Dependencies, checklistPath string) (string, bool, error) {
	return depsutil.BuildHeartbeatTask(d.BuildHeartbeatTask, checklistPath)
}

func buildHeartbeatMeta(d Dependencies, source string, interval time.Duration, checklistPath string, checklistEmpty bool, extra map[string]any) map[string]any {
	return depsutil.BuildHeartbeatMeta(d.BuildHeartbeatMeta, source, interval, checklistPath, checklistEmpty, extra)
}
