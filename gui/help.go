package gui

import (
	"image/color"
	"reflect"
	"strings"
	"sync"
	"time"
	"unsafe"

	"fynescope/disp16"
	"fynescope/disp7"
	"fynescope/selectscroll"
	"fynescope/sliderscroll"

	"fyne.io/fyne/v2"
	canvasPkg "fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// FocusCheck is a checkbox widget that gains input focus on mouse hover and tap.
type FocusCheck struct {
	widget.Check
	scp *ScpDesc
}

func (scp *ScpDesc) newFocusCheck(label string, changed func(bool)) *FocusCheck {
	fc := &FocusCheck{
		Check: widget.Check{Text: label, OnChanged: changed},
		scp:   scp,
	}
	fc.ExtendBaseWidget(fc)
	return fc
}

func (fc *FocusCheck) CreateRenderer() fyne.WidgetRenderer {
	r := fc.Check.CreateRenderer()
	fc.ExtendBaseWidget(fc)
	return r
}

func (fc *FocusCheck) focus() {
	if fc.scp != nil {
		fc.scp.focusWidget(fc)
		return
	}
	focusObject(fc)
}

func (fc *FocusCheck) MouseIn(e *desktop.MouseEvent) {
	fc.focus()
	fc.Check.MouseIn(e)
}

func (fc *FocusCheck) MouseMoved(e *desktop.MouseEvent) {
	fc.focus()
	fc.Check.MouseMoved(e)
}

func (fc *FocusCheck) Tapped(e *fyne.PointEvent) {
	fc.focus()
	fc.Check.Tapped(e)
}

// FocusButton is a button widget that gains input focus on mouse hover and tap.
type FocusButton struct {
	widget.Button
	scp *ScpDesc
}

func (scp *ScpDesc) newFocusButton(text string, tapped func()) *FocusButton {
	fb := &FocusButton{
		Button: widget.Button{Text: text, OnTapped: tapped},
		scp:    scp,
	}
	fb.ExtendBaseWidget(fb)
	return fb
}

func (scp *ScpDesc) newFocusButtonWithIcon(text string, icon fyne.Resource, tapped func()) *FocusButton {
	fb := &FocusButton{
		Button: widget.Button{Text: text, Icon: icon, OnTapped: tapped},
		scp:    scp,
	}
	fb.ExtendBaseWidget(fb)
	return fb
}

func (fb *FocusButton) CreateRenderer() fyne.WidgetRenderer {
	r := fb.Button.CreateRenderer()
	fb.ExtendBaseWidget(fb)
	return r
}

func (fb *FocusButton) focus() {
	if fb.scp != nil {
		fb.scp.focusWidget(fb)
		return
	}
	focusObject(fb)
}

func (fb *FocusButton) MouseIn(e *desktop.MouseEvent) {
	fb.focus()
	fb.Button.MouseIn(e)
}

func (fb *FocusButton) MouseMoved(e *desktop.MouseEvent) {
	fb.focus()
	fb.Button.MouseMoved(e)
}

func (fb *FocusButton) Tapped(e *fyne.PointEvent) {
	fb.focus()
	fb.Button.Tapped(e)
}

func canvasContains(root fyne.CanvasObject, target fyne.CanvasObject) bool {
	if root == nil || target == nil {
		return false
	}
	if root == target {
		return true
	}
	if c, ok := root.(*fyne.Container); ok {
		for _, child := range c.Objects {
			if canvasContains(child, target) {
				return true
			}
		}
	}
	return false
}

func focusObject(obj fyne.Focusable) {
	if obj == nil {
		return
	}
	app := fyne.CurrentApp()
	if app == nil || app.Driver() == nil {
		return
	}
	co, _ := obj.(fyne.CanvasObject)
	if co != nil {
		if c := app.Driver().CanvasForObject(co); c != nil {
			c.Focus(obj)
			if c.Focused() == obj {
				return
			}
		}
	}
	for _, w := range app.Driver().AllWindows() {
		if w != nil && w.Canvas() != nil {
			if co != nil && !canvasContains(w.Canvas().Content(), co) {
				continue
			}
			w.Canvas().Focus(obj)
			if w.Canvas().Focused() == obj {
				return
			}
		}
	}
}

// focusWidget focuses a focusable canvas object on the oscilloscope window.
func (scp *ScpDesc) focusWidget(obj fyne.Focusable) {
	if scp == nil || obj == nil {
		return
	}
	if scp.Window != nil && scp.Window.Canvas() != nil {
		scp.Window.Canvas().Focus(obj)
		if scp.Window.Canvas().Focused() == obj {
			return
		}
	}
	focusObject(obj)
}

func containsCanvasObject(objs []fyne.CanvasObject, target fyne.CanvasObject) bool {
	for _, obj := range objs {
		if obj == target {
			return true
		}
	}
	return false
}

// getTabButton returns the canvas object representing the button for a TabItem if initialized.
func getTabButton(item *container.TabItem) fyne.CanvasObject {
	if item == nil {
		return nil
	}
	v := reflect.ValueOf(item)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return nil
	}
	f := v.Elem().FieldByName("button")
	if !f.IsValid() {
		return nil
	}
	btnVal := reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()
	if btnVal.IsNil() {
		return nil
	}
	btn, ok := btnVal.Interface().(fyne.CanvasObject)
	if !ok {
		return nil
	}
	return btn
}

// isTabButtonHovered checks if the tab button has its internal hovered flag set.
func isTabButtonHovered(item *container.TabItem) bool {
	if item == nil {
		return false
	}
	v := reflect.ValueOf(item)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return false
	}
	f := v.Elem().FieldByName("button")
	if !f.IsValid() {
		return false
	}
	btnVal := reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()
	if btnVal.IsNil() {
		return false
	}
	elem := btnVal.Elem()
	if !elem.IsValid() {
		return false
	}
	h := elem.FieldByName("hovered")
	if !h.IsValid() {
		return false
	}
	hVal := reflect.NewAt(h.Type(), unsafe.Pointer(h.UnsafeAddr())).Elem()
	return hVal.Bool()
}

