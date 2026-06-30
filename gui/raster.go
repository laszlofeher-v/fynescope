package gui

import (
	"image"
	"image/color"
	"image/draw"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

const (
	scrollDelta = 10
	xSnapValue  = 100
)

var (
	chaColor   = color.RGBA{100, 180, 255, 255}
	blue       = color.RGBA{0, 0, 255, 255}
	green      = color.RGBA{0, 255, 0, 255}
	red        = color.RGBA{255, 0, 0, 255}
	yellow     = color.RGBA{255, 255, 0, 255}
	cyan       = color.RGBA{0, 255, 255, 255}
	lightGreen = color.RGBA{64, 200, 64, 255}
	background = color.RGBA{5, 5, 5, 255}
	white      = color.RGBA{255, 255, 255, 255}
	black      = color.RGBA{0, 0, 0, 255}
	gray       = color.RGBA{50, 50, 50, 255}
)

const (
	fontSize            = 20
	defaultLeftMargin   = 30.0
	defaultTopMargin    = 10.0
	defaultRightMargin  = 32.0
	defaultBottomMargin = 50.0
	defaultTimeMargin   = defaultBottomMargin - 8
	numberOfDivs        = 10
	minOffset           = -1e-32
)

type (
	rasterImage draw.RGBA64Image
	// the screen widget contains drawable elements
	rasterPartition struct {
		img         rasterImage
		imgRect     image.Rectangle
		refreshFlag bool
	}
	drawer interface {
		setRect(imgRect image.Rectangle)
		rect() (imgRect image.Rectangle)
		enableRefresh()
		disableRefresh()
		refresh() bool
		draw()
	}
	cursorable interface {
		cursor(x, y float32) (desktop.Cursor, bool)
	}
	mouser interface {
		mouseDown(button desktop.MouseButton, x, y float32)
		mouseUp(button desktop.MouseButton, x, y float32)
		mouseMoved(x, y float32)
	}
	scroller interface {
		scrolled(delta, x, y float32)
	}
	dragger interface {
		dragged(dx, dy, x, y float32)
	}
	keyer interface {
		typedKey(x, y float32, keyName fyne.KeyName)
	}
	screenRaster struct {
		mouseIn        bool
		mouseX, mouseY float32
		widget.BaseWidget
		Window fyne.Window
		raster *canvas.Raster
		min    fyne.Size
		scp    *ScpDesc
		isDft  bool
		isFv   bool
		isFf   bool
	}
)

func (scr *screenRaster) Drawers() []drawer {
	if scr.isDft {
		return scr.scp.dftDrawers
	}
	if scr.isFv {
		return scr.scp.fvDrawers
	}
	if scr.isFf {
		return scr.scp.ffDrawers
	}
	return scr.scp.ftDrawers
}

func (scrPart *rasterPartition) enableRefresh() {
	scrPart.refreshFlag = true
}
func (scrPart *rasterPartition) disableRefresh() {
	scrPart.refreshFlag = false
}
func (scrPart *rasterPartition) refresh() bool {
	return scrPart.refreshFlag
}
func (scrPart *rasterPartition) setRect(imgRect image.Rectangle) {
	scrPart.imgRect = imgRect
}
func (scrPart *rasterPartition) rect() (imgRect image.Rectangle) {
	return scrPart.imgRect
}

func (scp *ScpDesc) newScopeScreen(imgSize image.Point) rasterImage {
	p := image.NewRGBA(image.Rectangle{image.Point{0, 0}, imgSize})
	return p
}

func snapN(v float32, snap float32) float32 {
	v = snap * v
	v = float32(math.Round(float64(v))) / snap
	return v
}

// func (scp *ScpDesc) setXOffset(xOff float32) {
// 	var offset float32
// 	w := float32(scp.ftScopeSignalScreen.Bounds().Dx())
// 	switch {
// 	case xOff < 0:
// 		offset = 0
// 	case xOff > w:
// 		offset = w
// 	default:
// 		offset = xOff
// 	}
// 	scp.Settings.Time.TriggerTimeOffset = float64(offset/w) * scp.maxScreenTime
// 	slog.Debug("setXOffset", "XOffset", scp.Settings.Time.TriggerTimeOffset)
// }

func (raster *screenRaster) Tapped(event *fyne.PointEvent) {
	// fmt.Println("Tapped", event.Position.X, event.Position.Y)
	// raster.triggerX = event.Position.X
	// raster.triggerY = event.Position.Y
	// raster.Refresh()
	// fmt.Println(event.Scrolled, event.Position, event.AbsolutePosition)
}

func (raster *screenRaster) TappedSecondary(event *fyne.PointEvent) {
	// fmt.Println("TappedSecondary", event.Position.X, event.Position.Y)
	// raster.triggerX = event.Position.X
	// raster.triggerY = event.Position.Y
	// raster.Refresh()
	// fmt.Println(event.Scrolled, event.Position, event.AbsolutePosition)
}
func (raster *screenRaster) MouseIn(event *desktop.MouseEvent) {
	raster.Window.Canvas().Focus(raster)
	raster.mouseIn = true
	raster.mouseX = event.Position.X
	raster.mouseY = event.Position.Y
}
func (raster *screenRaster) MouseMoved(event *desktop.MouseEvent) {
	raster.mouseX = event.Position.X
	raster.mouseY = event.Position.Y
	drawers := raster.Drawers()
	for i := range drawers {
		switch m := drawers[i].(type) {
		case mouser:
			m.mouseMoved(event.Position.X, event.Position.Y)
		}
	}
}

func (raster *screenRaster) LeftMouseDown(event *desktop.MouseEvent) {
	raster.mouseX = event.Position.X
	raster.mouseY = event.Position.Y
	x := int(math.Round(float64(event.Position.X)))
	y := int(math.Round(float64(event.Position.Y)))
	p := image.Point{X: x, Y: y}

	if !raster.isFv && !raster.isFf {
		for channelIndex := range raster.scp.channelViewers {
			channelViewer := &raster.scp.channelViewers[channelIndex]
			if raster.isDft {
				channelViewer.dftDisplayOffsetFraction = raster.scp.offsetNToDftY(channelViewer.dftDisplayOffsetInt)
			} else {
				channelViewer.displayOffsetFraction = raster.scp.offsetNToFtY(channelViewer.displayOffsetInt)
			}
		}
		for channelIndex := range raster.scp.channelViewers {
			channel := &raster.scp.Settings.Channels[channelIndex]
			channelViewer := &raster.scp.channelViewers[channelIndex]
			if channel.Enabled {
				if raster.isDft {
					bounds := channelViewer.dftLabel.rect()
					if p.In(bounds) {
						raster.scp.displayMovedDivs = int(channelIndex) + 1
						channelViewer.dftLabel.enableRefresh()
						break
					}
				} else {
					bounds := channelViewer.label.imgRect.Bounds()
					if p.In(bounds) {
						raster.scp.displayMovedDivs = int(channelIndex) + 1
						channelViewer.label.enableRefresh()
						break
					}
				}
			}
		}
	}
	raster.Refresh()
}

func (raster *screenRaster) MouseDown(event *desktop.MouseEvent) {
	raster.mouseX = event.Position.X
	raster.mouseY = event.Position.Y
	canvas.Refresh(raster)
	switch {
	case event.Button == desktop.RightMouseButton:
	case event.Button == desktop.LeftMouseButton:
		raster.LeftMouseDown(event)
	default:
	}
	drawers := raster.Drawers()
	for i := range drawers {
		switch m := drawers[i].(type) {
		case mouser:
			m.mouseDown(event.Button, event.Position.X, event.Position.Y)
		}
	}
}

func (raster *screenRaster) MouseUp(event *desktop.MouseEvent) {
	raster.mouseX = event.Position.X
	raster.mouseY = event.Position.Y
	raster.scp.displayMovedDivs = 0
	drawers := raster.Drawers()
	for i := range drawers {
		switch m := drawers[i].(type) {
		case mouser:
			m.mouseUp(event.Button, event.Position.X, event.Position.Y)
		}
	}
	raster.Refresh()
}

func (raster *screenRaster) MouseOut() {
	raster.mouseIn = false
}

func (raster *screenRaster) DragEnd() {
	raster.scp.displayMovedDivs = 0
	drawers := raster.Drawers()
	for i := range drawers {
		switch m := drawers[i].(type) {
		case mouser:
			m.mouseUp(desktop.RightMouseButton, raster.mouseX, raster.mouseY)
			m.mouseUp(desktop.LeftMouseButton, raster.mouseX, raster.mouseY)
		}
	}
	raster.Refresh()
}
func (raster *screenRaster) TypedKey(k *fyne.KeyEvent) {
	drawers := raster.Drawers()
	for i := range drawers {
		switch m := drawers[i].(type) {
		case keyer:
			m.typedKey(raster.mouseX, raster.mouseY, k.Name)
		}
	}
}
func (raster *screenRaster) TypedRune(r rune) {
}

func (raster *screenRaster) FocusGained() {
}

func (raster *screenRaster) FocusLost() {
}

func (raster *screenRaster) TypedShortcut(r fyne.Shortcut) {
}

func (raster *screenRaster) KeyUp(k *fyne.KeyEvent) {
}
func (raster *screenRaster) KeyDown(k *fyne.KeyEvent) {
}

func (raster *screenRaster) Disable() {
}

func (raster *screenRaster) Enable() {
}

func (raster *screenRaster) Disabled() bool {
	return false
}

func (raster *screenRaster) AcceptsTab() bool {
	return false
}

func (raster *screenRaster) Cursor() desktop.Cursor {
	drawers := raster.Drawers()
	for i := range drawers {
		switch m := drawers[i].(type) {
		case cursorable:
			cp, ok := m.cursor(raster.mouseX, raster.mouseY)
			if ok {
				return cp
			}
		}
	}
	return desktop.DefaultCursor
}

var _ fyne.Disableable = (*screenRaster)(nil)
var _ fyne.Draggable = (*screenRaster)(nil)
var _ fyne.Focusable = (*screenRaster)(nil)
var _ fyne.Tappable = (*screenRaster)(nil)
var _ fyne.Widget = (*screenRaster)(nil)
var _ desktop.Mouseable = (*screenRaster)(nil)
var _ desktop.Keyable = (*screenRaster)(nil)
var _ fyne.Tabbable = (*screenRaster)(nil)

func (raster *screenRaster) Scrolled(event *fyne.ScrollEvent) {
	raster.mouseX = event.Position.X
	raster.mouseY = event.Position.Y
	x := event.Position.X
	y := event.Position.Y
	dy := float32(scrollDelta)

	// Adjust for scale to keep movement proportional to pixels
	if raster.Size().Width > 0 {
		scaleX := float32(raster.scp.dftScopeFullScreen.Bounds().Dx()) / raster.Size().Width
		if !math.IsNaN(float64(scaleX)) && !math.IsInf(float64(scaleX), 0) {
			dy *= scaleX
		}
	}

	if event.Scrolled.DY < 0 { // DY is 25 since fyne.io/fyne/v2 v2.4.2
		dy = -dy
	}
	drawers := raster.Drawers()
	for i := range drawers {
		switch m := drawers[i].(type) {
		case scroller:
			m.scrolled(dy, x, y)
		}
	}
}

func (raster *screenRaster) Dragged(event *fyne.DragEvent) {
	raster.mouseX = event.Position.X
	raster.mouseY = event.Position.Y
	x := event.Position.X
	y := event.Position.Y
	dx := event.Dragged.DX
	dy := event.Dragged.DY

	// Adjust for scale to keep movement proportional to pixels
	if raster.Size().Width > 0 {
		var scaleX float32
		if raster.isDft {
			scaleX = float32(raster.scp.dftScopeFullScreen.Bounds().Dx()) / raster.Size().Width
		} else if raster.isFv {
			scaleX = float32(raster.scp.fvScopeFullScreen.Bounds().Dx()) / raster.Size().Width
		} else if raster.isFf {
			scaleX = float32(raster.scp.ffScopeFullScreen.Bounds().Dx()) / raster.Size().Width
		} else {
			scaleX = float32(raster.scp.ftScopeFullScreen.Bounds().Dx()) / raster.Size().Width
		}
		if !math.IsNaN(float64(scaleX)) && !math.IsInf(float64(scaleX), 0) {
			dx *= scaleX
			dy *= scaleX
		}
	}

	drawers := raster.Drawers()
	for i := range drawers {
		switch m := drawers[i].(type) {
		case dragger:
			m.dragged(dx, dy, x, y)
		}
	}
}

func (scp *ScpDesc) newScreenRaster(generate func(w, h int) image.Image, window fyne.Window, isDft, isFv, isFf bool) (scr *screenRaster) {
	scr = &screenRaster{raster: canvas.NewRaster(generate), Window: window, scp: scp, isDft: isDft, isFv: isFv, isFf: isFf}
	scr.ExtendBaseWidget(scr)
	return scr
}

type (
	screenRasterRenderer struct {
		bg *canvas.Raster
	}
)

func (r *screenRaster) CreateRenderer() fyne.WidgetRenderer {
	r.min = fyne.Size{Width: 400, Height: 200}
	return &screenRasterRenderer{bg: canvas.NewRaster(r.raster.Generator)}
}

func (r *screenRaster) MinSize() fyne.Size {
	return r.min
}

func (r *screenRasterRenderer) Refresh() {
	fyne.Do(func() {
		canvas.Refresh(r.bg)
	})
}

func (r *screenRasterRenderer) Destroy() {
}

func (r *screenRasterRenderer) Layout(size fyne.Size) {
	r.bg.Resize(size)
}

func (r *screenRasterRenderer) MinSize() fyne.Size {
	return r.MinSize()
}

func (r *screenRasterRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.bg}
}
