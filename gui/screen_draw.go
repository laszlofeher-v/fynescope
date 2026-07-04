package gui

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log"
	"log/slog"
	"math"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

const dscp = 72

var (
	face      font.Face
	labelSrc  = &image.Uniform{} // reused across addLabel calls to avoid per-call heap allocation
)

func init() {
	f, err := opentype.Parse(gomono.TTF)
	if err != nil {
		log.Printf("Parse: %v", err)
		panic(9)
	}
	face, err = opentype.NewFace(f, &opentype.FaceOptions{
		Size:    fontSize,
		DPI:     dscp,
		Hinting: font.HintingNone,
	})
	if err != nil {
		log.Fatalf("NewFace: %v", err)
	}
}
func i26_6ToFloat64(i fixed.Int26_6) float64 {
	return float64(i>>6) + float64(i&0x3f)/float64(1000000)
}
func i26_6ToFloat32(i fixed.Int26_6) float32 {
	return float32(i26_6ToFloat64(i))
}

func (scp *ScpDesc) boundString(s string) (left, top, right, bottom float32) {
	bound26_6, _ := font.BoundString(face, s)
	left = i26_6ToFloat32(bound26_6.Min.X)
	right = i26_6ToFloat32(bound26_6.Max.X)
	top = i26_6ToFloat32(bound26_6.Min.Y)
	bottom = i26_6ToFloat32(bound26_6.Max.Y)
	return
}

func (scp *ScpDesc) addLabel(dst rasterImage, x, y int, label string, textColor color.Color) {
	labelSrc.C = textColor
	d := font.Drawer{ // Not thread safe
		Dst:  dst,
		Src:  labelSrc,
		Face: face,
		Dot:  fixed.P(x, y),
	}
	d.DrawString(label)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func drawHorizontalLine(img draw.Image, xf0, xf1, yf float32, c color.Color) {
}
func drawVerticalLine(img draw.Image, xf, yf0, yf1 float32, c color.Color) {
}
func drawLine(img draw.Image, xf0, yf0, xf1, yf1 float32, c color.Color) (err error) {

	x0 := int(math.Round(float64(xf0)))
	x1 := int(math.Round(float64(xf1)))
	y0 := int(math.Round(float64(yf0)))
	y1 := int(math.Round(float64(yf1)))
	// bresenham.DrawLine(img, x0, y0, x1, y1, c)
	dx := abs(x1 - x0)
	sx := -1
	if x0 < x1 {
		sx = 1
	}
	dy := -abs(y1 - y0)
	sy := -1
	if y0 < y1 {
		sy = 1
	}
	error := dx + dy
	n := 0
	for {
		n++
		if n > 10000 {
			slog.Debug("draw line", "xf0", xf0, "yf0", yf0, "xf1", xf1, "yf1", yf1)
			slog.Debug("draw line", "x0", x0, "y0", y0, "x1", x1, "y1", y1, "error", error)
			err = fmt.Errorf("draw line >10000")
			return
		}
		img.Set(x0, y0, c)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * error
		if e2 >= dy {
			if x0 == x1 {
				break
			}
			error = error + dy
			x0 = x0 + sx
		}
		if e2 <= dx {
			if y0 == y1 {
				break
			}
			error = error + dx
			y0 = y0 + sy
		}
	}
	return
}

func drawCircle(img draw.Image, x0, y0, r float32, c color.Color) {
	r34 := 3 * r / 4
	for r > r34 {
		x, y, dx, dy := (r - 1), float32(0), float32(1), float32(1)
		err := dx - (r * 2)
		for x >= y {
			x0px := int(math.Round(float64(x0 + x)))
			y0py := int(math.Round(float64(y0 + y)))
			x0py := int(math.Round(float64(x0 + y)))
			y0px := int(math.Round(float64(y0 + x)))
			x0my := int(math.Round(float64(x0 - y)))
			x0mx := int(math.Round(float64(x0 - x)))
			y0my := int(math.Round(float64(y0 - y)))
			y0mx := int(math.Round(float64(y0 - x)))
			img.Set(x0px, y0py, c)
			img.Set(x0py, y0px, c)
			img.Set(x0my, y0px, c)
			img.Set(x0mx, y0py, c)
			img.Set(x0mx, y0my, c)
			img.Set(x0my, y0mx, c)
			img.Set(x0py, y0mx, c)
			img.Set(x0px, y0my, c)
			if err <= 0 {
				y++
				err += dy
				dy += 2
			}
			if err > 0 {
				x--
				dx += 2
				err += dx - (r * 2)
			}
		}
		r--
	}
}
