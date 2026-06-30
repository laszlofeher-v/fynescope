//go:build !noscope && ps3000

package ps3000a

// #cgo CFLAGS: -g -Wall -I/opt/picoscope/include/libps6000a
// #cgo LDFLAGS: -L/opt/picoscope/lib/ -lps3000a
// #include <stdlib.h>
// #include "/opt/picoscope/include/libps3000/ps3000.h"
// #include "/opt/picoscope/include/libps3000a/PicoStatus.h"
// #include "/opt/picoscope/include/libps3000a/ps3000aApi.h"
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
	"log/slog"
	"fynescope/psc"

	// "strings"
	"unsafe"
)

func boolToint16(b bool) int16 {
	if b {
		return int16(1)
	}
	return int16(0)
}
func (ps *PsDesc) Handle() int16 {
	return ps.handle
}

func EnumerateUnits() (count int16, serials string, serialLth int16, err error) {
	const listLen = 1024
	var cstrPtr *C.schar
	cstrPtr = (*C.schar)(C.malloc(C.sizeof_schar * listLen))
	defer C.free(unsafe.Pointer(cstrPtr))
	serialLth = listLen
	stat := C.ps3000aEnumerateUnits((*C.short)(&count), cstrPtr, (*C.short)(&serialLth))
	if stat != C.PICO_OK {
		err = fmt.Errorf("EnumerateUnits:  %s", psc.StatStr(int(stat)))
		return
	}
	b := C.GoBytes(unsafe.Pointer(cstrPtr), (C.int)(serialLth-1))
	serials = string(b)
	return
}

func (ps *PsDesc) OpenUnit(serial string) (err error) {
	var p *C.schar
	sLength := len(serial)
	if sLength > 0 {
		p = (*C.schar)(C.CString(serial))
		defer C.free(unsafe.Pointer(p))
	}
	stat := C.ps3000aOpenUnit((*C.short)(&ps.handle), (*C.schar)(p))
	if stat != C.PICO_OK {
		err = fmt.Errorf("OpenUnit:  %s", psc.StatStr(int(stat)))
		return
	}
	return
}
func (ps *PsDesc) OpenUnitAsync(serial string) (err error) {
	var p *C.schar
	sLength := len(serial)
	if sLength > 0 {
		p = (*C.schar)(C.CString(serial))
		defer C.free(unsafe.Pointer(p))
	}
	stat := C.ps3000aOpenUnitAsync((*C.short)(&ps.handle), (*C.schar)(p))
	if stat != C.PICO_OK {
		err = fmt.Errorf("OpenUnit:  %s", psc.StatStr(int(stat)))
		return
	}
	return
}

func (ps *PsDesc) CloseUnit() (err error) {
	stat := C.ps3000aCloseUnit((C.short)(ps.handle))
	if stat != C.PICO_OK {
		err = fmt.Errorf("CloseUnit:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) GetUnitInfo(info PicoInfo) (infoString string, err error) {
	const listLen = 4096
	var cstrPtr *C.schar
	cstrPtr = (*C.schar)(C.malloc(C.sizeof_schar * listLen))
	defer C.free(unsafe.Pointer(cstrPtr))
	requiredSize := int16(listLen)
	stat := C.ps3000aGetUnitInfo((C.short)(ps.handle), cstrPtr, (C.short)(requiredSize),
		(*C.short)(&requiredSize), (C.PICO_INFO)(info))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetUnitInfo:  %s", psc.StatStr(int(stat)))
	}
	if requiredSize == 0 {
		infoString = "No answer from PS3000AGetUnitInfo "
		return
	}
	b := C.GoBytes(unsafe.Pointer(cstrPtr), (C.int)(requiredSize-1))
	infoString = string(b)
	return
}

func (ps *PsDesc) FlashLed(start int16) (err error) {
	slog.Debug("PsFlashLed start")
	stat := C.ps3000aFlashLed((C.short)(ps.handle), (C.short)(start))
	slog.Debug("PsFlashLed", "stat", stat, "err", err)
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

func (ps *PsDesc) GetValuesAsync(startIndex, noOfSamples, downSampleRatio uint32,
	downSampleRatioMode RatioMode, lpDataReadyGoPar DataReady, segmentIndex uint32,
	param interface{}) (err error) {
	regLpDataReadyGo = lpDataReadyGoPar
	stat := C.ps3000aGetValuesAsync((C.short)(ps.handle),
		(C.uint)(startIndex),
		(C.uint)(noOfSamples),
		(C.uint)(downSampleRatio),
		(C.PS3000A_RATIO_MODE)(downSampleRatioMode),
		(C.uint)(segmentIndex),
		(C.lpDataReady), // C callback function in callbacks.go
		unsafe.Pointer(&param))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetValuesAsync:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) GetValues(startIndex, reqNoOfSamples, downSampleRatio uint32,
	downSampleRatioMode RatioMode, segmentIndex uint32) (noOfSamples uint32, overflow int16, err error) {
	//	fmt.Println("cgo GetValues 1 reqNoOfSamples:", reqNoOfSamples)
	stat := C.ps3000aGetValues((C.short)(ps.handle),
		(C.uint)(startIndex),
		(*C.uint)(&reqNoOfSamples),
		(C.uint)(downSampleRatio),
		(C.PS3000A_RATIO_MODE)(downSampleRatioMode),
		(C.uint)(segmentIndex),
		(*C.short)(&overflow))
	//	fmt.Println("cgo GetValues 2")
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetValues:  %s", psc.StatStr(int(stat)))
	}
	//	fmt.Println("cgo GetValues 3 reqNoOfSamples:", reqNoOfSamples)
	noOfSamples = reqNoOfSamples
	return
}

func (ps *PsDesc) GetValuesBulk(reqNoOfSamples uint32, fromSegmentIndex, toSegmentIndex, downSampleRatio uint32,
	downSampleRatioMode RatioMode, overflow []int16) (noSamples uint32, err error) {
	stat := C.ps3000aGetValuesBulk((C.short)(ps.handle),
		(*C.uint)(&reqNoOfSamples),
		(C.uint)(fromSegmentIndex),
		(C.uint)(toSegmentIndex),
		(C.uint)(downSampleRatio),
		(C.PS3000A_RATIO_MODE)(downSampleRatioMode),
		(*C.short)(&overflow[0]))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetValuesBulk:  %s", psc.StatStr(int(stat)))
	}
	noSamples = reqNoOfSamples
	return
}

