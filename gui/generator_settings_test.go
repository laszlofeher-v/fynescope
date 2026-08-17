package gui

import (
	"fynescope/control"
	"fynescope/genericps"
	"fynescope/settings"
	"testing"

	"fyne.io/fyne/v2/container"
)

func TestSetGeneratorFreq(t *testing.T) {
	scp := &ScpDesc{
		Settings: &settings.PsSettings{
			FfGen: settings.GeneratorSettings{
				On:            true,
				Amplitude:     1000,
				OffsetVoltage: 0,
			},
			GenPanel: settings.GeneratorSettings{
				WaveType:      genericps.Sine,
				OffsetVoltage: 0,
				Amplitude:     1000,
			},
		},
		psControl: &control.PscDesc{
			SetGeneratorCh: make(chan *control.GeneratorDescMsg, 10),
			SetDemoGenCh:   make(chan *control.GeneratorDescMsg, 10),
		},
		controlTab: container.NewAppTabs(container.NewTabItem("Test", container.NewVBox())),
	}
	
	// Test normal flow
	scp.setGeneratorFreq(1000)
	
	// Check that a message was sent to SetGeneratorCh
	select {
	case <-scp.psControl.SetGeneratorCh:
		// success
	default:
		t.Error("expected message on SetGeneratorCh")
	}
}

func TestApplyFfGenSettings(t *testing.T) {
	scp := &ScpDesc{
		Settings: &settings.PsSettings{
			Ff: settings.FfSettings{
				MinFreq: 10,
				MaxFreq: 1000,
				DeltaT:  10,
			},
			FfGen: settings.GeneratorSettings{
				Amplitude:     1000,
				OffsetVoltage: 0,
			},
		},
		psControl: &control.PscDesc{
			SetGeneratorCh: make(chan *control.GeneratorDescMsg, 10),
		},
	}
	
	// Turn on
	scp.applyFfGenSettings(true)
	// Turn off
	scp.applyFfGenSettings(false)
}

func TestApplyFfDemoGenSettings(t *testing.T) {
	scp := &ScpDesc{
		channelCount: 2,
		Settings: &settings.PsSettings{
			Channels: []settings.ChSettings{
				{Enabled: true, RlcFilter: settings.RlcFilterSettings{GeneratorSource: 0}},
				{Enabled: true, RlcFilter: settings.RlcFilterSettings{GeneratorSource: 1}},
			},
			DemoGenPanel: []settings.GeneratorSettings{
				{On: true, WaveType: genericps.Sine, Amplitude: 1000},
				{On: false, WaveType: genericps.Square, Amplitude: 1000},
			},
			Ff: settings.FfSettings{
				MinFreq: 10,
				MaxFreq: 1000,
				DeltaT:  10,
			},
			FfGen: settings.GeneratorSettings{
				Amplitude:     1000,
				OffsetVoltage: 0,
			},
		},
		psControl: &control.PscDesc{
			SetDemoGenCh: make(chan *control.GeneratorDescMsg, 10),
		},
	}
	
	// Turn on
	scp.applyFfDemoGenSettings(true)
	// Turn off
	scp.applyFfDemoGenSettings(false)
}
