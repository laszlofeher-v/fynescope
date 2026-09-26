package netcontrol

import (
	"bytes"
	"encoding/gob"
	"image/color"
	"reflect"
	"testing"

	"fynescope/control"
	"fynescope/genericps"
	"fynescope/settings"
)

// roundtrip encodes val with gob into a buffer, decodes into out, and returns an error if any.
func roundtrip(t *testing.T, val interface{}, out interface{}) {
	t.Helper()
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(val); err != nil {
		t.Fatalf("gob.Encode failed: %v", err)
	}
	dec := gob.NewDecoder(&buf)
	if err := dec.Decode(out); err != nil {
		t.Fatalf("gob.Decode failed: %v", err)
	}
}

func TestCommandMessage_Roundtrip(t *testing.T) {
	tests := []struct {
		name    string
		cmdType string
		payload interface{}
	}{
		{
			name:    "SetTrigger",
			cmdType: TypeSetTrigger,
			payload: &SetTriggerCmd{
				Desc: control.TriggerDesc{
					Enabled:            true,
					TriggerADC:         1234,
					LowerTriggerADC:    -1234,
					HysteresisADC:      100,
					LowerHysteresisADC: 100,
					UpperHysteresis:    500,
					LowerHysteresis:    -500,
					Source:             genericps.ChA,
					ThresholdDirection: genericps.TriggerRising,
					ThresholdMode:      genericps.Level,
					Mode:               control.Auto,
					Type:               control.Simple,
					Mv:                 250,
					LowerMv:            -250,
					ComplexProperties: []genericps.TriggerChannelProperties{
						{
							ThresholdUpper:           100,
							ThresholdUpperHysteresis: 10,
							ThresholdLower:           -100,
							ThresholdLowerHysteresis: 10,
							Channel:                  genericps.ChA,
							ThresholdMode:            genericps.Level,
						},
					},
					ComplexConditions: []genericps.TriggerConditions{
						{
							ChannelA: genericps.CondTrue,
							ChannelB: genericps.CondFalse,
						},
					},
					ComplexDirections: []control.TriggerDirections{
						{
							ChannelA: genericps.TriggerRising,
						},
					},
					IntervalType:          genericps.PwTypeGreaterThan,
					IntervalTimeLower:     1e-6,
					IntervalTimeUpper:     2e-6,
					XOffset:               0.05,
					AutoTriggerMs:         500,
					DigitalTriggerEnabled: true,
					DigitalDirections: []genericps.DigitalChannelDirections{
						{Channel: 0, Direction: genericps.DigitalDirectionHigh},
					},
					DigitalAnalogOperand:   genericps.OperandOr,
					DigitalChannelsOperand: genericps.OperandAnd,
					EtsInterleave:          4,
					EtsCycles:              10,
				},
			},
		},
		{
			name:    "SetChannel",
			cmdType: TypeSetChannel,
			payload: &SetChannelCmd{
				Settings: settings.ChSettings{
					ID:         genericps.ChB,
					Inverted:   true,
					X10:        true,
					Col:        [2]color.NRGBA{{R: 255, G: 128, B: 0, A: 255}, {R: 128, G: 64, B: 0, A: 255}},
					VRange:     genericps.Range_1v,
					CoupleType: genericps.Dc,
					Enabled:    true,
					Offset:     0.25,
					RlcFilter: settings.RlcFilterSettings{
						GeneratorSource: genericps.ChA,
						Enabled:         true,
						Type:            "Lowpass RC",
						R:               1000,
						RUnit:           "Ohm",
						C:               1e-9,
						CUnit:           "F",
					},
					DigitalFilter: settings.DigitalFilterSettings{
						LowpassEnabled: true,
						LowpassFc:      50000,
					},
				},
			},
		},
		{
			name:    "SetDigitalPort",
			cmdType: TypeSetDigitalPort,
			payload: &SetDigitalPortCmd{
				Msg: control.DigitalPortMsg{
					Port: genericps.Port0,
					Settings: settings.DigitalPortSettings{
						Enabled:   true,
						Threshold: 1650,
					},
				},
			},
		},
		{
			name:    "SetGenerator",
			cmdType: TypeSetGenerator,
			payload: &SetGeneratorCmd{
				Desc: control.GeneratorDesc{
					OffsetVoltage:     500,
					PkToPK:            2000,
					WaveType:          genericps.Sine,
					StartFrequency:    1000.0,
					StopFrequency:     2000.0,
					Increment:         100.0,
					DwellTime:         0.01,
					SweepType:         genericps.SweepUp,
					Operation:         genericps.EsOff,
					Shots:             1,
					Sweeps:            2,
					TriggerType:       genericps.SigGenRising,
					TriggerSource:     genericps.SigGenScopeTrig,
					ExtInThreshold:    0,
					Phase:             90.0,
					Channel:           genericps.ChA,
					On:                true,
					ArbitraryWaveform: []int16{0, 1000, 2000, 0, -2000, -1000},
					StartDeltaPhase:   10,
					StopDeltaPhase:    20,
					IndexMode:         genericps.Single,
					SpiDataValue:      0x1234,
					I2cAddressValue:   0x50,
				},
			},
		},
		{
			name:    "SetDemoGen",
			cmdType: TypeSetDemoGen,
			payload: &SetDemoGenCmd{
				Desc: control.GeneratorDesc{
					OffsetVoltage:  0,
					PkToPK:         1500,
					WaveType:       genericps.Square,
					StartFrequency: 5000.0,
					On:             true,
				},
			},
		},
		{
			name:    "SetInterpolationMode",
			cmdType: TypeSetInterpolationMode,
			payload: &SetInterpolationModeCmd{
				Mode: settings.Sinc,
			},
		},
		{
			name:    "SetScopeScreenWidth",
			cmdType: TypeSetScopeScreenWidth,
			payload: &SetScopeScreenWidthCmd{
				Width: 1024.5,
			},
		},
		{
			name:    "SetMaxScreenTime",
			cmdType: TypeSetMaxScreenTime,
			payload: &SetMaxScreenTimeCmd{
				Time: 0.05,
			},
		},
		{
			name:    "SuggestSampleCount",
			cmdType: TypeSuggestSampleCount,
			payload: &SuggestSampleCountCmd{
				Count: 100000,
			},
		},
		{
			name:    "SetETSMode",
			cmdType: TypeSetETSMode,
			payload: &SetETSModeCmd{},
		},
		{
			name:    "SetBlockMode",
			cmdType: TypeSetBlockMode,
			payload: &SetBlockModeCmd{},
		},
		{
			name:    "Stop",
			cmdType: TypeStop,
			payload: &StopCmd{},
		},
		{
			name:    "Shutdown",
			cmdType: TypeShutdown,
			payload: &ShutdownCmd{},
		},
		{
			name:    "SetResolutionMode",
			cmdType: TypeSetResolutionMode,
			payload: &SetResolutionModeCmd{
				Mode: genericps.RatioModeDecimate,
			},
		},
		{
			name:    "SetDigitalPortEnabled",
			cmdType: TypeSetDigitalPortEnabled,
			payload: &SetDigitalPortEnabledCmd{
				Port:    1,
				Enabled: true,
			},
		},
		{
			name:    "SetMaxSamplingRate",
			cmdType: TypeSetMaxSamplingRate,
			payload: &SetMaxSamplingRateCmd{
				Rate: 500000000,
			},
		},
		{
			name:    "RequestRestart",
			cmdType: TypeRequestRestart,
			payload: &RequestRestartCmd{},
		},
		{
			name:    "SetStreamEnabled",
			cmdType: TypeSetStreamEnabled,
			payload: &SetStreamEnabledCmd{
				Enabled: true,
			},
		},
		{
			name:    "NewChannels",
			cmdType: TypeNewChannels,
			payload: &NewChannelsCmd{
				NumberOfChannels: 4,
			},
		},
		{
			name:    "SetInfo",
			cmdType: TypeSetInfo,
			payload: &SetInfoCmd{
				Info: "PicoScope 2206B",
			},
		},
		{
			name:    "SetScopeModel",
			cmdType: TypeSetScopeModel,
			payload: &SetScopeModelCmd{
				Model: control.Scope2206B,
			},
		},
		{
			name:    "SetDemoDigitalGen",
			cmdType: TypeSetDemoDigitalGen,
			payload: &SetDemoDigitalGenCmd{
				Port0Enabled: true,
				Port1Enabled: false,
				Frequency:    10000.0,
				Direction:    genericps.DigitalDemoGenDirectionUp,
				Encoding:     genericps.DigitalDemoGenEncodingGray,
				Mode:         genericps.DigitalDemoGenModeSynchronous,
				BitDelay:     0.0001,
			},
		},
		{
			name:    "SetDemoRlcFilter",
			cmdType: TypeSetDemoRlcFilter,
			payload: &SetDemoRlcFilterCmd{
				Channel:    genericps.ChA,
				GenSource:  genericps.ChB,
				Enabled:    true,
				FilterType: "Bandpass RLC",
				R:          100.0,
				RUnit:      "Ohm",
				L:          0.001,
				LUnit:      "H",
				C:          1e-6,
				CUnit:      "uF",
			},
		},
		{
			name:    "GetAnalogueOffsetReq",
			cmdType: TypeGetAnalogueOffset,
			payload: &GetAnalogueOffsetReq{
				VoltageRange: 5,
				Coupling:     genericps.Dc,
			},
		},
		{
			name:    "ChannelRangesReq",
			cmdType: TypeChannelRanges,
			payload: &ChannelRangesReq{
				Channel: genericps.ChA,
			},
		},
		{
			name:    "MinMaxValuesReq",
			cmdType: TypeMinMaxValues,
			payload: &MinMaxValuesReq{},
		},
		{
			name:    "UnitVariantInfoReq",
			cmdType: TypeUnitVariantInfo,
			payload: &UnitVariantInfoReq{},
		},
		{
			name:    "UnitBatchAndSerialInfoReq",
			cmdType: TypeUnitBatchAndSerialInfo,
			payload: &UnitBatchAndSerialInfoReq{},
		},
		{
			name:    "GetInfoReq",
			cmdType: TypeGetInfo,
			payload: &GetInfoReq{},
		},
		{
			name:    "GetScopeModelReq",
			cmdType: TypeGetScopeModel,
			payload: &GetScopeModelReq{},
		},
		{
			name:    "GetMaxSamplingRateReq",
			cmdType: TypeGetMaxSamplingRate,
			payload: &GetMaxSamplingRateReq{},
		},
		{
			name:    "GetStreamEnabledReq",
			cmdType: TypeGetStreamEnabled,
			payload: &GetStreamEnabledReq{},
		},
		{
			name:    "GetSamplingTimeIntervalReq",
			cmdType: TypeGetSamplingTimeInterval,
			payload: &GetSamplingTimeIntervalReq{},
		},
		{
			name:    "GetTimeBaseReq",
			cmdType: TypeGetTimeBase,
			payload: &GetTimeBaseReq{},
		},
		{
			name:    "NumberOfEnabledAnalogChannelsReq",
			cmdType: TypeNumberOfEnabledAnalogChannels,
			payload: &NumberOfEnabledAnalogChannelsReq{},
		},
		{
			name:    "GetEtsLimitsReq",
			cmdType: TypeGetEtsLimits,
			payload: &GetEtsLimitsReq{},
		},
		{
			name:    "IsDemoReq",
			cmdType: TypeIsDemo,
			payload: &IsDemoReq{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			orig := &CommandMessage{
				ID:      42,
				Type:    tc.cmdType,
				Payload: tc.payload,
			}
			var decoded CommandMessage
			roundtrip(t, orig, &decoded)

			if decoded.ID != orig.ID {
				t.Errorf("ID mismatch: got %d, want %d", decoded.ID, orig.ID)
			}
			if decoded.Type != orig.Type {
				t.Errorf("Type mismatch: got %s, want %s", decoded.Type, orig.Type)
			}
			if !reflect.DeepEqual(decoded.Payload, orig.Payload) {
				t.Errorf("Payload mismatch:\ngot:  %+v\nwant: %+v", decoded.Payload, orig.Payload)
			}
		})
	}
}

