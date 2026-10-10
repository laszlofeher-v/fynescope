package demo

type (
	// PwqConditions defines pulse-width qualifier trigger conditions across analog, external, aux, and digital inputs.
	PwqConditions struct {
		ChannelA TriggerState // Trigger condition for Channel A (e.g. ConditionDontCare, ConditionTrue, ConditionFalse).
		ChannelB TriggerState // Trigger condition for Channel B.
		ChannelC TriggerState // Trigger condition for Channel C.
		ChannelD TriggerState // Trigger condition for Channel D.
		External TriggerState // Trigger condition for External trigger input.
		Aux      TriggerState // Trigger condition for Aux trigger input.
		Digital  TriggerState // Trigger condition for digital MSO channels.
	}

	// Pwq specifies the full pulse-width qualifier configuration including conditions, directions, and sample duration window.
	Pwq struct {
		Conditions  PwqConditions      // Trigger conditions to qualify against.
		NConditions int16              // Number of conditions configured.
		Direction   ThresholdDirection // Threshold transition direction (e.g. Rising, Falling).
		Lower       uint32             // Minimum sample count duration for the pulse.
		Upper       uint32             // Maximum sample count duration for the pulse.
		Type        PulseWidthType     // Pulse width comparison type (e.g. InRange, OutOfRange, LessThan, GreaterThan).
	}

	// ChannelDesc describes the hardware configuration and status of an individual oscilloscope channel.
	ChannelDesc struct {
		CoupleType Coupling  // Input coupling (AC, DC).
		Range      RangeEnum // Input voltage range index.
		Enabled    bool      // True if the channel is enabled for acquisition.
		Inverted   bool      // True if the channel signal polarity is inverted.
		Offset     float32   // Analogue offset voltage in volts.
	}

	// PsDesc encapsulates a simulated oscilloscope hardware handle.
	PsDesc struct {
		handle int16 // Unique identifier handle for the simulated scope device.
	}

	// BlockReady is the callback invoked when a block-mode capture has finished and data is ready to retrieve.
	BlockReady func(handle int16, status int, param any)

	// DataReady is the callback invoked when an asynchronous get-values call has populated data buffers.
	DataReady func(handle int16, status int, noOfSamples uint32, overflow int16, param any)

	// StreamingReady is the callback invoked when a streaming-mode chunk of data is ready for processing.
	StreamingReady func(handle int16, noOfSamples int32, startIndex uint32, overflow int16,
		triggeredAt uint32, triggered, autoStop int16, param any) (err error)

	// TriggerChannelProperties defines upper/lower voltage thresholds, hysteresis, and threshold mode for a channel.
	TriggerChannelProperties struct {
		ThresholdUpper           int16           // Upper voltage threshold level in ADC counts.
		ThresholdUpperHysteresis uint16          // Upper threshold hysteresis band.
		ThresholdLower           int16           // Lower voltage threshold level in ADC counts.
		ThresholdLowerHysteresis uint16          // Lower threshold hysteresis band.
		Channel                  ChannelId       // Channel identifier this property set applies to.
		ThresholdMode            ThresholdModeId // Threshold comparison mode (Level or Window).
	}

	// TriggerConditions defines conditions required to trigger a capture across channels and qualifiers.
	TriggerConditions struct {
		ChannelA            TriggerState // Condition for Channel A.
		ChannelB            TriggerState // Condition for Channel B.
		ChannelC            TriggerState // Condition for Channel C.
		ChannelD            TriggerState // Condition for Channel D.
		External            TriggerState // Condition for External trigger.
		Aux                 TriggerState // Condition for Aux trigger.
		PulseWidthQualifier TriggerState // Condition for Pulse-Width Qualifier.
		Digital             TriggerState // Condition for Digital channels.
	}

	// TriggerDirections specifies threshold transition directions
	// (e.g., Rising, Falling) for analog/external channels.
	TriggerDirections struct {
		ChannelA ThresholdDirection // Direction for Channel A.
		ChannelB ThresholdDirection // Direction for Channel B.
		ChannelC ThresholdDirection // Direction for Channel C.
		ChannelD ThresholdDirection // Direction for Channel D.
		Ext      ThresholdDirection // Direction for External trigger.
		Aux      ThresholdDirection // Direction for Aux trigger.
	}

	// DigitalChannelDirections specifies digital pin trigger
	// conditions (e.g. Low, High, Rising, Falling).
	DigitalChannelDirections struct {
		Channel   DigitalChannel   // Digital pin identifier.
		Direction DigitalDirection // Transition or level direction condition for this digital pin.
	}
)
