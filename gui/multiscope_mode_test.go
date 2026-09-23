package gui

import (
	"image/color"
	"net"
	"testing"
	"time"

	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"fynescope/settings"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMultiscopeSwitchServerToClientStopsServer(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	w := app.NewWindow("Test")
	defer w.Close()

	// Find free port
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()

	indicator := canvas.NewCircle(color.NRGBA{R: 200, G: 0, B: 0, A: 255})

	scp := &ScpDesc{
		App:                        app,
		Window:                     w,
		channelCount:               2,
		Settings:                   settings.NewDefaultSettings(),
		multiClientsLabel:          widget.NewLabel(""),
		multiServerStatusIndicator: indicator,
	}
	scp.Settings.Multiscope.Mode = "Server"

	// 1. Start server
	scp.StartMultiServer(port)
	require.True(t, scp.IsMultiServerRunning(), "multiscope server should be running")
	assert.Equal(t, color.NRGBA{R: 0, G: 200, B: 0, A: 255}, indicator.FillColor, "status indicator should be green")

	// 2. Change server to client mode: server must stop first, then change to client
	scp.SetMultiscopeMode("Client")

	// 3. Verify server is stopped and mode is changed to Client
	assert.False(t, scp.IsMultiServerRunning(), "multiscope server should be stopped after changing to Client")
	assert.Equal(t, "Client", scp.Settings.Multiscope.Mode, "mode should be Client")
	assert.Equal(t, color.NRGBA{R: 200, G: 0, B: 0, A: 255}, indicator.FillColor, "status indicator should be red")

	// 4. Verify that port is free to bind (server actually released listener)
	verifyListener, err := net.Listen("tcp", l.Addr().String())
	assert.NoError(t, err, "server socket should be released")
	if verifyListener != nil {
		verifyListener.Close()
	}
}

func TestMultiscopeSwitchClientToServerDisconnectsClient(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	w := app.NewWindow("Test")
	defer w.Close()

	// Start a dummy TCP server to connect client to
	dummyListener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer dummyListener.Close()
	port := dummyListener.Addr().(*net.TCPAddr).Port

	scp := &ScpDesc{
		App:                  app,
		Window:               w,
		channelCount:         2,
		Settings:             settings.NewDefaultSettings(),
		multiSyncStatusLabel: widget.NewLabel(""),
	}
	scp.Settings.Multiscope.Mode = "Client"

	// Connect client
	scp.ConnectMultiClient("127.0.0.1", port)
	require.True(t, scp.IsMultiClientConnected(), "client should be connected")

	// Change client to server mode: client must disconnect first, then change to server
	scp.SetMultiscopeMode("Server")

	assert.False(t, scp.IsMultiClientConnected(), "client should be disconnected after changing to Server")
	assert.Equal(t, "Server", scp.Settings.Multiscope.Mode, "mode should be Server")
	assert.Equal(t, "Sync Status: Disconnected", scp.multiSyncStatusLabel.Text)
}

func TestMultiscopeConnectClientStopsRunningServer(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	w := app.NewWindow("Test")
	defer w.Close()

	// Find free port for server
	l1, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	serverPort := l1.Addr().(*net.TCPAddr).Port
	l1.Close()

	// Start remote dummy server for client target
	remoteListener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer remoteListener.Close()
	remotePort := remoteListener.Addr().(*net.TCPAddr).Port

	scp := &ScpDesc{
		App:                  app,
		Window:               w,
		channelCount:         2,
		Settings:             settings.NewDefaultSettings(),
		multiClientsLabel:    widget.NewLabel(""),
		multiSyncStatusLabel: widget.NewLabel(""),
	}

	// Start server on this instance
	scp.StartMultiServer(serverPort)
	require.True(t, scp.IsMultiServerRunning(), "server should be running")

	// Directly connect client to remote server - must stop running server first
	scp.ConnectMultiClient("127.0.0.1", remotePort)
	defer scp.DisconnectMultiClient()

	assert.False(t, scp.IsMultiServerRunning(), "server should be stopped when client connects")
	assert.True(t, scp.IsMultiClientConnected(), "client should be connected")
}

func TestMultiscopeStartServerDisconnectsRunningClient(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	w := app.NewWindow("Test")
	defer w.Close()

	// Start remote dummy server for client target
	remoteListener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer remoteListener.Close()
	remotePort := remoteListener.Addr().(*net.TCPAddr).Port

	// Find free port for server
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	serverPort := l.Addr().(*net.TCPAddr).Port
	l.Close()

	scp := &ScpDesc{
		App:                  app,
		Window:               w,
		channelCount:         2,
		Settings:             settings.NewDefaultSettings(),
		multiClientsLabel:    widget.NewLabel(""),
		multiSyncStatusLabel: widget.NewLabel(""),
	}

	// Connect client
	scp.ConnectMultiClient("127.0.0.1", remotePort)
	require.True(t, scp.IsMultiClientConnected(), "client should be connected")

	// Directly start server - must disconnect running client first
	scp.StartMultiServer(serverPort)
	defer scp.StopMultiServer()

	assert.False(t, scp.IsMultiClientConnected(), "client should be disconnected when server starts")
	assert.True(t, scp.IsMultiServerRunning(), "server should be running")
}

func TestMultiscopePanelSelectModeStopsServer(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	w := app.NewWindow("Test")
	defer w.Close()

	// Find free port
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()

	scp := &ScpDesc{
		App:          app,
		Window:       w,
		channelCount: 2,
		Settings:     settings.NewDefaultSettings(),
	}
	scp.Settings.Multiscope.Mode = "Server"
	scp.Settings.Multiscope.ServerPort = port

	panel := scp.newMultiscopePanel(false)
	require.NotNil(t, panel)

	// Start server
	scp.StartMultiServer(port)
	require.True(t, scp.IsMultiServerRunning())
	assert.Equal(t, color.NRGBA{R: 0, G: 200, B: 0, A: 255}, scp.multiServerStatusIndicator.FillColor)

	// Simulate selecting Client mode
	scp.SetMultiscopeMode("Client")
	time.Sleep(20 * time.Millisecond)

	assert.False(t, scp.IsMultiServerRunning(), "multiscope server should be stopped")
	assert.Equal(t, color.NRGBA{R: 200, G: 0, B: 0, A: 255}, scp.multiServerStatusIndicator.FillColor)
	assert.Equal(t, "Client", scp.Settings.Multiscope.Mode)
}
