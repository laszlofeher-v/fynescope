//go:build ps3000

package ps3000a

type (
	PwqConditions struct {
		ChannelA TriggerState
		ChannelB TriggerState
		ChannelC TriggerState
		ChannelD TriggerState
		External TriggerState
		Aux      TriggerState
		Digital  TriggerState
	}
	Pwq struct {
		Conditions  PwqConditions
		NConditions int16
		Direction   ThresholdDirection
		Lower       uint32
		Upper       uint32
		Type        PulseWidthType
	}

	ChannelDesc struct {
		CoupleType Coupling
		Range      RangeEnum
		Enabled    bool
		Inverted   bool
		Offset     float32
	}

	BlockReady     func(handle int16, status int, param any)
	DataReady      func(handle int16, status int, noOfSamples uint32, overflow int16, param any)
	StreamingReady func(handle int16, noOfSamples int32, startIndex uint32, overflow int16,
		triggeredAt uint32, triggered, autoStop int16, param any) (err error)

	TriggerChannelProperties struct {
		ThresholdUpper           int16
		ThresholdUpperHysteresis uint16
		ThresholdLower           int16
		ThresholdLowerHysteresis uint16
		Channel                  ChannelId
		ThresholdMode            ThresholdModeId
	}
	TriggerConditions struct {
		ChannelA            TriggerState
		ChannelB            TriggerState
		ChannelC            TriggerState
		ChannelD            TriggerState
		External            TriggerState
		Aux                 TriggerState
		PulseWidthQualifier TriggerState
		Digital             TriggerState
	}

	TriggerDirections struct {
		ChannelA ThresholdDirection
		ChannelB ThresholdDirection
		ChannelC ThresholdDirection
		ChannelD ThresholdDirection
		Ext      ThresholdDirection
		Aux      ThresholdDirection
	}

	DigitalChannelDirections struct {
		Channel   DigitalChannel
		Direction DigitalDirection
	}
)
