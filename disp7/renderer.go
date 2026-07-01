package disp7

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
)

// disp7Array[0] is the least significant digit
// if signed then disp7Array[len(disp7Array)-1] displays the sign
type disp7ArrayRenderer struct {
	d7array *DigitArray
	objects []fyne.CanvasObject
}

func (d7rend *disp7ArrayRenderer) Destroy() {
}

func (d7rend *disp7ArrayRenderer) Objects() []fyne.CanvasObject {
	return d7rend.objects
}

func (d7rend *disp7ArrayRenderer) refreshSign() {
	soMinus := d7rend.objects[signIndexMinus].(*canvas.Line)
	soPlus := d7rend.objects[signIndexPlus].(*canvas.Line)
	switch {
	case d7rend.d7array.Value == 0:
		soMinus.StrokeColor = d7rend.d7array.OffColor
		soPlus.StrokeColor = d7rend.d7array.OffColor
	case d7rend.d7array.Value > 0:
		soMinus.StrokeColor = d7rend.d7array.onColor
		soPlus.StrokeColor = d7rend.d7array.onColor
	case d7rend.d7array.Value < 0:
		soMinus.StrokeColor = d7rend.d7array.onColor
		soPlus.StrokeColor = d7rend.d7array.OffColor
	}
}

func (d7rend *disp7ArrayRenderer) refreshUnit() {
	d7rend.objects[unitIndex] = d7rend.d7array.unit
	text := d7rend.objects[unitIndex].(*canvas.Text)
	text.Color = d7rend.d7array.onColor
}

func (d7rend *disp7ArrayRenderer) refreshLabel() {
	d7rend.objects[labelIndex] = d7rend.d7array.label
	text := d7rend.objects[labelIndex].(*canvas.Text)
	text.Color = d7rend.d7array.onColor
}

func (d7rend *disp7ArrayRenderer) Refresh() {
	d7rend.d7array.OffColor = theme.Color(theme.ColorNameBackground)
	segmentWidth := d7rend.d7array.segmentWidth
	if d7rend.d7array.unit != nil {
		d7rend.refreshUnit()
	}
	if d7rend.d7array.label != nil {
		d7rend.refreshLabel()
	}
	if d7rend.d7array.signed == Signed {
		d7rend.refreshSign()
	}
	d7rend.refreshNumber(segmentWidth)
}

func (d7rend *disp7ArrayRenderer) refreshNumber(segmentWidth float32) {
	trailing := d7rend.d7array.trailingZeroesOn
	showDigitFrom := len(d7rend.d7array.digits) - 1
	if !trailing {
		showDigitFrom = d7rend.d7array.dpPos
		for i := len(d7rend.d7array.digits) - 1; i >= d7rend.d7array.dpPos; i-- {
			if d7rend.d7array.digits[i].val > 0 {
				showDigitFrom = i
				break
			}
		}
	}
	for i := numberIndex; i < len(d7rend.objects); i++ {
		digitIndex := (i - numberIndex) / (numOfSegments + 1) // 7 segments + 1 dp
		val := d7rend.d7array.digits[digitIndex].val
		var color color.Color
		if showDigitFrom >= digitIndex {
			color = d7rend.d7array.onColor
		} else {
			color = d7rend.d7array.OffColor
		}
		switch so := d7rend.objects[i].(type) {
		case *canvas.Line:
			so.StrokeWidth = segmentWidth / 2
			if segments[val][(i-numberIndex)%8] {
				if digitIndex == d7rend.d7array.digitCursor {
					so.StrokeColor = d7rend.d7array.CursorColor
				} else {
					so.StrokeColor = color
				}
			} else {
				so.StrokeColor = d7rend.d7array.OffColor
			}
			fyne.Do(so.Refresh)
		case *canvas.Circle:
			if d7rend.d7array.dpPos >= 0 {
				so.StrokeWidth = segmentWidth / 2
				if digitIndex > 0 && digitIndex == d7rend.d7array.dpPos {
					so.StrokeColor = d7rend.d7array.onColor
					so.FillColor = d7rend.d7array.onColor
				} else {
					so.StrokeColor = d7rend.d7array.OffColor
					so.FillColor = d7rend.d7array.OffColor
				}
			}
			fyne.Do(so.Refresh)
		default:
		}
	}
}

func placeLine(line *canvas.Line, p1, p2 fyne.Position) {
	line.Position1, line.Position2 = p1, p2
}

