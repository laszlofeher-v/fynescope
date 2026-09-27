package netcontrol

import (
	"net"
	"reflect"
	"testing"
	"time"

	"fynescope/control"
	"fynescope/genericps"
	"fynescope/settings"
)

// startTestServer is a helper that creates a Server backed by a mock controller,
// starts listening on a random port, and returns the server + mock.
func startTestServer(t *testing.T) (*Server, *mockScopeController) {
	t.Helper()
	mock := newMockScopeController()
	srv := NewServer(mock)
	if err := srv.Start("127.0.0.1:0"); err != nil {
		t.Fatalf("Server.Start failed: %v", err)
	}
	t.Cleanup(func() { srv.Close() })
	return srv, mock
}

// dialTestClientTyped dials the given server and returns a *Client.
func dialTestClientTyped(t *testing.T, srv *Server) *Client {
	t.Helper()
	c, err := Dial(srv.Addr().String())
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	t.Cleanup(func() { c.Close() })
	return c
}

// waitForCondition polls fn up to maxWait, sleeping pollInterval between tries.
func waitForCondition(t *testing.T, maxWait, pollInterval time.Duration, fn func() bool) {
	t.Helper()
	deadline := time.Now().Add(maxWait)
	for time.Now().Before(deadline) {
		if fn() {
			return
		}
		time.Sleep(pollInterval)
	}
	t.Helper()
	t.Fatal("condition not met within timeout")
}

// ---------------------------------------------------------------------------
// TestClient_Dial
// ---------------------------------------------------------------------------

func TestClient_Dial(t *testing.T) {
	srv, _ := startTestServer(t)

	c, err := Dial(srv.Addr().String())
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer c.Close()

	// Verify the server sees the client.
	waitForCondition(t, 500*time.Millisecond, 10*time.Millisecond, func() bool {
		return srv.ClientCount() == 1
	})
}

func TestClient_DialBadAddress(t *testing.T) {
	_, err := Dial("127.0.0.1:1") // port 1 is not open
	if err == nil {
		t.Fatal("expected Dial to fail with unreachable address")
	}
}

func TestClient_Close(t *testing.T) {
	srv, _ := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	waitForCondition(t, 500*time.Millisecond, 10*time.Millisecond, func() bool {
		return srv.ClientCount() == 1
	})

	if err := c.Close(); err != nil {
		t.Fatalf("Close returned unexpected error: %v", err)
	}

	// Server should drop the session after the connection closes.
	waitForCondition(t, 500*time.Millisecond, 10*time.Millisecond, func() bool {
		return srv.ClientCount() == 0
	})
}

// ---------------------------------------------------------------------------
// TestClient_FireAndForget — Commands that require no response
// ---------------------------------------------------------------------------

func TestClient_SetChannel(t *testing.T) {
	srv, mock := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	ch := settings.ChSettings{
		ID:      genericps.ChB,
		Enabled: true,
		VRange:  genericps.Range_500mv,
	}
	c.SetChannel(&ch)

	waitForCondition(t, 500*time.Millisecond, 10*time.Millisecond, func() bool {
		mock.mu.Lock()
		defer mock.mu.Unlock()
		return mock.lastChannelSettings.ID == genericps.ChB && mock.lastChannelSettings.Enabled
	})

	mock.mu.Lock()
	defer mock.mu.Unlock()
	if mock.lastChannelSettings.VRange != genericps.Range_500mv {
		t.Fatalf("unexpected VRange: %v", mock.lastChannelSettings.VRange)
	}
}

func TestClient_SetTrigger(t *testing.T) {
	srv, mock := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	done := make(chan struct{}, 1)
	msg := &control.TriggerDescMsg{
		TriggerDesc: control.TriggerDesc{
			Enabled:    true,
			TriggerADC: 1234,
			Mv:         250,
			Source:     genericps.ChA,
		},
		Done: done,
	}
	c.SetTrigger(msg)

	// Done should be signalled by the client immediately.
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("SetTrigger: Done channel was not signalled")
	}

	// Server should eventually receive the trigger.
	waitForCondition(t, 500*time.Millisecond, 10*time.Millisecond, func() bool {
		mock.mu.Lock()
		defer mock.mu.Unlock()
		return mock.lastTriggerDesc.Enabled && mock.lastTriggerDesc.TriggerADC == 1234
	})
}

