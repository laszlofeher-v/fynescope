// Package gui implements the Fynescope oscilloscope graphical user interface.
// The help subsystem in this file provides contextual, focus-driven help popups
// for every interactive widget in the UI.  When a user hovers over or focuses a
// control for longer than FocusHelpDelay, a small floating card appears near that
// widget explaining its purpose and usage.
//
// Architecture overview
//
//   - FocusCheck / FocusButton – thin wrappers around standard Fyne widgets that
//     steal canvas focus on mouse-enter so that the help monitor can detect which
//     control the user is looking at.
//
//   - widgetHelpRegistry – a package-level map that associates any fyne.CanvasObject
//     instance with a WidgetHelpInfo{Title, Description} pair.  Widgets register
//     themselves at construction time via RegisterWidgetHelp.
//
//   - TabFocusProxy – an invisible Focusable proxy placed in the help overlay for
//     every AppTabs tab button.  Because tab buttons are internal Fyne widgets that
//     cannot receive keyboard focus themselves, the proxy intercepts focus/hover and
//     forwards keyboard navigation (Left/Right/Up/Down) to adjacent tabs.
//
//   - getHelpForWidget – the central help-text resolver.  It first checks the
//     widgetHelpRegistry for an exact match, then falls back to a type-switch that
//     derives reasonable help text from widget type and properties (label text,
//     placeholder, option lists, etc.).
//
//   - startFocusHelp / checkFocusHelp – a 40 ms polling goroutine that monitors
//     which widget is focused.  When the same widget remains focused for
//     FocusHelpDelay without interruption, checkFocusHelp calls showHelpPopUp to
//     display the card.
package gui

