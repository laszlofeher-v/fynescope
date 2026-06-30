package gui

import (
	"math"
	"strconv"
)

// dftSyncRates / dftSyncUnits must match what newDftPanel creates.
var (
	dftSyncRates    = []string{"1", "2", "5", "10", "20", "50", "100", "200", "500"}
	dftSyncUnits    = []string{"S/s", "kS/s", "MS/s", "GS/s"}
	dftSyncUnitMuls = []float64{1.0, 1e3, 1e6, 1e9}
)

// syncDftToTimeDiv updates the DFT sample-rate widgets to match the current
// time/div setting using the formula:
//
//	fs = 1 / (timeDivSeconds / 10)  =  10 / timeDivSeconds
//
// Call after changing timeDiv or timeUnit.
func (scp *ScpDesc) syncDftToTimeDiv() {
	if scp.dftSampleRateSelect == nil || scp.dftSampleUnitSelect == nil {
		return
	}
	timeDivSeconds := float64(scp.timeDiv) * math.Pow(10, float64(scp.timeUnit))
	if timeDivSeconds <= 0 {
		return
	}
	fs := 10.0 / timeDivSeconds

	bestRateIdx := 0
	bestUnitIdx := 0
	bestDiff := math.MaxFloat64
	for ui, uMul := range dftSyncUnitMuls {
		for ri, rStr := range dftSyncRates {
			r, _ := strconv.ParseFloat(rStr, 64)
			val := r * uMul
			if val <= 0 {
				continue
			}
			relDiff := math.Abs(val-fs) / fs
			if relDiff < bestDiff {
				bestDiff = relDiff
				bestRateIdx = ri
				bestUnitIdx = ui
			}
		}
	}

	scp.dftSampleRateSelect.SilentSetSelectedIndex(bestRateIdx)
	scp.dftSampleUnitSelect.SilentSetSelectedIndex(bestUnitIdx)
	scp.Settings.Dft.SampleRate = dftSyncRates[bestRateIdx]
	scp.Settings.Dft.SampleRateUnit = dftSyncUnits[bestUnitIdx]
	scp.dftSampleRateSelect.Refresh()
	scp.dftSampleUnitSelect.Refresh()
}

// dftSettingsFs returns the sample rate in Hz from the current DFT settings.
func (scp *ScpDesc) dftSettingsFs() float64 {
	rate, _ := strconv.ParseFloat(scp.Settings.Dft.SampleRate, 64)
	unitMul := 1.0
	switch scp.Settings.Dft.SampleRateUnit {
	case "GS/s":
		unitMul = 1e9
	case "MS/s":
		unitMul = 1e6
	case "kS/s":
		unitMul = 1e3
	}
	return rate * unitMul
}

// syncTimeDivToDft updates the f(t) time/div widgets to match the current DFT
// sample-rate setting using the inverse formula:
//
//	timeDivSeconds = 10 / fs
//
// Call after changing the DFT sample rate or unit.
func (scp *ScpDesc) syncTimeDivToDft() {
	if scp.timeSelect == nil || scp.timeUnitSelect == nil {
		return
	}
	if units == nil {
		return
	}
	fs := scp.dftSettingsFs()
	if fs <= 0 {
		return
	}
	targetTimeDivSeconds := 10.0 / fs

	bestTimeStr := scp.timeSelect.Selected
	bestUnitStr := scp.timeUnitSelect.Selected
	bestDiff := math.MaxFloat64

	for _, unitStr := range units {
		unitPower := float64(tu[unitStr])
		uMul := math.Pow(10, unitPower)
		for _, timeStr := range times {
			tVal, _ := strconv.ParseFloat(timeStr, 64)
			val := tVal * uMul
			if val <= 0 {
				continue
			}
			relDiff := math.Abs(val-targetTimeDivSeconds) / targetTimeDivSeconds
			if relDiff < bestDiff {
				bestDiff = relDiff
				bestTimeStr = timeStr
				bestUnitStr = unitStr
			}
		}
	}

	prevTimeUnit := scp.timeUnit
	prevTime := scp.timeDiv

	scp.timeUnitSelect.SilentSetSelected(bestUnitStr)
	scp.timeUnit = tu[bestUnitStr]
	scp.Settings.Time.Unit = bestUnitStr
	scp.setTimeSelectOptions(bestUnitStr)
	scp.timeSelect.SilentSetSelected(bestTimeStr)
	intTimeDiv, _ := strconv.Atoi(bestTimeStr)
	scp.timeDiv = intTimeDiv
	scp.Settings.Time.TimeDiv = bestTimeStr
	// Update maxScreenTime so division lines are redrawn correctly.
	scp.maxScreenTime = float64(scp.timeDiv) * math.Pow(10, float64(scp.timeUnit)) * 10

	mul := math.Pow(10, float64(scp.timeUnit)) / math.Pow(10, float64(prevTimeUnit))
	if prevTime > 0 {
		mul *= float64(scp.timeDiv) / float64(prevTime)
	}
	scp.Settings.Time.TriggerTimeOffset *= mul
	scp.setTriggerTime(scp.Settings.Time.TriggerTimeOffset)

	scp.timeSelect.Refresh()
	scp.timeUnitSelect.Refresh()
}
