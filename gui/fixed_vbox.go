package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type fixedVBoxLayout struct {
	maxHeight float32
}

func (l *fixedVBoxLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	y := float32(0)
	for _, child := range objects {
		if !child.Visible() {
			continue
		}
		min := child.MinSize()
		child.Resize(fyne.NewSize(size.Width, min.Height))
		child.Move(fyne.NewPos(0, y))
		y += min.Height + theme.Padding()
	}
}

func (l *fixedVBoxLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	width, height := float32(0), float32(0)
	for _, child := range objects {
		min := child.MinSize()
		if min.Width > width {
			width = min.Width
		}
		height += min.Height + theme.Padding()
	}
	if len(objects) > 0 {
		height -= theme.Padding()
	}
	
	return fyne.NewSize(width, height)
}
