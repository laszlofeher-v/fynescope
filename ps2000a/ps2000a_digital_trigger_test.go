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

func TestPs2000aStreamingSim(t *testing.T) {
	handle := int16(1)

	// Set channel and data buffer
	err := ps2000aSetChannel(handle, ChA, true, Dc, Range_5v, 0)
	assert.NoError(t, err)

	buf := make([]int16, 1024)
	err = ps2000aSetDataBuffer(handle, ChA, buf, 0, RatioModeNone)
	assert.NoError(t, err)

	// Run streaming
	interval, err := ps2000aRunStreaming(handle, 1000, TuUs, 0, 1024, false, 1, RatioModeNone, 1024)
	assert.NoError(t, err)
	assert.Equal(t, uint32(1000), interval)

	// Get latest values
	var callbackCalled bool
	err = ps2000aGetStreamingLatestValues(handle, func(h int16, noOfSamples int32, startIndex uint32, overflow int16, triggerAt uint32, triggered, autoStop int16, param interface{}) error {
		callbackCalled = true
		return nil
	}, nil)
	assert.NoError(t, err)
	assert.True(t, callbackCalled)

	// Stop
	err = ps2000aStop(handle)
	assert.NoError(t, err)
}

func TestPs2000aEtsSim(t *testing.T) {
	handle := int16(1)

	sampleTime, err := ps2000aSetEts(handle, EtsFast, 40, 20)
	assert.NoError(t, err)
	assert.Equal(t, int32(100), sampleTime)
}
