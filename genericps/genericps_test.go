package genericps

import (
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockScopeHandler is a mock implementation of ScopeHandler for testing.
type MockScopeHandler struct {
	EnumerateUnitsFunc        func(bufferLen int16) (count int16, serials string, serialLth int16, err error)
	OpenUnitFunc              func(serial string, resolution int) (handle int16, err error)
	OpenUnitAsyncFunc         func(serial string, resolution int) (status int16, err error)
	OpenUnitProgressFunc      func() (retHandle, progressPercent, complete int16, err error)
	DispatchFunc              func(msg Message)
	IdVal                     string
	mutex                     sync.Mutex // Protects access to call counts
	EnumerateUnitsCallCount   int
	OpenUnitCallCount         int
	OpenUnitAsyncCallCount    int
	OpenUnitProgressCallCount int
	DispatchCallCount         int
}

func (m *MockScopeHandler) EnumerateUnits(bufferLen int16) (count int16, serials string, serialLth int16, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.EnumerateUnitsCallCount++
	if m.EnumerateUnitsFunc != nil {
		return m.EnumerateUnitsFunc(bufferLen)
	}
	return 0, "", 0, nil
}

func (m *MockScopeHandler) OpenUnit(serial string, resolution int) (handle int16, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.OpenUnitCallCount++
	if m.OpenUnitFunc != nil {
		return m.OpenUnitFunc(serial, resolution)
	}
	return 0, nil
}

func (m *MockScopeHandler) OpenUnitAsync(serial string, resolution int) (status int16, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.OpenUnitAsyncCallCount++
	if m.OpenUnitAsyncFunc != nil {
		return m.OpenUnitAsyncFunc(serial, resolution)
	}
	return 0, nil
}

func (m *MockScopeHandler) OpenUnitProgress() (retHandle, progressPercent, complete int16, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.OpenUnitProgressCallCount++
	if m.OpenUnitProgressFunc != nil {
		return m.OpenUnitProgressFunc()
	}
	return 0, 0, 0, nil
}

func (m *MockScopeHandler) Dispatch(msg Message) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.DispatchCallCount++
	if m.DispatchFunc != nil {
		m.DispatchFunc(msg)
	}
}

func (m *MockScopeHandler) Id() string {
	return m.IdVal
}

func TestRegister(t *testing.T) {
	// Arrange
	handler := ScopeHandler{Id: "test"}
	initialLength := len(implementedScopeHandlers)

	// Act
	Register(handler)

	// Assert
	if len(implementedScopeHandlers) != initialLength+1 {
		t.Errorf("Register() did not add a handler. Expected length: %d, Got: %d", initialLength+1, len(implementedScopeHandlers))
	}
	if implementedScopeHandlers[len(implementedScopeHandlers)-1].Id != "test" {
		t.Errorf("Register() did not add the correct handler.")
	}

	// Cleanup (optional, if you want to keep tests isolated)
	implementedScopeHandlers = implementedScopeHandlers[:initialLength]
}

func TestOpen_Success(t *testing.T) {
	// Arrange
	mockHandler := &MockScopeHandler{
		IdVal: "test_success",
		OpenUnitFunc: func(serial string, resolution int) (int16, error) {
			return 123, nil
		},
	}
	Register(mockHandler.ScopeHandler())
	defer func() { implementedScopeHandlers = []ScopeHandler{} }()

	// Act
	handle, err := Open("test_success", 0)

	// Assert
	if err != nil {
		t.Errorf("Open() failed: %v", err)
	}
	if handle != 123 {
		t.Errorf("Open() returned wrong handle. Expected: 123, Got: %d", handle)
	}

	mockHandler.mutex.Lock()
	defer mockHandler.mutex.Unlock()

	if mockHandler.OpenUnitCallCount != 1 {
		t.Errorf("OpenUnit() should be called once")
	}
}

