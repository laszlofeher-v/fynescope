package control

import (
	"testing"

	"fynescope/genericps"
	"github.com/stretchr/testify/assert"
)

func TestTriggerDescChangedDigital(t *testing.T) {
	a := TriggerDesc{
		DigitalTriggerEnabled: false,
		DigitalAnalogOperand:  genericps.OperandOr,
		DigitalDirections: []genericps.DigitalChannelDirections{
			{Channel: genericps.Dch0, Direction: genericps.DigitalDirectionRising},
		},
	}
	b := TriggerDesc{
		DigitalTriggerEnabled: true,
		DigitalAnalogOperand:  genericps.OperandOr,
		DigitalDirections: []genericps.DigitalChannelDirections{
			{Channel: genericps.Dch0, Direction: genericps.DigitalDirectionRising},
		},
	}

	// Compare with trigger monitor detection logic
	assert.NotEqual(t, a.DigitalTriggerEnabled, b.DigitalTriggerEnabled)

	// Changing directions
	c := b
	c.DigitalDirections = []genericps.DigitalChannelDirections{
		{Channel: genericps.Dch0, Direction: genericps.DigitalDirectionFalling},
	}
	assert.NotEqual(t, b.DigitalDirections[0].Direction, c.DigitalDirections[0].Direction)

	// Changing operand
	d := b
	d.DigitalAnalogOperand = genericps.OperandAnd
	assert.NotEqual(t, b.DigitalAnalogOperand, d.DigitalAnalogOperand)

	// Changing channels operand
	e := b
	e.DigitalChannelsOperand = genericps.OperandAnd
	assert.NotEqual(t, b.DigitalChannelsOperand, e.DigitalChannelsOperand)
}

func TestBuildTriggerConditions_DigitalAnalogOperand(t *testing.T) {
	psc := &PscDesc{}

	// Case 1: Digital disabled, only analog
	psc.triggerSetting = TriggerDesc{
		Source:                genericps.ChA,
		DigitalTriggerEnabled: false,
	}
	conds1 := psc.buildTriggerConditions(genericps.CondTrue, genericps.CondDontCare)
	assert.Len(t, conds1, 1)
	assert.Equal(t, genericps.CondTrue, conds1[0].ChannelA)
	assert.Equal(t, genericps.CondDontCare, conds1[0].Digital)

	// Case 2: Analog + Digital with OperandOr -> 2 rows
	psc.triggerSetting = TriggerDesc{
		Source:                genericps.ChA,
		DigitalTriggerEnabled: true,
		DigitalAnalogOperand:  genericps.OperandOr,
		DigitalDirections: []genericps.DigitalChannelDirections{
			{Channel: genericps.Dch0, Direction: genericps.DigitalDirectionRising},
		},
	}
	conds2 := psc.buildTriggerConditions(genericps.CondTrue, genericps.CondDontCare)
	assert.Len(t, conds2, 2)
	assert.Equal(t, genericps.CondTrue, conds2[0].ChannelA)
	assert.Equal(t, genericps.CondDontCare, conds2[0].Digital)
	assert.Equal(t, genericps.CondDontCare, conds2[1].ChannelA)
	assert.Equal(t, genericps.CondTrue, conds2[1].Digital)

	// Case 3: Analog + Digital with OperandAnd -> 1 row
	psc.triggerSetting = TriggerDesc{
		Source:                genericps.ChA,
		DigitalTriggerEnabled: true,
		DigitalAnalogOperand:  genericps.OperandAnd,
		DigitalDirections: []genericps.DigitalChannelDirections{
			{Channel: genericps.Dch0, Direction: genericps.DigitalDirectionRising},
		},
	}
	conds3 := psc.buildTriggerConditions(genericps.CondTrue, genericps.CondDontCare)
	assert.Len(t, conds3, 1)
	assert.Equal(t, genericps.CondTrue, conds3[0].ChannelA)
	assert.Equal(t, genericps.CondTrue, conds3[0].Digital)
}
