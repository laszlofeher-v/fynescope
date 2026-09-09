package demo

import (
	"fynescope/genericps"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDigitalTriggerRisingEdge(t *testing.T) {
	// Set up demo digital generator at 1000 Hz, binary counter
	SetDigitalGenPortEnabled(true, true)
	s := &SimDesc{}
	_ = s.SetDemoDigitalGen(true, true, 1000, genericps.DigitalDemoGenDirectionUp, genericps.DigitalDemoGenEncodingBinary, genericps.DigitalDemoGenModeSynchronous, 0)

	td := NewTriggerDetector(false, 0, 0, TriggerNone, ChA)
	td.SetDigitalPortProperties([]DigitalChannelDirections{
		{Channel: Dch0, Direction: DigitalDirectionRising},
	})

	dt := 1e-6 // 1 us time step
	reqSamples := uint32(1000)
	maxTime := 0.01 // 10 ms

	signalFunc := func(t float64, ch ChannelId) float64 {
		return 0
	}

	found, triggerTime := td.FindTriggerPoint(signalFunc, reqSamples, maxTime, dt)
	assert.True(t, found)
	assert.GreaterOrEqual(t, triggerTime, 0.0)

	// Verify that at triggerTime, D0 just underwent a rising transition (0 -> 1)
	p0Curr, _, _, _ := GetDemoDigitalGenValue(triggerTime)
	p0Prev, _, _, _ := GetDemoDigitalGenValue(triggerTime - dt)
	currD0 := p0Curr & 1
	prevD0 := p0Prev & 1
	assert.Equal(t, int16(1), currD0, "Current D0 should be 1 after rising edge")
	assert.Equal(t, int16(0), prevD0, "Previous D0 should be 0 before rising edge")
}

func TestDigitalTriggerFallingEdge(t *testing.T) {
	SetDigitalGenPortEnabled(true, true)
	s := &SimDesc{}
	_ = s.SetDemoDigitalGen(true, true, 1000, genericps.DigitalDemoGenDirectionUp, genericps.DigitalDemoGenEncodingBinary, genericps.DigitalDemoGenModeSynchronous, 0)

	td := NewTriggerDetector(false, 0, 0, TriggerNone, ChA)
	td.SetDigitalPortProperties([]DigitalChannelDirections{
		{Channel: Dch0, Direction: DigitalDirectionFalling},
	})

	dt := 1e-6
	reqSamples := uint32(1000)
	maxTime := 0.01

	signalFunc := func(t float64, ch ChannelId) float64 {
		return 0
	}

	found, triggerTime := td.FindTriggerPoint(signalFunc, reqSamples, maxTime, dt)
	assert.True(t, found)
	assert.GreaterOrEqual(t, triggerTime, 0.0)

	p0Curr, _, _, _ := GetDemoDigitalGenValue(triggerTime)
	p0Prev, _, _, _ := GetDemoDigitalGenValue(triggerTime - dt)
	currD0 := p0Curr & 1
	prevD0 := p0Prev & 1
	assert.Equal(t, int16(0), currD0, "Current D0 should be 0 after falling edge")
	assert.Equal(t, int16(1), prevD0, "Previous D0 should be 1 before falling edge")
}

func TestDigitalTriggerPatternMatching(t *testing.T) {
	SetDigitalGenPortEnabled(true, true)
	s := &SimDesc{}
	_ = s.SetDemoDigitalGen(true, true, 1000, genericps.DigitalDemoGenDirectionUp, genericps.DigitalDemoGenEncodingBinary, genericps.DigitalDemoGenModeSynchronous, 0)

	// Require D0 Rising and D1 High
	td := NewTriggerDetector(false, 0, 0, TriggerNone, ChA)
	td.SetDigitalPortProperties([]DigitalChannelDirections{
		{Channel: Dch0, Direction: DigitalDirectionRising},
		{Channel: Dch1, Direction: DigitalDirectionHigh},
	})

	dt := 1e-6
	reqSamples := uint32(1000)
	maxTime := 0.01

	signalFunc := func(t float64, ch ChannelId) float64 {
		return 0
	}

	found, triggerTime := td.FindTriggerPoint(signalFunc, reqSamples, maxTime, dt)
	assert.True(t, found)

	p0Curr, _, _, _ := GetDemoDigitalGenValue(triggerTime)
	p0Prev, _, _, _ := GetDemoDigitalGenValue(triggerTime - dt)
	currD0 := p0Curr & 1
	prevD0 := p0Prev & 1
	currD1 := (p0Curr >> 1) & 1

	assert.Equal(t, int16(1), currD0)
	assert.Equal(t, int16(0), prevD0)
	assert.Equal(t, int16(1), currD1, "D1 must be High")
}

func TestCombinedAnalogAndDigitalTriggerAND(t *testing.T) {
	SetDigitalGenPortEnabled(true, true)
	s := &SimDesc{}
	_ = s.SetDemoDigitalGen(true, true, 1000, genericps.DigitalDemoGenDirectionUp, genericps.DigitalDemoGenEncodingBinary, genericps.DigitalDemoGenModeSynchronous, 0)

	// Analog trigger on ChA rising through 1000
	td := NewTriggerDetector(true, 1000, 50, TriggerRising, ChA)
	td.SetChannelConditions([]TriggerConditions{
		{ChannelA: CondTrue, Digital: CondTrue},
	})
	// Digital trigger on D1 High
	td.SetDigitalPortProperties([]DigitalChannelDirections{
		{Channel: Dch1, Direction: DigitalDirectionHigh},
	})
	td.SetDigitalAnalogTriggerOperand(OperandAnd)

	dt := 1e-5
	reqSamples := uint32(1000)
	maxTime := 0.05

	// Sawtooth-like wave for easy testing
	signalFunc := func(t float64, ch ChannelId) float64 {
		if ch == ChA {
			return 2000.0 * (t*1000.0 - float64(int(t*1000.0)) - 0.5) * 2
		}
		return 0
	}

	found, triggerTime := td.FindTriggerPoint(signalFunc, reqSamples, maxTime, dt)
	assert.True(t, found)

	p0Curr, _, _, _ := GetDemoDigitalGenValue(triggerTime)
	currD1 := (p0Curr >> 1) & 1
	assert.Equal(t, int16(1), currD1, "D1 must be High when trigger fired in AND mode")
}

func TestDigitalGenerator_BlockMode(t *testing.T) {
	handle, err := openUnit("", 0)
	assert.NoError(t, err)

	s := &SimDesc{}
	err = s.SetDemoDigitalGen(true, true, 1000, genericps.DigitalDemoGenDirectionUp, genericps.DigitalDemoGenEncodingBinary, genericps.DigitalDemoGenModeSynchronous, 0)
	assert.NoError(t, err)

	err = simSetDigitalPort(handle, Port0, true, 0)
	assert.NoError(t, err)

	buf0 := make([]int16, 1000)
	err = simSetDataBuffer(handle, ChannelId(128), buf0, 0, RatioModeNone)
	assert.NoError(t, err)

	blockReadyCb := func(handle int16, status int, param any) {}
	_, err = simRunBlock(handle, 500, 500, 200, 0, 0, blockReadyCb, nil)
	assert.NoError(t, err)

	n, _, err := simGetValues(handle, 0, 1000, 1, RatioModeNone, 0)
	assert.NoError(t, err)
	assert.Greater(t, n, uint32(0))

	hasNonZero := false
	hasChanges := false
	for i := 1; i < int(n); i++ {
		if buf0[i] != 0 {
			hasNonZero = true
		}
		if buf0[i] != buf0[i-1] {
			hasChanges = true
		}
	}
	assert.True(t, hasNonZero, "Digital buffer should have non-zero samples")
	assert.True(t, hasChanges, "Digital buffer should have changing/toggling values")
}

func TestDigitalGenerator_StreamingMode(t *testing.T) {
	handle, err := openUnit("", 0)
	assert.NoError(t, err)

	s := &SimDesc{}
	err = s.SetDemoDigitalGen(true, true, 1000, genericps.DigitalDemoGenDirectionUp, genericps.DigitalDemoGenEncodingBinary, genericps.DigitalDemoGenModeSynchronous, 0)
	assert.NoError(t, err)

	err = simSetDigitalPort(handle, Port0, true, 0)
	assert.NoError(t, err)

	buf0 := make([]int16, 1000)
	err = simSetDataBuffer(handle, ChannelId(128), buf0, 0, RatioModeNone)
	assert.NoError(t, err)

	_, err = simRunStreaming(handle, 1, TuMs, 0, 1000, false, 1, RatioModeNone, 1000)
	assert.NoError(t, err)

	called := false
	var sampleCount int32
	readyCb := func(handle int16, noOfSamples int32, startIndex uint32, overflow int16, triggeredAt uint32, triggered int16, autoStop int16, p any) error {
		called = true
		sampleCount = noOfSamples
		return nil
	}

	// Sleep slightly to allow simulated time to advance
	time.Sleep(20 * time.Millisecond)
	err = simGetStreamingLatestValues(handle, readyCb, nil)
	assert.NoError(t, err)
	assert.True(t, called)
	assert.Greater(t, sampleCount, int32(0))

	hasNonZero := false
	for i := 0; i < int(sampleCount); i++ {
		if buf0[i] != 0 {
			hasNonZero = true
			break
		}
	}
	assert.True(t, hasNonZero, "Streaming digital buffer should have non-zero samples")
}

