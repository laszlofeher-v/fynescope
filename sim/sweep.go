package sim

import (
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
	sweepType        SweepTypeEnum
	initialSweepType SweepTypeEnum
	dwellTime        time.Duration
	lastStepTime     time.Time
}

// NewSweepController creates a new sweep controller with the specified parameters.
func NewSweepController(startFreq, stopFreq, stepFreq float64, sweepType SweepTypeEnum, dwellTime time.Duration) *SweepController {
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
		case SweepDown, SweepDownUp:
			sc.currentFreq = stopFreq
		case SweepUp, SweepUpDown:
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
	// No sweep if step frequency is zero
	// slog.Debug("sweep update 1")
	if sc.stepFreq == 0 {
		return
	}
	// slog.Debug("sweep update 2")

	// Check if dwell time has elapsed
	if time.Since(sc.lastStepTime) < sc.dwellTime {
		return
	}
	// slog.Debug("sweep update 3")

	// Update frequency based on sweep type
	switch sc.sweepType {
	case SweepUp:
		sc.updateSweepUp()
	case SweepUpDown:
		sc.updateSweepUpDown()
	case SweepDown:
		sc.updateSweepDown()
	case SweepDownUp:
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
	slog.Debug("UpDown", "currentFreq", sc.currentFreq, "stepFreq", sc.stepFreq, "stopFreq", sc.stopFreq)
	if sc.currentFreq+sc.stepFreq < sc.stopFreq {
		sc.currentFreq += sc.stepFreq
		sc.lastStepTime = time.Now()
	} else {
		sc.currentFreq = sc.stopFreq
		sc.lastStepTime = time.Now()
		sc.sweepType = SweepDownUp
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
	slog.Debug("DownUp", "currentFreq", sc.currentFreq, "stepFreq", sc.stepFreq, "stopFreq", sc.stopFreq)
	if sc.currentFreq-sc.stepFreq > sc.startFreq {
		sc.currentFreq -= sc.stepFreq
		sc.lastStepTime = time.Now()
	} else {
		sc.currentFreq = sc.startFreq
		sc.lastStepTime = time.Now()
		sc.sweepType = SweepUpDown
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
		case SweepDown, SweepDownUp:
			sc.currentFreq = sc.stopFreq
		case SweepUp, SweepUpDown:
			sc.currentFreq = sc.startFreq
		}
	}
}
