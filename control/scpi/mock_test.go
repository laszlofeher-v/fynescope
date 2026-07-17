// SPDX-License-Identifier: MIT

package scpi

import (
	"testing"
)

// TestParseHex tests the pure hex parsing helper (no USB hardware required).
func TestParseHex(t *testing.T) {
	tests := []struct {
		input    string
		expected uint16
	}{
		{"0x5345", 0x5345},
		{"5345", 0x5345},
		{"0X1234", 0x1234},
		{"", 0},
		{"invalid", 0},
		{"FFFF", 0xFFFF},
		{"  0x00ab  ", 0x00ab}, // trimmed whitespace
	}

	for _, tt := range tests {
		got := parseHex(tt.input)
		if got != tt.expected {
			t.Errorf("parseHex(%q) = 0x%04x, expected 0x%04x", tt.input, got, tt.expected)
		}
	}
}

// TestMockGenerator ensures that MockGenerator satisfies GeneratorIface and
// that every method returns without error (no USB hardware required).
func TestMockGenerator(t *testing.T) {
	m := NewMockGenerator()

	if err := m.Open(); err != nil {
		t.Fatalf("Open() error: %v", err)
	}

	calls := []struct {
		name string
		fn   func() error
	}{
		{"SetFrequency", func() error { return m.SetFrequency(Ch1, 1000.0) }},
		{"SetAmplitude", func() error { return m.SetAmplitude(Ch1, 1.0) }},
		{"SetOffset", func() error { return m.SetOffset(Ch1, 0.5) }},
		{"SetPhase", func() error { return m.SetPhase(Ch1, 90.0) }},
		{"SetOutputON", func() error { return m.SetOutput(Ch1, true) }},
		{"SetOutputOFF", func() error { return m.SetOutput(Ch1, false) }},
		{"SetWaveform", func() error { return m.SetWaveform(Ch1, "SINE") }},
		{"SetRampSymmetry", func() error { return m.SetRampSymmetry(Ch1, "RAMP:SYMMetry 50") }},
		{"SetImpedance", func() error { return m.SetImpedance(Ch1, "50") }},
		{"Query", func() error { return m.Query("*IDN?") }},
		{"Send", func() error { return m.Send(":FREQuency1 1000") }},
	}

	for _, c := range calls {
		t.Run(c.name, func(t *testing.T) {
			if err := c.fn(); err != nil {
				t.Errorf("%s returned unexpected error: %v", c.name, err)
			}
		})
	}

	resp, err := m.QueryString("*IDN?")
	if err != nil {
		t.Errorf("QueryString error: %v", err)
	}
	if resp == "" {
		t.Error("QueryString returned empty string for MockGenerator")
	}

	vid, pid := m.GetVidPid()
	if vid == "" || pid == "" {
		t.Errorf("GetVidPid returned empty values: vid=%q pid=%q", vid, pid)
	}

	if err := m.Close(); err != nil {
		t.Fatalf("Close() error: %v", err)
	}
}

// TestDummyGenerator verifies that the no-build-tag dummyGen satisfies the
// interface and that Open returns a descriptive error.
func TestDummyGenerator(t *testing.T) {
	g := New(Config{Port: "dummy"})

	err := g.Open()
	if err == nil {
		t.Fatal("expected Open() to return an error for dummyGen, got nil")
	}

	// All remaining methods should succeed silently.
	noErrCalls := []struct {
		name string
		fn   func() error
	}{
		{"SetFrequency", func() error { return g.SetFrequency(Ch1, 1000.0) }},
		{"SetAmplitude", func() error { return g.SetAmplitude(Ch1, 1.0) }},
		{"SetOffset", func() error { return g.SetOffset(Ch1, 0.0) }},
		{"SetPhase", func() error { return g.SetPhase(Ch1, 0.0) }},
		{"SetOutput", func() error { return g.SetOutput(Ch1, false) }},
		{"SetWaveform", func() error { return g.SetWaveform(Ch1, "SINE") }},
		{"SetRampSymmetry", func() error { return g.SetRampSymmetry(Ch1, "RAMP:SYMMetry 50") }},
		{"SetImpedance", func() error { return g.SetImpedance(Ch1, "50") }},
		{"Query", func() error { return g.Query("*IDN?") }},
		{"Send", func() error { return g.Send(":FREQuency1 1000") }},
		{"Close", func() error { return g.Close() }},
	}

	for _, c := range noErrCalls {
		t.Run(c.name, func(t *testing.T) {
			if err := c.fn(); err != nil {
				t.Errorf("%s: unexpected error: %v", c.name, err)
			}
		})
	}
}
