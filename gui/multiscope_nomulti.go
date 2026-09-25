//go:build !multi

package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

type multiClientInfo struct{}

// SetMultiscopeMode is a no-op when multiscope is not compiled in.
func (scp *ScpDesc) SetMultiscopeMode(mode string) {}

// IsMultiServerRunning always returns false when multiscope is not compiled in.
func (scp *ScpDesc) IsMultiServerRunning() bool { return false }

// IsMultiClientConnected always returns false when multiscope is not compiled in.
func (scp *ScpDesc) IsMultiClientConnected() bool { return false }

// StartMultiServer is a no-op when multiscope is not compiled in.
func (scp *ScpDesc) StartMultiServer(port int) {}

// StopMultiServer is a no-op when multiscope is not compiled in.
func (scp *ScpDesc) StopMultiServer() {}

// ConnectMultiClient is a no-op when multiscope is not compiled in.
func (scp *ScpDesc) ConnectMultiClient(ip string, port int) {}

// DisconnectMultiClient is a no-op when multiscope is not compiled in.
func (scp *ScpDesc) DisconnectMultiClient() {}

// newMultiscopePanel returns an empty container when multiscope is not compiled in.
func (scp *ScpDesc) newMultiscopePanel(undockable bool) *fyne.Container {
	return container.NewMax()
}

// broadcastMultiStart is a no-op when multiscope is not compiled in.
func (scp *ScpDesc) broadcastMultiStart() {}

// broadcastMultiStop is a no-op when multiscope is not compiled in.
func (scp *ScpDesc) broadcastMultiStop() {}

// sendClientWaveforms is a no-op when multiscope is not compiled in.
func (scp *ScpDesc) sendClientWaveforms() {}

// openRemoteChannelsWindow is a no-op when multiscope is not compiled in.
func (scp *ScpDesc) openRemoteChannelsWindow() {}
