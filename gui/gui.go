package gui

/*
VNA
Inputs:
  generator:
  fmin,fmax,multiplier,step,steptime,outputlevel

  channels (A,B):
  AC/DC, signal level, sample time interval

  display


SCOPE
*/
import (
	"fmt"
	"fynescope/control"
	"fynescope/control/scpi"
	"fynescope/disp7"
	"fynescope/genericps"
	"fynescope/selectscroll"
	"fynescope/settings"
	"image"
	"image/draw"
	"log/slog"
	"math"
	"strings"

	// "fynescope/sim"
	"strconv"

	// "fynescope/psi"
	// "fynescope/selectscroll"
	"sync"
	"time"

	"gonum.org/v1/gonum/dsp/fourier"

	"fyne.io/fyne/v2"
	// "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	maxSampling100M  = 100000000
	maxSampling200M  = 200000000
	maxSampling500M  = 500000000
	maxSampling1G    = 1000000000
	errorDisplayTime = 10 // seconds

	streamEnabledLabel  = "Stream: Enabled"
	streamDisabledLabel = "Stream: Disabled"

	// ErrFrequencyCannotBeDetected is returned when f(f) cannot lock onto the signal.
	ErrFrequencyCannotBeDetected = "Frequency cannot be detected"
	ErrWrongFfTrigger            = "Error: f(f) requires Simple, Advanced, or Window trigger"
)
const (
	ftTabIndex = iota
	fvTabIndex
	dftTabIndex
	ffTabIndex
	rlcTabIndex
	filterTabIndex
	genTabIndex
	extgenTabIndex
)

var (
	tabNames = []string{"f(t)", "f(v)", "FFT", "f(f)", "RLC", "filter", "gen", "extgen"}

	dontCare genericps.ChannelId = -1 // trigger is disabled
	chA                          = genericps.ChA
	chB                          = genericps.ChB
	chC                          = genericps.ChC
	chD                          = genericps.ChD
)

type (
	timeBaseDescs []struct {
		timeBase     uint32
		maxSamples   int32
		timeInterval int32
	}
	triggerDesc struct {
		enabled    bool
		threshold  int16
		hysteresis int32
		mv         int32
	}

	ScpDesc struct {
		Measure                             MeasureDesc
		running                             bool
		repartition, themeChanged           chan struct{}
		triggerSettingMsg                   control.TriggerDescMsg
		rangeMargin                         float32
		controlSamplingTimeInterval         float64
		controlXRoundError                  float64
		maxScreenTime                       float64
		signalTime                          float64
		App                                 fyne.App
		theme                               fyne.Theme
		Window, genWindow                   fyne.Window
		MaxChannel, triggerSource           genericps.ChannelId
		channelCount                        genericps.NumOfChannelEnum
		lastRange                           genericps.RangeEnum
		ratioMode                           genericps.RatioMode
		displayMovedDivs, timeDiv, timeUnit int // timeUnit: -12: ps, -9: ns, -6: us, -3:ms, 0: s
		MaxValue, MinValue                  int16
		controlTriggerTimeOffset            int64
		dftScopeFullScreen                  rasterImage
		dftScopeSignalScreen                rasterImage
		ftScopeFullScreen                   rasterImage
		ftScopeSignalScreen                 rasterImage
		ftPersistentLayers                  []*image.RGBA
		dftPersistentLayers                 []*image.RGBA
		delayedClearTimer                   *time.Timer
		GenFreqDelayStr                     string
		GenFreqStepStr                      string
		GenFreqStr                          string
		fvScopeFullScreen                   rasterImage
		fvScopeSignalScreen                 rasterImage
		ffScopeFullScreen                   rasterImage
		ffScopeSignalScreen                 rasterImage
		ffFullRefresh                       bool
		screenLocker                        sync.Mutex
		ffLocker                            sync.Mutex
		ffSweepQuit                         chan struct{}
		ffSweepDataReady                    chan struct{}
		ffSweepAcquireTime                  time.Time
		ffBufferDone                        chan struct{}
		currentFfFreq                       float64
		measuredFfFreq                      float64
		// statusChan / statusQuit for the status display goroutine
		statusChan chan string
		statusQuit chan struct{}
		// FFT caches for processFfData — reallocated only when sample count changes
		ffFftObj             *fourier.FFT
		ffFftBuf             []float64
		ffFftResult          []complex128
		ffFftSamples         int
		ffCurrentFreqDisp    *disp7.DigitArray
		ffMinFreqDisp        *disp7.DigitArray
		ffMaxFreqDisp        *disp7.DigitArray
		ffStepFreqDisp       *disp7.DigitArray
		ffDeltaTDisp         *disp7.DigitArray
		ffAmpDisp            *disp7.DigitArray
		ffOffsetDisp         *disp7.DigitArray
		ffPhaseDisp          *disp7.DigitArray
		ffRaiseFallTimeDisp  *disp7.DigitArray
		ffNoiseAmplitudeDisp *disp7.DigitArray
		ffPhaseNoiseDisp     *disp7.DigitArray

		bodeBuffers                [genericps.MaxChannel][]bodePoint
		maxSamplingRate            uint32
		segmentIndex               uint32 // maxSamplingRate: sample/sec
		controlTab                 *container.AppTabs
		dftTab                     *container.TabItem
		fraTab                     *container.TabItem
		ftTab                      *container.TabItem
		fvTab                      *container.TabItem
		ffTab                      *container.TabItem
		genTab                     *container.TabItem
		rlcTab                     *container.TabItem
		filterTab                  *container.TabItem
		extgenTab                  *container.TabItem
		setTab                     *container.TabItem
		psControl                  *control.PscDesc
		triggerHysteresisDisp      *disp7.DigitArray
		triggerThresholdDisp       *disp7.DigitArray
		boxTriggerHysteresisDisp   *fyne.Container
		triggerLowerThresholdDisp  *disp7.DigitArray
		triggerLowerHysteresisDisp *disp7.DigitArray
		// boxTriggerLowerDisp removed
		intervalTypeSelect           *selectscroll.SelectScroll
		intervalUnitSelect           *selectscroll.SelectScroll
		intervalTimeLowerDisp        *disp7.DigitArray
		intervalTimeUpperDisp        *disp7.DigitArray
		intervalTimeSingleDisp       *disp7.DigitArray
		boxTriggerIntervalDisp       *fyne.Container
		boxIntervalTimeSingle        *fyne.Container
		boxIntervalTimeRange         *fyne.Container
		digital                      *fyne.Container
		genLayout                    *fyne.Container
		rlcLayout                    *fyne.Container
		filterLayout                 *fyne.Container
		extgenLayout                 *fyne.Container
		extgenWindow                 fyne.Window
		triggerDisplays              *fyne.Container
		dftRaster                    *screenRaster
		ftRaster                     *screenRaster
		fvRaster                     *screenRaster
		ffRaster                     *screenRaster
		fvViewer                     *fvViewer
		ffViewer                     drawer // actually *ffViewer but drawer interface helps compilation here if not yet defined
		ipmSelect                    *selectscroll.SelectScroll
		sampleRateSelect             *selectscroll.SelectScroll
		sampleUnitSelect             *selectscroll.SelectScroll
		timeSelect                   *selectscroll.SelectScroll
		timeUnitSelect               *selectscroll.SelectScroll
		triggerCalculationModeSelect *selectscroll.SelectScroll
		triggerModeSelect            *selectscroll.SelectScroll
		triggerTypeSelect            *selectscroll.SelectScroll
		Settings                     *settings.PsSettings
		runblockButton               *widget.Button
		toolbar                      *fyne.Container
		streamEnableButton           *widget.Button
		etsCycles                    *widget.Entry
		// actualSampleTime                    *widget.Label
		status                     *widget.Label
		triggerCheck               []*widget.Check
		displayBuffers             [][]float32 // signal stored in mv
		channelViewers             []channelViewerDesc
		dftDrawers                 []drawer
		ftDrawers                  []drawer
		fvDrawers                  []drawer
		ffDrawers                  []drawer
		ftBottomLabelViewer        drawer
		dftBottomLabelViewer       drawer
		triggerPoint               drawer
		etsBuffer                  []int64
		triggerSources             []string
		dftDivsX                   [numberOfDivs + 1]float32
		dftDivsY                   [numberOfDivs + 1]float32
		ftDivsX                    [numberOfDivs + 1]float32
		ftDivsY                    [numberOfDivs + 1]float32
		fvDivsX                    [numberOfDivs + 1]float32
		fvDivsY                    [numberOfDivs + 1]float32
		binWidthLabel              *widget.Label
		dftDataCollectionTimeLabel *widget.Label
		dftSampleRateSelect        *selectscroll.SelectScroll
		dftSampleUnitSelect        *selectscroll.SelectScroll
		dftMaxFreqValSelect        *selectscroll.SelectScroll
		dftMaxFreqUnitSelect       *selectscroll.SelectScroll
		SettingFileName            string
		extGen                     control.ExtGenDesc
		ExtGenEnabled              bool
		useExtGenCheck             *widget.Check
		complexTriggerCheck        *widget.Check
		timeZoomButton             *widget.Button
		timeZoomWindow             fyne.Window
		timeZoomRaster             *screenRaster
		timeZoomMaxScreenTime      float64
		timeZoomScopeFullScreen    rasterImage
		timeZoomScopeSignalScreen  rasterImage
		timeZoomDrawers            []drawer
		timeZoomDivsX              [numberOfDivs + 1]float32
		timeZoomDivsY              [numberOfDivs + 1]float32
	}
)

