package control

import (
	"sync/atomic"
	"testing"
)

func TestPscDesc_checkOverflow(t *testing.T) {
	psControl := &PscDesc{}
	
	// Mock 4 channels
	psControl.chEnabled = make([]atomic.Bool, 4)
	
	var lastStatusMsg string
	var lastStatusErr ScopeError
	var statusCalls int
	
	psControl.DisplayStatus = func(s string, errorType ScopeError) {
		lastStatusMsg = s
		lastStatusErr = errorType
		statusCalls++
	}

	// Test 1: No overflow
	psControl.checkOverflow(0)
	if statusCalls != 0 {
		t.Errorf("Expected 0 status calls for no overflow, got %d", statusCalls)
	}

	// Test 2: Overflow on channel A, but channel A is disabled
	psControl.chEnabled[0].Store(false)
	psControl.checkOverflow(1) // Bit 0 is ChA
	if statusCalls != 0 {
		t.Errorf("Expected 0 status calls when disabled channel overflows, got %d", statusCalls)
	}

	// Test 3: Overflow on channel A, and channel A is enabled
	psControl.chEnabled[0].Store(true)
	psControl.checkOverflow(1)
	if statusCalls != 1 {
		t.Errorf("Expected 1 status call, got %d", statusCalls)
	}
	if lastStatusMsg != "Overflow error on channel:A " {
		t.Errorf("Unexpected status message: %q", lastStatusMsg)
	}
	if lastStatusErr != Warning {
		t.Errorf("Expected Warning error type, got %v", lastStatusErr)
	}

	// Test 4: Overflow on multiple channels (A and C)
	psControl.chEnabled[2].Store(true) // Enable ChC
	statusCalls = 0
	psControl.checkOverflow(5) // Bit 0 (ChA) and Bit 2 (ChC)
	if statusCalls != 1 {
		t.Errorf("Expected 1 status call, got %d", statusCalls)
	}
	if lastStatusMsg != "Overflow error on channels:A C " {
		t.Errorf("Unexpected status message: %q", lastStatusMsg)
	}
}
