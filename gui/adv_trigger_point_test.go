package gui

import (
	"fynescope/control"
	"fynescope/genericps"
	"fynescope/settings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSetTriggerUpperHysteresis_ValidSource(t *testing.T) {
	genericps.InputRanges = []int32{5000} // Mock InputRanges to prevent panic
	scp := &ScpDesc{
		triggerSource: 0,
		Settings: &settings.PsSettings{
			Channels: []settings.ChSettings{
				{VRange: genericps.Range_5v}, // valid channel 0
			},
		},
		psControl: &control.PscDesc{
			SetTriggerCh: make(chan *control.TriggerDescMsg, 1),
		},
	}
	
	// Force mvToUAdc to use mock/generic genericps behaviors if we were doing a full test,
	// but here mvToUAdc handles the calculation.

	scp.SetTriggerUpperHysteresis(500) // 500 mV

	assert.Equal(t, int32(500), scp.triggerSettingMsg.UpperHysteresis)

	// Consume the channel message to verify it was sent
	select {
	case msg := <-scp.psControl.SetTriggerCh:
		assert.NotNil(t, msg)
		msg.Done <- struct{}{} // Acknowledge
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected trigger message to be sent to channel")
	}
}

func TestSetTriggerUpperHysteresis_InvalidSource(t *testing.T) {
	genericps.InputRanges = []int32{5000}
	scp := &ScpDesc{
		triggerSource: -1, // invalid channel
		Settings: &settings.PsSettings{
			Channels: []settings.ChSettings{},
		},
		psControl: &control.PscDesc{
			SetTriggerCh: make(chan *control.TriggerDescMsg, 1),
		},
	}
	
	// Should not panic (tests the bounds check fix)
	scp.SetTriggerUpperHysteresis(200)

	assert.Equal(t, int32(200), scp.triggerSettingMsg.UpperHysteresis)
	
	// Consume the channel message to verify it was sent
	select {
	case msg := <-scp.psControl.SetTriggerCh:
		assert.NotNil(t, msg)
		msg.Done <- struct{}{}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected trigger message to be sent to channel")
	}
}

func TestSetTriggerLowerHysteresis(t *testing.T) {
	genericps.InputRanges = []int32{5000}
	scp := &ScpDesc{
		triggerSource: 0,
		Settings: &settings.PsSettings{
			Channels: []settings.ChSettings{
				{VRange: genericps.Range_1v},
			},
		},
		psControl: &control.PscDesc{
			SetTriggerCh: make(chan *control.TriggerDescMsg, 1),
		},
	}
	
	scp.SetTriggerLowerHysteresis(300)

	assert.Equal(t, int32(300), scp.triggerSettingMsg.LowerHysteresis)

	// Consume the channel message
	select {
	case msg := <-scp.psControl.SetTriggerCh:
		assert.NotNil(t, msg)
		msg.Done <- struct{}{}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected trigger message to be sent to channel")
	}
}