func createFlag() (ch chan struct{}) {
	ch = make(chan struct{}, 1)
	return
}
func setFlag(flag chan struct{}) {
	select {
	case flag <- struct{}{}:
	default:
	}
}

func getFlag(flag chan struct{}) bool {
	select {
	case <-flag:
		return true
	default:
		return false
	}
}

func (scp *ScpDesc) addFtDrawer(d drawer) {
	scp.ftDrawers = append(scp.ftDrawers, d)
}

func (scp *ScpDesc) deleteFtDrawer(d drawer) {
	if len(scp.ftDrawers) == 0 {
		return
	}
	i := 0
	for i < len(scp.ftDrawers) && scp.ftDrawers[i] != d {
		i++
		if i == len(scp.ftDrawers) {
			return
		}
	}
	scp.ftDrawers = scp.ftDrawers[:i+copy(scp.ftDrawers[i:], scp.ftDrawers[i+1:])]
}

func clearRGBA(img *image.RGBA) {
	if img == nil {
		return
	}
	for i := range img.Pix {
		img.Pix[i] = 0
	}
}

func (scp *ScpDesc) scheduleClearAllPersistentLayers() {
	if scp.delayedClearTimer != nil {
		scp.delayedClearTimer.Stop()
	}
	scp.delayedClearTimer = time.AfterFunc(500*time.Millisecond, func() {
		fyne.Do(func() {
			for i := range scp.ftPersistentLayers {
				clearRGBA(scp.ftPersistentLayers[i])
			}
			for i := range scp.dftPersistentLayers {
				clearRGBA(scp.dftPersistentLayers[i])
			}
		})
	})
}

func (scp *ScpDesc) clearFtPersistentLayer(chIndex genericps.ChannelId) {
	if int(chIndex) < len(scp.ftPersistentLayers) {
		clearRGBA(scp.ftPersistentLayers[chIndex])
	}
	scp.scheduleClearAllPersistentLayers()
}

func (scp *ScpDesc) clearAllFtPersistentLayers() {
	for i := range scp.ftPersistentLayers {
		clearRGBA(scp.ftPersistentLayers[i])
	}
	scp.scheduleClearAllPersistentLayers()
}

func (scp *ScpDesc) clearDftPersistentLayer(chIndex genericps.ChannelId) {
	if int(chIndex) < len(scp.dftPersistentLayers) {
		clearRGBA(scp.dftPersistentLayers[chIndex])
	}
	scp.scheduleClearAllPersistentLayers()
}

func (scp *ScpDesc) clearAllDftPersistentLayers() {
	for i := range scp.dftPersistentLayers {
		clearRGBA(scp.dftPersistentLayers[i])
	}
	scp.scheduleClearAllPersistentLayers()
}

func (scp *ScpDesc) addDftDrawer(d drawer) {
	scp.dftDrawers = append(scp.dftDrawers, d)
}

func (scp *ScpDesc) deleteDftDrawer(d drawer) {
	if len(scp.dftDrawers) == 0 {
		return
	}
	i := 0
	for i < len(scp.dftDrawers) && scp.dftDrawers[i] != d {
		i++
		if i == len(scp.dftDrawers) {
			return
		}
	}
	scp.dftDrawers = scp.dftDrawers[:i+copy(scp.dftDrawers[i:], scp.dftDrawers[i+1:])]
}

func (scp *ScpDesc) addFvDrawer(d drawer) {
	scp.fvDrawers = append(scp.fvDrawers, d)
}

func (scp *ScpDesc) deleteFvDrawer(d drawer) {
	if len(scp.fvDrawers) == 0 {
		return
	}
	i := 0
	for i < len(scp.fvDrawers) && scp.fvDrawers[i] != d {
		i++
		if i == len(scp.fvDrawers) {
			return
		}
	}
	scp.fvDrawers = scp.fvDrawers[:i+copy(scp.fvDrawers[i:], scp.fvDrawers[i+1:])]
}

func (scp *ScpDesc) SaveSettings() {
	if scp.SettingFileName == "" {
		return
	}
	if err := settings.Save(scp.SettingFileName, scp.Settings); err != nil {
		slog.Error("failed to save settings", "err", err)
	}
}

func (scp *ScpDesc) refreshRasters() {
	if scp.Settings == nil {
		return
	}
	targetFunction := scp.Settings.Window.Function
	if targetFunction == genTabIndex || targetFunction == filterTabIndex || targetFunction == extgenTabIndex {
		targetFunction = scp.Settings.Window.LastDispFunction
	}

	switch targetFunction {
	case dftTabIndex:
		if scp.dftRaster != nil {
			canvas.Refresh(scp.dftRaster)
		}
	case fvTabIndex:
		if scp.fvRaster != nil {
			canvas.Refresh(scp.fvRaster)
		}
	case ffTabIndex:
		if scp.ffRaster != nil {
			canvas.Refresh(scp.ffRaster)
		}
	default:
		if scp.ftRaster != nil {
			canvas.Refresh(scp.ftRaster)
		}
		if scp.timeZoomRaster != nil {
			canvas.Refresh(scp.timeZoomRaster)
		}
	}
}

func (scp *ScpDesc) adcToMv(raw float64, chRange genericps.RangeEnum) float64 {
	return (math.Round(float64(raw) * float64(genericps.InputRanges[chRange]) / float64(scp.MaxValue)))
}
func (scp *ScpDesc) mvToAdc(mv int32, chRange genericps.RangeEnum) int32 {
	adc := int32(math.Round(float64(mv)*float64(scp.MaxValue)) / float64(genericps.InputRanges[chRange]))
	if adc > math.MaxInt16 {
		adc = math.MaxInt16 - 1
	} else if adc < math.MinInt16 {
		adc = math.MinInt16 + 1
	}
	return adc
}

func (scp *ScpDesc) mvToUAdc(mv int32, chRange genericps.RangeEnum) int32 {
	adc := int32(math.Round(float64(mv)*float64(scp.MaxValue)) / float64(genericps.InputRanges[chRange]))
	if adc > math.MaxUint16 {
		adc = math.MaxInt16 - 1
	}
	return adc
}

func (scp *ScpDesc) getScreenScale() float32 {
	if scp.Settings == nil || scp.Settings.ScreenSize == "" {
		return 1.0
	}
	switch scp.Settings.ScreenSize {
	case settings.ScreenSize1920x1080:
		return 1.0
	case settings.ScreenSize1366x768:
		return 0.71
	case settings.ScreenSize1280x720:
		return 0.66
	case settings.ScreenSize1024x768:
		return 0.53
	default:
		return 1.0
	}
}

func (scp *ScpDesc) getScreenDimensions() (float32, float32) {
	if scp.Settings == nil || scp.Settings.ScreenSize == "" {
		return 1920, 1000
	}
	parts := strings.Split(scp.Settings.ScreenSize, "x")
	if len(parts) == 2 {
		w, err1 := strconv.ParseFloat(parts[0], 32)
		h, err2 := strconv.ParseFloat(parts[1], 32)
		if err1 == nil && err2 == nil {
			if w == 1920 && h == 1080 {
				return 1920, 1000
			}
			return float32(w), float32(h)
		}
	}
	return 1920, 1000
}

