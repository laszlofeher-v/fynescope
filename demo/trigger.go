package demo

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
	pwqConfig              PwqConfig
}

type PwqConfig struct {
	Enabled   bool
	Direction ThresholdDirection
	Lower     uint32
	Upper     uint32
	Type      PulseWidthType
	Condition [4]TriggerState // CondTrue, CondFalse, CondDontCare
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
			Enabled:                  true,
			Threshold:                threshold,
			Hysteresis:               hysteresis,
			ThresholdLower:           threshold,
			ThresholdLowerHysteresis: hysteresis,
			Direction:                direction,
			Condition:                CondTrue,
		}
	}
	return td
}

func (td *TriggerDetector) SetPulseWidthQualifier(conditions []PwqConditions, direction ThresholdDirection, lower, upper uint32, pwType PulseWidthType) {
	td.pwqConfig.Enabled = pwType != PwTypeNone
	td.pwqConfig.Direction = direction
	td.pwqConfig.Lower = lower
	td.pwqConfig.Upper = upper
	td.pwqConfig.Type = pwType

	if len(conditions) > 0 {
		td.pwqConfig.Condition[0] = conditions[0].ChannelA
		td.pwqConfig.Condition[1] = conditions[0].ChannelB
		td.pwqConfig.Condition[2] = conditions[0].ChannelC
		td.pwqConfig.Condition[3] = conditions[0].ChannelD
	} else {
		for i := 0; i < 4; i++ {
			td.pwqConfig.Condition[i] = CondDontCare
		}
	}
}

var old float64

type TriggerArmedState int

const (
	TriggerStateIdle TriggerArmedState = iota
	TriggerStateArmedRising
	TriggerStateArmedFalling
)