// TabFocusProxy is a Focusable proxy for an AppTabs TabItem, allowing tab buttons
// to receive keyboard/mouse focus, display a visual focus outline, and trigger
// the contextual help popup.
type TabFocusProxy struct {
	widget.BaseWidget
	scp     *ScpDesc
	item    *container.TabItem
	tabText string
}

func (scp *ScpDesc) getOrCreateTabProxy(item *container.TabItem) *TabFocusProxy {
	if scp == nil || item == nil {
		return nil
	}
	scp.helpMu.Lock()
	defer scp.helpMu.Unlock()
	if scp.tabFocusProxies == nil {
		scp.tabFocusProxies = make(map[*container.TabItem]*TabFocusProxy)
	}
	p, exists := scp.tabFocusProxies[item]
	if !exists {
		p = &TabFocusProxy{
			scp:     scp,
			item:    item,
			tabText: item.Text,
		}
		p.ExtendBaseWidget(p)
		scp.tabFocusProxies[item] = p
	}
	if scp.helpOverlay != nil && !containsCanvasObject(scp.helpOverlay.Objects, p) {
		scp.helpOverlay.Add(p)
	}
	return p
}

func (p *TabFocusProxy) CreateRenderer() fyne.WidgetRenderer {
	p.ExtendBaseWidget(p)
	return widget.NewSimpleRenderer(container.NewWithoutLayout())
}

func (p *TabFocusProxy) FocusGained() {
	if p.scp != nil {
		p.scp.setTabFocusHighlight(p.item, true)
	}
}

func (p *TabFocusProxy) FocusLost() {
	if p.scp != nil {
		p.scp.setTabFocusHighlight(p.item, false)
	}
}

func (p *TabFocusProxy) TypedRune(_ rune) {}

func (p *TabFocusProxy) TypedKey(e *fyne.KeyEvent) {
	if p.scp == nil || p.scp.controlTab == nil || p.item == nil {
		return
	}
	switch e.Name {
	case fyne.KeySpace, fyne.KeyReturn:
		p.scp.controlTab.Select(p.item)
	case fyne.KeyLeft, fyne.KeyUp:
		items := p.scp.controlTab.Items
		for i, it := range items {
			if it == p.item && i > 0 {
				prevProxy := p.scp.getOrCreateTabProxy(items[i-1])
				p.scp.focusWidget(prevProxy)
				p.scp.controlTab.Select(items[i-1])
				break
			}
		}
	case fyne.KeyRight, fyne.KeyDown:
		items := p.scp.controlTab.Items
		for i, it := range items {
			if it == p.item && i < len(items)-1 {
				nextProxy := p.scp.getOrCreateTabProxy(items[i+1])
				p.scp.focusWidget(nextProxy)
				p.scp.controlTab.Select(items[i+1])
				break
			}
		}
	}
}

func (p *TabFocusProxy) Position() fyne.Position {
	if btn := getTabButton(p.item); btn != nil {
		return btn.Position()
	}
	return p.BaseWidget.Position()
}

func (p *TabFocusProxy) Size() fyne.Size {
	if btn := getTabButton(p.item); btn != nil {
		return btn.Size()
	}
	return p.BaseWidget.Size()
}

func (scp *ScpDesc) setTabFocusHighlight(item *container.TabItem, visible bool) {
	if scp == nil {
		return
	}
	scp.helpMu.Lock()
	defer scp.helpMu.Unlock()

	if !visible || item == nil || scp.helpOverlay == nil {
		if scp.tabFocusRect != nil {
			scp.tabFocusRect.Hide()
			if scp.helpOverlay != nil {
				scp.helpOverlay.Refresh()
			}
		}
		return
	}

	btn := getTabButton(item)
	if btn == nil {
		return
	}

	var absPos fyne.Position
	if app := fyne.CurrentApp(); app != nil && app.Driver() != nil {
		absPos = app.Driver().AbsolutePositionForObject(btn)
	} else {
		absPos = btn.Position()
	}
	btnSize := btn.Size()
	if btnSize.Width <= 0 || btnSize.Height <= 0 {
		return
	}

	if scp.tabFocusRect == nil {
		scp.tabFocusRect = canvasPkg.NewRectangle(color.Transparent)
		scp.tabFocusRect.StrokeWidth = 2
	}
	th := theme.DefaultTheme()
	if app := fyne.CurrentApp(); app != nil && app.Settings() != nil && app.Settings().Theme() != nil {
		th = app.Settings().Theme()
	}
	v := fyne.CurrentApp().Settings().ThemeVariant()
	scp.tabFocusRect.StrokeColor = th.Color(theme.ColorNameFocus, v)
	scp.tabFocusRect.CornerRadius = th.Size(theme.SizeNameSelectionRadius)
	scp.tabFocusRect.Move(absPos)
	scp.tabFocusRect.Resize(btnSize)
	scp.tabFocusRect.Show()

	found := false
	for _, obj := range scp.helpOverlay.Objects {
		if obj == scp.tabFocusRect {
			found = true
			break
		}
	}
	if !found {
		scp.helpOverlay.Add(scp.tabFocusRect)
	}
	scp.helpOverlay.Refresh()
}

