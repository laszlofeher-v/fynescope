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
	"fynescope/disp7"
	"fynescope/genericps"
	"fynescope/selectscroll"
	"fynescope/settings"
	"image"
	"log/slog"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"gonum.org/v1/gonum/dsp/fourier"
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
		settingsLocker                      sync.Mutex
		ffSweepQuit                         chan struct{}
		ffSweepDataReady                    chan struct{}
		ffSweepAcquireTime                  time.Time
		ffBufferDone                        chan struct{}
		currentFfFreq                       float64
		measuredFfFreq                      float64
		// status field containing UI label and numeric code
		status *InitStatus
		// FFT caches for processFfData — reallocated only when sample count changes
		ffFftObj          *fourier.FFT
		ffFftBuf          []float64
		ffFftResult       []complex128
		ffFftSamples      int
		ffCurrentFreqDisp *disp7.DigitArray
		ffMinFreqDisp     *disp7.DigitArray
		ffMaxFreqDisp     *disp7.DigitArray
		ffStepFreqDisp    *disp7.DigitArray
		ffDeltaTDisp      *disp7.DigitArray
		ffAmpDisp         *disp7.DigitArray
		ffOffsetDisp      *disp7.DigitArray

		bodeBuffers                  [genericps.MaxChannel][]bodePoint
		maxSamplingRate              uint32
		segmentIndex                 uint32 // maxSamplingRate: sample/sec
		controlTab                   *container.AppTabs
		dftTab                       *container.TabItem
		fraTab                       *container.TabItem
		ftTab                        *container.TabItem
		fvTab                        *container.TabItem
		ffTab                        *container.TabItem
		genTab                       *container.TabItem
		rlcTab                       *container.TabItem
		filterTab                    *container.TabItem
		extgenTab                    *container.TabItem
		setTab                       *container.TabItem
		psControl                    *control.PscDesc
		triggerHysteresisDisp        *disp7.DigitArray
		triggerThresholdDisp         *disp7.DigitArray
		boxTriggerHysteresisDisp     *fyne.Container
		triggerLowerThresholdDisp    *disp7.DigitArray
		triggerLowerHysteresisDisp   *disp7.DigitArray
		intervalTypeSelect           *selectscroll.SelectScroll
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
		dftDivsX                   []float32
		dftDivsY                   [numberOfDivs + 1]float32
		ftDivsX                    [numberOfDivs + 1]float32
		ftDivsY                    [numberOfDivs + 1]float32
		fvDivsX                    [numberOfDivs + 1]float32
		fvDivsY                    [numberOfDivs + 1]float32
		binWidthLabel              *widget.Label
		dftDataCollectionTimeLabel *widget.Label
		dftSampleRateSelect        *selectscroll.SelectScroll
		dftSampleUnitSelect        *selectscroll.SelectScroll
		dftMinFreqDisp             *disp7.DigitArray
		dftMaxFreqDisp             *disp7.DigitArray
		SettingFileName            string
		extGen                     control.ExtGenDesc
		ExtGenEnabled              bool
		HighResUIEnabled           bool
		useExtGenCheck             *widget.Check
		hiResCheck                 *widget.Check
		complexTriggerCheck        *widget.Check
		timeZoomButton             *widget.Button
		timeZoomWindow             fyne.Window
		timeZoomRaster             *screenRaster
		timeZoomMaxScreenTime      float64
		timeZoomScopeFullScreen    rasterImage
		timeZoomScopeSignalScreen  rasterImage
		timeZoomDrawers            []drawer
		timeZoomBoxOffset          float64
		timeZoomBottomLabelViewer  drawer
		timeZoomDivsX              [numberOfDivs + 1]float32
		timeZoomDivsY              [numberOfDivs + 1]float32
		timeZoomTriggerPoint       drawer
		timeZoomTimeDiv            int
		timeZoomTimeUnit           int
		tzRepartition              chan struct{}
	}
)

