package control

import (
	_ "fynescope/demo"
	"fynescope/genericps"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestIdle_DrainsStaleStopBeforeTransition(t *testing.T) {
	for i := 0; i < 20; i++ {
		psControl := &PscDesc{
			shutdownCh:     make(chan struct{}, 1),
			restartChannel: make(chan struct{}, 1),
			stateChannel:   make(chan state, 1),
			stopChannel:    make(chan struct{}, 1),
		}
		psControl.stopChannel <- struct{}{}
		psControl.restartChannel <- struct{}{}
		psControl.stateChannel <- mockState

		nextState := idle(psControl)
		assert.Equal(t, functionName(mockState), functionName(nextState))
		assert.Empty(t, psControl.stopChannel, "stale stop request leaked to next state on iter %d", i)
		assert.Empty(t, psControl.restartChannel, "stale restart request leaked to next state on iter %d", i)
	}
}

func TestQuit_CompletesStopBeforeReturning(t *testing.T) {
	con := genericps.NewConnection()
	handle, err := genericps.OpenDemo(con, genericps.DemoId)
	require.NoError(t, err)
	con.Handle = handle
	defer con.CloseUnit()

	psControl := &PscDesc{Con: con}

	_, err = con.RunStreaming(1000, genericps.TuNs, 0, 100, false, 1, genericps.RatioModeNone, 100)
	require.NoError(t, err)

	require.NoError(t, psControl.quit())

	// Streaming must already be stopped before quit() returns so a subsequent
	// state's RunStreaming/RunBlock cannot be cancelled by an async Stop.
	err = con.GetStreamingLatestValues(func(_ int16, _ int32, _ uint32, _ int16, _ uint32, _, _ int16, _ any) error {
		return nil
	}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "streaming not running")
}
