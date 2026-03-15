package consolecmd

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/quailyquaily/mistermorph/agent"
	"github.com/quailyquaily/mistermorph/guard"
	runtimecore "github.com/quailyquaily/mistermorph/internal/channelruntime/core"
	"github.com/quailyquaily/mistermorph/internal/channelruntime/depsutil"
	"github.com/quailyquaily/mistermorph/internal/channelruntime/taskruntime"
	"github.com/quailyquaily/mistermorph/internal/chathistory"
	"github.com/quailyquaily/mistermorph/internal/daemonruntime"
	"github.com/quailyquaily/mistermorph/internal/llmstats"
	"github.com/quailyquaily/mistermorph/internal/llmutil"
	"github.com/quailyquaily/mistermorph/internal/logutil"
	"github.com/quailyquaily/mistermorph/internal/mcphost"
	"github.com/quailyquaily/mistermorph/internal/memoryruntime"
	"github.com/quailyquaily/mistermorph/internal/outputfmt"
	"github.com/quailyquaily/mistermorph/internal/skillsutil"
	"github.com/quailyquaily/mistermorph/internal/toolsutil"
	"github.com/quailyquaily/mistermorph/llm"
	"github.com/quailyquaily/mistermorph/memory"
	"github.com/quailyquaily/mistermorph/tools"
	"github.com/spf13/viper"
)

const (
	consoleLocalEndpointRef   = "ep_console_local"
	consoleLocalEndpointName  = "Console Local"
	consoleLocalEndpointURL   = "in-process://console-local"
	consoleConversationKey    = "console:main"
	consoleTaskOutputMaxChars = 4000
)

type consoleLocalTaskJob struct {
	TaskID          string
	ConversationKey string
	Task            string
	Model           string
	Timeout         time.Duration
	CreatedAt       time.Time
	Version         uint64
}

type consoleLocalRuntime struct {
	logger                  *slog.Logger
	store                   *daemonruntime.MemoryStore
	runner                  *runtimecore.ConversationRunner[string, consoleLocalTaskJob]
	commonDeps              depsutil.CommonDependencies
	taskRuntime             *taskruntime.Runtime
	mcpHost                 *mcphost.Host
	memoryEnabled           bool
	defaultTimeout          time.Duration
	defaultModel            string
	defaultProvider         string
	memoryInjectionEnabled  bool
	memoryInjectionMaxItems int
	memRuntime              runtimecore.MemoryRuntime
	handler                 http.Handler
	authToken               string
	cancelWorkers           context.CancelFunc
	seq                     atomic.Uint64
}

