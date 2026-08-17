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
