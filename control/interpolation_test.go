package control

import (
	"fynescope/settings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func setupPscDescForInterpolationTest() *PscDesc {
	return &PscDesc{
		restartChannel: make(chan struct{}, 1),
		SetInterpolationModeCh: make(chan settings.InterpolationType),
		getInterpolationModeCh: make(chan *getInterpolationModeMsg),
		getInterpolationMode: getInterpolationModeMsg{
			newSetting: make(chan bool),
		},
	}
}

func TestInterpolationMonitor_InitialState(t *testing.T) {
	psControl := setupPscDescForInterpolationTest()
	go psControl.interpolationMonitor()

	psControl.getInterpolationModeCh <- &psControl.getInterpolationMode
	select {
	case newData := <-psControl.getInterpolationMode.newSetting:
		assert.False(t, newData)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for response")
	}
}

func TestInterpolationMonitor_TransitionToChangedAndUnchanged(t *testing.T) {
	psControl := setupPscDescForInterpolationTest()
	go psControl.interpolationMonitor()

	psControl.SetInterpolationModeCh <- settings.Sinc
	time.Sleep(10 * time.Millisecond)

	select {
	case <-psControl.restartChannel:
		// Success
	default:
		t.Fatal("Expected restart to be requested when switching to Sinc")
	}

	psControl.getInterpolationModeCh <- &psControl.getInterpolationMode
	select {
	case newData := <-psControl.getInterpolationMode.newSetting:
		assert.True(t, newData)
		assert.Equal(t, settings.Sinc, psControl.getInterpolationMode.ipMode)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for response")
	}

	psControl.getInterpolationModeCh <- &psControl.getInterpolationMode
	select {
	case newData := <-psControl.getInterpolationMode.newSetting:
		assert.False(t, newData)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for response")
	}
}

func TestInterpolationMonitor_NoSincTransitionIsUnchanged(t *testing.T) {
	psControl := setupPscDescForInterpolationTest()
	go psControl.interpolationMonitor()

	psControl.SetInterpolationModeCh <- settings.Linear
	time.Sleep(10 * time.Millisecond)

	select {
	case <-psControl.restartChannel:
		t.Fatal("Did not expect restart for non-Sinc transition")
	default:
		// Success
	}

	psControl.getInterpolationModeCh <- &psControl.getInterpolationMode
	select {
	case newData := <-psControl.getInterpolationMode.newSetting:
		assert.False(t, newData)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for response")
	}
}
