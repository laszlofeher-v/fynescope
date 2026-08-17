package gui

import (
	"fynescope/genericps"
	"fynescope/settings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoveIndex(t *testing.T) {
	arr := []string{"A", "B", "C", "D"}
	
	arr = RemoveIndex(arr, 1)
	assert.Equal(t, []string{"A", "C", "D"}, arr)

	arr = RemoveIndex(arr, 2)
	assert.Equal(t, []string{"A", "C"}, arr)

	arr = RemoveIndex(arr, 0)
	assert.Equal(t, []string{"C"}, arr)
}

func TestSortMapString(t *testing.T) {
	testMap := map[string]genericps.RangeEnum{
		"10V":  genericps.Range_10v,  // higher numeric value
		"5V":   genericps.Range_5v,
		"20mV": genericps.Range_20mv, // lowest numeric value
	}

	// This function sorts keys descending by map values. 
	sorted := sortMapString(testMap)
	
	// Assuming Range10v > Range5v > Range20mv enum-wise
	assert.Equal(t, "10V", sorted[0])
	assert.Equal(t, "5V", sorted[1])
	assert.Equal(t, "20mV", sorted[2])
}

func TestNumberOfEnabledChannels(t *testing.T) {
	scp := &ScpDesc{
		channelCount: 4,
		Settings: &settings.PsSettings{
			Channels: []settings.ChSettings{
				{Enabled: true},  // bit 0 (value 1)
				{Enabled: false}, // bit 1 (value 2)
				{Enabled: true},  // bit 2 (value 4)
				{Enabled: false}, // bit 3 (value 8)
			},
		},
	}

	n, set := scp.numberOfEnabledChannels()
	assert.Equal(t, 2, n)
	assert.Equal(t, uint64(5), set) // 1 | 4 = 5
}

func TestNumberOfAllEnabledChannels(t *testing.T) {
	scp := &ScpDesc{
		channelCount: 2,
		Settings: &settings.PsSettings{
			Channels: []settings.ChSettings{
				{Enabled: true},
				{Enabled: false},
			},
			VirtualChannels: []settings.VirtualChSettings{
				{Enabled: true},
				{Enabled: true},
				{Enabled: false},
			},
		},
	}

	n := scp.numberOfAllEnabledChannels()
	assert.Equal(t, 3, n) // 1 physical + 2 virtual
}

func TestNthEnabledChannels(t *testing.T) {
	scp := &ScpDesc{
		channelCount: 4,
		Settings: &settings.PsSettings{
			Channels: []settings.ChSettings{
				{Enabled: false}, // index 0
				{Enabled: true},  // index 1
				{Enabled: false}, // index 2
				{Enabled: true},  // index 3
			},
		},
	}

	assert.Equal(t, 1, scp.nthEnabledChannels(0))
	assert.Equal(t, 3, scp.nthEnabledChannels(1))
	assert.Equal(t, -1, scp.nthEnabledChannels(2))
}
