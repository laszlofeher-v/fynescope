package control

import (
	"fynescope/genericps"
	"fynescope/settings"
	"log/slog"
)

func (psControl *PscDesc) setDigitalPort() (err error) {
	if psControl.getDigitalPortCh == nil {
		return nil
	}
	psControl.getDigitalPortCh <- &psControl.getDigitalPort
	for <-psControl.getDigitalPort.newSettings {
		chset := psControl.getDigitalPort.portSettings
		err = psControl.Con.SetDigitalPort(chset.Port, chset.Settings.Enabled, chset.Settings.Threshold)
		if err != nil {
			slog.Error("SetDigitalPort", "error:", err)
			err = nil // don't fail setEverything if digital ports are unsupported on this model
		}
		psControl.getDigitalPortCh <- &psControl.getDigitalPort
	}
	return
}

func (psControl *PscDesc) digitalPortMonitor() {
	type (
		eventHandlerFunc func() (nextFunc eventHandlerFunc)
	)
	var unchanged, changed eventHandlerFunc
	var oldChDesc [2]settings.DigitalPortSettings
	var changedSet [2]bool

	storeSettings := func(setMsg *DigitalPortMsg) (nextFunc eventHandlerFunc) {
		portIdx := int(setMsg.Port - genericps.Port0)
		if portIdx < 0 || portIdx > 1 {
			return unchanged
		}
		psControl.digitalPortsEnabled[portIdx].Store(setMsg.Settings.Enabled)
		if oldChDesc[portIdx] != setMsg.Settings {
			oldChDesc[portIdx] = setMsg.Settings
			changedSet[portIdx] = true
			psControl.requestRestart()
			return changed
		}
		return unchanged
	}

	unchanged = func() (nextFunc eventHandlerFunc) {
		select {
		case <-psControl.shutdownCh:
			return nil
		case setMsg := <-psControl.SetDigitalPortCh:
			return storeSettings(setMsg)
		case getMsg := <-psControl.getDigitalPortCh:
			getMsg.newSettings <- false
			return unchanged
		}
	}

	changed = func() (nextFunc eventHandlerFunc) {
		select {
		case <-psControl.shutdownCh:
			return nil
		case setMsg := <-psControl.SetDigitalPortCh:
			_ = storeSettings(setMsg)
			return changed
		case getMsg := <-psControl.getDigitalPortCh:
			for i := range changedSet {
				if changedSet[i] {
					portSettings := DigitalPortMsg{
						Port:     genericps.Port0 + genericps.DigitalPort(i),
						Settings: oldChDesc[i],
					}
					getMsg.portSettings = &portSettings
					changedSet[i] = false
					getMsg.newSettings <- true
					return changed
				}
			}
			getMsg.newSettings <- false
			return unchanged
		}
	}

	var nextFunc eventHandlerFunc = unchanged
	for nextFunc != nil {
		nextFunc = nextFunc()
	}
}
