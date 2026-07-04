package sliderscroll

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/test"
	"github.com/stretchr/testify/assert"
)

func TestSliderScroll_New(t *testing.T) {
	test.NewApp()

	ss := NewSliderScroll(0, 100)
	assert.NotNil(t, ss)
	assert.Equal(t, 0.0, ss.Value)
	assert.Equal(t, 0.0, ss.Min)
	assert.Equal(t, 100.0, ss.Max)
	assert.Equal(t, 1.0, ss.Step)
	assert.Equal(t, defaultMul, ss.mul)
}

func TestSliderScroll_SilentSetValue(t *testing.T) {
	test.NewApp()

	ss := NewSliderScroll(0, 100)
	onChangedCalled := false
	ss.OnChanged = func(val float64) {
		onChangedCalled = true
	}

	ss.SilentSetValue(50)
	assert.Equal(t, 50.0, ss.Value)
	assert.False(t, onChangedCalled, "OnChanged should not be called during SilentSetValue")

	// Verify standard SetValue still calls OnChanged
	ss.SetValue(60)
	assert.Equal(t, 60.0, ss.Value)
	assert.True(t, onChangedCalled, "OnChanged should be called during standard SetValue")
}

func TestSliderScroll_MouseDown_Multiplier(t *testing.T) {
	test.NewApp()

	ss := NewSliderScroll(0, 100)
	assert.Equal(t, 100.0, ss.mul)

	// Middle click (MouseButtonTertiary) increases multiplier by 10x
	ss.MouseDown(&desktop.MouseEvent{Button: desktop.MouseButtonTertiary})
	assert.Equal(t, 1000.0, ss.mul)

	// Right click (MouseButtonSecondary) decreases multiplier by 10x
	ss.MouseDown(&desktop.MouseEvent{Button: desktop.MouseButtonSecondary})
	assert.Equal(t, 100.0, ss.mul)

	// Test lower limit of multiplier (10)
	for i := 0; i < 5; i++ {
		ss.MouseDown(&desktop.MouseEvent{Button: desktop.MouseButtonSecondary})
	}
	assert.Equal(t, 1.0, ss.mul) // it stops decreasing below 10 because: if slScr.mul >= 10 { slScr.mul = slScr.mul / 10 }. Wait! 10 >= 10 -> 1.0. 1.0 < 10 -> does not divide anymore. So 1.0 is lower limit.

	// Test upper limit (1e6)
	for i := 0; i < 10; i++ {
		ss.MouseDown(&desktop.MouseEvent{Button: desktop.MouseButtonTertiary})
	}
	assert.Equal(t, 1e6, ss.mul)
}

func TestSliderScroll_Scrolled(t *testing.T) {
	test.NewApp()

	ss := NewSliderScroll(0, 1000)
	ss.SetValue(100)
	ss.mul = 50.0

	// Scroll up (positive DY) increases value by mul * DY
	ss.Scrolled(&fyne.ScrollEvent{Scrolled: fyne.NewDelta(0, 2.0)})
	assert.Equal(t, 200.0, ss.Value) // 100 + 50 * 2.0

	// Scroll down (negative DY) decreases value by mul * DY
	ss.Scrolled(&fyne.ScrollEvent{Scrolled: fyne.NewDelta(0, -3.0)})
	assert.Equal(t, 50.0, ss.Value) // 200 + 50 * -3.0
}
