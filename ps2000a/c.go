//go:build !noscope

package ps2000a

// #cgo CFLAGS: -g -Wall
// #cgo LDFLAGS: -L/opt/picoscope/lib/ -lps2000a
// #include <stdlib.h>
// #include "/opt/picoscope/include/libps2000/ps2000.h"
// #include "/opt/picoscope/include/libps2000a/PicoStatus.h"
// #include "/opt/picoscope/include/libps2000a/ps2000aApi.h"
/*
// Forward declarations
int lpBlockReady(int16_t handle, PICO_STATUS status, void * pParameter);
int lpDataReady(int16_t handle, PICO_STATUS status, uint32_t noOfSamples,
				int16_t overflow, void * pParameter);
int lpStreamingReady(int16_t handle, int32_t noOfSamples, uint32_t startIndex,
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
	scopeHandler.Id = "ps2000a"
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
		slog.Debug("ps2000aEnumerateUnits", "bufferLen", bufferLen)
		stat := C.ps2000aEnumerateUnits((*C.short)(&count), cstrPtr, (*C.short)(&serialLth))
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
	slog.Debug("ps2000aOpenUnit", "serial", serial)
	stat := C.ps2000aOpenUnit((*C.short)(&handle), (*C.schar)(p))
	if stat != C.PICO_OK {
		err = fmt.Errorf("OpenUnit:  %s", psc.StatStr(int(stat)))
		return
	}
	loadConstants()
	return
}
func openUnitAsync(serial string) (status int16, err error) {
	var p *C.schar
	sLength := len(serial)
	if sLength > 0 {
		p = (*C.schar)(C.CString(serial))
		defer C.free(unsafe.Pointer(p))
	}
	slog.Debug("ps2000aOpenUnitAsync", "serial", serial)
	stat := C.ps2000aOpenUnitAsync((*C.short)(&status), (*C.schar)(p))
	if stat != C.PICO_OK {
		err = fmt.Errorf("OpenUnit:  %s", psc.StatStr(int(stat)))
		return
	}
	return
}

func ps2000aCloseUnit(handle int16) (err error) {
	slog.Debug("ps2000aCloseUnit", "handle", handle)
	stat := C.ps2000aCloseUnit((C.short)(handle))
	if stat != C.PICO_OK {
		err = fmt.Errorf("CloseUnit:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aGetUnitInfo(handle int16, info PicoInfo) (infoString string, err error) {
	const listLen = 4096
	var cstrPtr *C.schar
	cstrPtr = (*C.schar)(C.malloc(C.sizeof_schar * listLen))
	defer C.free(unsafe.Pointer(cstrPtr))
	requiredSize := int16(listLen)
	slog.Debug("ps2000aGetUnitInfo", "handle", handle, "info", info)
	stat := C.ps2000aGetUnitInfo((C.short)(handle), cstrPtr, (C.short)(requiredSize),
		(*C.short)(&requiredSize), (C.PICO_INFO)(info))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetUnitInfo:  %s", psc.StatStr(int(stat)))
	}
	if requiredSize == 0 {
		infoString = "No answer from ps2000aGetUnitInfo "
		return
	}
	b := C.GoBytes(unsafe.Pointer(cstrPtr), (C.int)(requiredSize-1))
	infoString = string(b)
	return
}

func ps2000aFlashLed(handle int16, start int16) (err error) {
	slog.Debug("ps2000aFlashLed", "handle", handle, "start", start)
	stat := C.ps2000aFlashLed((C.short)(handle), (C.short)(start))
	if stat != C.PICO_OK {
		err = fmt.Errorf("FlashLed:  %s", psc.StatStr(int(stat)))
	}
	return
}

// cgo callback workaround
// 1. registers go callback function
// 2. registers C callback function
// 3. C calls registered C callback function
// 4. Registered C callback function calls bridge go function
// 5. Bridge go function calls registered go callback function
var regLpDataReadyGo DataReady // registered go callback function

// Bridge callback function. It is visible from C. (from callbacks.go lpDataReady C function)
// No space allowed before export!
//
//export lpDataReadyGo
func lpDataReadyGo(handle int16, status int, noOfSamples uint32, overflow int16, param interface{}) {
	if regLpDataReadyGo != nil {
		regLpDataReadyGo(handle, status, noOfSamples, overflow, param) // call registered go callback function
	}
	return
}

func ps2000aGetValuesAsync(handle int16, startIndex, noOfSamples, downSampleRatio uint32,
	downSampleRatioMode RatioMode, lpDataReadyGoPar DataReady, segmentIndex uint32,
	param interface{}) (err error) {
	regLpDataReadyGo = lpDataReadyGoPar
	slog.Debug("ps2000aGetValuesAsync", "handle", handle, "startIndex", startIndex, "noOfSamples", noOfSamples, "downSampleRatio", downSampleRatio, "downSampleRatioMode", downSampleRatioMode, "lpDataReadyGoPar", lpDataReadyGoPar, "segmentIndex", segmentIndex, "param", param)
	stat := C.ps2000aGetValuesAsync((C.short)(handle),
		(C.uint)(startIndex),
		(C.uint)(noOfSamples),
		(C.uint)(downSampleRatio),
		(C.PS2000A_RATIO_MODE)(downSampleRatioMode),
		(C.uint)(segmentIndex),
		(C.lpDataReady), // C callback function in callbacks.go
		unsafe.Pointer(&param))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetValuesAsync:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aGetValues(handle int16, startIndex, reqNoOfSamples, downSampleRatio uint32,
	downSampleRatioMode RatioMode, segmentIndex uint32) (noOfSamples uint32, overflow int16, err error) {
	slog.Debug("ps2000aGetValues", "handle", handle, "startIndex", startIndex, "reqNoOfSamples", reqNoOfSamples, "downSampleRatio", downSampleRatio, "downSampleRatioMode", downSampleRatioMode, "segmentIndex", segmentIndex)
	stat := C.ps2000aGetValues((C.short)(handle),
		(C.uint)(startIndex),
		(*C.uint)(&reqNoOfSamples),
		(C.uint)(downSampleRatio),
		(C.PS2000A_RATIO_MODE)(downSampleRatioMode),
		(C.uint)(segmentIndex),
		(*C.short)(&overflow))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetValues:  %s", psc.StatStr(int(stat)))
	}
	noOfSamples = reqNoOfSamples
	return
}

func ps2000aGetValuesBulk(handle int16, reqNoOfSamples uint32, fromSegmentIndex, toSegmentIndex, downSampleRatio uint32,
	downSampleRatioMode RatioMode, overflow []int16) (noSamples uint32, err error) {
	slog.Debug("ps2000aGetValuesBulk", "handle", handle, "reqNoOfSamples", reqNoOfSamples, "fromSegmentIndex", fromSegmentIndex, "toSegmentIndex", toSegmentIndex, "downSampleRatio", downSampleRatio, "downSampleRatioMode", downSampleRatioMode, "overflow", overflow)
	stat := C.ps2000aGetValuesBulk((C.short)(handle),
		(*C.uint)(&reqNoOfSamples),
		(C.uint)(fromSegmentIndex),
		(C.uint)(toSegmentIndex),
		(C.uint)(downSampleRatio),
		(C.PS2000A_RATIO_MODE)(downSampleRatioMode),
		(*C.short)(&overflow[0]))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetValuesBulk:  %s", psc.StatStr(int(stat)))
	}
	noSamples = reqNoOfSamples
	return
}

func ps2000aGetValuesOverlapped(handle int16, startIndex, reqNoOfSamples, downSampleRatio uint32,
	downSampleRatioMode RatioMode, segmentIndex uint32, overflow []int16) (noSamples uint32, err error) {
	slog.Debug("ps2000aGetValuesOverlapped", "handle", handle, "startIndex", startIndex, "reqNoOfSamples", reqNoOfSamples, "downSampleRatio", downSampleRatio, "downSampleRatioMode", downSampleRatioMode, "segmentIndex", segmentIndex, "overflow", overflow)
	stat := C.ps2000aGetValuesOverlapped((C.short)(handle),
		(C.uint)(startIndex),
		(*C.uint)(&reqNoOfSamples),
		(C.uint)(downSampleRatio),
		(C.PS2000A_RATIO_MODE)(downSampleRatioMode),
		(C.uint)(segmentIndex),
		(*C.short)(&overflow[0]))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetValuesOverlapped:  %s", psc.StatStr(int(stat)))
	}
	noSamples = reqNoOfSamples
	return
}

func ps2000aGetValuesOverlappedBulk(handle int16, startIndex, reqNoOfSamples, downSampleRatio uint32,
	downSampleRatioMode RatioMode, fromSegmentIndex, toSegmentIndex uint32, overflow []int16) (noSamples uint32, err error) {
	slog.Debug("ps2000aGetValuesOverlappedBulk", "handle", handle, "startIndex", startIndex, "reqNoOfSamples", reqNoOfSamples, "downSampleRatio", downSampleRatio, "downSampleRatioMode", downSampleRatioMode, "fromSegmentIndex", fromSegmentIndex, "toSegmentIndex", toSegmentIndex, "overflow", overflow)
	stat := C.ps2000aGetValuesOverlappedBulk((C.short)(handle),
		(C.uint)(startIndex),
		(*C.uint)(&reqNoOfSamples),
		(C.uint)(downSampleRatio),
		(C.PS2000A_RATIO_MODE)(downSampleRatioMode),
		(C.uint)(fromSegmentIndex),
		(C.uint)(toSegmentIndex),
		(*C.short)(&overflow[0]))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetValuesOverlappedBulk:  %s", psc.StatStr(int(stat)))
	}
	noSamples = reqNoOfSamples
	return
}

func ps2000aGetAnalogueOffset(handle int16, voltageRange int, coupling Coupling) (maximumVoltage, minimumVoltage float32, err error) {
	slog.Debug("ps2000aGetAnalogueOffset", "handle", handle, "voltageRange", voltageRange, "coupling", coupling)
	stat := C.ps2000aGetAnalogueOffset((C.short)(handle),
		(C.PS2000A_RANGE)(voltageRange),
		(C.PS2000A_COUPLING)(coupling), (*C.float)(&maximumVoltage), (*C.float)(&minimumVoltage))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetAnalogueOffset:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aGetChannelInformation(handle int16, info int16, probe int32, ranges []int32, channels ChannelId) (lengthOfRanges int32, err error) {
	lengthOfRanges = int32(len(ranges))
	slog.Debug("ps2000aGetChannelInformation", "handle", handle, "info", info, "probe", probe, "ranges", ranges, "channels", channels)
	stat := C.ps2000aGetChannelInformation((C.short)(handle), (C.PS2000A_CHANNEL_INFO)(info),
		(C.int)(probe), (*C.int)(&ranges[0]), (*C.int)(&lengthOfRanges), (C.int)(channels))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetChannelInformation:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aGetMaxDownSampleRatio(handle int16, noOfUnaggregatedSamples uint32, downSampleRatioMode RatioMode, segmentIndex int32) (maxDownSampleRatio uint32, err error) {
	slog.Debug("ps2000aGetMaxDownSampleRatio", "handle", handle, "noOfUnaggregatedSamples", noOfUnaggregatedSamples, "downSampleRatioMode", downSampleRatioMode, "segmentIndex", segmentIndex)
	stat := C.ps2000aGetMaxDownSampleRatio((C.short)(handle), (C.uint)(noOfUnaggregatedSamples),
		(*C.uint)(&maxDownSampleRatio), (C.PS2000A_RATIO_MODE)(downSampleRatioMode), (C.uint)(segmentIndex))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetMaxDownSampleRatio:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aGetMaxSegments(handle int16) (maxSegments uint32, err error) {
	slog.Debug("ps2000aGetMaxSegments", "handle", handle)
	stat := C.ps2000aGetMaxSegments((C.short)(handle), (*C.uint)(&maxSegments))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetMaxSegments:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aGetNumOfCaptures(handle int16) (nCaptures uint32, err error) {
	slog.Debug("ps2000aGetNoOfCaptures", "handle", handle)
	stat := C.ps2000aGetNoOfCaptures((C.short)(handle), (*C.uint)(&nCaptures))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetNoOfCaptures:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aGetNumOfProcessedCaptures(handle int16) (nCaptures uint32, err error) {
	slog.Debug("ps2000aGetNoOfProcessedCaptures", "handle", handle)
	stat := C.ps2000aGetNoOfProcessedCaptures((C.short)(handle), (*C.uint)(&nCaptures))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetNoOfProcessedCaptures:  %s", psc.StatStr(int(stat)))
	}
	return
}

// cgo callback workaround
// 1. registers go callback function
// 2. registers C callback function
// 3. C calls registered C callback function
// 4. Registered C callback function calls bridge go function
// 5. Bridge go function calls registered go callback function
var regLpStreamingReadyGo StreamingReady // registered go callback function

// Bridge callback function. It is visible from C. (from callbacks.go lpDataReady C function)
// No space allowed before export!
//
//export lpStreamingReadyGo
func lpStreamingReadyGo(handle int16, noOfSamples int32, startIndex uint32, overflow int16,
	triggeredAt uint32, triggered, autoStop int16, param interface{}) {
	if regLpStreamingReadyGo != nil {
		regLpStreamingReadyGo(handle, noOfSamples, startIndex, overflow, triggeredAt, autoStop, triggered, param) // call registered go callback function
	}
	return
}

func ps2000aGetStreamingLatestValues(handle int16, lpStreamingReadyGoPar StreamingReady, param interface{}) (err error) {
	regLpStreamingReadyGo = lpStreamingReadyGoPar
	slog.Debug("ps2000aGetStreamingLatestValues", "handle", handle, "lpStreamingReadyGoPar", lpStreamingReadyGoPar, "param", param)
	stat := C.ps2000aGetStreamingLatestValues((C.short)(handle),
		(C.ps2000aStreamingReady)(C.lpStreamingReady), // C callback function in callbacks.go
		unsafe.Pointer(&param))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetStreamingLatestValues:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aGetTimebase(handle int16, timeBase uint32, noOfSamples int32, overSample int16, segmentIndex uint32) (timeIntervalNanoseconds, maxSamples int32, err error) {
	slog.Debug("ps2000aGetTimebase", "handle", handle, "timeBase", timeBase, "noOfSamples", noOfSamples, "overSample", overSample, "segmentIndex", segmentIndex)
	stat := C.ps2000aGetTimebase((C.short)(handle), (C.uint)(timeBase), (C.int)(noOfSamples),
		(*C.int)(&timeIntervalNanoseconds), (C.short)(overSample),
		(*C.int)(&maxSamples), (C.uint)(segmentIndex))
	if stat != C.PICO_OK {
		slog.Error("GetTimebase", "noOfSamples", noOfSamples, "stat", psc.StatStr(int(stat)))
		err = fmt.Errorf("GetTimebase:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aGetTimebase2(handle int16, timeBase uint32, numOfSamples int32,
	overSample int16, segmentIndex uint32) (timeIntervalNanoseconds float32, maxSamples int32, err error) {
	slog.Debug("ps2000aGetTimebase2", "handle", handle, "timeBase", timeBase, "numOfSamples", numOfSamples, "overSample", overSample, "segmentIndex", segmentIndex)
	stat := C.ps2000aGetTimebase2((C.short)(handle), (C.uint)(timeBase), (C.int)(numOfSamples),
		(*C.float)(&timeIntervalNanoseconds), (C.short)(overSample),
		(*C.int)(&maxSamples), (C.uint)(segmentIndex))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetTimebase2:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aSetChannel(handle int16, channel ChannelId, enabled bool, couplingType Coupling, voltageRange RangeEnum, analogOffset float32) (err error) {
	slog.Debug("ps2000aSetChannel", "handle", handle, "channel", channel, "enabled", enabled, "couplingType", couplingType, "voltageRange", voltageRange, "analogOffset", analogOffset)
	stat := C.ps2000aSetChannel((C.short)(handle), (C.PS2000A_CHANNEL)(channel), (C.short)(boolToint16(enabled)),
		(C.PS2000A_COUPLING)(couplingType), (C.PS2000A_RANGE)(voltageRange),
		(C.float)(analogOffset))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetChannel:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aMaximumValue(handle int16) (value int16, err error) {
	slog.Debug("ps2000aMaximumValue", "handle", handle)
	stat := C.ps2000aMaximumValue((C.short)(handle), (*C.short)(&value))
	if stat != C.PICO_OK {
		err = fmt.Errorf("MaximumValue:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aMinimumValue(handle int16) (value int16, err error) {
	slog.Debug("ps2000aMinimumValue", "handle", handle)
	stat := C.ps2000aMinimumValue((C.short)(handle), (*C.short)(&value))
	if stat != C.PICO_OK {
		err = fmt.Errorf("MinimumValue:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aSetSimpleTrigger(handle int16, enable bool, source ChannelId, threshold int16,
	direction ThresholdDirection, delay uint32, autoTriggerMs int16) (err error) {
	slog.Debug("ps2000aSetSimpleTrigger", "handle", handle, "enable", enable, "source", source, "threshold", threshold, "direction", direction, "delay", delay, "autoTriggerMs", autoTriggerMs)
	stat := C.ps2000aSetSimpleTrigger((C.short)(handle), (C.short)(boolToint16(enable)),
		(C.PS2000A_CHANNEL)(source), (C.short)(threshold),
		(C.PS2000A_THRESHOLD_DIRECTION)(direction), (C.uint)(delay),
		(C.short)(autoTriggerMs))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetSimpleTrigger:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aSetDataBuffer(handle int16, ch ChannelId, bufferIn []int16, segmentIndex uint32,
	mode RatioMode) (err error) {

	slog.Debug("ps2000aSetDataBuffer", "handle", handle, "ch", ch /* "bufferIn", bufferIn,*/, "segmentIndex", segmentIndex, "mode", mode)
	stat := C.ps2000aSetDataBuffer((C.short)(handle), (C.int)(ch), (*C.short)(&bufferIn[0]),
		(C.int)(len(bufferIn)), (C.uint)(segmentIndex),
		(C.PS2000A_RATIO_MODE)(mode))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetDataBuffer:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aSetDataBuffers(handle int16, ch ChannelId, bufferMax, bufferMin []int16, segmentIndex uint32, mode RatioMode) (err error) {
	slog.Debug("ps2000aSetDataBuffers", "handle", handle, "ch", ch, "bufferMax", bufferMax, "bufferMin", bufferMin, "segmentIndex", segmentIndex, "mode", mode)
	stat := C.ps2000aSetDataBuffers((C.short)(handle), (C.int)(ch), (*C.short)(&bufferMax[0]),
		(*C.short)(&bufferMin[0]), (C.int)(len(bufferMax)), (C.uint)(segmentIndex),
		(C.PS2000A_RATIO_MODE)(mode))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetDataBuffers:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aSetUnscaledDataBuffers(handle int16, ch ChannelId, bufferMax, bufferMin []int16, segmentIndex uint32, mode RatioMode) (err error) {
	slog.Debug("ps2000aSetDataBuffers", "handle", handle, "ch", ch, "bufferMax", bufferMax, "bufferMin", bufferMin, "segmentIndex", segmentIndex, "mode", mode)
	stat := C.ps2000aSetDataBuffers((C.short)(handle), (C.int)(ch), (*C.short)(&bufferMax[0]),
		(*C.short)(&bufferMin[0]), (C.int)(len(bufferMax)), (C.uint)(segmentIndex),
		(C.PS2000A_RATIO_MODE)(mode))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetUnscaledDataBuffers:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aSetEtsTimeBuffer(handle int16, buffer []int64) (err error) {
	slog.Debug("ps2000aSetEtsTimeBuffer", "handle", handle, "buffer", buffer)
	stat := C.ps2000aSetEtsTimeBuffer((C.short)(handle), (*C.long)(&buffer[0]),
		(C.int)(len(buffer)))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetEtsTimeBuffer:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aSetEtsTimeBuffers(handle int16, timeUpper, timeLower []uint32) (err error) {
	slog.Debug("ps2000aSetEtsTimeBuffers", "handle", handle, "timeUpper", timeUpper, "timeLower", timeLower)
	stat := C.ps2000aSetEtsTimeBuffers((C.short)(handle), (*C.uint)(&timeUpper[0]),
		(*C.uint)(&timeLower[0]), (C.int)(len(timeUpper)))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetEtsTimeBuffer:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aSetEts(handle int16, mode EtsMode, etsCycles int16, etsInterLeave int16) (sampleTimePicoseconds int32, err error) {
	slog.Debug("ps2000aSetEts", "handle", handle, "mode", mode, "etsCycles", etsCycles, "etsInterLeave", etsInterLeave)
	stat := C.ps2000aSetEts((C.short)(handle), (C.PS2000A_ETS_MODE)(mode),
		(C.short)(etsCycles), (C.short)(etsInterLeave), (*C.int)(&sampleTimePicoseconds))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetEts:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aRunStreaming(handle int16, reqSampleInterval uint32, sampleIntervalTimeUnits TimeUnits,
	maxPreTriggerSamples, maxPostTriggerSamples uint32,
	autoStop bool, downSampleRatio uint32, downSampleRatioMode RatioMode,
	overviewBufferSize uint32) (sampleInterval uint32, err error) {
	slog.Debug("ps2000aRunStreaming", "handle", handle, "reqSampleInterval", reqSampleInterval, "sampleIntervalTimeUnits", sampleIntervalTimeUnits, "maxPreTriggerSamples", maxPreTriggerSamples, "maxPostTriggerSamples", maxPostTriggerSamples, "autoStop", autoStop, "downSampleRatio", downSampleRatio, "downSampleRatioMode", downSampleRatioMode, "overviewBufferSize", overviewBufferSize)
	stat := C.ps2000aRunStreaming((C.short)(handle), (*C.uint)(&reqSampleInterval),
		(C.PS2000A_TIME_UNITS)(sampleIntervalTimeUnits), (C.uint)(maxPreTriggerSamples),
		(C.uint)(maxPostTriggerSamples), (C.short)(boolToint16(autoStop)), (C.uint)(downSampleRatio),
		(C.PS2000A_RATIO_MODE)(downSampleRatioMode), (C.uint)(overviewBufferSize))
	if stat != C.PICO_OK {
		err = fmt.Errorf("RunStreaming:  %s", psc.StatStr(int(stat)))
	}
	sampleInterval = reqSampleInterval
	return
}

// cgo callback workaround
// 1. registers go callback function
// 2. registers C callback function
// 3. C calls registered C callback function
// 4. Registered C callback function calls bridge go function
// 5. Bridge go function calls registered go callback function
var regLpBlockReadyGo BlockReady // registered go callback function

// Bridge callback function. It is visible from C. (from callbacks.go lpDataReady C function)
// No space allowed before export!
//
//export lpBlockReadyGo
func lpBlockReadyGo(handle int16, status int, noOfSamples uint32, overflow int16, param interface{}) {
	if regLpBlockReadyGo != nil {
		regLpBlockReadyGo(handle, status, param) // call registered go callback function
	}
	return
}

func ps2000aRunBlock(handle int16, noOfPreTriggerSamples, noOfPostTriggerSamples int32,
	timeBase uint32, overSample int16, segmentIndex uint32, lpBlockReadyGoPar BlockReady,
	param interface{}) (timeIndisposedMs int32, err error) {
	regLpBlockReadyGo = lpBlockReadyGoPar
	nSamples := noOfPreTriggerSamples + noOfPostTriggerSamples
	if nSamples > 1<<25 { // avoid exception in cgo
		err = fmt.Errorf("RunBlock:  too many required samples %d", nSamples)
		return
	}
	slog.Debug("ps2000aRunBlock", "handle", handle, "noOfPreTriggerSamples", noOfPreTriggerSamples, "noOfPostTriggerSamples", noOfPostTriggerSamples, "timeBase", timeBase, "overSample", overSample, "segmentIndex", segmentIndex, "lpBlockReadyGoPar", lpBlockReadyGoPar, "param", param)
	stat := C.ps2000aRunBlock((C.short)(handle), (C.int)(noOfPreTriggerSamples),
		(C.int)(noOfPostTriggerSamples), (C.uint)(timeBase), (C.short)(overSample),
		(*C.int)(&timeIndisposedMs), (C.uint)(segmentIndex), (C.ps2000aBlockReady)(C.lpBlockReady),
		unsafe.Pointer(&param))
	if stat != C.PICO_OK {
		err = fmt.Errorf("RunBlock:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aSetTriggerChannelProperties(handle int16, channelProperties []TriggerChannelProperties, auxOutputEnable bool,
	autoTriggerMs int32) (err error) {
	var cTriggerChannelProperties []C.PS2000A_TRIGGER_CHANNEL_PROPERTIES
	if len(channelProperties) > 0 {
		cTriggerChannelProperties = make([]C.PS2000A_TRIGGER_CHANNEL_PROPERTIES, len(channelProperties))
		for i := range channelProperties {
			cTriggerChannelProperties[i].channel = (C.PS2000A_CHANNEL)(channelProperties[i].Channel)
			cTriggerChannelProperties[i].thresholdLowerHysteresis = (C.ushort)(channelProperties[i].ThresholdLowerHysteresis)
			cTriggerChannelProperties[i].thresholdLower = (C.short)(channelProperties[i].ThresholdLower)
			cTriggerChannelProperties[i].thresholdUpperHysteresis = (C.ushort)(channelProperties[i].ThresholdUpperHysteresis)
			cTriggerChannelProperties[i].thresholdUpper = (C.short)(channelProperties[i].ThresholdUpper)
			cTriggerChannelProperties[i].thresholdMode = (C.PS2000A_THRESHOLD_MODE)(channelProperties[i].ThresholdMode)
		}
	}
	pcTriggerChannelProperties := (*C.PS2000A_TRIGGER_CHANNEL_PROPERTIES)(nil)
	if len(channelProperties) > 0 {
		pcTriggerChannelProperties = &cTriggerChannelProperties[0]
	}
	slog.Debug("ps2000aSetTriggerChannelProperties", "handle", handle, "channelProperties", channelProperties, "auxOutputEnable", auxOutputEnable, "autoTriggerMs", autoTriggerMs)
	stat := C.ps2000aSetTriggerChannelProperties((C.short)(handle),
		(*C.PS2000A_TRIGGER_CHANNEL_PROPERTIES)(pcTriggerChannelProperties),
		(C.short)(len(channelProperties)), (C.short)(boolToint16(auxOutputEnable)), (C.int)(autoTriggerMs))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetTriggerChannelProperties:  %s", psc.StatStr(int(stat)))
	}
	return

}

func ps2000aSetTriggerChannelConditions(handle int16, triggerConditions []TriggerConditions) (err error) {
	cTriggerConditions := make([]C.PS2000A_TRIGGER_CONDITIONS, len(triggerConditions))
	for i := range triggerConditions {
		cTriggerConditions[i].channelA = (C.PS2000A_TRIGGER_STATE)(triggerConditions[i].ChannelA)
		cTriggerConditions[i].channelB = (C.PS2000A_TRIGGER_STATE)(triggerConditions[i].ChannelB)
		cTriggerConditions[i].channelC = (C.PS2000A_TRIGGER_STATE)(triggerConditions[i].ChannelC)
		cTriggerConditions[i].channelD = (C.PS2000A_TRIGGER_STATE)(triggerConditions[i].ChannelD)
		cTriggerConditions[i].external = (C.PS2000A_TRIGGER_STATE)(triggerConditions[i].External)
		cTriggerConditions[i].aux = (C.PS2000A_TRIGGER_STATE)(triggerConditions[i].Aux)
		cTriggerConditions[i].pulseWidthQualifier = (C.PS2000A_TRIGGER_STATE)(triggerConditions[i].PulseWidthQualifier)
		cTriggerConditions[i].digital = (C.PS2000A_TRIGGER_STATE)(triggerConditions[i].Digital)
	}
	slog.Debug("ps2000aSetTriggerChannelConditions", "handle", handle, "triggerConditions", triggerConditions)
	stat := C.ps2000aSetTriggerChannelConditions((C.short)(handle), (*C.PS2000A_TRIGGER_CONDITIONS)(nil), 0)
	pcTriggerConditions := (*C.PS2000A_TRIGGER_CONDITIONS)(nil)
	if len(triggerConditions) > 0 {
		pcTriggerConditions = &cTriggerConditions[0]
	}
	slog.Debug("ps2000aSetTriggerChannelConditions", "handle", handle, "triggerConditions", triggerConditions)
	stat = C.ps2000aSetTriggerChannelConditions((C.short)(handle),
		(*C.PS2000A_TRIGGER_CONDITIONS)(pcTriggerConditions),
		(C.short)(len(triggerConditions)))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetTriggerChannelConditions:  %s", psc.StatStr(int(stat)))
	}

	return
}

func ps2000aSetTriggerChannelDirections(handle int16, channelA, channelB, channelC, channelD, ext, aux ThresholdDirection) (err error) {
	slog.Debug("ps2000aSetTriggerChannelDirections", "handle", handle, "channelA", channelA, "channelB", channelB, "channelC", channelC, "channelD", channelD, "ext", ext, "aux", aux)
	stat := C.ps2000aSetTriggerChannelDirections((C.short)(handle),
		(C.PS2000A_THRESHOLD_DIRECTION)(channelA),
		(C.PS2000A_THRESHOLD_DIRECTION)(channelB),
		(C.PS2000A_THRESHOLD_DIRECTION)(channelC),
		(C.PS2000A_THRESHOLD_DIRECTION)(channelD),
		(C.PS2000A_THRESHOLD_DIRECTION)(ext),
		(C.PS2000A_THRESHOLD_DIRECTION)(aux))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetTriggerChannelDirections:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aSetTriggerDelay(handle int16, delay uint32) (err error) {
	slog.Debug("ps2000aSetTriggerDelay", "handle", handle, "delay", delay)
	stat := C.ps2000aSetTriggerDelay((C.short)(handle), (C.uint)(delay))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetTriggerDelay:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aSetPulseWidthQualifier(handle int16, conditions []PwqConditions, direction ThresholdDirection, lower, upper uint32,
	pwType PulseWidthType) (err error) {
	cPwqConditions := make([]C.PS2000A_PWQ_CONDITIONS, len(conditions))
	for i := range conditions {
		cPwqConditions[i].channelA = (C.PS2000A_TRIGGER_STATE)(conditions[i].ChannelA)
		cPwqConditions[i].channelB = (C.PS2000A_TRIGGER_STATE)(conditions[i].ChannelB)
		cPwqConditions[i].channelC = (C.PS2000A_TRIGGER_STATE)(conditions[i].ChannelC)
		cPwqConditions[i].channelD = (C.PS2000A_TRIGGER_STATE)(conditions[i].ChannelD)
		cPwqConditions[i].external = (C.PS2000A_TRIGGER_STATE)(conditions[i].External)
		cPwqConditions[i].aux = (C.PS2000A_TRIGGER_STATE)(conditions[i].Aux)
		cPwqConditions[i].digital = (C.PS2000A_TRIGGER_STATE)(conditions[i].Digital)
	}
	pcPwqConditions := (*C.PS2000A_PWQ_CONDITIONS)(nil)
	if len(conditions) > 0 {
		pcPwqConditions = &cPwqConditions[0]
	}
	slog.Debug("ps2000aSetPulseWidthQualifier", "handle", handle, "conditions", conditions, "direction", direction, "lower", lower, "upper", upper, "pwType", pwType)
	stat := C.ps2000aSetPulseWidthQualifier((C.short)(handle),
		pcPwqConditions, (C.short)(len(conditions)),
		(C.PS2000A_THRESHOLD_DIRECTION)(direction), (C.uint32_t)(lower), (C.uint32_t)(upper),
		(C.PS2000A_PULSE_WIDTH_TYPE)(pwType))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetPulseWidthQualifier:  %s", psc.StatStr(int(stat)))
	}
	triggerEnabled := int16(0)
	pulseWidthQualifierEnabled := int16(0)
	slog.Debug("ps2000aIsTriggerOrPulseWidthQualifierEnabled", "handle", handle)
	stat = C.ps2000aIsTriggerOrPulseWidthQualifierEnabled((C.short)(handle),
		(*C.short)(&triggerEnabled), (*C.short)(&pulseWidthQualifierEnabled))
	if stat != C.PICO_OK {
		err = fmt.Errorf("ps2000aIsTriggerOrPulseWidthQualifierEnabled:  %s", psc.StatStr(int(stat)))
	}
	return
}
func ps2000aSetTriggerDigitalPortProperties(handle int16, digitalDirections []DigitalChannelDirections) (err error) {
	cDigitalDirections := make([]C.PS2000A_DIGITAL_CHANNEL_DIRECTIONS, len(digitalDirections))
	for i := range digitalDirections {
		cDigitalDirections[i].channel = (C.PS2000A_DIGITAL_CHANNEL)(digitalDirections[i].Channel)
		cDigitalDirections[i].direction = (C.PS2000A_DIGITAL_DIRECTION)(digitalDirections[i].Direction)
	}
	slog.Debug("ps2000aSetTriggerDigitalPortProperties", "handle", handle, "digitalDirections", digitalDirections)
	stat := C.ps2000aSetTriggerDigitalPortProperties((C.short)(handle),
		(*C.PS2000A_DIGITAL_CHANNEL_DIRECTIONS)(&cDigitalDirections[0]), (C.short)(len(digitalDirections)))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetTriggerDigitalPortProperties:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aStop(handle int16) (err error) {
	slog.Debug("ps2000aStop", "handle", handle)
	stat := C.ps2000aStop((C.short)(handle))
	if stat != C.PICO_OK {
		err = fmt.Errorf("Stop:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aSetSigGenBuiltIn(handle int16, offsetVoltage int32, pkToPK uint32, waveType WaveTypeEnum,
	startFrequency, stopFrequency, increment, dwellTime float32, sweepType SweepTypeEnum,
	operation ExtraOperations, shots, sweeps uint32, triggerType SigGenTrigType,
	triggerSource SigGenTrigSource, extInThreshold int16) (err error) {
	slog.Debug("ps2000aSetSigGenBuiltIn", "handle", handle, "offsetVoltage", offsetVoltage, "pkToPK", pkToPK, "waveType", waveType, "startFrequency", startFrequency, "stopFrequency", stopFrequency, "increment", increment, "dwellTime", dwellTime, "sweepType", sweepType, "operation", operation, "shots", shots, "sweeps", sweeps, "triggerType", triggerType, "triggerSource", triggerSource, "extInThreshold", extInThreshold)
	stat := C.ps2000aSetSigGenBuiltIn((C.short)(handle), (C.int)(offsetVoltage),
		(C.uint)(pkToPK), (C.short)(waveType), (C.float)(startFrequency),
		(C.float)(stopFrequency), (C.float)(increment), (C.float)(dwellTime),
		(C.PS2000A_SWEEP_TYPE)(sweepType), (C.PS2000A_EXTRA_OPERATIONS)(operation),
		(C.uint)(shots), (C.uint)(sweeps), (C.PS2000A_SIGGEN_TRIG_TYPE)(triggerType),
		(C.PS2000A_SIGGEN_TRIG_SOURCE)(triggerSource), (C.short)(extInThreshold))

	if stat != C.PICO_OK {
		err = fmt.Errorf("SetSigGenBuiltIn:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aSetSigGenBuiltInV2(handle int16, offsetVoltage int32, pkToPK uint32, waveType WaveTypeEnum,
	startFrequency, stopFrequency, increment, dwellTime float64, sweepType SweepTypeEnum,
	operation ExtraOperations, shots, sweeps uint32, triggerType SigGenTrigType,
	triggerSource SigGenTrigSource, extInThreshold int16) (err error) {
	slog.Debug("ps2000aSetSigGenBuiltInV2", "handle", handle, "offsetVoltage", offsetVoltage, "pkToPK", pkToPK, "waveType", waveType, "startFrequency", startFrequency, "stopFrequency", stopFrequency, "increment", increment, "dwellTime", dwellTime, "sweepType", sweepType, "operation", operation, "shots", shots, "sweeps", sweeps, "triggerType", triggerType, "triggerSource", triggerSource, "extInThreshold", extInThreshold)
	stat := C.ps2000aSetSigGenBuiltInV2((C.short)(handle), (C.int)(offsetVoltage),
		(C.uint)(pkToPK), (C.short)(waveType), (C.double)(startFrequency),
		(C.double)(stopFrequency), (C.double)(increment), (C.double)(dwellTime),
		(C.PS2000A_SWEEP_TYPE)(sweepType), (C.PS2000A_EXTRA_OPERATIONS)(operation),
		(C.uint)(shots), (C.uint)(sweeps), (C.PS2000A_SIGGEN_TRIG_TYPE)(triggerType),
		(C.PS2000A_SIGGEN_TRIG_SOURCE)(triggerSource), (C.short)(extInThreshold))

	if stat != C.PICO_OK {
		err = fmt.Errorf("SetSigGenBuiltIn:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aSigGenFrequencyToPhase(handle int16, frequency float64, indexMode IndexMode, bufferLength uint32) (phase uint32, err error) {
	slog.Debug("ps2000aSigGenFrequencyToPhase", "handle", handle, "frequency", frequency, "indexMode", indexMode, "bufferLength", bufferLength)
	stat := C.ps2000aSigGenFrequencyToPhase((C.short)(handle), (C.double)(frequency),
		(C.PS2000A_INDEX_MODE)(indexMode), (C.uint)(bufferLength), (*C.uint)(&phase))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SigGenFrequencyToPhase:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aSetNoCaptures(handle int16, nCaptures uint32) (err error) {
	slog.Debug("ps2000aSetNoOfCaptures", "handle", handle, "nCaptures", nCaptures)
	stat := C.ps2000aSetNoOfCaptures((C.short)(handle), (C.uint)(nCaptures))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetNoCaptures:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aGetTriggerTimeOffset(handle int16, segmentIndex uint32) (timeUpper, timeLower uint32, timeUnits TimeUnits, err error) {
	slog.Debug("ps2000aGetTriggerTimeOffset", "handle", handle, "segmentIndex", segmentIndex)
	stat := C.ps2000aGetTriggerTimeOffset((C.short)(handle), (*C.uint)(&timeUpper),
		(*C.uint)(&timeLower), (*C.PS2000A_TIME_UNITS)(&timeUnits), (C.uint)(segmentIndex))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetTriggerTimeOffset:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aGetTriggerTimeOffset64(handle int16, segmentIndex uint32) (time int64, timeUnits TimeUnits, err error) {
	slog.Debug("ps2000aGetTriggerTimeOffset64", "handle", handle, "segmentIndex", segmentIndex)
	stat := C.ps2000aGetTriggerTimeOffset64((C.short)(handle), (*C.long)(&time),
		(*C.PS2000A_TIME_UNITS)(&timeUnits), (C.uint)(segmentIndex))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetTriggerTimeOffset64:  %s", psc.StatStr(int(stat)))
	}
	return

}

func ps2000aGetValuesTriggerTimeOffsetBulk(handle int16, timesUpper, timesLower []uint32, timeUnits []TimeUnits,
	fromSegmentIndex, toSegmentIndex uint32) (err error) {
	slog.Debug("ps2000aGetValuesTriggerTimeOffsetBulk", "handle", handle, "timesUpper", timesUpper, "timesLower", timesLower, "timeUnits", timeUnits, "fromSegmentIndex", fromSegmentIndex, "toSegmentIndex", toSegmentIndex)
	stat := C.ps2000aGetValuesTriggerTimeOffsetBulk((C.short)(handle), (*C.uint)(&timesUpper[0]),
		(*C.uint)(&timesLower[0]), (*C.PS2000A_TIME_UNITS)(&timeUnits[0]), (C.uint)(fromSegmentIndex),
		(C.uint)(toSegmentIndex))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetValuesTriggerTimeOffsetBulk:  %s", psc.StatStr(int(stat)))
	}

	return
}

func ps2000aGetValuesTriggerTimeOffsetBulk64(handle int16, times []int64, timeUnits []TimeUnits,
	fromSegmentIndex, toSegmentIndex uint32) (err error) {
	slog.Debug("ps2000aGetValuesTriggerTimeOffsetBulk64", "handle", handle, "times", times, "timeUnits", timeUnits, "fromSegmentIndex", fromSegmentIndex, "toSegmentIndex", toSegmentIndex)
	stat := C.ps2000aGetValuesTriggerTimeOffsetBulk64((C.short)(handle), (*C.long)(&times[0]),
		(*C.PS2000A_TIME_UNITS)(&timeUnits[0]), (C.uint)(fromSegmentIndex),
		(C.uint)(toSegmentIndex))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetValuesTriggerTimeOffsetBulk64:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aHoldOff(handle int16, holdOff uint64, holdOffType HoldOffType) (err error) {
	slog.Debug("ps2000aHoldOff", "handle", handle, "holdOff", holdOff, "holdOffType", holdOffType)
	stat := C.ps2000aHoldOff((C.short)(handle), (C.ulong)(holdOff), (C.PS2000A_HOLDOFF_TYPE)(holdOffType))
	if stat != C.PICO_OK {
		err = fmt.Errorf("HoldOff:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aIsReady(handle int16) (ready int16, err error) {
	slog.Debug("ps2000aIsReady", "handle", handle)
	stat := C.ps2000aIsReady((C.short)(handle), (*C.short)(&ready))
	if stat != C.PICO_OK {
		err = fmt.Errorf("LsReady:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aTriggerOrPulseWidthQualifierEnabled(handle int16) (triggerEnabled, pulseWidthQualifierEnabledint16 int16, err error) {
	slog.Debug("ps2000aIsTriggerOrPulseWidthQualifierEnabled", "handle", handle)
	stat := C.ps2000aIsTriggerOrPulseWidthQualifierEnabled((C.short)(handle),
		(*C.short)(&triggerEnabled), (*C.short)(&pulseWidthQualifierEnabledint16))
	if stat != C.PICO_OK {
		err = fmt.Errorf("TriggerOrPulseWidthQualifierEnabled:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aMemorySegments(handle int16, nSegments uint32) (nMaxSamples int32, err error) {
	slog.Debug("ps2000aMemorySegments", "handle", handle, "nSegments", nSegments)
	stat := C.ps2000aMemorySegments((C.short)(handle),
		(C.uint)(nSegments), (*C.int)(&nMaxSamples))
	if stat != C.PICO_OK {
		err = fmt.Errorf("MemorySegments:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aNumOfStreamingValues(handle int16) (numOfValues uint32, err error) {
	slog.Debug("ps2000aNoOfStreamingValues", "handle", handle)
	stat := C.ps2000aNoOfStreamingValues((C.short)(handle),
		(*C.uint)(&numOfValues))
	if stat != C.PICO_OK {
		err = fmt.Errorf("NumOfStreamingValues:  %s", psc.StatStr(int(stat)))
	}
	return
}

func openUnitProgress() (handle int16, progressPercent, complete int16, err error) {
	slog.Debug("ps2000aOpenUnitProgress")
	stat := C.ps2000aOpenUnitProgress((*C.short)(&handle),
		(*C.short)(&progressPercent), (*C.short)(&complete))
	if stat != C.PICO_OK {
		err = fmt.Errorf("OpenUnitProgress:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aPingUnit(handle int16) (err error) {
	slog.Debug("ps2000aPingUnit", "handle", handle)
	stat := C.ps2000aPingUnit((C.short)(handle))
	if stat != C.PICO_OK {
		err = fmt.Errorf("PingUnit:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aQueryOutputEdgeDetect(handle int16) (state int16, err error) {
	slog.Debug("ps2000aQueryOutputEdgeDetect", "handle", handle)
	stat := C.ps2000aQueryOutputEdgeDetect((C.short)(handle),
		(*C.short)(&state))
	if stat != C.PICO_OK {
		err = fmt.Errorf("OpenUnitProgress:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aSetDigitalAnalogTriggerOperand(handle int16, operand TriggerOperand) (err error) {
	slog.Debug("ps2000aSetDigitalAnalogTriggerOperand", "handle", handle, "operand", operand)
	stat := C.ps2000aSetDigitalAnalogTriggerOperand((C.short)(handle),
		(C.PS2000A_TRIGGER_OPERAND)(operand))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetDigitalAnalogTriggerOperand:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aSetDigitalPort(handle int16, port DigitalPort, enabled bool, logiclevel int16) (err error) {
	slog.Debug("ps2000aSetDigitalPort", "handle", handle, "port", port, "enabled", enabled, "logiclevel", logiclevel)
	stat := C.ps2000aSetDigitalPort((C.short)(handle),
		(C.PS2000A_DIGITAL_PORT)(port), (C.short)(boolToint16(enabled)), (C.short)(logiclevel))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetDigitalPort:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aSetOutputEdgeDetect(handle int16, state int16) (err error) {
	slog.Debug("ps2000aSetOutputEdgeDetect", "handle", handle, "state", state)
	stat := C.ps2000aSetOutputEdgeDetect((C.short)(handle), (C.short)(state))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetOutputEdgeDetect:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aSetPulseWidthDigitalPortProperties(handle int16, digitalDirections []DigitalChannelDirections) (err error) {
	cDigitalDirections := make([]C.PS2000A_DIGITAL_CHANNEL_DIRECTIONS, len(digitalDirections))
	for i := range digitalDirections {
		cDigitalDirections[i].channel = (C.PS2000A_DIGITAL_CHANNEL)(digitalDirections[i].Channel)
		cDigitalDirections[i].direction = (C.PS2000A_DIGITAL_DIRECTION)(digitalDirections[i].Direction)
	}
	slog.Debug("ps2000aSetPulseWidthDigitalPortProperties", "handle", handle, "digitalDirections", digitalDirections)
	stat := C.ps2000aSetPulseWidthDigitalPortProperties((C.short)(handle),
		(*C.PS2000A_DIGITAL_CHANNEL_DIRECTIONS)(&cDigitalDirections[0]), (C.short)(len(digitalDirections)))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetPulseWidthDigitalPortProperties:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aSetSigGenArbitrary(handle int16, offsetVoltage int32, pkToPK uint32,
	startDeltaPhase, stopDeltaPhase, deltaPhaseIncrement, dwellCount uint32,
	arbitraryWaveform []int16, sweepType SweepTypeEnum, operation ExtraOperations,
	indexMode IndexMode, shots, sweeps uint32, triggerType SigGenTrigType,
	triggerSource SigGenTrigSource, extInThreshold int16) (err error) {
	slog.Debug("ps2000aSetSigGenArbitrary", "handle", handle, "offsetVoltage", offsetVoltage, "pkToPK", pkToPK, "startDeltaPhase", startDeltaPhase, "stopDeltaPhase", stopDeltaPhase, "deltaPhaseIncrement", deltaPhaseIncrement, "dwellCount", dwellCount, "arbitraryWaveform", arbitraryWaveform, "sweepType", sweepType, "operation", operation, "indexMode", indexMode, "shots", shots, "sweeps", sweeps, "triggerType", triggerType, "triggerSource", triggerSource, "extInThreshold", extInThreshold)
	stat := C.ps2000aSetSigGenArbitrary((C.short)(handle), (C.int)(offsetVoltage),
		(C.uint)(pkToPK), (C.uint)(startDeltaPhase), (C.uint)(stopDeltaPhase),
		(C.uint32_t)(deltaPhaseIncrement), (C.uint32_t)(dwellCount),
		(*C.short)(&arbitraryWaveform[0]), (C.int32_t)(len(arbitraryWaveform)),
		(C.PS2000A_SWEEP_TYPE)(sweepType), (C.PS2000A_EXTRA_OPERATIONS)(operation),
		(C.PS2000A_INDEX_MODE)(indexMode),
		(C.uint)(shots), (C.uint)(sweeps), (C.PS2000A_SIGGEN_TRIG_TYPE)(triggerType),
		(C.PS2000A_SIGGEN_TRIG_SOURCE)(triggerSource), (C.short)(extInThreshold))

	if stat != C.PICO_OK {
		err = fmt.Errorf("SetSigGenArbitrary:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aSetSigGenPropertiesArbitrary(handle int16, offsetVoltage int32,
	startDeltaPhase, stopDeltaPhase, deltaPhaseIncrement, dwellCount uint32,
	sweepType SweepTypeEnum, operation ExtraOperations,
	indexMode IndexMode, shots, sweeps uint32, triggerType SigGenTrigType,
	triggerSource SigGenTrigSource, extInThreshold int16) (err error) {
	slog.Debug("ps2000aSetSigGenPropertiesArbitrary", "handle", handle, "offsetVoltage", offsetVoltage, "startDeltaPhase", startDeltaPhase, "stopDeltaPhase", stopDeltaPhase, "deltaPhaseIncrement", deltaPhaseIncrement, "dwellCount", dwellCount, "sweepType", sweepType, "operation", operation, "indexMode", indexMode, "shots", shots, "sweeps", sweeps, "triggerType", triggerType, "triggerSource", triggerSource, "extInThreshold", extInThreshold)
	stat := C.ps2000aSetSigGenPropertiesArbitrary((C.short)(handle),
		(C.uint)(startDeltaPhase), (C.uint)(stopDeltaPhase),
		(C.uint32_t)(deltaPhaseIncrement), (C.uint32_t)(dwellCount),
		(C.PS2000A_SWEEP_TYPE)(sweepType),
		(C.uint)(shots), (C.uint)(sweeps), (C.PS2000A_SIGGEN_TRIG_TYPE)(triggerType),
		(C.PS2000A_SIGGEN_TRIG_SOURCE)(triggerSource), (C.short)(extInThreshold))

	if stat != C.PICO_OK {
		err = fmt.Errorf("SetSigGenPropertiesArbitrary:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aSetSigGenPropertiesBuiltIn(handle int16, offsetVoltage int32,
	startFrequency, stopFrequency, increment, dwellTime float64,
	sweepType SweepTypeEnum,
	shots, sweeps uint32, triggerType SigGenTrigType,
	triggerSource SigGenTrigSource, extInThreshold int16) (err error) {
	slog.Debug("ps2000aSetSigGenPropertiesBuiltIn", "handle", handle, "offsetVoltage", offsetVoltage, "startFrequency", startFrequency, "stopFrequency", stopFrequency, "increment", increment, "dwellTime", dwellTime, "sweepType", sweepType, "shots", shots, "sweeps", sweeps, "triggerType", triggerType, "triggerSource", triggerSource, "extInThreshold", extInThreshold)
	stat := C.ps2000aSetSigGenPropertiesBuiltIn((C.short)(handle),
		(C.double)(startFrequency), (C.double)(stopFrequency),
		(C.double)(increment), (C.double)(dwellTime),
		(C.PS2000A_SWEEP_TYPE)(sweepType),
		(C.uint)(shots), (C.uint)(sweeps), (C.PS2000A_SIGGEN_TRIG_TYPE)(triggerType),
		(C.PS2000A_SIGGEN_TRIG_SOURCE)(triggerSource), (C.short)(extInThreshold))

	if stat != C.PICO_OK {
		err = fmt.Errorf("SetSigGenPropertiesBuiltIn:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aSigGenArbitraryMinMaxValues(handle int16) (minArbitraryWaveformValue, maxArbitraryWaveformValue int16,
	minArbitraryWaveformSize, maxArbitraryWaveformSize uint32, err error) {
	slog.Debug("ps2000aSigGenArbitraryMinMaxValues", "handle", handle)
	stat := C.ps2000aSigGenArbitraryMinMaxValues((C.short)(handle),
		(*C.short)(&minArbitraryWaveformValue), (*C.short)(&maxArbitraryWaveformValue),
		(*C.uint32_t)(&minArbitraryWaveformSize), (*C.uint32_t)(&maxArbitraryWaveformSize))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SigGenArbitraryMinMaxValues:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps2000aSigGenSoftwareControl(handle int16, state int16) (err error) {
	slog.Debug("ps2000aSigGenSoftwareControl", "handle", handle, "state", state)
	stat := C.ps2000aSigGenSoftwareControl((C.short)(handle),
		(C.short)(state))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetSigGenBuiltIn:  %s", psc.StatStr(int(stat)))
	}
	return
}
