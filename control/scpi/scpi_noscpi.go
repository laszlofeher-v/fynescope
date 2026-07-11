//go:build !scpi

package scpi

import "fmt"

func New(cfg Config) GeneratorIface {
	return &dummyGen{}
}

type dummyGen struct{}

func (d *dummyGen) Open() error {
	return fmt.Errorf("scpi module not included in this build")
}

func (d *dummyGen) Close() error { return nil }

func (d *dummyGen) SetFrequency(ch ChType, freq float64) error { return nil }

func (d *dummyGen) SetAmplitude(ch ChType, amp float64) error { return nil }

func (d *dummyGen) SetOffset(ch ChType, offset float64) error { return nil }

func (d *dummyGen) SetPhase(ch ChType, phase float64) error { return nil }

func (d *dummyGen) SetOutput(ch ChType, on bool) error { return nil }

func (d *dummyGen) SetWaveform(ch ChType, wave string) error { return nil }

func (d *dummyGen) SetRampSymmetry(ch ChType, wave string) error { return nil }

func (d *dummyGen) SetImpedance(ch ChType, val string) error { return nil }

func (d *dummyGen) Query(s string) error { return nil }

func (d *dummyGen) QueryString(s string) (string, error) { return "", nil }

func (d *dummyGen) GetVidPid() (string, string) { return "", "" }

func (d *dummyGen) Send(s string) error { return nil }
