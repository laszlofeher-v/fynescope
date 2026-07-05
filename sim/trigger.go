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

type TriggerArmedState int

const (
	TriggerStateIdle TriggerArmedState = iota
	TriggerStateArmedRising
	TriggerStateArmedFalling
)

type ChannelTriggerState struct {
	LevelState TriggerArmedState
	UpperState TriggerArmedState
	LowerState TriggerArmedState
}

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

	states := [4]ChannelTriggerState{}

	t := float64(0)
	for t < maxTime {
		allConditionsMet := true
		var edgeTriggerTime float64 = t - dt // Default trigger time if no edge triggers are used

		for i, cfg := range td.channels {
			if !cfg.Enabled || cfg.Condition == CondDontCare {
				continue
			}

			level := signalFunc(t, ChannelId(i))
			conditionMet := false
			fired := false

			if cfg.ThresholdMode == Window {
				conditionMet, fired = td.evaluateWindowTrigger(cfg, &states[i], level, signalFunc, t, dt, ChannelId(i))
			} else {
				conditionMet, fired = td.evaluateLevelTrigger(cfg, &states[i].LevelState, level)
			}

			if fired {
				edgeTriggerTime = t - dt
			}

			if (cfg.Condition == CondTrue && !conditionMet) || (cfg.Condition == CondFalse && conditionMet) {
				allConditionsMet = false
				break
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

func (td *TriggerDetector) evaluateLevelTrigger(cfg TriggerChannelConfig, state *TriggerArmedState, level float64) (conditionMet bool, fired bool) {
	thresh := float64(cfg.Threshold)
	hyst := float64(cfg.Hysteresis)

	if cfg.Direction == TriggerNone {
		// Treat TriggerNone as simple level > thresh
		return level >= thresh, false
	}

	if cfg.Direction == TriggerRaising {
		if *state == TriggerStateIdle && level <= (thresh-hyst) {
			*state = TriggerStateArmedRising
		} else if *state == TriggerStateArmedRising && level > thresh {
			*state = TriggerStateIdle
			return true, true
		}
	} else if cfg.Direction == TriggerFalling {
		if *state == TriggerStateIdle && level >= (thresh+hyst) {
			*state = TriggerStateArmedFalling
		} else if *state == TriggerStateArmedFalling && level < thresh {
			*state = TriggerStateIdle
			return true, true
		}
	} else if cfg.Direction == TriggerRisingOrFalling {
		if *state == TriggerStateIdle {
			if level <= (thresh - hyst) {
				*state = TriggerStateArmedRising
			} else if level >= (thresh + hyst) {
				*state = TriggerStateArmedFalling
			}
		} else if *state == TriggerStateArmedRising && level > thresh {
			*state = TriggerStateIdle
			return true, true
		} else if *state == TriggerStateArmedFalling && level < thresh {
			*state = TriggerStateIdle
			return true, true
		}
	}

	return false, false
}

func (td *TriggerDetector) evaluateWindowTrigger(
	cfg TriggerChannelConfig,
	state *ChannelTriggerState,
	level float64,
	signalFunc func(t float64, ch ChannelId) float64,
	t float64,
	dt float64,
	ch ChannelId,
) (conditionMet bool, fired bool) {

	upperCfg := TriggerChannelConfig{
		Threshold:  cfg.Threshold,
		Hysteresis: cfg.Hysteresis,
	}
	lowerCfg := TriggerChannelConfig{
		Threshold:  cfg.ThresholdLower,
		Hysteresis: cfg.ThresholdLowerHysteresis,
	}

	if cfg.Direction == TriggerEnter {
		upperCfg.Direction = TriggerFalling
		lowerCfg.Direction = TriggerRaising
	} else if cfg.Direction == TriggerExit {
		upperCfg.Direction = TriggerRaising
		lowerCfg.Direction = TriggerFalling
	} else if cfg.Direction == TriggerEnterOrExit {
		upperCfg.Direction = TriggerRisingOrFalling
		lowerCfg.Direction = TriggerRisingOrFalling
	}

	_, upperFired := td.evaluateLevelTrigger(upperCfg, &state.UpperState, level)
	_, lowerFired := td.evaluateLevelTrigger(lowerCfg, &state.LowerState, level)

	if upperFired && lowerFired {
		// Jumped over the entire window in one step (full swing). Not a window trigger event.
		return false, false
	}

	if upperFired || lowerFired {
		// One of the boundaries was crossed. Look ahead to see if it's a runt pulse!
		// Runt pulse: The same trigger fires again BEFORE the other trigger fires.
		// Full swing: The OTHER trigger fires before this trigger fires again.
		tempUpperState := state.UpperState
		tempLowerState := state.LowerState

		t_future := t
		limit := 100000 // Safely cap lookahead
		for i := 0; i < limit; i++ {
			t_future += dt
			futureLevel := signalFunc(t_future, ch)

			_, futureUpperFired := td.evaluateLevelTrigger(upperCfg, &tempUpperState, futureLevel)
			_, futureLowerFired := td.evaluateLevelTrigger(lowerCfg, &tempLowerState, futureLevel)

			if upperFired {
				if futureLowerFired {
					return false, false // Other trigger fired first! Full swing, reject.
				}
				if futureUpperFired {
					return true, true // Same trigger fired again! Runt pulse, accept.
				}
			} else if lowerFired {
				if futureUpperFired {
					return false, false // Other trigger fired first! Full swing, reject.
				}
				if futureLowerFired {
					return true, true // Same trigger fired again! Runt pulse, accept.
				}
			}
		}

		// If we exhausted the lookahead without crossing either boundary, assume valid
		return true, true
	}

	return false, false
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
