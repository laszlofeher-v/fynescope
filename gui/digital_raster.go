package gui

import (
	"image"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"golang.org/x/image/draw"
)

type digitalRaster struct {
	scp     *ScpDesc
	img     *canvas.Image
	imgRect image.Rectangle
	window  fyne.Window
}

func (scp *ScpDesc) newDigitalRaster(window fyne.Window) (*fyne.Container, *digitalRaster) {
	dr := &digitalRaster{
		scp:    scp,
		window: window,
	}

	dr.img = canvas.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, 1024, 200)))
	dr.img.FillMode = canvas.ImageFillStretch

	return container.NewMax(dr.img), dr
}

func (dr *digitalRaster) refresh() {
	if !dr.scp.Settings.Digital.Ports[0].Enabled && !dr.scp.Settings.Digital.Ports[1].Enabled {
		return
	}

	w := dr.imgRect.Dx()
	h := dr.imgRect.Dy()
	if w <= 0 || h <= 0 {
		return
	}

	bgCol := theme.BackgroundColor()
	draw.Draw(dr.img.Image.(*image.RGBA), dr.imgRect, &image.Uniform{bgCol}, image.Point{}, draw.Src)

	// We have 16 digital channels maximum
	activeChannels := 0
	if dr.scp.Settings.Digital.Ports[0].Enabled {
		activeChannels += 8
	}
	if dr.scp.Settings.Digital.Ports[1].Enabled {
		activeChannels += 8
	}

	if activeChannels == 0 {
		dr.img.Refresh()
		return
	}

	channelHeight := float64(h) / float64(activeChannels)

	// A very basic digital drawer (only drawing horizontal lines for logic 0 to start)
	lineCol := color.RGBA{0, 255, 0, 255}
	for i := 0; i < activeChannels; i++ {
		yPos := int(float64(i)*channelHeight + channelHeight/2)
		for x := 0; x < w; x++ {
			dr.img.Image.(*image.RGBA).Set(x, yPos, lineCol)
		}
	}
	dr.img.Refresh()
}
