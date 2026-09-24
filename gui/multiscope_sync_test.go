package gui

import (
	"encoding/json"
	"net"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"fynescope/control"
	"fynescope/genericps"
	"fynescope/settings"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	if len(genericps.InputRanges) == 0 {
		genericps.InputRanges = make([]int32, 40)
		for i := range genericps.InputRanges {
			genericps.InputRanges[i] = 1000
		}
	}
}

func TestBuildMultiSyncMessage(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	w := app.NewWindow("Test")
	defer w.Close()

	scp := &ScpDesc{
		Window:       w,
		channelCount: 2,
		Settings:     settings.NewDefaultSettings(),
	}

	scp.Settings.Multiscope.CommonTrigger = "Ch A"
	scp.Settings.Time.TimeDiv = "200"
	scp.Settings.Time.Unit = "us"
	scp.Settings.Time.TriggerTimeOffset = 0.001
	scp.Settings.Channels[0].VRange = genericps.Range_1v
	scp.Settings.Channels[0].CoupleType = genericps.Ac
	scp.Settings.Channels[0].X10 = true
	scp.Settings.Channels[0].Offset = 0.25
	scp.Settings.Channels[0].Trigger.Mv = 150
	scp.Settings.Channels[0].Trigger.TriggerDirection = genericps.TriggerRising

	msg, err := scp.buildMultiSyncMessage()
	require.NoError(t, err)
	require.NotNil(t, msg)
	assert.Equal(t, "SyncSettings", msg.Type)
	assert.Equal(t, "Ch A", msg.CommonTrigger)
	assert.Equal(t, "200", msg.TimeDiv)
	assert.Equal(t, "us", msg.TimeUnit)
	assert.Equal(t, 0.001, msg.TriggerTimeOffset)
	assert.Equal(t, "Rising", msg.TriggerDirection)
	assert.Equal(t, int32(150), msg.TriggerMv)
	require.Len(t, msg.Channels, 2)
	assert.Equal(t, 0, msg.Channels[0].Channel)
	assert.Equal(t, "±1V", msg.Channels[0].VRange)
	assert.Equal(t, "AC", msg.Channels[0].CoupleType)
	assert.True(t, msg.Channels[0].X10)
	assert.Equal(t, float32(0.25), msg.Channels[0].Offset)

	// Test JSON marshaling roundtrip
	data, err := json.Marshal(msg)
	require.NoError(t, err)

	var decoded MultiSyncMessage
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, msg.CommonTrigger, decoded.CommonTrigger)
	assert.Equal(t, msg.TimeDiv, decoded.TimeDiv)
	assert.Equal(t, msg.Channels[0].VRange, decoded.Channels[0].VRange)
}

func TestApplyCommonTrigger(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	w := app.NewWindow("Test")
	defer w.Close()

	setChan := make(chan *control.TriggerDescMsg, 10)
	psc := &control.PscDesc{
		SetTriggerCh: setChan,
	}

	scp := &ScpDesc{
		Window:       w,
		channelCount: 2,
		Settings:     settings.NewDefaultSettings(),
		psControl:    psc,
		triggerCheck: make([]*widget.Check, 2),
		IsMSO:        true,
	}
	scp.triggerCheck[0] = widget.NewCheck("Ch A", func(bool) {})
	scp.triggerCheck[1] = widget.NewCheck("Ch B", func(bool) {})

	// 1. Select Ch A
	scp.applyCommonTrigger("Ch A")
	assert.Equal(t, genericps.ChA, scp.triggerSource)
	assert.True(t, scp.Settings.Channels[0].TriggerSource)
	assert.False(t, scp.Settings.Channels[1].TriggerSource)
	assert.True(t, scp.triggerCheck[0].Checked)
	assert.False(t, scp.triggerCheck[1].Checked)
	assert.False(t, scp.Settings.Digital.Trigger.Enabled)

	// 2. Select Digital D3
	scp.applyCommonTrigger("D3")
	assert.Equal(t, dontCare, scp.triggerSource)
	assert.False(t, scp.Settings.Channels[0].TriggerSource)
	assert.False(t, scp.Settings.Channels[1].TriggerSource)
	assert.False(t, scp.triggerCheck[0].Checked)
	assert.False(t, scp.triggerCheck[1].Checked)
	assert.True(t, scp.Settings.Digital.Trigger.Enabled)
	assert.True(t, scp.Settings.Digital.Ports[0].Enabled)
	assert.True(t, scp.Settings.Digital.ChannelsEnabled[3])
	assert.Equal(t, genericps.DigitalDirectionRising, scp.Settings.Digital.Trigger.Directions[3])
	assert.Equal(t, genericps.DigitalDontCare, scp.Settings.Digital.Trigger.Directions[0])

	// 3. Select None
	scp.applyCommonTrigger("None")
	assert.Equal(t, dontCare, scp.triggerSource)
	assert.False(t, scp.Settings.Channels[0].TriggerSource)
	assert.False(t, scp.Settings.Channels[1].TriggerSource)
	assert.False(t, scp.Settings.Digital.Trigger.Enabled)
	for i := 0; i < 16; i++ {
		assert.Equal(t, genericps.DigitalDontCare, scp.Settings.Digital.Trigger.Directions[i])
	}
}

