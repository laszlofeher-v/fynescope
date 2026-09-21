package gui

import (
	"fmt"
	"image/color"

	"fynescope/checkcolorpick"
	"fynescope/control"
	"fynescope/disp7"
	"fynescope/genericps"
	"fynescope/selectscroll"
	"fynescope/settings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type framelessEntry struct {
	widget.Entry
}

func newFramelessEntry() *framelessEntry {
	e := &framelessEntry{}
	e.ExtendBaseWidget(e)
	return e
}

type tappableDnLabel struct {
	widget.Label
	onTapped func()
	negated  bool
	focused  bool
	scp      *ScpDesc
}

func newTappableDnLabel(scp *ScpDesc, text string, negated bool, tapped func()) *tappableDnLabel {
	l := &tappableDnLabel{onTapped: tapped, negated: negated, scp: scp}
	l.Text = text
	l.ExtendBaseWidget(l)
	return l
}

func (l *tappableDnLabel) Tapped(e *fyne.PointEvent) {
	if l.scp != nil {
		if l.scp.Window != nil && l.scp.Window.Canvas() != nil {
			l.scp.Window.Canvas().Focus(l)
		}
	}
	if l.onTapped != nil {
		l.onTapped()
	}
}

func (l *tappableDnLabel) TappedSecondary(e *fyne.PointEvent) {}

func (l *tappableDnLabel) setNegated(neg bool) {
	l.negated = neg
	l.Refresh()
}

func (l *tappableDnLabel) FocusGained() {
	l.focused = true
	l.Refresh()
}

func (l *tappableDnLabel) FocusLost() {
	l.focused = false
	l.Refresh()
}

func (l *tappableDnLabel) TypedRune(rune) {}

func (l *tappableDnLabel) TypedKey(e *fyne.KeyEvent) {
	if e.Name == fyne.KeySpace || e.Name == fyne.KeyReturn {
		l.Tapped(nil)
	}
}

func (l *tappableDnLabel) MouseIn(e *desktop.MouseEvent) {
	if l.scp != nil {
		if l.scp.Window != nil && l.scp.Window.Canvas() != nil {
			l.scp.Window.Canvas().Focus(l)
		}
	}
}

func (l *tappableDnLabel) MouseMoved(e *desktop.MouseEvent) {
	if l.scp != nil {
		if l.scp.Window != nil && l.scp.Window.Canvas() != nil {
			l.scp.Window.Canvas().Focus(l)
		}
	}
}

func (l *tappableDnLabel) MouseOut() {}

func (l *tappableDnLabel) CreateRenderer() fyne.WidgetRenderer {
	r := l.Label.CreateRenderer()
	line := canvas.NewLine(theme.ForegroundColor())
	line.StrokeWidth = 1
	if !l.negated {
		line.Hidden = true
	}
	focusBg := canvas.NewRectangle(theme.FocusColor())
	focusBg.Hidden = true
	return &tappableDnLabelRenderer{WidgetRenderer: r, label: l, line: line, focusBg: focusBg}
}

type tappableDnLabelRenderer struct {
	fyne.WidgetRenderer
	label   *tappableDnLabel
	line    *canvas.Line
	focusBg *canvas.Rectangle
}

func (r *tappableDnLabelRenderer) Layout(size fyne.Size) {
	r.WidgetRenderer.Layout(size)
	r.line.Position1 = fyne.NewPos(0, 2)
	r.line.Position2 = fyne.NewPos(size.Width, 2)
	r.focusBg.Resize(size)
}

func (r *tappableDnLabelRenderer) MinSize() fyne.Size {
	return r.WidgetRenderer.MinSize()
}

func (r *tappableDnLabelRenderer) Refresh() {
	r.line.StrokeColor = theme.ForegroundColor()
	r.line.Hidden = !r.label.negated
	r.focusBg.Hidden = !r.label.focused
	r.focusBg.FillColor = theme.FocusColor()
	r.WidgetRenderer.Refresh()
}

func (r *tappableDnLabelRenderer) Objects() []fyne.CanvasObject {
	baseObjs := r.WidgetRenderer.Objects()
	objs := make([]fyne.CanvasObject, len(baseObjs)+2)
	objs[0] = r.focusBg
	copy(objs[1:], baseObjs)
	objs[len(objs)-1] = r.line
	return objs
}

