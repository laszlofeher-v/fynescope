package gui

import (
	"fynescope/control"
	"fynescope/selectscroll"
	"math"
	"strconv"
)

func (scp *ScpDesc) updateAcquisitionParameters() {
	if scp.psControl == nil {
		return
	}

	effectiveFunction := scp.controlTab.SelectedIndex()
	if scp.controlTab.Selected() == scp.genTab || scp.controlTab.Selected() == scp.filterTab || scp.controlTab.Selected() == scp.extgenTab || scp.controlTab.Selected() == scp.vchTab {
		effectiveFunction = scp.Settings.Window.LastDispFunction
	}

	switch effectiveFunction {
	case dftTabIndex:
		// DFT mode: use sample rate and bins
		rate, _ := strconv.ParseFloat(scp.Settings.Dft.SampleRate, 64)
		unitMul := 1.0
		switch scp.Settings.Dft.SampleRateUnit {
		case selectscroll.UnitGSps:
			unitMul = 1e9
		case selectscroll.UnitMSps:
			unitMul = 1e6
		case selectscroll.UnitKSps:
			unitMul = 1e3
		case selectscroll.UnitSps:
			unitMul = 1.0
		}
		fs := rate * unitMul
		if fs <= 0 {
			fs = 1e6
		}

		// For DFT, we want at least 2 * Bins samples to avoid heavy zero padding
		samples := float64(scp.Settings.Dft.Bins * 2)
		scp.maxScreenTime = samples / fs
		scp.psControl.SetMaxScreenTime(scp.maxScreenTime)
		scp.psControl.SetScopeScreenWidth(samples)
	case ffTabIndex:
		// f(f) Bode mode: automatically adapt the capture window so there are
		// ~20 full periods of the current signal in the buffer.  This gives the
		// single-bin DFT plenty of cycles for good amplitude/phase SNR while
		// keeping the sampling rate high enough to resolve the waveform.
		//
		// Target: maxScreenTime = 20 / freq
		// Fallback when no signal measured yet: use MinFreq (lowest expected).
		// Use the app-controlled target frequency for acquisition window sizing.
		// This ensures the capture window matches the frequency the sweep is
		// currently targeting, preventing the feedback loop that caused the
		// old sweep to stall at higher frequencies.
		refFreq := scp.currentFfFreq
		if refFreq <= 0 {
			refFreq = scp.measuredFfFreq
		}
		if refFreq <= 0 {
			refFreq = scp.Settings.Ff.MinFreq
		}
		if refFreq <= 0 {
			refFreq = 10.0 // absolute fallback
		}
		const targetCycles = 20.0
		targetScreenTime := targetCycles / refFreq

		// Only update the acquisition window if it changed significantly (> 5%).
		// Because frequency measurements have small amounts of noise, updating
		// on every buffer would cause endless restarts of the scope.
		needsUpdate := false
		if scp.maxScreenTime == 0 {
			needsUpdate = true
		} else {
			diffRatio := math.Abs(scp.maxScreenTime-targetScreenTime) / scp.maxScreenTime
			if diffRatio > 0.05 {
				needsUpdate = true
			}
		}

		if needsUpdate {
			scp.maxScreenTime = targetScreenTime
			scp.psControl.SetMaxScreenTime(scp.maxScreenTime)
		}

		// Use the f(f) raster pixel width if available, else fall back to f(t) width.
		if scp.ffScopeSignalScreen != nil {
			scp.psControl.SetScopeScreenWidth(float64(scp.ffScopeSignalScreen.Bounds().Dx() - 1))
		} else if scp.ftScopeSignalScreen != nil {
			scp.psControl.SetScopeScreenWidth(float64(scp.ftScopeSignalScreen.Bounds().Dx() - 1))
		}
	default:
		// Time domain mode: use time/div
		scp.maxScreenTime = float64(scp.timeDiv) * math.Pow(10, float64(scp.timeUnit)) * 10 // 10 divs

		reqScreenTime := scp.maxScreenTime
		sampleMultiplier := 1.0

		if scp.timeZoomWindow != nil && scp.timeZoomMaxScreenTime > reqScreenTime {
			reqScreenTime = scp.timeZoomMaxScreenTime
			sampleMultiplier = scp.timeZoomMaxScreenTime / scp.maxScreenTime
		}

		scp.psControl.SetMaxScreenTime(reqScreenTime)
		if scp.ftScopeSignalScreen != nil {
			scp.psControl.SetScopeScreenWidth(float64(scp.ftScopeSignalScreen.Bounds().Dx()-1) * sampleMultiplier)
		} else {
			// Estimate the expected F(t) signal screen width if it hasn't been drawn yet
			w := float32(scp.Settings.Window.Width)
			h := float32(scp.Settings.Window.Height)
			if w == 0 {
				w = 1024
			}
			if h == 0 {
				h = 768
			}
			leftMargin, rightMargin := scp.clipFtChRangeScrs(w, h)
			expectedDx := int(math.Round(float64(w-rightMargin))) - int(math.Round(float64(leftMargin)))
			scp.psControl.SetScopeScreenWidth(float64(expectedDx) * sampleMultiplier)
		}
	}
	scp.updateDftDataCollectionTime()
	scp.updateStreamButtonVisibility()
}

func (scp *ScpDesc) inStreamMode() bool {
	if scp.psControl == nil {
		return false
	}
	return scp.maxScreenTime >= control.StreamThreshold && scp.psControl.StreamEnabled.Load()
}

func (scp *ScpDesc) updateStreamButtonVisibility() {
	if scp.streamEnableButton == nil {
		return
	}
	if scp.maxScreenTime >= control.StreamThreshold {
		scp.streamEnableButton.Show()
	} else {
		scp.streamEnableButton.Hide()
	}
	if scp.triggerDisplays != nil {
		if scp.inStreamMode() {
			scp.triggerDisplays.Hide()
		} else {
			if scp.triggerSource != dontCare {
				scp.triggerDisplays.Show()
			} else {
				scp.triggerDisplays.Hide()
			}
		}
	}
	if scp.toolbar != nil {
		scp.toolbar.Refresh()
	}
}

func (scp *ScpDesc) updateStreamButtonState() {
	if scp.streamEnableButton == nil || scp.psControl == nil {
		return
	}
	if scp.psControl.StreamEnabled.Load() {
		scp.streamEnableButton.SetText(streamEnabledLabel)
	} else {
		scp.streamEnableButton.SetText(streamDisabledLabel)
	}
}
