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
		if psControl.getGenerator.generatorSettings.WaveType == genericps.Arbitrary {
			waveform := psControl.getGenerator.generatorSettings.ArbitraryWaveform
			if len(waveform) == 0 {
				waveform = make([]int16, 8192)
			}
			startPhase, _ := psControl.Con.SigGenFrequencyToPhase(psControl.getGenerator.generatorSettings.StartFrequency, psControl.getGenerator.generatorSettings.IndexMode, uint32(len(waveform)))
			stopPhase, _ := psControl.Con.SigGenFrequencyToPhase(psControl.getGenerator.generatorSettings.StopFrequency, psControl.getGenerator.generatorSettings.IndexMode, uint32(len(waveform)))
			incPhase, _ := psControl.Con.SigGenFrequencyToPhase(psControl.getGenerator.generatorSettings.Increment, psControl.getGenerator.generatorSettings.IndexMode, uint32(len(waveform)))
			
			psControl.Con.SetSigGenArbitrary(psControl.getGenerator.generatorSettings.OffsetVoltage,
				psControl.getGenerator.generatorSettings.PkToPK,
				startPhase, stopPhase, incPhase,
				uint32(psControl.getGenerator.generatorSettings.DwellTime),
				waveform,
				psControl.getGenerator.generatorSettings.SweepType,
				psControl.getGenerator.generatorSettings.Operation,
				psControl.getGenerator.generatorSettings.IndexMode,
				psControl.getGenerator.generatorSettings.Shots,
				psControl.getGenerator.generatorSettings.Sweeps,
				psControl.getGenerator.generatorSettings.TriggerType,
				psControl.getGenerator.generatorSettings.TriggerSource,
				psControl.getGenerator.generatorSettings.ExtInThreshold)
		} else {
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
		if !storedSetting.Equals(&msg.GeneratorDesc) {
			storedSetting = msg.GeneratorDesc
			psControl.requestRestart() // restart the running state machine
			return changed
		}
		return unchanged
	}
	unchanged = func() (nextFunc eventHandlerFunc) {
		select {
		case <-psControl.shutdownCh:
			return nil
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
		case <-psControl.shutdownCh:
			return nil
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
	for eventHandler != nil {
		eventHandler = eventHandler()
	}
}

func (psControl *PscDesc) demoGeneratorMonitor() {
	var storedSettings [4]GeneratorDesc
	for {
		var msg *GeneratorDescMsg
		select {
		case <-psControl.shutdownCh:
			return
		case msg = <-psControl.SetDemoGenCh:
		}
		ch := int(msg.Channel)
		if ch >= 0 && ch < 4 {
			// If the demo connection isn't set up yet or is not the demo, we probably shouldn't panic, but let's check it.
			if psControl.Con != nil && psControl.Con.ID == genericps.DemoId {
				if !storedSettings[ch].Equals(&msg.GeneratorDesc) {
					storedSettings[ch] = msg.GeneratorDesc
					// Send to demo directly
					psControl.Con.SetDemoGen(msg.Channel, msg.On, msg.OffsetVoltage, msg.PkToPK, msg.WaveType,
						msg.StartFrequency, msg.StopFrequency, msg.Increment, msg.DwellTime, msg.SweepType,
						msg.Operation, msg.Shots, msg.Sweeps, msg.TriggerType, msg.TriggerSource,
						msg.ExtInThreshold, msg.Phase, msg.ArbitraryWaveform)
					psControl.requestRestart()
				}
			}
		}
		if msg.Done != nil {
			msg.Done <- struct{}{}
		}
	}
}
