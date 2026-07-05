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
	isComplex              bool
	channels               [4]TriggerChannelConfig
}

type TriggerChannelConfig struct {
	Enabled                  bool
	Threshold                int16
	Hysteresis               uint16
	ThresholdLower           int16
	ThresholdLowerHysteresis uint16
	ThresholdMode            ThresholdModeId
	Direction                ThresholdDirection
	Condition                TriggerState // CondTrue, CondFalse, CondDontCare
}

// NewTriggerDetector creates a new trigger detector with the specified parameters.
func NewTriggerDetector(enabled bool, threshold int16, hysteresis uint16, direction ThresholdDirection, source ChannelId) *TriggerDetector {
	td := &TriggerDetector{
		enabled:       enabled,
		threshold:     threshold,
		hysteresis:    hysteresis,
		direction:     direction,
		source:        source,
		maxIterations: maxTriggerTest,
	}
	// Setup simple trigger in channels array
	if enabled && source >= 0 && int(source) < len(td.channels) {
		td.channels[source] = TriggerChannelConfig{
			Enabled:    true,
			Threshold:  threshold,
			Hysteresis: hysteresis,
			Direction:  direction,
			Condition:  CondTrue,
		}
	}
	return td
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
func (td *TriggerDetector) FindTriggerPoint(signalFunc func(t float64, ch ChannelId) float64,
	reqSamples uint32, maxTime float64, dt float64) (triggerTime float64) {
	
	// Check if any channels are enabled
	anyEnabled := false
	for _, cfg := range td.channels {
		if cfg.Enabled {
			anyEnabled = true
			break
		}
	}

	if !anyEnabled {
		triggerTime = rand.Float64() * float64(reqSamples)
		return
	}

	state := [4]int{}
	
	t := float64(0)
	for t < maxTime {
		allConditionsMet := true
		var edgeTriggerTime float64 = t - dt // Default trigger time if no edge triggers are used

		for i, cfg := range td.channels {
			if !cfg.Enabled || cfg.Condition == CondDontCare {
				continue
			}
			
			level := signalFunc(t, ChannelId(i))
			
			if cfg.ThresholdMode == Window {
				threshUpper := float64(cfg.Threshold)
				threshLower := float64(cfg.ThresholdLower)
				upperHyst := float64(cfg.Hysteresis)
				lowerHyst := float64(cfg.ThresholdLowerHysteresis)
				
				isInside := level <= threshUpper && level >= threshLower
				
				if cfg.Direction == TriggerInside {
					conditionMet := isInside
					if (cfg.Condition == CondTrue && !conditionMet) || (cfg.Condition == CondFalse && conditionMet) {
						allConditionsMet = false
						break
					}
					continue
				} else if cfg.Direction == TriggerOutside {
					conditionMet := !isInside
					if (cfg.Condition == CondTrue && !conditionMet) || (cfg.Condition == CondFalse && conditionMet) {
						allConditionsMet = false
						break
					}
					continue
				}
				
				isDeepOutside := level > (threshUpper + upperHyst) || level < (threshLower - lowerHyst)
				isDeepInside := level <= (threshUpper - upperHyst) && level >= (threshLower + lowerHyst)
				
				conditionMet := false
				if cfg.Direction == TriggerEnter {
					if state[i] == 0 && isDeepOutside {
						state[i] = 1
					}
					if state[i] == 1 && isInside {
						conditionMet = true
						edgeTriggerTime = t - dt
					}
				} else if cfg.Direction == TriggerExit {
					if state[i] == 0 && isDeepInside {
						state[i] = 2
					}
					if state[i] == 2 && !isInside {
						conditionMet = true
						edgeTriggerTime = t - dt
					}
				} else if cfg.Direction == TriggerEnterOrExit {
					if state[i] == 0 {
						if isDeepOutside { state[i] = 1 }
						if isDeepInside { state[i] = 2 }
					}
					if state[i] == 1 && isInside {
						conditionMet = true
						edgeTriggerTime = t - dt
					}
					if state[i] == 2 && !isInside {
						conditionMet = true
						edgeTriggerTime = t - dt
					}
				}
				
				if (cfg.Condition == CondTrue && !conditionMet) || (cfg.Condition == CondFalse && conditionMet) {
					allConditionsMet = false
					break
				}
			} else {
				// Level mode
				thresh := float64(cfg.Threshold)
				hyst := float64(cfg.Hysteresis)
				
				// Level condition?
				if cfg.Direction == TriggerNone || cfg.Direction == TriggerRisingOrFalling {
					// Treat TriggerNone as simple level > thresh
					conditionMet := level >= thresh
					if (cfg.Condition == CondTrue && !conditionMet) || (cfg.Condition == CondFalse && conditionMet) {
						allConditionsMet = false
						break
					}
					continue
				}
				
				conditionMet := false
				if cfg.Direction == TriggerRaising {
					if state[i] == 0 && level <= (thresh - hyst) {
						state[i] = 1
					}
					if state[i] == 1 && level > thresh {
						conditionMet = true
						edgeTriggerTime = t - dt
					}
				} else if cfg.Direction == TriggerFalling {
					if state[i] == 0 && level >= (thresh + hyst) {
						state[i] = 2
					}
					if state[i] == 2 && level < thresh {
						conditionMet = true
						edgeTriggerTime = t - dt
					}
				}
				
				if (cfg.Condition == CondTrue && !conditionMet) || (cfg.Condition == CondFalse && conditionMet) {
					allConditionsMet = false
					break
				}
			}
		}

		if allConditionsMet {
			// Trigger found! We use the edgeTriggerTime (or current time if only levels were used)
			SetTriggerTimeOffset(0) // Simple boolean logic doesn't support sub-sample interpolation yet
			return edgeTriggerTime
		}

		t += dt
	}

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

// SetChannelProperties sets the multi-channel trigger properties.
func (td *TriggerDetector) SetChannelProperties(props []TriggerChannelProperties) {
	// First, clear existing enabled properties
	for i := range td.channels {
		td.channels[i].Enabled = false
	}
	for _, prop := range props {
		ch := int(prop.Channel)
		if ch >= 0 && ch < len(td.channels) {
			td.channels[ch].Enabled = true
			td.channels[ch].Threshold = prop.ThresholdUpper
			td.channels[ch].Hysteresis = prop.ThresholdUpperHysteresis
			td.channels[ch].ThresholdLower = prop.ThresholdLower
			td.channels[ch].ThresholdLowerHysteresis = prop.ThresholdLowerHysteresis
			td.channels[ch].ThresholdMode = prop.ThresholdMode
		}
	}
}

// SetChannelConditions sets the multi-channel trigger conditions.
func (td *TriggerDetector) SetChannelConditions(conds []TriggerConditions) {
	if len(conds) == 0 {
		return
	}
	cond := conds[0] // We only support one condition matrix block for now
	td.channels[ChA].Condition = cond.ChannelA
	td.channels[ChB].Condition = cond.ChannelB
	td.channels[ChC].Condition = cond.ChannelC
	td.channels[ChD].Condition = cond.ChannelD
}

// SetChannelDirections sets the multi-channel trigger directions.
func (td *TriggerDetector) SetChannelDirections(dirA, dirB, dirC, dirD ThresholdDirection) {
	td.channels[ChA].Direction = dirA
	td.channels[ChB].Direction = dirB
	td.channels[ChC].Direction = dirC
	td.channels[ChD].Direction = dirD
}

// SetTriggerCalculationMode sets the trigger calculation mode.
func (td *TriggerDetector) SetTriggerCalculationMode(mode int) {
	td.triggerCalculationMode = mode
}

// GetSource returns the trigger source channel.
func (td *TriggerDetector) GetSource() ChannelId {
	return td.source
}
