package gui

import (
	"fynescope/genericps"
	"fynescope/settings"
	"image"
	"image/draw"
	"log/slog"
	"math"

	"fynescope/control"
	"fynescope/disp7"
	"fynescope/selectscroll"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
)

const (
	nsUsMssubSetIndex = 3
	// psSubSetIndex     = 7
	picoSec       = "ps"
	nanoSec       = "ns"
	microSec      = "µs"
	milliSec      = "ms"
	sec           = "s"
	min           = "min"
	div           = "/div"
	dot           = "Dot"
	raw           = "Raw"
	linear        = "Linear"
	sinc          = "Sinc"
	autoTriggerMs = 1000 // TODO make it adjustable
)

type (
	timeLabelViewer struct {
		rasterPartition
		scp      *ScpDesc
		selected bool
	}
)

var (
	times = []string{"5000", "2000", "1000", "500", "200", "100",
		"50", "20", "10", "5", "2", "1"}
	nsUsMsTimesSubset = times[nsUsMssubSetIndex:]
	units             []string
	etsUnits          []string
	tu                map[string]int

	triggerModeOptions = []string{"Auto", "ETS", "Repeat", "Single"}
	triggerModes       = map[string]control.TriggerModes{
		triggerModeOptions[0]: control.Auto,
		triggerModeOptions[1]: control.ETS,
		triggerModeOptions[2]: control.Repeat,
		triggerModeOptions[3]: control.Single,
	}
	triggerTypeOptions = []string{"Simple", "Advanced", "Window"}
	triggerTypes       = map[string]control.TriggerTypes{
		triggerTypeOptions[0]: control.Simple,
		triggerTypeOptions[1]: control.Advanced,
		triggerTypeOptions[2]: control.Window,
	}
	sampleRates = []string{"900", "800", "700", "600", "500", "400", "300", "200", "100",
		"90", "80", "70", "60", "50", "40", "30", "20", "10",
		"9", "8", "7", "6", "5", "4", "3", "2", "1"}
	sampleUnits              = []string{"GS/s", "MS/s", "KS/s", "S/s"}
	interpolationModeOptions = []string{dot, raw, linear, sinc}
	interpolationModes       = map[string]settings.InterpolationType{
		interpolationModeOptions[settings.Sinc]:   settings.Sinc,
		interpolationModeOptions[settings.Linear]: settings.Linear,
		interpolationModeOptions[settings.Raw]:    settings.Raw,
		interpolationModeOptions[settings.Dot]:    settings.Dot,
	}
)

var (
	_ mouser     = (*timeLabelViewer)(nil)
	_ dragger    = (*timeLabelViewer)(nil)
	_ scroller   = (*timeLabelViewer)(nil)
	_ keyer      = (*timeLabelViewer)(nil)
	_ cursorable = (*timeLabelViewer)(nil)
	_ drawer     = (*timeLabelViewer)(nil)
)

func (tl *timeLabelViewer) cursor(x, y float32) (desktop.Cursor, bool) {
	if tl.mousIn(x, y) {
		return desktop.PointerCursor, true
	}
	return desktop.DefaultCursor, false
}

func (tl *timeLabelViewer) mouseMoved(x, y float32) {
}
func (tl *timeLabelViewer) mousIn(x, y float32) bool {
	p := image.Point{X: int(math.Round(float64(x))), Y: int(math.Round(float64(y)))}
	if p.In(tl.rect()) {
		return true
	}
	return false
}
func (tl *timeLabelViewer) mouseDown(button desktop.MouseButton, x, y float32) {
	tl.selected = tl.mousIn(x, y)
}
func (tl *timeLabelViewer) mouseUp(button desktop.MouseButton, x, y float32) {
	tl.selected = false
}

func (tl *timeLabelViewer) setDtDispXOffset(dx, x, y float32) {
	p := image.Point{X: int(x), Y: int(y)}
	if p.In(tl.rect()) {
		tl.scp.addFtXOffset(float64(dx))
		// tl.scp.setTriggerTimeRatio(tl.scp.Settings.Time.XOffsetRatio)
		tl.scp.setTriggerTime(tl.scp.Settings.Time.TriggerTimeOffset)
		tl.enableRefresh()
		tl.scp.clearAllFtPersistentLayers()
		tl.scp.clearAllDftPersistentLayers()
		tl.scp.refreshRasters()
	}
}
func (tl *timeLabelViewer) dragged(dx, dy, x, y float32) {
	if tl.selected {
		tl.setDtDispXOffset(dx, x, y)
	}
}