func (scp *ScpDesc) build2000Gui() {
	var (
		themeChangeAction *widget.Button
		changeSide        *widget.Button
		restoreScreen     *widget.Button
		fullScreen        *widget.Button
		logout            *widget.Button
		content           *fyne.Container
	)

	ftLayout := container.New(layout.NewVBoxLayout())
	scp.genLayout = container.New(layout.NewVBoxLayout())
	scp.ftTab = container.NewTabItem(tabNames[ftTabIndex], ftLayout)
	fvControls := container.New(layout.NewVBoxLayout())
	scp.genTab = container.NewTabItem(tabNames[genTabIndex], scp.genLayout)
	scp.fvTab = container.NewTabItem(tabNames[fvTabIndex], fvControls)
	dftLayout := container.New(layout.NewVBoxLayout())
	scp.dftTab = container.NewTabItem(tabNames[dftTabIndex], dftLayout)
	ffLayout := container.New(layout.NewVBoxLayout())
	scp.ffTab = container.NewTabItem(tabNames[ffTabIndex], ffLayout)
	scp.rlcLayout = container.New(layout.NewVBoxLayout())
	scp.rlcTab = container.NewTabItem(tabNames[rlcTabIndex], scp.rlcLayout)
	scp.filterLayout = container.NewMax()
	scp.filterTab = container.NewTabItem(tabNames[filterTabIndex], scp.filterLayout)
	scp.extgenLayout = container.New(layout.NewVBoxLayout())
	scp.extgenTab = container.NewTabItem(tabNames[extgenTabIndex], scp.extgenLayout)
	scp.controlTab = container.NewAppTabs(
		scp.ftTab, scp.fvTab, scp.dftTab, scp.ffTab, scp.rlcTab, scp.filterTab, scp.genTab, scp.extgenTab)

	if scp.psControl != nil && scp.psControl.Con.ID != genericps.SimId {
		scp.controlTab.Remove(scp.rlcTab)
	}
	if !scp.ExtGenEnabled {
		scp.controlTab.Remove(scp.extgenTab)
	}

	activeRasterContainer := container.NewMax(scp.ftRaster, scp.dftRaster, scp.fvRaster, scp.ffRaster)
	scp.controlTab.OnSelected = func(t *container.TabItem) {
		prevTab := scp.Settings.Window.Function
		newTab := scp.controlTab.SelectedIndex()
		scp.handleTabTransition(prevTab, newTab)

		if scp.Settings.Window.LastDispFunction != scp.Settings.Window.Function {
			switch scp.Settings.Window.Function {
			case ftTabIndex, fvTabIndex, dftTabIndex, ffTabIndex:
				scp.Settings.Window.LastDispFunction = scp.Settings.Window.Function
			}
		}
		scp.Settings.Window.Function = newTab
		slog.Debug("tab", "t", *t)

		targetFunction := scp.Settings.Window.Function
		if scp.controlTab.Selected() == scp.genTab ||
			scp.controlTab.Selected() == scp.filterTab ||
			scp.controlTab.Selected() == scp.extgenTab {
			targetFunction = scp.Settings.Window.LastDispFunction
		}

		switch targetFunction {
		case dftTabIndex:
			scp.ftRaster.Hide()
			scp.fvRaster.Hide()
			scp.ffRaster.Hide()
			scp.dftRaster.Show()
			if scp.timeZoomButton != nil {
				scp.timeZoomButton.Hide()
			}
		case fvTabIndex:
			scp.ftRaster.Hide()
			scp.dftRaster.Hide()
			scp.ffRaster.Hide()
			scp.fvRaster.Show()
			if scp.timeZoomButton != nil {
				scp.timeZoomButton.Hide()
			}
		case ffTabIndex:
			scp.ftRaster.Hide()
			scp.dftRaster.Hide()
			scp.fvRaster.Hide()
			scp.ffRaster.Show()
			if scp.timeZoomButton != nil {
				scp.timeZoomButton.Hide()
			}
		case rlcTabIndex:
			scp.ftRaster.Show()
			scp.dftRaster.Hide()
			scp.fvRaster.Hide()
			scp.ffRaster.Hide()
			if scp.timeZoomButton != nil {
				scp.timeZoomButton.Show()
			}
		default:
			scp.dftRaster.Hide()
			scp.fvRaster.Hide()
			scp.ffRaster.Hide()
			scp.ftRaster.Show()
			if scp.timeZoomButton != nil {
				scp.timeZoomButton.Show()
			}
		}

		scp.updateAcquisitionParameters()
		activeRasterContainer.Refresh()

		if scp.running {

			if targetFunction == fvTabIndex || targetFunction == ffTabIndex {
				if targetFunction == ffTabIndex && (scp.Settings.Trigger.Type == settings.TriggerTypeInterval || scp.Settings.Trigger.Type == settings.TriggerTypePulseWidth) {
					scp.StopRunning()
					scp.psControl.DisplayStatus(ErrWrongFfTrigger, control.Warning)
				} else {
					// Force block mode and ensure a trigger is set for f(v) and f(f)
					scp.triggerSettingMsg.Mode = control.Auto
					if scp.triggerSource == dontCare {
						// Set arbitrary simple trigger on ChA if none selected
						scp.triggerSource = chA
						scp.Settings.Channels[chA].TriggerSource = true
						scp.triggerSettingMsg.Source = chA
						scp.triggerSettingMsg.Enabled = true
						scp.triggerSettingMsg.Mv = 0
					}
				}
			}

		}
	}
	scp.controlTab.SelectTabIndex(scp.Settings.Window.Function)

	// Ensure the correct raster is displayed on startup since SelectTabIndex may not trigger OnSelected if already at 0
	targetFunctionInit := scp.Settings.Window.Function
	if scp.Settings.Window.Function == genTabIndex ||
		scp.Settings.Window.Function == filterTabIndex ||
		scp.Settings.Window.Function == extgenTabIndex {
		targetFunctionInit = scp.Settings.Window.LastDispFunction
	}
	switch targetFunctionInit {
	case dftTabIndex:
		scp.ftRaster.Hide()
		scp.fvRaster.Hide()
		scp.ffRaster.Hide()
	case fvTabIndex:
		scp.ftRaster.Hide()
		scp.dftRaster.Hide()
		scp.ffRaster.Hide()
	case ffTabIndex:
		scp.ftRaster.Hide()
		scp.dftRaster.Hide()
		scp.fvRaster.Hide()
	default:
		scp.dftRaster.Hide()
		scp.fvRaster.Hide()
		scp.ffRaster.Hide()
	}

	scp.timeZoomButton = widget.NewButtonWithIcon("", theme.SearchIcon(), func() {
		scp.openTimeZoomWindow()
	})
	if targetFunctionInit != ftTabIndex && targetFunctionInit != rlcTabIndex && targetFunctionInit != genTabIndex && targetFunctionInit != filterTabIndex && targetFunctionInit != extgenTabIndex {
		scp.timeZoomButton.Hide()
	}

	addToTest(scp.controlTab, ftFuncId)
	addToTest(scp.controlTab, fvFuncId)
	addToTest(scp.controlTab, dftFuncId)
	addToTest(scp.controlTab, ffFuncId)
	addToTest(scp.controlTab, rlcFuncId)
	addToTest(scp.controlTab, genFuncId)
	addToTest(scp.controlTab, filterFuncId)
	addToTest(scp.controlTab, extgenFuncId)
	scp.newChannelPanels(ftLayout)
	scp.newSetTimeDivPanel(ftLayout)

	scp.newFvPanel(fvControls)
	scp.newDftPanel(dftLayout)
	scp.newFfPanel(ffLayout)
	scp.newRlcPanel(scp.rlcLayout)
	scp.newDigitalFilterPanel(scp.filterLayout)
	if scp.ExtGenEnabled {
		scp.extgenLayout.Add(scp.newExtGenTab(true))
	}
	left := container.New(layout.NewVBoxLayout())
	themeChangeAction = widget.NewButtonWithIcon("", theme.CheckButtonIcon(), func() {
		if scp.theme == Theme(settings.DarkTheme) {
			scp.theme = Theme(settings.LightTheme)
			scp.Settings.Theme = settings.LightTheme
			scp.Settings.ChannelColorIndex = settings.LightChannel
			themeChangeAction.SetIcon(theme.CheckButtonIcon())
			for i := range scp.channelViewers {
				scp.channelViewers[i].enableCheckbox.Refresh()
			}
			fyne.CurrentApp().Settings().SetTheme(scp.theme)
		} else {
			scp.theme = Theme(settings.DarkTheme)
			scp.Settings.ChannelColorIndex = settings.DarkChannel
			scp.Settings.Theme = settings.DarkTheme
			themeChangeAction.SetIcon(theme.CheckButtonIcon())
			fyne.CurrentApp().Settings().SetTheme(scp.theme)
		}
		for i := range scp.channelViewers {
			col := scp.Settings.Channels[i].Col[scp.Settings.ChannelColorIndex]
			scp.channelViewers[i].enableCheckbox.SetColor(col)
			scp.SetChannelColors(col, genericps.ChannelId(i))
		}
		setFlag(scp.themeChanged)
		slog.Debug("themeChanged")
		if scp.toolbar != nil {
			scp.toolbar.Refresh()
		}
		scp.refreshRasters()
		scp.SaveSettings()
	})
	addToTest(themeChangeAction, themeChangeActionId)

	scp.streamEnableButton = widget.NewButton(streamEnabledLabel, func() {
		if scp.psControl == nil {
			return
		}
		newValue := !scp.psControl.StreamEnabled.Load()
		scp.psControl.StreamEnabled.Store(newValue)
		if scp.Settings != nil {
			scp.Settings.StreamEnabled = &newValue
			scp.SaveSettings()
		}
		scp.updateStreamButtonState()
		if scp.running {
			scp.psControl.RequestRestart()
		}
	})
	scp.updateStreamButtonState()

	scp.runblockButton = widget.NewButtonWithIcon("", theme.MediaPlayIcon(), func() {
		if !scp.running {
			scp.clearAllFtPersistentLayers()
			scp.clearAllDftPersistentLayers()
			if scp.controlTab.SelectedIndex() == fvTabIndex || scp.controlTab.SelectedIndex() == ffTabIndex {
				// Force block mode and ensure a trigger is set for f(v) and f(f)
				scp.triggerSettingMsg.Mode = control.Auto
				if scp.triggerSource == dontCare {
					// Set arbitrary simple trigger on ChA if none selected
					scp.triggerSource = chA
					scp.Settings.Channels[chA].TriggerSource = true
					scp.triggerSettingMsg.Source = chA
					scp.triggerSettingMsg.Enabled = true
					scp.triggerSettingMsg.Mv = 0
				}
			}
			if scp.status.Text == ErrFrequencyCannotBeDetected {
				scp.psControl.DisplayStatus("", control.Info)
			}
			if scp.controlTab.SelectedIndex() == ffTabIndex {
				if scp.Settings.Trigger.Type == settings.TriggerTypeInterval || scp.Settings.Trigger.Type == settings.TriggerTypePulseWidth {
					scp.psControl.DisplayStatus(ErrWrongFfTrigger, control.Warning)
					return
				}
				if scp.Settings.Ff.PtsDec <= 0 {
					scp.psControl.DisplayStatus("Error: Points per decade cannot be 0", control.Warning)
					return
				}
				// Set up the generator in non-sweep mode; the app controls stepping.
				if scp.psControl.Con != nil && scp.psControl.Con.ID == genericps.SimId {
					scp.applyFfSimGenSettings(false)
					scp.applyFfSimGenSettings(scp.Settings.FfGen.On)
				} else {
					scp.applyFfGenSettings(false)
					scp.applyFfGenSettings(scp.Settings.FfGen.On)
				}
			} else {
				if scp.Settings.GenPanel.On {
					scp.applyInternalGenSettings(true)
				}
				for i := 0; i < int(scp.channelCount); i++ {
					scp.applySimGenSettings(genericps.ChannelId(i), &scp.Settings.SimGenPanel[i])
				}
			}
			if scp.Settings.Ff.UseExternalGen && scp.extGen.Connected() {
				// External generator frequency will be set via setGeneratorFreq during the sweep.
			}

			if scp.controlTab.SelectedIndex() == ffTabIndex {
				scp.ResetFfSweep()
				// User requirement: When f(f) is selected then the run button starts the generators that are checked on.
				// Start the application-controlled logarithmic frequency sweep.
				scp.startFfSweep()
			}

			// Ensure acquisition parameters (screen time, sample width) match the
			// currently selected tab before starting the capture.  Without this,
			// pressing Run on the f(f) tab for the first time would use stale
			// f(t) parameters and the Bode sweep would not start.
			scp.updateAcquisitionParameters()

			var err error
			scp.running = true
			scp.runblockButton.SetIcon(theme.MediaPauseIcon())
			switch scp.triggerSettingMsg.Mode {
			case control.ETS:

				err = scp.psControl.SetETSMode()
			default: // Auto, Repeat, Single, and our forced Auto for f(v)
				err = scp.psControl.SetBlockMode()
				if scp.triggerSettingMsg.Mode == control.Single {
					scp.runblockButton.SetIcon(theme.MediaPlayIcon())
					scp.running = false
				}
			}
			if err == nil {

				// set unit
				// change default from GS/s to nsUsMssubSetIndex
				// set when running and timing changed

			} else {
				scp.runblockButton.SetIcon(theme.MediaPlayIcon())
				slog.Error("", "run error:", err)
			}
		} else {
			scp.StopRunning()
		}
	})
	addToTest(scp.runblockButton, runblockButtonId)
	setfullscreen := func() {
		scp.Settings.Window.Fullscreen = true
		scp.Window.SetFullScreen(true)
	}
	setnofullscreen := func() {
		scp.Settings.Window.Fullscreen = false
		scp.Window.SetFullScreen(false)
	}
	scp.initStatus()
	changeSideFunc := func() {
		if changeSide.Icon == theme.NavigateBackIcon() {
			scp.Settings.Window.LeftControl = true
			scp.toolbar.RemoveAll()
			scp.toolbar.Add(scp.runblockButton)
			scp.toolbar.Add(scp.streamEnableButton)
			scp.toolbar.Add(scp.timeZoomButton)
			scp.toolbar.Add(fullScreen)
			scp.toolbar.Add(restoreScreen)
			scp.toolbar.Add(changeSide)
			scp.toolbar.Add(themeChangeAction)
			scp.toolbar.Add(logout)
			scp.toolbar.Add(layout.NewSpacer())
			scp.toolbar.Add(scp.status)
			content = container.NewBorder(scp.toolbar, nil, scp.controlTab, left, activeRasterContainer)
			changeSide.SetIcon(theme.NavigateNextIcon())
		} else {
			scp.Settings.Window.LeftControl = false
			scp.toolbar.RemoveAll()
			scp.toolbar.Add(scp.status)
			scp.toolbar.Add(layout.NewSpacer())
			scp.toolbar.Add(scp.runblockButton)
			scp.toolbar.Add(scp.streamEnableButton)
			scp.toolbar.Add(scp.timeZoomButton)
			scp.toolbar.Add(fullScreen)
			scp.toolbar.Add(restoreScreen)
			scp.toolbar.Add(changeSide)
			scp.toolbar.Add(themeChangeAction)
			scp.toolbar.Add(logout)
			content = container.NewBorder(scp.toolbar, nil, left, scp.controlTab, activeRasterContainer)
			changeSide.SetIcon(theme.NavigateBackIcon())
		}
		scp.toolbar.Refresh()
		scp.Window.SetContent(content)
	}
	fullScreen = widget.NewButtonWithIcon("", theme.ViewFullScreenIcon(), setfullscreen)
	restoreScreen = widget.NewButtonWithIcon("", theme.ViewRestoreIcon(), setnofullscreen)
	changeSide = widget.NewButtonWithIcon("", theme.NavigateBackIcon(), changeSideFunc)
	if scp.Settings.Window.LeftControl {
		changeSide.SetIcon(theme.NavigateNextIcon())
	}
	addToTest(changeSide, changeSideId)
	logout = widget.NewButtonWithIcon("", theme.LogoutIcon(), func() {
		if scp.psControl != nil {
			scp.psControl.Shutdown()
		}
		if scp.statusQuit != nil {
			close(scp.statusQuit)
			scp.statusQuit = nil
		}
		scp.App.Quit()
	})
	if scp.Settings.Window.LeftControl {
		scp.toolbar = container.New(layout.NewHBoxLayout(), scp.runblockButton, scp.streamEnableButton, scp.timeZoomButton, fullScreen, restoreScreen, changeSide,
			themeChangeAction,
			logout,
			layout.NewSpacer(),
			scp.status)
		content = container.NewBorder(scp.toolbar, nil, scp.controlTab, left, activeRasterContainer)
	} else {
		scp.toolbar = container.New(layout.NewHBoxLayout(), scp.status, layout.NewSpacer(),
			scp.runblockButton, scp.streamEnableButton, scp.timeZoomButton, fullScreen, restoreScreen, changeSide,
			themeChangeAction,
			logout)
		content = container.NewBorder(scp.toolbar, nil, left, scp.controlTab, activeRasterContainer)
	}

	scp.updateStreamButtonVisibility()

	scp.psControl.BufferCallback = func(size int) {
		for i := range scp.displayBuffers {
			if size != len(scp.displayBuffers[i]) {
				if cap(scp.displayBuffers[i]) >= size {
					scp.displayBuffers[i] = scp.displayBuffers[i][:size]
				} else {
					scp.displayBuffers[i] = make([]float32, size)
				}
			}
		}
	}

	scp.psControl.EtsBufferCallback = func(size int) {
		scp.psControl.BufferCallback(size)
		if size > len(scp.etsBuffer) {
			if cap(scp.etsBuffer) >= size {
				scp.etsBuffer = scp.etsBuffer[:size]
			} else {
				scp.etsBuffer = make([]int64, size)
			}
		} else {
			scp.etsBuffer = scp.etsBuffer[:size]
		}
	}

	scp.psControl.RefreshEtsCallback = func(buffers [][]int16, etsInBuffer []int64, xRoundError float64) {
		copy(scp.etsBuffer, etsInBuffer)
		scp.psControl.RefreshCallback(buffers, 0, xRoundError, 0)
	}

	scp.psControl.RefreshCallback = func(buffers [][]int16, triggerTimeOffset int64,
		xRoundError, samplingTimeInterval float64) {
		scp.controlXRoundError = xRoundError
		scp.controlTriggerTimeOffset = triggerTimeOffset
		scp.controlSamplingTimeInterval = samplingTimeInterval

		fyne.Do(func() {
			scp.UpdateMeasurements(buffers, samplingTimeInterval)
			scp.updateBinWidth()
			scp.updateDftDataCollectionTime()
			scp.refreshRasters() // it calls draw method in signalviewer
		})
	}

	for i := range scp.channelViewers {
		if scp.Settings.Channels[i].Enabled {
			scp.channelViewers[i].enableCheckbox.Set()
			if scp.channelViewers[i].dftCheckbox != nil {
				scp.channelViewers[i].dftCheckbox.SetChecked(true)
			}
			scp.channelViewers[i].triggerCheckbox.SetChecked(scp.Settings.Channels[i].TriggerSource)
			scp.channelViewers[i].displayOffsetInt = scp.Settings.Channels[i].DisplayVOffset
			scp.channelViewers[i].dftDisplayOffsetInt = scp.Settings.Channels[i].DftDisplayVOffset
		}
	}
	scp.Window.SetContent(content)
}

