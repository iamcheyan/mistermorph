package heartbeat

import (
	"context"
	"fmt"

	"github.com/quailyquaily/mistermorph/internal/daemonruntime"
	"github.com/quailyquaily/mistermorph/internal/heartbeatutil"
)

type PokeRequest struct {
	Input  daemonruntime.PokeInput
	Result chan error
}

func Trigger(ctx context.Context, requests chan<- PokeRequest, input daemonruntime.PokeInput) error {
	if requests == nil {
		return fmt.Errorf("heartbeat poke is unavailable")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	req := PokeRequest{
		Input:  input.Normalize(),
		Result: make(chan error, 1),
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case requests <- req:
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-req.Result:
		return err
	}
}

func ErrorFromTickResult(result heartbeatutil.TickResult) error {
	switch result.Outcome {
	case heartbeatutil.TickEnqueued:
		return nil
	case heartbeatutil.TickBuildError:
		if result.BuildError != nil {
			return result.BuildError
		}
		return fmt.Errorf("heartbeat poke failed")
	case heartbeatutil.TickSkipped:
		switch result.SkipReason {
		case "", heartbeatutil.SkipReasonEmptyTask:
			return nil
		case heartbeatutil.SkipReasonAlreadyRunning:
			return daemonruntime.ErrPokeBusy
		default:
			return fmt.Errorf("heartbeat poke skipped: %s", result.SkipReason)
		}
	default:
		return fmt.Errorf("heartbeat poke failed")
	}
}
