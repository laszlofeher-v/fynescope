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
	readonly := d7.Readonly
	d7.lock.Unlock()

	if readonly {
		return
	}
	d7.clearDigitAt(event.Position.X)

	d7.lock.Lock()
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
	d7.lock.Lock()
	readonly := d7.Readonly
	cursor := d7.digitCursor
	value := d7.Value
	d7.lock.Unlock()

	if readonly {
		return
	}
	d := int(math.Round(math.Pow(10, float64(cursor))))
	switch k.Name {
	case fyne.KeyUp:
		d7.SetValue(value + d)
	case fyne.KeyDown:
		d7.down(cursor)
	case fyne.KeyLeft:
		d7.lock.Lock()
		d7.cursorLeft()
		d7.lock.Unlock()
	case fyne.KeyRight:
		d7.lock.Lock()
		d7.cursorRight()
		d7.lock.Unlock()
	case fyne.KeyDelete:
		d7.setDigitAtDigitCursor(0)
		d7.lock.Lock()
		d7.cursorRight()
		d7.lock.Unlock()
	case fyne.KeyBackspace:
		d7.setDigitAtDigitCursor(0)
		d7.lock.Lock()
		d7.cursorLeft()
		d7.lock.Unlock()
	case fyne.Key0:
		d7.setDigitAtDigitCursor(0)
		d7.lock.Lock()
		d7.cursorRight()
		d7.lock.Unlock()
	case fyne.Key1:
		d7.setDigitAtDigitCursor(1)
		d7.lock.Lock()
		d7.cursorRight()
		d7.lock.Unlock()
	case fyne.Key2:
		d7.setDigitAtDigitCursor(2)
		d7.lock.Lock()
		d7.cursorRight()
		d7.lock.Unlock()
	case fyne.Key3:
		d7.setDigitAtDigitCursor(3)
		d7.lock.Lock()
		d7.cursorRight()
		d7.lock.Unlock()
	case fyne.Key4:
		d7.setDigitAtDigitCursor(4)
		d7.lock.Lock()
		d7.cursorRight()
		d7.lock.Unlock()
	case fyne.Key5:
		d7.setDigitAtDigitCursor(5)
		d7.lock.Lock()
		d7.cursorRight()
		d7.lock.Unlock()
	case fyne.Key6:
		d7.setDigitAtDigitCursor(6)
		d7.lock.Lock()
		d7.cursorRight()
		d7.lock.Unlock()
	case fyne.Key7:
		d7.setDigitAtDigitCursor(7)
		d7.lock.Lock()
		d7.cursorRight()
		d7.lock.Unlock()
	case fyne.Key8:
		d7.setDigitAtDigitCursor(8)
		d7.lock.Lock()
		d7.cursorRight()
		d7.lock.Unlock()
	case fyne.Key9:
		d7.setDigitAtDigitCursor(9)
		d7.lock.Lock()
		d7.cursorRight()
		d7.lock.Unlock()
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
	d7.lock.Lock()
	tmp := d7.Value
	if tmp >= 0 {
		tmp -= int(math.Pow(10, float64(index))) * d7.digits[index].val
		tmp += int(math.Pow(10, float64(index))) * digit
	} else {
		tmp += int(math.Pow(10, float64(index))) * d7.digits[index].val
		tmp -= int(math.Pow(10, float64(index))) * digit
	}
	d7.lock.Unlock()

	d7.SetValue(tmp)
	d7.Refresh()
}

func (d7 *DigitArray) setDigitAtDigitCursor(digit int) {
	d7.lock.Lock()
	cursor := d7.digitCursor
	valid := cursor >= 0 && cursor < len(d7.digits)
	d7.lock.Unlock()

	if valid {
		d7.setDigitAtIndex(digit, cursor)
	}
}

