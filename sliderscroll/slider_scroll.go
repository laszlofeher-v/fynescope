package sliderscroll

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

const (
	defaultMul = 1000.0
)

type (
	Exception    int
	SliderScroll struct {
		widget.Slider
		mul float64
	}
)

func NewSliderScroll(min, max float64) *SliderScroll {
	sliderScroll := &SliderScroll{
		Slider: widget.Slider{
			Value:       0,
			Min:         min,
			Max:         max,
			Step:        1,
			Orientation: widget.Horizontal,
		},
		mul: defaultMul,
	}
	sliderScroll.ExtendBaseWidget(sliderScroll)
	return sliderScroll
}

func (slScr *SliderScroll) CreateRenderer() fyne.WidgetRenderer {
	r := slScr.Slider.CreateRenderer()
	slScr.ExtendBaseWidget(slScr)
	return r
}

func (slScr *SliderScroll) SilentSetValue(v float64) {
	savedOnChangedFunc := slScr.OnChanged
	slScr.OnChanged = nil
	slScr.SetValue(v)
	slScr.OnChanged = savedOnChangedFunc
}

func (slScr *SliderScroll) MouseDown(event *desktop.MouseEvent) {
	if slScr.mul == 0 {
		slScr.mul = defaultMul
	}
	if event.Button == desktop.MouseButtonTertiary {
		if slScr.mul < 1e6 {
			slScr.mul = 10 * slScr.mul
		}
	} else if event.Button == desktop.MouseButtonSecondary {
		if slScr.mul >= 10 {
			slScr.mul = slScr.mul / 10
		}
	}
}

func (slScr *SliderScroll) MouseUp(event *desktop.MouseEvent) {
}

func (slScr *SliderScroll) FocusGained() {
	slScr.Refresh()
}

func (slScr *SliderScroll) FocusLost() {
	slScr.Refresh()
}

func (slScr *SliderScroll) TypedRune(rune) {
}

func (slScr *SliderScroll) TypedKey(event *fyne.KeyEvent) {
	if slScr.mul == 0 {
		slScr.mul = defaultMul
	}
	switch event.Name {
	case fyne.KeyUp, fyne.KeyRight:
		slScr.SetValue(slScr.Value + slScr.mul)
	case fyne.KeyDown, fyne.KeyLeft:
		slScr.SetValue(slScr.Value - slScr.mul)
	case fyne.KeyPageUp:
		slScr.SetValue(slScr.Value + slScr.mul*10)
	case fyne.KeyPageDown:
		slScr.SetValue(slScr.Value - slScr.mul*10)
	}
}

func (slScr *SliderScroll) Scrolled(event *fyne.ScrollEvent) {
	if slScr.mul == 0 {
		slScr.mul = defaultMul
	}
	// if event.Scrolled.DY > 0 {
	slScr.SetValue(slScr.Value + slScr.mul*float64(event.Scrolled.DY))
	// } else {
	// slScr.SetValue(slScr.Value - slScr.mul)
	// }
}

func (slScr *SliderScroll) MouseIn(e *desktop.MouseEvent) {
}

func (slScr *SliderScroll) MouseOut() {
}

func (slScr *SliderScroll) MouseMoved(e *desktop.MouseEvent) {
}