// StopRunning stops the current capture or sweep operation and updates the run button UI.
func (scp *ScpDesc) StopRunning() {
	scp.stopFfSweep() // stop any running Bode sweep
	err := scp.psControl.Stop()
	if err == nil {
		scp.runblockButton.SetIcon(theme.MediaPlayIcon())
		scp.running = false
	} else {
		slog.Error("", "Stop returned", err)
	}
}

func (scp *ScpDesc) build2407Gui() {
	scp.build2000Gui()

	scp.newGenPanel(scp.genLayout)
}

func (scp *ScpDesc) build2000IMGui() {

	scp.build2000Gui()
	scp.newSimGenPanel(scp.genLayout, true)

}

func (scp *ScpDesc) SetVariant() (err error) {

	scp.psControl.Info, err = scp.psControl.UnitVariantInfo()
	slog.Info("scope ", "info string", scp.psControl.Info)
	// TODO select preconfigured gui description, including
	// channel count
	// e.g. ETS mode is available, voltage, frequency ranges
	// add refresh rate setting
	if err != nil {
		return
	}
	switch string(scp.psControl.Info[1]) {
	case "1":
		scp.MaxChannel = genericps.ChA
		scp.channelCount = 1
		scp.triggerSources = []string{"ChA", "None"}
	case "2":
		scp.MaxChannel = genericps.ChB
		scp.channelCount = 2
		scp.triggerSources = []string{"ChA", "ChB", "EXT", "None"}
	case "3":
		scp.MaxChannel = genericps.ChC
		scp.channelCount = 3
		scp.triggerSources = []string{"ChA", "ChB", "ChC", "None"}
	case "4":
		scp.channelCount = 4
		scp.MaxChannel = genericps.ChD
		scp.triggerSources = []string{"ChA", "ChB", "ChC", "ChD", "None"}
	default:
		scp.triggerSources = []string{"ChA", "ChB", "EXT", "None"}
		err = fmt.Errorf("getInfo: unknown variant info %s", scp.psControl.Info)
		return
	}
	scp.displayBuffers = make([][]float32, scp.channelCount)
	scp.psControl.MaxSamplingRate = scp.maxSamplingRate
	scp.channelViewers = make([]channelViewerDesc, scp.channelCount)
	scp.ftPersistentLayers = make([]*image.RGBA, scp.channelCount)
	scp.dftPersistentLayers = make([]*image.RGBA, scp.channelCount)
	scp.MinValue, scp.MaxValue, err = scp.psControl.MinMaxValues()

	switch scp.psControl.Info {
	case "2107SIM", "2207SIM", "2307SIM", "2407SIM":
		scp.maxSamplingRate = maxSampling1G
		scp.build2000IMGui()
	case "2204A":
		slog.Warn("2204A not tested")
		scp.maxSamplingRate = maxSampling100M
	case "2205A":
		slog.Warn("2205A not tested")
		scp.maxSamplingRate = maxSampling200M
	case "2206B":
		slog.Warn("2206B not tested")
		scp.maxSamplingRate = maxSampling500M
	case "2207B":
		slog.Warn("2207B not tested")
		scp.maxSamplingRate = maxSampling1G
	case "2208B":
		slog.Warn("2208B not tested")
		scp.maxSamplingRate = maxSampling1G
	case "2405A":
		slog.Warn("2208B not tested")
		scp.maxSamplingRate = maxSampling500M
	case "2406B":
		slog.Warn("2406B not tested")
		scp.maxSamplingRate = maxSampling1G
	case "2407B":
		scp.maxSamplingRate = maxSampling1G
		scp.build2407Gui()
	case "2408B":
		slog.Warn("2408B not tested")
		scp.maxSamplingRate = maxSampling1G
	// case "2205MSO":
	// case "2205AMSO":
	// case "2206BMSO":
	// case "2207BMSO":
	// case "2208BMSO":
	default:
		err = fmt.Errorf("getInfo: unknown variant info %s cannot set maximum sample rate",
			scp.psControl.Info)
		return
	}
	return
}

