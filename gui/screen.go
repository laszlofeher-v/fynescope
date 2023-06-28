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
	"image/color"
	"image/draw"
	"log"
	"math"
	"psgo/gui/inter"
	"psgo/ps2000a"

	//	"strings"

	// "fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"

	"strconv"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed" //"golang.org/x/image/font"
	//"golang.org/x/image/font/basicfont"
	//"golang.org/x/image/font/inconsolata"
	//"golang.org/x/image/math/fixed"
	//	"math/rand"
)

const (
	fontSize            = 16
	dpi                 = 72
	halfChWidth         = 5.0
	rangeMaxStrLen      = 6 //-500.0
	rangeMargin         = rangeMaxStrLen * halfChWidth * 2
	defaultLeftMargin   = 5.0
	defaultTopMargin    = 50.0
	defaultRightMargin  = 5.0
	defaultBottomMargin = 50.0
	numberOfDivs        = 10
)

type (
	screenImage draw.RGBA64Image
)

var (
	// scopeScreen draw.RGBA64Image
	chaColor   = color.RGBA{100, 180, 255, 255}
	green      = color.RGBA{0, 255, 0, 255}
	red        = color.RGBA{255, 0, 0, 255}
	yellow     = color.RGBA{255, 255, 0, 255}
	lightGreen = color.RGBA{64, 200, 64, 255}
	background = color.RGBA{5, 5, 5, 255}
)

func (ps *PsDesc) newScopeScreen(size image.Point) screenImage {
	// scopeScreen = image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{w, h}})
	p := image.NewRGBA(image.Rectangle{image.Point{0, 0}, size})
	//		hDiv: divisions{unitName: "ms", valDiv: 0.5},
	//		vDiv: divisions{unitName: "V", valDiv: 0.2}}
	return p
}

var face font.Face

func init() {
	f, err := opentype.Parse(gomono.TTF)
	if err != nil {
		log.Fatalf("Parse: %v", err)
	}
	face, err = opentype.NewFace(f, &opentype.FaceOptions{
		Size:    fontSize,
		DPI:     dpi,
		Hinting: font.HintingNone,
	})
	if err != nil {
		log.Fatalf("NewFace: %v", err)
	}
}

func (ps *PsDesc) addLabel(x, y int, label string, textColor color.Color) {

	d := font.Drawer{
		Dst:  ps.scopeScreen,
		Src:  image.NewUniform(textColor),
		Face: face,
		Dot:  fixed.P(x, y),
	}
	// r, a := d.BoundString(label)
	// fmt.Println("boundString:", label, r, a)
	d.DrawString(label)
}

func (ps *PsDesc) setXMargins(w float64) (leftMargin, rightMargin float64, chRangeXRelPos []float64) {
	numberOfEnabledChannels := ps.numberOfEnabledChannels()
	chRangeXRelPos = make([]float64, numberOfEnabledChannels)
	switch numberOfEnabledChannels {
	//set label pos, reduce internal width by increasing corresponding margin
	case 1: // left 1
		leftMargin = defaultLeftMargin + rangeMargin
		rightMargin = defaultRightMargin
		chRangeXRelPos[0] = defaultLeftMargin
	case 2: // left 1 right 1
		leftMargin = defaultLeftMargin + rangeMargin
		rightMargin = defaultRightMargin + rangeMargin
		chRangeXRelPos[0] = defaultLeftMargin
		chRangeXRelPos[1] = w - rangeMargin
	case 3: // left 2 right 1
		leftMargin = defaultLeftMargin + 2*rangeMargin
		rightMargin = defaultRightMargin + rangeMargin
		chRangeXRelPos[0] = defaultLeftMargin + rangeMargin
		chRangeXRelPos[1] = w - rangeMargin
		chRangeXRelPos[2] = defaultLeftMargin
	case 4: // left 2 right 2
		leftMargin = defaultLeftMargin + 2*rangeMargin
		rightMargin = defaultRightMargin + 2*rangeMargin
		chRangeXRelPos[0] = defaultLeftMargin + rangeMargin
		chRangeXRelPos[1] = w - rangeMargin
		chRangeXRelPos[2] = defaultLeftMargin
		chRangeXRelPos[3] = w - 2*rangeMargin
	default: // no channel enabled
		leftMargin = defaultLeftMargin
		rightMargin = defaultRightMargin
	}
	return
}

