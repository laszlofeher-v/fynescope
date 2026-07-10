package control

import (
	"fynescope/genericps"
	"log/slog"
	// "fynescope/psi"
)

type (
	TriggerModes int
	TriggerTypes int
)

const (
	autoTriggerMs = 1000
)
const (
	Auto TriggerModes = iota
	Repeat
	Single
	ETS
)
const (
	Simple TriggerTypes = iota
	Advanced
	Complex
	Window
	Interval
)

func (psControl *PscDesc) triggerMonitor() {
	type (
		eventHandlerFunc func() (nextFunc eventHandlerFunc)
	)
	var (
		unchanged, changed eventHandlerFunc
		triggerSetting     TriggerDesc
	)
	triggerDescChanged := func(a, b TriggerDesc) bool {
		return a.Enabled != b.Enabled ||
			a.TriggerADC != b.TriggerADC ||
			a.HysteresisADC != b.HysteresisADC ||
			a.UpperHysteresis != b.UpperHysteresis ||
			a.Source != b.Source ||
			a.ThresholdDirection != b.ThresholdDirection ||
			a.Mode != b.Mode ||
			a.Type != b.Type ||
			a.Mv != b.Mv ||
			a.LowerMv != b.LowerMv ||
			a.LowerTriggerADC != b.LowerTriggerADC ||
			a.LowerHysteresisADC != b.LowerHysteresisADC ||
			a.LowerHysteresis != b.LowerHysteresis ||
			a.XOffset != b.XOffset ||
			a.AutoTriggerMs != b.AutoTriggerMs ||
			a.IntervalType != b.IntervalType ||
			a.IntervalTimeLower != b.IntervalTimeLower ||
			a.IntervalTimeUpper != b.IntervalTimeUpper ||
			// For complex triggers, check slice lengths and pointer equality to detect updates
			len(a.ComplexProperties) != len(b.ComplexProperties) ||
			len(a.ComplexConditions) != len(b.ComplexConditions) ||
			len(a.ComplexDirections) != len(b.ComplexDirections)
	}
	storeSettings := func(setMsg *TriggerDescMsg) (nextFunc eventHandlerFunc) {
		if triggerDescChanged(setMsg.TriggerDesc, triggerSetting) {
			slog.Debug("trigger new", "TriggerADC", setMsg.TriggerADC)
			triggerSetting = setMsg.TriggerDesc
			psControl.requestRestart()
			return changed
		}
		slog.Debug("trigger not new", "TriggerADC", setMsg.TriggerADC)
		return unchanged
	}
	unchanged = func() (nextFunc eventHandlerFunc) {
		slog.Debug("trigger unchanged started")
		select {
		case <-psControl.shutdownCh:
			return nil
		case setMsg := <-psControl.SetTriggerCh:
			defer func() { setMsg.Done <- struct{}{} }()
			return storeSettings(setMsg)
		case getMsg := <-psControl.getTriggerCh:
			getMsg.newSettings <- false
			return unchanged
		}
	}
	changed = func() (nextFunc eventHandlerFunc) {
		slog.Debug("trigger changed started")
		select {
		case <-psControl.shutdownCh:
			return nil
		case setMsg := <-psControl.SetTriggerCh:
			defer func() { setMsg.Done <- struct{}{} }()
			_ = storeSettings(setMsg)
			return changed
		case getMsg := <-psControl.getTriggerCh:
			*getMsg.triggerSettings = triggerSetting
			getMsg.newSettings <- true
			slog.Debug("trigger changed sent")
			return unchanged
		}
	}
	eventHandler := unchanged
	for eventHandler != nil {
		eventHandler = eventHandler()
	}
}

func (psControl *PscDesc) sendSimpleTrigger() (err error) {
	at := int16(0)
	if psControl.triggerSetting.Mode == Auto {
		at = autoTriggerMs
	}
	err = psControl.Con.SetSimpleTrigger(psControl.triggerSetting.Enabled,
		psControl.triggerSetting.Source, psControl.triggerSetting.TriggerADC,
		psControl.triggerSetting.ThresholdDirection, 0, at)
	if err != nil {
		slog.Error("setSimpleTrigger", "error:", err)
		return
	}
	return
}

