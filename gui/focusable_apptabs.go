package gui

import (
	"log/slog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
)

// FocusableAppTabs is an extension of container.AppTabs that can receive focus.
type FocusableAppTabs struct {
	container.AppTabs
}

var _ desktop.Mouseable = (*FocusableAppTabs)(nil)
var _ desktop.Hoverable = (*FocusableAppTabs)(nil)

// NewFocusableAppTabs creates a new focusable app tabs container.
func NewFocusableAppTabs(items ...*container.TabItem) *FocusableAppTabs {
	tabs := &FocusableAppTabs{}
	tabs.Items = items
	tabs.ExtendBaseWidget(tabs)
	// Hook into OnSelected: whenever a tab is tapped/selected we claim canvas
	// focus. This is the only reliable entry point because the internal tab
	// buttons call AppTabs.Select() directly rather than routing through any
	// Tapped method on the outer widget.
	userOnSelected := tabs.OnSelected
	tabs.OnSelected = func(item *container.TabItem) {
		if userOnSelected != nil {
			userOnSelected(item)
		}
		tabs.requestFocus()
	}
	return tabs
}

// requestFocus asks the canvas of the active window to focus this widget.
func (f *FocusableAppTabs) requestFocus() {
	app := fyne.CurrentApp()
	if app == nil {
		return
	}
	for _, w := range app.Driver().AllWindows() {
		if c := w.Canvas(); c != nil {
			c.Focus(f)
			return
		}
	}
}

// FocusGained is called when the tabs gain focus.
func (f *FocusableAppTabs) FocusGained() {
	slog.Warn("FocusGained")
	f.Refresh()
}

// FocusLost is called when the tabs lose focus.
func (f *FocusableAppTabs) FocusLost() {
	slog.Warn("FocusLost")
	f.Refresh()
}

// TypedRune is called when the user types a rune.
func (f *FocusableAppTabs) TypedRune(rune) {
}

// TypedKey is called when the user types a key.
func (f *FocusableAppTabs) TypedKey(key *fyne.KeyEvent) {
}

// Tapped is called when the user taps the widget.
// Implementing Tappable allows Fyne to automatically grant focus when tapped.
func (f *FocusableAppTabs) Tapped(ev *fyne.PointEvent) {
	slog.Warn("Tapped")
	f.requestFocus()
}

// MouseDown is called when a desktop mouse button is pressed.
func (f *FocusableAppTabs) MouseDown(event *desktop.MouseEvent) {
	slog.Warn("MouseDown")
	f.requestFocus()
}

// MouseUp is called when a desktop mouse button is released.
func (f *FocusableAppTabs) MouseUp(event *desktop.MouseEvent) {
}
func (f *FocusableAppTabs) MouseIn(e *desktop.MouseEvent) {
	slog.Warn("MouseIn")
}
func (f *FocusableAppTabs) MouseOut() {
	slog.Warn("MouseOut")
}
func (f *FocusableAppTabs) MouseMoved(e *desktop.MouseEvent) {
	slog.Warn("MouseMoved")
}
