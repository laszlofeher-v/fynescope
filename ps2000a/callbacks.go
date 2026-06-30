//go:build !noscope

package ps2000a

/*
#include <stdio.h>
#include "/opt/picoscope/include/libps2000a/PicoStatus.h"
// C callback function
int lpDataReady(int16_t handle, PICO_STATUS status, uint32_t noOfSamples,
				int16_t overflow, void * pParameter)
{
	int lpDataReadyGo(int16_t handle, PICO_STATUS status, uint32_t noOfSamples,
				int16_t overflow, void * pParameter);
	return lpDataReadyGo(handle, status, noOfSamples, overflow, pParameter);
}

int lpBlockReady(int16_t handle, PICO_STATUS status, void * pParameter)
{
	int lpBlockReadyGo(int16_t handle, PICO_STATUS status, void * pParameter);
	return lpBlockReadyGo(handle, status,  pParameter);
}

int lpStreamingReady(int16_t handle, int32_t noOfSamples, uint32_t startIndex,
                int16_t overflow, uint32_t triggerAt, int16_t triggered,
                int16_t autoStop, void * pParameter)
{
	int lpStreamingReadyGo(int16_t handle, int32_t noOfSamples, uint32_t startIndex,
                int16_t overflow, uint32_t triggerAt, int16_t triggered,
                int16_t autoStop, void * pParameter);
	return lpStreamingReadyGo(handle,noOfSamples, startIndex,overflow,triggerAt, triggered,autoStop,pParameter);
}
*/
import "C"
