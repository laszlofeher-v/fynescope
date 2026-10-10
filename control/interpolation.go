package control

import (
	"fynescope/settings"
	"log/slog"
)

// interpolationMonitor runs as a concurrent actor monitoring changes to the active waveform
// interpolation mode (Raw, Linear, or Sinc).
// Transitions into or out of Sinc (Whittaker–Shannon) interpolation require the acquisition engine
// to restart to adjust sample count allocations and calculation buffers.
func (psControl *PscDesc) interpolationMonitor() {
	type (
		eventHandlerFunc func() (nextFunc eventHandlerFunc)
	)
	var (
		unchanged, changed eventHandlerFunc
		oldIpMode          settings.InterpolationType
	)

	// store evaluates whether the interpolation mode switch requires an acquisition restart.
	store := func(ipMode settings.InterpolationType) eventHandlerFunc {
		newData := false
		// Sinc mode changes affect buffer sizing, necessitating a restart
		if oldIpMode != ipMode &&
			(ipMode == settings.Sinc || oldIpMode == settings.Sinc) {
			newData = true
		}
		oldIpMode = ipMode // ipMode can be raw, linear, or sinc
		if newData {
			psControl.requestRestart()
			return changed
		}
		return unchanged
	}

	// unchanged state waits for new interpolation setting messages or query requests.
	unchanged = func() (nextFunc eventHandlerFunc) {
		select {
		case <-psControl.shutdownCh:
			return nil
		case ipMode := <-psControl.SetInterpolationModeCh:
			slog.Debug("unchanged received", "ipMode", ipMode)
			return store(ipMode)
		case getMsg := <-psControl.getInterpolationModeCh:
			slog.Debug("unchanged getMsg received", "getMsg", getMsg)
			getMsg.newSetting <- false
			return unchanged
		}
	}

	// changed state reports the updated interpolation mode back to the acquisition scheduler.
	changed = func() (nextFunc eventHandlerFunc) {
		select {
		case <-psControl.shutdownCh:
			return nil
		case ipMode := <-psControl.SetInterpolationModeCh:
			_ = store(ipMode)
			slog.Debug("changed set")
			return changed
		case getMsg := <-psControl.getInterpolationModeCh:
			slog.Debug("changed getMsg received", "getMsg", getMsg)
			getMsg.ipMode = oldIpMode
			getMsg.newSetting <- true
			return unchanged
		}
	}

	eventHandler := unchanged
	for eventHandler != nil {
		eventHandler = eventHandler()
	}
}