import (
	"image/color"
	"log/slog"
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

// FocusCheck is a checkbox widget that gains canvas input focus whenever the
// mouse enters its bounds or the user taps it.  By holding focus it triggers the
// contextual help monitor so that a descriptive popup appears after
// FocusHelpDelay milliseconds of hover.
type FocusCheck struct {
	widget.Check
	scp *ScpDesc // back-reference to the scope for focus routing
}

// newFocusCheck creates a FocusCheck with the given label and change callback.
// ExtendBaseWidget is called so that the custom MouseIn/MouseMoved/Tapped
// overrides are dispatched correctly by the Fyne renderer.
func (scp *ScpDesc) newFocusCheck(label string, changed func(bool)) *FocusCheck {
	fc := &FocusCheck{
		Check: widget.Check{Text: label, OnChanged: changed},
		scp:   scp,
	}
	fc.ExtendBaseWidget(fc)
	return fc
}

// CreateRenderer returns the standard Check renderer.  ExtendBaseWidget is
// re-called here defensively to cope with deserialization paths that bypass the
// constructor.
func (fc *FocusCheck) CreateRenderer() fyne.WidgetRenderer {
	r := fc.Check.CreateRenderer()
	fc.ExtendBaseWidget(fc)
	return r
}

// focus routes canvas focus to this checkbox.  When a scope reference is
// available it uses focusWidget, which restricts focus to the scope window;
// otherwise it falls back to the global focusObject helper.
func (fc *FocusCheck) focus() {
	if fc.scp != nil {
		fc.scp.focusWidget(fc)
		return
	}
	focusObject(fc)
}

// MouseIn acquires focus when the pointer enters the checkbox, triggering the
// help delay timer, then delegates the event to the underlying Check widget.
func (fc *FocusCheck) MouseIn(e *desktop.MouseEvent) {
	fc.focus()
	fc.Check.MouseIn(e)
}

// MouseMoved refreshes focus while the pointer moves within the checkbox.
// This keeps the help timer alive during slow mouse movement.
func (fc *FocusCheck) MouseMoved(e *desktop.MouseEvent) {
	fc.focus()
	fc.Check.MouseMoved(e)
}

// Tapped acquires focus on click before delegating the toggle to the Check.
func (fc *FocusCheck) Tapped(e *fyne.PointEvent) {
	fc.focus()
	fc.Check.Tapped(e)
}

// FocusButton is a button widget that acquires canvas focus on mouse-enter and
// tap so that the contextual help monitor can detect user intent.  It is used
// throughout the UI as a drop-in replacement for widget.Button wherever help
// text is desirable.
type FocusButton struct {
	widget.Button
	scp *ScpDesc // back-reference to the scope for focus routing
}

// newFocusButton creates a FocusButton with a text label and tap callback.
func (scp *ScpDesc) newFocusButton(text string, tapped func()) *FocusButton {
	fb := &FocusButton{
		Button: widget.Button{Text: text, OnTapped: tapped},
		scp:    scp,
	}
	fb.ExtendBaseWidget(fb)
	return fb
}

// newFocusButtonWithIcon creates a FocusButton that displays both an icon and a
// text label beside it.
func (scp *ScpDesc) newFocusButtonWithIcon(text string, icon fyne.Resource, tapped func()) *FocusButton {
	fb := &FocusButton{
		Button: widget.Button{Text: text, Icon: icon, OnTapped: tapped},
		scp:    scp,
	}
	fb.ExtendBaseWidget(fb)
	return fb
}

// CreateRenderer returns the standard Button renderer and re-calls
// ExtendBaseWidget to handle deserialized instances.
func (fb *FocusButton) CreateRenderer() fyne.WidgetRenderer {
	r := fb.Button.CreateRenderer()
	fb.ExtendBaseWidget(fb)
	return r
}

// focus routes canvas focus to this button.  Prefers the scope-window focus
// path when a scope reference is present.
func (fb *FocusButton) focus() {
	if fb.scp != nil {
		fb.scp.focusWidget(fb)
		return
	}
	focusObject(fb)
}

// MouseIn acquires focus on pointer-enter and forwards the event to Button.
func (fb *FocusButton) MouseIn(e *desktop.MouseEvent) {
	fb.focus()
	fb.Button.MouseIn(e)
}

// MouseMoved keeps focus alive while the pointer moves over the button.
func (fb *FocusButton) MouseMoved(e *desktop.MouseEvent) {
	fb.focus()
	fb.Button.MouseMoved(e)
}

// Tapped acquires focus then activates the button's tap handler.
func (fb *FocusButton) Tapped(e *fyne.PointEvent) {
	fb.focus()
	fb.Button.Tapped(e)
}

// canvasContains performs a depth-first walk of the Fyne object tree rooted at
// root and returns true if target is found anywhere within it.  It understands
// the common Fyne container types (Container, AppTabs, Scroll, Split) as well as
// generic widgets, whose renderer objects are inspected via CreateRenderer.
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

// isCanvasObjectMounted reports whether target is part of the visible object
// hierarchy of canvas c.  It checks both the canvas main content and the top
// overlay (e.g. a popup) so that widgets in popups are not incorrectly deemed
// unmounted.
func isCanvasObjectMounted(c fyne.Canvas, target fyne.CanvasObject) bool {
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

// focusObject searches all open application windows for one whose canvas
// contains obj (as a CanvasObject) and calls Canvas.Focus on it.  It returns
// after the first successful focus grant.  This is the fallback used when no
// ScpDesc reference is available.
func focusObject(obj fyne.Focusable) {
	if obj == nil {
		return
	}
	app := fyne.CurrentApp()
	if app == nil || app.Driver() == nil {
		return
	}
	co, _ := obj.(fyne.CanvasObject)
	for _, w := range app.Driver().AllWindows() {
		if w != nil && w.Canvas() != nil {
			if co != nil && !isCanvasObjectMounted(w.Canvas(), co) {
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
	co, _ := obj.(fyne.CanvasObject)
	if scp.Window != nil && scp.Window.Canvas() != nil {
		if co != nil && isCanvasObjectMounted(scp.Window.Canvas(), co) {
			scp.Window.Canvas().Focus(obj)
			if scp.Window.Canvas().Focused() == obj {
				return
			}
		}
	}
	focusObject(obj)
}

// containsCanvasObject returns true if target is present in the flat slice objs.
// Used to avoid adding duplicate children to the help overlay container.
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

// TabFocusProxy is an invisible Focusable proxy widget associated with one
// AppTabs TabItem.  Fyne's internal tab buttons cannot receive keyboard focus
// directly, so a zero-size proxy is laid out on top of each button inside the
// help overlay.  This lets the help system:
//
//  1. Detect which tab the mouse is hovering over via checkTabHoverFocus.
//  2. Show a focus highlight rectangle (tabFocusRect) around the active tab.
//  3. Expose contextual help text registered in tabHelpInfo / widgetHelpRegistry.
//  4. Allow Left/Right arrow keys to navigate between adjacent tabs.
type TabFocusProxy struct {
	widget.BaseWidget
	scp     *ScpDesc           // owning scope, used for focus routing and overlay access
	item    *container.TabItem // the real AppTabs item this proxy represents
	tabText string             // snapshot of item.Text at creation time (avoids pointer chasing)
}

// tabHelpInfo maps a tab item text to its contextual help title and description.
var tabHelpInfo = map[string]WidgetHelpInfo{
	"f(t)":    {Title: "f(t) Scope", Description: "Time domain oscilloscope waveform display mode."},
	"f(v)":    {Title: "f(v) X-Y Mode", Description: "Voltage vs voltage Lissajous X-Y waveform display mode."},
	"FFT":     {Title: "FFT Spectrum", Description: "Fast Fourier Transform frequency spectrum analyzer display mode."},
	"f(f)":    {Title: "f(f) Bode Plot", Description: "Frequency response analysis and Bode plot sweep mode."},
	"RLC":     {Title: "RLC Meter", Description: "Generator signal dispatch and simulated analog filters."},
	"digital": {Title: "Digital Channels", Description: "Mixed-signal logic analyzer digital input port configuration."},
	"gen":     {Title: "Signal Generator", Description: "Internal AWG waveform signal generator controls."},
	"extgen":  {Title: "External Generator", Description: "External SCPI programmable signal generator controls."},
	"digGen":  {Title: "digGen Digital Generator", Description: "Simulated digital signal generator control."},
	"vch":     {Title: "vch Virtual Channels", Description: "Virtual channel control."},
	"decode":  {Title: "decode Protocol Decoder", Description: "Serial bus protocol decoder control."},
	"filter":  {Title: "filter Digital Filter", Description: "Digital signal filter control."},
	"corr":    {Title: "corr Correlation", Description: "Signal correlation and comparison control."},
}

// getOrCreateTabProxy returns the TabFocusProxy for item, creating one if it
// does not yet exist.  On creation the proxy is:
//   - stored in scp.tabFocusProxies so subsequent calls return the same instance;
//   - registered in widgetHelpRegistry with the help text from tabHelpInfo;
//   - added to scp.helpOverlay so the proxy lives inside the floating overlay
//     container and can receive Fyne hit-test events.
//
// The method is safe to call concurrently; it locks helpMu for the map read/write
// and widgetHelpRegistryMtx for the registry write.
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
		// Register help text at proxy creation time.
		if info, ok := tabHelpInfo[item.Text]; ok {
			widgetHelpRegistryMtx.Lock()
			widgetHelpRegistry[p] = info
			widgetHelpRegistryMtx.Unlock()
		}
	}
	if scp.helpOverlay != nil && !containsCanvasObject(scp.helpOverlay.Objects, p) {
		scp.helpOverlay.Add(p)
	}
	return p
}

// CreateRenderer returns a no-op renderer wrapping an empty layout container.
// The proxy is intentionally invisible; its only purpose is to receive focus
// and participate in the Fyne event dispatch pipeline.
func (p *TabFocusProxy) CreateRenderer() fyne.WidgetRenderer {
	p.ExtendBaseWidget(p)
	return widget.NewSimpleRenderer(container.NewWithoutLayout())
}

// FocusGained shows the focus outline rectangle around the tab button.
func (p *TabFocusProxy) FocusGained() {
	if p.scp != nil {
		p.scp.setTabFocusHighlight(p.item, true)
	}
}

// FocusLost hides the focus outline rectangle when the proxy loses focus.
func (p *TabFocusProxy) FocusLost() {
	if p.scp != nil {
		p.scp.setTabFocusHighlight(p.item, false)
	}
}

// TypedRune satisfies fyne.Focusable; no action is taken for character input.
func (p *TabFocusProxy) TypedRune(_ rune) {}

// Tapped selects the tab when the user clicks on it.  The proxy is overlaid
// on top of the real tab button in the help overlay, so mouse clicks land
// here instead of on the underlying AppTabs button.
func (p *TabFocusProxy) Tapped(_ *fyne.PointEvent) {
	if p.scp == nil || p.scp.controlTab == nil || p.item == nil {
		return
	}
	p.scp.controlTab.Select(p.item)
	if p.scp != nil {
		p.scp.focusWidget(p)
	}
}

// TypedKey handles keyboard navigation for the focused tab proxy:
//   - Space / Return  – select the associated tab.
//   - Left / Up       – move focus to and select the previous tab.
//   - Right / Down    – move focus to and select the next tab.
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

// setTabFocusHighlight shows or hides a coloured stroke rectangle drawn on top
// of the tab button identified by item.  The rectangle is added to helpOverlay
// and positioned using the button's absolute canvas coordinates so it tracks
// the tab button exactly regardless of window resize or scrolling.
//
// When visible is false, or item is nil, the rectangle is simply hidden.
// Skips layout-stale buttons whose absolute position is (0,0) but whose
// relative position is non-zero.
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
	// Skip if the absolute position is (0,0) but the button's relative
	// position is not: the button hasn't been laid out in the canvas yet.
	if absPos == (fyne.Position{}) && btn.Position() != (fyne.Position{}) {
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

// findHoveredTabItem returns the AppTabs TabItem whose button is currently
// under the mouse cursor.  It uses two strategies:
//  1. Reflect into each tab button's unexported 'hovered' field via
//     isTabButtonHovered.  This is the fast path when Fyne has already set the
//     flag.
//  2. If no button reports itself as hovered (can happen on older Fyne builds
//     or during rapid mouse movement), fall back to reading raw GLFW cursor
//     coordinates and testing them against each button's absolute position.
func (scp *ScpDesc) findHoveredTabItem() *container.TabItem {
	if scp == nil || scp.controlTab == nil {
		slog.Warn("findHoveredTabItem return nil 1")
		return nil
	}
	for _, item := range scp.controlTab.Items {
		if isTabButtonHovered(item) {
			slog.Warn("findHoveredTabItem", "item", item)
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
				// slog.Warn("findHoveredTabItem -------------------", "item.Text", item.Text)
				var absPos fyne.Position
				if app != nil && app.Driver() != nil {
					absPos = app.Driver().AbsolutePositionForObject(btn)
				} else {
					absPos = btn.Position()
				}
				size := btn.Size()
				if size.Width <= 0 || size.Height <= 0 {
					continue
				}
				// Skip if absPos is (0,0) but the button's own relative
				// position is nonzero: the widget hasn't been laid out in
				// the canvas yet, so the absolute position is stale/invalid.
				// Using it would falsely match the toolbar area.
				if absPos == (fyne.Position{}) && btn.Position() != (fyne.Position{}) {
					continue
				}
				if cursorX >= absPos.X && cursorX <= absPos.X+size.Width &&
					cursorY >= absPos.Y && cursorY <= absPos.Y+size.Height {
					slog.Warn("findHoveredTabItem", "item", item)
					return item
				}
			}
		}
	}
	slog.Warn("findHoveredTabItem return nil 2")
	return nil
}

// checkTabHoverFocus is called every 40 ms by the help monitor goroutine.
// It detects which tab button the mouse is hovering over and routes canvas
// focus to the corresponding TabFocusProxy so that the help delay timer
// starts accumulating.  When the pointer leaves all tab buttons any previously
// focused proxy is un-focused and the highlight rectangle is hidden.
func (scp *ScpDesc) checkTabHoverFocus() {
	if IsFuzzer() || scp == nil || scp.controlTab == nil || !scp.IsHelpEnabled() {
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
	// before the contextual help popup is automatically displayed.  Lowering this
	// value makes help appear more quickly; raising it reduces accidental popups.
	FocusHelpDelay = 1500 * time.Millisecond

	// FocusHelpCheckInterval specifies how frequently the focus monitor checks
	// canvas focus.  Unused directly in the poll loop (which runs at 40 ms) but
	// kept as a named constant for tests that want to override it.
	FocusHelpCheckInterval = 100 * time.Millisecond

	// widgetHelpRegistry maps a canvas object instance to its help text.
	// Written at widget construction time via RegisterWidgetHelp; read by
	// getHelpForWidget.  Access is protected by widgetHelpRegistryMtx.
	widgetHelpRegistry    = make(map[fyne.CanvasObject]WidgetHelpInfo)
	widgetHelpRegistryMtx sync.RWMutex
)

// WidgetHelpInfo stores title and detailed description text for a widget.
type WidgetHelpInfo struct {
	Title       string
	Description string
}

// RegisterWidgetHelp associates custom help information directly with a canvas object.
// Call this at widget creation time to provide context-driven help text.
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

// toggleHelp flips the contextual help feature on/off.  Bound to the help
// toolbar button so the user can suppress popups when they are not wanted.
func (scp *ScpDesc) toggleHelp() {
	scp.SetHelpEnabled(!scp.IsHelpEnabled())
}

// updateHelpButtonState synchronises the help toolbar button's visual
// importance (highlighted vs dim) with the current IsHelpEnabled() value.
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

// setContentWithHelp wraps the window content in a two-layer stack:
//
//  1. A windowMouseTracker that wraps the original content and detects when
//     the pointer leaves the application window.
//  2. A helpOverlay container that floats above everything and holds:
//     - TabFocusProxy widgets (one per tab button, zero-size, invisible).
//     - The tabFocusRect highlight rectangle.
//     - The help card popup when it is visible.
//
// Must be called once during window setup, after all tabs have been created.
func (scp *ScpDesc) setContentWithHelp(content fyne.CanvasObject) {
	if scp == nil || scp.Window == nil {
		return
	}
	scp.helpMu.Lock()
	if scp.helpOverlay == nil {
		scp.helpOverlay = container.New(&helpOverlayLayout{scp: scp})
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

// getHelpForWidget resolves the contextual help title and description for the
// given focused widget.  Resolution uses two tiers:
//
//  1. widgetHelpRegistry exact lookup – widgets that called RegisterWidgetHelp
//     at construction time have their custom text returned immediately.
//
//  2. Heuristic type-switch fallback – derives help from the concrete widget
//     type and its visible properties (button label, checkbox text, entry
//     placeholder, select options, etc.).  This covers generic Fyne widgets
//     that were not individually registered.
//
// Returns empty strings for screenRaster (the oscilloscope display area), which
// must never show a help popup.
func (scp *ScpDesc) getHelpForWidget(focused fyne.Focusable) (string, string) {
	// Suppress waveform screen rasters explicitly: they never show help popups
	switch focused.(type) {
	case *screenRaster:
		return "", ""
	}

	co, isCanvasObj := focused.(fyne.CanvasObject)
	if isCanvasObj && co != nil {
		// 1. Direct object registry lookup (set via RegisterWidgetHelp at creation time)
		widgetHelpRegistryMtx.RLock()
		if info, exists := widgetHelpRegistry[co]; exists {
			widgetHelpRegistryMtx.RUnlock()
			return info.Title, info.Description
		}
		widgetHelpRegistryMtx.RUnlock()
	}

	// 2. Heuristic fallback based on widget concrete type
	switch w := focused.(type) {
	case *FocusButton:
		return getButtonHelp(w.Text)
	case *widget.Button:
		return getButtonHelp(w.Text)
	case *TabFocusProxy:
		title := w.tabText + " Tab"
		return title, "Selects the " + w.tabText + " function tab."
	case *FocusCheck:
		return getCheckHelp(w.Text)
	case *widget.Check:
		return getCheckHelp(w.Text)
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
		lower := strings.ToLower(title)
		switch {
		case strings.Contains(lower, "cutoff"):
			return title, "Sets filter cutoff frequency threshold value."
		case strings.Contains(lower, "threshold"):
			return title, "Sets voltage threshold for logic level detection."
		case strings.Contains(lower, "hysteresis"):
			return title, "Sets noise immunity hysteresis voltage band."
		case strings.Contains(lower, "expression") || strings.Contains(lower, "math"):
			return title, "Enter mathematical formula (e.g. ChA + ChB)."
		}
		return title, "Type text or numeric values into this input field."
	case *sliderscroll.SliderScroll, *widget.Slider:
		return "Value Slider", "Drag the slider knob or scroll mouse wheel to adjust value smoothly."
	case *selectscroll.SelectScroll:
		formatTitle := func(base string) string {
			if w.Selected != "" {
				return base + " (" + w.Selected + ")"
			}
			return base
		}
		if len(w.Options) == 2 && ((w.Options[0] == "AC" && w.Options[1] == "DC") || (w.Options[0] == "DC" && w.Options[1] == "AC")) {
			return formatTitle("Input Coupling"), "Selects input coupling: AC coupling (blocks DC bias) or DC coupling."
		}
		isFreqUnit := func(s string) bool {
			s = strings.ToLower(s)
			return s == "hz" || s == "khz" || s == "mhz" || strings.Contains(s, "unit")
		}
		isFreqOptions := len(w.Options) == 3 && w.Options[0] == "Hz" && w.Options[1] == "kHz" && w.Options[2] == "MHz"

		if isFreqUnit(w.PlaceHolder) || isFreqUnit(w.Selected) || isFreqOptions {
			return formatTitle("Frequency Unit"), "Selects frequency unit (Hz, kHz, MHz) for this cutoff or parameter."
		}

		if w.PlaceHolder != "" && w.PlaceHolder != w.Selected {
			c := strings.ToLower(w.PlaceHolder)
			switch {
			case strings.Contains(c, "bit order"):
				return formatTitle("Bit Order"), "Selects serial bit transmission order: LSB First or MSB First."
			case strings.Contains(c, "parity"):
				return formatTitle("Parity"), "Selects serial communication parity bit mode: None, Even, or Odd."
			case strings.Contains(c, "protocol"):
				return formatTitle("Protocol"), "Selects serial protocol: UART, SPI, I2C, or CAN."
			case strings.Contains(c, "baud"):
				return formatTitle("Baud Rate"), "Selects communication baud rate frequency for serial decoding."
			case strings.Contains(c, "data bits"):
				return formatTitle("Data Bits"), "Selects number of payload data bits per frame."
			case strings.Contains(c, "stop bits"):
				return formatTitle("Stop Bits"), "Selects number of stop bits terminating each frame."
			default:
				return formatTitle(w.PlaceHolder), "Click, use arrow keys, or scroll mouse wheel to choose " + w.PlaceHolder + "."
			}
		}
		return formatTitle("Selection Option"), "Click, use arrow keys, or scroll mouse wheel to choose from available configuration options."
	case *widget.Select:
		return "Selection Dropdown", "Click or scroll mouse wheel to choose from available configuration options."
	default:
		return "Control", "Focused user interface control. Interact using mouse or keyboard."
	}
}

// getButtonHelp returns a (title, description) pair for a button with the
// given label text.  It recognises several well-known labels by exact or
// prefix match and returns tailored descriptions.  Unknown labels fall back
// to a generic "click or press Space" message.
func getButtonHelp(text string) (string, string) {
	title := "Button"
	if text != "" {
		title = text + " Button"
	}
	switch strings.ToLower(text) {
	case "undock":
		return "Undock Panel", "Opens this control panel in an independent floating window."
	case "apply", "accept":
		return "Apply Changes", "Applies the current configuration changes."
	case "delete":
		return "Delete Item", "Removes the selected item or configuration."
	case "new":
		return "New Item", "Creates a new item or channel configuration."
	}
	if strings.HasPrefix(text, "Ch ") {
		return text + " Channel", "Selects channel configuration parameters to display."
	}
	return title, "Click or press Space/Enter to activate this action."
}

// getCheckHelp returns a (title, description) pair for a checkbox with the
// given label text.  Labels are matched case-insensitively against a set of
// known oscilloscope control names (Enabled, X-Axis, Inv, Trig, Pers, x10,
// CMPX, filter types, decode options, …).  Unrecognised labels fall back to
// a generic "toggle on/off" message.
func getCheckHelp(text string) (string, string) {
	title := "Checkbox"
	if text != "" {
		title = text
	}
	lower := strings.ToLower(text)
	switch {
	case lower == "enabled":
		return "Channel Enable", "Enables or disables signal acquisition and waveform display for this channel."
	case lower == "x-axis":
		return "X-Axis Mode", "Selects this channel as the horizontal X-axis signal for f(v) X-Y mode."
	case lower == "inv":
		return "Invert Waveform", "Inverts the voltage polarity of the channel waveform."
	case lower == "trig":
		return "Trigger Source", "Selects this analog channel as the primary source for hardware triggering."
	case lower == "pers":
		return "Persistence", "Accumulates past waveform traces on screen to visualize noise, jitter, and intermittent events."
	case lower == "x10" || lower == "x1":
		return "10x Probe Attenuation", "Applies 10x voltage scaling for passive oscilloscope probes with attenuation."
	case lower == "cmpx":
		return "Complex Trigger", "Enables advanced complex hardware triggering conditions and qualifiers."
	case strings.Contains(lower, "zero phase"):
		return "Zero Phase Filter (FiltFilt)", "Applies forward-backward filtering to eliminate phase distortion and signal group delay."
	case strings.Contains(lower, "lowpass"):
		return "Lowpass Filter Enable", "Enables or disables lowpass filtering to attenuate high-frequency noise above the cutoff frequency."
	case strings.Contains(lower, "highpass"):
		return "Highpass Filter Enable", "Enables or disables highpass filtering to remove DC offsets and low-frequency drift."
	case strings.Contains(lower, "bandpass"):
		return "Bandpass Filter Enable", "Enables or disables bandpass filtering to pass frequencies within the specified passband range."
	case strings.Contains(lower, "bandstop"):
		return "Bandstop Filter Enable", "Enables or disables notch / bandstop filtering to reject frequencies within the stopband range."
	case strings.Contains(lower, "bit lines"):
		return "Show Bit Lines", "Displays vertical timing marker lines for decoded serial bits on the waveform display."
	case strings.Contains(lower, "active high"):
		return "CS Active High", "Configures Chip Select polarity as active-high instead of active-low."
	case strings.Contains(lower, "cpol"):
		return "Clock Polarity (CPOL)", "Configures SPI idle clock polarity: 0 for low, 1 for high."
	case strings.Contains(lower, "cpha"):
		return "Clock Phase (CPHA)", "Configures SPI clock sampling phase: 0 for leading edge, 1 for trailing edge."
	case strings.Contains(lower, "invert"):
		return "Invert Signal", "Inverts the logic polarity of the decoded input signal."
	}
	return title, "Click or press Space to toggle this option on or off."
}

// findFocusedWidget returns the currently focused Focusable widget and the
// canvas it lives on.  It checks the scope's own window first; if nothing is
// focused there it iterates over all open application windows.  Returns
// (nil, nil) when nothing is focused anywhere.
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

// startFocusHelp launches the background help monitor goroutine.  The goroutine
// ticks every 40 ms and calls:
//   - checkTabHoverFocus  – routes focus to the tab proxy under the mouse.
//   - checkFocusHelp      – decides whether to show or hide the help card.
//
// A quit channel (helpQuit) is created here and consumed by stopFocusHelp.
// Calling startFocusHelp while the monitor is already running is a no-op.
func (scp *ScpDesc) startFocusHelp() {
	if IsFuzzer() {
		return
	}
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

// stopFocusHelp signals the monitor goroutine to exit by closing helpQuit and
// hides any currently visible help popup.  Safe to call multiple times.
func (scp *ScpDesc) stopFocusHelp() {
	scp.helpMu.Lock()
	if scp.helpQuit != nil {
		close(scp.helpQuit)
		scp.helpQuit = nil
	}
	scp.helpMu.Unlock()
	scp.hideHelpPopUp()
}

// checkFocusHelp is the core of the help state machine.  Called every 40 ms
// from the monitor goroutine (always on the Fyne main goroutine via fyne.Do).
//
// State transitions:
//   - Help disabled or mouse outside window → hide popup, clear state.
//   - Focus changed to a new widget          → reset timer, hide old popup.
//   - Same widget focused ≥ FocusHelpDelay   → resolve help text, show popup.
//   - Widget is a screenRaster               → always suppress popup.
func (scp *ScpDesc) checkFocusHelp() {
	if IsFuzzer() || !scp.IsHelpEnabled() {
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

// showHelpPopUp renders and positions a floating help card near the focused
// widget.  The card contains a bold title, a separator, and a word-wrapped
// description.  Placement logic ensures the card stays within canvas bounds:
//   - Horizontally clamped to [10, canvasWidth-cardWidth-10].
//   - Vertically: shown below the widget by default; flipped above if there
//     is not enough space below.
//
// If helpOverlay is available the card is embedded as an overlay child so that
// it participates in Fyne's layout and refresh cycle without becoming a modal
// popup.  Otherwise a traditional widget.PopUp is used as a fallback.
func (scp *ScpDesc) showHelpPopUp(focused fyne.Focusable, canvas fyne.Canvas, title, desc string) {
	scp.helpMu.Lock()
	defer scp.helpMu.Unlock()

	if IsFuzzer() || !scp.IsHelpEnabled() {
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

// hideHelpPopUp removes the help card from the overlay (or hides the popup
// widget) and clears scp.helpShownFor so that the help can be re-triggered
// if the same widget is focused again later.  Safe to call when no popup is
// visible.
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
