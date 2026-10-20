package gui

import (
	"fynescope/settings"
	"image"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/stretchr/testify/assert"
)

func TestFvViewer_MouseIn(t *testing.T) {
	rect := image.Rect(10, 20, 100, 50)
	img := image.NewRGBA(image.Rect(0, 0, 200, 100))
	scp := &ScpDesc{
		Settings: &settings.PsSettings{},
	}

	fv := newFvViewer(img, rect, scp)
	fv.labelBounds[0] = image.Rect(10, 20, 100, 50) // Mock a label bound

	assert.True(t, fv.mouseIn(15, 25), "Point inside should be true")
	assert.True(t, fv.mouseIn(10, 20), "Point on top-left border should be true")

	// Max bounds are exclusive in image.Rect
	assert.False(t, fv.mouseIn(100, 49), "Point on right border should be false")
	assert.False(t, fv.mouseIn(99, 50), "Point on bottom border should be false")
	assert.False(t, fv.mouseIn(5, 25), "Point outside left should be false")
	assert.False(t, fv.mouseIn(15, 10), "Point outside top should be false")
}

func TestFvViewer_MouseEvents(t *testing.T) {
	rect := image.Rect(10, 20, 100, 50)
	img := image.NewRGBA(image.Rect(0, 0, 200, 100))
	scp := &ScpDesc{
		Settings:            &settings.PsSettings{},
		fvScopeSignalScreen: image.NewRGBA(image.Rect(10, 20, 100, 50)),
	}

	fv := newFvViewer(img, rect, scp)

	// Initial state
	assert.False(t, fv.showInspector)
	assert.False(t, fv.refActive)
	assert.Equal(t, -1, fv.selectedChannel)

	// Mouse down inside rect with RightMouseButton -> shows inspector
	fv.mouseDown(desktop.MouseButtonSecondary, 0, 15, 25)
	assert.True(t, fv.showInspector)
	assert.False(t, fv.refActive)

	// Mouse up
	fv.mouseUp(desktop.MouseButtonSecondary, 0, 15, 25)
	assert.False(t, fv.showInspector)

	// Mouse down inside rect with RightMouseButton + Shift -> sets reference
	fv.mouseDown(desktop.MouseButtonSecondary, fyne.KeyModifierShift, 15, 25)
	assert.True(t, fv.refActive)
	assert.True(t, fv.refDragging)

	fv.mouseUp(desktop.MouseButtonSecondary, fyne.KeyModifierShift, 15, 25)
	assert.False(t, fv.refDragging)
}

func TestFvViewer_TypedKey(t *testing.T) {
	rect := image.Rect(10, 20, 100, 50)
	img := image.NewRGBA(image.Rect(0, 0, 200, 100))
	scp := &ScpDesc{
		Settings: &settings.PsSettings{},
	}

	fv := newFvViewer(img, rect, scp)

	// Make sure typedKey doesn't panic on a dummy scope
	fv.typedKey(20, 30, fyne.KeyDown)
	fv.typedKey(20, 30, fyne.KeyUp)

	assert.True(t, true)
}

func TestFvViewer_CalcValuesAt(t *testing.T) {
	st := settings.NewDefaultSettings()
	st.Channels[0].Enabled = true
	st.Channels[0].FvMode = settings.FvArgument
	st.Channels[0].VRange = 0 // 10mV or RangeValuesMv[0]
	st.Channels[0].DisplayVOffset = 0

	st.Channels[1].Enabled = true
	st.Channels[1].FvMode = settings.FvValue
	st.Channels[1].VRange = 0
	st.Channels[1].DisplayVOffset = 0

	scp := &ScpDesc{
		Settings:     st,
		channelCount: 2,
	}

	rect := image.Rect(0, 0, 200, 200)
	img := image.NewRGBA(rect)
	fv := newFvViewer(img, rect, scp)

	bounds := image.Rect(0, 0, 200, 200)
	w := float64(bounds.Dx() - 1)
	h := float64(bounds.Dy())

	// Test center: should be 0 for both X and Y
	curV := fv.calcValuesAt(float32(bounds.Min.X)+float32(w/2.0), float32(bounds.Min.Y)+float32(h/2.0), w, h, bounds, 0)
	assert.InDelta(t, float32(0.0), curV[0], 0.01)
	assert.InDelta(t, float32(0.0), curV[1], 0.01)

	// Even if displayBuffers are completely empty or have different data,
	// calcValuesAt depends ONLY on cursor position values and not on signal buffer values
	scp.displayBuffers = make([][]float32, 2)
	scp.displayBuffers[0] = []float32{999.0, -999.0}
	scp.displayBuffers[1] = []float32{-555.0, 555.0}

	curVWithBuffers := fv.calcValuesAt(float32(bounds.Min.X)+float32(w/2.0), float32(bounds.Min.Y)+float32(h/2.0), w, h, bounds, 0)
	assert.Equal(t, curV[0], curVWithBuffers[0])
	assert.Equal(t, curV[1], curVWithBuffers[1])
}

func TestFvViewer_DrawInspector(t *testing.T) {
	st := settings.NewDefaultSettings()
	st.Channels[0].Enabled = true
	st.Channels[0].FvMode = settings.FvArgument
	st.Channels[1].Enabled = true
	st.Channels[1].FvMode = settings.FvValue

	rect := image.Rect(0, 0, 400, 400)
	img := image.NewRGBA(rect)
	scp := &ScpDesc{
		Settings:            st,
		channelCount:        2,
		fvScopeFullScreen:   img,
		fvScopeSignalScreen: img,
	}

	fv := newFvViewer(img, rect, scp)
	fv.showInspector = true
	fv.mouseX = 200
	fv.mouseY = 200

	bounds := rect
	w := float64(bounds.Dx() - 1)
	h := float64(bounds.Dy())

	// Should not panic and should draw without signal values
	assert.NotPanics(t, func() {
		fv.drawInspector(w, h, bounds)
	})

	// With reference active
	fv.refActive = true
	fv.refX = 150
	fv.refY = 150
	assert.NotPanics(t, func() {
		fv.drawInspector(w, h, bounds)
	})
}

