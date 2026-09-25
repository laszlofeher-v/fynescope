//go:build multi

package gui

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// openRemoteChannelsWindow displays the remote channels in a dedicated floating/popup window,
// structured similarly to the f(t) channel control panel.
func (scp *ScpDesc) openRemoteChannelsWindow() {
	if scp.remoteChannelsWindow != nil {
		scp.updateRemoteChannelsWindow()
		scp.remoteChannelsWindow.Show()
		scp.remoteChannelsWindow.RequestFocus()
		return
	}

	w := scp.App.NewWindow("Remote Channels")
	scp.remoteChannelsWindow = w
	scp.remoteChannelsContainer = container.NewVBox()

	scp.updateRemoteChannelsWindow()

	scroll := container.NewVScroll(scp.remoteChannelsContainer)

	w.SetContent(container.NewPadded(scroll))
	w.Resize(fyne.NewSize(380, 520))
	w.SetOnClosed(func() {
		scp.remoteChannelsWindow = nil
		scp.remoteChannelsContainer = nil
	})
	w.Show()
}

// updateRemoteChannelsWindow refreshes the cards in the remote channels popup window.
func (scp *ScpDesc) updateRemoteChannelsWindow() {
	if scp.remoteChannelsContainer == nil {
		return
	}

	scp.remoteChannelsContainer.RemoveAll()

	scp.remoteChannelsMu.RLock()
	count := len(scp.remoteChannels)
	channelsCopy := make([]RemoteChannelDesc, count)
	copy(channelsCopy, scp.remoteChannels)
	scp.remoteChannelsMu.RUnlock()

	if count == 0 {
		emptyLabel := widget.NewLabel("No remote channels reported yet.")
		scp.remoteChannelsContainer.Add(emptyLabel)
		scp.remoteChannelsContainer.Refresh()
		return
	}

	for i := range channelsCopy {
		chIdx := i
		card := scp.buildRemoteChannelCard(chIdx, &channelsCopy[chIdx])
		scp.remoteChannelsContainer.Add(card)
		scp.remoteChannelsContainer.Add(widget.NewSeparator())
	}

	scp.remoteChannelsContainer.Refresh()
}

// buildRemoteChannelCard builds the UI card for a single remote channel, modeled after f(t) channel controls.
func (scp *ScpDesc) buildRemoteChannelCard(idx int, rch *RemoteChannelDesc) *fyne.Container {
	colorBox := canvas.NewRectangle(rch.Color)
	colorBox.SetMinSize(fyne.NewSize(16, 16))

	titleLabel := widget.NewLabelWithStyle(rch.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	header := container.NewHBox(colorBox, titleLabel, layout.NewSpacer())

	enableCheck := widget.NewCheck("Show on Raster", func(checked bool) {
		scp.remoteChannelsMu.Lock()
		if idx < len(scp.remoteChannels) {
			scp.remoteChannels[idx].Enabled = checked
		}
		scp.remoteChannelsMu.Unlock()

		setFlag(scp.repartition)
		scp.refreshRasters()
	})
	enableCheck.SetChecked(rch.Enabled)

	vRangeStr := rangeEnumToString[rch.VRange]
	if vRangeStr == "" {
		vRangeStr = "Default"
	}
	rangeLabel := widget.NewLabel(fmt.Sprintf("Range: %s", vRangeStr))

	couplingStr := "DC"
	if rch.CoupleType == 0 { // AC
		couplingStr = "AC"
	}
	coupleLabel := widget.NewLabel(fmt.Sprintf("Couple: %s", couplingStr))

	rowParams := container.NewHBox(rangeLabel, widget.NewLabel("|"), coupleLabel)

	offsetStr := fmt.Sprintf("Offset: %.2fV", rch.Offset)
	offsetLabel := widget.NewLabel(offsetStr)

	invertStr := "Inverted: No"
	if rch.Inverted {
		invertStr = "Inverted: Yes"
	}
	invertLabel := widget.NewLabel(invertStr)

	rowExtra := container.NewHBox(offsetLabel, widget.NewLabel("|"), invertLabel)

	resetOffsetBtn := widget.NewButtonWithIcon("Reset V-Offset", theme.ViewRefreshIcon(), func() {
		scp.remoteChannelsMu.Lock()
		if idx < len(scp.remoteChannels) {
			scp.remoteChannels[idx].DisplayVOffset = 0
		}
		scp.remoteChannelsMu.Unlock()
		scp.clearAllFtPersistentLayers()
		scp.refreshRasters()
	})

	cycleColorBtn := widget.NewButtonWithIcon("Color", theme.ColorPaletteIcon(), func() {
		scp.remoteChannelsMu.Lock()
		if idx < len(scp.remoteChannels) {
			// Cycle to next color in palette
			palette := []color.NRGBA{
				{R: 255, G: 165, B: 0, A: 255},   // Orange
				{R: 180, G: 80, B: 240, A: 255},  // Purple
				{R: 0, G: 210, B: 210, A: 255},   // Cyan
				{R: 255, G: 105, B: 180, A: 255}, // Hot Pink
				{R: 50, G: 205, B: 50, A: 255},   // Lime
				{R: 255, G: 215, B: 0, A: 255},   // Gold
			}
			curCol := scp.remoteChannels[idx].Color
			nextCol := palette[0]
			for pIdx, pCol := range palette {
				if pCol.R == curCol.R && pCol.G == curCol.G && pCol.B == curCol.B {
					nextCol = palette[(pIdx+1)%len(palette)]
					break
				}
			}
			scp.remoteChannels[idx].Color = nextCol
			colorBox.FillColor = nextCol
			colorBox.Refresh()
		}
		scp.remoteChannelsMu.Unlock()
		scp.refreshRasters()
	})

	actions := container.NewHBox(enableCheck, layout.NewSpacer(), cycleColorBtn, resetOffsetBtn)

	return container.NewVBox(
		header,
		rowParams,
		rowExtra,
		actions,
	)
}