func newConsoleLocalRuntime() (*consoleLocalRuntime, error) {
	logger, err := logutil.LoggerFromViper()
	if err != nil {
		return nil, err
	}
	slog.SetDefault(logger)
	logOpts := logutil.LogOptionsFromViper()

	llmValues := llmutil.RuntimeValuesFromViper()
	commonDeps := depsutil.CommonDependencies{
		Logger: func() (*slog.Logger, error) {
			return logger, nil
		},
		LogOptions: func() agent.LogOptions {
			return logOpts
		},
		ResolveLLMRoute: func(purpose string) (llmutil.ResolvedRoute, error) {
			return llmutil.ResolveRoute(llmValues, purpose)
		},
		CreateLLMClient: func(route llmutil.ResolvedRoute) (llm.Client, error) {
			base, err := llmutil.ClientFromConfigWithValues(route.ClientConfig, route.Values)
			if err != nil {
				return nil, err
			}
			return llmstats.WrapRuntimeClient(base, route.ClientConfig.Provider, route.ClientConfig.Endpoint, route.ClientConfig.Model, logger), nil
		},
		RuntimeToolsConfig: toolsutil.LoadRuntimeToolsRegisterConfigFromViper(),
		PromptSpec: func(ctx context.Context, logger *slog.Logger, logOpts agent.LogOptions, task string, client llm.Client, model string, stickySkills []string) (agent.PromptSpec, []string, error) {
			cfg := skillsutil.SkillsConfigFromViper()
			if len(stickySkills) > 0 {
				cfg.Requested = append(cfg.Requested, stickySkills...)
			}
			return skillsutil.PromptSpecWithSkills(ctx, logger, logOpts, task, client, model, cfg)
		},
	}

	baseRegistry, mcpHost := buildConsoleBaseRegistry(context.Background(), logger)
	sharedGuard := buildConsoleGuardFromViper(logger)
	commonDeps.Registry = func() *tools.Registry {
		return baseRegistry
	}
	commonDeps.Guard = func(_ *slog.Logger) *guard.Guard {
		return sharedGuard
	}
	execRuntime, err := taskruntime.Bootstrap(commonDeps, taskruntime.BootstrapOptions{
		AgentConfig:          consoleAgentConfigFromViper(),
		DefaultModelFallback: "gpt-5.2",
	})
	if err != nil {
		return nil, err
	}
	if warning := consoleLLMCredentialsWarning(execRuntime.MainRoute); warning != "" {
		logger.Warn("console_llm_credentials_missing",
			"provider", execRuntime.MainProvider,
			"hint", warning,
		)
	}

	defaultTimeout := viper.GetDuration("timeout")
	if defaultTimeout <= 0 {
		defaultTimeout = 10 * time.Minute
	}

	workersCtx, cancelWorkers := context.WithCancel(context.Background())
	memoryEnabled := viper.GetBool("memory.enabled")
	memRuntime, err := runtimecore.NewMemoryRuntime(commonDeps, runtimecore.MemoryRuntimeOptions{
		Enabled:       memoryEnabled,
		ShortTermDays: viper.GetInt("memory.short_term_days"),
		Logger:        logger,
	})
	if err != nil {
		cancelWorkers()
		return nil, err
	}
	if memRuntime.ProjectionWorker != nil {
		memRuntime.ProjectionWorker.Start(workersCtx)
	}

	authToken := strings.TrimSpace(viper.GetString("server.auth_token"))
	if authToken == "" {
		cancelWorkers()
		memRuntime.Cleanup()
		return nil, fmt.Errorf("missing server.auth_token (set via MISTER_MORPH_SERVER_AUTH_TOKEN) for console local runtime")
	}
	serverMaxQueue := viper.GetInt("server.max_queue")
	if serverMaxQueue <= 0 {
		serverMaxQueue = 100
	}

	store := daemonruntime.NewMemoryStore(serverMaxQueue)
	out := &consoleLocalRuntime{
		logger:                  logger,
		store:                   store,
		commonDeps:              commonDeps,
		taskRuntime:             execRuntime,
		mcpHost:                 mcpHost,
		memoryEnabled:           memoryEnabled,
		defaultTimeout:          defaultTimeout,
		defaultModel:            execRuntime.MainModel,
		defaultProvider:         execRuntime.MainProvider,
		memoryInjectionEnabled:  viper.GetBool("memory.injection.enabled"),
		memoryInjectionMaxItems: viper.GetInt("memory.injection.max_items"),
		memRuntime:              memRuntime,
		authToken:               authToken,
		cancelWorkers:           cancelWorkers,
	}
	out.runner = runtimecore.NewConversationRunner[string, consoleLocalTaskJob](
		workersCtx,
		make(chan struct{}, 1),
		16,
		func(workerCtx context.Context, conversationKey string, job consoleLocalTaskJob) {
			out.handleTaskJob(workerCtx, conversationKey, job)
		},
	)
	out.handler = daemonruntime.NewHandler(out.routesOptions(strings.TrimSpace(authToken)))
	return out, nil
}

func (r *consoleLocalRuntime) Close() {
	if r == nil {
		return
	}
	if r.cancelWorkers != nil {
		r.cancelWorkers()
	}
	if r.memRuntime.Cleanup != nil {
		r.memRuntime.Cleanup()
	}
	if r.mcpHost != nil {
		_ = r.mcpHost.Close()
	}
}

