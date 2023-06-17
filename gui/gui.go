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

	//	"math"

	"fyne.io/fyne/v2/theme"

	"image/color"
	//	"image/draw"

	//	"math/rand"
	"os"
	"os/exec"
	"psgo/gui/inter"
	"psgo/ps2000a"
	"psgo/ps2000sim"
	"strconv"
	"strings"

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
	GuiDesc struct {
		FirstRange                                  ps2000a.RangeEnum
		LastRange                                   ps2000a.RangeEnum
		SignalGenerator                             bool
		Ets                                         bool
		ChannelCount                                ps2000a.NumOfChannelEnum
		MaxChannel                                  ps2000a.ChannelId
		MaxValue                                    int16
		ChannelSettings                             [ps2000a.MaxChannels]ps2000a.ChannelDesc
		DigitalPorts                                int16
		AwgBufferSize                               int16
		AwgDACFrequency                             float64
		scaleVoltages                               bool
		timeBase                                    uint32
		overSample                                  int16
		channelPanels                               []*fyne.Container
		genPanel                                    *fyne.Container
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
	}
	PsDesc struct {
		ps2000a.PsDesc
		GuiDesc
	}
	DemoDesc struct {
		ps2000sim.Ps2000sim
		GuiDesc
	}
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
	_, err = ps.SetEts(ps2000a.EtsOff, 0, 0)
	if err != nil {
		return
	}
	for channel := ps2000a.ChA; channel < ps.MaxChannel; channel++ {
		err = ps.SetChannel(channel,
			ps.ChannelSettings[channel].Enabled,
			ps.ChannelSettings[channel].CoupleType,
			ps.ChannelSettings[channel].Range, 0.0)
		if err != nil {
			return
		}
	}
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

func (ps *PsDesc) Str(id int) string {
	return inter.Str(ps.Language, id)
}

func (ps *PsDesc) GetInfo() (err error) {
	ps.SignalGenerator = true
	ps.Ets = false
	ps.FirstRange = ps2000a.Range_20mv
	ps.LastRange = ps2000a.Range_20v
	ps.ChannelCount = ps2000a.QuadScope
	ps.DigitalPorts = 0
	ps.AwgBufferSize = ps2000a.AwgMaxSigGenBufferSize
	ps.scaleVoltages = true
	ps.timeBase = 0
	ps.overSample = 1
	info, err := ps.GetUnitInfo(ps2000a.PicoVariantInfo)
	if err != nil {
		return
	}
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
	// for ch := ps2000a.ChA; ch <= ps.MaxChannel; ch++ {
	// 	ps.ChannelSettings[ch].Enabled = true
	// 	ps.ChannelSettings[ch].CoupleType = ps2000a.Dc
	// 	ps.ChannelSettings[ch].Range = ps2000a.Range_2v
	// }
	ps.MaxValue, err = ps.MaximumValue()
	fmt.Println("........................................maxValue=", ps.MaxValue)
	if err != nil {
		return
	}
	return
}

func (ps *PsDesc) Menu() (err error) {
	dryRun := false
	//	os.Setenv("FYNE_THEME", "dark")
	//	os.Setenv("FYNE_THEME", "light")
	if !dryRun {
		err = ps.OpenUnit("")
		if err != nil {
			fmt.Println("Menu open err=", err)
			return
		}
		err = ps.GetInfo()
		if err != nil {
			fmt.Println("Menu GetInfo err=", err)
			return
		}
	}
	os.Setenv("FYNE_THEME", "dark")
	ps.ChannelColors = []color.Color{chaColor, red, green, yellow}
	//	ps.SetDefaults()
	myApp := app.New()
	ps.themeVariant = fyne.CurrentApp().Settings().ThemeVariant()
	// ps.ThemeVariant = ps.myApp.Settings().ThemeVariant()
	// fmt.Println(ps.ThemeVariant)
	myWindow := myApp.NewWindow("Border Layout")
	//	myApp.Settings().SetTheme(theme.DarkTheme())
	// top := widget.NewLabel("top")
	//	top.Resize()
	//left := widget.NewLabel("left")
	bottom := widget.NewLabel("bottom")
	//right := widget.NewLabel("right")
	// scopeScreen = image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{1024, 768}})
	ps.scopeScreen = ps.newScopeScreen(image.Point{1024, 768})
	//	draw.Draw(scopeScreen, scopeScreen.Bounds(), &image.Uniform{blue}, image.ZP, draw.Src)
	ps.raster = canvas.NewRaster(ps.generator)
	// cha := widget.NewCheck("CH A", func(c bool) {
	// 	raster.Refresh()
	// })

	// right := container.New(layout.NewVBoxLayout(), cha, layout.NewSpacer(), chb, layout.NewSpacer(),
	// 	chc, layout.NewSpacer(), chd, layout.NewSpacer())
	right := container.New(layout.NewVBoxLayout())
	ps.setChannels(right)
	ps.setGen(right)
	ps.setSample(right)

	left := widget.NewButton("VNA", ps.vna)
	// widget.NewButton("VNA", func() {
	// 	const nOfSamples = 32768 * 1024
	// 	ps.signals = &signalDesc{time: 1.0e9, length: nOfSamples} // 1E9 ns = 1 s
	// 	ps.signals.signal = make([][]int32, 4)
	// 	ps.signals.signal[0] = make([]int32, nOfSamples)
	// 	ps.signals.signal[1] = make([]int32, nOfSamples)
	// 	ps.signals.signal[2] = make([]int32, nOfSamples)
	// 	ps.signals.signal[3] = make([]int32, nOfSamples)
	// 	v := 0
	// 	d := 1
	// 	for i, _ := range ps.signals.signal[0] {
	// 		ps.signals.signal[0][i] = int32(100 * math.Cos(float64(i)/float64(100)))
	// 		ps.signals.signal[1][i] = int32(100 * (1 + math.Sin(float64(i)/float64(100))))
	// 		y := 0
	// 		if i%128 > 64 {
	// 			y = 1
	// 		}
	// 		ps.signals.signal[2][i] = int32(100 * y)
	// 		if i%100 == 0 {
	// 			d = -d
	// 		}
	// 		v = v + d
	// 		ps.signals.signal[3][i] = int32(10 * v)
	// 	}
	// 	ps.raster.Refresh()
	// 	//ps.vna()
	// })
	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.DocumentCreateIcon(), func() {
			log.Println("New document")
		}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.ContentCutIcon(), func() {}),
		widget.NewToolbarAction(theme.ContentCopyIcon(), func() {}),
		widget.NewToolbarAction(theme.ContentPasteIcon(), func() {}),
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.ViewFullScreenIcon(), func() { myWindow.SetFullScreen(true) }),
		widget.NewToolbarAction(theme.ViewRestoreIcon(), func() { myWindow.SetFullScreen(false) }),
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
	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(1024, 768))
	myWindow.ShowAndRun()
	fmt.Println("run")
	return
}
