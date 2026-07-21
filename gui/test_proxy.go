////go:build test

package gui

import (
	"fmt"
	"fynescope/disp7"
	"fynescope/selectscroll"
	"fynescope/sliderscroll"
	"io"
	"log"
	"log/slog"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

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

type TestControl struct {
	Obj fyne.CanvasObject
	Tab int
}

var (
	controls       map[string]TestControl
	FuzzerCommitID string
)

func init() {
	controls = make(map[string]TestControl)
}

func addToTest(obj fyne.CanvasObject, name string, tabID int) {
	controls[name] = TestControl{Obj: obj, Tab: tabID}
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

func randKey(name string) bool {
	ctrl, ok := controls[name]
	c := ctrl.Obj
	if !ok || c == nil || !c.Visible() {
		return false
	}
	slog.Debug("randKey", "name", name)
	switch c := c.(type) {
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
		return false
	}
	return true
}
func randTap(name string) bool {
	ctrl, ok := controls[name]
	c := ctrl.Obj
	if !ok || c == nil || !c.Visible() {
		return false
	}
	slog.Debug("randTap", "name", name)
	switch c := c.(type) {
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
			if obj, ok := controls[name]; ok && !obj.Obj.Visible() {
				return
			}
			c.Tapped(&fyne.PointEvent{AbsolutePosition: fyne.Position{X: 0, Y: 0}, Position: fyne.Position{X: 0, Y: 0}})
		})
	default:
		return false
	}
	return true
}
func randScroll(name string, n int) bool {
	ctrl, ok := controls[name]
	c := ctrl.Obj
	if !ok || c == nil || !c.Visible() {
		return false
	}
	slog.Debug("randScroll", "name", name)
	delta := float32(n)
	if n < 0 {
		n = -n
	}
	switch c := c.(type) {
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
		return false
	}
	return true
}
func randDrag(name string, delta float32) bool {
	ctrl, ok := controls[name]
	c := ctrl.Obj
	if !ok || c == nil || !c.Visible() {
		return false
	}
	slog.Debug("randDrag", "name", name)
	switch c := c.(type) {
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
		return false
	}
	return true
}
func tap(name string) {
	switch c := controls[name].Obj.(type) {
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
	switch c := controls[name].Obj.(type) {
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
	switch c := controls[name].Obj.(type) {
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

func (scp *ScpDesc) Test() {
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
		// tap(themeChangeActionId)
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
	scroll(ffMinFreqId, 100)
	wait()

	for !scp.running {
		wait()
		tap(runblockButtonId) // Start the sweep
	}
	// Wait for a few frequency steps to complete
	for i := 0; i < 200 && len(scp.bodeBuffers[0]) == 0; i++ {
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

type errorCountingWriter struct {
	target      io.Writer
	count       *uint64
	errorsMutex sync.Mutex
	firstErrors []string
}

func (w *errorCountingWriter) Write(p []byte) (n int, err error) {
	str := string(p)
	s := strings.ToLower(str)
	if strings.Contains(s, "level=error") {
		atomic.AddUint64(w.count, 1)
		w.errorsMutex.Lock()
		if len(w.firstErrors) < 100 {
			w.firstErrors = append(w.firstErrors, str)
		}
		w.errorsMutex.Unlock()
	}
	return w.target.Write(p)
}

func (scp *ScpDesc) Random(duration time.Duration, programVersion string, buildDate string, webport string) {
	commitID := FuzzerCommitID
	if commitID == "" {
		commitID = os.Getenv("FUZZER_COMMIT_ID")
	}
	
	if commitID == "" {
		fmt.Println("Error: Fuzzer requires a specified commit id.")
		fmt.Println("Please run with the required command, for example:")
		fmt.Println("FUZZER_COMMIT_ID=$(git rev-parse HEAD) go test -tags=\"noscope,testsw\" -v -run Test0 -timeout 105m")
		os.Exit(1)
	}

	var errorCount uint64
	var eventCount uint64
	startTime := time.Now()
	completed := false

	origLogWriter := log.Writer()
	customWriter := &errorCountingWriter{target: origLogWriter, count: &errorCount}
	log.SetOutput(customWriter)

	// Also configure slog if it was using the default handler
	origSlogHandler := slog.Default().Handler()
	slog.SetDefault(slog.New(slog.NewTextHandler(customWriter, &slog.HandlerOptions{Level: slog.LevelDebug})))

	defer func() {
		log.SetOutput(origLogWriter)
		slog.SetDefault(slog.New(origSlogHandler))
	}()

	statusWin := scp.App.NewWindow("Fuzzer Status")
	uptimeLabel := widget.NewLabel("Uptime: 0s")
	remainingLabel := widget.NewLabel(fmt.Sprintf("Remaining: %v", duration))
	eventsLabel := widget.NewLabel("Events: 0")
	errorsLabel := widget.NewLabel("Errors: 0")

	statusWin.SetContent(container.NewVBox(
		uptimeLabel,
		remainingLabel,
		eventsLabel,
		errorsLabel,
	))
	statusWin.Show()

	defer func() {
		fileName := fmt.Sprintf("fuzzer_%s.log", startTime.Format("0601021504"))
		if !completed {
			fileName = fmt.Sprintf("fuzzer_interrupted_%s.log", startTime.Format("0601021504"))
		}
		f, err := os.Create(fileName)
		if err == nil {
			uptime := time.Since(startTime)
			
			logBuildDate := buildDate
			if logBuildDate == "" {
				logBuildDate = startTime.Format("2006-01-02")
			}
			
			fmt.Fprintf(f, "Commit ID: %s\n", commitID)
			fmt.Fprintf(f, "Version: %s\n", programVersion)
			fmt.Fprintf(f, "Build Date: %s\n", logBuildDate)
			if webport != "" {
				fmt.Fprintf(f, "Webport: %s\n", webport)
			}
			
			tags := "none"
			if info, ok := debug.ReadBuildInfo(); ok {
				for _, s := range info.Settings {
					if s.Key == "-tags" {
						tags = s.Value
						break
					}
				}
			}
			fmt.Fprintf(f, "Build Tags: %s\n", tags)
			
			remaining := duration - uptime
			if remaining < 0 {
				remaining = 0
			}
			fmt.Fprintf(f, "Uptime: %v\n", uptime.Round(time.Second))
			fmt.Fprintf(f, "Remaining: %v\n", remaining.Round(time.Second))
			fmt.Fprintf(f, "Events: %d\n", atomic.LoadUint64(&eventCount))
			fmt.Fprintf(f, "Errors: %d\n", atomic.LoadUint64(&errorCount))
			
			customWriter.errorsMutex.Lock()
			if len(customWriter.firstErrors) > 0 {
				fmt.Fprintf(f, "\n--- First %d Errors ---\n", len(customWriter.firstErrors))
				for _, errStr := range customWriter.firstErrors {
					fmt.Fprint(f, errStr)
					if !strings.HasSuffix(errStr, "\n") {
						fmt.Fprintln(f)
					}
				}
				totalErrors := atomic.LoadUint64(&errorCount)
				if totalErrors > 100 {
					fmt.Fprintf(f, "\n... and %d more uncollected errors.\n", totalErrors-100)
				}
			}
			customWriter.errorsMutex.Unlock()
			
			f.Close()
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	arrayLen := len(controls)
	a := make([]string, arrayLen)
	i := 0
	for k := range controls {
		a[i] = k
		i++
	}
	op := 0
	ready := make(chan bool)
	deadline := time.Now().Add(duration)
	for time.Now().Before(deadline) {
		wait()
		var currentTab int
		if tabs, ok := controls[ftFuncId].Obj.(*container.AppTabs); ok {
			currentTab = tabs.SelectedIndex()
		}

		validKeys := make([]string, 0, 32)
		for k, ctrl := range controls {
			if ctrl.Tab == -1 || ctrl.Tab == currentTab {
				validKeys = append(validKeys, k)
			}
		}
		if len(validKeys) == 0 {
			continue
		}
		selectedKey := validKeys[rand.Intn(len(validKeys))]

		op = rand.Intn(4)
		go func() {
			n := 0
			executed := false
			switch op {
			case 0:
				n := rand.Intn(32) - 16
				executed = randDrag(selectedKey, float32(n))
			case 1:
				n = rand.Intn(10) - 5
				executed = randScroll(selectedKey, n)
			case 2:
				executed = randTap(selectedKey)
			case 3:
				executed = randKey(selectedKey)
			default:
				panic(8)
			}
			ready <- executed
		}()
		select {
		case <-sigChan:
			log.Println("Interrupted by signal")
			return
		case executed := <-ready:
			if executed {
				atomic.AddUint64(&eventCount, 1)
			}
			evs := atomic.LoadUint64(&eventCount)
			if evs%10 == 0 {
				uptime := time.Since(startTime)
				remaining := deadline.Sub(time.Now())
				if remaining < 0 {
					remaining = 0
				}
				errs := atomic.LoadUint64(&errorCount)
				fyne.Do(func() {
					uptimeLabel.SetText(fmt.Sprintf("Uptime: %v", uptime.Round(time.Second)))
					remainingLabel.SetText(fmt.Sprintf("Remaining: %v", remaining.Round(time.Second)))
					eventsLabel.SetText(fmt.Sprintf("Events: %d", evs))
					errorsLabel.SetText(fmt.Sprintf("Errors: %d", errs))
				})
			}
		case <-time.After(timeout):
			log.Println("Timed out ", selectedKey, op)
			buf := make([]byte, 1<<20)
			n := runtime.Stack(buf, true)
			log.Println(string(buf[:n]))
			panic(7)
		}
	}
	completed = true
}
