package control

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEtsTimes_2407B(t *testing.T) {
	ps := &PscDesc{
		Info:            "2407B",
		ScopeModel:      StringToScopeType("2407B"),
		MaxSamplingRate: MaxSampling500M,
	}

	tests := []struct {
		sampleTime int32
		wantInter  int16
		wantCycles int16
		wantErr    bool
	}{
		{1000, 2, 4, false},
		{500, 4, 8, false},
		{400, 5, 10, false},
		{200, 10, 20, false},
		{100, 20, 40, false},
		{50, 40, 80, false},
		{40, 0, 0, true},   // Below 50
		{1001, 0, 0, true}, // Above 1000
	}

	for _, tt := range tests {
		cycles, inter, err := ps.EtsTimes(tt.sampleTime)
		if tt.wantErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.wantInter, inter)
			assert.Equal(t, tt.wantCycles, cycles)
		}
	}
}

func TestEtsTimes_2207SIM(t *testing.T) {
	ps := &PscDesc{
		Info:            "2207SIM",
		ScopeModel:      StringToScopeType("2207SIM"),
		MaxSamplingRate: MaxSampling500M,
	}
	cycles, inter, err := ps.EtsTimes(200)
	assert.NoError(t, err)
	assert.Equal(t, int16(10), inter)
	assert.Equal(t, int16(20), cycles)
}
