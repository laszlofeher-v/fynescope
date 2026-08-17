package control

import (
	"fynescope/genericps"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func mockState(ps *PscDesc) state {
	return nil
}

func TestFunctionName(t *testing.T) {
	name := functionName(mockState)
	assert.Contains(t, name, "mockState")
}

func TestIdle_Shutdown(t *testing.T) {
	psControl := &PscDesc{
		shutdownCh:     make(chan struct{}, 1),
		restartChannel: make(chan struct{}, 1),
		stateChannel:   make(chan state, 1),
		stopChannel:    make(chan struct{}, 1),
	}

	psControl.shutdownCh <- struct{}{}
	nextState := idle(psControl)
	assert.Nil(t, nextState)
}

func TestIdle_StateTransition(t *testing.T) {
	psControl := &PscDesc{
		shutdownCh:     make(chan struct{}, 1),
		restartChannel: make(chan struct{}, 1),
		stateChannel:   make(chan state, 1),
		stopChannel:    make(chan struct{}, 1),
	}

	psControl.stateChannel <- mockState
	nextState := idle(psControl)
	
	// Check if the returned state is the same as we sent (comparing addresses or function name)
	assert.Equal(t, functionName(mockState), functionName(nextState))
}

func TestIdle_Restart(t *testing.T) {
	psControl := &PscDesc{
		shutdownCh:     make(chan struct{}, 1),
		restartChannel: make(chan struct{}, 1),
		stateChannel:   make(chan state, 1),
		stopChannel:    make(chan struct{}, 1),
	}

	// Send restart, but it should ignore it and block until another event.
	psControl.restartChannel <- struct{}{}
	psControl.stateChannel <- mockState

	nextState := idle(psControl)
	assert.Equal(t, functionName(mockState), functionName(nextState))
}

func TestStateMachine(t *testing.T) {
	psControl := &PscDesc{
		shutdownCh:     make(chan struct{}, 1),
		restartChannel: make(chan struct{}, 1),
		stateChannel:   make(chan state, 1),
		stopChannel:    make(chan struct{}, 1),
	}

	// Make idle return nil by sending shutdown immediately
	psControl.shutdownCh <- struct{}{}
	
	done := make(chan struct{})
	go func() {
		psControl.stateMachine()
		close(done)
	}()
	
	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("stateMachine did not return")
	}
}

func TestQuit(t *testing.T) {
	con := genericps.NewConnection()
	psControl := &PscDesc{
		Con: con,
	}

	err := psControl.quit()
	assert.NoError(t, err)

	// Since quit runs asynchronously, wait a moment
	time.Sleep(50 * time.Millisecond)
}

func TestRequestRestart(t *testing.T) {
	psControl := &PscDesc{
		restartChannel: make(chan struct{}, 1),
	}

	psControl.RequestRestart()
	
	// Verify channel has one item
	assert.Len(t, psControl.restartChannel, 1)

	// A second call should not block, despite channel capacity being 1
	psControl.RequestRestart()
	assert.Len(t, psControl.restartChannel, 1)
}
