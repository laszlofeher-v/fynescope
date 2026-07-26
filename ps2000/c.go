//go:build !noscope && ps2000

package ps2000

// #cgo CFLAGS: -g -Wall
// #cgo LDFLAGS: -L/opt/picoscope/lib/ -lps2000
// #include <stdlib.h>
// #include "/opt/picoscope/include/libps2000/ps2000.h"
/*
// Forward declarations
void lpStreamingReady2000(
  int16_t **overviewBuffers,
  int16_t   overflow,
  uint32_t  triggeredAt,
  int16_t   triggered,
  int16_t   auto_stop,
  uint32_t  nValues
);
*/
import "C"

import (
	"fmt"
	"fynescope/genericps"
	"log/slog"
	"time"
	"unsafe"
)

func init() {
	var scopeHandler genericps.ScopeHandler
	scopeHandler.EnumerateUnits = enumerateUnits
	scopeHandler.Dispatch = dispatch
	scopeHandler.OpenUnit = openUnit
	scopeHandler.OpenUnitAsync = openUnitAsync
	scopeHandler.OpenUnitProgress = openUnitProgress
	scopeHandler.Id = "ps2000"
	genericps.Register(scopeHandler)
}

func boolToint16(b bool) int16 {
	if b {
		return int16(1)
	}
	return int16(0)
}

func enumerateUnits(bufferLen int16) (count int16, serials string, serialLth int16, err error) {
	// ps2000 doesn't enumerate units natively by returning strings.
	// We'll check if a unit is connected by attempting to open it.
	handle := C.ps2000_open_unit()
	if handle > 0 {
		C.ps2000_close_unit(handle)
		count = 1
		serials = "PS2000_1"
		serialLth = int16(len(serials) + 1)
	} else {
		count = 0
		serials = ""
		serialLth = 0
	}
	return
}

func openUnit(serial string) (handle int16, err error) {
	slog.Debug("ps2000OpenUnit")
	stat := C.ps2000_open_unit()
	if stat <= 0 {
		err = fmt.Errorf("OpenUnit failed, handle %d", stat)
		return
	}
	handle = int16(stat)
	loadConstants()
	return
}

func openUnitAsync(serial string) (status int16, err error) {
	err = fmt.Errorf("openUnitAsync not supported on ps2000")
	return
}

func openUnitProgress() (handle int16, progressPercent, complete int16, err error) {
	err = fmt.Errorf("openUnitProgress not supported on ps2000")
	return
}

func ps2000CloseUnit(handle int16) (err error) {
	slog.Debug("ps2000CloseUnit", "handle", handle)
	C.ps2000_close_unit((C.short)(handle))
	return
}

func ps2000GetUnitInfo(handle int16, info PicoInfo) (infoString string, err error) {
	const listLen = 4096
	var cstrPtr *C.int8_t
	cstrPtr = (*C.int8_t)(C.malloc(C.sizeof_int8_t * listLen))
	defer C.free(unsafe.Pointer(cstrPtr))
	
	slog.Debug("ps2000GetUnitInfo", "handle", handle, "info", info)
	stat := C.ps2000_get_unit_info((C.short)(handle), cstrPtr, (C.short)(listLen), (C.short)(info))
	if stat == 0 {
		err = fmt.Errorf("GetUnitInfo failed")
		return
	}
	// find length manually since it doesn't return required size
	b := C.GoBytes(unsafe.Pointer(cstrPtr), (C.int)(listLen))
	n := 0
	for i, v := range b {
		if v == 0 {
			n = i
			break
		}
	}
	infoString = string(b[:n])
	return
}

func ps2000FlashLed(handle int16, start int16) (err error) {
	slog.Debug("ps2000FlashLed", "handle", handle, "start", start)
	C.ps2000_flash_led((C.short)(handle)) // does not take start on ps2000, or wait does it?
	return
}

// cgo callback workaround
var regLpStreamingReadyGo genericps.StreamingReady // registered go callback function

//export lpStreamingReadyGo2000
func lpStreamingReadyGo2000(overviewBuffers **C.int16_t, overflow C.int16_t, triggeredAt C.uint32_t, triggered C.int16_t, auto_stop C.int16_t, nValues C.uint32_t) {
	if regLpStreamingReadyGo != nil {
		// we can't easily extract data here, but typically streaming handles this differently on ps2000.
		// For simplicity, we just trigger the generic callback.
		regLpStreamingReadyGo(0, int32(nValues), 0, int16(overflow), uint32(triggeredAt), int16(triggered), int16(auto_stop), nil)
	}
}

