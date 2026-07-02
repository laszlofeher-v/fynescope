package gui

import (
	"fynescope/checkcolorpick"
	"fynescope/disp7"
	"fynescope/genericps"
	"fynescope/selectscroll"
	"fynescope/settings"
	"image/color"
	"log/slog"
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

const (
	checkColorPickMinSize = 24
)

type (
	ipModeType int
)

type (
	channelViewerDesc struct {
		displayOffsetInt         int
		displayOffsetFraction    float64
		dftDisplayOffsetInt      int
		dftDisplayOffsetFraction float64
		ffDisplayOffsetFraction  float64
		label                    ftChannelLabelViewer
		dftLabel                 dftChannelLabelViewer
		leftLabel                bool
		hasScreenPartition       bool
		enableCheckbox           *checkcolorpick.CheckColorPick
		enableChecks             []*widget.Check
		dftCheckbox              *widget.Check
		triggerCheckbox          *widget.Check
		persistenceCheckbox      *widget.Check
		dftPersistenceCheckbox   *widget.Check
		minV                     *disp7.DigitArray
		maxV                     *disp7.DigitArray
		offset                   *disp7.DigitArray
		frq                      *disp7.DigitArray
		period                   *disp7.DigitArray
		vRangeSelects            []*selectscroll.SelectScroll
		x10Checkboxes            []*widget.Check
		fvNameLabel              *canvas.Text
		dftNameLabel             *canvas.Text
		ffNameLabel              *canvas.Text
		rlcNameLabel             *canvas.Text
		simGenNameLabel          *canvas.Text
		filterWarning            *canvas.Text
		simGenDisplays           []*disp7.DigitArray
	}
)

const (
	raising = "Raising"
	failing = "Falling"
	ac      = "AC"
	dc      = "DC"
)

var (
	channelNames            = []string{"A", "B", "C", "D"}
	triggerDirectionOptions = []string{raising, failing}
	triggerDirections       map[string]genericps.ThresholdDirection
	triggerDirectionNames   map[genericps.ThresholdDirection]string
	coupleTypeNames         = []string{ac, dc}
	coupleTypes             map[string]genericps.Coupling
	vRanges                 map[string]genericps.RangeEnum
	x10vRanges              map[string]genericps.RangeEnum
	inputRanges             []string
	x10InputRanges          []string
	rangeEnumToString       map[genericps.RangeEnum]string
)

func initMaps() {
	coupleTypes = make(map[string]genericps.Coupling)
	coupleTypes[ac] = genericps.Ac
	coupleTypes[dc] = genericps.Dc
	triggerDirectionNames = make(map[genericps.ThresholdDirection]string)
	triggerDirectionNames[genericps.TriggerFalling] = failing
	triggerDirectionNames[genericps.TriggerRaising] = raising
	triggerDirections = make(map[string]genericps.ThresholdDirection)
	triggerDirections[failing] = genericps.TriggerFalling
	triggerDirections[raising] = genericps.TriggerRaising
}
func sortInputRanges() {
	vRanges = map[string]genericps.RangeEnum{
		"±10mV":  (genericps.Range_10mv),
		"±20mV":  genericps.Range_20mv,
		"±50mV":  genericps.Range_50mv,
		"±100mV": genericps.Range_100mv,
		"±200mV": genericps.Range_200mv,
		"±500mV": genericps.Range_500mv,
		"±1V":    genericps.Range_1v,
		"±2V":    genericps.Range_2v,
		"±5V":    genericps.Range_5v,
		"±10V":   genericps.Range_10v,
		"±20V":   genericps.Range_20v,
		"±50V":   (genericps.Range_50v),
	}
	x10vRanges = map[string]genericps.RangeEnum{
		"±100mV": genericps.Range_10mv,
		"±200mV": genericps.Range_20mv,
		"±500mV": genericps.Range_50mv,
		"±1V":    genericps.Range_100mv,
		"±2V":    genericps.Range_200mv,
		"±5V":    genericps.Range_500mv,
		"±10V":   genericps.Range_1v,
		"±20V":   genericps.Range_2v,
		"±50V":   (genericps.Range_5v),
	}
	inputRanges = sortMapString(vRanges)
	rangeEnumToString = make(map[genericps.RangeEnum]string)
	for s, enum := range vRanges {
		rangeEnumToString[enum] = s
	}
	slog.Debug("sortInputRanges", "inputRanges", inputRanges)
}

func sortMapString(strInt map[string]genericps.RangeEnum) []string {
	type keyValDesc struct {
		key string
		val genericps.RangeEnum
	}
	var keyVal []keyValDesc
	for key, val := range strInt {
		keyVal = append(keyVal, keyValDesc{key, val})
	}
	sort.Slice(keyVal, func(i, j int) bool {
		return keyVal[i].val > keyVal[j].val
	})
	sorted := make([]string, len(strInt))
	for i, kv := range keyVal {
		sorted[i] = kv.key
	}
	return sorted
}

func (scp *ScpDesc) numberOfEnabledChannels() (n int, set uint64) {
	n = 0
	pos := uint64(1)
	set = 0
	for i := 0; i < int(scp.channelCount); i++ {
		channel := scp.Settings.Channels[i]
		if channel.Enabled {
			set |= pos
			n++
		}
		pos += pos
	}
	return
}

func (scp *ScpDesc) nthEnabledChannels(n int) (ch int) {
	ch = 0
	for i := 0; i < int(scp.channelCount); i++ {
		channel := scp.Settings.Channels[i]
		if channel.Enabled {
			if n == 0 {
				ch = i
				return
			}
			n--
		}
	}
	ch = -1
	return
}

func RemoveIndex(s []string, index int) []string {
	return append(s[:index], s[index+1:]...)
}

func (scp *ScpDesc) isDigitalFilterEnabled(chIndex genericps.ChannelId) bool {
	if int(chIndex) >= len(scp.Settings.Channels) {
		return false
	}
	f := &scp.Settings.Channels[chIndex].DigitalFilter
	return f.LowpassEnabled || f.HighpassEnabled || f.BandpassEnabled || f.BandstopEnabled
}

func (scp *ScpDesc) refreshFilterWarning(chIndex genericps.ChannelId) {
	if int(chIndex) >= len(scp.channelViewers) {
		return
	}
	cv := &scp.channelViewers[chIndex]
	filtered := scp.isDigitalFilterEnabled(chIndex)

	if cv.filterWarning != nil {
		if filtered {
			cv.filterWarning.Show()
		} else {
			cv.filterWarning.Hide()
		}
		cv.filterWarning.Refresh()
	}

	chName := channelNames[chIndex]
	textBase := "Ch " + chName + ":"
	warnText := textBase + " ⚠️"

	if cv.fvNameLabel != nil {
		if filtered {
			cv.fvNameLabel.Text = warnText
		} else {
			cv.fvNameLabel.Text = textBase
		}
		cv.fvNameLabel.Refresh()
	}
	if cv.dftNameLabel != nil {
		if filtered {
			cv.dftNameLabel.Text = warnText
		} else {
			cv.dftNameLabel.Text = textBase
		}
		cv.dftNameLabel.Refresh()
	}
	if cv.ffNameLabel != nil {
		if filtered {
			cv.ffNameLabel.Text = warnText
		} else {
			cv.ffNameLabel.Text = textBase
		}
		cv.ffNameLabel.Refresh()
	}
	if cv.rlcNameLabel != nil {
		if filtered {
			cv.rlcNameLabel.Text = warnText
		} else {
			cv.rlcNameLabel.Text = textBase
		}
		cv.rlcNameLabel.Refresh()
	}
}

func (scp *ScpDesc) frqPeriodDisp(chIndex genericps.ChannelId) (
	frqPeriodBox *fyne.Container) {
	var err error
	const fontScale = 0.7
	scp.channelViewers[chIndex].frq, err = disp7.NewCustomDisp7Array(4, 0,
		maxFrqDisp, 0, disp7.UnSigned, disp7.NoTrailingZeroes,
		scp.Window, scp.Settings.Channels[chIndex].Col[scp.Settings.ChannelColorIndex],
		disp7.ReaOnly, fontScale*disp7.DefaultDigitWidth,
		fontScale*disp7.DeafultDigitHeight, 1,
		fontScale*disp7.DefaultVCursorSpace, "Frq:", " MHz")
	if err != nil {
		panic("error from disp7.NewCustomDisp7Array")
	}
	scp.channelViewers[chIndex].period, err = disp7.NewCustomDisp7Array(4, 0,
		maxPeriodDisp, 0, disp7.UnSigned, disp7.NoTrailingZeroes,
		scp.Window, scp.Settings.Channels[chIndex].Col[scp.Settings.ChannelColorIndex],
		disp7.ReaOnly, fontScale*disp7.DefaultDigitWidth,
		fontScale*disp7.DeafultDigitHeight, 1, fontScale*disp7.DefaultVCursorSpace, "  T:", " ms")
	if err != nil {
		panic("error from disp7.NewCustomDisp7Array")
	}
	scp.channelViewers[chIndex].period.SetValue(0)
	frqPeriodBox = container.New(layout.NewVBoxLayout(),
		scp.channelViewers[chIndex].frq, scp.channelViewers[chIndex].period)
	return
}

func (scp *ScpDesc) minMaxDisp(chIndex genericps.ChannelId) (
	vfBox *fyne.Container) {
	var err error
	const fontScale = 0.7
	scp.channelViewers[chIndex].maxV, err = disp7.NewCustomDisp7Array(5, 3,
		20000, -20000, disp7.Signed, disp7.NoTrailingZeroes, scp.Window,
		scp.Settings.Channels[chIndex].Col[scp.Settings.ChannelColorIndex],
		disp7.ReaOnly, fontScale*disp7.DefaultDigitWidth,
		fontScale*disp7.DeafultDigitHeight, 1,
		fontScale*disp7.DefaultVCursorSpace, " Max:", " V ")
	if err != nil {
		panic("error from disp7.NewCustomDisp7Array")
	}
	scp.channelViewers[chIndex].minV, err = disp7.NewCustomDisp7Array(5, 3,
		20000, -20000, disp7.Signed, disp7.NoTrailingZeroes, scp.Window,
		scp.Settings.Channels[chIndex].Col[scp.Settings.ChannelColorIndex],
		disp7.ReaOnly, fontScale*disp7.DefaultDigitWidth,
		fontScale*disp7.DeafultDigitHeight, 1,
		fontScale*disp7.DefaultVCursorSpace, " Min:", " V ")
	if err != nil {
		panic("error from disp7.NewCustomDisp7Array")
	}
	vfBox = container.New(layout.NewVBoxLayout(),
		scp.channelViewers[chIndex].maxV,
		scp.channelViewers[chIndex].minV)
	return
}

func (scp *ScpDesc) newChannel(chIndex genericps.ChannelId) *fyne.Container {
	var (
		invertTriggerIpm *fyne.Container
		invert           *widget.Check
		x10              *widget.Check
		vRange           *selectscroll.SelectScroll
		trigger          *widget.Check
		channelViewer    *channelViewerDesc
		channel          *settings.ChSettings
		ranges           []string
	)
	channelViewer = &scp.channelViewers[chIndex]
	channel = &scp.Settings.Channels[chIndex]
	chId := channelNames[chIndex]
	setChannel := func() {
		/*defer*/ setFlag(scp.repartition)
		channel.ID = chIndex
		channelCopy := scp.Settings.Channels[chIndex]
		scp.psControl.SetChannelCh <- &channelCopy
		if channelViewer.enableCheckbox.Val &&
			channelViewer.triggerCheckbox.Checked {
			scp.psControl.SetTriggerCh <- &scp.triggerSettingMsg
			<-scp.triggerSettingMsg.Done
		}
		scp.ResetFfSweep()
		scp.clearAllFtPersistentLayers()
		scp.clearAllDftPersistentLayers()
		scp.SaveSettings()
	}
	cChanged := func(option string, e selectscroll.Exception) {
		switch option {
		case "AC":
			scp.Settings.Channels[chIndex].CoupleType = genericps.Ac
		case "DC":
			scp.Settings.Channels[chIndex].CoupleType = genericps.Dc
		default:
		}
		scp.ftRaster.Refresh()
		setChannel()
	}
	vChanged := func(option string, e selectscroll.Exception) {
		scp.changeChannelRange(chIndex, option)
	}
	enableChanged := func(c bool, col color.Color) {
		scp.SetChannelColors(col, chIndex)
		scp.EnableChannel(chIndex, c)
	}
	inverted := func(c bool) {
		scp.Settings.Channels[chIndex].Inverted = c
		// channel.inverted = c
		//		fmt.Println("inverted ", chIndex, chName, c)
		scp.ResetFfSweep()
		scp.clearAllFtPersistentLayers()
		scp.clearAllDftPersistentLayers()
		scp.SaveSettings()
		//		setChannel()
	}
	x10Changed := func(c bool) {
		scp.changeChannelX10(chIndex, c)
	}
	triggerTypeChanged := func(option string, e selectscroll.Exception) {
		var direction genericps.ThresholdDirection
		switch {
		case option == failing:
			direction = genericps.TriggerFalling
		case option == raising:
			direction = genericps.TriggerRaising
		default:
			return
		}
		channel.Trigger.TriggerDirection = direction
		scp.Settings.Channels[chIndex].Trigger.TriggerDirection = direction
		if scp.triggerSource == chIndex && channelViewer.triggerCheckbox.Checked {
			scp.setTrigger(true, chIndex, channel.Trigger.Mv, direction, 1000,
				scp.Settings.Time.TriggerTimeOffset)
		}
		scp.refreshRasters()
		setChannel()
	}
	triggerSelected := func(checked bool) {
		if !channel.Enabled { // TODO is it impossible?
			channel.TriggerSource = false
			scp.triggerCheck[chIndex].Checked = false
			return
		}
		scp.Settings.Channels[chIndex].TriggerSource = checked
		if checked {
			for i, v := range scp.triggerCheck {
				if i != int(chIndex) {
					v.Checked = false
					scp.Settings.Channels[i].TriggerSource = false
					v.Refresh()
				}
			}
			scp.triggerSource = chIndex
			channel.TriggerSource = true
			if scp.triggerDisplays != nil {
				if !scp.inStreamMode() {
					scp.triggerDisplays.Show()
				} else {
					scp.triggerDisplays.Hide()
				}
				col := channel.Col[scp.Settings.ChannelColorIndex]
				scp.triggerThresholdDisp.SetOncolor(col)
				scp.triggerHysteresisDisp.SetOncolor(col)
			}
		} else {
			scp.triggerSource = dontCare
			scp.triggerDisplays.Hide()
			channel.TriggerSource = false
		}
		scp.triggerThresholdDisp.Refresh()
		scp.triggerHysteresisDisp.Refresh()
		scp.SetTriggerUpperHysteresis(channel.Trigger.Hysteresis)
		scp.setTrigger(checked, chIndex, channel.Trigger.Mv,
			channel.Trigger.TriggerDirection,
			1000, scp.Settings.Time.TriggerTimeOffset)
		scp.refreshRasters()
		setChannel()
	}
	channelOffset := func(chIndex genericps.ChannelId) (
		channelOffsetBox *fyne.Container) {
		var err error
		const fontScale = 0.7
		scp.channelViewers[chIndex].offset, err =
			disp7.NewCustomDisp7Array(5, 3, 20000, -20000, disp7.Signed,
				disp7.NoTrailingZeroes, scp.Window,
				scp.Settings.Channels[chIndex].Col[scp.Settings.ChannelColorIndex],
				disp7.ReadWrite,
				fontScale*disp7.DefaultDigitWidth,
				fontScale*disp7.DeafultDigitHeight, 1, fontScale*disp7.DefaultVCursorSpace,
				"Offs:", " V")
		if err != nil {
			panic(err.Error() + " error from disp7.NewCustomDisp7Array")
		}
		scp.channelViewers[chIndex].offset.SetFloatValue(float64(channel.Offset),
			3)
		scp.channelViewers[chIndex].offset.OnChanged = func(v float64) {
			channel.Offset = float32(v) / 1000.0
			setChannel()
		}
		channelOffsetBox = container.New(layout.NewHBoxLayout(),
			scp.channelViewers[chIndex].offset)
		addToTest(scp.channelViewers[chIndex].offset, chOffsetId+chId)
		return
	}

	idLabel := widget.NewLabel(chId)
	channelViewer.filterWarning = canvas.NewText("⚠️", color.NRGBA{255, 0, 0, 255})
	channelViewer.filterWarning.TextStyle.Bold = true
	if scp.isDigitalFilterEnabled(chIndex) {
		channelViewer.filterWarning.Show()
	} else {
		channelViewer.filterWarning.Hide()
	}

	channelViewer.enableCheckbox = checkcolorpick.NewCheckColorPick(scp.Window,
		enableChanged, scp.Settings.Channels[chIndex].Col[scp.Settings.ChannelColorIndex],
		fyne.Size{Width: checkColorPickMinSize, Height: checkColorPickMinSize})
	addToTest(channelViewer.enableCheckbox, chEnableId+chId)
	enableCh := container.New(layout.NewHBoxLayout(),
		channelViewer.enableCheckbox, idLabel, container.NewCenter(channelViewer.filterWarning))
	invert = widget.NewCheck("Inv:", inverted)
	invert.SetChecked(scp.Settings.Channels[chIndex].Inverted)
	addToTest(invert, invertId+chId)
	trigger = widget.NewCheck("Trig:", triggerSelected)
	channelViewer.triggerCheckbox = trigger
	scp.triggerCheck = append(scp.triggerCheck, trigger)
	addToTest(trigger, triggerCheckId+chId)

	persSelected := func(checked bool) {
		channel.Persistence = checked
		scp.Settings.Channels[chIndex].Persistence = checked
		if !checked {
			scp.clearFtPersistentLayer(chIndex)
		}
		scp.refreshRasters()
		scp.SaveSettings()
	}
	pers := widget.NewCheck("Pers:", persSelected)
	pers.SetChecked(scp.Settings.Channels[chIndex].Persistence)
	channelViewer.persistenceCheckbox = pers
	addToTest(pers, persId+chId)

	rangesEnum, err := scp.psControl.ChannelRanges(chIndex)
	switch {
	case err != nil:
		slog.Error("ChannelRanges", "error", err)
	default:
		for i := range rangesEnum {
			ranges = append(ranges, inputRanges[rangesEnum[i]])
		}
	}
	offsetBox := channelOffset(chIndex)
	vRange = selectscroll.NewSelectScroll(ranges, vChanged, "+500m")
	vr := scp.Settings.Channels[chIndex].VRange
	if s, ok := rangeEnumToString[vr]; ok {
		vRange.SetSelected(s)
	}
	x10 = widget.NewCheck("X10", x10Changed)
	x10.SetChecked(scp.Settings.Channels[chIndex].X10)
	addToTest(x10, x10Id+chId)
	channelViewer.x10Checkboxes = append(channelViewer.x10Checkboxes, x10)
	channelViewer.vRangeSelects = append(channelViewer.vRangeSelects, vRange)
	addToTest(vRange, vRangeId+chId)
	acdc := selectscroll.NewSelectScroll([]string{"AC", "DC"}, cChanged, "AC")
	acdc.SetSelected(coupleTypeNames[scp.Settings.Channels[chIndex].CoupleType])
	addToTest(acdc, acdcId+chId)
	triggerDirection := selectscroll.NewSelectScroll(triggerDirectionOptions,
		triggerTypeChanged, "Raising")
	addToTest(triggerDirection, triggerDirectionId)
	triggerDirection.SetSelected(
		triggerDirectionNames[scp.Settings.Channels[chIndex].Trigger.TriggerDirection])
	invertTriggerIpm = container.New(layout.NewHBoxLayout(), invert, trigger,
		triggerDirection, pers)
	enableCouplingRange := container.New(layout.NewHBoxLayout(), enableCh, acdc,
		vRange, x10)
	minMaxBox := scp.minMaxDisp(chIndex)
	frqPeriodBox := scp.frqPeriodDisp(chIndex)
	voltageBox := container.New(layout.NewVBoxLayout(), offsetBox, minMaxBox)
	vfBox := container.New(layout.NewCustomPaddedHBoxLayout(-20), voltageBox, frqPeriodBox)
	setChannel()
	box := container.New(layout.NewVBoxLayout(), enableCouplingRange,
		invertTriggerIpm, vfBox)
	return box
}

func (scp *ScpDesc) SetChannelColors(col color.Color,
	chIndex genericps.ChannelId) {
	channelViewer := scp.channelViewers[chIndex]
	cfg := &scp.Settings.Channels[chIndex]
	cfg.Col[scp.Settings.ChannelColorIndex] = col.(color.NRGBA)
	if channelViewer.enableCheckbox.Val &&
		channelViewer.triggerCheckbox.Checked {
		scp.triggerThresholdDisp.SetOncolor(
			cfg.Col[scp.Settings.ChannelColorIndex])
		scp.triggerHysteresisDisp.SetOncolor(
			cfg.Col[scp.Settings.ChannelColorIndex])
	}
	channelViewer.minV.SetOncolor(
		cfg.Col[scp.Settings.ChannelColorIndex])
	channelViewer.maxV.SetOncolor(
		cfg.Col[scp.Settings.ChannelColorIndex])
	channelViewer.offset.SetOncolor(
		cfg.Col[scp.Settings.ChannelColorIndex])
	channelViewer.frq.SetOncolor(
		cfg.Col[scp.Settings.ChannelColorIndex])
	channelViewer.period.SetOncolor(
		cfg.Col[scp.Settings.ChannelColorIndex])
	if channelViewer.fvNameLabel != nil {
		channelViewer.fvNameLabel.Color =
			cfg.Col[scp.Settings.ChannelColorIndex]
		channelViewer.fvNameLabel.Refresh()
	}
	if channelViewer.dftNameLabel != nil {
		channelViewer.dftNameLabel.Color =
			cfg.Col[scp.Settings.ChannelColorIndex]
		channelViewer.dftNameLabel.Refresh()
	}
	if channelViewer.ffNameLabel != nil {
		channelViewer.ffNameLabel.Color =
			cfg.Col[scp.Settings.ChannelColorIndex]
		channelViewer.ffNameLabel.Refresh()
	}
	if channelViewer.rlcNameLabel != nil {
		channelViewer.rlcNameLabel.Color =
			cfg.Col[scp.Settings.ChannelColorIndex]
		channelViewer.rlcNameLabel.Refresh()
	}
	if channelViewer.simGenNameLabel != nil {
		channelViewer.simGenNameLabel.Color =
			cfg.Col[scp.Settings.ChannelColorIndex]
		channelViewer.simGenNameLabel.Refresh()
	}
	for _, d := range channelViewer.simGenDisplays {
		if d != nil {
			d.SetOncolor(cfg.Col[scp.Settings.ChannelColorIndex])
		}
	}
	scp.SaveSettings()
}

func (scp *ScpDesc) changeChannelRange(chIndex genericps.ChannelId, option string) {
	channelViewer := &scp.channelViewers[chIndex]
	channel := &scp.Settings.Channels[chIndex]

	channelViewer.label.enableRefresh()
	channelViewer.dftLabel.enableRefresh()
	scp.Settings.Channels[chIndex].VRange = vRanges[option]
	if scp.Settings.Channels[chIndex].VRange < 0 {
		panic("<0")
	}
	// Synchronize all range selectors for this channel
	for _, vSel := range channelViewer.vRangeSelects {
		if vSel.Selected != option {
			vSel.SilentSetSelected(option)
		}
	}

	if chIndex == scp.triggerSource &&
		scp.triggerCheck[chIndex].Checked {
		slog.Debug("vRange -> trigger")
		scp.setTrigger(true, chIndex, channel.Trigger.Mv,
			channel.Trigger.TriggerDirection, 1000, scp.Settings.Time.TriggerTimeOffset)
	}
	max, min, err := scp.psControl.Con.GetAnalogueOffset(int(
		scp.Settings.Channels[chIndex].VRange),
		scp.Settings.Channels[chIndex].CoupleType)
	scp.channelViewers[chIndex].offset.SetMinMax(int(min*1000),
		int(max*1000))
	slog.Debug("AnalogueOffset", "max", max, "min", min, "err", err)
	scp.ResetFfSweep()
	scp.clearAllFtPersistentLayers()
	scp.clearAllDftPersistentLayers()

	// Update the device
	channelCopy := scp.Settings.Channels[chIndex]
	scp.psControl.SetChannelCh <- &channelCopy
	if channelViewer.enableCheckbox.Val &&
		channelViewer.triggerCheckbox.Checked {
		scp.psControl.SetTriggerCh <- &scp.triggerSettingMsg
		<-scp.triggerSettingMsg.Done
	}
	scp.SaveSettings()
}

func (scp *ScpDesc) changeChannelX10(chIndex genericps.ChannelId, c bool) {
	channelViewer := &scp.channelViewers[chIndex]
	channel := &scp.Settings.Channels[chIndex]

	scp.Settings.Channels[chIndex].X10 = c
	scp.clearAllFtPersistentLayers()
	scp.clearAllDftPersistentLayers()

	rangesEnum, _ := scp.psControl.ChannelRanges(chIndex)
	var ranges []string
	for _, r := range rangesEnum {
		ranges = append(ranges, inputRanges[r])
	}

	indexChanged := false
	var newOption string

	// Update all vRange selects
	for _, vRange := range channelViewer.vRangeSelects {
		p := vRange.SelectedIndex()
		if c {
			vRange.SetOptions(ranges[:len(ranges)-3])
			if p >= len(ranges)-3 {
				p = p - 3
				indexChanged = true
			}
			vRange.SetSelectedIndex(p)
			canvas.Refresh(vRange)
		} else {
			vRange.SetOptions(ranges)
			vRange.SetSelectedIndex(p)
			canvas.Refresh(vRange)
		}
		if p >= 0 && p < len(vRange.Options) {
			newOption = vRange.Options[p]
		}
	}

	// Synchronize all X10 checkboxes
	for _, x10Check := range channelViewer.x10Checkboxes {
		if x10Check.Checked != c {
			x10Check.Checked = c
			x10Check.Refresh()
		}
	}

	if indexChanged && newOption != "" {
		scp.changeChannelRange(chIndex, newOption)
	} else {
		// Just send channel update since X10 state changed
		channel.ID = chIndex
		channelCopy := scp.Settings.Channels[chIndex]
		scp.psControl.SetChannelCh <- &channelCopy
		if channelViewer.enableCheckbox.Val &&
			channelViewer.triggerCheckbox.Checked {
			scp.psControl.SetTriggerCh <- &scp.triggerSettingMsg
			<-scp.triggerSettingMsg.Done
		}
		scp.ResetFfSweep()
		scp.SaveSettings()
	}
}

func (scp *ScpDesc) EnableChannel(chIndex genericps.ChannelId, c bool) {
	/*defer*/ setFlag(scp.repartition)
	channel := &scp.Settings.Channels[chIndex]
	channelViewer := &scp.channelViewers[chIndex]

	if !c {
		if scp.triggerCheck[chIndex].Checked {
			scp.triggerCheck[chIndex].Checked = false
			scp.triggerCheck[chIndex].Refresh()
			channel.TriggerSource = false
			scp.triggerSource = dontCare
			scp.triggerDisplays.Hide()
			scp.setTrigger(c, chIndex, 0, channel.Trigger.TriggerDirection,
				1000, scp.Settings.Time.TriggerTimeOffset)
		}
	}
	col := channel.Col[scp.Settings.ChannelColorIndex]
	scp.SetChannelColors(col, chIndex)
	channelViewer.label.channelIndex = int(chIndex)
	channel.Enabled = c

	// Synchronize checkboxes
	if channelViewer.enableCheckbox != nil && channelViewer.enableCheckbox.Val != c {
		channelViewer.enableCheckbox.Val = c
		canvas.Refresh(channelViewer.enableCheckbox)
	}
	if channelViewer.dftCheckbox != nil && channelViewer.dftCheckbox.Checked != c {
		channelViewer.dftCheckbox.Checked = c
		channelViewer.dftCheckbox.Refresh()
	}
	for _, chk := range channelViewer.enableChecks {
		if chk != nil && chk.Checked != c {
			chk.Checked = c
			chk.Refresh()
		}
	}

	scp.ResetFfSweep()

	// Update device
	channel.ID = chIndex
	channelCopy := *channel
	scp.psControl.SetChannelCh <- &channelCopy
	if channel.Enabled && channel.TriggerSource {
		scp.psControl.SetTriggerCh <- &scp.triggerSettingMsg
		<-scp.triggerSettingMsg.Done
	}
}

func (scp *ScpDesc) newChannelPanels(container *fyne.Container) {
	initMaps()
	sortInputRanges()
	slog.Debug("newChannelPanels")
	scp.psControl.NewChannels(int(scp.channelCount))
	for i := range scp.channelViewers {
		if i > 0 {
			container.Add(layout.NewSpacer())
		}
		container.Add(scp.newChannel(genericps.ChannelId(i)))
	}
}
