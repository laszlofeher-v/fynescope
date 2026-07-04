package control

import (
	"fmt"
	"fynescope/control/scpi"
)

type ExtGenConfig struct {
	Port   string
	UsbVid string
	UsbPid string
}

// ExtGenDesc manages the lifecycle and commands for an external SCPI signal generator.
// The zero value is a valid, disconnected descriptor.
type ExtGenDesc struct {
	gen scpi.GeneratorIface
}

// Connect creates and opens the external generator described by cfg.
// Any previously open connection is closed first.
func (e *ExtGenDesc) Connect(cfg ExtGenConfig) error {
	// Close previous connection.
	if e.gen != nil {
		e.gen.Close()
		e.gen = nil
	}

	scpiCfg := scpi.Config{
		Port:   cfg.Port,
		UsbVid: cfg.UsbVid,
		UsbPid: cfg.UsbPid,
	}

	gen := scpi.New(scpiCfg)
	if err := gen.Open(); err != nil {
		return fmt.Errorf("open external generator: %w", err)
	}

	e.gen = gen
	return nil
}

// Disconnect closes the connection to the external generator.
func (e *ExtGenDesc) Disconnect() {
	if e.gen != nil {
		e.gen.Close()
		e.gen = nil
	}
}

// Connected reports whether a generator is currently open and ready.
func (e *ExtGenDesc) Connected() bool {
	return e.gen != nil
}

// GetVidPid returns the connected USB VID and PID, if available.
func (e *ExtGenDesc) GetVidPid() (string, string) {
	if e.gen != nil {
		return e.gen.GetVidPid()
	}
	return "", ""
}

// SetFrequency sets the output frequency on the given channel.
func (e *ExtGenDesc) SetFrequency(ch scpi.ChType, freq float64) error {
	if e.gen == nil {
		return fmt.Errorf("external generator is not connected")
	}
	return e.gen.SetFrequency(ch, freq)
}

// SetAmplitude sets the peak-to-peak output amplitude (in volts).
func (e *ExtGenDesc) SetAmplitude(ch scpi.ChType, amp float64) error {
	if e.gen == nil {
		return fmt.Errorf("external generator is not connected")
	}
	return e.gen.SetAmplitude(ch, amp)
}

// SetOffset sets the DC offset voltage (in volts).
func (e *ExtGenDesc) SetOffset(ch scpi.ChType, offset float64) error {
	if e.gen == nil {
		return fmt.Errorf("external generator is not connected")
	}
	return e.gen.SetOffset(ch, offset)
}

func (e *ExtGenDesc) SetPhase(ch scpi.ChType, phase float64) error {
	if e.gen == nil {
		return fmt.Errorf("external generator is not connected")
	}
	return e.gen.SetPhase(ch, phase)
}

// SetOutput enables or disables the signal output on the given channel.
func (e *ExtGenDesc) SetOutput(ch scpi.ChType, on bool) error {
	if e.gen == nil {
		return fmt.Errorf("external generator is not connected")
	}
	return e.gen.SetOutput(ch, on)
}

// SetWaveform sets the waveform type for simple shapes (e.g. "SINusoid", "SQUare").
func (e *ExtGenDesc) SetWaveform(ch scpi.ChType, wave string) error {
	if e.gen == nil {
		return fmt.Errorf("external generator is not connected")
	}
	return e.gen.SetWaveform(ch, wave)
}

// SetRampSymmetry sets the symmetry
func (e *ExtGenDesc) SetRampSymmetry(ch scpi.ChType, wave string) error {
	if e.gen == nil {
		return fmt.Errorf("external generator is not connected")
	}
	return e.gen.SetRampSymmetry(ch, wave)
}

// SetImpedance sets the output load impedance.
// val should be a numeric string (e.g. "50"), "INFinity", "MINimum", or "MAXimum".
func (e *ExtGenDesc) SetImpedance(ch scpi.ChType, val string) error {
	if e.gen == nil {
		return fmt.Errorf("external generator is not connected")
	}
	return e.gen.SetImpedance(ch, val)
}

// SendRaw sends a raw SCPI command string to the external generator.
func (e *ExtGenDesc) SendRaw(cmd string) error {
	if e.gen == nil {
		return fmt.Errorf("external generator is not connected")
	}
	return e.gen.Send(cmd)
}
