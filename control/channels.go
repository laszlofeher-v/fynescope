package control

import (
	"log/slog"
	"fynescope/genericps"
	"fynescope/settings"
	"sync/atomic"
)

func (psControl *PscDesc) channelStateMachine(numberOfChannels int) {
	type (
		eventHandlerFunc func() (nextFunc eventHandlerFunc)
	)
	var (
		unchanged, changed eventHandlerFunc
		changedSet     []bool
		oldChDesc    []genericps.SetChannelMsg
	)
	storeSettings := func(setMsg *settings.ChSettings) (nextFunc eventHandlerFunc) {
		targetRange := setMsg.VRange
		targetOffset := setMsg.Offset
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
			// slog.Debug("unchanged received", "setMsg", setMsg)
			// psControl.channels[setMsg.ID] = oldChDesc[setMsg.ID]
			changedSet[setMsg.ID] = true
			psControl.requestRestart()
			return changed
		}
		return unchanged
	}

	countEnabled := func() (n int) {
		for i := range oldChDesc {
			if oldChDesc[i].Enabled {
				n++
			}
		}
		return
	}
	unchanged = func() (nextFunc eventHandlerFunc) {
		slog.Debug("channel unchanged started")
		select {
		case setMsg := <-psControl.SetChannelCh:
			// slog.Debug("unchanged received", "setMsg", setMsg)
			// defer func() { setMsg.TransferDone <- struct{}{} }()
			// slog.Debug("unchanged done sent")
			return storeSettings(setMsg)
		case getMsg := <-psControl.getChannelCh:
			// slog.Debug("unchanged getMsg received -------------------------------", "getMsg", getMsg)
			getMsg.newSettings <- false
			return unchanged
		case getNumOfEnabledMsg := <-psControl.getNumOfEnabledCh:
			getNumOfEnabledMsg.n <- countEnabled()
			return unchanged
		}
	}
	changed = func() (nextFunc eventHandlerFunc) {
		slog.Debug("channel changed started")
		select {
		case setMsg := <-psControl.SetChannelCh:
			// slog.Debug("changed received", "setMsg", setMsg)
			_ = storeSettings(setMsg)
			// slog.Debug("changed done set")
			return changed
		case getMsg := <-psControl.getChannelCh:
			// slog.Debug("changed getMsg received**********************************", "getMsg", getMsg)
			for i := range changedSet {
				if changedSet[i] {
					// slog.Debug("changed", "i", i)
					channelSettings := oldChDesc[i]
					getMsg.channelSettings = &channelSettings
					getMsg.newSettings <- true
					changedSet[i] = false
					return changed
				}
			}
			// slog.Debug("no more changed")
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
	for {
		eventHandler = eventHandler()
	}
}
