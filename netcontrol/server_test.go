package netcontrol

import (
	"encoding/gob"
	"errors"
	"net"
	"reflect"
	"sync"
	"testing"
	"time"

	"fynescope/control"
	"fynescope/genericps"
	"fynescope/settings"
)

// mockScopeController implements control.ScopeController for testing the TCP server.
type mockScopeController struct {
	mu sync.Mutex

	// Callbacks registered by server
	refreshCallback    func(buffers [][]int16, buffersMin [][]int16, digitalBuffers [][]int16, startTimeOffset int64, xRoundError float64, samplingTimeInterval float64)
	refreshEtsCallback func(buffers [][]int16, etsOutBuffer []int64, xRoundError float64, samplingTimeInterval float64)
	bufferCallback     func(size int)
	etsBufferCallback  func(size int)
	displayStatus      func(s string, errorType control.ScopeError)

	// Recorded commands
	lastTriggerDesc     control.TriggerDesc
	lastChannelSettings settings.ChSettings
	lastDigitalPortMsg  control.DigitalPortMsg
	lastGeneratorDesc   control.GeneratorDesc
	lastDemoGenDesc     control.GeneratorDesc
	lastInterpMode      settings.InterpolationType
	lastScreenWidth     float64
	lastMaxScreenTime   float64
	lastSampleCount     uint64
	lastResolutionMode  genericps.RatioMode
	lastDigitalPort     int
	lastDigitalPortEn   bool
	lastMaxRate         uint32
	lastInfo            string
	lastModel           control.ScopeType
	lastStreamEnabled   bool
	lastNumChannels     int
	restarted           bool
	stopped             bool
	shutdown            bool

	// Queries return values
	mockInfo            string
	mockModel           control.ScopeType
	mockMaxRate         uint32
	mockSamplingTime    float64
	mockTimeBase        uint64
	mockEnabledAnalogCh int
	mockMinVal          int32
	mockMaxVal          int32
	mockVariantInfo     string
	mockBatchSerialInfo string
	mockChannelRanges   []int32
	mockOffsetMax       float32
	mockOffsetMin       float32
	mockEtsInterleave   int16
	mockEtsCycles       int16
	mockIsDemo          bool

	// Errors
	errEtsMode    error
	errBlockMode  error
	errStop       error
	errOffset     error
	errRanges     error
	errMinMax     error
	errVariant    error
	errBatch      error
	errDemoDig    error
	errDemoFilter error
}

func newMockScopeController() *mockScopeController {
	return &mockScopeController{
		mockInfo:            "PicoScope 2206B",
		mockModel:           control.Scope2206B,
		mockMaxRate:         1000000000,
		mockSamplingTime:    1e-9,
		mockTimeBase:        12345,
		mockEnabledAnalogCh: 2,
		mockMinVal:          -32768,
		mockMaxVal:          32767,
		mockVariantInfo:     "2206B",
		mockBatchSerialInfo: "ABC12345",
		mockChannelRanges:   []int32{100, 200, 500, 1000},
		mockOffsetMax:       5.0,
		mockOffsetMin:       -5.0,
		mockEtsInterleave:   4,
		mockEtsCycles:       10,
		mockIsDemo:          true,
	}
}

func (m *mockScopeController) SetTrigger(msg *control.TriggerDescMsg) {
	m.mu.Lock()
	m.lastTriggerDesc = msg.TriggerDesc
	m.mu.Unlock()
	if msg.Done != nil {
		msg.Done <- struct{}{}
	}
}

func (m *mockScopeController) SetChannel(msg *settings.ChSettings) {
	m.mu.Lock()
	m.lastChannelSettings = *msg
	m.mu.Unlock()
}

func (m *mockScopeController) SetDigitalPort(msg *control.DigitalPortMsg) {
	m.mu.Lock()
	m.lastDigitalPortMsg = *msg
	m.mu.Unlock()
}

func (m *mockScopeController) SetGenerator(msg *control.GeneratorDescMsg) {
	m.mu.Lock()
	m.lastGeneratorDesc = msg.GeneratorDesc
	m.mu.Unlock()
	if msg.Done != nil {
		msg.Done <- struct{}{}
	}
}

