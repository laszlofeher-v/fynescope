package disp7

import (
	"fmt"
	"image/color"
	"log/slog"
	"math"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
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
		relPos        fyne.Position
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
)

const (
	y = true
	n = false
)

var (
	segments = [][]bool{
		{y, y, y, y, y, y, n}, // 0
		{n, y, y, n, n, n, n}, // 1
		{y, y, n, y, y, n, y}, // 2
		{y, y, y, y, n, n, y}, // 3
		{n, y, y, n, n, y, y}, // 4
		{y, n, y, y, n, y, y}, // 5
		{y, n, y, y, y, y, y}, // 6
		{y, y, y, n, n, n, n}, // 7
		{y, y, y, y, y, y, y}, // 8
		{y, y, y, y, n, y, y}, // 9
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
	if label != "" {
		disp.label = canvas.NewText(label, onColor)
		disp.label.TextStyle = fyne.TextStyle{Monospace: true}
		disp.label.TextSize = digitHeight / 2
	}
	slog.Debug("sizes", "unit", unit, "W", disp.unit.MinSize().Width)
	dpSpace := float32(0)
	if numOfFractionDigits > 0 {
		dpSpace = spaceMultiplier * (disp.spaceBetweenDigits + disp.skew)
	}
	signSpace := float32(0)
	if signed != UnSigned {
		signSpace = disp.digitWidth + disp.spaceBetweenDigits + disp.skew
	}
	labelSpace := float32(0)
	if disp.label != nil {
		labelSpace = disp.label.MinSize().Width + disp.spaceBetweenDigits + disp.skew
	}
	disp.size.Width = float32(numOfDigits)*disp.digitWidth +
		float32(numOfDigits-1)*(disp.spaceBetweenDigits+disp.skew) +
		labelSpace +
		disp.unit.MinSize().Width + disp.spaceBetweenDigits + disp.skew +
		spaceMultiplier*(disp.spaceBetweenDigits+disp.skew)*
			(float32(numOfDigits)/3) + dpSpace + signSpace
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
	d7.lock.Lock()
	d7.maxValue = maxValue
	d7.minValue = minValue
	
	needsRefresh := false
	if d7.Value < d7.minValue {
		d7.silentSetValue(d7.minValue)
		needsRefresh = true
	} else if d7.Value > d7.maxValue {
		d7.silentSetValue(d7.maxValue)
		needsRefresh = true
	}
	d7.lock.Unlock()
	
	if needsRefresh {
		d7.Refresh()
	}
}

func (d7 *DigitArray) SetOncolor(col color.Color) {
	d7.lock.Lock()
	d7.onColor = col
	if d7.label != nil {
		d7.label.Color = col
	}
	if d7.unit != nil {
		d7.unit.Color = col
	}
	d7.lock.Unlock()
	d7.Refresh()
}

func (d7 *DigitArray) DpPos() int {
	defer d7.lock.Unlock()
	d7.lock.Lock()
	return d7.dpPos
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
	d7.lock.Lock()
	d7.silentSetValue(v)
	val := float64(d7.Value)
	onChanged := d7.OnChanged
	d7.lock.Unlock()
	
	if onChanged != nil {
		onChanged(val)
	}
}

func (d7 *DigitArray) SetUnit(unitName string) {
	d7.lock.Lock()
	d7.unit = canvas.NewText(unitName, d7.onColor)
	d7.unit.TextStyle = fyne.TextStyle{Monospace: true}
	d7.unit.TextSize = d7.size.Height / 2
	d7.lock.Unlock()
	d7.Refresh()
}
func (d7 *DigitArray) SetLabel(label string) {
	d7.lock.Lock()
	slog.Debug("set label", "label", label)
	d7.label = canvas.NewText(label, d7.onColor)
	d7.label.TextStyle = fyne.TextStyle{Monospace: true}
	d7.label.TextSize = d7.size.Height / 2
	d7.lock.Unlock()
	d7.Refresh()
}

func (d7 *DigitArray) SilentSetFloatValue(v float64, dpPos int) {
	defer d7.lock.Unlock()
	d7.lock.Lock()
	d7.silentSetFloatValue(v, dpPos)
}

func (d7 *DigitArray) silentSetFloatValue(v float64, dpPos int) {
	if dpPos < 0 || dpPos >= len(d7.digits) {
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

func (d7 *DigitArray) SetFloatValue(v float64, dpPos int) {
	d7.lock.Lock()
	if dpPos < 0 || dpPos >= len(d7.digits) {
		d7.lock.Unlock()
		return
	}
	d7.silentSetFloatValue(v, dpPos)
	val := float64(d7.Value)
	onChanged := d7.OnChanged
	d7.lock.Unlock()
	
	if onChanged != nil {
		onChanged(val)
	}
	d7.Refresh()
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
