package gui

import (
	"image"
	"image/color"
	"image/draw"
	"log"
	"math"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

const dscp = 72

var (
	faceCache  = make(map[float64]font.Face)
	faceMutex  sync.Mutex
	parsedFont *opentype.Font
	labelSrc   = &image.Uniform{} // reused across addLabel calls to avoid per-call heap allocation
)

func getFace(size float64) font.Face {
	faceMutex.Lock()
	defer faceMutex.Unlock()

	if f, ok := faceCache[size]; ok {
		return f
	}

	if parsedFont == nil {
		f, err := opentype.Parse(gomono.TTF)
		if err != nil {
			log.Printf("Parse: %v", err)
			panic(9)
		}
		parsedFont = f
	}

	f, err := opentype.NewFace(parsedFont, &opentype.FaceOptions{
		Size:    size,
		DPI:     dscp,
		Hinting: font.HintingNone,
	})
	if err != nil {
		log.Fatalf("NewFace: %v", err)
	}
	faceCache[size] = f
	return f
}

func init() {
	getFace(fontSize)
}

func i26_6ToFloat64(i fixed.Int26_6) float64 {
	return float64(i>>6) + float64(i&0x3f)/float64(1000000)
}
func i26_6ToFloat32(i fixed.Int26_6) float32 {
	return float32(i26_6ToFloat64(i))
}

func (scp *ScpDesc) boundString(s string, size ...float64) (left, top, right, bottom float32) {
	sz := float64(fontSize)
	if len(size) > 0 {
		sz = size[0]
	}
	bound26_6, _ := font.BoundString(getFace(sz), s)
	left = i26_6ToFloat32(bound26_6.Min.X)
	right = i26_6ToFloat32(bound26_6.Max.X)
	top = i26_6ToFloat32(bound26_6.Min.Y)
	bottom = i26_6ToFloat32(bound26_6.Max.Y)
	return
}

func (scp *ScpDesc) addLabel(dst rasterImage, x, y int, label string, textColor color.Color, size ...float64) {
	sz := float64(fontSize)
	if len(size) > 0 {
		sz = size[0]
	}
	labelSrc.C = textColor
	d := font.Drawer{ // Not thread safe
		Dst:  dst,
		Src:  labelSrc,
		Face: getFace(sz),
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
	drawLine(img, xf0, yf, xf1, yf, c)
}

func drawVerticalLine(img draw.Image, xf, yf0, yf1 float32, c color.Color) {
	drawLine(img, xf, yf0, xf, yf1, c)
}

func drawVerticalDashedLine(img draw.Image, xf, yf0, yf1 float32, c color.Color, dashLen, gapLen float32) {
	if yf0 > yf1 {
		yf0, yf1 = yf1, yf0
	}
	for y := yf0; y < yf1; y += dashLen + gapLen {
		yEnd := y + dashLen
		if yEnd > yf1 {
			yEnd = yf1
		}
		drawLine(img, xf, y, xf, yEnd, c)
	}
}

// clipLine clips line segment (x0, y0)-(x1, y1) to bounds using Cohen-Sutherland algorithm.
// Returns true if any part of the segment is inside bounds, false if completely outside.
func clipLine(bounds image.Rectangle, x0, y0, x1, y1 *int) bool {
	xMin := bounds.Min.X
	xMax := bounds.Max.X - 1
	yMin := bounds.Min.Y
	yMax := bounds.Max.Y - 1

	const (
		inside = 0      // 0000
		left   = 1 << 0 // 0001
		right  = 1 << 1 // 0010
		bottom = 1 << 2 // 0100
		top    = 1 << 3 // 1000
	)

	outCode := func(x, y int) int {
		code := inside
		if x < xMin {
			code |= left
		} else if x > xMax {
			code |= right
		}
		if y < yMin {
			code |= top
		} else if y > yMax {
			code |= bottom
		}
		return code
	}

	code0 := outCode(*x0, *y0)
	code1 := outCode(*x1, *y1)

	for {
		if (code0 | code1) == 0 {
			// Both points inside
			return true
		}
		if (code0 & code1) != 0 {
			// Both points share an outside zone (trivial reject)
			return false
		}

		// At least one point is outside
		codeOut := code0
		if codeOut == 0 {
			codeOut = code1
		}

		var x, y int
		if (codeOut & top) != 0 {
			if *y1 == *y0 {
				return false
			}
			x = *x0 + int(math.Round(float64(*x1-*x0)*float64(yMin-*y0)/float64(*y1-*y0)))
			y = yMin
		} else if (codeOut & bottom) != 0 {
			if *y1 == *y0 {
				return false
			}
			x = *x0 + int(math.Round(float64(*x1-*x0)*float64(yMax-*y0)/float64(*y1-*y0)))
			y = yMax
		} else if (codeOut & right) != 0 {
			if *x1 == *x0 {
				return false
			}
			y = *y0 + int(math.Round(float64(*y1-*y0)*float64(xMax-*x0)/float64(*x1-*x0)))
			x = xMax
		} else if (codeOut & left) != 0 {
			if *x1 == *x0 {
				return false
			}
			y = *y0 + int(math.Round(float64(*y1-*y0)*float64(xMin-*x0)/float64(*x1-*x0)))
			x = xMin
		}

		if codeOut == code0 {
			*x0 = x
			*y0 = y
			code0 = outCode(*x0, *y0)
		} else {
			*x1 = x
			*y1 = y
			code1 = outCode(*x1, *y1)
		}
	}
}

func drawLine(img draw.Image, xf0, yf0, xf1, yf1 float32, c color.Color) (err error) {
	if img == nil {
		return nil
	}
	bounds := img.Bounds()
	if bounds.Empty() {
		return nil
	}

	if math.IsNaN(float64(xf0)) || math.IsNaN(float64(yf0)) || math.IsNaN(float64(xf1)) || math.IsNaN(float64(yf1)) {
		return nil
	}
	if math.IsInf(float64(xf0), 0) || math.IsInf(float64(yf0), 0) || math.IsInf(float64(xf1), 0) || math.IsInf(float64(yf1), 0) {
		return nil
	}

	bMinX := float32(bounds.Min.X)
	bMaxX := float32(bounds.Max.X)
	bMinY := float32(bounds.Min.Y)
	bMaxY := float32(bounds.Max.Y)

	// Trivial rejection in float coordinates
	if (xf0 < bMinX && xf1 < bMinX) ||
		(xf0 >= bMaxX && xf1 >= bMaxX) ||
		(yf0 < bMinY && yf1 < bMinY) ||
		(yf0 >= bMaxY && yf1 >= bMaxY) {
		return nil
	}

	const maxCoord = 1e8
	const minCoord = -1e8
	clampF := func(v float32) float64 {
		if v > maxCoord {
			return maxCoord
		}
		if v < minCoord {
			return minCoord
		}
		return float64(v)
	}

	x0 := int(math.Round(clampF(xf0)))
	x1 := int(math.Round(clampF(xf1)))
	y0 := int(math.Round(clampF(yf0)))
	y1 := int(math.Round(clampF(yf1)))

	if !clipLine(bounds, &x0, &y0, &x1, &y1) {
		return nil
	}

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
	maxIter := bounds.Dx() + bounds.Dy() + 10
	n := 0
	for {
		n++
		if n > maxIter {
			break
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
	return nil
}

func drawCircle(img draw.Image, x0, y0, r float32, c color.Color) {
	if img == nil || r <= 0 {
		return
	}
	bounds := img.Bounds()
	if bounds.Empty() {
		return
	}
	bMinX := float32(bounds.Min.X)
	bMaxX := float32(bounds.Max.X)
	bMinY := float32(bounds.Min.Y)
	bMaxY := float32(bounds.Max.Y)
	if x0+r < bMinX || x0-r >= bMaxX || y0+r < bMinY || y0-r >= bMaxY {
		return
	}
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

const (
	triggerArrowLen   = float32(20)
	triggerArrowHalfH = float32(5)
)

func drawTriggerArrow(img draw.Image, tipX, tipY float32, pointLeft bool, c color.Color) {
	var baseX float32
	if pointLeft {
		baseX = tipX + triggerArrowLen
	} else {
		baseX = tipX - triggerArrowLen
	}
	_ = drawLine(img, baseX, tipY-triggerArrowHalfH, baseX, tipY+triggerArrowHalfH, c)
	_ = drawLine(img, baseX, tipY-triggerArrowHalfH, tipX, tipY, c)
	_ = drawLine(img, baseX, tipY+triggerArrowHalfH, tipX, tipY, c)
}