func (m *mockScopeController) SetDemoGen(msg *control.GeneratorDescMsg) {
	m.mu.Lock()
	m.lastDemoGenDesc = msg.GeneratorDesc
	m.mu.Unlock()
	if msg.Done != nil {
		msg.Done <- struct{}{}
	}
}

func (m *mockScopeController) SetInterpolationMode(mode settings.InterpolationType) {
	m.mu.Lock()
	m.lastInterpMode = mode
	m.mu.Unlock()
}

func (m *mockScopeController) SetScopeScreenWidth(width float64) {
	m.mu.Lock()
	m.lastScreenWidth = width
	m.mu.Unlock()
}

func (m *mockScopeController) SetMaxScreenTime(t float64) {
	m.mu.Lock()
	m.lastMaxScreenTime = t
	m.mu.Unlock()
}

func (m *mockScopeController) SuggestSampleCount(sc uint64) {
	m.mu.Lock()
	m.lastSampleCount = sc
	m.mu.Unlock()
}

func (m *mockScopeController) SetETSMode() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.errEtsMode
}

func (m *mockScopeController) SetBlockMode() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.errBlockMode
}

func (m *mockScopeController) Shutdown() {
	m.mu.Lock()
	m.shutdown = true
	m.mu.Unlock()
}

func (m *mockScopeController) Stop() error {
	m.mu.Lock()
	m.stopped = true
	err := m.errStop
	m.mu.Unlock()
	return err
}

func (m *mockScopeController) GetCon() *genericps.Connection {
	return nil
}

func (m *mockScopeController) GetInfo() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.mockInfo
}

func (m *mockScopeController) GetScopeModel() control.ScopeType {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.mockModel
}

func (m *mockScopeController) GetMaxSamplingRate() uint32 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.mockMaxRate
}

func (m *mockScopeController) GetStreamEnabled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastStreamEnabled
}

func (m *mockScopeController) SetStreamEnabled(b bool) {
	m.mu.Lock()
	m.lastStreamEnabled = b
	m.mu.Unlock()
}

func (m *mockScopeController) GetSamplingTimeInterval() float64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.mockSamplingTime
}

func (m *mockScopeController) GetTimeBase() uint64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.mockTimeBase
}

func (m *mockScopeController) ChannelRanges(chIndex genericps.ChannelId) ([]int32, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.mockChannelRanges, m.errRanges
}

func (m *mockScopeController) NumberOfEnabledAnalogChannels() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.mockEnabledAnalogCh
}

func (m *mockScopeController) MinMaxValues() (min, max int32, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.mockMinVal, m.mockMaxVal, m.errMinMax
}

func (m *mockScopeController) UnitVariantInfo() (info string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.mockVariantInfo, m.errVariant
}

func (m *mockScopeController) UnitBatchAndSerialInfo() (info string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.mockBatchSerialInfo, m.errBatch
}

func (m *mockScopeController) SetDigitalPortEnabled(port int, enabled bool) {
	m.mu.Lock()
	m.lastDigitalPort = port
	m.lastDigitalPortEn = enabled
	m.mu.Unlock()
}

func (m *mockScopeController) SetRefreshCallback(cb func(buffers [][]int16, buffersMin [][]int16, digitalBuffers [][]int16, startTimeOffset int64, xRoundError float64, samplingTimeInterval float64)) {
	m.mu.Lock()
	m.refreshCallback = cb
	m.mu.Unlock()
}

func (m *mockScopeController) SetRefreshEtsCallback(cb func(buffers [][]int16, etsOutBuffer []int64, xRoundError float64, samplingTimeInterval float64)) {
	m.mu.Lock()
	m.refreshEtsCallback = cb
	m.mu.Unlock()
}

func (m *mockScopeController) SetBufferCallback(cb func(size int)) {
	m.mu.Lock()
	m.bufferCallback = cb
	m.mu.Unlock()
}

func (m *mockScopeController) SetEtsBufferCallback(cb func(size int)) {
	m.mu.Lock()
	m.etsBufferCallback = cb
	m.mu.Unlock()
}

