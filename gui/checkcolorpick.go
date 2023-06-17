package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type (
	checkColorPick struct {
		widget.Check
	}
)

func newCheckColorPick(id string, changed func(bool)) *checkColorPick {
	ccp := &checkColorPick{}
	ccp.ExtendBaseWidget(ccp)
	ccp.OnChanged = changed
	//	scb.SetResource(res)
	ccp.Text = id
	return ccp
}

func (ccp *checkColorPick) TappedSecondary(event *fyne.PointEvent) {
	var (
		i int
		s string
	)
	fmt.Println("TappedSecondary", s, i)
	// fmt.Println(event.Scrolled, event.Position, event.AbsolutePosition)
}
