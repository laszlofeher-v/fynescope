package gui

import (
	"fmt"
	"fynescope/checkcolorpick"
	"fynescope/selectscroll"
	"fynescope/settings"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func (scp *ScpDesc) openVirtualChannelDialog() {
	// Re-focus if already open
	if scp.virtualChWindow != nil {
		scp.virtualChWindow.RequestFocus()
		return
	}

	win := scp.App.NewWindow("Virtual Channels")
	scp.virtualChWindow = win
	win.SetOnClosed(func() {
		scp.virtualChWindow = nil
	})

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
	exprEntry.SetMinRowsVisible(5)
	errorLabel := widget.NewLabel("")
	errorLabel.Hide()

	vRangeSelect := selectscroll.NewSelectScroll(inputRanges, func(s string, _ selectscroll.Exception) {
		if updatingForm || selectedIndex < 0 || selectedIndex >= len(scp.Settings.VirtualChannels) {
			return
		}
		scp.Settings.VirtualChannels[selectedIndex].VRange = vRanges[s]
		scp.refreshRasters()
	}, "±1V")

	offsetEntry := widget.NewEntry()
	offsetEntry.SetText("0.0")
	offsetEntry.OnChanged = func(s string) {
		if updatingForm || selectedIndex < 0 || selectedIndex >= len(scp.Settings.VirtualChannels) {
			return
		}
		var off float32
		if _, err := fmt.Sscanf(s, "%f", &off); err == nil {
			scp.Settings.VirtualChannels[selectedIndex].Offset = off
			scp.refreshRasters()
		}
	}

	invertCheck := widget.NewCheck("Invert", func(b bool) {
		if updatingForm || selectedIndex < 0 || selectedIndex >= len(scp.Settings.VirtualChannels) {
			return
		}
		scp.Settings.VirtualChannels[selectedIndex].Inverted = b
		scp.refreshRasters()
	})

	// CheckColorPick: left-click = toggle enabled, right-click = color picker
	minSz := fyne.NewSize(checkColorPickMinSize, checkColorPickMinSize)
	enableWidget := checkcolorpick.NewCheckColorPick(win, func(v bool, col color.Color) {
		if nrgba, ok := col.(color.NRGBA); ok {
			currentCol = nrgba
		} else {
			r, g, b, a := col.RGBA()
			currentCol = color.NRGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}
		}
		if updatingForm || selectedIndex < 0 || selectedIndex >= len(scp.Settings.VirtualChannels) {
			return
		}
		scp.Settings.VirtualChannels[selectedIndex].Enabled = v
		scp.Settings.VirtualChannels[selectedIndex].Col = [2]color.NRGBA{currentCol, currentCol}
		scp.refreshRasters()
	}, defaultCol, minSz)
	enableWidget.Val = true

	clearForm := func() {
		nameEntry.SetText("")
		exprEntry.SetText("A + B")
		vRangeSelect.SetSelected("±1V")
		offsetEntry.SetText("0.0")
		invertCheck.SetChecked(false)
		currentCol = defaultCol
		enableWidget.Val = true
		enableWidget.SetColor(defaultCol)
		errorLabel.Hide()
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
			offsetEntry.SetText(fmt.Sprintf("%g", vch.Offset))
			invertCheck.SetChecked(vch.Inverted)
			c := vch.Col[0]
			currentCol = c
			enableWidget.Val = vch.Enabled
			enableWidget.SetColor(c)
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

		var off float32
		fmt.Sscanf(offsetEntry.Text, "%f", &off)

		newVCh := settings.VirtualChSettings{
			Name:       nameEntry.Text,
			Expression: exprEntry.Text,
			VRange:     vRanges[vRangeSelect.Selected],
			Offset:     off,
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
		scp.refreshRasters() // immediately remove the trace from the screen
	})

	listButtons := container.NewHBox(newBtn, deleteBtn)
	leftPane := container.NewBorder(listButtons, nil, nil, nil, list)

	formContent := container.NewVBox(
		widget.NewLabel("Name:"), nameEntry,
		widget.NewLabel("Expression (A, B, C, D = physical channels):"), exprEntry,
		errorLabel,
		widget.NewLabel("V/div:"), vRangeSelect,
		widget.NewLabel("Offset (mV):"), offsetEntry,
		invertCheck,
		container.NewHBox(widget.NewLabel("Enable / Color:"), enableWidget),
	)

	buttons := container.NewHBox(layout.NewSpacer(), acceptBtn)

	rightPane := container.NewBorder(nil, buttons, nil, nil, container.NewVScroll(formContent))

	split := container.NewHSplit(leftPane, rightPane)
	split.SetOffset(0.3)

	win.SetContent(split)
	win.Resize(fyne.NewSize(640, 500))
	win.Show()
}
