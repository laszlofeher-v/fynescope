package gui

import (
	"fyne.io/fyne/v2"
)

type tightHBoxLayout struct {
	gap float32
}

func (l *tightHBoxLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	x := float32(0)
	for _, child := range objects {
		if !child.Visible() {
			continue
		}
		min := child.MinSize()
		child.Resize(min)
		child.Move(fyne.NewPos(x, 0))
		x += min.Width + l.gap
	}
}

func (l *tightHBoxLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	width, height := float32(0), float32(0)
	for _, child := range objects {
		if !child.Visible() {
			continue
		}
		min := child.MinSize()
		width += min.Width + l.gap
		if min.Height > height {
			height = min.Height
		}
	}
	if len(objects) > 0 {
		width -= l.gap
	}
	return fyne.NewSize(width, height)
}