func (ps *PsDesc) drawHDivs(leftMargin, w, h float64) {
	dh := h / numberOfDivs
	for y := 0.0; y <= h+1; y = y + dh {
		counter := 0
		for x := 0.0; x <= w; x = x + 1.0 {
			if counter%8 < 5 {
				ps.scopeScreen.Set(int(math.Round(x+leftMargin)), int(math.Round(y+defaultTopMargin)), lightGreen)
			}
			counter++
		}
	}
}

func (ps *PsDesc) drawVDivs(leftMargin, w, h float64) {
	dw := w / numberOfDivs
	for x := 0.0; x <= w+1; x = x + dw {
		counter := 0
		for y := 0.0; y <= h; y = y + 1.0 {
			if counter%8 < 5 {
				ps.scopeScreen.Set(int(math.Round(x+leftMargin)), int(math.Round(y+defaultTopMargin)), lightGreen)
			}
			counter++
		}
	}
}

func (ps *PsDesc) drawDivisions(leftMargin, w, h float64) {
	ps.drawHDivs(leftMargin, w, h)
	ps.drawVDivs(leftMargin, w, h)
}

func (ps *PsDesc) drawSignalsNew(leftMargin, w, h float64) (err error) {
	const notUsedY = 0xffffffff
	var prevY int
	setPixel := func(x, y int, col color.Color) {
		ps.scopeScreen.Set(x, y, col)
		if prevY != notUsedY {
			d := 1
			if y < prevY {
				d = -1
			}
			for {
				ps.scopeScreen.Set(x, prevY, col)
				if prevY == y {
					break
				}
				prevY += d
			}
		}
		prevY = y
	}
	ps.drawDivisions(leftMargin, w, h)
	zeroOffset := ps.scopeScreen.Bounds().Dy() / 2
	maxScreenTime := ps.timeDiv * float64(ps.timeUnit) * 10 // 10 divs
	fmt.Println("maxScreenTime=", maxScreenTime)
	pixelTimeStep := maxScreenTime / math.Round(w)
	minY := defaultTopMargin
	maxY := h + defaultTopMargin
	if ps.signals == nil {
		fmt.Println("ps.signals == nil")
		return
	}
	for ch := range ps.signals.signal {
		if ps.ChannelSettings[ch].Enabled {
			// for i := 0; i < len(ps.signals.signal[ch]); i++ {
			//	fmt.Println("i=", i, "time=", int(float64(i)*ps.signals.timeInterval), "v=", ps.signals.signal[ch][i])
			// if ps.signals.signal[ch][i] == 0 {
			// 	k := i
			// 	for ; ps.signals.signal[ch][k] < 100 && ps.signals.signal[ch][k] > -100; k++ {
			// 		fmt.Println("k=", k, "v=", ps.signals.signal[ch][k])

			// 	}
			// 	for j := k; j < len(ps.signals.signal[ch]); j++ {
			// 		if ps.signals.signal[ch][i] == 0 {
			// 			fmt.Println("period time=", float64(j-i)*ps.signals.timeInterval)
			// 			goto exit
			// 		}
			// 	}
			// }
			// }
			// exit:
			prevY = notUsedY
			fmt.Println("ch=", ch)
			col := ps.ChannelColors[ch]
			yScale := h / (2.0 * ps2000a.RangeValuesMv[ps.ChannelSettings[ch].Range])
			signalTime := 0.0
			displayTime := 0.0
			fmt.Println("pixelTimeStep=", pixelTimeStep, "ps.signals.timeInterval=", ps.signals.timeInterval)
			signalIndex := int32(0)
			for x := leftMargin; x <= w+leftMargin; x = x + 1.0 {
				sum := float64(0.0)
				n := 0
				if ps.signals.timeInterval < pixelTimeStep {
					for signalTime = displayTime; signalTime <= displayTime+pixelTimeStep; signalTime += ps.signals.timeInterval /*ps.signals.timeInterval*/ {
						if signalIndex >= ps.signals.length {
							break
						}
						sum += float64(ps.signals.signal[ch][signalIndex])
						signalIndex++
						n++
					}
					if n == 0 {
						fmt.Println("BREAK   signalTime=", signalTime, "displayTime=", displayTime)
						break
					}

					if ps.ChannelSettings[ch].Inverted {
						sum = -sum
					}
					//	fmt.Println("signalTime=", signalTime, "sum=", sum, "n=", n)
					sum = float64(zeroOffset) - yScale*sum/float64(n)
					// fmt.Println("minY=", minY, "maxY=", maxY, "x=", x, "displayTime=", displayTime, "signalTime=", signalTime, "sum=", sum)
				} else {
				}
				if sum >= minY && sum <= maxY {
					//	fmt.Println("minY=", minY, "maxY=", maxY, "x=", x, "displayTime=", displayTime, "signalTime=", signalTime, "sum=", sum)
					setPixel(int(math.Round(x)), int(math.Round(sum)), col)
				} else {
					fmt.Println("OUT")
				}
				displayTime += pixelTimeStep
			}
		}
	}
	return
}

