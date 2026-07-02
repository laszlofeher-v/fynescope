package control

import (
	"log/slog"
	"fynescope/settings"
)

func (psControl *PscDesc) interpolationMonitor() {
	type (
		eventHandlerFunc func() (nextFunc eventHandlerFunc)
	)
	var (
		unchanged, changed eventHandlerFunc
		oldIpMode    settings.InterpolationType
	)
	store := func(ipMode settings.InterpolationType) eventHandlerFunc {
		newData := false
		if oldIpMode != ipMode &&
			(ipMode == settings.Sinc || oldIpMode == settings.Sinc) {
			newData = true
		}
		oldIpMode = ipMode // ipMode can be raw, linear, sinc
		if newData {
			psControl.requestRestart()
			return changed
		}
		return unchanged
	}

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