func (scp *ScpDesc) setRangeMargin() {
	left, _, right, _ := scp.boundString("W-500.0")
	scp.rangeMargin = right - left
}

func (scp *ScpDesc) Menu(con *genericps.Connection, cfg *settings.PsSettings, fileName string) (err error) {
	scp.SettingFileName = fileName
	slog.Debug("menu", "cfg", *cfg)
	scp.triggerSettingMsg.Done = make(chan struct{})
	scp.psControl = control.NewControl(con)
	scp.Settings = cfg
	if scp.Settings.StreamEnabled != nil {
		scp.psControl.StreamEnabled.Store(*scp.Settings.StreamEnabled)
	}

	GlobalScreenScale = scp.getScreenScale()

	scp.Window = scp.App.NewWindow("")
	scp.theme = Theme(scp.Settings.Theme)
	fyne.CurrentApp().Settings().SetTheme(scp.theme)
	scp.ftScopeFullScreen = scp.newScopeScreen(image.Point{1024, 768})
	scp.dftScopeFullScreen = scp.newScopeScreen(image.Point{1024, 768})
	scp.fvScopeFullScreen = scp.newScopeScreen(image.Point{1024, 768})
	scp.ffScopeFullScreen = scp.newScopeScreen(image.Point{1024, 768})
	scp.setRangeMargin()
	scp.ftRaster = scp.newScreenRaster(scp.ftRasterGenerator, scp.Window, false, false, false)
	scp.dftRaster = scp.newScreenRaster(scp.dftRasterGenerator, scp.Window, true, false, false)
	scp.fvRaster = scp.newScreenRaster(scp.fvRasterGenerator, scp.Window, false, true, false)
	scp.ffRaster = scp.newScreenRaster(scp.ffRasterGenerator, scp.Window, false, false, true)
	scp.fvViewer = newFvViewer(scp.fvScopeFullScreen, image.Rect(0, 0, 1024, 768), scp)
	scp.addFvDrawer(scp.fvViewer)
	addToTest(scp.ftRaster, ftRasterId)
	addToTest(scp.dftRaster, dftRasterId)
	addToTest(scp.fvRaster, fvRasterId)
	addToTest(scp.ffRaster, ffRasterId)
	scp.themeChanged = createFlag()
	scp.repartition = createFlag()
	scp.ffBufferDone = make(chan struct{}, 1)

	err = scp.SetVariant()
	if err != nil {
		slog.Error("", "Menu GetInfo err=", err)
		return
	}
	scp.Window.SetTitle(scp.psControl.Info)

	sw, sh := scp.getScreenDimensions()
	winW := float32(scp.Settings.Window.Width)
	winH := float32(scp.Settings.Window.Height)
	if winW > sw {
		winW = sw
	}
	if winH > sh {
		winH = sh
	}
	scp.Window.Resize(fyne.NewSize(winW, winH))

	if scp.Settings.Window.Fullscreen {
		scp.Window.SetFullScreen(true)
	}
	scp.Window.Show()
	return
}
func (scp *ScpDesc) updateAcquisitionParameters() {
	if scp.psControl == nil {
		return
	}

	effectiveFunction := scp.controlTab.SelectedIndex()
	if scp.controlTab.Selected() == scp.genTab || scp.controlTab.Selected() == scp.filterTab || scp.controlTab.Selected() == scp.extgenTab {
		effectiveFunction = scp.Settings.Window.LastDispFunction
	}

	switch effectiveFunction {
	case dftTabIndex:
		// DFT mode: use sample rate and bins
		rate, _ := strconv.ParseFloat(scp.Settings.Dft.SampleRate, 64)
		unitMul := 1.0
		switch scp.Settings.Dft.SampleRateUnit {
		case selectscroll.UnitGSps:
			unitMul = 1e9
		case selectscroll.UnitMSps:
			unitMul = 1e6
		case selectscroll.UnitKSps:
			unitMul = 1e3
		case selectscroll.UnitSps:
			unitMul = 1.0
		}
		fs := rate * unitMul
		if fs <= 0 {
			fs = 1e6
		}

		// For DFT, we want at least 2 * Bins samples to avoid heavy zero padding
		samples := float64(scp.Settings.Dft.Bins * 2)
		scp.maxScreenTime = samples / fs
		scp.psControl.SetMaxScreenTimeCh <- scp.maxScreenTime
		scp.psControl.SetScopeScreenWidth(samples)
	case ffTabIndex:
		// f(f) Bode mode: automatically adapt the capture window so there are
		// ~20 full periods of the current signal in the buffer.  This gives the
		// single-bin DFT plenty of cycles for good amplitude/phase SNR while
		// keeping the sampling rate high enough to resolve the waveform.
		//
		// Target: maxScreenTime = 20 / freq
		// Fallback when no signal measured yet: use MinFreq (lowest expected).
		// Use the app-controlled target frequency for acquisition window sizing.
		// This ensures the capture window matches the frequency the sweep is
		// currently targeting, preventing the feedback loop that caused the
		// old sweep to stall at higher frequencies.
		refFreq := scp.currentFfFreq
		if refFreq <= 0 {
			refFreq = scp.measuredFfFreq
		}
		if refFreq <= 0 {
			refFreq = scp.Settings.Ff.MinFreq
		}
		if refFreq <= 0 {
			refFreq = 10.0 // absolute fallback
		}
		const targetCycles = 20.0
		targetScreenTime := targetCycles / refFreq

		// Only update the acquisition window if it changed significantly (> 5%).
		// Because frequency measurements have small amounts of noise, updating
		// on every buffer would cause endless restarts of the scope.
		needsUpdate := false
		if scp.maxScreenTime == 0 {
			needsUpdate = true
		} else {
			diffRatio := math.Abs(scp.maxScreenTime-targetScreenTime) / scp.maxScreenTime
			if diffRatio > 0.05 {
				needsUpdate = true
			}
		}

		if needsUpdate {
			scp.maxScreenTime = targetScreenTime
			scp.psControl.SetMaxScreenTimeCh <- scp.maxScreenTime
		}

		// Use the f(f) raster pixel width if available, else fall back to f(t) width.
		if scp.ffScopeSignalScreen != nil {
			scp.psControl.SetScopeScreenWidth(float64(scp.ffScopeSignalScreen.Bounds().Dx()))
		} else if scp.ftScopeSignalScreen != nil {
			scp.psControl.SetScopeScreenWidth(float64(scp.ftScopeSignalScreen.Bounds().Dx()))
		}
	default:
		// Time domain mode: use time/div
		scp.maxScreenTime = float64(scp.timeDiv) * math.Pow(10, float64(scp.timeUnit)) * 10 // 10 divs
		scp.psControl.SetMaxScreenTimeCh <- scp.maxScreenTime
		if scp.ftScopeSignalScreen != nil {
			scp.psControl.SetScopeScreenWidth(float64(scp.ftScopeSignalScreen.Bounds().Dx()))
		} else {
			// Estimate the expected F(t) signal screen width if it hasn't been drawn yet
			w := float32(scp.Settings.Window.Width)
			h := float32(scp.Settings.Window.Height)
			if w == 0 {
				w = 1024
			}
			if h == 0 {
				h = 768
			}
			leftMargin, rightMargin := scp.clipFtChRangeScrs(w, h)
			expectedDx := int(math.Round(float64(w-rightMargin))) - int(math.Round(float64(leftMargin)))
			scp.psControl.SetScopeScreenWidth(float64(expectedDx))
		}
	}
	scp.updateDftDataCollectionTime()
	scp.updateStreamButtonVisibility()
}