func TestClient_SetTrigger_NilDone(t *testing.T) {
	srv, mock := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	msg := &control.TriggerDescMsg{
		TriggerDesc: control.TriggerDesc{Enabled: true, TriggerADC: 777},
		Done:        nil,
	}
	// Must not panic when Done is nil.
	c.SetTrigger(msg)

	waitForCondition(t, 500*time.Millisecond, 10*time.Millisecond, func() bool {
		mock.mu.Lock()
		defer mock.mu.Unlock()
		return mock.lastTriggerDesc.TriggerADC == 777
	})
}

func TestClient_SetInterpolationMode(t *testing.T) {
	srv, mock := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	c.SetInterpolationMode(settings.Linear)

	waitForCondition(t, 500*time.Millisecond, 10*time.Millisecond, func() bool {
		mock.mu.Lock()
		defer mock.mu.Unlock()
		return mock.lastInterpMode == settings.Linear
	})
}

func TestClient_SetScopeScreenWidth(t *testing.T) {
	srv, mock := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	c.SetScopeScreenWidth(800.0)

	waitForCondition(t, 500*time.Millisecond, 10*time.Millisecond, func() bool {
		mock.mu.Lock()
		defer mock.mu.Unlock()
		return mock.lastScreenWidth == 800.0
	})
}

func TestClient_SetMaxScreenTime(t *testing.T) {
	srv, mock := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	c.SetMaxScreenTime(0.05)

	waitForCondition(t, 500*time.Millisecond, 10*time.Millisecond, func() bool {
		mock.mu.Lock()
		defer mock.mu.Unlock()
		return mock.lastMaxScreenTime == 0.05
	})
}

func TestClient_SuggestSampleCount(t *testing.T) {
	srv, mock := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	c.SuggestSampleCount(65536)

	waitForCondition(t, 500*time.Millisecond, 10*time.Millisecond, func() bool {
		mock.mu.Lock()
		defer mock.mu.Unlock()
		return mock.lastSampleCount == 65536
	})
}

func TestClient_SetDigitalPort(t *testing.T) {
	srv, mock := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	portMsg := &control.DigitalPortMsg{Port: 1}
	c.SetDigitalPort(portMsg)

	waitForCondition(t, 500*time.Millisecond, 10*time.Millisecond, func() bool {
		mock.mu.Lock()
		defer mock.mu.Unlock()
		return mock.lastDigitalPortMsg.Port == 1
	})
}

func TestClient_SetGenerator(t *testing.T) {
	srv, mock := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	done := make(chan struct{}, 1)
	msg := &control.GeneratorDescMsg{
		GeneratorDesc: control.GeneratorDesc{StartFrequency: 1000.0, On: true},
		Done:          done,
	}
	c.SetGenerator(msg)

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("SetGenerator: Done channel was not signalled")
	}

	waitForCondition(t, 500*time.Millisecond, 10*time.Millisecond, func() bool {
		mock.mu.Lock()
		defer mock.mu.Unlock()
		return mock.lastGeneratorDesc.StartFrequency == 1000.0 && mock.lastGeneratorDesc.On
	})
}

func TestClient_SetDemoGen(t *testing.T) {
	srv, mock := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	done := make(chan struct{}, 1)
	msg := &control.GeneratorDescMsg{
		GeneratorDesc: control.GeneratorDesc{StartFrequency: 500.0, On: true},
		Done:          done,
	}
	c.SetDemoGen(msg)

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("SetDemoGen: Done channel was not signalled")
	}

	waitForCondition(t, 500*time.Millisecond, 10*time.Millisecond, func() bool {
		mock.mu.Lock()
		defer mock.mu.Unlock()
		return mock.lastDemoGenDesc.StartFrequency == 500.0 && mock.lastDemoGenDesc.On
	})
}

func TestClient_SetStreamEnabled(t *testing.T) {
	srv, mock := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	c.SetStreamEnabled(true)

	waitForCondition(t, 500*time.Millisecond, 10*time.Millisecond, func() bool {
		mock.mu.Lock()
		defer mock.mu.Unlock()
		return mock.lastStreamEnabled
	})
}

func TestClient_RequestRestart(t *testing.T) {
	srv, mock := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	c.RequestRestart()

	waitForCondition(t, 500*time.Millisecond, 10*time.Millisecond, func() bool {
		mock.mu.Lock()
		defer mock.mu.Unlock()
		return mock.restarted
	})
}