func consoleLLMCredentialsWarning(route llmutil.ResolvedRoute) string {
	provider := strings.ToLower(strings.TrimSpace(route.ClientConfig.Provider))
	switch provider {
	case "bedrock":
		// Bedrock may rely on ambient AWS credentials outside llm.* config.
		return ""
	case "cloudflare":
		if strings.TrimSpace(route.Values.CloudflareAccountID) == "" {
			return "set llm.cloudflare.account_id to enable Console Local chat submit"
		}
		if strings.TrimSpace(route.ClientConfig.APIKey) == "" {
			return "set llm.cloudflare.api_token, llm.api_key, or MISTER_MORPH_LLM_API_KEY to enable Console Local chat submit"
		}
		return ""
	default:
		if strings.TrimSpace(route.ClientConfig.APIKey) == "" {
			return "set llm.api_key or MISTER_MORPH_LLM_API_KEY to enable Console Local chat submit"
		}
		return ""
	}
}

func (r *consoleLocalRuntime) Endpoint() runtimeEndpoint {
	return runtimeEndpoint{
		Ref:    consoleLocalEndpointRef,
		Name:   consoleLocalEndpointName,
		URL:    consoleLocalEndpointURL,
		Client: newInProcessRuntimeEndpointClient(r.handler, r.authToken),
	}
}

func (r *consoleLocalRuntime) routesOptions(authToken string) daemonruntime.RoutesOptions {
	return daemonruntime.RoutesOptions{
		Mode:          "console",
		AuthToken:     strings.TrimSpace(authToken),
		TaskReader:    r.store,
		Submit:        r.submitTask,
		HealthEnabled: true,
		Overview: func(ctx context.Context) (map[string]any, error) {
			return map[string]any{
				"llm": map[string]any{
					"provider": r.defaultProvider,
					"model":    r.defaultModel,
				},
				"channel": map[string]any{
					"configured":          true,
					"telegram_configured": false,
					"slack_configured":    false,
					"running":             "console",
					"telegram_running":    false,
					"slack_running":       false,
				},
			}, nil
		},
	}
}

func (r *consoleLocalRuntime) submitTask(ctx context.Context, req daemonruntime.SubmitTaskRequest) (daemonruntime.SubmitTaskResponse, error) {
	timeout := r.defaultTimeout
	if strings.TrimSpace(req.Timeout) != "" {
		d, err := time.ParseDuration(strings.TrimSpace(req.Timeout))
		if err != nil || d <= 0 {
			return daemonruntime.SubmitTaskResponse{}, daemonruntime.BadRequest("invalid timeout (use Go duration like 2m, 30s)")
		}
		timeout = d
	}
	model := strings.TrimSpace(req.Model)
	if model == "" {
		model = r.defaultModel
	}

	now := time.Now().UTC()
	seq := r.seq.Add(1)
	taskID := daemonruntime.BuildTaskID("console", now.UnixNano(), seq, rand.Uint64())
	r.store.Upsert(daemonruntime.TaskInfo{
		ID:        taskID,
		Status:    daemonruntime.TaskQueued,
		Task:      strings.TrimSpace(req.Task),
		Model:     model,
		Timeout:   timeout.String(),
		CreatedAt: now,
	})

	err := r.runner.Enqueue(ctx, consoleConversationKey, func(version uint64) consoleLocalTaskJob {
		return consoleLocalTaskJob{
			TaskID:          taskID,
			ConversationKey: consoleConversationKey,
			Task:            strings.TrimSpace(req.Task),
			Model:           model,
			Timeout:         timeout,
			CreatedAt:       now,
			Version:         version,
		}
	})
	if err != nil {
		runtimecore.MarkTaskFailed(r.store, taskID, strings.TrimSpace(err.Error()), daemonruntime.IsContextDeadline(ctx, err))
		return daemonruntime.SubmitTaskResponse{}, err
	}
	return daemonruntime.SubmitTaskResponse{
		ID:     taskID,
		Status: daemonruntime.TaskQueued,
	}, nil
}