func TestOpen_ScopeNotFound(t *testing.T) {
	// Arrange
	defer func() { implementedScopeHandlers = []ScopeHandler{} }()

	// Act
	_, err := Open("nonexistent", 0)

	// Assert
	if err == nil {
		t.Error("Open() should have returned an error.")
	}
	if err.Error() != "Scope not found" {
		t.Errorf("Open() returned wrong error message. Expected: Scope not found, Got: %s", err.Error())
	}
}
func (m *MockScopeHandler) ScopeHandler() ScopeHandler {
	return ScopeHandler{
		EnumerateUnits:   m.EnumerateUnits,
		OpenUnit:         m.OpenUnit,
		OpenUnitAsync:    m.OpenUnitAsync,
		OpenUnitProgress: m.OpenUnitProgress,
		Dispatch:         m.Dispatch,
		Id:               m.IdVal,
	}
}

func TestOpenDemo_Success(t *testing.T) {
	// Arrange
	mockHandler := &MockScopeHandler{
		IdVal: DemoId,
		OpenUnitFunc: func(serial string, resolution int) (int16, error) {
			return 456, nil
		},
		DispatchFunc: func(msg Message) {
			// Dummy dispatch implementation for testing
			msg.SetStatus(nil)
			msg.RspCh() <- struct{}{}
		},
	}
	Register(mockHandler.ScopeHandler())
	defer func() { implementedScopeHandlers = []ScopeHandler{} }()

	con := NewConnection()

	// Act
	handle, err := OpenDemo(con, DemoId)

	// Assert
	if err != nil {
		t.Fatalf("OpenDemo() failed: %v", err)
	}
	if handle != 456 {
		t.Errorf("OpenDemo() returned wrong handle. Expected: 456, Got: %d", handle)
	}

	// mockHandler.mutex.Lock()
	// defer mockHandler.mutex.Unlock()

	if mockHandler.OpenUnitCallCount != 1 {
		t.Errorf("OpenUnit() was not called once, called %d times", mockHandler.OpenUnitCallCount)
	}
	if mockHandler.DispatchCallCount != 0 {
		t.Errorf("Dispatch() was called, but should not called at this point, called %d times", mockHandler.DispatchCallCount)
	}

	//Act2 send a message to the opened connection.
	mockMsg := NullMsg{}
	mockMsg.rsp = &NullRsp{}
	con.Send(&mockMsg)

	//Assert2
	mockHandler.mutex.Lock()
	defer mockHandler.mutex.Unlock()
	if mockHandler.DispatchCallCount != 1 {
		t.Errorf("Dispatch() was not called after sending message, should be called once, but called %d times", mockHandler.DispatchCallCount)
	}

}

func TestOpenDemo_DemoNotFound(t *testing.T) {
	// Arrange
	defer func() { implementedScopeHandlers = []ScopeHandler{} }()
	con := NewConnection()
	// Act
	_, err := OpenDemo(con, DemoId)

	// Assert
	if err == nil {
		t.Error("OpenDemo() should have returned an error.")
	}
	if err.Error() != "Demo not found" {
		t.Errorf("OpenDemo() returned wrong error message. Expected: Demo not found, Got: %s", err.Error())
	}
}
func TestOpenUnit_Success(t *testing.T) {
	// Arrange
	mockHandler := &MockScopeHandler{
		IdVal: "mock_scope",
		OpenUnitFunc: func(serial string, resolution int) (int16, error) {
			return 789, nil
		},
		DispatchFunc: func(msg Message) {
			// Dummy dispatch implementation for testing
			msg.SetStatus(nil)
			msg.RspCh() <- struct{}{}
		},
	}
	Register(mockHandler.ScopeHandler())
	defer func() { implementedScopeHandlers = []ScopeHandler{} }()

	con := NewConnection()

	// Act
	handle, err := OpenUnit(con, "mock_scope", "", 0)

	// Assert
	if err != nil {
		t.Fatalf("OpenUnit() failed: %v", err)
	}
	if handle != 789 {
		t.Errorf("OpenUnit() returned wrong handle. Expected: 789, Got: %d", handle)
	}

	if mockHandler.OpenUnitCallCount != 1 {
		t.Errorf("OpenUnit() should be called once")
	}

	//Act2 send a message to the opened connection.
	mockMsg := NullMsg{}
	mockMsg.rsp = &NullRsp{}
	con.Send(&mockMsg)

	//Assert2
	mockHandler.mutex.Lock()
	defer mockHandler.mutex.Unlock()
	if mockHandler.DispatchCallCount != 1 {
		t.Errorf("Dispatch() was not called after sending message, should be called once")
	}
}

