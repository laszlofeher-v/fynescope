package gui

import (
	"fynescope/settings"
	"image"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/stretchr/testify/assert"
)

func TestFtVChannelLabelViewer_MouseIn(t *testing.T) {
	img := image.NewRGBA64(image.Rect(0, 0, 100, 100))
	rect := image.Rect(10, 10, 50, 50)

	scp := &ScpDesc{
		Settings: &settings.PsSettings{},
	}

	viewer := newFtVChannelLabelViewer(img, rect, 0, image.Rect(0, 0, 100, 100), false, scp, false)

	assert.Equal(t, rect, viewer.rect(), "rect() should return the configured rect")
	assert.True(t, viewer.mouseIn(20, 20))
	assert.True(t, viewer.mouseIn(10, 10))
	assert.False(t, viewer.mouseIn(50, 50)) // max bound excluded
	assert.False(t, viewer.mouseIn(5, 5))
	assert.False(t, viewer.mouseIn(60, 60))
}

func TestFtVChannelLabelViewer_Cursor(t *testing.T) {
	img := image.NewRGBA64(image.Rect(0, 0, 100, 100))
	rect := image.Rect(10, 10, 50, 50)

	scp := &ScpDesc{
		Settings: &settings.PsSettings{},
	}

	viewer := newFtVChannelLabelViewer(img, rect, 0, image.Rect(0, 0, 100, 100), false, scp, false)

	cursor, ok := viewer.cursor(20, 20)
	assert.True(t, ok)
	assert.Equal(t, desktop.PointerCursor, cursor)

	cursor, ok = viewer.cursor(0, 0)
	assert.False(t, ok)
	assert.Equal(t, desktop.DefaultCursor, cursor)
}

func TestFtVChannelLabelViewer_MouseDownUp(t *testing.T) {
	img := image.NewRGBA64(image.Rect(0, 0, 100, 100))
	rect := image.Rect(10, 10, 50, 50)

	scp := &ScpDesc{
		Settings: &settings.PsSettings{
			VirtualChannels: []settings.VirtualChSettings{
				{Enabled: true},
			},
		},
		channelViewers:      make([]channelViewerDesc, 1),
		ftScopeSignalScreen: image.NewRGBA64(image.Rect(0, 0, 800, 600)),
	}

	viewer := newFtVChannelLabelViewer(img, rect, 0, image.Rect(0, 0, 100, 100), false, scp, false)

	// Left click inside
	viewer.mouseDown(desktop.MouseButtonPrimary, 0, 20, 20)
	assert.True(t, viewer.selected)

	viewer.mouseUp(desktop.MouseButtonPrimary, 0, 20, 20)
	assert.False(t, viewer.selected)

	// Left click outside
	viewer.mouseDown(desktop.MouseButtonPrimary, 0, 0, 0)
	assert.False(t, viewer.selected)
}

func TestFtVChannelLabelViewer_TypedKey(t *testing.T) {
	img := image.NewRGBA64(image.Rect(0, 0, 100, 100))
	rect := image.Rect(10, 10, 50, 50)

	scp := &ScpDesc{
		Settings: &settings.PsSettings{
			VirtualChannels: []settings.VirtualChSettings{
				{Enabled: true},
			},
		},
		channelViewers:      make([]channelViewerDesc, 1),
		ftScopeSignalScreen: image.NewRGBA64(image.Rect(0, 0, 800, 600)),
	}

	viewer := newFtVChannelLabelViewer(img, rect, 0, image.Rect(0, 0, 100, 100), false, scp, false)

	// Verify it doesn't panic
	viewer.typedKey(20, 20, fyne.KeyDown)
	viewer.typedKey(20, 20, fyne.KeyUp)

	assert.True(t, true)
}
