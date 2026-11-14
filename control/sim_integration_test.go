//go:build sim

package control

import (
	"fynescope/demo"
	"fynescope/genericps"
	_ "fynescope/ps2000a"
	"fynescope/settings"
	"sync"
	"testing"
	"time"
)

func TestSimIntegration_GeneratorDisplaysSignal(t *testing.T) {
	testCases := []struct {
		name       string
		serial     string
		maxRate    uint32
		screenTime float64
		freq       float64
		pkToPk     uint32
		vRange     genericps.RangeEnum
		trigSource genericps.ChannelId
		trigMode   TriggerModes
		asyncGen   bool
	}{
		{
			name:       "2407BSIM_100kHz_BlockMode",
			serial:     "2407BSIM",
			maxRate:    MaxSampling1G,
			screenTime: 50e-6, // 5 us/div
			freq:       100000,
			pkToPk:     2000000, // 2Vpp (+/-1V)
			vRange:     genericps.Range_5v,
			trigSource: genericps.ChA,
			trigMode:   Auto,
			asyncGen:   false,
		},
		{
			name:       "2407BSIM_1MHz_100nsDiv_AsyncGen",
			serial:     "2407BSIM",
			maxRate:    MaxSampling1G,
			screenTime: 1e-6, // 100 ns/div
			freq:       1e6,
			pkToPk:     2000000,
			vRange:     genericps.Range_5v,
			trigSource: genericps.ChA,
			trigMode:   Auto,
			asyncGen:   true,
		},
		{
			name:       "2207BSIM_100kHz_5msDiv",
			serial:     "2207BSIM",
			maxRate:    MaxSampling1G,
			screenTime: 0.05, // 5 ms/div
			freq:       100000,
			pkToPk:     2000000,
			vRange:     genericps.Range_2v,
			trigSource: genericps.ChA,
			trigMode:   Auto,
			asyncGen:   true,
		},
		{
			name:       "2207BMSOSIM_1kHz_500usDiv",
			serial:     "2207BMSOSIM",
			maxRate:    MaxSampling1G,
			screenTime: 5e-3, // 500 us/div
			freq:       1000,
			pkToPk:     2000000,
			vRange:     genericps.Range_5v,
			trigSource: genericps.ChA,
			trigMode:   Auto,
			asyncGen:   true,
		},
		{
			name:       "2407BSIM_StreamMode_500msDiv",
			serial:     "2407BSIM",
			maxRate:    MaxSampling1G,
			screenTime: 5.0, // > StreamThreshold (2.1s)
			freq:       10,
			pkToPk:     2000000,
			vRange:     genericps.Range_5v,
			trigSource: genericps.ChA,
			trigMode:   Auto,
			asyncGen:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			demo.ScopeSimVariantInfo = tc.serial
			con := genericps.NewConnection()
			handle, err := genericps.OpenUnit(con, "ps2000a", tc.serial, 0)
			if err != nil {
				t.Fatalf("OpenUnit failed: %v", err)
			}
			con.Handle = handle
			defer con.CloseUnit()

			psControl := NewControl(con)
			defer psControl.Shutdown()

			psControl.SetInfo(tc.serial)
			psControl.SetScopeModel(StringToScopeType(tc.serial))
			psControl.SetMaxSamplingRate(tc.maxRate)
			psControl.SetStreamEnabled(tc.screenTime >= StreamThreshold)
			psControl.NewChannels(4)

			var mu sync.Mutex
			var gotFrames int
			var lastMin, lastMax int16
			var lastLen int
			var lastInterval float64
			var lastXRoundErr float64
			var triggeredCount int

			psControl.SetBufferCallback(func(size int) {})
			psControl.SetTriggerStateCallback(func(triggered bool) {
				mu.Lock()
				defer mu.Unlock()
				if triggered {
					triggeredCount++
				}
			})
			psControl.SetRefreshCallback(func(buffers [][]int16, buffersMin [][]int16, digitalBuffers [][]int16, startTimeOffset int64, xRoundError float64, samplingTimeInterval float64) {
				mu.Lock()
				defer mu.Unlock()
				gotFrames++
				lastInterval = samplingTimeInterval
				lastXRoundErr = xRoundError
				if len(buffers) > 0 && len(buffers[0]) > 0 {
					lastLen = len(buffers[0])
					minV, maxV := buffers[0][0], buffers[0][0]
					for _, s := range buffers[0] {
						if s < minV {
							minV = s
						}
						if s > maxV {
							maxV = s
						}
					}
					lastMin, lastMax = minV, maxV
				}
			})

			psControl.SetScopeScreenWidth(1000)
			psControl.SetMaxScreenTime(tc.screenTime)
			psControl.SetChannel(&settings.ChSettings{
				ID:         genericps.ChA,
				Enabled:    true,
				CoupleType: genericps.Dc,
				VRange:     tc.vRange,
			})

			trigDone := make(chan struct{}, 1)
			psControl.SetTrigger(&TriggerDescMsg{
				TriggerDesc: TriggerDesc{
					Enabled:            true,
					Source:             tc.trigSource,
					ThresholdDirection: genericps.TriggerRising,
					Mode:               tc.trigMode,
					Type:               Simple,
					TriggerADC:         0,
				},
				Done: trigDone,
			})
			<-trigDone

			genMsg := &GeneratorDescMsg{
				GeneratorDesc: GeneratorDesc{
					WaveType:       genericps.Sine,
					StartFrequency: tc.freq,
					StopFrequency:  tc.freq,
					Increment:      0,
					DwellTime:      1,
					SweepType:      genericps.SweepDown,
					PkToPK:         tc.pkToPk,
					OffsetVoltage:  0,
					On:             true,
				},
			}

			if tc.asyncGen {
				go psControl.SetGenerator(genMsg)
			} else {
				genDone := make(chan struct{}, 1)
				genMsg.Done = genDone
				psControl.SetGenerator(genMsg)
				<-genDone
			}

			if err := psControl.SetBlockMode(); err != nil {
				t.Fatalf("SetBlockMode failed: %v", err)
			}

			waitFor := 3 * time.Second
			deadline := time.Now().Add(waitFor)
			for time.Now().Before(deadline) {
				mu.Lock()
				done := gotFrames >= 2 && (lastMax > 100 || lastMin < -100)
				mu.Unlock()
				if done {
					break
				}
				time.Sleep(20 * time.Millisecond)
			}

			_ = psControl.Stop()

			mu.Lock()
			t.Logf("frames=%d triggered=%d len=%d min=%d max=%d interval=%g xRoundErr=%g",
				gotFrames, triggeredCount, lastLen, lastMin, lastMax, lastInterval, lastXRoundErr)
			if gotFrames == 0 {
				t.Errorf("expected at least 1 RefreshCallback frame, got 0")
			}
			if lastLen == 0 {
				t.Errorf("expected non-empty sample buffer")
			}
			if lastMax <= 100 && lastMin >= -100 {
				t.Errorf("expected non-zero generator signal, got min=%d max=%d", lastMin, lastMax)
			}
			mu.Unlock()
		})
	}
}