func (scp *ScpDesc) newFvPanel(panel *fyne.Container) {
	vbox := container.New(layout.NewVBoxLayout())
	var xChecks []*widget.Check

	for i := 0; i < int(scp.channelCount); i++ {
		chIndex := genericps.ChannelId(i)
		chName := channelNames[i]

		// Channel Label
		text := "Ch " + chName + ":"
		if scp.isDigitalFilterEnabled(chIndex) {
			text += " ⚠️"
		}
		label := canvas.NewText(text, scp.Settings.Channels[chIndex].Col[scp.Settings.ChannelColorIndex])
		label.TextStyle = fyne.TextStyle{Bold: true}
		label.TextSize = theme.TextSize()
		scp.channelViewers[chIndex].fvNameLabel = label

		// Enabled Checkbox
		enabledCheck := widget.NewCheck("Enabled", func(b bool) {
			scp.EnableChannel(chIndex, b)
		})
		enabledCheck.SetChecked(scp.Settings.Channels[chIndex].Enabled)
		scp.channelViewers[chIndex].enableChecks = append(scp.channelViewers[chIndex].enableChecks, enabledCheck)
		addToTest(enabledCheck, fvEnableId+chName)

		// X-Axis Check
		xCheck := widget.NewCheck("X-Axis", nil)
		if scp.Settings.Channels[chIndex].FvMode == settings.FvArgument {
			xCheck.SetChecked(true)
		}
		xChecks = append(xChecks, xCheck)
		addToTest(xCheck, fvXCheckId+chName)

		// Range Selector
		rangesEnum, _ := scp.psControl.ChannelRanges(chIndex)
		var ranges []string
		for _, r := range rangesEnum {
			ranges = append(ranges, inputRanges[r])
		}
		vRange := selectscroll.NewSelectScroll(ranges, func(option string, e selectscroll.Exception) {
			scp.changeChannelRange(chIndex, option)
		}, "+500m")
		scp.channelViewers[chIndex].vRangeSelects = append(scp.channelViewers[chIndex].vRangeSelects, vRange)
		addToTest(vRange, fvVRangeId+chName)

		vr := scp.Settings.Channels[chIndex].VRange
		if s, ok := rangeEnumToString[vr]; ok {
			vRange.SetSelected(s)
		}

		// X10 Checkbox
		x10Check := widget.NewCheck("X10", func(c bool) {
			scp.changeChannelX10(chIndex, c)
		})
		x10Check.SetChecked(scp.Settings.Channels[chIndex].X10)
		scp.channelViewers[chIndex].x10Checkboxes = append(scp.channelViewers[chIndex].x10Checkboxes, x10Check)
		addToTest(x10Check, fvX10Id+chName)

		// Arrange settings to minimize width (f(t) style)
		row1 := container.New(layout.NewHBoxLayout(), label, enabledCheck, xCheck)
		row2 := container.New(layout.NewHBoxLayout(), widget.NewLabel("Range:"), vRange, x10Check)

		chBox := container.New(layout.NewVBoxLayout(), row1, row2)
		if i > 0 {
			vbox.Add(layout.NewSpacer())
		}
		vbox.Add(chBox)
	}

	// Set up radio behavior for X-Axis checks
	for i := range xChecks {
		idx := i
		xChecks[idx].OnChanged = func(b bool) {
			if b {
				// Uncheck others
				for j, c := range xChecks {
					if idx != j {
						c.SetChecked(false)
					}
				}
				// Update settings
				for j := 0; j < int(scp.channelCount); j++ {
					if idx == j {
						scp.Settings.Channels[j].FvMode = settings.FvArgument
					} else {
						scp.Settings.Channels[j].FvMode = settings.FvValue
					}
				}
			} else {
				// If unchecked, set this one to FvValue as well
				scp.Settings.Channels[idx].FvMode = settings.FvValue
			}
			scp.refreshRasters()
			scp.SaveSettings()
		}
	}

	panel.Add(vbox)
}
func (scp *ScpDesc) setGeneratorFreq(f float64) {
	if scp.psControl == nil {
		return
	}

	if scp.ExtGenEnabled && scp.Settings.Ff.UseExternalGen && scp.extGen.Connected() {
		scp.setExtGenFrequency(f)
		return
	}

	if scp.controlTab.SelectedIndex() == ffTabIndex {
		if scp.psControl.Con != nil && scp.psControl.Con.ID == genericps.SimId {
			// Simulator mode: only sinus wave, for generators mapped via RLC
			activeGens := make([]bool, scp.channelCount)
			var missingGenChannels []string

			if scp.Settings.FfGen.On {
				for i := 0; i < int(scp.channelCount); i++ {
					if scp.Settings.Channels[i].Enabled {
						genSrc := int(scp.Settings.Channels[i].RlcFilter.GeneratorSource)
						if genSrc >= 0 && genSrc < int(scp.channelCount) {
							if scp.Settings.SimGenPanel[genSrc].On {
								activeGens[genSrc] = true
							} else {
								missingGenChannels = append(missingGenChannels, channelNames[i])
							}
						}
					}
				}
			}

			if len(missingGenChannels) > 0 {
				scp.psControl.DisplayStatus("Error: Channel "+strings.Join(missingGenChannels, ", ")+" has no active generator input", control.Warning)
			} else if strings.HasPrefix(scp.status.Text, "Error: Channel ") && strings.HasSuffix(scp.status.Text, " has no active generator input") {
				scp.psControl.DisplayStatus("", control.Info)
			}

			for i := 0; i < int(scp.channelCount); i++ {
				if activeGens[i] {
					msg := &control.GeneratorDescMsg{
						GeneratorDesc: control.GeneratorDesc{
							StartFrequency: f,
							StopFrequency:  f,
							Increment:      0,
							DwellTime:      1,
							SweepType:      genericps.SweepDown,
							WaveType:       genericps.Sine,
							OffsetVoltage:  0,
							PkToPK:         scp.Settings.Ff.Amplitude * 2,
							Channel:        genericps.ChannelId(i),
							On:             true,
							Phase:          0,
						},
					}
					scp.psControl.SetSimGenCh <- msg
				} else {
					offMsg := &control.GeneratorDescMsg{
						GeneratorDesc: control.GeneratorDesc{
							Channel: genericps.ChannelId(i),
							On:      false,
						},
					}
					scp.psControl.SetSimGenCh <- offMsg
				}
			}
			return
		}

		msg := &control.GeneratorDescMsg{
			GeneratorDesc: control.GeneratorDesc{
				StartFrequency: f,
				StopFrequency:  f,
				Increment:      0,
				DwellTime:      1,
				SweepType:      genericps.SweepDown, // No sweep
				WaveType:       genericps.Sine,      // Force Sine wave for Bode plots
				OffsetVoltage:  scp.Settings.FfGen.OffsetVoltage,
				PkToPK:         scp.Settings.FfGen.Amplitude * 2,
				On:             scp.Settings.FfGen.On,
			},
		}
		scp.psControl.SetGeneratorCh <- msg
		return
	}

	if scp.psControl.Con != nil && scp.psControl.Con.ID == genericps.SimId {
		// Simulator mode: update all enabled simulator channels
		for i := 0; i < int(scp.channelCount); i++ {
			msg := &control.GeneratorDescMsg{
				GeneratorDesc: control.GeneratorDesc{
					StartFrequency: f,
					StopFrequency:  f,
					Increment:      0,
					DwellTime:      1,
					SweepType:      genericps.SweepDown,
					WaveType:       scp.Settings.SimGenPanel[i].WaveType,
					OffsetVoltage:  scp.Settings.SimGenPanel[i].OffsetVoltage,
					PkToPK:         scp.Settings.SimGenPanel[i].Amplitude * 2,
					Channel:        genericps.ChannelId(i),
					On:             true,
				},
			}
			scp.psControl.SetSimGenCh <- msg
		}
		return
	}

	msg := &control.GeneratorDescMsg{
		GeneratorDesc: control.GeneratorDesc{
			StartFrequency: f,
			StopFrequency:  f,
			Increment:      0,
			DwellTime:      1,
			SweepType:      genericps.SweepDown, // No sweep
			WaveType:       scp.Settings.GenPanel.WaveType,
			OffsetVoltage:  scp.Settings.GenPanel.OffsetVoltage,
			PkToPK:         scp.Settings.GenPanel.Amplitude * 2,
		},
	}
	scp.psControl.SetGeneratorCh <- msg
}

