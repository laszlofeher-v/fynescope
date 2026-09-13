package gui

import (
	"testing"
	"time"

	"fynescope/checkcolorpick"
	"fynescope/control"
	"fynescope/disp7"
	"fynescope/genericps"
	"fynescope/selectscroll"
	"fynescope/settings"

	"github.com/stretchr/testify/assert"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)


func TestHelpToggle(t *testing.T) {
	scp := &ScpDesc{
		Settings: settings.NewDefaultSettings(),
	}
	scp.helpButton = scp.newFocusButtonWithIcon("", theme.HelpIcon(), scp.toggleHelp)
	scp.updateHelpButtonState()

	if !scp.IsHelpEnabled() {
		t.Error("Expected help to be enabled by default")
	}
	if scp.helpButton.Importance != widget.HighImportance {
		t.Errorf("Expected HighImportance when help is enabled, got %v", scp.helpButton.Importance)
	}

	// Toggle OFF
	scp.toggleHelp()
	if scp.IsHelpEnabled() {
		t.Error("Expected help to be disabled after toggle")
	}
	if scp.helpButton.Importance != widget.LowImportance {
		t.Errorf("Expected LowImportance when help is disabled, got %v", scp.helpButton.Importance)
	}

	// Toggle back ON
	scp.toggleHelp()
	if !scp.IsHelpEnabled() {
		t.Error("Expected help to be enabled after second toggle")
	}
	if scp.helpButton.Importance != widget.HighImportance {
		t.Errorf("Expected HighImportance when help is enabled, got %v", scp.helpButton.Importance)
	}
}

func TestHelpContentLookup(t *testing.T) {
	scp := &ScpDesc{
		Settings: settings.NewDefaultSettings(),
	}

	// 1. Direct registration lookup
	customBtn := widget.NewButton("Custom Action", func() {})
	RegisterWidgetHelp(customBtn, "Special Action", "Executes a special test operation.")
	title, desc := scp.getHelpForWidget(customBtn)
	if title != "Special Action" || desc != "Executes a special test operation." {
		t.Errorf("Direct registration lookup failed: got (%q, %q)", title, desc)
	}

	// 2. Named control lookup (registered via addToTest)
	testBtn := widget.NewButton("RunBlock", func() {})
	addToTest(testBtn, "runblockButton", -1)
	title, desc = scp.getHelpForWidget(testBtn)
	if title != "Run / Pause" {
		t.Errorf("Named lookup failed: expected 'Run / Pause', got %q", title)
	}

	// 3. Prefix rule lookup (e.g. channel enable)
	chEnableCheck := widget.NewCheck("Enable", func(bool) {})
	addToTest(chEnableCheck, chEnableId+"_ChA", -1)
	title, desc = scp.getHelpForWidget(chEnableCheck)
	if title != "Channel Enable" {
		t.Errorf("Prefix lookup failed: expected 'Channel Enable', got %q", title)
	}

	// 4. Heuristic fallback for Button
	heuristicBtn := widget.NewButton("AutoZero", func() {})
	title, _ = scp.getHelpForWidget(heuristicBtn)
	if title != "AutoZero Button" {
		t.Errorf("Button heuristic failed: expected 'AutoZero Button', got %q", title)
	}

	// 5. Heuristic fallback for Check
	heuristicCheck := widget.NewCheck("Log Scale", func(bool) {})
	title, _ = scp.getHelpForWidget(heuristicCheck)
	if title != "Log Scale" {
		t.Errorf("Check heuristic failed: expected 'Log Scale', got %q", title)
	}

	// 6. Heuristic fallback for DigitArray
	d7 := &disp7.DigitArray{}
	title, _ = scp.getHelpForWidget(d7)
	if title != "Numeric Value Editor" {
		t.Errorf("DigitArray heuristic failed: expected 'Numeric Value Editor', got %q", title)
	}
}