func TestServerMessage_Telemetry_Roundtrip(t *testing.T) {
	tests := []struct {
		name    string
		msgType string
		payload interface{}
	}{
		{
			name:    "DataMessage",
			msgType: TypeRefresh,
			payload: &DataMessage{
				Buffers: [][]int16{
					{0, 100, 200, 300, 400},
					{-100, -200, -300, -400, -500},
				},
				BuffersMin: [][]int16{
					{0, 50, 100, 150, 200},
					{-50, -100, -150, -200, -250},
				},
				DigitalBuffers: [][]int16{
					{1, 0, 1, 0, 1},
				},
				StartTimeOffset:      123456789,
				XRoundError:          0.00012,
				SamplingTimeInterval: 1e-8,
			},
		},
		{
			name:    "EtsDataMessage",
			msgType: TypeRefreshEts,
			payload: &EtsDataMessage{
				Buffers: [][]int16{
					{10, 20, 30, 40, 50},
				},
				EtsOutBuffer:         []int64{0, 100, 200, 300, 400},
				XRoundError:          0.0005,
				SamplingTimeInterval: 2e-9,
			},
		},
		{
			name:    "BufferResizeMessage",
			msgType: TypeBufferResize,
			payload: &BufferResizeMessage{
				Size: 4096,
			},
		},
		{
			name:    "EtsBufferResizeMessage",
			msgType: TypeEtsBufferResize,
			payload: &EtsBufferResizeMessage{
				Size: 8192,
			},
		},
		{
			name:    "StatusMessage",
			msgType: TypeStatus,
			payload: &StatusMessage{
				Text:      "Trigger ready",
				ErrorType: control.Info,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			orig := &ServerMessage{
				Type:    tc.msgType,
				Payload: tc.payload,
			}
			var decoded ServerMessage
			roundtrip(t, orig, &decoded)

			if decoded.Type != orig.Type {
				t.Errorf("Type mismatch: got %s, want %s", decoded.Type, orig.Type)
			}
			if !reflect.DeepEqual(decoded.Payload, orig.Payload) {
				t.Errorf("Payload mismatch:\ngot:  %+v\nwant: %+v", decoded.Payload, orig.Payload)
			}
		})
	}
}