func (ps *PsDesc) GetValuesOverlapped(startIndex, reqNoOfSamples, downSampleRatio uint32,
	downSampleRatioMode RatioMode, segmentIndex uint32, overflow []int16) (noSamples uint32, err error) {
	stat := C.ps3000aGetValuesOverlapped((C.short)(ps.handle),
		(C.uint)(startIndex),
		(*C.uint)(&reqNoOfSamples),
		(C.uint)(downSampleRatio),
		(C.PS3000A_RATIO_MODE)(downSampleRatioMode),
		(C.uint)(segmentIndex),
		(*C.short)(&overflow[0]))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetValuesOverlapped:  %s", psc.StatStr(int(stat)))
	}
	noSamples = reqNoOfSamples
	return
}

func (ps *PsDesc) GetValuesOverlappedBulk(startIndex, reqNoOfSamples, downSampleRatio uint32,
	downSampleRatioMode RatioMode, fromSegmentIndex, toSegmentIndex uint32, overflow []int16) (noSamples uint32, err error) {
	stat := C.ps3000aGetValuesOverlappedBulk((C.short)(ps.handle),
		(C.uint)(startIndex),
		(*C.uint)(&reqNoOfSamples),
		(C.uint)(downSampleRatio),
		(C.PS3000A_RATIO_MODE)(downSampleRatioMode),
		(C.uint)(fromSegmentIndex),
		(C.uint)(toSegmentIndex),
		(*C.short)(&overflow[0]))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetValuesOverlappedBulk:  %s", psc.StatStr(int(stat)))
	}
	noSamples = reqNoOfSamples
	return
}