func ps2000GetStreamingLatestValues(handle int16, lpStreamingReadyGoPar genericps.StreamingReady, param interface{}) (err error) {
	regLpStreamingReadyGo = lpStreamingReadyGoPar
	C.ps2000_get_streaming_last_values((C.short)(handle), (C.GetOverviewBuffersMaxMin)(C.lpStreamingReady2000))
	return
}

func ps2000GetTimebase(handle int16, timeBase uint32, noOfSamples int32, overSample int16, segmentIndex uint32) (timeIntervalNanoseconds, maxSamples int32, err error) {
	slog.Debug("ps2000GetTimebase", "handle", handle, "timeBase", timeBase, "noOfSamples", noOfSamples, "overSample", overSample)
	var timeUnits C.int16_t
	stat := C.ps2000_get_timebase((C.short)(handle), (C.short)(timeBase), (C.int)(noOfSamples),
		(*C.int)(&timeIntervalNanoseconds), &timeUnits, (C.short)(overSample),
		(*C.int)(&maxSamples))
	if stat == 0 {
		err = fmt.Errorf("GetTimebase failed")
	}
	// For ps2000, timeUnits indicates the units of timeInterval. We need to convert it to ns.
	var multiplier float64 = 1.0
	switch TimeUnits(timeUnits) {
	case TuFs:
		multiplier = 1e-6
	case TuPs:
		multiplier = 1e-3
	case TuNs:
		multiplier = 1
	case TuUs:
		multiplier = 1e3
	case TuMs:
		multiplier = 1e6
	case TuS:
		multiplier = 1e9
	}
	timeIntervalNanoseconds = int32(float64(timeIntervalNanoseconds) * multiplier)
	return
}

func ps2000SetChannel(handle int16, channel ChannelId, enabled bool, couplingType Coupling, voltageRange RangeEnum, analogOffset float32) (err error) {
	slog.Debug("ps2000SetChannel", "handle", handle, "channel", channel, "enabled", enabled, "couplingType", couplingType, "voltageRange", voltageRange)
	// ps2000 doesn't have analogOffset
	stat := C.ps2000_set_channel((C.short)(handle), (C.short)(channel), (C.short)(boolToint16(enabled)),
		(C.short)(couplingType), (C.short)(voltageRange))
	if stat == 0 {
		err = fmt.Errorf("SetChannel failed")
	}
	return
}

func ps2000MaximumValue(handle int16) (value int32, err error) {
	return 32767, nil
}

func ps2000MinimumValue(handle int16) (value int32, err error) {
	return -32767, nil
}

func ps2000SetSimpleTrigger(handle int16, enable bool, source ChannelId, threshold int16,
	direction ThresholdDirection, delay uint32, autoTriggerMs int16) (err error) {
	slog.Debug("ps2000SetSimpleTrigger", "handle", handle, "enable", enable, "src", source, "threshold", threshold, "direction", direction, "delay", delay, "autoTriggerMs", autoTriggerMs)
	var stat C.short
	if !enable {
		stat = C.ps2000_set_trigger((C.short)(handle), 5, 0, 0, 0, 0) // PS2000_NONE=5
	} else {
		stat = C.ps2000_set_trigger((C.short)(handle), (C.short)(source), (C.short)(threshold),
			(C.short)(direction), (C.short)(delay), (C.short)(autoTriggerMs))
	}
	if stat == 0 {
		err = fmt.Errorf("SetSimpleTrigger failed")
	}
	return
}

func ps2000RunStreaming(handle int16, reqSampleInterval uint32, sampleIntervalTimeUnits TimeUnits,
	maxPreTriggerSamples, maxPostTriggerSamples uint32,
	autoStop bool, downSampleRatio uint32, downSampleRatioMode RatioMode,
	overviewBufferSize uint32) (sampleInterval uint32, err error) {
	slog.Debug("ps2000RunStreaming")
	stat := C.ps2000_run_streaming_ns((C.short)(handle), (C.uint)(reqSampleInterval),
		(C.PS2000_TIME_UNITS)(sampleIntervalTimeUnits), (C.uint)(maxPreTriggerSamples+maxPostTriggerSamples),
		(C.short)(boolToint16(autoStop)), (C.uint)(downSampleRatio), (C.uint)(overviewBufferSize))
	if stat == 0 {
		err = fmt.Errorf("RunStreaming failed")
	}
	sampleInterval = reqSampleInterval
	return
}

