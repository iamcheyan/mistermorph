package telegram

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/quailyquaily/mistermorph/agent"
	"github.com/quailyquaily/mistermorph/guard"
	busruntime "github.com/quailyquaily/mistermorph/internal/bus"
	"github.com/quailyquaily/mistermorph/internal/channelruntime/depsutil"
	"github.com/quailyquaily/mistermorph/internal/chathistory"
	"github.com/quailyquaily/mistermorph/internal/memoryruntime"
	"github.com/quailyquaily/mistermorph/internal/promptprofile"
	"github.com/quailyquaily/mistermorph/internal/retryutil"
	"github.com/quailyquaily/mistermorph/internal/todo"
	"github.com/quailyquaily/mistermorph/internal/toolsutil"
	"github.com/quailyquaily/mistermorph/llm"
	"github.com/quailyquaily/mistermorph/memory"
	"github.com/quailyquaily/mistermorph/tools"
	telegramtools "github.com/quailyquaily/mistermorph/tools/telegram"
)

type runtimeTaskOptions struct {
	MemoryInjectionEnabled      bool
	MemoryInjectionMaxItems     int
	SecretsRequireSkillProfiles bool
	MemoryManager               *memory.Manager
	MemoryOrchestrator          *memoryruntime.Orchestrator
	MemoryProjectionWorker      *memoryruntime.ProjectionWorker
}

func runTelegramTask(ctx context.Context, d Dependencies, logger *slog.Logger, logOpts agent.LogOptions, client llm.Client, baseReg *tools.Registry, api *telegramAPI, filesEnabled bool, fileCacheDir string, filesMaxBytes int64, sharedGuard *guard.Guard, cfg agent.Config, allowedIDs map[int64]bool, job telegramJob, botUsername string, model string, history []chathistory.ChatHistoryItem, historyCap int, stickySkills []string, requestTimeout time.Duration, runtimeOpts runtimeTaskOptions, sendTelegramText func(context.Context, int64, string, string) error) (*agent.Final, *agent.Context, []string, *telegramtools.Reaction, error) {
	if sendTelegramText == nil {
		return nil, nil, nil, nil, fmt.Errorf("send telegram text callback is required")
	}
	task := job.Text
	historyWithCurrent := append([]chathistory.ChatHistoryItem(nil), history...)
	historyWithCurrent = append(historyWithCurrent, newTelegramInboundHistoryItem(job))
	historyRaw, err := json.MarshalIndent(map[string]any{
		"chat_history_messages": chathistory.BuildMessages(chathistory.ChannelTelegram, historyWithCurrent),
	}, "", "  ")
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("render telegram history context: %w", err)
	}
	llmHistory := []llm.Message{{Role: "user", Content: string(historyRaw)}}
	if baseReg == nil {
		return nil, nil, nil, nil, fmt.Errorf("base registry is nil")
	}

	// Per-run registry.
	reg := buildTelegramRegistry(baseReg, job.ChatType)
	toolsutil.RegisterRuntimeTools(reg, d.RuntimeToolsConfig, client, model)
	toolsutil.SetTodoUpdateToolAddContext(reg, todo.AddResolveContext{
		Channel:          "telegram",
		ChatType:         job.ChatType,
		ChatID:           job.ChatID,
		SpeakerUserID:    job.FromUserID,
		SpeakerUsername:  job.FromUsername,
		MentionUsernames: append([]string(nil), job.MentionUsers...),
		UserInputRaw:     job.Text,
	})
	toolAPI := newTelegramToolAPI(api)
	if api != nil {
		reg.Register(telegramtools.NewSendVoiceTool(toolAPI, job.ChatID, fileCacheDir, filesMaxBytes, nil))
		if filesEnabled {
			reg.Register(telegramtools.NewSendFileTool(toolAPI, job.ChatID, fileCacheDir, filesMaxBytes))
		}
	}
	var reactTool *telegramtools.ReactTool
	if api != nil && job.MessageID != 0 {
		reactTool = telegramtools.NewReactTool(toolAPI, job.ChatID, job.MessageID, allowedIDs)
		reg.Register(reactTool)
	}

	promptSpec, loadedSkills, skillAuthProfiles, err := depsutil.PromptSpecFromCommon(d, ctx, logger, logOpts, task, client, model, stickySkills)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	promptprofile.ApplyPersonaIdentity(&promptSpec, logger)
	promptprofile.AppendLocalToolNotesBlock(&promptSpec, logger)
	promptprofile.AppendPlanCreateGuidanceBlock(&promptSpec, reg)
	promptprofile.AppendTelegramRuntimeBlocks(&promptSpec, isGroupChat(job.ChatType), job.MentionUsers, strings.Join(telegramtools.StandardReactionEmojis(), ","))

	memSubjectID := telegramMemorySubjectID(job)
	if runtimeOpts.MemoryOrchestrator != nil && memSubjectID != "" && runtimeOpts.MemoryInjectionEnabled {
		snap, memErr := runtimeOpts.MemoryOrchestrator.PrepareInjectionWithAdapter(telegramMemoryInjectionAdapter{job: job}, runtimeOpts.MemoryInjectionMaxItems)
		if memErr != nil {
			if logger != nil {
				logger.Warn("memory_injection_error", "source", "telegram", "subject_id", memSubjectID, "chat_id", job.ChatID, "error", memErr.Error())
			}
		} else if strings.TrimSpace(snap) != "" {
			promptprofile.AppendMemorySummariesBlock(&promptSpec, snap)
			if logger != nil {
				logger.Info("memory_injection_applied", "source", "telegram", "subject_id", memSubjectID, "chat_id", job.ChatID, "snapshot_len", len(snap))
			}
		} else if logger != nil {
			logger.Debug("memory_injection_skipped", "source", "telegram", "reason", "empty_snapshot", "subject_id", memSubjectID, "chat_id", job.ChatID)
		}
	}

	planUpdateHook := func(runCtx *agent.Context, update agent.PlanStepUpdate) {
		if runCtx == nil || runCtx.Plan == nil {
			return
		}
		msg, err := generateTelegramPlanProgressMessage(ctx, client, model, task, runCtx.Plan, update, requestTimeout)
		if err != nil {
			logger.Warn("telegram_plan_progress_error", "error", err.Error())
			return
		}
		if strings.TrimSpace(msg) == "" {
			return
		}
		correlationID := fmt.Sprintf("telegram:plan:%d:%d", job.ChatID, job.MessageID)
		if err := sendTelegramText(context.Background(), job.ChatID, msg, correlationID); err != nil {
			logger.Warn("telegram_bus_publish_error", "channel", busruntime.ChannelTelegram, "chat_id", job.ChatID, "message_id", job.MessageID, "bus_error_code", busErrorCodeString(err), "error", err.Error())
		}
	}

	engineOpts := []agent.Option{
		agent.WithLogger(logger),
		agent.WithLogOptions(logOpts),
		agent.WithSkillAuthProfiles(skillAuthProfiles, runtimeOpts.SecretsRequireSkillProfiles),
		agent.WithGuard(sharedGuard),
		agent.WithPlanStepUpdate(planUpdateHook),
	}
	engine := agent.New(
		client,
		reg,
		cfg,
		promptSpec,
		engineOpts...,
	)
	meta := job.Meta
	if meta == nil {
		meta = map[string]any{
			"trigger":               "telegram",
			"telegram_chat_id":      job.ChatID,
			"telegram_message_id":   job.MessageID,
			"telegram_chat_type":    job.ChatType,
			"telegram_from_user_id": job.FromUserID,
		}
	}
	botUsername = strings.TrimPrefix(strings.TrimSpace(botUsername), "@")
	if botUsername != "" {
		meta["telegram_bot_username"] = botUsername
	}
	final, agentCtx, err := engine.Run(ctx, task, agent.RunOptions{
		Model:           model,
		History:         llmHistory,
		Meta:            meta,
		SkipTaskMessage: true,
	})
	if err != nil {
		return final, agentCtx, loadedSkills, nil, err
	}

	var reaction *telegramtools.Reaction
	if reactTool != nil {
		reaction = reactTool.LastReaction()
		if reaction != nil && logger != nil {
			logger.Info("message_reaction_applied",
				"chat_id", reaction.ChatID,
				"message_id", reaction.MessageID,
				"emoji", reaction.Emoji,
				"source", reaction.Source,
			)
		}
	}

	publishText := shouldPublishTelegramText(final)
	if shouldWriteMemory(publishText, runtimeOpts.MemoryOrchestrator, memSubjectID) {
		if err := recordMemoryFromJob(ctx, logger, client, model, runtimeOpts.MemoryOrchestrator, runtimeOpts.MemoryManager, job, history, historyCap, final, requestTimeout); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				retryutil.AsyncRetry(logger, "memory_update", 2*time.Second, requestTimeout, func(retryCtx context.Context) error {
					return recordMemoryFromJob(retryCtx, logger, client, model, runtimeOpts.MemoryOrchestrator, runtimeOpts.MemoryManager, job, history, historyCap, final, requestTimeout)
				})
			}
			logger.Warn("memory_update_error", "error", err.Error())
		} else if runtimeOpts.MemoryProjectionWorker != nil {
			runtimeOpts.MemoryProjectionWorker.NotifyRecordAppended()
		}
	}
	return final, agentCtx, loadedSkills, reaction, nil
}