func (scp *ScpDesc) findHoveredTabItem() *container.TabItem {
	if scp == nil || scp.controlTab == nil {
		return nil
	}
	for _, item := range scp.controlTab.Items {
		if isTabButtonHovered(item) {
			return item
		}
	}
	if scp.Window != nil {
		gw := getGLFWWindow(scp.Window)
		if gw != nil {
			xpos, ypos := gw.GetCursorPos()
			scale := float32(1)
			if scp.Window.Canvas() != nil && scp.Window.Canvas().Scale() > 0 {
				scale = scp.Window.Canvas().Scale()
			}
			cursorX := float32(xpos) / scale
			cursorY := float32(ypos) / scale
			app := fyne.CurrentApp()
			for _, item := range scp.controlTab.Items {
				btn := getTabButton(item)
				if btn == nil {
					continue
				}
				var absPos fyne.Position
				if app != nil && app.Driver() != nil {
					absPos = app.Driver().AbsolutePositionForObject(btn)
				} else {
					absPos = btn.Position()
				}
				size := btn.Size()
				if size.Width > 0 && size.Height > 0 &&
					cursorX >= absPos.X && cursorX <= absPos.X+size.Width &&
					cursorY >= absPos.Y && cursorY <= absPos.Y+size.Height {
					return item
				}
			}
		}
	}
	return nil
}

func (scp *ScpDesc) checkTabHoverFocus() {
	if scp == nil || scp.controlTab == nil || !scp.IsHelpEnabled() {
		return
	}
	hoveredItem := scp.findHoveredTabItem()
	var currFocused fyne.Focusable
	if scp.Window != nil && scp.Window.Canvas() != nil {
		currFocused = scp.Window.Canvas().Focused()
	}
	currProxy, isTabProxy := currFocused.(*TabFocusProxy)

	if hoveredItem != nil {
		if !isTabProxy || currProxy.item != hoveredItem {
			proxy := scp.getOrCreateTabProxy(hoveredItem)
			scp.focusWidget(proxy)
		}
	} else if isTabProxy {
		if scp.Window != nil && scp.Window.Canvas() != nil {
			scp.Window.Canvas().Unfocus()
		}
		scp.setTabFocusHighlight(nil, false)
	}
}

var (
	// FocusHelpDelay specifies how long a widget must remain continuously focused
	// before the contextual help popup is automatically displayed.
	FocusHelpDelay = 1500 * time.Millisecond

	// FocusHelpCheckInterval specifies how frequently the focus monitor checks canvas focus.
	FocusHelpCheckInterval = 100 * time.Millisecond

	widgetHelpRegistry    = make(map[fyne.CanvasObject]WidgetHelpInfo)
	widgetHelpRegistryMtx sync.RWMutex

	namedHelpRegistry = make(map[string]WidgetHelpInfo)
	prefixHelpRules   []prefixHelpRule
	helpInitOnce      sync.Once
)

type prefixHelpRule struct {
	prefix string
	info   WidgetHelpInfo
}

// WidgetHelpInfo stores title and detailed description text for a widget.
type WidgetHelpInfo struct {
	Title       string
	Description string
}

// RegisterWidgetHelp associates custom help information directly with a canvas object.
func RegisterWidgetHelp(obj fyne.CanvasObject, title, description string) {
	if obj == nil {
		return
	}
	widgetHelpRegistryMtx.Lock()
	widgetHelpRegistry[obj] = WidgetHelpInfo{
		Title:       title,
		Description: description,
	}
	widgetHelpRegistryMtx.Unlock()
}

