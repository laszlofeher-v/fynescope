package gui

import (
	"fynescope/disp7"
	"fynescope/genericps"
	"fynescope/selectscroll"
	"fynescope/settings"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (scp *ScpDesc) applyDemoDigitalGenSettings() {
	if scp.psControl != nil && scp.psControl.Con != nil {
		set := scp.Settings.DigitalDemoGenPanel
		go func(set settings.DigitalDemoGenSettings) {
			_ = scp.psControl.Con.SetDemoDigitalGen(
				set.Port0Enabled,
				set.Port1Enabled,
				set.Frequency,
				set.Direction,
				set.Encoding,
				set.Mode,
				set.BitDelay,
			)
		}(set)
		scp.SaveSettings()
	}
}

func (scp *ScpDesc) newDemoDigGenPanel(undockable bool) (box *fyne.Container, err error) {
	settings := &scp.Settings.DigitalDemoGenPanel

	var undockButton *widget.Button
	undock := "Undock"
	if undockable {
		undockButton = widget.NewButtonWithIcon(undock, theme.ViewFullScreenIcon(), func() {
			onWindowClose := func() {
				scp.digGenWindow.Hide()
				undockButton.Text = undock
				undockButton.Show()
				scp.digGenTab = container.NewTabItem(tabNames[digGenTabIndex], scp.digGenTab.Content)
				scp.dockTab(scp.digGenTab)
				scp.controlTab.SelectIndex(ftTabIndex)
				fyne.Do(scp.digGenTab.Content.Refresh)
			}
			scp.digGenWindow = scp.App.NewWindow("digGen")
			var digGenPanel *fyne.Container
			digGenPanel, err = scp.newDemoDigGenPanel(false)
			if err != nil {
				return
			}
			digGenControls := container.New(layout.NewVBoxLayout())
			digGenControls.Add(digGenPanel)
			scp.controlTab.Remove(scp.digGenTab)
			scp.digGenWindow.SetContent(digGenControls)
			scp.digGenWindow.SetOnClosed(onWindowClose)
			scp.controlTab.SelectIndex(ftTabIndex)
			scp.digGenWindow.Show()

			fyne.Do(undockButton.Refresh)
			fyne.Do(digGenControls.Refresh)
		})
	}

	// Port Checkboxes
	port0Check := widget.NewCheck("Port0", func(checked bool) {
		settings.Port0Enabled = checked
		scp.applyDemoDigitalGenSettings()
	})
	port0Check.SetChecked(settings.Port0Enabled)

	port1Check := widget.NewCheck("Port1", func(checked bool) {
		settings.Port1Enabled = checked
		scp.applyDemoDigitalGenSettings()
	})
	port1Check.SetChecked(settings.Port1Enabled)

	portGroup := container.NewHBox(port0Check, port1Check)

	// Frequency
	freqLabel := widget.NewLabel("Frequency:")
	freqDisp, err := disp7.NewCustomDisp7Array(8, 0, 10000000, 1,
		disp7.UnSigned, disp7.NoTrailingZeroes, scp.Window,
		color.White, disp7.ReadWrite, disp7.DefaultDigitWidth,
		disp7.DeafultDigitHeight, 1, disp7.DefaultVCursorSpace, "", " Hz")
	if err != nil {
		return nil, err
	}
	freqDisp.SilentSetValue(int(settings.Frequency))
	freqDisp.OnChanged = func(f float64) {
		settings.Frequency = f
		scp.applyDemoDigitalGenSettings()
	}

	// Direction Selector
	dirOptions := []string{"Up", "Down"}
	dirSelect := selectscroll.NewSelectScroll(dirOptions, func(s string, ex selectscroll.Exception) {
		if s == "Up" {
			settings.Direction = genericps.DigitalDemoGenDirectionUp
		} else {
			settings.Direction = genericps.DigitalDemoGenDirectionDown
		}
		scp.applyDemoDigitalGenSettings()
	}, "Direction")
	if settings.Direction == genericps.DigitalDemoGenDirectionDown {
		dirSelect.SilentSetSelected("Down")
	} else {
		dirSelect.SilentSetSelected("Up")
	}

	var updateVisibility func()

	// Encoding Selector
	encOptions := []string{"Binary", "Gray"}
	encSelect := selectscroll.NewSelectScroll(encOptions, func(s string, ex selectscroll.Exception) {
		if s == "Binary" {
			settings.Encoding = genericps.DigitalDemoGenEncodingBinary
		} else {
			settings.Encoding = genericps.DigitalDemoGenEncodingGray
		}
		if updateVisibility != nil {
			updateVisibility()
		}
		scp.applyDemoDigitalGenSettings()
	}, "Encoding")
	if settings.Encoding == genericps.DigitalDemoGenEncodingGray {
		encSelect.SilentSetSelected("Gray")
	} else {
		encSelect.SilentSetSelected("Binary")
	}

	// Mode Selector
	modeOptions := []string{"Synchronous", "Asynchronous"}
	modeSelect := selectscroll.NewSelectScroll(modeOptions, func(s string, ex selectscroll.Exception) {
		if s == "Synchronous" {
			settings.Mode = genericps.DigitalDemoGenModeSynchronous
		} else {
			settings.Mode = genericps.DigitalDemoGenModeAsynchronous
		}
		if updateVisibility != nil {
			updateVisibility()
		}
		scp.applyDemoDigitalGenSettings()
	}, "Mode")
	if settings.Mode == genericps.DigitalDemoGenModeAsynchronous {
		modeSelect.SilentSetSelected("Asynchronous")
	} else {
		modeSelect.SilentSetSelected("Synchronous")
	}

	// Bit Delay
	bitDelayLabel := widget.NewLabel("Bit Delay:")
	bitDelayDisp, err := disp7.NewCustomDisp7Array(10, 9, 1000000000, 0,
		disp7.UnSigned, disp7.NoTrailingZeroes, scp.Window,
		color.White, disp7.ReadWrite, disp7.DefaultDigitWidth,
		disp7.DeafultDigitHeight, 1, disp7.DefaultVCursorSpace, "", " s")
	if err != nil {
		return nil, err
	}
	bitDelayDisp.SilentSetValue(int(settings.BitDelay * 1000000000.0))
	bitDelayDisp.OnChanged = func(f float64) {
		settings.BitDelay = f / 1000000000.0
		scp.applyDemoDigitalGenSettings()
	}

	updateVisibility = func() {
		if settings.Encoding == genericps.DigitalDemoGenEncodingGray {
			modeSelect.Hide()
			settings.Mode = genericps.DigitalDemoGenModeSynchronous
			modeSelect.SilentSetSelected("Synchronous")
		} else {
			modeSelect.Show()
		}

		if settings.Mode == genericps.DigitalDemoGenModeSynchronous {
			bitDelayLabel.Hide()
			bitDelayDisp.Hide()
		} else {
			bitDelayLabel.Show()
			bitDelayDisp.Show()
		}
	}
	updateVisibility()

	if undockable {
		box = container.NewVBox(
			undockButton,
			portGroup,
			freqLabel,
			freqDisp,
			dirSelect,
			encSelect,
			modeSelect,
			bitDelayLabel,
			bitDelayDisp,
		)
	} else {
		box = container.NewVBox(
			portGroup,
			freqLabel,
			freqDisp,
			dirSelect,
			encSelect,
			modeSelect,
			bitDelayLabel,
			bitDelayDisp,
		)
	}
	return box, nil
}
