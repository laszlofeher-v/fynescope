package control

import (
	"fynescope/genericps"
	"fynescope/settings"
	"log/slog"
	"sync/atomic"
)

// channelStateMachine coordinates runtime configuration across all oscilloscope analog channels.
// It runs as a state machine actor alternating between unchanged and changed states:
// - Translates GUI channel settings into hardware driver commands (SetChannelMsg).
// - Automatically adjusts voltage ranges and analog offsets when 10x probe attenuation is enabled.
// - Atomically tracks which channels are currently enabled.
// - Triggers scope acquisition restarts whenever channel configuration changes.
func (psControl *PscDesc) channelStateMachine(numberOfChannels int) {
	type (
		eventHandlerFunc func() (nextFunc eventHandlerFunc)
	)
	var (
		unchanged, changed eventHandlerFunc
		changedSet         []bool
		oldChDesc          []genericps.SetChannelMsg
	)

	// storeSettings applies incoming channel parameters, adjusts for 10x probe attenuation,
	// and sets dirty flags if any parameter changed.
	storeSettings := func(setMsg *settings.ChSettings) (nextFunc eventHandlerFunc) {
		targetRange := setMsg.VRange
		targetOffset := setMsg.Offset
		// For 10x probe attenuation, shift voltage range index by 3 and divide offset by 10
		if setMsg.X10 {
			targetRange = setMsg.VRange - 3
			if targetRange < 0 {
				slog.Debug("X10 targetRange clamped", "original", setMsg.VRange-3)
				targetRange = 0
			}
			targetOffset = setMsg.Offset / 10
		}

		newData := false
		if oldChDesc[setMsg.ID].Channel != setMsg.ID {
			oldChDesc[setMsg.ID].Channel = setMsg.ID
			newData = true
		}
		if oldChDesc[setMsg.ID].CouplingType != setMsg.CoupleType {
			oldChDesc[setMsg.ID].CouplingType = setMsg.CoupleType
			newData = true
		}
		if oldChDesc[setMsg.ID].Enabled != setMsg.Enabled {
			oldChDesc[setMsg.ID].Enabled = setMsg.Enabled
			psControl.chEnabled[setMsg.ID].Store(setMsg.Enabled)
			newData = true
		}
		if oldChDesc[setMsg.ID].X10 != setMsg.X10 {
			oldChDesc[setMsg.ID].X10 = setMsg.X10
			newData = true
		}
		if oldChDesc[setMsg.ID].VoltageRange != targetRange {
			oldChDesc[setMsg.ID].VoltageRange = targetRange
			newData = true
		}
		if oldChDesc[setMsg.ID].AnalogOffset != targetOffset {
			oldChDesc[setMsg.ID].AnalogOffset = targetOffset
			newData = true
		}
		if newData {
			changedSet[setMsg.ID] = true
			psControl.requestRestart()
			return changed
		}
		return unchanged
	}

	// countEnabled calculates the count of currently enabled channels.
	countEnabled := func() (n int) {
		for i := range oldChDesc {
			if oldChDesc[i].Enabled {
				n++
			}
		}
		return
	}

	// unchanged state processes channel updates and serves channel queries when no pending changes exist.
	unchanged = func() (nextFunc eventHandlerFunc) {
		slog.Debug("channel unchanged started")
		select {
		case <-psControl.shutdownCh:
			return nil
		case setMsg := <-psControl.SetChannelCh:
			return storeSettings(setMsg)
		case getMsg := <-psControl.getChannelCh:
			getMsg.newSettings <- false
			return unchanged
		case getNumOfEnabledMsg := <-psControl.getNumOfEnabledCh:
			getNumOfEnabledMsg.n <- countEnabled()
			return unchanged
		}
	}

	// changed state delivers modified channel configurations one-by-one to getChannelCh.
	changed = func() (nextFunc eventHandlerFunc) {
		slog.Debug("channel changed started")
		select {
		case <-psControl.shutdownCh:
			return nil
		case setMsg := <-psControl.SetChannelCh:
			_ = storeSettings(setMsg)
			return changed
		case getMsg := <-psControl.getChannelCh:
			for i := range changedSet {
				if changedSet[i] {
					channelSettings := oldChDesc[i]
					getMsg.channelSettings = &channelSettings
					getMsg.newSettings <- true
					changedSet[i] = false
					return changed
				}
			}
			getMsg.newSettings <- false
			return unchanged
		case getNumOfEnabledMsg := <-psControl.getNumOfEnabledCh:
			getNumOfEnabledMsg.n <- countEnabled()
			return changed
		}
	}

	eventHandler := unchanged
	oldChDesc = make([]genericps.SetChannelMsg, numberOfChannels)
	changedSet = make([]bool, numberOfChannels)
	psControl.chEnabled = make([]atomic.Bool, numberOfChannels)
	for eventHandler != nil {
		eventHandler = eventHandler()
	}
}