type ChannelTriggerState struct {
	LevelState    TriggerArmedState
	UpperState    TriggerArmedState
	LowerState    TriggerArmedState
	RuntPending   bool
	RuntStartTime float64
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
	reqSamples uint32, maxTime float64, dt float64) (found bool, triggerTime float64) {
	
	DisablePhaseNoise(true)
	defer DisablePhaseNoise(false)

	// Check if any channels are enabled
	found = false
	anyEnabled := false
	for _, cfg := range td.channels {
		if cfg.Enabled {
			anyEnabled = true
			break
		}
	}

	if !anyEnabled {
		found = true
		triggerTime = rand.Float64() * float64(reqSamples)
		return
	}

	pwqStates := [4]ChannelTriggerState{}
	states := [4]ChannelTriggerState{}
	insideStates := [4]ChannelTriggerState{} // for isInsideWindow checks

	intervalActive := false
	var intervalDuration float64 = 0
	var intervalSourceCh int = -1

	t := float64(0)
	for t < maxTime {
		allConditionsMet := true
		var edgeTriggerTime float64 = t - dt // Default trigger time if no edge triggers are used

		var mainEdgeFired bool = false
		var pwqEdgeFired bool = false

		// 1. Evaluate PWQ triggers
		if td.pwqConfig.Enabled {
			for i, cond := range td.pwqConfig.Condition {
				if cond == CondDontCare {
					continue
				}
				level := signalFunc(t, ChannelId(i))

				// PWQ evaluates against the main channel's thresholds
				cfg := td.channels[i]

				// Determine start edge of the pulse
				pwqDir := TriggerRising
				if td.pwqConfig.Direction == TriggerFallingLower || td.pwqConfig.Direction == TriggerFalling {
					pwqDir = TriggerFalling
				} else if td.pwqConfig.Direction == TriggerRisingOrFalling {
					pwqDir = TriggerRisingOrFalling
				} else if td.pwqConfig.Direction == TriggerInside || td.pwqConfig.Direction == TriggerOutside || td.pwqConfig.Direction == TriggerEnter || td.pwqConfig.Direction == TriggerExit {
					pwqDir = td.pwqConfig.Direction
				}

				thresh := cfg.Threshold
				hyst := cfg.Hysteresis
				if td.pwqConfig.Direction == TriggerFallingLower || td.pwqConfig.Direction == TriggerRisingLower {
					thresh = cfg.ThresholdLower
					hyst = cfg.ThresholdLowerHysteresis
				}

				pwqCfg := TriggerChannelConfig{
					Threshold:                thresh,
					Hysteresis:               hyst,
					ThresholdLower:           cfg.ThresholdLower,
					ThresholdLowerHysteresis: cfg.ThresholdLowerHysteresis,
					Direction:                pwqDir,
				}

				var fired bool
				if cfg.ThresholdMode == Window {
					_, fired, _ = td.evaluateWindowTrigger(pwqCfg, &pwqStates[i], level, signalFunc, t, dt, ChannelId(i))
				} else {
					_, fired = td.evaluateLevelTrigger(pwqCfg, &pwqStates[i].LowerState, level)
				}
				if fired {
					pwqEdgeFired = true
					// pwqEdgeTimeOffset not strictly needed unless PWQ relies on edgeTriggerTime,
					// but PWQ usually doesn't dictate the main trigger time directly.
				}
			}
		}

		// 2. Evaluate main channel triggers
		for i, cfg := range td.channels {
			if !cfg.Enabled || cfg.Condition == CondDontCare {
				continue
			}

			level := signalFunc(t, ChannelId(i))
			conditionMet := false
			fired := false

			var timeOffset float64

			if cfg.ThresholdMode == Window {
				if td.pwqConfig.Enabled && td.pwqConfig.Condition[i] != CondDontCare {
					// In Window PW mode, the main trigger fires on the EXIT edge (opposite of PWQ).
					exitCfg := cfg
					switch cfg.Direction {
					case TriggerEnter, TriggerInside, TriggerAbove, TriggerRising:
						exitCfg.Direction = TriggerExit
					case TriggerExit, TriggerOutside, TriggerBelow, TriggerFalling:
						exitCfg.Direction = TriggerEnter
					}
					conditionMet, fired, timeOffset = td.evaluateWindowTrigger(exitCfg, &states[i], level, signalFunc, t, dt, ChannelId(i))
				} else {
					conditionMet, fired, timeOffset = td.evaluateWindowTrigger(cfg, &states[i], level, signalFunc, t, dt, ChannelId(i))
				}
				// Also maintain inside-window state for GreaterThan checks.
				td.updateInsideWindow(&insideStates[i], cfg, level)
			} else {
				conditionMet, fired = td.evaluateLevelTrigger(cfg, &states[i].LevelState, level)
			}

			if fired {
				edgeTriggerTime = t - dt + timeOffset
				if td.pwqConfig.Condition[i] != CondDontCare {
					mainEdgeFired = true
				}
			}

			if (cfg.Condition == CondTrue && !conditionMet) || (cfg.Condition == CondFalse && conditionMet) {
				allConditionsMet = false
				break
			}
		}

		// 3. Track Interval and Trigger
		if td.pwqConfig.Enabled {
			if intervalActive {
				intervalDuration += dt

				if td.pwqConfig.Type == PwTypeGreaterThan {
					// For Window PW GreaterThan: fire immediately when timer expires while
					// the signal is still inside the window. For regular PW GreaterThan
					// this is handled in the mainEdgeFired path below.
					if intervalSourceCh >= 0 && td.channels[intervalSourceCh].ThresholdMode == Window {
						lowerTime := float64(td.pwqConfig.Lower) * dt
						if intervalDuration >= lowerTime {
							isInside := insideStates[intervalSourceCh].UpperState == TriggerStateArmedFalling &&
								insideStates[intervalSourceCh].LowerState == TriggerStateArmedRising
							if isInside {
								SetTriggerTimeOffset(0)
								found = true
								triggerTime = t
								return
							}
						}
					}
				}
			}

			if mainEdgeFired && intervalActive {
				lowerTime := float64(td.pwqConfig.Lower) * dt
				upperTime := float64(td.pwqConfig.Upper) * dt

				intervalSatisfied := false
				switch td.pwqConfig.Type {
				case PwTypeLessThan:
					if intervalDuration < lowerTime {
						intervalSatisfied = true
					}
				case PwTypeGreaterThan:
					if intervalDuration > lowerTime {
						intervalSatisfied = true
					}
				case PwTypeInRange:
					if intervalDuration >= lowerTime && intervalDuration <= upperTime {
						intervalSatisfied = true
					}
				case PwTypeOutOfRange:
					if intervalDuration < lowerTime || intervalDuration > upperTime {
						intervalSatisfied = true
					}
				}

				if intervalSatisfied && allConditionsMet {
					SetTriggerTimeOffset(0) // Simple boolean logic doesn't support sub-sample interpolation yet
					found = true
					triggerTime = edgeTriggerTime
					return
				}
				intervalActive = false
			}

			if pwqEdgeFired {
				intervalActive = true
				intervalDuration = 0
				// Record which channel started the interval for Window PW GreaterThan checks.
				for i, cond := range td.pwqConfig.Condition {
					if cond != CondDontCare {
						intervalSourceCh = i
						break
					}
				}
			}
		} else {
			if allConditionsMet {
				SetTriggerTimeOffset(0)
				found = true
				triggerTime = edgeTriggerTime
				return
			}
		}

		t += dt
	}

	triggerTime = rand.Float64() * float64(reqSamples) * dt
	if !td.pwqConfig.Enabled {
		found = true
	}
	return
}