func TestFocusHelpTimingAndPopup(t *testing.T) {
	origDelay := FocusHelpDelay
	origInterval := FocusHelpCheckInterval
	FocusHelpDelay = 50 * time.Millisecond
	FocusHelpCheckInterval = 10 * time.Millisecond
	defer func() {
		FocusHelpDelay = origDelay
		FocusHelpCheckInterval = origInterval
	}()

	win := test.NewWindow(nil)
	b1 := widget.NewButton("Source Channel", func() {})
	b2 := widget.NewButton("Secondary Channel", func() {})
	form := widget.NewForm(
		widget.NewFormItem("Button 1", b1),
		widget.NewFormItem("Button 2", b2),
	)

	scp := &ScpDesc{
		Window:   win,
		Settings: settings.NewDefaultSettings(),
	}
	scp.setContentWithHelp(form)

	// Focus b1
	win.Canvas().Focus(b1)
	scp.checkFocusHelp()

	// Immediately after focus, help should NOT be shown yet (< FocusHelpDelay)
	if scp.IsHelpVisible() {
		t.Error("Help should not be shown immediately upon gaining focus")
	}

	// Wait for delay to elapse
	time.Sleep(70 * time.Millisecond)
	scp.checkFocusHelp()

	if !scp.IsHelpVisible() {
		t.Fatal("Help should be visible after FocusHelpDelay has elapsed")
	}

	// Now move focus to b2: help should immediately hide
	win.Canvas().Focus(b2)
	scp.checkFocusHelp()

	if scp.IsHelpVisible() {
		t.Error("Help should be hidden when focus changes to another widget")
	}

	// Now wait again on b2: help should show for b2
	time.Sleep(70 * time.Millisecond)
	scp.checkFocusHelp()

	if !scp.IsHelpVisible() {
		t.Fatal("Help should be visible for b2 after FocusHelpDelay has elapsed")
	}

	// Turn off help: help should immediately hide
	scp.SetHelpEnabled(false)
	if scp.IsHelpVisible() {
		t.Error("Help should be hidden when help is disabled")
	}

	// While help is disabled, waiting should NOT show help
	time.Sleep(70 * time.Millisecond)
	scp.checkFocusHelp()
	if scp.IsHelpVisible() {
		t.Error("Help should not show when help is disabled")
	}

	// Lifecycle test: start and stop
	scp.SetHelpEnabled(true)
	scp.startFocusHelp()
	time.Sleep(30 * time.Millisecond)
	scp.stopFocusHelp()
	if scp.helpQuit != nil {
		t.Error("helpQuit should be nil after stopFocusHelp")
	}

}

func TestScreenRasterSuppression(t *testing.T) {
	scp := &ScpDesc{
		Settings: settings.NewDefaultSettings(),
	}

	raster := &screenRaster{}
	title, desc := scp.getHelpForWidget(raster)
	if title != "" || desc != "" {
		t.Errorf("Screenraster must not produce help, got title=%q desc=%q", title, desc)
	}

	win := test.NewWindow(nil)
	scp.Window = win
	scp.setContentWithHelp(raster)

	// Focus raster
	win.Canvas().Focus(raster)
	scp.checkFocusHelp()

	// Wait beyond FocusHelpDelay
	origDelay := FocusHelpDelay
	FocusHelpDelay = 20 * time.Millisecond
	defer func() { FocusHelpDelay = origDelay }()

	time.Sleep(30 * time.Millisecond)
	scp.checkFocusHelp()

	if scp.IsHelpVisible() {
		t.Error("Screenraster must never show help popup when focused")
	}
}

