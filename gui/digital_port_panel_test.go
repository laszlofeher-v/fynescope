package gui

import (
	"fmt"
	"testing"
	"time"

	"fynescope/control"
	"fynescope/disp7"
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

		fc, isFC := negCtrl.Obj.(*FocusCheck)
		assert.True(t, isFC, "Expected %s to be *FocusCheck", negCheckId)
		title, _ := scp.getHelpForWidget(fc)
		assert.Equal(t, "Negate Label", title)

		assert.True(t, labelOk, "Expected %s to be registered in controls", labelEntryId)
		assert.NotNil(t, labelCtrl.Obj, "Expected %s object to be non-nil", labelEntryId)
		assert.Equal(t, digPortTabIndex, labelCtrl.Tab)
	}
}

func TestDigitalPortPanel_LogicLevelDisp(t *testing.T) {
	setCh := make(chan *control.DigitalPortMsg, 10)
	scp := &ScpDesc{
		Settings: &settings.PsSettings{
			Digital: settings.DigitalSettings{
				Ports: [2]settings.DigitalPortSettings{
					{Enabled: true, Threshold: 1500},
					{Enabled: false, Threshold: -2000},
				},
			},
		},
		psControl: &control.PscDesc{
			SetDigitalPortCh: setCh,
		},
	}

	scp.buildDigitalPortContent(false)

	// Verify Port 0 Logic Level Disp
	controlsMtx.RLock()
	p0Ctrl, p0Ok := controls["digPort0LogicLevelDisp"]
	p1Ctrl, p1Ok := controls["digPort1LogicLevelDisp"]
	controlsMtx.RUnlock()

	assert.True(t, p0Ok, "Expected digPort0LogicLevelDisp to be registered in controls")
	assert.NotNil(t, p0Ctrl.Obj, "digPort0LogicLevelDisp object should not be nil")
	assert.Equal(t, digPortTabIndex, p0Ctrl.Tab)

	d7_0, isD7_0 := p0Ctrl.Obj.(*disp7.DigitArray)
	assert.True(t, isD7_0, "digPort0LogicLevelDisp must be *disp7.DigitArray")
	assert.Equal(t, 1500, d7_0.Value)

	title0, desc0 := scp.getHelpForWidget(d7_0)
	assert.Equal(t, "Port 0 Logic Level", title0)
	assert.Contains(t, desc0, "–32767 (–5 V) to 32767 (+5 V)")

	// Verify Port 1 Logic Level Disp
	assert.True(t, p1Ok, "Expected digPort1LogicLevelDisp to be registered in controls")
	assert.NotNil(t, p1Ctrl.Obj, "digPort1LogicLevelDisp object should not be nil")
	assert.Equal(t, digPortTabIndex, p1Ctrl.Tab)

	d7_1, isD7_1 := p1Ctrl.Obj.(*disp7.DigitArray)
	assert.True(t, isD7_1, "digPort1LogicLevelDisp must be *disp7.DigitArray")
	assert.Equal(t, -2000, d7_1.Value)

	title1, desc1 := scp.getHelpForWidget(d7_1)
	assert.Equal(t, "Port 1 Logic Level", title1)
	assert.Contains(t, desc1, "–32767 (–5 V) to 32767 (+5 V)")

	// Drain any messages sent during initialization
	for len(setCh) > 0 {
		<-setCh
	}

	// Test changing Port 0 logic level via SetValue
	d7_0.SetValue(3000)
	assert.Equal(t, int16(3000), scp.Settings.Digital.Ports[0].Threshold)

	select {
	case msg := <-setCh:
		assert.Equal(t, genericps.Port0, msg.Port)
		assert.Equal(t, int16(3000), msg.Settings.Threshold)
		assert.True(t, msg.Settings.Enabled)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Expected SetDigitalPortCh message for Port 0")
	}

	// Test changing Port 1 logic level to negative value
	d7_1.SetValue(-10000)
	assert.Equal(t, int16(-10000), scp.Settings.Digital.Ports[1].Threshold)

	select {
	case msg := <-setCh:
		assert.Equal(t, genericps.Port1, msg.Port)
		assert.Equal(t, int16(-10000), msg.Settings.Threshold)
		assert.False(t, msg.Settings.Enabled)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Expected SetDigitalPortCh message for Port 1")
	}
}