func (psControl *PscDesc) getValidTriggerProperties() []genericps.TriggerChannelProperties {
	upper := psControl.triggerSetting.TriggerADC
	lower := psControl.triggerSetting.LowerTriggerADC
	upperHyst := psControl.triggerSetting.HysteresisADC
	lowerHyst := psControl.triggerSetting.LowerHysteresisADC
	slog.Debug("validprop", "genericps.Window", genericps.Window)
	if psControl.triggerSetting.ThresholdMode == genericps.Window {
		if lower >= upper {
			slog.Warn("getValidTriggerProperties: lower >= upper, correcting automatically", "lower", lower, "upper", upper)
			if lower > upper {
				lower, upper = upper, lower
			}
			if lower == upper {
				if upper < 32767 {
					upper++
				} else {
					lower--
				}
			}
		}
	}

	return []genericps.TriggerChannelProperties{{
		ThresholdUpper:           upper,
		ThresholdUpperHysteresis: upperHyst,
		ThresholdLower:           lower,
		ThresholdLowerHysteresis: lowerHyst,
		Channel:                  psControl.triggerSetting.Source,
		ThresholdMode:            psControl.triggerSetting.ThresholdMode,
	}}
}

func (psControl *PscDesc) sendAdvancedTrigger() (err error) {
	at := int32(0)
	if psControl.triggerSetting.Mode == Auto {
		at = autoTriggerMs
	}
	channelProperties := psControl.getValidTriggerProperties()

	var triggerConditions []genericps.TriggerConditions
	switch psControl.triggerSetting.Source {
	case genericps.ChA:
		triggerConditions = []genericps.TriggerConditions{{ChannelA: genericps.CondTrue, ChannelB: genericps.CondDontCare, ChannelC: genericps.CondDontCare,
			ChannelD: genericps.CondDontCare, External: genericps.CondDontCare, Aux: genericps.CondDontCare, PulseWidthQualifier: genericps.CondDontCare, Digital: genericps.CondDontCare}}
	case genericps.ChB:
		triggerConditions = []genericps.TriggerConditions{{ChannelA: genericps.CondDontCare, ChannelB: genericps.CondTrue, ChannelC: genericps.CondDontCare,
			ChannelD: genericps.CondDontCare, External: genericps.CondDontCare, Aux: genericps.CondDontCare, PulseWidthQualifier: genericps.CondDontCare, Digital: genericps.CondDontCare}}
	case genericps.ChC:
		triggerConditions = []genericps.TriggerConditions{{ChannelA: genericps.CondDontCare, ChannelB: genericps.CondDontCare, ChannelC: genericps.CondTrue,
			ChannelD: genericps.CondDontCare, External: genericps.CondDontCare, Aux: genericps.CondDontCare, PulseWidthQualifier: genericps.CondDontCare, Digital: genericps.CondDontCare}}
	case genericps.ChD:
		triggerConditions = []genericps.TriggerConditions{{ChannelA: genericps.CondDontCare, ChannelB: genericps.CondDontCare, ChannelC: genericps.CondDontCare,
			ChannelD: genericps.CondTrue, External: genericps.CondDontCare, Aux: genericps.CondDontCare, PulseWidthQualifier: genericps.CondDontCare, Digital: genericps.CondDontCare}}
	}

	slog.Debug("Prop", "prop", channelProperties)
	err = psControl.Con.SetTriggerChannelProperties(channelProperties, false, at)
	if err != nil {
		slog.Error("runblock SetTriggerChannelProperties:", "error:", err, "channelProperties:", channelProperties)
		return
	}

	err = psControl.Con.SetTriggerChannelConditions(triggerConditions)
	if err != nil {
		slog.Error("runblock SetTriggerChannelCondition:", "error:", err)
		return
	}

	channelA := genericps.TriggerNone
	channelB := genericps.TriggerNone
	channelC := genericps.TriggerNone
	channelD := genericps.TriggerNone
	ext := genericps.TriggerNone
	aux := genericps.TriggerNone
	dir := psControl.triggerSetting.ThresholdDirection

	switch psControl.triggerSetting.Source {
	case genericps.ChA:
		channelA = dir
	case genericps.ChB:
		channelB = dir
	case genericps.ChC:
		channelC = dir
	case genericps.ChD:
		channelD = dir
	}

	err = psControl.Con.SetTriggerChannelDirections(channelA,
		channelB,
		channelC,
		channelD,
		ext,
		aux)

	if err != nil {
		slog.Error("SetTriggerChannelDirections:", "error:", err)
		return
	}

	pwqDir := genericps.TriggerRaising
	if dir == genericps.TriggerFalling {
		pwqDir = genericps.TriggerFalling
	}
	err = psControl.Con.SetPulseWidthQualifier(nil, pwqDir, 0, 0, genericps.PwTypeNone)
	if err != nil {
		slog.Error("SetPulseWidthQualifier disable:", "error:", err)
	}
	return
}