func TestWidgetsFocusAndHelp(t *testing.T) {
	origDelay := FocusHelpDelay
	FocusHelpDelay = 20 * time.Millisecond
	defer func() { FocusHelpDelay = origDelay }()

	win := test.NewWindow(nil)
	scp := &ScpDesc{
		Window:   win,
		Settings: settings.NewDefaultSettings(),
	}

	// 1. AC/DC SelectScroll
	acdc := selectscroll.NewSelectScroll([]string{"AC", "DC"}, func(string, selectscroll.Exception) {}, "AC")
	addToTest(acdc, acdcId+"_ChA", -1)

	// 2. Voltage Range SelectScroll
	vRange := selectscroll.NewSelectScroll([]string{"±1V", "±2V"}, func(string, selectscroll.Exception) {}, "±1V")
	addToTest(vRange, vRangeId+"_ChA", -1)

	// 3. Time/Div SelectScroll
	timeDiv := selectscroll.NewSelectScroll([]string{"1", "2"}, func(string, selectscroll.Exception) {}, "1")
	addToTest(timeDiv, timeSelectId, -1)

	// 4. Invert FocusCheck
	inv := scp.newFocusCheck("Inv", func(bool) {})
	addToTest(inv, invertId+"_ChA", -1)

	// 5. Trigger FocusCheck
	trig := scp.newFocusCheck("Trig", func(bool) {})
	addToTest(trig, triggerCheckId+"_ChA", -1)

	// 6. X10 / X1 FocusCheck
	x10 := scp.newFocusCheck("X10", func(bool) {})
	addToTest(x10, x10Id+"_ChA", -1)

	// 7. Cmpx FocusCheck
	cmpx := scp.newFocusCheck("Cmpx", func(bool) {})
	addToTest(cmpx, "complexTriggerCheck", -1)

	cnt := container.NewVBox(acdc, vRange, timeDiv, inv, trig, x10, cmpx)
	scp.setContentWithHelp(cnt)

	widgetsToTest := []struct {
		name          string
		widget        fyne.Focusable
		expectedTitle string
	}{
		{"AC/DC", acdc, "Input Coupling"},
		{"Voltage Range", vRange, "Voltage Scale"},
		{"Time/Div", timeDiv, "Timebase Scale"},
		{"Inv", inv, "Invert Waveform"},
		{"Trig", trig, "Trigger Source"},
		{"X10", x10, "10x Probe Attenuation"},
		{"Cmpx", cmpx, "Complex Trigger"},
	}

	for _, tc := range widgetsToTest {
		t.Run(tc.name, func(t *testing.T) {
			title, desc := scp.getHelpForWidget(tc.widget)
			if title != tc.expectedTitle {
				t.Errorf("%s: expected title %q, got %q", tc.name, tc.expectedTitle, title)
			}
			if desc == "" {
				t.Errorf("%s: expected non-empty description", tc.name)
			}

			// Test focus acquisition via MouseIn
			switch w := tc.widget.(type) {
			case *selectscroll.SelectScroll:
				w.MouseIn(&desktop.MouseEvent{})
			case *FocusCheck:
				w.MouseIn(&desktop.MouseEvent{})
			}

			if win.Canvas().Focused() != tc.widget {
				t.Errorf("%s: expected canvas to focus widget after MouseIn, got %v", tc.name, win.Canvas().Focused())
			}

			scp.checkFocusHelp()
			time.Sleep(30 * time.Millisecond)
			scp.checkFocusHelp()

			if !scp.IsHelpVisible() {
				t.Errorf("%s: help popup should be visible after delay", tc.name)
			}
			scp.hideHelpPopUp()
		})
	}
}

func TestMouseLeaveRemovesFocusAndHelp(t *testing.T) {
	origDelay := FocusHelpDelay
	FocusHelpDelay = 20 * time.Millisecond
	defer func() { FocusHelpDelay = origDelay }()

	scp := &ScpDesc{
		Settings: settings.NewDefaultSettings(),
	}

	win := test.NewWindow(nil)
	scp.Window = win

	btn := scp.newFocusButton("Test Button", func() {})
	scp.setContentWithHelp(btn)

	// Focus the button
	btn.focus()
	if win.Canvas().Focused() != btn {
		t.Fatalf("expected button to be focused, got %v", win.Canvas().Focused())
	}

	// Trigger help display
	scp.checkFocusHelp()
	time.Sleep(30 * time.Millisecond)
	scp.checkFocusHelp()

	if !scp.IsHelpVisible() {
		t.Fatal("expected help popup to be visible after delay")
	}

	// Now simulate mouse leaving the window via onMouseLeftWindow
	scp.onMouseLeftWindow()

	if win.Canvas().Focused() != nil {
		t.Errorf("expected focus to be removed after mouse leaves window, got %v", win.Canvas().Focused())
	}
	if scp.IsHelpVisible() {
		t.Error("expected help popup to be removed after mouse leaves window")
	}
	if scp.currentFocused != nil {
		t.Errorf("expected currentFocused to be nil after mouse leaves window, got %v", scp.currentFocused)
	}
	if scp.helpShownFor != nil {
		t.Errorf("expected helpShownFor to be nil after mouse leaves window, got %v", scp.helpShownFor)
	}

	// Also test windowMouseTracker.MouseOut() directly
	btn.focus()
	if win.Canvas().Focused() != btn {
		t.Fatalf("expected button to be focused again, got %v", win.Canvas().Focused())
	}

	tracker := scp.newWindowMouseTracker(win, btn)
	tracker.MouseOut()

	if win.Canvas().Focused() != nil {
		t.Errorf("expected tracker.MouseOut() to remove focus, got %v", win.Canvas().Focused())
	}
}