func createFlag() (ch chan struct{}) {
	ch = make(chan struct{}, 1)
	return
}
func setFlag(flag chan struct{}) {
	if flag == nil {
		return
	}
	select {
	case flag <- struct{}{}:
	default:
	}
}

func getFlag(flag chan struct{}) bool {
	if flag == nil {
		return false
	}
	select {
	case <-flag:
		return true
	default:
		return false
	}
}

func (scp *ScpDesc) SaveSettings() {
	if scp.SettingFileName == "" {
		return
	}
	go func() {
		scp.settingsLocker.Lock()
		defer scp.settingsLocker.Unlock()
		if err := settings.Save(scp.SettingFileName, scp.Settings); err != nil {
			slog.Error("failed to save settings", "err", err)
		}
	}()
}

func (scp *ScpDesc) refreshRasters() {
	if scp.Settings == nil {
		return
	}
	fyne.Do(func() {
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
	})
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
					scp.psControl.DisplayStatus(ErrWrongFfTrigger, control.Warning)
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

	addToTest(scp.controlTab, ftFuncId, -1)
	addToTest(scp.controlTab, fvFuncId, -1)
	addToTest(scp.controlTab, dftFuncId, -1)
	addToTest(scp.controlTab, ffFuncId, -1)
	addToTest(scp.controlTab, rlcFuncId, -1)
	addToTest(scp.controlTab, genFuncId, -1)
	addToTest(scp.controlTab, filterFuncId, -1)
	addToTest(scp.controlTab, extgenFuncId, -1)
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
	addToTest(themeChangeAction, themeChangeActionId, -1)

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
			if scp.status != nil && scp.status.Code() == StatusFrequencyCannotBeDetected {
				scp.psControl.DisplayStatus("", control.Info)
			}
			if scp.controlTab.SelectedIndex() == ffTabIndex {
				if scp.Settings.Trigger.Type == settings.TriggerTypeInterval || scp.Settings.Trigger.Type == settings.TriggerTypePulseWidth {
					scp.psControl.DisplayStatus(ErrWrongFfTrigger, control.Warning)
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
	addToTest(scp.runblockButton, runblockButtonId, -1)
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
			scp.toolbar.Add(scp.status.label)
			content = container.NewBorder(scp.toolbar, nil, scp.controlTab, left, activeRasterContainer)
			changeSide.SetIcon(theme.NavigateNextIcon())
		} else {
			scp.Settings.Window.LeftControl = false
			scp.toolbar.RemoveAll()
			scp.toolbar.Add(scp.status.label)
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
	addToTest(changeSide, changeSideId, -1)
	logout = widget.NewButtonWithIcon("", theme.LogoutIcon(), func() {
		if scp.psControl != nil {
			scp.psControl.Shutdown()
		}
		if scp.status != nil && scp.status.statusQuit != nil {
			close(scp.status.statusQuit)
			scp.status.statusQuit = nil
		}
		scp.App.Quit()
	})
	if scp.Settings.Window.LeftControl {
		scp.toolbar = container.New(layout.NewHBoxLayout(), scp.runblockButton, scp.streamEnableButton, scp.timeZoomButton, fullScreen, restoreScreen, changeSide,
			themeChangeAction,
			logout,
			layout.NewSpacer(),
			scp.status.label)
		content = container.NewBorder(scp.toolbar, nil, scp.controlTab, left, activeRasterContainer)
	} else {
		scp.toolbar = container.New(layout.NewHBoxLayout(), scp.status.label, layout.NewSpacer(),
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
	scp.psControl.SetHiRes(scp.Settings.Time.HiRes)

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
	addToTest(scp.ftRaster, ftRasterId, ftTabIndex)
	addToTest(scp.dftRaster, dftRasterId, dftTabIndex)
	addToTest(scp.fvRaster, fvRasterId, fvTabIndex)
	addToTest(scp.ffRaster, ffRasterId, ffTabIndex)
	scp.themeChanged = createFlag()
	scp.repartition = createFlag()
	scp.tzRepartition = createFlag()
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
