package disp16

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/test"
	"github.com/stretchr/testify/assert"
)

func TestNewHexArray(t *testing.T) {
	test.NewApp()

	_, err := NewHexArray(0, nil, color.White, false, "Test")
	assert.Error(t, err)

	disp, err := NewHexArray(4, nil, color.White, false, "Test")
	assert.NoError(t, err)
	assert.NotNil(t, disp)
	assert.Equal(t, 4, disp.numOfDigits)
	assert.Equal(t, 4, len(disp.digits))
	assert.Equal(t, uint64(0), disp.GetValue())
}

func TestSetValue(t *testing.T) {
	test.NewApp()
	disp, _ := NewHexArray(2, nil, color.White, false, "")

	// Value within bounds (2 digits = max 0xFF)
	disp.SetValue(0x42)
	assert.Equal(t, uint64(0x42), disp.GetValue())
	assert.Equal(t, 4, disp.digits[0].val)
	assert.Equal(t, 2, disp.digits[1].val)

	// Value out of bounds (should truncate/clamp to maxVal which is 0xFF)
	disp.SetValue(0x100)
	assert.Equal(t, uint64(0xFF), disp.GetValue())
	assert.Equal(t, 0xF, disp.digits[0].val)
	assert.Equal(t, 0xF, disp.digits[1].val)
}

func TestOnChanged(t *testing.T) {
	test.NewApp()
	disp, _ := NewHexArray(4, nil, color.White, false, "")

	var changedVal uint64
	disp.OnChanged = func(v uint64) {
		changedVal = v
	}

	disp.SetValue(0x1234)
	assert.Equal(t, uint64(0x1234), changedVal)
}

func TestHexArray_DisableEnable(t *testing.T) {
	test.NewApp()
	disp, _ := NewHexArray(4, nil, color.White, false, "Test")

	assert.False(t, disp.AcceptsTab())
	assert.False(t, disp.Disabled())
	disp.Disable()
	assert.False(t, disp.Disabled())
	disp.Enable()
	assert.False(t, disp.Disabled())

	disp.TypedRune('X')
	disp.TypedShortcut(nil)
	disp.KeyUp(nil)
	disp.KeyDown(nil)
	disp.MouseUp(nil)
	disp.DragEnd()
}

func TestHexArray_FocusAndCursor(t *testing.T) {
	test.NewApp()
	disp, _ := NewHexArray(4, nil, color.White, false, "Test")

	// Initial cursor state is digitCursorOut
	assert.Equal(t, desktop.DefaultCursor, disp.Cursor())

	// Focus gained
	disp.FocusGained()
	// Cursor should now be valid if set, or if digitCursor is out, DefaultCursor
	disp.digitCursor = 0
	assert.Equal(t, desktop.PointerCursor, disp.Cursor())

	// Focus lost
	disp.FocusLost()
	assert.Equal(t, digitCursorOut, disp.digitCursor)
	assert.Equal(t, desktop.DefaultCursor, disp.Cursor())

	// Readonly mode
	disp.Readonly = true
	disp.digitCursor = 1
	assert.Equal(t, desktop.DefaultCursor, disp.Cursor())
	disp.FocusGained()
	disp.FocusLost()
}

func TestHexArray_CursorNavigationAndUpDown(t *testing.T) {
	test.NewApp()
	disp, _ := NewHexArray(4, nil, color.White, false, "Test")

	// Up/down when cursor is out should do nothing
	disp.digitCursor = digitCursorOut
	disp.up()
	disp.down()
	assert.Equal(t, uint64(0), disp.GetValue())

	// Move cursor
	disp.digitCursor = 0
	disp.cursorLeft()
	assert.Equal(t, 1, disp.digitCursor)
	disp.cursorLeft()
	assert.Equal(t, 2, disp.digitCursor)
	disp.cursorLeft()
	assert.Equal(t, 3, disp.digitCursor)
	disp.cursorLeft() // clamp at numOfDigits-1
	assert.Equal(t, 3, disp.digitCursor)

	disp.cursorRight()
	assert.Equal(t, 2, disp.digitCursor)
	disp.cursorRight()
	disp.cursorRight()
	assert.Equal(t, 0, disp.digitCursor)
	disp.cursorRight() // clamp at 0
	assert.Equal(t, 0, disp.digitCursor)

	// Test up and down
	disp.digitCursor = 0 // LSB
	disp.up()
	assert.Equal(t, uint64(1), disp.GetValue())
	disp.down()
	assert.Equal(t, uint64(0), disp.GetValue())
	disp.down() // underflow wrap to 0xF
	assert.Equal(t, uint64(0xF), disp.GetValue())
	disp.up() // overflow wrap to 0x0
	assert.Equal(t, uint64(0), disp.GetValue())
}

