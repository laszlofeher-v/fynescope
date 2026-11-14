//go:build sim

package gui

import (
	"fynescope/control"
	"fynescope/demo"
	"fynescope/genericps"
	_ "fynescope/ps2000a"
	"fynescope/settings"
	"image/color"
	"math"
	"path/filepath"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

func countNonBgPixels(scp *ScpDesc) (total int, tracePixels int) {
	if scp.ftScopeSignalScreen == nil {
		return 0, 0
	}
	bg := scp.theme.Color(ColorNameSignalBackground, 0)
	bgR, bgG, bgB, bgA := bg.RGBA()
	b := scp.ftScopeSignalScreen.Bounds()
	chCol := scp.Settings.Channels[0].Col[scp.Settings.ChannelColorIndex]
	cR, cG, cB, _ := chCol.RGBA()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			p := scp.ftScopeSignalScreen.At(x, y)
			r, g, bl, a := p.RGBA()
			if r != bgR || g != bgG || bl != bgB || a != bgA {
				total++
				if absDiff(r, cR) < 4096 && absDiff(g, cG) < 4096 && absDiff(bl, cB) < 4096 {
					tracePixels++
				}
			}
		}
	}
	_ = color.RGBA{}
	return total, tracePixels
}

func absDiff(a, b uint32) uint32 {
	if a > b {
		return a - b
	}
	return b - a
}

