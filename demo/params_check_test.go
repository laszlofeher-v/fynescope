package demo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDemo_ParameterChecks(t *testing.T) {
	behaviour = normal

	// 1. Invalid Handle Checks
	t.Run("InvalidHandle", func(t *testing.T) {
		assert.Error(t, simSetChannel(0, ChA, true, Dc, Range_1v, 0))
		assert.Error(t, simSetChannel(-1, ChA, true, Dc, Range_1v, 0))
		_, _, err := simGetAnalogueOffset(0, int(Range_1v), Dc)
		assert.Error(t, err)
		_, err = simGetChannelInformation(0, 0, 0, nil, ChA)
		assert.Error(t, err)
		assert.Error(t, simSetDataBuffer(0, ChA, []int16{1, 2}, 0, 0))
		assert.Error(t, simSetDataBuffers(0, ChA, []int16{1}, []int16{1}, 0, 0))
		_, err = simRunBlock(0, 10, 10, 2, 0, 0, nil, nil)
		assert.Error(t, err)
		_, err = simRunStreaming(0, 100, TuNs, 10, 10, false, 1, 0, 100)
		assert.Error(t, err)
		assert.Error(t, simSetSimpleTrigger(0, true, ChA, 0, TriggerRising, 0, 0))
		assert.Error(t, simSetTriggerChannelProperties(0, nil, false, 0))
		assert.Error(t, simSetTriggerChannelConditions(0, nil))
		assert.Error(t, simSetTriggerChannelDirections(0, TriggerNone, TriggerNone, TriggerNone, TriggerNone, TriggerNone, TriggerNone))
		assert.Error(t, simSetPulseWidthQualifier(0, nil, TriggerNone, 0, 0, PwTypeNone))
		assert.Error(t, simSetDigitalPort(0, Port0, true, 0))
	})

	// 2. SetChannel Parameter Checks
	t.Run("SetChannel", func(t *testing.T) {
		handle, err := openUnit("", 0)
		assert.NoError(t, err)

		// Invalid Channel
		assert.Error(t, simSetChannel(handle, -1, true, Dc, Range_1v, 0))
		assert.Error(t, simSetChannel(handle, ChannelId(MaxChannels), true, Dc, Range_1v, 0))

		// Invalid Voltage Range
		assert.Error(t, simSetChannel(handle, ChA, true, Dc, -1, 0))
		assert.Error(t, simSetChannel(handle, ChA, true, Dc, Range_50v+1, 0))

		// Invalid Coupling
		assert.Error(t, simSetChannel(handle, ChA, true, Coupling(99), Range_1v, 0))

		// Invalid Analog Offset (out of ±20V)
		assert.Error(t, simSetChannel(handle, ChA, true, Dc, Range_1v, 25.0))
		assert.Error(t, simSetChannel(handle, ChA, true, Dc, Range_1v, -25.0))

		// Valid Channel
		assert.NoError(t, simSetChannel(handle, ChA, true, Dc, Range_1v, 0))
		assert.NoError(t, simSetChannel(handle, ChB, true, Ac, Range_500mv, 1.5))
	})

	// 3. GetAnalogueOffset Parameter Checks
	t.Run("GetAnalogueOffset", func(t *testing.T) {
		handle, _ := openUnit("", 0)
		_, _, err := simGetAnalogueOffset(handle, -1, Dc)
		assert.Error(t, err)
		_, _, err = simGetAnalogueOffset(handle, int(Range_50v)+1, Dc)
		assert.Error(t, err)
		_, _, err = simGetAnalogueOffset(handle, int(Range_1v), Coupling(99))
		assert.Error(t, err)

		maxV, minV, err := simGetAnalogueOffset(handle, int(Range_1v), Dc)
		assert.NoError(t, err)
		assert.Equal(t, float32(20), maxV)
		assert.Equal(t, float32(-20), minV)
	})

	// 4. DataBuffer Parameter Checks
	t.Run("DataBuffers", func(t *testing.T) {
		handle, _ := openUnit("", 0)

		// Empty buffer
		assert.Error(t, simSetDataBuffer(handle, ChA, nil, 0, 0))
		assert.Error(t, simSetDataBuffer(handle, ChA, []int16{}, 0, 0))
		assert.Error(t, simSetDataBuffers(handle, ChA, nil, nil, 0, 0))

		// Invalid segment
		assert.Error(t, simSetDataBuffer(handle, ChA, []int16{1, 2}, 1, 0))
		assert.Error(t, simSetDataBuffers(handle, ChA, []int16{1}, []int16{1}, 1, 0))

		// Valid buffers
		assert.NoError(t, simSetDataBuffer(handle, ChA, make([]int16, 100), 0, 0))
		assert.NoError(t, simSetDataBuffers(handle, ChA, make([]int16, 100), make([]int16, 100), 0, 0))
	})

	// 5. Trigger Parameter Checks
	t.Run("TriggerChecks", func(t *testing.T) {
		handle, _ := openUnit("", 0)

		// Simple Trigger: invalid channel
		assert.Error(t, simSetSimpleTrigger(handle, true, -1, 0, TriggerRising, 0, 0))
		assert.Error(t, simSetSimpleTrigger(handle, true, ChannelId(10), 0, TriggerRising, 0, 0))

		// Simple Trigger: invalid direction
		assert.Error(t, simSetSimpleTrigger(handle, true, ChA, 0, -1, 0, 0))
		assert.Error(t, simSetSimpleTrigger(handle, true, ChA, 0, ThresholdDirection(99), 0, 0))

		// Simple Trigger: negative autoTriggerMs
		assert.Error(t, simSetSimpleTrigger(handle, true, ChA, 0, TriggerRising, 0, -5))

		// Advanced Trigger Channel Properties: Window mode lower > upper
		props := []TriggerChannelProperties{
			{
				Channel:        ChA,
				ThresholdMode:  Window,
				ThresholdLower: 1000,
				ThresholdUpper: 500, // lower > upper is invalid
			},
		}
		assert.Error(t, simSetTriggerChannelProperties(handle, props, false, 0))

		// PWQ: InRange with lower > upper
		assert.Error(t, simSetPulseWidthQualifier(handle, nil, TriggerRising, 1000, 500, PwTypeInRange))
		assert.NoError(t, simSetPulseWidthQualifier(handle, nil, TriggerRising, 500, 1000, PwTypeInRange))
	})

	// 6. Block & Streaming Parameter Checks
	t.Run("BlockAndStreaming", func(t *testing.T) {
		handle, _ := openUnit("", 0)

		// RunBlock: 0 samples
		_, err := simRunBlock(handle, 0, 0, 2, 0, 0, nil, nil)
		assert.Error(t, err)

		// RunBlock: negative samples
		_, err = simRunBlock(handle, -1, 10, 2, 0, 0, nil, nil)
		assert.Error(t, err)

		// RunStreaming: 0 sample interval
		_, err = simRunStreaming(handle, 0, TuNs, 10, 10, false, 1, 0, 100)
		assert.Error(t, err)

		// RunStreaming: invalid time units
		_, err = simRunStreaming(handle, 100, TimeUnits(99), 10, 10, false, 1, 0, 100)
		assert.Error(t, err)
	})

	// 7. Digital Port Checks
	t.Run("DigitalPort", func(t *testing.T) {
		handle, _ := openUnit("", 0)

		// Invalid port
		assert.Error(t, simSetDigitalPort(handle, -1, true, 0))
		assert.Error(t, simSetDigitalPort(handle, MaxDigitalPorts, true, 0))

		// Valid port
		assert.NoError(t, simSetDigitalPort(handle, Port0, true, 0))
		assert.NoError(t, simSetDigitalPort(handle, Port1, true, 1000))
	})
}
