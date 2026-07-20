////go:build test

package gui

import (
	"fynescope/disp7"
	"fynescope/selectscroll"
	"fynescope/sliderscroll"
	"log"
	"log/slog"
	"math/rand"
	"runtime"
	"time"

	"fyne.io/fyne/v2/container"

	"fyne.io/fyne/v2"
)

/*
save gui objects in control map[string]*object
controlSave(widget *Widget, name string)

in another file
go:build !test
dummy function

test refers :
test.Tap(control[enableCha])
*/
const (
	ftFuncId                       = "ftFunc"
	fvFuncId                       = "fvFunc"
	dftFuncId                      = "dftFunc"
	genFuncId                      = "genFunc"
	filterFuncId                   = "filterFunc"
	ffFuncId                       = "ffFunc"
	runblockButtonId               = "runblockButton"
	themeChangeActionId            = "themeChangeAction"
	genAmpdSetId                   = "genAmpSet"
	genFreqSetId                   = "genFreqSet"
	genFreqId                      = "genFreq"
	genShowId                      = "genShow"
	genCheckId                     = "genCheck"
	chEnableId                     = "chEnable"
	vRangeId                       = "vRange"
	invertId                       = "invert"
	x10Id                          = "x10"
	triggerCheckId                 = "triggerCheck"
	persId                         = "pers"
	timeId                         = "time"
	timeSelectId                   = "timeSelect"
	unitSelectId                   = "unitSelect"
	acdcId                         = "acdc"
	genMinFrqId                    = "genLowerLimit"
	genStepFreqId                  = "genStepFreq"
	genMaxFrqId                    = "genUpperLimit"
	genAmpId                       = "genAmp"
	genOffsetId                    = "genOffset"
	genSweepId                     = "genSweepId"
	ftRasterId                     = "ftRaster"
	dftRasterId                    = "dftRaster"
	fvRasterId                     = "fvRaster"
	fvEnableId                     = "fvEnable"
	fvXCheckId                     = "fvXCheck"
	fvVRangeId                     = "fvVRange"
	fvX10Id                        = "fvX10"
	ffRasterId                     = "ffRaster"
	changeSideId                   = "changeSide"
	triggerThresholdDispId         = "triggerThresholdDisp"
	triggerHysteresisDispId        = "triggerHysteresisDisp"
	triggerModeSelectId            = "triggerModeSelect"
	triggerTypeSelectId            = "triggerTypeSelect"
	triggerCalculationModeSelectId = "triggerCalculationModeSelect"
	intervalTypeSelectId           = "intervalTypeSelect"
	intervalTimeLowerDispId        = "intervalTimeLowerDisp"
	intervalTimeUpperDispId        = "intervalTimeUpperDisp"
	intervalTimeSingleDispId       = "intervalTimeSingleDisp"
	dftEnableId                    = "dftEnable"
	dftPersId                      = "dftPers"
	dftVRangeId                    = "dftVRange"
	dftX10Id                       = "dftX10"
	dftWindowId                    = "dftWindow"
	dftModeId                      = "dftMode"
	dftMaxFreqValId                = "dftMaxFreqVal"
	dftMaxFreqUnitId               = "dftMaxFreqUnit"
	dftBinId                       = "dftBin"
	dftSampleRateId                = "dftSampleRate"
	dftSampleUnitId                = "dftSampleUnit"
	extGenOnOffId                  = "extGenOnOff"
	extGenWaveTypeId               = "extGenWaveType"
	extGenFreqId                   = "extGenFreq"
	extGenAmpId                    = "extGenAmp"
	extGenOffsetId                 = "extGenOffset"
	extGenPhaseId                  = "extGenPhase"
	extGenImpOhmsId                = "extGenImpOhms"
	extGenImpModeId                = "extGenImpMode"
	ipmId                          = "ipm"
	triggerDirectionId             = "triggerDirection"
	chOffsetId                     = "chOffsetId"
	genRiseFallTimeId              = "genRiseFallTime"
	genNoiseAmpId                  = "genNoiseAmp"
	genPhaseNoiseId                = "genPhaseNoise"
	genPhaseId                     = "genPhase"
	genDwellTimeId                 = "genDwellTime"
	genWaveTypeId                  = "genWaveType"
	genOperationId                 = "genOperation"
	ffMinFreqId                    = "ffMinFreq"
	ffMaxFreqId                    = "ffMaxFreq"
	ffSweepButtonId                = "ffSweepButton"
	ffStopButtonId                 = "ffStopButton"
	ffExtGenSelectId               = "ffExtGenSelect"
	ffDispModeSelectId             = "ffDispModeSelect"

	ffPhaseCheckId = "ffPhaseCheck"
	ffRefCheckId   = "ffRefCheck"
	ffEnableId     = "ffEnable"
	ffX10Id        = "ffX10"
	ffVRangeId     = "ffVRange"
	rlcFuncId      = "rlcFunc"
	extgenFuncId   = "extgenFunc"
	rlcEnableId    = "rlcEnable"
	rlcTypeId      = "rlcType"
	rlcGenSourceId = "rlcGenSource"
	rlcRId         = "rlcR"
	rlcRUnitId     = "rlcRUnit"
	rlcLId         = "rlcL"
	rlcLUnitId     = "rlcLUnit"
	rlcCId         = "rlcC"
	rlcCUnitId     = "rlcCUnit"
	sleepTime      = 100 * time.Millisecond
	timeout        = time.Duration(30) * time.Second
)

