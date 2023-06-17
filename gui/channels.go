package gui

import (
	//	"fmt"
	"psgo/gui/inter"
	"psgo/ps2000a"
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

var (
	//TODO language change
	channelNames = []string{"CH A", "CH B", "CH C", "CH D"}
	vRanges      = map[string]ps2000a.RangeEnum{
		"±10mV":  (ps2000a.Range_10mv),
		"±20mV":  ps2000a.Range_20mv,
		"±50mV":  ps2000a.Range_50mv,
		"±100mV": ps2000a.Range_100mv,
		"±200mV": ps2000a.Range_200mv,
		"±500mV": ps2000a.Range_500mv,
		"±1V":    ps2000a.Range_1v,
		"±2V":    ps2000a.Range_2v,
		"±5V":    ps2000a.Range_5v,
		"±10V":   ps2000a.Range_10v,
		"±20V":   ps2000a.Range_20v,
		"±50V":   (ps2000a.Range_50v),
	}
)

func (ps *PsDesc) numberOfEnabledChannels() int {
	n := 0
	for i := range ps.ChannelSettings {
		if ps.ChannelSettings[i].Enabled {
			n++
		}
	}
	return n
}

func sortMapString(strInt map[string]ps2000a.RangeEnum) []string {
	type keyValDesc struct {
		key string
		val ps2000a.RangeEnum
	}
	var keyVal []keyValDesc
	for key, val := range strInt {
		keyVal = append(keyVal, keyValDesc{key, val})
	}
	sort.Slice(keyVal, func(i, j int) bool {
		return keyVal[i].val > keyVal[j].val
	})
	sorted := make([]string, len(strInt))
	for i, kv := range keyVal {
		sorted[i] = kv.key
	}
	return sorted
}

func (ps *PsDesc) newChannelSettings(chIndex int) *fyne.Container {
	chName := channelNames[chIndex]
	cChanged := func(option string) {
		switch option {
		case "AC":
			ps.ChannelSettings[chIndex].CoupleType = ps2000a.Ac
			//			fmt.Println("AC Changed ", chIndex, chName, option)
		case "DC":
			ps.ChannelSettings[chIndex].CoupleType = ps2000a.Dc
			//			fmt.Println("DC Changed ", chIndex, chName, option)
		default:
		}
		ps.raster.Refresh()
	}
	vChanged := func(option string) {
		v := vRanges[option]
		//		fmt.Println("Changed ", chIndex, chName, option, v)
		ps.ChannelSettings[chIndex].Range = v
		ps.raster.Refresh()
	}
	checked := func(c bool) {
		ps.ChannelSettings[chIndex].Enabled = c
		ps.raster.Refresh()
		//		fmt.Println("Checked ", chIndex, chName, c)
	}
	inverted := func(c bool) {
		ps.ChannelSettings[chIndex].Inverted = c
		//		fmt.Println("inverted ", chIndex, chName, c)
		ps.raster.Refresh()
	}
	//TODO new widget type: left click check, right click color picker
	// check := widget.NewCheck(chName, checked)
	check := newCheckColorPick(chName, checked)
	invert := widget.NewCheck("Invert", inverted)
	inputRanges := sortMapString(vRanges)
	vRange := newSelectScroll(inputRanges, vChanged, "+500m")
	vRange.SetSelected("±1V")
	acdc := newSelectScroll([]string{ps.Str(inter.Ac), ps.Str(inter.Dc)}, cChanged,
		ps.Str(inter.Ac))
	acdc.SetSelected(ps.Str(inter.Dc))
	nameInvert := container.New(layout.NewHBoxLayout(), check, invert)
	couplingRange := container.New(layout.NewHBoxLayout(), acdc, vRange)
	box := container.New(layout.NewVBoxLayout(), nameInvert, couplingRange)
	return box
}

func (ps *PsDesc) setChannels(container *fyne.Container) {
	//	ps.ChannelSettings = make([]*fyne.Container, ps.ChannelCount)
	ps.channelPanels = make([]*fyne.Container, ps.ChannelCount)
	for i := range ps.ChannelSettings {
		if i > 0 {
			container.Add(layout.NewSpacer())
		}
		ps.channelPanels[i] = ps.newChannelSettings(i)
		container.Add(ps.channelPanels[i])
	}
	//	fmt.Println(ps.channelPanels)
	// var cha *fyne.Container
	// cha = ps.newChannelSettings("CH A", ps.chAChanged)
	// chb := ps.newChannelSettings("CH B", ps.chBChanged)
	// chc := ps.newChannelSettings("CH C", ps.chCChanged)
	// chd := ps.newChannelSettings("CH D", ps.chDChanged)

}
