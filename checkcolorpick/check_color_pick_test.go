package checkcolorpick

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/test"
	"github.com/stretchr/testify/assert"
)

func TestCheckColorPick_New(t *testing.T) {
	test.NewApp()
	w := test.NewWindow(nil)
	defer w.Close()

	changedCalled := false
	var changedVal bool
	var changedCol color.Color

	ccp := NewCheckColorPick(w, func(v bool, col color.Color) {
		changedCalled = true
		changedVal = v
		changedCol = col
	}, color.NRGBA{R: 255, G: 0, B: 0, A: 255}, fyne.NewSize(50, 50))

	assert.NotNil(t, ccp)
	assert.Equal(t, fyne.NewSize(50, 50), ccp.MinSize())
	assert.False(t, ccp.Val)

	// Set() method
	ccp.Set()
	assert.True(t, ccp.Val)
	assert.True(t, changedCalled)
	assert.True(t, changedVal)
	assert.Equal(t, color.NRGBA{R: 255, G: 0, B: 0, A: 255}, changedCol)

	// SetColor() method
	changedCalled = false
	ccp.SetColor(color.NRGBA{R: 0, G: 255, B: 0, A: 255})
	assert.True(t, changedCalled)
	assert.Equal(t, color.NRGBA{R: 0, G: 255, B: 0, A: 255}, ccp.col)
}

func TestCheckColorPick_Interactions(t *testing.T) {
	test.NewApp()
	w := test.NewWindow(nil)
	defer w.Close()

	var changedVal bool
	ccp := NewCheckColorPick(w, func(v bool, col color.Color) {
		changedVal = v
	}, color.White, fyne.NewSize(30, 30))

	// Tapped toggles Val
	ccp.Tapped(&fyne.PointEvent{})
	assert.True(t, ccp.Val)
	assert.True(t, changedVal)

	ccp.Tapped(&fyne.PointEvent{})
	assert.False(t, ccp.Val)
	assert.False(t, changedVal)

	// Focus handling
	ccp.FocusGained()
	assert.True(t, ccp.focused)

	ccp.FocusLost()
	assert.False(t, ccp.focused)

	// TypedKey toggles Val
	ccp.TypedKey(&fyne.KeyEvent{})
	assert.True(t, ccp.Val)

	// TypedRune (noop coverage)
	ccp.TypedRune('a')

	// Disable/Enable/Disabled checks
	ccp.Disable()
	ccp.Enable()
	assert.False(t, ccp.Disabled())

	// Hide check
	ccp.Hide()
}

func TestCheckColorPick_MouseEvents(t *testing.T) {
	test.NewApp()
	w := test.NewWindow(nil)
	defer w.Close()

	ccp := NewCheckColorPick(w, func(v bool, col color.Color) {}, color.White, fyne.NewSize(30, 30))

	w.SetContent(ccp)
	w.Show()

	// Mouse events
	ccp.MouseIn(&desktop.MouseEvent{})
	ccp.MouseDown(&desktop.MouseEvent{})
	ccp.MouseMoved(&desktop.MouseEvent{})
	ccp.MouseUp(&desktop.MouseEvent{})
	ccp.MouseOut()
}

func TestCheckColorPick_TappedSecondary(t *testing.T) {
	test.NewApp()
	w := test.NewWindow(nil)
	defer w.Close()

	ccp := NewCheckColorPick(w, func(v bool, col color.Color) {}, color.White, fyne.NewSize(30, 30))

	w.SetContent(ccp)
	w.Show()

	// Simply verify that calling TappedSecondary doesn't crash (dialog shows up asynchronously in test harness)
	ccp.TappedSecondary(&fyne.PointEvent{})
}

func TestCheckColorPick_Renderer(t *testing.T) {
	test.NewApp()
	w := test.NewWindow(nil)
	defer w.Close()

	ccp := NewCheckColorPick(w, func(v bool, col color.Color) {}, color.White, fyne.NewSize(30, 30))
	renderer := ccp.CreateRenderer()
	assert.NotNil(t, renderer)

	assert.Equal(t, fyne.NewSize(1, 1), renderer.MinSize())

	objects := renderer.Objects()
	assert.Len(t, objects, 2)

	renderer.Layout(fyne.NewSize(40, 40))

	ccp.focused = true
	renderer.Refresh()

	ccp.focused = false
	renderer.Refresh()

	renderer.Destroy()
}