func (tl *timeLabelViewer) scrolled(delta, x, y float32) {
	nX := (float32(tl.scp.ftScopeSignalScreen.Bounds().Dx()) / float32(numberOfDivs)) / 10
	tl.setDtDispXOffset(delta*nX, x, y)
}

func (tl *timeLabelViewer) typedKey(x, y float32, keyName fyne.KeyName) {
	switch keyName {
	case fyne.KeyLeft:
		tl.scrolled(-scrollDelta, x, y)
	case fyne.KeyRight:
		tl.scrolled(scrollDelta, x, y)
	}
}
func newTimelLabelViewer(img rasterImage, imgRect image.Rectangle, scp *ScpDesc) *timeLabelViewer {
	tl := &timeLabelViewer{rasterPartition: rasterPartition{img: img, imgRect: imgRect, refreshFlag: true},
		scp: scp}
	return tl
}

func switchUpTimeUnit(dt float32, timeUnit int) (newDt float32, unitName string) {
	switch timeUnit {
	case -12:
		newDt = dt / 1000
		unitName = nanoSec
	case -9:
		newDt = dt / 1000
		unitName = microSec
	case -6:
		newDt = dt / 1000
		unitName = milliSec
	case -3:
		newDt = dt / 1000
		unitName = sec
	case 0:
		newDt = dt / 60
		unitName = min
	default:
		slog.Error("switchUpTimeUnit", "timeUnit", timeUnit)
		newDt = dt
		unitName = "?"
	}
	return
}

func getTimeUnitName(timeUnit int) (unitName string) {
	switch timeUnit {
	case -12:
		unitName = picoSec
	case -9:
		unitName = nanoSec
	case -6:
		unitName = microSec
	case -3:
		unitName = milliSec
	case 0:
		unitName = sec
	default:
		slog.Error("getTimeUnitName", "timeUnit", timeUnit)
		unitName = "S"
	}
	return
}

func (tl *timeLabelViewer) draw() {
	if !tl.refreshFlag {
		return
	}
	if tl.scp.controlTab.SelectedIndex() == dftTabIndex {
		return
	}
	tl.clear()
	var unitName string
	dt := float32(tl.scp.timeDiv)
	if tl.scp.timeDiv < 100 {
		unitName = getTimeUnitName(tl.scp.timeUnit)
	} else {
		dt, unitName = switchUpTimeUnit(dt, tl.scp.timeUnit)
	}
	bounds := tl.rect()
	y := bounds.Max.Y - fontSize
	w := float64(tl.scp.ftScopeSignalScreen.Bounds().Dx() - 1)
	v := float32(dt)
	zeroAt := w*tl.scp.Settings.Time.TriggerTimeOffset/tl.scp.maxScreenTime + float64(tl.scp.ftScopeSignalScreen.Bounds().Min.X)
	diff := float32(10000000)
	// bestI := 0
	for i := 0; i < len(tl.scp.ftDivsX); i++ {
		newDiff := float32(zeroAt) - tl.scp.ftDivsX[i]
		if newDiff < 0 {
			newDiff = -newDiff
		}
		if newDiff < diff {
			// bestI = i
			diff = newDiff
		} else if newDiff > diff {
			break
		}
		v -= float32(dt)
	}
	for i, x := range tl.scp.ftDivsX {
		if v > -dt/8 && v < dt/8 { // avoid -0.0
			v = 0
		}
		vstr := strconv.FormatFloat(float64(v), 'f', 1, 32)
		if i == 0 { // 											first label
			vstr = vstr + " " + unitName
		}
		left, _, right, _ := tl.scp.boundString(vstr)
		tl.scp.addLabel(tl.scp.ftScopeFullScreen, int(math.Round(float64(x-(right+left)/2))), y, vstr, theme.ForegroundColor())
		v += float32(dt)
	}
	tl.disableRefresh()
}

func (tl *timeLabelViewer) clear() {
	draw.Draw(tl.img, tl.rect(), &image.Uniform{theme.BackgroundColor()}, image.ZP, draw.Src)
}

func (scp *ScpDesc) newUnitList() {
	if units == nil {
		units = []string{sec + div, milliSec + div, microSec + div, nanoSec + div, picoSec + div}
		etsUnits = units[3:]
		tu = map[string]int{
			sec + div:      0,
			milliSec + div: -3,
			microSec + div: -6,
			nanoSec + div:  -9,
			picoSec + div:  -12,
		}
	}
}

func (scp *ScpDesc) setMaxScreenTime() {
	scp.updateAcquisitionParameters()
}

