package integration

import (
	"log/slog"
	"strings"

	"github.com/quailyquaily/mistermorph/internal/llmconfig"
	"github.com/quailyquaily/mistermorph/internal/llminspect"
	"github.com/quailyquaily/mistermorph/internal/llmutil"
	"github.com/quailyquaily/mistermorph/llm"
)

var integrationBaseClientBuilder = llmutil.ClientFromConfigWithValues

func buildIntegrationLLMClient(route llmutil.ResolvedRoute, logger *slog.Logger, wrap llmutil.ClientWrapFunc) (llm.Client, error) {
	return llmutil.BuildRouteClient(route, nil, integrationBaseClientBuilder, wrap, logger)
}

func inspectClientWrap(promptInspector *llminspect.PromptInspector, requestInspector *llminspect.RequestInspector) llmutil.ClientWrapFunc {
	if promptInspector == nil && requestInspector == nil {
		return nil
	}
	return func(client llm.Client, cfg llmconfig.ClientConfig, _ string) llm.Client {
		return llminspect.WrapClient(client, llminspect.ClientOptions{
			PromptInspector:  promptInspector,
			RequestInspector: requestInspector,
			APIBase:          cfg.Endpoint,
			Model:            strings.TrimSpace(cfg.Model),
		})
	}
}