func TestOpenUnit_DemoNotFound(t *testing.T) {
	// Arrange
	defer func() { implementedScopeHandlers = []ScopeHandler{} }()
	con := NewConnection()

	// Act
	_, err := OpenUnit(con, "nonexistent_id", "", 0)

	// Assert
	if err == nil {
		t.Error("OpenUnit() should have returned an error.")
	}
	if err.Error() != "Device not found" {
		t.Errorf("OpenUnit() returned wrong error message. Expected: Device not found, Got: %s", err.Error())
	}
}
func TestEnumerateUnits_Success(t *testing.T) {
	// Arrange
	mockHandler := &MockScopeHandler{
		IdVal: DemoId,
		EnumerateUnitsFunc: func(bufferLen int16) (count int16, serials string, serialLth int16, err error) {
			return 2, "serial1,serial2", 15, nil
		},
	}
	Register(mockHandler.ScopeHandler())
	defer func() { implementedScopeHandlers = []ScopeHandler{} }()

	// Act
	count, serials, serialLth, err := EnumerateUnits(100)

	// Assert
	if err != nil {
		t.Errorf("EnumerateUnits() failed: %v", err)
	}
	if count != 2 {
		t.Errorf("EnumerateUnits() returned wrong count. Expected: 2, Got: %d", count)
	}
	if serials != "serial1,serial2" {
		t.Errorf("EnumerateUnits() returned wrong serials. Expected: serial1,serial2, Got: %s", serials)
	}
	if serialLth != 15 {
		t.Errorf("EnumerateUnits() returned wrong serialLth. Expected: 15, Got: %d", serialLth)
	}
	mockHandler.mutex.Lock()
	defer mockHandler.mutex.Unlock()
	if mockHandler.EnumerateUnitsCallCount != 1 {
		t.Errorf("EnumerateUnits should be called once, got: %d", mockHandler.EnumerateUnitsCallCount)
	}
}

func TestEnumerateUnits_DemoNotFound(t *testing.T) {
	// Arrange
	defer func() { implementedScopeHandlers = []ScopeHandler{} }()
	// Act
	_, _, _, err := EnumerateUnits(100)

	// Assert
	if err != nil {
		t.Fatalf("EnumerateUnits should not return with error, got: %v", err)
	}
}

func TestEnumerateUnits_Error(t *testing.T) {
	// Arrange
	mockHandler := &MockScopeHandler{
		IdVal: DemoId,
		EnumerateUnitsFunc: func(bufferLen int16) (count int16, serials string, serialLth int16, err error) {
			return 0, "", 0, errors.New("enumerate error")
		},
	}
	Register(mockHandler.ScopeHandler())
	defer func() { implementedScopeHandlers = []ScopeHandler{} }()
	// Act
	_, _, _, err := EnumerateUnits(100)

	// Assert
	if err == nil {
		t.Fatalf("EnumerateUnits should return with error")
	}
	if err.Error() != "enumerate error" {
		t.Fatalf("EnumerateUnits returned with wrong error")
	}
}

func TestUnRegister(t *testing.T) {
	initialLength := len(implementedScopeHandlers)
	handler1 := ScopeHandler{Id: "test1"}
	handler2 := ScopeHandler{Id: "test2"}
	Register(handler1)
	Register(handler2)
	assert.Equal(t, initialLength+2, len(implementedScopeHandlers))

	UnRegister("test1")
	assert.Equal(t, initialLength+1, len(implementedScopeHandlers))
	assert.Equal(t, "test2", implementedScopeHandlers[len(implementedScopeHandlers)-1].Id)

	UnRegister("nonexistent")
	assert.Equal(t, initialLength+1, len(implementedScopeHandlers))

	UnRegister("test2")
	assert.Equal(t, initialLength, len(implementedScopeHandlers))
}

