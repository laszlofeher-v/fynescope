package gui

import (
	"fynescope/disp7"
	"fynescope/genericps"
	"fynescope/selectscroll"
	"log/slog"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"fynescope/settings"
)

// newFfPanel initializes the graphical control panel container for the F(f) tab.
// It constructs channel checkboxes (Enabled, Phase, Reference), voltage range dropdowns,
// digit displays (Start/Stop frequency, Step, Delta T), and binds state updates
// to sweep calculations and generator registers.
func (scp *ScpDesc) newFfPanel(panel *fyne.Container) {
	// Clamp configuration settings to safe valid ranges to prevent out-of-bounds panics or initialization errors
	if scp.Settings.Ff.ReferenceChannel < 0 || scp.Settings.Ff.ReferenceChannel >= int(scp.channelCount) {
		scp.Settings.Ff.ReferenceChannel = 0
	}

	maxPossibleFreq := genericps.SineMaxFrequency
	if scp.Settings.Ff.UseExternalGen {
		maxPossibleFreq = 100000000.0 // 100 MHz
	} else {
		// Ensure enough digits are allocated for external generator even if currently using internal generator
		maxPossibleFreq = 100000000.0 // 100 MHz
	}

	if scp.Settings.Ff.MinFreq < genericps.MinFrequency {
		scp.Settings.Ff.MinFreq = genericps.MinFrequency
	}
	// Initial bounds check
	if scp.Settings.Ff.MaxFreq > maxPossibleFreq {
		scp.Settings.Ff.MaxFreq = maxPossibleFreq
	}
	if scp.Settings.Ff.MinFreq > scp.Settings.Ff.MaxFreq {
		scp.Settings.Ff.MinFreq = scp.Settings.Ff.MaxFreq
	}

	if scp.Settings.Ff.DeltaT < 0.001 {
		scp.Settings.Ff.DeltaT = 1.0
	}

	maxV := 2000000 // 2V peak-to-peak
	if scp.Settings.Ff.Amplitude <= 0 {
		scp.Settings.Ff.Amplitude = 2000000
	} else if scp.Settings.Ff.Amplitude > uint32(maxV) {
		scp.Settings.Ff.Amplitude = uint32(maxV)
	}

	vbox := container.New(layout.NewVBoxLayout())
	var refChecks []*widget.Check

	for i := 0; i < int(scp.channelCount); i++ {
		chIndex := genericps.ChannelId(i)
		chName := channelNames[i]

		// Channel Label
		text := "Ch " + chName + ":"
		if scp.isDigitalFilterEnabled(chIndex) {
			text += " ⚠️"
		}
		label := canvas.NewText(text, scp.Settings.Channels[chIndex].Col[scp.Settings.ChannelColorIndex])
		label.TextStyle = fyne.TextStyle{Bold: true}
		label.TextSize = theme.TextSize()
		scp.channelViewers[chIndex].ffNameLabel = label

		// Enabled Checkbox
		enabledCheck := widget.NewCheck("Enabled", func(b bool) {
			scp.EnableChannel(chIndex, b)
		})
		enabledCheck.SetChecked(scp.Settings.Channels[chIndex].Enabled)
		scp.channelViewers[chIndex].enableChecks = append(scp.channelViewers[chIndex].enableChecks, enabledCheck)

		// Phase Checkbox
		phaseCheck := widget.NewCheck("Phase", func(b bool) {
			scp.Settings.Channels[chIndex].FfPhaseEnabled = b
			scp.ResetFfSweep()
		})
		phaseCheck.SetChecked(scp.Settings.Channels[chIndex].FfPhaseEnabled)

		// Ref Radio (Check)
		refCheck := widget.NewCheck("Ref", nil)
		refChecks = append(refChecks, refCheck)

		// Range Selector
		rangesEnum, _ := scp.psControl.ChannelRanges(chIndex)
		var ranges []string
		for _, r := range rangesEnum {
			ranges = append(ranges, inputRanges[r])
		}
		vRange := selectscroll.NewSelectScroll(ranges, func(opt string, ex selectscroll.Exception) {
			scp.changeChannelRange(chIndex, opt)
		}, "+500m")
		scp.channelViewers[chIndex].vRangeSelects = append(scp.channelViewers[chIndex].vRangeSelects, vRange)
		vr := scp.Settings.Channels[chIndex].VRange
		if s, ok := rangeEnumToString[vr]; ok {
			vRange.SetSelected(s)
		}

		// X10 Checkbox
		x10Check := widget.NewCheck("X10", func(c bool) {
			scp.changeChannelX10(chIndex, c)
		})
		x10Check.SetChecked(scp.Settings.Channels[chIndex].X10)
		scp.channelViewers[chIndex].x10Checkboxes = append(scp.channelViewers[chIndex].x10Checkboxes, x10Check)

		// Arrange settings to minimize width
		row1 := container.New(layout.NewHBoxLayout(), label, enabledCheck, phaseCheck, refCheck)
		row2 := container.New(layout.NewHBoxLayout(), widget.NewLabel("Range:"), vRange, x10Check)

		chBox := container.New(layout.NewVBoxLayout(), row1, row2)
		if i > 0 {
			vbox.Add(layout.NewSpacer())
		}
		vbox.Add(chBox)

		addToTest(enabledCheck, ffEnableId+chName)
		addToTest(phaseCheck, ffPhaseCheckId+chName)
		addToTest(refCheck, ffRefCheckId+chName)
		addToTest(vRange, ffVRangeId+chName)
		addToTest(x10Check, ffX10Id+chName)
	}

	// Declare disp7 widgets first so they can be referenced in the OnChanged closure
	// deltaDisp, deltaTDisp removed

	for i := 0; i < int(scp.channelCount); i++ {
		idx := i
		refChecks[idx].OnChanged = func(b bool) {
			if b {
				scp.Settings.Ff.ReferenceChannel = idx
				for j, rc := range refChecks {
					if j != idx {
						rc.SetChecked(false)
					}
				}

				// Update disp7 widget colors based on reference channel
				refCol := scp.Settings.Channels[idx].Col[scp.Settings.ChannelColorIndex]
				if scp.ffMinFreqDisp != nil {
					scp.ffMinFreqDisp.SetOncolor(refCol)
					scp.ffMaxFreqDisp.SetOncolor(refCol)
					// deltaDisp and deltaTDisp color updates removed
					if scp.ffCurrentFreqDisp != nil {
						scp.ffCurrentFreqDisp.SetOncolor(refCol)
					}
				}

				scp.ResetFfSweep()
			}
		}
		if scp.Settings.Ff.ReferenceChannel == i {
			refChecks[i].SetChecked(true)
		}
	}

	panel.Add(vbox)

	// Horizontal Sweep Disp7 Widgets
	slog.Debug("newFfPanel", "maxPossibleFreq", maxPossibleFreq)
	numOfFractionDigits := 2
	numOfDigits := numOfFractionDigits
	f := int(math.Round(maxPossibleFreq))
	for f > 0 {
		f /= 10
		numOfDigits++
	}
	size := float32(0.8)
	refCol := scp.Settings.Channels[scp.Settings.Ff.ReferenceChannel].Col[scp.Settings.ChannelColorIndex]

	scp.ffMinFreqDisp, _ = disp7.NewCustomDisp7Array(numOfDigits, numOfFractionDigits,
		int(maxPossibleFreq)*pow10tab[numOfFractionDigits],
		int(genericps.MinFrequency)*pow10tab[numOfFractionDigits],
		disp7.UnSigned, disp7.NoTrailingZeroes, scp.Window,
		refCol,
		disp7.ReadWrite, size*disp7.DefaultDigitWidth,
		disp7.DeafultDigitHeight, 1,
		disp7.DefaultVCursorSpace, "Start:", " Hz")

	scp.ffMaxFreqDisp, _ = disp7.NewCustomDisp7Array(numOfDigits, numOfFractionDigits,
		int(maxPossibleFreq)*100, int(genericps.MinFrequency)*pow10tab[numOfFractionDigits],
		disp7.UnSigned, disp7.NoTrailingZeroes, scp.Window,
		refCol,
		disp7.ReadWrite, size*disp7.DefaultDigitWidth,
		disp7.DeafultDigitHeight, 1,
		disp7.DefaultVCursorSpace, "Stop :", " Hz")

	scp.ffMinFreqDisp.SetFloatValue(scp.Settings.Ff.MinFreq, 2)
	scp.ffMaxFreqDisp.SetFloatValue(scp.Settings.Ff.MaxFreq, 2)

	scp.ffCurrentFreqDisp, _ = disp7.NewCustomDisp7Array(numOfDigits, numOfFractionDigits,
		int(maxPossibleFreq)*pow10tab[numOfFractionDigits],
		int(genericps.MinFrequency)*pow10tab[numOfFractionDigits],
		disp7.UnSigned, disp7.NoTrailingZeroes, scp.Window,
		refCol,
		disp7.ReaOnly, size*disp7.DefaultDigitWidth,
		disp7.DeafultDigitHeight, 1,
		disp7.DefaultVCursorSpace, "Freq :", " Hz")
	scp.updateFfCurrentFreq()

	scp.ffStepFreqDisp, _ = disp7.NewCustomDisp7Array(3, 0,
		500,
		5,
		disp7.UnSigned, disp7.NoTrailingZeroes, scp.Window,
		refCol,
		disp7.ReadWrite, size*disp7.DefaultDigitWidth,
		disp7.DeafultDigitHeight, 1,
		disp7.DefaultVCursorSpace, "Pts/dec:", "")

	scp.ffStepFreqDisp.OnChanged = func(v float64) {
		scp.Settings.Ff.PtsDec = v
		scp.SaveSettings()
	}
	if scp.Settings.Ff.PtsDec < 5 {
		scp.Settings.Ff.PtsDec = 5
	}
	scp.ffStepFreqDisp.SetValue(int(scp.Settings.Ff.PtsDec))

	scp.ffDeltaTDisp, _ = disp7.NewCustomDisp7Array(5, 3,
		10000, 0,
		disp7.UnSigned, disp7.NoTrailingZeroes, scp.Window,
		refCol,
		disp7.ReadWrite, size*disp7.DefaultDigitWidth,
		disp7.DeafultDigitHeight, 1,
		disp7.DefaultVCursorSpace, "ΔT:", " s")
	scp.ffDeltaTDisp.SetFloatValue(scp.Settings.Ff.DeltaT, 3)

	syncGenStartStopAndStep := func() {
		scp.Settings.FfGen.StartFrequency = scp.Settings.Ff.MinFreq
		scp.Settings.FfGen.StopFrequency = scp.Settings.Ff.MaxFreq
		scp.Settings.FfGen.Frequency = scp.Settings.Ff.MinFreq
		scp.Settings.FfGen.Dwelltime = scp.Settings.Ff.DeltaT
		scp.SaveSettings()

		if scp.psControl != nil && scp.psControl.Con != nil && scp.psControl.Con.ID == genericps.SimId {
			scp.applyFfSimGenSettings(false)
			scp.applyFfSimGenSettings(scp.Settings.FfGen.On)
		} else {
			scp.applyFfGenSettings(false)
			scp.applyFfGenSettings(scp.Settings.FfGen.On)
		}
	}

	scp.ffDeltaTDisp.OnChanged = func(v float64) {
		scp.Settings.Ff.DeltaT = v / 1000.0
		scp.ResetFfSweep()
		syncGenStartStopAndStep()
	}

	scp.ffMinFreqDisp.OnChanged = func(v float64) {
		scp.Settings.Ff.MinFreq = v / 100.0
		if scp.Settings.Ff.MinFreq > scp.Settings.Ff.MaxFreq {
			scp.Settings.Ff.MaxFreq = scp.Settings.Ff.MinFreq
			if scp.ffMaxFreqDisp != nil {
				scp.ffMaxFreqDisp.SetFloatValue(scp.Settings.Ff.MaxFreq, 2)
			}
		}
		scp.ResetFfSweep()
		syncGenStartStopAndStep()
	}
	scp.ffMaxFreqDisp.OnChanged = func(v float64) {
		scp.Settings.Ff.MaxFreq = v / 100.0
		if scp.Settings.Ff.MaxFreq < scp.Settings.Ff.MinFreq {
			scp.Settings.Ff.MinFreq = scp.Settings.Ff.MaxFreq
			if scp.ffMinFreqDisp != nil {
				scp.ffMinFreqDisp.SetFloatValue(scp.Settings.Ff.MinFreq, 2)
			}
		}
		scp.ResetFfSweep()
		syncGenStartStopAndStep()
	}

	if scp.Settings.Dft.DisplayMode == "" {
		scp.Settings.Dft.DisplayMode = "dB"
	}
	dispModeSelect := selectscroll.NewSelectScroll([]string{settings.ModeVoltage, settings.ModeDB}, func(opt string, ex selectscroll.Exception) {
		scp.Settings.Dft.DisplayMode = opt
		scp.ResetFfSweep()
		scp.SaveSettings()
	}, settings.ModeVoltage)
	dispModeSelect.SilentSetSelected(scp.Settings.Dft.DisplayMode)

	dispModeControls := container.NewHBox(widget.NewLabel(" Mode:"), dispModeSelect)

	// Generator controls container
	genHeader := widget.NewLabelWithStyle("Generator Settings", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	isSim := false
	if scp.psControl != nil && scp.psControl.Con != nil && scp.psControl.Con.ID == genericps.SimId {
		isSim = true
	}

	var genPanel *fyne.Container
	var genErr error
	if isSim {
		genPanel, genErr = scp.newFfSimGenPanel()
	} else {
		genPanel, genErr = scp.newFfGenPanel()
	}
	if genErr != nil {
		slog.Error("Failed to create generator panel", "err", genErr)
	}

	scp.useExtGenCheck = widget.NewCheck("Use external generator", func(checked bool) {
		scp.Settings.Ff.UseExternalGen = checked
		scp.updateFfWidgetLimits()
		scp.SaveSettings()
		
		if scp.psControl != nil && scp.psControl.Con != nil && scp.psControl.Con.ID == genericps.SimId {
			scp.applyFfSimGenSettings(scp.Settings.FfGen.On)
		} else {
			scp.applyFfGenSettings(scp.Settings.FfGen.On)
		}
	})
	scp.useExtGenCheck.SetChecked(scp.Settings.Ff.UseExternalGen)
	if !scp.ExtGenEnabled || !scp.extGen.Connected() {
		scp.useExtGenCheck.Hide()
	}

	genVBox := container.NewVBox(
		layout.NewSpacer(),
		genHeader,
		scp.useExtGenCheck,
	)
	if genPanel != nil {
		genVBox.Add(genPanel)
	}

	infoHead := widget.NewLabelWithStyle("Status / Info", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	genVBox.Add(layout.NewSpacer())
	genVBox.Add(infoHead)
	genVBox.Add(scp.ffCurrentFreqDisp)

	genSettings := container.New(layout.NewVBoxLayout(), scp.ffMinFreqDisp,
		scp.ffMaxFreqDisp, scp.ffDeltaTDisp, scp.ffStepFreqDisp,
		dispModeControls, genVBox)
	panel.Add(genSettings)

	addToTest(scp.ffMinFreqDisp, ffMinFreqId)
	addToTest(scp.ffMaxFreqDisp, ffMaxFreqId)
	addToTest(scp.ffCurrentFreqDisp, ffCurrentFreqId)
	addToTest(dispModeSelect, ffDispModeSelectId)
	addToTest(scp.useExtGenCheck, ffExtGenSelectId)

	scp.updateFfWidgetLimits()
}

// updateFfWidgetLimits adjusts the min/max limits of the frequency display widgets
// to match the selected generator (internal scope generator or external USB device).
// It also clamps any stored settings values that fall outside the new limits.
func (scp *ScpDesc) updateFfWidgetLimits() {
	if scp.ffMinFreqDisp == nil || scp.ffMaxFreqDisp == nil || scp.ffCurrentFreqDisp == nil {
		return
	}

	var minF, maxF float64
	if scp.Settings.Ff.UseExternalGen {
		minF = 0.01        // 10 mHz
		maxF = 100000000.0 // 100 MHz
	} else {
		minF = genericps.MinFrequency
		maxF = genericps.SineMaxFrequency
	}

	fractionWidth := 2
	scale := int(math.Pow10(fractionWidth))

	minLimitVal := int(minF * float64(scale))
	maxLimitVal := int(maxF * float64(scale))

	scp.ffMinFreqDisp.Value = int(scp.Settings.Ff.MinFreq * float64(scale))
	scp.ffMaxFreqDisp.Value = int(scp.Settings.Ff.MaxFreq * float64(scale))

	scp.ffMinFreqDisp.SetMinMax(minLimitVal, maxLimitVal)
	scp.ffMaxFreqDisp.SetMinMax(minLimitVal, maxLimitVal)
	scp.ffCurrentFreqDisp.SetMinMax(minLimitVal, maxLimitVal)

	// In case the values were clamped, update settings to match
	scp.Settings.Ff.MinFreq = float64(scp.ffMinFreqDisp.Value) / float64(scale)
	scp.Settings.Ff.MaxFreq = float64(scp.ffMaxFreqDisp.Value) / float64(scale)

}