func (d7 *DigitArray) setDigitAtX(x float32, digit int) {
	d7.lock.Lock()
	targetIndex := -1
	for digitIndex := range d7.digits {
		pos := d7.digits[digitIndex].relPos
		if x > pos.X {
			targetIndex = digitIndex
			break
		}
	}
	d7.lock.Unlock()

	if targetIndex != -1 {
		d7.setDigitAtIndex(digit, targetIndex)
	}
}

func (d7 *DigitArray) clearDigitAt(x float32) {
	d7.setDigitAtX(x, 0)
}
func (d7 *DigitArray) Dragged(event *fyne.DragEvent) {
	d7.lock.Lock()
	readonly := d7.Readonly
	d7.lock.Unlock()

	if readonly {
		return
	}
	d7.clearDigitAt(event.Position.X)
}
func (d7 *DigitArray) DragEnd() {
}

func (d7 *DigitArray) MouseDown(event *desktop.MouseEvent) {
	d7.lock.Lock()
	readonly := d7.Readonly
	d7.lock.Unlock()

	if readonly {
		return
	}
	if event.Button == desktop.MouseButtonPrimary {
		d7.clearDigitAt(event.Position.X)
	}
}

func (d7 *DigitArray) MouseUp(event *desktop.MouseEvent) {
}

func (d7 *DigitArray) down(digitIndex int) {
	d7.lock.Lock()
	if digitIndex < 0 || digitIndex >= len(d7.digits) || d7.Value <= d7.minValue {
		d7.lock.Unlock()
		return
	}
	step := int(math.Pow(10, float64(digitIndex)))
	var newVal *int
	switch {
	case d7.digits[digitIndex].val >= 1:
		v := d7.Value - step
		newVal = &v
	case d7.digits[digitIndex].val == 0:
		for i := digitIndex + 1; i < len(d7.digits); i++ {
			if d7.digits[i].val > 0 {
				v := d7.Value - step
				newVal = &v
				break
			}
		}
		if newVal == nil {
			for i := digitIndex; i >= 0; i-- {
				step = int(math.Pow(10, float64(i)))
				if d7.digits[i].val > 0 {
					v := d7.Value - step
					newVal = &v
					break
				}
			}
		}
		if newVal == nil {
			step = int(math.Pow(10, float64(digitIndex)))
			v := d7.Value - step
			newVal = &v
		}
	}
	d7.lock.Unlock()

	if newVal != nil {
		d7.SetValue(*newVal)
	}
}

func (d7 *DigitArray) Scrolled(event *fyne.ScrollEvent) {
	d7.lock.Lock()
	readonly := d7.Readonly
	cursor := d7.digitCursor
	value := d7.Value
	d7.lock.Unlock()

	if readonly {
		return
	}
	step := int(math.Pow(10, float64(cursor)))
	if event.Scrolled.DY > 0 {
		d7.SetValue(value + step)
	} else {
		d7.down(cursor)
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
	d7.lock.Lock()
	readonly := d7.Readonly
	d7.lock.Unlock()

	if readonly {
		return
	}
	d7.lock.Lock()
	d7.mousePos = e.Position
	focus := d7.setDigitCursor(e.Position.X)
	d7.lock.Unlock()

	if focus {
		d7.Window.Canvas().Focus(d7)
	}
	d7.Refresh()
}

func (d7 *DigitArray) MouseMoved(e *desktop.MouseEvent) {
	d7.lock.Lock()
	readonly := d7.Readonly
	d7.lock.Unlock()

	if readonly {
		return
	}
	d7.lock.Lock()
	d7.mousePos = e.Position
	focus := d7.setDigitCursor(e.Position.X)
	d7.lock.Unlock()

	if focus {
		d7.Window.Canvas().Focus(d7)
	}
	d7.Refresh()
}

func (d7 *DigitArray) MouseOut() {
	d7.lock.Lock()
	readonly := d7.Readonly
	d7.lock.Unlock()

	if readonly {
		return
	}
	d7.Window.Canvas().Unfocus()
	d7.lock.Lock()
	d7.digitCursor = digitCursorOut
	d7.lock.Unlock()
	d7.Refresh()
}
