// Package checkcolorpick provides a custom Fyne widget that combines a checkbox
// toggle with a color picker. Left-clicking toggles the boolean state, while
// right-clicking opens a color picker dialog to change the associated color.
// The widget supports keyboard focus, mouse hover focus, and implements the
// Fyne Focusable, Tappable, Widget, and desktop.Mouseable interfaces.
package checkcolorpick

import (
	"image/color"

	"fyne.io/fyne/v2/driver/desktop"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"

	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type (
	// CheckColorPick is a custom widget that displays a colored square with a
	// checkbox-like toggle behaviour. The color fills the interior of the square,
	// and the boolean state (Val) is reflected by visual border/fill changes.
	CheckColorPick struct {
		widget.BaseWidget
		raster  *canvas.Raster    // pixel-level renderer that draws the colored square
		min     fyne.Size         // minimum size of the widget
		changed func(v bool, col color.Color) // callback invoked whenever Val or col changes
		col     color.Color       // current selected color
		Val     bool              // current toggle state (true = checked / enabled)
		window  fyne.Window       // parent window, used for the color picker dialog and focus management
		focused bool              // whether the widget currently holds keyboard focus
	}
)

// NewCheckColorPick creates and returns a new CheckColorPick widget.
//
//   - window:  the parent window, required for opening the color-picker dialog
//     and for managing canvas focus on mouse hover.
//   - changed: callback called with the current (Val, col) whenever either changes.
//   - col:     the initial color to display.
//   - minSize: the minimum size the widget should occupy.
//
// The pixel generator closure draws the widget as a colored square:
//   - A 3-pixel transparent border creates the outer edge.
//   - When Val is true (checked):  a 2-pixel semi-opaque ring (0xa0 mask) is
//     drawn between the outer border and the inner fill, indicating the enabled state.
//   - When Val is false (unchecked): the inner area (inside a 5-pixel border) is
//     rendered white, and the ring is drawn with a lighter mask (0xf0), indicating
//     the disabled / unchecked state.
func NewCheckColorPick(window fyne.Window, changed func(v bool, col color.Color), col color.Color, minSize fyne.Size) (ccp *CheckColorPick) {
	// generator is the per-pixel color function passed to canvas.NewRasterWithPixels.
	// (x, y) is the pixel coordinate; (w, h) is the total raster size in pixels.
	generator := func(x, y, w, h int) color.Color {
		r, g, b, a := ccp.col.RGBA()

		// Outer 3-pixel border: make these pixels fully transparent so the
		// widget appears to have a rounded/floating look against the background.
		if !(x > 2 && x < w-3 &&
			y > 2 && y < h-3) {
			a &= 0x0
			r &= 0x0
			g &= 0x0
			b &= 0x0
		} else {
			if ccp.Val {
				// Checked state: draw a semi-opaque ring (pixels between the
				// 3-pixel and 5-pixel borders) using a 0xa0 alpha mask.
				// Pixels inside the 5-pixel border retain the full widget color.
				if !(x > 4 && x < w-5 &&
					y > 4 && y < h-5) {
					a &= 0xa0
					r &= 0xa0
					g &= 0xa0
					b &= 0xa0
				}
			} else {
				// Unchecked state: the inner region (inside the 5-pixel border)
				// is rendered solid white, making it visually "empty / off".
				// The ring region uses a 0xf0 alpha mask for a lighter appearance.
				if x > 4 && x < w-5 &&
					y > 4 && y < h-5 {
					return color.Gray16{0xff} // solid white inner fill
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

// SetColor updates the widget's color to col, triggers the changed callback,
// and refreshes the canvas to redraw the widget.
func (ccp *CheckColorPick) SetColor(col color.Color) {
	ccp.col = col
	ccp.changed(ccp.Val, ccp.col)
	canvas.Refresh(ccp)
}

// Set sets the toggle state to true (checked), triggers the changed callback,
// and refreshes the canvas.
func (ccp *CheckColorPick) Set() {
	ccp.Val = true
	ccp.changed(ccp.Val, ccp.col)
	canvas.Refresh(ccp)
}

// SetVal sets the toggle state to val, triggers the changed callback,
// and refreshes the canvas.
func (ccp *CheckColorPick) SetVal(val bool) {
	ccp.Val = val
	ccp.changed(ccp.Val, ccp.col)
	canvas.Refresh(ccp)
}

// Tapped handles a primary (left) tap/click: it toggles Val, triggers the
// changed callback, and refreshes the canvas.
func (ccp *CheckColorPick) Tapped(event *fyne.PointEvent) {
	ccp.Val = !ccp.Val
	ccp.changed(ccp.Val, ccp.col)
	canvas.Refresh(ccp)
}

// TappedSecondary handles a secondary (right) tap/click: it opens the Fyne
// color-picker dialog in advanced mode pre-filled with the current color.
// When the user confirms a new color, the callback updates ccp.col and
// triggers the changed callback.
func (ccp *CheckColorPick) TappedSecondary(event *fyne.PointEvent) {
	callback := func(c color.Color) {
		// Normalise the returned color to a concrete color.NRGBA value so that
		// subsequent RGBA() calls on ccp.col behave consistently.
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

// MinSize returns the minimum size of the widget as configured at construction time.
func (ccp *CheckColorPick) MinSize() fyne.Size {
	return ccp.min
}

// FocusGained is called by Fyne when the widget receives keyboard focus.
// It sets the focused flag and refreshes so the focus indicator becomes visible.
func (ccp *CheckColorPick) FocusGained() {
	ccp.focused = true
	ccp.Refresh()
}

// FocusLost is called by Fyne when the widget loses keyboard focus.
// It clears the focused flag and refreshes so the focus indicator is hidden.
func (ccp *CheckColorPick) FocusLost() {
	ccp.focused = false
	ccp.Refresh()
}

// TypedKey is called when a key is pressed while the widget has focus.
// Any key press toggles the Val state, triggering the changed callback.
func (ccp *CheckColorPick) TypedKey(k *fyne.KeyEvent) {
	ccp.Val = !ccp.Val
	ccp.changed(ccp.Val, ccp.col)
	canvas.Refresh(ccp)
}

// TypedRune is required by the fyne.Focusable interface but is intentionally
// a no-op: rune input has no meaning for this toggle+color widget.
func (ccp *CheckColorPick) TypedRune(r rune) {

}

type (
	// checkColorPickRenderer is the private renderer for CheckColorPick.
	// It holds the raster background and a circular focus indicator overlay.
	checkColorPickRenderer struct {
		col            color.Color      // cached color (currently unused in renderer logic)
		ccp            *CheckColorPick  // reference to the owning widget
		bg             *canvas.Raster   // raster that draws the colored square
		focusIndicator *canvas.Circle   // circle overlaid on the widget to indicate focus
	}
)

// CreateRenderer satisfies fyne.Widget and returns a new renderer for this widget.
// The renderer owns a canvas.Raster (sharing the pixel generator from construction)
// and a canvas.Circle used as a focus indicator ring.
func (ccp *CheckColorPick) CreateRenderer() fyne.WidgetRenderer {
	focusIndicator := canvas.NewCircle(theme.BackgroundColor())
	r := &checkColorPickRenderer{}
	r.bg = canvas.NewRaster(ccp.raster.Generator)
	r.focusIndicator = focusIndicator
	r.ccp = ccp
	r.Refresh()
	return r
}

// MinSize returns the minimum size of the rendered background raster.
func (ccp *checkColorPickRenderer) MinSize() fyne.Size {
	return ccp.bg.MinSize()
}

// Objects returns the canvas objects managed by this renderer. The background
// raster is drawn first, then the focus indicator circle is drawn on top.
func (ccp *checkColorPickRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{ccp.bg, ccp.focusIndicator}
}

// Refresh updates the focus indicator color and redraws the widget.
// When the widget is focused the indicator uses the theme focus color;
// otherwise it is fully transparent (invisible).
func (ccpr *checkColorPickRenderer) Refresh() {
	if ccpr.ccp.focused {
		ccpr.focusIndicator.FillColor = theme.FocusColor()
	} else {
		ccpr.focusIndicator.FillColor = color.Transparent
	}
	ccpr.focusIndicator.Refresh()
	canvas.Refresh(ccpr.ccp)
}

// Destroy is called when the renderer is no longer needed.
// No resources require explicit cleanup for this renderer.
func (ccpr *checkColorPickRenderer) Destroy() {
}

// Layout positions and sizes the focus indicator and the background raster.
// The widget is always forced to be square by clamping both dimensions to
// the smaller of width and height.
func (ccpr *checkColorPickRenderer) Layout(size fyne.Size) {
	// Enforce a square aspect ratio.
	if size.Width < size.Height {
		size.Height = size.Width
	} else {
		size.Width = size.Height
	}
	ccpr.focusIndicator.Resize(size)
	ccpr.bg.Resize(size)
}

// canvasContains reports whether target is present anywhere in the object tree
// rooted at root. It recurses into Containers, AppTabs (selected tab only),
// Scroll containers, Split containers, and generic Widgets via their renderers.
func canvasContains(root fyne.CanvasObject, target fyne.CanvasObject) bool {
	if root == nil || target == nil {
		return false
	}
	if root == target {
		return true
	}
	switch c := root.(type) {
	case *fyne.Container:
		for _, child := range c.Objects {
			if canvasContains(child, target) {
				return true
			}
		}
	case *container.AppTabs:
		// Only search inside the currently selected tab's content.
		if sel := c.Selected(); sel != nil {
			if canvasContains(sel.Content, target) {
				return true
			}
		}
	case *container.Scroll:
		return canvasContains(c.Content, target)
	case *container.Split:
		return canvasContains(c.Leading, target) || canvasContains(c.Trailing, target)
	case fyne.Widget:
		// For generic widgets, introspect their renderer's child objects.
		r := c.CreateRenderer()
		if r != nil {
			for _, child := range r.Objects() {
				if canvasContains(child, target) {
					return true
				}
			}
		}
	}
	return false
}

// isCanvasMounted reports whether target is reachable from canvas c, either
// through the main content tree or through the topmost overlay (e.g. a dialog).
// Returns false if either argument is nil.
func isCanvasMounted(c fyne.Canvas, target fyne.CanvasObject) bool {
	if c == nil || target == nil {
		return false
	}
	// Check the main content hierarchy.
	if canvasContains(c.Content(), target) {
		return true
	}
	// Also check the top overlay (dialogs, menus, etc.).
	if c.Overlays() != nil && c.Overlays().Top() != nil {
		if canvasContains(c.Overlays().Top(), target) {
			return true
		}
	}
	return false
}

// MouseIn is called when the mouse cursor enters the widget area.
// If the widget is mounted on the canvas it requests keyboard focus so that
// TypedKey events are delivered without an explicit click, then refreshes.
func (ccp *CheckColorPick) MouseIn(e *desktop.MouseEvent) {
	if ccp.window != nil && ccp.window.Canvas() != nil && isCanvasMounted(ccp.window.Canvas(), ccp) {
		ccp.window.Canvas().Focus(ccp)
	}
	ccp.Refresh()
}

// MouseDown satisfies desktop.Mouseable; no action is required on mouse button press.
func (ccp *CheckColorPick) MouseDown(e *desktop.MouseEvent) {
}

// MouseUp satisfies desktop.Mouseable; no action is required on mouse button release.
func (ccp *CheckColorPick) MouseUp(e *desktop.MouseEvent) {
}

// MouseMoved satisfies desktop.Mouseable; no action is required on mouse movement.
func (ccp *CheckColorPick) MouseMoved(e *desktop.MouseEvent) {
}

// MouseOut is called when the mouse cursor leaves the widget area.
// It releases keyboard focus and refreshes the widget.
func (ccp *CheckColorPick) MouseOut() {
	ccp.window.Canvas().Unfocus()
	ccp.Refresh()
}

// Compile-time interface assertions: ensure CheckColorPick satisfies the required
// Fyne interfaces. These will cause a build error if any method is missing.
var _ fyne.Focusable = (*CheckColorPick)(nil)
var _ fyne.Tappable = (*CheckColorPick)(nil)
var _ fyne.Widget = (*CheckColorPick)(nil)
var _ desktop.Mouseable = (*CheckColorPick)(nil)
