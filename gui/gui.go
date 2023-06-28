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
	"image"
	"log"
	"math"
	"psgo/psi"

	//	"math"

	"fyne.io/fyne/v2/theme"

	"image/color"
	//	"image/draw"

	//	"math/rand"
	"os"
	"os/exec"
	"psgo/gui/inter"
	"psgo/ps2000a"

	// "psgo/ps2000asim"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	//	"fyne.io/fyne/v2/data/binding"

	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type (
	signalDesc struct {
		signal       [][]int32
		length       int32
		time         float64
		timeInterval float64
	}
	ChannelSetMsg struct {
		channel      ps2000a.ChannelId
		enabled      bool
		couplingType ps2000a.Coupling
		voltageRange ps2000a.RangeEnum
		analogOffset float32
	}
	// GuiDesc struct {
	// 	con                                         *psi.ConnectionDesc
	// 	FirstRange                                  ps2000a.RangeEnum
	// 	LastRange                                   ps2000a.RangeEnum
	// 	SignalGenerator                             bool
	// 	ChannelCount                                ps2000a.NumOfChannelEnum
	// 	MaxChannel                                  ps2000a.ChannelId
	// 	MaxValue                                    int16
	// 	ChannelSettings                             [ps2000a.MaxChannels]ps2000a.ChannelDesc
	// 	DigitalPorts                                int16
	// 	AwgBufferSize                               int16
	// 	AwgDACFrequency                             float64
	// 	scaleVoltages                               bool
	// 	timeBase                                    uint32
	// 	overSample                                  int16
	// 	channelPanels                               []*fyne.Container
	// 	genPanel                                    *fyne.Container
	// 	etsPanel                                    *fyne.Container
	// 	etsCycles                                   *widget.Entry
	// 	etsInterleave                               *widget.Entry
	// 	etsSampleCountEntry                         *widget.Entry
	// 	samplePanel                                 *fyne.Container
	// 	Language                                    int
	// 	themeVariant                                fyne.ThemeVariant
	// 	ChannelColors                               []color.Color
	// 	raster                                      *canvas.Raster
	// 	GenFreq                                     float64
	// 	GenFreqStep                                 float64
	// 	GenFreqDelay                                float64
	// 	GenFreqStr, GenFreqStepStr, GenFreqDelayStr string
	// 	scopeScreen                                 screenImage
	// 	signals                                     *signalDesc
	// 	timeUnit                                    float64 // 1: ns, 1000: us, 1000000:ms
	// 	timeDiv                                     float64
	// 	EtsMode                                     bool
	// 	screenLocker                                sync.Mutex
	// 	refreshTime                                 time.Time
	// 	Window                                      fyne.Window
	// 	stopChannel                                 chan struct{}
	// }
	PsDesc struct {
		//		ps2000a.PsDesc
		Con                                         *psi.ConnectionDesc
		FirstRange                                  ps2000a.RangeEnum
		LastRange                                   ps2000a.RangeEnum
		SignalGenerator                             bool
		ChannelCount                                ps2000a.NumOfChannelEnum
		MaxChannel                                  ps2000a.ChannelId
		MaxValue                                    int16
		MinValue                                    int16
		ChannelSettings                             []ps2000a.ChannelDesc
		DigitalPorts                                int16
		AwgBufferSize                               int16
		AwgDACFrequency                             float64
		scaleVoltages                               bool
		timeBase                                    uint32
		overSample                                  int16
		channelPanels                               []*fyne.Container
		genPanel                                    *fyne.Container
		etsPanel                                    *fyne.Container
		etsCycles                                   *widget.Entry
		etsInterleave                               *widget.Entry
		etsSampleCountEntry                         *widget.Entry
		samplePanel                                 *fyne.Container
		Language                                    int
		themeVariant                                fyne.ThemeVariant
		ChannelColors                               []color.Color
		raster                                      *canvas.Raster
		GenFreq                                     float64
		GenFreqStep                                 float64
		GenFreqDelay                                float64
		GenFreqStr, GenFreqStepStr, GenFreqDelayStr string
		scopeScreen                                 screenImage
		signals                                     *signalDesc
		timeUnit                                    float64 // 1: ns, 1000: us, 1000000:ms
		timeDiv                                     float64
		EtsMode                                     bool
		screenLocker                                sync.Mutex
		refreshTime                                 time.Time
		Window                                      fyne.Window
		stopChannel                                 chan struct{}
		runLock                                     chan struct{}
		buffers                                     [][]int16
		ratioMode                                   ps2000a.RatioMode
		segmentIndex                                uint32
	}
	// DemoDesc struct {
	// 	//		ps2000asim.SimDesc
	// 	GuiDesc
	// }
)
type (
	appInfo struct {
		name string
		icon fyne.Resource
		//	canv bool
		run func(fyne.Window) fyne.CanvasObject
	}
)

