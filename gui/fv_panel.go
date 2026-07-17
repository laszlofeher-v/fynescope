package gui

import (
	"fynescope/genericps"
	"fynescope/selectscroll"
	"fynescope/settings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (scp *ScpDesc) newFvPanel(panel *fyne.Container) {
	vbox := container.New(layout.NewVBoxLayout())
	var xChecks []*widget.Check

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
		scp.channelViewers[chIndex].fvNameLabel = label

		// Enabled Checkbox
		enabledCheck := widget.NewCheck("Enabled", func(b bool) {
			scp.EnableChannel(chIndex, b)
		})
		enabledCheck.SetChecked(scp.Settings.Channels[chIndex].Enabled)
		scp.channelViewers[chIndex].enableChecks = append(scp.channelViewers[chIndex].enableChecks, enabledCheck)
		addToTest(enabledCheck, fvEnableId+chName)

		// X-Axis Check
		xCheck := widget.NewCheck("X-Axis", nil)
		if scp.Settings.Channels[chIndex].FvMode == settings.FvArgument {
			xCheck.SetChecked(true)
		}
		xChecks = append(xChecks, xCheck)
		addToTest(xCheck, fvXCheckId+chName)

		// Range Selector
		rangesEnum, _ := scp.psControl.ChannelRanges(chIndex)
		var ranges []string
		for _, r := range rangesEnum {
			ranges = append(ranges, inputRanges[r])
		}
		vRange := selectscroll.NewSelectScroll(ranges, func(option string, e selectscroll.Exception) {
			scp.changeChannelRange(chIndex, option)
		}, "+500m")
		scp.channelViewers[chIndex].vRangeSelects = append(scp.channelViewers[chIndex].vRangeSelects, vRange)
		addToTest(vRange, fvVRangeId+chName)

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
		addToTest(x10Check, fvX10Id+chName)

		// Arrange settings to minimize width (f(t) style)
		row1 := container.New(layout.NewHBoxLayout(), label, enabledCheck, xCheck)
		row2 := container.New(layout.NewHBoxLayout(), widget.NewLabel("Range:"), vRange, x10Check)

		chBox := container.New(layout.NewVBoxLayout(), row1, row2)
		if i > 0 {
			vbox.Add(layout.NewSpacer())
		}
		vbox.Add(chBox)
	}

	// Set up radio behavior for X-Axis checks
	for i := range xChecks {
		idx := i
		xChecks[idx].OnChanged = func(b bool) {
			if b {
				// Uncheck others
				for j, c := range xChecks {
					if idx != j {
						c.SetChecked(false)
					}
				}
				// Update settings
				for j := 0; j < int(scp.channelCount); j++ {
					if idx == j {
						scp.Settings.Channels[j].FvMode = settings.FvArgument
					} else {
						scp.Settings.Channels[j].FvMode = settings.FvValue
					}
				}
			} else {
				// If unchecked, set this one to FvValue as well
				scp.Settings.Channels[idx].FvMode = settings.FvValue
			}
			scp.refreshRasters()
			scp.SaveSettings()
		}
	}

	panel.Add(vbox)
}
