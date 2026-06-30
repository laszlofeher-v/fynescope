package disp7

import (
	"fmt"
	"image/color"
	"log/slog"
	"math"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type (
	trailingZeroesType int
	signType           int
	accessModeType     int
)

const (
	ReaOnly accessModeType = iota
	ReadWrite
)
const (
	TrailingZeroes trailingZeroesType = iota
	NoTrailingZeroes
)

const (
	Signed signType = iota
	SignedHidden
	UnSigned
)
const (
	DefaultDigitWidth   = float32(17)
	DeafultDigitHeight  = float32(36)
	DefaultVCursorSpace = float32(8)
	DefaultSkew         = float32(2)
	numOfSegments       = 7
	DefaultCursorColor  = "gray"
	segmentWidthRatio   = 0.25
	digitCursorOut      = -1
	labelIndex          = 0
	signIndexMinus      = 1
	signIndexPlus       = 2
	unitIndex           = 3 // sign has 2 objects
	numberIndex         = 4
	spaceMultiplier     = 1.5
)

type (
	disp7Desc struct {
		val int
		//0 -> a ... 6 ->g
		segmentStates []bool
		// dp            bool
		relPos fyne.Position
	}
	DigitArray struct {
		widget.BaseWidget
		onColor            color.Color
		OffColor           color.Color
		CursorColor        color.Color
		Readonly           bool
		size               fyne.Size
		digitWidth         float32
		cursorVSpace       float32
		digits             []disp7Desc
		Value              int
		dpPos              int
		maxValue           int
		minValue           int
		OnChanged          func(v float64)
		Window             fyne.Window
		mousePos           fyne.Position
		digitCursor        int
		spaceBetweenDigits float32
		skew               float32
		segmentWidth       float32
		signed             signType
		trailingZeroesOn   bool
		label              *canvas.Text
		unit               *canvas.Text
		numberOfFullDigits int
		lock               sync.Mutex
	}
	// disp7Array[0] is the least significant digit
	// if signed then disp7Array[len(disp7Array)-1] displays the sign
	disp7ArrayRenderer struct {
		d7array *DigitArray
		objects []fyne.CanvasObject
	}
)

const (
	y = true
	n = false
)

var (
	// red      = color.RGBA{255, 0, 0, 255}
	// green    = color.RGBA{0, 255, 0, 255}
	segments = [][]bool{
		{y, y, y, y, y, y, n}, // 0
		//{n, n, n, n, y, y, n}, // 1 alternative
		{n, y, y, n, n, n, n}, // 1
		{y, y, n, y, y, n, y}, // 2
		{y, y, y, y, n, n, y}, // 3
		{n, y, y, n, n, y, y}, // 4
		{y, n, y, y, n, y, y}, // 5
		{y, n, y, y, y, y, y}, // 6
		//{n, n, y, y, y, y, y}, // 6 alternative
		{y, y, y, n, n, n, n}, // 7
		//{y, y, y, n, n, y, n}, // 7 alternative
		{y, y, y, y, y, y, y}, // 8
		{y, y, y, y, n, y, y}, // 9
		//{y, y, y, n, n, y, y}, // 9 alternative
	}
)

func NewCustomDisp7Array(numOfDigits int, numOfFractionDigits, maxValue,
	minValue int, signed signType, trailingZeroes trailingZeroesType, w fyne.Window, onColor color.Color,
	readOnly accessModeType, digitWidth, digitHeight, skew, cursorVSpace float32,
	label, unit string) (disp *DigitArray, err error) {
	if numOfDigits <= 0 {
		err = fmt.Errorf("numOfDigits is 0")
		return
	}
	if signed == UnSigned && minValue < 0 {
		err = fmt.Errorf("minValue < 0 and unsigned")
		return
	}
	disp = &DigitArray{digits: make([]disp7Desc, numOfDigits)}
	for i := range disp.digits {
		disp.digits[i] = disp7Desc{}
	}
	disp.signed = signed
	disp.trailingZeroesOn = trailingZeroes == TrailingZeroes
	disp.dpPos = numOfFractionDigits
	disp.ExtendBaseWidget(disp)
	disp.spaceBetweenDigits = digitWidth / 3
	disp.size.Height = digitHeight
	disp.digitWidth = digitWidth
	disp.skew = skew
	disp.segmentWidth = segmentWidthRatio * disp.digitWidth
	disp.unit = canvas.NewText(unit, onColor)
	disp.unit.TextStyle = fyne.TextStyle{Monospace: true}
	disp.unit.TextSize = digitHeight / 2
	disp.label = canvas.NewText(label, onColor)
	disp.label.TextStyle = fyne.TextStyle{Monospace: true}
	disp.label.TextSize = digitHeight / 2
	slog.Debug("sizes", "unit", unit, "W", disp.unit.MinSize().Width)
	dpSpace := float32(0)
	if numOfFractionDigits > 0 {
		dpSpace = spaceMultiplier * (disp.spaceBetweenDigits + disp.skew)
	}
	signSpace := float32(0)
	if signed != UnSigned {
		signSpace = disp.digitWidth + disp.spaceBetweenDigits + disp.skew
	}
	disp.size.Width = float32(numOfDigits)*disp.digitWidth +
		float32(numOfDigits-1)*(disp.spaceBetweenDigits+disp.skew) +
		disp.label.MinSize().Width + disp.spaceBetweenDigits + disp.skew +
		disp.unit.MinSize().Width + disp.spaceBetweenDigits + disp.skew +
		// disp.segmentWidth/2 + disp.skew +
		spaceMultiplier*(disp.spaceBetweenDigits+disp.skew)*
			(float32(numOfDigits /*-numOfFractionDigits*/)/3) + dpSpace + signSpace
	disp.cursorVSpace = cursorVSpace
	disp.onColor = onColor
	disp.CursorColor = theme.PrimaryColorNamed(DefaultCursorColor)
	disp.maxValue = maxValue
	disp.minValue = minValue
	disp.Window = w
	disp.digitCursor = digitCursorOut
	disp.Readonly = readOnly == ReaOnly
	return
}

func NewDisp7Array(numOfDigits int, numOfFractionDigits, maxValue, minValue int,
	signed signType, w fyne.Window, onColor color.Color, readOnly accessModeType) (disp *DigitArray, err error) {
	disp, err = NewCustomDisp7Array(numOfDigits, numOfFractionDigits, maxValue,
		minValue, signed, TrailingZeroes, w, onColor, readOnly, DefaultDigitWidth,
		DeafultDigitHeight, DefaultSkew, DefaultVCursorSpace, "", "")
	return
}

func (d7 *DigitArray) SetNumOfFractionDigits(numOfFractionDigits int) {
	defer d7.lock.Unlock()
	d7.lock.Lock()
	d7.dpPos = numOfFractionDigits
}

func (d7 *DigitArray) SetMinMax(minValue, maxValue int) {
	defer d7.lock.Unlock()
	d7.lock.Lock()
	d7.maxValue = maxValue
	d7.minValue = minValue
	if d7.Value < d7.minValue {
		d7.silentSetValue(d7.minValue)
		d7.Refresh()
	} else if d7.Value > d7.maxValue {
		d7.silentSetValue(d7.maxValue)
		d7.Refresh()
	}
}

func (d7 *DigitArray) SetOncolor(col color.Color) {
	defer d7.lock.Unlock()
	d7.lock.Lock()
	d7.onColor = col
	d7.Refresh()
}

func (d7 *DigitArray) DpPos() int {
	defer d7.lock.Unlock()
	d7.lock.Lock()
	return d7.dpPos
}

var _ fyne.Disableable = (*DigitArray)(nil)
var _ fyne.Draggable = (*DigitArray)(nil)
var _ fyne.Focusable = (*DigitArray)(nil)
var _ fyne.Tappable = (*DigitArray)(nil)
var _ fyne.Widget = (*DigitArray)(nil)
var _ desktop.Mouseable = (*DigitArray)(nil)
var _ desktop.Keyable = (*DigitArray)(nil)
var _ fyne.Tabbable = (*DigitArray)(nil)

func (d7 *DigitArray) AcceptsTab() bool {
	return false
}

func (d7 *DigitArray) Disable() {
}
func (d7 *DigitArray) Enable() {
}
func (d7 *DigitArray) Disabled() bool {
	return false
}
func (d7 *DigitArray) Tapped(event *fyne.PointEvent) {
	d7.lock.Lock()
	if d7.Readonly {
		d7.lock.Unlock()
		return
	}
	d7.clearDigitAt(event.Position.X)
	d7.mousePos = event.Position
	focus := d7.setDigitCursor(event.Position.X)
	d7.lock.Unlock()
	
	if focus {
		d7.Window.Canvas().Focus(d7)
	}
	d7.Refresh()
}

func (d7 *DigitArray) Cursor() desktop.Cursor {
	defer d7.lock.Unlock()
	d7.lock.Lock()
	if d7.Readonly || d7.digitCursor == digitCursorOut {
		return desktop.DefaultCursor
	}
	return desktop.PointerCursor
}
func (d7 *DigitArray) cursorLeft() {
	if d7.digitCursor < len(d7.digits)-1 {
		d7.digitCursor++
	}
}
func (d7 *DigitArray) cursorRight() {
	if d7.digitCursor > 0 {
		d7.digitCursor--
	}
}
func (d7 *DigitArray) TypedKey(k *fyne.KeyEvent) {
	defer d7.lock.Unlock()
	d7.lock.Lock()
	if d7.Readonly {
		return
	}
	d := int(math.Round(math.Pow(10, float64(d7.digitCursor))))
	switch k.Name {
	case fyne.KeyUp:
		d7.setValue(d7.Value + d)
	case fyne.KeyDown:
		d7.down(d7.digitCursor)
	case fyne.KeyLeft:
		d7.cursorLeft()
	case fyne.KeyRight:
		d7.cursorRight()
	case fyne.KeyDelete:
		d7.setDigitAtDigitCursor(0)
		d7.cursorRight()
	case fyne.KeyBackspace:
		d7.setDigitAtDigitCursor(0)
		d7.cursorLeft()
	case fyne.Key0:
		d7.setDigitAtDigitCursor(0)
		d7.cursorRight()
	case fyne.Key1:
		d7.setDigitAtDigitCursor(1)
		d7.cursorRight()
	case fyne.Key2:
		d7.setDigitAtDigitCursor(2)
		d7.cursorRight()
	case fyne.Key3:
		d7.setDigitAtDigitCursor(3)
		d7.cursorRight()
	case fyne.Key4:
		d7.setDigitAtDigitCursor(4)
		d7.cursorRight()
	case fyne.Key5:
		d7.setDigitAtDigitCursor(5)
		d7.cursorRight()
	case fyne.Key6:
		d7.setDigitAtDigitCursor(6)
		d7.cursorRight()
	case fyne.Key7:
		d7.setDigitAtDigitCursor(7)
		d7.cursorRight()
	case fyne.Key8:
		d7.setDigitAtDigitCursor(8)
		d7.cursorRight()
	case fyne.Key9:
		d7.setDigitAtDigitCursor(9)
		d7.cursorRight()
	default:
	}
	d7.Refresh()
}
func (d7 *DigitArray) TypedRune(r rune) {
}
func (d7 *DigitArray) FocusGained() {
	defer d7.lock.Unlock()
	d7.lock.Lock()
	if d7.Readonly {
		return
	}
	d7.digitCursor = len(d7.digits) - 1 // default when TAB pressed
	d7.Refresh()
}

func (d7 *DigitArray) FocusLost() {
	defer d7.lock.Unlock()
	d7.lock.Lock()
	if d7.Readonly {
		return
	}
	d7.digitCursor = digitCursorOut
	d7.Refresh()
}

func (d7 *DigitArray) TypedShortcut(r fyne.Shortcut) {
}
func (d7 *DigitArray) KeyUp(k *fyne.KeyEvent) {
}
func (d7 *DigitArray) KeyDown(k *fyne.KeyEvent) {
}
func (d7 *DigitArray) SilentSetValue(v int) {
	defer d7.lock.Unlock()
	d7.lock.Lock()
	d7.silentSetValue(v)
}
func (d7 *DigitArray) silentSetValue(v int) {
	switch {
	case v > d7.maxValue:
		v = d7.maxValue
	case v < d7.minValue:
		v = d7.minValue
	case d7.Value == d7.minValue && v > 10*d7.minValue && d7.signed == UnSigned:
		v = v - d7.minValue
	}
	d7.Value = v
	if v < 0 {
		v = -v
	}
	for pos := 0; pos < len(d7.digits); pos++ {
		d7.digits[pos].val = v % 10
		d7.digits[pos].segmentStates = segments[d7.digits[pos].val]
		v = v / 10
	}
}

func (d7 *DigitArray) SetValue(v int) {
	defer d7.lock.Unlock()
	d7.lock.Lock()
	d7.setValue(v)
}

func (d7 *DigitArray) setValue(v int) {
	d7.silentSetValue(v)
	if d7.OnChanged != nil {
		d7.OnChanged(float64(d7.Value))
	}
}

func (d7 *DigitArray) SetUnit(unitName string) {
	defer d7.lock.Unlock()
	d7.lock.Lock()
	d7.unit = canvas.NewText(unitName, d7.onColor)
	d7.unit.TextStyle = fyne.TextStyle{Monospace: true}
	d7.unit.TextSize = d7.size.Height / 2
	d7.Refresh()
}
func (d7 *DigitArray) SetLabel(label string) {
	defer d7.lock.Unlock()
	d7.lock.Lock()
	slog.Debug("set label", "label", label)
	d7.label = canvas.NewText(label, d7.onColor)
	d7.label.TextStyle = fyne.TextStyle{Monospace: true}
	d7.label.TextSize = d7.size.Height / 2
	d7.Refresh()
}

func (d7 *DigitArray) SilentSetFloatValue(v float64, dpPos int) {
	defer d7.lock.Unlock()
	d7.lock.Lock()
	d7.silentSetFloatValue(v, dpPos)
}

func (d7 *DigitArray) silentSetFloatValue(v float64, dpPos int) {
	if v > float64(d7.maxValue) || v < float64(d7.minValue) || dpPos < 0 ||
		dpPos >= len(d7.digits) {
		return
	}
	d7.dpPos = dpPos
	for k := d7.dpPos; k > 0; k-- {
		v = v * 10
	}
	v = math.Round(v)
	i, _ := math.Modf(v)
	d7.silentSetValue(int(i))
}

// func (d7 *DigitArray) silentSetFloatValue(v float64, dpPos int) {
// 	if v > float64(d7.maxValue) || v < float64(d7.minValue) || dpPos < 0 ||
// 		dpPos >= len(d7.digits) {
// 		return
// 	}
// 	i, _ := math.Modf(v)
// 	if i < 0 {
// 		i = -i
// 	}
// 	n := 0
// 	for i > 0 {
// 		i /= 10
// 		n++
// 	}
// 	d7.dpPos = len(d7.digits) - n - 1
// 	for k := d7.dpPos; k > 0; k-- {
// 		v = v * 10
// 	}
// 	v = math.Round(v)
// 	i, _ = math.Modf(v)
// 	d7.silentSetValue(int(i))
// }

func (d7 *DigitArray) SetFloatValue(v float64, dpPos int) {
	defer d7.lock.Unlock()
	d7.lock.Lock()
	if v > float64(d7.maxValue) || v < -float64(d7.maxValue) || dpPos < 0 ||
		dpPos >= len(d7.digits) {
		return
	}
	d7.silentSetFloatValue(v, dpPos)
	if d7.OnChanged != nil {
		d7.OnChanged(float64(d7.Value))
	}
	d7.Refresh()
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
	// limit := len(d7rend.objects) - 1
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
	// d7rend.d7array.trailingZeroesOn = true
	trailing := d7rend.d7array.trailingZeroesOn
	// slog.Debug("segments", "trailing", trailing)
	showDigitFrom := len(d7rend.d7array.digits) - 1
	// slog.Debug("refreshNumber", "trailing", trailing, "val", d7rend.d7array.Value,
	// "dpPos", d7rend.d7array.dpPos)
	if !trailing {
		showDigitFrom = d7rend.d7array.dpPos
		for i := len(d7rend.d7array.digits) - 1; i >= d7rend.d7array.dpPos; i-- {
			// slog.Debug("refreshNumber", "i", i, "digit val", d7rend.d7array.digits[i].val)
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
			// if i < limit {
			so.StrokeWidth = segmentWidth / 2
			// slog.Debug("segments", "val", val, "i", i)
			if segments[val][(i-numberIndex)%8] {
				if digitIndex == d7rend.d7array.digitCursor {
					so.StrokeColor = d7rend.d7array.CursorColor
				} else {
					so.StrokeColor = color
				}
			} else {
				so.StrokeColor = d7rend.d7array.OffColor
			}
			// }
			// canvas.Refresh(so)
			fyne.Do(so.Refresh)
		case *canvas.Circle:
			if d7rend.d7array.dpPos >= 0 {
				// slog.Debug("Circle at:", "d7rend.d7array.Value",
				// 	d7rend.d7array.Value, "digitIndex", digitIndex, "dpPos", d7rend.d7array.dpPos)
				so.StrokeWidth = segmentWidth / 2
				if digitIndex > 0 && digitIndex == d7rend.d7array.dpPos {
					so.StrokeColor = d7rend.d7array.onColor
					so.FillColor = d7rend.d7array.onColor
				} else {
					so.StrokeColor = d7rend.d7array.OffColor
					so.FillColor = d7rend.d7array.OffColor
				}
			}
			// canvas.Refresh(so)
			fyne.Do(so.Refresh)
		default:
			// log.Printf("Bad type: %T\n", so)
			// panic(0)
		}
	}
}
func (d7 *DigitArray) floatDigitSetRelPos(numOfDigits int, segmentThickness float32) {
	plusSpace := float32(0)
	thirdCounter := 0
	for digitIndex := d7.dpPos - 1; digitIndex < numOfDigits; digitIndex++ {
		if (thirdCounter)%3 == 0 {
			plusSpace = (d7.spaceBetweenDigits + d7.skew) * spaceMultiplier
		} else {
			plusSpace = 0
		}
		d7.digits[digitIndex].relPos = fyne.Position{X: plusSpace,
			Y: segmentThickness / 2}
		thirdCounter++
	}
	thirdCounter = 0
	for digitIndex := d7.dpPos - 1; digitIndex >= 0; digitIndex-- {
		if (thirdCounter)%3 == 0 {
			plusSpace = (d7.spaceBetweenDigits + d7.skew) * spaceMultiplier
		} else {
			plusSpace = 0
		}
		d7.digits[digitIndex].relPos = fyne.Position{X: plusSpace,
			Y: segmentThickness / 2}
		thirdCounter++
	}
}

func (d7 *DigitArray) intDigitSetRelPos(segmentThickness float32) {
	plusSpace := float32(0)
	thirdCounter := 1
	for digitIndex := 0; digitIndex < len(d7.digits); digitIndex++ {
		if (thirdCounter)%3 == 0 {
			plusSpace = (d7.spaceBetweenDigits + d7.skew) * spaceMultiplier
		} else {
			plusSpace = 0
		}
		d7.digits[digitIndex].relPos = fyne.Position{X: plusSpace,
			Y: segmentThickness / 2}
		thirdCounter++
	}
}

func (d7 *DigitArray) digitSetRelPos() {
	numOfDigits := len(d7.digits)
	segmentThickness := d7.segmentWidth
	for digitIndex := numOfDigits - 1; digitIndex >= 0; digitIndex-- {
		d7.digits[digitIndex].relPos.X = 0
	}
	plusSpace := float32(0)
	if d7.label != nil {
		plusSpace += d7.label.MinSize().Width + d7.spaceBetweenDigits + d7.skew
	}
	if d7.signed != UnSigned {
		plusSpace += d7.digitWidth + d7.spaceBetweenDigits + d7.skew
	}
	if d7.dpPos > 0 {
		d7.floatDigitSetRelPos(numOfDigits, segmentThickness)
	} else {
		d7.intDigitSetRelPos(segmentThickness)
	}
	d7.digits[numOfDigits-1].relPos.X += plusSpace
	for digitIndex := numOfDigits - 2; digitIndex >= 0; digitIndex-- {
		d7.digits[digitIndex].relPos.X += d7.digits[digitIndex+1].relPos.X +
			d7.digitWidth + d7.spaceBetweenDigits + d7.skew
		d7.digits[digitIndex].relPos.Y = segmentThickness / 2
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
	// start := 2 /*sign*/ + 1 /*label*/ + 1 /*unit*/
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

func (d7 *DigitArray) setDigitAtIndex(digit, index int) {
	// if d7.signed {
	// 	if index == len(d7.digits)-1 {
	// 		return
	// 	}
	// }
	tmp := d7.Value
	if tmp >= 0 {
		tmp -= int(math.Pow(10, float64(index))) * d7.digits[index].val
		tmp += int(math.Pow(10, float64(index))) * digit
	} else {
		tmp += int(math.Pow(10, float64(index))) * d7.digits[index].val
		tmp -= int(math.Pow(10, float64(index))) * digit
	}
	d7.setValue(tmp)
	d7.Refresh()
}

func (d7 *DigitArray) setDigitAtDigitCursor(digit int) {
	if d7.digitCursor >= 0 && d7.digitCursor < len(d7.digits) {
		d7.setDigitAtIndex(digit, d7.digitCursor)
	}
}

func (d7 *DigitArray) setDigitAtX(x float32, digit int) {
	for digitIndex := range d7.digits {
		pos := d7.digits[digitIndex].relPos
		if x > pos.X {
			d7.setDigitAtIndex(digit, digitIndex)
			break
		}
	}
}

func (d7 *DigitArray) clearDigitAt(x float32) {
	d7.setDigitAtX(x, 0)
}
func (d7 *DigitArray) Dragged(event *fyne.DragEvent) {
	if d7.Readonly {
		return
	}
	d7.clearDigitAt(event.Position.X)
}
func (d7 *DigitArray) DragEnd() {
}

func (d7 *DigitArray) MouseDown(event *desktop.MouseEvent) {
	if d7.Readonly {
		return
	}
	if event.Button == desktop.MouseButtonPrimary {
		d7.clearDigitAt(event.Position.X)
	}
}

func (d7 *DigitArray) MouseUp(event *desktop.MouseEvent) {
}

func (d7 *DigitArray) down(digitIndex int) {
	if digitIndex < 0 || digitIndex >= len(d7.digits) || d7.Value <= d7.minValue {
		return
	}
	step := int(math.Pow(10, float64(digitIndex)))
	switch {
	case d7.digits[digitIndex].val >= 1:
		d7.setValue(d7.Value - step)
	case d7.digits[digitIndex].val == 0:
		for i := digitIndex + 1; i < len(d7.digits); i++ {
			if d7.digits[i].val > 0 {
				d7.setValue(d7.Value - step)
				return
			}
		}
		for i := digitIndex; i >= 0; i-- {
			step = int(math.Pow(10, float64(i)))
			if d7.digits[i].val > 0 {
				d7.setValue(d7.Value - step)
				return
			}
		}
		step = int(math.Pow(10, float64(digitIndex)))
		d7.setValue(d7.Value - step)
	}
}

func (d7 *DigitArray) Scrolled(event *fyne.ScrollEvent) {
	if d7.Readonly {
		return
	}
	// digitIndex, out := d7.digitIndex(event.Position.X)
	// if out {
	// 	return
	// }
	// step := int(math.Pow(10, float64(digitIndex)))
	step := int(math.Pow(10, float64(d7.digitCursor)))
	if event.Scrolled.DY > 0 {
		d7.setValue(d7.Value + step)
	} else {
		d7.down(d7.digitCursor)
	}
	d7.Refresh()
}

func (d7 *DigitArray) setDigitCursor(x float32) bool {
	for digitIndex := range d7.digits {
		pos := d7.digits[digitIndex].relPos
		if x > pos.X {
			// if d7.signed && digitIndex == len(d7.digits)-1 {
			// 	d7.digitCursor = digitCursorOut
			// 	return false
			// }
			d7.digitCursor = digitIndex
			break
		}
	}
	return true
}
func (d7 *DigitArray) MouseIn(e *desktop.MouseEvent) {
	if d7.Readonly {
		return
	}
	d7.mousePos = e.Position
	if d7.setDigitCursor(e.Position.X) {
		d7.Window.Canvas().Focus(d7)
	}
	d7.Refresh()
}

func (d7 *DigitArray) MouseMoved(e *desktop.MouseEvent) {
	if d7.Readonly {
		return
	}
	d7.mousePos = e.Position
	if d7.setDigitCursor(e.Position.X) {
		d7.Window.Canvas().Focus(d7)
	}
	d7.Refresh()
}

func (d7 *DigitArray) MouseOut() {
	if d7.Readonly {
		return
	}
	d7.Window.Canvas().Unfocus()
	d7.digitCursor = digitCursorOut
	d7.Refresh()
}
