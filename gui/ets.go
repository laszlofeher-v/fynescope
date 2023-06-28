package gui

import (
	"fmt"
	"psgo/gui/inter"
	"psgo/psi"

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
		mode := ps2000a.EtsOff
		if c {
			ps.halt()
			mode = ps2000a.EtsSlow
		}
		//		_, err := ps.SetEts(ps2000a.EtsOff, 20, 4)
		err := ps.Con.Send(&psi.SetEtsMsg{Mode: mode, EtsCycles: 40, EtsInterleave: 4})
		if err != nil {
			fmt.Println("checked SetEts:", err)
			return
		}
		resp, err := ps.Con.Receive()
		sampleTimePs := resp.(*psi.SetEtsRsp).SampleTimePicoseconds
		if err != nil || resp.Status() != nil {
			fmt.Println("checked SetEts:", err, resp.Status())
			return
		}
		fmt.Println("sampleTimePs:", sampleTimePs)
		if c {

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