func (m *mockScopeController) SetDisplayStatus(cb func(s string, errorType control.ScopeError)) {
	m.mu.Lock()
	m.displayStatus = cb
	m.mu.Unlock()
}

func (m *mockScopeController) ShowDisplayStatus(s string, errorType control.ScopeError) {
	m.mu.Lock()
	cb := m.displayStatus
	m.mu.Unlock()
	if cb != nil {
		cb(s, errorType)
	}
}

func (m *mockScopeController) NewChannels(numberOfChannels int) {
	m.mu.Lock()
	m.lastNumChannels = numberOfChannels
	m.mu.Unlock()
}

func (m *mockScopeController) SetMaxSamplingRate(rate uint32) {
	m.mu.Lock()
	m.lastMaxRate = rate
	m.mu.Unlock()
}

func (m *mockScopeController) RequestRestart() {
	m.mu.Lock()
	m.restarted = true
	m.mu.Unlock()
}

func (m *mockScopeController) SetResolutionMode(mode genericps.RatioMode) {
	m.mu.Lock()
	m.lastResolutionMode = mode
	m.mu.Unlock()
}

func (m *mockScopeController) CallBufferCallback(size int) {
	m.mu.Lock()
	cb := m.bufferCallback
	m.mu.Unlock()
	if cb != nil {
		cb(size)
	}
}

func (m *mockScopeController) CallRefreshCallback(buffers [][]int16, buffersMin [][]int16, digitalBuffers [][]int16, startTimeOffset int64, xRoundError float64, samplingTimeInterval float64) {
	m.mu.Lock()
	cb := m.refreshCallback
	m.mu.Unlock()
	if cb != nil {
		cb(buffers, buffersMin, digitalBuffers, startTimeOffset, xRoundError, samplingTimeInterval)
	}
}

func (m *mockScopeController) SetInfo(info string) {
	m.mu.Lock()
	m.lastInfo = info
	m.mu.Unlock()
}

func (m *mockScopeController) SetScopeModel(model control.ScopeType) {
	m.mu.Lock()
	m.lastModel = model
	m.mu.Unlock()
}

func (m *mockScopeController) GetEtsLimits() (maxInterleave, maxCycles int16) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.mockEtsInterleave, m.mockEtsCycles
}

func (m *mockScopeController) GetAnalogueOffset(voltageRange int, coupling genericps.Coupling) (maximumVoltage, minimumVoltage float32, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.mockOffsetMax, m.mockOffsetMin, m.errOffset
}

func (m *mockScopeController) IsDemo() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.mockIsDemo
}

func (m *mockScopeController) SetDemoDigitalGen(port0Enabled, port1Enabled bool, freq float64, dir genericps.DigitalDemoGenDirection, enc genericps.DigitalDemoGenEncoding, mode genericps.DigitalDemoGenMode, bitDelay float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.errDemoDig
}

func (m *mockScopeController) SetDemoRlcFilter(channel genericps.ChannelId, genSource genericps.ChannelId, enabled bool, filterType string, r float64, runit string, l float64, lunit string, cval float64, cunit string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.errDemoFilter
}

// Helper to create client connection and gob encoders/decoders.
func dialTestClient(t *testing.T, s *Server) (net.Conn, *gob.Encoder, *gob.Decoder) {
	t.Helper()
	conn, err := net.Dial("tcp", s.Addr().String())
	if err != nil {
		t.Fatalf("failed to dial server: %v", err)
	}
	return conn, gob.NewEncoder(conn), gob.NewDecoder(conn)
}

func TestServer_LifecycleAndConnection(t *testing.T) {
	mock := newMockScopeController()
	srv := NewServer(mock)

	err := srv.Start("127.0.0.1:0")
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer srv.Close()

	if srv.Addr() == nil {
		t.Fatal("expected non-nil server address")
	}

	if srv.ClientCount() != 0 {
		t.Fatalf("expected 0 clients, got %d", srv.ClientCount())
	}

	conn, _, _ := dialTestClient(t, srv)
	defer conn.Close()

	time.Sleep(50 * time.Millisecond)
	if srv.ClientCount() != 1 {
		t.Fatalf("expected 1 client, got %d", srv.ClientCount())
	}
}

