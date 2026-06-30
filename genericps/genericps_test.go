package genericps

import (
	"errors"
	"sync"
	"testing"
)

// MockScopeHandler is a mock implementation of ScopeHandler for testing.
type MockScopeHandler struct {
	EnumerateUnitsFunc        func(bufferLen int16) (count int16, serials string, serialLth int16, err error)
	OpenUnitFunc              func(serial string) (handle int16, err error)
	OpenUnitAsyncFunc         func(serial string) (status int16, err error)
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

func (m *MockScopeHandler) OpenUnit(serial string) (handle int16, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.OpenUnitCallCount++
	if m.OpenUnitFunc != nil {
		return m.OpenUnitFunc(serial)
	}
	return 0, nil
}

func (m *MockScopeHandler) OpenUnitAsync(serial string) (status int16, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.OpenUnitAsyncCallCount++
	if m.OpenUnitAsyncFunc != nil {
		return m.OpenUnitAsyncFunc(serial)
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
		OpenUnitFunc: func(serial string) (int16, error) {
			return 123, nil
		},
	}
	Register(mockHandler.ScopeHandler())
	defer func() { implementedScopeHandlers = []ScopeHandler{} }()

	// Act
	handle, err := Open("test_success")

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
	_, err := Open("nonexistent")

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

func TestOpenSimulator_Success(t *testing.T) {
	// Arrange
	mockHandler := &MockScopeHandler{
		IdVal: SimId,
		OpenUnitFunc: func(serial string) (int16, error) {
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
	handle, err := OpenSimulator(con, SimId)

	// Assert
	if err != nil {
		t.Fatalf("OpenSimulator() failed: %v", err)
	}
	if handle != 456 {
		t.Errorf("OpenSimulator() returned wrong handle. Expected: 456, Got: %d", handle)
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

func TestOpenSimulator_SimulatorNotFound(t *testing.T) {
	// Arrange
	defer func() { implementedScopeHandlers = []ScopeHandler{} }()
	con := NewConnection()
	// Act
	_, err := OpenSimulator(con, SimId)

	// Assert
	if err == nil {
		t.Error("OpenSimulator() should have returned an error.")
	}
	if err.Error() != "Simulator not found" {
		t.Errorf("OpenSimulator() returned wrong error message. Expected: Simulator not found, Got: %s", err.Error())
	}
}
func TestOpenUnit_Success(t *testing.T) {
	// Arrange
	mockHandler := &MockScopeHandler{
		IdVal: "mock_scope",
		OpenUnitFunc: func(serial string) (int16, error) {
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
	handle, err := OpenUnit(con, "mock_scope", "")

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

func TestOpenUnit_SimulatorNotFound(t *testing.T) {
	// Arrange
	defer func() { implementedScopeHandlers = []ScopeHandler{} }()
	con := NewConnection()

	// Act
	_, err := OpenUnit(con, "nonexistent_id", "")

	// Assert
	if err == nil {
		t.Error("OpenUnit() should have returned an error.")
	}
	if err.Error() != "Simulator not found" {
		t.Errorf("OpenUnit() returned wrong error message. Expected: Simulator not found, Got: %s", err.Error())
	}
}
func TestEnumerateUnits_Success(t *testing.T) {
	// Arrange
	mockHandler := &MockScopeHandler{
		IdVal: SimId,
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

func TestEnumerateUnits_SimulatorNotFound(t *testing.T) {
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
		IdVal: SimId,
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
