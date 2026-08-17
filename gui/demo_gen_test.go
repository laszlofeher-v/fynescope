package gui

import (
	"fynescope/control"
	"fynescope/genericps"
	"fynescope/settings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestApplyDemoGenSettings(t *testing.T) {
	scp := &ScpDesc{
		psControl: &control.PscDesc{
			SetDemoGenCh: make(chan *control.GeneratorDescMsg, 1),
		},
	}

	// Test 1: No Sweep
	genSettings := &settings.GeneratorSettings{
		On:            true,
		WaveType:      genericps.Sine,
		Frequency:     1000,
		Amplitude:     500,
		OffsetVoltage: 100,
		Sweep:         genericps.NoSweep,
	}

	scp.applyDemoGenSettings(genericps.ChA, genSettings)

	select {
	case msg := <-scp.psControl.SetDemoGenCh:
		assert.Equal(t, genericps.ChA, msg.Channel)
		assert.True(t, msg.On)
		assert.Equal(t, genericps.Sine, msg.WaveType)
		assert.Equal(t, float64(1000), msg.StartFrequency)
		assert.Equal(t, float64(1000), msg.StopFrequency)
		assert.Equal(t, genericps.SweepDown, msg.SweepType) // NoSweep mapped to SweepDown internally
		assert.Equal(t, uint32(1000), msg.PkToPK) // Amplitude * 2
		assert.Equal(t, int32(100), msg.OffsetVoltage)
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for demo gen message")
	}

	// Test 2: Sweep
	genSettings.Sweep = genericps.SweepUp
	genSettings.StartFrequency = 100
	genSettings.StopFrequency = 2000
	genSettings.Increment = 10
	genSettings.Dwelltime = 0.5

	scp.applyDemoGenSettings(genericps.ChB, genSettings)

	select {
	case msg := <-scp.psControl.SetDemoGenCh:
		assert.Equal(t, genericps.ChB, msg.Channel)
		assert.Equal(t, genericps.SweepUp, msg.SweepType)
		assert.Equal(t, float64(100), msg.StartFrequency)
		assert.Equal(t, float64(2000), msg.StopFrequency)
		assert.Equal(t, float64(10), msg.Increment)
		assert.Equal(t, float64(0.5), msg.DwellTime)
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for demo gen message")
	}

	// Test 3: Off
	genSettings.On = false
	scp.applyDemoGenSettings(genericps.ChA, genSettings)

	select {
	case msg := <-scp.psControl.SetDemoGenCh:
		assert.False(t, msg.On)
		assert.Equal(t, float64(0), msg.DwellTime)
		assert.Equal(t, int32(0), msg.OffsetVoltage)
		assert.Equal(t, uint32(0), msg.PkToPK)
		assert.Equal(t, genericps.DcVoltage, msg.WaveType)
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for demo gen message")
	}
}
