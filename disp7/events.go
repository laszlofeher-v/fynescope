package disp7

import (
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

var _ fyne.Disableable = (*DigitArray)(nil)
var _ fyne.Draggable = (*DigitArray)(nil)
var _ fyne.Focusable = (*DigitArray)(nil)
var _ fyne.Tappable = (*DigitArray)(nil)
var _ fyne.Widget = (*DigitArray)(nil)
var _ desktop.Mouseable = (*DigitArray)(nil)
var _ desktop.Keyable = (*DigitArray)(nil)
var _ fyne.Tabbable = (*DigitArray)(nil)

func (d7 *DigitArray) AcceptsTab() bool {
	return false
}

func (d7 *DigitArray) Disable() {
}
func (d7 *DigitArray) Enable() {
}
func (d7 *DigitArray) Disabled() bool {
	return false
}
func (d7 *DigitArray) Tapped(event *fyne.PointEvent) {
	d7.lock.Lock()
	if d7.Readonly {
		d7.lock.Unlock()
		return
	}
	d7.clearDigitAt(event.Position.X)
	d7.mousePos = event.Position
	focus := d7.setDigitCursor(event.Position.X)
	d7.lock.Unlock()

	if focus {
		d7.Window.Canvas().Focus(d7)
	}
	d7.Refresh()
}

func (d7 *DigitArray) Cursor() desktop.Cursor {
	defer d7.lock.Unlock()
	d7.lock.Lock()
	if d7.Readonly || d7.digitCursor == digitCursorOut {
		return desktop.DefaultCursor
	}
	return desktop.PointerCursor
}
func (d7 *DigitArray) cursorLeft() {
	if d7.digitCursor < len(d7.digits)-1 {
		d7.digitCursor++
	}
}
func (d7 *DigitArray) cursorRight() {
	if d7.digitCursor > 0 {
		d7.digitCursor--
	}
}
func (d7 *DigitArray) TypedKey(k *fyne.KeyEvent) {
	defer d7.lock.Unlock()
	d7.lock.Lock()
	if d7.Readonly {
		return
	}
	d := int(math.Round(math.Pow(10, float64(d7.digitCursor))))
	switch k.Name {
	case fyne.KeyUp:
		d7.SetValue(d7.Value + d)
	case fyne.KeyDown:
		d7.down(d7.digitCursor)
	case fyne.KeyLeft:
		d7.cursorLeft()
	case fyne.KeyRight:
		d7.cursorRight()
	case fyne.KeyDelete:
		d7.setDigitAtDigitCursor(0)
		d7.cursorRight()
	case fyne.KeyBackspace:
		d7.setDigitAtDigitCursor(0)
		d7.cursorLeft()
	case fyne.Key0:
		d7.setDigitAtDigitCursor(0)
		d7.cursorRight()
	case fyne.Key1:
		d7.setDigitAtDigitCursor(1)
		d7.cursorRight()
	case fyne.Key2:
		d7.setDigitAtDigitCursor(2)
		d7.cursorRight()
	case fyne.Key3:
		d7.setDigitAtDigitCursor(3)
		d7.cursorRight()
	case fyne.Key4:
		d7.setDigitAtDigitCursor(4)
		d7.cursorRight()
	case fyne.Key5:
		d7.setDigitAtDigitCursor(5)
		d7.cursorRight()
	case fyne.Key6:
		d7.setDigitAtDigitCursor(6)
		d7.cursorRight()
	case fyne.Key7:
		d7.setDigitAtDigitCursor(7)
		d7.cursorRight()
	case fyne.Key8:
		d7.setDigitAtDigitCursor(8)
		d7.cursorRight()
	case fyne.Key9:
		d7.setDigitAtDigitCursor(9)
		d7.cursorRight()
	default:
	}
	d7.Refresh()
}
func (d7 *DigitArray) TypedRune(r rune) {
}
func (d7 *DigitArray) FocusGained() {
	defer d7.lock.Unlock()
	d7.lock.Lock()
	if d7.Readonly {
		return
	}
	d7.digitCursor = len(d7.digits) - 1 // default when TAB pressed
	d7.Refresh()
}

func (d7 *DigitArray) FocusLost() {
	defer d7.lock.Unlock()
	d7.lock.Lock()
	if d7.Readonly {
		return
	}
	d7.digitCursor = digitCursorOut
	d7.Refresh()
}

func (d7 *DigitArray) TypedShortcut(r fyne.Shortcut) {
}
func (d7 *DigitArray) KeyUp(k *fyne.KeyEvent) {
}
func (d7 *DigitArray) KeyDown(k *fyne.KeyEvent) {
}

func (d7 *DigitArray) setDigitAtIndex(digit, index int) {
	tmp := d7.Value
	if tmp >= 0 {
		tmp -= int(math.Pow(10, float64(index))) * d7.digits[index].val
		tmp += int(math.Pow(10, float64(index))) * digit
	} else {
		tmp += int(math.Pow(10, float64(index))) * d7.digits[index].val
		tmp -= int(math.Pow(10, float64(index))) * digit
	}
	d7.SetValue(tmp)
	d7.Refresh()
}

func (d7 *DigitArray) setDigitAtDigitCursor(digit int) {
	if d7.digitCursor >= 0 && d7.digitCursor < len(d7.digits) {
		d7.setDigitAtIndex(digit, d7.digitCursor)
	}
}

func (d7 *DigitArray) setDigitAtX(x float32, digit int) {
	for digitIndex := range d7.digits {
		pos := d7.digits[digitIndex].relPos
		if x > pos.X {
			d7.setDigitAtIndex(digit, digitIndex)
			break
		}
	}
}

func (d7 *DigitArray) clearDigitAt(x float32) {
	d7.setDigitAtX(x, 0)
}
func (d7 *DigitArray) Dragged(event *fyne.DragEvent) {
	if d7.Readonly {
		return
	}
	d7.clearDigitAt(event.Position.X)
}
func (d7 *DigitArray) DragEnd() {
}

func (d7 *DigitArray) MouseDown(event *desktop.MouseEvent) {
	if d7.Readonly {
		return
	}
	if event.Button == desktop.MouseButtonPrimary {
		d7.clearDigitAt(event.Position.X)
	}
}

func (d7 *DigitArray) MouseUp(event *desktop.MouseEvent) {
}

func (d7 *DigitArray) down(digitIndex int) {
	if digitIndex < 0 || digitIndex >= len(d7.digits) || d7.Value <= d7.minValue {
		return
	}
	step := int(math.Pow(10, float64(digitIndex)))
	switch {
	case d7.digits[digitIndex].val >= 1:
		d7.SetValue(d7.Value - step)
	case d7.digits[digitIndex].val == 0:
		for i := digitIndex + 1; i < len(d7.digits); i++ {
			if d7.digits[i].val > 0 {
				d7.SetValue(d7.Value - step)
				return
			}
		}
		for i := digitIndex; i >= 0; i-- {
			step = int(math.Pow(10, float64(i)))
			if d7.digits[i].val > 0 {
				d7.SetValue(d7.Value - step)
				return
			}
		}
		step = int(math.Pow(10, float64(digitIndex)))
		d7.SetValue(d7.Value - step)
	}
}

func (d7 *DigitArray) Scrolled(event *fyne.ScrollEvent) {
	if d7.Readonly {
		return
	}
	step := int(math.Pow(10, float64(d7.digitCursor)))
	if event.Scrolled.DY > 0 {
		d7.SetValue(d7.Value + step)
	} else {
		d7.down(d7.digitCursor)
	}
	d7.Refresh()
}

func (d7 *DigitArray) setDigitCursor(x float32) bool {
	for digitIndex := range d7.digits {
		pos := d7.digits[digitIndex].relPos
		if x > pos.X {
			d7.digitCursor = digitIndex
			break
		}
	}
	return true
}
func (d7 *DigitArray) MouseIn(e *desktop.MouseEvent) {
	if d7.Readonly {
		return
	}
	d7.mousePos = e.Position
	if d7.setDigitCursor(e.Position.X) {
		d7.Window.Canvas().Focus(d7)
	}
	d7.Refresh()
}

func (d7 *DigitArray) MouseMoved(e *desktop.MouseEvent) {
	if d7.Readonly {
		return
	}
	d7.mousePos = e.Position
	if d7.setDigitCursor(e.Position.X) {
		d7.Window.Canvas().Focus(d7)
	}
	d7.Refresh()
}

func (d7 *DigitArray) MouseOut() {
	if d7.Readonly {
		return
	}
	d7.Window.Canvas().Unfocus()
	d7.digitCursor = digitCursorOut
	d7.Refresh()
}