func (d7rend *disp7ArrayRenderer) Layout(size fyne.Size) {
	segmentThickness := d7rend.d7array.segmentWidth
	halfSegmentThickness := segmentThickness / 2
	doubleVerticalSegmentLength := (d7rend.d7array.size.Height -
		segmentThickness - d7rend.d7array.cursorVSpace)
	verticalSegmentLength := doubleVerticalSegmentLength / 2
	horizontalSegmentLength := d7rend.d7array.digitWidth - segmentThickness
	offset := d7rend.d7array.skew
	doubleOffset := 2 * offset
	d7rend.d7array.digitSetRelPos()
	for digitIndex := range d7rend.d7array.digits {
		pos := d7rend.d7array.digits[digitIndex].relPos
		start := numberIndex + digitIndex*(numOfSegments+1)
		switch d7rend.objects[start].(type) {
		case *canvas.Line:
			placeLine(d7rend.objects[start].(*canvas.Line),
				pos.Add(fyne.NewPos(halfSegmentThickness+doubleOffset, 0)),
				pos.Add(fyne.NewPos(horizontalSegmentLength+
					halfSegmentThickness+doubleOffset, 0))) // segment a
			placeLine(d7rend.objects[1+start].(*canvas.Line),
				pos.Add(fyne.NewPos(horizontalSegmentLength+
					halfSegmentThickness+doubleOffset, 0)),
				pos.Add(fyne.NewPos(horizontalSegmentLength+
					halfSegmentThickness+offset, verticalSegmentLength))) // segment b
			placeLine(d7rend.objects[2+start].(*canvas.Line),
				pos.Add(fyne.NewPos(horizontalSegmentLength+
					halfSegmentThickness+offset, verticalSegmentLength)),
				pos.Add(fyne.NewPos(horizontalSegmentLength+
					halfSegmentThickness, doubleVerticalSegmentLength))) // segment c
			placeLine(d7rend.objects[3+start].(*canvas.Line),
				pos.Add(fyne.NewPos(halfSegmentThickness, doubleVerticalSegmentLength)),
				pos.Add(fyne.NewPos(horizontalSegmentLength+
					halfSegmentThickness, doubleVerticalSegmentLength))) // segment d
			placeLine(d7rend.objects[4+start].(*canvas.Line),
				pos.Add(fyne.NewPos(halfSegmentThickness+offset, verticalSegmentLength)),
				pos.Add(fyne.NewPos(halfSegmentThickness, doubleVerticalSegmentLength))) // segment e
			placeLine(d7rend.objects[5+start].(*canvas.Line),
				pos.Add(fyne.NewPos(halfSegmentThickness+doubleOffset, 0)),
				pos.Add(fyne.NewPos(halfSegmentThickness+offset, verticalSegmentLength))) // segment f
			placeLine(d7rend.objects[6+start].(*canvas.Line),
				pos.Add(fyne.NewPos(halfSegmentThickness+offset, verticalSegmentLength)),
				pos.Add(fyne.NewPos(horizontalSegmentLength+
					halfSegmentThickness+offset, verticalSegmentLength))) // segment g
			d7rend.objects[7+start].(*canvas.Circle).Position1 =
				pos.Add(fyne.NewPos(d7rend.d7array.digitWidth+halfSegmentThickness,
					doubleVerticalSegmentLength-halfSegmentThickness))
			posd := fyne.NewPos(halfSegmentThickness*2.5, halfSegmentThickness*2.5)
			d7rend.objects[7+start].(*canvas.Circle).Position2 =
				d7rend.objects[7+start].(*canvas.Circle).Position1.Add(posd)
		case *canvas.Text:
			panic("Unexpedted text")
		}
	}
	pos := d7rend.d7array.digits[len(d7rend.d7array.digits)-1].relPos
	pos = pos.SubtractXY(d7rend.d7array.digitWidth, 0)
	placeLine(d7rend.objects[signIndexMinus].(*canvas.Line),
		pos.Add(fyne.NewPos(halfSegmentThickness+offset, verticalSegmentLength)),
		pos.Add(fyne.NewPos(horizontalSegmentLength+
			halfSegmentThickness+offset, verticalSegmentLength)))
	placeLine(d7rend.objects[signIndexPlus].(*canvas.Line),
		pos.Add(fyne.NewPos(horizontalSegmentLength/2+
			halfSegmentThickness/2+doubleOffset, verticalSegmentLength/2)),
		pos.Add(fyne.NewPos(horizontalSegmentLength/2+
			halfSegmentThickness/2+offset, 3*verticalSegmentLength/2)))
	if d7rend.d7array.unit != nil {
		pos = d7rend.d7array.digits[0].relPos
		pos = pos.AddXY(d7rend.d7array.digitWidth, 0)
		d7rend.d7array.unit.Move(pos)
	}
	if d7rend.d7array.label != nil {
		pos = fyne.Position{X: 0, Y: pos.Y}
		pos = pos.AddXY(d7rend.d7array.digitWidth-
			d7rend.d7array.spaceBetweenDigits-d7rend.d7array.skew, 0)
		d7rend.d7array.label.Move(pos)
	}
}

func (d7rend *disp7ArrayRenderer) MinSize() fyne.Size {
	return fyne.NewSize(
		d7rend.d7array.size.Width,
		d7rend.d7array.size.Height)
}

func (d7 *DigitArray) CreateRenderer() fyne.WidgetRenderer {
	d7.numberOfFullDigits = len(d7.digits)
	numberOfSegmentObjects := d7.numberOfFullDigits*8 + 2 /*sign*/ +
		1 /*label*/ + 1 /*unit*/
	objects := make([]fyne.CanvasObject, numberOfSegmentObjects)
	r := &disp7ArrayRenderer{
		d7array: d7,
		objects: objects,
	}
	col := d7.OffColor
	for i := numberIndex; i < numberOfSegmentObjects; i++ {
		if (i-numberIndex+1)%8 != 0 {
			objects[i] = canvas.NewLine(col)
		} else {
			objects[i] = canvas.NewCircle(col)
		}
	}
	objects[signIndexMinus] = canvas.NewLine(col)
	objects[signIndexPlus] = canvas.NewLine(col)
	r.Refresh()
	return r
}
