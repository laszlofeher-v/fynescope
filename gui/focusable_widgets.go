package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// FocusableLabel is a Label that is Focusable, Mousable, and Hoverable.
type FocusableLabel struct {
	widget.Label
}

var _ fyne.Focusable = (*FocusableLabel)(nil)
var _ desktop.Mouseable = (*FocusableLabel)(nil)
var _ desktop.Hoverable = (*FocusableLabel)(nil)

func NewFocusableLabel(text string) *FocusableLabel {
	l := &FocusableLabel{}
	l.Text = text
	l.ExtendBaseWidget(l)
	return l
}

func (l *FocusableLabel) requestFocus() {
	if app := fyne.CurrentApp(); app != nil {
		for _, w := range app.Driver().AllWindows() {
			if c := w.Canvas(); c != nil {
				c.Focus(l)
				return
			}
		}
	}
}

func (l *FocusableLabel) FocusGained()            { l.Refresh() }
func (l *FocusableLabel) FocusLost()              { l.Refresh() }
func (l *FocusableLabel) TypedRune(rune)          {}
func (l *FocusableLabel) TypedKey(*fyne.KeyEvent) {}

func (l *FocusableLabel) MouseDown(*desktop.MouseEvent)  { l.requestFocus() }
func (l *FocusableLabel) MouseUp(*desktop.MouseEvent)    {}
func (l *FocusableLabel) MouseIn(*desktop.MouseEvent)    {}
func (l *FocusableLabel) MouseOut()                      {}
func (l *FocusableLabel) MouseMoved(*desktop.MouseEvent) {}

// FocusableIcon is an Icon that is Focusable, Mousable, and Hoverable.
type FocusableIcon struct {
	widget.Icon
}

var _ fyne.Focusable = (*FocusableIcon)(nil)
var _ desktop.Mouseable = (*FocusableIcon)(nil)
var _ desktop.Hoverable = (*FocusableIcon)(nil)

func NewFocusableIcon(res fyne.Resource) *FocusableIcon {
	i := &FocusableIcon{}
	i.SetResource(res)
	i.ExtendBaseWidget(i)
	return i
}

func (i *FocusableIcon) requestFocus() {
	if app := fyne.CurrentApp(); app != nil {
		for _, w := range app.Driver().AllWindows() {
			if c := w.Canvas(); c != nil {
				c.Focus(i)
				return
			}
		}
	}
}

func (i *FocusableIcon) FocusGained()            { i.Refresh() }
func (i *FocusableIcon) FocusLost()              { i.Refresh() }
func (i *FocusableIcon) TypedRune(rune)          {}
func (i *FocusableIcon) TypedKey(*fyne.KeyEvent) {}

func (i *FocusableIcon) MouseDown(*desktop.MouseEvent)  { i.requestFocus() }
func (i *FocusableIcon) MouseUp(*desktop.MouseEvent)    {}
func (i *FocusableIcon) MouseIn(*desktop.MouseEvent)    {}
func (i *FocusableIcon) MouseOut()                      {}
func (i *FocusableIcon) MouseMoved(*desktop.MouseEvent) {}

// FocusableProgressBar is a ProgressBar that is Focusable, Mousable, and Hoverable.
type FocusableProgressBar struct {
	widget.ProgressBar
}

var _ fyne.Focusable = (*FocusableProgressBar)(nil)
var _ desktop.Mouseable = (*FocusableProgressBar)(nil)
var _ desktop.Hoverable = (*FocusableProgressBar)(nil)

func NewFocusableProgressBar() *FocusableProgressBar {
	p := &FocusableProgressBar{}
	p.ExtendBaseWidget(p)
	return p
}

func (p *FocusableProgressBar) requestFocus() {
	if app := fyne.CurrentApp(); app != nil {
		for _, w := range app.Driver().AllWindows() {
			if c := w.Canvas(); c != nil {
				c.Focus(p)
				return
			}
		}
	}
}

func (p *FocusableProgressBar) FocusGained()            { p.Refresh() }
func (p *FocusableProgressBar) FocusLost()              { p.Refresh() }
func (p *FocusableProgressBar) TypedRune(rune)          {}
func (p *FocusableProgressBar) TypedKey(*fyne.KeyEvent) {}

func (p *FocusableProgressBar) MouseDown(*desktop.MouseEvent)  { p.requestFocus() }
func (p *FocusableProgressBar) MouseUp(*desktop.MouseEvent)    {}
func (p *FocusableProgressBar) MouseIn(*desktop.MouseEvent)    {}
func (p *FocusableProgressBar) MouseOut()                      {}
func (p *FocusableProgressBar) MouseMoved(*desktop.MouseEvent) {}

func NewFocusableLabelWithStyle(text string, alignment fyne.TextAlign, style fyne.TextStyle) *FocusableLabel {
	l := &FocusableLabel{}
	l.Text = text
	l.Alignment = alignment
	l.TextStyle = style
	l.ExtendBaseWidget(l)
	return l
}
