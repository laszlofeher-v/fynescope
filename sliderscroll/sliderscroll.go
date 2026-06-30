package sliderscroll

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

const (
	defaultMul = 100.0
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

func (slScr *SliderScroll) SilentSetValue(v float64) {
	savedOnChangedFunc := slScr.OnChanged
	slScr.OnChanged = nil
	slScr.SetValue(v)
	slScr.OnChanged = savedOnChangedFunc
}

func (slScr *SliderScroll) MouseDown(event *desktop.MouseEvent) {
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

// TODO up,down left,right, pagedown,pageup
func (slScr *SliderScroll) Scrolled(event *fyne.ScrollEvent) {
	// if event.Scrolled.DY > 0 {
	slScr.SetValue(slScr.Value + slScr.mul*float64(event.Scrolled.DY))
	// } else {
	// slScr.SetValue(slScr.Value - slScr.mul)
	// }
}
