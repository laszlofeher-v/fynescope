//go:build sim

package ps2000a

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPs2000aDigitalTriggerSim(t *testing.T) {
	handle := int16(1)

	// Test ps2000aSetDigitalPort
	err := ps2000aSetDigitalPort(handle, Port0, true, 1000)
	assert.NoError(t, err)

	err = ps2000aSetDigitalPort(handle, Port1, true, 1000)
	assert.NoError(t, err)

	// Test ps2000aSetTriggerDigitalPortProperties
	dirs := []DigitalChannelDirections{
		{Channel: Dch0, Direction: DigitalDirectionRising},
		{Channel: Dch1, Direction: DigitalDirectionHigh},
	}
	err = ps2000aSetTriggerDigitalPortProperties(handle, dirs)
	assert.NoError(t, err)

	// Test ps2000aSetDigitalAnalogTriggerOperand
	err = ps2000aSetDigitalAnalogTriggerOperand(handle, OperandAnd)
	assert.NoError(t, err)

	// Test ps2000aSetTriggerChannelConditions with Digital: ConditionTrue
	conds := []TriggerConditions{
		{
			ChannelA: CondTrue,
			Digital:  CondTrue,
		},
	}
	err = ps2000aSetTriggerChannelConditions(handle, conds)
	assert.NoError(t, err)
}