func TestHexArray_TypedKey(t *testing.T) {
	test.NewApp()
	disp, _ := NewHexArray(4, nil, color.White, false, "Test")
	disp.digitCursor = 3 // MSB

	// TypedKey when Readonly does nothing
	disp.Readonly = true
	disp.TypedKey(&fyne.KeyEvent{Name: fyne.KeyUp})
	assert.Equal(t, uint64(0), disp.GetValue())
	disp.Readonly = false

	// Arrows
	disp.TypedKey(&fyne.KeyEvent{Name: fyne.KeyUp})
	assert.Equal(t, uint64(0x1000), disp.GetValue())
	disp.TypedKey(&fyne.KeyEvent{Name: fyne.KeyDown})
	assert.Equal(t, uint64(0), disp.GetValue())

	disp.TypedKey(&fyne.KeyEvent{Name: fyne.KeyRight})
	assert.Equal(t, 2, disp.digitCursor)
	disp.TypedKey(&fyne.KeyEvent{Name: fyne.KeyLeft})
	assert.Equal(t, 3, disp.digitCursor)

	// Digits 0-9 and A-F
	keys := []struct {
		key      fyne.KeyName
		expected int
	}{
		{fyne.KeyA, 10},
		{fyne.KeyB, 11},
		{fyne.KeyC, 12},
		{fyne.KeyD, 13},
		{fyne.KeyE, 14},
		{fyne.KeyF, 15},
		{fyne.Key0, 0},
		{fyne.Key1, 1},
		{fyne.Key2, 2},
		{fyne.Key3, 3},
		{fyne.Key4, 4},
		{fyne.Key5, 5},
		{fyne.Key6, 6},
		{fyne.Key7, 7},
		{fyne.Key8, 8},
		{fyne.Key9, 9},
	}

	for _, k := range keys {
		disp.digitCursor = 1
		disp.TypedKey(&fyne.KeyEvent{Name: k.key})
		assert.Equal(t, k.expected, disp.digits[2].val)
		assert.Equal(t, 0, disp.digitCursor) // cursor moves right
	}

	// Delete and Backspace
	disp.digitCursor = 1
	disp.TypedKey(&fyne.KeyEvent{Name: fyne.KeyBackspace})
	assert.Equal(t, 0, disp.digits[2].val)

	disp.digitCursor = 1
	disp.TypedKey(&fyne.KeyEvent{Name: fyne.KeyDelete})
	assert.Equal(t, 0, disp.digits[2].val)
}

func TestHexArray_MouseAndInteractions(t *testing.T) {
	a := test.NewApp()
	w := a.NewWindow("Test")
	disp, _ := NewHexArray(4, w, color.White, false, "HEX")
	disp.size = fyne.NewSize(200, 50)

	// Set digit cursor via position
	assert.False(t, disp.setDigitCursor(-10))
	assert.False(t, disp.setDigitCursor(300))

	// MouseIn
	disp.MouseIn(&desktop.MouseEvent{PointEvent: fyne.PointEvent{Position: fyne.NewPos(100, 25)}})
	assert.NotEqual(t, digitCursorOut, disp.digitCursor)

	// MouseMoved
	disp.MouseMoved(&desktop.MouseEvent{PointEvent: fyne.PointEvent{Position: fyne.NewPos(120, 25)}})

	// MouseDown
	disp.MouseDown(&desktop.MouseEvent{Button: desktop.MouseButtonPrimary, PointEvent: fyne.PointEvent{Position: fyne.NewPos(100, 25)}})

	// Tapped
	disp.Tapped(&fyne.PointEvent{Position: fyne.NewPos(100, 25)})

	// Dragged
	disp.Dragged(&fyne.DragEvent{PointEvent: fyne.PointEvent{Position: fyne.NewPos(100, 25)}})

	// Scrolled
	disp.digitCursor = 0
	valBefore := disp.GetValue()
	disp.Scrolled(&fyne.ScrollEvent{PointEvent: fyne.PointEvent{Position: fyne.NewPos(100, 25)}, Scrolled: fyne.NewDelta(0, 10)})
	assert.NotEqual(t, valBefore, disp.GetValue())
	disp.Scrolled(&fyne.ScrollEvent{PointEvent: fyne.PointEvent{Position: fyne.NewPos(100, 25)}, Scrolled: fyne.NewDelta(0, -10)})
	assert.Equal(t, valBefore, disp.GetValue())

	// MouseOut
	disp.MouseOut()
	assert.Equal(t, digitCursorOut, disp.digitCursor)

	// Readonly versions return early
	disp.Readonly = true
	disp.MouseIn(&desktop.MouseEvent{PointEvent: fyne.PointEvent{Position: fyne.NewPos(100, 25)}})
	disp.MouseMoved(&desktop.MouseEvent{PointEvent: fyne.PointEvent{Position: fyne.NewPos(100, 25)}})
	disp.MouseDown(&desktop.MouseEvent{Button: desktop.MouseButtonPrimary, PointEvent: fyne.PointEvent{Position: fyne.NewPos(100, 25)}})
	disp.Tapped(&fyne.PointEvent{Position: fyne.NewPos(100, 25)})
	disp.Dragged(&fyne.DragEvent{PointEvent: fyne.PointEvent{Position: fyne.NewPos(100, 25)}})
	disp.Scrolled(&fyne.ScrollEvent{Scrolled: fyne.NewDelta(0, 10)})
	disp.MouseOut()
}

func TestHexArray_Renderer(t *testing.T) {
	test.NewApp()
	disp, _ := NewHexArray(4, nil, color.White, false, "LABEL")
	disp.SetValue(0xABCD)
	disp.digitCursor = 1 // test cursor coloring

	r := disp.CreateRenderer()
	assert.NotNil(t, r)
	assert.NotEmpty(t, r.Objects())

	min := r.MinSize()
	assert.True(t, min.Width > 0 && min.Height > 0)

	r.Layout(fyne.NewSize(200, 50))
	r.Refresh()
	r.Destroy()

	// Renderer with nil label
	dispNoLabel, _ := NewHexArray(2, nil, color.White, false, "")
	r2 := dispNoLabel.CreateRenderer()
	r2.Layout(fyne.NewSize(100, 40))
	r2.Refresh()
}
