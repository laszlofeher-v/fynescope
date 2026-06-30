package scpi

import "log/slog"

// MockGenerator implements GeneratorIface for use when the simulator is active.
// All methods succeed silently and log the call at Debug level.
type MockGenerator struct{}

// NewMockGenerator returns a MockGenerator ready for use.
func NewMockGenerator() GeneratorIface {
	return &MockGenerator{}
}

func (m *MockGenerator) Open() error {
	slog.Debug("MockGenerator.Open")
	return nil
}

func (m *MockGenerator) Close() error {
	slog.Debug("MockGenerator.Close")
	return nil
}

func (m *MockGenerator) SetFrequency(ch ChType, freq float64) error {
	slog.Debug("MockGenerator.SetFrequency", "ch", ch, "freq", freq)
	return nil
}

func (m *MockGenerator) SetAmplitude(ch ChType, amp float64) error {
	slog.Debug("MockGenerator.SetAmplitude", "ch", ch, "amp", amp)
	return nil
}

func (m *MockGenerator) SetOffset(ch ChType, offset float64) error {
	slog.Debug("MockGenerator.SetOffset", "ch", ch, "offset", offset)
	return nil
}

func (m *MockGenerator) SetPhase(ch ChType, phase float64) error {
	slog.Debug("MockGenerator.SetPhase", "ch", ch, "phase", phase)
	return nil
}

func (m *MockGenerator) SetOutput(ch ChType, on bool) error {
	slog.Debug("MockGenerator.SetOutput", "ch", ch, "on", on)
	return nil
}

func (m *MockGenerator) SetWaveform(ch ChType, wave string) error {
	slog.Debug("MockGenerator.SetWaveform", "ch", ch, "wave", wave)
	return nil
}
func (m *MockGenerator) SetRampSymmetry(ch ChType, wave string) error {
	slog.Debug("MockGenerator.SetWaveform", "ch", ch, "wave", wave)
	return nil
}

func (m *MockGenerator) SetImpedance(ch ChType, val string) error {
	slog.Debug("MockGenerator.SetImpedance", "ch", ch, "val", val)
	return nil
}

func (m *MockGenerator) Query(s string) error {
	slog.Debug("MockGenerator.Query", "cmd", s)
	return nil
}

func (m *MockGenerator) QueryString(s string) (string, error) {
	slog.Debug("MockGenerator.QueryString", "cmd", s)
	return "MockGenerator", nil
}

func (m *MockGenerator) GetVidPid() (string, string) {
	return "0x0000", "0x0000"
}

func (m *MockGenerator) Send(s string) error {
	slog.Debug("MockGenerator.Send", "cmd", s)
	return nil
}
