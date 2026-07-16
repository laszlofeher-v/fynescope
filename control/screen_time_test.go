package control

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupPscDescForScreenTimeTest() *PscDesc {
	return &PscDesc{
		restartChannel: make(chan struct{}, 1),
	}
}

func TestSetMaxScreenTime_SetNewScreenTime(t *testing.T) {
	psControl := setupPscDescForScreenTimeTest()

	psControl.SetMaxScreenTime(2.5)

	select {
	case <-psControl.restartChannel:
		// Success
	default:
		t.Fatal("Expected restart to be requested when setting new max screen time")
	}

	assert.Equal(t, 2.5, psControl.maxScreenTime)
}

func TestSetMaxScreenTime_SetSameScreenTime(t *testing.T) {
	psControl := setupPscDescForScreenTimeTest()

	psControl.SetMaxScreenTime(3.0)

	select {
	case <-psControl.restartChannel:
		// Expected for the first time
	default:
		t.Fatal("Expected restart to be requested when setting new max screen time")
	}

	psControl.SetMaxScreenTime(3.0)

	select {
	case <-psControl.restartChannel:
		t.Fatal("Did not expect restart when setting identical max screen time")
	default:
		// Success
	}

	assert.Equal(t, 3.0, psControl.maxScreenTime)
}
