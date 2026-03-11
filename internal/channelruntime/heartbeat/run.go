package heartbeat

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/quailyquaily/mistermorph/agent"
	"github.com/quailyquaily/mistermorph/guard"
	"github.com/quailyquaily/mistermorph/internal/channelruntime/depsutil"
	"github.com/quailyquaily/mistermorph/internal/heartbeatutil"
	"github.com/quailyquaily/mistermorph/internal/llmutil"
	"github.com/quailyquaily/mistermorph/internal/memoryruntime"
	"github.com/quailyquaily/mistermorph/internal/promptprofile"
	"github.com/quailyquaily/mistermorph/internal/statepaths"
	"github.com/quailyquaily/mistermorph/internal/toolsutil"
	"github.com/quailyquaily/mistermorph/llm"
	"github.com/quailyquaily/mistermorph/memory"
	"github.com/quailyquaily/mistermorph/tools"
)

type RunOptions struct {
	Interval                time.Duration
	InitialDelay            time.Duration
	TaskTimeout             time.Duration
	RequestTimeout          time.Duration
	AgentLimits             agent.Limits
	Source                  string
	ChecklistPath           string
	MemoryEnabled           bool
	MemoryShortTermDays     int
	MemoryInjectionEnabled  bool
	MemoryInjectionMaxItems int
	Notifier                Notifier
	PokeRequests            <-chan PokeRequest
}

type Dependencies = depsutil.HeartbeatDependencies

func Run(ctx context.Context, d Dependencies, opts RunOptions) error {
	return runHeartbeatLoop(ctx, d, resolveRuntimeLoopOptionsFromRunOptions(opts))
}

func runHeartbeatLoop(ctx context.Context, d Dependencies, opts runtimeLoopOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}
	common := depsutil.CommonFromHeartbeat(d)

	logger, err := depsutil.LoggerFromCommon(common)
	if err != nil {
		return err
	}
	logOpts := depsutil.LogOptionsFromCommon(common)

	route, err := depsutil.ResolveLLMRouteFromCommon(common, llmutil.RoutePurposeHeartbeat)
	if err != nil {
		return err
	}
	client, err := depsutil.CreateClient(d.CreateLLMClient, route)
	if err != nil {
		return err
	}
	model := strings.TrimSpace(route.ClientConfig.Model)
	if model == "" {
		return fmt.Errorf("missing model")
	}

	baseReg := depsutil.RegistryFromCommon(common)
	if baseReg == nil {
		return fmt.Errorf("base registry is nil")
	}
	sharedGuard := depsutil.GuardFromCommon(common, logger)
	cfg := opts.AgentLimits.ToConfig()

	orchestrator, projectionWorker, cleanup, err := newHeartbeatOrchestrator(common, opts)
	if err != nil {
		return err
	}
	defer cleanup()
	if projectionWorker != nil {
		projectionWorker.Start(ctx)
	}

	state := &heartbeatutil.State{}
	var wg sync.WaitGroup

	runTaskAsync := func(task string, checklistEmpty bool) string {
		if ctx.Err() != nil {
			return "context_canceled"
		}
		runAt := time.Now().UTC()
		taskRunID := heartbeatTaskRunID(runAt)
		meta := depsutil.BuildHeartbeatMetaFromDeps(d, opts.Source, opts.Interval, opts.ChecklistPath, checklistEmpty, map[string]any{
			"task_run_id": taskRunID,
		})
		wg.Add(1)
		go func() {
			defer wg.Done()

			summary, runErr := runHeartbeatTask(ctx, d, heartbeatTaskOptions{
				Logger:                  logger,
				LogOptions:              logOpts,
				Client:                  client,
				Model:                   model,
				Task:                    task,
				Meta:                    meta,
				TaskRunID:               taskRunID,
				BaseRegistry:            baseReg,
				SharedGuard:             sharedGuard,
				Config:                  cfg,
				TaskTimeout:             opts.TaskTimeout,
				MemoryOrchestrator:      orchestrator,
				MemoryProjectionWorker:  projectionWorker,
				MemoryInjectionEnabled:  opts.MemoryInjectionEnabled,
				MemoryInjectionMaxItems: opts.MemoryInjectionMaxItems,
			})
			if runErr != nil {
				displayErr := depsutil.FormatRuntimeError(runErr)
				alert, alertMsg := state.EndFailure(errors.New(displayErr))
				if alert {
					logger.Warn("heartbeat_alert", "source", opts.Source, "message", alertMsg)
					notifyHeartbeat(ctx, opts.Notifier, logger, alertMsg)
					return
				}
				logger.Warn("heartbeat_error", "source", opts.Source, "error", displayErr)
				return
			}
			state.EndSuccess(time.Now())
			if summary == "" {
				summary = "empty"
			}
			logger.Info("heartbeat_summary", "source", opts.Source, "message", summary)
		}()
		return ""
	}

	runTick := func() heartbeatutil.TickResult {
		result := heartbeatutil.Tick(
			state,
			func() (string, bool, error) {
				return depsutil.BuildHeartbeatTaskFromDeps(d, opts.ChecklistPath)
			},
			runTaskAsync,
		)
		switch result.Outcome {
		case heartbeatutil.TickBuildError:
			if strings.TrimSpace(result.AlertMessage) != "" {
				logger.Warn("heartbeat_alert", "source", opts.Source, "message", result.AlertMessage)
				notifyHeartbeat(ctx, opts.Notifier, logger, result.AlertMessage)
			} else if result.BuildError != nil {
				logger.Warn("heartbeat_task_error", "source", opts.Source, "error", result.BuildError.Error())
			}
		case heartbeatutil.TickSkipped:
			logger.Debug("heartbeat_skip", "source", opts.Source, "reason", result.SkipReason)
		}
		return result
	}

	handlePoke := func(req PokeRequest) {
		err := ErrorFromTickResult(runTick())
		if req == nil {
			return
		}
		select {
		case req <- err:
		default:
		}
	}

	pokeRequests := opts.PokeRequests

	if opts.InitialDelay > 0 {
		initialTimer := time.NewTimer(opts.InitialDelay)
		defer initialTimer.Stop()
		initialTriggered := false
		for !initialTriggered {
			select {
			case <-ctx.Done():
				wg.Wait()
				return nil
			case req, ok := <-pokeRequests:
				if !ok {
					pokeRequests = nil
					continue
				}
				handlePoke(req)
				initialTriggered = true
			case <-initialTimer.C:
				runTick()
				initialTriggered = true
			}
		}
	} else {
		runTick()
	}

	ticker := time.NewTicker(opts.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			return nil
		case req, ok := <-pokeRequests:
			if !ok {
				pokeRequests = nil
				continue
			}
			handlePoke(req)
		case <-ticker.C:
			runTick()
		}
	}
}

