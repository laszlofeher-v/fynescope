package gui

import (
	"fynescope/control"
	"fynescope/genericps"
	"log/slog"
	"math"

	"fyne.io/fyne/v2/theme"

	"fynescope/disp7"
	"fynescope/selectscroll"
	"fynescope/sliderscroll"
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

const (
	fractionWidth     = 2
	dwellTimeScale    = 10_000_000 // Dwelltime stored in seconds; display unit is 100ns steps
	minDwellTimeNs    = 500        // Minimum dwell time in 100ns units
	dwellTimeDigits   = 11
	dwellTimeFraction = 7
	voltageDigits     = 7
	voltageFraction   = 6
)

var pow10tab [32]int

func init() {
	n := 1
	for i := range pow10tab {
		pow10tab[i] = n
		n *= 10
	}
}
func (scp *ScpDesc) applyInternalGenSettings(on bool) {
	msg := &control.GeneratorDescMsg{}
	if on {
		if scp.Settings.GenPanel.Sweep == genericps.NoSweep {
			slog.Debug("setGen NoSweep", "freq 0", scp.Settings.GenPanel.Frequency,
				"freq 1", scp.Settings.GenPanel.StopFrequency,
				"amp", scp.Settings.GenPanel.Amplitude)
			msg.StartFrequency = scp.Settings.GenPanel.Frequency
			msg.Increment = 0
			msg.DwellTime = 1
			msg.StopFrequency = scp.Settings.GenPanel.Frequency
			msg.SweepType = genericps.SweepDown // there is no "no sweep"
		} else {
			slog.Debug("setGen Sweep", "freq 0", scp.Settings.GenPanel.StartFrequency,
				"freq 1", scp.Settings.GenPanel.StopFrequency,
				"amp", scp.Settings.GenPanel.Amplitude)
			msg.DwellTime = scp.Settings.GenPanel.Dwelltime
			msg.Increment = scp.Settings.GenPanel.Increment
			msg.StartFrequency = scp.Settings.GenPanel.StartFrequency
			msg.StopFrequency = scp.Settings.GenPanel.StopFrequency
			msg.SweepType = scp.Settings.GenPanel.Sweep
		}
		msg.WaveType = scp.Settings.GenPanel.WaveType
		msg.OffsetVoltage = scp.Settings.GenPanel.OffsetVoltage
		msg.PkToPK = scp.Settings.GenPanel.Amplitude * 2
		// msg.WaveType = scp.waveType
	} else {
		msg.DwellTime = 0
		msg.OffsetVoltage = 0
		msg.PkToPK = 0
		msg.WaveType = genericps.DcVoltage
		msg.StopFrequency = scp.Settings.GenPanel.StopFrequency
		msg.SweepType = genericps.SweepDown // there is no "no sweep"
	}
	msg.Operation = genericps.EsOff //TODO ui
	msg.Shots = 0
	msg.Sweeps = 0
	msg.TriggerType = genericps.SigGenRising
	msg.TriggerSource = genericps.SigGenNone
	msg.ExtInThreshold = 0
	if scp.psControl != nil && scp.psControl.SetGeneratorCh != nil {
		scp.psControl.SetGeneratorCh <- msg
	}
}

func (scp *ScpDesc) newGenPanel(cont *fyne.Container) (err error) {

	const (
		maxV   = 2000000
		undock = "Undock"
	)
	const (
		size = 0.8
	)
	const (
		sweepOff    = "Off"
		sweepUp     = "Up"
		sweepDown   = "Down"
		sweepUpDown = "Up down"
		sweepDownUp = "Down up"
	)

	var (
		waveTypeMap     map[string]genericps.WaveTypeEnum
		waveTypeOptions []string
		sweepOptions    = []string{sweepOff, sweepUp, sweepDown, sweepUpDown, sweepDownUp}
		reloadData      chan struct{}

		// triggerCalculationOptions = []string{"Interpolated", "Fine-grained"}
		// triggerCalculationModes   = map[string]int{
		// 	triggerCalculationOptions[0]: sim.InterpolatedTrigger,
		// 	triggerCalculationOptions[1]: sim.FineGrainedTrigger,
		// }
	)
	sortWaveTypes := func() {
		type keyValDesc struct {
			key string
			val genericps.WaveTypeEnum
		}
		waveTypeMap = map[string]genericps.WaveTypeEnum{
			"Sine":      genericps.Sine,
			"Square":    genericps.Square,
			"Triangle":  genericps.Triangle,
			"RampUp":    genericps.RampUp,
			"RampDown":  genericps.RampDown,
			"SinC":      genericps.SinC,
			"Gaussian":  genericps.Gaussian,
			"HalfSine":  genericps.HalfSine,
			"DcVoltage": genericps.DcVoltage,
		}
		// log.Println("Wave consts:", genericps.Sine, genericps.Square,
		//genericps.Triangle, genericps.RampUp, genericps.RampDown, genericps.SinC,
		// genericps.Gaussian, genericps.HalfSine, genericps.DcVoltage)
		var keyVal []keyValDesc
		for key, val := range waveTypeMap {
			keyVal = append(keyVal, keyValDesc{key, val})
		}
		// log.Println("keyVal:", keyVal)
		sort.Slice(keyVal, func(i, j int) bool {
			return keyVal[i].val < keyVal[j].val
		})
		waveTypeOptions = make([]string, len(waveTypeMap))
		for i, kv := range keyVal {
			waveTypeOptions[i] = kv.key
		}
		// log.Println(waveTypeOptions)
	}

	var newGenSettings func(undockable bool) (box *fyne.Container, err error)
	newGenSettings = func(undockable bool) (box *fyne.Container, err error) {

		// func (scp *ScpDesc) newGenSettings(undockable bool) (box *fyne.Container, err error) {
		var (
			top, analog, digital, sweepBox, frqBox *fyne.Container
			freqSetAnalog, ampSetAnalog            *sliderscroll.SliderScroll
			frequency                              *disp7.DigitArray
			startFrqDisp                           *disp7.DigitArray
			stopFrqDisp                            *disp7.DigitArray
			stepFreq                               *disp7.DigitArray
			offset                                 *disp7.DigitArray
			amp                                    *disp7.DigitArray
			// raiseFallTimeDisp, noiseAmplitudeDisp, phaseNoiseDisp *disp7.DigitArray
			dwellTime    *disp7.DigitArray
			undockButton *widget.Button
		)
		if reloadData == nil {
			reloadData = make(chan struct{})
		}
		checked := func(c bool) {
			scp.Settings.GenPanel.On = c
			scp.applyInternalGenSettings(scp.Settings.GenPanel.On)
		}
		showChanged := func(c bool) {
			scp.Settings.GenPanel.Digital = c
			if c {
				analog.Hide()
				amp.SilentSetValue(int(scp.Settings.GenPanel.Amplitude))
				frequency.SilentSetValue(int(scp.Settings.GenPanel.Frequency) * pow10tab[fractionWidth])
				digital.Show()
				fyne.Do(digital.Refresh)
			} else {
				digital.Hide()
				ampSetAnalog.SilentSetValue(float64(scp.Settings.GenPanel.Amplitude))
				freqSetAnalog.SilentSetValue(scp.Settings.GenPanel.Frequency * float64(pow10tab[fractionWidth]))
				analog.Show()
				fyne.Do(analog.Refresh)
			}
		}
		show := widget.NewCheck("Digital", showChanged)
		addToTest(show, genShowId)
		check := widget.NewCheck("On", checked)
		check.Checked = scp.Settings.GenPanel.On

		addToTest(check, genCheckId)
		dwellTimeChanged := func(v float64) {
			scp.Settings.GenPanel.Dwelltime = v / dwellTimeScale
			scp.applyInternalGenSettings(check.Checked)
		}
		freqChanged := func(v float64) {
			scp.Settings.GenPanel.Frequency = v / float64(pow10tab[fractionWidth])
			scp.applyInternalGenSettings(check.Checked)
		}
		startFreqChanged := func(v float64) {
			scp.Settings.GenPanel.StartFrequency = v / float64(pow10tab[fractionWidth])
			scp.applyInternalGenSettings(check.Checked)
		}
		stopFreqChanged := func(v float64) {
			scp.Settings.GenPanel.StopFrequency = v / float64(pow10tab[fractionWidth])
			scp.applyInternalGenSettings(check.Checked)
		}
		stepFreqChanged := func(v float64) {
			scp.Settings.GenPanel.Increment = v / float64(pow10tab[fractionWidth])
			scp.applyInternalGenSettings(check.Checked)
		}
		setOffsetMinMax := func() {
			maxOffset := maxV - int(scp.Settings.GenPanel.Amplitude)
			minOffset := -maxV + int(scp.Settings.GenPanel.Amplitude)
			offset.SetMinMax(minOffset, maxOffset)
		}
		ampChanged := func(v float64) {
			scp.Settings.GenPanel.Amplitude = uint32(v)
			setOffsetMinMax()
			scp.applyInternalGenSettings(check.Checked)
		}
		offsetChanged := func(v float64) {
			scp.Settings.GenPanel.OffsetVoltage = int32(v)
			scp.applyInternalGenSettings(check.Checked)
		}
		waveTypeChanged := func(option string, e selectscroll.Exception) {
			// scp.waveType = waveTypeMap[option]
			scp.Settings.GenPanel.WaveType = waveTypeMap[option]
			// if scp.psControl.Con.ID == genericps.SimId {
			// 	switch scp.Settings.Generator.WaveType {
			// 	case genericps.Square:
			// 		fallthrough
			// 	case genericps.RampUp:
			// 		fallthrough
			// 	case genericps.RampDown:
			// 		raiseFallTimeDisp.Show()
			// 	default:
			// 		raiseFallTimeDisp.Hide()
			// 	}
			// }
			scp.applyInternalGenSettings(check.Checked)
		}
		sweepChanged := func(option string, e selectscroll.Exception) {
			if option == sweepOff {
				sweepBox.Hide()
				scp.Settings.GenPanel.Sweep = genericps.NoSweep
				scp.applyInternalGenSettings(check.Checked)
				frqBox.Show()
			} else {
				frqBox.Hide()
				sweepBox.Show()
				scp.Settings.GenPanel.Sweep = func(option string) (st genericps.SweepTypeEnum) {
					switch option {
					case sweepDown:
						st = genericps.SweepDown
					case sweepDownUp:
						st = genericps.SweepDownUp
					case sweepUp:
						st = genericps.SweepUp
					case sweepUpDown:
						st = genericps.SweepUpDown
					}
					return
				}(option)
			}
			scp.applyInternalGenSettings(check.Checked)
		}
		waveType := selectscroll.NewSelectScroll(waveTypeOptions, waveTypeChanged, waveTypeOptions[genericps.DcVoltage])
		waveType.SetSelected(waveTypeOptions[scp.Settings.GenPanel.WaveType])
		if undockable {
			undockButton = widget.NewButtonWithIcon(undock, theme.ViewFullScreenIcon(), func() {
				// Errors logged with Fyne 2.6.0, 2.6.1 2.7.0
				onWindowClose := func() {
					scp.genWindow.Hide()
					undockButton.Text = undock
					undockButton.Show()
					show.SetChecked(scp.Settings.GenPanel.Digital)
					frequency.SilentSetValue(int(scp.Settings.GenPanel.Frequency) * pow10tab[fractionWidth])
					ampSetAnalog.SilentSetValue(float64(scp.Settings.GenPanel.Amplitude))
					freqSetAnalog.SilentSetValue(scp.Settings.GenPanel.Frequency * float64(pow10tab[fractionWidth]))
					amp.SilentSetValue(int(scp.Settings.GenPanel.Amplitude))
					offset.SilentSetValue(int(scp.Settings.GenPanel.OffsetVoltage))
					scp.genTab = container.NewTabItem(tabNames[genTabIndex], scp.genTab.Content)
					check.Checked = scp.Settings.GenPanel.On
					stepFreq.SilentSetFloatValue(scp.Settings.GenPanel.Increment, fractionWidth)
					dwellTime.SilentSetValue(int(scp.Settings.GenPanel.Dwelltime * dwellTimeScale))
					startFrqDisp.SilentSetValue(int(scp.Settings.GenPanel.StartFrequency * float64(pow10tab[fractionWidth])))
					stopFrqDisp.SilentSetValue(int(scp.Settings.GenPanel.StopFrequency * float64(pow10tab[fractionWidth])))
					scp.dockTab(scp.genTab)
					scp.controlTab.SelectIndex(ftTabIndex)
					fyne.Do(scp.genTab.Content.Refresh)
				}
				scp.genWindow = scp.App.NewWindow("gen")
				var genPanel *fyne.Container
				genPanel, err = newGenSettings(false)
				if err != nil {
					return
				}
				genControls := container.New(layout.NewVBoxLayout())
				genControls.Add(genPanel)
				scp.controlTab.Remove(scp.genTab)
				scp.genWindow.SetContent(genControls)
				scp.genWindow.SetOnClosed(onWindowClose)
				scp.controlTab.SelectIndex(ftTabIndex)
				scp.genWindow.Show()

				fyne.Do(undockButton.Refresh)
				fyne.Do(genControls.Refresh)
				// scp.genFunc.Content.Refresh()
			})
		}
		freqSetAnalog = sliderscroll.NewSliderScroll(genericps.MinFrequency, genericps.SineMaxFrequency)
		addToTest(freqSetAnalog, genFreqSetId)
		freqSetAnalog.OnChanged = freqChanged
		ampSetAnalog = sliderscroll.NewSliderScroll(0, maxV)
		ampSetAnalog.SilentSetValue(float64(scp.Settings.GenPanel.Amplitude))
		addToTest(ampSetAnalog, genAmpdSetId)
		ampSetAnalog.OnChanged = ampChanged
		// fractionWidth := 2
		disp7Width := fractionWidth
		f := int(math.Round(genericps.SineMaxFrequency))
		for f > 0 {
			f /= 10
			disp7Width++
		}
		frequency, err = disp7.NewCustomDisp7Array(disp7Width, fractionWidth,
			int(genericps.SineMaxFrequency)*pow10tab[fractionWidth],
			int(genericps.MinFrequency)*pow10tab[fractionWidth],
			disp7.UnSigned, disp7.NoTrailingZeroes, scp.Window,
			scp.theme.Color(ColorNameGeneratorDisp, 0),
			disp7.ReadWrite, size*disp7.DefaultDigitWidth,
			disp7.DeafultDigitHeight, 1,
			disp7.DefaultVCursorSpace, "Frq : ", " Hz")
		if err != nil {
			return
		}
		frequency.OnChanged = freqChanged
		frequency.SilentSetValue(int(scp.Settings.GenPanel.Frequency) * pow10tab[fractionWidth])
		addToTest(frequency, genFreqId)

		dwellTime, err = disp7.NewCustomDisp7Array(dwellTimeDigits, dwellTimeFraction,
			int(genericps.MaxDwellTime)*1000, minDwellTimeNs,
			disp7.UnSigned, disp7.NoTrailingZeroes, scp.Window,
			scp.theme.Color(ColorNameGeneratorDisp, 0),
			disp7.ReadWrite, size*disp7.DefaultDigitWidth,
			disp7.DeafultDigitHeight, 1,
			disp7.DefaultVCursorSpace, "∆t   :", " s") //TODO change unit
		if err != nil {
			return
		}
		dwellTime.OnChanged = dwellTimeChanged
		dwellTime.SilentSetValue(int(scp.Settings.GenPanel.Dwelltime * dwellTimeScale))
		addToTest(dwellTime, genDwellTimeId)

		startFrqDisp, err = disp7.NewCustomDisp7Array(disp7Width, fractionWidth,
			int(genericps.SineMaxFrequency)*pow10tab[fractionWidth],
			int(genericps.MinFrequency)*pow10tab[fractionWidth],
			disp7.UnSigned, disp7.NoTrailingZeroes, scp.Window,
			scp.theme.Color(ColorNameGeneratorDisp, 0),
			disp7.ReadWrite, size*disp7.DefaultDigitWidth,
			disp7.DeafultDigitHeight, 1,
			disp7.DefaultVCursorSpace, "Low :", " Hz")
		if err != nil {
			return
		}
		startFrqDisp.OnChanged = startFreqChanged
		startFrqDisp.SilentSetValue(int(scp.Settings.GenPanel.StartFrequency) * pow10tab[fractionWidth])
		addToTest(startFrqDisp, genMinFrqId)
		stopFrqDisp, err = disp7.NewCustomDisp7Array(disp7Width, fractionWidth,
			int(genericps.SineMaxFrequency)*100, int(genericps.MinFrequency)*pow10tab[fractionWidth],
			disp7.UnSigned, disp7.NoTrailingZeroes, scp.Window,
			scp.theme.Color(ColorNameGeneratorDisp, 0),
			disp7.ReadWrite, size*disp7.DefaultDigitWidth,
			disp7.DeafultDigitHeight, 1,
			disp7.DefaultVCursorSpace, "High:", " Hz")
		if err != nil {
			return
		}
		addToTest(stopFrqDisp, genMaxFrqId)
		stopFrqDisp.OnChanged = stopFreqChanged
		stopFrqDisp.SilentSetValue(int(scp.Settings.GenPanel.StopFrequency) * pow10tab[fractionWidth])
		stepFreq, err = disp7.NewCustomDisp7Array(disp7Width, fractionWidth,
			int(genericps.SineMaxFrequency)*100, 0,
			disp7.UnSigned, disp7.NoTrailingZeroes, scp.Window,
			scp.theme.Color(ColorNameGeneratorDisp, 0),
			disp7.ReadWrite, size*disp7.DefaultDigitWidth,
			disp7.DeafultDigitHeight, 1,
			disp7.DefaultVCursorSpace, "Step :", " Hz")
		if err != nil {
			return
		}
		addToTest(stepFreq, genStepFreqId)
		stepFreq.OnChanged = stepFreqChanged
		stepFreq.SilentSetValue(int(scp.Settings.GenPanel.Increment) * pow10tab[fractionWidth])
		amp, err = disp7.NewCustomDisp7Array(7, 6, maxV, 0,
			disp7.SignedHidden, disp7.NoTrailingZeroes, scp.Window,
			scp.theme.Color(ColorNameGeneratorDisp, 0),
			disp7.ReadWrite, size*disp7.DefaultDigitWidth,
			disp7.DeafultDigitHeight, 1,
			disp7.DefaultVCursorSpace, "Amplitude:", " V")
		if err != nil {
			return
		}
		amp.SilentSetValue(int(scp.Settings.GenPanel.Amplitude))
		addToTest(amp, genAmpId)
		amp.OnChanged = ampChanged
		offset, err = disp7.NewCustomDisp7Array(7, 6,
			maxV, -maxV,
			disp7.Signed, disp7.NoTrailingZeroes, scp.Window,
			scp.theme.Color(ColorNameGeneratorDisp, 0),
			disp7.ReadWrite, size*disp7.DefaultDigitWidth,
			disp7.DeafultDigitHeight, 1,
			disp7.DefaultVCursorSpace, "Offset   :", " V")
		if err != nil {
			return
		}
		addToTest(offset, genOffsetId)
		offset.OnChanged = offsetChanged
		setOffsetMinMax()
		offset.SilentSetValue(int(scp.Settings.GenPanel.OffsetVoltage))

		fLabel := widget.NewLabel("Freq")
		voltLabel := widget.NewLabel("Amp")
		shortLineF := container.NewBorder(nil, nil, nil, fLabel, freqSetAnalog)
		shortLineA := container.NewBorder(nil, nil, nil, voltLabel, ampSetAnalog)
		analog = container.New(layout.NewVBoxLayout(), shortLineF, shortLineA)
		sweepBox = container.New(layout.NewVBoxLayout(), startFrqDisp, stopFrqDisp,
			stepFreq, dwellTime)
		frqBox = container.New(layout.NewVBoxLayout(), frequency)
		sweepMenu := selectscroll.NewSelectScroll(sweepOptions, sweepChanged, sweepDownUp)
		sweepMenu.SetSelected(sweepOptions[scp.Settings.GenPanel.Sweep+1])
		if undockable {
			top = container.New(layout.NewHBoxLayout(), show, check,
				container.New(layout.NewVBoxLayout(), waveType), undockButton)
		} else {
			top = container.New(layout.NewHBoxLayout(), show, check,
				container.New(layout.NewVBoxLayout(), waveType))
		}

		addToTest(sweepMenu, genSweepId)
		sweepMenuBox := container.New(layout.NewHBoxLayout(),
			widget.NewLabel("Sweep "), sweepMenu)
		digital = container.New(layout.NewVBoxLayout(), sweepMenuBox, frqBox,
			sweepBox, amp, offset)
		// }

		box = container.New(layout.NewVBoxLayout(), top, analog, digital)
		show.SetChecked(scp.Settings.GenPanel.Digital)
		showChanged(scp.Settings.GenPanel.Digital)
		return
	} //newGenSettings

	sortWaveTypes()
	var genPanel *fyne.Container
	genPanel, err = newGenSettings(true)
	if err != nil {
		return
	}
	cont.Add(genPanel)
	return
}
