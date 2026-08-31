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
		if !child.Visible() {
			continue
		}
		min := child.MinSize()
		if min.Width > width {
			width = min.Width
		}
		height += min.Height + theme.Padding()
	}
	if height > 0 {
		height -= theme.Padding()
	}

	return fyne.NewSize(width, height)
}

// stableVBoxLayout is like fixedVBoxLayout but reserves height for every
// child slot regardless of visibility. This prevents sibling widgets from
// jumping when a row is shown or hidden, because the container's MinSize
// stays constant.
//
// It uses a ratchet: each slot's height only ever increases. Once a child
// has been measured at its natural size the value is cached, so even if Fyne
// returns 0 for a hidden container the reserved space is kept.
type stableVBoxLayout struct {
	reserved []float32 // per-slot max height observed so far
}

func (l *stableVBoxLayout) ensure(n int) {
	for len(l.reserved) < n {
		l.reserved = append(l.reserved, 0)
	}
}

// slotHeight returns the reserved height for index i, updating the ratchet if
// the child reports a larger value now.
func (l *stableVBoxLayout) slotHeight(i int, child fyne.CanvasObject) float32 {
	h := child.MinSize().Height
	if h > l.reserved[i] {
		l.reserved[i] = h
	}
	return l.reserved[i]
}

func (l *stableVBoxLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	l.ensure(len(objects))
	y := float32(0)
	for i, child := range objects {
		h := l.slotHeight(i, child)
		if h == 0 {
			continue // never measured yet — skip the slot entirely
		}
		if child.Visible() {
			child.Resize(fyne.NewSize(size.Width, child.MinSize().Height))
			child.Move(fyne.NewPos(0, y))
		}
		y += h + theme.Padding()
	}
}

func (l *stableVBoxLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	l.ensure(len(objects))
	width, height := float32(0), float32(0)
	slots := 0
	for i, child := range objects {
		h := l.slotHeight(i, child)
		w := child.MinSize().Width
		if w > width {
			width = w
		}
		if h > 0 {
			height += h + theme.Padding()
			slots++
		}
	}
	if slots > 0 {
		height -= theme.Padding()
	}
	return fyne.NewSize(width, height)
}

// fixedMaxLayout is like container.NewMax but always reserves the maximum
// size needed by ANY child (visible or hidden). This prevents the container
// from shrinking when some children are hidden, which would otherwise cause
// sibling widgets in an outer layout to jump.
type fixedMaxLayout struct{}

func (l *fixedMaxLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, child := range objects {
		if !child.Visible() {
			continue
		}
		child.Resize(size)
		child.Move(fyne.NewPos(0, 0))
	}
}

func (l *fixedMaxLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	var width, height float32
	for _, child := range objects {
		// Include ALL children regardless of visibility so that the
		// container always reserves the maximum possible space.
		min := child.MinSize()
		if min.Width > width {
			width = min.Width
		}
		if min.Height > height {
			height = min.Height
		}
	}
	return fyne.NewSize(width, height)
}
