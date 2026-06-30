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
		// slog.Debug("idle before select")
		select {
		case <-psControl.restartChannel:
			// slog.Debug("idle restart received")
			// Do not upload settings now
			// next prepare will do
			// err := psControl.setEverything()
			// if err != nil && psControl.DisplayStatus != nil {
			// 	psControl.DisplayStatus(err.Error())
			// }
		case nextState := <-psControl.stateChannel:
			slog.Debug("nextState", "new state", functionName(nextState))
			return nextState
		case <-psControl.stopChannel:
			slog.Debug("idle stop received")
		}
	}
	// return nil
}

func (psControl *PscDesc) stateMachine() {
	f := idle
	for f != nil {
		f = f(psControl)
	}
}

// func (psControl *PscDesc) stopIdle() {
// }
// func (psControl *PscDesc) runIdle() {
// }

func (psControl *PscDesc) quit() (err error) {
	// Run asynchronously to prevent state machine deadlocks if the driver hangs
	go func() {
		stopErr := psControl.Con.Stop()
		slog.Debug("Stop called from quit", "error", stopErr)
	}()
	// err = psControl.Con.Send(&psi.StopMsg{})
	// if err != nil {
	// 	slog.Error("quit", "error", err)
	// 	return
	// }
	// resp, err := psControl.Con.Receive()
	// if err != nil {
	// 	slog.Error("quit response:", "error", err)
	// 	return
	// }
	// if resp.Status() != nil {
	// 	err = fmt.Errorf("%v", resp.Status())
	// 	slog.Error("quit response:", "status", resp.Status())
	// }
	return
}

func (psControl *PscDesc) requestRestart() {
	// slog.Debug("Restart")
	// slog.Debug("Restart", "goid", goid())
	select {
	case psControl.restartChannel <- struct{}{}:
		// slog.Debug("Restart sent")
	default: // Do not block. There is an earlier restart message in the channel.
		// slog.Debug("Restart default")
	}
}

func (psControl *PscDesc) RequestRestart() {
	psControl.requestRestart()
}