func TestTopLineIconsFocusAndHelp(t *testing.T) {
	origDelay := FocusHelpDelay
	FocusHelpDelay = 20 * time.Millisecond
	defer func() { FocusHelpDelay = origDelay }()

	scp := &ScpDesc{
		Settings: settings.NewDefaultSettings(),
	}
	win := test.NewWindow(nil)
	scp.Window = win

	topButtons := []struct {
		id            string
		btn           *FocusButton
		expectedTitle string
	}{
		{"runblockButton", scp.newFocusButtonWithIcon("", theme.MediaPlayIcon(), func() {}), "Run / Pause"},
		{"streamEnableButton", scp.newFocusButton(streamEnabledLabel, func() {}), "Streaming Mode"},
		{"timeZoomButton", scp.newFocusButtonWithIcon("", theme.SearchIcon(), func() {}), "Time Zoom"},
		{"recordGifButton", scp.newFocusButtonWithIcon("GIF", theme.MediaRecordIcon(), func() {}), "Record GIF"},
		{"saveRasterButton", scp.newFocusButtonWithIcon("R", theme.DocumentSaveIcon(), func() {}), "Save Raster (R)"},
		{"saveWindowButton", scp.newFocusButtonWithIcon("W", theme.DocumentSaveIcon(), func() {}), "Save Window (W)"},
		{fullScreenId, scp.newFocusButtonWithIcon("", theme.ViewFullScreenIcon(), func() {}), "Full Screen"},
		{restoreScreenId, scp.newFocusButtonWithIcon("", theme.ViewRestoreIcon(), func() {}), "Restore Window"},
		{changeSideId, scp.newFocusButtonWithIcon("", theme.NavigateBackIcon(), func() {}), "Move Controls"},
		{themeChangeActionId, scp.newFocusButtonWithIcon("", theme.CheckButtonIcon(), func() {}), "Toggle Theme"},
		{helpButtonId, scp.newFocusButtonWithIcon("", theme.HelpIcon(), func() {}), "Help Toggle (?)"},
		{"logout", scp.newFocusButtonWithIcon("", theme.LogoutIcon(), func() {}), "Disconnect & Quit"},
	}

	for _, tb := range topButtons {
		t.Run(tb.id, func(t *testing.T) {
			addToTest(tb.btn, tb.id, -1)
			scp.setContentWithHelp(tb.btn)

			title, desc := scp.getHelpForWidget(tb.btn)
			if title != tb.expectedTitle {
				t.Errorf("%s: expected title %q, got %q", tb.id, tb.expectedTitle, title)
			}
			if desc == "" {
				t.Errorf("%s: expected non-empty description", tb.id)
			}

			// Hover over the button to gain focus
			tb.btn.MouseIn(&desktop.MouseEvent{})
			if win.Canvas().Focused() != tb.btn {
				t.Errorf("%s: expected button to gain focus on MouseIn, got %v", tb.id, win.Canvas().Focused())
			}

			// Verify help is shown after delay
			scp.checkFocusHelp()
			time.Sleep(30 * time.Millisecond)
			scp.checkFocusHelp()

			if !scp.IsHelpVisible() {
				t.Errorf("%s: expected help popup to be visible", tb.id)
			}
			scp.hideHelpPopUp()
		})
	}
}

func TestTabFocusAndHelp(t *testing.T) {
	origDelay := FocusHelpDelay
	FocusHelpDelay = 20 * time.Millisecond
	defer func() { FocusHelpDelay = origDelay }()

	scp := &ScpDesc{
		Settings: settings.NewDefaultSettings(),
	}
	win := test.NewWindow(nil)
	scp.Window = win

	tabItems := []*container.TabItem{
		container.NewTabItem("f(t)", widget.NewLabel("f(t) content")),
		container.NewTabItem("f(v)", widget.NewLabel("f(v) content")),
		container.NewTabItem("FFT", widget.NewLabel("FFT content")),
		container.NewTabItem("f(f)", widget.NewLabel("f(f) content")),
		container.NewTabItem("RLC", widget.NewLabel("RLC content")),
		container.NewTabItem("digital", widget.NewLabel("digital content")),
		container.NewTabItem("gen", widget.NewLabel("gen content")),
		container.NewTabItem("extgen", widget.NewLabel("extgen content")),
		container.NewTabItem("digGen", widget.NewLabel("digGen content")),
		container.NewTabItem("vch", widget.NewLabel("vch content")),
		container.NewTabItem("decode", widget.NewLabel("decode content")),
		container.NewTabItem("filter", widget.NewLabel("filter content")),
	}

	scp.controlTab = container.NewAppTabs(tabItems...)
	scp.setContentWithHelp(scp.controlTab)

	for _, item := range tabItems {
		t.Run(item.Text, func(t *testing.T) {
			proxy := scp.getOrCreateTabProxy(item)
			assert.NotNil(t, proxy)

			title, desc := scp.getHelpForWidget(proxy)
			assert.NotEmpty(t, title, "tab %s should have non-empty title", item.Text)
			assert.NotEmpty(t, desc, "tab %s should have non-empty description", item.Text)

			// Focus the tab proxy
			scp.focusWidget(proxy)
			assert.Equal(t, proxy, win.Canvas().Focused())

			// Check help display
			scp.checkFocusHelp()
			time.Sleep(30 * time.Millisecond)
			scp.checkFocusHelp()

			assert.True(t, scp.IsHelpVisible(), "expected help popup for tab %s", item.Text)
			scp.hideHelpPopUp()

			// Test key activation (Space selects tab)
			proxy.TypedKey(&fyne.KeyEvent{Name: fyne.KeySpace})
			assert.Equal(t, item, scp.controlTab.Selected())
		})
	}
}

