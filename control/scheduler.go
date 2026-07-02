package control

import (
	"log/slog"
	"reflect"
	"runtime"
)

func functionName(i any) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

type (
	state func(ps *PscDesc) state
)

func idle(psControl *PscDesc) state {
	for {
		select {
		case <-psControl.shutdownCh:
			slog.Debug("idle quit received")
			return nil
		case <-psControl.restartChannel:
			// Do not upload settings now
			// next prepare will do
		case nextState := <-psControl.stateChannel:
			slog.Debug("nextState", "new state", functionName(nextState))
			return nextState
		case <-psControl.stopChannel:
			slog.Debug("idle stop received")
		}
	}
}

func (psControl *PscDesc) stateMachine() {
	f := idle
	for f != nil {
		f = f(psControl)
	}
}

func (psControl *PscDesc) quit() (err error) {
	// Run asynchronously to prevent state machine deadlocks if the driver hangs
	go func() {
		stopErr := psControl.Con.Stop()
		slog.Debug("Stop called from quit", "error", stopErr)
	}()
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
