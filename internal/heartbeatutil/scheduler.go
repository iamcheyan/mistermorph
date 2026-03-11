package heartbeatutil

import "strings"

type TickOutcome int

const (
	TickEnqueued TickOutcome = iota
	TickSkipped
	TickBuildError
)

const (
	SkipReasonInvalidConfig  = "invalid_config"
	SkipReasonAlreadyRunning = "already_running"
	SkipReasonEmptyTask      = "empty_task"
)

type TaskBuilder func() (task string, checklistEmpty bool, err error)

type TaskEnqueuer func(task string, checklistEmpty bool) (skipReason string)

type TickResult struct {
	Outcome      TickOutcome
	SkipReason   string
	BuildError   error
	AlertMessage string
}

func Tick(state *State, buildTask TaskBuilder, enqueueTask TaskEnqueuer) TickResult {
	if state == nil || buildTask == nil || enqueueTask == nil {
		return TickResult{
			Outcome:    TickSkipped,
			SkipReason: SkipReasonInvalidConfig,
		}
	}
	if !state.Start() {
		return TickResult{
			Outcome:    TickSkipped,
			SkipReason: SkipReasonAlreadyRunning,
		}
	}

	task, checklistEmpty, err := buildTask()
	if err != nil {
		alert, msg := state.EndFailure(err)
		result := TickResult{
			Outcome:    TickBuildError,
			BuildError: err,
		}
		if alert {
			result.AlertMessage = strings.TrimSpace(msg)
		}
		return result
	}
	if strings.TrimSpace(task) == "" {
		state.EndSkipped()
		return TickResult{
			Outcome:    TickSkipped,
			SkipReason: SkipReasonEmptyTask,
		}
	}

	reason := strings.TrimSpace(enqueueTask(task, checklistEmpty))
	if reason != "" {
		state.EndSkipped()
		return TickResult{
			Outcome:    TickSkipped,
			SkipReason: reason,
		}
	}

	return TickResult{Outcome: TickEnqueued}
}
