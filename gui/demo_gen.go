package gui

import (
	"fmt"
	"fynescope/control"
	"fynescope/disp16"
	"fynescope/genericps"
	"image/color"
	"log/slog"
	"math"

	"fyne.io/fyne/v2/theme"

	"fynescope/demo"
	"fynescope/disp7"
	"fynescope/selectscroll"
	"fynescope/settings"
	"fynescope/sliderscroll"
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func (scp *ScpDesc) applyDemoGenSettings(ch genericps.ChannelId, genSettings *settings.GeneratorSettings) {
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
		msg.ArbitraryWaveform = genSettings.ArbitraryWaveform
	} else {
		msg.DwellTime = 0
		msg.OffsetVoltage = 0
		msg.PkToPK = 0
		msg.WaveType = genericps.DcVoltage
		msg.StopFrequency = genSettings.StopFrequency
		msg.SweepType = genericps.SweepDown // there is no "no sweep"
	}
	msg.Operation = genSettings.Operation
	msg.Shots = 0
	msg.Sweeps = 0
	msg.TriggerType = genericps.SigGenRising
	msg.TriggerSource = genericps.SigGenNone
	msg.ExtInThreshold = 0
	msg.Phase = genSettings.Phase
	msg.SpiDataValue = genSettings.SpiDataValue
	msg.I2cAddressValue = genSettings.I2cAddressValue

	if genSettings.On {
		demo.SetNoiseAmplitude(int(ch), genSettings.NoiseAmplitude)
		demo.SetPhaseNoiseDegree(int(ch), genSettings.PhaseNoiseDegree)
		demo.SetRaiseFallTimePercent(genSettings.RaiseFallTimePercent / 100.0)
	}

	if scp.psControl != nil && scp.psControl.SetDemoGenCh != nil {
		msgCopy := msg
		go func() { scp.psControl.SetDemoGenCh <- msgCopy }()
	}
	go scp.SaveSettings()
}