func (scp *ScpDesc) setTrigger(enable bool, source genericps.ChannelId, mv int32, direction genericps.ThresholdDirection,
	autoTriggerMs int16, xOffset float64) {
	vRange := scp.Settings.Channels[genericps.ChA].VRange
	if scp.triggerSource != dontCare && int(scp.triggerSource) < len(scp.Settings.Channels) {
		vRange = scp.Settings.Channels[scp.triggerSource].VRange
	}

	var upperHysteresis int32
	var lowerMv int32
	var lowerHysteresis int32
	var thresholdMode genericps.ThresholdModeId

	if source != dontCare && int(source) < len(scp.Settings.Channels) {
		trig := scp.Settings.Channels[source].Trigger
		upperHysteresis = trig.Hysteresis
		lowerMv = trig.LowerMv
		lowerHysteresis = trig.LowerHysteresis
		thresholdMode = trig.ThresholdMode
	} else {
		upperHysteresis = scp.triggerSettingMsg.UpperHysteresis
		lowerMv = scp.triggerSettingMsg.LowerMv
		lowerHysteresis = scp.triggerSettingMsg.LowerHysteresis
		thresholdMode = scp.triggerSettingMsg.ThresholdMode
	}

	hysteresisADC := uint16(scp.mvToUAdc(upperHysteresis, vRange))
	triggerADC := int16(scp.mvToAdc(mv, vRange))
	lowerTriggerADC := int16(scp.mvToAdc(lowerMv, vRange))
	lowerHysteresisADC := uint16(scp.mvToUAdc(lowerHysteresis, vRange))

	if scp.triggerSettingMsg.Enabled != enable ||
		scp.triggerSettingMsg.Source != source ||
		scp.triggerSettingMsg.Mv != mv ||
		scp.triggerSettingMsg.AutoTriggerMs != autoTriggerMs ||
		scp.triggerSettingMsg.XOffset != xOffset ||
		scp.triggerSettingMsg.ThresholdDirection != direction ||
		scp.triggerSettingMsg.HysteresisADC != hysteresisADC ||
		scp.triggerSettingMsg.TriggerADC != triggerADC ||
		scp.triggerSettingMsg.UpperHysteresis != upperHysteresis ||
		scp.triggerSettingMsg.LowerMv != lowerMv ||
		scp.triggerSettingMsg.LowerHysteresis != lowerHysteresis ||
		scp.triggerSettingMsg.LowerHysteresisADC != lowerHysteresisADC ||
		scp.triggerSettingMsg.LowerTriggerADC != lowerTriggerADC ||
		scp.triggerSettingMsg.ThresholdMode != thresholdMode {

		// slog.Debug("new trigger", "old triggerADC", scp.triggerSettingMsg.TriggerADC)
		scp.triggerSettingMsg.Enabled = enable
		scp.triggerSettingMsg.Source = source
		scp.triggerSettingMsg.Mv = mv
		scp.triggerSettingMsg.UpperHysteresis = upperHysteresis
		scp.triggerSettingMsg.HysteresisADC = hysteresisADC
		scp.triggerSettingMsg.TriggerADC = triggerADC
		scp.triggerSettingMsg.AutoTriggerMs = autoTriggerMs
		scp.triggerSettingMsg.XOffset = xOffset
		scp.triggerSettingMsg.ThresholdDirection = direction
		scp.triggerSettingMsg.LowerMv = lowerMv
		scp.triggerSettingMsg.LowerHysteresis = lowerHysteresis
		scp.triggerSettingMsg.LowerHysteresisADC = lowerHysteresisADC
		scp.triggerSettingMsg.LowerTriggerADC = lowerTriggerADC
		scp.triggerSettingMsg.ThresholdMode = thresholdMode
		scp.psControl.SetTriggerCh <- &scp.triggerSettingMsg
		<-scp.triggerSettingMsg.Done
	} else {
		slog.Debug("not new trigger")
	}
}

func (scp *ScpDesc) setTimeSelectOptions(unitOption string) {
	if unitOption == "S/div" {
		if scp.timeSelect.Options[0] != times[0] {
			scp.timeSelect.Options = times
			index := scp.timeSelect.SelectedIndex()
			scp.timeSelect.SetSelectedIndex(index)
			scp.Settings.Time.TimeDiv = scp.timeSelect.Selected
		}
	} else {
		if scp.timeSelect.Options[0] == times[0] {
			scp.timeSelect.Options = nsUsMsTimesSubset
			index := scp.timeSelect.SelectedIndex()
			if index <= nsUsMssubSetIndex {
				scp.timeSelect.SilentSetSelectedIndex(0)
				scp.Settings.Time.TimeDiv = times[0]
			}
		}
	}
}

