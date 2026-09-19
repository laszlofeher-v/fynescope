package gui

import (
	"fmt"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// computeCorrelation returns the Pearson correlation coefficient between two
// equal-length sample slices. It returns NaN if either slice has zero variance
// or if the slices are empty / mismatched in length.
func computeCorrelation(a, b []float32) float64 {
	n := len(a)
	if n == 0 || n != len(b) {
		return math.NaN()
	}

	var sumA, sumB float64
	for i := 0; i < n; i++ {
		sumA += float64(a[i])
		sumB += float64(b[i])
	}
	meanA := sumA / float64(n)
	meanB := sumB / float64(n)

	var cov, varA, varB float64
	for i := 0; i < n; i++ {
		da := float64(a[i]) - meanA
		db := float64(b[i]) - meanB
		cov += da * db
		varA += da * da
		varB += db * db
	}
	if varA == 0 || varB == 0 {
		return math.NaN()
	}
	return cov / math.Sqrt(varA*varB)
}

// newCorrPanel builds the cross-correlation UI into the given container.
// panel should be a container.NewMax() so the scroll view fills the tab area.
// Label references are stored in scp.corrLabels / scp.corrStrengthLabels so
// UpdateCorrelation can push live values into them.
func (scp *ScpDesc) newCorrPanel(panel *fyne.Container, undockable bool) {
	vbox := container.New(layout.NewVBoxLayout())

	var undockBtn *widget.Button
	if undockable {
		undockBtn = widget.NewButtonWithIcon("Undock", theme.ViewFullScreenIcon(), func() {
			if scp.corrWindow != nil {
				scp.corrWindow.RequestFocus()
				return
			}
			onWindowClose := func() {
				scp.corrWindow = nil
				scp.dockTab(scp.corrTab)
				scp.controlTab.SelectIndex(ftTabIndex)
				fyne.Do(scp.corrTab.Content.Refresh)
			}
			scp.corrWindow = scp.App.NewWindow("Cross-Correlation")
			winContent := container.NewMax()
			scp.newCorrPanel(winContent, false)
			scp.controlTab.Remove(scp.corrTab)
			scp.corrWindow.SetContent(winContent)
			scp.corrWindow.SetOnClosed(onWindowClose)
			scp.corrWindow.Resize(fyne.NewSize(450, 500))
			scp.controlTab.SelectIndex(ftTabIndex)
			scp.corrWindow.Show()
			fyne.Do(winContent.Refresh)
		})
		addToTest(undockBtn, "corrUndockBtn", corrTabIndex)
		RegisterWidgetHelp(undockBtn, "Undock Correlation", "Opens the cross-correlation panel in an independent floating window.")
		vbox.Add(container.NewHBox(undockBtn))
		vbox.Add(widget.NewSeparator())
	}

	title := widget.NewLabelWithStyle("Cross-Correlation (Pearson r)", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	vbox.Add(title)

	hint := widget.NewLabel("Correlation between enabled channel pairs, updated live.\nRange: −1 (inverse) … 0 (none) … +1 (perfect)")
	hint.Wrapping = fyne.TextWrapWord
	vbox.Add(hint)

	vbox.Add(widget.NewSeparator())

	// Header row
	headerHBox := container.NewHBox(
		corrCell("Pair", 70, true),
		corrCell("r", 100, true),
		corrCell("Strength", 140, true),
	)
	vbox.Add(headerHBox)
	vbox.Add(widget.NewSeparator())

	nch := int(scp.channelCount)
	for i := 0; i < nch; i++ {
		for j := i + 1; j < nch; j++ {
			ci, cj := i, j // capture

			nameI := string(rune('A' + ci))
			nameJ := string(rune('A' + cj))

			col := scp.Settings.Channels[ci].Col[scp.Settings.ChannelColorIndex]
			pairLabel := canvas.NewText(nameI+"↔"+nameJ, col)
			pairLabel.TextStyle.Bold = true
			pairLabel.TextSize = 13

			rLabel := widget.NewLabel("---")
			rLabel.Alignment = fyne.TextAlignLeading

			strengthLabel := widget.NewLabel("")

			scp.corrLabels[ci][cj] = rLabel
			scp.corrStrengthLabels[ci][cj] = strengthLabel

			row := container.NewHBox(
				container.New(layout.NewGridWrapLayout(fyne.NewSize(70, 24)), pairLabel),
				container.New(layout.NewGridWrapLayout(fyne.NewSize(100, 24)), rLabel),
				container.New(layout.NewGridWrapLayout(fyne.NewSize(140, 24)), strengthLabel),
			)
			vbox.Add(row)
		}
	}

	// Use VScroll so the list is scrollable on small screens; Max on the panel
	// makes the scroll fill the entire available tab space.
	panel.Add(container.NewVScroll(vbox))
}

// corrCell is a helper that returns a label in a fixed-width cell for column alignment.
func corrCell(text string, width float32, bold bool) *fyne.Container {
	lbl := widget.NewLabel(text)
	lbl.TextStyle.Bold = bold
	return container.New(layout.NewGridWrapLayout(fyne.NewSize(width, 24)), lbl)
}

// corrStrength returns a human-readable label for a Pearson r value.
func corrStrength(r float64) string {
	abs := math.Abs(r)
	switch {
	case abs >= 0.9:
		return "Very strong"
	case abs >= 0.7:
		return "Strong"
	case abs >= 0.5:
		return "Moderate"
	case abs >= 0.3:
		return "Weak"
	default:
		return "Negligible"
	}
}

// UpdateCorrelation computes Pearson r for every enabled channel pair and
// pushes the results to the UI labels. Safe to call from any goroutine.
func (scp *ScpDesc) UpdateCorrelation() {
	// Guard: panel not yet built (all labels are nil).
	nch := int(scp.channelCount)
	if nch < 2 {
		return
	}
	if scp.corrLabels[0][1] == nil {
		return
	}

	type result struct {
		i, j int
		r     float64
	}
	results := make([]result, 0, nch*(nch-1)/2)

	for i := 0; i < nch; i++ {
		for j := i + 1; j < nch; j++ {
			if scp.corrLabels[i][j] == nil {
				continue
			}
			chI := &scp.Settings.Channels[i]
			chJ := &scp.Settings.Channels[j]
			if !chI.Enabled || !chJ.Enabled {
				results = append(results, result{i, j, math.NaN()})
				continue
			}
			if i >= len(scp.displayBuffers) || j >= len(scp.displayBuffers) {
				results = append(results, result{i, j, math.NaN()})
				continue
			}
			r := computeCorrelation(scp.displayBuffers[i], scp.displayBuffers[j])
			results = append(results, result{i, j, r})
		}
	}

	fyne.Do(func() {
		for _, res := range results {
			lbl := scp.corrLabels[res.i][res.j]
			str := scp.corrStrengthLabels[res.i][res.j]
			if lbl == nil {
				continue
			}
			if math.IsNaN(res.r) {
				lbl.SetText("N/A")
				if str != nil {
					str.SetText("—")
				}
			} else {
				lbl.SetText(fmt.Sprintf("%.4f", res.r))
				if str != nil {
					str.SetText(corrStrength(res.r))
				}
			}
		}
	})
}