func ps2000RunBlock(handle int16, noOfPreTriggerSamples, noOfPostTriggerSamples int32,
	timeBase uint32, overSample int16, segmentIndex uint32, lpBlockReadyGoPar genericps.BlockReady,
	param interface{}) (timeIndisposedMs int32, err error) {
	nSamples := noOfPreTriggerSamples + noOfPostTriggerSamples
	slog.Debug("ps2000RunBlock", "handle", handle, "timeBase", timeBase)
	
	stat := C.ps2000_run_block((C.short)(handle), (C.int)(nSamples),
		(C.short)(timeBase), (C.short)(overSample), (*C.int)(&timeIndisposedMs))
	if stat == 0 {
		err = fmt.Errorf("RunBlock failed")
		return
	}
	
	// ps2000 doesn't support a callback for RunBlock, so we start a goroutine to poll.
	go func() {
		for {
			ready := C.ps2000_ready((C.short)(handle))
			if ready > 0 {
				if lpBlockReadyGoPar != nil {
					lpBlockReadyGoPar(handle, 1, param)
				}
				break
			}
			if ready < 0 {
				// Error occurred
				if lpBlockReadyGoPar != nil {
					lpBlockReadyGoPar(handle, 0, param)
				}
				break
			}
			time.Sleep(1 * time.Millisecond)
		}
	}()
	return
}

func ps2000Stop(handle int16) (err error) {
	slog.Debug("ps2000Stop", "handle", handle)
	C.ps2000_stop((C.short)(handle))
	return
}

func ps2000GetValues(handle int16, reqNoOfSamples uint32, bufferA, bufferB, bufferC, bufferD []int16) (noOfSamples uint32, overflow int16, err error) {
	slog.Debug("ps2000GetValues", "handle", handle, "reqNoOfSamples", reqNoOfSamples)
	
	pA := (*C.short)(nil)
	if len(bufferA) > 0 {
		pA = (*C.short)(&bufferA[0])
	}
	pB := (*C.short)(nil)
	if len(bufferB) > 0 {
		pB = (*C.short)(&bufferB[0])
	}
	pC := (*C.short)(nil)
	if len(bufferC) > 0 {
		pC = (*C.short)(&bufferC[0])
	}
	pD := (*C.short)(nil)
	if len(bufferD) > 0 {
		pD = (*C.short)(&bufferD[0])
	}
	
	stat := C.ps2000_get_values((C.short)(handle), pA, pB, pC, pD, (*C.short)(&overflow), (C.int)(reqNoOfSamples))
	if stat == 0 {
		err = fmt.Errorf("GetValues failed")
	}
	noOfSamples = uint32(stat)
	return
}

func ps2000SetSigGenBuiltIn(handle int16, offsetVoltage int32, pkToPK uint32, waveType WaveTypeEnum,
	startFrequency, stopFrequency, increment, dwellTime float32, sweepType SweepTypeEnum,
	operation ExtraOperations, shots, sweeps uint32, triggerType SigGenTrigType,
	triggerSource SigGenTrigSource, extInThreshold int16) (err error) {
	slog.Debug("ps2000SetSigGenBuiltIn")
	stat := C.ps2000_set_sig_gen_built_in((C.short)(handle), (C.int)(offsetVoltage),
		(C.uint)(pkToPK), (C.PS2000_WAVE_TYPE)(waveType), (C.float)(startFrequency),
		(C.float)(stopFrequency), (C.float)(increment), (C.float)(dwellTime),
		(C.PS2000_SWEEP_TYPE)(sweepType), (C.uint)(sweeps))
	if stat == 0 {
		err = fmt.Errorf("SetSigGenBuiltIn failed")
	}
	return
}

func ps2000IsReady(handle int16) (ready int16, err error) {
	slog.Debug("ps2000IsReady", "handle", handle)
	r := C.ps2000_ready((C.short)(handle))
	ready = int16(r)
	return
}