func (r *consoleLocalRuntime) handleTaskJob(workerCtx context.Context, conversationKey string, job consoleLocalTaskJob) {
	if r == nil {
		return
	}
	runtimecore.MarkTaskRunning(r.store, job.TaskID)

	runCtx, cancel := context.WithTimeout(workerCtx, job.Timeout)
	final, agentCtx, runErr := r.runTask(runCtx, conversationKey, job)
	contextDeadline := daemonruntime.IsContextDeadline(runCtx, runErr)
	cancel()

	if runErr != nil {
		displayErr := strings.TrimSpace(outputfmt.FormatErrorForDisplay(runErr))
		if displayErr == "" {
			displayErr = strings.TrimSpace(runErr.Error())
		}
		runtimecore.MarkTaskFailed(r.store, job.TaskID, displayErr, contextDeadline)
		return
	}

	if pendingID, ok := pendingApprovalID(final); ok {
		pendingAt := time.Now().UTC()
		r.store.Update(job.TaskID, func(info *daemonruntime.TaskInfo) {
			info.Status = daemonruntime.TaskPending
			info.PendingAt = &pendingAt
			info.ApprovalRequestID = pendingID
			info.Result = buildConsoleTaskResult(final, agentCtx)
		})
		return
	}

	finishedAt := time.Now().UTC()
	r.store.Update(job.TaskID, func(info *daemonruntime.TaskInfo) {
		output := strings.TrimSpace(outputfmt.FormatFinalOutput(final))
		info.Status = daemonruntime.TaskDone
		info.Error = ""
		info.FinishedAt = &finishedAt
		info.Result = buildConsoleTaskResult(final, agentCtx)
		if strings.TrimSpace(output) != "" {
			// Keep output summary for old readers that only consume result.output.
			if resultMap, ok := info.Result.(map[string]any); ok {
				resultMap["output"] = daemonruntime.TruncateUTF8(output, consoleTaskOutputMaxChars)
				info.Result = resultMap
			}
		}
	})
}

func (r *consoleLocalRuntime) runTask(ctx context.Context, conversationKey string, job consoleLocalTaskJob) (*agent.Final, *agent.Context, error) {
	if r == nil {
		return nil, nil, fmt.Errorf("console runtime is not initialized")
	}
	ctx = llmstats.WithRunID(ctx, job.TaskID)
	task := strings.TrimSpace(job.Task)
	if task == "" {
		return nil, nil, fmt.Errorf("empty console task")
	}
	model := strings.TrimSpace(job.Model)
	if model == "" {
		model = r.defaultModel
	}
	memSubjectID := buildConsoleMemorySubjectID(conversationKey)
	memoryHooks := taskruntime.MemoryHooks{
		Source:    "console",
		SubjectID: memSubjectID,
	}
	if r.memoryEnabled && r.memRuntime.Orchestrator != nil && memSubjectID != "" {
		memoryHooks.InjectionEnabled = r.memoryInjectionEnabled
		memoryHooks.InjectionMaxItems = r.memoryInjectionMaxItems
		memoryHooks.PrepareInjection = func(maxItems int) (string, error) {
			return r.memRuntime.Orchestrator.PrepareInjection(memoryruntime.PrepareInjectionRequest{
				SubjectID:      memSubjectID,
				RequestContext: memory.ContextPrivate,
				MaxItems:       maxItems,
			})
		}
		memoryHooks.Record = func(_ *agent.Final, finalOutput string) error {
			_, err := r.memRuntime.Orchestrator.Record(buildConsoleMemoryRecordRequest(job, memSubjectID, finalOutput))
			return err
		}
		memoryHooks.NotifyRecorded = func() {
			if r.memRuntime.ProjectionWorker != nil {
				r.memRuntime.ProjectionWorker.NotifyRecordAppended()
			}
		}
	}
	result, err := r.taskRuntime.Run(ctx, taskruntime.RunRequest{
		Task:  task,
		Model: model,
		Scene: "console.loop",
		Meta: map[string]any{
			"trigger":         "console",
			"console_task_id": job.TaskID,
		},
		Memory: memoryHooks,
	})
	if err != nil {
		return result.Final, result.Context, err
	}
	return result.Final, result.Context, nil
}

func buildConsoleTaskResult(final *agent.Final, runCtx *agent.Context) map[string]any {
	out := map[string]any{
		"final": final,
	}
	if runCtx != nil {
		out["metrics"] = runCtx.Metrics
		out["steps"] = summarizeConsoleSteps(runCtx)
	}
	return out
}

