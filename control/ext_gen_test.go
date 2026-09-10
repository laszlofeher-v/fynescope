package control

import (
	"testing"

	"fynescope/control/scpi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtGenDesc_Disconnected(t *testing.T) {
	var e ExtGenDesc

	assert.False(t, e.Connected())
	vid, pid := e.GetVidPid()
	assert.Equal(t, "", vid)
	assert.Equal(t, "", pid)

	assert.ErrorContains(t, e.SetFrequency(scpi.Ch1, 1000), "not connected")
	assert.ErrorContains(t, e.SetAmplitude(scpi.Ch1, 1.0), "not connected")
	assert.ErrorContains(t, e.SetOffset(scpi.Ch1, 0.0), "not connected")
	assert.ErrorContains(t, e.SetPhase(scpi.Ch1, 0.0), "not connected")
	assert.ErrorContains(t, e.SetOutput(scpi.Ch1, true), "not connected")
	assert.ErrorContains(t, e.SetWaveform(scpi.Ch1, "SIN"), "not connected")
	assert.ErrorContains(t, e.SetRampSymmetry(scpi.Ch1, "50%"), "not connected")
	assert.ErrorContains(t, e.SetImpedance(scpi.Ch1, "50"), "not connected")
	assert.ErrorContains(t, e.SendRaw("*IDN?"), "not connected")

	// Disconnect on disconnected descriptor should not panic
	e.Disconnect()
	assert.False(t, e.Connected())
}

func TestExtGenDesc_Connected(t *testing.T) {
	var e ExtGenDesc
	mockGen := scpi.NewMockGenerator()
	e.gen = mockGen

	assert.True(t, e.Connected())

	vid, pid := e.GetVidPid()
	assert.Equal(t, "0x0000", vid)
	assert.Equal(t, "0x0000", pid)

	require.NoError(t, e.SetFrequency(scpi.Ch1, 1000.0))
	require.NoError(t, e.SetAmplitude(scpi.Ch1, 2.5))
	require.NoError(t, e.SetOffset(scpi.Ch1, 0.5))
	require.NoError(t, e.SetPhase(scpi.Ch1, 90.0))
	require.NoError(t, e.SetOutput(scpi.Ch1, true))
	require.NoError(t, e.SetWaveform(scpi.Ch1, "SINusoid"))
	require.NoError(t, e.SetRampSymmetry(scpi.Ch1, "50"))
	require.NoError(t, e.SetImpedance(scpi.Ch1, "50"))
	require.NoError(t, e.SendRaw("*RST"))

	// Test Disconnect
	e.Disconnect()
	assert.False(t, e.Connected())
}

func TestExtGenDesc_Connect(t *testing.T) {
	var e ExtGenDesc
	// If e.gen is currently set, Connect closes it first.
	e.gen = scpi.NewMockGenerator()
	assert.True(t, e.Connected())

	cfg := ExtGenConfig{
		Port:   "COM1",
		UsbVid: "0x1234",
		UsbPid: "0x5678",
	}

	// Connect invokes scpi.New(cfg).Open(), which under demo/!scpi build returns an error, closing existing gen.
	_ = e.Connect(cfg)

	// Test connecting when gen is nil
	_ = e.Connect(cfg)
}
