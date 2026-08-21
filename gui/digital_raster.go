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
	dr.imgRect = dr.img.Image.Bounds()

	return container.NewMax(dr.img), dr
}

func (dr *digitalRaster) refresh() {
	if !dr.scp.Settings.Digital.Ports[0].Enabled && !dr.scp.Settings.Digital.Ports[1].Enabled {
		return
	}

	if dr.img != nil && dr.img.Image != nil {
		dr.imgRect = dr.img.Image.Bounds()
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

	lineCol := color.RGBA{0, 255, 0, 255}
	
	chIdx := 0
	for port := 0; port < 2; port++ {
		if !dr.scp.Settings.Digital.Ports[port].Enabled {
			continue
		}
		var buf []uint8
		if len(dr.scp.digitalDisplayBuffer) > port {
			buf = dr.scp.digitalDisplayBuffer[port]
		}
		samples := len(buf)
		
		for c := 0; c < 8; c++ {
			yBase := float64(chIdx) * channelHeight
			yHigh := int(yBase + channelHeight*0.2)
			yLow := int(yBase + channelHeight*0.8)
			
			for x := 0; x < w; x++ {
				yPos := yLow
				sampleIdx := 0
				if samples > 0 {
					sampleIdx = int(float64(x) * float64(samples) / float64(w))
					if sampleIdx < samples {
						val := buf[sampleIdx]
						if (val & (1 << c)) != 0 {
							yPos = yHigh
						}
					}
				}
				dr.img.Image.(*image.RGBA).Set(x, yPos, lineCol)
				
				if x > 0 && samples > 0 {
					prevSampleIdx := int(float64(x-1) * float64(samples) / float64(w))
					if sampleIdx < samples && prevSampleIdx < samples {
						prevVal := buf[prevSampleIdx]
						val := buf[sampleIdx]
						prevBit := (prevVal & (1 << c)) != 0
						currBit := (val & (1 << c)) != 0
						if prevBit != currBit {
							for y := yHigh; y <= yLow; y++ {
								dr.img.Image.(*image.RGBA).Set(x, y, lineCol)
							}
						}
					}
				}
			}
			chIdx++
		}
	}
	dr.img.Refresh()
}