func TestSimGuiRepro(t *testing.T) {
	runCase := func(t *testing.T, name string, openFn func() (*genericps.Connection, error), loadPath string, configureFn func(scp *ScpDesc)) {
		t.Run(name, func(t *testing.T) {
			app := test.NewApp()
			defer app.Quit()
			w := app.NewWindow("SimRepro")
			w.Resize(fyne.NewSize(1280, 800))

			con, err := openFn()
			if err != nil {
				t.Fatalf("open failed: %v", err)
			}

			var cfg *settings.PsSettings
			if loadPath != "" {
				cfg, err = settings.Load(loadPath)
				if err != nil {
					t.Fatalf("settings.Load(%q) failed: %v", loadPath, err)
				}
			} else {
				cfg = settings.NewDefaultSettings()
			}
			tmpSettings := filepath.Join(t.TempDir(), "settings.yaml")

			scp := &ScpDesc{
				App:    app,
				Window: w,
			}
			// Run Menu in UI thread/directly in Fyne test app
			if err := scp.Menu(con, cfg, tmpSettings); err != nil {
				t.Fatalf("Menu failed: %v", err)
			}
			defer func() {
				if scp.running {
					scp.StopRunning()
				}
				if scp.psControl != nil {
					scp.psControl.Shutdown()
				}
			}()

			// Render once to partition ftScopeSignalScreen and set scopeScreenWidth
			scp.ftRasterGenerator(900, 500)

			if configureFn != nil {
				configureFn(scp)
			}

			// Start running if not already running
			if !scp.running {
				scp.StartRunning()
			}

			deadline := time.Now().Add(3 * time.Second)
			var bufLen int
			var minV, maxV float32
			for time.Now().Before(deadline) {
				scp.screenLocker.Lock()
				if len(scp.displayBuffers) > 0 && len(scp.displayBuffers[0]) > 0 {
					buf := scp.displayBuffers[0]
					bufLen = len(buf)
					minV, maxV = buf[0], buf[0]
					for _, v := range buf[1:] {
						if v < minV {
							minV = v
						}
						if v > maxV {
							maxV = v
						}
					}
					if maxV != minV && math.Abs(float64(maxV-minV)) > 1 {
						scp.screenLocker.Unlock()
						break
					}
				}
				scp.screenLocker.Unlock()
				time.Sleep(50 * time.Millisecond)
			}

			scp.ftRasterGenerator(900, 500)
			nonBg, chPx := countNonBgPixels(scp)
			deltaT := 0.0
			if scp.maxScreenTime > 0 && scp.ftScopeSignalScreen != nil {
				wPx := float64(scp.ftScopeSignalScreen.Bounds().Dx() - 1)
				deltaT = (wPx / scp.maxScreenTime) * scp.controlSamplingTimeInterval
			}
			t.Logf("%s: runningMode=%v info=%q bufLen=%d min=%.2f max=%.2f trigStat=%q genOn=%v genFreq=%.2f genAmp=%d ch0Enabled=%v ch0Range=%v maxScreenTime=%g dt=%g xRoundErr=%g trigOff=%d deltaT=%g nonBgPx=%d tracePx=%d",
				name, scp.runningMode, scp.psControl.GetInfo(), bufLen, minV, maxV,
				scp.triggerStatusLabel.Text, scp.Settings.GenPanel.On, scp.Settings.GenPanel.Frequency, scp.Settings.GenPanel.Amplitude,
				scp.Settings.Channels[0].Enabled, scp.Settings.Channels[0].VRange,
				scp.maxScreenTime, scp.controlSamplingTimeInterval, scp.controlXRoundError, scp.controlTriggerTimeOffset, deltaT, nonBg, chPx)
			if bufLen == 0 {
				t.Errorf("%s: expected non-empty display buffer", name)
			}
			if scp.controlSamplingTimeInterval <= 0 {
				t.Errorf("%s: expected positive sampling interval, got %g", name, scp.controlSamplingTimeInterval)
			}
			if maxV <= minV {
				t.Errorf("%s: expected non-flat signal, got min=%.2f max=%.2f", name, minV, maxV)
			}
			if chPx == 0 {
				t.Errorf("%s: expected rendered trace pixels on ftScopeSignalScreen", name)
			}
		})
	}

	openSimFn := func(serial string) func() (*genericps.Connection, error) {
		return func() (*genericps.Connection, error) {
			demo.ScopeSimVariantInfo = serial
			con := genericps.NewConnection()
			var err error
			con.Handle, err = genericps.OpenUnit(con, "ps2000a", serial, 0)
			if err != nil {
				return nil, err
			}
			con.ID = "ps2000a"
			return con, nil
		}
	}

	// 1. Demo mode with default settings + turning generator A on
	runCase(t, "Demo_Default_GenOn", func() (*genericps.Connection, error) {
		demo.ScopeSimVariantInfo = ""
		con := genericps.NewConnection()
		var err error
		con.Handle, err = genericps.OpenDemo(con, genericps.DemoId)
		if err != nil {
			return nil, err
		}
		con.ID = genericps.DemoId
		return con, nil
	}, "", func(scp *ScpDesc) {
		scp.Settings.Channels[0].Enabled = true
		scp.Settings.DemoGenPanel[0].On = true
		scp.Settings.DemoGenPanel[0].Frequency = 1000
		scp.Settings.DemoGenPanel[0].Amplitude = 1000000
		scp.applyDemoGenSettings(genericps.ChA, &scp.Settings.DemoGenPanel[0])
		time.Sleep(50 * time.Millisecond)
	})

	// 2. Sim 2407B with default settings + GenPanel On
	runCase(t, "Sim_2407B_Default_GenOn", openSimFn("2407BSIM"), "", func(scp *ScpDesc) {
		scp.Settings.Channels[0].Enabled = true
		scp.Settings.GenPanel.On = true
		scp.Settings.GenPanel.Frequency = 1000
		scp.Settings.GenPanel.Amplitude = 1000000
		scp.applyInternalGenSettings(true)
		time.Sleep(50 * time.Millisecond)
	})

	// 3. Sim 2407B via switching to gen tab & tapping genCheck (without manually setting Amplitude!)
	runCase(t, "Sim_2407B_Default_TabSelect_GenCheck", openSimFn("2407BSIM"), "", func(scp *ScpDesc) {
		scp.controlTab.Select(scp.genTab)
		internalTap(genCheckId, false)
		time.Sleep(100 * time.Millisecond)
	})

	// 4. Sim 2407B with 1 MHz signal at 100 ns/div
	runCase(t, "Sim_2407B_1MHz_100nsDiv", openSimFn("2407BSIM"), "", func(scp *ScpDesc) {
		scp.Settings.Channels[0].Enabled = true
		scp.timeSelect.SetSelected("100")
		scp.timeUnitSelect.SetSelected(nanoSec + div)
		scp.Settings.GenPanel.On = true
		scp.Settings.GenPanel.Digital = false
		scp.Settings.GenPanel.Frequency = 1e6
		scp.Settings.GenPanel.Amplitude = 1000000
		scp.applyInternalGenSettings(false)
		time.Sleep(80 * time.Millisecond)
	})

	// 5. Sim 2207BMSO with 1 kHz signal at 200 ns/div + digital port & trigger enabled
	runCase(t, "Sim_2207BMSO_DigitalTrig_200nsDiv", openSimFn("2207BMSOSIM"), "", func(scp *ScpDesc) {
		scp.Settings.Channels[0].Enabled = true
		scp.timeSelect.SetSelected("200")
		scp.timeUnitSelect.SetSelected(nanoSec + div)
		scp.Settings.Digital.Ports[0].Enabled = true
		scp.Settings.Digital.Trigger.Enabled = true
		scp.Settings.GenPanel.On = true
		scp.Settings.GenPanel.Frequency = 1000
		scp.Settings.GenPanel.Amplitude = 1000000
		scp.applyInternalGenSettings(false)
		time.Sleep(80 * time.Millisecond)
	})

	// 6. Sim 2207B with 100 kHz signal at 5 ms/div
	runCase(t, "Sim_2207B_100kHz_5msDiv", openSimFn("2207BSIM"), "", func(scp *ScpDesc) {
		scp.Settings.Channels[0].Enabled = true
		scp.timeSelect.SetSelected("5")
		scp.timeUnitSelect.SetSelected(milliSec + div)
		scp.Settings.GenPanel.On = true
		scp.Settings.GenPanel.Frequency = 100000
		scp.Settings.GenPanel.Amplitude = 1000000
		scp.applyInternalGenSettings(false)
		time.Sleep(80 * time.Millisecond)
	})

	// 7. Sim 2407B: StartRunning FIRST, THEN switch to genTab, tap Digital/Analog toggle, tap genCheck ON, switch back to ftTab
	runCase(t, "Sim_2407B_StartThenToggleGen", openSimFn("2407BSIM"), "", func(scp *ScpDesc) {
		scp.StartRunning()
		time.Sleep(150 * time.Millisecond)
		scp.controlTab.Select(scp.genTab)
		internalTap(genShowId, false) // Digital -> Analog
		t.Logf("after toggle to Analog: genAmp=%d genDigital=%v", scp.Settings.GenPanel.Amplitude, scp.Settings.GenPanel.Digital)
		internalTap(genCheckId, false) // Turn On while running
		time.Sleep(150 * time.Millisecond)
		scp.controlTab.Select(scp.ftTab)
		time.Sleep(150 * time.Millisecond)
	})

	// 8. Sim 2407B: toggle genShowId multiple times / scroll amp
	runCase(t, "Sim_2407B_ToggleDigitalAnalogAndScroll", openSimFn("2407BSIM"), "", func(scp *ScpDesc) {
		scp.controlTab.Select(scp.genTab)
		internalTap(genCheckId, false)
		internalTap(genShowId, false) // Analog
		t.Logf("after genShowId(1): genAmp=%d genFreq=%g", scp.Settings.GenPanel.Amplitude, scp.Settings.GenPanel.Frequency)
		internalTap(genShowId, false) // Digital
		t.Logf("after genShowId(2): genAmp=%d genFreq=%g", scp.Settings.GenPanel.Amplitude, scp.Settings.GenPanel.Frequency)
		internalTap(genShowId, false) // Analog again
		t.Logf("after genShowId(3): genAmp=%d genFreq=%g", scp.Settings.GenPanel.Amplitude, scp.Settings.GenPanel.Frequency)
	})

	// 9. Sim 2407B Stream mode (500 ms/div)
	runCase(t, "Sim_2407B_StreamMode_500ms", openSimFn("2407BSIM"), "", func(scp *ScpDesc) {
		scp.Settings.Channels[0].Enabled = true
		scp.timeSelect.SetSelected("500")
		scp.timeUnitSelect.SetSelected(milliSec + div)
		trueVal := true
		scp.Settings.StreamEnabled = &trueVal
		scp.psControl.SetStreamEnabled(true)
		scp.updateAcquisitionParameters()
		scp.Settings.GenPanel.On = true
		scp.Settings.GenPanel.Frequency = 10
		scp.Settings.GenPanel.Amplitude = 1000000
		scp.applyInternalGenSettings(true)
		time.Sleep(50 * time.Millisecond)
	})

	// 10. Demo Stream mode (500 ms/div) for comparison
	runCase(t, "Demo_StreamMode_500ms", func() (*genericps.Connection, error) {
		demo.ScopeSimVariantInfo = ""
		con := genericps.NewConnection()
		var err error
		con.Handle, err = genericps.OpenDemo(con, genericps.DemoId)
		if err != nil {
			return nil, err
		}
		con.ID = genericps.DemoId
		return con, nil
	}, "", func(scp *ScpDesc) {
		scp.Settings.Channels[0].Enabled = true
		scp.timeSelect.SetSelected("500")
		scp.timeUnitSelect.SetSelected(milliSec + div)
		trueVal := true
		scp.Settings.StreamEnabled = &trueVal
		scp.psControl.SetStreamEnabled(true)
		scp.updateAcquisitionParameters()
		scp.Settings.DemoGenPanel[0].On = true
		scp.Settings.DemoGenPanel[0].Frequency = 10
		scp.Settings.DemoGenPanel[0].Amplitude = 1000000
		scp.applyDemoGenSettings(genericps.ChA, &scp.Settings.DemoGenPanel[0])
		time.Sleep(50 * time.Millisecond)
	})

	// 11. Race repro: 10 fresh starts of Sim 2407B with GenPanel.On = true already
	t.Run("Sim_2407B_StartRace_10x", func(t *testing.T) {
		failures := 0
		for iter := 0; iter < 10; iter++ {
			app := test.NewApp()
			w := app.NewWindow("Race")
			w.Resize(fyne.NewSize(1280, 800))
			con, err := openSimFn("2407BSIM")()
			if err != nil {
				t.Fatalf("open failed: %v", err)
			}
			cfg := settings.NewDefaultSettings()
			cfg.Channels[0].Enabled = true
			cfg.GenPanel.On = true
			cfg.GenPanel.Frequency = 1000
			cfg.GenPanel.Amplitude = 1000000
			scp := &ScpDesc{App: app, Window: w}
			if err := scp.Menu(con, cfg, filepath.Join(t.TempDir(), "settings.yaml")); err != nil {
				t.Fatalf("Menu failed: %v", err)
			}
			scp.ftRasterGenerator(900, 500)
			scp.StartRunning()
			deadline := time.Now().Add(800 * time.Millisecond)
			gotSignal := false
			for time.Now().Before(deadline) {
				scp.screenLocker.Lock()
				if len(scp.displayBuffers) > 0 && len(scp.displayBuffers[0]) > 0 {
					minV, maxV := scp.displayBuffers[0][0], scp.displayBuffers[0][0]
					for _, v := range scp.displayBuffers[0][1:] {
						if v < minV {
							minV = v
						}
						if v > maxV {
							maxV = v
						}
					}
					if maxV-minV > 100 {
						gotSignal = true
						scp.screenLocker.Unlock()
						break
					}
				}
				scp.screenLocker.Unlock()
				time.Sleep(25 * time.Millisecond)
			}
			if !gotSignal {
				failures++
				t.Logf("iter %d FAILED to get signal! dt=%g trigStat=%q", iter, scp.controlSamplingTimeInterval, scp.triggerStatusLabel.Text)
			}
			if scp.running {
				scp.StopRunning()
			}
			scp.psControl.Shutdown()
			app.Quit()
		}
		t.Logf("Sim_2407B_StartRace_10x: %d/10 failures", failures)
		if failures > 0 {
			t.Errorf("expected 0 startup failures across 10 runs, got %d", failures)
		}
	})

	// 12. ETS mode on Sim 2407B with generator ON (comparing ETS acquisition & display)
	runCase(t, "Sim_2407B_ETS_GenOn", openSimFn("2407BSIM"), "", func(scp *ScpDesc) {
		scp.Settings.Channels[0].Enabled = true
		scp.timeSelect.SetSelected("1")
		scp.timeUnitSelect.SetSelected(microSec + div)
		scp.triggerModeSelect.SetSelected("ETS")
		scp.Settings.GenPanel.On = true
		scp.Settings.GenPanel.Frequency = 100000
		scp.Settings.GenPanel.Amplitude = 1000000
		scp.applyInternalGenSettings(false)
		time.Sleep(80 * time.Millisecond)
	})

	// 13a. 2207BMSO with 500 us/div + 1 kHz + digital enabled & digital trigger enabled
	runCase(t, "Sim_2207BMSO_DigitalTrigEnabled_GenOn", openSimFn("2207BMSOSIM"), "", func(scp *ScpDesc) {
		scp.Settings.Channels[0].Enabled = true
		scp.timeSelect.SetSelected("500")
		scp.timeUnitSelect.SetSelected(microSec + div)
		scp.Settings.Digital.Ports[0].Enabled = true
		scp.Settings.Digital.Trigger.Enabled = true
		scp.Settings.GenPanel.On = true
		scp.Settings.GenPanel.Frequency = 1000
		scp.Settings.GenPanel.Amplitude = 1000000
		scp.applyInternalGenSettings(false)
		time.Sleep(80 * time.Millisecond)
	})

	// 13b. 2207BMSO with 500 us/div + 1 kHz + digital disabled
	runCase(t, "Sim_2207BMSO_DigitalDisabled_GenOn", openSimFn("2207BMSOSIM"), "", func(scp *ScpDesc) {
		scp.Settings.Channels[0].Enabled = true
		scp.timeSelect.SetSelected("500")
		scp.timeUnitSelect.SetSelected(microSec + div)
		scp.Settings.Digital.Ports[0].Enabled = false
		scp.Settings.Digital.Trigger.Enabled = false
		scp.Settings.GenPanel.On = true
		scp.Settings.GenPanel.Frequency = 1000
		scp.Settings.GenPanel.Amplitude = 1000000
		scp.applyInternalGenSettings(false)
		time.Sleep(80 * time.Millisecond)
	})

	// 14. Sim 2407B Gen ON -> OFF -> inspect signal -> ON -> inspect signal
	t.Run("Sim_2407B_GenOnOffOn", func(t *testing.T) {
		app := test.NewApp()
		defer app.Quit()
		w := app.NewWindow("GenOnOffOn")
		w.Resize(fyne.NewSize(1280, 800))
		con, err := openSimFn("2407BSIM")()
		if err != nil {
			t.Fatalf("open failed: %v", err)
		}
		cfg := settings.NewDefaultSettings()
		cfg.Channels[0].Enabled = true
		cfg.GenPanel.On = true
		cfg.GenPanel.Frequency = 1000
		cfg.GenPanel.Amplitude = 1000000
		scp := &ScpDesc{App: app, Window: w}
		if err := scp.Menu(con, cfg, filepath.Join(t.TempDir(), "settings.yaml")); err != nil {
			t.Fatalf("Menu failed: %v", err)
		}
		scp.ftRasterGenerator(900, 500)
		scp.StartRunning()
		defer func() {
			if scp.running {
				scp.StopRunning()
			}
			scp.psControl.Shutdown()
		}()
		time.Sleep(300 * time.Millisecond)
		ampRange := func() float64 {
			scp.screenLocker.Lock()
			defer scp.screenLocker.Unlock()
			if len(scp.displayBuffers) == 0 || len(scp.displayBuffers[0]) == 0 {
				return 0
			}
			minV, maxV := scp.displayBuffers[0][0], scp.displayBuffers[0][0]
			for _, v := range scp.displayBuffers[0][1:] {
				if v < minV {
					minV = v
				}
				if v > maxV {
					maxV = v
				}
			}
			return float64(maxV - minV)
		}
		on1Range := ampRange()
		// Turn generator OFF via checkbox
		scp.controlTab.Select(scp.genTab)
		internalTap(genCheckId, false)
		time.Sleep(350 * time.Millisecond)
		offRange := ampRange()
		// Turn generator back ON via checkbox
		internalTap(genCheckId, false)
		time.Sleep(350 * time.Millisecond)
		on2Range := ampRange()
		t.Logf("Sim_2407B_GenOnOffOn: on1Range=%.2f offRange=%.2f on2Range=%.2f genOn=%v", on1Range, offRange, on2Range, scp.Settings.GenPanel.On)
		if on1Range <= 500 {
			t.Errorf("expected non-zero initial generator signal, got %.2f mVpp", on1Range)
		}
		if offRange >= 50 {
			t.Errorf("expected near-zero signal after turning generator off, got %.2f mVpp", offRange)
		}
		if on2Range <= 500 {
			t.Errorf("expected generator signal after turning back on, got %.2f mVpp", on2Range)
		}
	})

	_ = control.Auto
}