func shouldWriteMemory(publishText bool, orchestrator *memoryruntime.Orchestrator, subjectID string) bool {
	if !publishText || orchestrator == nil {
		return false
	}
	return strings.TrimSpace(subjectID) != ""
}

func buildTelegramRegistry(baseReg *tools.Registry, chatType string) *tools.Registry {
	reg := tools.NewRegistry()
	if baseReg == nil {
		return reg
	}
	groupChat := isGroupChat(chatType)
	for _, t := range baseReg.All() {
		name := strings.TrimSpace(t.Name())
		if groupChat && strings.EqualFold(name, "contacts_send") {
			continue
		}
		reg.Register(t)
	}
	return reg
}

func generateTelegramPlanProgressMessage(ctx context.Context, client llm.Client, model string, task string, plan *agent.Plan, update agent.PlanStepUpdate, requestTimeout time.Duration) (string, error) {
	_ = ctx
	_ = client
	_ = model
	_ = task
	_ = requestTimeout

	if plan == nil || update.CompletedIndex < 0 {
		return "", nil
	}
	stepText := firstNonEmpty(
		strings.TrimSpace(update.CompletedStep),
		stepByIndex(plan, update.CompletedIndex),
		strings.TrimSpace(update.StartedStep),
		stepByIndex(plan, update.StartedIndex),
	)
	if stepText == "" {
		return "", nil
	}
	return stepText, nil
}

func stepByIndex(plan *agent.Plan, index int) string {
	if plan == nil || index < 0 || index >= len(plan.Steps) {
		return ""
	}
	return strings.TrimSpace(plan.Steps[index].Step)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v != "" {
			return v
		}
	}
	return ""
}
