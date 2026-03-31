package integration

import (
	"context"
	"errors"
	"testing"

	"github.com/quailyquaily/mistermorph/agent"
	"github.com/quailyquaily/mistermorph/internal/llmconfig"
	"github.com/quailyquaily/mistermorph/internal/llmutil"
	"github.com/quailyquaily/mistermorph/llm"
)

type stubIntegrationLLMClient struct {
	chatFn func(context.Context, llm.Request) (llm.Result, error)
}

func (c *stubIntegrationLLMClient) Chat(ctx context.Context, req llm.Request) (llm.Result, error) {
	if c == nil || c.chatFn == nil {
		return llm.Result{}, nil
	}
	return c.chatFn(ctx, req)
}

func TestNewRunEngineUsesFallbackProfiles(t *testing.T) {
	oldBuilder := integrationBaseClientBuilder
	t.Cleanup(func() {
		integrationBaseClientBuilder = oldBuilder
	})

	buildModels := make([]string, 0, 2)
	requestModels := make([]string, 0, 2)
	integrationBaseClientBuilder = func(cfg llmconfig.ClientConfig, _ llmutil.RuntimeValues) (llm.Client, error) {
		buildModels = append(buildModels, cfg.Model)
		switch cfg.Model {
		case "gpt-5.2":
			return &stubIntegrationLLMClient{
				chatFn: func(_ context.Context, req llm.Request) (llm.Result, error) {
					requestModels = append(requestModels, req.Model)
					return llm.Result{}, errors.New("openai API request failed with status 429: too many requests")
				},
			}, nil
		case "gpt-4.1-mini":
			return &stubIntegrationLLMClient{
				chatFn: func(_ context.Context, req llm.Request) (llm.Result, error) {
					requestModels = append(requestModels, req.Model)
					return llm.Result{Text: `{"type":"final","output":"fallback ok"}`}, nil
				},
			}, nil
		default:
			t.Fatalf("unexpected model build %q", cfg.Model)
			return nil, nil
		}
	}

	cfg := DefaultConfig()
	cfg.Features.Skills = false
	cfg.Set("llm.provider", "openai")
	cfg.Set("llm.model", "gpt-5.2")
	cfg.Set("llm.profiles", map[string]any{
		"cheap": map[string]any{
			"model": "gpt-4.1-mini",
		},
	})
	cfg.Set("llm.routes", map[string]any{
		"main_loop": map[string]any{
			"fallback_profiles": []string{"cheap"},
		},
	})

	rt := New(cfg)
	prepared, err := rt.NewRunEngine(context.Background(), "ping")
	if err != nil {
		t.Fatalf("NewRunEngine() error = %v", err)
	}
	defer func() { _ = prepared.Cleanup() }()

	final, _, err := prepared.Engine.Run(context.Background(), "ping", agent.RunOptions{Model: prepared.Model})
	if err != nil {
		t.Fatalf("Engine.Run() error = %v", err)
	}
	if final == nil {
		t.Fatal("final is nil")
	}
	if final.Output != "fallback ok" {
		t.Fatalf("final output = %q, want fallback ok", final.Output)
	}
	if len(buildModels) < 2 {
		t.Fatalf("build models = %#v, want at least primary + fallback", buildModels)
	}
	if buildModels[0] != "gpt-5.2" || buildModels[1] != "gpt-4.1-mini" {
		t.Fatalf("build models prefix = %#v, want [gpt-5.2 gpt-4.1-mini]", buildModels[:2])
	}
	if len(requestModels) != 2 {
		t.Fatalf("request models = %#v, want primary + fallback", requestModels)
	}
	if requestModels[0] != "gpt-5.2" || requestModels[1] != "gpt-4.1-mini" {
		t.Fatalf("request models = %#v, want [gpt-5.2 gpt-4.1-mini]", requestModels)
	}
}

func TestSharedDependenciesCreateLLMClientUsesFallbackProfiles(t *testing.T) {
	oldBuilder := integrationBaseClientBuilder
	t.Cleanup(func() {
		integrationBaseClientBuilder = oldBuilder
	})

	buildModels := make([]string, 0, 2)
	integrationBaseClientBuilder = func(cfg llmconfig.ClientConfig, _ llmutil.RuntimeValues) (llm.Client, error) {
		buildModels = append(buildModels, cfg.Model)
		switch cfg.Model {
		case "gpt-5.2":
			return &stubIntegrationLLMClient{
				chatFn: func(_ context.Context, req llm.Request) (llm.Result, error) {
					if req.Model != "gpt-5.2" {
						t.Fatalf("primary request model = %q, want gpt-5.2", req.Model)
					}
					return llm.Result{}, errors.New("openai API request failed with status 500: upstream error")
				},
			}, nil
		case "gpt-4.1-mini":
			return &stubIntegrationLLMClient{
				chatFn: func(_ context.Context, req llm.Request) (llm.Result, error) {
					if req.Model != "gpt-4.1-mini" {
						t.Fatalf("fallback request model = %q, want gpt-4.1-mini", req.Model)
					}
					return llm.Result{Text: req.Model}, nil
				},
			}, nil
		default:
			t.Fatalf("unexpected model build %q", cfg.Model)
			return nil, nil
		}
	}

	cfg := DefaultConfig()
	cfg.Set("llm.provider", "openai")
	cfg.Set("llm.model", "gpt-5.2")
	cfg.Set("llm.profiles", map[string]any{
		"cheap": map[string]any{
			"model": "gpt-4.1-mini",
		},
	})
	cfg.Set("llm.routes", map[string]any{
		"main_loop": map[string]any{
			"fallback_profiles": []string{"cheap"},
		},
	})

	rt := New(cfg)
	snap := rt.snapshot()
	deps := rt.sharedDependencies(snap)
	route, err := deps.ResolveLLMRoute(llmutil.RoutePurposeMainLoop)
	if err != nil {
		t.Fatalf("ResolveLLMRoute() error = %v", err)
	}
	client, err := deps.CreateLLMClient(route)
	if err != nil {
		t.Fatalf("CreateLLMClient() error = %v", err)
	}

	result, err := client.Chat(context.Background(), llm.Request{Model: route.ClientConfig.Model})
	if err != nil {
		t.Fatalf("Chat() error = %v", err)
	}
	if result.Text != "gpt-4.1-mini" {
		t.Fatalf("result text = %q, want gpt-4.1-mini", result.Text)
	}
	if len(buildModels) < 2 {
		t.Fatalf("build models = %#v, want at least primary + fallback", buildModels)
	}
	if buildModels[0] != "gpt-5.2" || buildModels[1] != "gpt-4.1-mini" {
		t.Fatalf("build models prefix = %#v, want [gpt-5.2 gpt-4.1-mini]", buildModels[:2])
	}
}