func makeTestChannelViewers(count int) []channelViewerDesc {
	chViewers := make([]channelViewerDesc, count)
	for i := range chViewers {
		chViewers[i] = channelViewerDesc{
			enableCheckbox:  &checkcolorpick.CheckColorPick{},
			triggerCheckbox: &widget.Check{},
			minV:            &disp7.DigitArray{},
			maxV:            &disp7.DigitArray{},
			offset:          &disp7.DigitArray{},
			frq:             &disp7.DigitArray{},
			period:          &disp7.DigitArray{},
		}
	}
	return chViewers
}

func TestFvPanelCheckboxesFocusAndHelp(t *testing.T) {
	origDelay := FocusHelpDelay
	FocusHelpDelay = 20 * time.Millisecond
	defer func() { FocusHelpDelay = origDelay }()

	reqCh := make(chan genericps.Message)
	scp := &ScpDesc{
		Settings:       settings.NewDefaultSettings(),
		channelCount:   2,
		channelViewers: makeTestChannelViewers(2),
		psControl: &control.PscDesc{
			Con: &genericps.Connection{
				MsgCh: reqCh,
				ID:    "",
			},
		},
	}
	win := test.NewWindow(nil)
	scp.Window = win

	go func() {
		for msg := range reqCh {
			if getInfoMsg, ok := msg.(*genericps.GetChannelInformationMsg); ok {
				rsp := getInfoMsg.Rsp().(*genericps.GetChannelInformationRsp)
				rsp.LengthOfRanges = 0
			}
			ch := msg.RspCh()
			if ch != nil {
				close(ch)
			}
		}
	}()
	defer close(reqCh)

	layout := container.NewVBox()
	scp.newFvPanel(layout)
	scp.setContentWithHelp(layout)

	for ch := 0; ch < int(scp.channelCount); ch++ {
		chName := channelNames[ch]

		// Enabled checkbox
		assert.Greater(t, len(scp.channelViewers[ch].enableChecks), 0)
		enableCheck := scp.channelViewers[ch].enableChecks[0]
		enableFC, ok := controls[fvEnableId+chName].Obj.(*FocusCheck)
		assert.True(t, ok, "enableCheck should be a *FocusCheck")
		assert.NotNil(t, enableCheck)

		enableFC.MouseIn(&desktop.MouseEvent{})
		assert.Equal(t, enableFC, win.Canvas().Focused())
		title, desc := scp.getHelpForWidget(enableFC)
		assert.Contains(t, title, "Channel Enable")
		assert.NotEmpty(t, desc)

		scp.checkFocusHelp()
		time.Sleep(30 * time.Millisecond)
		scp.checkFocusHelp()
		assert.True(t, scp.IsHelpVisible(), "expected help popup for fv enable %s", chName)
		scp.hideHelpPopUp()

		// X10 checkbox
		assert.Greater(t, len(scp.channelViewers[ch].x10Checkboxes), 0)
		x10FC, ok := controls[fvX10Id+chName].Obj.(*FocusCheck)
		assert.True(t, ok, "x10Check should be a *FocusCheck")

		x10FC.MouseIn(&desktop.MouseEvent{})
		assert.Equal(t, x10FC, win.Canvas().Focused())
		title, desc = scp.getHelpForWidget(x10FC)
		assert.Equal(t, "10x Probe Attenuation", title)
		assert.NotEmpty(t, desc)

		scp.checkFocusHelp()
		time.Sleep(30 * time.Millisecond)
		scp.checkFocusHelp()
		assert.True(t, scp.IsHelpVisible(), "expected help popup for fv x10 %s", chName)
		scp.hideHelpPopUp()

		// X-Axis check
		xCheckFC, ok := controls[fvXCheckId+chName].Obj.(*FocusCheck)
		assert.True(t, ok, "xCheck should be a *FocusCheck")

		xCheckFC.MouseIn(&desktop.MouseEvent{})
		assert.Equal(t, xCheckFC, win.Canvas().Focused())
		title, desc = scp.getHelpForWidget(xCheckFC)
		assert.Equal(t, "X-Axis Mode", title)
		assert.NotEmpty(t, desc)
	}
}