func (scp *ScpDesc) inStreamMode() bool {
	if scp.psControl == nil {
		return false
	}
	return scp.maxScreenTime >= control.StreamThreshold && scp.psControl.StreamEnabled.Load()
}

func (scp *ScpDesc) updateStreamButtonVisibility() {
	if scp.streamEnableButton == nil {
		return
	}
	if scp.maxScreenTime >= control.StreamThreshold {
		scp.streamEnableButton.Show()
	} else {
		scp.streamEnableButton.Hide()
	}
	if scp.triggerDisplays != nil {
		if scp.inStreamMode() {
			scp.triggerDisplays.Hide()
		} else {
			if scp.triggerSource != dontCare {
				scp.triggerDisplays.Show()
			} else {
				scp.triggerDisplays.Hide()
			}
		}
	}
	if scp.toolbar != nil {
		scp.toolbar.Refresh()
	}
}

func (scp *ScpDesc) updateStreamButtonState() {
	if scp.streamEnableButton == nil || scp.psControl == nil {
		return
	}
	if scp.psControl.StreamEnabled.Load() {
		scp.streamEnableButton.SetText(streamEnabledLabel)
	} else {
		scp.streamEnableButton.SetText(streamDisabledLabel)
	}
}

func (scp *ScpDesc) applyFfGenSettings(on bool) {
	if scp.ExtGenEnabled && scp.Settings.Ff.UseExternalGen && scp.extGen.Connected() {
		if on {
			scp.extGen.SetAmplitude(scpi.Ch1, float64(scp.Settings.FfGen.Amplitude)/1000000.0)
			scp.extGen.SetOffset(scpi.Ch1, float64(scp.Settings.FfGen.OffsetVoltage)/1000000.0)
			scp.extGen.SetWaveform(scpi.Ch1, "SINusoid")
			scp.extGen.SetOutput(scpi.Ch1, true)
		} else {
			scp.extGen.SetOutput(scpi.Ch1, false)
		}

		// Ensure internal generator is turned off
		msg := &control.GeneratorDescMsg{}
		msg.Operation = genericps.EsOff
		if scp.psControl != nil && scp.psControl.SetGeneratorCh != nil {
			scp.psControl.SetGeneratorCh <- msg
		}
		return
	}

	msg := &control.GeneratorDescMsg{}
	if on {
		msg.StartFrequency = scp.Settings.Ff.MinFreq
		msg.StopFrequency = scp.Settings.Ff.MaxFreq
		msg.Increment = 0 // App controls frequency stepping; no hardware sweep
		msg.DwellTime = scp.Settings.Ff.DeltaT
		msg.SweepType = genericps.SweepUp
		msg.WaveType = genericps.Sine
		msg.OffsetVoltage = scp.Settings.FfGen.OffsetVoltage
		msg.PkToPK = scp.Settings.FfGen.Amplitude * 2
	} else {
		msg.DwellTime = 0
		msg.OffsetVoltage = 0
		msg.PkToPK = 0
		msg.WaveType = genericps.DcVoltage
		msg.StopFrequency = scp.Settings.FfGen.StopFrequency
		msg.SweepType = genericps.SweepUp
	}
	msg.Operation = genericps.EsOff
	msg.Shots = 0
	msg.Sweeps = 0
	msg.TriggerType = genericps.SigGenRising
	msg.TriggerSource = genericps.SigGenNone
	msg.ExtInThreshold = 0
	if scp.psControl != nil && scp.psControl.SetGeneratorCh != nil {
		scp.psControl.SetGeneratorCh <- msg
	}
}

