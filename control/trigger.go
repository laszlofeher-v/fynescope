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
)

func (psControl *PscDesc) triggerMonitor() {
	type (
		eventHandlerFunc func() (nextFunc eventHandlerFunc)
	)
	var (
		unchanged, changed eventHandlerFunc
		triggerSetting     TriggerDesc
	)
	storeSettings := func(setMsg *TriggerDescMsg) (nextFunc eventHandlerFunc) {
		if setMsg.TriggerDesc != triggerSetting {
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

func (psControl *PscDesc) sendComplexTrigger() (err error) {

	at := int32(0)
	if psControl.triggerSetting.Mode == Auto {

		at = autoTriggerMs
	} else {

	}
	channelProperties := []genericps.TriggerChannelProperties{{ThresholdUpper: psControl.triggerSetting.TriggerADC,
		ThresholdUpperHysteresis: psControl.triggerSetting.HysteresisADC, ThresholdLower: psControl.triggerSetting.TriggerADC,
		ThresholdLowerHysteresis: psControl.triggerSetting.HysteresisADC, Channel: psControl.triggerSetting.Source, ThresholdMode: genericps.Level}}

	slog.Debug("Prop", "prop", channelProperties)
	err = psControl.Con.SetTriggerChannelProperties(channelProperties, false, at)
	if err != nil {
		slog.Error("runblock SetTriggerChannelProperties:", "error:", err, "channelProperties:", channelProperties)
		return
	}

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
	switch psControl.triggerSetting.Source {
	case genericps.ChA:

		channelA = psControl.triggerSetting.ThresholdDirection
	case genericps.ChB:

		channelB = psControl.triggerSetting.ThresholdDirection
	case genericps.ChC:

		channelC = psControl.triggerSetting.ThresholdDirection
	case genericps.ChD:

		channelD = psControl.triggerSetting.ThresholdDirection
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

	return
}
