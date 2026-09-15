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
	app := fyne.CurrentApp()
	if app == nil || app.Driver() == nil {
		return
	}
	for _, obj := range objects {
		proxy, ok := obj.(*TabFocusProxy)
		if !ok {
			continue
		}
		btn := getTabButton(proxy.item)
		if btn == nil {
			// No button yet — keep proxy visible but with zero size so it
			// stays in the canvas tree (focusable) but doesn't block clicks.
			proxy.Move(fyne.NewPos(-1, -1))
			proxy.Resize(fyne.NewSize(0, 0))
			continue
		}
		btnSize := btn.Size()
		absPos := app.Driver().AbsolutePositionForObject(btn)
		btnRelPos := btn.Position()

		// If AbsolutePositionForObject returns (0,0) but the button's own
		// relative position is non-zero, the widget hasn't been laid out in
		// the canvas tree yet.  Park the proxy off-screen at (-1,-1) with
		// zero size so it stays focusable but can't intercept mouse events
		// at the top-left corner (where Run/Zoom toolbar buttons live).
		if btnSize.Width <= 0 || btnSize.Height <= 0 ||
			(absPos == (fyne.Position{}) && btnRelPos != (fyne.Position{})) {
			proxy.Move(fyne.NewPos(-1, -1))
			proxy.Resize(fyne.NewSize(0, 0))
			continue
		}
		proxy.Move(absPos)
		proxy.Resize(btnSize)
	}
}