func initNamedHelp() {
	helpInitOnce.Do(func() {
		// Top toolbar actions
		namedHelpRegistry[helpButtonId] = WidgetHelpInfo{
			Title:       "Help Toggle (?)",
			Description: "Toggles contextual focus help on or off. When enabled, holding focus on any control for a moment displays helpful guidance on its functionality.",
		}
		namedHelpRegistry["runblockButton"] = WidgetHelpInfo{
			Title:       "Run / Pause",
			Description: "Starts or stops continuous oscilloscope waveform acquisition and display updates.",
		}
		namedHelpRegistry["streamEnableButton"] = WidgetHelpInfo{
			Title:       "Streaming Mode",
			Description: "Toggles real-time streaming acquisition mode for continuous gapless data capture.",
		}
		namedHelpRegistry["timeZoomButton"] = WidgetHelpInfo{
			Title:       "Time Zoom",
			Description: "Opens the Time Zoom window to examine high-resolution details of captured waveforms.",
		}
		namedHelpRegistry["recordGifButton"] = WidgetHelpInfo{
			Title:       "Record GIF",
			Description: "Records live oscilloscope display frames and saves them as an animated GIF image file.",
		}
		namedHelpRegistry["saveRasterButton"] = WidgetHelpInfo{
			Title:       "Save Raster (R)",
			Description: "Saves the current waveform signal screen to a PNG image file.",
		}
		namedHelpRegistry["saveWindowButton"] = WidgetHelpInfo{
			Title:       "Save Window (W)",
			Description: "Captures and saves the entire application window to a PNG image file.",
		}
		namedHelpRegistry[fullScreenId] = WidgetHelpInfo{
			Title:       "Full Screen",
			Description: "Expands the oscilloscope window to occupy the entire display monitor.",
		}
		namedHelpRegistry[restoreScreenId] = WidgetHelpInfo{
			Title:       "Restore Window",
			Description: "Restores the oscilloscope window from full screen to normal windowed size.",
		}
		namedHelpRegistry[changeSideId] = WidgetHelpInfo{
			Title:       "Move Controls",
			Description: "Relocates the side control panel between the left and right sides of the waveform screen.",
		}
		namedHelpRegistry["themeChangeAction"] = WidgetHelpInfo{
			Title:       "Toggle Theme",
			Description: "Switches the user interface color scheme between dark theme and light theme.",
		}
		namedHelpRegistry["logout"] = WidgetHelpInfo{
			Title:       "Disconnect & Quit",
			Description: "Safely stops data capture, shuts down scope hardware, saves settings, and quits the application.",
		}

		// Channel controls
		namedHelpRegistry[triggerDirectionId] = WidgetHelpInfo{
			Title:       "Trigger Direction",
			Description: "Selects the trigger threshold edge slope: Rising edge, Falling edge, Above, or Below.",
		}
		namedHelpRegistry[timeSelectId] = WidgetHelpInfo{
			Title:       "Timebase Scale",
			Description: "Adjusts horizontal time scale per division across the waveform screen.",
		}
		namedHelpRegistry[unitSelectId] = WidgetHelpInfo{
			Title:       "Timebase Unit",
			Description: "Selects time scale unit: nanoseconds (ns), microseconds (µs), milliseconds (ms), or seconds (s).",
		}
		namedHelpRegistry["complexTriggerCheck"] = WidgetHelpInfo{
			Title:       "Complex Trigger",
			Description: "Enables advanced complex hardware triggering conditions and qualifiers.",
		}

		// Function tabs
		namedHelpRegistry[ftFuncId] = WidgetHelpInfo{
			Title:       "f(t) Scope",
			Description: "Time domain oscilloscope waveform display mode.",
		}
		namedHelpRegistry["f(t)"] = namedHelpRegistry[ftFuncId]

		namedHelpRegistry[fvFuncId] = WidgetHelpInfo{
			Title:       "f(v) X-Y Mode",
			Description: "Voltage vs voltage Lissajous X-Y waveform display mode.",
		}
		namedHelpRegistry["f(v)"] = namedHelpRegistry[fvFuncId]

		namedHelpRegistry[dftFuncId] = WidgetHelpInfo{
			Title:       "FFT Spectrum",
			Description: "Fast Fourier Transform frequency spectrum analyzer display mode.",
		}
		namedHelpRegistry["FFT"] = namedHelpRegistry[dftFuncId]

		namedHelpRegistry[ffFuncId] = WidgetHelpInfo{
			Title:       "f(f) Bode Plot",
			Description: "Frequency response analysis and Bode plot sweep mode.",
		}
		namedHelpRegistry["f(f)"] = namedHelpRegistry[ffFuncId]

		namedHelpRegistry[rlcFuncId] = WidgetHelpInfo{
			Title:       "RLC Meter",
			Description: "Component measurement mode for resistance, inductance, and capacitance.",
		}
		namedHelpRegistry["RLC"] = namedHelpRegistry[rlcFuncId]

		namedHelpRegistry[digPortFuncId] = WidgetHelpInfo{
			Title:       "Digital Channels",
			Description: "Mixed-signal logic analyzer digital input port configuration.",
		}
		namedHelpRegistry["digital"] = namedHelpRegistry[digPortFuncId]

		namedHelpRegistry[genFuncId] = WidgetHelpInfo{
			Title:       "Signal Generator",
			Description: "Internal AWG waveform signal generator controls.",
		}
		namedHelpRegistry["gen"] = namedHelpRegistry[genFuncId]

		namedHelpRegistry[extgenFuncId] = WidgetHelpInfo{
			Title:       "External Generator",
			Description: "External SCPI programmable signal generator controls.",
		}
		namedHelpRegistry["extgen"] = namedHelpRegistry[extgenFuncId]

		namedHelpRegistry[digGenFuncId] = WidgetHelpInfo{
			Title:       "Digital Generator",
			Description: "Digital pattern generator stimulus output controls.",
		}
		namedHelpRegistry["digGen"] = namedHelpRegistry[digGenFuncId]

		namedHelpRegistry[vchFuncId] = WidgetHelpInfo{
			Title:       "Virtual Channels",
			Description: "Mathematical and virtual channel waveform definitions.",
		}
		namedHelpRegistry["vch"] = namedHelpRegistry[vchFuncId]

		namedHelpRegistry[decodeFuncId] = WidgetHelpInfo{
			Title:       "Protocol Decoder",
			Description: "Serial protocol decoding for UART, SPI, I2C, and CAN buses.",
		}
		namedHelpRegistry["decode"] = namedHelpRegistry[decodeFuncId]

		namedHelpRegistry[filterFuncId] = WidgetHelpInfo{
			Title:       "Digital Filter",
			Description: "Digital signal filter configuration and parameters.",
		}
		namedHelpRegistry["filter"] = namedHelpRegistry[filterFuncId]

		// Prefix rules for per-channel or dynamically named controls
		prefixHelpRules = []prefixHelpRule{
			// f(v) controls
			{
				prefix: fvEnableId,
				info: WidgetHelpInfo{
					Title:       "Channel Enable (f(v))",
					Description: "Enables or disables this channel for X-Y voltage vs voltage display.",
				},
			},
			{
				prefix: fvXCheckId,
				info: WidgetHelpInfo{
					Title:       "X-Axis Mode",
					Description: "Selects this channel as the horizontal X-axis signal for f(v) X-Y mode.",
				},
			},
			{
				prefix: fvX10Id,
				info: WidgetHelpInfo{
					Title:       "10x Probe Attenuation",
					Description: "Applies 10x voltage scaling for passive oscilloscope probes with attenuation.",
				},
			},
			{
				prefix: fvVRangeId,
				info: WidgetHelpInfo{
					Title:       "Voltage Scale",
					Description: "Sets vertical sensitivity scale (volts or millivolts per screen division).",
				},
			},
			// FFT (DFT) controls
			{
				prefix: dftEnableId,
				info: WidgetHelpInfo{
					Title:       "Channel Enable (FFT)",
					Description: "Enables or disables frequency spectrum display for this channel.",
				},
			},
			{
				prefix: dftPersId,
				info: WidgetHelpInfo{
					Title:       "Persistence",
					Description: "Accumulates past spectral traces on screen to visualize intermittent frequency peaks and noise floor.",
				},
			},
			{
				prefix: dftX10Id,
				info: WidgetHelpInfo{
					Title:       "10x Probe Attenuation",
					Description: "Applies 10x voltage scaling for passive oscilloscope probes with attenuation.",
				},
			},
			{
				prefix: dftVRangeId,
				info: WidgetHelpInfo{
					Title:       "Voltage Scale",
					Description: "Sets vertical sensitivity scale (volts or millivolts per screen division).",
				},
			},
			{
				prefix: timeSelectId,
				info: WidgetHelpInfo{
					Title:       "Timebase Scale",
					Description: "Adjusts horizontal time scale per division across the waveform screen.",
				},
			},
			{
				prefix: unitSelectId,
				info: WidgetHelpInfo{
					Title:       "Timebase Unit",
					Description: "Selects time scale unit: nanoseconds (ns), microseconds (µs), milliseconds (ms), or seconds (s).",
				},
			},
			{
				prefix: "complexTrigger",
				info: WidgetHelpInfo{
					Title:       "Complex Trigger",
					Description: "Enables advanced complex hardware triggering conditions and qualifiers.",
				},
			},

			{
				prefix: chEnableId,
				info: WidgetHelpInfo{
					Title:       "Channel Enable",
					Description: "Enables or disables waveform display and signal acquisition for this channel.",
				},
			},
			{
				prefix: chOffsetId,
				info: WidgetHelpInfo{
					Title:       "Voltage Offset",
					Description: "Adjusts vertical DC voltage offset level for this channel.",
				},
			},
			{
				prefix: vRangeId,
				info: WidgetHelpInfo{
					Title:       "Voltage Scale",
					Description: "Sets vertical sensitivity scale (volts or millivolts per screen division).",
				},
			},
			{
				prefix: acdcId,
				info: WidgetHelpInfo{
					Title:       "Input Coupling",
					Description: "Selects input coupling: DC coupling, AC coupling (blocks DC bias), or 50-ohm termination.",
				},
			},
			{
				prefix: invertId,
				info: WidgetHelpInfo{
					Title:       "Invert Waveform",
					Description: "Inverts the voltage polarity of the channel waveform.",
				},
			},
			{
				prefix: triggerCheckId,
				info: WidgetHelpInfo{
					Title:       "Trigger Source",
					Description: "Selects this analog channel as the primary source for hardware triggering.",
				},
			},
			{
				prefix: persId,
				info: WidgetHelpInfo{
					Title:       "Persistence",
					Description: "Accumulates past waveform traces on screen to visualize noise, jitter, and intermittent events.",
				},
			},
			{
				prefix: x10Id,
				info: WidgetHelpInfo{
					Title:       "10x Probe Attenuation",
					Description: "Applies 10x voltage scaling for passive oscilloscope probes with attenuation.",
				},
			},
			// Timebase & Acquisition
			{
				prefix: "timeDiv",
				info: WidgetHelpInfo{
					Title:       "Timebase Scale",
					Description: "Adjusts horizontal time scale per division across the waveform screen.",
				},
			},
			{
				prefix: "timeUnit",
				info: WidgetHelpInfo{
					Title:       "Timebase Unit",
					Description: "Selects time scale unit: nanoseconds (ns), microseconds (µs), milliseconds (ms), or seconds (s).",
				},
			},
			{
				prefix: "sampleRate",
				info: WidgetHelpInfo{
					Title:       "Sample Rate",
					Description: "Configures ADC sampling frequency for digitizing input signals.",
				},
			},
			{
				prefix: "sampleUnit",
				info: WidgetHelpInfo{
					Title:       "Sample Rate Unit",
					Description: "Selects sampling rate frequency unit: kS/s, MS/s, or GS/s.",
				},
			},
			// FFT / Spectrum analyzer
			{
				prefix: dftModeId + "MinFreq",
				info: WidgetHelpInfo{
					Title:       "Spectrum Min Frequency",
					Description: "Sets the lower frequency limit displayed on the FFT spectrum analyzer.",
				},
			},
			{
				prefix: "dftMinFreq",
				info: WidgetHelpInfo{
					Title:       "Spectrum Min Frequency",
					Description: "Sets the lower frequency limit displayed on the FFT spectrum analyzer.",
				},
			},
			{
				prefix: dftMaxFreqValId,
				info: WidgetHelpInfo{
					Title:       "Spectrum Max Frequency",
					Description: "Sets the upper frequency limit displayed on the FFT spectrum analyzer.",
				},
			},
			{
				prefix: dftWindowId,
				info: WidgetHelpInfo{
					Title:       "FFT Window",
					Description: "Selects the windowing filter function (Hann, Hamming, Blackman, Rectangular, etc.) to control spectral leakage.",
				},
			},
			{
				prefix: dftModeId,
				info: WidgetHelpInfo{
					Title:       "Spectrum Scale",
					Description: "Sets spectral vertical scale: dBFS, dBm, or Linear magnitude.",
				},
			},
			{
				prefix: dftBinId,
				info: WidgetHelpInfo{
					Title:       "FFT Resolution Bins",
					Description: "Selects the number of frequency bins (FFT length) for spectral analysis.",
				},
			},
			// Frequency response / Bode sweep
			{
				prefix: ffMinFreqId,
				info: WidgetHelpInfo{
					Title:       "Sweep Start Frequency",
					Description: "Start frequency for Bode frequency response analysis sweep.",
				},
			},
			{
				prefix: ffMaxFreqId,
				info: WidgetHelpInfo{
					Title:       "Sweep Stop Frequency",
					Description: "Stop frequency for Bode frequency response analysis sweep.",
				},
			},
			{
				prefix: ffDispUnitSelectId,
				info: WidgetHelpInfo{
					Title:       "Bode Display Units",
					Description: "Selects frequency response display unit.",
				},
			},
			{
				prefix: ffExtGenSelectId,
				info: WidgetHelpInfo{
					Title:       "External Generator Stimulus",
					Description: "Routes Bode frequency sweep stimulus through external signal generator.",
				},
			},
			// Signal generator / AWG
			{
				prefix: extGenOnOffId,
				info: WidgetHelpInfo{
					Title:       "Generator Output",
					Description: "Enables or disables signal generator output.",
				},
			},
			{
				prefix: extGenWaveTypeId,
				info: WidgetHelpInfo{
					Title:       "Waveform Shape",
					Description: "Selects signal generator output shape: Sine, Square, Triangle, Ramp, DC, or Noise.",
				},
			},
			{
				prefix: extGenFreqId,
				info: WidgetHelpInfo{
					Title:       "Generator Frequency",
					Description: "Sets periodic waveform frequency generated by the hardware.",
				},
			},
			{
				prefix: extGenAmpId,
				info: WidgetHelpInfo{
					Title:       "Generator Amplitude",
					Description: "Sets peak-to-peak output voltage amplitude of the generated waveform.",
				},
			},
			{
				prefix: extGenOffsetId,
				info: WidgetHelpInfo{
					Title:       "Generator Offset",
					Description: "Sets DC bias voltage added to the generated signal output.",
				},
			},
			{
				prefix: extGenPhaseId,
				info: WidgetHelpInfo{
					Title:       "Generator Phase",
					Description: "Sets relative phase angle shift for signal generator waveform.",
				},
			},
			{
				prefix: extGenImpOhmsId,
				info: WidgetHelpInfo{
					Title:       "Output Impedance",
					Description: "Selects generator output termination impedance: 50 ohms or 600 ohms.",
				},
			},
			// Digital / MSO logic analyzer
			{
				prefix: "digPortTrigEnable",
				info: WidgetHelpInfo{
					Title:       "Digital Pattern Trigger",
					Description: "Enables multi-channel digital logic pattern triggering.",
				},
			},
			{
				prefix: "digPortOperandSelect",
				info: WidgetHelpInfo{
					Title:       "Logic Operand",
					Description: "Selects logical combination operator (AND / OR) for digital trigger channels.",
				},
			},
			{
				prefix: "digPort0EnableCheck",
				info: WidgetHelpInfo{
					Title:       "Port 0 Enable",
					Description: "Enables or disables digital input channels D0-D7.",
				},
			},
			{
				prefix: "digPort1EnableCheck",
				info: WidgetHelpInfo{
					Title:       "Port 1 Enable",
					Description: "Enables or disables digital input channels D8-D15.",
				},
			},
		}
	})
}

