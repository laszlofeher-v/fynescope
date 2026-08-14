package disp16

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
)

func (h *HexArray) CreateRenderer() fyne.WidgetRenderer {
	h.OffColor = theme.Color(theme.ColorNameBackground)
	var objects []fyne.CanvasObject

	if h.label != nil {
		objects = append(objects, h.label)
	} else {
		objects = append(objects, canvas.NewText("", h.onColor)) // Placeholder for labelIndex
	}

	// For each digit, we have 35 dots
	for i := 0; i < h.numOfDigits; i++ {
		for d := 0; d < 35; d++ {
			rect := canvas.NewRectangle(h.OffColor)
			objects = append(objects, rect)
		}
	}

	r := &hexArrayRenderer{
		hArray:  h,
		objects: objects,
	}
	return r
}

type hexArrayRenderer struct {
	hArray  *HexArray
	objects []fyne.CanvasObject
}

func (r *hexArrayRenderer) Destroy() {
}

func (r *hexArrayRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *hexArrayRenderer) refreshLabel() {
	if r.hArray.label != nil {
		r.objects[labelIndex] = r.hArray.label
		text := r.objects[labelIndex].(*canvas.Text)
		text.Color = r.hArray.onColor
	}
}

func (r *hexArrayRenderer) Refresh() {
	r.hArray.OffColor = theme.Color(theme.ColorNameBackground)
	r.refreshLabel()
	r.refreshNumber()
	canvas.Refresh(r.hArray)
}

func (r *hexArrayRenderer) refreshNumber() {
	for i := 0; i < r.hArray.numOfDigits; i++ {
		val := r.hArray.digits[i].val
		fontData := hexFont[val]
		// slog.Debug("refreshNumber", "r.hArray.digitCursor", r.hArray.digitCursor)
		digitColor := r.hArray.onColor
		if r.hArray.numOfDigits-i-1 == r.hArray.digitCursor {
			digitColor = r.hArray.CursorColor
		}

		for row := 0; row < 7; row++ {
			rowBits := fontData[row]
			for col := 0; col < 5; col++ {
				// 0th col is the leftmost bit, which is bit 4 of rowBits.
				// Wait, the font was defined where LSB is rightmost pixel.
				// So col 0 (left) is bit 4. col 4 (right) is bit 0.
				bit := (rowBits >> (4 - col)) & 1

				idx := numberIndex + i*35 + row*5 + col
				rect := r.objects[idx].(*canvas.Rectangle)

				if bit == 1 {
					rect.FillColor = digitColor
				} else {
					rect.FillColor = r.hArray.OffColor
				}
			}
		}
	}
}

func (r *hexArrayRenderer) Layout(size fyne.Size) {
	labelW := float32(0)
	if r.hArray.label != nil {
		text := r.objects[labelIndex].(*canvas.Text)
		text.Move(fyne.NewPos(0, (size.Height-text.MinSize().Height)/2))
		text.Resize(text.MinSize())
		labelW = text.MinSize().Width + r.hArray.spaceBetweenDigits
	}

	digitWidth := float32(5)*r.hArray.dotSize + float32(4)*r.hArray.dotSpacing

	for i := 0; i < r.hArray.numOfDigits; i++ {
		startX := labelW + float32(i)*(digitWidth+r.hArray.spaceBetweenDigits)
		startY := float32(0)

		r.hArray.digits[i].relPos = fyne.NewPos(startX, startY)

		for row := 0; row < 7; row++ {
			for col := 0; col < 5; col++ {
				x := startX + float32(col)*(r.hArray.dotSize+r.hArray.dotSpacing)
				y := startY + float32(row)*(r.hArray.dotSize+r.hArray.dotSpacing)

				idx := numberIndex + i*35 + row*5 + col
				rect := r.objects[idx].(*canvas.Rectangle)
				rect.Move(fyne.NewPos(x, y))
				rect.Resize(fyne.NewSize(r.hArray.dotSize, r.hArray.dotSize))
			}
		}
	}
}

func (r *hexArrayRenderer) MinSize() fyne.Size {
	return r.hArray.size
}