func (td *TriggerDetector) evaluateLevelTrigger(cfg TriggerChannelConfig, state *TriggerArmedState, level float64) (conditionMet bool, fired bool) {
	thresh := float64(cfg.Threshold)
	hyst := float64(cfg.Hysteresis)

	if cfg.Direction == TriggerNone {
		// Treat TriggerNone as simple level > thresh
		return level >= thresh, false
	}

	switch cfg.Direction {
	case TriggerRising, TriggerRisingLower:
		if *state == TriggerStateIdle && level <= (thresh-hyst) {
			*state = TriggerStateArmedRising
		} else if *state == TriggerStateArmedRising && level > thresh {
			*state = TriggerStateIdle
			return true, true
		}
	case TriggerFalling, TriggerFallingLower:
		if *state == TriggerStateIdle && level >= (thresh+hyst) {
			*state = TriggerStateArmedFalling
		} else if *state == TriggerStateArmedFalling && level < thresh {
			*state = TriggerStateIdle
			return true, true
		}
	case TriggerRisingOrFalling:
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

// updateInsideWindow tracks whether the signal is currently inside the window by
// maintaining an armed state for both upper and lower thresholds. This is used
// for the GreaterThan Window PW check, where we need to know if the signal is
// still within the window without consuming the edge state machines used by the
// main trigger.
func (td *TriggerDetector) updateInsideWindow(state *ChannelTriggerState, cfg TriggerChannelConfig, level float64) {
	upper := float64(cfg.Threshold)
	lower := float64(cfg.ThresholdLower)
	upperHyst := float64(cfg.Hysteresis)
	lowerHyst := float64(cfg.ThresholdLowerHysteresis)

	// Upper threshold: arm falling (above window) → fire falling into window
	if state.UpperState == TriggerStateIdle && level >= (upper+upperHyst) {
		state.UpperState = TriggerStateArmedFalling
	} else if state.UpperState == TriggerStateArmedFalling && level <= upper {
		state.UpperState = TriggerStateIdle
	}

	// Lower threshold: arm rising (below window) → fire rising into window
	if state.LowerState == TriggerStateIdle && level <= (lower-lowerHyst) {
		state.LowerState = TriggerStateArmedRising
	} else if state.LowerState == TriggerStateArmedRising && level >= lower {
		state.LowerState = TriggerStateIdle
	}
}

func (td *TriggerDetector) evaluateWindowTrigger(
	cfg TriggerChannelConfig,
	state *ChannelTriggerState,
	level float64,
	signalFunc func(t float64, ch ChannelId) float64,
	t float64,
	dt float64,
	ch ChannelId,
) (conditionMet bool, fired bool, timeOffset float64) {

	upperCfg := TriggerChannelConfig{
		Threshold:  cfg.Threshold,
		Hysteresis: cfg.Hysteresis,
	}
	lowerCfg := TriggerChannelConfig{
		Threshold:  cfg.ThresholdLower,
		Hysteresis: cfg.ThresholdLowerHysteresis,
	}

	isRunt := cfg.Direction == TriggerPositiveRunt || cfg.Direction == TriggerNegativeRunt

	if !isRunt {
		switch cfg.Direction {
		case TriggerEnter, TriggerInside, TriggerAbove, TriggerRising:
			upperCfg.Direction = TriggerFalling
			lowerCfg.Direction = TriggerRising
		case TriggerExit, TriggerOutside, TriggerBelow, TriggerFalling:
			upperCfg.Direction = TriggerRising
			lowerCfg.Direction = TriggerFalling
		case TriggerEnterOrExit, TriggerRisingOrFalling:
			upperCfg.Direction = TriggerRisingOrFalling
			lowerCfg.Direction = TriggerRisingOrFalling
		}

		_, upperFired := td.evaluateLevelTrigger(upperCfg, &state.UpperState, level)
		_, lowerFired := td.evaluateLevelTrigger(lowerCfg, &state.LowerState, level)

		if upperFired && lowerFired {
			// Jumped over the entire window in one step (full swing). Not a window trigger event.
			return false, false, 0
		}
		if upperFired || lowerFired {
			return true, true, 0
		}
		return false, false, 0
	}

	// State-machine based Runt trigger
	if cfg.Direction == TriggerPositiveRunt {
		// Positive Runt looks for rising across lower threshold initially,
		// and rejects if it rises across upper threshold.
		upperCfg.Direction = TriggerRising
		lowerCfg.Direction = TriggerRising
	} else {
		// Negative Runt looks for falling across upper threshold initially,
		// and rejects if it falls across lower threshold.
		upperCfg.Direction = TriggerFalling
		lowerCfg.Direction = TriggerFalling
	}

	_, upperFired := td.evaluateLevelTrigger(upperCfg, &state.UpperState, level)
	_, lowerFired := td.evaluateLevelTrigger(lowerCfg, &state.LowerState, level)

	if upperFired && lowerFired {
		// Full swing in one step, ignore
		state.RuntPending = false
		return false, false, 0
	}

	if state.RuntPending {
		if cfg.Direction == TriggerPositiveRunt {
			if upperFired {
				// Reached upper threshold -> Full swing, not a runt.
				state.RuntPending = false
			} else if level < float64(cfg.ThresholdLower) {
				// Fell back below lower threshold without crossing upper -> Confirmed runt!
				state.RuntPending = false
				return true, true, state.RuntStartTime - t
			}
		} else { // TriggerNegativeRunt
			if lowerFired {
				// Reached lower threshold -> Full swing, not a runt.
				state.RuntPending = false
			} else if level > float64(cfg.Threshold) {
				// Rose back above upper threshold without crossing lower -> Confirmed runt!
				state.RuntPending = false
				return true, true, state.RuntStartTime - t
			}
		}
	} else {
		// Not pending. Look for the initial crossing to start a runt pulse.
		if cfg.Direction == TriggerPositiveRunt {
			if lowerFired {
				state.RuntPending = true
				state.RuntStartTime = t
			}
		} else { // TriggerNegativeRunt
			if upperFired {
				state.RuntPending = true
				state.RuntStartTime = t
			}
		}
	}

	return false, false, 0
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