func (ps *PsDesc) GetAnalogueOffset(voltageRange int, coupling Coupling) (maximumVoltage, minimumVoltage float32, err error) {
	stat := C.ps3000aGetAnalogueOffset((C.short)(ps.handle),
		(C.PS3000A_RANGE)(voltageRange),
		(C.PS3000A_COUPLING)(coupling), (*C.float)(&maximumVoltage), (*C.float)(&minimumVoltage))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetAnalogueOffset:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) GetChannelInformation(info int16, probe int32, ranges []int32, channels ChannelId) (lengthOfRanges int32, err error) {
	lengthOfRanges = int32(len(ranges))
	stat := C.ps3000aGetChannelInformation((C.short)(ps.handle), (C.PS3000A_CHANNEL_INFO)(info),
		(C.int)(probe), (*C.int)(&ranges[0]), (*C.int)(&lengthOfRanges), (C.int)(channels))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetChannelInformation:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) GetMaxDownSampleRatio(noOfUnaggregatedSamples uint32, downSampleRatioMode RatioMode, segmentIndex int32) (maxDownSampleRatio uint32, err error) {
	stat := C.ps3000aGetMaxDownSampleRatio((C.short)(ps.handle), (C.uint)(noOfUnaggregatedSamples),
		(*C.uint)(&maxDownSampleRatio), (C.PS3000A_RATIO_MODE)(downSampleRatioMode), (C.uint)(segmentIndex))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetMaxDownSampleRatio:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) GetMaxSegments() (maxSegments uint32, err error) {
	stat := C.ps3000aGetMaxSegments((C.short)(ps.handle), (*C.uint)(&maxSegments))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetMaxSegments:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) GetNoOfCaptures() (nCaptures uint32, err error) {
	stat := C.ps3000aGetNoOfCaptures((C.short)(ps.handle), (*C.uint)(&nCaptures))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetNoOfCaptures:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) GetNoOfProcessedCaptures() (nCaptures uint32, err error) {
	stat := C.ps3000aGetNoOfProcessedCaptures((C.short)(ps.handle), (*C.uint)(&nCaptures))
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

func (ps *PsDesc) GetStreamingLatestValues(lpStreamingReadyGoPar StreamingReady, param interface{}) (err error) {
	regLpStreamingReadyGo = lpStreamingReadyGoPar
	stat := C.ps3000aGetStreamingLatestValues((C.short)(ps.handle),
		(C.ps3000aStreamingReady)(C.lpStreamingReady), // C callback function in callbacks.go
		unsafe.Pointer(&param))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetStreamingLatestValues:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) GetTimebase(timeBase uint32, noOfSamples int32, overSample int16, segmentIndex uint32) (timeIntervalNanoseconds, maxSamples int32, err error) {
	stat := C.ps3000aGetTimebase((C.short)(ps.handle), (C.uint)(timeBase), (C.int)(noOfSamples),
		(*C.int)(&timeIntervalNanoseconds), (C.short)(overSample),
		(*C.int)(&maxSamples), (C.uint)(segmentIndex))
	if stat != C.PICO_OK {
		slog.Error("GetTimebase", "noOfSamples", noOfSamples, "stat", psc.StatStr(int(stat)))
		err = fmt.Errorf("GetTimebase:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) GetTimebase2(timeBase uint32, noOfSamples int32, overSample int16, segmentIndex uint32) (timeIntervalNanoseconds float32, maxSamples int32, err error) {
	stat := C.ps3000aGetTimebase2((C.short)(ps.handle), (C.uint)(timeBase), (C.int)(noOfSamples),
		(*C.float)(&timeIntervalNanoseconds), (C.short)(overSample),
		(*C.int)(&maxSamples), (C.uint)(segmentIndex))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetTimebase2:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) SetChannel(channel ChannelId, enabled bool, couplingType Coupling, voltageRange RangeEnum, analogOffset float32) (err error) {
	// log.Println("channel=", channel, " enabled=", enabled, " dcCoupled=", couplingType,
	// 	" range=", voltageRange, " offset=", analogOffset)
	stat := C.ps3000aSetChannel((C.short)(ps.handle), (C.PS3000A_CHANNEL)(channel), (C.short)(boolToint16(enabled)),
		(C.PS3000A_COUPLING)(couplingType), (C.PS3000A_RANGE)(voltageRange),
		(C.float)(analogOffset))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetChannel:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) MaximumValue() (value int16, err error) {
	stat := C.ps3000aMaximumValue((C.short)(ps.handle), (*C.short)(&value))
	if stat != C.PICO_OK {
		err = fmt.Errorf("MaximumValue:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) MinimumValue() (value int16, err error) {
	stat := C.ps3000aMinimumValue((C.short)(ps.handle), (*C.short)(&value))
	if stat != C.PICO_OK {
		err = fmt.Errorf("MinimumValue:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) SetSimpleTrigger(enable bool, source ChannelId, threshold int16,
	direction ThresholdDirection, delay uint32, autoTriggerMs int16) (err error) {
	stat := C.ps3000aSetSimpleTrigger((C.short)(ps.handle), (C.short)(boolToint16(enable)),
		(C.PS3000A_CHANNEL)(source), (C.short)(threshold),
		(C.PS3000A_THRESHOLD_DIRECTION)(direction), (C.uint)(delay),
		(C.short)(autoTriggerMs))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetSimpleTrigger:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) SetDataBuffer(ch ChannelId, bufferIn []int16, segmentIndex uint32,
	mode RatioMode) (err error) {

	stat := C.ps3000aSetDataBuffer((C.short)(ps.handle), (C.PS3000A_CHANNEL)(ch), (*C.short)(&bufferIn[0]),
		(C.int)(len(bufferIn)), (C.uint)(segmentIndex),
		(C.PS3000A_RATIO_MODE)(mode))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetDataBuffer:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) SetDataBuffers(ch ChannelId, bufferMax, bufferMin []int16, segmentIndex uint32, mode RatioMode) (err error) {
	// log.Println("handle=", ps.handle, " channel=", ch, " &bufferMax[0]", &bufferMin[0], " &bufferMax[0]", &bufferMin[0],
	// 	" len=", len(bufferMax), "segmentIndex", segmentIndex, " RatioMode=", mode)
	stat := C.ps3000aSetDataBuffers((C.short)(ps.handle), (C.PS3000A_CHANNEL)(ch), (*C.short)(&bufferMax[0]),
		(*C.short)(&bufferMin[0]), (C.int)(len(bufferMax)), (C.uint)(segmentIndex),
		(C.PS3000A_RATIO_MODE)(mode))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetDataBuffers:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) SetUnscaledDataBuffers(ch ChannelId, bufferMax, bufferMin []int16, segmentIndex uint32, mode RatioMode) (err error) {
	// log.Println("handle=", ps.handle, " channel=", ch, " &bufferMax[0]", &bufferMin[0], " &bufferMax[0]", &bufferMin[0],
	// 	" len=", len(bufferMax), "segmentIndex", segmentIndex, " RatioMode=", mode)
	stat := C.ps3000aSetDataBuffers((C.short)(ps.handle), (C.PS3000A_CHANNEL)(ch), (*C.short)(&bufferMax[0]),
		(*C.short)(&bufferMin[0]), (C.int)(len(bufferMax)), (C.uint)(segmentIndex),
		(C.PS3000A_RATIO_MODE)(mode))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetUnscaledDataBuffers:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) SetEtsTimeBuffer(buffer []int64) (err error) {
	stat := C.ps3000aSetEtsTimeBuffer((C.short)(ps.handle), (*C.long)(&buffer[0]),
		(C.int)(len(buffer)))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetEtsTimeBuffer:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) SetEtsTimeBuffers(timeUpper, timeLower []uint32) (err error) {
	stat := C.ps3000aSetEtsTimeBuffers((C.short)(ps.handle), (*C.uint)(&timeUpper[0]),
		(*C.uint)(&timeLower[0]), (C.int)(len(timeUpper)))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetEtsTimeBuffer:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) SetEts(mode EtsMode, etsCycles int16, etsInterLeave int16) (sampleTimePicoseconds int32, err error) {
	stat := C.ps3000aSetEts((C.short)(ps.handle), (C.PS3000A_ETS_MODE)(mode),
		(C.short)(etsCycles), (C.short)(etsInterLeave), (*C.int)(&sampleTimePicoseconds))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetEts:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) RunStreaming(reqSampleInterval uint32, sampleIntervalTimeUnits TimeUnits,
	maxPreTriggerSamples, maxPostTriggerSamples uint32,
	autoStop bool, downSampleRatio uint32, downSampleRatioMode RatioMode,
	overviewBufferSize uint32) (sampleInterval uint32, err error) {
	stat := C.ps3000aRunStreaming((C.short)(ps.handle), (*C.uint)(&reqSampleInterval),
		(C.PS3000A_TIME_UNITS)(sampleIntervalTimeUnits), (C.uint)(maxPreTriggerSamples),
		(C.uint)(maxPostTriggerSamples), (C.short)(boolToint16(autoStop)), (C.uint)(downSampleRatio),
		(C.PS3000A_RATIO_MODE)(downSampleRatioMode), (C.uint)(overviewBufferSize))
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

func (ps *PsDesc) RunBlock(noOfPreTriggerSamples, noOfPostTriggerSamples int32,
	timeBase uint32, overSample int16, segmentIndex uint32, lpBlockReadyGoPar BlockReady,
	param interface{}) (timeIndisposedMs int32, err error) {
	regLpBlockReadyGo = lpBlockReadyGoPar
	//	fmt.Println("noOfPreTriggerSamples=", noOfPreTriggerSamples, " noOfPostTriggerSamples=", noOfPostTriggerSamples,
	//		" timeBase=", timeBase, " overSample=", overSample, " segmentIndex=", segmentIndex)
	stat := C.ps3000aRunBlock((C.short)(ps.handle), (C.int)(noOfPreTriggerSamples),
		(C.int)(noOfPostTriggerSamples), (C.uint)(timeBase), (C.short)(overSample),
		(*C.int)(&timeIndisposedMs), (C.uint)(segmentIndex), (C.ps3000aBlockReady)(C.lpBlockReady),
		unsafe.Pointer(&param))
	if stat != C.PICO_OK {
		err = fmt.Errorf("RunBlock:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) SetTriggerChannelProperties(channelProperties []TriggerChannelProperties, auxOutputEnable bool,
	autoTriggerMs int32) (err error) {
	var cTriggerChannelProperties []C.PS3000A_TRIGGER_CHANNEL_PROPERTIES
	// log.Println("SetTriggerChannelProperties autoTriggerMs:", autoTriggerMs)
	if len(channelProperties) > 0 {
		cTriggerChannelProperties = make([]C.PS3000A_TRIGGER_CHANNEL_PROPERTIES, len(channelProperties))
		for i := range channelProperties {
			cTriggerChannelProperties[i].channel = (C.PS3000A_CHANNEL)(channelProperties[i].Channel)
			cTriggerChannelProperties[i].thresholdLowerHysteresis = (C.ushort)(channelProperties[i].ThresholdLowerHysteresis)
			cTriggerChannelProperties[i].thresholdLower = (C.short)(channelProperties[i].ThresholdLower)
			cTriggerChannelProperties[i].thresholdUpperHysteresis = (C.ushort)(channelProperties[i].ThresholdUpperHysteresis)
			cTriggerChannelProperties[i].thresholdUpper = (C.short)(channelProperties[i].ThresholdUpper)
		}
	}
	pcTriggerChannelProperties := (*C.PS3000A_TRIGGER_CHANNEL_PROPERTIES)(nil)
	if len(channelProperties) > 0 {
		pcTriggerChannelProperties = &cTriggerChannelProperties[0]
	}
	stat := C.ps3000aSetTriggerChannelProperties((C.short)(ps.handle),
		(*C.PS3000A_TRIGGER_CHANNEL_PROPERTIES)(pcTriggerChannelProperties),
		(C.short)(len(channelProperties)), (C.short)(boolToint16(auxOutputEnable)), (C.int)(autoTriggerMs))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetTriggerChannelProperties:  %s", psc.StatStr(int(stat)))
	}
	return

}

func (ps *PsDesc) SetTriggerChannelConditions(triggerConditions []TriggerConditions) (err error) {
	cTriggerConditions := make([]C.PS3000A_TRIGGER_CONDITIONS, len(triggerConditions))
	for i := range triggerConditions {
		cTriggerConditions[i].channelA = (C.PS3000A_TRIGGER_STATE)(triggerConditions[i].ChannelA)
		cTriggerConditions[i].channelB = (C.PS3000A_TRIGGER_STATE)(triggerConditions[i].ChannelB)
		cTriggerConditions[i].channelC = (C.PS3000A_TRIGGER_STATE)(triggerConditions[i].ChannelC)
		cTriggerConditions[i].channelD = (C.PS3000A_TRIGGER_STATE)(triggerConditions[i].ChannelD)
		cTriggerConditions[i].external = (C.PS3000A_TRIGGER_STATE)(triggerConditions[i].External)
		cTriggerConditions[i].aux = (C.PS3000A_TRIGGER_STATE)(triggerConditions[i].Aux)
		cTriggerConditions[i].pulseWidthQualifier = (C.PS3000A_TRIGGER_STATE)(triggerConditions[i].PulseWidthQualifier)
		// cTriggerConditions[i].digital = (C.PS3000A_TRIGGER_STATE)(triggerConditions[i].Digital)
	}
	pcTriggerConditions := (*C.PS3000A_TRIGGER_CONDITIONS)(nil)
	if len(triggerConditions) > 0 {
		pcTriggerConditions = &cTriggerConditions[0]
	}
	stat := C.ps3000aSetTriggerChannelConditions((C.short)(ps.handle),
		(*C.PS3000A_TRIGGER_CONDITIONS)(pcTriggerConditions),
		(C.short)(len(triggerConditions)))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetTriggerChannelConditions:  %s", psc.StatStr(int(stat)))
	}

	return
}

func (ps *PsDesc) SetTriggerChannelDirections(channelA, channelB, channelC, channelD, ext, aux ThresholdDirection) (err error) {
	stat := C.ps3000aSetTriggerChannelDirections((C.short)(ps.handle),
		(C.PS3000A_THRESHOLD_DIRECTION)(channelA),
		(C.PS3000A_THRESHOLD_DIRECTION)(channelB),
		(C.PS3000A_THRESHOLD_DIRECTION)(channelC),
		(C.PS3000A_THRESHOLD_DIRECTION)(channelD),
		(C.PS3000A_THRESHOLD_DIRECTION)(ext),
		(C.PS3000A_THRESHOLD_DIRECTION)(aux))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetTriggerChannelDirections:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) SetTriggerDelay(delay uint32) (err error) {
	stat := C.ps3000aSetTriggerDelay((C.short)(ps.handle), (C.uint)(delay))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetTriggerDelay:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) SetPulseWidthQualifier(conditions []PwqConditions, direction ThresholdDirection, lower, upper uint32,
	pwType PulseWidthType) (err error) {
	cPwqConditions := make([]C.PS3000A_PWQ_CONDITIONS, len(conditions))
	for i := range conditions {
		cPwqConditions[i].channelA = (C.PS3000A_TRIGGER_STATE)(conditions[i].ChannelA)
		cPwqConditions[i].channelB = (C.PS3000A_TRIGGER_STATE)(conditions[i].ChannelB)
		cPwqConditions[i].channelC = (C.PS3000A_TRIGGER_STATE)(conditions[i].ChannelC)
		cPwqConditions[i].channelD = (C.PS3000A_TRIGGER_STATE)(conditions[i].ChannelD)
		cPwqConditions[i].external = (C.PS3000A_TRIGGER_STATE)(conditions[i].External)
		cPwqConditions[i].aux = (C.PS3000A_TRIGGER_STATE)(conditions[i].Aux)
		// cPwqConditions[i].digital = (C.PS3000A_TRIGGER_STATE)(conditions[i].Digital)
	}
	stat := C.ps3000aSetPulseWidthQualifier((C.short)(ps.handle),
		(*C.PS3000A_PWQ_CONDITIONS)(&cPwqConditions[0]), (C.short)(len(conditions)),
		(C.PS3000A_THRESHOLD_DIRECTION)(direction), (C.uint)(lower), (C.uint)(upper),
		(C.PS3000A_PULSE_WIDTH_TYPE)(pwType))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetPulseWidthQualifier:  %s", psc.StatStr(int(stat)))
	}
	return
}
func (ps *PsDesc) SetTriggerDigitalPortProperties(digitalDirections []DigitalChannelDirections) (err error) {
	cDigitalDirections := make([]C.PS3000A_DIGITAL_CHANNEL_DIRECTIONS, len(digitalDirections))
	for i := range digitalDirections {
		cDigitalDirections[i].channel = (C.PS3000A_DIGITAL_CHANNEL)(digitalDirections[i].Channel)
		cDigitalDirections[i].direction = (C.PS3000A_DIGITAL_DIRECTION)(digitalDirections[i].Direction)
	}
	stat := C.ps3000aSetTriggerDigitalPortProperties((C.short)(ps.handle),
		(*C.PS3000A_DIGITAL_CHANNEL_DIRECTIONS)(&cDigitalDirections[0]), (C.short)(len(digitalDirections)))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetTriggerDigitalPortProperties:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) Stop() (err error) {
	stat := C.ps3000aStop((C.short)(ps.handle))
	if stat != C.PICO_OK {
		err = fmt.Errorf("Stop:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) SetSigGenBuiltIn(offsetVoltage int32, pkToPK uint32, waveType WaveTypeEnum,
	startFrequency, stopFrequency, increment, dwellTime float32, sweepType SweepTypeEnum,
	operation ExtraOperations, shots, sweeps uint32, triggerType SigGenTrigType,
	triggerSource SigGenTrigSource, extInThreshold int16) (err error) {
	stat := C.ps3000aSetSigGenBuiltIn((C.short)(ps.handle), (C.int)(offsetVoltage),
		(C.uint)(pkToPK), (C.short)(waveType), (C.float)(startFrequency),
		(C.float)(stopFrequency), (C.float)(increment), (C.float)(dwellTime),
		(C.PS3000A_SWEEP_TYPE)(sweepType), (C.PS3000A_EXTRA_OPERATIONS)(operation),
		(C.uint)(shots), (C.uint)(sweeps), (C.PS3000A_SIGGEN_TRIG_TYPE)(triggerType),
		(C.PS3000A_SIGGEN_TRIG_SOURCE)(triggerSource), (C.short)(extInThreshold))

	if stat != C.PICO_OK {
		err = fmt.Errorf("SetSigGenBuiltIn:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) SetSigGenBuiltInV2(offsetVoltage int32, pkToPK uint32, waveType WaveTypeEnum,
	startFrequency, stopFrequency, increment, dwellTime float64, sweepType SweepTypeEnum,
	operation ExtraOperations, shots, sweeps uint32, triggerType SigGenTrigType,
	triggerSource SigGenTrigSource, extInThreshold int16) (err error) {
	stat := C.ps3000aSetSigGenBuiltInV2((C.short)(ps.handle), (C.int)(offsetVoltage),
		(C.uint)(pkToPK), (C.short)(waveType), (C.double)(startFrequency),
		(C.double)(stopFrequency), (C.double)(increment), (C.double)(dwellTime),
		(C.PS3000A_SWEEP_TYPE)(sweepType), (C.PS3000A_EXTRA_OPERATIONS)(operation),
		(C.uint)(shots), (C.uint)(sweeps), (C.PS3000A_SIGGEN_TRIG_TYPE)(triggerType),
		(C.PS3000A_SIGGEN_TRIG_SOURCE)(triggerSource), (C.short)(extInThreshold))

	if stat != C.PICO_OK {
		err = fmt.Errorf("SetSigGenBuiltIn:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) SigGenFrequencyToPhase(frequency float64, indexMode IndexMode, bufferLength uint32) (phase uint32, err error) {
	stat := C.ps3000aSigGenFrequencyToPhase((C.short)(ps.handle), (C.double)(frequency),
		(C.PS3000A_INDEX_MODE)(indexMode), (C.uint)(bufferLength), (*C.uint)(&phase))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SigGenFrequencyToPhase:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) SetNoCaptures(nCaptures uint32) (err error) {
	stat := C.ps3000aSetNoOfCaptures((C.short)(ps.handle), (C.uint)(nCaptures))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetNoCaptures:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) GetTriggerTimeOffset(segmentIndex uint32) (timeUpper, timeLower uint32, timeUnits TimeUnits, err error) {
	stat := C.ps3000aGetTriggerTimeOffset((C.short)(ps.handle), (*C.uint)(&timeUpper),
		(*C.uint)(&timeLower), (*C.PS3000A_TIME_UNITS)(&timeUnits), (C.uint)(segmentIndex))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetTriggerTimeOffset:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) GetTriggerTimeOffset64(segmentIndex uint32) (time int64, timeUnits TimeUnits, err error) {
	stat := C.ps3000aGetTriggerTimeOffset64((C.short)(ps.handle), (*C.long)(&time),
		(*C.PS3000A_TIME_UNITS)(&timeUnits), (C.uint)(segmentIndex))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetTriggerTimeOffset64:  %s", psc.StatStr(int(stat)))
	}
	return

}

func (ps *PsDesc) GetValuesTriggerTimeOffsetBulk(timesUpper, timesLower []uint32, timeUnits []TimeUnits,
	fromSegmentIndex, toSegmentIndex uint32) (err error) {
	stat := C.ps3000aGetValuesTriggerTimeOffsetBulk((C.short)(ps.handle), (*C.uint)(&timesUpper[0]),
		(*C.uint)(&timesLower[0]), (*C.PS3000A_TIME_UNITS)(&timeUnits[0]), (C.uint)(fromSegmentIndex),
		(C.uint)(toSegmentIndex))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetValuesTriggerTimeOffsetBulk:  %s", psc.StatStr(int(stat)))
	}

	return
}

func (ps *PsDesc) GetValuesTriggerTimeOffsetBulk64(times []int64, timeUnits []TimeUnits,
	fromSegmentIndex, toSegmentIndex uint32) (err error) {
	stat := C.ps3000aGetValuesTriggerTimeOffsetBulk64((C.short)(ps.handle), (*C.long)(&times[0]),
		(*C.PS3000A_TIME_UNITS)(&timeUnits[0]), (C.uint)(fromSegmentIndex),
		(C.uint)(toSegmentIndex))
	if stat != C.PICO_OK {
		err = fmt.Errorf("GetValuesTriggerTimeOffsetBulk64:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) HoldOff(holdOff uint64, holdOffType HoldOffType) (err error) {
	stat := C.ps3000aHoldOff((C.short)(ps.handle), (C.ulong)(holdOff), (C.PS3000A_HOLDOFF_TYPE)(holdOffType))
	if stat != C.PICO_OK {
		err = fmt.Errorf("HoldOff:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) LsReady() (ready int16, err error) {
	stat := C.ps3000aIsReady((C.short)(ps.handle), (*C.short)(&ready))
	if stat != C.PICO_OK {
		err = fmt.Errorf("LsReady:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) TriggerOrPulseWidthQualifierEnabled() (triggerEnabled, pulseWidthQualifierEnabledint16 int16, err error) {
	stat := C.ps3000aIsTriggerOrPulseWidthQualifierEnabled((C.short)(ps.handle),
		(*C.short)(&triggerEnabled), (*C.short)(&pulseWidthQualifierEnabledint16))
	if stat != C.PICO_OK {
		err = fmt.Errorf("TriggerOrPulseWidthQualifierEnabled:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) MemorySegments(nSegments uint32) (nMaxSamples int32, err error) {
	stat := C.ps3000aMemorySegments((C.short)(ps.handle),
		(C.uint)(nSegments), (*C.int)(&nMaxSamples))
	if stat != C.PICO_OK {
		err = fmt.Errorf("MemorySegments:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) NoOfStreamingValues() (noOfValues uint32, err error) {
	stat := C.ps3000aNoOfStreamingValues((C.short)(ps.handle),
		(*C.uint)(&noOfValues))
	if stat != C.PICO_OK {
		err = fmt.Errorf("NoOfStreamingValues:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) OpenUnitProgress() (retHandle, progressPercent, complete int16, err error) {
	stat := C.ps3000aOpenUnitProgress((*C.short)(&retHandle),
		(*C.short)(&progressPercent), (*C.short)(&complete))
	if stat != C.PICO_OK {
		err = fmt.Errorf("OpenUnitProgress:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) PingUnit() (err error) {
	stat := C.ps3000aPingUnit((C.short)(ps.handle))
	if stat != C.PICO_OK {
		err = fmt.Errorf("PingUnit:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) QueryOutputEdgeDetect() (state int16, err error) {
	stat := C.ps3000aQueryOutputEdgeDetect((C.short)(ps.handle),
		(*C.short)(&state))
	if stat != C.PICO_OK {
		err = fmt.Errorf("OpenUnitProgress:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) SetDigitalPort(port DigitalPort, enabled bool, logiclevel int16) (err error) {
	stat := C.ps3000aSetDigitalPort((C.short)(ps.handle),
		(C.PS3000A_DIGITAL_PORT)(port), (C.short)(boolToint16(enabled)), (C.short)(logiclevel))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetDigitalPort:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) SetOutputEdgeDetect(state int16) (err error) {
	stat := C.ps3000aSetOutputEdgeDetect((C.short)(ps.handle), (C.short)(state))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetOutputEdgeDetect:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) SetPulseWidthDigitalPortProperties(digitalDirections []DigitalChannelDirections) (err error) {
	cDigitalDirections := make([]C.PS3000A_DIGITAL_CHANNEL_DIRECTIONS, len(digitalDirections))
	for i := range digitalDirections {
		cDigitalDirections[i].channel = (C.PS3000A_DIGITAL_CHANNEL)(digitalDirections[i].Channel)
		cDigitalDirections[i].direction = (C.PS3000A_DIGITAL_DIRECTION)(digitalDirections[i].Direction)
	}
	stat := C.ps3000aSetPulseWidthDigitalPortProperties((C.short)(ps.handle),
		(*C.PS3000A_DIGITAL_CHANNEL_DIRECTIONS)(&cDigitalDirections[0]), (C.short)(len(digitalDirections)))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetPulseWidthDigitalPortProperties:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) SetSigGenArbitrary(offsetVoltage int32, pkToPK uint32,
	startDeltaPhase, stopDeltaPhase, deltaPhaseIncrement, dwellCount uint32,
	arbitraryWaveform []int16, sweepType SweepTypeEnum, operation ExtraOperations,
	indexMode IndexMode, shots, sweeps uint32, triggerType SigGenTrigType,
	triggerSource SigGenTrigSource, extInThreshold int16) (err error) {
	stat := C.ps3000aSetSigGenArbitrary((C.short)(ps.handle), (C.int)(offsetVoltage),
		(C.uint)(pkToPK), (C.uint)(startDeltaPhase), (C.uint)(stopDeltaPhase),
		(C.uint32_t)(deltaPhaseIncrement), (C.uint32_t)(dwellCount),
		(*C.short)(&arbitraryWaveform[0]), (C.int32_t)(len(arbitraryWaveform)),
		(C.PS3000A_SWEEP_TYPE)(sweepType), (C.PS3000A_EXTRA_OPERATIONS)(operation),
		(C.PS3000A_INDEX_MODE)(indexMode),
		(C.uint)(shots), (C.uint)(sweeps), (C.PS3000A_SIGGEN_TRIG_TYPE)(triggerType),
		(C.PS3000A_SIGGEN_TRIG_SOURCE)(triggerSource), (C.short)(extInThreshold))

	if stat != C.PICO_OK {
		err = fmt.Errorf("SetSigGenArbitrary:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) SetSigGenPropertiesArbitrary(offsetVoltage int32,
	startDeltaPhase, stopDeltaPhase, deltaPhaseIncrement, dwellCount uint32,
	sweepType SweepTypeEnum, operation ExtraOperations,
	indexMode IndexMode, shots, sweeps uint32, triggerType SigGenTrigType,
	triggerSource SigGenTrigSource, extInThreshold int16) (err error) {
	stat := C.ps3000aSetSigGenPropertiesArbitrary((C.short)(ps.handle),
		(C.uint)(startDeltaPhase), (C.uint)(stopDeltaPhase),
		(C.uint32_t)(deltaPhaseIncrement), (C.uint32_t)(dwellCount),
		(C.PS3000A_SWEEP_TYPE)(sweepType),
		(C.uint)(shots), (C.uint)(sweeps), (C.PS3000A_SIGGEN_TRIG_TYPE)(triggerType),
		(C.PS3000A_SIGGEN_TRIG_SOURCE)(triggerSource), (C.short)(extInThreshold))

	if stat != C.PICO_OK {
		err = fmt.Errorf("SetSigGenPropertiesArbitrary:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) SetSigGenPropertiesBuiltIn(offsetVoltage int32,
	startFrequency, stopFrequency, increment, dwellTime float64,
	sweepType SweepTypeEnum,
	shots, sweeps uint32, triggerType SigGenTrigType,
	triggerSource SigGenTrigSource, extInThreshold int16) (err error) {
	stat := C.ps3000aSetSigGenPropertiesBuiltIn((C.short)(ps.handle),
		(C.double)(startFrequency), (C.double)(stopFrequency),
		(C.double)(increment), (C.double)(dwellTime),
		(C.PS3000A_SWEEP_TYPE)(sweepType),
		(C.uint)(shots), (C.uint)(sweeps), (C.PS3000A_SIGGEN_TRIG_TYPE)(triggerType),
		(C.PS3000A_SIGGEN_TRIG_SOURCE)(triggerSource), (C.short)(extInThreshold))

	if stat != C.PICO_OK {
		err = fmt.Errorf("SetSigGenPropertiesBuiltIn:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) SigGenArbitraryMinMaxValues() (minArbitraryWaveformValue, maxArbitraryWaveformValue int16,
	minArbitraryWaveformSize, maxArbitraryWaveformSize uint32, err error) {
	stat := C.ps3000aSigGenArbitraryMinMaxValues((C.short)(ps.handle),
		(*C.short)(&minArbitraryWaveformValue), (*C.short)(&maxArbitraryWaveformValue),
		(*C.uint32_t)(&minArbitraryWaveformSize), (*C.uint32_t)(&maxArbitraryWaveformSize))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SigGenArbitraryMinMaxValues:  %s", psc.StatStr(int(stat)))
	}
	return
}

func (ps *PsDesc) SigGenSoftwareControl(state int16) (err error) {
	stat := C.ps3000aSigGenSoftwareControl((C.short)(ps.handle),
		(C.short)(state))
	if stat != C.PICO_OK {
		err = fmt.Errorf("SetSigGenBuiltIn:  %s", psc.StatStr(int(stat)))
	}
	return
}