func TestClient_NewChannels(t *testing.T) {
	srv, mock := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	c.NewChannels(4)

	waitForCondition(t, 500*time.Millisecond, 10*time.Millisecond, func() bool {
		mock.mu.Lock()
		defer mock.mu.Unlock()
		return mock.lastNumChannels == 4
	})
}

func TestClient_SetMaxSamplingRate(t *testing.T) {
	srv, mock := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	c.SetMaxSamplingRate(500_000_000)

	waitForCondition(t, 500*time.Millisecond, 10*time.Millisecond, func() bool {
		mock.mu.Lock()
		defer mock.mu.Unlock()
		return mock.lastMaxRate == 500_000_000
	})
}

func TestClient_SetResolutionMode(t *testing.T) {
	srv, mock := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	c.SetResolutionMode(genericps.RatioModeAggregate)

	waitForCondition(t, 500*time.Millisecond, 10*time.Millisecond, func() bool {
		mock.mu.Lock()
		defer mock.mu.Unlock()
		return mock.lastResolutionMode == genericps.RatioModeAggregate
	})
}

func TestClient_SetDigitalPortEnabled(t *testing.T) {
	srv, mock := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	c.SetDigitalPortEnabled(2, true)

	waitForCondition(t, 500*time.Millisecond, 10*time.Millisecond, func() bool {
		mock.mu.Lock()
		defer mock.mu.Unlock()
		return mock.lastDigitalPort == 2 && mock.lastDigitalPortEn
	})
}

func TestClient_SetInfo(t *testing.T) {
	srv, mock := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	c.SetInfo("test-info-string")

	waitForCondition(t, 500*time.Millisecond, 10*time.Millisecond, func() bool {
		mock.mu.Lock()
		defer mock.mu.Unlock()
		return mock.lastInfo == "test-info-string"
	})
}

func TestClient_SetScopeModel(t *testing.T) {
	srv, mock := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	c.SetScopeModel(control.Scope3406D)

	waitForCondition(t, 500*time.Millisecond, 10*time.Millisecond, func() bool {
		mock.mu.Lock()
		defer mock.mu.Unlock()
		return mock.lastModel == control.Scope3406D
	})
}

func TestClient_Shutdown(t *testing.T) {
	srv, mock := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	c.Shutdown()

	waitForCondition(t, 500*time.Millisecond, 10*time.Millisecond, func() bool {
		mock.mu.Lock()
		defer mock.mu.Unlock()
		return mock.shutdown
	})
}

// ---------------------------------------------------------------------------
// TestClient_RPC — Synchronous query methods
// ---------------------------------------------------------------------------

func TestClient_Stop_Success(t *testing.T) {
	srv, _ := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	if err := c.Stop(); err != nil {
		t.Fatalf("Stop returned error: %v", err)
	}
}

func TestClient_Stop_Error(t *testing.T) {
	srv, mock := startTestServer(t)
	mock.errStop = net.ErrClosed
	c := dialTestClientTyped(t, srv)

	err := c.Stop()
	if err == nil {
		t.Fatal("expected Stop to return an error")
	}
}

func TestClient_SetETSMode(t *testing.T) {
	srv, _ := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	if err := c.SetETSMode(); err != nil {
		t.Fatalf("SetETSMode returned error: %v", err)
	}
}

func TestClient_SetBlockMode(t *testing.T) {
	srv, _ := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	if err := c.SetBlockMode(); err != nil {
		t.Fatalf("SetBlockMode returned error: %v", err)
	}
}

func TestClient_SetBlockMode_Error(t *testing.T) {
	srv, mock := startTestServer(t)
	mock.errBlockMode = net.ErrClosed
	c := dialTestClientTyped(t, srv)

	if err := c.SetBlockMode(); err == nil {
		t.Fatal("expected SetBlockMode to propagate server error")
	}
}

func TestClient_GetInfo(t *testing.T) {
	srv, _ := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	info := c.GetInfo()
	if info != "PicoScope 2206B" {
		t.Fatalf("unexpected GetInfo result: %q", info)
	}
}

func TestClient_GetScopeModel(t *testing.T) {
	srv, _ := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	model := c.GetScopeModel()
	if model != control.Scope2206B {
		t.Fatalf("unexpected model: %v", model)
	}
}

func TestClient_GetMaxSamplingRate(t *testing.T) {
	srv, _ := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	rate := c.GetMaxSamplingRate()
	if rate != 1_000_000_000 {
		t.Fatalf("unexpected rate: %d", rate)
	}
}