func (ps *PsDesc) SetDefaults() (err error) {
	//_, err = ps.SetEts(ps2000a.EtsOff, 0, 0)
	ps.EtsMode = false
	if err != nil {
		return
	}
	// for channel := ps2000a.ChA; channel <= ps.MaxChannel; channel++ {
	// 	err = ps.SetChannel(channel,
	// 		ps.ChannelSettings[channel].Enabled,
	// 		ps.ChannelSettings[channel].CoupleType,
	// 		ps.ChannelSettings[channel].Range, 0.0)
	// 	if err != nil {
	// 		return
	// 	}
	// }
	return
}

func getX() (pixelTime float64, resX, resY, hSpan, vSpan int64) {
	cmd := exec.Command("xvidtune", "-show")
	stdout, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
		panic(-1)
	}
	res := strings.Fields(string(stdout))
	pixelClock, _ := strconv.ParseFloat(res[1], 64)
	pixelTime = 1 / (1e6 * pixelClock)
	resX, _ = strconv.ParseInt(res[2], 10, 64)
	hSpan, _ = strconv.ParseInt(res[5], 10, 64)
	resY, _ = strconv.ParseInt(res[6], 10, 64)
	vSpan, _ = strconv.ParseInt(res[9], 10, 64)
	return
}

func (ps *PsDesc) adcToMv(raw int32, ch ps2000a.RangeEnum) int32 {
	return int32(math.Round(float64(raw) * float64(ps2000a.InputRanges(ch)) / float64(ps.MaxValue)))
}
func (ps *PsDesc) mvToAdc(raw int32, ch ps2000a.RangeEnum) int16 {
	return int16(math.Round(float64(raw)*float64(ps.MaxValue)) / float64(ps2000a.InputRanges(ch)))
}

func (ps *PsDesc) Str(id int) string {
	return inter.Str(ps.Language, id)
}

func (ps *PsDesc) GetInfo() (err error) {
	ps.SignalGenerator = true
	ps.EtsMode = false
	ps.FirstRange = ps2000a.Range_20mv
	ps.LastRange = ps2000a.Range_20v
	ps.ChannelCount = ps2000a.QuadScope
	ps.DigitalPorts = 0
	ps.AwgBufferSize = ps2000a.AwgMaxSigGenBufferSize
	ps.scaleVoltages = true
	ps.timeBase = 0
	ps.overSample = 1
	// info, err := ps.GetUnitInfo(ps2000a.PicoVariantInfo)
	err = ps.Con.Send(&psi.GetUnitInfoMsg{Info: ps2000a.PicoVariantInfo})
	if err != nil {
		return
	}
	resp, err := ps.Con.Receive()
	if err != nil || resp.Status() != nil {
		fmt.Println("GetInfo:", err, resp.Status())
	}
	info := resp.(*psi.GetUnitInfoRsp).InfoString
	fmt.Println("info:", info)
	switch string(info[1]) {
	case "2":
		if len(info) == 4 || (len(info) == 5 && info[4] == 'A' || info == "2205MSO") {
			ps.FirstRange = ps2000a.Range_50mv
		}
		ps.MaxChannel = ps2000a.ChB
	case "4":
		ps.ChannelCount = ps2000a.QuadScope
		ps.MaxChannel = ps2000a.ChD
	default:
		fmt.Println(info)
		fmt.Println(byte(info[2]))
		err = fmt.Errorf("OpenUnit: unknown variant info %s", info)
		return
	}
	if strings.Contains(info, "MSO") {
		ps.DigitalPorts = 2
	}
	fmt.Println("ChannelCount:", ps.ChannelCount, " MaxChannel:", ps.MaxChannel)
	ps.ChannelSettings = make([]ps2000a.ChannelDesc, ps.ChannelCount)
	ps.buffers = make([][]int16, ps.ChannelCount)
	ps.signals = &signalDesc{time: 0}
	ps.signals.signal = make([][]int32, ps.ChannelCount)

	// for ch := ps2000a.ChA; ch <= ps.MaxChannel; ch++ {
	// 	ps.ChannelSettings[ch].Enabled = true
	// 	ps.ChannelSettings[ch].CoupleType = ps2000a.Dc
	// 	ps.ChannelSettings[ch].Range = ps2000a.Range_2v
	// }
	// ps.MaxValue, err = ps.MaximumValue()
	// fmt.Println("........................................maxValue=", ps.MaxValue)
	// if err != nil {
	// 	return
	// }
	err = ps.Con.Send(&psi.MaximumValueMsg{})
	if err != nil {
		fmt.Println("MaximumValueMsg", err)
	}
	resp, err = ps.Con.Receive()
	if err != nil || resp.Status() != nil {
		fmt.Println("MaximumValueMResp", err, resp.Status())
	}
	ps.MaxValue = resp.(*psi.MaximumValueRsp).Value

	err = ps.Con.Send(&psi.MinimumValueMsg{})
	if err != nil {
		fmt.Println("MinimumValueMsg", err)
	}
	resp, err = ps.Con.Receive()
	if err != nil || resp.Status() != nil {
		fmt.Println("MinimumValueResp", err, resp.Status())
	}
	ps.MinValue = resp.(*psi.MinimumValueResp).Value
	fmt.Println("ps.MinValue:", ps.MinValue, "ps.MaxValue", ps.MaxValue)
	return
}