func (scp *ScpDesc) timeUnitUp() {
	if scp.timeUnitSelect.Selected == units[0] {
		scp.timeSelect.Options = times
	} else {
		if scp.triggerSettingMsg.Mode != control.ETS ||
			scp.timeUnitSelect.Selected == picoSec+div {
			scp.timeSelect.SilentSetSelectedIndex(len(scp.timeSelect.Options) - 1)
			scp.Settings.Time.TimeDiv = times[scp.timeSelect.SelectedIndex()]
			index := scp.timeUnitSelect.SelectedIndex()
			scp.timeUnitSelect.SilentSetSelectedIndex(index - 1)
			scp.Settings.Time.Unit = scp.timeUnitSelect.Selected
		}
	}
}

func (scp *ScpDesc) timeUnitDown() {
	if scp.timeUnitSelect.Selected == units[0] {
		scp.timeSelect.Options = nsUsMsTimesSubset
	}
	index := scp.timeUnitSelect.SelectedIndex()
	if index < len(scp.timeUnitSelect.Options)-1 {
		scp.timeSelect.SilentSetSelectedIndex(0)
		scp.Settings.Time.TimeDiv = times[scp.timeSelect.SelectedIndex()]
		scp.timeUnitSelect.SilentSetSelectedIndex(index + 1)
		scp.Settings.Time.Unit = scp.timeUnitSelect.Selected
	}
}

func (scp *ScpDesc) onTimeUnitChange(option string, ex selectscroll.Exception) {
	prevTimeUnit := scp.timeUnit
	scp.setTimeSelectOptions(option)
	scp.timeUnit = tu[scp.timeUnitSelect.Selected]
	scp.Settings.Time.Unit = scp.timeUnitSelect.Selected
	intTimeDiv, _ := strconv.Atoi(scp.timeSelect.Selected)
	scp.timeDiv = intTimeDiv
	scp.Settings.Time.TimeDiv = scp.timeSelect.Selected
	scp.setMaxScreenTime()
	scp.syncDftToTimeDiv()
	scp.clearAllFtPersistentLayers()
	scp.clearAllDftPersistentLayers()
	scp.timeSelect.Refresh()
	scp.timeUnitSelect.Refresh()
	scp.refreshRasters()
	mul := math.Pow(10, float64(scp.timeUnit)) / math.Pow(10, float64(prevTimeUnit))
	scp.Settings.Time.TriggerTimeOffset *= mul
	scp.setTriggerTime(scp.Settings.Time.TriggerTimeOffset)
	scp.SaveSettings()
}

func (scp *ScpDesc) onTimeDivChange(option string, ex selectscroll.Exception) {
	tl := scp.ftBottomLabelViewer.(*timeLabelViewer)
	switch {
	case ex == selectscroll.Over:
		scp.timeUnitDown()
	case ex == selectscroll.Under:
		scp.timeUnitUp()
	default:
	}
	prevTimeUnit := scp.timeUnit
	scp.timeUnit = tu[scp.timeUnitSelect.Selected]
	mul := math.Pow(10, float64(scp.timeUnit)) / math.Pow(10, float64(prevTimeUnit))
	prevTime := scp.timeDiv
	intTimeDiv, _ := strconv.Atoi(scp.timeSelect.Selected)
	scp.timeDiv = intTimeDiv
	scp.Settings.Time.TimeDiv = scp.timeSelect.Selected
	tl.enableRefresh()
	scp.setMaxScreenTime()
	scp.syncDftToTimeDiv()
	scp.clearAllFtPersistentLayers()
	scp.clearAllDftPersistentLayers()
	scp.timeSelect.Refresh()
	scp.refreshRasters()
	mul *= float64(scp.timeDiv) / float64(prevTime)
	scp.Settings.Time.TriggerTimeOffset *= mul
	scp.setTriggerTime(scp.Settings.Time.TriggerTimeOffset)
	scp.SaveSettings()
}

func (scp *ScpDesc) onInterpolationModeChange(option string, e selectscroll.Exception) {
	scp.psControl.SetInterpolationModeCh <- interpolationModes[option]
	scp.Settings.Time.Interpolation = interpolationModes[option]
	scp.refreshRasters()
	scp.SaveSettings()
}

func (scp *ScpDesc) sampleUnitUp() {
	scp.sampleRateSelect.SilentSetSelectedIndex(len(scp.sampleRateSelect.Options) - 1)
	index := scp.sampleUnitSelect.SelectedIndex()
	scp.sampleUnitSelect.SilentSetSelectedIndex(index - 1)
}

