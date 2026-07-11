package scpi

type (
	ChType int
)

const (
	Ch1 = ChType(1)
	Ch2 = ChType(2)
)

type Config struct {
	Port   string // e.g. "/dev/ttyUSB0"
	UsbVid string
	UsbPid string
}

// GeneratorIface is the common interface satisfied by both the real USB-backed
// Generator and the MockGenerator used in simulator mode.
type GeneratorIface interface {
	Open() error
	Close() error
	SetFrequency(ch ChType, freq float64) error
	SetAmplitude(ch ChType, amp float64) error
	SetOffset(ch ChType, offset float64) error
	SetPhase(ch ChType, phase float64) error
	SetOutput(ch ChType, on bool) error
	SetWaveform(ch ChType, wave string) error
	SetRampSymmetry(ch ChType, wave string) error
	SetImpedance(ch ChType, val string) error
	Query(s string) error
	QueryString(s string) (string, error)
	GetVidPid() (string, string)
	Send(s string) error
}
