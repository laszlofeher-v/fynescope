package control

import (
	"log/slog"
)

func (psControl *PscDesc) screenTimeMonitor() {
	type (
		eventHandlerFunc func() (nextFunc eventHandlerFunc)
	)
	var (
		unchanged, changed        eventHandlerFunc
		storedMaxScreenTime float64
	)
	store := func(screenTime float64) eventHandlerFunc {
		if storedMaxScreenTime != screenTime {
			storedMaxScreenTime = screenTime
			psControl.requestRestart()
			return changed
		}
		return unchanged
	}

	unchanged = func() (nextFunc eventHandlerFunc) {
		slog.Debug("unchanged before select")
		select {
		case <-psControl.shutdownCh:
			return nil
		case screenTime := <-psControl.SetMaxScreenTimeCh:
			slog.Debug("unchanged received", "screenTime", screenTime)
			return store(screenTime)
		case getMsg := <-psControl.getMaxScreenTimeCh:
			slog.Debug("unchanged getMsg received", "getMsg", getMsg)
			getMsg.newSetting <- false
			return unchanged
		}
	}
	changed = func() (nextFunc eventHandlerFunc) {
		slog.Debug("changed before select")
		select {
		case <-psControl.shutdownCh:
			return nil
		case screenTime := <-psControl.SetMaxScreenTimeCh:
			_ = store(screenTime)
			slog.Debug("changed set")
			return changed
		case getMsg := <-psControl.getMaxScreenTimeCh:
			slog.Debug("changed getMsg received", "getMsg", getMsg)
			getMsg.maxScreenTime = storedMaxScreenTime
			getMsg.newSetting <- true
			return unchanged
		}
	}
	eventHandler := unchanged
	for eventHandler != nil {
		eventHandler = eventHandler()
	}
}
