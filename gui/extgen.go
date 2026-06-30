package gui

import (
	"fmt"
	"log/slog"
	"math"
	"fynescope/control"
	"fynescope/control/scpi"
	"fynescope/disp7"
	"fynescope/genericps"
	"fynescope/selectscroll"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// newExtGenTab builds the full external-generator tab content and returns it
// as a container ready to be placed inside the controlTab AppTabs.
func (scp *ScpDesc) newExtGenTab(undockable bool) *fyne.Container {
	root := container.New(layout.NewVBoxLayout())

	// Status label: reflects current connection state.
	statusLabel := widget.NewLabel("Not connected")

	updateStatus := func() {
		if scp.extGen.Connected() {
			statusLabel.SetText("Connected")
			if scp.useExtGenCheck != nil {
				scp.useExtGenCheck.Show()
			}
		} else {
			statusLabel.SetText("Not connected")
			if scp.useExtGenCheck != nil {
				scp.useExtGenCheck.Hide()
			}
		}
	}

	// ── USB VID / PID entries ────────────────────────────────────────────────
	extGenVid := newDigitEntry()
	extGenVid.SetPlaceHolder("     VID")
	extGenVid.SetText(scp.Settings.Ff.ExternalGenUsbVid)

	extGenPid := newDigitEntry()
	extGenPid.SetPlaceHolder("     PID")
	extGenPid.SetText(scp.Settings.Ff.ExternalGenUsbPid)

	extGenVid.OnChanged = func(s string) {
		scp.Settings.Ff.ExternalGenUsbVid = s
		scp.extGen.Disconnect()
		updateStatus()
		scp.SaveSettings()
	}
	extGenPid.OnChanged = func(s string) {
		scp.Settings.Ff.ExternalGenUsbPid = s
		scp.extGen.Disconnect()
		updateStatus()
		scp.SaveSettings()
	}

	// ── Connect / Disconnect buttons ─────────────────────────────────────────
	connectButton := widget.NewButton("Connect", func() {
		cfg := control.ExtGenConfig{
			Port:   scp.Settings.Ff.ExternalGenPort,
			UsbVid: scp.Settings.Ff.ExternalGenUsbVid,
			UsbPid: scp.Settings.Ff.ExternalGenUsbPid,
		}
		if err := scp.extGen.Connect(cfg); err != nil {
			slog.Error("external generator connect failed", "err", err)
			statusLabel.SetText("Error: " + err.Error())
			if scp.psControl != nil && scp.psControl.DisplayStatus != nil {
				scp.psControl.DisplayStatus("ExtGen: "+err.Error(), control.Warning)
			}
		} else {
			// If autodetection occurred, update widgets and settings
			vid, pid := scp.extGen.GetVidPid()
			if vid != "" && pid != "" && (vid != scp.Settings.Ff.ExternalGenUsbVid || pid != scp.Settings.Ff.ExternalGenUsbPid) {
				vidChanged := extGenVid.OnChanged
				pidChanged := extGenPid.OnChanged
				extGenVid.OnChanged = nil
				extGenPid.OnChanged = nil

				extGenVid.SetText(vid)
				extGenPid.SetText(pid)
				scp.Settings.Ff.ExternalGenUsbVid = vid
				scp.Settings.Ff.ExternalGenUsbPid = pid
				scp.SaveSettings()

				extGenVid.OnChanged = vidChanged
				extGenPid.OnChanged = pidChanged
			}

			updateStatus()
			scp.syncExtGenSettings()
		}
	})

	disconnectButton := widget.NewButton("Disconnect", func() {
		scp.extGen.Disconnect()
		updateStatus()
	})

	var undockButton *widget.Button
	if undockable {
		undockButton = widget.NewButtonWithIcon("Undock", theme.ViewFullScreenIcon(), func() {
			if scp.extgenWindow == nil {
				scp.extgenWindow = scp.App.NewWindow("External Generator")
				scp.controlTab.Remove(scp.extgenTab)

				// Create completely new UI tree for the undocked window to avoid Fyne canvas crashes
				newLayout := scp.newExtGenTab(false)
				scrollContainer := container.NewScroll(newLayout)
				scp.extgenWindow.SetContent(scrollContainer)
				scp.extgenWindow.Resize(fyne.NewSize(450, 700))

				scp.extgenWindow.SetOnClosed(func() {
					// Re-dock
					// Recreate the docked UI tree to sync with any settings changed while undocked
					scp.extgenLayout.RemoveAll()
					scp.extgenLayout.Add(scp.newExtGenTab(true))
					scp.extgenTab.Content = scp.extgenLayout
					scp.dockTab(scp.extgenTab)
					scp.extgenWindow = nil
				})

				scp.extgenWindow.Show()
			} else {
				scp.extgenWindow.Close()
			}
		})
	}

	createChannelColumn := func(chIdx int, scpiCh scpi.ChType) *fyne.Container {
		// chLabel := widget.NewLabelWithStyle(fmt.Sprintf("Channel %d", chIdx+1), fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
		chName := fmt.Sprintf("Channel %d", chIdx+1)
		onOffCheck := widget.NewCheck(chName, func(b bool) {
			scp.Settings.ExtGen[chIdx].On = b
			if scp.extGen.Connected() {
				if err := scp.extGen.SetOutput(scpiCh, b); err != nil {
					slog.Error("extgen set output", "err", err)
				}
			}
			scp.SaveSettings()
		})
		onOffCheck.SetChecked(scp.Settings.ExtGen[chIdx].On)

		waveTypeOptions := []string{"Sine", "Square", "Triangle", "RampUp", "RampDown"}
		waveTypeMap := map[string]genericps.WaveTypeEnum{
			"Sine":     genericps.Sine,
			"Square":   genericps.Square,
			"Triangle": genericps.Triangle,
			"RampUp":   genericps.RampUp,
			"RampDown": genericps.RampDown,
		}

		waveTypeSelect := selectscroll.NewSelectScroll(waveTypeOptions, func(option string, ex selectscroll.Exception) {
			scp.Settings.ExtGen[chIdx].WaveType = waveTypeMap[option]
			if scp.extGen.Connected() {
				var err error
				switch scp.Settings.ExtGen[chIdx].WaveType {
				case genericps.Sine:
					err = scp.extGen.SetWaveform(scpiCh, "SINusoid")
				case genericps.Square:
					err = scp.extGen.SetWaveform(scpiCh, "SQUare")
				case genericps.Triangle:
					if err = scp.extGen.SetWaveform(scpiCh, "RAMP"); err == nil {
						err = scp.extGen.SetRampSymmetry(scpiCh, "RAMP:SYMMetry 50")
					}
				case genericps.RampUp:
					if err = scp.extGen.SetWaveform(scpiCh, "RAMP"); err == nil {
						err = scp.extGen.SetRampSymmetry(scpiCh, "RAMP:SYMMetry 100")
					}
				case genericps.RampDown:
					if err = scp.extGen.SetWaveform(scpiCh, "RAMP"); err == nil {
						err = scp.extGen.SetRampSymmetry(scpiCh, "RAMP:SYMMetry 0")
					}
				default:
					err = scp.extGen.SetWaveform(scpiCh, "SINusoid")
				}
				if err != nil {
					slog.Error("extgen set waveform", "err", err)
				}
			}
			scp.SaveSettings()
		}, "RampDown")

		for key, val := range waveTypeMap {
			if val == scp.Settings.ExtGen[chIdx].WaveType {
				waveTypeSelect.SilentSetSelected(key)
				break
			}
		}

		waveRow := container.NewHBox(widget.NewLabel("Waveform:"), waveTypeSelect)

		const maxV = 2000000
		const size = 0.8

		const maxFreq = 100000000.0 // 100 MHz

		disp7Width := fractionWidth
		// Allocate enough digits for the maximum supported frequency (100 MHz).
		f := int(math.Round(maxFreq))
		for f > 0 {
			f /= 10
			disp7Width++
		}

		frequency, _ := disp7.NewCustomDisp7Array(disp7Width, fractionWidth,
			int(maxFreq)*pow10tab[fractionWidth],
			int(genericps.MinFrequency)*pow10tab[fractionWidth],
			disp7.UnSigned, disp7.NoTrailingZeroes, scp.Window,
			scp.theme.Color(ColorNameGeneratorDisp, 0),
			disp7.ReadWrite, size*disp7.DefaultDigitWidth,
			disp7.DeafultDigitHeight, 1,
			disp7.DefaultVCursorSpace, "Frq: ", " Hz")
		frequency.SilentSetValue(int(scp.Settings.ExtGen[chIdx].Frequency) * pow10tab[fractionWidth])
		frequency.OnChanged = func(v float64) {
			scp.Settings.ExtGen[chIdx].Frequency = v / float64(pow10tab[fractionWidth])
			if scp.extGen.Connected() {
				if err := scp.extGen.SetFrequency(scpiCh, scp.Settings.ExtGen[chIdx].Frequency); err != nil {
					slog.Error("external gen set freq failed", "err", err)
					if scp.psControl != nil && scp.psControl.DisplayStatus != nil {
						scp.psControl.DisplayStatus("ExtGen: "+err.Error(), control.Warning)
					}
				}
			}
			scp.SaveSettings()
		}

		amp, _ := disp7.NewCustomDisp7Array(7, 6, maxV, 0,
			disp7.SignedHidden, disp7.NoTrailingZeroes, scp.Window,
			scp.theme.Color(ColorNameGeneratorDisp, 0),
			disp7.ReadWrite, size*disp7.DefaultDigitWidth,
			disp7.DeafultDigitHeight, 1,
			disp7.DefaultVCursorSpace, "Amp:", " V")
		amp.SilentSetValue(int(scp.Settings.ExtGen[chIdx].Amplitude))
		amp.OnChanged = func(v float64) {
			scp.Settings.ExtGen[chIdx].Amplitude = uint32(v)
			if scp.extGen.Connected() {
				if err := scp.extGen.SetAmplitude(scpiCh, float64(v)/1000000.0); err != nil {
					slog.Error("extgen set amplitude", "err", err)
				}
			}
			scp.SaveSettings()
		}

		offset, _ := disp7.NewCustomDisp7Array(7, 6, maxV, -maxV,
			disp7.Signed, disp7.NoTrailingZeroes, scp.Window,
			scp.theme.Color(ColorNameGeneratorDisp, 0),
			disp7.ReadWrite, size*disp7.DefaultDigitWidth,
			disp7.DeafultDigitHeight, 1,
			disp7.DefaultVCursorSpace, "Offs:", " V")
		offset.SilentSetValue(int(scp.Settings.ExtGen[chIdx].OffsetVoltage))
		offset.OnChanged = func(v float64) {
			scp.Settings.ExtGen[chIdx].OffsetVoltage = int32(v)
			if scp.extGen.Connected() {
				if err := scp.extGen.SetOffset(scpiCh, float64(v)/1000000.0); err != nil {
					slog.Error("extgen set offset", "err", err)
				}
			}
			scp.SaveSettings()
		}

		phase, _ := disp7.NewCustomDisp7Array(4, 1, 3600, 0,
			disp7.UnSigned, disp7.NoTrailingZeroes, scp.Window,
			scp.theme.Color(ColorNameGeneratorDisp, 0),
			disp7.ReadWrite, size*disp7.DefaultDigitWidth,
			disp7.DeafultDigitHeight, 1,
			disp7.DefaultVCursorSpace, "Phase:", " °")
		phase.SilentSetValue(int(scp.Settings.ExtGen[chIdx].Phase * 10))
		phase.OnChanged = func(v float64) {
			scp.Settings.ExtGen[chIdx].Phase = v / 10.0
			if scp.extGen.Connected() {
				if err := scp.extGen.SetPhase(scpiCh, scp.Settings.ExtGen[chIdx].Phase); err != nil {
					slog.Error("extgen set phase", "err", err)
				}
			}
			scp.SaveSettings()
		}

		// ── Impedance ─────────────────────────────────────────────────────────
		impedanceModeOptions := []string{"Ohms", "INFinity", "MINimum", "MAXimum"}
		// default to INFinity if not set
		if scp.Settings.ExtGen[chIdx].ImpedanceMode == "" {
			scp.Settings.ExtGen[chIdx].ImpedanceMode = "INFinity"
		}
		if scp.Settings.ExtGen[chIdx].ImpedanceOhms == 0 {
			scp.Settings.ExtGen[chIdx].ImpedanceOhms = 50
		}

		impedanceOhmsWidget, _ := disp7.NewCustomDisp7Array(5, 0, 10000, 1,
			disp7.UnSigned, disp7.NoTrailingZeroes, scp.Window,
			scp.theme.Color(ColorNameGeneratorDisp, 0),
			disp7.ReadWrite, size*disp7.DefaultDigitWidth,
			disp7.DeafultDigitHeight, 1,
			disp7.DefaultVCursorSpace, "Imp:", " Ω")
		impedanceOhmsWidget.SilentSetValue(scp.Settings.ExtGen[chIdx].ImpedanceOhms)
		impedanceOhmsWidget.OnChanged = func(v float64) {
			scp.Settings.ExtGen[chIdx].ImpedanceOhms = int(v)
			if scp.extGen.Connected() {
				if err := scp.extGen.SetImpedance(scpiCh, fmt.Sprintf("%d", int(v))); err != nil {
					slog.Error("extgen set impedance", "err", err)
				}
			}
			scp.SaveSettings()
		}

		// Show/hide the ohms widget depending on selected mode.
		setImpedanceOhmsVisible := func(mode string) {
			if mode == "Ohms" {
				impedanceOhmsWidget.Show()
			} else {
				impedanceOhmsWidget.Hide()
			}
		}

		impedanceModeSelect := selectscroll.NewSelectScroll(impedanceModeOptions, func(option string, ex selectscroll.Exception) {
			scp.Settings.ExtGen[chIdx].ImpedanceMode = option
			setImpedanceOhmsVisible(option)
			if scp.extGen.Connected() {
				var val string
				if option == "Ohms" {
					val = fmt.Sprintf("%d", scp.Settings.ExtGen[chIdx].ImpedanceOhms)
				} else {
					val = option
				}
				if err := scp.extGen.SetImpedance(scpiCh, val); err != nil {
					slog.Error("extgen set impedance", "err", err)
				}
			}
			scp.SaveSettings()
		}, "INFinity")
		impedanceModeSelect.SilentSetSelected(scp.Settings.ExtGen[chIdx].ImpedanceMode)
		setImpedanceOhmsVisible(scp.Settings.ExtGen[chIdx].ImpedanceMode)

		impedanceRow := container.NewHBox(widget.NewLabel("Impedance:"), impedanceModeSelect)

		return container.NewVBox(
			// chLabel,
			onOffCheck,
			waveRow,
			frequency,
			amp,
			offset,
			phase,
			impedanceRow,
			impedanceOhmsWidget,
		)
	}

	ch1Col := createChannelColumn(0, scpi.Ch1)
	ch2Col := createChannelColumn(1, scpi.Ch2)
	channelsHBox := container.NewVBox(ch1Col, widget.NewSeparator(), ch2Col)

	scpiTestEntry := widget.NewEntry()
	scpiTestEntry.SetPlaceHolder("Test SCPI Cmd (e.g. :VOLT1 1.0)")
	scpiSendBtn := widget.NewButton("Send", func() {
		if scp.extGen.Connected() {
			if err := scp.extGen.SendRaw(scpiTestEntry.Text); err != nil {
				slog.Error("scpi test error", "err", err)
			}
		}
	})
	scpiRow := container.NewBorder(nil, nil, widget.NewLabel("Raw Cmd:"), scpiSendBtn, scpiTestEntry)

	// Assemble layout
	vidPidRow := container.NewHBox(
		widget.NewLabel("VID:"), extGenVid,
		widget.NewLabel("PID:"), extGenPid,
	)
	var btnRow *fyne.Container
	if undockable {
		btnRow = container.NewHBox(connectButton, disconnectButton, statusLabel, layout.NewSpacer())
	} else {
		btnRow = container.NewHBox(connectButton, disconnectButton, statusLabel)
	}
	if undockable {
		root.Add(undockButton)
	}
	root.Add(scpiRow)
	root.Add(vidPidRow)
	root.Add(btnRow)
	root.Add(channelsHBox)

	updateStatus()
	return root
}

// setExtGenFrequency forwards a frequency command to the external generator
// over its SCPI control layer. It logs errors and shows them in the status bar.
func (scp *ScpDesc) setExtGenFrequency(f float64) {
	if err := scp.extGen.SetFrequency(scpi.Ch1, f); err != nil {
		slog.Error("external gen set frequency failed", "err", err)
		if scp.psControl != nil && scp.psControl.DisplayStatus != nil {
			scp.psControl.DisplayStatus("ExtGen: "+err.Error(), control.Warning)
		}
	}
}

// syncExtGenSettings pushes the complete configuration of both channels
// to the external generator immediately upon connection.
func (scp *ScpDesc) syncExtGenSettings() {
	if !scp.extGen.Connected() {
		return
	}

	for chIdx, scpiCh := range []scpi.ChType{scpi.Ch1, scpi.Ch2} {
		settings := scp.Settings.ExtGen[chIdx]

		// Set output ON/OFF
		if err := scp.extGen.SetOutput(scpiCh, settings.On); err != nil {
			slog.Error("extgen sync output", "err", err)
		}

		// Set waveform
		var err error
		switch settings.WaveType {
		case genericps.Sine:
			err = scp.extGen.SetWaveform(scpiCh, "SINusoid")
		case genericps.Square:
			err = scp.extGen.SetWaveform(scpiCh, "SQUare")
		case genericps.Triangle:
			if err = scp.extGen.SetWaveform(scpiCh, "RAMP"); err == nil {
				err = scp.extGen.SetRampSymmetry(scpiCh, "RAMP:SYMMetry 50")
			}
		case genericps.RampUp:
			if err = scp.extGen.SetWaveform(scpiCh, "RAMP"); err == nil {
				err = scp.extGen.SetRampSymmetry(scpiCh, "RAMP:SYMMetry 100")
			}
		case genericps.RampDown:
			if err = scp.extGen.SetWaveform(scpiCh, "RAMP"); err == nil {
				err = scp.extGen.SetRampSymmetry(scpiCh, "RAMP:SYMMetry 0")
			}
		default:
			err = scp.extGen.SetWaveform(scpiCh, "SINusoid")
		}
		if err != nil {
			slog.Error("extgen sync waveform", "err", err)
		}

		// Set frequency
		if err := scp.extGen.SetFrequency(scpiCh, settings.Frequency); err != nil {
			slog.Error("extgen sync frequency", "err", err)
		}

		// Set amplitude
		if err := scp.extGen.SetAmplitude(scpiCh, float64(settings.Amplitude)/1000000.0); err != nil {
			slog.Error("extgen sync amplitude", "err", err)
		}

		// Set offset
		if err := scp.extGen.SetOffset(scpiCh, float64(settings.OffsetVoltage)/1000000.0); err != nil {
			slog.Error("extgen sync offset", "err", err)
		}

		// Set phase
		if err := scp.extGen.SetPhase(scpiCh, settings.Phase); err != nil {
			slog.Error("extgen sync phase", "err", err)
		}

		// Set impedance
		var impedanceVal string
		if settings.ImpedanceMode == "Ohms" {
			if settings.ImpedanceOhms == 0 {
				impedanceVal = "INFinity"
			} else {
				impedanceVal = fmt.Sprintf("%d", settings.ImpedanceOhms)
			}
		} else if settings.ImpedanceMode != "" {
			impedanceVal = settings.ImpedanceMode
		} else {
			impedanceVal = "INFinity"
		}
		if err := scp.extGen.SetImpedance(scpiCh, impedanceVal); err != nil {
			slog.Error("extgen sync impedance", "err", err)
		}
	}
}
