package gui

import (
	"fynescope/disp7"
	"fynescope/genericps"
	"fynescope/selectscroll"
	"fynescope/settings"
	"math"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"gonum.org/v1/gonum/dsp/fourier"
)

func (scp *ScpDesc) newDftPanel(layout *fyne.Container) {
	// FFT specific controls
	chControls := container.NewVBox()
	chControls.Add(widget.NewLabel("Channels:"))

	for i := 0; i < int(scp.channelCount); i++ {
		chIdx := genericps.ChannelId(i)
		chName := channelNames[chIdx]
		channel := &scp.Settings.Channels[chIdx]
		channelViewer := &scp.channelViewers[chIdx]

		// Enable Checkbox
		check := widget.NewCheck("", func(checked bool) {
			scp.EnableChannel(chIdx, checked)
		})
		check.SetChecked(channel.Enabled)
		channelViewer.dftCheckbox = check
		addToTest(check, dftEnableId+chName, dftTabIndex)

		persSelected := func(checked bool) {
			channel.DftPersistence = checked
			scp.Settings.Channels[chIdx].DftPersistence = checked
			if !checked {
				scp.clearDftPersistentLayer(chIdx)
			}
			scp.refreshRasters()
			scp.SaveSettings()
		}
		persCheck := widget.NewCheck("Pers", persSelected)
		persCheck.SetChecked(channel.DftPersistence)
		channelViewer.dftPersistenceCheckbox = persCheck
		addToTest(persCheck, dftPersId+chName, dftTabIndex)

		// Channel Label
		text := "Ch " + chName + ":"
		if scp.isDigitalFilterEnabled(chIdx) {
			text += " ⚠️"
		}
		label := canvas.NewText(text, channel.Col[scp.Settings.ChannelColorIndex])
		label.TextStyle = fyne.TextStyle{Bold: true}
		label.TextSize = theme.TextSize()
		channelViewer.dftNameLabel = label

		// Input Range Selector
		rangesEnum, err := scp.psControl.ChannelRanges(chIdx)
		var ranges []string
		if err == nil {
			for idx := range rangesEnum {
				ranges = append(ranges, inputRanges[rangesEnum[idx]])
			}
		} else {
			ranges = inputRanges
		}

		vRange := selectscroll.NewSelectScroll(ranges, func(option string, e selectscroll.Exception) {
			scp.changeChannelRange(chIdx, option)
		}, "±200mV")
		addToTest(vRange, dftVRangeId+chName, dftTabIndex)

		vr := scp.Settings.Channels[chIdx].VRange
		if s, ok := rangeEnumToString[vr]; ok {
			vRange.SilentSetSelected(s)
		}

		// Add to synchronization list
		channelViewer.vRangeSelects = append(channelViewer.vRangeSelects, vRange)

		// X10 Checkbox
		x10Check := widget.NewCheck("X10", func(c bool) {
			scp.changeChannelX10(chIdx, c)
		})
		x10Check.SetChecked(scp.Settings.Channels[chIdx].X10)
		channelViewer.x10Checkboxes = append(channelViewer.x10Checkboxes, x10Check)
		addToTest(x10Check, dftX10Id+chName, dftTabIndex)

		// Each channel gets its own row
		chRow := container.NewHBox(check, label, vRange, x10Check, persCheck)
		chControls.Add(chRow)
	}

	// Window selector row

	windowSelector := selectscroll.NewSelectScroll([]string{settings.WindowBartlettHann,
		settings.WindowBlackman, settings.WindowBlackmanHarris, settings.WindowBlackmanNuttall, settings.WindowFlatTop, settings.WindowHamming, settings.WindowHann,
		settings.WindowLanczos, settings.WindowNuttall, settings.WindowTriangular, settings.WindowRectangular}, func(selected string, _ selectscroll.Exception) {
		scp.Settings.Dft.Window = selected
		scp.clearAllDftPersistentLayers()
		if scp.dftRaster != nil {
			scp.dftRaster.Refresh()
		}
		scp.SaveSettings()
	}, settings.WindowRectangular)
	windowSelector.SilentSetSelected(scp.Settings.Dft.Window)
	addToTest(windowSelector, dftWindowId, dftTabIndex)
	var arbDbRefContainer *fyne.Container
	arbDbRefDisp, _ := disp7.NewCustomDisp7Array(5, 3, 20000, 1, disp7.UnSigned, disp7.NoTrailingZeroes, scp.Window,
		scp.theme.Color(ColorNameGeneratorDisp, 0), disp7.ReadWrite, disp7.DefaultDigitWidth, disp7.DeafultDigitHeight,
		disp7.DefaultSkew, disp7.DefaultVCursorSpace, "0dB Ref :", " V")

	arbDbRefDisp.SetFloatValue(scp.Settings.Dft.ArbitraryDbRefV, 3)
	arbDbRefDisp.OnChanged = func(val float64) {
		scp.Settings.Dft.ArbitraryDbRefV = val / math.Pow(10, float64(arbDbRefDisp.DpPos()))
		scp.clearAllDftPersistentLayers()
		if scp.dftRaster != nil {
			scp.dftRaster.Refresh()
		}
		scp.SaveSettings()
	}

	arbDbRefContainer = container.NewVBox(arbDbRefDisp)
	if scp.Settings.Dft.DisplayMode != settings.ModeArbitraryDB {
		arbDbRefContainer.Hide()
	}
	addToTest(arbDbRefDisp, dftBinId+"ArbRef", dftTabIndex)

	// Display mode selector row
	modeSelector := selectscroll.NewSelectScroll([]string{settings.ModeDBFS, settings.ModeVoltage, settings.ModeDBV, settings.ModeDBU, settings.ModeDBM, settings.ModeArbitraryDB}, func(selected string, _ selectscroll.Exception) {
		scp.Settings.Dft.DisplayMode = selected
		if selected == settings.ModeArbitraryDB {
			arbDbRefContainer.Show()
		} else {
			arbDbRefContainer.Hide()
		}
		scp.clearAllDftPersistentLayers()
		for i := range scp.channelViewers {
			scp.channelViewers[i].dftLabel.enableRefresh()
		}
		scp.refreshRasters()
		if scp.dftRaster != nil {
			scp.dftRaster.Refresh()
		}
		scp.SaveSettings()
	}, settings.ModeVoltage)
	modeSelector.SilentSetSelected(scp.Settings.Dft.DisplayMode)
	addToTest(modeSelector, dftModeId, dftTabIndex)

	maxPossibleFreq := 500000000.0
	numOfFractionDigits := 2
	numOfDigits := numOfFractionDigits
	f := int(math.Round(maxPossibleFreq))
	for f > 0 {
		f /= 10
		numOfDigits++
	}
	size := float32(0.8)
	refCol := scp.theme.Color(ColorNameGeneratorDisp, 0)

	scp.dftMinFreqDisp, _ = disp7.NewCustomDisp7Array(numOfDigits, numOfFractionDigits,
		int(maxPossibleFreq)*pow10tab[numOfFractionDigits],
		0,
		disp7.UnSigned, disp7.NoTrailingZeroes, scp.Window,
		refCol,
		disp7.ReadWrite, size*disp7.DefaultDigitWidth,
		disp7.DeafultDigitHeight, 1,
		disp7.DefaultVCursorSpace, "Min:", " Hz")

	scp.dftMaxFreqDisp, _ = disp7.NewCustomDisp7Array(numOfDigits, numOfFractionDigits,
		int(maxPossibleFreq)*pow10tab[numOfFractionDigits],
		0,
		disp7.UnSigned, disp7.NoTrailingZeroes, scp.Window,
		refCol,
		disp7.ReadWrite, size*disp7.DefaultDigitWidth,
		disp7.DeafultDigitHeight, 1,
		disp7.DefaultVCursorSpace, "Max:", " Hz")

	scp.dftMinFreqDisp.SetFloatValue(scp.Settings.Dft.MinFreq, 2)
	scp.dftMaxFreqDisp.SetFloatValue(scp.Settings.Dft.MaxFreq, 2)

	scp.dftMinFreqDisp.OnChanged = func(v float64) {
		go func() {
			scp.Settings.Dft.MinFreq = v / 100.0
			if scp.Settings.Dft.MinFreq > scp.Settings.Dft.MaxFreq {
				scp.Settings.Dft.MaxFreq = scp.Settings.Dft.MinFreq
				if scp.dftMaxFreqDisp != nil {
					scp.dftMaxFreqDisp.SetFloatValue(scp.Settings.Dft.MaxFreq, 2)
				}
			}
			scp.setDftHDivsX()
			if scp.dftBottomLabelViewer != nil {
				scp.dftBottomLabelViewer.(*frqLabelViewer).enableRefresh()
			}
			scp.clearAllDftPersistentLayers()
			scp.refreshRasters()
			scp.SaveSettings()
		}()
	}

	scp.dftMaxFreqDisp.OnChanged = func(v float64) {
		go func() {
			scp.Settings.Dft.MaxFreq = v / 100.0
			if scp.Settings.Dft.MaxFreq < scp.Settings.Dft.MinFreq {
				scp.Settings.Dft.MinFreq = scp.Settings.Dft.MaxFreq
				if scp.dftMinFreqDisp != nil {
					scp.dftMinFreqDisp.SetFloatValue(scp.Settings.Dft.MinFreq, 2)
				}
			}
			scp.setDftHDivsX()
			if scp.dftBottomLabelViewer != nil {
				scp.dftBottomLabelViewer.(*frqLabelViewer).enableRefresh()
			}
			scp.clearAllDftPersistentLayers()
			scp.refreshRasters()
			scp.SaveSettings()
		}()
	}
	addToTest(scp.dftMaxFreqDisp, dftMaxFreqValId, dftTabIndex)
	addToTest(scp.dftMinFreqDisp, dftModeId+"MinFreq", dftTabIndex)

	// Bins Selector
	binLabels := []string{"128", "256", "512", "1024", "2048", "4096", "8192", "16384", "32768", "65536", "131072", "262144", "524288", "1048576"}
	binSelector := selectscroll.NewSelectScroll(binLabels, func(selected string, _ selectscroll.Exception) {
		val, _ := strconv.Atoi(selected)
		scp.Settings.Dft.Bins = val
		m = scp.Settings.Dft.Bins * 2
		fft = fourier.NewFFT(scp.Settings.Dft.Bins * 2)
		samples = make([]float64, m)
		fftResult = make([]complex128, fft.Len()/2+1)
		scp.updateBinWidth()
		scp.updateAcquisitionParameters()
		scp.clearAllDftPersistentLayers()
		if scp.dftRaster != nil {
			scp.dftRaster.Refresh()
		}
		scp.SaveSettings()
	}, "1024")
	binSelector.SilentSetSelected(strconv.Itoa(scp.Settings.Dft.Bins))
	addToTest(binSelector, dftBinId, dftTabIndex)

	scp.binWidthLabel = widget.NewLabel("BW: -")
	scp.updateBinWidth()

	scp.dftDataCollectionTimeLabel = widget.NewLabel("Coll: -")
	scp.updateDftDataCollectionTime()

	// Sample Rate Selector
	dftSampleRates := []string{"1", "2", "5", "10", "20", "50", "100", "200", "500"}
	dftSampleUnits := []string{selectscroll.UnitSps, selectscroll.UnitKSps, selectscroll.UnitMSps, selectscroll.UnitGSps}

	scp.dftSampleRateSelect = selectscroll.NewSelectScroll(dftSampleRates, func(selected string, ex selectscroll.Exception) {
		if ex == selectscroll.Over {
			scp.dftSampleUnitUp()
			return
		}
		if ex == selectscroll.Under {
			scp.dftSampleUnitDown()
			return
		}
		scp.Settings.Dft.SampleRate = selected
		scp.syncTimeDivToDft()
		scp.updateAcquisitionParameters()
	}, "100")
	addToTest(scp.dftSampleRateSelect, dftSampleRateId, dftTabIndex)

	scp.dftSampleUnitSelect = selectscroll.NewSelectScroll(dftSampleUnits, func(selected string, _ selectscroll.Exception) {
		scp.Settings.Dft.SampleRateUnit = selected
		scp.syncTimeDivToDft()
		scp.updateAcquisitionParameters()
	}, selectscroll.UnitMSps)
	addToTest(scp.dftSampleUnitSelect, dftSampleUnitId, dftTabIndex)

	scp.dftSampleRateSelect.SilentSetSelected(scp.Settings.Dft.SampleRate)
	scp.dftSampleUnitSelect.SilentSetSelected(scp.Settings.Dft.SampleRateUnit)

	windowCol := container.NewVBox()
	windowCol.Add(widget.NewLabel("Window:"))

	windowCol.Add(windowSelector)
	windowCol.Add(widget.NewLabel("Mode:"))
	windowCol.Add(modeSelector)
	windowCol.Add(arbDbRefContainer)
	windowCol.Add(scp.dftMinFreqDisp)
	logXCheck := widget.NewCheck("Log X", func(b bool) {
		scp.Settings.Dft.XAxisLog = b
		scp.clearAllDftPersistentLayers()
		scp.refreshRasters()
		scp.SaveSettings()
	})
	logXCheck.Checked = scp.Settings.Dft.XAxisLog
	addToTest(logXCheck, dftModeId+"LogX", dftTabIndex)

	windowCol.Add(scp.dftMaxFreqDisp)
	windowCol.Add(logXCheck)
	windowCol.Add(widget.NewLabel("Sample Rate:"))
	windowCol.Add(container.NewHBox(scp.dftSampleRateSelect, scp.dftSampleUnitSelect))
	windowCol.Add(widget.NewLabel("Bins:"))
	windowCol.Add(binSelector)
	windowCol.Add(scp.binWidthLabel)
	windowCol.Add(scp.dftDataCollectionTimeLabel)

	layout.Add(container.NewVBox(chControls, windowCol))
}
