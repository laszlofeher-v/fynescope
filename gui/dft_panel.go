package gui

import (
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
	// Display mode selector row
	modeSelector := selectscroll.NewSelectScroll([]string{settings.ModeDB, settings.ModeVoltage}, func(selected string, _ selectscroll.Exception) {
		scp.Settings.Dft.DisplayMode = selected
		scp.clearAllDftPersistentLayers()
		for i := range scp.channelViewers {
			scp.channelViewers[i].dftLabel.enableRefresh()
		}
		if scp.dftRaster != nil {
			scp.dftRaster.Refresh()
		}
		scp.SaveSettings()
	}, settings.ModeVoltage)
	modeSelector.SilentSetSelected(scp.Settings.Dft.DisplayMode)
	addToTest(modeSelector, dftModeId, dftTabIndex)

	// Freq Range Selector
	valLabels := []string{"1", "2", "5", "10", "20", "50", "100", "200", "500"}
	unitLabels := []string{settings.UnitHz, settings.UnitKHz, settings.UnitMHz}
	unitVals := map[string]float64{settings.UnitHz: 1, settings.UnitKHz: 1e3, settings.UnitMHz: 1e6}

	updateMaxFreq := func() {
		v, _ := strconv.ParseFloat(scp.dftMaxFreqValSelect.Selected, 64)
		u := unitVals[scp.dftMaxFreqUnitSelect.Selected]
		scp.Settings.Dft.MaxFreq = v * u
		scp.setDftHDivsX()
		scp.clearAllDftPersistentLayers()
		if scp.dftBottomLabelViewer != nil {
			scp.dftBottomLabelViewer.(*frqLabelViewer).enableRefresh()
		}
		if scp.dftRaster != nil {
			fyne.Do(func() { scp.dftRaster.Refresh() })
		}
		scp.SaveSettings()
	}

	scp.dftMaxFreqValSelect = selectscroll.NewSelectScroll(valLabels, func(selected string, ex selectscroll.Exception) {
		if ex == selectscroll.Over {
			scp.dftMaxFreqUnitUp()
			return
		}
		if ex == selectscroll.Under {
			scp.dftMaxFreqUnitDown()
			return
		}
		updateMaxFreq()
	}, "500")
	addToTest(scp.dftMaxFreqValSelect, dftMaxFreqValId, dftTabIndex)

	scp.dftMaxFreqUnitSelect = selectscroll.NewSelectScroll(unitLabels, func(selected string, _ selectscroll.Exception) {
		updateMaxFreq()
	}, "MHz")
	addToTest(scp.dftMaxFreqUnitSelect, dftMaxFreqUnitId, dftTabIndex)

	// Initialize selectors from current MaxFreq
	currentMaxFreq := scp.Settings.Dft.MaxFreq
	bestVal := "1"
	bestUnit := "MHz"
	if currentMaxFreq > 0 {
		// Find best match
		for _, u := range unitLabels {
			uv := unitVals[u]
			for _, v := range valLabels {
				vv, _ := strconv.ParseFloat(v, 64)
				if math.Abs(vv*uv-currentMaxFreq) < 1e-6 {
					bestVal = v
					bestUnit = u
					goto found
				}
			}
		}
	found:
	}
	scp.dftMaxFreqValSelect.SilentSetSelected(bestVal)
	scp.dftMaxFreqUnitSelect.SilentSetSelected(bestUnit)

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
	windowCol.Add(widget.NewLabel("Range:"))
	windowCol.Add(container.NewHBox(scp.dftMaxFreqValSelect, scp.dftMaxFreqUnitSelect))
	windowCol.Add(widget.NewLabel("Sample Rate:"))
	windowCol.Add(container.NewHBox(scp.dftSampleRateSelect, scp.dftSampleUnitSelect))
	windowCol.Add(widget.NewLabel("Bins:"))
	windowCol.Add(binSelector)
	windowCol.Add(scp.binWidthLabel)
	windowCol.Add(scp.dftDataCollectionTimeLabel)

	layout.Add(container.NewVBox(chControls, windowCol))
}