func (scp *ScpDesc) newDemoGenPanel(cont *fyne.Container, undockable bool) (err error) {

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
			triggerCalculationOptions[0]: demo.InterpolatedTrigger,
			triggerCalculationOptions[1]: demo.FineGrainedTrigger,
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
			"Arbitrary": genericps.Arbitrary,
			"SpiClock":  genericps.SpiClock,
			"SpiData":   genericps.SpiData,
			"I2cClock":  genericps.I2cClock,
			"I2cData":   genericps.I2cData,
		}
		var keyVal []keyValDesc
		for key, val := range waveTypeMap {
			keyVal = append(keyVal, keyValDesc{key, val})
		}
		sort.Slice(keyVal, func(i, j int) bool {
			return keyVal[i].val < keyVal[j].val
		})
		waveTypeOptions = make([]string, len(waveTypeMap))
		for i, kv := range keyVal {
			waveTypeOptions[i] = kv.key
		}
	}

	getWaveTypeString := func(wt genericps.WaveTypeEnum) string {
		for k, v := range waveTypeMap {
			if v == wt {
				return k
			}
		}
		return "Sine"
	}

	onTriggerCalculationModeChange := func(option string, ex selectscroll.Exception) {
		mode := triggerCalculationModes[option]
		scp.Settings.Trigger.CalculationMode = mode
		demo.SetTriggerCalculationMode(mode)
		scp.SaveSettings()
	}

	var newGenSettings func(ch genericps.ChannelId, undockable bool) (box *fyne.Container, err error)

	newGenSettings = func(ch genericps.ChannelId, undockable bool) (box *fyne.Container, err error) {
		genSettings := &scp.Settings.DemoGenPanel[ch]
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
			spiDataDisp, i2cAddressDisp                                      *disp16.HexArray
			undockButton                                                     *widget.Button
			nameLabel                                                        *canvas.Text
		)
		if reloadData == nil {
			reloadData = make(chan struct{})
		}
		chCol := scp.Settings.Channels[ch].Col[scp.Settings.ChannelColorIndex]
		scp.channelViewers[ch].demoGenDisplays = nil
		checked := func(c bool) {
			genSettings.On = c
			scp.applyDemoGenSettings(ch, genSettings)
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
		addToTest(show, genShowId, genTabIndex)
		check := widget.NewCheck("On", checked)
		check.Checked = genSettings.On

		addToTest(check, genCheckId, genTabIndex)
		dwellTimeChanged := func(v float64) {
			genSettings.Dwelltime = v / 10000000
			scp.applyDemoGenSettings(ch, genSettings)
		}

		// Initialize channel name label
		nameLabel = canvas.NewText(channelNames[ch], chCol)
		nameLabel.TextStyle.Bold = true
		scp.channelViewers[ch].demoGenNameLabel = nameLabel

		freqChanged := func(v float64) {
			genSettings.Frequency = v / 100
			scp.applyDemoGenSettings(ch, genSettings)
		}
		startFreqChanged := func(v float64) {
			genSettings.StartFrequency = v / 100
			scp.applyDemoGenSettings(ch, genSettings)
		}
		stopFreqChanged := func(v float64) {
			genSettings.StopFrequency = v / 100
			scp.applyDemoGenSettings(ch, genSettings)
		}
		stepFreqChanged := func(v float64) {
			genSettings.Increment = v / 100
			scp.applyDemoGenSettings(ch, genSettings)
		}
		setOffsetMinMax := func() {
			maxOffset := maxV - int(genSettings.Amplitude)
			minOffset := -maxV + int(genSettings.Amplitude)
			offset.SetMinMax(minOffset, maxOffset)
		}
		ampChanged := func(v float64) {
			genSettings.Amplitude = uint32(v)
			setOffsetMinMax()
			scp.applyDemoGenSettings(ch, genSettings)
		}
		offsetChanged := func(v float64) {
			genSettings.OffsetVoltage = int32(v)
			scp.applyDemoGenSettings(ch, genSettings)
		}
		demo.SetRaiseFallTimePercent(genSettings.RaiseFallTimePercent / 100.0)
		raiseFallTimeDisp, err = disp7.NewCustomDisp7Array(5, 2,
			10000, 0,
			disp7.UnSigned, disp7.NoTrailingZeroes, scp.Window,
			chCol,
			disp7.ReadWrite, size*disp7.DefaultDigitWidth,
			disp7.DeafultDigitHeight, 1,
			disp7.DefaultVCursorSpace, "Rise/Fall:", " %")
		scp.channelViewers[ch].demoGenDisplays = append(scp.channelViewers[ch].demoGenDisplays, raiseFallTimeDisp)
		addToTest(raiseFallTimeDisp, genRiseFallTimeId, genTabIndex)
		if err != nil {
			return
		}
		raiseFallTimeDisp.OnChanged = func(v float64) {
			genSettings.RaiseFallTimePercent = v / 100.0 // Value comes back in % * 100, e.g. 159 for 1.59%
			demo.SetRaiseFallTimePercent(v / 10000.0)    // 1.59% -> 0.0159
			scp.applyDemoGenSettings(ch, genSettings)
		}
		raiseFallTimeDisp.SilentSetValue(int(genSettings.RaiseFallTimePercent * 100))

		demo.SetNoiseAmplitude(int(ch), genSettings.NoiseAmplitude)
		noiseAmplitudeDisp, err = disp7.NewCustomDisp7Array(5, 0,
			10000, 0,
			disp7.UnSigned, disp7.NoTrailingZeroes, scp.Window,
			chCol,
			disp7.ReadWrite, size*disp7.DefaultDigitWidth,
			disp7.DeafultDigitHeight, 1,
			disp7.DefaultVCursorSpace, "Noise:", " mV")
		scp.channelViewers[ch].demoGenDisplays = append(scp.channelViewers[ch].demoGenDisplays, noiseAmplitudeDisp)
		addToTest(noiseAmplitudeDisp, genNoiseAmpId, genTabIndex)
		if err != nil {
			return
		}
		noiseAmplitudeDisp.OnChanged = func(v float64) {
			genSettings.NoiseAmplitude = v
			demo.SetNoiseAmplitude(int(ch), v)
			scp.applyDemoGenSettings(ch, genSettings)
		}
		noiseAmplitudeDisp.SilentSetValue(int(genSettings.NoiseAmplitude))
		demo.SetPhaseNoiseDegree(int(ch), genSettings.PhaseNoiseDegree)
		phaseNoiseDisp, err = disp7.NewCustomDisp7Array(5, 2,
			36000, 0,
			disp7.UnSigned, disp7.NoTrailingZeroes, scp.Window,
			chCol,
			disp7.ReadWrite, size*disp7.DefaultDigitWidth,
			disp7.DeafultDigitHeight, 1,
			disp7.DefaultVCursorSpace, "Phase Noise:", " °")
		scp.channelViewers[ch].demoGenDisplays = append(scp.channelViewers[ch].demoGenDisplays, phaseNoiseDisp)
		addToTest(phaseNoiseDisp, genPhaseNoiseId, genTabIndex)
		if err != nil {
			return
		}
		phaseNoiseDisp.OnChanged = func(v float64) {
			genSettings.PhaseNoiseDegree = v / 100.0
			slog.Debug("phaseNoiseDisp", "PhaseNoiseDegree", genSettings.PhaseNoiseDegree)
			demo.SetPhaseNoiseDegree(int(ch), v/100.0)
			scp.applyDemoGenSettings(ch, genSettings)
		}
		phaseNoiseDisp.SilentSetValue(int(math.Round(genSettings.PhaseNoiseDegree * 100)))
		phaseDisp, err = disp7.NewCustomDisp7Array(3, 0,
			360, 0,
			disp7.UnSigned, disp7.NoTrailingZeroes, scp.Window,
			chCol,
			disp7.ReadWrite, size*disp7.DefaultDigitWidth,
			disp7.DeafultDigitHeight, 1,
			disp7.DefaultVCursorSpace, "Phase      :", " °")
		scp.channelViewers[ch].demoGenDisplays = append(scp.channelViewers[ch].demoGenDisplays, phaseDisp)
		addToTest(phaseDisp, genPhaseId, genTabIndex)
		if err != nil {
			return
		}
		phaseDisp.OnChanged = func(v float64) {
			genSettings.Phase = v
			scp.applyDemoGenSettings(ch, genSettings)
		}
		phaseDisp.SilentSetValue(int(genSettings.Phase))

		waveTypeChanged := func(option string, e selectscroll.Exception) {
			genSettings.WaveType = waveTypeMap[option]
			switch genSettings.WaveType {
			case genericps.Square, genericps.RampUp, genericps.RampDown:
				raiseFallTimeDisp.Show()
			default:
				raiseFallTimeDisp.Hide()
			}
			if spiDataDisp != nil {
				if genSettings.WaveType == genericps.SpiData || genSettings.WaveType == genericps.I2cData {
					spiDataDisp.Show()
				} else {
					spiDataDisp.Hide()
				}
			}
			if i2cAddressDisp != nil {
				if genSettings.WaveType == genericps.I2cData {
					i2cAddressDisp.Show()
				} else {
					i2cAddressDisp.Hide()
				}
			}
			scp.applyDemoGenSettings(ch, genSettings)
		}

		const (
			operationNormal     = "Normal"
			operationPrbs       = "PRBS"
			operationWhiteNoise = "White Noise"
		)
		operationMap := map[string]genericps.ExtraOperations{
			operationNormal:     genericps.EsOff,
			operationPrbs:       genericps.Prbs,
			operationWhiteNoise: genericps.WhiteNoise,
		}
		operationOptions := []string{operationNormal, operationPrbs, operationWhiteNoise}
		operationChanged := func(option string, e selectscroll.Exception) {
			genSettings.Operation = operationMap[option]
			scp.applyDemoGenSettings(ch, genSettings)
			scp.SaveSettings()
		}
		sweepChanged := func(option string, e selectscroll.Exception) {
			if option == sweepOff {
				sweepBox.Hide()
				genSettings.Sweep = genericps.NoSweep
				scp.applyDemoGenSettings(ch, genSettings)
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
			scp.applyDemoGenSettings(ch, genSettings)
		}
		waveType := selectscroll.NewSelectScroll(waveTypeOptions, waveTypeChanged, getWaveTypeString(genericps.DcVoltage))
		waveType.SetSelected(getWaveTypeString(genSettings.WaveType))
		addToTest(waveType, genWaveTypeId, genTabIndex)
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
					raiseFallTimeDisp.SilentSetValue(int(genSettings.RaiseFallTimePercent * 100))
					spiDataDisp.SetValue(uint64(genSettings.SpiDataValue))
					i2cAddressDisp.SetValue(uint64(genSettings.I2cAddressValue))
					noiseAmplitudeDisp.SilentSetValue(int(genSettings.NoiseAmplitude))
					phaseNoiseDisp.SilentSetValue(int(math.Round(genSettings.PhaseNoiseDegree * 100)))
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
			})
			addToTest(undockButton, "demoGenUndockBtn", genTabIndex)
		}
		freqSetAnalog = sliderscroll.NewSliderScroll(genericps.MinFrequency, genericps.SineMaxFrequency)
		addToTest(freqSetAnalog, genFreqSetId, genTabIndex)
		freqSetAnalog.OnChanged = freqChanged
		ampSetAnalog = sliderscroll.NewSliderScroll(0, maxV)
		ampSetAnalog.SilentSetValue(float64(genSettings.Amplitude))
		addToTest(ampSetAnalog, genAmpdSetId, genTabIndex)
		ampSetAnalog.OnChanged = ampChanged
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
		scp.channelViewers[ch].demoGenDisplays = append(scp.channelViewers[ch].demoGenDisplays, frequency)
		if err != nil {
			return
		}
		frequency.OnChanged = freqChanged
		frequency.SilentSetValue(int(genSettings.Frequency) * pow10tab[fractionWidth])
		addToTest(frequency, genFreqId, genTabIndex)

		dwellTime, err = disp7.NewCustomDisp7Array(11, 7,
			int(genericps.MaxDwellTime)*1000, int(500),
			disp7.UnSigned, disp7.NoTrailingZeroes, scp.Window,
			chCol,
			disp7.ReadWrite, size*disp7.DefaultDigitWidth,
			disp7.DeafultDigitHeight, 1,
			disp7.DefaultVCursorSpace, "∆t   :", " s")
		scp.channelViewers[ch].demoGenDisplays = append(scp.channelViewers[ch].demoGenDisplays, dwellTime)
		addToTest(dwellTime, genDwellTimeId, genTabIndex)
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
		scp.channelViewers[ch].demoGenDisplays = append(scp.channelViewers[ch].demoGenDisplays, startFrqDisp)
		if err != nil {
			return
		}
		startFrqDisp.OnChanged = startFreqChanged
		startFrqDisp.SilentSetValue(int(genSettings.StartFrequency) * pow10tab[fractionWidth])
		addToTest(startFrqDisp, genMinFrqId, genTabIndex)
		stopFrqDisp, err = disp7.NewCustomDisp7Array(disp7Width, fractionWidth,
			int(genericps.SineMaxFrequency)*100, int(genericps.MinFrequency)*pow10tab[fractionWidth],
			disp7.UnSigned, disp7.NoTrailingZeroes, scp.Window,
			chCol,
			disp7.ReadWrite, size*disp7.DefaultDigitWidth,
			disp7.DeafultDigitHeight, 1,
			disp7.DefaultVCursorSpace, "High:", " Hz")
		scp.channelViewers[ch].demoGenDisplays = append(scp.channelViewers[ch].demoGenDisplays, stopFrqDisp)
		if err != nil {
			return
		}
		addToTest(stopFrqDisp, genMaxFrqId, genTabIndex)
		stopFrqDisp.OnChanged = stopFreqChanged
		stopFrqDisp.SilentSetValue(int(genSettings.StopFrequency) * pow10tab[fractionWidth])
		stepFreq, err = disp7.NewCustomDisp7Array(disp7Width, fractionWidth,
			int(genericps.SineMaxFrequency)*100, 0,
			disp7.UnSigned, disp7.NoTrailingZeroes, scp.Window,
			chCol,
			disp7.ReadWrite, size*disp7.DefaultDigitWidth,
			disp7.DeafultDigitHeight, 1,
			disp7.DefaultVCursorSpace, "Step :", " Hz")
		scp.channelViewers[ch].demoGenDisplays = append(scp.channelViewers[ch].demoGenDisplays, stepFreq)
		if err != nil {
			return
		}
		addToTest(stepFreq, genStepFreqId, genTabIndex)
		stepFreq.OnChanged = stepFreqChanged
		stepFreq.SilentSetValue(int(genSettings.Increment) * pow10tab[fractionWidth])
		amp, err = disp7.NewCustomDisp7Array(7, 6, maxV, 0,
			disp7.SignedHidden, disp7.NoTrailingZeroes, scp.Window,
			chCol,
			disp7.ReadWrite, size*disp7.DefaultDigitWidth,
			disp7.DeafultDigitHeight, 1,
			disp7.DefaultVCursorSpace, "Amplitude:", " V")
		scp.channelViewers[ch].demoGenDisplays = append(scp.channelViewers[ch].demoGenDisplays, amp)
		if err != nil {
			return
		}
		amp.SilentSetValue(int(genSettings.Amplitude))
		addToTest(amp, genAmpId, genTabIndex)
		amp.OnChanged = ampChanged
		offset, err = disp7.NewCustomDisp7Array(7, 6,
			maxV, -maxV,
			disp7.Signed, disp7.NoTrailingZeroes, scp.Window,
			chCol,
			disp7.ReadWrite, size*disp7.DefaultDigitWidth,
			disp7.DeafultDigitHeight, 1,
			disp7.DefaultVCursorSpace, "Offset   :", " V")
		scp.channelViewers[ch].demoGenDisplays = append(scp.channelViewers[ch].demoGenDisplays, offset)
		if err != nil {
			return
		}
		addToTest(offset, genOffsetId, genTabIndex)
		offset.OnChanged = offsetChanged
		setOffsetMinMax()
		offset.SilentSetValue(int(genSettings.OffsetVoltage))

		spiDataDisp, err = disp16.NewHexArray(4, scp.Window, chCol, false, "SPI/I2C Data:")
		if err != nil {
			return
		}
		spiDataDisp.OnChanged = func(v uint64) {
			genSettings.SpiDataValue = uint32(v)
			scp.applyDemoGenSettings(ch, genSettings)
		}
		spiDataDisp.SetValue(uint64(genSettings.SpiDataValue))
		if genSettings.WaveType != genericps.SpiData && genSettings.WaveType != genericps.I2cData {
			spiDataDisp.Hide()
		}

		i2cAddressDisp, err = disp16.NewHexArray(2, scp.Window, chCol, false, "I2C Addr:")
		if err != nil {
			return
		}
		i2cAddressDisp.OnChanged = func(v uint64) {
			genSettings.I2cAddressValue = uint32(v)
			scp.applyDemoGenSettings(ch, genSettings)
		}
		i2cAddressDisp.SetValue(uint64(genSettings.I2cAddressValue))
		if genSettings.WaveType != genericps.I2cData {
			i2cAddressDisp.Hide()
		}

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
		awgEditorBtn := widget.NewButton("Open Waveform Editor", func() {
			scp.showAwgEditor(func(wf []int16) {
				genSettings.ArbitraryWaveform = wf
				genSettings.WaveType = genericps.Arbitrary
				waveType.SetSelected("Arbitrary")
				scp.applyDemoGenSettings(ch, genSettings)
			})
		})
		addToTest(awgEditorBtn, "demoAwgEditorBtn", genTabIndex)

		if undockable {
			top = container.New(layout.NewHBoxLayout(), nameLabel, show, check,
				container.New(layout.NewVBoxLayout(), waveType), undockButton)
		} else {
			top = container.New(layout.NewHBoxLayout(), nameLabel, show, check,
				container.New(layout.NewVBoxLayout(), waveType))
		}

		addToTest(sweepMenu, genSweepId, genTabIndex)
		sweepMenuBox := container.New(layout.NewHBoxLayout(),
			widget.NewLabel("Sweep "), sweepMenu)

		operationSelect := selectscroll.NewSelectScroll(operationOptions, operationChanged, operationNormal)
		if genSettings.Operation == genericps.Prbs {
			operationSelect.SetSelected(operationPrbs)
		} else if genSettings.Operation == genericps.WhiteNoise {
			operationSelect.SetSelected(operationWhiteNoise)
		} else {
			operationSelect.SetSelected(operationNormal)
		}
		addToTest(operationSelect, genOperationId, genTabIndex)
		operationBox := container.New(layout.NewHBoxLayout(),
			widget.NewLabel("Operation:"), operationSelect)

		scp.triggerCalculationModeSelect = selectscroll.NewSelectScroll(triggerCalculationOptions, onTriggerCalculationModeChange, triggerCalculationOptions[0])
		addToTest(scp.triggerCalculationModeSelect, triggerCalculationModeSelectId, genTabIndex)
		scp.triggerCalculationModeSelect.SetSelected(triggerCalculationOptions[scp.Settings.Trigger.CalculationMode])
		// initialize simulator with saved setting
		demo.SetTriggerCalculationMode(scp.Settings.Trigger.CalculationMode)

		label := widget.NewLabel("Trigger Calc:")
		calcBox := container.New(layout.NewHBoxLayout(), label, scp.triggerCalculationModeSelect)

		digital = container.New(layout.NewVBoxLayout(), sweepMenuBox, frqBox,
			sweepBox, amp, offset, phaseDisp, raiseFallTimeDisp, spiDataDisp, i2cAddressDisp,
			noiseAmplitudeDisp, phaseNoiseDisp, operationBox, calcBox, awgEditorBtn)
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
				scp.newDemoGenPanel(scp.genLayout, true)

				scp.genTab = container.NewTabItem(tabNames[genTabIndex], scp.genLayout)
				scp.dockTab(scp.genTab)
				scp.controlTab.SelectIndex(ftTabIndex)
				fyne.Do(scp.genTab.Content.Refresh)
			}
			scp.genWindow = scp.App.NewWindow("Demo Generator")
			windowLayout := container.New(layout.NewVBoxLayout())
			err = scp.newDemoGenPanel(windowLayout, false)
			if err != nil {
				slog.Error("newDemoGenPanel error", "err", err)
				return
			}
			scp.controlTab.Remove(scp.genTab)
			scp.genWindow.SetContent(windowLayout)
			scp.genWindow.SetOnClosed(onWindowClose)
			scp.genWindow.Show()
		})
		cont.Add(undockButton)
		addToTest(undockButton, "demoGenUndockBtnMain", genTabIndex)
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

	if scp.Settings.Window.DemoGenActiveTab >= 0 && scp.Settings.Window.DemoGenActiveTab < len(tabs.Items) {
		tabs.SelectIndex(scp.Settings.Window.DemoGenActiveTab)
	}
	tabs.OnSelected = func(tab *container.TabItem) {
		scp.Settings.Window.DemoGenActiveTab = tabs.SelectedIndex()
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
