package sim

import (
	"math/rand"
)

const (
	InterpolatedTrigger = iota
	FineGrainedTrigger
)

// TriggerDetector handles trigger point detection for signal acquisition.
// It implements hysteresis-based triggering to avoid false triggers from noise.
type TriggerDetector struct {
	enabled                bool
	threshold              int16
	hysteresis             uint16
	direction              ThresholdDirection
	source                 ChannelId
	maxIterations          int
	triggerCalculationMode int
}

// NewTriggerDetector creates a new trigger detector with the specified parameters.
func NewTriggerDetector(enabled bool, threshold int16, hysteresis uint16, direction ThresholdDirection, source ChannelId) *TriggerDetector {
	return &TriggerDetector{
		enabled:       enabled,
		threshold:     threshold,
		hysteresis:    hysteresis,
		direction:     direction,
		source:        source,
		maxIterations: maxTriggerTest,
	}
}

var old float64

// FindTriggerPoint searches for a trigger point in the signal.
// It returns the time offset (in seconds) where the trigger condition is met.
// If no trigger is found within maxTime, it returns a random value.
//
// Parameters:
//   - signalFunc: Function that generates signal level at a given time (in seconds)
//   - reqSamples: Number of samples requested (used for random fallback)
//   - maxTime: Maximum time in seconds to search for a trigger
//   - dt: Time step in seconds for searching
//
// Returns:
//   - triggerTime: Time offset in seconds where trigger occurred
func (td *TriggerDetector) FindTriggerPoint(signalFunc func(t float64) float64,
	reqSamples uint32, maxTime float64, dt float64) (triggerTime float64) {
	// slog.Debug("FindTriggerPoint", "reqSamples", reqSamples, "timeIntervalNanoseconds", timeIntervalNanoseconds)
	// If trigger is disabled, return random trigger time
	if !td.enabled {
		triggerTime = rand.Float64() * float64(reqSamples)
		return
	}

	// Calculate threshold with hysteresis
	// For rising edge: hysteresis is below threshold
	// For falling edge: hysteresis is above threshold
	thresholdWithHysteresis := float64(td.threshold - int16(td.hysteresis))
	if td.direction == TriggerFalling {
		thresholdWithHysteresis = float64(td.threshold + int16(td.hysteresis))
	}

	// Track previous value for edge detection
	prev := 1e12
	if td.direction == TriggerFalling {
		prev = -prev
	}
	hysteresisPassed := false
	t := float64(0)
	threshold := float64(td.threshold)
	isRaising := td.direction == TriggerRaising
	// Search for trigger point
	for t < maxTime {
		// Get signal level at current time
		levelFloat := signalFunc(t)
		// Check if hysteresis condition is met
		if (isRaising && levelFloat <= thresholdWithHysteresis) ||
			(!isRaising && levelFloat >= thresholdWithHysteresis) {
			hysteresisPassed = true
		}
		// Check for trigger condition
		if hysteresisPassed {
			fineDt := dt / 1e3 // Sub-sample precision (1000 steps per sample)
			triggerTime = t - dt
			if (isRaising && levelFloat > threshold) ||
				(!isRaising && levelFloat < threshold) {
				switch td.triggerCalculationMode {
				case InterpolatedTrigger:
					v0 := signalFunc(t - dt)
					v1 := levelFloat
					a := (v1 - v0) / dt
					// slog.Debug("trg", "v0", v0, "v1", v1, "a", a)
					if a != 0 {
						SetTriggerTimeOffset(-(threshold - v0) / a)
					} else {
						SetTriggerTimeOffset(0)
					}
					return
				case FineGrainedTrigger:
					// TriggerTimeOffset = triggerTime + dt/2
					// slog.Debug("trg", "TriggerTimeOffset", TriggerTimeOffset, "triggerTime", triggerTime)
					// slog.Debug("trg", "TriggerTimeOffset (s)", TriggerTimeOffset, "triggerTime (s)", triggerTime)
					// Fine-grained search in the previous interval
					count := 0
					for tt := triggerTime; tt < t; tt += fineDt {
						levelFloat = signalFunc(tt)
						if (isRaising && levelFloat > threshold) ||
							(!isRaising && levelFloat < threshold) {
							SetTriggerTimeOffset(-(tt - triggerTime))
							if old != GetTriggerTimeOffset() {
								// slog.Debug("trg", "threshold", threshold, "levelFloat", levelFloat)
								// slog.Debug("trg", "TriggerTimeOffset", TriggerTimeOffset)
								// slog.Debug("trg", "t", t, "triggerTime", triggerTime)
								// slog.Debug("trg", "tt", tt, "fineDt", fineDt)
								old = GetTriggerTimeOffset()
							}
							return
						}
						count++
					}
				}
			}
		}
		prev = levelFloat
		t += dt
	}

	// No trigger found within max time, return random value
	// This simulates auto-trigger behavior
	triggerTime = rand.Float64() * float64(reqSamples) * dt
	return
}

// SetMaxIterations sets the maximum number of iterations to search for a trigger.
func (td *TriggerDetector) SetMaxIterations(iterations int) {
	td.maxIterations = iterations
}

// SetEnabled enables or disables the trigger.
func (td *TriggerDetector) SetEnabled(enabled bool) {
	td.enabled = enabled
}

// SetThreshold sets the trigger threshold level.
func (td *TriggerDetector) SetThreshold(threshold int16) {
	td.threshold = threshold
}

// SetHysteresis sets the trigger hysteresis value.
func (td *TriggerDetector) SetHysteresis(hysteresis uint16) {
	td.hysteresis = hysteresis
}

// SetDirection sets the trigger direction (rising/falling).
func (td *TriggerDetector) SetDirection(direction ThresholdDirection) {
	td.direction = direction
}

// SetSource sets the trigger source channel.
func (td *TriggerDetector) SetSource(source ChannelId) {
	td.source = source
}

// SetTriggerCalculationMode sets the trigger calculation mode.
func (td *TriggerDetector) SetTriggerCalculationMode(mode int) {
	td.triggerCalculationMode = mode
}

// GetSource returns the trigger source channel.
func (td *TriggerDetector) GetSource() ChannelId {
	return td.source
}