type heartbeatTaskOptions struct {
	Logger                  *slog.Logger
	LogOptions              agent.LogOptions
	Client                  llm.Client
	Model                   string
	Task                    string
	Meta                    map[string]any
	TaskRunID               string
	BaseRegistry            *tools.Registry
	SharedGuard             *guard.Guard
	Config                  agent.Config
	TaskTimeout             time.Duration
	MemoryOrchestrator      *memoryruntime.Orchestrator
	MemoryProjectionWorker  *memoryruntime.ProjectionWorker
	MemoryInjectionEnabled  bool
	MemoryInjectionMaxItems int
}

func runHeartbeatTask(ctx context.Context, d Dependencies, opts heartbeatTaskOptions) (string, error) {
	task := strings.TrimSpace(opts.Task)
	if task == "" {
		return "", fmt.Errorf("heartbeat task is empty")
	}

	runCtx := ctx
	cancel := func() {}
	if opts.TaskTimeout > 0 {
		runCtx, cancel = context.WithTimeout(ctx, opts.TaskTimeout)
	}
	defer cancel()

	promptSpec, _, err := depsutil.PromptSpecFromCommon(depsutil.CommonFromHeartbeat(d), runCtx, opts.Logger, opts.LogOptions, task, opts.Client, strings.TrimSpace(opts.Model), nil)
	if err != nil {
		return "", err
	}

	reg := cloneRegistry(opts.BaseRegistry)
	toolsutil.RegisterRuntimeTools(reg, d.RuntimeToolsConfig, toolsutil.RuntimeToolLLMOptions{
		DefaultClient: opts.Client,
		DefaultModel:  strings.TrimSpace(opts.Model),
	})
	promptprofile.ApplyPersonaIdentity(&promptSpec, opts.Logger)
	promptprofile.AppendLocalToolNotesBlock(&promptSpec, opts.Logger)
	promptprofile.AppendPlanCreateGuidanceBlock(&promptSpec, reg)
	promptprofile.AppendTodoWorkflowBlock(&promptSpec, reg)
	if opts.MemoryOrchestrator != nil && opts.MemoryInjectionEnabled {
		snap, memErr := opts.MemoryOrchestrator.PrepareInjection(memoryruntime.PrepareInjectionRequest{
			SubjectID:      heartbeatMemorySubjectID,
			RequestContext: memory.ContextPrivate,
			MaxItems:       opts.MemoryInjectionMaxItems,
		})
		if memErr != nil {
			if opts.Logger != nil {
				opts.Logger.Warn("memory_injection_error", "source", "heartbeat", "error", memErr.Error())
			}
		} else if strings.TrimSpace(snap) != "" {
			promptprofile.AppendMemorySummariesBlock(&promptSpec, snap)
		}
	}

	engine := agent.New(
		opts.Client,
		reg,
		opts.Config,
		promptSpec,
		agent.WithLogger(opts.Logger),
		agent.WithLogOptions(opts.LogOptions),
		agent.WithGuard(opts.SharedGuard),
	)
	final, _, err := engine.Run(runCtx, task, agent.RunOptions{
		Model: strings.TrimSpace(opts.Model),
		Scene: "heartbeat.loop",
		Meta:  opts.Meta,
	})
	if err != nil {
		return "", err
	}

	summary := strings.TrimSpace(depsutil.FormatFinalOutput(final))
	if opts.MemoryOrchestrator != nil {
		if _, memErr := opts.MemoryOrchestrator.Record(memoryruntime.RecordRequest{
			TaskRunID: opts.TaskRunID,
			SessionID: heartbeatMemorySessionID,
			SubjectID: heartbeatMemorySubjectID,
			Channel:   "heartbeat",
			Participants: []memory.MemoryParticipant{{
				ID:       "0",
				Nickname: "agent",
				Protocol: "",
			}},
			TaskText:    task,
			FinalOutput: summary,
			SessionContext: memory.SessionContext{
				ConversationID: heartbeatMemorySubjectID,
			},
		}); memErr != nil && opts.Logger != nil {
			opts.Logger.Warn("memory_record_error", "source", "heartbeat", "error", memErr.Error())
		} else if opts.MemoryProjectionWorker != nil {
			opts.MemoryProjectionWorker.NotifyRecordAppended()
		}
	}

	return summary, nil
}

