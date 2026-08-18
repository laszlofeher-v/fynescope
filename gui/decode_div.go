package gui

import (
	"strconv"

	"fynescope/selectscroll"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (scp *ScpDesc) buildDecodeContent(undockable bool) fyne.CanvasObject {
	settings := &scp.Settings.Decode

	enableCheck := widget.NewCheck("Enable Decoder", func(b bool) {
		settings.Enabled = b
	})
	enableCheck.SetChecked(settings.Enabled)

	protocolSelect := selectscroll.NewSelectScroll([]string{"UART", "SPI", "I2C"}, func(s string, _ selectscroll.Exception) {
		if settings.Protocol != s {
			settings.Protocol = s
			scp.refreshDecodeTab() // Refresh to show/hide Ch2
		}
	}, "Protocol")
	protocolSelect.SetSelected(settings.Protocol)

	chNames := []string{"Ch A", "Ch B", "Ch C", "Ch D"}

	ch1Select := selectscroll.NewSelectScroll(chNames, func(s string, _ selectscroll.Exception) {
		for i, name := range chNames {
			if s == name {
				if settings.Channel1 != i {
					settings.Channel1 = i
					scp.refreshDecodeTab()
				}
				break
			}
		}
	}, "Channel 1")
	if settings.Channel1 >= 0 && settings.Channel1 < len(chNames) {
		ch1Select.SetSelected(chNames[settings.Channel1])
	}

	ch2Select := selectscroll.NewSelectScroll(chNames, func(s string, _ selectscroll.Exception) {
		for i, name := range chNames {
			if s == name {
				settings.Channel2 = i
				break
			}
		}
	}, "Channel 2")
	if settings.Channel2 >= 0 && settings.Channel2 < len(chNames) {
		ch2Select.SetSelected(chNames[settings.Channel2])
	}

	chNamesWithNone := []string{"None", "Ch A", "Ch B", "Ch C", "Ch D"}

	ch3Select := selectscroll.NewSelectScroll(chNamesWithNone, func(s string, _ selectscroll.Exception) {
		if s == "None" {
			settings.Channel3 = -1
			return
		}
		for i, name := range chNames {
			if s == name {
				settings.Channel3 = i
				break
			}
		}
	}, "Channel 3")
	if settings.Channel3 >= 0 && settings.Channel3 < len(chNames) {
		ch3Select.SetSelected(chNames[settings.Channel3])
	} else {
		ch3Select.SetSelected("None")
	}

	ch4Select := selectscroll.NewSelectScroll(chNamesWithNone, func(s string, _ selectscroll.Exception) {
		if s == "None" {
			settings.Channel4 = -1
			return
		}
		for i, name := range chNames {
			if s == name {
				settings.Channel4 = i
				break
			}
		}
	}, "Channel 4")
	if settings.Channel4 >= 0 && settings.Channel4 < len(chNames) {
		ch4Select.SetSelected(chNames[settings.Channel4])
	} else {
		ch4Select.SetSelected("None")
	}

	csActiveHighCheck := widget.NewCheck("", func(b bool) {
		settings.CSActiveHigh = b
	})
	csActiveHighCheck.SetChecked(settings.CSActiveHigh)

	baudrates := []string{"300", "600", "1200", "2400", "4800", "9600", "14400", "19200", "38400", "57600", "115200", "128000", "256000", "921600"}
	baudSelect := selectscroll.NewSelectScroll(baudrates, func(s string, _ selectscroll.Exception) {
		if val, err := strconv.Atoi(s); err == nil {
			settings.BaudRate = val
		}
	}, "Baud Rate")
	baudSelect.SetSelected(strconv.Itoa(settings.BaudRate))

	invertCheck := widget.NewCheck("", func(b bool) {
		settings.Invert = b
	})
	invertCheck.SetChecked(settings.Invert)

	showBitstartsCheck := widget.NewCheck("", func(b bool) {
		settings.ShowBitstarts = b
	})
	showBitstartsCheck.SetChecked(settings.ShowBitstarts)

	chRange := scp.Settings.Channels[settings.Channel1].VRange

	thresholdMv := scp.adcToMv(float64(settings.Threshold), chRange)
	thresholdEntry := widget.NewEntry()
	thresholdEntry.SetText(strconv.FormatFloat(thresholdMv/1000.0, 'f', -1, 64))
	thresholdEntry.OnChanged = func(s string) {
		if valV, err := strconv.ParseFloat(s, 64); err == nil {
			currRange := scp.Settings.Channels[settings.Channel1].VRange
			settings.Threshold = int16(scp.mvToAdc(int32(valV*1000.0), currRange))
		}
	}

	hysteresisMv := scp.adcToMv(float64(settings.Hysteresis), chRange)
	hysteresisEntry := widget.NewEntry()
	hysteresisEntry.SetText(strconv.FormatFloat(hysteresisMv/1000.0, 'f', -1, 64))
	hysteresisEntry.OnChanged = func(s string) {
		if valV, err := strconv.ParseFloat(s, 64); err == nil {
			currRange := scp.Settings.Channels[settings.Channel1].VRange
			settings.Hysteresis = int32(scp.mvToAdc(int32(valV*1000.0), currRange))
		}
	}

	ch1Label := "Channel 1 (Rx/CLK)"
	if settings.Protocol == "I2C" {
		ch1Label = "Channel 1 (SCL)"
	}

	form := widget.NewForm(
		widget.NewFormItem("Protocol", protocolSelect),
		widget.NewFormItem(ch1Label, ch1Select),
	)

	if settings.Protocol == "SPI" {
		form.Append("Channel 2 (MOSI)", ch2Select)
		form.Append("Channel 3 (MISO)", ch3Select)
		form.Append("Channel 4 (CS)", ch4Select)
		form.Append("CS Active High", csActiveHighCheck)
	}

	if settings.Protocol == "UART" {
		form.Append("Channel 2 (Tx)", ch2Select)
	}

	if settings.Protocol == "I2C" {
		form.Append("Channel 2 (SDA)", ch2Select)
	}

	if settings.Protocol == "UART" {
		form.Append("Baud / Clock (Hz)", baudSelect)
		dataBitsOptions := []string{"5", "6", "7", "8", "9"}
		dataBitsSelect := selectscroll.NewSelectScroll(dataBitsOptions, func(s string, _ selectscroll.Exception) {
			if val, err := strconv.Atoi(s); err == nil {
				settings.DataBits = val
			}
		}, "Data Bits")
		dataBitsSelect.SetSelected(strconv.Itoa(settings.DataBits))

		stopBitsOptions := []string{"1", "1.5", "2"}
		stopBitsSelect := selectscroll.NewSelectScroll(stopBitsOptions, func(s string, _ selectscroll.Exception) {
			settings.StopBits = s
		}, "Stop Bits")
		stopBitsSelect.SetSelected(settings.StopBits)

		parityOptions := []string{"None", "Even", "Odd", "Mark", "Space"}
		paritySelect := selectscroll.NewSelectScroll(parityOptions, func(s string, _ selectscroll.Exception) {
			settings.Parity = s
		}, "Parity")
		paritySelect.SetSelected(settings.Parity)

		bitOrderOptions := []string{"LSB First", "MSB First"}
		bitOrderSelect := selectscroll.NewSelectScroll(bitOrderOptions, func(s string, _ selectscroll.Exception) {
			settings.BitOrder = s
		}, "Bit Order")
		if settings.BitOrder == "" {
			settings.BitOrder = "LSB First"
		}
		bitOrderSelect.SetSelected(settings.BitOrder)

		form.Append("Data Bits", dataBitsSelect)
		form.Append("Stop Bits", stopBitsSelect)
		form.Append("Parity", paritySelect)
		form.Append("Bit Order", bitOrderSelect)
	}
	form.Append("Invert", invertCheck)
	if settings.Protocol == "UART" {
		form.Append("Show Bit Lines", showBitstartsCheck)
	}
	form.Append("Threshold (V)", thresholdEntry)
	form.Append("Hysteresis (V)", hysteresisEntry)

	var undockBtn *widget.Button
	if undockable {
		undockBtn = widget.NewButtonWithIcon("Undock", theme.ViewFullScreenIcon(), func() {
			if scp.decodeWindow == nil {
				scp.decodeWindow = scp.App.NewWindow("Decoder")
				scp.controlTab.Remove(scp.decodeTab)

				// Create completely new UI tree for the undocked window
				newLayout := scp.buildDecodeContent(false)
				scrollContainer := container.NewScroll(newLayout)
				scp.decodeWindow.SetContent(scrollContainer)
				scp.decodeWindow.Resize(fyne.NewSize(350, 600))

				scp.decodeWindow.SetOnClosed(func() {
					// Re-dock
					scp.decodeLayout.RemoveAll()
					scp.decodeLayout.Add(scp.buildDecodeContent(true))
					scp.decodeTab.Content = scp.decodeLayout
					scp.dockTab(scp.decodeTab)
					scp.decodeWindow = nil
				})

				scp.decodeWindow.Show()
			} else {
				scp.decodeWindow.Close()
			}
		})
		addToTest(undockBtn, "decodeUndockBtn", decodeTabIndex)
	}

	topRow := container.NewHBox(layout.NewSpacer())
	if undockable {
		topRow = container.NewHBox(layout.NewSpacer(), undockBtn)
	}

	return container.NewVScroll(container.New(layout.NewVBoxLayout(),
		topRow,
		enableCheck,
		form,
	))
}

func (scp *ScpDesc) refreshDecodeTab() {
	if scp.decodeLayout != nil {
		scp.decodeLayout.Objects = []fyne.CanvasObject{scp.buildDecodeContent(scp.decodeWindow == nil)}
		scp.decodeLayout.Refresh()
	}

	// If undocked, we need to refresh the undocked window's layout too
	if scp.decodeWindow != nil {
		newLayout := scp.buildDecodeContent(false)
		scp.decodeWindow.SetContent(container.NewScroll(newLayout))
	}
}