func TestClient_GetStreamEnabled(t *testing.T) {
	srv, mock := startTestServer(t)
	mock.lastStreamEnabled = true
	c := dialTestClientTyped(t, srv)

	if !c.GetStreamEnabled() {
		t.Fatal("expected GetStreamEnabled to return true")
	}
}

func TestClient_GetSamplingTimeInterval(t *testing.T) {
	srv, _ := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	interval := c.GetSamplingTimeInterval()
	if interval != 1e-9 {
		t.Fatalf("unexpected interval: %v", interval)
	}
}

func TestClient_GetTimeBase(t *testing.T) {
	srv, _ := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	tb := c.GetTimeBase()
	if tb != 12345 {
		t.Fatalf("unexpected time base: %d", tb)
	}
}

func TestClient_NumberOfEnabledAnalogChannels(t *testing.T) {
	srv, _ := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	n := c.NumberOfEnabledAnalogChannels()
	if n != 2 {
		t.Fatalf("unexpected channel count: %d", n)
	}
}

func TestClient_MinMaxValues(t *testing.T) {
	srv, _ := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	min, max, err := c.MinMaxValues()
	if err != nil {
		t.Fatalf("MinMaxValues error: %v", err)
	}
	if min != -32768 || max != 32767 {
		t.Fatalf("unexpected min/max: %d/%d", min, max)
	}
}

func TestClient_MinMaxValues_Error(t *testing.T) {
	srv, mock := startTestServer(t)
	mock.errMinMax = net.ErrClosed
	c := dialTestClientTyped(t, srv)

	if _, _, err := c.MinMaxValues(); err == nil {
		t.Fatal("expected MinMaxValues to propagate server error")
	}
}

func TestClient_UnitVariantInfo(t *testing.T) {
	srv, _ := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	info, err := c.UnitVariantInfo()
	if err != nil {
		t.Fatalf("UnitVariantInfo error: %v", err)
	}
	if info != "2206B" {
		t.Fatalf("unexpected variant info: %q", info)
	}
}

func TestClient_UnitBatchAndSerialInfo(t *testing.T) {
	srv, _ := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	info, err := c.UnitBatchAndSerialInfo()
	if err != nil {
		t.Fatalf("UnitBatchAndSerialInfo error: %v", err)
	}
	if info != "ABC12345" {
		t.Fatalf("unexpected batch/serial info: %q", info)
	}
}

func TestClient_ChannelRanges(t *testing.T) {
	srv, _ := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	ranges, err := c.ChannelRanges(genericps.ChA)
	if err != nil {
		t.Fatalf("ChannelRanges error: %v", err)
	}
	expected := []int32{100, 200, 500, 1000}
	if !reflect.DeepEqual(ranges, expected) {
		t.Fatalf("unexpected ranges: %v", ranges)
	}
}

func TestClient_ChannelRanges_Error(t *testing.T) {
	srv, mock := startTestServer(t)
	mock.errRanges = net.ErrClosed
	c := dialTestClientTyped(t, srv)

	if _, err := c.ChannelRanges(genericps.ChA); err == nil {
		t.Fatal("expected ChannelRanges to propagate server error")
	}
}

func TestClient_GetAnalogueOffset(t *testing.T) {
	srv, _ := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	maxV, minV, err := c.GetAnalogueOffset(1000, genericps.Dc)
	if err != nil {
		t.Fatalf("GetAnalogueOffset error: %v", err)
	}
	if maxV != 5.0 || minV != -5.0 {
		t.Fatalf("unexpected offset: max=%v min=%v", maxV, minV)
	}
}

func TestClient_GetAnalogueOffset_Error(t *testing.T) {
	srv, mock := startTestServer(t)
	mock.errOffset = net.ErrClosed
	c := dialTestClientTyped(t, srv)

	if _, _, err := c.GetAnalogueOffset(1000, genericps.Dc); err == nil {
		t.Fatal("expected GetAnalogueOffset to propagate server error")
	}
}

func TestClient_GetEtsLimits(t *testing.T) {
	srv, _ := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	maxI, maxC := c.GetEtsLimits()
	if maxI != 4 || maxC != 10 {
		t.Fatalf("unexpected ETS limits: interleave=%d cycles=%d", maxI, maxC)
	}
}

