//go:build multi

package gui

import (
	"encoding/json"
	"image/color"
	"net"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
	"fynescope/genericps"
	"fynescope/settings"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMultiscopeRemoteChannelAnnouncementAndWaveforms(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	serverScp := &ScpDesc{
		App:          app,
		channelCount: 2,
		Settings: &settings.PsSettings{
			Multiscope: settings.MultiscopeSettings{
				ServerPort: 45789,
			},
			Channels: []settings.ChSettings{
				{Enabled: true, VRange: genericps.Range_1v},
				{Enabled: false, VRange: genericps.Range_1v},
			},
		},
	}

	// Start server
	serverScp.startMultiServer(45789)
	defer serverScp.stopMultiServer()
	time.Sleep(50 * time.Millisecond)

	// Connect a simulated client
	conn, err := net.Dial("tcp", "127.0.0.1:45789")
	require.NoError(t, err)
	defer conn.Close()

	// Handshake
	infoMsg := MultiClientInfoMsg{Type: "ClientInfo", ScopeType: "TestClientScope"}
	b, _ := json.Marshal(infoMsg)
	conn.Write(append(b, '\n'))
	time.Sleep(50 * time.Millisecond)

	// Send RemoteChannelsAnnounce
	announce := RemoteChannelsAnnounceMsg{
		Type:     "RemoteChannelsAnnounce",
		ClientID: "127.0.0.1:fake",
		Channels: []RemoteChannelInfo{
			{
				RemoteChanIdx: 0,
				Name:          "Ch A",
				Enabled:       true,
				VRange:        "±2V",
				CoupleType:    "DC",
				Offset:        0.1,
				Inverted:      false,
				X10:           false,
				ColorR:        255,
				ColorG:        100,
				ColorB:        50,
			},
			{
				RemoteChanIdx: 1,
				Name:          "Ch B",
				Enabled:       true,
				VRange:        "±500mV",
				CoupleType:    "AC",
				Offset:        -0.2,
				Inverted:      true,
				X10:           true,
			},
		},
	}
	b, _ = json.Marshal(announce)
	conn.Write(append(b, '\n'))
	time.Sleep(100 * time.Millisecond)

	// Verify server registered the channels
	serverScp.remoteChannelsMu.RLock()
	require.Len(t, serverScp.remoteChannels, 2)
	assert.Equal(t, 0, serverScp.remoteChannels[0].RemoteChanIdx)
	assert.Equal(t, genericps.Range_2v, serverScp.remoteChannels[0].VRange)
	assert.Equal(t, genericps.Dc, serverScp.remoteChannels[0].CoupleType)
	assert.Equal(t, color.NRGBA{R: 255, G: 100, B: 50, A: 255}, serverScp.remoteChannels[0].Color)

	assert.Equal(t, 1, serverScp.remoteChannels[1].RemoteChanIdx)
	assert.Equal(t, genericps.Range_500mv, serverScp.remoteChannels[1].VRange)
	assert.Equal(t, genericps.Ac, serverScp.remoteChannels[1].CoupleType)
	assert.True(t, serverScp.remoteChannels[1].Inverted)
	serverScp.remoteChannelsMu.RUnlock()

	// Verify numberOfAllEnabledChannels includes remote channels
	assert.Equal(t, 3, serverScp.numberOfAllEnabledChannels()) // 1 local + 2 remote

	// Send Waveform data for Ch A
	samples := []float32{100.5, 200.3, 300.1, 400.0}
	wfMsg := RemoteWaveformMsg{
		Type:          "RemoteWaveform",
		ClientID:      conn.LocalAddr().String(),
		RemoteChanIdx: 0,
		Name:          "Ch A",
		Samples:       samples,
	}
	b, _ = json.Marshal(wfMsg)
	conn.Write(append(b, '\n'))
	time.Sleep(100 * time.Millisecond)

	// Verify server updated buffer
	serverScp.remoteChannelsMu.RLock()
	assert.Equal(t, samples, serverScp.remoteChannels[0].Buffer)
	serverScp.remoteChannelsMu.RUnlock()

	// Test server broadcasting StartAcquisition and StopAcquisition
	receivedMsgs := make(chan string, 10)
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				return
			}
			receivedMsgs <- string(buf[:n])
		}
	}()

	serverScp.broadcastMultiStart()
	select {
	case msg := <-receivedMsgs:
		assert.Contains(t, msg, "StartAcquisition")
	case <-time.After(1 * time.Second):
		t.Fatal("Timed out waiting for StartAcquisition")
	}

	serverScp.broadcastMultiStop()
	select {
	case msg := <-receivedMsgs:
		assert.Contains(t, msg, "StopAcquisition")
	case <-time.After(1 * time.Second):
		t.Fatal("Timed out waiting for StopAcquisition")
	}
}

func TestRemoteChannelsPanelWindow(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	scp := &ScpDesc{
		App: app,
		remoteChannels: []RemoteChannelDesc{
			{
				ClientID:       "127.0.0.1:5000",
				RemoteChanIdx:  0,
				Name:           "Ch A (127.0.0.1:5000)",
				Enabled:        true,
				VRange:         genericps.Range_1v,
				CoupleType:     genericps.Dc,
				Offset:         0.0,
				Color:          color.NRGBA{R: 255, G: 165, B: 0, A: 255},
				DisplayVOffset: 0,
			},
		},
	}

	scp.openRemoteChannelsWindow()
	require.NotNil(t, scp.remoteChannelsWindow)
	assert.Equal(t, "Remote Channels", scp.remoteChannelsWindow.Title())
	require.NotNil(t, scp.remoteChannelsContainer)

	// Modify channel and update window
	scp.remoteChannels[0].Offset = 0.5
	scp.updateRemoteChannelsWindow()
	assert.True(t, len(scp.remoteChannelsContainer.Objects) > 0)
}
