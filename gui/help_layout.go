package gui

import (
	"fyne.io/fyne/v2"
)

type helpOverlayLayout struct {
	scp *ScpDesc
}

func (l *helpOverlayLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(0, 0)
}

func (l *helpOverlayLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if app := fyne.CurrentApp(); app != nil && app.Driver() != nil {
		for _, obj := range objects {
			if proxy, ok := obj.(*TabFocusProxy); ok {
				if btn := getTabButton(proxy.item); btn != nil {
					pos := app.Driver().AbsolutePositionForObject(btn)
					proxy.Move(pos)
					proxy.Resize(btn.Size())
				}
			}
		}
	}
}
