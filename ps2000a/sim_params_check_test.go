//go:build sim

package ps2000a

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPs2000a_SimParameterChecks(t *testing.T) {
	// 1. Invalid Handle Checks
	t.Run("InvalidHandle", func(t *testing.T) {
		assert.Error(t, ps2000aCloseUnit(0))
		assert.Error(t, ps2000aCloseUnit(-1))

		_, _, err := ps2000aGetAnalogueOffset(0, 0, Dc)
		assert.Error(t, err)

		assert.Error(t, ps2000aSetChannel(0, ChA, true, Dc, Range_1v, 0))
		assert.Error(t, ps2000aSetSimpleTrigger(0, true, ChA, 0, TriggerRising, 0, 0))

		buf := make([]int16, 10)
		assert.Error(t, ps2000aSetDataBuffer(0, ChA, buf, 0, RatioModeNone))
		assert.Error(t, ps2000aSetDataBuffers(0, ChA, buf, buf, 0, RatioModeNone))

		_, err = ps2000aRunBlock(0, 10, 10, 2, 0, 0, nil, nil)
		assert.Error(t, err)

		_, err = ps2000aRunStreaming(0, 100, TuNs, 10, 10, false, 1, RatioModeNone, 100)
		assert.Error(t, err)

		assert.Error(t, ps2000aSetDigitalPort(0, Port0, true, 0))
		assert.Error(t, ps2000aStop(0))
	})

	// 2. SetChannel Range Checks
	t.Run("SetChannel", func(t *testing.T) {
		handle := int16(1)

		// Invalid Channel
		assert.Error(t, ps2000aSetChannel(handle, ChannelId(-1), true, Dc, Range_1v, 0))
		assert.Error(t, ps2000aSetChannel(handle, ChannelId(4), true, Dc, Range_1v, 0))

		// Invalid Coupling (not AC/DC)
		assert.Error(t, ps2000aSetChannel(handle, ChA, true, Coupling(-1), Range_1v, 0))
		assert.Error(t, ps2000aSetChannel(handle, ChA, true, Coupling(2), Range_1v, 0))

		// Invalid Voltage Range
		assert.Error(t, ps2000aSetChannel(handle, ChA, true, Dc, RangeEnum(-1), 0))
		assert.Error(t, ps2000aSetChannel(handle, ChA, true, Dc, RangeEnum(20), 0))

		// Valid Channel
		assert.NoError(t, ps2000aSetChannel(handle, ChA, true, Dc, Range_1v, 0))
		assert.NoError(t, ps2000aSetChannel(handle, ChB, true, Ac, Range_500mv, 0))
	})

	// 3. SetSimpleTrigger Checks
	t.Run("SetSimpleTrigger", func(t *testing.T) {
		handle := int16(1)

		// Invalid Source
		assert.Error(t, ps2000aSetSimpleTrigger(handle, true, ChannelId(-1), 0, TriggerRising, 0, 0))
		assert.Error(t, ps2000aSetSimpleTrigger(handle, true, ChannelId(10), 0, TriggerRising, 0, 0))

		// Invalid Direction
		assert.Error(t, ps2000aSetSimpleTrigger(handle, true, ChA, 0, ThresholdDirection(-1), 0, 0))
		assert.Error(t, ps2000aSetSimpleTrigger(handle, true, ChA, 0, ThresholdDirection(10), 0, 0))

		// Invalid autoTriggerMs
		assert.Error(t, ps2000aSetSimpleTrigger(handle, true, ChA, 0, TriggerRising, 0, -1))

		// Valid SimpleTrigger
		assert.NoError(t, ps2000aSetSimpleTrigger(handle, true, ChA, 1000, TriggerRising, 0, 100))
	})

	// 4. Signal Generator Limits
	t.Run("SigGen", func(t *testing.T) {
		handle := int16(1)

		// Output over voltage (> 4V total: pkToPk + 2*|offset| = 1V + 2*2V = 5V)
		assert.Error(t, ps2000aSetSigGenBuiltIn(handle, 2000000, 1000000, Sine, 1000, 1000, 0, 0, SweepUp, EsOff, 0, 0, SigGenRising, SigGenNone, 0))

		// Invalid wave type
		assert.Error(t, ps2000aSetSigGenBuiltIn(handle, 0, 1000000, WaveTypeEnum(99), 1000, 1000, 0, 0, SweepUp, EsOff, 0, 0, SigGenRising, SigGenNone, 0))

		// Invalid frequency on AC waveform
		assert.Error(t, ps2000aSetSigGenBuiltIn(handle, 0, 1000000, Sine, 0, 0, 0, 0, SweepUp, EsOff, 0, 0, SigGenRising, SigGenNone, 0))
		assert.Error(t, ps2000aSetSigGenBuiltInV2(handle, 0, 1000000, Sine, -1, 1000, 0, 0, SweepUp, EsOff, 0, 0, SigGenRising, SigGenNone, 0))

		// Valid SigGen
		assert.NoError(t, ps2000aSetSigGenBuiltIn(handle, 0, 1000000, Sine, 1000, 1000, 0, 0, SweepUp, EsOff, 0, 0, SigGenRising, SigGenNone, 0))

		// Turning generator off uses DcVoltage with 0 Hz / 0 Vpp
		assert.NoError(t, ps2000aSetSigGenBuiltIn(handle, 0, 0, DcVoltage, 0, 0, 0, 0, SweepUp, EsOff, 0, 0, SigGenRising, SigGenNone, 0))
		assert.NoError(t, ps2000aSetSigGenBuiltInV2(handle, 0, 0, DcVoltage, 0, 0, 0, 0, SweepUp, EsOff, 0, 0, SigGenRising, SigGenNone, 0))
	})

	// 5. Digital Port Checks
	t.Run("DigitalPort", func(t *testing.T) {
		handle := int16(1)

		// Invalid digital port
		assert.Error(t, ps2000aSetDigitalPort(handle, DigitalPort(0x90), true, 0))

		// Invalid logic level
		assert.Error(t, ps2000aSetDigitalPort(handle, Port0, true, -32768))

		// Valid digital port
		assert.NoError(t, ps2000aSetDigitalPort(handle, Port0, true, 0))
		assert.NoError(t, ps2000aSetDigitalPort(handle, Port1, true, 1000))
	})

	// 6. Pulse Width Qualifier Checks
	t.Run("PulseWidthQualifier", func(t *testing.T) {
		handle := int16(1)
		conds := []PwqConditions{
			{ChannelA: CondTrue},
		}

		// GreaterThan with lower=1, upper=0 is valid (upper is unused for GreaterThan)
		assert.NoError(t, ps2000aSetPulseWidthQualifier(handle, conds, TriggerFallingLower, 1, 0, PwTypeGreaterThan))

		// LessThan with lower=1, upper=0 is valid
		assert.NoError(t, ps2000aSetPulseWidthQualifier(handle, conds, TriggerRisingLower, 1, 0, PwTypeLessThan))

		// InRange with lower <= upper is valid
		assert.NoError(t, ps2000aSetPulseWidthQualifier(handle, conds, TriggerRisingLower, 10, 20, PwTypeInRange))

		// InRange with lower > upper is invalid
		assert.Error(t, ps2000aSetPulseWidthQualifier(handle, conds, TriggerRisingLower, 20, 10, PwTypeInRange))

		// OutOfRange with lower > upper is invalid
		assert.Error(t, ps2000aSetPulseWidthQualifier(handle, conds, TriggerRisingLower, 20, 10, PwTypeOutOfRange))
	})

	// 7. Trigger Channel Properties (Window Mode)
	t.Run("TriggerChannelPropertiesWindow", func(t *testing.T) {
		handle := int16(1)

		// Window mode with upper - lower < 512 is rejected
		pSmall := []TriggerChannelProperties{
			{ThresholdUpper: 256, ThresholdLower: 0, Channel: ChA, ThresholdMode: Window},
		}
		assert.Error(t, ps2000aSetTriggerChannelProperties(handle, pSmall, false, 0))

		// Window mode with upper - lower = 511 is rejected
		p511 := []TriggerChannelProperties{
			{ThresholdUpper: 511, ThresholdLower: 0, Channel: ChA, ThresholdMode: Window},
		}
		assert.Error(t, ps2000aSetTriggerChannelProperties(handle, p511, false, 0))

		// Window mode with upper - lower >= 512 is accepted
		p512 := []TriggerChannelProperties{
			{ThresholdUpper: 512, ThresholdLower: 0, Channel: ChA, ThresholdMode: Window},
		}
		assert.NoError(t, ps2000aSetTriggerChannelProperties(handle, p512, false, 0))
	})

	// 8. Normal mode (RatioModeNone) capture with generator on produces non-zero samples
	t.Run("SimRunBlockRatioModeNonePopulatesSamples", func(t *testing.T) {
		handle := int16(1)
		simOpenUnit(handle)

		require.NoError(t, ps2000aSetChannel(handle, ChA, true, Dc, Range_1v, 0))
		require.NoError(t, ps2000aSetSigGenBuiltInV2(handle, 0, 2000000, Sine, 1000, 1000, 0, 1, SweepDown, EsOff, 0, 0, SigGenRising, SigGenNone, 0))
		require.NoError(t, ps2000aSetSimpleTrigger(handle, true, ChA, 0, TriggerRising, 0, 1000))

		buf := make([]int16, 1000)
		require.NoError(t, ps2000aSetDataBuffer(handle, ChA, buf, 0, RatioModeNone))

		_, err := ps2000aRunBlock(handle, 500, 500, 1000, 0, 0, nil, nil)
		require.NoError(t, err)

		deadline := time.Now().Add(2 * time.Second)
		ready := int16(0)
		for time.Now().Before(deadline) {
			r, err := ps2000aIsReady(handle)
			require.NoError(t, err)
			if r != 0 {
				ready = r
				break
			}
			time.Sleep(5 * time.Millisecond)
		}
		require.NotZero(t, ready, "simRunBlock did not finish in time")

		hasNonZero := false
		for _, v := range buf {
			if v != 0 {
				hasNonZero = true
				break
			}
		}
		assert.True(t, hasNonZero, "expected non-zero samples in RatioModeNone buffer when generator is on")
	})

	// 9. Streaming callback preserves triggered and autoStop parameter order
	t.Run("StreamingCallbackOrder", func(t *testing.T) {
		prevCb := regLpStreamingReadyGo
		prevParam := regLpStreamingReadyParam
		defer func() {
			regLpStreamingReadyGo = prevCb
			regLpStreamingReadyParam = prevParam
		}()

		var gotTriggered, gotAutoStop int16
		regLpStreamingReadyGo = func(_ int16, _ int32, _ uint32, _ int16, _ uint32, triggered, autoStop int16, _ any) error {
			gotTriggered = triggered
			gotAutoStop = autoStop
			return nil
		}
		lpStreamingReadyGo(1, 10, 0, 0, 5, 1, 0, nil)
		assert.Equal(t, int16(1), gotTriggered)
		assert.Equal(t, int16(0), gotAutoStop)
	})
}
