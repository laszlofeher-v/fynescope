package gui

import (
	"fynescope/control"
	"fynescope/genericps"
)

// buildComplexTriggerMessage converts the GUI settings (in mV) into ADC values
// and prepares the arrays for the backend control logic.
func (scp *ScpDesc) buildComplexTriggerMessage() {
	var props []genericps.TriggerChannelProperties
	var dirs []control.TriggerDirections
	var directionA, directionB, directionC, directionD genericps.ThresholdDirection
	directionA = genericps.TriggerNone
	directionB = genericps.TriggerNone
	directionC = genericps.TriggerNone
	directionD = genericps.TriggerNone

	condition := genericps.TriggerConditions{
		ChannelA:            genericps.CondDontCare,
		ChannelB:            genericps.CondDontCare,
		ChannelC:            genericps.CondDontCare,
		ChannelD:            genericps.CondDontCare,
		External:            genericps.CondDontCare,
		Aux:                 genericps.CondDontCare,
		PulseWidthQualifier: genericps.CondDontCare,
		Digital:             genericps.CondDontCare,
	}

	for i, ch := range scp.Settings.Channels {
		chCfg := ch.Trigger
		if chCfg.Condition != genericps.CondDontCare {
			vRange := ch.VRange

			props = append(props, genericps.TriggerChannelProperties{
				ThresholdUpper:           int16(scp.mvToAdc(chCfg.Mv, vRange)),
				ThresholdUpperHysteresis: uint16(scp.mvToUAdc(chCfg.Hysteresis, vRange)),
				ThresholdLower:           int16(scp.mvToAdc(chCfg.LowerMv, vRange)),
				ThresholdLowerHysteresis: uint16(scp.mvToUAdc(chCfg.LowerHysteresis, vRange)),
				Channel:                  genericps.ChannelId(i),
				ThresholdMode:            chCfg.ThresholdMode,
			})

			switch genericps.ChannelId(i) {
			case genericps.ChA:
				condition.ChannelA = chCfg.Condition
				directionA = chCfg.TriggerDirection
			case genericps.ChB:
				condition.ChannelB = chCfg.Condition
				directionB = chCfg.TriggerDirection
			case genericps.ChC:
				condition.ChannelC = chCfg.Condition
				directionC = chCfg.TriggerDirection
			case genericps.ChD:
				condition.ChannelD = chCfg.Condition
				directionD = chCfg.TriggerDirection
			}
		}
	}

	dirs = append(dirs, control.TriggerDirections{
		ChannelA: directionA,
		ChannelB: directionB,
		ChannelC: directionC,
		ChannelD: directionD,
		Ext:      genericps.TriggerNone,
		Aux:      genericps.TriggerNone,
	})

	scp.triggerSettingMsg.ComplexProperties = props
	scp.triggerSettingMsg.ComplexConditions = []genericps.TriggerConditions{condition}
	scp.triggerSettingMsg.ComplexDirections = dirs
}