func (psControl *PscDesc) sendWindowTrigger() (err error) {
	at := int32(0)
	if psControl.triggerSetting.Mode == Auto {
		at = autoTriggerMs
	}
	var triggerConditions []genericps.TriggerConditions
	pwqCond := genericps.CondDontCare
	condMain := genericps.CondTrue

	switch psControl.triggerSetting.Source {
	case genericps.ChA:
		triggerConditions = []genericps.TriggerConditions{{ChannelA: condMain, ChannelB: genericps.CondDontCare, ChannelC: genericps.CondDontCare,
			ChannelD: genericps.CondDontCare, External: genericps.CondDontCare, Aux: genericps.CondDontCare, PulseWidthQualifier: pwqCond, Digital: genericps.CondDontCare}}
	case genericps.ChB:
		triggerConditions = []genericps.TriggerConditions{{ChannelA: genericps.CondDontCare, ChannelB: condMain, ChannelC: genericps.CondDontCare,
			ChannelD: genericps.CondDontCare, External: genericps.CondDontCare, Aux: genericps.CondDontCare, PulseWidthQualifier: pwqCond, Digital: genericps.CondDontCare}}
	case genericps.ChC:
		triggerConditions = []genericps.TriggerConditions{{ChannelA: genericps.CondDontCare, ChannelB: genericps.CondDontCare, ChannelC: condMain,
			ChannelD: genericps.CondDontCare, External: genericps.CondDontCare, Aux: genericps.CondDontCare, PulseWidthQualifier: pwqCond, Digital: genericps.CondDontCare}}
	case genericps.ChD:
		triggerConditions = []genericps.TriggerConditions{{ChannelA: genericps.CondDontCare, ChannelB: genericps.CondDontCare, ChannelC: genericps.CondDontCare,
			ChannelD: condMain, External: genericps.CondDontCare, Aux: genericps.CondDontCare, PulseWidthQualifier: pwqCond, Digital: genericps.CondDontCare}}
	}

	channelProperties := psControl.getValidTriggerProperties()

	slog.Debug("Prop", "prop", channelProperties)
	err = psControl.Con.SetTriggerChannelProperties(channelProperties, false, at)
	if err != nil {
		slog.Error("runblock SetTriggerChannelProperties:", "error:", err, "channelProperties:", channelProperties)
		return
	}

	err = psControl.Con.SetTriggerChannelConditions(triggerConditions)
	if err != nil {
		slog.Error("runblock SetTriggerChannelCondition:", "error:", err)
		return
	}

	channelA := genericps.TriggerNone
	channelB := genericps.TriggerNone
	channelC := genericps.TriggerNone
	channelD := genericps.TriggerNone
	ext := genericps.TriggerNone
	aux := genericps.TriggerNone
	dir := psControl.triggerSetting.ThresholdDirection

	switch psControl.triggerSetting.Source {
	case genericps.ChA:
		channelA = dir
	case genericps.ChB:
		channelB = dir
	case genericps.ChC:
		channelC = dir
	case genericps.ChD:
		channelD = dir
	}

	err = psControl.Con.SetTriggerChannelDirections(channelA,
		channelB,
		channelC,
		channelD,
		ext,
		aux)

	if err != nil {
		slog.Error("SetTriggerChannelDirections:", "error:", err)
		return
	}

	pwqDir := genericps.TriggerRaising
	if dir == genericps.TriggerFalling {
		pwqDir = genericps.TriggerFalling
	}
	err = psControl.Con.SetPulseWidthQualifier(nil, pwqDir, 0, 0, genericps.PwTypeNone)
	if err != nil {
		slog.Error("SetPulseWidthQualifier disable:", "error:", err)
	}
	return
}

