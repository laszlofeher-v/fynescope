package selectscroll

import (
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type (
	Exception    int
	SelectScroll struct {
		widget.Select
		overflowIndex int
		scrolled      bool
	}
)

const (
	None Exception = iota
	Over
	Under
)

func NewSelectScroll(options []string, changed func(option string, exception Exception), placeHolder string) *SelectScroll {
	var selScr *SelectScroll
	wrapper := func(option string) {
		if !selScr.scrolled {
			selScr.overflowIndex = selScr.SelectedIndex()
			changed(option, None)
		} else {
			selScr.scrolled = false
			switch {
			case selScr.overflowIndex > selScr.SelectedIndex():
				changed(option, Over)
			case selScr.overflowIndex < selScr.SelectedIndex():
				changed(option, Under)
			default:
				changed(option, None)
			}
			selScr.overflowIndex = selScr.SelectedIndex()
		}
	}
	selScr = &SelectScroll{Select: widget.Select{Options: options, OnChanged: wrapper, PlaceHolder: placeHolder}}
	selScr.ExtendBaseWidget(selScr)
	return selScr
}

func (selScr *SelectScroll) SilentSetSelectedIndex(index int) {
	savedOnChangedFunc := selScr.OnChanged
	selScr.OnChanged = nil
	selScr.SetSelectedIndex(index)
	selScr.OnChanged = savedOnChangedFunc
}

func (selScr *SelectScroll) SilentSetSelected(option string) {
	savedOnChangedFunc := selScr.OnChanged
	selScr.OnChanged = nil
	selScr.SetSelected(option)
	selScr.OnChanged = savedOnChangedFunc
}

func parseOptionToValue(s string) (float64, bool) {
	s = strings.ReplaceAll(s, "±", "")
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "/div", "")

	multiplier := 1.0
	valStr := s

	if strings.HasSuffix(s, "GS/s") {
		multiplier = 1e9
		valStr = strings.TrimSuffix(s, "GS/s")
	} else if strings.HasSuffix(s, "MS/s") {
		multiplier = 1e6
		valStr = strings.TrimSuffix(s, "MS/s")
	} else if strings.HasSuffix(s, "kS/s") || strings.HasSuffix(s, "KS/s") {
		multiplier = 1e3
		valStr = strings.TrimSuffix(s, "kS/s")
		valStr = strings.TrimSuffix(valStr, "KS/s")
	} else if strings.HasSuffix(s, "S/s") {
		multiplier = 1.0
		valStr = strings.TrimSuffix(s, "S/s")
	} else if strings.HasSuffix(s, "MHz") {
		multiplier = 1e6
		valStr = strings.TrimSuffix(s, "MHz")
	} else if strings.HasSuffix(s, "kHz") {
		multiplier = 1e3
		valStr = strings.TrimSuffix(s, "kHz")
	} else if strings.HasSuffix(s, "Hz") {
		multiplier = 1.0
		valStr = strings.TrimSuffix(s, "Hz")
	} else if strings.HasSuffix(s, "mV") {
		multiplier = 1e-3
		valStr = strings.TrimSuffix(s, "mV")
	} else if strings.HasSuffix(s, "V") {
		multiplier = 1.0
		valStr = strings.TrimSuffix(s, "V")
	} else if strings.HasSuffix(s, "ms") {
		multiplier = 1e-3
		valStr = strings.TrimSuffix(s, "ms")
	} else if strings.HasSuffix(s, "us") || strings.HasSuffix(s, "µs") {
		multiplier = 1e-6
		valStr = strings.TrimSuffix(s, "us")
		valStr = strings.TrimSuffix(valStr, "µs")
	} else if strings.HasSuffix(s, "ns") {
		multiplier = 1e-9
		valStr = strings.TrimSuffix(s, "ns")
	} else if strings.HasSuffix(s, "ps") {
		multiplier = 1e-12
		valStr = strings.TrimSuffix(s, "ps")
	} else if strings.HasSuffix(s, "s") {
		multiplier = 1.0
		valStr = strings.TrimSuffix(s, "s")
	} else if strings.HasSuffix(s, "MΩ") {
		multiplier = 1e6
		valStr = strings.TrimSuffix(s, "MΩ")
	} else if strings.HasSuffix(s, "kΩ") {
		multiplier = 1e3
		valStr = strings.TrimSuffix(s, "kΩ")
	} else if strings.HasSuffix(s, "mΩ") {
		multiplier = 1e-3
		valStr = strings.TrimSuffix(s, "mΩ")
	} else if strings.HasSuffix(s, "Ω") {
		multiplier = 1.0
		valStr = strings.TrimSuffix(s, "Ω")
	} else if strings.HasSuffix(s, "µH") || strings.HasSuffix(s, "uH") {
		multiplier = 1e-6
		valStr = strings.TrimSuffix(s, "µH")
		valStr = strings.TrimSuffix(valStr, "uH")
	} else if strings.HasSuffix(s, "mH") {
		multiplier = 1e-3
		valStr = strings.TrimSuffix(s, "mH")
	} else if strings.HasSuffix(s, "H") {
		multiplier = 1.0
		valStr = strings.TrimSuffix(s, "H")
	} else if strings.HasSuffix(s, "mF") {
		multiplier = 1e-3
		valStr = strings.TrimSuffix(s, "mF")
	} else if strings.HasSuffix(s, "µF") || strings.HasSuffix(s, "uF") {
		multiplier = 1e-6
		valStr = strings.TrimSuffix(s, "µF")
		valStr = strings.TrimSuffix(valStr, "uF")
	} else if strings.HasSuffix(s, "nF") {
		multiplier = 1e-9
		valStr = strings.TrimSuffix(s, "nF")
	} else if strings.HasSuffix(s, "pF") {
		multiplier = 1e-12
		valStr = strings.TrimSuffix(s, "pF")
	}

	val, err := strconv.ParseFloat(valStr, 64)
	if err == nil {
		return val * multiplier, true
	}

	unitOnlyMultipliers := map[string]float64{
		"s":    1.0,
		"ms":   1e-3,
		"us":   1e-6,
		"µs":   1e-6,
		"ns":   1e-9,
		"ps":   1e-12,
		"Hz":   1.0,
		"kHz":  1e3,
		"MHz":  1e6,
		"S/s":  1.0,
		"kS/s": 1e3,
		"KS/s": 1e3,
		"MS/s": 1e6,
		"GS/s": 1e9,
		"mΩ":   1e-3,
		"Ω":    1.0,
		"kΩ":   1e3,
		"MΩ":   1e6,
		"µH":   1e-6,
		"mH":   1e-3,
		"H":    1.0,
		"pF":   1e-12,
		"nF":   1e-9,
		"µF":   1e-6,
		"mF":   1e-3,
	}
	if m, ok := unitOnlyMultipliers[s]; ok {
		return m, true
	}

	return 0, false
}

