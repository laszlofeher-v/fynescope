package sim

import (
	"fynescope/genericps"
	"testing"
	"time"
)

func TestBoolToint16(t *testing.T) {
	tests := []struct {
		name string
		b    bool
		want int16
	}{
		{"true", true, 1},
		{"false", false, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := boolToint16(tt.b); got != tt.want {
				t.Errorf("boolToint16() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNextInt16(t *testing.T) {
	next := nextInt16()
	if next() != 1 {
		t.Errorf("nextInt16() = %v, want %v", 1, 1)
	}
	if next() != 2 {
		t.Errorf("nextInt16() = %v, want %v", 2, 2)
	}
}

func TestEnumerateUnits(t *testing.T) {
	tests := []struct {
		name          string
		behaviour     returnStatus
		wantCount     int16
		wantSerials   string
		wantSerialLth int16
		wantErr       bool
	}{
		{"normal", normal, 1, scopeBathAndSerialInfo, int16(len(scopeBathAndSerialInfo)), false},
		{"faulty", faulty, 0, "", 0, true},
		{"timeoutNormal", timeoutNormal, 1, scopeBathAndSerialInfo, int16(len(scopeBathAndSerialInfo)), false},
		{"timeoutfaulty", timeoutfaulty, 0, "", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			behaviour = tt.behaviour
			if tt.behaviour == timeoutNormal || tt.behaviour == timeoutfaulty {
				timeout = 10 * time.Millisecond
			}
			gotCount, gotSerials, gotSerialLth, err := EnumerateUnits(0)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnumerateUnits() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotCount != tt.wantCount {
				t.Errorf("EnumerateUnits() gotCount = %v, want %v", gotCount, tt.wantCount)
			}
			if gotSerials != tt.wantSerials {
				t.Errorf("EnumerateUnits() gotSerials = %v, want %v", gotSerials, tt.wantSerials)
			}
			if gotSerialLth != tt.wantSerialLth {
				t.Errorf("EnumerateUnits() gotSerialLth = %v, want %v", gotSerialLth, tt.wantSerialLth)
			}
			if tt.behaviour == timeoutNormal || tt.behaviour == timeoutfaulty {
				//reset timeout
				timeout = 8 * time.Second
			}
		})
	}
}

func TestOpenUnit(t *testing.T) {
	tests := []struct {
		name      string
		behaviour returnStatus
		wantErr   bool
	}{
		{"normal", normal, false},
		{"faulty", faulty, true},
		{"timeoutNormal", timeoutNormal, false},
		{"timeoutfaulty", timeoutfaulty, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			behaviour = tt.behaviour
			if tt.behaviour == timeoutNormal || tt.behaviour == timeoutfaulty {
				timeout = 10 * time.Millisecond
			}
			_, err := openUnit("test")
			if (err != nil) != tt.wantErr {
				t.Errorf("openUnit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.behaviour == timeoutNormal || tt.behaviour == timeoutfaulty {
				//reset timeout
				timeout = 8 * time.Second
			}
		})
	}
}

func TestOpenUnitAsync(t *testing.T) {
	tests := []struct {
		name       string
		behaviour  returnStatus
		wantStatus int16
		wantErr    bool
	}{
		{"normal", normal, 0, false},
		{"faulty", faulty, 0, true},
		{"timeoutNormal", timeoutNormal, 0, false},
		{"timeoutfaulty", timeoutfaulty, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			behaviour = tt.behaviour
			if tt.behaviour == timeoutNormal || tt.behaviour == timeoutfaulty {
				timeout = 10 * time.Millisecond
			}
			gotStatus, err := openUnitAsync("test")
			if (err != nil) != tt.wantErr {
				t.Errorf("openUnitAsync() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotStatus != tt.wantStatus {
				t.Errorf("openUnitAsync() status = %v, wantStatus %v", gotStatus, tt.wantStatus)
			}
			if tt.behaviour == timeoutNormal || tt.behaviour == timeoutfaulty {
				//reset timeout
				timeout = 8 * time.Second
			}
		})
	}
}

func TestOpenUnitProgress(t *testing.T) {
	tests := []struct {
		name                string
		behaviour           returnStatus
		wantHandle          int16
		wantProgressPercent int16
		wantComplete        int16
		wantErr             bool
	}{
		{"normal", normal, 1, 100, 1, false},
		{"faulty", faulty, 0, 0, 0, true},
		{"timeoutNormal", timeoutNormal, 1, 100, 1, false},
		{"timeoutfaulty", timeoutfaulty, 0, 0, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			behaviour = tt.behaviour
			if tt.behaviour == timeoutNormal || tt.behaviour == timeoutfaulty {
				timeout = 10 * time.Millisecond
			}
			gotHandle, gotProgressPercent, gotComplete, err := openUnitProgress()
			if (err != nil) != tt.wantErr {
				t.Errorf("openUnitProgress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantHandle > 0 && gotHandle <= 0 {
				t.Errorf("openUnitProgress() handle = %v, want handle to be grater than %v", gotHandle, tt.wantHandle)
				return
			}

			if gotProgressPercent != tt.wantProgressPercent {
				t.Errorf("openUnitProgress() progressPercent = %v, want %v", gotProgressPercent, tt.wantProgressPercent)
			}
			if gotComplete != tt.wantComplete {
				t.Errorf("openUnitProgress() complete = %v, want %v", gotComplete, tt.wantComplete)
			}
			if tt.behaviour == timeoutNormal || tt.behaviour == timeoutfaulty {
				//reset timeout
				timeout = 8 * time.Second
			}
		})
	}
}

func TestSimCloseUnit(t *testing.T) {
	err := simCloseUnit(1)
	if err != nil {
		t.Errorf("simCloseUnit() error = %v", err)
	}
}

// Add more tests when functions are more complex.
func TestLoadConstants(t *testing.T) {
	loadConstants()
	//just test if it runs, but it should run without panic
}
func TestDispatchNullMsg(t *testing.T) {
	rspCh := make(chan struct{}, 1)
	msg := &genericps.NullMsg{}
	msg.SetRspCh(rspCh)

	dispatch(msg)

	select {
	case <-rspCh:
		// Test passes if the channel receives a value
	case <-time.After(100 * time.Millisecond):
		t.Errorf("dispatch() did not send a value on the channel")
	}
}

func TestSimGetValues(t *testing.T) {
	// Initialize things
	loadConstants()
	behaviour = normal

	// Open unit
	h, err := openUnit("serial")
	if err != nil {
		t.Fatalf("openUnit failed: %v", err)
	}

	// Set channel
	err = simSetChannel(h, ChA, true, Dc, Range_1v, 0)
	if err != nil {
		t.Fatalf("simSetChannel failed: %v", err)
	}

	// Enable channel buffer
	buf := make([]int16, 1000)
	buffers[ChA] = buf

	// Set signal generator directly (since simSetSigGenBuiltInV2 is a no-op)
	channels[ChA].genOn = true
	channels[ChA].genWaveFunction = NewWaveformGenerator(Sine)
	channels[ChA].genPkToPk = 2000
	channels[ChA].genOffsetVoltage = 0
	channels[ChA].phase = 0
	channels[ChA].sweepController = NewSweepController(10000000, 10000000, 0, SweepUp, 0)

	// Set trigger detector
	err = simSetSimpleTrigger(h, true, ChA, 0, TriggerRaising, 0, 0)
	if err != nil {
		t.Fatalf("simSetSimpleTrigger failed: %v", err)
	}

	// Get values
	noSamples, overflow, err := simGetValues(h, 0, 1000, 1, RatioModeNone, 0)
	if err != nil {
		t.Fatalf("simGetValues failed: %v", err)
	}

	if noSamples != 1000 {
		t.Errorf("expected 1000 samples, got %d", noSamples)
	}

	// Check if buffer is not all zeros (random check)
	allZero := true
	for _, v := range buf {
		if v != 0 {
			allZero = false
			break
		}
	}
	// Sine wave should have non-zero values
	if allZero {
		t.Errorf("buffer contains all zeros")
	}

	// Manual check of overflow if needed
	if overflow != 0 {
		t.Logf("Overflow: %d", overflow)
	}
}
