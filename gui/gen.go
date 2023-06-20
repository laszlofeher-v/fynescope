package gui

import (
	"fmt"
	"psgo/gui/inter"
	"psgo/ps2000a"

	// "psgo/ps2000a"
	// "sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	//	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

var ()

func (ps *PsDesc) genOff() {
	var (
		offsetVoltage  int32                   = 0
		pkToPK         uint32                  = 0 // microvolts = 0V
		waveType       ps2000a.WaveTypeEnum    = ps2000a.DcVoltage
		sweepType      ps2000a.SweepTypeEnum   = ps2000a.SweepUp
		operation      ps2000a.ExtraOperations = ps2000a.EsOff
		shots          uint32                  = 0
		sweeps         uint32                  = 0
		triggerType    ps2000a.SigGenTrigType
		triggerSource  ps2000a.SigGenTrigSource = ps2000a.SigGenNone
		extInThreshold int16
	)
	startFrequency := ps.GenFreq
	stopFrequency := startFrequency + 10000.0 // +100 kHz
	increment := ps.GenFreqStep
	dwellTime := ps.GenFreqDelay
	err := ps.SetSigGenBuiltInV2(offsetVoltage, pkToPK, waveType, startFrequency, stopFrequency, increment, dwellTime,
		sweepType, operation, shots, sweeps, triggerType, triggerSource, extInThreshold)
	if err != nil {
		fmt.Println("genOff", err)
	}
}

func (ps *PsDesc) genOn() {
	var (
		offsetVoltage  int32                   = 0
		pkToPK         uint32                  = 2000000 // microvolts
		waveType       ps2000a.WaveTypeEnum    = ps2000a.Sine
		sweepType      ps2000a.SweepTypeEnum   = ps2000a.SweepUp
		operation      ps2000a.ExtraOperations = ps2000a.EsOff
		shots          uint32                  = 0
		sweeps         uint32                  = 0
		triggerType    ps2000a.SigGenTrigType
		triggerSource  ps2000a.SigGenTrigSource = ps2000a.SigGenNone
		extInThreshold int16
	)
	startFrequency := ps.GenFreq
	stopFrequency := startFrequency
	increment := ps.GenFreqStep
	dwellTime := ps.GenFreqDelay
	fmt.Println(offsetVoltage, pkToPK, waveType, startFrequency, stopFrequency, increment, dwellTime,
		sweepType, operation, shots, sweeps, triggerType, triggerSource, extInThreshold)
	err := ps.SetSigGenBuiltInV2(offsetVoltage, pkToPK, waveType, startFrequency, stopFrequency, increment, dwellTime,
		sweepType, operation, shots, sweeps, triggerType, triggerSource, extInThreshold)
	if err != nil {
		fmt.Println("genOn", err)
	}
}

func (ps *PsDesc) newGenSettings() *fyne.Container {
	// changed := func(option string) {
	// 	fmt.Println("Changed Gen", option)
	// }
	checked := func(c bool) {
		fmt.Println("Checked Gen", c)
		if c {
			ps.genOn()
		} else {
			ps.genOff()
		}
	}
	check := widget.NewCheck(ps.Str(inter.Gen), checked)
	freqChanged := func(v float64) {
		fmt.Println("freqChanged Gen", v)
		ps.GenFreq = v
		if check.Checked {
			ps.genOn()
		} else {
			ps.genOff()
		}
	}
	// freqStepChanged := func(v float64) {
	// 	fmt.Println("freqStepChanged Gen", v)
	// 	ps.GenFreqStep = v
	// 	if check.Checked {
	// 		ps.genOn()
	// 	} else {
	// 		ps.genOff()
	// 	}
	// }
	// freqDelayChanged := func(v float64) {
	// 	fmt.Println("freqDelayChanged Gen", v)
	// 	ps.GenFreqDelay = v
	// 	if check.Checked {
	// 		ps.genOn()
	// 	} else {
	// 		ps.genOff()
	// 	}
	// }
	// inverted := func(c bool) {
	// 	fmt.Println("inverted ", chIndex, chName, c)
	// }

	// invert := widget.NewCheck("Invert", inverted)
	// inputRanges := sortMapString(Ranges)
	// vRange := newSelectScroll(inputRanges, changed, chName)
	// //	vRange := newSelectScroll([]string{"±10mV", "±20mV", "±20V"}, changed)
	// vRange.SetSelected("±20V")
	//	acdc := newSelectScroll([]string{"AC", "DC"}, changed, chName)
	// acdc := widget.NewRadioGroup([]string{"AC", "DC"}, func(value string) {
	// 	log.Println("Select set to", value)
	// })
	// acdc.SetSelected("AC")
	// ps.GenFreq
	// ps.GenFreqStep
	// ps.GenFreqDelay
	// ps.GenFreq = binding.NewFloat()
	// ps.GenFreqStep = binding.NewFloat()
	// ps.GenFreqDelay = binding.NewFloat()
	// ps.GenFreqStr = binding.FloatToString(ps.GenFreq)
	// ps.GenFreqStepStr = binding.FloatToString(ps.GenFreqStep)
	// ps.GenFreqDelayStr = binding.FloatToString(ps.GenFreqDelay)
	// freqSetEntry := widget.NewEntryWithData(ps.GenFreqStr)
	// freqStepEntry := widget.NewEntryWithData(ps.GenFreqStepStr)
	// freqDelayEntry := widget.NewEntryWithData(ps.GenFreqDelayStr)
	// freqSet := widget.NewSliderWithData(0.0, 1.0e6, ps.GenFreq)
	// freqStep := widget.NewSliderWithData(0.0, 1.0e4, ps.GenFreqStep)
	// freqDelay := widget.NewSliderWithData(0.1, 1.0e3, ps.GenFreqDelay)
	//	freqSetEntry := widget.NewEntry()
	// freqStepEntry := widget.NewEntry()
	// freqDelayEntry := widget.NewEntry()
	freqSet := widget.NewSlider(ps2000a.MinFrequency, ps2000a.SineMaxFrequency)
	freqSet.OnChanged = freqChanged
	// freqStep := widget.NewSlider(0.0, 1.0e4)
	// freqStep.OnChanged = freqStepChanged
	// freqDelay := widget.NewSlider(0.1, 1.0e3)
	// freqDelay.OnChanged = freqDelayChanged
	// box := container.New(layout.NewVBoxLayout(), check, freqSetEntry, freqSet, freqStepEntry, freqStep, freqDelayEntry, freqDelay)
	box := container.New(layout.NewVBoxLayout(), check, freqSet)
	return box
}

func (ps *PsDesc) setGen(container *fyne.Container) {
	container.Add(layout.NewSpacer())
	ps.genPanel = ps.newGenSettings()
	container.Add(ps.genPanel)
}