func TestParseSerials(t *testing.T) {
	assert.Empty(t, parseSerials("", 0))
	assert.Empty(t, parseSerials("ABC", 0))

	assert.Equal(t, []string{"ABC"}, parseSerials("ABC", 1))
	assert.Equal(t, []string{"ABC", "DEF"}, parseSerials("ABC,DEF", 2))
	assert.Equal(t, []string{"A", "B", "C"}, parseSerials("A,B,C,", 3))
}

func TestEnumerateAllDevices(t *testing.T) {
	defer func() { implementedScopeHandlers = []ScopeHandler{} }()
	implementedScopeHandlers = []ScopeHandler{}

	// No handlers
	devices, err := EnumerateAllDevices(100)
	assert.Error(t, err)
	assert.Empty(t, devices)

	// Add a handler with error
	mockErr := &MockScopeHandler{
		IdVal: "errHandler",
		EnumerateUnitsFunc: func(bufferLen int16) (int16, string, int16, error) {
			return 0, "", 0, errors.New("enum failed")
		},
	}
	Register(mockErr.ScopeHandler())
	devices, err = EnumerateAllDevices(100)
	assert.Error(t, err)

	// Add a handler with devices
	mockOk := &MockScopeHandler{
		IdVal: "ps2000a",
		EnumerateUnitsFunc: func(bufferLen int16) (int16, string, int16, error) {
			return 2, "SN1,SN2", 7, nil
		},
	}
	Register(mockOk.ScopeHandler())
	devices, err = EnumerateAllDevices(100)
	assert.NoError(t, err)
	assert.Len(t, devices, 2)
	assert.Equal(t, "ps2000a", devices[0].Id)
	assert.Equal(t, "SN1", devices[0].Serial)
	assert.False(t, devices[0].IsDemo)

	// Add demo handler
	mockDemo := &MockScopeHandler{
		IdVal: DemoId,
		EnumerateUnitsFunc: func(bufferLen int16) (int16, string, int16, error) {
			return 1, "DEMO_1", 6, nil
		},
	}
	Register(mockDemo.ScopeHandler())
	devices, err = EnumerateAllDevices(100)
	assert.NoError(t, err)
	assert.Len(t, devices, 3)
	assert.True(t, devices[2].IsDemo)
}

func TestTimeUnitToVal(t *testing.T) {
	assert.Equal(t, 1e-15, TimeUnitToVal(TuFs))
	assert.Equal(t, 0.0, TimeUnitToVal(TimeUnits(999)))
}

func TestGetMinThresholdDiff(t *testing.T) {
	RangeValuesMv = map[RangeEnum]float64{
		RangeEnum(1): 100.0,
		RangeEnum(2): 1000.0,
	}
	MinThresholdDiff = 200

	assert.Equal(t, int32(200), GetMinThresholdDiff(RangeEnum(-1)))
	assert.Equal(t, int32(200), GetMinThresholdDiff(RangeEnum(999)))
	assert.Equal(t, int32(5), GetMinThresholdDiff(RangeEnum(1)))  // 100 * 0.05 = 5
	assert.Equal(t, int32(50), GetMinThresholdDiff(RangeEnum(2))) // 1000 * 0.05 = 50
}