func (scp *ScpDesc) sampleUnitDown() {
	index := scp.sampleUnitSelect.SelectedIndex()
	if index < len(scp.sampleUnitSelect.Options)-1 {
		scp.sampleRateSelect.SilentSetSelectedIndex(0)
		scp.sampleUnitSelect.SilentSetSelectedIndex(index + 1)
	}
}

func (scp *ScpDesc) onSampleRateChange(_ string, ex selectscroll.Exception) {
	switch {
	case ex == selectscroll.Over:
		scp.sampleUnitDown()
	case ex == selectscroll.Under:
		if scp.sampleUnitSelect.SelectedIndex() > 0 {
			scp.sampleUnitUp()
		}
	default:
	}
}

func (scp *ScpDesc) onSampleUnitChange(_ string, _ selectscroll.Exception) {
}

func (scp *ScpDesc) setETSTimeDiv() {
	scp.ipmSelect.SetOptions(interpolationModeOptions[:3])
	if scp.ipmSelect.Selected == sinc {
		scp.ipmSelect.SetSelected(linear)
	}
	scp.timeUnitSelect.SetOptions(etsUnits)
	if scp.timeUnitSelect.Selected != nanoSec+div &&
		scp.timeUnitSelect.Selected != picoSec+div {
		scp.timeUnitSelect.SetSelected(nanoSec + div)
	}
	scp.timeUnitSelect.Refresh()
}

func (scp *ScpDesc) setNotETSTimeDiv() {
	scp.ipmSelect.SetOptions(interpolationModeOptions)
	scp.timeUnitSelect.SetOptions(units)
}

func (scp *ScpDesc) onTriggerModeChange(option string, ex selectscroll.Exception) {
	prev := scp.triggerSettingMsg.Mode
	scp.Settings.Trigger.Mode = option
	if triggerModes[option] == control.ETS {
		if prev != control.ETS {
			scp.setETSTimeDiv()
			for i := range scp.channelViewers { // Uncheck and disable all channels
				scp.channelViewers[i].triggerCheckbox.SetChecked(false)
				scp.channelViewers[i].triggerCheckbox.Disable()
			}
			scp.triggerSource = genericps.ChA // only channel A is allowed
			//TODO Is it ps2000 specific?
			channelViewer := &scp.channelViewers[genericps.ChA]
			channel := &scp.Settings.Channels[genericps.ChA]
			channelViewer.enableCheckbox.Set()
			channelViewer.triggerCheckbox.Enable()
			channelViewer.triggerCheckbox.SetChecked(true)
			scp.setTrigger(true, genericps.ChA, channel.Trigger.Mv,
				channel.Trigger.TriggerDirection, autoTriggerMs, float64(scp.Settings.Time.TriggerTimeOffset))
			if scp.running {
				err := scp.psControl.Stop()
				if err != nil {
					slog.Error("onTriggerModeChange", "stop error:", err)
					return
				}
				err = scp.psControl.SetETSMode()
				if err != nil {
					slog.Error("onTriggerModeChange", "SetETSMode error:", err)
					return
				}
			}
		}
	} else {
		if prev == control.ETS {
			scp.setNotETSTimeDiv()
			for i := range scp.channelViewers {
				scp.channelViewers[i].triggerCheckbox.Enable()
			}
			err := scp.psControl.Stop()
			if err != nil {
				slog.Error("onTriggerModeChange", "stop error:", err)
				return
			}
			err = scp.psControl.SetBlockMode()
			if err != nil {
				slog.Error("onTriggerModeChange", "SetBlockMode error:", err)
				return
			}
		}
		if triggerModes[option] == control.Single {
			if scp.running {
				err := scp.psControl.Stop()
				if err != nil {
					slog.Error("onTriggerModeChange", "stop error:", err)
					return
				}
				scp.runblockButton.SetIcon(theme.MediaPlayIcon())
				scp.running = false
			}
		}
	}
	scp.triggerSettingMsg.Mode = triggerModes[option]
	scp.psControl.SetTriggerCh <- &scp.triggerSettingMsg // send to control
	<-scp.triggerSettingMsg.Done                         // wait for done
	setFlag(scp.repartition)
	scp.clearAllFtPersistentLayers()
	scp.clearAllDftPersistentLayers()
	scp.refreshRasters()
	scp.SaveSettings()
}

