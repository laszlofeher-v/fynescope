package gui

import (
	"testing"

	"fyne.io/fyne/v2/test"
	"github.com/stretchr/testify/assert"
)

func TestAwgEditor_Clear(t *testing.T) {
	test.NewApp()
	editor := newAwgEditorWidget()
	editor.values = []float64{0.5, -0.5, 1.0, -1.0}
	
	editor.clear()
	
	for _, v := range editor.values {
		assert.Equal(t, 0.0, v)
	}
	assert.False(t, editor.drawing)
}

func TestAwgEditor_GenerateWaveform(t *testing.T) {
	test.NewApp()
	editor := newAwgEditorWidget()
	editor.values = make([]float64, 4) // 4 samples for easy validation
	
	// Test DC
	editor.generateWaveform("DC")
	for _, v := range editor.values {
		assert.Equal(t, 0.0, v)
	}

	// Test Square
	editor.generateWaveform("Square")
	assert.Equal(t, 1.0, editor.values[0])
	assert.Equal(t, 1.0, editor.values[1])
	assert.Equal(t, -1.0, editor.values[2])
	assert.Equal(t, -1.0, editor.values[3])

	// Test RampUp
	editor.generateWaveform("RampUp")
	assert.Equal(t, -1.0, editor.values[0]) // -1 + 0*2 = -1
	assert.Equal(t, -0.5, editor.values[1]) // -1 + 0.25*2 = -0.5
	assert.Equal(t, 0.0, editor.values[2])  // -1 + 0.5*2 = 0
	assert.Equal(t, 0.5, editor.values[3])  // -1 + 0.75*2 = 0.5
}

func TestAwgEditor_ResizeValues(t *testing.T) {
	test.NewApp()
	editor := newAwgEditorWidget()
	editor.values = []float64{0.0, 0.5, 1.0, 0.5, 0.0} // 5 elements

	// Resize up to 10
	editor.resizeValues(10)
	assert.Len(t, editor.values, 10)
	
	// Check a few linearly interpolated values
	assert.Equal(t, 0.0, editor.values[0])
	assert.Equal(t, 1.0, editor.values[4]) // Element 4 maps to origIdx 2 (1.0)
	assert.Equal(t, 0.0, editor.values[9]) // Element 9 maps to origIdx 4 (0.0)

	// Resize back down to 3
	editor.resizeValues(3)
	assert.Len(t, editor.values, 3)
}