func notifyHeartbeat(ctx context.Context, notifier Notifier, logger *slog.Logger, message string) {
	if notifier == nil {
		return
	}
	if err := notifier.Notify(ctx, strings.TrimSpace(message)); err != nil && logger != nil {
		logger.Warn("heartbeat_notify_error", "error", err.Error())
	}
}

func cloneRegistry(base *tools.Registry) *tools.Registry {
	reg := tools.NewRegistry()
	if base == nil {
		return reg
	}
	for _, t := range base.All() {
		reg.Register(t)
	}
	return reg
}

func newHeartbeatOrchestrator(common depsutil.CommonDependencies, opts runtimeLoopOptions) (*memoryruntime.Orchestrator, *memoryruntime.ProjectionWorker, func(), error) {
	if !opts.MemoryEnabled {
		return nil, nil, func() {}, nil
	}
	mgr := memory.NewManager(statepaths.MemoryDir(), opts.MemoryShortTermDays)
	journal := mgr.NewJournal(memory.JournalOptions{})
	draftResolver, err := memoryruntime.NewConfiguredDraftResolver(memoryruntime.DraftResolverFactoryOptions{
		ResolveLLMRoute: common.ResolveLLMRoute,
		CreateLLMClient: func(route llmutil.ResolvedRoute) (llm.Client, error) {
			return depsutil.CreateClientFromCommon(common, route)
		},
	})
	if err != nil {
		return nil, nil, func() {}, err
	}
	projector := memory.NewProjector(mgr, journal, memory.ProjectorOptions{
		DraftResolver: draftResolver,
	})
	orchestrator, err := memoryruntime.New(mgr, journal, projector, memoryruntime.OrchestratorOptions{})
	if err != nil {
		return nil, nil, func() {}, err
	}
	projectionWorker, err := memoryruntime.NewProjectionWorker(journal, projector, memoryruntime.ProjectionWorkerOptions{})
	if err != nil {
		return nil, nil, func() {}, err
	}
	cleanup := func() {
		_ = journal.Close()
	}
	return orchestrator, projectionWorker, cleanup, nil
}
