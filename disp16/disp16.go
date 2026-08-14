package disp16

import (
	"fmt"
	"image/color"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	DefaultDotSize      = float32(4)
	DefaultDotSpacing   = float32(2)
	DefaultCursorColor  = "gray"
	digitCursorOut      = -1
	labelIndex          = 0
	numberIndex         = 1
	spaceMultiplier     = 1.5
)

var (
	// 5x7 font for 0-9, A-F. Each row is 5 bits (LSB is rightmost pixel).
	hexFont = [16][7]byte{
		{0x0E, 0x11, 0x13, 0x15, 0x19, 0x11, 0x0E}, // 0
		{0x04, 0x0C, 0x04, 0x04, 0x04, 0x04, 0x0E}, // 1
		{0x0E, 0x11, 0x01, 0x02, 0x04, 0x08, 0x1F}, // 2
		{0x1F, 0x02, 0x04, 0x02, 0x01, 0x11, 0x0E}, // 3
		{0x02, 0x06, 0x0A, 0x12, 0x1F, 0x02, 0x02}, // 4
		{0x1F, 0x10, 0x1E, 0x01, 0x01, 0x11, 0x0E}, // 5
		{0x06, 0x08, 0x10, 0x1E, 0x11, 0x11, 0x0E}, // 6
		{0x1F, 0x01, 0x02, 0x04, 0x08, 0x08, 0x08}, // 7
		{0x0E, 0x11, 0x11, 0x0E, 0x11, 0x11, 0x0E}, // 8
		{0x0E, 0x11, 0x11, 0x0F, 0x01, 0x02, 0x0C}, // 9
		{0x0E, 0x11, 0x11, 0x1F, 0x11, 0x11, 0x11}, // A
		{0x1E, 0x11, 0x11, 0x1E, 0x11, 0x11, 0x1E}, // B
		{0x0E, 0x11, 0x10, 0x10, 0x10, 0x11, 0x0E}, // C
		{0x1E, 0x11, 0x11, 0x11, 0x11, 0x11, 0x1E}, // D
		{0x1F, 0x10, 0x10, 0x1E, 0x10, 0x10, 0x1F}, // E
		{0x1F, 0x10, 0x10, 0x1E, 0x10, 0x10, 0x10}, // F
	}
)

type (
	disp16Desc struct {
		val    int
		relPos fyne.Position
	}
	HexArray struct {
		widget.BaseWidget
		onColor            color.Color
		OffColor           color.Color
		CursorColor        color.Color
		Readonly           bool
		size               fyne.Size
		dotSize            float32
		dotSpacing         float32
		digits             []disp16Desc
		Value              uint64
		numOfDigits        int
		OnChanged          func(v uint64)
		Window             fyne.Window
		mousePos           fyne.Position
		digitCursor        int
		spaceBetweenDigits float32
		label              *canvas.Text
		lock               sync.Mutex
	}
)

func NewHexArray(numOfDigits int, w fyne.Window, onColor color.Color, readOnly bool, label string) (*HexArray, error) {
	if numOfDigits <= 0 {
		return nil, fmt.Errorf("numOfDigits is 0")
	}
	disp := &HexArray{
		numOfDigits: numOfDigits,
		digits:      make([]disp16Desc, numOfDigits),
	}
	for i := range disp.digits {
		disp.digits[i] = disp16Desc{}
	}
	disp.ExtendBaseWidget(disp)
	
	disp.dotSize = DefaultDotSize
	disp.dotSpacing = DefaultDotSpacing
	disp.spaceBetweenDigits = disp.dotSize * 3

	digitWidth := float32(5)*disp.dotSize + float32(4)*disp.dotSpacing
	digitHeight := float32(7)*disp.dotSize + float32(6)*disp.dotSpacing

	disp.size.Height = digitHeight

	if label != "" {
		disp.label = canvas.NewText(label, onColor)
		disp.label.TextStyle = fyne.TextStyle{Monospace: true}
		disp.label.TextSize = digitHeight / 2
	}
	
	labelSpace := float32(0)
	if disp.label != nil {
		labelSpace = disp.label.MinSize().Width + disp.spaceBetweenDigits
	}
	
	disp.size.Width = float32(numOfDigits)*digitWidth +
		float32(numOfDigits-1)*disp.spaceBetweenDigits +
		labelSpace

	disp.onColor = onColor
	disp.CursorColor = theme.PrimaryColorNamed(DefaultCursorColor)
	disp.Window = w
	disp.digitCursor = digitCursorOut
	disp.Readonly = readOnly
	return disp, nil
}

func (h *HexArray) SetValue(val uint64) {
	h.lock.Lock()
	h.silentSetValue(val)
	h.lock.Unlock()
	h.Refresh()
	if h.OnChanged != nil {
		h.OnChanged(h.Value)
	}
}

func (h *HexArray) silentSetValue(val uint64) {
	// Max value based on digits
	maxVal := uint64(1<<(h.numOfDigits*4)) - 1
	if val > maxVal {
		val = maxVal
	}
	h.Value = val

	temp := val
	for i := h.numOfDigits - 1; i >= 0; i-- {
		h.digits[i].val = int(temp & 0x0F)
		temp >>= 4
	}
}

func (h *HexArray) GetValue() uint64 {
	h.lock.Lock()
	defer h.lock.Unlock()
	return h.Value
}
