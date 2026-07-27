//go:build !noscope && ps4000

package ps4000

// #cgo CFLAGS: -g -Wall -I/opt/picoscope/include/libps4000
// #cgo LDFLAGS: -L/opt/picoscope/lib/ -lps4000
// #include <stdlib.h>
// #include "/opt/picoscope/include/libps4000/PicoStatus.h"
// #include "/opt/picoscope/include/libps4000/ps4000Api.h"
/*
// Forward declarations
int ps4000LpBlockReady(int16_t handle, PICO_STATUS status, void * pParameter);
int ps4000LpDataReady(int16_t handle, PICO_STATUS status, uint32_t noOfSamples,
				int16_t overflow, void * pParameter);
int ps4000LpStreamingReady(int16_t handle, int32_t noOfSamples, uint32_t startIndex,
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
	scopeHandler.Id = "ps4000"
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
		slog.Debug("ps4000EnumerateUnits", "bufferLen", bufferLen)
		stat := C.ps4000EnumerateUnits((*C.short)(&count), cstrPtr, (*C.short)(&serialLth))
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
	slog.Debug("ps4000OpenUnit")
	stat := C.ps4000OpenUnit((*C.short)(&handle))
	if stat != C.PICO_OK {
		err = fmt.Errorf("OpenUnit:  %s", psc.StatStr(int(stat)))
		return
	}
	return
}

func openUnitAsync(serial string) (status int16, err error) {
	slog.Debug("ps4000OpenUnitAsync")
	stat := C.ps4000OpenUnitAsync((*C.short)(&status))
	if stat != C.PICO_OK {
		err = fmt.Errorf("OpenUnitAsync:  %s", psc.StatStr(int(stat)))
		return
	}
	return
}

func openUnitProgress() (handle int16, progressPercent int16, complete int16, err error) {
	stat := C.ps4000OpenUnitProgress((*C.short)(&handle),
		(*C.short)(&progressPercent), (*C.short)(&complete))
	if stat != C.PICO_OK {
		err = fmt.Errorf("OpenUnitProgress:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps4000CloseUnit(handle int16) (err error) {
	slog.Debug("ps4000CloseUnit", "handle", handle)
	stat := C.ps4000CloseUnit((C.short)(handle))
	if stat != C.PICO_OK {
		err = fmt.Errorf("CloseUnit:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps4000GetUnitInfo(handle int16, info PicoInfo) (infoString string, err error) {
	const listLen = 4096
	var cstrPtr *C.schar
	cstrPtr = (*C.schar)(C.malloc(C.sizeof_schar * listLen))
	defer C.free(unsafe.Pointer(cstrPtr))
	requiredSize := int16(listLen)
	slog.Debug("ps4000GetUnitInfo", "handle", handle, "info", info)
	stat := C.ps4000GetUnitInfo((C.short)(handle), cstrPtr, (C.short)(requiredSize),
		(*C.short)(&requiredSize), (C.PICO_INFO)(info))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetUnitInfo:  %s", psc.StatStr(int(stat)))
	}
	if requiredSize == 0 {
		infoString = "No answer from ps4000GetUnitInfo "
		return
	}
	b := C.GoBytes(unsafe.Pointer(cstrPtr), (C.int)(requiredSize-1))
	infoString = string(b)
	return
}

func ps4000FlashLed(handle int16, start int16) (err error) {
	slog.Debug("ps4000FlashLed", "handle", handle, "start", start)
	stat := C.ps4000FlashLed((C.short)(handle), (C.short)(start))
	if stat != C.PICO_OK {
		err = fmt.Errorf("FlashLed:  %s", psc.StatStr(int(stat)))
	}
	return
}

var regLpDataReadyGo DataReady // registered go callback function

//export ps4000LpDataReadyGo
func ps4000LpDataReadyGo(handle int16, status int, noOfSamples uint32, overflow int16, param interface{}) {
	if regLpDataReadyGo != nil {
		regLpDataReadyGo(handle, status, noOfSamples, overflow, param) // call registered go callback function
	}
	return
}

func ps4000GetValuesAsync(handle int16, startIndex, noOfSamples, downSampleRatio uint32,
	downSampleRatioMode RatioMode, lpDataReadyGoPar DataReady, segmentIndex uint32,
	param interface{}) (err error) {
	regLpDataReadyGo = lpDataReadyGoPar
	slog.Debug("ps4000GetValuesAsync", "handle", handle, "startIndex", startIndex, "noOfSamples", noOfSamples, "downSampleRatio", downSampleRatio, "downSampleRatioMode", downSampleRatioMode, "lpDataReadyGoPar", lpDataReadyGoPar, "segmentIndex", segmentIndex, "param", param)
	stat := C.ps4000GetValuesAsync((C.short)(handle),
		(C.uint)(startIndex),
		(C.uint)(noOfSamples),
		(C.uint)(downSampleRatio),
		(C.int16_t)(downSampleRatioMode),
		(C.uint16_t)(segmentIndex),
		(C.ps4000LpDataReady), // C callback function in callbacks.go
		unsafe.Pointer(&param))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetValuesAsync:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps4000GetValues(handle int16, startIndex, reqNoOfSamples, downSampleRatio uint32,
	downSampleRatioMode RatioMode, segmentIndex uint32) (noOfSamples uint32, overflow int16, err error) {
	slog.Debug("ps4000GetValues", "handle", handle, "startIndex", startIndex, "reqNoOfSamples", reqNoOfSamples, "downSampleRatio", downSampleRatio, "downSampleRatioMode", downSampleRatioMode, "segmentIndex", segmentIndex)
	stat := C.ps4000GetValues((C.short)(handle),
		(C.uint)(startIndex),
		(*C.uint)(&reqNoOfSamples),
		(C.uint)(downSampleRatio),
		(C.int16_t)(downSampleRatioMode),
		(C.uint16_t)(segmentIndex),
		(*C.short)(&overflow))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetValues:  %s", psc.StatStr(int(stat)))
	}
	noOfSamples = reqNoOfSamples
	return
}

func ps4000GetValuesBulk(handle int16, reqNoOfSamples uint32, fromSegmentIndex, toSegmentIndex, downSampleRatio uint32,
	downSampleRatioMode RatioMode, overflow []int16) (noSamples uint32, err error) {
	slog.Debug("ps4000GetValuesBulk", "handle", handle, "reqNoOfSamples", reqNoOfSamples, "fromSegmentIndex", fromSegmentIndex, "toSegmentIndex", toSegmentIndex, "downSampleRatio", downSampleRatio, "downSampleRatioMode", downSampleRatioMode, "overflow", overflow)
	stat := C.ps4000GetValuesBulk((C.short)(handle),
		(*C.uint)(&reqNoOfSamples),
		(C.uint16_t)(fromSegmentIndex),
		(C.uint16_t)(toSegmentIndex),
		(*C.short)(&overflow[0]))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetValuesBulk:  %s", psc.StatStr(int(stat)))
	}
	noSamples = reqNoOfSamples
	return
}

func ps4000GetValuesOverlapped(handle int16, startIndex, reqNoOfSamples, downSampleRatio uint32,
	downSampleRatioMode RatioMode, segmentIndex uint32, overflow []int16) (noSamples uint32, err error) {
	return 0, fmt.Errorf("GetValuesOverlapped not supported on ps4000")
}

func ps4000GetValuesOverlappedBulk(handle int16, startIndex, reqNoOfSamples, downSampleRatio uint32,
	downSampleRatioMode RatioMode, fromSegmentIndex, toSegmentIndex uint32, overflow []int16) (noSamples uint32, err error) {
	return 0, fmt.Errorf("GetValuesOverlappedBulk not supported on ps4000")
}

func ps4000GetAnalogueOffset(handle int16, voltageRange int, coupling Coupling) (maximumVoltage, minimumVoltage float32, err error) {
	return 0, 0, fmt.Errorf("GetAnalogueOffset not supported on ps4000")
}

func ps4000GetChannelInformation(handle int16, info int16, probe int32, ranges []int32, channels ChannelId) (lengthOfRanges int32, err error) {
	lengthOfRanges = int32(len(ranges))
	slog.Debug("ps4000GetChannelInformation", "handle", handle, "info", info, "probe", probe, "ranges", ranges, "channels", channels)
	stat := C.ps4000GetChannelInformation((C.short)(handle), (C.PS4000_CHANNEL_INFO)(info),
		(C.int)(probe), (*C.int)(&ranges[0]), (*C.int)(&lengthOfRanges), (C.int)(channels))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetChannelInformation:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps4000GetMaxDownSampleRatio(handle int16, noOfUnaggregatedSamples uint32, downSampleRatioMode RatioMode, segmentIndex int32) (maxDownSampleRatio uint32, err error) {
	slog.Debug("ps4000GetMaxDownSampleRatio", "handle", handle, "noOfUnaggregatedSamples", noOfUnaggregatedSamples, "downSampleRatioMode", downSampleRatioMode, "segmentIndex", segmentIndex)
	stat := C.ps4000GetMaxDownSampleRatio((C.short)(handle), (C.uint32_t)(noOfUnaggregatedSamples),
		(*C.uint32_t)(&maxDownSampleRatio), (C.int16_t)(downSampleRatioMode), (C.uint16_t)(segmentIndex))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetMaxDownSampleRatio:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps4000GetMaxSegments(handle int16) (maxSegments uint32, err error) {
	return 0, fmt.Errorf("GetMaxSegments not supported on ps4000")
}

// ps4000ChangePowerSource
// ps4000CurrentPowerSource
func ps4000GetNumOfCaptures(handle int16) (nCaptures uint32, err error) {
	slog.Debug("ps4000GetNoOfCaptures", "handle", handle)
	var captures uint16
	stat := C.ps4000GetNoOfCaptures((C.short)(handle), (*C.uint16_t)(&captures))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetNoOfCaptures:  %s", psc.StatStr(int(stat)))
	}
	nCaptures = uint32(captures)
	return
}

func ps4000GetNumOfProcessedCaptures(handle int16) (nCaptures uint32, err error) {
	return 0, fmt.Errorf("GetNoOfProcessedCaptures is not implemented")
}

var regLpStreamingReadyGo StreamingReady // registered go callback function

//export ps4000LpStreamingReadyGo
func ps4000LpStreamingReadyGo(handle int16, noOfSamples int32, startIndex uint32, overflow int16,
	triggeredAt uint32, triggered, autoStop int16, param interface{}) {
	if regLpStreamingReadyGo != nil {
		regLpStreamingReadyGo(handle, noOfSamples, startIndex, overflow, triggeredAt, autoStop, triggered, param) // call registered go callback function
	}
	return
}

func ps4000GetStreamingLatestValues(handle int16, lpStreamingReadyGoPar StreamingReady, param interface{}) (err error) {
	regLpStreamingReadyGo = lpStreamingReadyGoPar
	slog.Debug("ps4000GetStreamingLatestValues", "handle", handle, "lpStreamingReadyGoPar", lpStreamingReadyGoPar, "param", param)
	stat := C.ps4000GetStreamingLatestValues((C.short)(handle),
		(C.ps4000StreamingReady)(C.ps4000LpStreamingReady), // C callback function in callbacks.go
		unsafe.Pointer(&param))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetStreamingLatestValues:  %s", psc.StatStr(int(stat)))
	}
	return
}

// ps4000CheckForUpdate
// ps4000StartFirmwareUpdate
func ps4000GetTimebase(handle int16, timeBase uint32, noOfSamples int32, overSample int16, segmentIndex uint32) (timeIntervalNanoseconds, maxSamples int32, err error) {
	slog.Debug("ps4000GetTimebase", "handle", handle, "timeBase", timeBase, "noOfSamples", noOfSamples, "overSample", overSample, "segmentIndex", segmentIndex)
	stat := C.ps4000GetTimebase((C.short)(handle), (C.uint)(timeBase), (C.int)(noOfSamples),
		(*C.int)(&timeIntervalNanoseconds), (C.short)(overSample),
		(*C.int)(&maxSamples), (C.uint16_t)(segmentIndex))
	if stat != C.PICO_OK {
		slog.Error("GetTimebase", "noOfSamples", noOfSamples, "stat", psc.StatStr(int(stat)))
		err = fmt.Errorf("GetTimebase:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps4000GetTimebase2(handle int16, timeBase uint32, numOfSamples int32,
	overSample int16, segmentIndex uint32) (timeIntervalNanoseconds float32, maxSamples int32, err error) {
	slog.Debug("ps4000GetTimebase2", "handle", handle, "timeBase", timeBase, "numOfSamples", numOfSamples, "overSample", overSample, "segmentIndex", segmentIndex)
	stat := C.ps4000GetTimebase2((C.short)(handle), (C.uint)(timeBase), (C.int)(numOfSamples),
		(*C.float)(&timeIntervalNanoseconds), (C.short)(overSample),
		(*C.int32_t)(&maxSamples), (C.uint16_t)(segmentIndex))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetTimebase2:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps4000SetChannel(handle int16, channel ChannelId, enabled bool, couplingType Coupling, voltageRange RangeEnum, analogOffset float32) (err error) {
	slog.Debug("ps4000SetChannel", "handle", handle, "channel", channel, "enabled", enabled, "couplingType", couplingType, "voltageRange", voltageRange)
	stat := C.ps4000SetChannel((C.short)(handle), (C.PS4000_CHANNEL)(channel), (C.short)(boolToint16(enabled)), (C.short)(couplingType), (C.PS4000_RANGE)(voltageRange))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetChannel:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps4000MaximumValue(handle int16) (value int32, err error) {
	return 32767, nil
}

func ps4000MinimumValue(handle int16) (value int32, err error) {
	return -32767, nil
}

func ps4000SetSimpleTrigger(handle int16, enable bool, source ChannelId, threshold int16,
	direction ThresholdDirection, delay uint32, autoTriggerMs int16) (err error) {
	slog.Debug("ps4000SetSimpleTrigger", "handle", handle, "enable", enable, "src", source, "threshold", threshold, "direction", direction, "delay", delay, "autoTriggerMs", autoTriggerMs)
	stat := C.ps4000SetSimpleTrigger((C.short)(handle), (C.short)(boolToint16(enable)),
		(C.PS4000_CHANNEL)(source), (C.short)(threshold),
		(C.THRESHOLD_DIRECTION)(direction), (C.uint)(delay),
		(C.short)(autoTriggerMs))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetSimpleTrigger:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps4000SetDataBuffer(handle int16, ch ChannelId, bufferIn []int16, segmentIndex uint32,
	mode RatioMode) (err error) {

	slog.Debug("ps4000SetDataBuffer", "handle", handle, "ch", ch, "segmentIndex", segmentIndex, "mode", mode)
	stat := C.ps4000SetDataBuffer((C.short)(handle), (C.PS4000_CHANNEL)(ch), (*C.short)(&bufferIn[0]),
		(C.int)(len(bufferIn)))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetDataBuffer:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps4000SetDataBuffers(handle int16, ch ChannelId, bufferMax, bufferMin []int16, segmentIndex uint32, mode RatioMode) (err error) {
	slog.Debug("ps4000SetDataBuffers", "handle", handle, "ch", ch, "segmentIndex", segmentIndex, "mode", mode)
	stat := C.ps4000SetDataBuffers((C.short)(handle), (C.PS4000_CHANNEL)(ch), (*C.short)(&bufferMax[0]),
		(*C.short)(&bufferMin[0]), (C.int)(len(bufferMax)))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetDataBuffers:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps4000SetUnscaledDataBuffers(handle int16, ch ChannelId, bufferMax, bufferMin []int16, segmentIndex uint32, mode RatioMode) (err error) {
	slog.Debug("ps4000SetUnscaledDataBuffers", "handle", handle, "ch", ch, "segmentIndex", segmentIndex, "mode", mode)
	stat := C.ps4000SetDataBuffers((C.short)(handle), (C.PS4000_CHANNEL)(ch), (*C.short)(&bufferMax[0]),
		(*C.short)(&bufferMin[0]), (C.int)(len(bufferMax)))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetUnscaledDataBuffers:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps4000SetEtsTimeBuffer(handle int16, buffer []int64) (err error) {
	slog.Debug("ps4000SetEtsTimeBuffer", "handle", handle, "buffer", buffer)
	stat := C.ps4000SetEtsTimeBuffer((C.short)(handle), (*C.long)(&buffer[0]),
		(C.int)(len(buffer)))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetEtsTimeBuffer:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps4000SetEtsTimeBuffers(handle int16, timeUpper, timeLower []uint32) (err error) {
	slog.Debug("ps4000SetEtsTimeBuffers", "handle", handle, "timeUpper", timeUpper, "timeLower", timeLower)
	stat := C.ps4000SetEtsTimeBuffers((C.short)(handle), (*C.uint)(&timeUpper[0]),
		(*C.uint)(&timeLower[0]), (C.int)(len(timeUpper)))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetEtsTimeBuffer:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps4000SetEts(handle int16, mode EtsMode, etsCycles int16, etsInterLeave int16) (sampleTimePicoseconds int32, err error) {
	slog.Debug("ps4000SetEts", "handle", handle, "mode", mode, "etsCycles", etsCycles, "etsInterLeave", etsInterLeave)
	stat := C.ps4000SetEts((C.short)(handle), (C.PS4000_ETS_MODE)(mode),
		(C.short)(etsCycles), (C.short)(etsInterLeave), (*C.int)(&sampleTimePicoseconds))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetEts:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps4000RunStreaming(handle int16, reqSampleInterval uint32, sampleIntervalTimeUnits TimeUnits,
	maxPreTriggerSamples, maxPostTriggerSamples uint32,
	autoStop bool, downSampleRatio uint32, downSampleRatioMode RatioMode,
	overviewBufferSize uint32) (sampleInterval uint32, err error) {
	slog.Debug("ps4000RunStreaming", "handle", handle, "reqSampleInterval", reqSampleInterval, "sampleIntervalTimeUnits", sampleIntervalTimeUnits, "maxPreTriggerSamples", maxPreTriggerSamples, "maxPostTriggerSamples", maxPostTriggerSamples, "autoStop", autoStop, "downSampleRatio", downSampleRatio, "downSampleRatioMode", downSampleRatioMode, "overviewBufferSize", overviewBufferSize)
	stat := C.ps4000RunStreaming((C.short)(handle), (*C.uint)(&reqSampleInterval),
		(C.PS4000_TIME_UNITS)(sampleIntervalTimeUnits), (C.uint)(maxPreTriggerSamples),
		(C.uint)(maxPostTriggerSamples), (C.short)(boolToint16(autoStop)), (C.uint)(downSampleRatio),
		(C.uint)(overviewBufferSize))
	if stat != C.PICO_OK {
		err = fmt.Errorf("RunStreaming:  %s", psc.StatStr(int(stat)))
	}
	sampleInterval = reqSampleInterval
	return
}

var regLpBlockReadyGo BlockReady // registered go callback function

//export ps4000LpBlockReadyGo
func ps4000LpBlockReadyGo(handle int16, status int, noOfSamples uint32, overflow int16, param interface{}) {
	if regLpBlockReadyGo != nil {
		regLpBlockReadyGo(handle, status, param) // call registered go callback function
	}
	return
}

func ps4000RunBlock(handle int16, noOfPreTriggerSamples, noOfPostTriggerSamples int32,
	timeBase uint32, overSample int16, segmentIndex uint32, lpBlockReadyGoPar BlockReady,
	param interface{}) (timeIndisposedMs int32, err error) {
	regLpBlockReadyGo = lpBlockReadyGoPar
	slog.Debug("ps4000RunBlock", "handle", handle, "noOfPreTriggerSamples", noOfPreTriggerSamples, "noOfPostTriggerSamples", noOfPostTriggerSamples, "timeBase", timeBase, "overSample", overSample, "segmentIndex", segmentIndex, "lpBlockReadyGoPar", lpBlockReadyGoPar, "param", param)
	stat := C.ps4000RunBlock((C.short)(handle), (C.int)(noOfPreTriggerSamples),
		(C.int)(noOfPostTriggerSamples), (C.uint)(timeBase), (C.short)(overSample),
		(*C.int)(&timeIndisposedMs), (C.uint16_t)(segmentIndex), (C.ps4000BlockReady)(C.ps4000LpBlockReady),
		unsafe.Pointer(&param))
	if stat != C.PICO_OK {
		err = fmt.Errorf("RunBlock:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps4000SetTriggerChannelProperties(handle int16, channelProperties []TriggerChannelProperties, auxOutputEnable bool,
	autoTriggerMs int32) (err error) {
	var cTriggerChannelProperties []C.TRIGGER_CHANNEL_PROPERTIES
	if len(channelProperties) > 0 {
		cTriggerChannelProperties = make([]C.TRIGGER_CHANNEL_PROPERTIES, len(channelProperties))
		for i := range channelProperties {
			cTriggerChannelProperties[i].channel = (C.PS4000_CHANNEL)(channelProperties[i].Channel)
			cTriggerChannelProperties[i].thresholdLowerHysteresis = (C.ushort)(channelProperties[i].ThresholdLowerHysteresis)
			cTriggerChannelProperties[i].thresholdLower = (C.short)(channelProperties[i].ThresholdLower)
			cTriggerChannelProperties[i].thresholdUpperHysteresis = (C.ushort)(channelProperties[i].ThresholdUpperHysteresis)
			cTriggerChannelProperties[i].thresholdUpper = (C.short)(channelProperties[i].ThresholdUpper)
		}
	}
	pcTriggerChannelProperties := (*C.TRIGGER_CHANNEL_PROPERTIES)(nil)
	if len(channelProperties) > 0 {
		pcTriggerChannelProperties = &cTriggerChannelProperties[0]
	}
	slog.Debug("ps4000SetTriggerChannelProperties", "handle", handle, "channelProperties", channelProperties, "auxOutputEnable", auxOutputEnable, "autoTriggerMs", autoTriggerMs)
	stat := C.ps4000SetTriggerChannelProperties((C.short)(handle),
		(*C.TRIGGER_CHANNEL_PROPERTIES)(pcTriggerChannelProperties),
		(C.short)(len(channelProperties)), (C.short)(boolToint16(auxOutputEnable)), (C.int)(autoTriggerMs))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetTriggerChannelProperties:  %s", psc.StatStr(int(stat)))
	}
	return

}

func ps4000SetTriggerChannelConditions(handle int16, triggerConditions []TriggerConditions) (err error) {
	cTriggerConditions := make([]C.TRIGGER_CONDITIONS, len(triggerConditions))
	for i := range triggerConditions {
		cTriggerConditions[i].channelA = (C.TRIGGER_STATE)(triggerConditions[i].ChannelA)
		cTriggerConditions[i].channelB = (C.TRIGGER_STATE)(triggerConditions[i].ChannelB)
		cTriggerConditions[i].channelC = (C.TRIGGER_STATE)(triggerConditions[i].ChannelC)
		cTriggerConditions[i].channelD = (C.TRIGGER_STATE)(triggerConditions[i].ChannelD)
		cTriggerConditions[i].external = (C.TRIGGER_STATE)(triggerConditions[i].External)
		cTriggerConditions[i].aux = (C.TRIGGER_STATE)(triggerConditions[i].Aux)
		cTriggerConditions[i].pulseWidthQualifier = (C.TRIGGER_STATE)(triggerConditions[i].PulseWidthQualifier)
	}
	pcTriggerConditions := (*C.TRIGGER_CONDITIONS)(nil)
	if len(triggerConditions) > 0 {
		pcTriggerConditions = &cTriggerConditions[0]
	}
	slog.Debug("ps4000SetTriggerChannelConditions", "handle", handle, "triggerConditions", triggerConditions)
	stat := C.ps4000SetTriggerChannelConditions((C.short)(handle),
		(*C.TRIGGER_CONDITIONS)(pcTriggerConditions),
		(C.short)(len(triggerConditions)))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetTriggerChannelConditions:  %s", psc.StatStr(int(stat)))
	}

	return
}

func ps4000SetTriggerChannelDirections(handle int16, channelA, channelB, channelC, channelD, ext, aux ThresholdDirection) (err error) {
	slog.Debug("ps4000SetTriggerChannelDirections", "handle", handle, "channelA", channelA, "channelB", channelB, "channelC", channelC, "channelD", channelD, "ext", ext, "aux", aux)
	stat := C.ps4000SetTriggerChannelDirections((C.short)(handle),
		(C.THRESHOLD_DIRECTION)(channelA),
		(C.THRESHOLD_DIRECTION)(channelB),
		(C.THRESHOLD_DIRECTION)(channelC),
		(C.THRESHOLD_DIRECTION)(channelD),
		(C.THRESHOLD_DIRECTION)(ext),
		(C.THRESHOLD_DIRECTION)(aux))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetTriggerChannelDirections:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps4000SetTriggerDelay(handle int16, delay uint32) (err error) {
	slog.Debug("ps4000SetTriggerDelay", "handle", handle, "delay", delay)
	stat := C.ps4000SetTriggerDelay((C.short)(handle), (C.uint)(delay))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetTriggerDelay:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps4000SetPulseWidthQualifier(handle int16, conditions []PwqConditions, direction ThresholdDirection, lower, upper uint32,
	pwType PulseWidthType) (err error) {
	cPwqConditions := make([]C.PWQ_CONDITIONS, len(conditions))
	for i := range conditions {
		cPwqConditions[i].channelA = (C.TRIGGER_STATE)(conditions[i].ChannelA)
		cPwqConditions[i].channelB = (C.TRIGGER_STATE)(conditions[i].ChannelB)
		cPwqConditions[i].channelC = (C.TRIGGER_STATE)(conditions[i].ChannelC)
		cPwqConditions[i].channelD = (C.TRIGGER_STATE)(conditions[i].ChannelD)
		cPwqConditions[i].external = (C.TRIGGER_STATE)(conditions[i].External)
	}
	slog.Debug("ps4000SetPulseWidthQualifier", "handle", handle, "conditions", conditions, "direction", direction, "lower", lower, "upper", upper, "pwType", pwType)
	stat := C.ps4000SetPulseWidthQualifier((C.short)(handle),
		(*C.PWQ_CONDITIONS)(&cPwqConditions[0]), (C.short)(len(conditions)),
		(C.THRESHOLD_DIRECTION)(direction), (C.uint)(lower), (C.uint)(upper),
		(C.PULSE_WIDTH_TYPE)(pwType))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetPulseWidthQualifier:  %s", psc.StatStr(int(stat)))
	}
	return
}
func ps4000SetTriggerDigitalPortProperties(handle int16, digitalDirections []DigitalChannelDirections) (err error) {
	return fmt.Errorf("SetTriggerDigitalPortProperties not supported on ps4000")
}

func ps4000Stop(handle int16) (err error) {
	slog.Debug("ps4000Stop", "handle", handle)
	stat := C.ps4000Stop((C.short)(handle))
	if stat != C.PICO_OK {
		err = fmt.Errorf("Stop:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps4000SetSigGenBuiltIn(handle int16, offsetVoltage int32, pkToPK uint32, waveType WaveTypeEnum,
	startFrequency, stopFrequency, increment, dwellTime float32, sweepType SweepTypeEnum,
	operation ExtraOperations, shots, sweeps uint32, triggerType SigGenTrigType,
	triggerSource SigGenTrigSource, extInThreshold int16) (err error) {
	slog.Debug("ps4000SetSigGenBuiltIn", "handle", handle, "offsetVoltage", offsetVoltage, "pkToPK", pkToPK, "waveType", waveType, "startFrequency", startFrequency, "stopFrequency", stopFrequency, "increment", increment, "dwellTime", dwellTime, "sweepType", sweepType, "operation", operation, "shots", shots, "sweeps", sweeps, "triggerType", triggerType, "triggerSource", triggerSource, "extInThreshold", extInThreshold)
	stat := C.ps4000SetSigGenBuiltIn((C.short)(handle), (C.int)(offsetVoltage),
		(C.uint)(pkToPK), (C.short)(waveType), (C.float)(startFrequency),
		(C.float)(stopFrequency), (C.float)(increment), (C.float)(dwellTime),
		(C.SWEEP_TYPE)(sweepType), (C.int16_t)(operation),
		(C.uint)(shots), (C.uint)(sweeps), (C.SIGGEN_TRIG_TYPE)(triggerType),
		(C.SIGGEN_TRIG_SOURCE)(triggerSource), (C.short)(extInThreshold))

	if stat != C.PICO_OK {
		err = fmt.Errorf("SetSigGenBuiltIn:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps4000SetSigGenBuiltInV2(handle int16, offsetVoltage int32, pkToPK uint32, waveType WaveTypeEnum,
	startFrequency, stopFrequency, increment, dwellTime float64, sweepType SweepTypeEnum,
	operation ExtraOperations, shots, sweeps uint32, triggerType SigGenTrigType,
	triggerSource SigGenTrigSource, extInThreshold int16) (err error) {
	return fmt.Errorf("SetSigGenBuiltInV2 not supported on ps4000")
}

func ps4000SigGenFrequencyToPhase(handle int16, frequency float64, indexMode IndexMode, bufferLength uint32) (phase uint32, err error) {
	slog.Debug("ps4000SigGenFrequencyToPhase", "handle", handle, "frequency", frequency, "indexMode", indexMode, "bufferLength", bufferLength)
	stat := C.ps4000SigGenFrequencyToPhase((C.short)(handle), (C.double)(frequency),
		(C.INDEX_MODE)(indexMode), (C.uint)(bufferLength), (*C.uint)(&phase))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SigGenFrequencyToPhase:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps4000SetNoCaptures(handle int16, nCaptures uint32) (err error) {
	slog.Debug("ps4000SetNoOfCaptures", "handle", handle, "nCaptures", nCaptures)
	stat := C.ps4000SetNoOfCaptures((C.short)(handle), (C.uint16_t)(nCaptures))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetNoCaptures:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps4000GetTriggerTimeOffset(handle int16, segmentIndex uint32) (timeUpper, timeLower uint32, timeUnits TimeUnits, err error) {
	slog.Debug("ps4000GetTriggerTimeOffset", "handle", handle, "segmentIndex", segmentIndex)
	stat := C.ps4000GetTriggerTimeOffset((C.short)(handle), (*C.uint)(&timeUpper),
		(*C.uint)(&timeLower), (*C.PS4000_TIME_UNITS)(&timeUnits), (C.uint16_t)(segmentIndex))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetTriggerTimeOffset:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps4000GetTriggerTimeOffset64(handle int16, segmentIndex uint32) (time int64, timeUnits TimeUnits, err error) {
	slog.Debug("ps4000GetTriggerTimeOffset64", "handle", handle, "segmentIndex", segmentIndex)
	stat := C.ps4000GetTriggerTimeOffset64((C.short)(handle), (*C.long)(&time),
		(*C.PS4000_TIME_UNITS)(&timeUnits), (C.uint16_t)(segmentIndex))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetTriggerTimeOffset64:  %s", psc.StatStr(int(stat)))
	}
	return

}

func ps4000GetValuesTriggerTimeOffsetBulk(handle int16, timesUpper, timesLower []uint32, timeUnits []TimeUnits,
	fromSegmentIndex, toSegmentIndex uint32) (err error) {
	slog.Debug("ps4000GetValuesTriggerTimeOffsetBulk", "handle", handle, "timesUpper", timesUpper, "timesLower", timesLower, "timeUnits", timeUnits, "fromSegmentIndex", fromSegmentIndex, "toSegmentIndex", toSegmentIndex)
	stat := C.ps4000GetValuesTriggerTimeOffsetBulk((C.short)(handle), (*C.uint)(&timesUpper[0]),
		(*C.uint)(&timesLower[0]), (*C.PS4000_TIME_UNITS)(&timeUnits[0]), (C.uint16_t)(fromSegmentIndex),
		(C.uint16_t)(toSegmentIndex))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetValuesTriggerTimeOffsetBulk:  %s", psc.StatStr(int(stat)))
	}

	return
}

func ps4000GetValuesTriggerTimeOffsetBulk64(handle int16, times []int64, timeUnits []TimeUnits,
	fromSegmentIndex, toSegmentIndex uint32) (err error) {
	slog.Debug("ps4000GetValuesTriggerTimeOffsetBulk64", "handle", handle, "times", times, "timeUnits", timeUnits, "fromSegmentIndex", fromSegmentIndex, "toSegmentIndex", toSegmentIndex)
	stat := C.ps4000GetValuesTriggerTimeOffsetBulk64((C.short)(handle), (*C.long)(&times[0]),
		(*C.PS4000_TIME_UNITS)(&timeUnits[0]), (C.uint16_t)(fromSegmentIndex),
		(C.uint16_t)(toSegmentIndex))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetValuesTriggerTimeOffsetBulk64:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps4000HoldOff(handle int16, holdOff uint64, holdOffType HoldOffType) (err error) {
	slog.Debug("ps4000HoldOff", "handle", handle, "holdOff", holdOff, "holdOffType", holdOffType)
	stat := C.ps4000HoldOff((C.short)(handle), (C.ulong)(holdOff), (C.PS4000_HOLDOFF_TYPE)(holdOffType))
	if stat != C.PICO_OK {
		err = fmt.Errorf("HoldOff:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps4000LsReady(handle int16) (ready int16, err error) {
	slog.Debug("ps4000LsReady", "handle", handle)
	stat := C.ps4000IsReady((C.short)(handle), (*C.short)(&ready))
	if stat != C.PICO_OK {
		err = fmt.Errorf("LsReady:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps4000TriggerOrPulseWidthQualifierEnabled(handle int16) (triggerEnabled, pulseWidthQualifierEnabledint16 int16, err error) {
	slog.Debug("ps4000TriggerOrPulseWidthQualifierEnabled", "handle", handle)
	stat := C.ps4000IsTriggerOrPulseWidthQualifierEnabled((C.short)(handle),
		(*C.short)(&triggerEnabled), (*C.short)(&pulseWidthQualifierEnabledint16))
	if stat != C.PICO_OK {
		err = fmt.Errorf("TriggerOrPulseWidthQualifierEnabled:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps4000MemorySegments(handle int16, nSegments uint32) (nMaxSamples int32, err error) {
	slog.Debug("ps4000MemorySegments", "handle", handle, "nSegments", nSegments)
	stat := C.ps4000MemorySegments((C.short)(handle),
		(C.uint16_t)(nSegments), (*C.int)(&nMaxSamples))
	if stat != C.PICO_OK {
		err = fmt.Errorf("MemorySegments:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps4000NoOfStreamingValues(handle int16) (noOfValues uint32, err error) {
	slog.Debug("ps4000NoOfStreamingValues", "handle", handle)
	stat := C.ps4000NoOfStreamingValues((C.short)(handle),
		(*C.uint)(&noOfValues))
	if stat != C.PICO_OK {
		err = fmt.Errorf("NoOfStreamingValues:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps4000PingUnit(handle int16) (err error) {
	slog.Debug("ps4000PingUnit", "handle", handle)
	stat := C.ps4000PingUnit((C.short)(handle))
	if stat != C.PICO_OK {
		err = fmt.Errorf("PingUnit:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps4000QueryOutputEdgeDetect(handle int16) (state int16, err error) {
	return 0, fmt.Errorf("QueryOutputEdgeDetect not supported on ps4000")
}

func ps4000SetDigitalPort(handle int16, port DigitalPort, enabled bool, logiclevel int16) (err error) {
	return fmt.Errorf("SetDigitalPort not supported on ps4000")
}

func ps4000SetOutputEdgeDetect(handle int16, state int16) (err error) {
	return fmt.Errorf("SetOutputEdgeDetect not supported on ps4000")
}

// ps4000GetScalingValues
func ps4000SetPulseWidthDigitalPortProperties(handle int16, digitalDirections []DigitalChannelDirections) (err error) {
	return fmt.Errorf("SetPulseWidthDigitalPortProperties not supported on ps4000")
}

func ps4000SetSigGenArbitrary(handle int16, offsetVoltage int32, pkToPK uint32,
	startDeltaPhase, stopDeltaPhase, deltaPhaseIncrement, dwellCount uint32,
	arbitraryWaveform []int16, sweepType SweepTypeEnum, operation ExtraOperations,
	indexMode IndexMode, shots, sweeps uint32, triggerType SigGenTrigType,
	triggerSource SigGenTrigSource, extInThreshold int16) (err error) {
	slog.Debug("ps4000SetSigGenArbitrary", "handle", handle, "offsetVoltage", offsetVoltage, "pkToPK", pkToPK, "startDeltaPhase", startDeltaPhase, "stopDeltaPhase", stopDeltaPhase, "deltaPhaseIncrement", deltaPhaseIncrement, "dwellCount", dwellCount, "arbitraryWaveform", arbitraryWaveform, "sweepType", sweepType, "operation", operation, "indexMode", indexMode, "shots", shots, "sweeps", sweeps, "triggerType", triggerType, "triggerSource", triggerSource, "extInThreshold", extInThreshold)
	stat := C.ps4000SetSigGenArbitrary((C.short)(handle), (C.int)(offsetVoltage),
		(C.uint)(pkToPK), (C.uint)(startDeltaPhase), (C.uint)(stopDeltaPhase),
		(C.uint32_t)(deltaPhaseIncrement), (C.uint32_t)(dwellCount),
		(*C.short)(&arbitraryWaveform[0]), (C.int32_t)(len(arbitraryWaveform)),
		(C.SWEEP_TYPE)(sweepType), (C.int16_t)(operation),
		(C.INDEX_MODE)(indexMode),
		(C.uint)(shots), (C.uint)(sweeps), (C.SIGGEN_TRIG_TYPE)(triggerType),
		(C.SIGGEN_TRIG_SOURCE)(triggerSource), (C.short)(extInThreshold))

	if stat != C.PICO_OK {
		err = fmt.Errorf("SetSigGenArbitrary:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps4000SetSigGenPropertiesArbitrary(handle int16, offsetVoltage int32,
	startDeltaPhase, stopDeltaPhase, deltaPhaseIncrement, dwellCount uint32,
	sweepType SweepTypeEnum, operation ExtraOperations,
	indexMode IndexMode, shots, sweeps uint32, triggerType SigGenTrigType,
	triggerSource SigGenTrigSource, extInThreshold int16) (err error) {
	return fmt.Errorf("SetSigGenPropertiesArbitrary not supported on ps4000")
}

func ps4000SetSigGenPropertiesBuiltIn(handle int16, offsetVoltage int32,
	startFrequency, stopFrequency, increment, dwellTime float64,
	sweepType SweepTypeEnum,
	shots, sweeps uint32, triggerType SigGenTrigType,
	triggerSource SigGenTrigSource, extInThreshold int16) (err error) {
	return fmt.Errorf("SetSigGenPropertiesBuiltIn not supported on ps4000")
}

func ps4000SigGenArbitraryMinMaxValues(handle int16) (minArbitraryWaveformValue, maxArbitraryWaveformValue int16,
	minArbitraryWaveformSize, maxArbitraryWaveformSize uint32, err error) {
	slog.Debug("ps4000SigGenArbitraryMinMaxValues", "handle", handle)
	stat := C.ps4000SigGenArbitraryMinMaxValues((C.short)(handle),
		(*C.short)(&minArbitraryWaveformValue), (*C.short)(&maxArbitraryWaveformValue),
		(*C.uint32_t)(&minArbitraryWaveformSize), (*C.uint32_t)(&maxArbitraryWaveformSize))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SigGenArbitraryMinMaxValues:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps4000SigGenSoftwareControl(handle int16, state int16) (err error) {
	slog.Debug("ps4000SigGenSoftwareControl", "handle", handle, "state", state)
	stat := C.ps4000SigGenSoftwareControl((C.short)(handle),
		(C.short)(state))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetSigGenBuiltIn:  %s", psc.StatStr(int(stat)))
	}
	return
}