// IsHelpEnabled returns whether the contextual focus help feature is currently active.
func (scp *ScpDesc) IsHelpEnabled() bool {
	if scp == nil || scp.Settings == nil || scp.Settings.Window.HelpEnabled == nil {
		return true
	}
	return *scp.Settings.Window.HelpEnabled
}

// SetHelpEnabled turns the contextual focus help feature on or off and updates settings and UI.
func (scp *ScpDesc) SetHelpEnabled(enabled bool) {
	if scp == nil || scp.Settings == nil {
		return
	}
	scp.Settings.Window.HelpEnabled = &enabled
	scp.updateHelpButtonState()
	if !enabled {
		scp.hideHelpPopUp()
	}
	scp.SaveSettings()
}

func (scp *ScpDesc) toggleHelp() {
	scp.SetHelpEnabled(!scp.IsHelpEnabled())
}

func (scp *ScpDesc) updateHelpButtonState() {
	if scp == nil || scp.helpButton == nil {
		return
	}
	if scp.IsHelpEnabled() {
		scp.helpButton.Importance = widget.HighImportance
	} else {
		scp.helpButton.Importance = widget.LowImportance
	}
	scp.helpButton.Refresh()
}

// IsHelpVisible returns whether a help card or popup is currently on screen.
func (scp *ScpDesc) IsHelpVisible() bool {
	if scp == nil {
		return false
	}
	scp.helpMu.Lock()
	defer scp.helpMu.Unlock()
	return (scp.helpCard != nil && scp.helpCard.Visible()) || scp.helpPopUp != nil
}