func (r *tappableDnLabelRenderer) Destroy() {
	r.WidgetRenderer.Destroy()
}

func (e *framelessEntry) MinSize() fyne.Size {
	s := e.Entry.MinSize()
	s.Width = fyne.MeasureText("WWWWWW", theme.TextSize(), fyne.TextStyle{}).Width
	return s
}

func (e *framelessEntry) CreateRenderer() fyne.WidgetRenderer {
	r := e.Entry.CreateRenderer()
	fr := &framelessEntryRenderer{WidgetRenderer: r}
	fr.Refresh()
	return fr
}

type framelessEntryRenderer struct {
	fyne.WidgetRenderer
}

func (fr *framelessEntryRenderer) Refresh() {
	fr.WidgetRenderer.Refresh()
	for _, obj := range fr.WidgetRenderer.Objects() {
		if rect, ok := obj.(*canvas.Rectangle); ok {
			rect.FillColor = color.Transparent
			rect.StrokeColor = color.Transparent
			rect.StrokeWidth = 0
		}
		if line, ok := obj.(*canvas.Line); ok {
			line.StrokeColor = color.Transparent
			line.StrokeWidth = 0
		}
	}
}

const (
	up        = "↑"
	down      = "↓"
	upDown    = "↑↓"
	digitalDc = "X"
	low       = "L"
	high      = "H"
)

func (scp *ScpDesc) updateDigitalTrigger() {
	var dirs []genericps.DigitalChannelDirections
	for i := 0; i < 16; i++ {
		portIdx := i / 8
		if !scp.Settings.Digital.ChannelsEnabled[i] || !scp.Settings.Digital.Ports[portIdx].Enabled {
			continue
		}
		d := scp.Settings.Digital.Trigger.Directions[i]
		if d != genericps.DigitalDontCare {
			dirs = append(dirs, genericps.DigitalChannelDirections{
				Channel:   genericps.DigitalChannel(i),
				Direction: d,
			})
		}
	}
	scp.triggerSettingMsg.DigitalTriggerEnabled = scp.Settings.Digital.Trigger.Enabled
	scp.triggerSettingMsg.DigitalDirections = dirs

	operand := scp.Settings.Digital.Trigger.Operand
	if operand == genericps.OperandNone {
		if scp.Settings.Digital.Trigger.AnalogLogic == "AND" || scp.Settings.Digital.Trigger.Logic == "AND" {
			operand = genericps.OperandAnd
		} else {
			operand = genericps.OperandOr
		}
	}
	scp.triggerSettingMsg.DigitalAnalogOperand = operand

	channelsOperand := genericps.OperandAnd
	if scp.Settings.Digital.Trigger.ChannelsLogic == "OR" {
		channelsOperand = genericps.OperandOr
	}
	scp.triggerSettingMsg.DigitalChannelsOperand = channelsOperand

	triggerCopy := scp.triggerSettingMsg
	triggerCopy.Done = make(chan struct{}, 1)
	if scp.psControl != nil && scp.psControl.SetTriggerCh != nil {
		go func(t control.TriggerDescMsg) {
			scp.psControl.SetTriggerCh <- &t
			<-t.Done
		}(triggerCopy)
	}
}

