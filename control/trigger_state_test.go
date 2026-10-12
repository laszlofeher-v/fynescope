package control

import (
	"sync/atomic"
	"testing"
	"time"

	_ "fynescope/demo"
	"fynescope/genericps"

	"github.com/stretchr/testify/assert"
)

func TestBlockTriggered(t *testing.T) {
	tests := []struct {
		name    string
		active  bool
		autoMs  int32
		waited  time.Duration
		trigger bool
	}{
		{"disabled trigger never fires", false, 0, 0, false},
		{"disabled trigger with auto never fires", false, 1000, 10 * time.Millisecond, false},
		{"no auto trigger completes only on event", true, 0, 30 * time.Second, true},
		{"quick event with auto trigger", true, 1000, 5 * time.Millisecond, true},
		{"event just before the timeout", true, 1000, 850 * time.Millisecond, true},
		{"released by auto trigger", true, 1000, 1000 * time.Millisecond, false},
		{"auto trigger plus latency", true, 1000, 1080 * time.Millisecond, false},
		{"auto trigger at detection boundary", true, 1000, 900 * time.Millisecond, false},
		{"negative wait estimate counts as event", true, 1000, -time.Second, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.trigger, blockTriggered(tt.active, tt.autoMs, tt.waited))
		})
	}
}

func TestSetTriggerStateCallback(t *testing.T) {
	ps := &PscDesc{}
	// no callback registered must be harmless
	ps.notifyTriggerState(true)

	var got []bool
	ps.SetTriggerStateCallback(func(triggered bool) { got = append(got, triggered) })
	ps.notifyTriggerState(true)
	ps.notifyTriggerState(false)
	assert.Equal(t, []bool{true, false}, got)

	ps.SetTriggerStateCallback(nil)
	ps.notifyTriggerState(true)
	assert.Equal(t, []bool{true, false}, got, "removed callback must not be called")
}

func TestReportBlockTriggerState(t *testing.T) {
	var last atomic.Int32 // 0 = never called, 1 = false, 2 = true
	newPs := func(enabled bool, autoMs int32, armedAgo time.Duration) *PscDesc {
		ps := &PscDesc{}
		ps.triggerSetting.Enabled = enabled
		ps.activeAutoTriggerMs.Store(autoMs)
		ps.blockRunStart = time.Now().Add(-armedAgo)
		last.Store(0)
		ps.SetTriggerStateCallback(func(triggered bool) {
			if triggered {
				last.Store(2)
			} else {
				last.Store(1)
			}
		})
		return ps
	}

	t.Run("quick trigger", func(t *testing.T) {
		newPs(true, 1000, 20*time.Millisecond).reportBlockTriggerState()
		assert.EqualValues(t, 2, last.Load())
	})
	t.Run("auto trigger timeout", func(t *testing.T) {
		newPs(true, 1000, 1200*time.Millisecond).reportBlockTriggerState()
		assert.EqualValues(t, 1, last.Load())
	})
	t.Run("long capture with real trigger", func(t *testing.T) {
		ps := newPs(true, 1000, 3*time.Second)
		// 2.5 s of the elapsed time is pure sample acquisition
		ps.SampleCountRequired = 2500
		ps.SamplingTimeInterval = 1e-3
		ps.reportBlockTriggerState()
		assert.EqualValues(t, 2, last.Load())
	})
	t.Run("long capture released by auto trigger", func(t *testing.T) {
		ps := newPs(true, 1000, 3600*time.Millisecond)
		ps.SampleCountRequired = 2500
		ps.SamplingTimeInterval = 1e-3
		ps.reportBlockTriggerState()
		assert.EqualValues(t, 1, last.Load())
	})
	t.Run("trigger disabled", func(t *testing.T) {
		newPs(false, 1000, 20*time.Millisecond).reportBlockTriggerState()
		assert.EqualValues(t, 1, last.Load())
	})
	t.Run("not armed yet", func(t *testing.T) {
		ps := newPs(true, 1000, 0)
		ps.blockRunStart = time.Time{}
		ps.reportBlockTriggerState()
		assert.EqualValues(t, 0, last.Load(), "nothing to report before the first RunBlock")
	})
	t.Run("no listener", func(t *testing.T) {
		ps := newPs(true, 1000, 0)
		ps.SetTriggerStateCallback(nil)
		assert.NotPanics(t, ps.reportBlockTriggerState)
	})
}

// The simulator completes the capture instantly, so its elapsed time must not be
// reduced by the acquisition time of a real scope.
func TestReportBlockTriggerStateDemoIgnoresCaptureTime(t *testing.T) {
	con := genericps.NewConnection()
	handle, err := genericps.OpenDemo(con, genericps.DemoId)
	if err != nil {
		t.Fatalf("Failed to open simulator: %v", err)
	}
	con.Handle = handle
	con.ID = genericps.DemoId
	defer con.CloseUnit()

	ps := &PscDesc{Con: con}
	if !ps.IsDemo() {
		t.Fatal("expected the simulator connection to be recognised as demo")
	}
	var triggered atomic.Bool
	triggered.Store(true)
	ps.triggerSetting.Enabled = true
	ps.activeAutoTriggerMs.Store(1000)
	ps.blockRunStart = time.Now().Add(-1100 * time.Millisecond)
	ps.SampleCountRequired = 2000
	ps.SamplingTimeInterval = 1e-3 // would hide the auto trigger if applied
	ps.SetTriggerStateCallback(func(b bool) { triggered.Store(b) })
	ps.reportBlockTriggerState()
	assert.False(t, triggered.Load())
}

func TestTriggerWrappersRecordAutoTrigger(t *testing.T) {
	con := genericps.NewConnection()
	handle, err := genericps.OpenDemo(con, genericps.DemoId)
	if err != nil {
		t.Fatalf("Failed to open simulator: %v", err)
	}
	con.Handle = handle
	defer con.CloseUnit()

	ps := &PscDesc{Con: con}
	assert.NoError(t, ps.setSimpleTrigger(true, genericps.ChA, 0, genericps.TriggerRising, 0, 1000))
	assert.EqualValues(t, 1000, ps.activeAutoTriggerMs.Load())

	assert.NoError(t, ps.setSimpleTrigger(true, genericps.ChA, 0, genericps.TriggerRising, 0, 0))
	assert.EqualValues(t, 0, ps.activeAutoTriggerMs.Load())

	props := []genericps.TriggerChannelProperties{{Channel: genericps.ChA}}
	_ = ps.setTriggerChannelProperties(props, false, 500)
	assert.EqualValues(t, 500, ps.activeAutoTriggerMs.Load())
}