func (ps *PsDesc) drawSignals(leftMargin, w, h float64) (err error) {
	ps.drawDivisions(leftMargin, w, h)
	//TODO offset setting per channel
	zeroOffset := ps.scopeScreen.Bounds().Dy() / 2
	if ps.signals == nil {
		fmt.Println("ps.signals == nil!!!!!!!!!!!!!!!!!!!")
		return
	}
	// fmt.Println("ps.signals.length:", ps.signals.length, "ps.signals.time:", ps.signals.time)
	maxScreenTime := ps.timeDiv * float64(ps.timeUnit) * 10 // 10 divs
	// if ps.EtsMode {
	// 	maxScreenTime=maxScreenTime/1000.0
	// }
	// fmt.Println("ps.timeDiv:", ps.timeDiv, "ps.timeUnit:", ps.timeUnit)
	// fmt.Println("ps.signals.length:", ps.signals.length, "ps.signals.time:", ps.signals.time, "maxScreenTime", maxScreenTime)
	// maxTimeIndex := (maxScreenTime / ps.signals.time) * float64(ps.signals.length)
	// dt := maxTimeIndex / math.Round(w)

	maxTimeIndex := (maxScreenTime / ps.signals.timeInterval)
	dt := maxTimeIndex / math.Round(w)
	if maxTimeIndex > float64(ps.signals.length-1) {
		maxTimeIndex = float64(ps.signals.length - 1)
	}
	minY := int(math.Round(defaultTopMargin))
	maxY := int(math.Round(h + defaultTopMargin))
	//	fmt.Println("screen ps.ChannelSettings[ch].Range:", ps.ChannelSettings[0].Range, " RangeValuesMv:", ps2000a.RangeValuesMv[ps.ChannelSettings[0].Range])
	for ch := range ps.signals.signal {
		if ps.signals.signal[ch] != nil && ps.ChannelSettings[ch].Enabled {
			col := ps.ChannelColors[ch]
			yScale := h / (2.0 * ps2000a.RangeValuesMv[ps.ChannelSettings[ch].Range])
			var prev int
			t := 0.0
			for i := 0; t < maxTimeIndex; i = i + 1 {
				index := int(math.Round(t))
				if index >= len(ps.signals.signal[ch]) {
					break
				}
				s := ps.signals.signal[ch][index]
				t += dt
				if ps.ChannelSettings[ch].Inverted {
					s = -s
				}
				x := int(math.Round(float64(i) + leftMargin))
				if x >= int(math.Round(w+leftMargin)) {
					break
				}
				y := int(-yScale*float64(s)) + zeroOffset
				if y >= minY && y <= maxY {
					ps.scopeScreen.Set(x, y, col)
				}
				if i > 1 {
					d := 1
					if y < prev {
						d = -1
					}
					for {
						if prev >= minY && prev <= maxY {
							ps.scopeScreen.Set(x, prev, col)
						}
						if prev == y {
							break
						}
						prev += d
					}
				}
				prev = y
			}
		}
	}
	return
}

