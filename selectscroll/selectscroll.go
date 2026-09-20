package selectscroll

import (
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
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

// Sample-rate unit strings shared across packages.
const (
	UnitGSps = "GS/s"
	UnitMSps = "MS/s"
	UnitKSps = "kS/s"
	UnitSps  = "S/s"
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

func (selScr *SelectScroll) CreateRenderer() fyne.WidgetRenderer {
	r := selScr.Select.CreateRenderer()
	selScr.ExtendBaseWidget(selScr)
	return r
}

func (selScr *SelectScroll) SilentSetSelectedIndex(index int) {
	savedOnChangedFunc := selScr.OnChanged
	selScr.OnChanged = nil
	defer func() { selScr.OnChanged = savedOnChangedFunc }()
	selScr.SetSelectedIndex(index)
}

func (selScr *SelectScroll) SilentSetSelected(option string) {
	savedOnChangedFunc := selScr.OnChanged
	selScr.OnChanged = nil
	defer func() { selScr.OnChanged = savedOnChangedFunc }()
	selScr.SetSelected(option)
}

func parseOptionToValue(s string) (float64, bool) {
	s = strings.ReplaceAll(s, "±", "")
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "/div", "")

	units := []struct {
		suffix     string
		multiplier float64
	}{
		{UnitGSps, 1e9}, {UnitMSps, 1e6}, {UnitKSps, 1e3}, {"KS/s", 1e3}, {UnitSps, 1.0},
		{"MHz", 1e6}, {"kHz", 1e3}, {"Hz", 1.0},
		{"mV", 1e-3}, {"V", 1.0},
		{"ms", 1e-3}, {"us", 1e-6}, {"µs", 1e-6}, {"ns", 1e-9}, {"ps", 1e-12}, {"s", 1.0},
		{"MΩ", 1e6}, {"kΩ", 1e3}, {"mΩ", 1e-3}, {"Ω", 1.0},
		{"µH", 1e-6}, {"uH", 1e-6}, {"mH", 1e-3}, {"H", 1.0},
		{"mF", 1e-3}, {"µF", 1e-6}, {"uF", 1e-6}, {"nF", 1e-9}, {"pF", 1e-12},
	}

	multiplier := 1.0
	valStr := s

	for _, u := range units {
		if strings.HasSuffix(s, u.suffix) {
			multiplier = u.multiplier
			valStr = strings.TrimSuffix(s, u.suffix)
			break
		}
	}

	val, err := strconv.ParseFloat(valStr, 64)
	if err == nil {
		return val * multiplier, true
	}

	for _, u := range units {
		if s == u.suffix {
			return u.multiplier, true
		}
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

func canvasContains(root fyne.CanvasObject, target fyne.CanvasObject) bool {
	if root == nil || target == nil {
		return false
	}
	if root == target {
		return true
	}
	switch c := root.(type) {
	case *fyne.Container:
		for _, child := range c.Objects {
			if canvasContains(child, target) {
				return true
			}
		}
	case *container.AppTabs:
		if sel := c.Selected(); sel != nil {
			if canvasContains(sel.Content, target) {
				return true
			}
		}
	case *container.Scroll:
		return canvasContains(c.Content, target)
	case *container.Split:
		return canvasContains(c.Leading, target) || canvasContains(c.Trailing, target)
	case fyne.Widget:
		r := c.CreateRenderer()
		if r != nil {
			for _, child := range r.Objects() {
				if canvasContains(child, target) {
					return true
				}
			}
		}
	}
	return false
}

func isCanvasMounted(c fyne.Canvas, target fyne.CanvasObject) bool {
	if c == nil || target == nil {
		return false
	}
	if canvasContains(c.Content(), target) {
		return true
	}
	if c.Overlays() != nil && c.Overlays().Top() != nil {
		if canvasContains(c.Overlays().Top(), target) {
			return true
		}
	}
	return false
}

func (selScr *SelectScroll) focus() {
	if app := fyne.CurrentApp(); app != nil && app.Driver() != nil {
		if c := app.Driver().CanvasForObject(selScr); c != nil {
			if isCanvasMounted(c, selScr) {
				c.Focus(selScr)
				if c.Focused() == selScr {
					return
				}
			}
		}
		for _, w := range app.Driver().AllWindows() {
			if w != nil && w.Canvas() != nil {
				if !isCanvasMounted(w.Canvas(), selScr) {
					continue
				}
				w.Canvas().Focus(selScr)
				if w.Canvas().Focused() == selScr {
					return
				}
			}
		}
	}
}

// Tapped focuses this select widget and opens the dropdown menu.
func (selScr *SelectScroll) Tapped(event *fyne.PointEvent) {
	selScr.focus()
	selScr.Select.Tapped(event)
}

// MouseIn focuses this select widget when the mouse cursor enters its bounding area.
func (selScr *SelectScroll) MouseIn(event *desktop.MouseEvent) {
	selScr.focus()
	selScr.Select.MouseIn(event)
}

func (selScr *SelectScroll) Scrolled(event *fyne.ScrollEvent) {
	selScr.focus()
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