type (
	testCase struct {
		cmd             string
		resultImageName string
	}
)

var (
	controls map[string]fyne.CanvasObject
)
var visibleCh chan bool

func init() {
	controls = make(map[string]fyne.CanvasObject)
	visibleCh = make(chan bool, 1)
}

func addToTest(obj fyne.CanvasObject, name string) {
	controls[name] = obj
}
func wait() {
	time.Sleep(sleepTime)
}

var keyNames = []fyne.KeyName{
	fyne.KeyUp, fyne.KeyDown, fyne.KeyLeft,
	fyne.KeyRight, fyne.KeyDelete, fyne.KeyBackspace,
	fyne.Key0, fyne.Key1, fyne.Key2,
	fyne.Key3, fyne.Key4, fyne.Key5,
	fyne.Key6, fyne.Key7, fyne.Key8, fyne.Key9}

func randKey(name string) {
	if c, ok := controls[name]; !ok || c == nil || !c.Visible() {
		return
	}
	slog.Debug("randKey", "name", name)
	switch c := controls[name].(type) {
	case *disp7.DigitArray:
		wait()
		key := keyNames[rand.Intn(len(keyNames))]
		fyne.Do(func() {
			if !c.Visible() {
				return
			}
			c.Window.Canvas().Focus(c)
			c.KeyDown(&fyne.KeyEvent{Name: key})
		})
		wait()
		fyne.Do(func() {
			if !c.Visible() {
				return
			}
			c.KeyUp(&fyne.KeyEvent{Name: key})
			c.Window.Canvas().Unfocus()
		})
	case *digitEntry:
		wait()
		fyne.Do(func() {
			if !c.Visible() {
				return
			}
			if rand.Float32() < 0.2 && len(c.Text) > 0 {
				c.TypedKey(&fyne.KeyEvent{Name: fyne.KeyBackspace})
			} else {
				runes := []rune("0123456789abcdefABCDEF")
				r := runes[rand.Intn(len(runes))]
				c.TypedRune(r)
			}
		})
		wait()
	default:
	}
}
func randTap(name string) {
	if c, ok := controls[name]; !ok || c == nil || !c.Visible() {
		return
	}
	slog.Debug("randTap", "name", name)
	switch c := controls[name].(type) {
	case *container.AppTabs:
		fyne.Do(func() {
			var targetText string
			switch name {
			case ftFuncId:
				targetText = "f(t)"
			case fvFuncId:
				targetText = "f(v)"
			case dftFuncId:
				targetText = "FFT"
			case ffFuncId:
				targetText = "f(f)"
			case rlcFuncId:
				targetText = "RLC"
			case filterFuncId:
				targetText = "filter"
			case genFuncId:
				targetText = "gen"
			case extgenFuncId:
				targetText = "extgen"
			}
			if targetText != "" {
				for idx, item := range c.Items {
					if item.Text == targetText {
						c.SelectIndex(idx)
						break
					}
				}
			}
		})
	case *selectscroll.SelectScroll:
		n := rand.Intn(len(c.Options))
		wait()
		fyne.Do(func() {
			if !c.Visible() {
				return
			}
			c.SetSelectedIndex(n)
		})
	case fyne.Tappable:
		wait()
		fyne.Do(func() {
			if obj, ok := controls[name]; ok && !obj.Visible() {
				return
			}
			c.Tapped(&fyne.PointEvent{AbsolutePosition: fyne.Position{X: 0, Y: 0}, Position: fyne.Position{X: 0, Y: 0}})
		})
	default:
	}
}
func randScroll(name string, n int) {
	if c, ok := controls[name]; !ok || c == nil || !c.Visible() {
		return
	}
	slog.Debug("randScroll", "name", name)
	delta := float32(n)
	if n < 0 {
		n = -n
	}
	switch c := controls[name].(type) {
	case *screenRaster:
		if n > 2 {
			n = 2
		} // Limit iterations to avoid watchdog timeouts
		for ; n > 0; n-- {
			wait()
			fyne.Do(func() {
				if !c.Visible() || int(c.Size().Width) <= 0 || int(c.Size().Height) <= 0 {
					return
				}
				ap := c.Position() // The absolute position of the event
				x := rand.Intn(int(c.Size().Width))
				y := rand.Intn(int(c.Size().Height))
				p := fyne.Position{X: float32(x), Y: float32(y)} // The relative position of the event
				e := &fyne.ScrollEvent{}
				e.Scrolled.DY = delta
				e.Scrolled.DY = delta
				e.AbsolutePosition = ap
				e.Position = p
				c.Scrolled(e)
			})
		}
	case *sliderscroll.SliderScroll:
		wait()
		e := &fyne.ScrollEvent{Scrolled: fyne.Delta{DX: delta, DY: delta}}
		fyne.Do(func() {
			if !c.Visible() {
				return
			}
			c.Scrolled(e)
		})
	case *selectscroll.SelectScroll:
		if n > 2 {
			n = 2
		} // Limit iterations to avoid timeouts on heavy OnChanged callbacks
		for ; n > 0; n-- {
			wait()
			e := &fyne.ScrollEvent{Scrolled: fyne.Delta{DX: delta, DY: delta}}
			fyne.Do(func() {
				if !c.Visible() {
					return
				}
				c.Scrolled(e)
			})
		}
	case *disp7.DigitArray:
		if n > 2 {
			n = 2
		} // Limit iterations to avoid timeouts
		for ; n > 0; n-- {
			wait()
			fyne.Do(func() {
				if !c.Visible() || int(c.Size().Width) <= 0 || int(c.Size().Height) <= 0 {
					return
				}
				ap := c.Position() // The absolute position of the event
				digit := rand.Intn(int(c.Size().Width))
				p := fyne.Position{X: float32(digit), Y: 1} // The relative position of the event
				e := &fyne.ScrollEvent{}
				e.Scrolled.DY = delta
				e.Scrolled.DY = delta
				e.AbsolutePosition = ap
				e.Position = p
				c.Scrolled(e)
			})
		}
	default:
	}
}
func randDrag(name string, delta float32) {
	if c, ok := controls[name]; !ok || c == nil || !c.Visible() {
		return
	}
	slog.Debug("randDrag", "name", name)
	switch c := controls[name].(type) {
	case *screenRaster:
		wait()
		fyne.Do(func() {
			if !c.Visible() || int(c.Size().Width) <= 0 || int(c.Size().Height) <= 0 {
				return
			}
			ap := c.Position() // The absolute position of the event
			x := rand.Intn(int(c.Size().Width))
			y := rand.Intn(int(c.Size().Height))
			p := fyne.Position{X: float32(x), Y: float32(y)} // The relative position of the event
			e := fyne.PointEvent{}
			e.AbsolutePosition = ap
			e.Position = p
			c.Dragged(&fyne.DragEvent{PointEvent: e, Dragged: fyne.NewDelta(delta, delta)})
		})
	case *sliderscroll.SliderScroll:
		wait()
		fyne.Do(func() {
			if !c.Visible() {
				return
			}
			c.Dragged(&fyne.DragEvent{Dragged: fyne.NewDelta(delta, delta)})
		})
	case *disp7.DigitArray:
		wait()
		fyne.Do(func() {
			if !c.Visible() {
				return
			}
			c.Dragged(&fyne.DragEvent{Dragged: fyne.NewDelta(delta, delta)})
		})
	default:
	}
}
func tap(name string) {
	switch c := controls[name].(type) {
	case *container.AppTabs:
		fyne.Do(func() {
			var targetText string
			switch name {
			case ftFuncId:
				targetText = "f(t)"
			case fvFuncId:
				targetText = "f(v)"
			case dftFuncId:
				targetText = "FFT"
			case ffFuncId:
				targetText = "f(f)"
			case rlcFuncId:
				targetText = "RLC"
			case filterFuncId:
				targetText = "filter"
			case genFuncId:
				targetText = "gen"
			case extgenFuncId:
				targetText = "extgen"
			}
			if targetText != "" {
				for idx, item := range c.Items {
					if item.Text == targetText {
						c.SelectIndex(idx)
						break
					}
				}
			}
		})
	case fyne.Tappable:
		wait()
		fyne.Do(func() {
			c.Tapped(&fyne.PointEvent{AbsolutePosition: fyne.Position{X: 0, Y: 0}, Position: fyne.Position{X: 0, Y: 0}})
		})
	default:
		log.Printf("%s cannot use type %T\n", name, c)
	}
}
func scroll(name string, n int) {
	delta := float32(n)
	if n < 0 {
		n = -n
	}
	switch c := controls[name].(type) {
	case *screenRaster:
		for ; n > 0; n-- {
			wait()
			ap := c.Position() // The absolute position of the event
			x := 0
			y := 0
			p := fyne.Position{X: float32(x), Y: float32(y)} // The relative position of the event
			e := &fyne.ScrollEvent{}
			e.Scrolled.DY = delta
			e.Scrolled.DY = delta
			e.AbsolutePosition = ap
			e.Position = p
			fyne.Do(func() {
				c.Scrolled(e)
			})
		}
	case *sliderscroll.SliderScroll:
		wait()
		e := &fyne.ScrollEvent{Scrolled: fyne.Delta{DX: delta, DY: delta}}
		fyne.Do(func() {
			c.Scrolled(e)
		})
	case *selectscroll.SelectScroll:
		for ; n > 0; n-- {
			wait()
			e := &fyne.ScrollEvent{Scrolled: fyne.Delta{DX: delta, DY: delta}}
			fyne.Do(func() {
				c.Scrolled(e)
			})
		}
	case *disp7.DigitArray:
		for ; n > 0; n-- {
			wait()
			ap := c.Position() // The absolute position of the event
			if int(c.Size().Width) <= 0 {
				return
			}
			digit := rand.Intn(int(c.Size().Width))
			p := fyne.Position{X: float32(digit), Y: 1} // The relative position of the event
			e := &fyne.ScrollEvent{}
			e.Scrolled.DY = delta
			e.Scrolled.DY = delta
			e.AbsolutePosition = ap
			e.Position = p
			fyne.Do(func() {
				c.Scrolled(e)
			})
		}
	default:
	}
}
func drag(name string, delta float32) {
	switch c := controls[name].(type) {
	case *screenRaster:
		wait()
		ap := c.Position() // The absolute position of the event
		x := 0
		y := 0
		p := fyne.Position{X: float32(x), Y: float32(y)} // The relative position of the event
		e := fyne.PointEvent{}
		e.AbsolutePosition = ap
		e.Position = p
		fyne.Do(func() {
			c.Dragged(&fyne.DragEvent{PointEvent: e, Dragged: fyne.NewDelta(delta, delta)})
		})
	case *sliderscroll.SliderScroll:
		wait()
		fyne.Do(func() {
			c.Dragged(&fyne.DragEvent{Dragged: fyne.NewDelta(delta, delta)})
		})
	case *disp7.DigitArray:
		wait()
		fyne.Do(func() {
			c.Dragged(&fyne.DragEvent{Dragged: fyne.NewDelta(delta, delta)})
		})
	default:
	}
}

