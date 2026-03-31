package llmutil

import (
	"context"
	"hash/fnv"
	"io"
	"log/slog"
	"strings"

	"github.com/quailyquaily/mistermorph/internal/llmconfig"
	"github.com/quailyquaily/mistermorph/internal/llmstats"
	"github.com/quailyquaily/mistermorph/llm"
)

type weightedRouteCandidate struct {
	Profile string
	Model   string
	Weight  int
	Client  llm.Client
}

type weightedRouteClient struct {
	identity   string
	candidates []weightedRouteCandidate
	fallbacks  []FallbackCandidate
	logger     *slog.Logger
}

func buildWeightedRouteClient(route ResolvedRoute, primaryOverride *llmconfig.ClientConfig, build BaseClientBuilder, wrap ClientWrapFunc, logger *slog.Logger) (llm.Client, error) {
	candidates := make([]weightedRouteCandidate, 0, len(route.Candidates))
	for _, candidate := range route.Candidates {
		cfg := candidate.ClientConfig
		if primaryOverride != nil {
			cfg = mergeClientConfig(cfg, *primaryOverride)
		}
		client, err := build(cfg, candidate.Values)
		if err != nil {
			return nil, err
		}
		if wrap != nil {
			client = wrap(client, cfg, candidate.Profile)
		}
		candidates = append(candidates, weightedRouteCandidate{
			Profile: candidate.Profile,
			Model:   strings.TrimSpace(cfg.Model),
			Weight:  candidate.Weight,
			Client:  client,
		})
	}

	fallbacks := make([]FallbackCandidate, 0, len(route.Fallbacks))
	for _, fallback := range route.Fallbacks {
		client, err := build(fallback.ClientConfig, fallback.Values)
		if err != nil {
			return nil, err
		}
		if wrap != nil {
			client = wrap(client, fallback.ClientConfig, fallback.Profile)
		}
		fallbacks = append(fallbacks, FallbackCandidate{
			Profile: fallback.Profile,
			Model:   strings.TrimSpace(fallback.ClientConfig.Model),
			Client:  client,
		})
	}

	return &weightedRouteClient{
		identity:   strings.TrimSpace(route.Identity),
		candidates: candidates,
		fallbacks:  fallbacks,
		logger:     logger,
	}, nil
}

func mergeClientConfig(base llmconfig.ClientConfig, override llmconfig.ClientConfig) llmconfig.ClientConfig {
	if provider := strings.TrimSpace(override.Provider); provider != "" {
		base.Provider = provider
	}
	if endpoint := strings.TrimSpace(override.Endpoint); endpoint != "" {
		base.Endpoint = endpoint
	}
	if apiKey := strings.TrimSpace(override.APIKey); apiKey != "" {
		base.APIKey = apiKey
	}
	if model := strings.TrimSpace(override.Model); model != "" {
		base.Model = model
	}
	if override.RequestTimeout > 0 {
		base.RequestTimeout = override.RequestTimeout
	}
	return base
}

func (c *weightedRouteClient) Chat(ctx context.Context, req llm.Request) (llm.Result, error) {
	if c == nil || len(c.candidates) == 0 {
		return llm.Result{}, io.ErrClosedPipe
	}
	primaryIdx := c.pickPrimaryIndex(ctx, req)
	primary := c.candidates[primaryIdx]
	fallbacks := make([]FallbackCandidate, 0, len(c.candidates)-1+len(c.fallbacks))
	for idx, candidate := range c.candidates {
		if idx == primaryIdx {
			continue
		}
		fallbacks = append(fallbacks, FallbackCandidate{
			Profile: candidate.Profile,
			Model:   candidate.Model,
			Client:  candidate.Client,
		})
	}
	fallbacks = append(fallbacks, c.fallbacks...)
	client := NewFallbackClient(FallbackClientOptions{
		Primary:        primary.Client,
		PrimaryProfile: primary.Profile,
		PrimaryModel:   primary.Model,
		Fallbacks:      fallbacks,
		Logger:         c.logger,
	})
	fallbackReq := req
	fallbackReq.Model = primary.Model
	return client.Chat(ctx, fallbackReq)
}

func (c *weightedRouteClient) Close() error {
	if c == nil {
		return nil
	}
	var firstErr error
	closeClient := func(client llm.Client) {
		closer, ok := client.(io.Closer)
		if !ok {
			return
		}
		if err := closer.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	for _, candidate := range c.candidates {
		closeClient(candidate.Client)
	}
	for _, fallback := range c.fallbacks {
		closeClient(fallback.Client)
	}
	return firstErr
}

func (c *weightedRouteClient) pickPrimaryIndex(ctx context.Context, req llm.Request) int {
	totalWeight := 0
	for _, candidate := range c.candidates {
		if candidate.Weight > 0 {
			totalWeight += candidate.Weight
		}
	}
	if totalWeight <= 0 {
		return 0
	}
	key := selectionKey(ctx, req)
	if key == "" {
		return 0
	}
	hasher := fnv.New32a()
	_, _ = hasher.Write([]byte(key))
	target := int(hasher.Sum32() % uint32(totalWeight))
	acc := 0
	for idx, candidate := range c.candidates {
		acc += candidate.Weight
		if target < acc {
			return idx
		}
	}
	return 0
}

func selectionKey(ctx context.Context, req llm.Request) string {
	if runID := strings.TrimSpace(llmstats.RunIDFromContext(ctx)); runID != "" {
		return runID
	}
	if originEventID := strings.TrimSpace(llmstats.OriginEventIDFromContext(ctx)); originEventID != "" {
		return originEventID
	}
	if scene := strings.TrimSpace(req.Scene); scene != "" {
		return scene
	}
	return ""
}
