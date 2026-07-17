//go:build scpi

package scpi

import (
	"bytes"
	"testing"
)

func TestGeneratorCommands(t *testing.T) {
	var buf bytes.Buffer
	gIface := New(Config{Port: "dummy"})
	g := gIface.(*Generator)
	g.Writer = &buf

	tests := []struct {
		name     string
		action   func() error
		expected string
	}{
		{
			name:     "SetFrequency",
			action:   func() error { return g.SetFrequency(Ch1, 1234.56) },
			expected: ":FREQuency1 1234.560000\n",
		},
		{
			name:     "SetAmplitude",
			action:   func() error { return g.SetAmplitude(Ch1, 2.5) },
			expected: "SOURce1:VOLTage:AMPLitude 2.500000\n",
		},
		{
			name:     "SetOffset",
			action:   func() error { return g.SetOffset(Ch1, 0.1) },
			expected: "SOURce1:VOLTage:OFFSet 0.100000\n",
		},
		{
			name:     "SetOutputON",
			action:   func() error { return g.SetOutput(Ch1, true) },
			expected: "OUTPut1:STATe ON\n",
		},
		{
			name:     "SetOutputOFF",
			action:   func() error { return g.SetOutput(Ch1, false) },
			expected: "OUTPut1:STATe OFF\n",
		},
		{
			name:     "SetWaveform",
			action:   func() error { return g.SetWaveform(Ch1, "SINE") },
			expected: "SOURce1:FUNCtion:SHAPe SINE\n",
		},
		{
			name:     "SetPhase",
			action:   func() error { return g.SetPhase(Ch1, 90.0) },
			expected: "SOUR1:PHAS 90.000000DEG\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			err := tt.action()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := buf.String()
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}