// windowMouseTracker is an invisible widget that wraps window content and
// implements desktop.Hoverable. When MouseOut fires (the pointer leaves the
// widget, i.e. the app window), it calls onMouseLeftWindow() so that keyboard
// focus and any open help popups are immediately cleared whenever the mouse exits.
type windowMouseTracker struct {
	widget.BaseWidget
	scp     *ScpDesc
	win     fyne.Window
	content fyne.CanvasObject
}

func (scp *ScpDesc) newWindowMouseTracker(win fyne.Window, content fyne.CanvasObject) *windowMouseTracker {
	w := &windowMouseTracker{scp: scp, win: win, content: content}
	w.ExtendBaseWidget(w)
	return w
}

func (t *windowMouseTracker) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.content)
}

func (t *windowMouseTracker) MouseIn(_ *desktop.MouseEvent)    {}
func (t *windowMouseTracker) MouseMoved(_ *desktop.MouseEvent) {}
func (t *windowMouseTracker) MouseOut() {
	if t.scp != nil {
		t.scp.onMouseLeftWindow()
	} else if t.win != nil && t.win.Canvas() != nil {
		t.win.Canvas().Unfocus()
	}
}

// setContentWithHelp wraps the window content with a floating help overlay layer.
func (scp *ScpDesc) setContentWithHelp(content fyne.CanvasObject) {
	if scp == nil || scp.Window == nil {
		return
	}
	scp.helpMu.Lock()
	if scp.helpOverlay == nil {
		scp.helpOverlay = container.NewWithoutLayout()
	}
	overlay := scp.helpOverlay
	if scp.tabFocusProxies != nil {
		for _, p := range scp.tabFocusProxies {
			if !containsCanvasObject(overlay.Objects, p) {
				overlay.Add(p)
			}
		}
	}
	scp.helpMu.Unlock()
	tracker := scp.newWindowMouseTracker(scp.Window, content)
	root := container.NewStack(tracker, overlay)
	scp.Window.SetContent(root)
	scp.hookWindowMouseLeave(scp.Window)
}