func TestApplyMultiSyncParams(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	w := app.NewWindow("Test")
	defer w.Close()

	scp := &ScpDesc{
		Window:               w,
		channelCount:         2,
		Settings:             settings.NewDefaultSettings(),
		channelViewers:       make([]channelViewerDesc, 2),
		triggerCheck:         make([]*widget.Check, 2),
		multiSyncStatusLabel: widget.NewLabel(""),
	}
	scp.triggerCheck[0] = widget.NewCheck("Ch A", func(bool) {})
	scp.triggerCheck[1] = widget.NewCheck("Ch B", func(bool) {})

	msg := &MultiSyncMessage{
		Type:              "SyncSettings",
		CommonTrigger:     "Ch B",
		TriggerMode:       "Auto",
		TimeDiv:           "100",
		TimeUnit:          "ms",
		TriggerTimeOffset: 0.05,
		Channels: []MultiChannelSyncParams{
			{
				Channel:           0,
				Enabled:           true,
				VRange:            "±2V",
				CoupleType:        "AC",
				X10:               false,
				Offset:            0.5,
				Inverted:          false,
				TriggerSource:     false,
				TriggerDirection:  "Rising",
				TriggerMv:         200,
				TriggerHysteresis: 10,
			},
			{
				Channel:           1,
				Enabled:           true,
				VRange:            "±500mV",
				CoupleType:        "DC",
				X10:               true,
				Offset:            -0.2,
				Inverted:          true,
				TriggerSource:     true,
				TriggerDirection:  "Falling",
				TriggerMv:         -100,
				TriggerHysteresis: 15,
			},
		},
	}

	scp.applyMultiSyncParams(msg)

	assert.Equal(t, "100", scp.Settings.Time.TimeDiv)
	assert.Equal(t, "ms", scp.Settings.Time.Unit)
	assert.Equal(t, 0.05, scp.Settings.Time.TriggerTimeOffset)
	assert.Equal(t, "Auto", scp.Settings.Trigger.Mode)
	assert.Equal(t, "Ch B", scp.Settings.Multiscope.CommonTrigger)

	// Channel 0 checks
	assert.Equal(t, genericps.Range_2v, scp.Settings.Channels[0].VRange)
	assert.Equal(t, genericps.Ac, scp.Settings.Channels[0].CoupleType)
	assert.False(t, scp.Settings.Channels[0].X10)
	assert.Equal(t, float32(0.5), scp.Settings.Channels[0].Offset)

	// Channel 1 checks
	assert.Equal(t, genericps.Range_50mv, scp.Settings.Channels[1].VRange) // ±500mV with x10 maps to 50mV base range
	assert.Equal(t, genericps.Dc, scp.Settings.Channels[1].CoupleType)
	assert.True(t, scp.Settings.Channels[1].X10)
	assert.Equal(t, float32(-0.2), scp.Settings.Channels[1].Offset)
	assert.True(t, scp.Settings.Channels[1].Inverted)
	assert.Equal(t, genericps.TriggerFalling, scp.Settings.Channels[1].Trigger.TriggerDirection)

	// Status label
	assert.Contains(t, scp.multiSyncStatusLabel.Text, "Synced: Trig=Ch B")
}

func TestMultiscopeTCPSyncRoundtrip(t *testing.T) {
	genericps.ChA = genericps.ChannelId(0)
	genericps.ChB = genericps.ChannelId(1)
	app := test.NewApp()
	defer app.Quit()
	w := app.NewWindow("Test")
	defer w.Close()

	// Find free port
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()

	// Server instance
	server := &ScpDesc{
		Window:                 w,
		channelCount:           2,
		Settings:               settings.NewDefaultSettings(),
		multiClientsLabel:      widget.NewLabel(""),
		multiServerStatusLabel: widget.NewLabel(""),
		psControl: &control.PscDesc{
			Info: "ServerScope",
		},
	}
	server.Settings.Multiscope.CommonTrigger = "Ch B"
	server.triggerSource = genericps.ChB
	server.Settings.Time.TimeDiv = "50"
	server.Settings.Time.Unit = "ns"
	server.startMultiServer(port)
	defer server.stopMultiServer()

	time.Sleep(50 * time.Millisecond)

	// Client instance
	client := &ScpDesc{
		Window:               w,
		channelCount:         2,
		Settings:             settings.NewDefaultSettings(),
		multiSyncStatusLabel: widget.NewLabel(""),
		channelViewers:       make([]channelViewerDesc, 2),
		triggerCheck:         make([]*widget.Check, 2),
		psControl: &control.PscDesc{
			Info: "ClientScope",
		},
	}
	client.triggerCheck[0] = widget.NewCheck("Ch A", func(bool) {})
	client.triggerCheck[1] = widget.NewCheck("Ch B", func(bool) {})

	client.connectMultiClient("127.0.0.1", port)
	defer client.disconnectMultiClient()

	// Wait for connection and handshake
	require.Eventually(t, func() bool {
		server.multiServerMu.Lock()
		defer server.multiServerMu.Unlock()
		return len(server.multiClients) == 1 && server.multiClients[0].ScopeType == "ClientScope"
	}, 2*time.Second, 50*time.Millisecond)

	// Verify server UI updated with client scope type
	assert.Contains(t, server.multiClientsLabel.Text, "ClientScope")

	// Server publishes sync settings
	server.publishMultiSync()

	// Verify client receives and applies settings
	require.Eventually(t, func() bool {
		return client.Settings.Multiscope.CommonTrigger == "Ch B" &&
			client.Settings.Time.TimeDiv == "50" &&
			client.Settings.Time.Unit == "ns"
	}, 2*time.Second, 50*time.Millisecond)

	assert.Contains(t, client.multiSyncStatusLabel.Text, "Synced: Trig=Ch B")
	assert.Contains(t, server.multiServerStatusLabel.Text, "Synced to 1 client(s)")
}
