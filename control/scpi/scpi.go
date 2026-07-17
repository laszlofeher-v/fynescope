//go:build scpi

package scpi

import (
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/google/gousb"
)

type Generator struct {
	cfg         Config
	ctx         *gousb.Context
	dev         *gousb.Device
	intfDone    func()
	outEndpoint *gousb.OutEndpoint
	inEndpoint  *gousb.InEndpoint
	Writer      io.Writer // For testing
}

func New(cfg Config) GeneratorIface {
	return &Generator{cfg: cfg}
}


func (g *Generator) setupDevice(dev *gousb.Device) error {
	slog.Debug("Device found", "dev", dev)
	g.dev = dev
	// Ensure kernel driver detachment
	g.dev.SetAutoDetach(true)

	intf, done, err := g.dev.DefaultInterface()
	if err != nil {
		g.dev = nil
		return fmt.Errorf("failed to claim default USB interface: %w", err)
	}
	g.intfDone = done
	slog.Debug("Find bulk OUT endpoint")
	g.outEndpoint, err = intf.OutEndpoint(1)
	g.inEndpoint, err = intf.InEndpoint(1)

	if g.outEndpoint == nil {
		g.intfDone()
		g.intfDone = nil
		g.dev = nil
		return fmt.Errorf("no Bulk OUT endpoint found on USB device")
	}
	if g.inEndpoint == nil {
		slog.Warn("no Bulk IN endpoint found on USB device")
	}
	slog.Debug("endpoints", "out", g.outEndpoint, "in", g.inEndpoint)

	return nil
}

func (g *Generator) Open() error {
	var targetVid, targetPid gousb.ID

	if g.cfg.UsbVid != "" {
		if vid := parseHex(g.cfg.UsbVid); vid != 0 {
			targetVid = gousb.ID(vid)
		}
	}
	if g.cfg.UsbPid != "" {
		if pid := parseHex(g.cfg.UsbPid); pid != 0 {
			targetPid = gousb.ID(pid)
		}
	}

	g.ctx = gousb.NewContext()

	if targetVid != 0 || targetPid != 0 {
		dev, err := g.ctx.OpenDeviceWithVIDPID(targetVid, targetPid)
		if err != nil || dev == nil {
			return fmt.Errorf("Open err: %v", err)
		}
		return g.setupDevice(dev)
	}

	// Autodetection loop
	devs, err := g.ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		// Skip root hubs and standard hubs to avoid unnecessary access errors
		if desc.Class == gousb.ClassHub {
			return false
		}
		return true
	})
	if err != nil {
		slog.Warn("OpenDevices encountered errors (e.g. access denied on some devices)", "err", err)
	}

	var foundDev *gousb.Device
	for _, dev := range devs {
		if foundDev != nil {
			dev.Close()
			continue
		}

		err := g.setupDevice(dev)
		if err == nil {
			resp, err := g.QueryString("*IDN?")
			if err == nil && resp != "" {
				// Assuming any SCPI device that responds to *IDN? is what we want
				foundDev = dev
			} else {
				g.Close()
				dev.Close()
			}
		} else {
			dev.Close()
		}
	}

	if foundDev == nil {
		return fmt.Errorf("no suitable SCPI generator found during autodetection")
	}

	return nil
}

func (g *Generator) Close() error {
	var errs []string

	g.outEndpoint = nil
	g.inEndpoint = nil

	if g.intfDone != nil {
		g.intfDone()
		g.intfDone = nil
	}

	if g.dev != nil {
		if err := g.dev.Close(); err != nil {
			errs = append(errs, err.Error())
		}
		g.dev = nil
	}

	if g.ctx != nil {
		if err := g.ctx.Close(); err != nil {
			errs = append(errs, err.Error())
		}
		g.ctx = nil
	}

	if len(errs) > 0 {
		return fmt.Errorf("cleanup errors: %s", strings.Join(errs, "; "))
	}
	return nil
}

func (g *Generator) send(cmd string) error {
	if g.Writer != nil {
		_, err := g.Writer.Write([]byte(cmd + "\n"))
		return err
	}
	if g.outEndpoint == nil {
		return fmt.Errorf("generator is not open (outEndpoint is nil)")
	}
	l, err := g.outEndpoint.Write([]byte(cmd + "\n"))
	slog.Debug("SCPI write", "cmd", cmd, "len", l)
	if err != nil {
		return fmt.Errorf("write %s error:%w", cmd, err)
	}
	return nil
}

func (g *Generator) Send(s string) error {
	return g.send(s)
}

func (g *Generator) SetFrequency(ch ChType, freq float64) error {
	return g.send(fmt.Sprintf(":FREQuency%d %f", ch, freq))
}

func (g *Generator) Query(s string) error {
	_, err := g.QueryString(s)
	return err
}

func (g *Generator) QueryString(s string) (string, error) {
	err := g.send(s)
	if err != nil {
		return "", err
	}
	if g.inEndpoint == nil {
		if g.Writer != nil {
			return "", nil
		}
		return "", fmt.Errorf("generator is not open (inEndpoint is nil)")
	}
	buf := make([]byte, 512)
	slog.Debug("usb send", "inEndpoint", g.inEndpoint)
	n, err := g.inEndpoint.Read(buf)
	if err != nil {
		return "", err
	}
	if n > 0 {
		fmt.Printf("Data: %s\n", string(buf[:n]))
		return string(buf[:n]), nil
	}
	return "", nil
}

func (g *Generator) GetVidPid() (string, string) {
	if g.dev != nil {
		return fmt.Sprintf("0x%04x", g.dev.Desc.Vendor), fmt.Sprintf("0x%04x", g.dev.Desc.Product)
	}
	return g.cfg.UsbVid, g.cfg.UsbPid
}

func (g *Generator) SetAmplitude(ch ChType, amp float64) error {
	return g.send(fmt.Sprintf("SOURce%d:VOLTage:AMPLitude %f", ch, amp))
}

func (g *Generator) SetOffset(ch ChType, offset float64) error {
	return g.send(fmt.Sprintf("SOURce%d:VOLTage:OFFSet %f", ch, offset))
}

func (g *Generator) SetPhase(ch ChType, phase float64) error {
	return g.send(fmt.Sprintf("SOUR%d:PHAS %fDEG", ch, phase))
}

func (g *Generator) SetOutput(ch ChType, on bool) error {
	val := "OFF"
	if on {
		val = "ON"
	}
	return g.send(fmt.Sprintf("OUTPut%d:STATe %s", ch, val))
}

func (g *Generator) SetWaveform(ch ChType, wave string) error {
	// Common SCPI wave types: SIN, SQU, TRI
	return g.send(fmt.Sprintf("SOURce%d:FUNCtion:SHAPe %s", ch, wave))
}

// SetRampSymmetry uses a colon separator so that ramp sub-commands
// (e.g. "RAMP:SYMMetry 50") produce the correct form
// SOURce1:FUNCtion:RAMP:SYMMetry 50.
func (g *Generator) SetRampSymmetry(ch ChType, wave string) error {
	return g.send(fmt.Sprintf("SOURce%d:FUNCtion:%s", ch, wave))
}

// SetImpedance sets the output load impedance.
// val should be a numeric string (e.g. "50"), "INFinity", "MINimum", or "MAXimum".
func (g *Generator) SetImpedance(ch ChType, val string) error {
	return g.send(fmt.Sprintf("OUTPut%d:IMPedance %s", ch, val))
}