func isAscending(options []string) bool {
	var firstVal, lastVal float64
	var foundFirst, foundLast bool

	for _, opt := range options {
		if val, ok := parseOptionToValue(opt); ok {
			firstVal = val
			foundFirst = true
			break
		}
	}

	for idx := len(options) - 1; idx >= 0; idx-- {
		if val, ok := parseOptionToValue(options[idx]); ok {
			lastVal = val
			foundLast = true
			break
		}
	}

	if foundFirst && foundLast {
		return firstVal < lastVal
	}
	return false
}

func (selScr *SelectScroll) Scrolled(event *fyne.ScrollEvent) {
	var (
		i int
		s string
	)
	selScr.scrolled = true
	for i, s = range selScr.Options {
		if s == selScr.Selected {
			break
		}
	}

	ascending := isAscending(selScr.Options)
	var nextIndex int

	if event.Scrolled.DY > 0 { // Scroll Up (increases value)
		if ascending {
			nextIndex = i + 1
		} else {
			nextIndex = i - 1
		}
	} else { // Scroll Down (decreases value)
		if ascending {
			nextIndex = i - 1
		} else {
			nextIndex = i + 1
		}
	}

	selScr.overflowIndex = nextIndex
	selScr.SetSelectedIndex(nextIndex)
	if selScr.overflowIndex != selScr.SelectedIndex() {
		selScr.OnChanged(selScr.Selected)
	}
}
