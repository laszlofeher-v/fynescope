package gui

import (
	"fmt"
	"fynescope/control"
	"fynescope/genericps"
	"fynescope/selectscroll"
	"fynescope/settings"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

func (scp *ScpDesc) FunctionalTestUnit() error {
	log.Println("FunctionalTestUnit started")

	// Ensure clean initial state: stop acquisition if running, and disable digital trigger
	fyne.Do(func() {
		if scp.running {
			scp.StopRunning()
		}
		scp.Settings.Digital.Trigger.Enabled = false
		scp.updateDigitalTrigger()
	})
	wait()

	// 1. Select generator tab
	tap(genFuncId)
	wait()

	// 2. In demo mode select channel A
	fyne.Do(func() {
		scp.Settings.Window.DemoGenActiveTab = 0
		if len(scp.genLayout.Objects) > 0 {
			if tabs, ok := scp.genLayout.Objects[0].(*container.AppTabs); ok {
				tabs.SelectIndex(0)
			}
		}
	})
	wait()

	// 3. Set the generator to 1kHz sinusoid, 1V, normal mode, no offset, no sweep
	// 4. In demo mode: 0 phase, 0mv noise, 0 phase noise
	// 5. Switch the generator on
	genDone := make(chan struct{}, 1)
	fyne.Do(func() {
		genSettings := &scp.Settings.DemoGenPanel[genericps.ChA]
		genSettings.WaveType = genericps.Sine
		genSettings.Frequency = 1000
		genSettings.Amplitude = 1000000 // 1V amplitude (pk-to-pk 2V = 2000000 uV)
		genSettings.Sweep = genericps.NoSweep
		genSettings.OffsetVoltage = 0
		genSettings.Phase = 0
		genSettings.NoiseAmplitude = 0
		genSettings.PhaseNoiseDegree = 0
		genSettings.Operation = genericps.EsOff
		genSettings.On = true
		scp.applyDemoGenSettings(genericps.ChA, genSettings)

		if scp.psControl != nil && scp.psControl.SetDemoGenCh != nil {
			syncMsg := &control.GeneratorDescMsg{
				Done: genDone,
			}
			go func() { scp.psControl.SetDemoGenCh <- syncMsg }()
		} else {
			genDone <- struct{}{}
		}
	})
	<-genDone
	wait()

	// 6. Go to f(t)
	tap(ftFuncId)
	wait()

	// 7. Enable channel A, disable others
	// 8. Set 5V range, 0v offset
	fyne.Do(func() {
		scp.EnableChannel(genericps.ChA, true)
		for ch := genericps.ChB; int(ch) < int(scp.channelCount); ch++ {
			scp.EnableChannel(ch, false)
		}
		for i := 0; i < 2; i++ {
			scp.psControl.SetDigitalPortCh <- &control.DigitalPortMsg{
				Port:     genericps.Port0 + genericps.DigitalPort(i),
				Settings: settings.DigitalPortSettings{Enabled: false},
			}
		}

		scp.changeChannelRange(genericps.ChA, "±5V")
		scp.Settings.Channels[genericps.ChA].Offset = 0
	})
	wait()

	// 9. Set channel A trigger
	// 10. Set auto advanced trigger, 0V threshold, 0V hysteresis
	// 11. Move trigger point in the middle
	fyne.Do(func() {
		if int(genericps.ChA) < len(scp.triggerCheck) && scp.triggerCheck[genericps.ChA] != nil {
			scp.triggerCheck[genericps.ChA].SetChecked(true)
		}

		if scp.triggerModeSelect != nil {
			scp.triggerModeSelect.SetSelected(settings.TriggerModeAuto)
		}
		if scp.triggerTypeSelect != nil {
			scp.triggerTypeSelect.SetSelected(settings.TriggerTypeAdvanced)
		}
		
		scp.onThresholdChange(0)
		scp.onHysteresisChange(0)

		// Move trigger point in the middle
		scp.Settings.Time.TriggerTimeOffset = scp.maxScreenTime / 2.0
		scp.setTriggerTime(scp.Settings.Time.TriggerTimeOffset)
	})
	wait()

	// 12. Set normal sampling mode
	// 13. Set 1ms/div
	// 14. Set raw signal display mode
	fyne.Do(func() {
		if scp.psControl != nil {
			scp.psControl.StreamEnabled.Store(false)
		}

		if scp.resSelect != nil {
			scp.resSelect.SetSelected("Normal")
		}

		// Set 1ms/div
		if scp.timeUnitSelect != nil && scp.timeSelect != nil {
			scp.timeUnitSelect.SilentSetSelected("ms/div")
			scp.timeSelect.SilentSetSelected("1")
			scp.onTimeUnitChange("ms/div", selectscroll.None)
			scp.onTimeDivChange("1", selectscroll.None)
		}

		// Maintain trigger point in the middle of the screen
		scp.Settings.Time.TriggerTimeOffset = scp.maxScreenTime / 2.0
		scp.setTriggerTime(scp.Settings.Time.TriggerTimeOffset)

		// Set raw signal display mode
		if scp.ipmSelect != nil {
			scp.ipmSelect.SetSelected(raw)
		}
	})
	wait()

	// 15. Start acquisition
	if !scp.running {
		tap(runblockButtonId)
		for !scp.running {
			wait()
		}
	}

	// Wait and poll for acquisitions and raster render
	var minVal, maxVal float32 = 1e9, -1e9
	var bufLen int
	for i := 0; i < 50; i++ {
		wait()
		scp.screenLocker.Lock()
		buf := scp.displayBuffers[genericps.ChA]
		bufLen = len(buf)
		if bufLen > 0 {
			minVal, maxVal = 1e9, -1e9
			for _, v := range buf {
				if v < minVal {
					minVal = v
				}
				if v > maxVal {
					maxVal = v
				}
			}
			if maxVal >= 500 && minVal <= -500 {
				scp.screenLocker.Unlock()
				break
			}
		}
		scp.screenLocker.Unlock()
	}

	// 16. Check the displayed signal on raster
	var testErr error
	if bufLen == 0 {
		testErr = fmt.Errorf("expected display buffer for Channel A to have data, got 0 samples")
	} else if maxVal < 500 || minVal > -500 {
		// 1V amplitude sinusoid with 0V offset has expected peaks at ~ +1000mV and -1000mV
		testErr = fmt.Errorf("signal amplitude mismatch on Channel A: min=%.1fmV, max=%.1fmV (expected ~ +/-1000mV)", minVal, maxVal)
	}

	// 17. Stop the acquisition
	fyne.Do(func() {
		if scp.running {
			scp.StopRunning()
		}
	})
	wait()

	if testErr != nil {
		return testErr
	}

	log.Println("FunctionalTestUnit passed successfully")
	return nil
}