func (ps *PsDesc) switchUpTimeUnit(dt, timeUnit float64) (newDt float64, unitName string) {
	//	fmt.Println("up:", timeUnit)
	switch timeUnit {
	case 0.1:
		newDt = dt / 1000.0
		unitName = ps.Str(inter.Nsec)
	case 1.0:
		newDt = dt / 1000.0
		unitName = ps.Str(inter.Usec)
	case 1000.0:
		newDt = dt / 1000.0
		unitName = ps.Str(inter.Msec)
	case 1.0e6:
		newDt = dt / 1000.0
		unitName = ps.Str(inter.Sec)
	case 1.0e9:
		newDt = dt
		unitName = ps.Str(inter.Sec)
	default:
		newDt = dt
		unitName = ps.Str(inter.Sec)
	}
	return
}

func (ps *PsDesc) getTimeUnitName(timeUnit float64) (unitName string) {
	switch timeUnit {
	case 0.1:
		unitName = ps.Str(inter.Psec)
	case 1.0:
		unitName = ps.Str(inter.Nsec)
	case 1000.0:
		unitName = ps.Str(inter.Usec)
	case 1.0e6:
		unitName = ps.Str(inter.Msec)
	case 1.0e9:
		unitName = ps.Str(inter.Sec)
	default:
		unitName = ps.Str(inter.Sec)
	}
	return
}
func (ps *PsDesc) drawTimeLabels(leftMargin, w, h float64, textColor color.Color) {
	//TODO call only when changes
	var unitName string
	dt := ps.timeDiv
	switch dt {
	case 1.0:
		unitName = ps.getTimeUnitName(ps.timeUnit)
	case 2.0:
		unitName = ps.getTimeUnitName(ps.timeUnit)
	case 5.0:
		unitName = ps.getTimeUnitName(ps.timeUnit)
	case 10.0:
		unitName = ps.getTimeUnitName(ps.timeUnit)
	case 20.0:
		unitName = ps.getTimeUnitName(ps.timeUnit)
	case 50.0:
		unitName = ps.getTimeUnitName(ps.timeUnit)
	case 100.0:
		dt, unitName = ps.switchUpTimeUnit(dt, ps.timeUnit)
	case 200.0:
		dt, unitName = ps.switchUpTimeUnit(dt, ps.timeUnit)
	case 500.0:
		dt, unitName = ps.switchUpTimeUnit(dt, ps.timeUnit)
	case 1000.0:
		unitName = ps.getTimeUnitName(ps.timeUnit)
		//TODO change to hour
	case 2000.0:
		unitName = ps.getTimeUnitName(ps.timeUnit)
	case 5000.0:
		unitName = ps.getTimeUnitName(ps.timeUnit)
	default:
		fmt.Println("==================unhandled case==================", dt)
	}
	ps.addLabel(int(math.Round(leftMargin+4*halfChWidth)), int(math.Round(h+defaultTopMargin+defaultBottomMargin/float64(2.0))),
		unitName, textColor)
	for v, x := 0.0, leftMargin; x <= (w + leftMargin + 1); x, v = x+w/10, v+dt {
		//TODO font metric
		vstr := strconv.FormatFloat(v, 'f', 1, 32)
		n := float64(len(vstr))
		ps.addLabel(int(math.Round(x-n*halfChWidth)), int(math.Round(h+defaultTopMargin+defaultBottomMargin/float64(2.0))),
			vstr, textColor)
	}
}

