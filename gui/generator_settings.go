package gui

import (
	"fynescope/control"
	"fynescope/control/scpi"
	"fynescope/genericps"
	"strings"
)

func (scp *ScpDesc) setGeneratorFreq(f float64) {
	if scp.psControl == nil {
		return
	}

	if scp.ExtGenEnabled && scp.Settings.Ff.UseExternalGen && scp.extGen.Connected() {
		scp.setExtGenFrequency(f)
		return
	}

	if scp.controlTab.SelectedIndex() == ffTabIndex {
		if scp.psControl.Con != nil && scp.psControl.Con.ID == genericps.SimId {
			// Simulator mode: only sinus wave, for generators mapped via RLC
			activeGens := make([]bool, scp.channelCount)
			var missingGenChannels []string

			if scp.Settings.FfGen.On {
				for i := 0; i < int(scp.channelCount); i++ {
					if scp.Settings.Channels[i].Enabled {
						genSrc := int(scp.Settings.Channels[i].RlcFilter.GeneratorSource)
						if genSrc >= 0 && genSrc < int(scp.channelCount) {
							if scp.Settings.SimGenPanel[genSrc].On {
								activeGens[genSrc] = true
							} else {
								missingGenChannels = append(missingGenChannels, channelNames[i])
							}
						}
					}
				}
			}

			if len(missingGenChannels) > 0 {
				scp.psControl.DisplayStatus("Error: Channel "+strings.Join(missingGenChannels, ", ")+" has no active generator input", control.Warning)
			} else if scp.status.Code() == StatusChannelNoActiveGen {
				scp.psControl.DisplayStatus("", control.Info)
			}

			for i := 0; i < int(scp.channelCount); i++ {
				if activeGens[i] {
					msg := &control.GeneratorDescMsg{
						GeneratorDesc: control.GeneratorDesc{
							StartFrequency: f,
							StopFrequency:  f,
							Increment:      0,
							DwellTime:      1,
							SweepType:      genericps.SweepDown,
							WaveType:       genericps.Sine,
							OffsetVoltage:  0,
							PkToPK:         scp.Settings.Ff.Amplitude * 2,
							Channel:        genericps.ChannelId(i),
							On:             true,
							Phase:          0,
						},
					}
					scp.psControl.SetSimGenCh <- msg
				} else {
					offMsg := &control.GeneratorDescMsg{
						GeneratorDesc: control.GeneratorDesc{
							Channel: genericps.ChannelId(i),
							On:      false,
						},
					}
					scp.psControl.SetSimGenCh <- offMsg
				}
			}
			return
		}

		msg := &control.GeneratorDescMsg{
			GeneratorDesc: control.GeneratorDesc{
				StartFrequency: f,
				StopFrequency:  f,
				Increment:      0,
				DwellTime:      1,
				SweepType:      genericps.SweepDown, // No sweep
				WaveType:       genericps.Sine,      // Force Sine wave for Bode plots
				OffsetVoltage:  scp.Settings.FfGen.OffsetVoltage,
				PkToPK:         scp.Settings.FfGen.Amplitude * 2,
				On:             scp.Settings.FfGen.On,
			},
		}
		scp.psControl.SetGeneratorCh <- msg
		return
	}

	if scp.psControl.Con != nil && scp.psControl.Con.ID == genericps.SimId {
		// Simulator mode: update all enabled simulator channels
		for i := 0; i < int(scp.channelCount); i++ {
			msg := &control.GeneratorDescMsg{
				GeneratorDesc: control.GeneratorDesc{
					StartFrequency: f,
					StopFrequency:  f,
					Increment:      0,
					DwellTime:      1,
					SweepType:      genericps.SweepDown,
					WaveType:       scp.Settings.SimGenPanel[i].WaveType,
					OffsetVoltage:  scp.Settings.SimGenPanel[i].OffsetVoltage,
					PkToPK:         scp.Settings.SimGenPanel[i].Amplitude * 2,
					Channel:        genericps.ChannelId(i),
					On:             true,
				},
			}
			scp.psControl.SetSimGenCh <- msg
		}
		return
	}

	msg := &control.GeneratorDescMsg{
		GeneratorDesc: control.GeneratorDesc{
			StartFrequency: f,
			StopFrequency:  f,
			Increment:      0,
			DwellTime:      1,
			SweepType:      genericps.SweepDown, // No sweep
			WaveType:       scp.Settings.GenPanel.WaveType,
			OffsetVoltage:  scp.Settings.GenPanel.OffsetVoltage,
			PkToPK:         scp.Settings.GenPanel.Amplitude * 2,
		},
	}
	scp.psControl.SetGeneratorCh <- msg
}