func TestResponseMessage_Roundtrip(t *testing.T) {
	tests := []struct {
		name    string
		result  interface{}
		success bool
		errMsg  string
	}{
		{
			name:    "GetAnalogueOffsetRsp",
			result:  &GetAnalogueOffsetRsp{Max: 20.0, Min: -20.0},
			success: true,
		},
		{
			name:    "ChannelRangesRsp",
			result:  &ChannelRangesRsp{Ranges: []int32{1, 2, 5, 10}},
			success: true,
		},
		{
			name:    "MinMaxValuesRsp",
			result:  &MinMaxValuesRsp{Min: -32768, Max: 32767},
			success: true,
		},
		{
			name:    "UnitVariantInfoRsp",
			result:  &UnitVariantInfoRsp{Info: "2206B"},
			success: true,
		},
		{
			name:    "UnitBatchAndSerialInfoRsp",
			result:  &UnitBatchAndSerialInfoRsp{Info: "KP381/0078"},
			success: true,
		},
		{
			name:    "GetInfoRsp",
			result:  &GetInfoRsp{Info: "PicoScope 2000 series"},
			success: true,
		},
		{
			name:    "GetScopeModelRsp",
			result:  &GetScopeModelRsp{Model: control.Scope2206B},
			success: true,
		},
		{
			name:    "GetMaxSamplingRateRsp",
			result:  &GetMaxSamplingRateRsp{Rate: 1000000000},
			success: true,
		},
		{
			name:    "GetStreamEnabledRsp",
			result:  &GetStreamEnabledRsp{Enabled: true},
			success: true,
		},
		{
			name:    "GetSamplingTimeIntervalRsp",
			result:  &GetSamplingTimeIntervalRsp{Interval: 8e-9},
			success: true,
		},
		{
			name:    "GetTimeBaseRsp",
			result:  &GetTimeBaseRsp{TimeBase: 125},
			success: true,
		},
		{
			name:    "NumberOfEnabledAnalogChannelsRsp",
			result:  &NumberOfEnabledAnalogChannelsRsp{Count: 2},
			success: true,
		},
		{
			name:    "GetEtsLimitsRsp",
			result:  &GetEtsLimitsRsp{MaxInterleave: 10, MaxCycles: 20},
			success: true,
		},
		{
			name:    "IsDemoRsp",
			result:  &IsDemoRsp{IsDemo: true},
			success: true,
		},
		{
			name:    "FailureResponse",
			result:  nil,
			success: false,
			errMsg:  "device disconnected",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			origRsp := &ResponseMessage{
				ID:      101,
				Success: tc.success,
				Error:   tc.errMsg,
				Result:  tc.result,
			}
			origServerMsg := &ServerMessage{
				Type:    TypeResponse,
				Payload: origRsp,
			}

			var decodedServerMsg ServerMessage
			roundtrip(t, origServerMsg, &decodedServerMsg)

			if decodedServerMsg.Type != TypeResponse {
				t.Fatalf("ServerMessage Type mismatch: got %s, want %s", decodedServerMsg.Type, TypeResponse)
			}
			decodedRsp, ok := decodedServerMsg.Payload.(*ResponseMessage)
			if !ok {
				t.Fatalf("expected *ResponseMessage payload, got %T", decodedServerMsg.Payload)
			}
			if decodedRsp.ID != origRsp.ID {
				t.Errorf("ID mismatch: got %d, want %d", decodedRsp.ID, origRsp.ID)
			}
			if decodedRsp.Success != origRsp.Success {
				t.Errorf("Success mismatch: got %v, want %v", decodedRsp.Success, origRsp.Success)
			}
			if decodedRsp.Error != origRsp.Error {
				t.Errorf("Error mismatch: got %s, want %s", decodedRsp.Error, origRsp.Error)
			}
			if !reflect.DeepEqual(decodedRsp.Result, origRsp.Result) {
				t.Errorf("Result mismatch:\ngot:  %+v\nwant: %+v", decodedRsp.Result, origRsp.Result)
			}
		})
	}
}

