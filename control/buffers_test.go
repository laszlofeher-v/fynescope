package control

import (
	"sync/atomic"
	"testing"

	_ "fynescope/demo"
	"fynescope/genericps"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupDemoControl(t *testing.T) (*PscDesc, *genericps.Connection) {
	con := genericps.NewConnection()
	handle, err := genericps.OpenDemo(con, genericps.DemoId)
	require.NoError(t, err)
	con.Handle = handle

	psControl := &PscDesc{
		Con:                 con,
		receiveBuffer:       make([][]int16, 2),
		receiveBufferMin:    make([][]int16, 2),
		displayBuffer:       make([][]float32, 2),
		downSampleRatioMode: genericps.RatioModeNone,
		digitalPortsEnabled: [2]atomic.Bool{},
		chEnabled:           make([]atomic.Bool, 2),
	}
	psControl.chEnabled[0].Store(true)
	_ = psControl.setEtsBuffer(10000, 0)
	return psControl, con
}

func TestPscDesc_checkOverflow(t *testing.T) {
	psControl := &PscDesc{}

	// Mock 4 channels
	psControl.chEnabled = make([]atomic.Bool, 4)

	var lastStatusMsg string
	var lastStatusErr ScopeError
	var statusCalls int

	psControl.DisplayStatus = func(s string, errorType ScopeError) {
		lastStatusMsg = s
		lastStatusErr = errorType
		statusCalls++
	}

	// Test 1: No overflow
	psControl.checkOverflow(0)
	assert.Equal(t, 0, statusCalls)

	// Test 2: Overflow on channel A, but channel A is disabled
	psControl.chEnabled[0].Store(false)
	psControl.checkOverflow(1) // Bit 0 is ChA
	assert.Equal(t, 0, statusCalls)

	// Test 3: Overflow on channel A, and channel A is enabled
	psControl.chEnabled[0].Store(true)
	psControl.checkOverflow(1)
	assert.Equal(t, 1, statusCalls)
	assert.Equal(t, "Overflow error on channel:A ", lastStatusMsg)
	assert.Equal(t, Warning, lastStatusErr)

	// Test 4: Overflow on multiple channels (A and C)
	psControl.chEnabled[2].Store(true) // Enable ChC
	statusCalls = 0
	psControl.checkOverflow(5) // Bit 0 (ChA) and Bit 2 (ChC)
	assert.Equal(t, 1, statusCalls)
	assert.Equal(t, "Overflow error on channels:A C ", lastStatusMsg)
}

func TestPscDesc_SetBuffers(t *testing.T) {
	psControl, con := setupDemoControl(t)
	defer con.CloseUnit()

	psControl.digitalPortsEnabled[0].Store(true)

	// Call setBuffers with 100 samples
	err := psControl.setBuffers(100, 0)
	require.NoError(t, err)

	assert.Len(t, psControl.receiveBuffer[0], 100)
	assert.Len(t, psControl.displayBuffer[0], 100)
	assert.Len(t, psControl.digitalReceiveBuffer[0], 100)

	// Re-call setBuffers with larger sampleCount (150)
	err = psControl.setBuffers(150, 0)
	require.NoError(t, err)
	assert.Len(t, psControl.receiveBuffer[0], 150)

	// Test with RatioModeAggregate
	psControl.downSampleRatioMode = genericps.RatioModeAggregate
	err = psControl.setBuffers(100, 0)
	require.NoError(t, err)
	assert.Len(t, psControl.receiveBufferMin[0], 100)
	assert.Len(t, psControl.digitalReceiveBufferMin[0], 100)
}

func TestPscDesc_SetEtsBuffer(t *testing.T) {
	psControl, con := setupDemoControl(t)
	defer con.CloseUnit()

	err := psControl.setEtsBuffer(500, 0)
	require.NoError(t, err)
	assert.Len(t, psControl.EtsInBuffer, 500)

	// Call again with smaller count
	err = psControl.setEtsBuffer(300, 0)
	require.NoError(t, err)
	assert.Len(t, psControl.EtsInBuffer, 300)

	// Call again with larger count to leave large capacity for global demo generator
	err = psControl.setEtsBuffer(100000, 0)
	require.NoError(t, err)
	assert.Len(t, psControl.EtsInBuffer, 100000)
}

func TestPscDesc_MemorySegments(t *testing.T) {
	psControl, con := setupDemoControl(t)
	defer con.CloseUnit()

	sc, err := psControl.memorySegments(4)
	require.NoError(t, err)
	assert.Greater(t, sc, uint64(0))
}

func TestPscDesc_GetData(t *testing.T) {
	psControl, con := setupDemoControl(t)
	defer con.CloseUnit()

	refreshed := false
	refreshedEts := false

	psControl.EtsInBuffer = make([]int64, 100)
	psControl.NPre = 10
	psControl.SamplingTimeInterval = 1e-6
	psControl.RefreshCallback = func(rec, min [][]int16, dig [][]int16, trigOffset int64, roundErr float64, dt float64) {
		refreshed = true
	}
	psControl.RefreshEtsCallback = func(rec [][]int16, etsTime []int64, roundErr float64, dt float64) {
		refreshedEts = true
	}
	psControl.DisplayStatus = func(s string, errType ScopeError) {}

	_ = con.SetChannel(genericps.ChA, true, genericps.Dc, 7, 0)
	_, _, _ = con.GetTimebase2(1, 100, 0, 0)
	dummyReady := genericps.BlockReady(func(handle int16, status int, param any) {})
	_, _ = con.RunBlock(0, 100, 1, 0, 0, dummyReady, nil)

	// First set buffers to allocate driver buffers
	err := psControl.setBuffers(100, 0)
	require.NoError(t, err)

	// 1. Standard getData (ets = false)
	err = psControl.getData(100, 0, false)
	require.NoError(t, err)
	assert.True(t, refreshed)

	// 2. ETS getData (ets = true)
	err = psControl.getData(100, 0, true)
	require.NoError(t, err)
	assert.True(t, refreshedEts)
}
