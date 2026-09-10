package control

import (
	"testing"
	"time"

	_ "fynescope/demo"
	"fynescope/genericps"
	"fynescope/settings"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDigitalPortMonitor(t *testing.T) {
	psControl := &PscDesc{
		shutdownCh:       make(chan struct{}),
		restartChannel:   make(chan struct{}, 10),
		SetDigitalPortCh: make(chan *DigitalPortMsg, 10),
		getDigitalPortCh: make(chan *getDigitalPortMsg, 10),
	}

	go psControl.digitalPortMonitor()

	// 1. Initial getDigitalPort check (should report no new settings)
	getMsg := &getDigitalPortMsg{
		newSettings: make(chan bool, 1),
	}
	psControl.getDigitalPortCh <- getMsg
	select {
	case changed := <-getMsg.newSettings:
		assert.False(t, changed)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for getDigitalPort")
	}

	// 2. Send invalid port msg (should be ignored)
	psControl.SetDigitalPortCh <- &DigitalPortMsg{
		Port: genericps.DigitalPort(99),
		Settings: settings.DigitalPortSettings{
			Enabled:   true,
			Threshold: 1500,
		},
	}
	// Give monitor time to process
	time.Sleep(10 * time.Millisecond)

	// 3. Send valid Port0 settings
	setMsg := &DigitalPortMsg{
		Port: genericps.Port0,
		Settings: settings.DigitalPortSettings{
			Enabled:   true,
			Threshold: 1650,
		},
	}
	psControl.SetDigitalPortCh <- setMsg
	time.Sleep(10 * time.Millisecond)

	assert.True(t, psControl.digitalPortsEnabled[0].Load())

	// 4. Send identical settings (should stay in unchanged state)
	psControl.SetDigitalPortCh <- setMsg
	time.Sleep(10 * time.Millisecond)

	// 5. Query digital port settings
	getMsg = &getDigitalPortMsg{
		newSettings: make(chan bool, 1),
	}
	psControl.getDigitalPortCh <- getMsg
	select {
	case changed := <-getMsg.newSettings:
		assert.True(t, changed)
		require.NotNil(t, getMsg.portSettings)
		assert.Equal(t, genericps.Port0, getMsg.portSettings.Port)
		assert.Equal(t, int16(1650), getMsg.portSettings.Settings.Threshold)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for getDigitalPort after change")
	}

	// 6. Shutdown monitor
	close(psControl.shutdownCh)
}

func TestSetDigitalPort_WithMonitor(t *testing.T) {
	con := genericps.NewConnection()
	handle, err := genericps.OpenDemo(con, genericps.DemoId)
	require.NoError(t, err)
	con.Handle = handle
	defer con.CloseUnit()

	psControl := &PscDesc{
		Con:              con,
		shutdownCh:       make(chan struct{}),
		restartChannel:   make(chan struct{}, 10),
		SetDigitalPortCh: make(chan *DigitalPortMsg, 10),
		getDigitalPortCh: make(chan *getDigitalPortMsg, 10),
		getDigitalPort: getDigitalPortMsg{
			newSettings: make(chan bool, 10),
		},
	}
	go psControl.digitalPortMonitor()
	defer close(psControl.shutdownCh)

	// Send setting for Port0
	psControl.SetDigitalPortCh <- &DigitalPortMsg{
		Port: genericps.Port0,
		Settings: settings.DigitalPortSettings{
			Enabled:   true,
			Threshold: 1500,
		},
	}
	time.Sleep(10 * time.Millisecond)

	done := make(chan error, 1)
	go func() {
		done <- psControl.setDigitalPort()
	}()

	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(1 * time.Second):
		t.Fatal("setDigitalPort timed out")
	}
}

func TestSetDigitalPort_NilChannel(t *testing.T) {
	psControl := &PscDesc{
		getDigitalPortCh: nil,
	}
	err := psControl.setDigitalPort()
	assert.NoError(t, err)
}
