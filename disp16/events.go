package disp16

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

var _ fyne.Disableable = (*HexArray)(nil)
var _ fyne.Draggable = (*HexArray)(nil)
var _ fyne.Focusable = (*HexArray)(nil)
var _ fyne.Tappable = (*HexArray)(nil)
var _ fyne.Widget = (*HexArray)(nil)
var _ desktop.Mouseable = (*HexArray)(nil)
var _ desktop.Keyable = (*HexArray)(nil)
var _ fyne.Tabbable = (*HexArray)(nil)

func (h *HexArray) AcceptsTab() bool {
	return false
}

func (h *HexArray) Disable() {}
func (h *HexArray) Enable()  {}
func (h *HexArray) Disabled() bool {
	return false
}

func (h *HexArray) Tapped(event *fyne.PointEvent) {
	h.lock.Lock()
	readonly := h.Readonly
	h.lock.Unlock()

	if readonly {
		return
	}

	h.lock.Lock()
	h.mousePos = event.Position
	focus := h.setDigitCursor(event.Position.X)
	h.lock.Unlock()
	h.setDigitAtDigitCursor(0)

	if focus {
		h.Window.Canvas().Focus(h)
	}
	h.Refresh()
}

func (h *HexArray) Cursor() desktop.Cursor {
	h.lock.Lock()
	defer h.lock.Unlock()
	if h.Readonly || h.digitCursor == digitCursorOut {
		return desktop.DefaultCursor
	}
	return desktop.PointerCursor
}

func (h *HexArray) cursorLeft() {
	if h.digitCursor < h.numOfDigits-1 {
		h.digitCursor++
	}
}

func (h *HexArray) cursorRight() {
	if h.digitCursor > 0 {
		h.digitCursor--
	}
}

func (h *HexArray) setDigitAtDigitCursor(val int) {
	if h.digitCursor == digitCursorOut {
		return
	}
	h.setDigitAtIndex(val)
}

func (h *HexArray) setDigitAtIndex(digit int) {
	h.lock.Lock()
	tmp := h.Value
	// clear the current digit at index
	shift := h.digitCursor * 4
	mask := uint64(0xF) << shift
	tmp = tmp & ^mask
	// set the new digit
	tmp = tmp | (uint64(digit) << shift)
	h.lock.Unlock()
	h.SetValue(tmp)
	h.Refresh()
}

func (h *HexArray) TypedKey(k *fyne.KeyEvent) {
	h.lock.Lock()
	readonly := h.Readonly
	h.lock.Unlock()

	if readonly {
		return
	}

	switch k.Name {
	case fyne.KeyUp:
		h.up()
	case fyne.KeyDown:
		h.down()
	case fyne.KeyLeft:
		h.lock.Lock()
		h.cursorLeft()
		h.lock.Unlock()
	case fyne.KeyRight:
		h.lock.Lock()
		h.cursorRight()
		h.lock.Unlock()
	case fyne.KeyDelete, fyne.KeyBackspace, fyne.Key0:
		h.setDigitAtDigitCursor(0)
		h.lock.Lock()
		h.cursorRight()
		h.lock.Unlock()
	case fyne.Key1:
		h.setDigitAtDigitCursor(1)
		h.lock.Lock()
		h.cursorRight()
		h.lock.Unlock()
	case fyne.Key2:
		h.setDigitAtDigitCursor(2)
		h.lock.Lock()
		h.cursorRight()
		h.lock.Unlock()
	case fyne.Key3:
		h.setDigitAtDigitCursor(3)
		h.lock.Lock()
		h.cursorRight()
		h.lock.Unlock()
	case fyne.Key4:
		h.setDigitAtDigitCursor(4)
		h.lock.Lock()
		h.cursorRight()
		h.lock.Unlock()
	case fyne.Key5:
		h.setDigitAtDigitCursor(5)
		h.lock.Lock()
		h.cursorRight()
		h.lock.Unlock()
	case fyne.Key6:
		h.setDigitAtDigitCursor(6)
		h.lock.Lock()
		h.cursorRight()
		h.lock.Unlock()
	case fyne.Key7:
		h.setDigitAtDigitCursor(7)
		h.lock.Lock()
		h.cursorRight()
		h.lock.Unlock()
	case fyne.Key8:
		h.setDigitAtDigitCursor(8)
		h.lock.Lock()
		h.cursorRight()
		h.lock.Unlock()
	case fyne.Key9:
		h.setDigitAtDigitCursor(9)
		h.lock.Lock()
		h.cursorRight()
		h.lock.Unlock()
	case fyne.KeyA:
		h.setDigitAtDigitCursor(10)
		h.lock.Lock()
		h.cursorRight()
		h.lock.Unlock()
	case fyne.KeyB:
		h.setDigitAtDigitCursor(11)
		h.lock.Lock()
		h.cursorRight()
		h.lock.Unlock()
	case fyne.KeyC:
		h.setDigitAtDigitCursor(12)
		h.lock.Lock()
		h.cursorRight()
		h.lock.Unlock()
	case fyne.KeyD:
		h.setDigitAtDigitCursor(13)
		h.lock.Lock()
		h.cursorRight()
		h.lock.Unlock()
	case fyne.KeyE:
		h.setDigitAtDigitCursor(14)
		h.lock.Lock()
		h.cursorRight()
		h.lock.Unlock()
	case fyne.KeyF:
		h.setDigitAtDigitCursor(15)
		h.lock.Lock()
		h.cursorRight()
		h.lock.Unlock()
	default:
	}
	h.Refresh()
}

