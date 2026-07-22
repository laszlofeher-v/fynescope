package gui

import (
	"math"
	"fynescope/genericps"
	"fynescope/selectscroll"
	"fynescope/settings"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func toHz(value float64, unit string) float64 {
	switch unit {
	case settings.UnitHz:
		return value
	case settings.UnitKHz:
		return value * 1e3
	case settings.UnitMHz:
		return value * 1e6
	default:
		return value
	}
}

func fromHz(hz float64) (float64, string) {
	if hz >= 1e6 {
		return hz / 1e6, settings.UnitMHz
	} else if hz >= 1e3 {
		return hz / 1e3, settings.UnitKHz
	} else {
		return hz, settings.UnitHz
	}
}

// applyFIR applies a 2-tap or 3-tap FIR filter on the buffer.
func applyFIR(input []float32, coeffs []float64) []float32 {
	out := make([]float32, len(input))
	if len(coeffs) == 2 {
		b0, b1 := float32(coeffs[0]), float32(coeffs[1])
		for i := 0; i < len(input); i++ {
			x0 := input[i]
			x1 := input[0]
			if i > 0 {
				x1 = input[i-1]
			}
			out[i] = b0*x0 + b1*x1
		}
	} else if len(coeffs) == 3 {
		b0, b1, b2 := float32(coeffs[0]), float32(coeffs[1]), float32(coeffs[2])
		for i := 0; i < len(input); i++ {
			x0 := input[i]
			x1 := input[0]
			if i > 0 {
				x1 = input[i-1]
			}
			x2 := input[0]
			if i > 1 {
				x2 = input[i-2]
			}
			out[i] = b0*x0 + b1*x1 + b2*x2
		}
	} else {
		copy(out, input)
	}
	return out
}

func (scp *ScpDesc) applyDigitalFilters(chIdx int, buf []float32, samplingTimeInterval float64) {
	if chIdx >= len(scp.Settings.Channels) {
		return
	}
	if len(buf) == 0 {
		return
	}
	chSettings := &scp.Settings.Channels[chIdx]
	filter := &chSettings.DigitalFilter

	fs := 1.0 / samplingTimeInterval
	if fs <= 0 {
		return
	}

	applyFilterFwd := func(data []float32) {
		// 1. Lowpass (First-Order IIR)
		if filter.LowpassEnabled {
			fc := filter.LowpassFc
			if fc < fs/2 {
				alpha := 1.0 - math.Exp(-2.0*math.Pi*fc*samplingTimeInterval)
				y := float64(data[0])
				for i := 0; i < len(data); i++ {
					y = y + alpha*(float64(data[i])-y)
					data[i] = float32(y)
				}
			}
		}

		// 2. Highpass (First-Order IIR)
		if filter.HighpassEnabled {
			fc := filter.HighpassFc
			if fc < fs/2 {
				alpha := math.Exp(-2.0*math.Pi*fc*samplingTimeInterval)
				x1 := float64(data[0])
				y1 := float64(0)
				b0 := (1.0 + alpha) / 2.0
				b1 := -b0
				a1 := -alpha
				for i := 0; i < len(data); i++ {
					x := float64(data[i])
					out := b0*x + b1*x1 - a1*y1
					x1 = x
					y1 = out
					data[i] = float32(out)
				}
			} else {
				for i := range data {
					data[i] = 0
				}
			}
		}

		// 3. Bandpass (Second-Order IIR Biquad)
		if filter.BandpassEnabled {
			fc1 := filter.BandpassFc1
			fc2 := filter.BandpassFc2
			f0 := math.Sqrt(fc1 * fc2)
			bw := fc2 - fc1
			if bw <= 0 {
				bw = 1.0
			}
			omega0 := 2.0 * math.Pi * f0 * samplingTimeInterval
			if omega0 < math.Pi {
				q := f0 / bw
				if q < 0.1 {
					q = 0.1
				}
				alpha := math.Sin(omega0) / (2.0 * q)
				// Constant 0 dB peak gain bandpass
				b0 := alpha
				b1 := 0.0
				b2 := -alpha
				a0 := 1.0 + alpha
				a1 := -2.0 * math.Cos(omega0)
				a2 := 1.0 - alpha

				b0 /= a0
				b1 /= a0
				b2 /= a0
				a1 /= a0
				a2 /= a0

				x1, x2 := float64(data[0]), float64(data[0])
				y1, y2 := float64(0), float64(0)
				for i := 0; i < len(data); i++ {
					x := float64(data[i])
					out := b0*x + b1*x1 + b2*x2 - a1*y1 - a2*y2
					x2 = x1
					x1 = x
					y2 = y1
					y1 = out
					data[i] = float32(out)
				}
			}
		}

		// 4. Bandstop (Second-Order IIR Biquad)
		if filter.BandstopEnabled {
			fc1 := filter.BandstopFc1
			fc2 := filter.BandstopFc2
			f0 := math.Sqrt(fc1 * fc2)
			bw := fc2 - fc1
			if bw <= 0 {
				bw = 1.0
			}
			omega0 := 2.0 * math.Pi * f0 * samplingTimeInterval
			if omega0 < math.Pi {
				q := f0 / bw
				if q < 0.1 {
					q = 0.1
				}
				alpha := math.Sin(omega0) / (2.0 * q)
				b0 := 1.0
				b1 := -2.0 * math.Cos(omega0)
				b2 := 1.0
				a0 := 1.0 + alpha
				a1 := -2.0 * math.Cos(omega0)
				a2 := 1.0 - alpha

				b0 /= a0
				b1 /= a0
				b2 /= a0
				a1 /= a0
				a2 /= a0

				x1, x2 := float64(data[0]), float64(data[0])
				y1, y2 := float64(data[0]), float64(data[0])
				for i := 0; i < len(data); i++ {
					x := float64(data[i])
					out := b0*x + b1*x1 + b2*x2 - a1*y1 - a2*y2
					x2 = x1
					x1 = x
					y2 = y1
					y1 = out
					data[i] = float32(out)
				}
			}
		}
	}

	n := len(buf)
	// Padlen is typically 3 * max(order), but for low cutoffs we need more.
	// We use an odd extension of length equal to the buffer size to give it ample time to settle.
	padlen := n
	if padlen > 1000 {
		padlen = 1000
	}
	padded := make([]float32, n+2*padlen)

	// Odd extension at the start: x[-i] = 2*x[0] - x[i]
	firstVal := buf[0]
	for i := 0; i < padlen; i++ {
		padded[padlen-1-i] = 2*firstVal - buf[i]
	}
	// Copy original data
	copy(padded[padlen:padlen+n], buf)
	// Odd extension at the end: x[n-1+i] = 2*x[n-1] - x[n-1-i]
	lastVal := buf[n-1]
	for i := 0; i < padlen; i++ {
		padded[padlen+n+i] = 2*lastVal - buf[n-1-i]
	}

	// Apply filter forward
	applyFilterFwd(padded)

	// Apply filter backward if ZeroPhase
	if filter.ZeroPhaseEnabled {
		for i, j := 0, len(padded)-1; i < j; i, j = i+1, j-1 {
			padded[i], padded[j] = padded[j], padded[i]
		}
		applyFilterFwd(padded)
		for i, j := 0, len(padded)-1; i < j; i, j = i+1, j-1 {
			padded[i], padded[j] = padded[j], padded[i]
		}
	}

	// Copy back the middle valid section
	copy(buf, padded[padlen:padlen+n])
}

func (scp *ScpDesc) newDigitalFilterPanel(panel *fyne.Container) {
	channelTabs := container.NewAppTabs()

	for i := 0; i < int(scp.channelCount); i++ {
		chIdx := i
		chStr := string(rune('A' + chIdx))
		chSettings := &scp.Settings.Channels[chIdx]
		col := chSettings.Col[scp.Settings.ChannelColorIndex]

		chBox := container.NewVBox()

		notify := func() {
			go func() {
				scp.SaveSettings()
				scp.notifyDigitalFilter(chIdx)
				scp.refreshRasters()
			}()
		}

		lblCh := canvas.NewText("Digital Filters for Channel "+chStr, col)
		lblCh.TextStyle.Bold = true
		lblCh.TextSize = 16
		chBox.Add(container.NewHBox(lblCh))

		zeroPhaseCheck := widget.NewCheck("Zero Phase Delay (FiltFilt)", func(checked bool) {
			chSettings.DigitalFilter.ZeroPhaseEnabled = checked
			notify()
		})
		zeroPhaseCheck.SetChecked(chSettings.DigitalFilter.ZeroPhaseEnabled)
		chBox.Add(zeroPhaseCheck)

		// 1. Lowpass filter
		var lpControls *fyne.Container
		lpCheck := widget.NewCheck("Enable Lowpass Filter", func(checked bool) {
			chSettings.DigitalFilter.LowpassEnabled = checked
			if lpControls != nil {
				if checked {
					lpControls.Show()
				} else {
					lpControls.Hide()
				}
				chBox.Refresh()
			}
			notify()
		})
		lpCheck.SetChecked(chSettings.DigitalFilter.LowpassEnabled)

		lpVal, lpUnit := fromHz(chSettings.DigitalFilter.LowpassFc)
		lpEntry := widget.NewEntry()
		lpEntry.SetText(strconv.FormatFloat(lpVal, 'f', -1, 64))

		lpUnitSelect := selectscroll.NewSelectScroll([]string{settings.UnitHz, settings.UnitKHz, settings.UnitMHz}, func(s string, exc selectscroll.Exception) {
			text := lpEntry.Text
			go func() {
				val, _ := strconv.ParseFloat(text, 64)
				chSettings.DigitalFilter.LowpassFc = toHz(val, s)
				notify()
			}()
		}, "kHz")
		lpUnitSelect.SetSelected(lpUnit)

		lpEntry.OnChanged = func(s string) {
			unit := lpUnitSelect.Selected
			go func() {
				v, err := strconv.ParseFloat(s, 64)
				if err == nil {
					chSettings.DigitalFilter.LowpassFc = toHz(v, unit)
					notify()
				}
			}()
		}
		lpEntryContainer := container.New(layout.NewGridWrapLayout(fyne.NewSize(80, 35)), lpEntry)
		lpControls = container.NewHBox(widget.NewLabel("Cutoff Frequency:"), lpEntryContainer, lpUnitSelect)
		if chSettings.DigitalFilter.LowpassEnabled {
			lpControls.Show()
		} else {
			lpControls.Hide()
		}

		// 2. Highpass filter
		var hpControls *fyne.Container
		hpCheck := widget.NewCheck("Enable Highpass Filter", func(checked bool) {
			chSettings.DigitalFilter.HighpassEnabled = checked
			if hpControls != nil {
				if checked {
					hpControls.Show()
				} else {
					hpControls.Hide()
				}
				chBox.Refresh()
			}
			notify()
		})
		hpCheck.SetChecked(chSettings.DigitalFilter.HighpassEnabled)

		hpVal, hpUnit := fromHz(chSettings.DigitalFilter.HighpassFc)
		hpEntry := widget.NewEntry()
		hpEntry.SetText(strconv.FormatFloat(hpVal, 'f', -1, 64))

		hpUnitSelect := selectscroll.NewSelectScroll([]string{settings.UnitHz, settings.UnitKHz, settings.UnitMHz}, func(s string, exc selectscroll.Exception) {
			text := hpEntry.Text
			go func() {
				val, _ := strconv.ParseFloat(text, 64)
				chSettings.DigitalFilter.HighpassFc = toHz(val, s)
				notify()
			}()
		}, "kHz")
		hpUnitSelect.SetSelected(hpUnit)

		hpEntry.OnChanged = func(s string) {
			unit := hpUnitSelect.Selected
			go func() {
				v, err := strconv.ParseFloat(s, 64)
				if err == nil {
					chSettings.DigitalFilter.HighpassFc = toHz(v, unit)
					notify()
				}
			}()
		}
		hpEntryContainer := container.New(layout.NewGridWrapLayout(fyne.NewSize(80, 35)), hpEntry)
		hpControls = container.NewHBox(widget.NewLabel("Cutoff Frequency:"), hpEntryContainer, hpUnitSelect)
		if chSettings.DigitalFilter.HighpassEnabled {
			hpControls.Show()
		} else {
			hpControls.Hide()
		}

		// 3. Bandpass filter
		var bpControls *fyne.Container
		bpCheck := widget.NewCheck("Enable Bandpass Filter", func(checked bool) {
			chSettings.DigitalFilter.BandpassEnabled = checked
			if bpControls != nil {
				if checked {
					bpControls.Show()
				} else {
					bpControls.Hide()
				}
				chBox.Refresh()
			}
			notify()
		})
		bpCheck.SetChecked(chSettings.DigitalFilter.BandpassEnabled)

		bpVal1, bpUnit1 := fromHz(chSettings.DigitalFilter.BandpassFc1)
		bpEntry1 := widget.NewEntry()
		bpEntry1.SetText(strconv.FormatFloat(bpVal1, 'f', -1, 64))

		bpUnitSelect1 := selectscroll.NewSelectScroll([]string{settings.UnitHz, settings.UnitKHz, settings.UnitMHz}, func(s string, exc selectscroll.Exception) {
			text := bpEntry1.Text
			go func() {
				val, _ := strconv.ParseFloat(text, 64)
				chSettings.DigitalFilter.BandpassFc1 = toHz(val, s)
				notify()
			}()
		}, "kHz")
		bpUnitSelect1.SetSelected(bpUnit1)

		bpEntry1.OnChanged = func(s string) {
			unit := bpUnitSelect1.Selected
			go func() {
				v, err := strconv.ParseFloat(s, 64)
				if err == nil {
					chSettings.DigitalFilter.BandpassFc1 = toHz(v, unit)
					notify()
				}
			}()
		}
		bpEntryContainer1 := container.New(layout.NewGridWrapLayout(fyne.NewSize(80, 35)), bpEntry1)

		bpVal2, bpUnit2 := fromHz(chSettings.DigitalFilter.BandpassFc2)
		bpEntry2 := widget.NewEntry()
		bpEntry2.SetText(strconv.FormatFloat(bpVal2, 'f', -1, 64))

		bpUnitSelect2 := selectscroll.NewSelectScroll([]string{settings.UnitHz, settings.UnitKHz, settings.UnitMHz}, func(s string, exc selectscroll.Exception) {
			text := bpEntry2.Text
			go func() {
				val, _ := strconv.ParseFloat(text, 64)
				chSettings.DigitalFilter.BandpassFc2 = toHz(val, s)
				notify()
			}()
		}, "kHz")
		bpUnitSelect2.SetSelected(bpUnit2)

		bpEntry2.OnChanged = func(s string) {
			unit := bpUnitSelect2.Selected
			go func() {
				v, err := strconv.ParseFloat(s, 64)
				if err == nil {
					chSettings.DigitalFilter.BandpassFc2 = toHz(v, unit)
					notify()
				}
			}()
		}
		bpEntryContainer2 := container.New(layout.NewGridWrapLayout(fyne.NewSize(80, 35)), bpEntry2)
		bpControls = container.NewVBox(
			widget.NewLabel("Lower Cutoff:"), bpEntryContainer1, bpUnitSelect1,
			widget.NewLabel("Upper Cutoff:"), bpEntryContainer2, bpUnitSelect2,
		)
		if chSettings.DigitalFilter.BandpassEnabled {
			bpControls.Show()
		} else {
			bpControls.Hide()
		}

		// 4. Bandstop filter
		var bsControls *fyne.Container
		bsCheck := widget.NewCheck("Enable Bandstop Filter", func(checked bool) {
			chSettings.DigitalFilter.BandstopEnabled = checked
			if bsControls != nil {
				if checked {
					bsControls.Show()
				} else {
					bsControls.Hide()
				}
				chBox.Refresh()
			}
			notify()
		})
		bsCheck.SetChecked(chSettings.DigitalFilter.BandstopEnabled)

		bsVal1, bsUnit1 := fromHz(chSettings.DigitalFilter.BandstopFc1)
		bsEntry1 := widget.NewEntry()
		bsEntry1.SetText(strconv.FormatFloat(bsVal1, 'f', -1, 64))

		bsUnitSelect1 := selectscroll.NewSelectScroll([]string{settings.UnitHz, settings.UnitKHz, settings.UnitMHz}, func(s string, exc selectscroll.Exception) {
			text := bsEntry1.Text
			go func() {
				val, _ := strconv.ParseFloat(text, 64)
				chSettings.DigitalFilter.BandstopFc1 = toHz(val, s)
				notify()
			}()
		}, "kHz")
		bsUnitSelect1.SetSelected(bsUnit1)

		bsEntry1.OnChanged = func(s string) {
			unit := bsUnitSelect1.Selected
			go func() {
				v, err := strconv.ParseFloat(s, 64)
				if err == nil {
					chSettings.DigitalFilter.BandstopFc1 = toHz(v, unit)
					notify()
				}
			}()
		}
		bsEntryContainer1 := container.New(layout.NewGridWrapLayout(fyne.NewSize(80, 35)), bsEntry1)

		bsVal2, bsUnit2 := fromHz(chSettings.DigitalFilter.BandstopFc2)
		bsEntry2 := widget.NewEntry()
		bsEntry2.SetText(strconv.FormatFloat(bsVal2, 'f', -1, 64))

		bsUnitSelect2 := selectscroll.NewSelectScroll([]string{settings.UnitHz, settings.UnitKHz, settings.UnitMHz}, func(s string, exc selectscroll.Exception) {
			text := bsEntry2.Text
			go func() {
				val, _ := strconv.ParseFloat(text, 64)
				chSettings.DigitalFilter.BandstopFc2 = toHz(val, s)
				notify()
			}()
		}, "kHz")
		bsUnitSelect2.SetSelected(bsUnit2)

		bsEntry2.OnChanged = func(s string) {
			unit := bsUnitSelect2.Selected
			go func() {
				v, err := strconv.ParseFloat(s, 64)
				if err == nil {
					chSettings.DigitalFilter.BandstopFc2 = toHz(v, unit)
					notify()
				}
			}()
		}
		bsEntryContainer2 := container.New(layout.NewGridWrapLayout(fyne.NewSize(80, 35)), bsEntry2)
		bsControls = container.NewVBox(
			widget.NewLabel("Lower Cutoff:"), bsEntryContainer1, bsUnitSelect1,
			widget.NewLabel("Upper Cutoff:"), bsEntryContainer2, bsUnitSelect2,
		)
		if chSettings.DigitalFilter.BandstopEnabled {
			bsControls.Show()
		} else {
			bsControls.Hide()
		}

		chBox.Add(widget.NewSeparator())
		chBox.Add(lpCheck)
		chBox.Add(lpControls)
		chBox.Add(widget.NewSeparator())
		chBox.Add(hpCheck)
		chBox.Add(hpControls)
		chBox.Add(widget.NewSeparator())
		chBox.Add(bpCheck)
		chBox.Add(bpControls)
		chBox.Add(widget.NewSeparator())
		chBox.Add(bsCheck)
		chBox.Add(bsControls)

		addToTest(lpCheck, "lpCheck"+chStr, filterTabIndex)
		addToTest(lpEntry, "lpEntry"+chStr, filterTabIndex)
		addToTest(lpUnitSelect, "lpUnitSelect"+chStr, filterTabIndex)
		addToTest(hpCheck, "hpCheck"+chStr, filterTabIndex)
		addToTest(hpEntry, "hpEntry"+chStr, filterTabIndex)
		addToTest(hpUnitSelect, "hpUnitSelect"+chStr, filterTabIndex)
		addToTest(bpCheck, "bpCheck"+chStr, filterTabIndex)
		addToTest(bpEntry1, "bpEntry1"+chStr, filterTabIndex)
		addToTest(bpUnitSelect1, "bpUnitSelect1"+chStr, filterTabIndex)
		addToTest(bpEntry2, "bpEntry2"+chStr, filterTabIndex)
		addToTest(bpUnitSelect2, "bpUnitSelect2"+chStr, filterTabIndex)
		addToTest(bsCheck, "bsCheck"+chStr, filterTabIndex)
		addToTest(bsEntry1, "bsEntry1"+chStr, filterTabIndex)
		addToTest(bsUnitSelect1, "bsUnitSelect1"+chStr, filterTabIndex)
		addToTest(bsEntry2, "bsEntry2"+chStr, filterTabIndex)
		addToTest(bsUnitSelect2, "bsUnitSelect2"+chStr, filterTabIndex)

		tabItem := container.NewTabItem("Ch "+chStr, container.NewScroll(chBox))
		channelTabs.Append(tabItem)
		scp.notifyDigitalFilter(chIdx)
	}

	if scp.Settings.Window.FilterActiveTab >= 0 && scp.Settings.Window.FilterActiveTab < len(channelTabs.Items) {
		channelTabs.SelectIndex(scp.Settings.Window.FilterActiveTab)
	}
	channelTabs.OnSelected = func(item *container.TabItem) {
		scp.Settings.Window.FilterActiveTab = channelTabs.SelectedIndex()
		scp.SaveSettings()
	}

	panel.Add(channelTabs)
}

func (scp *ScpDesc) notifyDigitalFilter(chIdx int) {
	scp.refreshFilterWarning(genericps.ChannelId(chIdx))
}