func (scp *ScpDesc) getHelpForWidget(focused fyne.Focusable) (string, string) {
	initNamedHelp()

	// Suppress waveform screen rasters explicitly: they never show help popups
	switch focused.(type) {
	case *screenRaster:
		return "", ""
	}

	co, isCanvasObj := focused.(fyne.CanvasObject)
	if isCanvasObj && co != nil {
		// 1. Direct object registry lookup
		widgetHelpRegistryMtx.RLock()
		if info, exists := widgetHelpRegistry[co]; exists {
			widgetHelpRegistryMtx.RUnlock()
			return info.Title, info.Description
		}
		widgetHelpRegistryMtx.RUnlock()

		// 2. Control name lookup via test_proxy reverse map
		if name, ok := getControlNameByObject(co); ok && name != "" {
			if info, exists := namedHelpRegistry[name]; exists {
				return info.Title, info.Description
			}
			for _, rule := range prefixHelpRules {
				if strings.HasPrefix(name, rule.prefix) {
					return rule.info.Title, rule.info.Description
				}
			}
		}
	}

	// 3. Heuristic fallback based on widget concrete type
	switch w := focused.(type) {
	case *FocusButton:
		title := "Button"
		if w.Text != "" {
			title = w.Text + " Button"
		}
		return title, "Click or press Space/Enter to activate this action."
	case *widget.Button:
		title := "Button"
		if w.Text != "" {
			title = w.Text + " Button"
		}
		return title, "Click or press Space/Enter to activate this action."
	case *TabFocusProxy:
		if info, exists := namedHelpRegistry[w.tabText]; exists {
			return info.Title, info.Description
		}
		title := w.tabText + " Tab"
		return title, "Selects the " + w.tabText + " function tab."
	case *FocusCheck:
		title := "Checkbox"
		if w.Text != "" {
			title = w.Text
		}
		switch strings.ToLower(w.Text) {
		case "enabled":
			return "Channel Enable", "Enables or disables signal acquisition and waveform display for this channel."
		case "x-axis":
			return "X-Axis Mode", "Selects this channel as the horizontal X-axis signal for f(v) X-Y mode."
		case "inv":
			return "Invert Waveform", "Inverts the voltage polarity of the channel waveform."
		case "trig":
			return "Trigger Source", "Selects this analog channel as the primary source for hardware triggering."
		case "pers":
			return "Persistence", "Accumulates past waveform traces on screen to visualize noise, jitter, and intermittent events."
		case "x10", "x1":
			return "10x Probe Attenuation", "Applies 10x voltage scaling for passive oscilloscope probes with attenuation."
		case "cmpx":
			return "Complex Trigger", "Enables advanced complex hardware triggering conditions and qualifiers."
		}
		return title, "Click or press Space to toggle this option on or off."
	case *widget.Check:
		title := "Checkbox"
		if w.Text != "" {
			title = w.Text
		}
		switch strings.ToLower(w.Text) {
		case "enabled":
			return "Channel Enable", "Enables or disables signal acquisition and waveform display for this channel."
		case "x-axis":
			return "X-Axis Mode", "Selects this channel as the horizontal X-axis signal for f(v) X-Y mode."
		case "inv":
			return "Invert Waveform", "Inverts the voltage polarity of the channel waveform."
		case "trig":
			return "Trigger Source", "Selects this analog channel as the primary source for hardware triggering."
		case "pers":
			return "Persistence", "Accumulates past waveform traces on screen to visualize noise, jitter, and intermittent events."
		case "x10", "x1":
			return "10x Probe Attenuation", "Applies 10x voltage scaling for passive oscilloscope probes with attenuation."
		case "cmpx":
			return "Complex Trigger", "Enables advanced complex hardware triggering conditions and qualifiers."
		}
		return title, "Click or press Space to toggle this option on or off."
	case *disp7.DigitArray:
		return "Numeric Value Editor", "Interactive 7-segment numeric display. Click individual digits, use arrow keys, or scroll mouse wheel to increment or decrement the value."
	case *disp16.HexArray:
		return "Hex Value Editor", "Interactive 16-segment hexadecimal display. Click characters or use keyboard/mouse wheel to modify value."
	case *screenRaster:
		return "", ""

	case *widget.Entry:
		title := "Text Entry"
		if w.PlaceHolder != "" {
			title = w.PlaceHolder
		}
		return title, "Type text or numeric values into this input field."
	case *sliderscroll.SliderScroll, *widget.Slider:
		return "Value Slider", "Drag the slider knob or scroll mouse wheel to adjust value smoothly."
	case *selectscroll.SelectScroll:
		title := "Selection Option"
		if w.Selected != "" {
			title = w.Selected
		}
		if len(w.Options) == 2 && ((w.Options[0] == "AC" && w.Options[1] == "DC") || (w.Options[0] == "DC" && w.Options[1] == "AC")) {
			return "Input Coupling (" + title + ")", "Selects input coupling: AC coupling (blocks DC bias) or DC coupling."
		}
		return "Selection Option (" + title + ")", "Click, use arrow keys, or scroll mouse wheel to choose from available configuration options."
	case *widget.Select:
		return "Selection Dropdown", "Click or scroll mouse wheel to choose from available configuration options."
	default:
		return "Control", "Focused user interface control. Interact using mouse or keyboard."
	}
}

func (scp *ScpDesc) findFocusedWidget() (fyne.Focusable, fyne.Canvas) {
	if scp == nil {
		return nil, nil
	}
	if scp.Window != nil && scp.Window.Canvas() != nil {
		if f := scp.Window.Canvas().Focused(); f != nil {
			return f, scp.Window.Canvas()
		}
	}
	if scp.App != nil && scp.App.Driver() != nil {
		for _, w := range scp.App.Driver().AllWindows() {
			if w != nil && w.Canvas() != nil {
				if f := w.Canvas().Focused(); f != nil {
					return f, w.Canvas()
				}
			}
		}
	}
	return nil, nil
}

func (scp *ScpDesc) startFocusHelp() {
	scp.hookAllWindows()
	scp.helpMu.Lock()
	if scp.helpQuit != nil {
		scp.helpMu.Unlock()
		return
	}
	scp.helpQuit = make(chan struct{})
	scp.helpMu.Unlock()

	go func() {
		ticker := time.NewTicker(40 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-scp.helpQuit:
				return
			case <-ticker.C:
				fyne.Do(func() {
					scp.checkTabHoverFocus()
					scp.checkFocusHelp()
				})
			}
		}
	}()
}

