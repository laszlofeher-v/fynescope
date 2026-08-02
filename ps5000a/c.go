//go:build !demo && ps5000a

package ps5000a

// #cgo CFLAGS: -g -Wall -I/opt/picoscope/include/libps5000a
// #cgo LDFLAGS: -L/opt/picoscope/lib/ -lps5000a
// #include <stdlib.h>
// #include "/opt/picoscope/include/libps5000a/PicoStatus.h"
// #include "/opt/picoscope/include/libps5000a/ps5000aApi.h"
/*
// Forward declarations
int ps5000aLpBlockReady(int16_t handle, PICO_STATUS status, void * pParameter);
int ps5000aLpDataReady(int16_t handle, PICO_STATUS status, uint32_t noOfSamples,
				int16_t overflow, void * pParameter);
int ps5000aLpStreamingReady(int16_t handle, int32_t noOfSamples, uint32_t startIndex,
                int16_t overflow, uint32_t triggerAt, int16_t triggered,
                int16_t autoStop, void * pParameter);
*/
import "C"

import (
	"fmt"
	"fynescope/genericps"
	"fynescope/psc"
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
	scopeHandler.Id = "ps5000a"
	genericps.Register(scopeHandler)
}

func boolToint16(b bool) int16 {
	if b {
		return int16(1)
	}
	return int16(0)
}

func enumerateUnits(bufferLen int16) (count int16, serials string, serialLth int16, err error) {
	c := make(chan struct{}, 1)
	go func() {
		var cstrPtr *C.schar
		cstrPtr = (*C.schar)(C.malloc(C.sizeof_schar * (C.ulong)(bufferLen)))
		defer C.free(unsafe.Pointer(cstrPtr))
		serialLth = bufferLen
		slog.Debug("ps5000aEnumerateUnits", "bufferLen", bufferLen)
		stat := C.ps5000aEnumerateUnits((*C.short)(&count), cstrPtr, (*C.short)(&serialLth))
		if stat != C.PICO_OK {
			err = fmt.Errorf("EnumerateUnits:  %s", psc.StatStr(int(stat)))
			c <- struct{}{}
			return
		}
		b := C.GoBytes(unsafe.Pointer(cstrPtr), (C.int)(serialLth-1))
		serials = string(b)
		c <- struct{}{}
	}()
	select {
	case res := <-c:
		fmt.Println(res)
	case <-time.After(10 * time.Second):
		fmt.Errorf("EnumerateUnits:timeout")
	}
	return
}

func openUnit(serial string) (handle int16, err error) {
	var p *C.schar
	sLength := len(serial)
	if sLength > 0 {
		p = (*C.schar)(C.CString(serial))
		defer C.free(unsafe.Pointer(p))
	}
	slog.Debug("ps5000aOpenUnit", "serial", serial)
	stat := C.ps5000aOpenUnit((*C.short)(&handle), (*C.schar)(p), C.PS5000A_DR_8BIT)
	if stat != C.PICO_OK {
		err = fmt.Errorf("OpenUnit:  %s", psc.StatStr(int(stat)))
		return
	}
	return
}

func openUnitAsync(serial string) (status int16, err error) {
	var p *C.schar
	sLength := len(serial)
	if sLength > 0 {
		p = (*C.schar)(C.CString(serial))
		defer C.free(unsafe.Pointer(p))
	}
	slog.Debug("ps5000aOpenUnitAsync", "serial", serial)
	stat := C.ps5000aOpenUnitAsync((*C.short)(&status), (*C.schar)(p), C.PS5000A_DR_8BIT)
	if stat != C.PICO_OK {
		err = fmt.Errorf("OpenUnitAsync:  %s", psc.StatStr(int(stat)))
		return
	}
	return
}

