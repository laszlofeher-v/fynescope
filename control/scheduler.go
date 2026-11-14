package control

import (
	"log/slog"
	"reflect"
	"runtime"
	"time"
)

func functionName(i any) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

type (
	state func(ps *PscDesc) state
)

func drainSignalChannel(ch <-chan struct{}) {
	for {
		select {
		case <-ch:
		default:
			return
		}
	}
}

func idle(psControl *PscDesc) state {
	for {
		select {
		case <-psControl.shutdownCh:
			slog.Debug("idle quit received")
			return nil
		default:
		}
		select {
		case <-psControl.shutdownCh:
			slog.Debug("idle quit received")
			return nil
		case <-psControl.restartChannel:
			// Do not upload settings now
			// next prepare will do
		case nextState := <-psControl.stateChannel:
			// Drain any stop/restart requests queued before or concurrently with
			// entering the next state (such as SetBlockMode/SetETSMode calling
			// Stop() while already idle) so the next state does not immediately abort.
			drainSignalChannel(psControl.stopChannel)
			drainSignalChannel(psControl.restartChannel)
			slog.Debug("nextState", "new state", functionName(nextState))
			return nextState
		case <-psControl.stopChannel:
			slog.Debug("idle stop received")
		}
	}
}

func (psControl *PscDesc) stateMachine() {
	if psControl.stateMachineDone != nil {
		defer close(psControl.stateMachineDone)
	}
	f := idle
	for f != nil {
		f = f(psControl)
	}
}

func (psControl *PscDesc) quit() (err error) {
	// Run asynchronously to prevent deadlocks if the driver hangs, while
	// waiting briefly for normal Stop completion so a late Stop from the
	// previous state cannot cancel a newly started next state (e.g. blockMode
	// handing off to streamMode).
	done := make(chan struct{})
	go func() {
		defer close(done)
		stopErr := psControl.stopHardware()
		slog.Debug("Stop called from quit", "error", stopErr)
	}()
	timer := time.NewTimer(100 * time.Millisecond)
	defer timer.Stop()
	select {
	case <-done:
	case <-timer.C:
		slog.Debug("quit stopHardware timed out waiting for completion")
	}
	return
}

func (psControl *PscDesc) requestRestart() {
	select {
	case psControl.restartChannel <- struct{}{}:
	default: // Do not block. There is an earlier restart message in the channel.
	}
}

func (psControl *PscDesc) RequestRestart() {
	psControl.requestRestart()
}