func (ps *PsDesc) halt() {
	fmt.Println("Stop ")
	select {
	case ps.stopChannel <- struct{}{}:
	case <-time.After(100 * time.Millisecond):
	}
	fmt.Println("Stop sent")
	err := ps.Con.Send(&psi.StopMsg{})
	if err != nil {
		fmt.Println(err)
		return
	}
	resp, err := ps.Con.Receive()
	if err != nil || resp.Status() != nil {
		fmt.Println(err, resp.Status())
		return
	}
}
func (ps *PsDesc) Menu() (err error) {
	var (
		runblockButton *widget.Button
		running        bool
	)
	runButtonHandler := func() {
		if !running {
			running = true
			runblockButton.SetText("Stop")
			ps.BlockMode()
		} else {
			running = false
			runblockButton.SetText("Run")
			ps.halt()
		}
	}
	ps.stopChannel = make(chan struct{})
	ps.runLock = make(chan struct{}, 1)
	dryRun := false
	//	os.Setenv("FYNE_THEME", "dark")
	//	os.Setenv("FYNE_THEME", "light")
	if !dryRun {
		// err = ps.OpenUnit("")
		// if err != nil {
		// 	fmt.Println("Menu open err=", err)
		// 	return
		// }
		err = ps.GetInfo()
		if err != nil {
			fmt.Println("Menu GetInfo err=", err)
			return
		}
	}
	os.Setenv("FYNE_THEME", "dark")
	ps.ChannelColors = []color.Color{chaColor, red, green, yellow}
	myApp := app.New()
	ps.themeVariant = fyne.CurrentApp().Settings().ThemeVariant()
	ps.Window = myApp.NewWindow("Border Layout")
	bottom := widget.NewLabel("bottom")
	ps.scopeScreen = ps.newScopeScreen(image.Point{1024, 768})
	ps.raster = canvas.NewRaster(ps.generator)
	right := container.New(layout.NewVBoxLayout())
	ps.setChannels(right)
	ps.setGen(right)
	ps.setSample(right)
	ps.setEtsPanel(right)
	left := container.New(layout.NewVBoxLayout())
	// vna := widget.NewButton("VNA", ps.vna)
	runblockButton = widget.NewButton("Run", runButtonHandler)
	stop := widget.NewButton("STOP", ps.halt)
	// left.Add(vna)
	left.Add(runblockButton)
	left.Add(stop)
	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.DocumentCreateIcon(), func() {
			log.Println("New document")
		}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.ContentCutIcon(), func() {}),
		widget.NewToolbarAction(theme.ContentCopyIcon(), func() {}),
		widget.NewToolbarAction(theme.ContentPasteIcon(), func() {}),
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.ViewFullScreenIcon(), func() { ps.Window.SetFullScreen(true) }),
		widget.NewToolbarAction(theme.ViewRestoreIcon(), func() { ps.Window.SetFullScreen(false) }),
		widget.NewToolbarAction(theme.SettingsIcon(), func() {}),
		widget.NewToolbarAction(theme.HelpIcon(), func() {
			log.Println("Display help")
			// ps.themeVariant = fyne.CurrentApp().Settings().ThemeVariant()
			ps.themeVariant = (ps.themeVariant + 1) % 2
			fmt.Println(ps.themeVariant)
			if ps.themeVariant != theme.VariantLight {
				fyne.CurrentApp().Settings().SetTheme(theme.DarkTheme())
			} else {
				fyne.CurrentApp().Settings().SetTheme(theme.LightTheme())
			}
		}),
	)
	content := container.NewBorder(toolbar, bottom, left, right, ps.raster)
	ps.Window.SetContent(content)
	ps.Window.Resize(fyne.NewSize(1024, 768))
	ps.refreshTime = time.Now()
	ps.SetDefaults()
	ps.Window.ShowAndRun()
	fmt.Println("run")
	return
}