func openUnitProgress() (handle int16, progressPercent int16, complete int16, err error) {
	stat := C.ps5000aOpenUnitProgress((*C.short)(&handle),
		(*C.short)(&progressPercent), (*C.short)(&complete))
	if stat != C.PICO_OK {
		err = fmt.Errorf("OpenUnitProgress:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps5000aCloseUnit(handle int16) (err error) {
	slog.Debug("ps5000aCloseUnit", "handle", handle)
	stat := C.ps5000aCloseUnit((C.short)(handle))
	if stat != C.PICO_OK {
		err = fmt.Errorf("CloseUnit:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps5000aGetUnitInfo(handle int16, info PicoInfo) (infoString string, err error) {
	const listLen = 4096
	var cstrPtr *C.schar
	cstrPtr = (*C.schar)(C.malloc(C.sizeof_schar * listLen))
	defer C.free(unsafe.Pointer(cstrPtr))
	requiredSize := int16(listLen)
	slog.Debug("ps5000aGetUnitInfo", "handle", handle, "info", info)
	stat := C.ps5000aGetUnitInfo((C.short)(handle), cstrPtr, (C.short)(requiredSize),
		(*C.short)(&requiredSize), (C.PICO_INFO)(info))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetUnitInfo:  %s", psc.StatStr(int(stat)))
	}
	if requiredSize == 0 {
		infoString = "No answer from ps5000aGetUnitInfo "
		return
	}
	b := C.GoBytes(unsafe.Pointer(cstrPtr), (C.int)(requiredSize-1))
	infoString = string(b)
	return
}

func ps5000aFlashLed(handle int16, start int16) (err error) {
	slog.Debug("ps5000aFlashLed", "handle", handle, "start", start)
	stat := C.ps5000aFlashLed((C.short)(handle), (C.short)(start))
	if stat != C.PICO_OK {
		err = fmt.Errorf("FlashLed:  %s", psc.StatStr(int(stat)))
	}
	return
}

var regLpDataReadyGo DataReady // registered go callback function

//export ps5000aLpDataReadyGo
func ps5000aLpDataReadyGo(handle int16, status int, noOfSamples uint32, overflow int16, param interface{}) {
	if regLpDataReadyGo != nil {
		regLpDataReadyGo(handle, status, noOfSamples, overflow, param) // call registered go callback function
	}
	return
}

func ps5000aGetValuesAsync(handle int16, startIndex, noOfSamples, downSampleRatio uint32,
	downSampleRatioMode RatioMode, lpDataReadyGoPar DataReady, segmentIndex uint32,
	param interface{}) (err error) {
	regLpDataReadyGo = lpDataReadyGoPar
	slog.Debug("ps5000aGetValuesAsync", "handle", handle, "startIndex", startIndex, "noOfSamples", noOfSamples, "downSampleRatio", downSampleRatio, "downSampleRatioMode", downSampleRatioMode, "lpDataReadyGoPar", lpDataReadyGoPar, "segmentIndex", segmentIndex, "param", param)
	stat := C.ps5000aGetValuesAsync((C.short)(handle),
		(C.uint)(startIndex),
		(C.uint)(noOfSamples),
		(C.uint)(downSampleRatio),
		(C.PS5000A_RATIO_MODE)(downSampleRatioMode),
		(C.uint)(segmentIndex),
		(C.ps5000aLpDataReady), // C callback function in callbacks.go
		unsafe.Pointer(&param))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetValuesAsync:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps5000aGetValues(handle int16, startIndex, reqNoOfSamples, downSampleRatio uint32,
	downSampleRatioMode RatioMode, segmentIndex uint32) (noOfSamples uint32, overflow int16, err error) {
	slog.Debug("ps5000aGetValues", "handle", handle, "startIndex", startIndex, "reqNoOfSamples", reqNoOfSamples, "downSampleRatio", downSampleRatio, "downSampleRatioMode", downSampleRatioMode, "segmentIndex", segmentIndex)
	stat := C.ps5000aGetValues((C.short)(handle),
		(C.uint)(startIndex),
		(*C.uint)(&reqNoOfSamples),
		(C.uint)(downSampleRatio),
		(C.PS5000A_RATIO_MODE)(downSampleRatioMode),
		(C.uint)(segmentIndex),
		(*C.short)(&overflow))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetValues:  %s", psc.StatStr(int(stat)))
	}
	noOfSamples = reqNoOfSamples
	return
}

func ps5000aGetValuesBulk(handle int16, reqNoOfSamples uint32, fromSegmentIndex, toSegmentIndex, downSampleRatio uint32,
	downSampleRatioMode RatioMode, overflow []int16) (noSamples uint32, err error) {
	slog.Debug("ps5000aGetValuesBulk", "handle", handle, "reqNoOfSamples", reqNoOfSamples, "fromSegmentIndex", fromSegmentIndex, "toSegmentIndex", toSegmentIndex, "downSampleRatio", downSampleRatio, "downSampleRatioMode", downSampleRatioMode, "overflow", overflow)
	stat := C.ps5000aGetValuesBulk((C.short)(handle),
		(*C.uint)(&reqNoOfSamples),
		(C.uint)(fromSegmentIndex),
		(C.uint)(toSegmentIndex),
		(C.uint)(downSampleRatio),
		(C.PS5000A_RATIO_MODE)(downSampleRatioMode),
		(*C.short)(&overflow[0]))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetValuesBulk:  %s", psc.StatStr(int(stat)))
	}
	noSamples = reqNoOfSamples
	return
}

func ps5000aGetValuesOverlapped(handle int16, startIndex, reqNoOfSamples, downSampleRatio uint32,
	downSampleRatioMode RatioMode, segmentIndex uint32, overflow []int16) (noSamples uint32, err error) {
	slog.Debug("ps5000aGetValuesOverlapped", "handle", handle, "startIndex", startIndex, "reqNoOfSamples", reqNoOfSamples, "downSampleRatio", downSampleRatio, "downSampleRatioMode", downSampleRatioMode, "segmentIndex", segmentIndex, "overflow", overflow)
	stat := C.ps5000aGetValuesOverlapped((C.short)(handle),
		(C.uint)(startIndex),
		(*C.uint)(&reqNoOfSamples),
		(C.uint)(downSampleRatio),
		(C.PS5000A_RATIO_MODE)(downSampleRatioMode),
		(C.uint)(segmentIndex),
		(*C.short)(&overflow[0]))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetValuesOverlapped:  %s", psc.StatStr(int(stat)))
	}
	noSamples = reqNoOfSamples
	return
}

func ps5000aGetValuesOverlappedBulk(handle int16, startIndex, reqNoOfSamples, downSampleRatio uint32,
	downSampleRatioMode RatioMode, fromSegmentIndex, toSegmentIndex uint32, overflow []int16) (noSamples uint32, err error) {
	slog.Debug("ps5000aGetValuesOverlappedBulk", "handle", handle, "startIndex", startIndex, "reqNoOfSamples", reqNoOfSamples, "downSampleRatio", downSampleRatio, "downSampleRatioMode", downSampleRatioMode, "fromSegmentIndex", fromSegmentIndex, "toSegmentIndex", toSegmentIndex, "overflow", overflow)
	stat := C.ps5000aGetValuesOverlappedBulk((C.short)(handle),
		(C.uint)(startIndex),
		(*C.uint)(&reqNoOfSamples),
		(C.uint)(downSampleRatio),
		(C.PS5000A_RATIO_MODE)(downSampleRatioMode),
		(C.uint)(fromSegmentIndex),
		(C.uint)(toSegmentIndex),
		(*C.short)(&overflow[0]))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetValuesOverlappedBulk:  %s", psc.StatStr(int(stat)))
	}
	noSamples = reqNoOfSamples
	return
}

