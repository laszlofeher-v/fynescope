package control

import (
	"log/slog"
	"fynescope/genericps"
	// "fynescope/psi"
)

func (psControl *PscDesc) setGenerator() (err error) {
	slog.Debug("control setGenerator")
	psControl.getGeneratorCh <- &psControl.getGenerator
	if <-psControl.getGenerator.newSetting {
		psControl.Con.SetSigGenBuiltInV2(psControl.getGenerator.generatorSettings.OffsetVoltage,
			psControl.getGenerator.generatorSettings.PkToPK,
			psControl.getGenerator.generatorSettings.WaveType,
			psControl.getGenerator.generatorSettings.StartFrequency,
			psControl.getGenerator.generatorSettings.StopFrequency,
			psControl.getGenerator.generatorSettings.Increment,
			psControl.getGenerator.generatorSettings.DwellTime,
			psControl.getGenerator.generatorSettings.SweepType,
			psControl.getGenerator.generatorSettings.Operation,
			psControl.getGenerator.generatorSettings.Shots,
			psControl.getGenerator.generatorSettings.Sweeps,
			psControl.getGenerator.generatorSettings.TriggerType,
			psControl.getGenerator.generatorSettings.TriggerSource,
			psControl.getGenerator.generatorSettings.ExtInThreshold)
	}
	return
}

func (psControl *PscDesc) generatorMonitor() {
	type (
		eventHandlerFunc func() (nextFunc eventHandlerFunc)
	)
	var (
		unchanged, changed  eventHandlerFunc
		storedSetting GeneratorDesc
	)
	storeSettings := func(msg *GeneratorDescMsg) (nextFunc eventHandlerFunc) {
		// slog.Debug("storeSettings", "*msg", *msg)
		if storedSetting != msg.GeneratorDesc {
			storedSetting = msg.GeneratorDesc
			psControl.requestRestart() // restart the running state machine
			return changed
		}
		return unchanged
	}
	unchanged = func() (nextFunc eventHandlerFunc) {
		select {
		case msg := <-psControl.SetGeneratorCh:
			// slog.Debug("generatorMonitor unchanged set received", "*msg", *msg)
			return storeSettings(msg)
		case getMsg := <-psControl.getGeneratorCh:
			getMsg.newSetting <- false
			return unchanged
		}
	}
	changed = func() (nextFunc eventHandlerFunc) {
		select {
		case msg := <-psControl.SetGeneratorCh:
			_ = storeSettings(msg)
			return changed
		case getMsg := <-psControl.getGeneratorCh:
			getMsg.generatorSettings = &storedSetting
			getMsg.newSetting <- true
			return unchanged
		}
	}
	eventHandler := unchanged
	for {
		eventHandler = eventHandler()
	}
}

func (psControl *PscDesc) simGeneratorMonitor() {
	var storedSettings [4]GeneratorDesc
	for {
		msg := <-psControl.SetSimGenCh
		ch := int(msg.Channel)
		if ch >= 0 && ch < 4 {
			// If the simulator connection isn't set up yet or is not the simulator, we probably shouldn't panic, but let's check it.
			if psControl.Con != nil && psControl.Con.ID == genericps.SimId {
				if storedSettings[ch] != msg.GeneratorDesc {
					storedSettings[ch] = msg.GeneratorDesc
					// Send to simulator directly
					psControl.Con.SetSimGen(msg.Channel, msg.On, msg.OffsetVoltage, msg.PkToPK, msg.WaveType,
						msg.StartFrequency, msg.StopFrequency, msg.Increment, msg.DwellTime, msg.SweepType,
						msg.Operation, msg.Shots, msg.Sweeps, msg.TriggerType, msg.TriggerSource,
						msg.ExtInThreshold, msg.Phase)
					psControl.requestRestart()
				}
			}
		}
		if msg.Done != nil {
			msg.Done <- struct{}{}
		}
	}
}
