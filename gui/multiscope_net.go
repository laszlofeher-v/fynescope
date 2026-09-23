package gui

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"strings"

	"fyne.io/fyne/v2"
)

func (scp *ScpDesc) startMultiServer(port int) {
	scp.multiServerMu.Lock()
	defer scp.multiServerMu.Unlock()

	if scp.multiServer != nil {
		slog.Warn("Multiscope server is already running")
		return
	}

	addr := fmt.Sprintf("0.0.0.0:%d", port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		slog.Error("Failed to start multiscope server", "err", err)
		return
	}
	scp.multiServer = l
	slog.Info("Multiscope server started", "addr", addr)

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				// Server was probably closed
				return
			}
			scp.multiServerMu.Lock()
			scp.multiClients = append(scp.multiClients, conn)
			scp.updateConnectedClientsLabel()
			scp.multiServerMu.Unlock()

			go scp.handleMultiClient(conn)
		}
	}()
}

func (scp *ScpDesc) stopMultiServer() {
	scp.multiServerMu.Lock()
	defer scp.multiServerMu.Unlock()

	if scp.multiServer != nil {
		scp.multiServer.Close()
		scp.multiServer = nil
	}

	for _, conn := range scp.multiClients {
		conn.Close()
	}
	scp.multiClients = nil
	scp.updateConnectedClientsLabel()
	slog.Info("Multiscope server stopped")
}

func (scp *ScpDesc) updateConnectedClientsLabel() {
	if scp.multiClientsLabel == nil {
		return
	}

	var sb strings.Builder
	sb.WriteString("Connected Clients:\n")
	if len(scp.multiClients) == 0 {
		sb.WriteString("None")
	} else {
		for i, conn := range scp.multiClients {
			sb.WriteString(fmt.Sprintf("%d. %s (Type: Unknown)\n", i+1, conn.RemoteAddr().String()))
		}
	}

	fyne.Do(func() {
		scp.multiClientsLabel.SetText(sb.String())
	})
}

func (scp *ScpDesc) handleMultiClient(conn net.Conn) {
	defer func() {
		conn.Close()
		scp.multiServerMu.Lock()
		// Remove from slice
		for i, c := range scp.multiClients {
			if c == conn {
				scp.multiClients = append(scp.multiClients[:i], scp.multiClients[i+1:]...)
				break
			}
		}
		scp.updateConnectedClientsLabel()
		scp.multiServerMu.Unlock()
	}()

	slog.Info("Client connected", "addr", conn.RemoteAddr().String())

	// Basic read loop for now
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		slog.Info("Received from client", "data", line)
	}
}

func (scp *ScpDesc) connectMultiClient(ip string, port int) {
	scp.multiServerMu.Lock()
	defer scp.multiServerMu.Unlock()

	if scp.multiClientConn != nil {
		slog.Warn("Multiscope client is already connected")
		return
	}

	addr := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		slog.Error("Failed to connect to multiscope server", "err", err)
		return
	}
	
	scp.multiClientConn = conn
	slog.Info("Connected to multiscope server", "addr", addr)

	go func() {
		// Basic read loop for now
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			line := scanner.Text()
			slog.Info("Received from server", "data", line)
		}
		
		// If scanner ends, connection was lost
		scp.multiServerMu.Lock()
		if scp.multiClientConn == conn {
			scp.multiClientConn = nil
		}
		scp.multiServerMu.Unlock()
		slog.Info("Disconnected from multiscope server")
	}()
}

func (scp *ScpDesc) disconnectMultiClient() {
	scp.multiServerMu.Lock()
	defer scp.multiServerMu.Unlock()

	if scp.multiClientConn != nil {
		scp.multiClientConn.Close()
		scp.multiClientConn = nil
		slog.Info("Multiscope client disconnected")
	}
}