func (scp *ScpDesc) onTriggerTypeChange(option string, ex selectscroll.Exception) {
	scp.Settings.Trigger.Type = option
	scp.triggerSettingMsg.Type = triggerTypes[option]
	
	if scp.triggerSettingMsg.Type == control.Window {
		scp.triggerSettingMsg.ThresholdMode = genericps.Window
	} else {
		scp.triggerSettingMsg.ThresholdMode = genericps.Level
	}

	for i := range scp.channelViewers {
		scp.Settings.Channels[i].Trigger.ThresholdMode = scp.triggerSettingMsg.ThresholdMode
		if scp.channelViewers[i].triggerDirectionSelect != nil {
			var activeOpts []string
			if option == "Window" {
				activeOpts = triggerWindowDirectionOptions
			} else {
				activeOpts = triggerDirectionOptions
			}
			scp.channelViewers[i].triggerDirectionSelect.SetOptions(activeOpts)
			// Also reset the selection if it's invalid for the new options
			currSel := scp.channelViewers[i].triggerDirectionSelect.Selected
			valid := false
			for _, opt := range activeOpts {
				if opt == currSel {
					valid = true
					break
				}
			}
			if !valid && len(activeOpts) > 0 {
				scp.channelViewers[i].triggerDirectionSelect.SilentSetSelected(activeOpts[0])
				scp.Settings.Channels[i].Trigger.TriggerDirection = triggerDirections[activeOpts[0]]
				if scp.triggerSource == genericps.ChannelId(i) {
					scp.triggerSettingMsg.ThresholdDirection = triggerDirections[activeOpts[0]]
				}
			}
		}
	}

	if option == "Complex" {
		scp.buildComplexTriggerMessage()
		scp.showComplexTriggerPopup()
	}
	
	scp.psControl.SetTriggerCh <- &scp.triggerSettingMsg
	<-scp.triggerSettingMsg.Done
	if scp.boxTriggerHysteresisDisp != nil {
		if scp.triggerSettingMsg.Type == control.Simple {
			scp.boxTriggerHysteresisDisp.Hide()
			if scp.boxTriggerLowerDisp != nil {
				scp.boxTriggerLowerDisp.Hide()
			}
		} else if scp.triggerSettingMsg.Type == control.Advanced {
			scp.boxTriggerHysteresisDisp.Show()
			if scp.boxTriggerLowerDisp != nil {
				scp.boxTriggerLowerDisp.Hide()
			}
		} else if scp.triggerSettingMsg.Type == control.Window {
			scp.boxTriggerHysteresisDisp.Show()
			if scp.boxTriggerLowerDisp != nil {
				scp.boxTriggerLowerDisp.Show()
			}
		} else { // Complex
			scp.boxTriggerHysteresisDisp.Show()
			if scp.boxTriggerLowerDisp != nil {
				scp.boxTriggerLowerDisp.Hide()
			}
		}
	}
	setFlag(scp.repartition)
	scp.refreshRasters()
	scp.SaveSettings()
}

func (scp *ScpDesc) onThresholdChange(v float64) {
	if scp.triggerSource == dontCare {
		return
	}
	intV := int32(math.Round(v))
	if scp.Settings.Trigger.Type == "Window" {
		lowerMv := scp.Settings.Channels[scp.triggerSource].Trigger.LowerMv
		if intV < lowerMv {
			intV = lowerMv
			scp.triggerThresholdDisp.SilentSetValue(int(intV))
		}
	}
	scp.Settings.Channels[scp.triggerSource].Trigger.Mv = intV
	scp.triggerSettingMsg.Mv = intV
	scp.triggerSettingMsg.TriggerADC = int16(scp.mvToAdc(intV,
		scp.Settings.Channels[scp.triggerSource].VRange))
	scp.psControl.SetTriggerCh <- &scp.triggerSettingMsg
	<-scp.triggerSettingMsg.Done
	setFlag(scp.repartition)
	scp.clearAllFtPersistentLayers()
	scp.clearAllDftPersistentLayers()
	scp.refreshRasters()
	scp.SaveSettings()
}

func (scp *ScpDesc) onHysteresisChange(v float64) {
	if scp.triggerSource == dontCare {
		return
	}
	intV := int32(math.Round(v))
	scp.Settings.Channels[scp.triggerSource].Trigger.Hysteresis = intV
	scp.SetTriggerUpperHysteresis(intV)
	setFlag(scp.repartition)
	scp.clearAllFtPersistentLayers()
	scp.clearAllDftPersistentLayers()
	scp.refreshRasters()
	scp.SaveSettings()
}

