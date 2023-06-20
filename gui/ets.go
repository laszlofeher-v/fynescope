package gui

import (
	"fmt"
	"psgo/gui/inter"

	"psgo/ps2000a"

	// "psgo/ps2000a"
	// "sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	//	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

var ()

func (ps *PsDesc) newEtsSettings() *fyne.Container {
	// changed := func(option string) {
	// 	fmt.Println("Changed Gen", option)
	// }
	checked := func(c bool) {
		fmt.Println("Checked ETS", c)
		ps.EtsMode = c
		if !c {
			_, err := ps.SetEts(ps2000a.EtsOff, 20, 4)
			if err != nil {
				fmt.Println("checked SetEts:", err)
				return
			}
			err = ps.Stop()
			if err != nil {
				fmt.Println("checked Stop:", err)
				return
			}
		}
	}
	check := widget.NewCheck(ps.Str(inter.Ets), checked)
	ps.etsCycles = widget.NewEntry()
	ps.etsInterleave = widget.NewEntry()
	ps.etsSampleCountEntry = widget.NewEntry()
	box := container.New(layout.NewVBoxLayout(), check, ps.etsCycles, ps.etsInterleave, ps.etsSampleCountEntry)
	return box
}

func (ps *PsDesc) setEtsPanel(container *fyne.Container) {
	container.Add(layout.NewSpacer())
	ps.etsPanel = ps.newEtsSettings()
	container.Add(ps.etsPanel)
}
