package gui

import (
	"fmt"
	"testing"
	"time"

	"fynescope/control"
	"fynescope/genericps"
	"fynescope/settings"
	"github.com/stretchr/testify/assert"
)

func TestUpdateDigitalTrigger(t *testing.T) {
	genericps.DigitalDontCare = 0
	genericps.DigitalDirectionLow = 1
	genericps.DigitalDirectionHigh = 2
	genericps.DigitalDirectionRising = 3
	genericps.DigitalDirectionFalling = 4
	genericps.DigitalDirectionRisingOrFalling = 5

	genericps.OperandOr = 1
	genericps.OperandAnd = 2

	scp := &ScpDesc{
		Settings: &settings.PsSettings{
			Digital: settings.DigitalSettings{
				Ports: [2]settings.DigitalPortSettings{
					{Enabled: true},
					{Enabled: true},
				},
				ChannelsEnabled: [16]bool{
					0: true,
					1: true,
				},
				Trigger: settings.DigitalTriggerSettings{
					Enabled: true,
					Logic:   "AND",
					Operand: genericps.OperandAnd,
					Directions: [16]genericps.DigitalDirection{
						0: genericps.DigitalDirectionRising,
						1: genericps.DigitalDirectionHigh,
					},
				},
			},
		},
		psControl: &control.PscDesc{
			SetTriggerCh: make(chan *control.TriggerDescMsg, 1),
		},
	}

	scp.updateDigitalTrigger()

	select {
	case msg := <-scp.psControl.SetTriggerCh:
		assert.NotNil(t, msg)
		assert.True(t, msg.DigitalTriggerEnabled)
		assert.Equal(t, genericps.OperandAnd, msg.DigitalAnalogOperand)
		assert.Len(t, msg.DigitalDirections, 2)
		assert.Equal(t, genericps.DigitalChannel(0), msg.DigitalDirections[0].Channel)
		assert.Equal(t, genericps.DigitalDirectionRising, msg.DigitalDirections[0].Direction)
		assert.Equal(t, genericps.DigitalChannel(1), msg.DigitalDirections[1].Channel)
		assert.Equal(t, genericps.DigitalDirectionHigh, msg.DigitalDirections[1].Direction)
		msg.Done <- struct{}{}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected digital trigger message to be sent to channel")
	}

	// Test with disabled channel: D1 disabled -> only D0 should be in DigitalDirections
	scp.Settings.Digital.ChannelsEnabled[1] = false
	scp.updateDigitalTrigger()

	select {
	case msg := <-scp.psControl.SetTriggerCh:
		assert.NotNil(t, msg)
		assert.Len(t, msg.DigitalDirections, 1)
		assert.Equal(t, genericps.DigitalChannel(0), msg.DigitalDirections[0].Channel)
		msg.Done <- struct{}{}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected digital trigger message to be sent to channel")
	}

	// Test with disabled port: Port 0 disabled -> 0 directions
	scp.Settings.Digital.Ports[0].Enabled = false
	scp.updateDigitalTrigger()

	select {
	case msg := <-scp.psControl.SetTriggerCh:
		assert.NotNil(t, msg)
		assert.Len(t, msg.DigitalDirections, 0)
		msg.Done <- struct{}{}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected digital trigger message to be sent to channel")
	}
}

func TestDigitalPortPanel_DnLabelsAddToTest(t *testing.T) {
	scp := &ScpDesc{
		Settings: &settings.PsSettings{
			Digital: settings.DigitalSettings{
				Ports: [2]settings.DigitalPortSettings{
					{Enabled: true},
					{Enabled: true},
				},
			},
		},
	}

	scp.buildDigitalPortContent(false)

	for i := 0; i < 16; i++ {
		labelId := "digPortDnLabel_" + string(rune('0'+i/10)) + string(rune('0'+i%10))
		if i < 10 {
			labelId = "digPortDnLabel_" + string(rune('0'+i))
		}
		controlsMtx.RLock()
		ctrl, ok := controls[labelId]
		controlsMtx.RUnlock()
		assert.True(t, ok, "Expected %s to be registered in controls", labelId)
		assert.NotNil(t, ctrl.Obj, "Expected %s object to be non-nil", labelId)
		assert.Equal(t, digPortTabIndex, ctrl.Tab)

		// Test tapping toggles negation
		dnLabel, isDnLabel := ctrl.Obj.(*tappableDnLabel)
		assert.True(t, isDnLabel, "Expected %s to be *tappableDnLabel", labelId)
		initialNeg := scp.Settings.Digital.ChannelNegated[i]
		dnLabel.Tapped(nil)
		assert.Equal(t, !initialNeg, scp.Settings.Digital.ChannelNegated[i])
		assert.Equal(t, !initialNeg, dnLabel.negated)
	}
}

func TestDigitalPortPanel_NegCheckAndLabelAddToTest(t *testing.T) {
	scp := &ScpDesc{
		Settings: &settings.PsSettings{
			Digital: settings.DigitalSettings{
				Ports: [2]settings.DigitalPortSettings{
					{Enabled: true},
					{Enabled: true},
				},
			},
		},
	}

	scp.buildDigitalPortContent(false)

	for i := 0; i < 16; i++ {
		negCheckId := fmt.Sprintf("digPortNegCheck_%d", i)
		labelEntryId := fmt.Sprintf("digPortLabelEntry_%d", i)

		controlsMtx.RLock()
		negCtrl, negOk := controls[negCheckId]
		labelCtrl, labelOk := controls[labelEntryId]
		controlsMtx.RUnlock()

		assert.True(t, negOk, "Expected %s to be registered in controls", negCheckId)
		assert.NotNil(t, negCtrl.Obj, "Expected %s object to be non-nil", negCheckId)
		assert.Equal(t, digPortTabIndex, negCtrl.Tab)

		assert.True(t, labelOk, "Expected %s to be registered in controls", labelEntryId)
		assert.NotNil(t, labelCtrl.Obj, "Expected %s object to be non-nil", labelEntryId)
		assert.Equal(t, digPortTabIndex, labelCtrl.Tab)
	}
}
