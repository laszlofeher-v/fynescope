//go:build !noscope && ps2000

package ps2000

/*
#include <stdio.h>
#include <stdint.h>
// C callback function
void lpStreamingReady2000(
  int16_t **overviewBuffers,
  int16_t   overflow,
  uint32_t  triggeredAt,
  int16_t   triggered,
  int16_t   auto_stop,
  uint32_t  nValues
)
{
	void lpStreamingReadyGo2000(int16_t **overviewBuffers, int16_t overflow, uint32_t triggeredAt, int16_t triggered, int16_t auto_stop, uint32_t nValues);
	lpStreamingReadyGo2000(overviewBuffers, overflow, triggeredAt, triggered, auto_stop, nValues);
}
*/
import "C"