func (scp *ScpDesc) stopFocusHelp() {
	scp.helpMu.Lock()
	if scp.helpQuit != nil {
		close(scp.helpQuit)
		scp.helpQuit = nil
	}
	scp.helpMu.Unlock()
	scp.hideHelpPopUp()
}

func (scp *ScpDesc) checkFocusHelp() {
	if !scp.IsHelpEnabled() {
		if scp.IsHelpVisible() {
			scp.hideHelpPopUp()
		}
		return
	}

	if !scp.isMouseInsideAppWindow() {
		scp.onMouseLeftWindow()
		return
	}

	focused, canvas := scp.findFocusedWidget()
	if focused == nil || canvas == nil {
		if scp.helpShownFor != nil {
			scp.hideHelpPopUp()
		}
		scp.helpMu.Lock()
		scp.currentFocused = nil
		scp.helpMu.Unlock()
		return
	}

	// Explicitly suppress waveform display rasters: they never show help
	switch focused.(type) {
	case *screenRaster:
		if scp.helpShownFor != nil {

			scp.hideHelpPopUp()
		}
		scp.helpMu.Lock()
		scp.currentFocused = nil
		scp.helpMu.Unlock()
		return
	}

	scp.helpMu.Lock()

	if focused != scp.currentFocused {
		scp.currentFocused = focused
		scp.focusedSince = time.Now()
		scp.helpShownFor = nil
		scp.helpMu.Unlock()
		scp.hideHelpPopUp()
		return
	}

	// Same widget remains focused: check if threshold has elapsed
	duration := time.Since(scp.focusedSince)
	shouldShow := scp.helpShownFor == nil && duration >= FocusHelpDelay
	scp.helpMu.Unlock()

	if shouldShow {
		title, desc := scp.getHelpForWidget(focused)
		if desc != "" {
			scp.helpMu.Lock()
			scp.helpShownFor = focused
			scp.helpMu.Unlock()
			scp.showHelpPopUp(focused, canvas, title, desc)
		}
	}

}

func (scp *ScpDesc) showHelpPopUp(focused fyne.Focusable, canvas fyne.Canvas, title, desc string) {
	scp.helpMu.Lock()
	defer scp.helpMu.Unlock()

	if !scp.IsHelpEnabled() {
		return
	}

	co, isCanvasObj := focused.(fyne.CanvasObject)
	if !isCanvasObj || co == nil || canvas == nil {
		return
	}

	titleLabel := widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	helpIcon := widget.NewIcon(theme.HelpIcon())
	closeBtn := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
		scp.hideHelpPopUp()
	})
	closeBtn.Importance = widget.LowImportance

	header := container.NewBorder(nil, nil, helpIcon, closeBtn, titleLabel)

	descLabel := widget.NewLabel(desc)
	descLabel.Wrapping = fyne.TextWrapWord

	inner := container.NewVBox(
		header,
		widget.NewSeparator(),
		descLabel,
	)

	var absPos fyne.Position
	targetObj := co
	if tp, ok := focused.(*TabFocusProxy); ok {
		if btn := getTabButton(tp.item); btn != nil {
			targetObj = btn
		}
	}
	if app := fyne.CurrentApp(); app != nil && app.Driver() != nil {
		absPos = app.Driver().AbsolutePositionForObject(targetObj)
	} else {
		absPos = targetObj.Position()
	}
	objSize := targetObj.Size()
	canvasSize := canvas.Size()

	const cardWidth = float32(280)
	const cardHeight = float32(110)

	targetX := absPos.X
	if canvasSize.Width > cardWidth && targetX+cardWidth > canvasSize.Width-10 {
		targetX = canvasSize.Width - cardWidth - 10
	}
	if targetX < 10 {
		targetX = 10
	}

	targetY := absPos.Y + objSize.Height + 6
	if canvasSize.Height > cardHeight && targetY+cardHeight > canvasSize.Height-10 {
		targetY = absPos.Y - cardHeight - 6
	}
	if targetY < 10 {
		targetY = 10
	}

	if scp.helpOverlay != nil {
		bg := canvasPkg.NewRectangle(theme.MenuBackgroundColor())
		bg.StrokeColor = theme.PrimaryColor()
		bg.StrokeWidth = 1
		bg.CornerRadius = 6

		card := container.NewStack(bg, container.NewPadded(inner))
		wrapper := container.NewWithoutLayout(card)
		card.Move(fyne.NewPos(targetX, targetY))
		card.Resize(fyne.NewSize(cardWidth, cardHeight))

		scp.helpCard = wrapper
		if scp.tabFocusRect != nil && scp.tabFocusRect.Visible() {
			scp.helpOverlay.Objects = []fyne.CanvasObject{scp.tabFocusRect, wrapper}
		} else {
			scp.helpOverlay.Objects = []fyne.CanvasObject{wrapper}
		}
		scp.helpOverlay.Refresh()
	} else {
		if scp.helpPopUp != nil {
			scp.helpPopUp.Hide()
			scp.helpPopUp = nil
		}
		popup := widget.NewPopUp(container.NewPadded(inner), canvas)
		scp.helpPopUp = popup
		popup.ShowAtPosition(fyne.NewPos(targetX, targetY))
	}
}

func (scp *ScpDesc) hideHelpPopUp() {
	scp.helpMu.Lock()
	if scp.helpCard != nil {
		if scp.helpOverlay != nil {
			var remaining []fyne.CanvasObject
			for _, obj := range scp.helpOverlay.Objects {
				if obj != scp.helpCard {
					remaining = append(remaining, obj)
				}
			}
			scp.helpOverlay.Objects = remaining
			scp.helpOverlay.Refresh()
		}
		scp.helpCard = nil
	}
	p := scp.helpPopUp
	scp.helpPopUp = nil
	scp.helpShownFor = nil
	scp.helpMu.Unlock()

	if p != nil {
		p.Hide()
	}
}