func (scp *ScpDesc) Test( /*w *sync.WaitGroup*/ ) {
	log.Println("Test started")
	tap(ftFuncId)
	tap(genFuncId)
	scroll(genMinFrqId, 5)
	scroll(genMaxFrqId, 5)
	scroll(genStepFreqId, 5)
	for i := 0; i < 1; i++ {
		randTap(runblockButtonId)
		tap(themeChangeActionId)
		drag(genFreqSetId, 100000)
		tap(genCheckId)
		scroll(acdcId+"A", -1)
		tap(themeChangeActionId)
		scroll(acdcId+"B", 1)
		scroll(genFreqSetId, 1000)
		tap(chEnableId + "A")
		scroll(genFreqSetId, 1000)
		tap(chEnableId + "B")
		tap(chEnableId + "C")
		tap(chEnableId + "D")
		scroll(vRangeId+"A", 1)
		scroll(vRangeId+"B", 2)
		scroll(vRangeId+"C", 3)
		wait()
		scroll(genFreqSetId, -1000)
		wait()
		tap(chEnableId + "C")
		tap(chEnableId + "B")
		scroll(genFreqSetId, -500)
		scroll(vRangeId+"A", -1)
		scroll(vRangeId+"B", -2)
		scroll(vRangeId+"C", -3)
		scroll(genFreqSetId, -500)
		tap(chEnableId + "A")
		tap(genCheckId)
		tap(runblockButtonId)
		tap(themeChangeActionId)
		wait()
		wait()
		wait()
		wait()
		tap(changeSideId)
		tap(changeSideId)
		tap(runblockButtonId)

		wait()
	}

	fyne.Do(func() {
		// Set time division to 10 ms/div so that low frequency signals (10Hz-1000Hz) have enough periods on screen to measure frequency
		scp.timeUnitSelect.SilentSetSelected("ms/div")
		scp.timeSelect.SilentSetSelected("10")
		scp.onTimeUnitChange("ms/div", selectscroll.None)
		scp.triggerModeSelect.SetSelected("Auto")
	})
	tap(ffFuncId)
	if scp.running {
		tap(runblockButtonId) // Stop before switching to ensure sweep starts cleanly
	}
	tap(runblockButtonId) // Start the sweep

	// Wait for a few frequency steps to complete
	for i := 0; i < 60; i++ {
		wait()
	}

	log.Printf("BODE BUFFER SIZE: %d", len(scp.bodeBuffers[0]))
	if len(scp.bodeBuffers[0]) == 0 {
		log.Fatalf("FAIL: expected bode buffers to have data before clear")
	}
	// Pause the run if it hasn't automatically finished
	fyne.Do(func() {
		if scp.running {
			scp.StopRunning()
		}
	})
	wait()
	fyne.Do(func() {
		scp.ResetFfSweep()
	})
	wait()
	scp.ffLocker.Lock()
	bufLen := len(scp.bodeBuffers[0])
	scp.ffLocker.Unlock()
	if bufLen != 0 {
		log.Fatalf("FAIL: expected bode buffers to be empty after clear, got: %d", bufLen)
	}
}