func TestServer_CommandsDispatch(t *testing.T) {
	mock := newMockScopeController()
	srv := NewServer(mock)

	if err := srv.Start("127.0.0.1:0"); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer srv.Close()

	conn, enc, dec := dialTestClient(t, srv)
	defer conn.Close()

	time.Sleep(20 * time.Millisecond)

	// 1. SetTrigger (with ID for ACK)
	trigCmd := &CommandMessage{
		ID:   1,
		Type: TypeSetTrigger,
		Payload: &SetTriggerCmd{
			Desc: control.TriggerDesc{
				Enabled:    true,
				TriggerADC: 500,
				Mv:         100,
				Source:     genericps.ChA,
			},
		},
	}
	if err := enc.Encode(trigCmd); err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	var trigRsp ServerMessage
	if err := dec.Decode(&trigRsp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if trigRsp.Type != TypeResponse {
		t.Fatalf("expected TypeResponse, got %s", trigRsp.Type)
	}
	rsp := trigRsp.Payload.(*ResponseMessage)
	if rsp.ID != 1 || !rsp.Success {
		t.Fatalf("unexpected response: %+v", rsp)
	}

	mock.mu.Lock()
	if !mock.lastTriggerDesc.Enabled || mock.lastTriggerDesc.TriggerADC != 500 {
		t.Fatalf("mock did not receive trigger: %+v", mock.lastTriggerDesc)
	}
	mock.mu.Unlock()

	// 2. SetChannel (fire-and-forget, ID = 0)
	chCmd := &CommandMessage{
		ID:   0,
		Type: TypeSetChannel,
		Payload: &SetChannelCmd{
			Settings: settings.ChSettings{
				ID:      genericps.ChB,
				Enabled: true,
				VRange:  genericps.Range_2v,
			},
		},
	}
	if err := enc.Encode(chCmd); err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	time.Sleep(30 * time.Millisecond)
	mock.mu.Lock()
	if mock.lastChannelSettings.ID != genericps.ChB || !mock.lastChannelSettings.Enabled {
		t.Fatalf("mock did not receive channel settings: %+v", mock.lastChannelSettings)
	}
	mock.mu.Unlock()

	// 3. SetInterpolationMode & SetScopeScreenWidth
	if err := enc.Encode(&CommandMessage{ID: 0, Type: TypeSetInterpolationMode, Payload: &SetInterpolationModeCmd{Mode: settings.Linear}}); err != nil {
		t.Fatal(err)
	}
	if err := enc.Encode(&CommandMessage{ID: 0, Type: TypeSetScopeScreenWidth, Payload: &SetScopeScreenWidthCmd{Width: 1024.0}}); err != nil {
		t.Fatal(err)
	}

	time.Sleep(30 * time.Millisecond)
	mock.mu.Lock()
	if mock.lastInterpMode != settings.Linear || mock.lastScreenWidth != 1024.0 {
		t.Fatalf("mock did not receive screen / interp settings: %v, %v", mock.lastInterpMode, mock.lastScreenWidth)
	}
	mock.mu.Unlock()
}

func TestServer_QueriesExecution(t *testing.T) {
	mock := newMockScopeController()
	srv := NewServer(mock)

	if err := srv.Start("127.0.0.1:0"); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer srv.Close()

	conn, enc, dec := dialTestClient(t, srv)
	defer conn.Close()

	tests := []struct {
		reqType  string
		payload  interface{}
		verifyFn func(t *testing.T, rsp *ResponseMessage)
	}{
		{
			reqType: TypeGetInfo,
			payload: &GetInfoReq{},
			verifyFn: func(t *testing.T, rsp *ResponseMessage) {
				r, ok := rsp.Result.(*GetInfoRsp)
				if !ok || r.Info != "PicoScope 2206B" {
					t.Fatalf("unexpected GetInfoRsp: %+v", rsp.Result)
				}
			},
		},
		{
			reqType: TypeGetScopeModel,
			payload: &GetScopeModelReq{},
			verifyFn: func(t *testing.T, rsp *ResponseMessage) {
				r, ok := rsp.Result.(*GetScopeModelRsp)
				if !ok || r.Model != control.Scope2206B {
					t.Fatalf("unexpected GetScopeModelRsp: %+v", rsp.Result)
				}
			},
		},
		{
			reqType: TypeMinMaxValues,
			payload: &MinMaxValuesReq{},
			verifyFn: func(t *testing.T, rsp *ResponseMessage) {
				r, ok := rsp.Result.(*MinMaxValuesRsp)
				if !ok || r.Min != -32768 || r.Max != 32767 {
					t.Fatalf("unexpected MinMaxValuesRsp: %+v", rsp.Result)
				}
			},
		},
		{
			reqType: TypeChannelRanges,
			payload: &ChannelRangesReq{Channel: genericps.ChA},
			verifyFn: func(t *testing.T, rsp *ResponseMessage) {
				r, ok := rsp.Result.(*ChannelRangesRsp)
				if !ok || !reflect.DeepEqual(r.Ranges, []int32{100, 200, 500, 1000}) {
					t.Fatalf("unexpected ChannelRangesRsp: %+v", rsp.Result)
				}
			},
		},
		{
			reqType: TypeGetAnalogueOffset,
			payload: &GetAnalogueOffsetReq{VoltageRange: 1000, Coupling: genericps.Dc},
			verifyFn: func(t *testing.T, rsp *ResponseMessage) {
				r, ok := rsp.Result.(*GetAnalogueOffsetRsp)
				if !ok || r.Max != 5.0 || r.Min != -5.0 {
					t.Fatalf("unexpected GetAnalogueOffsetRsp: %+v", rsp.Result)
				}
			},
		},
		{
			reqType: TypeIsDemo,
			payload: &IsDemoReq{},
			verifyFn: func(t *testing.T, rsp *ResponseMessage) {
				r, ok := rsp.Result.(*IsDemoRsp)
				if !ok || !r.IsDemo {
					t.Fatalf("unexpected IsDemoRsp: %+v", rsp.Result)
				}
			},
		},
	}

	for i, tc := range tests {
		reqID := uint64(i + 100)
		cmd := &CommandMessage{
			ID:      reqID,
			Type:    tc.reqType,
			Payload: tc.payload,
		}
		if err := enc.Encode(cmd); err != nil {
			t.Fatalf("[%s] encode failed: %v", tc.reqType, err)
		}

		var sMsg ServerMessage
		if err := dec.Decode(&sMsg); err != nil {
			t.Fatalf("[%s] decode failed: %v", tc.reqType, err)
		}
		if sMsg.Type != TypeResponse {
			t.Fatalf("[%s] expected TypeResponse, got %s", tc.reqType, sMsg.Type)
		}
		rsp := sMsg.Payload.(*ResponseMessage)
		if rsp.ID != reqID {
			t.Fatalf("[%s] expected ID %d, got %d", tc.reqType, reqID, rsp.ID)
		}
		if !rsp.Success {
			t.Fatalf("[%s] query returned error: %s", tc.reqType, rsp.Error)
		}
		tc.verifyFn(t, rsp)
	}
}

func TestServer_TelemetryBroadcast(t *testing.T) {
	mock := newMockScopeController()
	srv := NewServer(mock)

	if err := srv.Start("127.0.0.1:0"); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer srv.Close()

	conn1, _, dec1 := dialTestClient(t, srv)
	defer conn1.Close()
	conn2, _, dec2 := dialTestClient(t, srv)
	defer conn2.Close()

	time.Sleep(30 * time.Millisecond)
	if srv.ClientCount() != 2 {
		t.Fatalf("expected 2 clients, got %d", srv.ClientCount())
	}

	// 1. Test Refresh Telemetry
	mockBufs := [][]int16{{10, 20, 30}, {-10, -20, -30}}
	mock.CallRefreshCallback(mockBufs, nil, nil, 1000, 0.05, 1e-6)

	// Verify both clients receive the Refresh message
	for idx, dec := range []*gob.Decoder{dec1, dec2} {
		var sMsg ServerMessage
		if err := dec.Decode(&sMsg); err != nil {
			t.Fatalf("client %d decode error: %v", idx+1, err)
		}
		if sMsg.Type != TypeRefresh {
			t.Fatalf("client %d expected TypeRefresh, got %s", idx+1, sMsg.Type)
		}
		data := sMsg.Payload.(*DataMessage)
		if !reflect.DeepEqual(data.Buffers, mockBufs) {
			t.Fatalf("client %d buffer mismatch: %+v", idx+1, data.Buffers)
		}
		if data.StartTimeOffset != 1000 || data.SamplingTimeInterval != 1e-6 {
			t.Fatalf("client %d meta mismatch", idx+1)
		}
	}

	// 2. Test Buffer Resize
	mock.CallBufferCallback(4096)
	for idx, dec := range []*gob.Decoder{dec1, dec2} {
		var sMsg ServerMessage
		if err := dec.Decode(&sMsg); err != nil {
			t.Fatalf("client %d decode error: %v", idx+1, err)
		}
		if sMsg.Type != TypeBufferResize {
			t.Fatalf("client %d expected TypeBufferResize, got %s", idx+1, sMsg.Type)
		}
		bMsg := sMsg.Payload.(*BufferResizeMessage)
		if bMsg.Size != 4096 {
			t.Fatalf("client %d size mismatch: %d", idx+1, bMsg.Size)
		}
	}

	// 3. Test Display Status
	mock.ShowDisplayStatus("Sampling completed", control.Info)
	for idx, dec := range []*gob.Decoder{dec1, dec2} {
		var sMsg ServerMessage
		if err := dec.Decode(&sMsg); err != nil {
			t.Fatalf("client %d decode error: %v", idx+1, err)
		}
		if sMsg.Type != TypeStatus {
			t.Fatalf("client %d expected TypeStatus, got %s", idx+1, sMsg.Type)
		}
		stMsg := sMsg.Payload.(*StatusMessage)
		if stMsg.Text != "Sampling completed" || stMsg.ErrorType != control.Info {
			t.Fatalf("client %d status mismatch: %+v", idx+1, stMsg)
		}
	}
}

func TestServer_ErrorHandling(t *testing.T) {
	mock := newMockScopeController()
	mock.errRanges = errors.New("hardware communication failure")
	mock.errBlockMode = errors.New("device not configured for block mode")

	srv := NewServer(mock)
	if err := srv.Start("127.0.0.1:0"); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer srv.Close()

	conn, enc, dec := dialTestClient(t, srv)
	defer conn.Close()

	// 1. ChannelRanges returns error
	req := &CommandMessage{
		ID:      50,
		Type:    TypeChannelRanges,
		Payload: &ChannelRangesReq{Channel: genericps.ChA},
	}
	if err := enc.Encode(req); err != nil {
		t.Fatal(err)
	}

	var sMsg ServerMessage
	if err := dec.Decode(&sMsg); err != nil {
		t.Fatal(err)
	}
	rsp := sMsg.Payload.(*ResponseMessage)
	if rsp.ID != 50 || rsp.Success {
		t.Fatalf("expected failed response, got: %+v", rsp)
	}
	if rsp.Error != "hardware communication failure" {
		t.Fatalf("unexpected error message: %s", rsp.Error)
	}

	// 2. SetBlockMode returns error
	req2 := &CommandMessage{
		ID:   51,
		Type: TypeSetBlockMode,
	}
	if err := enc.Encode(req2); err != nil {
		t.Fatal(err)
	}

	if err := dec.Decode(&sMsg); err != nil {
		t.Fatal(err)
	}
	rsp2 := sMsg.Payload.(*ResponseMessage)
	if rsp2.ID != 51 || rsp2.Success {
		t.Fatalf("expected failed response, got: %+v", rsp2)
	}
	if rsp2.Error != "device not configured for block mode" {
		t.Fatalf("unexpected error message: %s", rsp2.Error)
	}
}
