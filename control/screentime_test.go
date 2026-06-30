package control

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func setupPscDescForScreenTimeTest() *PscDesc {
	return &PscDesc{
		restartChannel: make(chan struct{}, 1),
		SetMaxScreenTimeCh: make(chan float64),
		getMaxScreenTimeCh: make(chan *getMaxScreenTimeMsg),
		getMaxScreenTime: getMaxScreenTimeMsg{
			newSetting: make(chan bool),
		},
	}
}

func TestScreenTimeMonitor_InitialState(t *testing.T) {
	psControl := setupPscDescForScreenTimeTest()
	go psControl.screenTimeMonitor()

	psControl.getMaxScreenTimeCh <- &psControl.getMaxScreenTime
	select {
	case newData := <-psControl.getMaxScreenTime.newSetting:
		assert.False(t, newData)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for response")
	}
}

func TestScreenTimeMonitor_SetNewScreenTime(t *testing.T) {
	psControl := setupPscDescForScreenTimeTest()
	go psControl.screenTimeMonitor()

	psControl.SetMaxScreenTimeCh <- 2.5
	time.Sleep(10 * time.Millisecond)

	select {
	case <-psControl.restartChannel:
		// Success
	default:
		t.Fatal("Expected restart to be requested when setting new max screen time")
	}

	psControl.getMaxScreenTimeCh <- &psControl.getMaxScreenTime
	select {
	case newData := <-psControl.getMaxScreenTime.newSetting:
		assert.True(t, newData)
		assert.Equal(t, 2.5, psControl.getMaxScreenTime.maxScreenTime)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for response")
	}

	psControl.getMaxScreenTimeCh <- &psControl.getMaxScreenTime
	select {
	case newData := <-psControl.getMaxScreenTime.newSetting:
		assert.False(t, newData)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for response")
	}
}

func TestScreenTimeMonitor_SetSameScreenTime(t *testing.T) {
	psControl := setupPscDescForScreenTimeTest()
	go psControl.screenTimeMonitor()

	psControl.SetMaxScreenTimeCh <- 3.0
	time.Sleep(10 * time.Millisecond)

	select {
	case <-psControl.restartChannel:
	default:
	}
	psControl.getMaxScreenTimeCh <- &psControl.getMaxScreenTime
	<-psControl.getMaxScreenTime.newSetting

	psControl.SetMaxScreenTimeCh <- 3.0
	time.Sleep(10 * time.Millisecond)

	select {
	case <-psControl.restartChannel:
		t.Fatal("Did not expect restart when setting identical max screen time")
	default:
		// Success
	}

	psControl.getMaxScreenTimeCh <- &psControl.getMaxScreenTime
	select {
	case newData := <-psControl.getMaxScreenTime.newSetting:
		assert.False(t, newData)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for response")
	}
}
