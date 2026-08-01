//go:build demo

package ps2000a

import (
	"math"
	"sync/atomic"
)

var simParams struct {
	triggerTimeOffset atomic.Uint64 // float64 bits
}

// SetTriggerTimeOffset sets the trigger time offset (sampling goroutine → read goroutine).
func SetTriggerTimeOffset(v float64) {
	simParams.triggerTimeOffset.Store(math.Float64bits(v))
}

// GetTriggerTimeOffset returns the current trigger time offset.
func GetTriggerTimeOffset() float64 {
	return math.Float64frombits(simParams.triggerTimeOffset.Load())
}
