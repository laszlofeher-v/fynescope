package gui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/test"
	"github.com/stretchr/testify/assert"
)

func TestScreenRaster_MouseIn(t *testing.T) {
	scr := &screenRaster{
		Window: test.NewWindow(nil),
	}

	scr.MouseIn(&desktop.MouseEvent{})
	assert.True(t, scr.mouseIn, "MouseIn should set mouseIn to true")
}

func TestScreenRaster_MouseOut(t *testing.T) {
	scr := &screenRaster{mouseIn: true}

	scr.MouseOut()
	assert.False(t, scr.mouseIn, "MouseOut should set mouseIn to false")
}

func TestScreenRaster_MouseMoved(t *testing.T) {
	scp := &ScpDesc{}
	scr := &screenRaster{
		scp: scp,
	}

	event := &desktop.MouseEvent{
		PointEvent: fyne.PointEvent{
			Position: fyne.NewPos(10, 20),
		},
	}
	
	// mock mouseIn=true
	scr.mouseIn = true
	scr.MouseMoved(event)
	
	assert.Equal(t, float32(10), scr.mouseX)
	assert.Equal(t, float32(20), scr.mouseY)
}