func TestDftPanelCheckboxesFocusAndHelp(t *testing.T) {
	origDelay := FocusHelpDelay
	FocusHelpDelay = 20 * time.Millisecond
	defer func() { FocusHelpDelay = origDelay }()

	reqCh := make(chan genericps.Message)
	scp := &ScpDesc{
		theme:          Theme(settings.DarkTheme),
		Settings:       settings.NewDefaultSettings(),
		channelCount:   2,
		channelViewers: makeTestChannelViewers(2),
		psControl: &control.PscDesc{
			Con: &genericps.Connection{
				MsgCh: reqCh,
				ID:    "",
			},
		},
	}
	win := test.NewWindow(nil)
	scp.Window = win

	go func() {
		for msg := range reqCh {
			if getInfoMsg, ok := msg.(*genericps.GetChannelInformationMsg); ok {
				rsp := getInfoMsg.Rsp().(*genericps.GetChannelInformationRsp)
				rsp.LengthOfRanges = 0
			}
			ch := msg.RspCh()
			if ch != nil {
				close(ch)
			}
		}
	}()
	defer close(reqCh)

	layout := container.NewVBox()
	scp.newDftPanel(layout)
	scp.setContentWithHelp(layout)

	for ch := 0; ch < int(scp.channelCount); ch++ {
		chName := channelNames[ch]

		// Channel enable checkbox (DFT)
		chEnableFC, ok := controls[dftEnableId+chName].Obj.(*FocusCheck)
		assert.True(t, ok, "dft enable check should be a *FocusCheck")

		chEnableFC.MouseIn(&desktop.MouseEvent{})
		assert.Equal(t, chEnableFC, win.Canvas().Focused())
		title, desc := scp.getHelpForWidget(chEnableFC)
		assert.Equal(t, "Channel Enable (FFT)", title)
		assert.NotEmpty(t, desc)

		scp.checkFocusHelp()
		time.Sleep(30 * time.Millisecond)
		scp.checkFocusHelp()
		assert.True(t, scp.IsHelpVisible(), "expected help popup for dft enable %s", chName)
		scp.hideHelpPopUp()

		// Pers checkbox (DFT)
		persFC, ok := controls[dftPersId+chName].Obj.(*FocusCheck)
		assert.True(t, ok, "persCheck should be a *FocusCheck")

		persFC.MouseIn(&desktop.MouseEvent{})
		assert.Equal(t, persFC, win.Canvas().Focused())
		title, desc = scp.getHelpForWidget(persFC)
		assert.Equal(t, "Persistence", title)
		assert.NotEmpty(t, desc)

		scp.checkFocusHelp()
		time.Sleep(30 * time.Millisecond)
		scp.checkFocusHelp()
		assert.True(t, scp.IsHelpVisible(), "expected help popup for dft pers %s", chName)
		scp.hideHelpPopUp()

		// X10 checkbox (DFT)
		x10FC, ok := controls[dftX10Id+chName].Obj.(*FocusCheck)
		assert.True(t, ok, "dft x10Check should be a *FocusCheck")

		x10FC.MouseIn(&desktop.MouseEvent{})
		assert.Equal(t, x10FC, win.Canvas().Focused())
		title, desc = scp.getHelpForWidget(x10FC)
		assert.Equal(t, "10x Probe Attenuation", title)
		assert.NotEmpty(t, desc)

		scp.checkFocusHelp()
		time.Sleep(30 * time.Millisecond)
		scp.checkFocusHelp()
		assert.True(t, scp.IsHelpVisible(), "expected help popup for dft x10 %s", chName)
		scp.hideHelpPopUp()
	}
}


