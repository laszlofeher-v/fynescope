//go:build multi

package gui

import (
	"fmt"
	"fynescope/disp7"
	"fynescope/settings"
	"net"

	"fynescope/tastybutton"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func getLocalIPs() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "Error fetching IPs"
	}

	var ips string
	for _, i := range interfaces {
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			if ips != "" {
				ips += ", "
			}
			ips += ip.String()
		}
	}
	if ips == "" {
		return "No IP found"
	}
	return ips
}

// SetMultiscopeMode updates the multiscope mode ("Server" or "Client").
// If switching to "Client" while the server is running, the server is stopped first.
// If switching to "Server" while the client is connected, the client is disconnected first.
func (scp *ScpDesc) SetMultiscopeMode(mode string) {
	if mode == "Client" {
		if scp.IsMultiServerRunning() {
			scp.StopMultiServer()
		}
	} else if mode == "Server" {
		if scp.IsMultiClientConnected() {
			scp.DisconnectMultiClient()
		}
	}
	if scp.Settings != nil {
		scp.Settings.Multiscope.Mode = mode
		scp.SaveSettings()
	}
}

func (scp *ScpDesc) newMultiscopePanel(undockable bool) *fyne.Container {
	if scp.Settings == nil {
		scp.Settings = settings.NewDefaultSettings()
	}
	ipStr := getLocalIPs()
	ipLabel := widget.NewLabel(fmt.Sprintf("Host IP(s): %s", ipStr))

	var serverBtn, clientBtn *tastybutton.TastyButton
	contentContainer := container.NewMax()

	// Server View Components
	serverPortDisp, _ := disp7.NewDisp7Array(5, 0, 65535, 0, disp7.UnSigned, scp.Window, theme.PrimaryColorNamed("green"), disp7.ReadWrite)
	if scp.Settings.Multiscope.ServerPort == 0 {
		scp.Settings.Multiscope.ServerPort = 50000
	}
	serverPortDisp.SilentSetValue(scp.Settings.Multiscope.ServerPort)
	serverPortDisp.OnChanged = func(v float64) {
		scp.Settings.Multiscope.ServerPort = int(v)
		scp.SaveSettings()
	}

	scp.multiClientsLabel = widget.NewLabel("Connected Clients:\nNone")
	
	statusColor := color.NRGBA{R: 200, G: 0, B: 0, A: 255}
	if scp.IsMultiServerRunning() {
		statusColor = color.NRGBA{R: 0, G: 200, B: 0, A: 255}
	}
	statusIndicator := canvas.NewCircle(statusColor)
	scp.multiServerStatusIndicator = statusIndicator
	statusIndicatorContainer := container.NewGridWrap(fyne.NewSize(16, 16), statusIndicator)

	startServerBtn := tastybutton.NewTastyButton("Start Server", tastybutton.Green, func() {
		scp.startMultiServer(serverPortDisp.Value)
	})
	stopServerBtn := tastybutton.NewTastyButton("Stop Server", tastybutton.Red, func() {
		scp.stopMultiServer()
	})
	
	
	syncSettingsBtn := tastybutton.NewTastyButton("Sync Settings", tastybutton.Green, func() {
		err := scp.publishMultiSync()
		if err != nil {
			dialog.ShowError(err, scp.Window)
		}
	})

	scp.multiServerStatusLabel = widget.NewLabel("")
	
	remoteChannelsBtn := tastybutton.NewTastyButton("Remote Channels...", tastybutton.Orange, func() {
		scp.openRemoteChannelsWindow()
	})

	serverView := container.NewVBox(
		widget.NewLabel("Server Port:"),
		serverPortDisp,
		container.NewHBox(startServerBtn, stopServerBtn, widget.NewLabel("Status:"), statusIndicatorContainer),
		syncSettingsBtn,
		remoteChannelsBtn,
		scp.multiServerStatusLabel,
		scp.multiClientsLabel,
	)

	// Client View Components
	clientIPEntry := widget.NewEntry()
	clientIPEntry.SetPlaceHolder("Enter Server IP")
	clientIPEntry.SetText(scp.Settings.Multiscope.ClientIP)
	clientIPEntry.OnChanged = func(s string) {
		scp.Settings.Multiscope.ClientIP = s
		scp.SaveSettings()
	}
	clientPortDisp, _ := disp7.NewDisp7Array(5, 0, 65535, 0, disp7.UnSigned, scp.Window, theme.PrimaryColorNamed("green"), disp7.ReadWrite)
	if scp.Settings.Multiscope.ClientPort == 0 {
		scp.Settings.Multiscope.ClientPort = 50000
	}
	clientPortDisp.SilentSetValue(scp.Settings.Multiscope.ClientPort)
	clientPortDisp.OnChanged = func(v float64) {
		scp.Settings.Multiscope.ClientPort = int(v)
		scp.SaveSettings()
	}

	connectClientBtn := tastybutton.NewTastyButton("Connect", tastybutton.Green, func() {
		scp.connectMultiClient(clientIPEntry.Text, clientPortDisp.Value)
	})
	disconnectClientBtn := tastybutton.NewTastyButton("Disconnect", tastybutton.Red, func() {
		scp.disconnectMultiClient()
	})

	scp.multiSyncStatusLabel = widget.NewLabel("Sync Status: Idle")

	clientView := container.NewVBox(
		widget.NewLabel("Server IP:"),
		clientIPEntry,
		widget.NewLabel("Server Port:"),
		clientPortDisp,
		container.NewHBox(connectClientBtn, disconnectClientBtn),
		widget.NewLabel("Multiscope Sync:"),
		scp.multiSyncStatusLabel,
	)

	selectMode := func(mode string) {
		scp.SetMultiscopeMode(mode)

		contentContainer.Objects = nil
		if mode == "Server" {
			serverBtn.Style = tastybutton.Green
			clientBtn.Style = tastybutton.Orange
			contentContainer.Add(serverView)
		} else {
			serverBtn.Style = tastybutton.Orange
			clientBtn.Style = tastybutton.Green
			contentContainer.Add(clientView)
		}
		serverBtn.Refresh()
		clientBtn.Refresh()
		contentContainer.Refresh()
	}

	serverBtn = tastybutton.NewTastyButton("Server", tastybutton.Green, func() { selectMode("Server") })
	clientBtn = tastybutton.NewTastyButton("Client", tastybutton.Orange, func() { selectMode("Client") })

	modeSelection := container.NewHBox(serverBtn, clientBtn)
	
	if scp.Settings.Multiscope.Mode == "" {
		scp.Settings.Multiscope.Mode = "Server"
	}
	selectMode(scp.Settings.Multiscope.Mode)

	var topRow fyne.CanvasObject = ipLabel

	if undockable {
		undockBtn := widget.NewButtonWithIcon("Undock", theme.ViewFullScreenIcon(), func() {
			if scp.multiWindow != nil {
				scp.multiWindow.RequestFocus()
				return
			}
			onWindowClose := func() {
				scp.multiWindow = nil
				scp.multiLayout = scp.newMultiscopePanel(true)
				scp.multiTab.Content = scp.multiLayout
				scp.dockTab(scp.multiTab)
				scp.controlTab.SelectIndex(ftTabIndex)
				fyne.Do(scp.multiTab.Content.Refresh)
			}
			scp.multiWindow = scp.App.NewWindow("Multiscope")
			winContent := scp.newMultiscopePanel(false)
			scp.controlTab.Remove(scp.multiTab)
			scp.multiWindow.SetContent(winContent)
			scp.multiWindow.SetOnClosed(onWindowClose)
			scp.multiWindow.Resize(fyne.NewSize(350, 700))
			scp.controlTab.SelectIndex(ftTabIndex)
			scp.multiWindow.Show()
			fyne.Do(winContent.Refresh)
		})
		addToTest(undockBtn, "multiUndockBtn", multiTabIndex)
		topRow = container.NewVBox(undockBtn, ipLabel)
	}

	return container.NewVBox(
		topRow,
		modeSelection,
		contentContainer,
	)
}
