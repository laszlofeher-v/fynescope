package tastybutton

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/test"
	"github.com/stretchr/testify/assert"
)

func TestTastyButton_New(t *testing.T) {
	test.NewApp()

	tappedCalled := false
	tapped := func() {
		tappedCalled = true
	}
	b := NewTastyButton("Click Me", Green, tapped)

	assert.NotNil(t, b)
	assert.Equal(t, "Click Me", b.Text)
	assert.Equal(t, Green, b.Style)
	assert.False(t, b.hovered)
	assert.False(t, b.pressed)

	b.Tapped(&fyne.PointEvent{})
	assert.True(t, tappedCalled, "OnTapped should be called on Tapped")
}

func TestTastyButton_MouseStates(t *testing.T) {
	test.NewApp()

	b := NewTastyButton("Hover Me", Red, nil)
	renderer := b.CreateRenderer()
	assert.NotNil(t, renderer)

	// MouseIn -> Hovered
	b.MouseIn(&desktop.MouseEvent{})
	assert.True(t, b.hovered)

	// MouseDown -> Pressed (if Primary button)
	b.MouseDown(&desktop.MouseEvent{Button: desktop.MouseButtonPrimary})
	assert.True(t, b.pressed)

	// MouseUp -> Not pressed (if Primary button)
	b.MouseUp(&desktop.MouseEvent{Button: desktop.MouseButtonPrimary})
	assert.False(t, b.pressed)

	// MouseDown with secondary button -> Not pressed
	b.MouseDown(&desktop.MouseEvent{Button: desktop.MouseButtonSecondary})
	assert.False(t, b.pressed)

	// MouseOut -> Not hovered or pressed
	b.MouseIn(&desktop.MouseEvent{})
	b.MouseDown(&desktop.MouseEvent{Button: desktop.MouseButtonPrimary})
	b.MouseOut()
	assert.False(t, b.hovered)
	assert.False(t, b.pressed)

	// Call MouseMoved to cover the empty function
	b.MouseMoved(&desktop.MouseEvent{})
}

func TestTastyButton_Renderer_Refresh(t *testing.T) {
	test.NewApp()

	styles := []Style{Green, Red, Orange}

	for _, s := range styles {
		b := NewTastyButton("Test", s, nil)
		r := b.CreateRenderer().(*tastyButtonRenderer)

		assert.Equal(t, "Test", r.label.Text)

		// Test neutral state colors
		r.Refresh()
		assert.Equal(t, color.White, r.label.Color)

		// Test hover colors
		b.hovered = true
		r.Refresh()

		// Test press colors
		b.hovered = false
		b.pressed = true
		r.Refresh()

		// Ensure MinSize doesn't panic
		min := r.MinSize()
		assert.Greater(t, min.Width, float32(0))
		assert.Greater(t, min.Height, float32(0))

		// Ensure Destroy doesn't panic
		r.Destroy()
	}
}
