package gui

import (
	//	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type (
	selectScroll struct {
		widget.Select
		OverflowIndex int
	}
)

func newSelectScroll(options []string, changed func(string), placeHolder string) *selectScroll {
	selScr := &selectScroll{}
	selScr.Options = options
	selScr.OnChanged = changed
	selScr.ExtendBaseWidget(selScr)
	selScr.PlaceHolder = placeHolder
	//	scb.SetResource(res)
	return selScr
}

func (selScr *selectScroll) Scrolled(event *fyne.ScrollEvent) {
	var (
		i int
		s string
	)
	for i, s = range selScr.Options {
		if s == selScr.Selected {
			break
		}
	}
	//	fmt.Println(s, i)
	if event.Scrolled.DY > 0 {
		selScr.OverflowIndex--
		selScr.SetSelectedIndex(i - 1)
		if selScr.OverflowIndex != selScr.SelectedIndex() {
			selScr.OnChanged(selScr.Selected)
		}
	} else {
		selScr.OverflowIndex++
		selScr.SetSelectedIndex(i + 1)
		if selScr.OverflowIndex != selScr.SelectedIndex() {
			selScr.OnChanged(selScr.Selected)
		}
	}
	// fmt.Println(event.Scrolled, event.Position, event.AbsolutePosition)
}