func ps5000aGetAnalogueOffset(handle int16, voltageRange int, coupling Coupling) (maximumVoltage, minimumVoltage float32, err error) {
	slog.Debug("ps5000aGetAnalogueOffset", "handle", handle, "voltageRange", voltageRange, "coupling", coupling)
	stat := C.ps5000aGetAnalogueOffset((C.short)(handle),
		(C.PS5000A_RANGE)(voltageRange),
		(C.PS5000A_COUPLING)(coupling), (*C.float)(&maximumVoltage), (*C.float)(&minimumVoltage))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetAnalogueOffset:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps5000aGetChannelInformation(handle int16, info int16, probe int32, ranges []int32, channels ChannelId) (lengthOfRanges int32, err error) {
	lengthOfRanges = int32(len(ranges))
	slog.Debug("ps5000aGetChannelInformation", "handle", handle, "info", info, "probe", probe, "ranges", ranges, "channels", channels)
	stat := C.ps5000aGetChannelInformation((C.short)(handle), (C.PS5000A_CHANNEL_INFO)(info),
		(C.int)(probe), (*C.int)(&ranges[0]), (*C.int)(&lengthOfRanges), (C.int)(channels))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetChannelInformation:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps5000aGetMaxDownSampleRatio(handle int16, noOfUnaggregatedSamples uint32, downSampleRatioMode RatioMode, segmentIndex int32) (maxDownSampleRatio uint32, err error) {
	slog.Debug("ps5000aGetMaxDownSampleRatio", "handle", handle, "noOfUnaggregatedSamples", noOfUnaggregatedSamples, "downSampleRatioMode", downSampleRatioMode, "segmentIndex", segmentIndex)
	stat := C.ps5000aGetMaxDownSampleRatio((C.short)(handle), (C.uint)(noOfUnaggregatedSamples),
		(*C.uint)(&maxDownSampleRatio), (C.PS5000A_RATIO_MODE)(downSampleRatioMode), (C.uint)(segmentIndex))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetMaxDownSampleRatio:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps5000aGetMaxSegments(handle int16) (maxSegments uint32, err error) {
	slog.Debug("ps5000aGetMaxSegments", "handle", handle)
	stat := C.ps5000aGetMaxSegments((C.short)(handle), (*C.uint)(&maxSegments))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetMaxSegments:  %s", psc.StatStr(int(stat)))
	}
	return
}

// ps5000aChangePowerSource
// ps5000aCurrentPowerSource
func ps5000aGetNumOfCaptures(handle int16) (nCaptures uint32, err error) {
	slog.Debug("ps5000aGetNoOfCaptures", "handle", handle)
	stat := C.ps5000aGetNoOfCaptures((C.short)(handle), (*C.uint)(&nCaptures))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetNoOfCaptures:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps5000aGetNumOfProcessedCaptures(handle int16) (nCaptures uint32, err error) {
	slog.Debug("ps5000aGetNoOfProcessedCaptures", "handle", handle)
	stat := C.ps5000aGetNoOfProcessedCaptures((C.short)(handle), (*C.uint)(&nCaptures))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetNoOfProcessedCaptures:  %s", psc.StatStr(int(stat)))
	}
	return
}

var regLpStreamingReadyGo StreamingReady // registered go callback function

//export ps5000aLpStreamingReadyGo
func ps5000aLpStreamingReadyGo(handle int16, noOfSamples int32, startIndex uint32, overflow int16,
	triggeredAt uint32, triggered, autoStop int16, param interface{}) {
	if regLpStreamingReadyGo != nil {
		regLpStreamingReadyGo(handle, noOfSamples, startIndex, overflow, triggeredAt, autoStop, triggered, param) // call registered go callback function
	}
	return
}

func ps5000aGetStreamingLatestValues(handle int16, lpStreamingReadyGoPar StreamingReady, param interface{}) (err error) {
	regLpStreamingReadyGo = lpStreamingReadyGoPar
	slog.Debug("ps5000aGetStreamingLatestValues", "handle", handle, "lpStreamingReadyGoPar", lpStreamingReadyGoPar, "param", param)
	stat := C.ps5000aGetStreamingLatestValues((C.short)(handle),
		(C.ps5000aStreamingReady)(C.ps5000aLpStreamingReady), // C callback function in callbacks.go
		unsafe.Pointer(&param))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetStreamingLatestValues:  %s", psc.StatStr(int(stat)))
	}
	return
}

// ps5000aCheckForUpdate
// ps5000aStartFirmwareUpdate
func ps5000aGetTimebase(handle int16, timeBase uint32, noOfSamples int32, overSample int16, segmentIndex uint32) (timeIntervalNanoseconds, maxSamples int32, err error) {
	slog.Debug("ps5000aGetTimebase", "handle", handle, "timeBase", timeBase, "noOfSamples", noOfSamples, "overSample", overSample, "segmentIndex", segmentIndex)
	stat := C.ps5000aGetTimebase((C.short)(handle), (C.uint)(timeBase), (C.int)(noOfSamples),
		(*C.int)(&timeIntervalNanoseconds),
		(*C.int)(&maxSamples), (C.uint)(segmentIndex))
	if stat != C.PICO_OK {
		slog.Error("GetTimebase", "noOfSamples", noOfSamples, "stat", psc.StatStr(int(stat)))
		err = fmt.Errorf("GetTimebase:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps5000aGetTimebase2(handle int16, timeBase uint32, numOfSamples int32,
	overSample int16, segmentIndex uint32) (timeIntervalNanoseconds float32, maxSamples int32, err error) {
	slog.Debug("ps5000aGetTimebase2", "handle", handle, "timeBase", timeBase, "numOfSamples", numOfSamples, "overSample", overSample, "segmentIndex", segmentIndex)
	stat := C.ps5000aGetTimebase2((C.short)(handle), (C.uint)(timeBase), (C.int)(numOfSamples),
		(*C.float)(&timeIntervalNanoseconds),
		(*C.int)(&maxSamples), (C.uint)(segmentIndex))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetTimebase2:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps5000aSetChannel(handle int16, channel ChannelId, enabled bool, couplingType Coupling, voltageRange RangeEnum, analogOffset float32) (err error) {
	slog.Debug("ps5000aSetChannel", "handle", handle, "channel", channel, "enabled", enabled, "couplingType", couplingType, "voltageRange", voltageRange, "analogOffset", analogOffset)
	stat := C.ps5000aSetChannel((C.short)(handle), (C.PS5000A_CHANNEL)(channel), (C.short)(boolToint16(enabled)),
		(C.PS5000A_COUPLING)(couplingType), (C.PS5000A_RANGE)(voltageRange),
		(C.float)(analogOffset))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetChannel:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps5000aMaximumValue(handle int16) (value int32, err error) {
	slog.Debug("ps5000aMaximumValue", "handle", handle)
	var val int16
	stat := C.ps5000aMaximumValue((C.short)(handle), (*C.short)(&val))
	value = int32(val)
	if stat != C.PICO_OK {
		err = fmt.Errorf("MaximumValue:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps5000aMinimumValue(handle int16) (value int32, err error) {
	slog.Debug("ps5000aMinimumValue", "handle", handle)
	var val int16
	stat := C.ps5000aMinimumValue((C.short)(handle), (*C.short)(&val))
	value = int32(val)
	if stat != C.PICO_OK {
		err = fmt.Errorf("MinimumValue:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps5000aSetSimpleTrigger(handle int16, enable bool, source ChannelId, threshold int16,
	direction ThresholdDirection, delay uint32, autoTriggerMs int16) (err error) {
	slog.Debug("ps5000aSetSimpleTrigger", "handle", handle, "enable", enable, "src", source, "threshold", threshold, "direction", direction, "delay", delay, "autoTriggerMs", autoTriggerMs)
	stat := C.ps5000aSetSimpleTrigger((C.short)(handle), (C.short)(boolToint16(enable)),
		(C.PS5000A_CHANNEL)(source), (C.short)(threshold),
		(C.PS5000A_THRESHOLD_DIRECTION)(direction), (C.uint)(delay),
		(C.short)(autoTriggerMs))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetSimpleTrigger:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps5000aSetDataBuffer(handle int16, ch ChannelId, bufferIn []int16, segmentIndex uint32,
	mode RatioMode) (err error) {

	slog.Debug("ps5000aSetDataBuffer", "handle", handle, "ch", ch, "segmentIndex", segmentIndex, "mode", mode)
	stat := C.ps5000aSetDataBuffer((C.short)(handle), (C.PS5000A_CHANNEL)(ch), (*C.short)(&bufferIn[0]),
		(C.int)(len(bufferIn)), (C.uint)(segmentIndex),
		(C.PS5000A_RATIO_MODE)(mode))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetDataBuffer:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps5000aSetDataBuffers(handle int16, ch ChannelId, bufferMax, bufferMin []int16, segmentIndex uint32, mode RatioMode) (err error) {
	slog.Debug("ps5000aSetDataBuffers", "handle", handle, "ch", ch, "segmentIndex", segmentIndex, "mode", mode)
	stat := C.ps5000aSetDataBuffers((C.short)(handle), (C.PS5000A_CHANNEL)(ch), (*C.short)(&bufferMax[0]),
		(*C.short)(&bufferMin[0]), (C.int)(len(bufferMax)), (C.uint)(segmentIndex),
		(C.PS5000A_RATIO_MODE)(mode))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetDataBuffers:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps5000aSetUnscaledDataBuffers(handle int16, ch ChannelId, bufferMax, bufferMin []int16, segmentIndex uint32, mode RatioMode) (err error) {
	slog.Debug("ps5000aSetUnscaledDataBuffers", "handle", handle, "ch", ch, "segmentIndex", segmentIndex, "mode", mode)
	stat := C.ps5000aSetDataBuffers((C.short)(handle), (C.PS5000A_CHANNEL)(ch), (*C.short)(&bufferMax[0]),
		(*C.short)(&bufferMin[0]), (C.int)(len(bufferMax)), (C.uint)(segmentIndex),
		(C.PS5000A_RATIO_MODE)(mode))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetUnscaledDataBuffers:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps5000aSetEtsTimeBuffer(handle int16, buffer []int64) (err error) {
	slog.Debug("ps5000aSetEtsTimeBuffer", "handle", handle, "buffer", buffer)
	stat := C.ps5000aSetEtsTimeBuffer((C.short)(handle), (*C.long)(&buffer[0]),
		(C.int)(len(buffer)))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetEtsTimeBuffer:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps5000aSetEtsTimeBuffers(handle int16, timeUpper, timeLower []uint32) (err error) {
	slog.Debug("ps5000aSetEtsTimeBuffers", "handle", handle, "timeUpper", timeUpper, "timeLower", timeLower)
	stat := C.ps5000aSetEtsTimeBuffers((C.short)(handle), (*C.uint)(&timeUpper[0]),
		(*C.uint)(&timeLower[0]), (C.int)(len(timeUpper)))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetEtsTimeBuffer:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps5000aSetEts(handle int16, mode EtsMode, etsCycles int16, etsInterLeave int16) (sampleTimePicoseconds int32, err error) {
	slog.Debug("ps5000aSetEts", "handle", handle, "mode", mode, "etsCycles", etsCycles, "etsInterLeave", etsInterLeave)
	stat := C.ps5000aSetEts((C.short)(handle), (C.PS5000A_ETS_MODE)(mode),
		(C.short)(etsCycles), (C.short)(etsInterLeave), (*C.int)(&sampleTimePicoseconds))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetEts:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps5000aRunStreaming(handle int16, reqSampleInterval uint32, sampleIntervalTimeUnits TimeUnits,
	maxPreTriggerSamples, maxPostTriggerSamples uint32,
	autoStop bool, downSampleRatio uint32, downSampleRatioMode RatioMode,
	overviewBufferSize uint32) (sampleInterval uint32, err error) {
	slog.Debug("ps5000aRunStreaming", "handle", handle, "reqSampleInterval", reqSampleInterval, "sampleIntervalTimeUnits", sampleIntervalTimeUnits, "maxPreTriggerSamples", maxPreTriggerSamples, "maxPostTriggerSamples", maxPostTriggerSamples, "autoStop", autoStop, "downSampleRatio", downSampleRatio, "downSampleRatioMode", downSampleRatioMode, "overviewBufferSize", overviewBufferSize)
	stat := C.ps5000aRunStreaming((C.short)(handle), (*C.uint)(&reqSampleInterval),
		(C.PS5000A_TIME_UNITS)(sampleIntervalTimeUnits), (C.uint)(maxPreTriggerSamples),
		(C.uint)(maxPostTriggerSamples), (C.short)(boolToint16(autoStop)), (C.uint)(downSampleRatio),
		(C.PS5000A_RATIO_MODE)(downSampleRatioMode), (C.uint)(overviewBufferSize))
	if stat != C.PICO_OK {
		err = fmt.Errorf("RunStreaming:  %s", psc.StatStr(int(stat)))
	}
	sampleInterval = reqSampleInterval
	return
}

var regLpBlockReadyGo BlockReady // registered go callback function

//export ps5000aLpBlockReadyGo
func ps5000aLpBlockReadyGo(handle int16, status int, noOfSamples uint32, overflow int16, param interface{}) {
	if regLpBlockReadyGo != nil {
		regLpBlockReadyGo(handle, status, param) // call registered go callback function
	}
	return
}

func ps5000aRunBlock(handle int16, noOfPreTriggerSamples, noOfPostTriggerSamples int32,
	timeBase uint32, overSample int16, segmentIndex uint32, lpBlockReadyGoPar BlockReady,
	param interface{}) (timeIndisposedMs int32, err error) {
	regLpBlockReadyGo = lpBlockReadyGoPar
	slog.Debug("ps5000aRunBlock", "handle", handle, "noOfPreTriggerSamples", noOfPreTriggerSamples, "noOfPostTriggerSamples", noOfPostTriggerSamples, "timeBase", timeBase, "overSample", overSample, "segmentIndex", segmentIndex, "lpBlockReadyGoPar", lpBlockReadyGoPar, "param", param)
	stat := C.ps5000aRunBlock((C.short)(handle), (C.int)(noOfPreTriggerSamples),
		(C.int)(noOfPostTriggerSamples), (C.uint)(timeBase),
		(*C.int)(&timeIndisposedMs), (C.uint)(segmentIndex), (C.ps5000aBlockReady)(C.ps5000aLpBlockReady),
		unsafe.Pointer(&param))
	if stat != C.PICO_OK {
		err = fmt.Errorf("RunBlock:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps5000aSetTriggerChannelProperties(handle int16, channelProperties []TriggerChannelProperties, auxOutputEnable bool,
	autoTriggerMs int32) (err error) {
	var cTriggerChannelProperties []C.PS5000A_TRIGGER_CHANNEL_PROPERTIES
	if len(channelProperties) > 0 {
		cTriggerChannelProperties = make([]C.PS5000A_TRIGGER_CHANNEL_PROPERTIES, len(channelProperties))
		for i := range channelProperties {
			cTriggerChannelProperties[i].channel = (C.PS5000A_CHANNEL)(channelProperties[i].Channel)
			cTriggerChannelProperties[i].thresholdLowerHysteresis = (C.ushort)(channelProperties[i].ThresholdLowerHysteresis)
			cTriggerChannelProperties[i].thresholdLower = (C.short)(channelProperties[i].ThresholdLower)
			cTriggerChannelProperties[i].thresholdUpperHysteresis = (C.ushort)(channelProperties[i].ThresholdUpperHysteresis)
			cTriggerChannelProperties[i].thresholdUpper = (C.short)(channelProperties[i].ThresholdUpper)
		}
	}
	pcTriggerChannelProperties := (*C.PS5000A_TRIGGER_CHANNEL_PROPERTIES)(nil)
	if len(channelProperties) > 0 {
		pcTriggerChannelProperties = &cTriggerChannelProperties[0]
	}
	slog.Debug("ps5000aSetTriggerChannelProperties", "handle", handle, "channelProperties", channelProperties, "auxOutputEnable", auxOutputEnable, "autoTriggerMs", autoTriggerMs)
	stat := C.ps5000aSetTriggerChannelProperties((C.short)(handle),
		(*C.PS5000A_TRIGGER_CHANNEL_PROPERTIES)(pcTriggerChannelProperties),
		(C.short)(len(channelProperties)), (C.short)(boolToint16(auxOutputEnable)), (C.int)(autoTriggerMs))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetTriggerChannelProperties:  %s", psc.StatStr(int(stat)))
	}
	return

}

func ps5000aSetTriggerChannelConditions(handle int16, triggerConditions []TriggerConditions) (err error) {
	cConditions := make([]C.PS5000A_CONDITION, 0, len(triggerConditions)*5)
	for _, tc := range triggerConditions {
		if tc.ChannelA != CondDontCare {
			cConditions = append(cConditions, C.PS5000A_CONDITION{source: C.PS5000A_CHANNEL_A, condition: C.PS5000A_TRIGGER_STATE(tc.ChannelA)})
		}
		if tc.ChannelB != CondDontCare {
			cConditions = append(cConditions, C.PS5000A_CONDITION{source: C.PS5000A_CHANNEL_B, condition: C.PS5000A_TRIGGER_STATE(tc.ChannelB)})
		}
		if tc.ChannelC != CondDontCare {
			cConditions = append(cConditions, C.PS5000A_CONDITION{source: C.PS5000A_CHANNEL_C, condition: C.PS5000A_TRIGGER_STATE(tc.ChannelC)})
		}
		if tc.ChannelD != CondDontCare {
			cConditions = append(cConditions, C.PS5000A_CONDITION{source: C.PS5000A_CHANNEL_D, condition: C.PS5000A_TRIGGER_STATE(tc.ChannelD)})
		}
		if tc.External != CondDontCare {
			cConditions = append(cConditions, C.PS5000A_CONDITION{source: C.PS5000A_EXTERNAL, condition: C.PS5000A_TRIGGER_STATE(tc.External)})
		}
	}
	pcConditions := (*C.PS5000A_CONDITION)(nil)
	if len(cConditions) > 0 {
		pcConditions = &cConditions[0]
	}
	slog.Debug("ps5000aSetTriggerChannelConditions", "handle", handle, "triggerConditions", triggerConditions)
	stat := C.ps5000aSetTriggerChannelConditionsV2((C.short)(handle),
		pcConditions, (C.short)(len(cConditions)), C.PS5000A_CLEAR|C.PS5000A_ADD)
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetTriggerChannelConditions: %s", psc.StatStr(int(stat)))
	}
	return
}

func ps5000aSetTriggerChannelDirections(handle int16, channelA, channelB, channelC, channelD, ext, aux ThresholdDirection) (err error) {
	slog.Debug("ps5000aSetTriggerChannelDirections", "handle", handle, "channelA", channelA, "channelB", channelB, "channelC", channelC, "channelD", channelD, "ext", ext, "aux", aux)
	stat := C.ps5000aSetTriggerChannelDirections((C.short)(handle),
		(C.PS5000A_THRESHOLD_DIRECTION)(channelA),
		(C.PS5000A_THRESHOLD_DIRECTION)(channelB),
		(C.PS5000A_THRESHOLD_DIRECTION)(channelC),
		(C.PS5000A_THRESHOLD_DIRECTION)(channelD),
		(C.PS5000A_THRESHOLD_DIRECTION)(ext),
		(C.PS5000A_THRESHOLD_DIRECTION)(aux))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetTriggerChannelDirections:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps5000aSetTriggerDelay(handle int16, delay uint32) (err error) {
	slog.Debug("ps5000aSetTriggerDelay", "handle", handle, "delay", delay)
	stat := C.ps5000aSetTriggerDelay((C.short)(handle), (C.uint)(delay))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetTriggerDelay:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps5000aSetPulseWidthQualifier(handle int16, conditions []PwqConditions, direction ThresholdDirection, lower, upper uint32,
	pwType PulseWidthType) (err error) {
	return fmt.Errorf("SetPulseWidthQualifier not supported on ps5000a")
}
func ps5000aSetTriggerDigitalPortProperties(handle int16, digitalDirections []DigitalChannelDirections) (err error) {
	return fmt.Errorf("SetTriggerDigitalPortProperties not supported on ps5000a")
}

func ps5000aStop(handle int16) (err error) {
	slog.Debug("ps5000aStop", "handle", handle)
	stat := C.ps5000aStop((C.short)(handle))
	if stat != C.PICO_OK {
		err = fmt.Errorf("Stop:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps5000aSetSigGenBuiltIn(handle int16, offsetVoltage int32, pkToPK uint32, waveType WaveTypeEnum,
	startFrequency, stopFrequency, increment, dwellTime float32, sweepType SweepTypeEnum,
	operation ExtraOperations, shots, sweeps uint32, triggerType SigGenTrigType,
	triggerSource SigGenTrigSource, extInThreshold int16) (err error) {
	slog.Debug("ps5000aSetSigGenBuiltIn", "handle", handle, "offsetVoltage", offsetVoltage, "pkToPK", pkToPK, "waveType", waveType, "startFrequency", startFrequency, "stopFrequency", stopFrequency, "increment", increment, "dwellTime", dwellTime, "sweepType", sweepType, "operation", operation, "shots", shots, "sweeps", sweeps, "triggerType", triggerType, "triggerSource", triggerSource, "extInThreshold", extInThreshold)
	stat := C.ps5000aSetSigGenBuiltIn((C.short)(handle), (C.int)(offsetVoltage),
		(C.uint)(pkToPK), (C.PS5000A_WAVE_TYPE)(waveType), (C.float)(startFrequency),
		(C.float)(stopFrequency), (C.float)(increment), (C.float)(dwellTime),
		(C.PS5000A_SWEEP_TYPE)(sweepType), (C.PS5000A_EXTRA_OPERATIONS)(operation),
		(C.uint)(shots), (C.uint)(sweeps), (C.PS5000A_SIGGEN_TRIG_TYPE)(triggerType),
		(C.PS5000A_SIGGEN_TRIG_SOURCE)(triggerSource), (C.short)(extInThreshold))

	if stat != C.PICO_OK {
		err = fmt.Errorf("SetSigGenBuiltIn:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps5000aSetSigGenBuiltInV2(handle int16, offsetVoltage int32, pkToPK uint32, waveType WaveTypeEnum,
	startFrequency, stopFrequency, increment, dwellTime float64, sweepType SweepTypeEnum,
	operation ExtraOperations, shots, sweeps uint32, triggerType SigGenTrigType,
	triggerSource SigGenTrigSource, extInThreshold int16) (err error) {
	slog.Debug("ps5000aSetSigGenBuiltInV2", "handle", handle, "offsetVoltage", offsetVoltage, "pkToPK", pkToPK, "waveType", waveType, "startFrequency", startFrequency, "stopFrequency", stopFrequency, "increment", increment, "dwellTime", dwellTime, "sweepType", sweepType, "operation", operation, "shots", shots, "sweeps", sweeps, "triggerType", triggerType, "triggerSource", triggerSource, "extInThreshold", extInThreshold)
	stat := C.ps5000aSetSigGenBuiltInV2((C.short)(handle), (C.int)(offsetVoltage),
		(C.uint)(pkToPK), (C.PS5000A_WAVE_TYPE)(waveType), (C.double)(startFrequency),
		(C.double)(stopFrequency), (C.double)(increment), (C.double)(dwellTime),
		(C.PS5000A_SWEEP_TYPE)(sweepType), (C.PS5000A_EXTRA_OPERATIONS)(operation),
		(C.uint)(shots), (C.uint)(sweeps), (C.PS5000A_SIGGEN_TRIG_TYPE)(triggerType),
		(C.PS5000A_SIGGEN_TRIG_SOURCE)(triggerSource), (C.short)(extInThreshold))

	if stat != C.PICO_OK {
		err = fmt.Errorf("SetSigGenBuiltIn:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps5000aSigGenFrequencyToPhase(handle int16, frequency float64, indexMode IndexMode, bufferLength uint32) (phase uint32, err error) {
	slog.Debug("ps5000aSigGenFrequencyToPhase", "handle", handle, "frequency", frequency, "indexMode", indexMode, "bufferLength", bufferLength)
	stat := C.ps5000aSigGenFrequencyToPhase((C.short)(handle), (C.double)(frequency),
		(C.PS5000A_INDEX_MODE)(indexMode), (C.uint)(bufferLength), (*C.uint)(&phase))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SigGenFrequencyToPhase:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps5000aSetNoCaptures(handle int16, nCaptures uint32) (err error) {
	slog.Debug("ps5000aSetNoOfCaptures", "handle", handle, "nCaptures", nCaptures)
	stat := C.ps5000aSetNoOfCaptures((C.short)(handle), (C.uint)(nCaptures))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetNoCaptures:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps5000aGetTriggerTimeOffset(handle int16, segmentIndex uint32) (timeUpper, timeLower uint32, timeUnits TimeUnits, err error) {
	slog.Debug("ps5000aGetTriggerTimeOffset", "handle", handle, "segmentIndex", segmentIndex)
	stat := C.ps5000aGetTriggerTimeOffset((C.short)(handle), (*C.uint)(&timeUpper),
		(*C.uint)(&timeLower), (*C.PS5000A_TIME_UNITS)(&timeUnits), (C.uint)(segmentIndex))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetTriggerTimeOffset:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps5000aGetTriggerTimeOffset64(handle int16, segmentIndex uint32) (time int64, timeUnits TimeUnits, err error) {
	slog.Debug("ps5000aGetTriggerTimeOffset64", "handle", handle, "segmentIndex", segmentIndex)
	stat := C.ps5000aGetTriggerTimeOffset64((C.short)(handle), (*C.long)(&time),
		(*C.PS5000A_TIME_UNITS)(&timeUnits), (C.uint)(segmentIndex))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetTriggerTimeOffset64:  %s", psc.StatStr(int(stat)))
	}
	return

}

func ps5000aGetValuesTriggerTimeOffsetBulk(handle int16, timesUpper, timesLower []uint32, timeUnits []TimeUnits,
	fromSegmentIndex, toSegmentIndex uint32) (err error) {
	slog.Debug("ps5000aGetValuesTriggerTimeOffsetBulk", "handle", handle, "timesUpper", timesUpper, "timesLower", timesLower, "timeUnits", timeUnits, "fromSegmentIndex", fromSegmentIndex, "toSegmentIndex", toSegmentIndex)
	stat := C.ps5000aGetValuesTriggerTimeOffsetBulk((C.short)(handle), (*C.uint)(&timesUpper[0]),
		(*C.uint)(&timesLower[0]), (*C.PS5000A_TIME_UNITS)(&timeUnits[0]), (C.uint)(fromSegmentIndex),
		(C.uint)(toSegmentIndex))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetValuesTriggerTimeOffsetBulk:  %s", psc.StatStr(int(stat)))
	}

	return
}

func ps5000aGetValuesTriggerTimeOffsetBulk64(handle int16, times []int64, timeUnits []TimeUnits,
	fromSegmentIndex, toSegmentIndex uint32) (err error) {
	slog.Debug("ps5000aGetValuesTriggerTimeOffsetBulk64", "handle", handle, "times", times, "timeUnits", timeUnits, "fromSegmentIndex", fromSegmentIndex, "toSegmentIndex", toSegmentIndex)
	stat := C.ps5000aGetValuesTriggerTimeOffsetBulk64((C.short)(handle), (*C.long)(&times[0]),
		(*C.PS5000A_TIME_UNITS)(&timeUnits[0]), (C.uint)(fromSegmentIndex),
		(C.uint)(toSegmentIndex))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetValuesTriggerTimeOffsetBulk64:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps5000aHoldOff(handle int16, holdOff uint64, holdOffType HoldOffType) (err error) {
	return fmt.Errorf("HoldOff not supported on ps5000a")
}

func ps5000aLsReady(handle int16) (ready int16, err error) {
	slog.Debug("ps5000aLsReady", "handle", handle)
	stat := C.ps5000aIsReady((C.short)(handle), (*C.short)(&ready))
	if stat != C.PICO_OK {
		err = fmt.Errorf("LsReady:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps5000aTriggerOrPulseWidthQualifierEnabled(handle int16) (triggerEnabled, pulseWidthQualifierEnabledint16 int16, err error) {
	slog.Debug("ps5000aTriggerOrPulseWidthQualifierEnabled", "handle", handle)
	stat := C.ps5000aIsTriggerOrPulseWidthQualifierEnabled((C.short)(handle),
		(*C.short)(&triggerEnabled), (*C.short)(&pulseWidthQualifierEnabledint16))
	if stat != C.PICO_OK {
		err = fmt.Errorf("TriggerOrPulseWidthQualifierEnabled:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps5000aMemorySegments(handle int16, nSegments uint64) (nMaxSamples int64, err error) {
	slog.Debug("ps5000aMemorySegments", "handle", handle, "nSegments", nSegments)
	var maxSamples C.int
	stat := C.ps5000aMemorySegments((C.short)(handle), (C.uint)(nSegments), &maxSamples)
	if stat != C.PICO_OK {
		err = fmt.Errorf("MemorySegments:  %s", psc.StatStr(int(stat)))
	}
	nMaxSamples = int64(maxSamples)
	return
}

func ps5000aNoOfStreamingValues(handle int16) (noOfValues uint32, err error) {
	slog.Debug("ps5000aNoOfStreamingValues", "handle", handle)
	stat := C.ps5000aNoOfStreamingValues((C.short)(handle),
		(*C.uint)(&noOfValues))
	if stat != C.PICO_OK {
		err = fmt.Errorf("NoOfStreamingValues:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps5000aPingUnit(handle int16) (err error) {
	slog.Debug("ps5000aPingUnit", "handle", handle)
	stat := C.ps5000aPingUnit((C.short)(handle))
	if stat != C.PICO_OK {
		err = fmt.Errorf("PingUnit:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps5000aQueryOutputEdgeDetect(handle int16) (state int16, err error) {
	slog.Debug("ps5000aQueryOutputEdgeDetect", "handle", handle)
	stat := C.ps5000aQueryOutputEdgeDetect((C.short)(handle),
		(*C.short)(&state))
	if stat != C.PICO_OK {
		err = fmt.Errorf("QueryOutputEdgeDetect:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps5000aSetDigitalPort(handle int16, port DigitalPort, enabled bool, logiclevel int16) (err error) {
	return fmt.Errorf("SetDigitalPort not supported on ps5000a")
}

func ps5000aSetOutputEdgeDetect(handle int16, state int16) (err error) {
	slog.Debug("ps5000aSetOutputEdgeDetect", "handle", handle, "state", state)
	stat := C.ps5000aSetOutputEdgeDetect((C.short)(handle), (C.short)(state))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetOutputEdgeDetect:  %s", psc.StatStr(int(stat)))
	}
	return
}

// ps5000aGetScalingValues
func ps5000aSetPulseWidthDigitalPortProperties(handle int16, digitalDirections []DigitalChannelDirections) (err error) {
	return fmt.Errorf("SetPulseWidthDigitalPortProperties not supported on ps5000a")
}

func ps5000aSetSigGenArbitrary(handle int16, offsetVoltage int32, pkToPK uint32,
	startDeltaPhase, stopDeltaPhase, deltaPhaseIncrement, dwellCount uint32,
	arbitraryWaveform []int16, sweepType SweepTypeEnum, operation ExtraOperations,
	indexMode IndexMode, shots, sweeps uint32, triggerType SigGenTrigType,
	triggerSource SigGenTrigSource, extInThreshold int16) (err error) {
	slog.Debug("ps5000aSetSigGenArbitrary", "handle", handle, "offsetVoltage", offsetVoltage, "pkToPK", pkToPK, "startDeltaPhase", startDeltaPhase, "stopDeltaPhase", stopDeltaPhase, "deltaPhaseIncrement", deltaPhaseIncrement, "dwellCount", dwellCount, "arbitraryWaveform", arbitraryWaveform, "sweepType", sweepType, "operation", operation, "indexMode", indexMode, "shots", shots, "sweeps", sweeps, "triggerType", triggerType, "triggerSource", triggerSource, "extInThreshold", extInThreshold)
	stat := C.ps5000aSetSigGenArbitrary((C.short)(handle), (C.int)(offsetVoltage),
		(C.uint)(pkToPK), (C.uint)(startDeltaPhase), (C.uint)(stopDeltaPhase),
		(C.uint32_t)(deltaPhaseIncrement), (C.uint32_t)(dwellCount),
		(*C.short)(&arbitraryWaveform[0]), (C.int32_t)(len(arbitraryWaveform)),
		(C.PS5000A_SWEEP_TYPE)(sweepType), (C.PS5000A_EXTRA_OPERATIONS)(operation),
		(C.PS5000A_INDEX_MODE)(indexMode),
		(C.uint)(shots), (C.uint)(sweeps), (C.PS5000A_SIGGEN_TRIG_TYPE)(triggerType),
		(C.PS5000A_SIGGEN_TRIG_SOURCE)(triggerSource), (C.short)(extInThreshold))

	if stat != C.PICO_OK {
		err = fmt.Errorf("SetSigGenArbitrary:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps5000aSetSigGenPropertiesArbitrary(handle int16, offsetVoltage int32,
	startDeltaPhase, stopDeltaPhase, deltaPhaseIncrement, dwellCount uint32,
	sweepType SweepTypeEnum, operation ExtraOperations,
	indexMode IndexMode, shots, sweeps uint32, triggerType SigGenTrigType,
	triggerSource SigGenTrigSource, extInThreshold int16) (err error) {
	slog.Debug("ps5000aSetSigGenPropertiesArbitrary", "handle", handle, "offsetVoltage", offsetVoltage, "startDeltaPhase", startDeltaPhase, "stopDeltaPhase", stopDeltaPhase, "deltaPhaseIncrement", deltaPhaseIncrement, "dwellCount", dwellCount, "sweepType", sweepType, "operation", operation, "indexMode", indexMode, "shots", shots, "sweeps", sweeps, "triggerType", triggerType, "triggerSource", triggerSource, "extInThreshold", extInThreshold)
	stat := C.ps5000aSetSigGenPropertiesArbitrary((C.short)(handle),
		(C.uint)(startDeltaPhase), (C.uint)(stopDeltaPhase),
		(C.uint32_t)(deltaPhaseIncrement), (C.uint32_t)(dwellCount),
		(C.PS5000A_SWEEP_TYPE)(sweepType),
		(C.uint)(shots), (C.uint)(sweeps), (C.PS5000A_SIGGEN_TRIG_TYPE)(triggerType),
		(C.PS5000A_SIGGEN_TRIG_SOURCE)(triggerSource), (C.short)(extInThreshold))

	if stat != C.PICO_OK {
		err = fmt.Errorf("SetSigGenPropertiesArbitrary:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps5000aSetSigGenPropertiesBuiltIn(handle int16, offsetVoltage int32,
	startFrequency, stopFrequency, increment, dwellTime float64,
	sweepType SweepTypeEnum,
	shots, sweeps uint32, triggerType SigGenTrigType,
	triggerSource SigGenTrigSource, extInThreshold int16) (err error) {
	slog.Debug("ps5000aSetSigGenPropertiesBuiltIn", "handle", handle, "offsetVoltage", offsetVoltage, "startFrequency", startFrequency, "stopFrequency", stopFrequency, "increment", increment, "dwellTime", dwellTime, "sweepType", sweepType, "shots", shots, "sweeps", sweeps, "triggerType", triggerType, "triggerSource", triggerSource, "extInThreshold", extInThreshold)
	stat := C.ps5000aSetSigGenPropertiesBuiltIn((C.short)(handle),
		(C.double)(startFrequency), (C.double)(stopFrequency),
		(C.double)(increment), (C.double)(dwellTime),
		(C.PS5000A_SWEEP_TYPE)(sweepType),
		(C.uint)(shots), (C.uint)(sweeps), (C.PS5000A_SIGGEN_TRIG_TYPE)(triggerType),
		(C.PS5000A_SIGGEN_TRIG_SOURCE)(triggerSource), (C.short)(extInThreshold))

	if stat != C.PICO_OK {
		err = fmt.Errorf("SetSigGenPropertiesBuiltIn:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps5000aSigGenArbitraryMinMaxValues(handle int16) (minArbitraryWaveformValue, maxArbitraryWaveformValue int16,
	minArbitraryWaveformSize, maxArbitraryWaveformSize uint32, err error) {
	slog.Debug("ps5000aSigGenArbitraryMinMaxValues", "handle", handle)
	stat := C.ps5000aSigGenArbitraryMinMaxValues((C.short)(handle),
		(*C.short)(&minArbitraryWaveformValue), (*C.short)(&maxArbitraryWaveformValue),
		(*C.uint32_t)(&minArbitraryWaveformSize), (*C.uint32_t)(&maxArbitraryWaveformSize))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SigGenArbitraryMinMaxValues:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps5000aSigGenSoftwareControl(handle int16, state int16) (err error) {
	slog.Debug("ps5000aSigGenSoftwareControl", "handle", handle, "state", state)
	stat := C.ps5000aSigGenSoftwareControl((C.short)(handle),
		(C.short)(state))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetSigGenBuiltIn:  %s", psc.StatStr(int(stat)))
	}
	return
}
