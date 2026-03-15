package consolecmd

import (
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/quailyquaily/mistermorph/agent"
	"github.com/quailyquaily/mistermorph/internal/channelruntime/depsutil"
	"github.com/quailyquaily/mistermorph/internal/channelruntime/taskruntime"
	"github.com/quailyquaily/mistermorph/internal/llmconfig"
	"github.com/quailyquaily/mistermorph/internal/llmutil"
	"github.com/quailyquaily/mistermorph/internal/toolsutil"
	"github.com/quailyquaily/mistermorph/llm"
	"github.com/quailyquaily/mistermorph/tools"
)

func TestConsoleLLMCredentialsWarning_OpenAIWithoutAPIKey(t *testing.T) {
	route := llmutil.ResolvedRoute{
		ClientConfig: llmconfig.ClientConfig{
			Provider: "openai",
			Model:    "gpt-5.2",
		},
	}
	warning := consoleLLMCredentialsWarning(route)
	if !strings.Contains(warning, "MISTER_MORPH_LLM_API_KEY") {
		t.Fatalf("warning = %q, want env hint", warning)
	}
}

func TestConsoleLLMCredentialsWarning_OpenAIWithAPIKey(t *testing.T) {
	route := llmutil.ResolvedRoute{
		ClientConfig: llmconfig.ClientConfig{
			Provider: "openai",
			APIKey:   "dev-key",
			Model:    "gpt-5.2",
		},
	}
	if warning := consoleLLMCredentialsWarning(route); warning != "" {
		t.Fatalf("warning = %q, want empty", warning)
	}
}

func TestConsoleLLMCredentialsWarning_BedrockSkipsWarning(t *testing.T) {
	route := llmutil.ResolvedRoute{
		ClientConfig: llmconfig.ClientConfig{
			Provider: "bedrock",
			Model:    "anthropic.claude-3-7-sonnet",
		},
	}
	if warning := consoleLLMCredentialsWarning(route); warning != "" {
		t.Fatalf("warning = %q, want empty", warning)
	}
}

type stubConsoleLLMClient struct {
	requests []llm.Request
}

func (c *stubConsoleLLMClient) Chat(_ context.Context, req llm.Request) (llm.Result, error) {
	c.requests = append(c.requests, req)
	return llm.Result{Text: `{"type":"final","reasoning":"done","output":"ok"}`}, nil
}

func TestConsoleRunTaskUsesConfiguredClient(t *testing.T) {
	client := &stubConsoleLLMClient{}
	resolvedRoute := llmutil.ResolvedRoute{
		ClientConfig: llmconfig.ClientConfig{
			Provider: "openai",
			Model:    "gpt-5.2",
		},
	}
	execRuntime, err := taskruntime.Bootstrap(depsutil.CommonDependencies{
		Logger: func() (*slog.Logger, error) {
			return slog.Default(), nil
		},
		LogOptions: func() agent.LogOptions {
			return agent.LogOptions{}
		},
		ResolveLLMRoute: func(string) (llmutil.ResolvedRoute, error) {
			return resolvedRoute, nil
		},
		CreateLLMClient: func(llmutil.ResolvedRoute) (llm.Client, error) {
			return client, nil
		},
		Registry: func() *tools.Registry {
			return tools.NewRegistry()
		},
		RuntimeToolsConfig: toolsutil.RuntimeToolsRegisterConfig{},
		PromptSpec: func(_ context.Context, _ *slog.Logger, _ agent.LogOptions, _ string, _ llm.Client, _ string, _ []string) (agent.PromptSpec, []string, error) {
			return agent.DefaultPromptSpec(), nil, nil
		},
	}, taskruntime.BootstrapOptions{
		AgentConfig:          agent.Config{MaxSteps: 3, ParseRetries: 1, ToolRepeatLimit: 3},
		DefaultModelFallback: "gpt-5.2",
	})
	if err != nil {
		t.Fatalf("Bootstrap() error = %v", err)
	}
	rt := &consoleLocalRuntime{
		taskRuntime:  execRuntime,
		defaultModel: "gpt-5.2",
	}

	final, runCtx, err := rt.runTask(context.Background(), consoleConversationKey, consoleLocalTaskJob{
		TaskID: "task_console_1",
		Task:   "ping",
		Model:  "gpt-5.4",
	})
	if err != nil {
		t.Fatalf("runTask() error = %v", err)
	}
	if final == nil {
		t.Fatal("final is nil")
	}
	if runCtx == nil {
		t.Fatal("runCtx is nil")
	}
	if len(client.requests) != 1 {
		t.Fatalf("client requests = %d, want 1", len(client.requests))
	}
	if client.requests[0].Model != "gpt-5.4" {
		t.Fatalf("request model = %q, want %q", client.requests[0].Model, "gpt-5.4")
	}
	if client.requests[0].Scene != "console.loop" {
		t.Fatalf("request scene = %q, want %q", client.requests[0].Scene, "console.loop")
	}
}
