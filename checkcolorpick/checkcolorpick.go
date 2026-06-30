package checkcolorpick

import (
	"image/color"

	"fyne.io/fyne/v2/driver/desktop"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/dialog"

	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// TODO TAB focus, space check
type (
	CheckColorPick struct {
		widget.BaseWidget
		raster  *canvas.Raster
		min     fyne.Size
		changed func(v bool, col color.Color)
		col     color.Color
		Val     bool
		window  fyne.Window
		focused bool
		//TODO show fucus
	}
)

func NewCheckColorPick(window fyne.Window, changed func(v bool, col color.Color), col color.Color, minSize fyne.Size) (ccp *CheckColorPick) {
	generator := func(x, y, w, h int) color.Color {
		r, g, b, a := ccp.col.RGBA()
		if !(x > 2 && x < w-3 &&
			y > 2 && y < h-3) {
			a &= 0x0
			r &= 0x0
			g &= 0x0
			b &= 0x0
		} else {
			if ccp.Val {
				if !(x > 4 && x < w-5 &&
					y > 4 && y < h-5) {
					a &= 0xa0
					r &= 0xa0
					g &= 0xa0
					b &= 0xa0
				}
			} else {
				if x > 4 && x < w-5 &&
					y > 4 && y < h-5 {
					return color.Gray16{0xff}
				} else {
					a &= 0xf0
					r &= 0xf0
					g &= 0xf0
					b &= 0xf0
				}
			}
		}
		return &color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a)}
	}
	raster := canvas.NewRasterWithPixels(generator)
	ccp = &CheckColorPick{raster: raster, changed: changed, col: col, window: window, min: minSize}
	ccp.ExtendBaseWidget(ccp)
	return ccp
}

func (ccp *CheckColorPick) Hide() {
}

func (ccp *CheckColorPick) SetColor(col color.Color) {
	ccp.col = col
	ccp.changed(ccp.Val, ccp.col)
	canvas.Refresh(ccp)
}
func (ccp *CheckColorPick) Set() {
	ccp.Val = true
	ccp.changed(ccp.Val, ccp.col)
	canvas.Refresh(ccp)
}
func (ccp *CheckColorPick) Tapped(event *fyne.PointEvent) {
	ccp.Val = !ccp.Val
	ccp.changed(ccp.Val, ccp.col)
	canvas.Refresh(ccp)
}

func (ccp *CheckColorPick) TappedSecondary(event *fyne.PointEvent) {
	callback := func(c color.Color) {
		switch v := c.(type) {
		case *color.NRGBA:
			ccp.col = *v
		case color.NRGBA:
			ccp.col = v
		}
		ccp.changed(ccp.Val, ccp.col)
		canvas.Refresh(ccp)
	}
	cp := dialog.NewColorPicker("Colors", "Select", callback, ccp.window)
	cp.Advanced = true
	cp.SetColor(ccp.col)
	cp.Show()
}

func (ccp *CheckColorPick) MinSize() fyne.Size {
	return ccp.min
}

func (ccp *CheckColorPick) FocusGained() {
	ccp.focused = true
	ccp.Refresh()
}
func (ccp *CheckColorPick) FocusLost() {
	ccp.focused = false
	ccp.Refresh()
}
func (ccp *CheckColorPick) TypedKey(k *fyne.KeyEvent) {
	ccp.Val = !ccp.Val
	ccp.changed(ccp.Val, ccp.col)
	canvas.Refresh(ccp)
}
func (ccp *CheckColorPick) TypedRune(r rune) {

}

var _ fyne.Focusable = (*CheckColorPick)(nil)

type (
	checkColorPickRenderer struct {
		col            color.Color
		ccp            *CheckColorPick
		bg             *canvas.Raster
		focusIndicator *canvas.Circle
	}
)

func (ccp *CheckColorPick) CreateRenderer() fyne.WidgetRenderer {
	focusIndicator := canvas.NewCircle(theme.BackgroundColor())
	r := &checkColorPickRenderer{}
	r.bg = canvas.NewRaster(ccp.raster.Generator)
	r.focusIndicator = focusIndicator
	r.ccp = ccp
	r.Refresh()
	return r
}
func (ccp *checkColorPickRenderer) MinSize() fyne.Size {
	return ccp.bg.MinSize()
}

func (ccp *checkColorPickRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{ccp.bg, ccp.focusIndicator}
}

func (ccpr *checkColorPickRenderer) Refresh() {
	if ccpr.ccp.focused {
		ccpr.focusIndicator.FillColor = theme.FocusColor()
	} else {
		ccpr.focusIndicator.FillColor = color.Transparent
	}
	ccpr.focusIndicator.Refresh()
	canvas.Refresh(ccpr.ccp)
}
func (ccpr *checkColorPickRenderer) Destroy() {
}
func (ccpr *checkColorPickRenderer) Layout(size fyne.Size) {
	if size.Width < size.Height {
		size.Height = size.Width
	} else {
		size.Width = size.Height
	}
	ccpr.focusIndicator.Resize(size)
	ccpr.bg.Resize(size)
}
func (ccp *CheckColorPick) MouseIn(e *desktop.MouseEvent) {
	ccp.window.Canvas().Focus(ccp)
	ccp.Refresh()
}
func (ccp *CheckColorPick) MouseDown(e *desktop.MouseEvent) {
}
func (ccp *CheckColorPick) MouseUp(e *desktop.MouseEvent) {
}
func (ccp *CheckColorPick) MouseMoved(e *desktop.MouseEvent) {
}
func (ccp *CheckColorPick) MouseOut() {
	ccp.window.Canvas().Unfocus()
	ccp.Refresh()
}
func (ccp *CheckColorPick) Disable() {
}
func (ccp *CheckColorPick) Enable() {
}
func (ccp *CheckColorPick) Disabled() bool {
	return false
}

var _ fyne.Disableable = (*CheckColorPick)(nil)

// var _ fyne.Draggable = (*checkColorPick)(nil)
var _ fyne.Focusable = (*CheckColorPick)(nil)
var _ fyne.Tappable = (*CheckColorPick)(nil)
var _ fyne.Widget = (*CheckColorPick)(nil)
var _ desktop.Mouseable = (*CheckColorPick)(nil)

// var _ desktop.Keyable = (*checkColorPick)(nil)