func (ps *PsDesc) drawVoltLabels(leftMargin, w, h float64, chRangeXRelPos []float64, textColor color.Color) {
	enabledChannelIndex := 0
	padding := []byte("      ")
	for i := range ps.ChannelSettings {
		if ps.ChannelSettings[i].Enabled {
			startValue := ps2000a.RangeValuesMv[ps.ChannelSettings[i].Range]
			x := chRangeXRelPos[enabledChannelIndex]
			var spacer, unitName string
			if enabledChannelIndex%2 == 0 {
				spacer = "    "
			}
			if startValue >= 1000.0 {
				startValue = startValue / 1000.0
				unitName = ps.Str(inter.Volt)
			} else {
				unitName = ps.Str(inter.Millivolt)
			}
			ps.addLabel(int(math.Round(x)), int(math.Round(defaultTopMargin+2*fontSize)),
				spacer+unitName, ps.ChannelColors[i])
			dv := startValue / 5.0
			startValue = -startValue
			for v, y := startValue, defaultBottomMargin; y <= (h + defaultTopMargin + 1.0); y, v = y+h/10, v+dv {
				//TODO font metric
				vstr := strconv.FormatFloat(v, 'f', 1, 32)
				if enabledChannelIndex%2 == 0 {
					vstr = string(padding[0:6-len(vstr)]) + vstr
				}
				ps.addLabel(int(math.Round(x)), int(math.Round(h-y+defaultTopMargin+defaultBottomMargin+fontSize/3.0)),
					vstr, ps.ChannelColors[i])
			}
			enabledChannelIndex++
		}
	}
}

func (ps *PsDesc) drawLabels(leftMargin, w, h float64, chRangeXRelPos []float64, textColor color.Color) {
	//	metrics := face.Metrics()
	//	fmt.Println("metrics:", metrics)
	ps.drawTimeLabels(leftMargin, w, h, textColor)
	ps.drawVoltLabels(leftMargin, w, h, chRangeXRelPos, textColor)
}

func (ps *PsDesc) setColors() (bgColor, textColor color.Color) {
	var systemThemVariant = fyne.CurrentApp().Settings().ThemeVariant() // does not work
	if systemThemVariant != ps.themeVariant {
		//		fmt.Println("systemThemVariant!=ps.themeVariant")
	}
	themeVariant := ps.themeVariant
	switch themeVariant {
	case theme.VariantDark:
		bgColor = background
		textColor = color.White
	case theme.VariantLight:
		bgColor = color.White
		textColor = background
	default:
		bgColor = background
		textColor = color.White
	}
	return
}
func (ps *PsDesc) generator(wInt int, hInt int) image.Image {
	defer ps.screenLocker.Unlock()
	ps.screenLocker.Lock()
	w := float64(wInt)
	h := float64(hInt)
	rect := ps.scopeScreen.Bounds()
	if wInt != rect.Max.X-rect.Min.X || hInt != rect.Max.Y-rect.Min.Y { // window resized
		ps.scopeScreen = ps.newScopeScreen(image.Point{wInt, hInt})
		rect = ps.scopeScreen.Bounds()
	}
	bgColor, textColor := ps.setColors()
	//TODO clear changed areas only
	draw.Draw(ps.scopeScreen, rect, &image.Uniform{bgColor}, image.ZP, draw.Src)
	leftMargin, rightMargin, chRangeXRelPos := ps.setXMargins(w)
	if leftMargin < 4*halfChWidth {
		leftMargin = 4 * halfChWidth
	}
	if rightMargin < 5*halfChWidth {
		rightMargin = 5 * halfChWidth
	}
	w -= leftMargin + rightMargin
	h -= defaultTopMargin + defaultBottomMargin
	if w < numberOfDivs || h < numberOfDivs {
		return ps.scopeScreen
	}
	ps.drawLabels(leftMargin, w, h, chRangeXRelPos, textColor)
	ps.drawSignals(leftMargin, w, h)
	return ps.scopeScreen
}
