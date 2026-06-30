package gui

import (
	"fmt"
	"image/color"
	"log/slog"
	"math"
	"fynescope/control"
	"fynescope/genericps"

	"fyne.io/fyne/v2/theme"

	"fynescope/disp7"
	"fynescope/selectscroll"
	"fynescope/settings"
	"fynescope/sim"
	"fynescope/sliderscroll"
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func (scp *ScpDesc) applySimGenSettings(ch genericps.ChannelId, genSettings *settings.GeneratorSettings) {
	msg := &control.GeneratorDescMsg{}
	msg.Channel = ch
	msg.On = genSettings.On
	if genSettings.On {
		if genSettings.Sweep == genericps.NoSweep {
			slog.Debug("setGen NoSweep", "freq 0", genSettings.Frequency,
				"freq 1", genSettings.StopFrequency,
				"amp", genSettings.Amplitude)
			msg.StartFrequency = genSettings.Frequency
			msg.Increment = 0
			msg.DwellTime = 1
			msg.StopFrequency = genSettings.Frequency
			msg.SweepType = genericps.SweepDown // there is no "no sweep"
		} else {
			slog.Debug("setGen Sweep", "freq 0", genSettings.StartFrequency,
				"freq 1", genSettings.StopFrequency,
				"amp", genSettings.Amplitude)
			msg.DwellTime = genSettings.Dwelltime
			msg.Increment = genSettings.Increment
			msg.StartFrequency = genSettings.StartFrequency
			msg.StopFrequency = genSettings.StopFrequency
			msg.SweepType = genSettings.Sweep
		}
		msg.WaveType = genSettings.WaveType
		msg.OffsetVoltage = genSettings.OffsetVoltage
		msg.PkToPK = genSettings.Amplitude * 2
	} else {
		msg.DwellTime = 0
		msg.OffsetVoltage = 0
		msg.PkToPK = 0
		msg.WaveType = genericps.DcVoltage
		msg.StopFrequency = genSettings.StopFrequency
		msg.SweepType = genericps.SweepDown // there is no "no sweep"
	}
	msg.Operation = genericps.EsOff //TODO ui
	msg.Shots = 0
	msg.Sweeps = 0
	msg.TriggerType = genericps.SigGenRising
	msg.TriggerSource = genericps.SigGenNone
	msg.ExtInThreshold = 0
	msg.Phase = genSettings.Phase
	if scp.psControl != nil && scp.psControl.SetSimGenCh != nil {
		scp.psControl.SetSimGenCh <- msg
	}
	scp.SaveSettings()
}

func (scp *ScpDesc) newSimGenPanel(cont *fyne.Container, undockable bool) (err error) {

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

		triggerCalculationOptions = []string{"Interpolated", "Fine-grained"}
		triggerCalculationModes   = map[string]int{
			triggerCalculationOptions[0]: sim.InterpolatedTrigger,
			triggerCalculationOptions[1]: sim.FineGrainedTrigger,
		}
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

	onTriggerCalculationModeChange := func(option string, ex selectscroll.Exception) {
		mode := triggerCalculationModes[option]
		scp.Settings.Trigger.CalculationMode = mode
		sim.SetTriggerCalculationMode(mode)
		scp.SaveSettings()
	}

	var newGenSettings func(ch genericps.ChannelId, undockable bool) (box *fyne.Container, err error)

	newGenSettings = func(ch genericps.ChannelId, undockable bool) (box *fyne.Container, err error) {
		genSettings := &scp.Settings.SimGenPanel[ch]
		var (
			top, analog, digital, sweepBox, frqBox                           *fyne.Container
			freqSetAnalog, ampSetAnalog                                      *sliderscroll.SliderScroll
			frequency                                                        *disp7.DigitArray
			startFrqDisp                                                     *disp7.DigitArray
			stopFrqDisp                                                      *disp7.DigitArray
			stepFreq                                                         *disp7.DigitArray
			offset                                                           *disp7.DigitArray
			amp                                                              *disp7.DigitArray
			raiseFallTimeDisp, noiseAmplitudeDisp, phaseNoiseDisp, phaseDisp *disp7.DigitArray
			dwellTime                                                        *disp7.DigitArray
			undockButton                                                     *widget.Button
			nameLabel                                                        *canvas.Text
		)
		if reloadData == nil {
			reloadData = make(chan struct{})
		}
		chCol := scp.Settings.Channels[ch].Col[scp.Settings.ChannelColorIndex]
		scp.channelViewers[ch].simGenDisplays = nil
		checked := func(c bool) {
			genSettings.On = c
			scp.applySimGenSettings(ch, genSettings)
		}
		showChanged := func(c bool) {
			genSettings.Digital = c
			if c {
				analog.Hide()
				amp.SilentSetValue(int(genSettings.Amplitude))
				frequency.SilentSetValue(int(genSettings.Frequency) * pow10tab[fractionWidth])
				digital.Show()
				fyne.Do(digital.Refresh)
			} else {
				digital.Hide()
				ampSetAnalog.SilentSetValue(float64(genSettings.Amplitude))
				freqSetAnalog.SilentSetValue(genSettings.Frequency * float64(pow10tab[fractionWidth]))
				analog.Show()
				fyne.Do(analog.Refresh)
			}
		}
		show := widget.NewCheck("Digital", showChanged)
		addToTest(show, genShowId)
		check := widget.NewCheck("On", checked)
		check.Checked = genSettings.On

		addToTest(check, genCheckId)
		dwellTimeChanged := func(v float64) {
			genSettings.Dwelltime = v / 10000000
			scp.applySimGenSettings(ch, genSettings)
		}

		// Initialize channel name label
		nameLabel = canvas.NewText(channelNames[ch], chCol)
		nameLabel.TextStyle.Bold = true
		scp.channelViewers[ch].simGenNameLabel = nameLabel

		freqChanged := func(v float64) {
			genSettings.Frequency = v / 100
			scp.applySimGenSettings(ch, genSettings)
		}
		startFreqChanged := func(v float64) {
			genSettings.StartFrequency = v / 100
			scp.applySimGenSettings(ch, genSettings)
		}
		stopFreqChanged := func(v float64) {
			genSettings.StopFrequency = v / 100
			scp.applySimGenSettings(ch, genSettings)
		}
		stepFreqChanged := func(v float64) {
			genSettings.Increment = v / 100
			scp.applySimGenSettings(ch, genSettings)
		}
		setOffsetMinMax := func() {
			maxOffset := maxV - int(genSettings.Amplitude)
			minOffset := -maxV + int(genSettings.Amplitude)
			offset.SetMinMax(minOffset, maxOffset)
		}
		ampChanged := func(v float64) {
			genSettings.Amplitude = uint32(v)
			setOffsetMinMax()
			scp.applySimGenSettings(ch, genSettings)
		}
		offsetChanged := func(v float64) {
			genSettings.OffsetVoltage = int32(v)
			scp.applySimGenSettings(ch, genSettings)
		}
		// if scp.psControl.Con.ID == genericps.SimId {
		sim.SetRaiseFallTimePercent(genSettings.RaiseFallTimePercent / 100.0)
		raiseFallTimeDisp, err = disp7.NewCustomDisp7Array(5, 2,
			10000, 0,
			disp7.UnSigned, disp7.NoTrailingZeroes, scp.Window,
			chCol,
			disp7.ReadWrite, size*disp7.DefaultDigitWidth,
			disp7.DeafultDigitHeight, 1,
			disp7.DefaultVCursorSpace, "Rise/Fall:", " %")
		scp.channelViewers[ch].simGenDisplays = append(scp.channelViewers[ch].simGenDisplays, raiseFallTimeDisp)
		addToTest(raiseFallTimeDisp, genRiseFallTimeId)
		if err != nil {
			return
		}
		raiseFallTimeDisp.OnChanged = func(v float64) {
			genSettings.RaiseFallTimePercent = v / 100.0 // Value comes back in % * 100, e.g. 159 for 1.59%
			sim.SetRaiseFallTimePercent(v / 10000.0)     // 1.59% -> 0.0159
			scp.applySimGenSettings(ch, genSettings)
		}
		raiseFallTimeDisp.SilentSetValue(int(genSettings.RaiseFallTimePercent * 100))

		sim.SetNoiseAmplitude(genSettings.NoiseAmplitude)
		noiseAmplitudeDisp, err = disp7.NewCustomDisp7Array(5, 0,
			10000, 0,
			disp7.UnSigned, disp7.NoTrailingZeroes, scp.Window,
			chCol,
			disp7.ReadWrite, size*disp7.DefaultDigitWidth,
			disp7.DeafultDigitHeight, 1,
			disp7.DefaultVCursorSpace, "Noise:", " mV")
		scp.channelViewers[ch].simGenDisplays = append(scp.channelViewers[ch].simGenDisplays, noiseAmplitudeDisp)
		addToTest(noiseAmplitudeDisp, genNoiseAmpId)
		if err != nil {
			return
		}
		noiseAmplitudeDisp.OnChanged = func(v float64) {
			genSettings.NoiseAmplitude = v
			sim.SetNoiseAmplitude(v)
			scp.applySimGenSettings(ch, genSettings)
		}
		noiseAmplitudeDisp.SilentSetValue(int(genSettings.NoiseAmplitude))
		sim.SetPhaseNoiseDegree(genSettings.PhaseNoiseDegree)
		phaseNoiseDisp, err = disp7.NewCustomDisp7Array(5, 2,
			36000, 0,
			disp7.UnSigned, disp7.NoTrailingZeroes, scp.Window,
			chCol,
			disp7.ReadWrite, size*disp7.DefaultDigitWidth,
			disp7.DeafultDigitHeight, 1,
			disp7.DefaultVCursorSpace, "Phase Noise:", " °")
		scp.channelViewers[ch].simGenDisplays = append(scp.channelViewers[ch].simGenDisplays, phaseNoiseDisp)
		addToTest(phaseNoiseDisp, genPhaseNoiseId)
		if err != nil {
			return
		}
		phaseNoiseDisp.OnChanged = func(v float64) {
			genSettings.PhaseNoiseDegree = v / 100.0
			slog.Debug("phaseNoiseDisp", "PhaseNoiseDegree", genSettings.PhaseNoiseDegree)
			sim.SetPhaseNoiseDegree(v / 100.0)
			scp.applySimGenSettings(ch, genSettings)
		}
		phaseNoiseDisp.SilentSetValue(int(math.Round(genSettings.PhaseNoiseDegree * 100)))
		// }
		phaseDisp, err = disp7.NewCustomDisp7Array(3, 0,
			360, 0,
			disp7.UnSigned, disp7.NoTrailingZeroes, scp.Window,
			chCol,
			disp7.ReadWrite, size*disp7.DefaultDigitWidth,
			disp7.DeafultDigitHeight, 1,
			disp7.DefaultVCursorSpace, "Phase      :", " °")
		scp.channelViewers[ch].simGenDisplays = append(scp.channelViewers[ch].simGenDisplays, phaseDisp)
		addToTest(phaseDisp, genPhaseId)
		if err != nil {
			return
		}
		phaseDisp.OnChanged = func(v float64) {
			genSettings.Phase = v
			scp.applySimGenSettings(ch, genSettings)
		}
		phaseDisp.SilentSetValue(int(genSettings.Phase))

		waveTypeChanged := func(option string, e selectscroll.Exception) {
			// scp.waveType = waveTypeMap[option]
			genSettings.WaveType = waveTypeMap[option]
			// if scp.psControl.Con.ID == genericps.SimId {
			switch genSettings.WaveType {
			case genericps.Square:
				fallthrough
			case genericps.RampUp:
				fallthrough
			case genericps.RampDown:
				raiseFallTimeDisp.Show()
			default:
				raiseFallTimeDisp.Hide()
			}
			// }
			scp.applySimGenSettings(ch, genSettings)
		}
		sweepChanged := func(option string, e selectscroll.Exception) {
			if option == sweepOff {
				sweepBox.Hide()
				genSettings.Sweep = genericps.NoSweep
				scp.applySimGenSettings(ch, genSettings)
				frqBox.Show()
			} else {
				frqBox.Hide()
				sweepBox.Show()
				genSettings.Sweep = func(option string) (st genericps.SweepTypeEnum) {
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
			scp.applySimGenSettings(ch, genSettings)
		}
		waveType := selectscroll.NewSelectScroll(waveTypeOptions, waveTypeChanged, waveTypeOptions[genericps.DcVoltage])
		waveType.SetSelected(waveTypeOptions[genSettings.WaveType])
		if undockable {
			undockButton = widget.NewButtonWithIcon(undock, theme.ViewFullScreenIcon(), func() {
				// Errors logged with Fyne 2.6.0, 2.6.1 2.7.0
				onWindowClose := func() {
					scp.genWindow.Hide()
					undockButton.Text = undock
					undockButton.Show()
					show.SetChecked(genSettings.Digital)
					frequency.SilentSetValue(int(genSettings.Frequency) * pow10tab[fractionWidth])
					ampSetAnalog.SilentSetValue(float64(genSettings.Amplitude))
					freqSetAnalog.SilentSetValue(genSettings.Frequency * float64(pow10tab[fractionWidth]))
					amp.SilentSetValue(int(genSettings.Amplitude))
					offset.SilentSetValue(int(genSettings.OffsetVoltage))
					// if scp.psControl.Con.ID == genericps.SimId && raiseFallTimeDisp != nil {
					raiseFallTimeDisp.SilentSetValue(int(genSettings.RaiseFallTimePercent * 100))
					// triggerTimeOffsetDisp.SilentSetValue(int(genSettings.TriggerTimeOffset))
					noiseAmplitudeDisp.SilentSetValue(int(genSettings.NoiseAmplitude))
					phaseNoiseDisp.SilentSetValue(int(math.Round(genSettings.PhaseNoiseDegree * 100)))
					// }
					scp.genTab = container.NewTabItem(tabNames[genTabIndex], scp.genTab.Content)
					check.Checked = genSettings.On
					stepFreq.SilentSetFloatValue(genSettings.Increment, fractionWidth)
					dwellTime.SilentSetValue(int(genSettings.Dwelltime * 10000000))
					startFrqDisp.SilentSetValue(int(genSettings.StartFrequency * 100))
					stopFrqDisp.SilentSetValue(int(genSettings.StopFrequency * 100))
					scp.dockTab(scp.genTab)
					scp.controlTab.SelectIndex(ftTabIndex)
					fyne.Do(scp.genTab.Content.Refresh)
				}
				scp.genWindow = scp.App.NewWindow("gen")
				var genPanel *fyne.Container
				genPanel, err = newGenSettings(ch, false)
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
		ampSetAnalog.SilentSetValue(float64(genSettings.Amplitude))
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
			chCol,
			disp7.ReadWrite, size*disp7.DefaultDigitWidth,
			disp7.DeafultDigitHeight, 1,
			disp7.DefaultVCursorSpace, "Frq : ", " Hz")
		scp.channelViewers[ch].simGenDisplays = append(scp.channelViewers[ch].simGenDisplays, frequency)
		if err != nil {
			return
		}
		frequency.OnChanged = freqChanged
		frequency.SilentSetValue(int(genSettings.Frequency) * pow10tab[fractionWidth])
		addToTest(frequency, genFreqId)

		dwellTime, err = disp7.NewCustomDisp7Array(11, 7,
			int(genericps.MaxDwellTime)*1000, int(500),
			disp7.UnSigned, disp7.NoTrailingZeroes, scp.Window,
			chCol,
			disp7.ReadWrite, size*disp7.DefaultDigitWidth,
			disp7.DeafultDigitHeight, 1,
			disp7.DefaultVCursorSpace, "∆t   :", " s") //TODO change unit
		scp.channelViewers[ch].simGenDisplays = append(scp.channelViewers[ch].simGenDisplays, dwellTime)
		addToTest(dwellTime, genDwellTimeId)
		if err != nil {
			return
		}
		dwellTime.OnChanged = dwellTimeChanged
		dwellTime.SilentSetValue(int(genSettings.Dwelltime * 10000000))

		startFrqDisp, err = disp7.NewCustomDisp7Array(disp7Width, fractionWidth,
			int(genericps.SineMaxFrequency)*100,
			int(genericps.MinFrequency)*pow10tab[fractionWidth],
			disp7.UnSigned, disp7.NoTrailingZeroes, scp.Window,
			chCol,
			disp7.ReadWrite, size*disp7.DefaultDigitWidth,
			disp7.DeafultDigitHeight, 1,
			disp7.DefaultVCursorSpace, "Low :", " Hz")
		scp.channelViewers[ch].simGenDisplays = append(scp.channelViewers[ch].simGenDisplays, startFrqDisp)
		if err != nil {
			return
		}
		startFrqDisp.OnChanged = startFreqChanged
		startFrqDisp.SilentSetValue(int(genSettings.StartFrequency) * pow10tab[fractionWidth])
		addToTest(startFrqDisp, genMinFrqId)
		stopFrqDisp, err = disp7.NewCustomDisp7Array(disp7Width, fractionWidth,
			int(genericps.SineMaxFrequency)*100, int(genericps.MinFrequency)*pow10tab[fractionWidth],
			disp7.UnSigned, disp7.NoTrailingZeroes, scp.Window,
			chCol,
			disp7.ReadWrite, size*disp7.DefaultDigitWidth,
			disp7.DeafultDigitHeight, 1,
			disp7.DefaultVCursorSpace, "High:", " Hz")
		scp.channelViewers[ch].simGenDisplays = append(scp.channelViewers[ch].simGenDisplays, stopFrqDisp)
		if err != nil {
			return
		}
		addToTest(stopFrqDisp, genMaxFrqId)
		stopFrqDisp.OnChanged = stopFreqChanged
		stopFrqDisp.SilentSetValue(int(genSettings.StopFrequency) * pow10tab[fractionWidth])
		stepFreq, err = disp7.NewCustomDisp7Array(disp7Width, fractionWidth,
			int(genericps.SineMaxFrequency)*100, 0,
			disp7.UnSigned, disp7.NoTrailingZeroes, scp.Window,
			chCol,
			disp7.ReadWrite, size*disp7.DefaultDigitWidth,
			disp7.DeafultDigitHeight, 1,
			disp7.DefaultVCursorSpace, "Step :", " Hz")
		scp.channelViewers[ch].simGenDisplays = append(scp.channelViewers[ch].simGenDisplays, stepFreq)
		if err != nil {
			return
		}
		addToTest(stepFreq, genStepFreqId)
		stepFreq.OnChanged = stepFreqChanged
		stepFreq.SilentSetValue(int(genSettings.Increment) * pow10tab[fractionWidth])
		//TODO  arbitrary waveform
		amp, err = disp7.NewCustomDisp7Array(7, 6, maxV, 0,
			disp7.SignedHidden, disp7.NoTrailingZeroes, scp.Window,
			chCol,
			disp7.ReadWrite, size*disp7.DefaultDigitWidth,
			disp7.DeafultDigitHeight, 1,
			disp7.DefaultVCursorSpace, "Amplitude:", " V")
		scp.channelViewers[ch].simGenDisplays = append(scp.channelViewers[ch].simGenDisplays, amp)
		if err != nil {
			return
		}
		amp.SilentSetValue(int(genSettings.Amplitude))
		addToTest(amp, genAmpId)
		amp.OnChanged = ampChanged
		offset, err = disp7.NewCustomDisp7Array(7, 6,
			maxV, -maxV,
			disp7.Signed, disp7.NoTrailingZeroes, scp.Window,
			chCol,
			disp7.ReadWrite, size*disp7.DefaultDigitWidth,
			disp7.DeafultDigitHeight, 1,
			disp7.DefaultVCursorSpace, "Offset   :", " V")
		scp.channelViewers[ch].simGenDisplays = append(scp.channelViewers[ch].simGenDisplays, offset)
		if err != nil {
			return
		}
		addToTest(offset, genOffsetId)
		offset.OnChanged = offsetChanged
		setOffsetMinMax()
		offset.SilentSetValue(int(genSettings.OffsetVoltage))

		fLabel := widget.NewLabel("Freq")
		voltLabel := widget.NewLabel("Amp")
		shortLineF := container.NewBorder(nil, nil, nil, fLabel, freqSetAnalog)
		shortLineA := container.NewBorder(nil, nil, nil, voltLabel, ampSetAnalog)
		analog = container.New(layout.NewVBoxLayout(), shortLineF, shortLineA)
		sweepBox = container.New(layout.NewVBoxLayout(), startFrqDisp, stopFrqDisp,
			stepFreq, dwellTime)
		frqBox = container.New(layout.NewVBoxLayout(), frequency)
		sweepMenu := selectscroll.NewSelectScroll(sweepOptions, sweepChanged, sweepDownUp)
		sweepMenu.SetSelected(sweepOptions[genSettings.Sweep+1])
		if undockable {
			top = container.New(layout.NewHBoxLayout(), nameLabel, show, check,
				container.New(layout.NewVBoxLayout(), waveType), undockButton)
		} else {
			top = container.New(layout.NewHBoxLayout(), nameLabel, show, check,
				container.New(layout.NewVBoxLayout(), waveType))
		}

		addToTest(sweepMenu, genSweepId)
		sweepMenuBox := container.New(layout.NewHBoxLayout(),
			widget.NewLabel("Sweep "), sweepMenu)

		// if scp.psControl.Con.ID == genericps.SimId {
		scp.triggerCalculationModeSelect = selectscroll.NewSelectScroll(triggerCalculationOptions, onTriggerCalculationModeChange, triggerCalculationOptions[0])
		addToTest(scp.triggerCalculationModeSelect, triggerCalculationModeSelectId)
		scp.triggerCalculationModeSelect.SetSelected(triggerCalculationOptions[scp.Settings.Trigger.CalculationMode])
		// initialize simulator with saved setting
		sim.SetTriggerCalculationMode(scp.Settings.Trigger.CalculationMode)

		label := widget.NewLabel("Trigger Calc:")
		calcBox := container.New(layout.NewHBoxLayout(), label, scp.triggerCalculationModeSelect)

		digital = container.New(layout.NewVBoxLayout(), sweepMenuBox, frqBox,
			sweepBox, amp, offset, phaseDisp, raiseFallTimeDisp /*triggerTimeOffsetDisp,*/, noiseAmplitudeDisp, phaseNoiseDisp, calcBox)
		// } else {
		// 	digital = container.New(layout.NewVBoxLayout(), sweepMenuBox, frqBox,
		// 		sweepBox, amp, offset)
		// }

		box = container.New(layout.NewVBoxLayout(), top, analog, digital)
		show.SetChecked(genSettings.Digital)
		showChanged(genSettings.Digital)
		return
	} //newGenSettings

	sortWaveTypes()

	if undockable {
		undockButton := widget.NewButtonWithIcon(undock, theme.ViewFullScreenIcon(), func() {
			onWindowClose := func() {
				scp.genWindow.Hide()
				scp.genLayout.RemoveAll()
				scp.newSimGenPanel(scp.genLayout, true)

				scp.genTab = container.NewTabItem(tabNames[genTabIndex], scp.genLayout)
				scp.dockTab(scp.genTab)
				scp.controlTab.SelectIndex(ftTabIndex)
				fyne.Do(scp.genTab.Content.Refresh)
			}
			scp.genWindow = scp.App.NewWindow("Simulator Generator")
			windowLayout := container.New(layout.NewVBoxLayout())
			err = scp.newSimGenPanel(windowLayout, false)
			if err != nil {
				slog.Error("newSimGenPanel error", "err", err)
				return
			}
			scp.controlTab.Remove(scp.genTab)
			scp.genWindow.SetContent(windowLayout)
			scp.genWindow.SetOnClosed(onWindowClose)
			scp.genWindow.Show()
		})
		cont.Add(undockButton)
	}

	tabs := container.NewAppTabs()
	for i := 0; i < int(scp.channelCount); i++ {
		chId := genericps.ChannelId(i)
		genPanel, err := newGenSettings(chId, false)
		if err != nil {
			return err
		}
		chName := channelNames[i]
		chCol := scp.Settings.Channels[i].Col[scp.Settings.ChannelColorIndex]
		tabItem := container.NewTabItem("Ch "+chName, genPanel)
		tabItem.Icon = coloredCircleResource(colorToHex(chCol))
		tabs.Append(tabItem)
	}

	if scp.Settings.Window.SimGenActiveTab >= 0 && scp.Settings.Window.SimGenActiveTab < len(tabs.Items) {
		tabs.SelectIndex(scp.Settings.Window.SimGenActiveTab)
	}
	tabs.OnSelected = func(item *container.TabItem) {
		scp.Settings.Window.SimGenActiveTab = tabs.SelectedIndex()
		scp.SaveSettings()
	}

	cont.Add(tabs)
	return
}

func colorToHex(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02x%02x%02x", uint8(r>>8), uint8(g>>8), uint8(b>>8))
}

func coloredCircleResource(colorStr string) fyne.Resource {
	svg := fmt.Sprintf(`<svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><circle cx="12" cy="12" r="10" fill="%s" /></svg>`, colorStr)
	return fyne.NewStaticResource("color_"+colorStr, []byte(svg))
}