func summarizeConsoleSteps(ctx *agent.Context) []map[string]any {
	if ctx == nil || len(ctx.Steps) == 0 {
		return nil
	}
	out := make([]map[string]any, 0, len(ctx.Steps))
	for _, s := range ctx.Steps {
		m := map[string]any{
			"step":        s.StepNumber,
			"action":      s.Action,
			"duration_ms": s.Duration.Milliseconds(),
		}
		if s.Error != nil {
			m["error"] = s.Error.Error()
		}
		out = append(out, m)
	}
	return out
}

func buildConsoleMemorySubjectID(conversationKey string) string {
	key := strings.TrimSpace(conversationKey)
	if key == "" {
		key = "main"
	}
	if strings.HasPrefix(strings.ToLower(key), "console:") {
		return key
	}
	return "console:" + key
}

func buildConsoleMemoryRecordRequest(job consoleLocalTaskJob, subjectID, output string) memoryruntime.RecordRequest {
	subjectID = strings.TrimSpace(subjectID)
	if subjectID == "" {
		subjectID = "console:main"
	}
	now := time.Now().UTC()
	sentAt := job.CreatedAt
	if sentAt.IsZero() {
		sentAt = now
	}
	inbound := chathistory.ChatHistoryItem{
		Channel:   "console",
		Kind:      chathistory.KindInboundUser,
		ChatID:    subjectID,
		ChatType:  "private",
		MessageID: job.TaskID,
		SentAt:    sentAt,
		Sender: chathistory.ChatHistorySender{
			UserID:     "console:user",
			Username:   "console",
			Nickname:   "Console User",
			DisplayRef: "console:user",
		},
		Text: strings.TrimSpace(job.Task),
	}
	outbound := chathistory.ChatHistoryItem{
		Channel:          "console",
		Kind:             chathistory.KindOutboundAgent,
		ChatID:           subjectID,
		ChatType:         "private",
		ReplyToMessageID: job.TaskID,
		SentAt:           now,
		Sender: chathistory.ChatHistorySender{
			UserID:     "0",
			Username:   "agent",
			Nickname:   "MisterMorph",
			IsBot:      true,
			DisplayRef: "agent",
		},
		Text: strings.TrimSpace(output),
	}
	return memoryruntime.RecordRequest{
		TaskRunID:   strings.TrimSpace(job.TaskID),
		SessionID:   subjectID,
		SubjectID:   subjectID,
		Channel:     "console",
		TaskText:    strings.TrimSpace(job.Task),
		FinalOutput: strings.TrimSpace(output),
		Participants: []memory.MemoryParticipant{
			{ID: "console:user", Nickname: "Console User", Protocol: "console"},
			{ID: 0, Nickname: "MisterMorph", Protocol: ""},
		},
		SourceHistory: []chathistory.ChatHistoryItem{
			inbound,
			outbound,
		},
		SessionContext: memory.SessionContext{
			ConversationID:     subjectID,
			ConversationType:   "private",
			CounterpartyID:     "console:user",
			CounterpartyName:   "Console User",
			CounterpartyHandle: "console",
			CounterpartyLabel:  "[Console User](console:user)",
		},
	}
}

func pendingApprovalID(final *agent.Final) (string, bool) {
	if final == nil || final.Output == nil {
		return "", false
	}
	switch v := final.Output.(type) {
	case agent.PendingOutput:
		if strings.EqualFold(strings.TrimSpace(v.Status), "pending") && strings.TrimSpace(v.ApprovalRequestID) != "" {
			return strings.TrimSpace(v.ApprovalRequestID), true
		}
	case *agent.PendingOutput:
		if v != nil && strings.EqualFold(strings.TrimSpace(v.Status), "pending") && strings.TrimSpace(v.ApprovalRequestID) != "" {
			return strings.TrimSpace(v.ApprovalRequestID), true
		}
	case map[string]any:
		st, _ := v["status"].(string)
		id, _ := v["approval_request_id"].(string)
		if strings.EqualFold(strings.TrimSpace(st), "pending") && strings.TrimSpace(id) != "" {
			return strings.TrimSpace(id), true
		}
	}
	return "", false
}