func (scp *ScpDesc) Random(duration time.Duration) {
	arrayLen := len(controls)
	a := make([]string, arrayLen)
	i := 0
	for k := range controls {
		a[i] = k
		i++
	}
	op := 0
	control := 0
	ready := make(chan struct{})
	deadline := time.Now().Add(duration)
	for time.Now().Before(deadline) {
		wait()
		control = rand.Intn(len(controls))
		var targetTab int
		if len(a[control]) >= 6 && (a[control][0:6] == "extGen" || a[control][0:6] == "extgen") {
			targetTab = 7
		} else if len(a[control]) >= 6 && a[control][0:6] == "filter" {
			targetTab = 5
		} else if len(a[control]) >= 3 && a[control][0:3] == "gen" {
			targetTab = 6
		} else if len(a[control]) >= 3 && a[control][0:3] == "rlc" {
			targetTab = 4
		} else if len(a[control]) >= 2 && a[control][0:2] == "ff" {
			targetTab = 3
		} else if len(a[control]) >= 3 && a[control][0:3] == "dft" {
			targetTab = 2
		} else if len(a[control]) >= 2 && a[control][0:2] == "fv" {
			targetTab = 1
		} else {
			targetTab = 0
		}
		var currentTab int
		if tabs, ok := controls[ftFuncId].(*container.AppTabs); ok {
			currentTab = tabs.SelectedIndex()
		}

		isAlwaysVisible := false
		switch a[control] {
		case ftFuncId, fvFuncId, dftFuncId, ffFuncId, rlcFuncId, genFuncId, filterFuncId, extgenFuncId,
			runblockButtonId, themeChangeActionId, changeSideId:
			isAlwaysVisible = true
		}

		if currentTab != targetTab && !isAlwaysVisible {
			continue
		}
		op = rand.Intn(4)
		go func() {
			n := 0
			switch op {
			case 0:
				n := rand.Intn(32) - 16
				randDrag(a[control], float32(n))
			case 1:
				n = rand.Intn(10) - 5
				randScroll(a[control], n)
			case 2:
				randTap(a[control])
			case 3:
				randKey(a[control])
			default:
				panic(8)
			}
			ready <- struct{}{}
		}()
		select {
		case <-ready:
		case <-time.After(timeout):
			log.Println("Timed out ", a[control], op)
			buf := make([]byte, 1<<20)
			n := runtime.Stack(buf, true)
			log.Println(string(buf[:n]))
			panic(7)
		}
	}
}