func (scp *ScpDesc) onLowerThresholdChange(v float64) {
	if scp.triggerSource == dontCare {
		return
	}
	intV := int32(math.Round(v))
	if scp.Settings.Trigger.Type == "Window" {
		upperMv := scp.Settings.Channels[scp.triggerSource].Trigger.Mv
		if intV > upperMv {
			intV = upperMv
			scp.triggerLowerThresholdDisp.SilentSetValue(int(intV))
		}
	}
	scp.Settings.Channels[scp.triggerSource].Trigger.LowerMv = intV
	scp.triggerSettingMsg.LowerMv = intV
	scp.triggerSettingMsg.LowerTriggerADC = int16(scp.mvToAdc(intV,
		scp.Settings.Channels[scp.triggerSource].VRange))
	scp.psControl.SetTriggerCh <- &scp.triggerSettingMsg
	<-scp.triggerSettingMsg.Done
	setFlag(scp.repartition)
	scp.clearAllFtPersistentLayers()
	scp.clearAllDftPersistentLayers()
	scp.refreshRasters()
	scp.SaveSettings()
}

func (scp *ScpDesc) onLowerHysteresisChange(v float64) {
	if scp.triggerSource == dontCare {
		return
	}
	intV := int32(math.Round(v))
	scp.Settings.Channels[scp.triggerSource].Trigger.LowerHysteresis = intV
	scp.triggerSettingMsg.LowerHysteresis = intV
	scp.triggerSettingMsg.LowerHysteresisADC = uint16(scp.mvToUAdc(intV, scp.Settings.Channels[scp.triggerSource].VRange))
	scp.psControl.SetTriggerCh <- &scp.triggerSettingMsg
	<-scp.triggerSettingMsg.Done
	setFlag(scp.repartition)
	scp.clearAllFtPersistentLayers()
	scp.clearAllDftPersistentLayers()
	scp.refreshRasters()
	scp.SaveSettings()
}



func (scp *ScpDesc) newTimeSelectionUI() *fyne.Container {
	scp.timeUnitSelect = selectscroll.NewSelectScroll(units, scp.onTimeUnitChange, milliSec+div)
	scp.timeUnitSelect.SilentSetSelected(scp.Settings.Time.Unit)
	scp.timeUnit = tu[scp.timeUnitSelect.Selected]
	addToTest(scp.timeUnitSelect, unitSelectId)
	tOption := nsUsMsTimesSubset
	if scp.timeUnitSelect.Selected == sec+div {
		tOption = times
	}
	scp.timeSelect = selectscroll.NewSelectScroll(tOption, scp.onTimeDivChange, strconv.Itoa(500))
	addToTest(scp.timeSelect, timeSelectId)
	scp.timeSelect.SilentSetSelected(scp.Settings.Time.TimeDiv)
	intTimeDiv, _ := strconv.Atoi(scp.timeSelect.Selected)
	scp.timeDiv = intTimeDiv
	scp.ipmSelect = selectscroll.NewSelectScroll(interpolationModeOptions, scp.onInterpolationModeChange, linear)
	scp.ipmSelect.SetSelected(interpolationModeOptions[scp.Settings.Time.Interpolation])
	addToTest(scp.ipmSelect, ipmId)
	return container.New(layout.NewHBoxLayout(), scp.timeSelect, scp.timeUnitSelect, scp.ipmSelect)
}