func TestConnectionWrappers(t *testing.T) {
	con := NewConnection()
	con.Handle = 42

	// Launch responder goroutine
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case msg := <-con.MsgCh:
				if msg != nil && msg.RspCh() != nil {
					msg.RspCh() <- struct{}{}
				}
			case <-quit:
				return
			}
		}
	}()
	defer close(quit)

	// Test getters/setters on message
	msg := &CloseUnitMsg{}
	msg.rsp = &CloseUnitRsp{}
	msg.SetHandle(10)
	assert.Equal(t, int16(10), msg.Handle())
	msg.SetStatus(nil)
	assert.NoError(t, msg.Status())
	rspCh := make(chan struct{}, 1)
	msg.SetRspCh(rspCh)
	assert.Equal(t, rspCh, msg.RspCh())

	// Test all Connection command wrappers
	_ = con.CloseUnit()
	_ = con.FlashLed(1)
	_ = con.PingUnit()
	_, _, _ = con.GetAnalogueOffset(0, 0)
	_, _ = con.GetChannelInformation(0, 0, nil, 0)
	_, _ = con.GetMaxDownSampleRatio(100, 0, 0)
	_, _ = con.GetMaxSegments()
	_, _ = con.GetNumOfCaptures()
	_, _ = con.GetNumOfProcessedCaptures()
	_ = con.GetStreamingLatestValues(nil, nil)
	_, _, _ = con.GetTimebase(0, 100, 0, 0)
	_, _, _ = con.GetTimebase2(0, 100, 0, 0)
	_, _, _, _ = con.GetTriggerTimeOffset(0)
	_, _, _ = con.GetTriggerTimeOffset64(0)
	_, _ = con.GetUnitInfo(0)
	_, _, _ = con.GetValues(0, 100, 1, 0, 0)
	_ = con.GetValuesAsync(0, 100, 1, 0, nil, 0, nil)
	_, _ = con.GetValuesBulk(0, 0, 0, 1, 0, nil)
	_, _ = con.GetValuesOverlapped(0, 0, 1, 0, 0, nil)
	_, _ = con.GetValuesOverlappedBulk(0, 0, 1, 0, 0, 0, nil)
	_ = con.GetValuesTriggerTimeOffsetBulk(nil, nil, nil, 0, 0)
	_ = con.GetValuesTriggerTimeOffsetBulk64(nil, nil, 0, 0)
	_ = con.HoldOff(0, 0)
	_, _ = con.LsReady()
	_, _ = con.MaximumValue()
	_, _ = con.MemorySegments(1)
	_, _ = con.MinimumValue()
	_, _ = con.NumOfStreamingValues()
	_, _ = con.QueryOutputEdgeDetect()
	_, _ = con.RunBlock(0, 100, 1, 0, 0, nil, nil)
	_, _ = con.RunStreaming(0, 0, 100, 100, false, 1, 0, 100)
	_ = con.SetChannel(0, true, 0, 0, 0)
	_ = con.SetDataBuffer(0, nil, 0, 0)
	_ = con.SetDataBuffers(0, nil, nil, 0, 0)
	_ = con.SetUnscaledDataBuffers(0, nil, nil, 0, 0)
	_ = con.SetDigitalAnalogTriggerOperand(0)
	_ = con.SetDigitalPort(0, true, 0)
	_, _ = con.SetEts(0, 0, 0)
	_ = con.SetEtsTimeBuffer(nil)
	_ = con.SetEtsTimeBuffers(nil, nil)
	_ = con.SetNoCaptures(1)
	_ = con.SetOutputEdgeDetect(0)
	_ = con.SetPulseWidthDigitalPortProperties(nil)
	_ = con.SetPulseWidthQualifier(nil, 0, 0, 0, 0)
	_ = con.SetSigGenArbitrary(0, 0, 0, 0, 0, 0, nil, 0, 0, 0, 0, 0, 0, 0, 0)
	_, _, _, _, _ = con.SigGenArbitraryMinMaxValues()
	_ = con.SetSigGenBuiltIn(0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)
	_ = con.SetSigGenBuiltInV2(0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)
	_ = con.SetDemoGen(0, true, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, nil, 0, 0)
	_ = con.SetDemoDigitalGen(false, false, 1000, 0, 0, 0, 0)
	_ = con.SetDemoRlcFilter(0, 0, false, "", 0, "", 0, "", 0, "")
	_ = con.SetSigGenPropertiesArbitrary(0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)
	_ = con.SetSigGenPropertiesBuiltIn(0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)
	_ = con.SetSimpleTrigger(false, 0, 0, 0, 0, 0)
	_ = con.SetTriggerChannelConditions(nil)
	_ = con.SetTriggerChannelDirections(0, 0, 0, 0, 0, 0)
	_ = con.SetTriggerChannelProperties(nil, false, 0)
	_ = con.SetTriggerDelay(0)
	_ = con.SetTriggerDigitalPortProperties(nil)
	_, _ = con.SigGenFrequencyToPhase(0, 0, 0)
	_ = con.Stop()
	_, _, _ = con.TriggerOrPulseWidthQualifierEnabled()
	_ = con.SigGenSoftwareControl(0)
	_ = OpenUnitAsync("123")
	_, _, _, _ = OpenUnitProgress()
}
