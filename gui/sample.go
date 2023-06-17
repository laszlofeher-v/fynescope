package gui

import (
	"fmt"
	"psgo/gui/inter"

	//	"psgo/ps2000a"

	//	"psgo/ps2000a"
	//	"sort"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	// "fyne.io/fyne/v2/widget"
)

var (
	//	times       = []string{"1", "2", "5", "10", "20", "50", "100", "200", "500", "1000", "2000", "5000"}
	times = []string{"5000", "2000", "1000", "500", "200", "100", "50", "20", "10", "5", "2", "1"}
	// timesSubset = times[:9]
	timesSubset = times[3:]
	units       []string
	// tu          []float64 = []float64{1.0, 1.0e3, 1.0e6, 1.0e9}
	tu []float64 = []float64{1.0e9, 1.0e6, 1.0e3, 1.0}
)

// TODO language change must change this
func (ps *PsDesc) newUnitList() []string {
	if units == nil {
		units = []string{ps.Str(inter.SDiv), ps.Str(inter.MsDiv), ps.Str(inter.UsDiv), ps.Str(inter.NsDiv)}
	}
	return units
}

func (ps *PsDesc) newSampleSettings() *fyne.Container {
	var (
		timeSelect, unitSelect *selectScroll
	)
	unitChanged := func(option string) {
		fmt.Println("Units changed ", option, timeSelect.SelectedIndex(), timeSelect.OverflowIndex)
		ps.timeUnit = tu[unitSelect.SelectedIndex()]
		// ps.ChannelSettings[chIndex].Range = v
		if option == ps.Str(inter.SDiv) {
			// fmt.Println("S selected")
			timeSelect.Options = times
		} else {
			// fmt.Println("S unselected")
			selectIdex := timeSelect.SelectedIndex()
			if selectIdex > 8 {
				timeSelect.Selected = times[selectIdex-3]
			}
			timeSelect.Options = timesSubset
		}
		timeSelect.Refresh()
		unitSelect.Refresh()
		ps.raster.Refresh()
	}
	timeChanged := func(option string) {
		// fmt.Println("time changed ", option, timeSelect.SelectedIndex(), timeSelect.OverflowIndex)
		switch {
		case timeSelect.OverflowIndex > timeSelect.SelectedIndex():
			uIndex := unitSelect.SelectedIndex() + 1
			if uIndex < len(units) {
				unitSelect.SetSelectedIndex(uIndex)
				timeSelect.Selected = timeSelect.Options[0]
				unitSelect.Refresh()
			}
		case timeSelect.OverflowIndex < timeSelect.SelectedIndex():
			uIndex := unitSelect.SelectedIndex()
			fmt.Println("uIndex ", uIndex)
			if uIndex > 0 {
				unitSelect.SetSelectedIndex(uIndex - 1)
				timeSelect.Selected = timeSelect.Options[len(timeSelect.Options)-1]
				unitSelect.Refresh()
			}
		default:
		}
		//vstr := strconv.FormatFloat(v, 'f', 1, 32)
		intTimeDiv, _ := strconv.Atoi(timeSelect.Selected)
		ps.timeDiv = float64(intTimeDiv)
		timeSelect.OverflowIndex = timeSelect.SelectedIndex()
		timeSelect.Refresh()
		ps.raster.Refresh()
	}
	timeSelect = newSelectScroll(timesSubset, timeChanged, strconv.Itoa(5000)) // first create widgets
	unitSelect = newSelectScroll(ps.newUnitList(), unitChanged, ps.Str(inter.MsDiv))
	timeSelect.SetSelected("1") // these will call both unitSelect and timeSelect
	unitSelect.SetSelected(ps.Str(inter.UsDiv))
	box := container.New(layout.NewHBoxLayout(), timeSelect, unitSelect)
	return box
}
func (ps *PsDesc) setSample(container *fyne.Container) {
	container.Add(layout.NewSpacer())
	ps.samplePanel = ps.newSampleSettings()
	container.Add(ps.samplePanel)
}