func (psControl *PscDesc) sendIntervalTrigger() (err error) {
	at := int32(0) // Pulse Width Qualifier requires autoTriggerMilliseconds to be 0

	channelProperties := psControl.getValidTriggerProperties()

	pwqCond := genericps.CondDontCare
	isIntervalActive := psControl.triggerSetting.Type == Interval && psControl.triggerSetting.IntervalType != genericps.PwTypeNone
	if isIntervalActive {
		pwqCond = genericps.CondTrue
	}

	condMain := genericps.CondTrue

	var triggerConditions []genericps.TriggerConditions
	switch psControl.triggerSetting.Source {
	case genericps.ChA:
		triggerConditions = []genericps.TriggerConditions{{ChannelA: condMain, ChannelB: genericps.CondDontCare, ChannelC: genericps.CondDontCare,
			ChannelD: genericps.CondDontCare, External: genericps.CondDontCare, Aux: genericps.CondDontCare, PulseWidthQualifier: pwqCond, Digital: genericps.CondDontCare}}
	case genericps.ChB:
		triggerConditions = []genericps.TriggerConditions{{ChannelA: genericps.CondDontCare, ChannelB: condMain, ChannelC: genericps.CondDontCare,
			ChannelD: genericps.CondDontCare, External: genericps.CondDontCare, Aux: genericps.CondDontCare, PulseWidthQualifier: pwqCond, Digital: genericps.CondDontCare}}
	case genericps.ChC:
		triggerConditions = []genericps.TriggerConditions{{ChannelA: genericps.CondDontCare, ChannelB: genericps.CondDontCare, ChannelC: condMain,
			ChannelD: genericps.CondDontCare, External: genericps.CondDontCare, Aux: genericps.CondDontCare, PulseWidthQualifier: pwqCond, Digital: genericps.CondDontCare}}
	case genericps.ChD:
		triggerConditions = []genericps.TriggerConditions{{ChannelA: genericps.CondDontCare, ChannelB: genericps.CondDontCare, ChannelC: genericps.CondDontCare,
			ChannelD: condMain, External: genericps.CondDontCare, Aux: genericps.CondDontCare, PulseWidthQualifier: pwqCond, Digital: genericps.CondDontCare}}
	}

	slog.Debug("Prop", "prop", channelProperties)
	err = psControl.Con.SetTriggerChannelProperties(channelProperties, false, at)
	if err != nil {
		slog.Error("runblock SetTriggerChannelProperties:", "error:", err, "channelProperties:", channelProperties)
		return
	}

	err = psControl.Con.SetTriggerChannelConditions(triggerConditions)
	if err != nil {
		slog.Error("runblock SetTriggerChannelCondition:", "error:", err)
		return
	}

	channelA := genericps.TriggerNone
	channelB := genericps.TriggerNone
	channelC := genericps.TriggerNone
	channelD := genericps.TriggerNone
	ext := genericps.TriggerNone
	aux := genericps.TriggerNone
	dir := psControl.triggerSetting.ThresholdDirection

	switch psControl.triggerSetting.Source {
	case genericps.ChA:
		channelA = dir
	case genericps.ChB:
		channelB = dir
	case genericps.ChC:
		channelC = dir
	case genericps.ChD:
		channelD = dir
	}

	err = psControl.Con.SetTriggerChannelDirections(channelA,
		channelB,
		channelC,
		channelD,
		ext,
		aux)

	if err != nil {
		slog.Error("SetTriggerChannelDirections:", "error:", err)
		return
	}

	if isIntervalActive {
		var pwqConditions []genericps.PwqConditions
		switch psControl.triggerSetting.Source {
		case genericps.ChA:
			pwqConditions = []genericps.PwqConditions{{ChannelA: genericps.CondTrue, ChannelB: genericps.CondDontCare, ChannelC: genericps.CondDontCare, ChannelD: genericps.CondDontCare, External: genericps.CondDontCare, Aux: genericps.CondDontCare, Digital: genericps.CondDontCare}}
		case genericps.ChB:
			pwqConditions = []genericps.PwqConditions{{ChannelA: genericps.CondDontCare, ChannelB: genericps.CondTrue, ChannelC: genericps.CondDontCare, ChannelD: genericps.CondDontCare, External: genericps.CondDontCare, Aux: genericps.CondDontCare, Digital: genericps.CondDontCare}}
		case genericps.ChC:
			pwqConditions = []genericps.PwqConditions{{ChannelA: genericps.CondDontCare, ChannelB: genericps.CondDontCare, ChannelC: genericps.CondTrue, ChannelD: genericps.CondDontCare, External: genericps.CondDontCare, Aux: genericps.CondDontCare, Digital: genericps.CondDontCare}}
		case genericps.ChD:
			pwqConditions = []genericps.PwqConditions{{ChannelA: genericps.CondDontCare, ChannelB: genericps.CondDontCare, ChannelC: genericps.CondDontCare, ChannelD: genericps.CondTrue, External: genericps.CondDontCare, Aux: genericps.CondDontCare, Digital: genericps.CondDontCare}}
		}

		lowerSamples := uint32(1)
		if psControl.triggerSetting.IntervalTimeLower > 0 && psControl.SamplingTimeInterval > 0 {
			samples := uint32(psControl.triggerSetting.IntervalTimeLower / psControl.SamplingTimeInterval)
			if samples > 0 {
				lowerSamples = samples
			}
		}
		upperSamples := uint32(1)
		if psControl.triggerSetting.IntervalTimeUpper > 0 && psControl.SamplingTimeInterval > 0 {
			samples := uint32(psControl.triggerSetting.IntervalTimeUpper / psControl.SamplingTimeInterval)
			if samples > 0 {
				upperSamples = samples
			}
		}

		if lowerSamples > 16777215 {
			lowerSamples = 16777215
		}
		if upperSamples > 16777215 {
			upperSamples = 16777215
		}

		if psControl.triggerSetting.IntervalType == genericps.PwTypeInRange || psControl.triggerSetting.IntervalType == genericps.PwTypeOutOfRange {
			if lowerSamples >= upperSamples {
				upperSamples = lowerSamples + 1
				if upperSamples > 16777215 {
					lowerSamples = 16777214
				}
			}
		}
		pwqDir := dir

		// The PicoScope driver uses the 'lower' parameter for the time limit in single-value modes.
		if psControl.triggerSetting.IntervalType == genericps.PwTypeLessThan {
			lowerSamples = upperSamples
			upperSamples = 0
		} else if psControl.triggerSetting.IntervalType == genericps.PwTypeGreaterThan {
			upperSamples = 0
		}

		slog.Debug("isIntervalActive", "pwqConditions", pwqConditions, "pwqDir", pwqDir,
			"intervalType", psControl.triggerSetting.IntervalType,
			"lowerSamples", lowerSamples, "upperSamples", upperSamples,
			"rawTimeLower", psControl.triggerSetting.IntervalTimeLower,
			"rawTimeUpper", psControl.triggerSetting.IntervalTimeUpper,
			"samplingInterval", psControl.SamplingTimeInterval)
		err = psControl.Con.SetPulseWidthQualifier(pwqConditions, pwqDir, lowerSamples, upperSamples, psControl.triggerSetting.IntervalType)
		if err != nil {
			slog.Error("SetPulseWidthQualifier:", "error:", err)
			return
		}
	} else {
		pwqDir := genericps.TriggerRaising
		if dir == genericps.TriggerFalling {
			pwqDir = genericps.TriggerFalling
		}
		err = psControl.Con.SetPulseWidthQualifier(nil, pwqDir, 0, 0, genericps.PwTypeNone)
		if err != nil {
			slog.Error("SetPulseWidthQualifier disable:", "error:", err)
		}
	}

	return
}