func TestLargeWaveformBuffer_Roundtrip(t *testing.T) {
	// Simulate 2 channels of 65536 samples (128 KB of int16 data)
	samples := 65536
	chA := make([]int16, samples)
	chB := make([]int16, samples)
	for i := 0; i < samples; i++ {
		chA[i] = int16(i % 32767)
		chB[i] = int16(-i % 32767)
	}

	orig := &ServerMessage{
		Type: TypeRefresh,
		Payload: &DataMessage{
			Buffers:              [][]int16{chA, chB},
			StartTimeOffset:      1000000,
			XRoundError:          0.00001,
			SamplingTimeInterval: 1e-9,
		},
	}

	var decoded ServerMessage
	roundtrip(t, orig, &decoded)

	data, ok := decoded.Payload.(*DataMessage)
	if !ok {
		t.Fatalf("expected *DataMessage payload, got %T", decoded.Payload)
	}
	if len(data.Buffers) != 2 {
		t.Fatalf("expected 2 buffers, got %d", len(data.Buffers))
	}
	if len(data.Buffers[0]) != samples || len(data.Buffers[1]) != samples {
		t.Fatalf("buffer lengths mismatch: chA=%d, chB=%d", len(data.Buffers[0]), len(data.Buffers[1]))
	}
	for i := 0; i < 100; i++ { // check sample points
		if data.Buffers[0][i] != chA[i] || data.Buffers[1][i] != chB[i] {
			t.Fatalf("sample %d mismatch: A=%d vs %d, B=%d vs %d", i, data.Buffers[0][i], chA[i], data.Buffers[1][i], chB[i])
		}
	}
}
