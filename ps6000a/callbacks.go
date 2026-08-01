//go:build !demo && ps6000a

package ps6000a

/*
#include <stdio.h>
#include "/opt/picoscope/include/libps6000a/PicoStatus.h"
// C callback function
int ps6000aLpDataReady(int16_t handle, PICO_STATUS status, uint32_t noOfSamples,
				int16_t overflow, void * pParameter)
{
	int ps6000aLpDataReadyGo(int16_t handle, PICO_STATUS status, uint32_t noOfSamples,
				int16_t overflow, void * pParameter);
	return ps6000aLpDataReadyGo(handle, status, noOfSamples, overflow, pParameter);
}

int ps6000aLpBlockReady(int16_t handle, PICO_STATUS status, void * pParameter)
{
	int ps6000aLpBlockReadyGo(int16_t handle, PICO_STATUS status, void * pParameter);
	return ps6000aLpBlockReadyGo(handle, status,  pParameter);
}

int ps6000aLpStreamingReady(int16_t handle, int32_t noOfSamples, uint32_t startIndex,
                int16_t overflow, uint32_t triggerAt, int16_t triggered,
                int16_t autoStop, void * pParameter)
{
	int ps6000aLpStreamingReadyGo(int16_t handle, int32_t noOfSamples, uint32_t startIndex,
                int16_t overflow, uint32_t triggerAt, int16_t triggered,
                int16_t autoStop, void * pParameter);
	return ps6000aLpStreamingReadyGo(handle,noOfSamples, startIndex,overflow,triggerAt, triggered,autoStop,pParameter);
}
*/
import "C"
