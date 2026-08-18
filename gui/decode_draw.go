package gui

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// drawDecode renders the decoded bytes onto the given image.
func (scp *ScpDesc) drawDecode(img rasterImage, bounds image.Rectangle, w float64, zeroOffset int, isTimeZoom bool) {
	if !scp.Settings.Decode.Enabled || len(scp.DecodeState.Bytes) == 0 {
		return
	}

	maxScreenTime := scp.maxScreenTime
	var offset float64
	if isTimeZoom {
		maxScreenTime = scp.timeZoomMaxScreenTime
		offset = scp.timeZoomBoxOffset
	}

	pixelsPerSecond := w / maxScreenTime
	timeBaseOffset := offset

	// Font configuration
	f := basicfont.Face7x13
	textColor := color.RGBA{255, 255, 255, 255}
	boxColor := color.RGBA{50, 150, 50, 200}
	errColor := color.RGBA{200, 50, 50, 200}

	// Position just below the horizontal center axis
	yPos := zeroOffset + 40

	for _, result := range scp.DecodeState.Bytes {
		// Convert time to pixels relative to screen left
		xStart := (result.StartTime - timeBaseOffset) * pixelsPerSecond
		xEnd := (result.EndTime - timeBaseOffset) * pixelsPerSecond

		// If it's completely off screen, skip
		if xEnd < 0 || xStart > w {
			continue
		}

		// Clip to screen
		if xStart < 0 {
			xStart = 0
		}
		if xEnd > w {
			xEnd = w
		}

		drawBox := func(val uint16, y int, bgColor color.Color, label string) {
			rect := image.Rect(int(xStart)+bounds.Min.X, y, int(xEnd)+bounds.Min.X, y+20)

			// Draw background box
			draw.Draw(img.(draw.Image), rect, &image.Uniform{bgColor}, image.ZP, draw.Src)
			// Draw border
			borderRect := rect
			borderRect.Max.X = borderRect.Min.X + 1
			draw.Draw(img.(draw.Image), borderRect, &image.Uniform{color.RGBA{255, 255, 255, 255}}, image.ZP, draw.Src)

			// Draw text
			text := fmt.Sprintf("%02X", val)
			if label != "" {
				text = label
			}
			asciiText := ""
			if scp.Settings.Decode.Protocol == "UART" && val >= 32 && val <= 126 {
				asciiText = fmt.Sprintf(" ('%c')", val)
			}

			fullText := text + asciiText
			textWidth := len(fullText) * f.Width

			if float64(rect.Dx()) > float64(textWidth)+4 {
				text = fullText
			} else {
				textWidth = len(text) * f.Width
				if float64(rect.Dx()) <= float64(textWidth)+4 {
					text = ""
				}
			}

			if text != "" {
				// Center text
				x := rect.Min.X + (rect.Dx()-textWidth)/2
				yText := rect.Min.Y + 14 // Baseline

				d := &font.Drawer{
					Dst:  img.(draw.Image),
					Src:  image.NewUniform(textColor),
					Face: f,
					Dot:  fixed.Point26_6{X: fixed.I(x), Y: fixed.I(yText)},
				}
				d.DrawString(text)
			}
		}

		bgColor := boxColor
		if result.Error {
			bgColor = errColor
		}

		drawBox(result.Value, yPos, bgColor, result.Label)
		
		if result.HasValue2 {
			misoColor := color.RGBA{50, 50, 150, 200}
			if result.Error {
				misoColor = errColor
			}
			drawBox(result.Value2, yPos+22, misoColor, "")
		}
	}

	if scp.Settings.Decode.Protocol == "UART" && scp.Settings.Decode.ShowBitstarts {
		orangeColor := color.RGBA{255, 165, 0, 150} // semi-transparent orange
		for _, bit := range scp.DecodeState.Bits {
			// x := ((bit.StartTime+bit.EndTime)/2 - timeBaseOffset) * pixelsPerSecond
			x := (bit.StartTime - timeBaseOffset) * pixelsPerSecond
			if x >= 0 && x <= w {
				pixelX := int(x) + bounds.Min.X
				drawLine(img.(draw.Image), float32(pixelX), float32(bounds.Min.Y), float32(pixelX), float32(bounds.Max.Y), orangeColor)
			}
		}
	}
}