func (psControl *PscDesc) sendComplexTrigger() (err error) {

	at := int32(0)
	if psControl.triggerSetting.Mode == Auto {
		at = autoTriggerMs
	}

	err = psControl.Con.SetTriggerChannelProperties(psControl.triggerSetting.ComplexProperties, false, at)
	if err != nil {
		slog.Error("SetTriggerChannelProperties (Complex):", "error:", err, "properties:", psControl.triggerSetting.ComplexProperties)
		return
	}
	err = psControl.Con.SetTriggerChannelConditions(psControl.triggerSetting.ComplexConditions)
	if err != nil {
		slog.Error("SetTriggerChannelCondition (Complex):", "error:", err)
		return
	}
	dirs := psControl.triggerSetting.ComplexDirections
	if len(dirs) > 0 {
		err = psControl.Con.SetTriggerChannelDirections(dirs[0].ChannelA, dirs[0].ChannelB, dirs[0].ChannelC, dirs[0].ChannelD, dirs[0].Ext, dirs[0].Aux)
	} else {
		err = psControl.Con.SetTriggerChannelDirections(genericps.TriggerNone, genericps.TriggerNone, genericps.TriggerNone, genericps.TriggerNone, genericps.TriggerNone, genericps.TriggerNone)
	}
	if err != nil {
		slog.Error("SetTriggerChannelDirections (Complex):", "error:", err)
		return
	}
	return
}