func (scp *ScpDesc) newTriggerSelectionUI() (*fyne.Container, error) {
	triggerColor := scp.theme.Color(ColorNameGeneratorDisp, 0)
	for i := 0; i < int(scp.channelCount); i++ {
		if scp.Settings.Channels[i].TriggerSource {
			triggerColor = scp.Settings.Channels[i].Col[scp.Settings.ChannelColorIndex]
			break
		}
	}

	const fontScale = 0.7
	var err error
	scp.triggerThresholdDisp, err = disp7.NewCustomDisp7Array(5, 3, 20000, -20000,
		disp7.Signed, disp7.NoTrailingZeroes, scp.Window, triggerColor, disp7.ReadWrite,
		fontScale*disp7.DefaultDigitWidth, fontScale*disp7.DeafultDigitHeight,
		1, disp7.DefaultVCursorSpace, "Threshold :", " V")
	if err != nil {
		return nil, err
	}
	addToTest(scp.triggerThresholdDisp, triggerThresholdDispId)
	scp.triggerThresholdDisp.OnChanged = scp.onThresholdChange
	scp.triggerHysteresisDisp, err = disp7.NewCustomDisp7Array(5, 3, 20000, 0,
		disp7.SignedHidden, disp7.NoTrailingZeroes, scp.Window, triggerColor, disp7.ReadWrite,
		fontScale*disp7.DefaultDigitWidth, fontScale*disp7.DeafultDigitHeight,
		1, disp7.DefaultVCursorSpace, "Hysteresis:", " V")
	if err != nil {
		return nil, err
	}
	addToTest(scp.triggerHysteresisDisp, triggerHysteresisDispId)
	scp.triggerHysteresisDisp.OnChanged = scp.onHysteresisChange
	scp.boxTriggerHysteresisDisp = container.New(layout.NewHBoxLayout(), scp.triggerHysteresisDisp)
	if triggerTypes[scp.Settings.Trigger.Type] == control.Simple {
		scp.boxTriggerHysteresisDisp.Hide()
	}
	scp.triggerLowerThresholdDisp, err = disp7.NewCustomDisp7Array(5, 3, 20000, -20000,
		disp7.Signed, disp7.NoTrailingZeroes, scp.Window, triggerColor, disp7.ReadWrite,
		fontScale*disp7.DefaultDigitWidth, fontScale*disp7.DeafultDigitHeight,
		1, disp7.DefaultVCursorSpace, "Low Thres :", " V")
	if err != nil {
		return nil, err
	}
	scp.triggerLowerThresholdDisp.OnChanged = scp.onLowerThresholdChange

	scp.triggerLowerHysteresisDisp, err = disp7.NewCustomDisp7Array(5, 3, 20000, 0,
		disp7.SignedHidden, disp7.NoTrailingZeroes, scp.Window, triggerColor, disp7.ReadWrite,
		fontScale*disp7.DefaultDigitWidth, fontScale*disp7.DeafultDigitHeight,
		1, disp7.DefaultVCursorSpace, "Low Hyst  :", " V")
	if err != nil {
		return nil, err
	}
	scp.triggerLowerHysteresisDisp.OnChanged = scp.onLowerHysteresisChange
	scp.boxTriggerLowerDisp = container.New(layout.NewVBoxLayout(), scp.triggerLowerThresholdDisp, scp.triggerLowerHysteresisDisp)

	if triggerTypes[scp.Settings.Trigger.Type] != control.Window {
		scp.boxTriggerLowerDisp.Hide()
	}

	scp.triggerModeSelect = selectscroll.NewSelectScroll(triggerModeOptions, scp.onTriggerModeChange, triggerModeOptions[2])
	addToTest(scp.triggerModeSelect, triggerModeSelectId)
	scp.triggerModeSelect.SilentSetSelected(scp.Settings.Trigger.Mode)
	scp.triggerSettingMsg.Mode = triggerModes[scp.Settings.Trigger.Mode]

	// Build trigger type options based on whether complex trigger is enabled
	var activeTypeOptions []string
	if scp.ComplexTriggerEnabled {
		activeTypeOptions = []string{"Simple", "Advanced", "Window", "Complex"}
		triggerTypes["Complex"] = control.Complex
	} else {
		activeTypeOptions = []string{"Simple", "Advanced", "Window"}
		// If settings had Complex selected but flag is off, fall back to Advanced
		if scp.Settings.Trigger.Type == "Complex" {
			scp.Settings.Trigger.Type = "Advanced"
		}
	}
	scp.triggerTypeSelect = selectscroll.NewSelectScroll(activeTypeOptions, scp.onTriggerTypeChange, "Advanced")
	addToTest(scp.triggerTypeSelect, triggerTypeSelectId)
	scp.triggerTypeSelect.SilentSetSelected(scp.Settings.Trigger.Type)
	scp.triggerSettingMsg.Type = triggerTypes[scp.Settings.Trigger.Type]
	if scp.Settings.Trigger.Type == "Complex" {
		scp.buildComplexTriggerMessage()
	}

	boxMode := container.New(layout.NewHBoxLayout(), scp.triggerModeSelect, scp.triggerTypeSelect)
	boxThresh := container.New(layout.NewHBoxLayout(), scp.triggerThresholdDisp)
	scp.triggerDisplays = container.New(layout.NewVBoxLayout(), boxMode, boxThresh, scp.boxTriggerHysteresisDisp, scp.boxTriggerLowerDisp)
	return scp.triggerDisplays, nil
}

func (scp *ScpDesc) newTimeDivSettings() (box *fyne.Container, err error) {
	scp.newUnitList()
	box0 := scp.newTimeSelectionUI()
	// box1 := scp.newSamplingUI()
	triggerUI, err := scp.newTriggerSelectionUI()
	if err != nil {
		return nil, err
	}
	box = container.New(layout.NewVBoxLayout(), box0 /* box1,*/, triggerUI)
	scp.setMaxScreenTime()
	return box, nil
}

func (scp *ScpDesc) newSetTimeDivPanel(container *fyne.Container) (err error) {
	container.Add(layout.NewSpacer())
	var timeDivPanel *fyne.Container
	timeDivPanel, err = scp.newTimeDivSettings()
	container.Add(timeDivPanel)
	return
}
