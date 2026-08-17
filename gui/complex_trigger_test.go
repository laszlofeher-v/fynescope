package gui

import (
	"fynescope/control"
	"fynescope/genericps"
	"fynescope/settings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildComplexTriggerMessage(t *testing.T) {
	// Mock runtime uninitialized genericps variables
	genericps.CondDontCare = genericps.TriggerRespBase(0)
	genericps.CondTrue = genericps.TriggerRespBase(1)
	genericps.CondFalse = genericps.TriggerRespBase(2)

	genericps.TriggerNone = genericps.ThresholdDirection(0)
	genericps.TriggerRising = genericps.ThresholdDirection(1)
	genericps.TriggerFalling = genericps.ThresholdDirection(2)

	genericps.Level = genericps.ThresholdModeId(0)
	genericps.Window = genericps.ThresholdModeId(1)

	genericps.ChA = genericps.ChannelId(0)
	genericps.ChB = genericps.ChannelId(1)
	genericps.ChC = genericps.ChannelId(2)
	genericps.ChD = genericps.ChannelId(3)

	// RangeEnum is used as an index into InputRanges — use index 0 for all channels
	vRange := genericps.RangeEnum(0)
	genericps.InputRanges = []int32{5000}

	scp := &ScpDesc{
		triggerSettingMsg: control.TriggerDescMsg{},
		Settings: &settings.PsSettings{
			Channels: []settings.ChSettings{
				{
					VRange: vRange,
					Trigger: settings.ChTriggerSettings{
						Condition:        genericps.CondTrue, // ChA
						TriggerDirection: genericps.TriggerRising,
						Mv:               1000,
						LowerMv:          500,
						Hysteresis:       10,
						LowerHysteresis:  5,
						ThresholdMode:    genericps.Level,
					},
				},
				{
					VRange: vRange,
					Trigger: settings.ChTriggerSettings{
						Condition: genericps.CondDontCare, // ChB - skipped
					},
				},
				{
					VRange: vRange,
					Trigger: settings.ChTriggerSettings{
						Condition:        genericps.CondFalse, // ChC
						TriggerDirection: genericps.TriggerFalling,
						Mv:               -500,
						ThresholdMode:    genericps.Window,
					},
				},
			},
		},
		MaxValue: 32767,
		MinValue: -32768,
	}

	scp.buildComplexTriggerMessage()

	// Assert ComplexConditions
	assert.Len(t, scp.triggerSettingMsg.ComplexConditions, 1)
	cond := scp.triggerSettingMsg.ComplexConditions[0]
	assert.Equal(t, genericps.CondTrue, cond.ChannelA)
	assert.Equal(t, genericps.CondDontCare, cond.ChannelB)
	assert.Equal(t, genericps.CondFalse, cond.ChannelC)
	assert.Equal(t, genericps.CondDontCare, cond.ChannelD)

	// Assert ComplexDirections
	assert.Len(t, scp.triggerSettingMsg.ComplexDirections, 1)
	dir := scp.triggerSettingMsg.ComplexDirections[0]
	assert.Equal(t, genericps.TriggerRising, dir.ChannelA)
	assert.Equal(t, genericps.TriggerNone, dir.ChannelB)
	assert.Equal(t, genericps.TriggerFalling, dir.ChannelC)
	assert.Equal(t, genericps.TriggerNone, dir.ChannelD)

	// Assert ComplexProperties - only ChA (i=0) and ChC (i=2) should be present
	assert.Len(t, scp.triggerSettingMsg.ComplexProperties, 2)
	propA := scp.triggerSettingMsg.ComplexProperties[0]
	assert.Equal(t, genericps.ChA, propA.Channel)
	assert.Equal(t, genericps.Level, propA.ThresholdMode)

	propC := scp.triggerSettingMsg.ComplexProperties[1]
	assert.Equal(t, genericps.ChC, propC.Channel)
	assert.Equal(t, genericps.Window, propC.ThresholdMode)
}