func TestClient_IsDemo(t *testing.T) {
	srv, _ := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	if !c.IsDemo() {
		t.Fatal("expected IsDemo to return true")
	}
}

func TestClient_SetDemoDigitalGen(t *testing.T) {
	srv, _ := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	err := c.SetDemoDigitalGen(true, false, 1000.0,
		genericps.DigitalDemoGenDirectionUp,
		genericps.DigitalDemoGenEncodingBinary,
		genericps.DigitalDemoGenModeSynchronous,
		0.001)
	if err != nil {
		t.Fatalf("SetDemoDigitalGen error: %v", err)
	}
}

func TestClient_SetDemoRlcFilter(t *testing.T) {
	srv, _ := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	err := c.SetDemoRlcFilter(genericps.ChA, genericps.ChA, true, "LowPass",
		100.0, "Ω", 1e-3, "H", 1e-6, "F")
	if err != nil {
		t.Fatalf("SetDemoRlcFilter error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestClient_Telemetry — Server-pushed callbacks reach the Client
// ---------------------------------------------------------------------------

func TestClient_TelemetryRefresh(t *testing.T) {
	srv, mock := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	// Wait for server to register the client session.
	waitForCondition(t, 500*time.Millisecond, 10*time.Millisecond, func() bool {
		return srv.ClientCount() == 1
	})

	received := make(chan *DataMessage, 1)
	c.SetRefreshCallback(func(buffers [][]int16, buffersMin [][]int16, digitalBuffers [][]int16, startTimeOffset int64, xRoundError float64, samplingTimeInterval float64) {
		received <- &DataMessage{
			Buffers:              buffers,
			BuffersMin:           buffersMin,
			DigitalBuffers:       digitalBuffers,
			StartTimeOffset:      startTimeOffset,
			XRoundError:          xRoundError,
			SamplingTimeInterval: samplingTimeInterval,
		}
	})

	bufs := [][]int16{{1, 2, 3}, {4, 5, 6}}
	mock.CallRefreshCallback(bufs, nil, nil, 999, 0.01, 1e-7)

	select {
	case data := <-received:
		if !reflect.DeepEqual(data.Buffers, bufs) {
			t.Fatalf("buffer mismatch: %v", data.Buffers)
		}
		if data.StartTimeOffset != 999 {
			t.Fatalf("unexpected start time offset: %d", data.StartTimeOffset)
		}
		if data.SamplingTimeInterval != 1e-7 {
			t.Fatalf("unexpected sampling interval: %v", data.SamplingTimeInterval)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout: Refresh callback not called")
	}
}

func TestClient_TelemetryRefreshEts(t *testing.T) {
	srv, mock := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	waitForCondition(t, 500*time.Millisecond, 10*time.Millisecond, func() bool {
		return srv.ClientCount() == 1
	})

	received := make(chan *EtsDataMessage, 1)
	c.SetRefreshEtsCallback(func(buffers [][]int16, etsOutBuffer []int64, xRoundError float64, samplingTimeInterval float64) {
		received <- &EtsDataMessage{
			Buffers:              buffers,
			EtsOutBuffer:         etsOutBuffer,
			XRoundError:          xRoundError,
			SamplingTimeInterval: samplingTimeInterval,
		}
	})

	bufs := [][]int16{{10, 20}}
	etsOut := []int64{100, 200, 300}

	mock.mu.Lock()
	cb := mock.refreshEtsCallback
	mock.mu.Unlock()
	if cb != nil {
		cb(bufs, etsOut, 0.02, 2e-7)
	}

	select {
	case data := <-received:
		if !reflect.DeepEqual(data.Buffers, bufs) {
			t.Fatalf("buffer mismatch: %v", data.Buffers)
		}
		if !reflect.DeepEqual(data.EtsOutBuffer, etsOut) {
			t.Fatalf("ets buffer mismatch: %v", data.EtsOutBuffer)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout: RefreshEts callback not called")
	}
}

func TestClient_TelemetryBufferResize(t *testing.T) {
	srv, mock := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	waitForCondition(t, 500*time.Millisecond, 10*time.Millisecond, func() bool {
		return srv.ClientCount() == 1
	})

	received := make(chan int, 1)
	c.SetBufferCallback(func(size int) {
		received <- size
	})

	mock.CallBufferCallback(8192)

	select {
	case size := <-received:
		if size != 8192 {
			t.Fatalf("unexpected buffer size: %d", size)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout: BufferCallback not called")
	}
}

func TestClient_TelemetryEtsBufferResize(t *testing.T) {
	srv, mock := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	waitForCondition(t, 500*time.Millisecond, 10*time.Millisecond, func() bool {
		return srv.ClientCount() == 1
	})

	received := make(chan int, 1)
	c.SetEtsBufferCallback(func(size int) {
		received <- size
	})

	mock.mu.Lock()
	cb := mock.etsBufferCallback
	mock.mu.Unlock()
	if cb != nil {
		cb(4096)
	}

	select {
	case size := <-received:
		if size != 4096 {
			t.Fatalf("unexpected ETS buffer size: %d", size)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout: EtsBufferCallback not called")
	}
}

func TestClient_TelemetryStatus(t *testing.T) {
	srv, mock := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	waitForCondition(t, 500*time.Millisecond, 10*time.Millisecond, func() bool {
		return srv.ClientCount() == 1
	})

	received := make(chan *StatusMessage, 1)
	c.SetDisplayStatus(func(s string, errorType control.ScopeError) {
		received <- &StatusMessage{Text: s, ErrorType: errorType}
	})

	mock.ShowDisplayStatus("acquiring", control.Info)

	select {
	case st := <-received:
		if st.Text != "acquiring" || st.ErrorType != control.Info {
			t.Fatalf("unexpected status: %+v", st)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout: DisplayStatus callback not called")
	}
}

// ---------------------------------------------------------------------------
// TestClient_GetCon
// ---------------------------------------------------------------------------

func TestClient_GetCon(t *testing.T) {
	srv, _ := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	if c.GetCon() != nil {
		t.Fatal("expected GetCon to return nil for remote client")
	}
}

// ---------------------------------------------------------------------------
// TestClient_ConcurrentRPCs — Multiple in-flight queries resolved correctly
// ---------------------------------------------------------------------------

func TestClient_ConcurrentRPCs(t *testing.T) {
	srv, _ := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	done := make(chan string, 10)
	for i := 0; i < 5; i++ {
		go func() {
			info := c.GetInfo()
			done <- info
		}()
	}

	for i := 0; i < 5; i++ {
		select {
		case info := <-done:
			if info != "PicoScope 2206B" {
				t.Errorf("unexpected GetInfo result: %q", info)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("concurrent RPC %d timed out", i+1)
		}
	}
}

// ---------------------------------------------------------------------------
// TestClient_CallbackHelpers — CallBufferCallback / CallRefreshCallback / ShowDisplayStatus
// ---------------------------------------------------------------------------

func TestClient_CallBufferCallback(t *testing.T) {
	srv, _ := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	called := make(chan int, 1)
	c.SetBufferCallback(func(size int) { called <- size })
	c.CallBufferCallback(2048)

	select {
	case size := <-called:
		if size != 2048 {
			t.Fatalf("unexpected size: %d", size)
		}
	default:
		t.Fatal("CallBufferCallback did not invoke the registered callback")
	}
}

func TestClient_CallRefreshCallback(t *testing.T) {
	srv, _ := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	called := make(chan [][]int16, 1)
	c.SetRefreshCallback(func(buffers [][]int16, _ [][]int16, _ [][]int16, _ int64, _ float64, _ float64) {
		called <- buffers
	})

	bufs := [][]int16{{7, 8, 9}}
	c.CallRefreshCallback(bufs, nil, nil, 0, 0, 0)

	select {
	case got := <-called:
		if !reflect.DeepEqual(got, bufs) {
			t.Fatalf("unexpected buffers: %v", got)
		}
	default:
		t.Fatal("CallRefreshCallback did not invoke the registered callback")
	}
}

func TestClient_ShowDisplayStatus(t *testing.T) {
	srv, _ := startTestServer(t)
	c := dialTestClientTyped(t, srv)

	called := make(chan string, 1)
	c.SetDisplayStatus(func(s string, _ control.ScopeError) { called <- s })
	c.ShowDisplayStatus("hello", control.Info)

	select {
	case got := <-called:
		if got != "hello" {
			t.Fatalf("unexpected status text: %q", got)
		}
	default:
		t.Fatal("ShowDisplayStatus did not invoke the registered callback")
	}
}

// ---------------------------------------------------------------------------
// TestClient_InterfaceCompliance — compile-time check
// ---------------------------------------------------------------------------

var _ control.ScopeController = (*Client)(nil)
