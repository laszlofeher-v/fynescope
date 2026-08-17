package gui

import (
	"fynescope/settings"
	"image"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/stretchr/testify/assert"
)

func TestSignalViewer_MouseIn(t *testing.T) {
	rect := image.Rect(10, 20, 100, 50)
	img := image.NewRGBA(image.Rect(0, 0, 200, 100))
	scp := &ScpDesc{
		Settings: &settings.PsSettings{},
	}

	sv := newSignalViewer(img, rect, scp, false)

	assert.True(t, sv.mouseIn(15, 25), "Point inside should be true")
	assert.True(t, sv.mouseIn(10, 20), "Point on top-left border should be true")

	// Max bounds are exclusive in image.Rect
	assert.False(t, sv.mouseIn(100, 49), "Point on right border should be false")
	assert.False(t, sv.mouseIn(99, 50), "Point on bottom border should be false")
	assert.False(t, sv.mouseIn(5, 25), "Point outside left should be false")
	assert.False(t, sv.mouseIn(15, 10), "Point outside top should be false")
}

func TestSignalViewer_MouseEvents(t *testing.T) {
	rect := image.Rect(10, 20, 100, 50)
	img := image.NewRGBA(image.Rect(0, 0, 200, 100))
	scp := &ScpDesc{
		Settings: &settings.PsSettings{},
	}

	sv := newSignalViewer(img, rect, scp, false)

	// Initial state
	assert.False(t, sv.showInspector)
	assert.False(t, sv.refActive)

	// Mouse down inside rect with RightMouseButton -> shows inspector
	sv.mouseDown(desktop.MouseButtonSecondary, 0, 15, 25)
	assert.True(t, sv.showInspector)
	assert.False(t, sv.refActive)

	// Mouse up
	sv.mouseUp(desktop.MouseButtonSecondary, 0, 15, 25)
	assert.False(t, sv.showInspector)

	// Mouse down inside rect with RightMouseButton + Shift -> sets reference
	sv.mouseDown(desktop.MouseButtonSecondary, fyne.KeyModifierShift, 15, 25)
	assert.True(t, sv.refActive)
	assert.True(t, sv.refDragging)
	
	sv.mouseUp(desktop.MouseButtonSecondary, fyne.KeyModifierShift, 15, 25)
	assert.False(t, sv.refDragging)
	// refActive remains true after mouseUp until deleted
}

func TestSignalViewer_TypedKey(t *testing.T) {
	rect := image.Rect(10, 20, 100, 50)
	img := image.NewRGBA(image.Rect(0, 0, 200, 100))
	scp := &ScpDesc{
		Settings: &settings.PsSettings{},
	}

	sv := newSignalViewer(img, rect, scp, false)

	// Make sure typedKey doesn't panic on a dummy scope
	sv.typedKey(20, 30, fyne.KeyDown)
	sv.typedKey(20, 30, fyne.KeyUp)
	
	assert.True(t, true)
}
