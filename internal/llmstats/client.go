package llmstats

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/quailyquaily/mistermorph/internal/llminspect"
	"github.com/quailyquaily/mistermorph/internal/statepaths"
	"github.com/quailyquaily/mistermorph/llm"
)

type ClientOptions struct {
	Provider           string
	APIBase            string
	DefaultModel       string
	JournalDir         string
	RotateMaxFileBytes int64
	Logger             *slog.Logger
}

type UsageClient struct {
	Base         llm.Client
	Journal      *Journal
	Provider     string
	APIBase      string
	DefaultModel string
	Logger       *slog.Logger
	now          func() time.Time
}

func WrapClient(base llm.Client, opts ClientOptions) llm.Client {
	if base == nil {
		return nil
	}
	journalDir := strings.TrimSpace(opts.JournalDir)
	if journalDir == "" {
		return base
	}
	return &UsageClient{
		Base:         base,
		Journal:      NewJournal(journalDir, JournalOptions{MaxFileBytes: opts.RotateMaxFileBytes}),
		Provider:     normalizeProvider(opts.Provider),
		APIBase:      normalizeAPIBase(opts.APIBase),
		DefaultModel: normalizeModel(opts.DefaultModel),
		Logger:       opts.Logger,
		now:          time.Now,
	}
}

func WrapRuntimeClient(base llm.Client, provider, apiBase, defaultModel string, logger *slog.Logger) llm.Client {
	return WrapClient(base, ClientOptions{
		Provider:     provider,
		APIBase:      apiBase,
		DefaultModel: defaultModel,
		JournalDir:   statepaths.LLMUsageJournalDir(),
		Logger:       logger,
	})
}

func (c *UsageClient) Chat(ctx context.Context, req llm.Request) (llm.Result, error) {
	if c == nil || c.Base == nil {
		return llm.Result{}, fmt.Errorf("usage client is not initialized")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	start := c.now()
	res, err := c.Base.Chat(ctx, req)
	if err != nil {
		return res, err
	}
	if c.Journal == nil {
		return res, nil
	}

	rec := normalizeRequestRecord(RequestRecord{
		TS:            c.now().UTC().Format(time.RFC3339),
		RunID:         RunIDFromContext(ctx),
		OriginEventID: OriginEventIDFromContext(ctx),
		Provider:      c.Provider,
		APIBase:       c.APIBase,
		Model:         firstNonEmpty(strings.TrimSpace(req.Model), c.DefaultModel),
		Scene:         llminspect.ModelSceneFromContext(ctx),
		InputTokens:   int64(res.Usage.InputTokens),
		OutputTokens:  int64(res.Usage.OutputTokens),
		TotalTokens:   int64(res.Usage.TotalTokens),
		CostUSD:       res.Usage.Cost,
		DurationMs:    durationMillis(res.Duration, c.now().Sub(start)),
	})
	if _, recErr := c.Journal.Append(rec); recErr != nil && c.Logger != nil {
		c.Logger.Warn(
			"llm_usage_record_error",
			"error", recErr.Error(),
			"provider", rec.Provider,
			"api_host", rec.APIHost,
			"model", rec.Model,
		)
	}
	return res, nil
}

func (c *UsageClient) SetDebugFn(fn func(label, payload string)) {
	if c == nil || c.Base == nil {
		return
	}
	if setter, ok := c.Base.(interface {
		SetDebugFn(func(label, payload string))
	}); ok {
		setter.SetDebugFn(fn)
	}
}

func (c *UsageClient) Close() error {
	if c == nil {
		return nil
	}
	var firstErr error
	if c.Journal != nil {
		if err := c.Journal.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if closer, ok := c.Base.(io.Closer); ok {
		if err := closer.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func firstNonEmpty(values ...string) string {
	for _, raw := range values {
		if s := strings.TrimSpace(raw); s != "" {
			return s
		}
	}
	return ""
}

func durationMillis(values ...time.Duration) int64 {
	for _, d := range values {
		if d > 0 {
			return d.Milliseconds()
		}
	}
	return 0
}
