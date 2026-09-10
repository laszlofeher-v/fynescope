package control

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringToScopeType(t *testing.T) {
	allScopes := []struct {
		str      string
		expected ScopeType
	}{
		{"2204A", Scope2204A},
		{"2205A", Scope2205A},
		{"2205A_MSO", Scope2205A_MSO},
		{"2205A MSO", Scope2205A_MSO},
		{"2206B", Scope2206B},
		{"2206B_MSO", Scope2206B_MSO},
		{"2207B", Scope2207B},
		{"2207B_MSO", Scope2207B_MSO},
		{"2208B", Scope2208B},
		{"2208B_MSO", Scope2208B_MSO},
		{"2405A", Scope2405A},
		{"2406B", Scope2406B},
		{"2407B", Scope2407B},
		{"2408B", Scope2408B},
		{"3203D", Scope3203D},
		{"3203D_MSO", Scope3203D_MSO},
		{"3204D", Scope3204D},
		{"3204D_MSO", Scope3204D_MSO},
		{"3205D", Scope3205D},
		{"3205D_MSO", Scope3205D_MSO},
		{"3206D", Scope3206D},
		{"3206D_MSO", Scope3206D_MSO},
		{"3403D", Scope3403D},
		{"3403D_MSO", Scope3403D_MSO},
		{"3404D", Scope3404D},
		{"3404D_MSO", Scope3404D_MSO},
		{"3405D", Scope3405D},
		{"3405D_MSO", Scope3405D_MSO},
		{"3406D", Scope3406D},
		{"3406D_MSO", Scope3406D_MSO},
		{"4262", Scope4262},
		{"4444", Scope4444},
		{"4823", Scope4823},
		{"4224A", Scope4224A},
		{"4424A", Scope4424A},
		{"4824A", Scope4824A},
		{"4225A", Scope4225A},
		{"4425A", Scope4425A},
		{"5242D", Scope5242D},
		{"5242D_MSO", Scope5242D_MSO},
		{"5243D", Scope5243D},
		{"5243D_MSO", Scope5243D_MSO},
		{"5244D", Scope5244D},
		{"5244D_MSO", Scope5244D_MSO},
		{"5442D", Scope5442D},
		{"5442D_MSO", Scope5442D_MSO},
		{"5443D", Scope5443D},
		{"5443D_MSO", Scope5443D_MSO},
		{"5444D", Scope5444D},
		{"5444D_MSO", Scope5444D_MSO},
		{"6403E", Scope6403E},
		{"6404E", Scope6404E},
		{"6405E", Scope6405E},
		{"6406E", Scope6406E},
		{"6424E", Scope6424E},
		{"6425E", Scope6425E},
		{"6426E", Scope6426E},
		{"6428E-D", Scope6428E_D},
		{"6428E_D", Scope6428E_D},
		{"6804E", Scope6804E},
		{"6824E", Scope6824E},
	}

	for _, tc := range allScopes {
		t.Run(tc.str, func(t *testing.T) {
			got := StringToScopeType(tc.str)
			assert.Equal(t, tc.expected, got)
		})
	}

	// Test SIM and DEMO variants
	assert.Equal(t, Scope2206B+ScopeSimulatedOffset, StringToScopeType("2206BSIM"))
	assert.Equal(t, Scope2206B_MSO+ScopeSimulatedOffset, StringToScopeType("2206BMSOSIM"))
	assert.Equal(t, Scope2206B_MSO+ScopeSimulatedOffset, StringToScopeType("2206B_MSOSIM"))
	assert.Equal(t, Scope2206B_MSO+ScopeSimulatedOffset, StringToScopeType("2206B MSOSIM"))
	assert.Equal(t, Scope2206B_MSO+ScopeSimulatedOffset, StringToScopeType("2206B-MSOSIM"))
	assert.Equal(t, Scope2206B+ScopeSimulatedOffset, StringToScopeType("2206BDEMO"))
	// Legacy missing "B"
	assert.Equal(t, Scope2407B+ScopeSimulatedOffset, StringToScopeType("2407SIM"))
	assert.Equal(t, Scope2207B+ScopeSimulatedOffset, StringToScopeType("2207DEMO"))

	// Unknown cases
	assert.Equal(t, ScopeUnknown, StringToScopeType("NonExistentScope"))
	assert.Equal(t, ScopeUnknown, StringToScopeType("NonExistentSIM"))
}

func TestScopeTypeString(t *testing.T) {
	for st := ScopeUnknown; st <= Scope6824E; st++ {
		str := st.String()
		assert.NotEmpty(t, str)
		if st == ScopeUnknown {
			assert.Equal(t, "Unknown", str)
		} else {
			assert.NotEqual(t, "Unknown", str, fmt.Sprintf("Missing String() representation for %d", st))
		}
	}

	// Test simulated string formatting
	simType := Scope2206B + ScopeSimulatedOffset
	assert.Equal(t, "2206BSIM", simType.String())

	simMso := Scope2206B_MSO + ScopeSimulatedOffset
	assert.Equal(t, "2206BMSOSIM", simMso.String())

	unknownSim := ScopeType(9999)
	assert.Equal(t, "Unknown", unknownSim.String())
}

func TestScopeTypeBase(t *testing.T) {
	assert.Equal(t, Scope2206B, (Scope2206B + ScopeSimulatedOffset).Base())
	assert.Equal(t, Scope3204D, Scope3204D.Base())
	assert.Equal(t, ScopeUnknown, ScopeUnknown.Base())
}

func TestScopeTypeIsETSCapable(t *testing.T) {
	// ETS capable scopes (2000B, 3000, 5000, 6000)
	assert.True(t, Scope2206B.IsETSCapable())
	assert.True(t, (Scope2206B + ScopeSimulatedOffset).IsETSCapable())
	assert.True(t, Scope3206D.IsETSCapable())
	assert.True(t, Scope5444D.IsETSCapable())
	assert.True(t, Scope6404E.IsETSCapable())

	// Non-ETS scopes
	assert.False(t, ScopeUnknown.IsETSCapable())
	assert.False(t, Scope4262.IsETSCapable())
	assert.False(t, Scope4444.IsETSCapable())
	assert.False(t, Scope4824A.IsETSCapable())
}