func (h *HexArray) TypedRune(r rune) {}
func (h *HexArray) FocusGained() {
	h.lock.Lock()
	defer h.lock.Unlock()
	if h.Readonly {
		return
	}
	h.Refresh()
}

func (h *HexArray) FocusLost() {
	h.lock.Lock()
	defer h.lock.Unlock()
	if h.Readonly {
		return
	}
	h.digitCursor = digitCursorOut
	h.Refresh()
}

func (h *HexArray) TypedShortcut(r fyne.Shortcut) {}
func (h *HexArray) KeyUp(k *fyne.KeyEvent)        {}
func (h *HexArray) KeyDown(k *fyne.KeyEvent)      {}

func (h *HexArray) MouseIn(e *desktop.MouseEvent) {
	h.lock.Lock()
	readonly := h.Readonly
	h.lock.Unlock()

	if readonly {
		return
	}
	h.lock.Lock()
	h.mousePos = e.Position
	focus := h.setDigitCursor(e.Position.X)
	h.lock.Unlock()

	if focus {
		h.Window.Canvas().Focus(h)
	}
	h.Refresh()
}
func (h *HexArray) MouseOut() {
	h.lock.Lock()
	readonly := h.Readonly
	h.lock.Unlock()

	if readonly {
		return
	}
	h.Window.Canvas().Unfocus()
	h.lock.Lock()
	h.digitCursor = digitCursorOut
	h.lock.Unlock()
	h.Refresh()
}
func (h *HexArray) MouseMoved(e *desktop.MouseEvent) {
	h.lock.Lock()
	readonly := h.Readonly
	h.lock.Unlock()

	if readonly {
		return
	}
	h.lock.Lock()
	h.mousePos = e.Position
	focus := h.setDigitCursor(e.Position.X)
	h.lock.Unlock()
	if focus {
		h.Window.Canvas().Focus(h)
	}
	h.Refresh()

}
func (h *HexArray) MouseDown(event *desktop.MouseEvent) {
	h.lock.Lock()
	readonly := h.Readonly
	h.lock.Unlock()

	if readonly {
		return
	}
	if event.Button == desktop.MouseButtonPrimary {
		focus := h.setDigitCursor(event.Position.X)
		h.setDigitAtDigitCursor(0)
		if focus {
			h.Window.Canvas().Focus(h)
		}
		h.Refresh()
	}
}
func (h *HexArray) MouseUp(event *desktop.MouseEvent) {}

func (h *HexArray) Dragged(event *fyne.DragEvent) {
	h.lock.Lock()
	readonly := h.Readonly
	h.lock.Unlock()
	if readonly {
		return
	}
	focus := h.setDigitCursor(event.Position.X)
	h.setDigitAtDigitCursor(0)
	if focus {
		h.Window.Canvas().Focus(h)
	}
	h.Refresh()
}
func (h *HexArray) DragEnd() {}

func (h *HexArray) setDigitCursor(x float32) bool {
	if x < 0 || x > h.size.Width {
		h.digitCursor = digitCursorOut
		return false
	}

	// reverse map X position to digit index
	digitWidth := float32(5)*h.dotSize + float32(4)*h.dotSpacing
	labelW := float32(0)
	if h.label != nil {
		labelW = h.label.MinSize().Width + h.spaceBetweenDigits
	}

	if x < labelW {
		return false
	}

	x -= labelW
	idx := int(x / (digitWidth + h.spaceBetweenDigits))
	if idx >= 0 && idx < h.numOfDigits {
		h.digitCursor = h.numOfDigits - 1 - idx
		return true
	}
	return false
}

func (h *HexArray) up() {
	if h.digitCursor == digitCursorOut {
		return
	}
	h.lock.Lock()
	val := h.digits[h.numOfDigits-1-h.digitCursor].val
	h.lock.Unlock()
	val = (val + 1) & 0xF
	h.setDigitAtIndex(val)
}

func (h *HexArray) down() {
	if h.digitCursor == digitCursorOut {
		return
	}
	h.lock.Lock()
	val := h.digits[h.numOfDigits-1-h.digitCursor].val
	h.lock.Unlock()
	val = (val - 1) & 0xF
	h.setDigitAtIndex(val)
}

func (h *HexArray) Scrolled(event *fyne.ScrollEvent) {
	h.lock.Lock()
	readonly := h.Readonly
	h.lock.Unlock()

	if readonly {
		return
	}

	h.lock.Lock()
	focus := h.setDigitCursor(event.Position.X)
	h.lock.Unlock()

	if !focus {
		return
	}

	h.Window.Canvas().Focus(h)
	if event.Scrolled.DY > 0 {
		h.up()
	} else if event.Scrolled.DY < 0 {
		h.down()
	}
}