func (scp *ScpDesc) buildDigitalPortContent(undockable bool) fyne.CanvasObject {
	if genericps.DigitalDirectionRising == 0 && genericps.DigitalDontCare == 0 {
		genericps.DigitalDontCare = 0
		genericps.DigitalDirectionLow = 1
		genericps.DigitalDirectionHigh = 2
		genericps.DigitalDirectionRising = 3
		genericps.DigitalDirectionFalling = 4
		genericps.DigitalDirectionRisingOrFalling = 5
	}
	dirOptions := []string{digitalDc, low, high, up, down, upDown}
	dirMap := map[string]genericps.DigitalDirection{
		digitalDc: genericps.DigitalDontCare,
		low:       genericps.DigitalDirectionLow,
		high:      genericps.DigitalDirectionHigh,
		up:        genericps.DigitalDirectionRising,
		down:      genericps.DigitalDirectionFalling,
		upDown:    genericps.DigitalDirectionRisingOrFalling,
	}
	dirReverseMap := map[genericps.DigitalDirection]string{
		genericps.DigitalDontCare:                 digitalDc,
		genericps.DigitalDirectionLow:             low,
		genericps.DigitalDirectionHigh:            high,
		genericps.DigitalDirectionRising:          up,
		genericps.DigitalDirectionFalling:         down,
		genericps.DigitalDirectionRisingOrFalling: upDown,
	}

	mainBox := container.NewVBox()

	var undockBtn *widget.Button
	if undockable {
		undockBtn = widget.NewButtonWithIcon("Undock", theme.ViewFullScreenIcon(), func() {
			if scp.digPortWindow != nil {
				scp.digPortWindow.RequestFocus()
				return
			}
			onWindowClose := func() {
				scp.digPortWindow = nil
				scp.digPortLayout = container.NewVBox(scp.buildDigitalPortContent(true))
				scp.digPortTab.Content = scp.digPortLayout
				scp.dockTab(scp.digPortTab)
				scp.controlTab.SelectIndex(ftTabIndex)
				fyne.Do(scp.digPortTab.Content.Refresh)
			}
			scp.digPortWindow = scp.App.NewWindow("Digital Port")
			winContent := scp.buildDigitalPortContent(false)
			scp.controlTab.Remove(scp.digPortTab)
			scp.digPortWindow.SetContent(winContent)
			scp.digPortWindow.SetOnClosed(onWindowClose)
			scp.digPortWindow.Resize(fyne.NewSize(350, 700))
			scp.controlTab.SelectIndex(ftTabIndex)
			scp.digPortWindow.Show()
			fyne.Do(winContent.Refresh)
		})
	}

	title := NewFocusableLabelWithStyle("Digital Channels", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	trigEnableCheck := widget.NewCheck("Enable Digital Trigger", func(v bool) {
		scp.Settings.Digital.Trigger.Enabled = v
		scp.SaveSettings()
		scp.updateDigitalTrigger()
	})
	trigEnableCheck.SetChecked(scp.Settings.Digital.Trigger.Enabled)

	channelsLogicOptions := []string{"AND", "OR"}
	initialChannelsLogic := "AND"
	if scp.Settings.Digital.Trigger.ChannelsLogic == "OR" {
		initialChannelsLogic = "OR"
	}
	channelsLogicSelect := selectscroll.NewSelectScroll(channelsLogicOptions, func(sel string, ex selectscroll.Exception) {
		scp.Settings.Digital.Trigger.ChannelsLogic = sel
		scp.SaveSettings()
		scp.updateDigitalTrigger()
	}, "AND")
	channelsLogicSelect.SetSelected(initialChannelsLogic)

	operandOptions := []string{"OR", "AND"}
	initialOperand := "OR"
	if scp.Settings.Digital.Trigger.Operand == genericps.OperandAnd || scp.Settings.Digital.Trigger.AnalogLogic == "AND" || scp.Settings.Digital.Trigger.Logic == "AND" {
		initialOperand = "AND"
	}
	operandSelect := selectscroll.NewSelectScroll(operandOptions, func(sel string, ex selectscroll.Exception) {
		if sel == "AND" {
			scp.Settings.Digital.Trigger.Operand = genericps.OperandAnd
			scp.Settings.Digital.Trigger.AnalogLogic = "AND"
			scp.Settings.Digital.Trigger.Logic = "AND"
		} else {
			scp.Settings.Digital.Trigger.Operand = genericps.OperandOr
			scp.Settings.Digital.Trigger.AnalogLogic = "OR"
			scp.Settings.Digital.Trigger.Logic = "OR"
		}
		scp.SaveSettings()
		scp.updateDigitalTrigger()
	}, "AND")
	operandSelect.SetSelected(initialOperand)

	trigHeader := container.NewVBox(
		trigEnableCheck,
		layout.NewSpacer(),
		NewFocusableLabel("Channels:"),
		channelsLogicSelect,
		NewFocusableLabel("Analog/Digital:"),
		operandSelect,
	)

	if undockable && undockBtn != nil {
		addToTest(undockBtn, "digPortUndockBtn", digPortTabIndex)
		mainBox.Add(container.NewVBox(undockBtn, layout.NewSpacer(), title, trigHeader))
	} else {
		mainBox.Add(container.NewVBox(title, trigHeader))
	}
	addToTest(trigEnableCheck, "digPortTrigEnable", digPortTabIndex)
	addToTest(channelsLogicSelect, "digPortChannelsLogicSelect", digPortTabIndex)
	addToTest(operandSelect, "digPortOperandSelect", digPortTabIndex)
	port0Box := container.NewVBox()
	port1Box := container.NewVBox()

	port0EnableCheck := widget.NewCheck("Enable Port 0 (D0-D7)", func(v bool) {
		scp.Settings.Digital.Ports[0].Enabled = v
		scp.SaveSettings()
		scp.updateDigitalSplit()
		scp.updateDigitalTrigger()
		scp.updateTriggerModeOptions()
		if scp.digitalRaster != nil {
			scp.digitalRaster.refresh()
		}
		if scp.psControl != nil {
			go func(p settings.DigitalPortSettings) {
				scp.psControl.SetDigitalPortCh <- &control.DigitalPortMsg{Port: genericps.Port0, Settings: p}
			}(scp.Settings.Digital.Ports[0])
		}
		if scp.runningMode == genericps.DemoMode {
			if v {
				if scp.Settings.DigitalDemoGenPanel.Frequency <= 0 {
					scp.Settings.DigitalDemoGenPanel.Frequency = 1000
				}
				scp.Settings.DigitalDemoGenPanel.Port0Enabled = true
			} else {
				scp.Settings.DigitalDemoGenPanel.Port0Enabled = false
			}
			scp.applyDemoDigitalGenSettings()
		}
	})
	port0EnableCheck.SetChecked(scp.Settings.Digital.Ports[0].Enabled)
	addToTest(port0EnableCheck, "digPort0EnableCheck", digPortTabIndex)
	port0Box.Add(port0EnableCheck)

	port0ChannelsBox := container.NewVBox()

	fontScale := float32(0.7) * scp.getScreenScale()
	dispColor := theme.ForegroundColor()
	if scp.theme != nil {
		dispColor = scp.theme.Color(ColorNameGeneratorDisp, 0)
	}

	port0LogicDisp, err0 := disp7.NewCustomDisp7Array(4, 3, 5000, -5000,
		disp7.Signed, disp7.NoTrailingZeroes, scp.Window,
		dispColor, disp7.ReadWrite,
		fontScale*disp7.DefaultDigitWidth, fontScale*disp7.DeafultDigitHeight,
		1, disp7.DefaultVCursorSpace, "Threshold: ", " V")
	if err0 == nil {
		port0LogicDisp.SilentSetValue(int(float64(scp.Settings.Digital.Ports[0].Threshold) * 5000.0 / 32767.0))
		port0LogicDisp.OnChanged = func(val float64) {
			scp.Settings.Digital.Ports[0].Threshold = int16(val * 32767.0 / 5000.0)
			scp.SaveSettings()
			if scp.psControl != nil {
				go func(p settings.DigitalPortSettings) {
					scp.psControl.SetDigitalPortCh <- &control.DigitalPortMsg{Port: genericps.Port0, Settings: p}
				}(scp.Settings.Digital.Ports[0])
			}
		}
		addToTest(port0LogicDisp, "digPort0LogicLevelDisp", digPortTabIndex)
		scp.digPortLogicLevelDisp[0] = port0LogicDisp
		if scp.runningMode != genericps.DemoMode {
			port0Box.Add(container.NewHBox(port0LogicDisp))
		}
	}
	port0Box.Add(port0ChannelsBox)

	port1EnableCheck := widget.NewCheck("Enable Port 1 (D8-D15)", func(v bool) {
		scp.Settings.Digital.Ports[1].Enabled = v
		scp.SaveSettings()
		scp.updateDigitalSplit()
		scp.updateDigitalTrigger()
		scp.updateTriggerModeOptions()
		if scp.digitalRaster != nil {
			scp.digitalRaster.refresh()
		}
		if scp.psControl != nil {
			go func(p settings.DigitalPortSettings) {
				scp.psControl.SetDigitalPortCh <- &control.DigitalPortMsg{Port: genericps.Port1, Settings: p}
			}(scp.Settings.Digital.Ports[1])
		}
		if scp.runningMode == genericps.DemoMode {
			if v {
				if scp.Settings.DigitalDemoGenPanel.Frequency <= 0 {
					scp.Settings.DigitalDemoGenPanel.Frequency = 1000
				}
				scp.Settings.DigitalDemoGenPanel.Port1Enabled = true
			} else {
				scp.Settings.DigitalDemoGenPanel.Port1Enabled = false
			}
			scp.applyDemoDigitalGenSettings()
		}
	})
	port1EnableCheck.SetChecked(scp.Settings.Digital.Ports[1].Enabled)
	addToTest(port1EnableCheck, "digPort1EnableCheck", digPortTabIndex)
	port1Box.Add(port1EnableCheck)

	port1ChannelsBox := container.NewVBox()

	port1LogicDisp, err1 := disp7.NewCustomDisp7Array(4, 3, 5000, -5000,
		disp7.Signed, disp7.NoTrailingZeroes, scp.Window,
		dispColor, disp7.ReadWrite,
		fontScale*disp7.DefaultDigitWidth, fontScale*disp7.DeafultDigitHeight,
		1, disp7.DefaultVCursorSpace, "Logic Level: ", " V")
	if err1 == nil {
		port1LogicDisp.SilentSetValue(int(float64(scp.Settings.Digital.Ports[1].Threshold) * 5000.0 / 32767.0))
		port1LogicDisp.OnChanged = func(val float64) {
			scp.Settings.Digital.Ports[1].Threshold = int16(val * 32767.0 / 5000.0)
			scp.SaveSettings()
			if scp.psControl != nil {
				go func(p settings.DigitalPortSettings) {
					scp.psControl.SetDigitalPortCh <- &control.DigitalPortMsg{Port: genericps.Port1, Settings: p}
				}(scp.Settings.Digital.Ports[1])
			}
		}
		addToTest(port1LogicDisp, "digPort1LogicLevelDisp", digPortTabIndex)
		scp.digPortLogicLevelDisp[1] = port1LogicDisp
		if scp.runningMode != genericps.DemoMode {
			port1Box.Add(container.NewHBox(port1LogicDisp))
		}
	}
	port1Box.Add(port1ChannelsBox)

	var port0Rows []fyne.CanvasObject
	var port1Rows []fyne.CanvasObject
	var trigSelects [16]*selectscroll.SelectScroll

	for i := 0; i < 16; i++ {
		chIdx := i
		dn := fmt.Sprintf("D%d", chIdx)

		var dnLabel *tappableDnLabel
		dnLabel = newTappableDnLabel(scp, dn, scp.Settings.Digital.ChannelNegated[chIdx], func() {
			scp.Settings.Digital.ChannelNegated[chIdx] = !scp.Settings.Digital.ChannelNegated[chIdx]
			dnLabel.setNegated(scp.Settings.Digital.ChannelNegated[chIdx])
			scp.SaveSettings()
			if scp.digitalRaster != nil {
				scp.digitalRaster.refresh()
			}
		})
		addToTest(dnLabel, fmt.Sprintf("digPortDnLabel_%d", chIdx), digPortTabIndex)
		// 1. Editable label (max 6 chars, frameless)
		labelEntry := newFramelessEntry()
		labelEntry.SetPlaceHolder("Lbl")
		labelEntry.SetText(scp.Settings.Digital.ChannelLabels[chIdx])
		labelEntry.OnChanged = func(s string) {
			runes := []rune(s)
			if len(runes) > 6 {
				s = string(runes[:6])
				labelEntry.SetText(s)
			}
			scp.Settings.Digital.ChannelLabels[chIdx] = s
			scp.SaveSettings()
			if scp.digitalRaster != nil {
				scp.digitalRaster.refresh()
			}
		}
		addToTest(labelEntry, fmt.Sprintf("digPortLabelEntry_%d", chIdx), digPortTabIndex)
		// 2. Color & Enable picker
		col := scp.Settings.Digital.ChannelColors[chIdx]
		ccp := checkcolorpick.NewCheckColorPick(scp.Window, func(v bool, c color.Color) {
			nrgba := color.NRGBAModel.Convert(c).(color.NRGBA)
			scp.Settings.Digital.ChannelColors[chIdx] = nrgba
			scp.Settings.Digital.ChannelsEnabled[chIdx] = v
			if scp.digitalRaster != nil {
				scp.digitalRaster.refresh()
			}
			scp.SaveSettings()
			scp.updateDigitalTrigger()
		}, col, fyne.NewSize(20, 20))
		ccp.SetVal(scp.Settings.Digital.ChannelsEnabled[chIdx])
		addToTest(ccp, fmt.Sprintf("digPortCheckColorPick_%d", chIdx), digPortTabIndex)
		// 3. Trigger mode
		initialDirStr := dirReverseMap[scp.Settings.Digital.Trigger.Directions[chIdx]]
		if initialDirStr == "" {
			initialDirStr = digitalDc
		}
		trigSelect := selectscroll.NewSelectScroll(dirOptions, func(sel string, ex selectscroll.Exception) {
			if dirVal, ok := dirMap[sel]; ok {
				scp.Settings.Digital.Trigger.Directions[chIdx] = dirVal

				if sel == up || sel == down || sel == upDown {
					for j := 0; j < 16; j++ {
						if j != chIdx {
							d := scp.Settings.Digital.Trigger.Directions[j]
							if d == genericps.DigitalDirectionRising || d == genericps.DigitalDirectionFalling || d == genericps.DigitalDirectionRisingOrFalling {
								scp.Settings.Digital.Trigger.Directions[j] = genericps.DigitalDontCare
								if trigSelects[j] != nil && trigSelects[j].Selected != digitalDc {
									trigSelects[j].SetSelected(digitalDc)
								}
							}
						}
					}
				}

				scp.SaveSettings()
				scp.updateDigitalTrigger()
			}
		}, upDown)
		trigSelects[chIdx] = trigSelect
		trigSelect.SetSelected(initialDirStr)
		addToTest(trigSelect, fmt.Sprintf("digPortTrigSelect_%d", chIdx), digPortTabIndex)
		negCheck := widget.NewCheck("Neg", func(v bool) {
			scp.Settings.Digital.LabelNegated[chIdx] = v
			if scp.digitalRaster != nil {
				scp.digitalRaster.refresh()
			}
			scp.SaveSettings()
		})
		negCheck.SetChecked(scp.Settings.Digital.LabelNegated[chIdx])
		addToTest(negCheck, fmt.Sprintf("digPortNegCheck_%d", chIdx), digPortTabIndex)
		row := container.NewHBox(
			dnLabel,
			labelEntry,
			negCheck,
			container.NewCenter(ccp),
			NewFocusableLabel("Trig:"),
			trigSelect,
		)

		if i < 8 {
			port0Rows = append(port0Rows, row)
		} else {
			port1Rows = append(port1Rows, row)
		}
	}

	updateStack := func() {
		port0ChannelsBox.Objects = nil
		port1ChannelsBox.Objects = nil
		if scp.Settings.Digital.D0AtBottom {
			for i := len(port0Rows) - 1; i >= 0; i-- {
				port0ChannelsBox.Add(port0Rows[i])
			}
			for i := len(port1Rows) - 1; i >= 0; i-- {
				port1ChannelsBox.Add(port1Rows[i])
			}
		} else {
			for i := 0; i < len(port0Rows); i++ {
				port0ChannelsBox.Add(port0Rows[i])
			}
			for i := 0; i < len(port1Rows); i++ {
				port1ChannelsBox.Add(port1Rows[i])
			}
		}
		port0ChannelsBox.Refresh()
		port1ChannelsBox.Refresh()
	}
	updateStack()

	d0BottomCheck := widget.NewCheck("D0 at Bottom", func(v bool) {
		scp.Settings.Digital.D0AtBottom = v
		scp.SaveSettings()
		updateStack()
		if scp.digitalRaster != nil {
			scp.digitalRaster.refresh()
		}
	})
	d0BottomCheck.SetChecked(scp.Settings.Digital.D0AtBottom)
	addToTest(d0BottomCheck, "digPortD0BottomCheck", digPortTabIndex)
	portTabs := container.NewAppTabs(
		container.NewTabItem("Port 0 (D0-D7)", port0Box),
		container.NewTabItem("Port 1 (D8-D15)", port1Box),
	)
	addToTest(portTabs, "digPortSubTabs", digPortTabIndex)
	mainBox.Add(d0BottomCheck)
	mainBox.Add(portTabs)

	return container.NewVBox(mainBox, layout.NewSpacer())
}