func (scp *ScpDesc) applyFfGenSettings(on bool) {
	if scp.ExtGenEnabled && scp.Settings.Ff.UseExternalGen && scp.extGen.Connected() {
		if on {
			scp.extGen.SetAmplitude(scpi.Ch1, float64(scp.Settings.FfGen.Amplitude)/1000000.0)
			scp.extGen.SetOffset(scpi.Ch1, float64(scp.Settings.FfGen.OffsetVoltage)/1000000.0)
			scp.extGen.SetWaveform(scpi.Ch1, "SINusoid")
			scp.extGen.SetOutput(scpi.Ch1, true)
		} else {
			scp.extGen.SetOutput(scpi.Ch1, false)
		}

		// Ensure internal generator is turned off
		msg := &control.GeneratorDescMsg{}
		msg.Operation = genericps.EsOff
		if scp.psControl != nil && scp.psControl.SetGeneratorCh != nil {
			scp.psControl.SetGeneratorCh <- msg
		}
		return
	}

	msg := &control.GeneratorDescMsg{}
	if on {
		msg.StartFrequency = scp.Settings.Ff.MinFreq
		msg.StopFrequency = scp.Settings.Ff.MaxFreq
		msg.Increment = 0 // App controls frequency stepping; no hardware sweep
		msg.DwellTime = scp.Settings.Ff.DeltaT
		msg.SweepType = genericps.SweepUp
		msg.WaveType = genericps.Sine
		msg.OffsetVoltage = scp.Settings.FfGen.OffsetVoltage
		msg.PkToPK = scp.Settings.FfGen.Amplitude * 2
	} else {
		msg.DwellTime = 0
		msg.OffsetVoltage = 0
		msg.PkToPK = 0
		msg.WaveType = genericps.DcVoltage
		msg.StopFrequency = scp.Settings.FfGen.StopFrequency
		msg.SweepType = genericps.SweepUp
	}
	msg.Operation = genericps.EsOff
	msg.Shots = 0
	msg.Sweeps = 0
	msg.TriggerType = genericps.SigGenRising
	msg.TriggerSource = genericps.SigGenNone
	msg.ExtInThreshold = 0
	if scp.psControl != nil && scp.psControl.SetGeneratorCh != nil {
		scp.psControl.SetGeneratorCh <- msg
	}
}

func (scp *ScpDesc) applyFfSimGenSettings(on bool) {
	if scp.ExtGenEnabled && scp.Settings.Ff.UseExternalGen && scp.extGen.Connected() {
		if on {
			scp.extGen.SetAmplitude(scpi.Ch1, float64(scp.Settings.FfGen.Amplitude)/1000000.0)
			scp.extGen.SetOffset(scpi.Ch1, float64(scp.Settings.FfGen.OffsetVoltage)/1000000.0)
			scp.extGen.SetWaveform(scpi.Ch1, "SINusoid")
			scp.extGen.SetOutput(scpi.Ch1, true)
		} else {
			scp.extGen.SetOutput(scpi.Ch1, false)
		}

		// Ensure internal simulator generators are turned off
		for i := 0; i < int(scp.channelCount); i++ {
			msg := &control.GeneratorDescMsg{}
			msg.Channel = genericps.ChannelId(i)
			msg.Operation = genericps.EsOff
			if scp.psControl != nil && scp.psControl.SetSimGenCh != nil {
				scp.psControl.SetSimGenCh <- msg
			}
		}
		return
	}

	if scp.psControl != nil && scp.psControl.SetSimGenCh != nil {
		activeGens := make([]bool, scp.channelCount)
		var missingGenChannels []string

		if on {
			for i := 0; i < int(scp.channelCount); i++ {
				if scp.Settings.Channels[i].Enabled {
					genSrc := int(scp.Settings.Channels[i].RlcFilter.GeneratorSource)
					if genSrc >= 0 && genSrc < int(scp.channelCount) {
						if scp.Settings.SimGenPanel[genSrc].On {
							activeGens[genSrc] = true
						} else {
							missingGenChannels = append(missingGenChannels, channelNames[i])
						}
					}
				}
			}
		}

		if len(missingGenChannels) > 0 {
			if scp.status != nil {
				scp.psControl.DisplayStatus("Error: Channel "+strings.Join(missingGenChannels, ", ")+" has no active generator input", control.Warning)
			}
		} else if scp.status != nil && scp.status.Code() == StatusChannelNoActiveGen {
			scp.psControl.DisplayStatus("", control.Info)
		}

		for i := 0; i < int(scp.channelCount); i++ {
			msg := &control.GeneratorDescMsg{}
			msg.Channel = genericps.ChannelId(i)
			if activeGens[i] {
				msg.On = true
				msg.StartFrequency = scp.Settings.Ff.MinFreq
				msg.StopFrequency = scp.Settings.Ff.MaxFreq
				msg.Increment = 0 // App controls frequency stepping; no hardware sweep
				msg.DwellTime = scp.Settings.Ff.DeltaT
				msg.SweepType = genericps.SweepUp
				msg.WaveType = genericps.Sine
				msg.OffsetVoltage = scp.Settings.FfGen.OffsetVoltage
				msg.PkToPK = scp.Settings.FfGen.Amplitude * 2
				msg.Phase = 0
			} else {
				msg.On = false
				msg.DwellTime = 0
				msg.OffsetVoltage = 0
				msg.PkToPK = 0
				msg.WaveType = genericps.DcVoltage
				msg.StopFrequency = scp.Settings.FfGen.StopFrequency
				msg.SweepType = genericps.SweepDown
			}
			msg.Operation = genericps.EsOff
			msg.Shots = 0
			msg.Sweeps = 0
			msg.TriggerType = genericps.SigGenRising
			msg.TriggerSource = genericps.SigGenNone
			msg.ExtInThreshold = 0
			scp.psControl.SetSimGenCh <- msg
		}
	}
}
