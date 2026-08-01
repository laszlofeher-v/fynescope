//go:build demo

package ps2000a

import (
	"fynescope/genericps"
	"log/slog"
	"time"
)

// SweepController manages frequency sweep state and updates.
// It encapsulates the sweep logic for signal generator frequency control.
type SweepController struct {
	currentFreq      float64
	startFreq        float64
	stopFreq         float64
	stepFreq         float64
	sweepType        genericps.SweepTypeEnum
	initialSweepType genericps.SweepTypeEnum
	dwellTime        time.Duration
	lastStepTime     time.Time
}

// NewSweepController creates a new sweep controller with the specified parameters.
func NewSweepController(startFreq, stopFreq, stepFreq float64, sweepType genericps.SweepTypeEnum, dwellTime time.Duration) *SweepController {
	sc := &SweepController{
		startFreq:        startFreq,
		stopFreq:         stopFreq,
		stepFreq:         stepFreq,
		sweepType:        sweepType,
		initialSweepType: sweepType,
		dwellTime:        dwellTime,
		lastStepTime:     time.Now(),
	}

	// Initialize current frequency based on sweep type
	if stepFreq == 0 {
		sc.currentFreq = startFreq
	} else {
		switch sweepType {
		case genericps.SweepDown, genericps.SweepDownUp:
			sc.currentFreq = stopFreq
		case genericps.SweepUp, genericps.SweepUpDown:
			sc.currentFreq = startFreq
		}
	}

	return sc
}

// GetCurrentFrequency returns the current frequency value.
func (sc *SweepController) GetCurrentFrequency() float64 {
	return sc.currentFreq
}

// Update updates the sweep state if the dwell time has elapsed.
// This should be called periodically (e.g., after each sample acquisition).
func (sc *SweepController) Update() {
	if sc.stepFreq == 0 {
		return
	}

	// Check if dwell time has elapsed
	if time.Since(sc.lastStepTime) < sc.dwellTime {
		return
	}

	// Update frequency based on sweep type
	switch sc.sweepType {
	case genericps.SweepUp:
		sc.updateSweepUp()
	case genericps.SweepUpDown:
		sc.updateSweepUpDown()
	case genericps.SweepDown:
		sc.updateSweepDown()
	case genericps.SweepDownUp:
		sc.updateSweepDownUp()
	default:
		slog.Error("Bad sweepType", "sc.sweepType", sc.sweepType)
	}
}

// updateSweepUp handles sweep up logic.
func (sc *SweepController) updateSweepUp() {
	if sc.currentFreq+sc.stepFreq < sc.stopFreq {
		sc.currentFreq += sc.stepFreq
		sc.lastStepTime = time.Now()
	} else {
		sc.currentFreq = sc.stopFreq
		sc.lastStepTime = time.Now()
	}
}

// updateSweepUpDown handles sweep up-down logic.
// When reaching the stop frequency, it switches to sweep down.
func (sc *SweepController) updateSweepUpDown() {
	if sc.currentFreq+sc.stepFreq < sc.stopFreq {
		sc.currentFreq += sc.stepFreq
		sc.lastStepTime = time.Now()
	} else {
		sc.currentFreq = sc.stopFreq
		sc.lastStepTime = time.Now()
		sc.sweepType = genericps.SweepDownUp
	}
}

// updateSweepDown handles sweep down logic.
func (sc *SweepController) updateSweepDown() {
	if sc.currentFreq-sc.stepFreq > sc.startFreq {
		sc.currentFreq -= sc.stepFreq
		sc.lastStepTime = time.Now()
	} else {
		sc.currentFreq = sc.startFreq
		sc.lastStepTime = time.Now()
	}
}

// updateSweepDownUp handles sweep down-up logic.
// When reaching the start frequency, it switches to sweep up.
func (sc *SweepController) updateSweepDownUp() {
	if sc.currentFreq-sc.stepFreq > sc.startFreq {
		sc.currentFreq -= sc.stepFreq
		sc.lastStepTime = time.Now()
	} else {
		sc.currentFreq = sc.startFreq
		sc.lastStepTime = time.Now()
		sc.sweepType = genericps.SweepUpDown
	}
}

// Reset resets the sweep controller to its initial state.
func (sc *SweepController) Reset() {
	sc.lastStepTime = time.Now()
	sc.sweepType = sc.initialSweepType
	if sc.stepFreq == 0 {
		sc.currentFreq = sc.startFreq
	} else {
		switch sc.sweepType {
		case genericps.SweepDown, genericps.SweepDownUp:
			sc.currentFreq = sc.stopFreq
		case genericps.SweepUp, genericps.SweepUpDown:
			sc.currentFreq = sc.startFreq
		}
	}
}
