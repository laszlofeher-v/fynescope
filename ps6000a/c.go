//go:build !demo && ps6000a

package ps6000a

// #include <stdlib.h>
// #include "/opt/picoscope/include/libps6000a/PicoStatus.h"
// #include "/opt/picoscope/include/libps6000a/ps6000aApi.h"
/*
// Forward declarations
int ps6000aLpBlockReady(int16_t handle, PICO_STATUS status, void * pParameter);
int ps6000aLpDataReady(int16_t handle, PICO_STATUS status, uint32_t noOfSamples,
				int16_t overflow, void * pParameter);
int lpStreamingReady(int16_t handle, int32_t noOfSamples, uint32_t startIndex,
                int16_t overflow, uint32_t triggerAt, int16_t triggered,
                int16_t autoStop, void * pParameter);

static inline PICO_STATUS wrap_ps6000aGetValuesAsync(
	int16_t handle,
	uint64_t startIndex,
	uint64_t noOfSamples,
	uint64_t downSampleRatio,
	PICO_RATIO_MODE downSampleRatioMode,
	uint64_t segmentIndex,
	void *lpDataReady,
	void *pParameter)
{
	return ps6000aGetValuesAsync(handle, startIndex, noOfSamples, downSampleRatio, downSampleRatioMode, segmentIndex, (ps6000aDataReady)lpDataReady, (PICO_POINTER)pParameter);
}

static inline PICO_STATUS wrap_ps6000aGetValuesBulkAsync(
	int16_t handle,
	uint64_t startIndex,
	uint64_t noOfSamples,
	uint64_t fromSegmentIndex,
	uint64_t toSegmentIndex,
	uint64_t downSampleRatio,
	PICO_RATIO_MODE downSampleRatioMode,
	void *lpDataReady,
	void *pParameter)
{
	return ps6000aGetValuesBulkAsync(handle, startIndex, noOfSamples, fromSegmentIndex, toSegmentIndex, downSampleRatio, downSampleRatioMode, (ps6000aDataReady)lpDataReady, (PICO_POINTER)pParameter);
}
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
	scopeHandler.Id = "ps6000a"
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
		slog.Debug("ps6000aEnumerateUnits", "bufferLen", bufferLen)
		stat := C.ps6000aEnumerateUnits((*C.short)(&count), cstrPtr, (*C.short)(&serialLth))
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
	slog.Debug("ps6000aOpenUnit", "serial", serial)
	stat := C.ps6000aOpenUnit((*C.short)(&handle),
		(*C.schar)(p), (C.PICO_DEVICE_RESOLUTION)(resolution))
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
	slog.Debug("ps6000aOpenUnitAsync", "serial", serial)
	stat := C.ps6000aOpenUnitAsync((*C.short)(&status), (*C.schar)(p),
		(C.PICO_DEVICE_RESOLUTION)(resolution))
	if stat != C.PICO_OK {
		err = fmt.Errorf("OpenUnitAsync:  %s", psc.StatStr(int(stat)))
		return
	}
	return
}

func openUnitProgress() (handle int16, progressPercent int16, complete int16, err error) {
	stat := C.ps6000aOpenUnitProgress((*C.short)(&handle),
		(*C.short)(&progressPercent), (*C.short)(&complete))
	if stat != C.PICO_OK {
		err = fmt.Errorf("OpenUnitProgress:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aCloseUnit(handle int16) (err error) {
	slog.Debug("ps6000aCloseUnit", "handle", handle)
	stat := C.ps6000aCloseUnit((C.short)(handle))
	if stat != C.PICO_OK {
		err = fmt.Errorf("CloseUnit:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aGetUnitInfo(handle int16, info PicoInfo) (infoString string, err error) {
	const listLen = 4096
	var cstrPtr *C.schar
	cstrPtr = (*C.schar)(C.malloc(C.sizeof_schar * listLen))
	defer C.free(unsafe.Pointer(cstrPtr))
	requiredSize := int16(listLen)
	slog.Debug("ps6000aGetUnitInfo", "handle", handle, "info", info)
	stat := C.ps6000aGetUnitInfo((C.short)(handle), cstrPtr, (C.short)(requiredSize),
		(*C.short)(&requiredSize), (C.PICO_INFO)(info))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetUnitInfo:  %s", psc.StatStr(int(stat)))
	}
	if requiredSize == 0 {
		infoString = "No answer from ps6000aGetUnitInfo "
		return
	}
	b := C.GoBytes(unsafe.Pointer(cstrPtr), (C.int)(requiredSize-1))
	infoString = string(b)
	return
}

func ps6000aFlashLed(handle int16, start int16) (err error) {
	slog.Debug("ps6000aFlashLed", "handle", handle, "start", start)
	stat := C.ps6000aFlashLed((C.short)(handle), (C.short)(start))
	if stat != C.PICO_OK {
		err = fmt.Errorf("FlashLed:  %s", psc.StatStr(int(stat)))
	}
	return
}

var regLpDataReadyGo DataReady // registered go callback function

//export ps6000aLpDataReadyGo
func ps6000aLpDataReadyGo(handle int16, status int, noOfSamples uint32, overflow int16, param interface{}) {
	if regLpDataReadyGo != nil {
		regLpDataReadyGo(handle, status, noOfSamples, overflow, param) // call registered go callback function
	}
	return
}

func ps6000aGetValuesAsync(handle int16, startIndex, noOfSamples, downSampleRatio uint64,
	downSampleRatioMode RatioMode, lpDataReadyGoPar DataReady, segmentIndex uint64,
	param interface{}) (err error) {
	regLpDataReadyGo = lpDataReadyGoPar
	slog.Debug("ps6000aGetValuesAsync", "handle", handle, "startIndex", startIndex, "noOfSamples", noOfSamples, "downSampleRatio", downSampleRatio, "downSampleRatioMode", downSampleRatioMode, "lpDataReadyGoPar", lpDataReadyGoPar, "segmentIndex", segmentIndex, "param", param)
	stat := C.wrap_ps6000aGetValuesAsync((C.short)(handle),
		(C.uint64_t)(startIndex),
		(C.uint64_t)(noOfSamples),
		(C.uint64_t)(downSampleRatio),
		(C.PICO_RATIO_MODE)(downSampleRatioMode),
		(C.uint64_t)(segmentIndex),
		unsafe.Pointer(C.ps6000aLpDataReady),
		unsafe.Pointer(&param))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetValuesAsync:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aGetValues(handle int16, startIndex, reqNoOfSamples, downSampleRatio uint64,
	downSampleRatioMode RatioMode, segmentIndex uint64) (noOfSamples uint64, overflow int16, err error) {
	slog.Debug("ps6000aGetValues", "handle", handle, "startIndex", startIndex, "reqNoOfSamples", reqNoOfSamples, "downSampleRatio", downSampleRatio, "downSampleRatioMode", downSampleRatioMode, "segmentIndex", segmentIndex)
	stat := C.ps6000aGetValues((C.int16_t)(handle),
		(C.uint64_t)(startIndex),
		(*C.uint64_t)(&reqNoOfSamples),
		(C.uint64_t)(downSampleRatio),
		(C.PICO_RATIO_MODE)(downSampleRatioMode),
		(C.uint64_t)(segmentIndex),
		(*C.int16_t)(&overflow))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetValues:  %s", psc.StatStr(int(stat)))
	}
	noOfSamples = reqNoOfSamples
	return
}

func ps6000aGetValuesBulk(handle int16, noOfSamples, fromSegmentIndex, toSegmentIndex, downSampleRatio uint64,
	downSampleRatioMode RatioMode, overflow []int16) (noSamples uint64, err error) {
	slog.Debug("ps6000aGetValuesBulk", "handle", handle, "noOfSamples", noOfSamples, "fromSegmentIndex", fromSegmentIndex, "toSegmentIndex", toSegmentIndex, "downSampleRatio", downSampleRatio, "downSampleRatioMode", downSampleRatioMode, "overflow", overflow)
	stat := C.ps6000aGetValuesBulk((C.int16_t)(handle),
		(C.uint64_t)(0), // startIndex
		(*C.uint64_t)(&noOfSamples),
		(C.uint64_t)(fromSegmentIndex),
		(C.uint64_t)(toSegmentIndex),
		(C.uint64_t)(downSampleRatio),
		(C.PICO_RATIO_MODE)(downSampleRatioMode),
		(*C.int16_t)(&overflow[0]))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetValuesBulk:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aGetValuesOverlapped(handle int16, startIndex, reqNoOfSamples, downSampleRatio uint32,
	downSampleRatioMode RatioMode, segmentIndex uint32, overflow []int16) (noSamples uint32, err error) {
	slog.Debug("ps6000aGetValuesOverlapped", "handle", handle, "startIndex", startIndex, "reqNoOfSamples", reqNoOfSamples, "downSampleRatio", downSampleRatio, "downSampleRatioMode", downSampleRatioMode, "segmentIndex", segmentIndex, "overflow", overflow)
	reqNoOfSamples64 := C.uint64_t(reqNoOfSamples)
	stat := C.ps6000aGetValuesOverlapped((C.int16_t)(handle),
		(C.uint64_t)(startIndex),
		(*C.uint64_t)(&reqNoOfSamples64),
		(C.uint64_t)(downSampleRatio),
		(C.PICO_RATIO_MODE)(downSampleRatioMode),
		(C.uint64_t)(segmentIndex), // fromSegmentIndex
		(C.uint64_t)(segmentIndex), // toSegmentIndex
		(*C.int16_t)(&overflow[0]))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetValuesOverlapped:  %s", psc.StatStr(int(stat)))
	}
	noSamples = uint32(reqNoOfSamples64)
	return
}

func ps6000aGetValuesOverlappedBulk(handle int16, startIndex, reqNoOfSamples,
	downSampleRatio uint32, downSampleRatioMode RatioMode, fromSegmentIndex,
	toSegmentIndex uint32, overflow []int16) (noSamples uint32, err error) {
	slog.Debug("ps6000aGetValuesOverlappedBulk", "handle", handle, "startIndex", startIndex, "reqNoOfSamples", reqNoOfSamples, "downSampleRatio", downSampleRatio, "downSampleRatioMode", downSampleRatioMode, "fromSegmentIndex", fromSegmentIndex, "toSegmentIndex", toSegmentIndex, "overflow", overflow)
	err = fmt.Errorf("Not supported on ps6000a")
	return
}

func ps6000aGetAnalogueOffset(handle int16, voltageRange int, coupling Coupling) (maximumVoltage, minimumVoltage float32, err error) {
	err = fmt.Errorf("Not supported on ps6000a")
	return
}

func ps6000aGetChannelInformation(handle int16, info int16, probe int32, ranges []int32, channels ChannelId) (lengthOfRanges int32, err error) {
	return 0, fmt.Errorf("GetChannelInformation not supported on ps6000a")
}

func ps6000aGetMaxDownSampleRatio(handle int16, noOfUnaggregatedSamples uint32, downSampleRatioMode RatioMode, segmentIndex int32) (maxDownSampleRatio uint32, err error) {
	return 0, fmt.Errorf("GetMaxDownSampleRatio not supported on ps6000a")
}

func ps6000aGetMaxSegments(handle int16) (maxSegments uint32, err error) {
	return 0, fmt.Errorf("GetMaxSegments not supported on ps6000a")
}

// ps6000aChangePowerSource
// ps6000aCurrentPowerSource
func ps6000aGetNumOfCaptures(handle int16) (nCaptures uint64, err error) {
	slog.Debug("ps6000aGetNoOfCaptures", "handle", handle)
	stat := C.ps6000aGetNoOfCaptures((C.short)(handle), (*C.uint64_t)(&nCaptures))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetNoOfCaptures:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aGetNumOfProcessedCaptures(handle int16) (nCaptures uint64, err error) {
	slog.Debug("ps6000aGetNoOfProcessedCaptures", "handle", handle)
	stat := C.ps6000aGetNoOfProcessedCaptures((C.short)(handle), (*C.uint64_t)(&nCaptures))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetNoOfProcessedCaptures:  %s", psc.StatStr(int(stat)))
	}
	return
}

var regLpStreamingReadyGo StreamingReady // registered go callback function

//export ps6000aLpStreamingReadyGo
func ps6000aLpStreamingReadyGo(handle int16, noOfSamples int32, startIndex uint32, overflow int16,
	triggeredAt uint32, triggered, autoStop int16, param interface{}) {
	if regLpStreamingReadyGo != nil {
		regLpStreamingReadyGo(handle, noOfSamples, startIndex, overflow, triggeredAt, autoStop, triggered, param) // call registered go callback function
	}
	return
}

func ps6000aGetStreamingLatestValues(handle int16, lpStreamingReadyGoPar StreamingReady, param interface{}) (err error) {
	return fmt.Errorf("GetStreamingLatestValues requires updating for ps6000a new standard")
}

// ps6000aCheckForUpdate
// ps6000aStartFirmwareUpdate
func ps6000aGetTimebase(handle int16, timeBase uint32, noOfSamples uint64,
	segmentIndex uint32) (timeIntervalNanoseconds float64, maxSamples uint64, err error) {
	slog.Debug("ps6000aGetTimebase", "handle", handle, "timeBase", timeBase,
		"noOfSamples", noOfSamples, "segmentIndex", segmentIndex)
	stat := C.ps6000aGetTimebase((C.short)(handle), (C.uint32_t)(timeBase), (C.uint64_t)(noOfSamples),
		(*C.double)(&timeIntervalNanoseconds),
		(*C.uint64_t)(&maxSamples), (C.uint64_t)(segmentIndex))
	if stat != C.PICO_OK {
		slog.Error("GetTimebase", "noOfSamples", noOfSamples, "stat", psc.StatStr(int(stat)))
		err = fmt.Errorf("GetTimebase:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aGetTimebase2(handle int16, timeBase uint64, numOfSamples uint64,
	overSample int16, segmentIndex uint64) (timeIntervalNanoseconds float64, maxSamples uint64, err error) {
	return 0, 0, fmt.Errorf("GetTimebase2 not supported on ps6000a")
}

func ps6000aSetChannel(handle int16, channel ChannelId, enabled bool, couplingType Coupling, voltageRange RangeEnum, analogOffset float32) (err error) {
	if enabled {
		stat := C.ps6000aSetChannelOn((C.short)(handle), (C.PICO_CHANNEL)(channel), (C.PICO_COUPLING)(couplingType), (C.PICO_CONNECT_PROBE_RANGE)(voltageRange), (C.double)(analogOffset), C.PICO_BW_FULL)
		if stat != C.PICO_OK {
			return fmt.Errorf("SetChannelOn: %s", psc.StatStr(int(stat)))
		}
	} else {
		stat := C.ps6000aSetChannelOff((C.short)(handle), (C.PICO_CHANNEL)(channel))
		if stat != C.PICO_OK {
			return fmt.Errorf("SetChannelOff: %s", psc.StatStr(int(stat)))
		}
	}
	return nil
}

func ps6000aMaximumValue(handle int16) (value int32, err error) {
	slog.Debug("ps6000aMaximumValue", "handle", handle)
	return 0, fmt.Errorf("ps6000aMaximumValue not supported on ps6000a")
}

func ps6000aMinimumValue(handle int16) (value int32, err error) {
	slog.Debug("ps6000aMinimumValue", "handle", handle)
	var minValue, maxValue int16
	stat := C.ps6000aGetAdcLimits((C.short)(handle), C.PICO_DR_8BIT, (*C.short)(&minValue), (*C.short)(&maxValue))
	if stat != C.PICO_OK {
		return -32767, fmt.Errorf("%s", psc.StatStr(int(stat)))
	}
	return int32(minValue), nil
}

func ps6000aSetSimpleTrigger(handle int16, enable bool, source ChannelId, threshold int16,
	direction ThresholdDirection, delay uint32, autoTriggerMs int16) (err error) {
	slog.Debug("ps6000aSetSimpleTrigger", "handle", handle, "enable", enable, "source", source, "threshold", threshold, "direction", direction, "delay", delay, "autoTriggerMs", autoTriggerMs)
	stat := C.ps6000aSetSimpleTrigger((C.short)(handle), (C.short)(boolToint16(enable)),
		(C.PICO_CHANNEL)(source), (C.short)(threshold), (C.PICO_THRESHOLD_DIRECTION)(direction),
		(C.uint64_t)(delay), (C.uint32_t)(autoTriggerMs*1000))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetSimpleTrigger:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aSetDataBuffer(handle int16, ch ChannelId, bufferIn []int16, segmentIndex uint64,
	mode RatioMode) (err error) {
	slog.Debug("ps6000aSetDataBuffer", "handle", handle, "ch", ch, "segmentIndex", segmentIndex, "mode", mode)
	var pBuf unsafe.Pointer
	if len(bufferIn) > 0 {
		pBuf = unsafe.Pointer(&bufferIn[0])
	}
	stat := C.ps6000aSetDataBuffer((C.short)(handle), (C.PICO_CHANNEL)(ch), (C.PICO_POINTER)(pBuf),
		(C.int32_t)(len(bufferIn)), C.PICO_INT16_T, (C.uint64_t)(segmentIndex),
		(C.PICO_RATIO_MODE)(mode), C.PICO_ADD)
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetDataBuffer:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aSetDataBuffers(handle int16, ch ChannelId, bufferMax, bufferMin []int16, segmentIndex uint32, mode RatioMode) (err error) {
	slog.Debug("ps6000aSetDataBuffers", "handle", handle, "ch", ch, "segmentIndex", segmentIndex, "mode", mode)
	var pMax, pMin unsafe.Pointer
	if len(bufferMax) > 0 {
		pMax = unsafe.Pointer(&bufferMax[0])
	}
	if len(bufferMin) > 0 {
		pMin = unsafe.Pointer(&bufferMin[0])
	}
	stat := C.ps6000aSetDataBuffers((C.short)(handle), (C.PICO_CHANNEL)(ch), (C.PICO_POINTER)(pMax), (C.PICO_POINTER)(pMin),
		(C.int32_t)(len(bufferMax)), C.PICO_INT16_T, (C.uint64_t)(segmentIndex), (C.PICO_RATIO_MODE)(mode), C.PICO_ADD)
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetDataBuffers:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aSetUnscaledDataBuffers(handle int16, ch ChannelId, bufferMax, bufferMin []int16, segmentIndex uint32, mode RatioMode) (err error) {
	return fmt.Errorf("Not supported on ps6000a")
}

func ps6000aSetEtsTimeBuffer(handle int16, buffer []int64) (err error) {
	return fmt.Errorf("Not supported on ps6000a")
}

func ps6000aSetEtsTimeBuffers(handle int16, timeUpper, timeLower []uint32) (err error) {
	return fmt.Errorf("Not supported on ps6000a")
}

func ps6000aSetEts(handle int16, mode EtsMode, etsCycles int16, etsInterLeave int16) (sampleTimePicoseconds int32, err error) {
	return 0, fmt.Errorf("Not supported on ps6000a")
}

func ps6000aRunStreaming(handle int16, reqSampleInterval uint32, sampleIntervalTimeUnits TimeUnits,
	maxPreTriggerSamples, maxPostTriggerSamples uint32,
	autoStop bool, downSampleRatio uint32, downSampleRatioMode RatioMode,
	overviewBufferSize uint32) (sampleInterval uint32, err error) {
	slog.Debug("ps6000aRunStreaming", "handle", handle, "reqSampleInterval", reqSampleInterval)
	var intervalSec float64 = float64(reqSampleInterval)
	stat := C.ps6000aRunStreaming((C.short)(handle), (*C.double)(&intervalSec),
		(C.PICO_TIME_UNITS)(sampleIntervalTimeUnits), (C.uint64_t)(maxPreTriggerSamples),
		(C.uint64_t)(maxPostTriggerSamples), (C.short)(boolToint16(autoStop)), (C.uint64_t)(downSampleRatio),
		(C.PICO_RATIO_MODE)(downSampleRatioMode))
	if stat != C.PICO_OK {
		err = fmt.Errorf("RunStreaming:  %s", psc.StatStr(int(stat)))
	}
	sampleInterval = uint32(intervalSec)
	return
}

var regLpBlockReadyGo BlockReady

//export ps6000aLpBlockReadyGo
func ps6000aLpBlockReadyGo(handle int16, status int, param unsafe.Pointer) {
	if regLpBlockReadyGo != nil {
		regLpBlockReadyGo(handle, status, param)
	}
}

func ps6000aRunBlock(handle int16, noOfPreTriggerSamples, noOfPostTriggerSamples uint64,
	timeBase uint64, overSample int16, segmentIndex uint64, lpBlockReadyGoPar BlockReady,
	param interface{}) (timeIndisposedMs int32, err error) {
	regLpBlockReadyGo = lpBlockReadyGoPar
	slog.Debug("ps6000aRunBlock", "handle", handle, "noOfPreTriggerSamples", noOfPreTriggerSamples, "noOfPostTriggerSamples", noOfPostTriggerSamples, "timeBase", timeBase, "segmentIndex", segmentIndex)
	var timeMs float64
	stat := C.ps6000aRunBlock((C.short)(handle), (C.uint64_t)(noOfPreTriggerSamples), (C.uint64_t)(noOfPostTriggerSamples),
		(C.uint32_t)(timeBase), (*C.double)(&timeMs), (C.uint64_t)(segmentIndex), (C.ps6000aBlockReady)(C.ps6000aLpBlockReady), nil)
	if stat != C.PICO_OK {
		err = fmt.Errorf("RunBlock: %s", psc.StatStr(int(stat)))
	}
	timeIndisposedMs = int32(timeMs)
	return
}

func ps6000aSetTriggerChannelProperties(handle int16, channelProperties []TriggerChannelProperties, auxOutputEnable bool, autoTriggerMicroSeconds uint32) (err error) {
	slog.Debug("ps6000aSetTriggerChannelProperties", "handle", handle)
	cProperties := make([]C.PICO_TRIGGER_CHANNEL_PROPERTIES, len(channelProperties))
	for i := range channelProperties {
		cProperties[i].thresholdUpper = (C.int16_t)(channelProperties[i].ThresholdUpper)
		cProperties[i].thresholdUpperHysteresis = (C.uint16_t)(channelProperties[i].ThresholdUpperHysteresis)
		cProperties[i].thresholdLower = (C.int16_t)(channelProperties[i].ThresholdLower)
		cProperties[i].thresholdLowerHysteresis = (C.uint16_t)(channelProperties[i].ThresholdLowerHysteresis)
		cProperties[i].channel = (C.PICO_CHANNEL)(channelProperties[i].Channel)
	}
	var pProp *C.PICO_TRIGGER_CHANNEL_PROPERTIES
	if len(cProperties) > 0 {
		pProp = &cProperties[0]
	}
	stat := C.ps6000aSetTriggerChannelProperties((C.short)(handle), pProp, (C.short)(len(cProperties)), (C.short)(boolToint16(auxOutputEnable)), (C.uint32_t)(autoTriggerMicroSeconds))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetTriggerChannelProperties: %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aSetTriggerChannelConditions(handle int16, triggerConditions []TriggerConditions) (err error) {
	cConditions := make([]C.PICO_CONDITION, 0, len(triggerConditions)*4)
	for _, tc := range triggerConditions {
		if tc.ChannelA != CondDontCare {
			cConditions = append(cConditions, C.PICO_CONDITION{source: C.PICO_CHANNEL_A, condition: C.PICO_TRIGGER_STATE(tc.ChannelA)})
		}
		if tc.ChannelB != CondDontCare {
			cConditions = append(cConditions, C.PICO_CONDITION{source: C.PICO_CHANNEL_B, condition: C.PICO_TRIGGER_STATE(tc.ChannelB)})
		}
		if tc.ChannelC != CondDontCare {
			cConditions = append(cConditions, C.PICO_CONDITION{source: C.PICO_CHANNEL_C, condition: C.PICO_TRIGGER_STATE(tc.ChannelC)})
		}
		if tc.ChannelD != CondDontCare {
			cConditions = append(cConditions, C.PICO_CONDITION{source: C.PICO_CHANNEL_D, condition: C.PICO_TRIGGER_STATE(tc.ChannelD)})
		}
	}
	var pcConditions *C.PICO_CONDITION
	if len(cConditions) > 0 {
		pcConditions = &cConditions[0]
	}
	slog.Debug("ps6000aSetTriggerChannelConditions", "handle", handle)
	stat := C.ps6000aSetTriggerChannelConditions((C.short)(handle), pcConditions, (C.short)(len(cConditions)), C.PICO_CLEAR_ALL|C.PICO_ADD)
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetTriggerChannelConditions: %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aSetTriggerChannelDirections(handle int16, channelA, channelB, channelC, channelD, ext, aux ThresholdDirection) (err error) {
	cDirections := []C.PICO_DIRECTION{
		{channel: C.PICO_CHANNEL_A, direction: C.PICO_THRESHOLD_DIRECTION(channelA)},
		{channel: C.PICO_CHANNEL_B, direction: C.PICO_THRESHOLD_DIRECTION(channelB)},
		{channel: C.PICO_CHANNEL_C, direction: C.PICO_THRESHOLD_DIRECTION(channelC)},
		{channel: C.PICO_CHANNEL_D, direction: C.PICO_THRESHOLD_DIRECTION(channelD)},
		{channel: C.PICO_EXTERNAL, direction: C.PICO_THRESHOLD_DIRECTION(ext)},
	}
	slog.Debug("ps6000aSetTriggerChannelDirections", "handle", handle)
	stat := C.ps6000aSetTriggerChannelDirections((C.short)(handle), &cDirections[0], (C.short)(len(cDirections)))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetTriggerChannelDirections: %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aSetTriggerDelay(handle int16, delay uint64) (err error) {
	slog.Debug("ps6000aSetTriggerDelay", "handle", handle, "delay", delay)
	stat := C.ps6000aSetTriggerDelay((C.short)(handle), (C.uint64_t)(delay))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetTriggerDelay: %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aSetPulseWidthQualifier(handle int16, conditions []PwqConditions, direction ThresholdDirection, lower, upper uint32, pwType PulseWidthType) (err error) {
	return fmt.Errorf("Not supported on ps6000a")
}

func ps6000aSetTriggerDigitalPortProperties(handle int16, port DigitalPort, digitalDirections []DigitalChannelDirections) (err error) {
	slog.Debug("ps6000aSetTriggerDigitalPortProperties", "handle", handle, "port", port)
	cDirections := make([]C.PICO_DIGITAL_CHANNEL_DIRECTIONS, len(digitalDirections))
	for i := range digitalDirections {
		cDirections[i].channel = (C.PICO_PORT_DIGITAL_CHANNEL)(digitalDirections[i].Channel)
		cDirections[i].direction = (C.PICO_DIGITAL_DIRECTION)(digitalDirections[i].Direction)
	}
	var pDir *C.PICO_DIGITAL_CHANNEL_DIRECTIONS
	if len(cDirections) > 0 {
		pDir = &cDirections[0]
	}
	stat := C.ps6000aSetTriggerDigitalPortProperties((C.short)(handle), (C.PICO_CHANNEL)(port), pDir, (C.short)(len(cDirections)))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetTriggerDigitalPortProperties: %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aStop(handle int16) (err error) {
	slog.Debug("ps6000aStop", "handle", handle)
	stat := C.ps6000aStop((C.short)(handle))
	if stat != C.PICO_OK {
		err = fmt.Errorf("Stop:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aSetSigGenBuiltIn(handle int16, offsetVoltage int32, pkToPK uint32, waveType WaveTypeEnum,
	startFrequency, stopFrequency, increment, dwellTime float32, sweepType SweepTypeEnum,
	operation ExtraOperations, shots, sweeps uint32, triggerType SigGenTrigType,
	triggerSource SigGenTrigSource, extInThreshold int16) (err error) {
	return fmt.Errorf("Not supported on ps6000a")
}

func ps6000aSetSigGenBuiltInV2(handle int16, offsetVoltage int32, pkToPK uint32, waveType WaveTypeEnum,
	startFrequency, stopFrequency, increment, dwellTime float64, sweepType SweepTypeEnum,
	operation ExtraOperations, shots, sweeps uint32, triggerType SigGenTrigType,
	triggerSource SigGenTrigSource, extInThreshold int16) (err error) {
	return fmt.Errorf("Not supported on ps6000a")
}

func ps6000aSigGenFrequencyToPhase(handle int16, frequency float64, indexMode IndexMode, bufferLength uint32) (phase uint32, err error) {
	return 0, fmt.Errorf("Not supported on ps6000a")
}

func ps6000aSetNoCaptures(handle int16, nCaptures uint32) (err error) {
	slog.Debug("ps6000aSetNoOfCaptures", "handle", handle, "nCaptures", nCaptures)
	stat := C.ps6000aSetNoOfCaptures((C.short)(handle), (C.uint64_t)(nCaptures))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetNoCaptures:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aGetTriggerTimeOffset(handle int16, segmentIndex uint64) (timeUpper, timeLower uint32, timeUnits TimeUnits, err error) {
	var time int64
	time, timeUnits, err = ps6000aGetTriggerTimeOffset64(handle, segmentIndex)
	timeUpper = uint32(time >> 32)
	timeLower = uint32(time)
	return
}

func ps6000aGetTriggerTimeOffset64(handle int16, segmentIndex uint64) (time int64, timeUnits TimeUnits, err error) {
	slog.Debug("ps6000aGetTriggerTimeOffset64", "handle", handle, "segmentIndex", segmentIndex)
	stat := C.ps6000aGetTriggerTimeOffset((C.short)(handle), (*C.int64_t)(unsafe.Pointer(&time)),
		(*C.PICO_TIME_UNITS)(&timeUnits), (C.uint64_t)(segmentIndex))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetTriggerTimeOffset64:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aGetValuesTriggerTimeOffsetBulk(handle int16, times []int64, timeUnits []TimeUnits, fromSegmentIndex, toSegmentIndex uint32) (err error) {
	slog.Debug("ps6000aGetValuesTriggerTimeOffsetBulk", "handle", handle, "fromSegmentIndex", fromSegmentIndex, "toSegmentIndex", toSegmentIndex)
	var pTimes *C.int64_t
	if len(times) > 0 {
		pTimes = (*C.int64_t)(unsafe.Pointer(&times[0]))
	}
	var pUnits *C.PICO_TIME_UNITS
	if len(timeUnits) > 0 {
		pUnits = (*C.PICO_TIME_UNITS)(&timeUnits[0])
	}
	stat := C.ps6000aGetValuesTriggerTimeOffsetBulk((C.short)(handle), pTimes, pUnits, (C.uint64_t)(fromSegmentIndex), (C.uint64_t)(toSegmentIndex))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetValuesTriggerTimeOffsetBulk: %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aGetValuesTriggerTimeOffsetBulk64(handle int16, times []int64, timeUnits []TimeUnits, fromSegmentIndex, toSegmentIndex uint32) (err error) {
	return fmt.Errorf("Not supported on ps6000a")
}

func ps6000aHoldOff(handle int16, holdOff uint64, holdOffType HoldOffType) (err error) {
	return fmt.Errorf("Not supported on ps6000a")
}

func ps6000aLsReady(handle int16) (ready int16, err error) {
	slog.Debug("ps6000aLsReady", "handle", handle)
	stat := C.ps6000aIsReady((C.short)(handle), (*C.short)(&ready))
	if stat != C.PICO_OK {
		err = fmt.Errorf("LsReady:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aTriggerOrPulseWidthQualifierEnabled(handle int16) (triggerEnabled, pulseWidthQualifierEnabledint16 int16, err error) {
	return 0, 0, fmt.Errorf("TriggerOrPulseWidthQualifierEnabled not supported on ps6000a")
}

func ps6000aMemorySegments(handle int16, nSegments uint64) (nMaxSamples uint64, err error) {
	slog.Debug("ps6000aMemorySegments", "handle", handle, "nSegments", nSegments)
	stat := C.ps6000aMemorySegments((C.short)(handle), (C.uint64_t)(nSegments), (*C.uint64_t)(&nMaxSamples))
	if stat != C.PICO_OK {
		err = fmt.Errorf("MemorySegments:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aNoOfStreamingValues(handle int16) (noOfValues uint64, err error) {
	slog.Debug("ps6000aNoOfStreamingValues", "handle", handle)
	stat := C.ps6000aNoOfStreamingValues((C.short)(handle), (*C.uint64_t)(&noOfValues))
	if stat != C.PICO_OK {
		err = fmt.Errorf("NoOfStreamingValues:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aPingUnit(handle int16) (err error) {
	slog.Debug("ps6000aPingUnit", "handle", handle)
	stat := C.ps6000aPingUnit((C.short)(handle))
	if stat != C.PICO_OK {
		err = fmt.Errorf("PingUnit:  %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aQueryOutputEdgeDetect(handle int16) (state int16, err error) {
	slog.Debug("ps6000aQueryOutputEdgeDetect", "handle", handle)
	stat := C.ps6000aQueryOutputEdgeDetect((C.short)(handle), (*C.short)(&state))
	if stat != C.PICO_OK {
		err = fmt.Errorf("QueryOutputEdgeDetect: %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aSetDigitalPort(handle int16, port DigitalPort, enabled bool, logiclevel int16) (err error) {
	return fmt.Errorf("Not supported on ps6000a")
}

func ps6000aSetOutputEdgeDetect(handle int16, state int16) (err error) {
	slog.Debug("ps6000aSetOutputEdgeDetect", "handle", handle, "state", state)
	stat := C.ps6000aSetOutputEdgeDetect((C.short)(handle), (C.short)(state))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetOutputEdgeDetect: %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aSetPulseWidthDigitalPortProperties(handle int16, port DigitalPort, digitalDirections []DigitalChannelDirections) (err error) {
	slog.Debug("ps6000aSetPulseWidthDigitalPortProperties", "handle", handle, "port", port)
	cDirections := make([]C.PICO_DIGITAL_CHANNEL_DIRECTIONS, len(digitalDirections))
	for i := range digitalDirections {
		cDirections[i].channel = (C.PICO_PORT_DIGITAL_CHANNEL)(digitalDirections[i].Channel)
		cDirections[i].direction = (C.PICO_DIGITAL_DIRECTION)(digitalDirections[i].Direction)
	}
	var pDir *C.PICO_DIGITAL_CHANNEL_DIRECTIONS
	if len(cDirections) > 0 {
		pDir = &cDirections[0]
	}
	stat := C.ps6000aSetPulseWidthDigitalPortProperties((C.short)(handle), (C.PICO_CHANNEL)(port), pDir, (C.short)(len(cDirections)))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetPulseWidthDigitalPortProperties: %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aSetSigGenArbitrary(handle int16, offsetVoltage int32, pkToPK uint32,
	startDeltaPhase, stopDeltaPhase, deltaPhaseIncrement, dwellCount uint32,
	arbitraryWaveform []int16, sweepType SweepTypeEnum, operation ExtraOperations,
	indexMode IndexMode, shots, sweeps uint32, triggerType SigGenTrigType,
	triggerSource SigGenTrigSource, extInThreshold int16) (err error) {
	return fmt.Errorf("Not supported on ps6000a")
}

func ps6000aSetSigGenPropertiesArbitrary(handle int16, offsetVoltage int32,
	startDeltaPhase, stopDeltaPhase, deltaPhaseIncrement, dwellCount uint32,
	sweepType SweepTypeEnum, operation ExtraOperations,
	indexMode IndexMode, shots, sweeps uint32, triggerType SigGenTrigType,
	triggerSource SigGenTrigSource, extInThreshold int16) (err error) {
	return fmt.Errorf("Not supported on ps6000a")
}

func ps6000aSetSigGenPropertiesBuiltIn(handle int16, offsetVoltage int32,
	startFrequency, stopFrequency, increment, dwellTime float64,
	sweepType SweepTypeEnum,
	shots, sweeps uint32, triggerType SigGenTrigType,
	triggerSource SigGenTrigSource, extInThreshold int16) (err error) {
	return fmt.Errorf("Not supported on ps6000a")
}

func ps6000aSigGenArbitraryMinMaxValues(handle int16) (minArbitraryWaveformValue, maxArbitraryWaveformValue int16,
	minArbitraryWaveformSize, maxArbitraryWaveformSize uint32, err error) {
	return 0, 0, 0, 0, fmt.Errorf("Not supported on ps6000a")
}

func ps6000aSigGenSoftwareControl(handle int16, state int16) (err error) {
	return fmt.Errorf("Not supported on ps6000a")
}

// Additional ps6000aApi.h bindings

func ps6000aGetAccessoryInfo(handle int16, channel ChannelId, listLen int16, info PicoInfo) (infoString string, requiredSize int16, err error) {
	var cstrPtr *C.schar
	cstrPtr = (*C.schar)(C.malloc(C.sizeof_schar * C.ulong(listLen)))
	defer C.free(unsafe.Pointer(cstrPtr))
	slog.Debug("ps6000aGetAccessoryInfo", "handle", handle, "channel", channel, "info", info)
	stat := C.ps6000aGetAccessoryInfo((C.short)(handle), (C.PICO_CHANNEL)(channel), cstrPtr, (C.short)(listLen), (*C.short)(&requiredSize), (C.PICO_INFO)(info))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetAccessoryInfo: %s", psc.StatStr(int(stat)))
		return
	}
	b := C.GoBytes(unsafe.Pointer(cstrPtr), (C.int)(requiredSize-1))
	infoString = string(b)
	return
}

func ps6000aSetAuxIoMode(handle int16, auxIoMode int) (err error) {
	slog.Debug("ps6000aSetAuxIoMode", "handle", handle, "auxIoMode", auxIoMode)
	stat := C.ps6000aSetAuxIoMode((C.short)(handle), (C.PICO_AUXIO_MODE)(auxIoMode))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetAuxIoMode: %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aSetTriggerHoldoffCounterBySamples(handle int16, samples uint64) (err error) {
	slog.Debug("ps6000aSetTriggerHoldoffCounterBySamples", "handle", handle, "samples", samples)
	stat := C.ps6000aSetTriggerHoldoffCounterBySamples((C.short)(handle), (C.uint64_t)(samples))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetTriggerHoldoffCounterBySamples: %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aMemorySegmentsBySamples(handle int16, nSamples uint64) (nMaxSegments uint64, err error) {
	slog.Debug("ps6000aMemorySegmentsBySamples", "handle", handle, "nSamples", nSamples)
	stat := C.ps6000aMemorySegmentsBySamples((C.short)(handle), (C.uint64_t)(nSamples), (*C.uint64_t)(&nMaxSegments))
	if stat != C.PICO_OK {
		err = fmt.Errorf("MemorySegmentsBySamples: %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aGetMaximumAvailableMemory(handle int16, resolution int) (nMaxSamples uint64, err error) {
	slog.Debug("ps6000aGetMaximumAvailableMemory", "handle", handle, "resolution", resolution)
	stat := C.ps6000aGetMaximumAvailableMemory((C.short)(handle), (*C.uint64_t)(&nMaxSamples), (C.PICO_DEVICE_RESOLUTION)(resolution))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetMaximumAvailableMemory: %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aQueryMaxSegmentsBySamples(handle int16, nSamples uint64, nChannelEnabled uint32, resolution int) (nMaxSegments uint64, err error) {
	slog.Debug("ps6000aQueryMaxSegmentsBySamples", "handle", handle, "nSamples", nSamples, "nChannelEnabled", nChannelEnabled, "resolution", resolution)
	stat := C.ps6000aQueryMaxSegmentsBySamples((C.short)(handle), (C.uint64_t)(nSamples), (C.uint32_t)(nChannelEnabled), (*C.uint64_t)(&nMaxSegments), (C.PICO_DEVICE_RESOLUTION)(resolution))
	if stat != C.PICO_OK {
		err = fmt.Errorf("QueryMaxSegmentsBySamples: %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aGetScopeState(handle int16) (scopeState int, err error) {
	slog.Debug("ps6000aGetScopeState", "handle", handle)
	var cState C.PICO_SCOPE_STATE
	stat := C.ps6000aGetScopeState((C.short)(handle), &cState)
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetScopeState: %s", psc.StatStr(int(stat)))
	} else {
		scopeState = int(cState)
	}
	return
}

func ps6000aSetDigitalPortOn(handle int16, port DigitalPort, logicThresholdLevel []int16, hysteresis int) (err error) {
	slog.Debug("ps6000aSetDigitalPortOn", "handle", handle, "port", port, "hysteresis", hysteresis)
	var pLevel *C.short
	if len(logicThresholdLevel) > 0 {
		pLevel = (*C.short)(&logicThresholdLevel[0])
	}
	stat := C.ps6000aSetDigitalPortOn((C.short)(handle), (C.PICO_CHANNEL)(port), pLevel, (C.short)(len(logicThresholdLevel)), (C.PICO_DIGITAL_PORT_HYSTERESIS)(hysteresis))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetDigitalPortOn: %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aSetDigitalPortOff(handle int16, port DigitalPort) (err error) {
	slog.Debug("ps6000aSetDigitalPortOff", "handle", handle, "port", port)
	stat := C.ps6000aSetDigitalPortOff((C.short)(handle), (C.PICO_CHANNEL)(port))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetDigitalPortOff: %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aSigGenWaveform(handle int16, waveType WaveTypeEnum, buffer []int16) (err error) {
	slog.Debug("ps6000aSigGenWaveform", "handle", handle, "waveType", waveType)
	var pBuf *C.short
	if len(buffer) > 0 {
		pBuf = (*C.short)(&buffer[0])
	}
	stat := C.ps6000aSigGenWaveform((C.short)(handle), (C.PICO_WAVE_TYPE)(waveType), pBuf, (C.uint64_t)(len(buffer)))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SigGenWaveform: %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aSigGenRange(handle int16, peakToPeakVolts, offsetVolts float64) (err error) {
	slog.Debug("ps6000aSigGenRange", "handle", handle, "peakToPeakVolts", peakToPeakVolts, "offsetVolts", offsetVolts)
	stat := C.ps6000aSigGenRange((C.short)(handle), (C.double)(peakToPeakVolts), (C.double)(offsetVolts))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SigGenRange: %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aSigGenWaveformDutyCycle(handle int16, dutyCyclePercent float64) (err error) {
	slog.Debug("ps6000aSigGenWaveformDutyCycle", "handle", handle, "dutyCyclePercent", dutyCyclePercent)
	stat := C.ps6000aSigGenWaveformDutyCycle((C.short)(handle), (C.double)(dutyCyclePercent))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SigGenWaveformDutyCycle: %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aSigGenTrigger(handle int16, triggerType SigGenTrigType, triggerSource SigGenTrigSource, cycles uint64, autoTriggerPicoSeconds uint64) (err error) {
	slog.Debug("ps6000aSigGenTrigger", "handle", handle, "cycles", cycles)
	stat := C.ps6000aSigGenTrigger((C.short)(handle), (C.PICO_SIGGEN_TRIG_TYPE)(triggerType), (C.PICO_SIGGEN_TRIG_SOURCE)(triggerSource), (C.uint64_t)(cycles), (C.uint64_t)(autoTriggerPicoSeconds))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SigGenTrigger: %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aSigGenFilter(handle int16, filterState int) (err error) {
	slog.Debug("ps6000aSigGenFilter", "handle", handle, "filterState", filterState)
	stat := C.ps6000aSigGenFilter((C.short)(handle), (C.PICO_SIGGEN_FILTER_STATE)(filterState))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SigGenFilter: %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aSigGenFrequency(handle int16, frequencyHz float64) (err error) {
	slog.Debug("ps6000aSigGenFrequency", "handle", handle, "frequencyHz", frequencyHz)
	stat := C.ps6000aSigGenFrequency((C.short)(handle), (C.double)(frequencyHz))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SigGenFrequency: %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aSigGenFrequencySweep(handle int16, stopFrequencyHz, frequencyIncrement, dwellTimeSeconds float64, sweepType SweepTypeEnum) (err error) {
	slog.Debug("ps6000aSigGenFrequencySweep", "handle", handle)
	stat := C.ps6000aSigGenFrequencySweep((C.short)(handle), (C.double)(stopFrequencyHz), (C.double)(frequencyIncrement), (C.double)(dwellTimeSeconds), (C.PICO_SWEEP_TYPE)(sweepType))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SigGenFrequencySweep: %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aSigGenPhase(handle int16, deltaPhase uint64) (err error) {
	slog.Debug("ps6000aSigGenPhase", "handle", handle, "deltaPhase", deltaPhase)
	stat := C.ps6000aSigGenPhase((C.short)(handle), (C.uint64_t)(deltaPhase))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SigGenPhase: %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aSigGenPhaseSweep(handle int16, stopDeltaPhase, deltaPhaseIncrement, dwellCount uint64, sweepType SweepTypeEnum) (err error) {
	slog.Debug("ps6000aSigGenPhaseSweep", "handle", handle)
	stat := C.ps6000aSigGenPhaseSweep((C.short)(handle), (C.uint64_t)(stopDeltaPhase), (C.uint64_t)(deltaPhaseIncrement), (C.uint64_t)(dwellCount), (C.PICO_SWEEP_TYPE)(sweepType))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SigGenPhaseSweep: %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aSigGenClockManual(handle int16, dacClockFrequency float64, prescaleRatio uint64) (err error) {
	slog.Debug("ps6000aSigGenClockManual", "handle", handle)
	stat := C.ps6000aSigGenClockManual((C.short)(handle), (C.double)(dacClockFrequency), (C.uint64_t)(prescaleRatio))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SigGenClockManual: %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aSigGenSoftwareTriggerControl(handle int16, triggerState SigGenTrigType) (err error) {
	slog.Debug("ps6000aSigGenSoftwareTriggerControl", "handle", handle)
	stat := C.ps6000aSigGenSoftwareTriggerControl((C.short)(handle), (C.PICO_SIGGEN_TRIG_TYPE)(triggerState))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SigGenSoftwareTriggerControl: %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aSigGenApply(handle int16, sigGenEnabled, sweepEnabled, triggerEnabled, automaticClockOptimisationEnabled, overrideAutomaticClockAndPrescale int16) (frequency, stopFrequency, frequencyIncrement, dwellTime float64, err error) {
	slog.Debug("ps6000aSigGenApply", "handle", handle)
	stat := C.ps6000aSigGenApply((C.short)(handle), (C.short)(sigGenEnabled), (C.short)(sweepEnabled), (C.short)(triggerEnabled), (C.short)(automaticClockOptimisationEnabled), (C.short)(overrideAutomaticClockAndPrescale), (*C.double)(&frequency), (*C.double)(&stopFrequency), (*C.double)(&frequencyIncrement), (*C.double)(&dwellTime))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SigGenApply: %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aSigGenPause(handle int16) (err error) {
	slog.Debug("ps6000aSigGenPause", "handle", handle)
	stat := C.ps6000aSigGenPause((C.short)(handle))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SigGenPause: %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aSigGenRestart(handle int16) (err error) {
	slog.Debug("ps6000aSigGenRestart", "handle", handle)
	stat := C.ps6000aSigGenRestart((C.short)(handle))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SigGenRestart: %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aTriggerWithinPreTriggerSamples(handle int16, state int) (err error) {
	slog.Debug("ps6000aTriggerWithinPreTriggerSamples", "handle", handle, "state", state)
	stat := C.ps6000aTriggerWithinPreTriggerSamples((C.short)(handle), (C.PICO_TRIGGER_WITHIN_PRE_TRIGGER)(state))
	if stat != C.PICO_OK {
		err = fmt.Errorf("TriggerWithinPreTriggerSamples: %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aSetPulseWidthQualifierProperties(handle int16, lower, upper uint32, pwType PulseWidthType) (err error) {
	slog.Debug("ps6000aSetPulseWidthQualifierProperties", "handle", handle, "lower", lower, "upper", upper, "pwType", pwType)
	stat := C.ps6000aSetPulseWidthQualifierProperties((C.short)(handle), (C.uint32_t)(lower), (C.uint32_t)(upper), (C.PICO_PULSE_WIDTH_TYPE)(pwType))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetPulseWidthQualifierProperties: %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aSetPulseWidthQualifierConditions(handle int16, triggerConditions []TriggerConditions) (err error) {
	cConditions := make([]C.PICO_CONDITION, 0, len(triggerConditions)*4)
	for _, tc := range triggerConditions {
		if tc.ChannelA != CondDontCare {
			cConditions = append(cConditions, C.PICO_CONDITION{source: C.PICO_CHANNEL_A, condition: C.PICO_TRIGGER_STATE(tc.ChannelA)})
		}
		if tc.ChannelB != CondDontCare {
			cConditions = append(cConditions, C.PICO_CONDITION{source: C.PICO_CHANNEL_B, condition: C.PICO_TRIGGER_STATE(tc.ChannelB)})
		}
	}
	var pcConditions *C.PICO_CONDITION
	if len(cConditions) > 0 {
		pcConditions = &cConditions[0]
	}
	slog.Debug("ps6000aSetPulseWidthQualifierConditions", "handle", handle)
	stat := C.ps6000aSetPulseWidthQualifierConditions((C.short)(handle), pcConditions, (C.short)(len(cConditions)), C.PICO_CLEAR_ALL|C.PICO_ADD)
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetPulseWidthQualifierConditions: %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aSetPulseWidthQualifierDirections(handle int16, channelA, channelB ThresholdDirection) (err error) {
	cDirections := []C.PICO_DIRECTION{
		{channel: C.PICO_CHANNEL_A, direction: C.PICO_THRESHOLD_DIRECTION(channelA)},
		{channel: C.PICO_CHANNEL_B, direction: C.PICO_THRESHOLD_DIRECTION(channelB)},
	}
	slog.Debug("ps6000aSetPulseWidthQualifierDirections", "handle", handle)
	stat := C.ps6000aSetPulseWidthQualifierDirections((C.short)(handle), &cDirections[0], (C.short)(len(cDirections)))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetPulseWidthQualifierDirections: %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aGetValuesBulkAsync(handle int16, startIndex, noOfSamples, fromSegmentIndex, toSegmentIndex, downSampleRatio uint64, downSampleRatioMode RatioMode, param interface{}) (err error) {
	slog.Debug("ps6000aGetValuesBulkAsync", "handle", handle, "startIndex", startIndex, "noOfSamples", noOfSamples)
	stat := C.wrap_ps6000aGetValuesBulkAsync((C.short)(handle), (C.uint64_t)(startIndex), (C.uint64_t)(noOfSamples), (C.uint64_t)(fromSegmentIndex), (C.uint64_t)(toSegmentIndex), (C.uint64_t)(downSampleRatio),
		(C.PICO_RATIO_MODE)(downSampleRatioMode), unsafe.Pointer(C.ps6000aLpDataReady), nil)
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetValuesBulkAsync: %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aStopUsingGetValuesOverlapped(handle int16) (err error) {
	slog.Debug("ps6000aStopUsingGetValuesOverlapped", "handle", handle)
	stat := C.ps6000aStopUsingGetValuesOverlapped((C.short)(handle))
	if stat != C.PICO_OK {
		err = fmt.Errorf("StopUsingGetValuesOverlapped: %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aGetAnalogueOffsetLimits(handle int16, rangeVal RangeEnum, coupling Coupling) (maximumVoltage, minimumVoltage float64, err error) {
	slog.Debug("ps6000aGetAnalogueOffsetLimits", "handle", handle, "rangeVal", rangeVal, "coupling", coupling)
	stat := C.ps6000aGetAnalogueOffsetLimits((C.short)(handle), (C.PICO_CONNECT_PROBE_RANGE)(rangeVal), (C.PICO_COUPLING)(coupling), (*C.double)(&maximumVoltage), (*C.double)(&minimumVoltage))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetAnalogueOffsetLimits: %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aGetMinimumTimebaseStateless(handle int16, enabledChannelFlags int, resolution int) (timebase uint32, timeInterval float64, err error) {
	slog.Debug("ps6000aGetMinimumTimebaseStateless", "handle", handle)
	stat := C.ps6000aGetMinimumTimebaseStateless((C.short)(handle), (C.PICO_CHANNEL_FLAGS)(enabledChannelFlags), (*C.uint32_t)(&timebase), (*C.double)(&timeInterval), (C.PICO_DEVICE_RESOLUTION)(resolution))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetMinimumTimebaseStateless: %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aSetDeviceResolution(handle int16, resolution int) (err error) {
	slog.Debug("ps6000aSetDeviceResolution", "handle", handle, "resolution", resolution)
	stat := C.ps6000aSetDeviceResolution((C.short)(handle), (C.PICO_DEVICE_RESOLUTION)(resolution))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetDeviceResolution: %s", psc.StatStr(int(stat)))
	}
	return
}

func ps6000aGetDeviceResolution(handle int16) (resolution int, err error) {
	slog.Debug("ps6000aGetDeviceResolution", "handle", handle)
	var cRes C.PICO_DEVICE_RESOLUTION
	stat := C.ps6000aGetDeviceResolution((C.short)(handle), &cRes)
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetDeviceResolution: %s", psc.StatStr(int(stat)))
	} else {
		resolution = int(cRes)
	}
	return
}

func ps6000aGetAdcLimits(handle int16, resolution int) (minValue, maxValue int16, err error) {
	slog.Debug("ps6000aGetAdcLimits", "handle", handle, "resolution", resolution)
	stat := C.ps6000aGetAdcLimits((C.short)(handle), (C.PICO_DEVICE_RESOLUTION)(resolution), (*C.short)(&minValue), (*C.short)(&maxValue))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetAdcLimits: %s", psc.StatStr(int(stat)))
	}
	return
}