func (scp *ScpDesc) applyFfSimGenSettings(on bool) {
	if scp.ExtGenEnabled && scp.Settings.Ff.UseExternalGen && scp.extGen.Connected() {
		if on {
			scp.extGen.SetAmplitude(scpi.Ch1, float64(scp.Settings.FfGen.Amplitude)/1000000.0)
			scp.extGen.SetOffset(scpi.Ch1, float64(scp.Settings.FfGen.OffsetVoltage)/1000000.0)
			scp.extGen.SetWaveform(scpi.Ch1, "SINusoid")
			scp.extGen.SetOutput(scpi.Ch1, true)
		} else {
			scp.extGen.SetOutput(scpi.Ch1, false)
		}

		// Ensure internal simulator generators are turned off
		for i := 0; i < int(scp.channelCount); i++ {
			msg := &control.GeneratorDescMsg{}
			msg.Channel = genericps.ChannelId(i)
			msg.Operation = genericps.EsOff
			if scp.psControl != nil && scp.psControl.SetSimGenCh != nil {
				scp.psControl.SetSimGenCh <- msg
			}
		}
		return
	}

	if scp.psControl != nil && scp.psControl.SetSimGenCh != nil {
		activeGens := make([]bool, scp.channelCount)
		var missingGenChannels []string

		if on {
			for i := 0; i < int(scp.channelCount); i++ {
				if scp.Settings.Channels[i].Enabled {
					genSrc := int(scp.Settings.Channels[i].RlcFilter.GeneratorSource)
					if genSrc >= 0 && genSrc < int(scp.channelCount) {
						if scp.Settings.SimGenPanel[genSrc].On {
							activeGens[genSrc] = true
						} else {
							missingGenChannels = append(missingGenChannels, channelNames[i])
						}
					}
				}
			}
		}

		if len(missingGenChannels) > 0 {
			if scp.status != nil {
				scp.psControl.DisplayStatus("Error: Channel "+strings.Join(missingGenChannels, ", ")+" has no active generator input", control.Warning)
			}
		} else if scp.status != nil && strings.HasPrefix(scp.status.Text, "Error: Channel ") && strings.HasSuffix(scp.status.Text, " has no active generator input") {
			scp.psControl.DisplayStatus("", control.Info)
		}

		for i := 0; i < int(scp.channelCount); i++ {
			msg := &control.GeneratorDescMsg{}
			msg.Channel = genericps.ChannelId(i)
			if activeGens[i] {
				msg.On = true
				msg.StartFrequency = scp.Settings.Ff.MinFreq
				msg.StopFrequency = scp.Settings.Ff.MaxFreq
				msg.Increment = 0 // App controls frequency stepping; no hardware sweep
				msg.DwellTime = scp.Settings.Ff.DeltaT
				msg.SweepType = genericps.SweepUp
				msg.WaveType = genericps.Sine
				msg.OffsetVoltage = scp.Settings.FfGen.OffsetVoltage
				msg.PkToPK = scp.Settings.FfGen.Amplitude * 2
				msg.Phase = 0
			} else {
				msg.On = false
				msg.DwellTime = 0
				msg.OffsetVoltage = 0
				msg.PkToPK = 0
				msg.WaveType = genericps.DcVoltage
				msg.StopFrequency = scp.Settings.FfGen.StopFrequency
				msg.SweepType = genericps.SweepDown
			}
			msg.Operation = genericps.EsOff
			msg.Shots = 0
			msg.Sweeps = 0
			msg.TriggerType = genericps.SigGenRising
			msg.TriggerSource = genericps.SigGenNone
			msg.ExtInThreshold = 0
			scp.psControl.SetSimGenCh <- msg
		}
	}
}

func (scp *ScpDesc) handleTabTransition(prevTab, newTab int) {
	if prevTab == newTab {
		return
	}

	// Transitioning from non-f(f) to f(f)
	if prevTab != ffTabIndex && newTab == ffTabIndex {
		// Force Sine wave for Bode plots across all generator configurations
		scp.Settings.GenPanel.WaveType = genericps.Sine
		for i := 0; i < len(scp.Settings.ExtGen); i++ {
			scp.Settings.ExtGen[i].WaveType = genericps.Sine
		}
		for i := 0; i < len(scp.Settings.SimGenPanel); i++ {
			scp.Settings.SimGenPanel[i].WaveType = genericps.Sine
		}

		scp.SaveSettings()

		// Refresh generator panels to reflect Sine wave selection in UI
		if scp.genLayout != nil {
			scp.genLayout.RemoveAll()
			if scp.psControl != nil && scp.psControl.Con != nil && scp.psControl.Con.ID == genericps.SimId {
				scp.newSimGenPanel(scp.genLayout, true)
			} else {
				scp.newGenPanel(scp.genLayout)
			}
		}

		if scp.extgenLayout != nil && scp.ExtGenEnabled {
			scp.extgenLayout.RemoveAll()
			scp.extgenLayout.Add(scp.newExtGenTab(true))
		}

		if scp.ffAmpDisp != nil {
			scp.ffAmpDisp.SetValue(int(scp.Settings.FfGen.Amplitude))
			scp.ffAmpDisp.Refresh()
		}

		if scp.psControl != nil {
			if scp.Settings.Ff.UseExternalGen && scp.extGen.Connected() {
				scp.syncExtGenSettings()
			} else if scp.psControl.Con != nil && scp.psControl.Con.ID == genericps.SimId {
				scp.applyFfSimGenSettings(false)
				scp.applyFfSimGenSettings(scp.Settings.FfGen.On)
			} else {
				scp.applyFfGenSettings(false)
				scp.applyFfGenSettings(scp.Settings.FfGen.On)
			}
		}
		if scp.running {
			scp.ResetFfSweep()
			scp.startFfSweep()
		}
	}

	// Transitioning from f(f) to non-f(f)
	if prevTab == ffTabIndex && newTab != ffTabIndex {
		if scp.status.Text == ErrWrongFfTrigger {
			scp.psControl.DisplayStatus("", control.Info)
		}
		scp.stopFfSweep() // stop any running Bode sweep
		if scp.psControl != nil && scp.psControl.Con.ID == genericps.SimId {
			for i := 0; i < int(scp.channelCount); i++ {
				scp.applySimGenSettings(genericps.ChannelId(i), &scp.Settings.SimGenPanel[i])
			}
		} else {
			scp.applyInternalGenSettings(scp.Settings.GenPanel.On)
		}
	}
}

func (scp *ScpDesc) shouldDrawRaster(targetTabIndex int) bool {
	if scp.controlTab == nil {
		return false
	}
	selectedIndex := scp.controlTab.SelectedIndex()
	if selectedIndex == targetTabIndex {
		return true
	}
	if targetTabIndex == ftTabIndex && selectedIndex == rlcTabIndex {
		return true
	}
	if scp.Settings.Window.LastDispFunction == targetTabIndex {
		sel := scp.controlTab.Selected()
		if sel == scp.genTab || sel == scp.filterTab || sel == scp.extgenTab {
			return true
		}
	}
	return false
}

func (scp *ScpDesc) dockTab(tab *container.TabItem) {
	if tab == nil || scp.controlTab == nil {
		return
	}
	// ensure tab is not already in Items
	for _, t := range scp.controlTab.Items {
		if t == tab {
			scp.controlTab.Select(tab)
			return
		}
	}
	// Global ordered list of all possible tabs
	allTabs := []*container.TabItem{
		scp.ftTab, scp.fvTab, scp.dftTab, scp.ffTab, scp.rlcTab, scp.filterTab, scp.genTab, scp.extgenTab,
	}

	var newItems []*container.TabItem
	for _, t := range allTabs {
		if t == nil {
			continue
		}
		if t == tab {
			newItems = append(newItems, t)
			continue
		}
		for _, existing := range scp.controlTab.Items {
			if existing == t {
				newItems = append(newItems, t)
				break
			}
		}
	}

	scp.controlTab.Items = newItems
	scp.controlTab.Refresh()
	scp.controlTab.Select(tab)
}

func (scp *ScpDesc) timeZoomGenerator(wInt int, hInt int) image.Image {
	defer scp.screenLocker.Unlock()
	scp.screenLocker.Lock()

	w := float32(wInt)
	h := float32(hInt)

	if scp.timeZoomScopeFullScreen == nil || scp.timeZoomScopeFullScreen.Bounds().Dx() != wInt || scp.timeZoomScopeFullScreen.Bounds().Dy() != hInt {
		scp.timeZoomScopeFullScreen = scp.newScopeScreen(image.Point{wInt, hInt})
		scp.partitionTzScreen(w, h)
		draw.Draw(scp.timeZoomScopeFullScreen, scp.timeZoomScopeFullScreen.Bounds(), &image.Uniform{scp.theme.Color(ColorNameSignalBackground, 0)}, image.ZP, draw.Src)
		scp.setTzVDivsY()
		scp.setTzHDivsX()
	} else {
		draw.Draw(scp.timeZoomScopeFullScreen, scp.timeZoomScopeSignalScreen.Bounds(), &image.Uniform{scp.theme.Color(ColorNameSignalBackground, 0)}, image.ZP, draw.Src)
		scp.setTzVDivsY()
		scp.setTzHDivsX()
	}

	for i := range scp.timeZoomDrawers {
		scp.timeZoomDrawers[i].draw()
	}
	return scp.timeZoomScopeFullScreen
}

func (scp *ScpDesc) openTimeZoomWindow() {
	if scp.timeZoomWindow != nil {
		scp.timeZoomWindow.RequestFocus()
		return
	}
	scp.timeZoomWindow = scp.App.NewWindow("Time Zoom")
	scp.timeZoomWindow.SetOnClosed(func() {
		scp.timeZoomWindow = nil
		scp.timeZoomRaster = nil
		scp.timeZoomScopeFullScreen = nil
		scp.timeZoomScopeSignalScreen = nil
		scp.timeZoomDrawers = nil
	})

	scp.timeZoomMaxScreenTime = scp.maxScreenTime

	scp.timeZoomRaster = scp.newScreenRaster(scp.timeZoomGenerator, scp.timeZoomWindow, false, false, false)
	scp.timeZoomRaster.disableInput = true

	scp.timeZoomWindow.SetContent(scp.timeZoomRaster)
	scp.timeZoomWindow.Resize(fyne.NewSize(800, 600))
	scp.timeZoomWindow.Show()
}
