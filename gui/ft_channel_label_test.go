package gui

import (
	"fynescope/settings"
	"image"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/stretchr/testify/assert"
)

func TestFtChannelLabelViewer_MouseIn(t *testing.T) {
	img := image.NewRGBA64(image.Rect(0, 0, 100, 100))
	rect := image.Rect(10, 10, 50, 50)

	scp := &ScpDesc{
		Settings: &settings.PsSettings{},
	}

	viewer := newFtChannelLabelViewer(img, rect, 0, image.Rect(0, 0, 100, 100), scp, false)

	assert.Equal(t, rect, viewer.rect(), "rect() should return the configured rect")
	assert.True(t, viewer.mouseIn(20, 20))
	assert.True(t, viewer.mouseIn(10, 10))
	assert.False(t, viewer.mouseIn(50, 50)) // rect bound excludes max bounds
	assert.False(t, viewer.mouseIn(5, 5))
	assert.False(t, viewer.mouseIn(60, 60))
}

func TestFtChannelLabelViewer_Cursor(t *testing.T) {
	img := image.NewRGBA64(image.Rect(0, 0, 100, 100))
	rect := image.Rect(10, 10, 50, 50)

	scp := &ScpDesc{
		Settings: &settings.PsSettings{},
	}

	viewer := newFtChannelLabelViewer(img, rect, 0, image.Rect(0, 0, 100, 100), scp, false)

	cursor, ok := viewer.cursor(20, 20)
	assert.True(t, ok)
	assert.Equal(t, desktop.PointerCursor, cursor)

	cursor, ok = viewer.cursor(0, 0)
	assert.False(t, ok)
	assert.Equal(t, desktop.DefaultCursor, cursor)
}

func TestFtChannelLabelViewer_MouseDownUp(t *testing.T) {
	img := image.NewRGBA64(image.Rect(0, 0, 100, 100))
	rect := image.Rect(10, 10, 50, 50)

	scp := &ScpDesc{
		Settings: &settings.PsSettings{
			Channels: []settings.ChSettings{
				{Enabled: true},
			},
		},
		channelViewers:      make([]channelViewerDesc, 1),
		ftScopeSignalScreen: image.NewRGBA64(image.Rect(0, 0, 800, 600)),
	}

	viewer := newFtChannelLabelViewer(img, rect, 0, image.Rect(0, 0, 100, 100), scp, false)

	// Left click inside
	viewer.mouseDown(desktop.MouseButtonPrimary, 0, 20, 20)
	assert.True(t, viewer.selected)

	viewer.mouseUp(desktop.MouseButtonPrimary, 0, 20, 20)
	assert.False(t, viewer.selected)

	// Left click outside
	viewer.mouseDown(desktop.MouseButtonPrimary, 0, 0, 0)
	assert.False(t, viewer.selected)
}

func TestFtChannelLabelViewer_TypedKey(t *testing.T) {
	img := image.NewRGBA64(image.Rect(0, 0, 100, 100))
	rect := image.Rect(10, 10, 50, 50)

	scp := &ScpDesc{
		Settings: &settings.PsSettings{
			Channels: []settings.ChSettings{
				{Enabled: true},
			},
		},
		channelViewers:      make([]channelViewerDesc, 1),
		ftScopeSignalScreen: image.NewRGBA64(image.Rect(0, 0, 800, 600)),
	}

	viewer := newFtChannelLabelViewer(img, rect, 0, image.Rect(0, 0, 100, 100), scp, false)

	// Verify it doesn't panic
	viewer.typedKey(20, 20, fyne.KeyDown)
	viewer.typedKey(20, 20, fyne.KeyUp)

	assert.True(t, true)
}
