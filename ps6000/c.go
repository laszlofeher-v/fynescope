//go:build !demo && ps6000

package ps6000

// #cgo CFLAGS: -g -Wall -I/opt/picoscope/include/libps6000
// #cgo LDFLAGS: -L/opt/picoscope/lib/ -lps6000
// #include <stdlib.h>
// #include "/opt/picoscope/include/libps6000/PicoStatus.h"
// #include "/opt/picoscope/include/libps6000/ps6000Api.h"
/*
// Forward declarations
int ps6000LpBlockReady(int16_t handle, PICO_STATUS status, void * pParameter);
int ps6000LpDataReady(int16_t handle, PICO_STATUS status, uint32_t noOfSamples,
				int16_t overflow, void * pParameter);
int ps6000LpStreamingReady(int16_t handle, int32_t noOfSamples, uint32_t startIndex,
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
	scopeHandler.Id = "ps6000"
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
		slog.Debug("ps6000EnumerateUnits", "bufferLen", bufferLen)
		stat := C.ps6000EnumerateUnits((*C.short)(&count), cstrPtr, (*C.short)(&serialLth))
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

func openUnit(serial string, resolution int) (handle int16, err error) {
	var p *C.schar
	sLength := len(serial)
	if sLength > 0 {
		p = (*C.schar)(C.CString(serial))
		defer C.free(unsafe.Pointer(p))
	}
	slog.Debug("ps6000OpenUnit", "serial", serial)
	stat := C.ps6000OpenUnit((*C.short)(&handle), (*C.schar)(p))
	if stat != C.PICO_OK {
		err = fmt.Errorf("OpenUnit:  %s", psc.StatStr(int(stat)))
		return
	}
	return
}

func openUnitAsync(serial string, resolution int) (status int16, err error) {
	var p *C.schar
	sLength := len(serial)
	if sLength > 0 {
		p = (*C.schar)(C.CString(serial))
		defer C.free(unsafe.Pointer(p))
	}
	slog.Debug("ps6000OpenUnitAsync", "serial", serial)
	stat := C.ps6000OpenUnitAsync((*C.short)(&status), (*C.schar)(p))
	if stat != C.PICO_OK {
		err = fmt.Errorf("OpenUnitAsync:  %s", psc.StatStr(int(stat)))
		return
	}
	return
}

func openUnitProgress() (handle int16, progressPercent int16, complete int16, err error) {
	stat := C.ps6000OpenUnitProgress((*C.short)(&handle),
		(*C.short)(&progressPercent), (*C.short)(&complete))
	if stat != C.PICO_OK {
		err = fmt.Errorf("OpenUnitProgress:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000CloseUnit(handle int16) (err error) {
	slog.Debug("ps6000CloseUnit", "handle", handle)
	stat := C.ps6000CloseUnit((C.short)(handle))
	if stat != C.PICO_OK {
		err = fmt.Errorf("CloseUnit:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000GetUnitInfo(handle int16, info PicoInfo) (infoString string, err error) {
	const listLen = 4096
	var cstrPtr *C.schar
	cstrPtr = (*C.schar)(C.malloc(C.sizeof_schar * listLen))
	defer C.free(unsafe.Pointer(cstrPtr))
	requiredSize := int16(listLen)
	slog.Debug("ps6000GetUnitInfo", "handle", handle, "info", info)
	stat := C.ps6000GetUnitInfo((C.short)(handle), cstrPtr, (C.short)(requiredSize),
		(*C.short)(&requiredSize), (C.PICO_INFO)(info))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetUnitInfo:  %s", psc.StatStr(int(stat)))
	}
	if requiredSize == 0 {
		infoString = "No answer from ps6000GetUnitInfo "
		return
	}
	b := C.GoBytes(unsafe.Pointer(cstrPtr), (C.int)(requiredSize-1))
	infoString = string(b)
	return
}

func ps6000FlashLed(handle int16, start int16) (err error) {
	slog.Debug("ps6000FlashLed", "handle", handle, "start", start)
	stat := C.ps6000FlashLed((C.short)(handle), (C.short)(start))
	if stat != C.PICO_OK {
		err = fmt.Errorf("FlashLed:  %s", psc.StatStr(int(stat)))
	}
	return
}

var regLpDataReadyGo DataReady // registered go callback function

//export ps6000LpDataReadyGo
func ps6000LpDataReadyGo(handle int16, status int, noOfSamples uint32, overflow int16, param interface{}) {
	if regLpDataReadyGo != nil {
		regLpDataReadyGo(handle, status, noOfSamples, overflow, param) // call registered go callback function
	}
	return
}

func ps6000GetValuesAsync(handle int16, startIndex, noOfSamples, downSampleRatio uint32,
	downSampleRatioMode RatioMode, lpDataReadyGoPar DataReady, segmentIndex uint32,
	param interface{}) (err error) {
	regLpDataReadyGo = lpDataReadyGoPar
	slog.Debug("ps6000GetValuesAsync", "handle", handle, "startIndex", startIndex, "noOfSamples", noOfSamples, "downSampleRatio", downSampleRatio, "downSampleRatioMode", downSampleRatioMode, "lpDataReadyGoPar", lpDataReadyGoPar, "segmentIndex", segmentIndex, "param", param)
	stat := C.ps6000GetValuesAsync((C.short)(handle),
		(C.uint32_t)(startIndex),
		(C.uint32_t)(noOfSamples),
		(C.uint32_t)(downSampleRatio),
		(C.PS6000_RATIO_MODE)(downSampleRatioMode),
		(C.uint32_t)(segmentIndex),
		(C.ps6000LpDataReady), // C callback function in callbacks.go
		unsafe.Pointer(&param))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetValuesAsync:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000GetValues(handle int16, startIndex, reqNoOfSamples, downSampleRatio uint32,
	downSampleRatioMode RatioMode, segmentIndex uint32) (noOfSamples uint32, overflow int16, err error) {
	slog.Debug("ps6000GetValues", "handle", handle, "startIndex", startIndex, "reqNoOfSamples", reqNoOfSamples, "downSampleRatio", downSampleRatio, "downSampleRatioMode", downSampleRatioMode, "segmentIndex", segmentIndex)
	stat := C.ps6000GetValues((C.short)(handle),
		(C.uint)(startIndex),
		(*C.uint)(&reqNoOfSamples),
		(C.uint)(downSampleRatio),
		(C.PS6000_RATIO_MODE)(downSampleRatioMode),
		(C.uint)(segmentIndex),
		(*C.short)(&overflow))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetValues:  %s", psc.StatStr(int(stat)))
	}
	noOfSamples = reqNoOfSamples
	return
}

func ps6000GetValuesBulk(handle int16, reqNoOfSamples uint32, fromSegmentIndex, toSegmentIndex, downSampleRatio uint32,
	downSampleRatioMode RatioMode, overflow []int16) (noSamples uint32, err error) {
	slog.Debug("ps6000GetValuesBulk", "handle", handle, "reqNoOfSamples", reqNoOfSamples, "fromSegmentIndex", fromSegmentIndex, "toSegmentIndex", toSegmentIndex, "downSampleRatio", downSampleRatio, "downSampleRatioMode", downSampleRatioMode, "overflow", overflow)
	stat := C.ps6000GetValuesBulk((C.short)(handle),
		(*C.uint)(&reqNoOfSamples),
		(C.uint)(fromSegmentIndex),
		(C.uint)(toSegmentIndex),
		(C.uint)(downSampleRatio),
		(C.PS6000_RATIO_MODE)(downSampleRatioMode),
		(*C.short)(&overflow[0]))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetValuesBulk:  %s", psc.StatStr(int(stat)))
	}
	noSamples = reqNoOfSamples
	return
}

func ps6000GetValuesOverlapped(handle int16, startIndex, reqNoOfSamples, downSampleRatio uint32,
	downSampleRatioMode RatioMode, segmentIndex uint32, overflow []int16) (noSamples uint32, err error) {
	err = fmt.Errorf("GetValuesOverlapped not supported on ps6000")
	return
}

func ps6000GetValuesOverlappedBulk(handle int16, startIndex, reqNoOfSamples, downSampleRatio uint32,
	downSampleRatioMode RatioMode, fromSegmentIndex, toSegmentIndex uint32, overflow []int16) (noSamples uint32, err error) {
	err = fmt.Errorf("GetValuesOverlappedBulk not supported on ps6000")
	return
}

func ps6000GetAnalogueOffset(handle int16, voltageRange int, coupling Coupling) (maximumVoltage, minimumVoltage float32, err error) {
	return 0, 0, fmt.Errorf("GetAnalogueOffset not supported on ps6000")
}

func ps6000GetChannelInformation(handle int16, info int16, probe int32, ranges []int32, channels ChannelId) (lengthOfRanges int32, err error) {
	return 0, fmt.Errorf("GetChannelInformation not supported on ps6000")
}

func ps6000GetMaxDownSampleRatio(handle int16, noOfUnaggregatedSamples uint32, downSampleRatioMode RatioMode, segmentIndex int32) (maxDownSampleRatio uint32, err error) {
	slog.Debug("ps6000GetMaxDownSampleRatio", "handle", handle, "noOfUnaggregatedSamples", noOfUnaggregatedSamples, "downSampleRatioMode", downSampleRatioMode, "segmentIndex", segmentIndex)
	stat := C.ps6000GetMaxDownSampleRatio((C.short)(handle), (C.uint32_t)(noOfUnaggregatedSamples),
		(*C.uint32_t)(&maxDownSampleRatio), (C.PS6000_RATIO_MODE)(downSampleRatioMode), (C.uint32_t)(segmentIndex))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetMaxDownSampleRatio:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000GetMaxSegments(handle int16) (maxSegments uint32, err error) {
	return 0, fmt.Errorf("GetMaxSegments not supported on ps6000")
}

// ps6000ChangePowerSource
// ps6000CurrentPowerSource
func ps6000GetNumOfCaptures(handle int16) (nCaptures uint32, err error) {
	slog.Debug("ps6000GetNoOfCaptures", "handle", handle)
	stat := C.ps6000GetNoOfCaptures((C.short)(handle), (*C.uint)(&nCaptures))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetNoOfCaptures:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000GetNumOfProcessedCaptures(handle int16) (nCaptures uint32, err error) {
	slog.Debug("ps6000GetNoOfProcessedCaptures", "handle", handle)
	stat := C.ps6000GetNoOfProcessedCaptures((C.short)(handle), (*C.uint)(&nCaptures))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetNoOfProcessedCaptures:  %s", psc.StatStr(int(stat)))
	}
	return
}

var regLpStreamingReadyGo StreamingReady // registered go callback function

//export ps6000LpStreamingReadyGo
func ps6000LpStreamingReadyGo(handle int16, noOfSamples int32, startIndex uint32, overflow int16,
	triggeredAt uint32, triggered, autoStop int16, param interface{}) {
	if regLpStreamingReadyGo != nil {
		regLpStreamingReadyGo(handle, noOfSamples, startIndex, overflow, triggeredAt, autoStop, triggered, param) // call registered go callback function
	}
	return
}

func ps6000GetStreamingLatestValues(handle int16, lpStreamingReadyGoPar StreamingReady, param interface{}) (err error) {
	regLpStreamingReadyGo = lpStreamingReadyGoPar
	slog.Debug("ps6000GetStreamingLatestValues", "handle", handle, "lpStreamingReadyGoPar", lpStreamingReadyGoPar, "param", param)
	stat := C.ps6000GetStreamingLatestValues((C.short)(handle),
		(C.ps6000StreamingReady)(C.ps6000LpStreamingReady), // C callback function in callbacks.go
		unsafe.Pointer(&param))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetStreamingLatestValues:  %s", psc.StatStr(int(stat)))
	}
	return
}

// ps6000CheckForUpdate
// ps6000StartFirmwareUpdate
func ps6000GetTimebase(handle int16, timeBase uint32, noOfSamples uint32, overSample int16, segmentIndex uint32) (timeIntervalNanoseconds int32, maxSamples uint32, err error) {
	slog.Debug("ps6000GetTimebase", "handle", handle, "timeBase", timeBase, "noOfSamples", noOfSamples, "overSample", overSample, "segmentIndex", segmentIndex)
	stat := C.ps6000GetTimebase((C.short)(handle), (C.uint32_t)(timeBase), (C.uint32_t)(noOfSamples),
		(*C.int32_t)(&timeIntervalNanoseconds), (C.short)(overSample),
		(*C.uint32_t)(&maxSamples), (C.uint)(segmentIndex))
	if stat != C.PICO_OK {
		slog.Error("GetTimebase", "noOfSamples", noOfSamples, "stat", psc.StatStr(int(stat)))
		err = fmt.Errorf("GetTimebase:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000GetTimebase2(handle int16, timeBase uint32, numOfSamples uint32,
	overSample int16, segmentIndex uint32) (timeIntervalNanoseconds float32, maxSamples int32, err error) {
	slog.Debug("ps6000GetTimebase2", "handle", handle, "timeBase", timeBase, "numOfSamples", numOfSamples, "overSample", overSample, "segmentIndex", segmentIndex)
	stat := C.ps6000GetTimebase2((C.short)(handle), (C.uint32_t)(timeBase), (C.uint32_t)(numOfSamples),
		(*C.float)(&timeIntervalNanoseconds), (C.short)(overSample),
		(*C.uint32_t)(&maxSamples), (C.uint32_t)(segmentIndex))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetTimebase2:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000SetChannel(handle int16, channel ChannelId, enabled bool, couplingType Coupling, voltageRange RangeEnum, analogOffset float32) (err error) {
	slog.Debug("ps6000SetChannel", "handle", handle, "channel", channel, "enabled", enabled, "couplingType", couplingType, "voltageRange", voltageRange)
	stat := C.ps6000SetChannel((C.short)(handle), (C.PS6000_CHANNEL)(channel), (C.short)(boolToint16(enabled)), (C.PS6000_COUPLING)(couplingType), (C.PS6000_RANGE)(voltageRange), (C.float)(analogOffset), (C.PS6000_BANDWIDTH_LIMITER)(0))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetChannel:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000MaximumValue(handle int16) (value int32, err error) {
	return 32767, nil
}

func ps6000MinimumValue(handle int16) (value int32, err error) {
	return -32767, nil
}

func ps6000SetSimpleTrigger(handle int16, enable bool, source ChannelId, threshold int16,
	direction ThresholdDirection, delay uint32, autoTriggerMs int16) (err error) {
	slog.Debug("ps6000SetSimpleTrigger", "handle", handle, "enable", enable, "src", source, "threshold", threshold, "direction", direction, "delay", delay, "autoTriggerMs", autoTriggerMs)
	stat := C.ps6000SetSimpleTrigger((C.short)(handle), (C.short)(boolToint16(enable)),
		(C.PS6000_CHANNEL)(source), (C.short)(threshold),
		(C.PS6000_THRESHOLD_DIRECTION)(direction), (C.uint)(delay),
		(C.short)(autoTriggerMs))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetSimpleTrigger:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000SetDataBuffer(handle int16, ch ChannelId, bufferIn []int16, segmentIndex uint32,
	mode RatioMode) (err error) {

	slog.Debug("ps6000SetDataBuffer", "handle", handle, "ch", ch, "segmentIndex", segmentIndex, "mode", mode)
	stat := C.ps6000SetDataBufferBulk((C.short)(handle), (C.PS6000_CHANNEL)(ch), (*C.short)(&bufferIn[0]),
		(C.uint32_t)(len(bufferIn)), (C.uint32_t)(segmentIndex),
		(C.PS6000_RATIO_MODE)(mode))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetDataBuffer:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000SetDataBuffers(handle int16, ch ChannelId, bufferMax, bufferMin []int16, segmentIndex uint32, mode RatioMode) (err error) {
	slog.Debug("ps6000SetDataBuffers", "handle", handle, "ch", ch, "segmentIndex", segmentIndex, "mode", mode)
	stat := C.ps6000SetDataBuffersBulk((C.short)(handle), (C.PS6000_CHANNEL)(ch), (*C.short)(&bufferMax[0]),
		(*C.short)(&bufferMin[0]), (C.uint32_t)(len(bufferMax)), (C.uint32_t)(segmentIndex),
		(C.PS6000_RATIO_MODE)(mode))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetDataBuffers:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000SetUnscaledDataBuffers(handle int16, ch ChannelId, bufferMax, bufferMin []int16, segmentIndex uint32, mode RatioMode) (err error) {
	slog.Debug("ps6000SetUnscaledDataBuffers", "handle", handle, "ch", ch, "segmentIndex", segmentIndex, "mode", mode)
	stat := C.ps6000SetDataBuffersBulk((C.short)(handle), (C.PS6000_CHANNEL)(ch), (*C.short)(&bufferMax[0]),
		(*C.short)(&bufferMin[0]), (C.uint32_t)(len(bufferMax)), (C.uint32_t)(segmentIndex),
		(C.PS6000_RATIO_MODE)(mode))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetUnscaledDataBuffers:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000SetEtsTimeBuffer(handle int16, buffer []int64) (err error) {
	slog.Debug("ps6000SetEtsTimeBuffer", "handle", handle, "buffer", buffer)
	stat := C.ps6000SetEtsTimeBuffer((C.short)(handle), (*C.long)(&buffer[0]),
		(C.int)(len(buffer)))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetEtsTimeBuffer:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000SetEtsTimeBuffers(handle int16, timeUpper, timeLower []uint32) (err error) {
	slog.Debug("ps6000SetEtsTimeBuffers", "handle", handle, "timeUpper", timeUpper, "timeLower", timeLower)
	stat := C.ps6000SetEtsTimeBuffers((C.short)(handle), (*C.uint)(&timeUpper[0]),
		(*C.uint)(&timeLower[0]), (C.int)(len(timeUpper)))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetEtsTimeBuffer:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000SetEts(handle int16, mode EtsMode, etsCycles int16, etsInterLeave int16) (sampleTimePicoseconds int32, err error) {
	slog.Debug("ps6000SetEts", "handle", handle, "mode", mode, "etsCycles", etsCycles, "etsInterLeave", etsInterLeave)
	stat := C.ps6000SetEts((C.short)(handle), (C.PS6000_ETS_MODE)(mode),
		(C.short)(etsCycles), (C.short)(etsInterLeave), (*C.int)(&sampleTimePicoseconds))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetEts:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000RunStreaming(handle int16, reqSampleInterval uint32, sampleIntervalTimeUnits TimeUnits,
	maxPreTriggerSamples, maxPostTriggerSamples uint32,
	autoStop bool, downSampleRatio uint32, downSampleRatioMode RatioMode,
	overviewBufferSize uint32) (sampleInterval uint32, err error) {
	slog.Debug("ps6000RunStreaming", "handle", handle, "reqSampleInterval", reqSampleInterval, "sampleIntervalTimeUnits", sampleIntervalTimeUnits, "maxPreTriggerSamples", maxPreTriggerSamples, "maxPostTriggerSamples", maxPostTriggerSamples, "autoStop", autoStop, "downSampleRatio", downSampleRatio, "downSampleRatioMode", downSampleRatioMode, "overviewBufferSize", overviewBufferSize)
	stat := C.ps6000RunStreaming((C.short)(handle), (*C.uint)(&reqSampleInterval),
		(C.PS6000_TIME_UNITS)(sampleIntervalTimeUnits), (C.uint)(maxPreTriggerSamples),
		(C.uint)(maxPostTriggerSamples), (C.short)(boolToint16(autoStop)), (C.uint)(downSampleRatio),
		(C.int16_t)(downSampleRatioMode), (C.uint)(overviewBufferSize))
	if stat != C.PICO_OK {
		err = fmt.Errorf("RunStreaming:  %s", psc.StatStr(int(stat)))
	}
	sampleInterval = reqSampleInterval
	return
}

var regLpBlockReadyGo BlockReady // registered go callback function

//export ps6000LpBlockReadyGo
func ps6000LpBlockReadyGo(handle int16, status int, noOfSamples uint32, overflow int16, param interface{}) {
	if regLpBlockReadyGo != nil {
		regLpBlockReadyGo(handle, status, param) // call registered go callback function
	}
	return
}

func ps6000RunBlock(handle int16, noOfPreTriggerSamples, noOfPostTriggerSamples int32,
	timeBase uint32, overSample int16, segmentIndex uint32, lpBlockReadyGoPar BlockReady,
	param interface{}) (timeIndisposedMs int32, err error) {
	regLpBlockReadyGo = lpBlockReadyGoPar
	slog.Debug("ps6000RunBlock", "handle", handle, "noOfPreTriggerSamples", noOfPreTriggerSamples, "noOfPostTriggerSamples", noOfPostTriggerSamples, "timeBase", timeBase, "overSample", overSample, "segmentIndex", segmentIndex, "lpBlockReadyGoPar", lpBlockReadyGoPar, "param", param)
	stat := C.ps6000RunBlock((C.short)(handle), (C.uint32_t)(noOfPreTriggerSamples),
		(C.uint32_t)(noOfPostTriggerSamples), (C.uint32_t)(timeBase), (C.short)(overSample),
		(*C.int)(&timeIndisposedMs), (C.uint)(segmentIndex), (C.ps6000BlockReady)(C.ps6000LpBlockReady),
		unsafe.Pointer(&param))
	if stat != C.PICO_OK {
		err = fmt.Errorf("RunBlock:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000SetTriggerChannelProperties(handle int16, channelProperties []TriggerChannelProperties, auxOutputEnable bool,
	autoTriggerMs int32) (err error) {
	var cTriggerChannelProperties []C.PS6000_TRIGGER_CHANNEL_PROPERTIES
	if len(channelProperties) > 0 {
		cTriggerChannelProperties = make([]C.PS6000_TRIGGER_CHANNEL_PROPERTIES, len(channelProperties))
		for i := range channelProperties {
			cTriggerChannelProperties[i].channel = (C.PS6000_CHANNEL)(channelProperties[i].Channel)
			cTriggerChannelProperties[i].thresholdLowerHysteresis = (C.ushort)(channelProperties[i].ThresholdLowerHysteresis)
			cTriggerChannelProperties[i].thresholdLower = (C.short)(channelProperties[i].ThresholdLower)
			cTriggerChannelProperties[i].thresholdUpperHysteresis = (C.ushort)(channelProperties[i].ThresholdUpperHysteresis)
			cTriggerChannelProperties[i].thresholdUpper = (C.short)(channelProperties[i].ThresholdUpper)
		}
	}
	pcTriggerChannelProperties := (*C.PS6000_TRIGGER_CHANNEL_PROPERTIES)(nil)
	if len(channelProperties) > 0 {
		pcTriggerChannelProperties = &cTriggerChannelProperties[0]
	}
	slog.Debug("ps6000SetTriggerChannelProperties", "handle", handle, "channelProperties", channelProperties, "auxOutputEnable", auxOutputEnable, "autoTriggerMs", autoTriggerMs)
	stat := C.ps6000SetTriggerChannelProperties((C.short)(handle),
		(*C.PS6000_TRIGGER_CHANNEL_PROPERTIES)(pcTriggerChannelProperties),
		(C.short)(len(channelProperties)), (C.short)(boolToint16(auxOutputEnable)), (C.int)(autoTriggerMs))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetTriggerChannelProperties:  %s", psc.StatStr(int(stat)))
	}
	return

}

func ps6000SetTriggerChannelConditions(handle int16, triggerConditions []TriggerConditions) (err error) {
	cTriggerConditions := make([]C.PS6000_TRIGGER_CONDITIONS, len(triggerConditions))
	for i := range triggerConditions {
		cTriggerConditions[i].channelA = (C.PS6000_TRIGGER_STATE)(triggerConditions[i].ChannelA)
		cTriggerConditions[i].channelB = (C.PS6000_TRIGGER_STATE)(triggerConditions[i].ChannelB)
		cTriggerConditions[i].channelC = (C.PS6000_TRIGGER_STATE)(triggerConditions[i].ChannelC)
		cTriggerConditions[i].channelD = (C.PS6000_TRIGGER_STATE)(triggerConditions[i].ChannelD)
		cTriggerConditions[i].external = (C.PS6000_TRIGGER_STATE)(triggerConditions[i].External)
		cTriggerConditions[i].aux = (C.PS6000_TRIGGER_STATE)(triggerConditions[i].Aux)
		cTriggerConditions[i].pulseWidthQualifier = (C.PS6000_TRIGGER_STATE)(triggerConditions[i].PulseWidthQualifier)
	}
	pcTriggerConditions := (*C.PS6000_TRIGGER_CONDITIONS)(nil)
	if len(triggerConditions) > 0 {
		pcTriggerConditions = &cTriggerConditions[0]
	}
	slog.Debug("ps6000SetTriggerChannelConditions", "handle", handle, "triggerConditions", triggerConditions)
	stat := C.ps6000SetTriggerChannelConditions((C.short)(handle),
		(*C.PS6000_TRIGGER_CONDITIONS)(pcTriggerConditions),
		(C.short)(len(triggerConditions)))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetTriggerChannelConditions:  %s", psc.StatStr(int(stat)))
	}

	return
}

func ps6000SetTriggerChannelDirections(handle int16, channelA, channelB, channelC, channelD, ext, aux ThresholdDirection) (err error) {
	slog.Debug("ps6000SetTriggerChannelDirections", "handle", handle, "channelA", channelA, "channelB", channelB, "channelC", channelC, "channelD", channelD, "ext", ext, "aux", aux)
	stat := C.ps6000SetTriggerChannelDirections((C.short)(handle),
		(C.PS6000_THRESHOLD_DIRECTION)(channelA),
		(C.PS6000_THRESHOLD_DIRECTION)(channelB),
		(C.PS6000_THRESHOLD_DIRECTION)(channelC),
		(C.PS6000_THRESHOLD_DIRECTION)(channelD),
		(C.PS6000_THRESHOLD_DIRECTION)(ext),
		(C.PS6000_THRESHOLD_DIRECTION)(aux))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetTriggerChannelDirections:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000SetTriggerDelay(handle int16, delay uint32) (err error) {
	slog.Debug("ps6000SetTriggerDelay", "handle", handle, "delay", delay)
	stat := C.ps6000SetTriggerDelay((C.short)(handle), (C.uint)(delay))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetTriggerDelay:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000SetPulseWidthQualifier(handle int16, conditions []PwqConditions, direction ThresholdDirection, lower, upper uint32,
	pwType PulseWidthType) (err error) {
	cPwqConditions := make([]C.PS6000_PWQ_CONDITIONS, len(conditions))
	for i := range conditions {
		cPwqConditions[i].channelA = (C.PS6000_TRIGGER_STATE)(conditions[i].ChannelA)
		cPwqConditions[i].channelB = (C.PS6000_TRIGGER_STATE)(conditions[i].ChannelB)
		cPwqConditions[i].channelC = (C.PS6000_TRIGGER_STATE)(conditions[i].ChannelC)
		cPwqConditions[i].channelD = (C.PS6000_TRIGGER_STATE)(conditions[i].ChannelD)
		cPwqConditions[i].external = (C.PS6000_TRIGGER_STATE)(conditions[i].External)
	}
	slog.Debug("ps6000SetPulseWidthQualifier", "handle", handle, "conditions", conditions, "direction", direction, "lower", lower, "upper", upper, "pwType", pwType)
	stat := C.ps6000SetPulseWidthQualifier((C.short)(handle),
		(*C.PS6000_PWQ_CONDITIONS)(&cPwqConditions[0]), (C.short)(len(conditions)),
		(C.PS6000_THRESHOLD_DIRECTION)(direction), (C.uint)(lower), (C.uint)(upper),
		(C.PS6000_PULSE_WIDTH_TYPE)(pwType))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetPulseWidthQualifier:  %s", psc.StatStr(int(stat)))
	}
	return
}
func ps6000SetTriggerDigitalPortProperties(handle int16, digitalDirections []DigitalChannelDirections) (err error) {
	return fmt.Errorf("SetTriggerDigitalPortProperties not supported on ps6000")
}

func ps6000Stop(handle int16) (err error) {
	slog.Debug("ps6000Stop", "handle", handle)
	stat := C.ps6000Stop((C.short)(handle))
	if stat != C.PICO_OK {
		err = fmt.Errorf("Stop:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000SetSigGenBuiltIn(handle int16, offsetVoltage int32, pkToPK uint32, waveType WaveTypeEnum,
	startFrequency, stopFrequency, increment, dwellTime float32, sweepType SweepTypeEnum,
	operation ExtraOperations, shots, sweeps uint32, triggerType SigGenTrigType,
	triggerSource SigGenTrigSource, extInThreshold int16) (err error) {
	slog.Debug("ps6000SetSigGenBuiltIn", "handle", handle, "offsetVoltage", offsetVoltage, "pkToPK", pkToPK, "waveType", waveType, "startFrequency", startFrequency, "stopFrequency", stopFrequency, "increment", increment, "dwellTime", dwellTime, "sweepType", sweepType, "operation", operation, "shots", shots, "sweeps", sweeps, "triggerType", triggerType, "triggerSource", triggerSource, "extInThreshold", extInThreshold)
	stat := C.ps6000SetSigGenBuiltIn((C.short)(handle), (C.int)(offsetVoltage),
		(C.uint)(pkToPK), (C.short)(waveType), (C.float)(startFrequency),
		(C.float)(stopFrequency), (C.float)(increment), (C.float)(dwellTime),
		(C.PS6000_SWEEP_TYPE)(sweepType), (C.int)(operation),
		(C.uint)(shots), (C.uint)(sweeps), (C.PS6000_SIGGEN_TRIG_TYPE)(triggerType),
		(C.PS6000_SIGGEN_TRIG_SOURCE)(triggerSource), (C.short)(extInThreshold))

	if stat != C.PICO_OK {
		err = fmt.Errorf("SetSigGenBuiltIn:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000SetSigGenBuiltInV2(handle int16, offsetVoltage int32, pkToPK uint32, waveType WaveTypeEnum,
	startFrequency, stopFrequency, increment, dwellTime float64, sweepType SweepTypeEnum,
	operation ExtraOperations, shots, sweeps uint32, triggerType SigGenTrigType,
	triggerSource SigGenTrigSource, extInThreshold int16) (err error) {
	return fmt.Errorf("SetSigGenBuiltInV2 not supported on ps6000")
}

func ps6000SigGenFrequencyToPhase(handle int16, frequency float64, indexMode IndexMode, bufferLength uint32) (phase uint32, err error) {
	slog.Debug("ps6000SigGenFrequencyToPhase", "handle", handle, "frequency", frequency, "indexMode", indexMode, "bufferLength", bufferLength)
	stat := C.ps6000SigGenFrequencyToPhase((C.short)(handle), (C.double)(frequency),
		(C.PS6000_INDEX_MODE)(indexMode), (C.uint)(bufferLength), (*C.uint)(&phase))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SigGenFrequencyToPhase:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000SetNoCaptures(handle int16, nCaptures uint32) (err error) {
	slog.Debug("ps6000SetNoOfCaptures", "handle", handle, "nCaptures", nCaptures)
	stat := C.ps6000SetNoOfCaptures((C.short)(handle), (C.uint)(nCaptures))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetNoCaptures:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000GetTriggerTimeOffset(handle int16, segmentIndex uint32) (timeUpper,
	timeLower uint32, timeUnits TimeUnits, err error) {
	slog.Debug("ps6000GetTriggerTimeOffset", "handle", handle, "segmentIndex", segmentIndex)
	stat := C.ps6000GetTriggerTimeOffset((C.short)(handle), (*C.uint)(&timeUpper),
		(*C.uint)(&timeLower), (*C.PS6000_TIME_UNITS)(&timeUnits), (C.uint)(segmentIndex))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetTriggerTimeOffset:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000GetTriggerTimeOffset64(handle int16, segmentIndex uint32) (time int64, timeUnits TimeUnits, err error) {
	slog.Debug("ps6000GetTriggerTimeOffset64", "handle", handle, "segmentIndex", segmentIndex)
	stat := C.ps6000GetTriggerTimeOffset64((C.short)(handle), (*C.long)(&time),
		(*C.PS6000_TIME_UNITS)(&timeUnits), (C.uint)(segmentIndex))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetTriggerTimeOffset64:  %s", psc.StatStr(int(stat)))
	}
	return

}

func ps6000GetValuesTriggerTimeOffsetBulk(handle int16, timesUpper, timesLower []uint32, timeUnits []TimeUnits,
	fromSegmentIndex, toSegmentIndex uint32) (err error) {
	slog.Debug("ps6000GetValuesTriggerTimeOffsetBulk", "handle", handle, "timesUpper", timesUpper, "timesLower", timesLower, "timeUnits", timeUnits, "fromSegmentIndex", fromSegmentIndex, "toSegmentIndex", toSegmentIndex)
	stat := C.ps6000GetValuesTriggerTimeOffsetBulk((C.short)(handle), (*C.uint)(&timesUpper[0]),
		(*C.uint)(&timesLower[0]), (*C.PS6000_TIME_UNITS)(&timeUnits[0]), (C.uint)(fromSegmentIndex),
		(C.uint)(toSegmentIndex))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetValuesTriggerTimeOffsetBulk:  %s", psc.StatStr(int(stat)))
	}

	return
}

func ps6000GetValuesTriggerTimeOffsetBulk64(handle int16, times []int64, timeUnits []TimeUnits,
	fromSegmentIndex, toSegmentIndex uint32) (err error) {
	slog.Debug("ps6000GetValuesTriggerTimeOffsetBulk64", "handle", handle, "times", times, "timeUnits", timeUnits, "fromSegmentIndex", fromSegmentIndex, "toSegmentIndex", toSegmentIndex)
	stat := C.ps6000GetValuesTriggerTimeOffsetBulk64((C.short)(handle), (*C.long)(&times[0]),
		(*C.PS6000_TIME_UNITS)(&timeUnits[0]), (C.uint)(fromSegmentIndex),
		(C.uint)(toSegmentIndex))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetValuesTriggerTimeOffsetBulk64:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000HoldOff(handle int16, holdOff uint64, holdOffType HoldOffType) (err error) {
	slog.Debug("ps6000HoldOff", "handle", handle, "holdOff", holdOff, "holdOffType", holdOffType)
	err = fmt.Errorf("ps6000HoldOff not supported on ps6000")
	return
}

func ps6000LsReady(handle int16) (ready int16, err error) {
	slog.Debug("ps6000LsReady", "handle", handle)
	stat := C.ps6000IsReady((C.short)(handle), (*C.short)(&ready))
	if stat != C.PICO_OK {
		err = fmt.Errorf("LsReady:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000TriggerOrPulseWidthQualifierEnabled(handle int16) (triggerEnabled, pulseWidthQualifierEnabledint16 int16, err error) {
	slog.Debug("ps6000TriggerOrPulseWidthQualifierEnabled", "handle", handle)
	stat := C.ps6000IsTriggerOrPulseWidthQualifierEnabled((C.short)(handle),
		(*C.short)(&triggerEnabled), (*C.short)(&pulseWidthQualifierEnabledint16))
	if stat != C.PICO_OK {
		err = fmt.Errorf("TriggerOrPulseWidthQualifierEnabled:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000MemorySegments(handle int16, nSegments uint64) (nMaxSamples int64, err error) {
	slog.Debug("ps6000MemorySegments", "handle", handle, "nSegments", nSegments)
	var maxSamples C.int
	stat := C.ps6000MemorySegments((C.short)(handle), (C.uint)(nSegments), &maxSamples)
	if stat != C.PICO_OK {
		err = fmt.Errorf("MemorySegments:  %s", psc.StatStr(int(stat)))
	}
	nMaxSamples = int64(maxSamples)
	return
}

func ps6000NoOfStreamingValues(handle int16) (noOfValues uint32, err error) {
	slog.Debug("ps6000NoOfStreamingValues", "handle", handle)
	stat := C.ps6000NoOfStreamingValues((C.short)(handle),
		(*C.uint)(&noOfValues))
	if stat != C.PICO_OK {
		err = fmt.Errorf("NoOfStreamingValues:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000PingUnit(handle int16) (err error) {
	slog.Debug("ps6000PingUnit", "handle", handle)
	stat := C.ps6000PingUnit((C.short)(handle))
	if stat != C.PICO_OK {
		err = fmt.Errorf("PingUnit:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000QueryOutputEdgeDetect(handle int16) (state int16, err error) {
	return 0, fmt.Errorf("QueryOutputEdgeDetect not supported on ps6000")
}

func ps6000SetDigitalPort(handle int16, port DigitalPort, enabled bool, logiclevel int16) (err error) {
	return fmt.Errorf("SetDigitalPort not supported on ps6000")
}

func ps6000SetOutputEdgeDetect(handle int16, state int16) (err error) {
	return fmt.Errorf("SetOutputEdgeDetect not supported on ps6000")
}

// ps6000GetScalingValues
func ps6000SetPulseWidthDigitalPortProperties(handle int16, digitalDirections []DigitalChannelDirections) (err error) {
	return fmt.Errorf("SetPulseWidthDigitalPortProperties not supported on ps6000")
}

func ps6000SetSigGenArbitrary(handle int16, offsetVoltage int32, pkToPK uint32,
	startDeltaPhase, stopDeltaPhase, deltaPhaseIncrement, dwellCount uint32,
	arbitraryWaveform []int16, sweepType SweepTypeEnum, operation ExtraOperations,
	indexMode IndexMode, shots, sweeps uint32, triggerType SigGenTrigType,
	triggerSource SigGenTrigSource, extInThreshold int16) (err error) {
	slog.Debug("ps6000SetSigGenArbitrary", "handle", handle, "offsetVoltage", offsetVoltage, "pkToPK", pkToPK, "startDeltaPhase", startDeltaPhase, "stopDeltaPhase", stopDeltaPhase, "deltaPhaseIncrement", deltaPhaseIncrement, "dwellCount", dwellCount, "arbitraryWaveform", arbitraryWaveform, "sweepType", sweepType, "operation", operation, "indexMode", indexMode, "shots", shots, "sweeps", sweeps, "triggerType", triggerType, "triggerSource", triggerSource, "extInThreshold", extInThreshold)
	stat := C.ps6000SetSigGenArbitrary((C.short)(handle), (C.int)(offsetVoltage),
		(C.uint)(pkToPK), (C.uint)(startDeltaPhase), (C.uint)(stopDeltaPhase),
		(C.uint32_t)(deltaPhaseIncrement), (C.uint32_t)(dwellCount),
		(*C.short)(&arbitraryWaveform[0]), (C.int32_t)(len(arbitraryWaveform)),
		(C.PS6000_SWEEP_TYPE)(sweepType), (C.int)(operation),
		(C.PS6000_INDEX_MODE)(indexMode),
		(C.uint)(shots), (C.uint)(sweeps), (C.PS6000_SIGGEN_TRIG_TYPE)(triggerType),
		(C.PS6000_SIGGEN_TRIG_SOURCE)(triggerSource), (C.short)(extInThreshold))

	if stat != C.PICO_OK {
		err = fmt.Errorf("SetSigGenArbitrary:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000SetSigGenPropertiesArbitrary(handle int16, offsetVoltage int32,
	startDeltaPhase, stopDeltaPhase, deltaPhaseIncrement, dwellCount uint32,
	sweepType SweepTypeEnum, operation ExtraOperations,
	indexMode IndexMode, shots, sweeps uint32, triggerType SigGenTrigType,
	triggerSource SigGenTrigSource, extInThreshold int16) (err error) {
	return fmt.Errorf("SetSigGenPropertiesArbitrary not supported on ps6000")
}

func ps6000SetSigGenPropertiesBuiltIn(handle int16, offsetVoltage int32,
	startFrequency, stopFrequency, increment, dwellTime float64,
	sweepType SweepTypeEnum,
	shots, sweeps uint32, triggerType SigGenTrigType,
	triggerSource SigGenTrigSource, extInThreshold int16) (err error) {
	return fmt.Errorf("SetSigGenPropertiesBuiltIn not supported on ps6000")
}

func ps6000SigGenArbitraryMinMaxValues(handle int16) (minArbitraryWaveformValue, maxArbitraryWaveformValue int16,
	minArbitraryWaveformSize, maxArbitraryWaveformSize uint32, err error) {
	slog.Debug("ps6000SigGenArbitraryMinMaxValues", "handle", handle)
	stat := C.ps6000SigGenArbitraryMinMaxValues((C.short)(handle),
		(*C.short)(&minArbitraryWaveformValue), (*C.short)(&maxArbitraryWaveformValue),
		(*C.uint32_t)(&minArbitraryWaveformSize), (*C.uint32_t)(&maxArbitraryWaveformSize))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SigGenArbitraryMinMaxValues:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000SigGenSoftwareControl(handle int16, state int16) (err error) {
	slog.Debug("ps6000SigGenSoftwareControl", "handle", handle, "state", state)
	stat := C.ps6000SigGenSoftwareControl((C.short)(handle),
		(C.short)(state))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetSigGenBuiltIn:  %s", psc.StatStr(int(stat)))
	}
	return
}
