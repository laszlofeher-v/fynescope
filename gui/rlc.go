package gui

import (
	"fynescope/genericps"
	"fynescope/selectscroll"
	"fynescope/settings"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func (scp *ScpDesc) newRlcPanel(panel *fyne.Container) {
	vbox := container.New(layout.NewVBoxLayout())

	// Add title label for clarity
	title := widget.NewLabelWithStyle("Simulator RLC Filter", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	vbox.Add(title)

	for i := 0; i < int(scp.channelCount); i++ {
		chIdx := i // capture loop variable
		chSettings := &scp.Settings.Channels[chIdx]
		chStr := string(rune('A' + i))

		text := "Ch " + chStr + ":"
		if scp.isDigitalFilterEnabled(genericps.ChannelId(chIdx)) {
			text += " ⚠️"
		}
		channelLabel := canvas.NewText(text, chSettings.Col[scp.Settings.ChannelColorIndex])
		channelLabel.TextStyle.Bold = true
		scp.channelViewers[chIdx].rlcNameLabel = channelLabel

		notifySim := func() {
			if scp.psControl != nil && scp.psControl.Con != nil && scp.psControl.Con.ID == genericps.SimId {
				scp.psControl.Con.SetSimRlcFilter(genericps.ChannelId(chIdx), chSettings.RlcFilter.GeneratorSource, chSettings.RlcFilter.Enabled,
					chSettings.RlcFilter.Type, chSettings.RlcFilter.R, chSettings.RlcFilter.RUnit,
					chSettings.RlcFilter.L, chSettings.RlcFilter.LUnit,
					chSettings.RlcFilter.C, chSettings.RlcFilter.CUnit)
			}
		}

		genSourceOptions := []string{}
		for j := 0; j < int(scp.channelCount); j++ {
			genSourceOptions = append(genSourceOptions, "Gen "+string(rune('A'+j)))
		}

		genSourceSelect := selectscroll.NewSelectScroll(genSourceOptions, func(s string, exc selectscroll.Exception) {
			for j, opt := range genSourceOptions {
				if opt == s {
					chSettings.RlcFilter.GeneratorSource = genericps.ChannelId(j)
					break
				}
			}
			scp.SaveSettings()
			notifySim()
		}, "")

		// Ensure it defaults to something valid
		if int(chSettings.RlcFilter.GeneratorSource) >= int(scp.channelCount) {
			chSettings.RlcFilter.GeneratorSource = genericps.ChannelId(chIdx)
		}
		genSourceSelect.SetSelected("Gen " + string(rune('A'+int(chSettings.RlcFilter.GeneratorSource))))
		addToTest(genSourceSelect, "rlcGenSource"+chStr)

		filterTypes := []string{settings.RlcFilterTypeDisabled, "Lowpass RC", "Lowpass RL", "Highpass RC", "Highpass RL", "Lowpass LC", "Highpass LC"}
		typeSelect := selectscroll.NewSelectScroll(filterTypes, func(s string, exc selectscroll.Exception) {
			chSettings.RlcFilter.Type = s
			chSettings.RlcFilter.Enabled = (s != settings.RlcFilterTypeDisabled)
			scp.SaveSettings()
			notifySim()
		}, "")
		if chSettings.RlcFilter.Type == "" {
			chSettings.RlcFilter.Type = "Lowpass RC"
		}
		typeSelect.SetSelected(chSettings.RlcFilter.Type)

		rUnits := []string{"mΩ", "Ω", "kΩ", "MΩ"}
		lUnits := []string{"µH", "mH", "H"}
		cUnits := []string{"pF", "nF", "µF", "mF"}

		// Helper to build entry + unit selector
		buildInput := func(labelStr string, value *float64, unit *string, units []string, valId string, unitId string) *fyne.Container {
			lbl := widget.NewLabel(labelStr)
			entry := widget.NewEntry()

			// Format correctly without trailing zeros
			entry.SetText(strconv.FormatFloat(*value, 'f', -1, 64))

			unitSelect := selectscroll.NewSelectScroll(units, func(s string, exc selectscroll.Exception) {
				*unit = s
				scp.SaveSettings()
				notifySim()
			}, "")
			if *unit == "" {
				*unit = units[1]
			}
			unitSelect.SetSelected(*unit)

			entry.OnChanged = func(s string) {
				v, err := strconv.ParseFloat(s, 64)
				if err == nil {
					*value = v
					scp.SaveSettings()
					notifySim()
				}
			}

			// Keep entry width manageable
			entryContainer := container.New(layout.NewGridWrapLayout(fyne.NewSize(80, 35)), entry)

			addToTest(entry, valId)
			addToTest(unitSelect, unitId)
			return container.NewHBox(lbl, entryContainer, unitSelect)
		}

		rBox := buildInput("R:", &chSettings.RlcFilter.R, &chSettings.RlcFilter.RUnit, rUnits, rlcRId+chStr, rlcRUnitId+chStr)
		lBox := buildInput("L:", &chSettings.RlcFilter.L, &chSettings.RlcFilter.LUnit, lUnits, rlcLId+chStr, rlcLUnitId+chStr)
		cBox := buildInput("C:", &chSettings.RlcFilter.C, &chSettings.RlcFilter.CUnit, cUnits, rlcCId+chStr, rlcCUnitId+chStr)

		// Hide irrelevant inputs based on filter type
		updateVisibility := func() {
			t := chSettings.RlcFilter.Type
			if t == settings.RlcFilterTypeDisabled {
				rBox.Hide()
				lBox.Hide()
				cBox.Hide()
			} else if strings.Contains(t, "RC") {
				rBox.Show()
				cBox.Show()
				lBox.Hide()
			} else if strings.Contains(t, "RL") {
				rBox.Show()
				lBox.Show()
				cBox.Hide()
			} else if strings.Contains(t, "LC") {
				lBox.Show()
				cBox.Show()
				rBox.Hide()
			}
		}

		// Initial visibility update and tie to type selector
		updateVisibility()
		originalOnChanged := typeSelect.OnChanged
		typeSelect.OnChanged = func(s string) {
			originalOnChanged(s)
			updateVisibility()
		}

		row1 := container.NewHBox(
			channelLabel,
			widget.NewLabel("Source:"),
			genSourceSelect,
		)

		controls := container.NewVBox(
			row1,
			typeSelect,
			rBox,
			lBox,
			cBox,
		)

		addToTest(typeSelect, rlcTypeId+chStr)

		notifySim() // initialize sim with current settings on startup
		vbox.Add(controls)
	}

	panel.Add(vbox)
}
