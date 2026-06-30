package control

import (
	"fynescope/genericps"
	"fynescope/settings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func init() {
	genericps.Range_5v = 8
}

func setupPscDescForChannelsTest() *PscDesc {
	return &PscDesc{
		restartChannel: make(chan struct{}, 1),
		SetChannelCh:   make(chan *settings.ChSettings),
		getChannelCh:   make(chan *getChannelMsg),
		getChannel: getChannelMsg{
			newSettings: make(chan bool),
		},
		getNumOfEnabledCh: make(chan *getNumOfEnabledChMsg),
		getNumOfEnabled: getNumOfEnabledChMsg{
			n: make(chan int),
		},
	}
}

func TestChannelStateMachine_InitialState(t *testing.T) {
	psControl := setupPscDescForChannelsTest()
	go psControl.channelStateMachine(4)

	// Unchanged initial get
	psControl.getChannelCh <- &psControl.getChannel
	select {
	case newSettings := <-psControl.getChannel.newSettings:
		assert.False(t, newSettings)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for response")
	}

	// Channel count enabled initially should be 0
	msg := &getNumOfEnabledChMsg{
		n: make(chan int),
	}
	psControl.getNumOfEnabledCh <- msg
	select {
	case n := <-msg.n:
		assert.Equal(t, 0, n)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for response")
	}
}

func TestChannelStateMachine_SetChannel(t *testing.T) {
	psControl := setupPscDescForChannelsTest()
	go psControl.channelStateMachine(4)

	chSetting := &settings.ChSettings{
		ID:         genericps.ChannelId(0), // ChA
		Enabled:    true,
		CoupleType: genericps.Dc,
		VRange:     genericps.Range_5v,
		Offset:     1.5,
	}

	psControl.SetChannelCh <- chSetting
	time.Sleep(10 * time.Millisecond)

	// Check if restart was triggered
	select {
	case <-psControl.restartChannel:
		// Success
	default:
		t.Fatal("Expected restart to be requested when setting channel")
	}

	// Verify channel enabled count is now 1
	msg := &getNumOfEnabledChMsg{
		n: make(chan int),
	}
	psControl.getNumOfEnabledCh <- msg
	select {
	case n := <-msg.n:
		assert.Equal(t, 1, n)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout")
	}

	// Verify we can get the changed channel settings
	psControl.getChannelCh <- &psControl.getChannel
	select {
	case newSettings := <-psControl.getChannel.newSettings:
		assert.True(t, newSettings)
		assert.NotNil(t, psControl.getChannel.channelSettings)
		assert.Equal(t, genericps.ChannelId(0), psControl.getChannel.channelSettings.Channel)
		assert.True(t, psControl.getChannel.channelSettings.Enabled)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout")
	}

	// Should be unchanged now
	psControl.getChannelCh <- &psControl.getChannel
	select {
	case newSettings := <-psControl.getChannel.newSettings:
		assert.False(t, newSettings)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout")
	}
}

func TestChannelStateMachine_SetChannelX10Scaling(t *testing.T) {
	psControl := setupPscDescForChannelsTest()
	go psControl.channelStateMachine(4)

	chSetting := &settings.ChSettings{
		ID:         genericps.ChannelId(1), // ChB
		Enabled:    true,
		X10:        true,
		CoupleType: genericps.Ac,
		VRange:     genericps.Range_5v, // 5V
		Offset:     10.0,
	}

	psControl.SetChannelCh <- chSetting
	time.Sleep(10 * time.Millisecond)

	// Get the settings
	psControl.getChannelCh <- &psControl.getChannel
	select {
	case newSettings := <-psControl.getChannel.newSettings:
		assert.True(t, newSettings)
		// With X10 enabled:
		// VoltageRange = VRange - 3
		// AnalogOffset = Offset / 10
		assert.Equal(t, genericps.RangeEnum(chSetting.VRange-3), psControl.getChannel.channelSettings.VoltageRange)
		assert.Equal(t, float32(1.0), psControl.getChannel.channelSettings.AnalogOffset)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout")
	}
}
