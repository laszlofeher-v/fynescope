//go:build !demo && ps4000a

package ps4000a

/*
#include <stdio.h>
#include "/opt/picoscope/include/libps4000a/PicoStatus.h"
// C callback function
int ps4000aLpDataReady(int16_t handle, PICO_STATUS status, uint32_t noOfSamples,
				int16_t overflow, void * pParameter)
{
	int ps4000aLpDataReadyGo(int16_t handle, PICO_STATUS status, uint32_t noOfSamples,
				int16_t overflow, void * pParameter);
	return ps4000aLpDataReadyGo(handle, status, noOfSamples, overflow, pParameter);
}

int ps4000aLpBlockReady(int16_t handle, PICO_STATUS status, void * pParameter)
{
	int ps4000aLpBlockReadyGo(int16_t handle, PICO_STATUS status, void * pParameter);
	return ps4000aLpBlockReadyGo(handle, status,  pParameter);
}

int ps4000aLpStreamingReady(int16_t handle, int32_t noOfSamples, uint32_t startIndex,
                int16_t overflow, uint32_t triggerAt, int16_t triggered,
                int16_t autoStop, void * pParameter)
{
	int ps4000aLpStreamingReadyGo(int16_t handle, int32_t noOfSamples, uint32_t startIndex,
                int16_t overflow, uint32_t triggerAt, int16_t triggered,
                int16_t autoStop, void * pParameter);
	return ps4000aLpStreamingReadyGo(handle,noOfSamples, startIndex,overflow,triggerAt, triggered,autoStop,pParameter);
}
*/
import "C"
