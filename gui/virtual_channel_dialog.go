package gui

import (
	"fmt"
	"fynescope/checkcolorpick"
	"fynescope/disp7"
	"fynescope/selectscroll"
	"fynescope/settings"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const vchUndockLabel = "Undock"

// openVirtualChannelDialog selects the Virtual Channels tab in the side panel.
// If the tab was undocked into a floating window, focus that window instead.
func (scp *ScpDesc) openVirtualChannelDialog() {
	if scp.virtualChWindow != nil {
		scp.virtualChWindow.RequestFocus()
		return
	}
	if scp.controlTab == nil || scp.vchTab == nil {
		return
	}
	scp.dockTab(scp.vchTab)
}

// buildVirtualChannelContent constructs the Virtual Channels editor UI and returns
// it as a CanvasObject suitable for embedding in a tab. Layout is a single narrow
// column so it fits comfortably in the side panel.
func (scp *ScpDesc) buildVirtualChannelContent(undockable bool) fyne.CanvasObject {
	var selectedIndex int = -1
	var updatingForm bool

	// Default color for new virtual channels
	defaultCol := color.NRGBA{R: 0xff, G: 0x80, B: 0x00, A: 0xff}
	currentCol := defaultCol

	list := widget.NewList(
		func() int { return len(scp.Settings.VirtualChannels) },
		func() fyne.CanvasObject { return widget.NewLabel("Virtual Channel 123456") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(scp.Settings.VirtualChannels[i].Name)
		},
	)

	nameEntry := widget.NewEntry()
	exprEntry := widget.NewMultiLineEntry()
	exprEntry.SetMinRowsVisible(3)
	errorLabel := widget.NewLabel("")
	errorLabel.Hide()

	vRangeSelect := selectscroll.NewSelectScroll(inputRanges, func(s string, _ selectscroll.Exception) {
		if updatingForm || selectedIndex < 0 || selectedIndex >= len(scp.Settings.VirtualChannels) {
			return
		}
		scp.Settings.VirtualChannels[selectedIndex].VRange = vRanges[s]
		if selectedIndex < len(scp.ftVChannelLabels) {
			scp.ftVChannelLabels[selectedIndex].enableRefresh()
		}
		if selectedIndex < len(scp.tzVChannelLabels) {
			scp.tzVChannelLabels[selectedIndex].enableRefresh()
		}
		scp.refreshRasters()
	}, "±50V")

	fontScale := float32(0.7) * scp.getScreenScale()
	var err error
	scp.vchMaxV, err = disp7.NewCustomDisp7Array(5, 3, 99999, -99999, disp7.Signed,
		disp7.NoTrailingZeroes, scp.Window, defaultCol,
		disp7.ReaOnly, fontScale*disp7.DefaultDigitWidth,
		fontScale*disp7.DeafultDigitHeight, 1, fontScale*disp7.DefaultVCursorSpace, " Max:", " V ")
	if err != nil {
		panic(err.Error() + " error from disp7.NewCustomDisp7Array (vch maxV)")
	}

	scp.vchMinV, err = disp7.NewCustomDisp7Array(5, 3, 99999, -99999, disp7.Signed,
		disp7.NoTrailingZeroes, scp.Window, defaultCol,
		disp7.ReaOnly, fontScale*disp7.DefaultDigitWidth,
		fontScale*disp7.DeafultDigitHeight, 1, fontScale*disp7.DefaultVCursorSpace, " Min:", " V ")
	if err != nil {
		panic(err.Error() + " error from disp7.NewCustomDisp7Array (vch minV)")
	}

	scp.vchFrq, err = disp7.NewCustomDisp7Array(4, 2, 9999, 0, disp7.UnSigned,
		disp7.NoTrailingZeroes, scp.Window, defaultCol,
		disp7.ReaOnly, fontScale*disp7.DefaultDigitWidth,
		fontScale*disp7.DeafultDigitHeight, 1, fontScale*disp7.DefaultVCursorSpace, "Frq:", " MHz")
	if err != nil {
		panic(err.Error() + " error from disp7.NewCustomDisp7Array (vch frq)")
	}

	scp.vchPeriod, err = disp7.NewCustomDisp7Array(4, 2, 9999, 0, disp7.UnSigned,
		disp7.NoTrailingZeroes, scp.Window, defaultCol,
		disp7.ReaOnly, fontScale*disp7.DefaultDigitWidth,
		fontScale*disp7.DeafultDigitHeight, 1, fontScale*disp7.DefaultVCursorSpace, "  T:", " ms")
	if err != nil {
		panic(err.Error() + " error from disp7.NewCustomDisp7Array (vch period)")
	}
	scp.vchPeriod.SilentSetValue(0)

	invertCheck := widget.NewCheck("Invert", func(b bool) {
		if updatingForm || selectedIndex < 0 || selectedIndex >= len(scp.Settings.VirtualChannels) {
			return
		}
		scp.Settings.VirtualChannels[selectedIndex].Inverted = b
		if selectedIndex < len(scp.ftVChannelLabels) {
			scp.ftVChannelLabels[selectedIndex].enableRefresh()
		}
		if selectedIndex < len(scp.tzVChannelLabels) {
			scp.tzVChannelLabels[selectedIndex].enableRefresh()
		}
		scp.refreshRasters()
	})

	// CheckColorPick: left-click = toggle enabled, right-click = color picker
	minSz := fyne.NewSize(checkColorPickMinSize, checkColorPickMinSize)
	enableWidget := checkcolorpick.NewCheckColorPick(scp.Window, func(v bool, col color.Color) {
		if nrgba, ok := col.(color.NRGBA); ok {
			currentCol = nrgba
		} else {
			r, g, b, a := col.RGBA()
			currentCol = color.NRGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}
		}
		scp.vchMinV.SetOncolor(currentCol)
		scp.vchMaxV.SetOncolor(currentCol)
		scp.vchFrq.SetOncolor(currentCol)
		scp.vchPeriod.SetOncolor(currentCol)
		if updatingForm || selectedIndex < 0 || selectedIndex >= len(scp.Settings.VirtualChannels) {
			return
		}

		// If toggling enabled state, we must repartition the screen to add/remove the label from the draw loop.
		if scp.Settings.VirtualChannels[selectedIndex].Enabled != v {
			setFlag(scp.repartition)
			setFlag(scp.tzRepartition)
		} else {
			// Just a color change, refresh the label.
			if selectedIndex < len(scp.ftVChannelLabels) {
				scp.ftVChannelLabels[selectedIndex].enableRefresh()
			}
			if selectedIndex < len(scp.tzVChannelLabels) {
				scp.tzVChannelLabels[selectedIndex].enableRefresh()
			}
		}

		scp.Settings.VirtualChannels[selectedIndex].Enabled = v
		scp.Settings.VirtualChannels[selectedIndex].Col = [2]color.NRGBA{currentCol, currentCol}
		scp.refreshRasters()
	}, defaultCol, minSz)
	enableWidget.Val = true

	extraFields := container.NewVBox(
		widget.NewLabel("V/div:"), vRangeSelect,
		container.NewHBox(scp.vchMaxV, scp.vchFrq),
		container.NewHBox(scp.vchMinV, scp.vchPeriod),
		invertCheck,
		container.NewHBox(widget.NewLabel("Enable / Color:"), enableWidget),
	)
	extraFields.Hide()

	clearForm := func() {
		nameEntry.SetText("")
		exprEntry.SetText("A + B")
		vRangeSelect.SetSelected("±1V")
		invertCheck.SetChecked(false)
		currentCol = defaultCol
		enableWidget.Val = true
		enableWidget.SetColor(defaultCol)
		scp.vchMeasureIndex = -1
		scp.vchMinV.SilentSetValue(0)
		scp.vchMaxV.SilentSetValue(0)
		scp.vchFrq.SilentSetValue(0)
		scp.vchPeriod.SilentSetValue(0)
		scp.vchMinV.SetOncolor(defaultCol)
		scp.vchMaxV.SetOncolor(defaultCol)
		scp.vchFrq.SetOncolor(defaultCol)
		scp.vchPeriod.SetOncolor(defaultCol)
		errorLabel.Hide()
		extraFields.Hide()
	}

	var updateForm = func(idx int) {
		updatingForm = true
		defer func() { updatingForm = false }()

		if idx >= 0 && idx < len(scp.Settings.VirtualChannels) {
			vch := scp.Settings.VirtualChannels[idx]
			nameEntry.SetText(vch.Name)
			exprEntry.SetText(vch.Expression)
			if s, ok := rangeEnumToString[vch.VRange]; ok {
				vRangeSelect.SetSelected(s)
			}
			invertCheck.SetChecked(vch.Inverted)
			c := vch.Col[0]
			currentCol = c
			enableWidget.Val = vch.Enabled
			enableWidget.SetColor(c)
			scp.vchMeasureIndex = idx
			scp.vchMinV.SetOncolor(c)
			scp.vchMaxV.SetOncolor(c)
			scp.vchFrq.SetOncolor(c)
			scp.vchPeriod.SetOncolor(c)
			extraFields.Show()
		} else {
			clearForm()
		}
		errorLabel.Hide()
	}

	list.OnSelected = func(id widget.ListItemID) {
		selectedIndex = int(id)
		updateForm(selectedIndex)
	}

	newBtn := widget.NewButton("New", func() {
		list.UnselectAll()
		selectedIndex = -1
		clearForm()
		extraFields.Show()
	})

	acceptBtn := widget.NewButton("Accept", func() {
		if nameEntry.Text == "" {
			errorLabel.SetText("Error: channel name must not be empty")
			errorLabel.Show()
			return
		}

		eng, err := CompileVirtualChannel(exprEntry.Text)
		if err != nil {
			errorLabel.SetText(fmt.Sprintf("Error: %v", err))
			errorLabel.Show()
			return
		}

		newVCh := settings.VirtualChSettings{
			Name:       nameEntry.Text,
			Expression: exprEntry.Text,
			VRange:     vRanges[vRangeSelect.Selected],
			Inverted:   invertCheck.Checked,
			Enabled:    enableWidget.Val,
			Col:        [2]color.NRGBA{currentCol, currentCol},
		}

		// Find existing channel with same name to overwrite
		found := -1
		for i, vch := range scp.Settings.VirtualChannels {
			if vch.Name == newVCh.Name {
				found = i
				break
			}
		}

		var newSelectedIndex int
		if found >= 0 {
			scp.Settings.VirtualChannels[found] = newVCh
			for len(scp.virtualChannelEngines) <= found {
				scp.virtualChannelEngines = append(scp.virtualChannelEngines, nil)
			}
			scp.virtualChannelEngines[found] = eng
			newSelectedIndex = found
		} else if selectedIndex >= 0 && selectedIndex < len(scp.Settings.VirtualChannels) {
			scp.Settings.VirtualChannels[selectedIndex] = newVCh
			scp.virtualChannelEngines[selectedIndex] = eng
			newSelectedIndex = selectedIndex
		} else {
			scp.Settings.VirtualChannels = append(scp.Settings.VirtualChannels, newVCh)
			scp.virtualChannelEngines = append(scp.virtualChannelEngines, eng)
			newSelectedIndex = len(scp.Settings.VirtualChannels) - 1
		}

		// Extend buffer slices for the new virtual channel slot.
		needed := int(scp.channelCount) + len(scp.Settings.VirtualChannels)
		for len(scp.displayBuffers) < needed {
			scp.displayBuffers = append(scp.displayBuffers, nil)
		}
		for len(scp.ftPersistentLayers) < needed {
			scp.ftPersistentLayers = append(scp.ftPersistentLayers, nil)
		}
		for len(scp.dftPersistentLayers) < needed {
			scp.dftPersistentLayers = append(scp.dftPersistentLayers, nil)
		}

		list.Refresh()
		list.Select(newSelectedIndex)
		errorLabel.Hide()
		setFlag(scp.repartition)
		setFlag(scp.tzRepartition)
		scp.refreshRasters()
	})

	deleteBtn := widget.NewButton("Delete", func() {
		if selectedIndex < 0 || selectedIndex >= len(scp.Settings.VirtualChannels) {
			return
		}
		// Remove from settings
		scp.Settings.VirtualChannels = append(
			scp.Settings.VirtualChannels[:selectedIndex],
			scp.Settings.VirtualChannels[selectedIndex+1:]...,
		)
		// Remove engine
		if selectedIndex < len(scp.virtualChannelEngines) {
			scp.virtualChannelEngines = append(
				scp.virtualChannelEngines[:selectedIndex],
				scp.virtualChannelEngines[selectedIndex+1:]...,
			)
		}
		// Splice the deleted channel's buffer slot out of the middle so that
		// any remaining virtual channels stay correctly aligned at channelCount+i.
		bufIdx := int(scp.channelCount) + selectedIndex
		if bufIdx < len(scp.displayBuffers) {
			scp.displayBuffers = append(scp.displayBuffers[:bufIdx], scp.displayBuffers[bufIdx+1:]...)
		}
		if bufIdx < len(scp.ftPersistentLayers) {
			scp.ftPersistentLayers = append(scp.ftPersistentLayers[:bufIdx], scp.ftPersistentLayers[bufIdx+1:]...)
		}
		if bufIdx < len(scp.dftPersistentLayers) {
			scp.dftPersistentLayers = append(scp.dftPersistentLayers[:bufIdx], scp.dftPersistentLayers[bufIdx+1:]...)
		}

		selectedIndex = -1
		list.UnselectAll()
		list.Refresh()
		clearForm()
		setFlag(scp.repartition)
		setFlag(scp.tzRepartition)
		scp.refreshRasters() // immediately remove the trace from the screen
	})

	// Undock button: pops the content into a floating window and removes the tab.
	var undockBtn *widget.Button
	if undockable {
		undockBtn = widget.NewButtonWithIcon(vchUndockLabel, theme.ViewFullScreenIcon(), func() {
			onWindowClose := func() {
				scp.virtualChWindow = nil
				scp.dockTab(scp.vchTab)
				scp.controlTab.SelectIndex(ftTabIndex)
				fyne.Do(scp.vchTab.Content.Refresh)
			}
			scp.virtualChWindow = scp.App.NewWindow("Virtual Channels")
			winContent := scp.buildVirtualChannelContent(false)
			scp.controlTab.Remove(scp.vchTab)
			scp.virtualChWindow.SetContent(winContent)
			scp.virtualChWindow.SetOnClosed(onWindowClose)
			scp.virtualChWindow.Resize(fyne.NewSize(500, 600))
			scp.controlTab.SelectIndex(ftTabIndex)
			scp.virtualChWindow.Show()
			fyne.Do(winContent.Refresh)
		})
	}

	// Single-column layout: channel list in top pane, form fields in bottom pane.
	// VSplit lets the user resize the divider as needed.
	listBox := container.NewBorder(nil, nil, nil, nil, list)

	formContent := container.NewVBox(
		widget.NewLabel("Name:"), nameEntry,
		widget.NewLabel("Expression (A,B,C,D = physical ch.):"), exprEntry,
		errorLabel,
		extraFields,
	)

	actionRow := container.NewHBox(newBtn, deleteBtn, layout.NewSpacer(), acceptBtn)
	var topRow fyne.CanvasObject
	if undockable {
		topRow = container.NewHBox(layout.NewSpacer(), undockBtn)
	}

	topPane := container.NewBorder(topRow, actionRow, nil, nil, listBox)
	bottomPane := container.NewVScroll(formContent)

	split := container.NewVSplit(topPane, bottomPane)
	split.SetOffset(0.35)

	return split
}
