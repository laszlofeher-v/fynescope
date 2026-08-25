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
}
