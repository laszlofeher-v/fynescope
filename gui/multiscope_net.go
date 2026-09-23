package gui

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fynescope/genericps"
	"fynescope/settings"
)

type multiClientInfo struct {
	Conn      net.Conn
	ScopeType string
}

type MultiClientInfoMsg struct {
	Type      string `json:"type"` // "ClientInfo"
	ScopeType string `json:"scope_type"`
}

type MultiSyncMessage struct {
	Type              string                     `json:"type"` // "SyncSettings"
	CommonTrigger     string                     `json:"common_trigger"`
	TriggerMode       string                     `json:"trigger_mode"`
	TriggerDirection  string                     `json:"trigger_direction,omitempty"`
	TriggerMv         int32                      `json:"trigger_mv,omitempty"`
	TriggerHysteresis int32                      `json:"trigger_hysteresis,omitempty"`
	DigitalChannel    int                        `json:"digital_channel,omitempty"`
	DigitalDirection  genericps.DigitalDirection `json:"digital_direction,omitempty"`
	TimeDiv           string                     `json:"time_div"`
	TimeUnit          string                     `json:"time_unit"`
	TriggerTimeOffset float64                    `json:"trigger_time_offset"`
	SampleRate        string                     `json:"sample_rate,omitempty"`
	SampleRateUnit    string                     `json:"sample_rate_unit,omitempty"`
	Channels          []MultiChannelSyncParams   `json:"channels"`
}

type MultiChannelSyncParams struct {
	Channel           int     `json:"channel"`
	Enabled           bool    `json:"enabled"`
	VRange            string  `json:"v_range"`
	CoupleType        string  `json:"couple_type"`
	X10               bool    `json:"x10"`
	Offset            float32 `json:"offset"`
	Inverted          bool    `json:"inverted"`
	TriggerSource     bool    `json:"trigger_source"`
	TriggerDirection  string  `json:"trigger_direction"`
	TriggerMv         int32   `json:"trigger_mv"`
	TriggerHysteresis int32   `json:"trigger_hysteresis"`
}

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
			scp.multiClients = append(scp.multiClients, multiClientInfo{Conn: conn, ScopeType: "Unknown"})
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

	for _, client := range scp.multiClients {
		if client.Conn != nil {
			client.Conn.Close()
		}
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
		for i, client := range scp.multiClients {
			scopeType := client.ScopeType
			if scopeType == "" {
				scopeType = "Unknown"
			}
			remote := "Unknown"
			if client.Conn != nil {
				remote = client.Conn.RemoteAddr().String()
			}
			sb.WriteString(fmt.Sprintf("%d. %s (Type: %s)\n", i+1, remote, scopeType))
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
		for i, c := range scp.multiClients {
			if c.Conn == conn {
				scp.multiClients = append(scp.multiClients[:i], scp.multiClients[i+1:]...)
				break
			}
		}
		scp.updateConnectedClientsLabel()
		scp.multiServerMu.Unlock()
	}()

	slog.Info("Client connected", "addr", conn.RemoteAddr().String())

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		slog.Info("Received from client", "data", line)

		var base struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal([]byte(line), &base); err == nil {
			if base.Type == "ClientInfo" {
				var info MultiClientInfoMsg
				if err := json.Unmarshal([]byte(line), &info); err == nil {
					scp.multiServerMu.Lock()
					for idx := range scp.multiClients {
						if scp.multiClients[idx].Conn == conn {
							scp.multiClients[idx].ScopeType = info.ScopeType
							break
						}
					}
					scp.updateConnectedClientsLabel()
					scp.multiServerMu.Unlock()
				}
			}
		}
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
		if scp.multiSyncStatusLabel != nil {
			fyne.Do(func() {
				scp.multiSyncStatusLabel.SetText("Sync Status: Connection Failed")
			})
		}
		return
	}

	scp.multiClientConn = conn
	slog.Info("Connected to multiscope server", "addr", addr)
	if scp.multiSyncStatusLabel != nil {
		fyne.Do(func() {
			scp.multiSyncStatusLabel.SetText("Sync Status: Connected, Waiting for Settings")
		})
	}

	go func() {
		// Send handshake with local scope type
		scopeType := "Demo"
		if scp.psControl != nil && scp.psControl.Info != "" {
			scopeType = scp.psControl.Info
		}
		infoMsg := MultiClientInfoMsg{
			Type:      "ClientInfo",
			ScopeType: scopeType,
		}
		if b, err := json.Marshal(infoMsg); err == nil {
			b = append(b, '\n')
			conn.Write(b)
		}

		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			line := scanner.Text()
			slog.Info("Received from server", "data", line)

			var base struct {
				Type string `json:"type"`
			}
			if err := json.Unmarshal([]byte(line), &base); err == nil {
				if base.Type == "SyncSettings" {
					var syncMsg MultiSyncMessage
					if err := json.Unmarshal([]byte(line), &syncMsg); err == nil {
						fyne.Do(func() {
							scp.applyMultiSyncParams(&syncMsg)
						})
					}
				}
			}
		}

		// If scanner ends, connection was lost
		scp.multiServerMu.Lock()
		if scp.multiClientConn == conn {
			scp.multiClientConn = nil
		}
		scp.multiServerMu.Unlock()
		slog.Info("Disconnected from multiscope server")
		if scp.multiSyncStatusLabel != nil {
			fyne.Do(func() {
				scp.multiSyncStatusLabel.SetText("Sync Status: Disconnected")
			})
		}
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
	if scp.multiSyncStatusLabel != nil {
		fyne.Do(func() {
			scp.multiSyncStatusLabel.SetText("Sync Status: Disconnected")
		})
	}
}

func parseTriggerDirection(dirStr string) genericps.ThresholdDirection {
	switch strings.ToLower(strings.TrimSpace(dirStr)) {
	case "falling":
		return genericps.TriggerFalling
	case "rising":
		return genericps.TriggerRising
	case "enter":
		return genericps.TriggerEnter
	case "exit":
		return genericps.TriggerExit
	case "either":
		return genericps.TriggerEnterOrExit
	case "positive runt":
		return genericps.TriggerPositiveRunt
	case "negative runt":
		return genericps.TriggerNegativeRunt
	default:
		return genericps.TriggerRising
	}
}

func formatTriggerDirection(dir genericps.ThresholdDirection) string {
	switch dir {
	case genericps.TriggerFalling:
		return "Falling"
	case genericps.TriggerRising:
		return "Rising"
	case genericps.TriggerEnter:
		return "Enter"
	case genericps.TriggerExit:
		return "Exit"
	case genericps.TriggerEnterOrExit:
		return "Either"
	case genericps.TriggerPositiveRunt:
		return "Positive Runt"
	case genericps.TriggerNegativeRunt:
		return "Negative Runt"
	default:
		return "Rising"
	}
}

func (scp *ScpDesc) buildMultiSyncMessage() *MultiSyncMessage {
	if scp.Settings == nil {
		return &MultiSyncMessage{Type: "SyncSettings"}
	}

	commonTrig := scp.Settings.Multiscope.CommonTrigger
	if commonTrig == "" {
		if scp.multiCommonTrigSelect != nil && scp.multiCommonTrigSelect.Selected != "" {
			commonTrig = scp.multiCommonTrigSelect.Selected
		} else {
			commonTrig = "Ch A"
		}
	}

	timeDiv := scp.Settings.Time.TimeDiv
	if scp.timeSelect != nil && scp.timeSelect.Selected != "" {
		timeDiv = scp.timeSelect.Selected
	}
	if timeDiv == "" {
		timeDiv = "500"
	}

	timeUnit := scp.Settings.Time.Unit
	if scp.timeUnitSelect != nil && scp.timeUnitSelect.Selected != "" {
		timeUnit = scp.timeUnitSelect.Selected
	}
	if timeUnit == "" {
		timeUnit = "us"
	}

	msg := &MultiSyncMessage{
		Type:              "SyncSettings",
		CommonTrigger:     commonTrig,
		TriggerMode:       scp.Settings.Trigger.Mode,
		TimeDiv:           timeDiv,
		TimeUnit:          timeUnit,
		TriggerTimeOffset: scp.Settings.Time.TriggerTimeOffset,
		SampleRate:        scp.Settings.Dft.SampleRate,
		SampleRateUnit:    scp.Settings.Dft.SampleRateUnit,
	}

	// Single-channel trigger details
	if chIdx, err := parseChannelIdentifier(commonTrig); err == nil && chIdx >= 0 && chIdx < len(scp.Settings.Channels) {
		ch := scp.Settings.Channels[chIdx]
		msg.TriggerDirection = formatTriggerDirection(ch.Trigger.TriggerDirection)
		msg.TriggerMv = ch.Trigger.Mv
		msg.TriggerHysteresis = ch.Trigger.Hysteresis
	} else if strings.HasPrefix(strings.ToUpper(commonTrig), "D") {
		digIdx, err := strconv.Atoi(strings.TrimPrefix(strings.ToUpper(commonTrig), "D"))
		if err == nil && digIdx >= 0 && digIdx < 16 {
			msg.DigitalChannel = digIdx
			msg.DigitalDirection = scp.Settings.Digital.Trigger.Directions[digIdx]
		}
	}

	chCount := int(scp.channelCount)
	if chCount == 0 {
		chCount = len(scp.Settings.Channels)
	}
	for i := 0; i < len(scp.Settings.Channels) && i < chCount; i++ {
		ch := scp.Settings.Channels[i]
		vRangeStr := rangeEnumToString[ch.VRange]
		if vRangeStr == "" {
			vRangeStr = "±1V"
		}
		coupleStr := "DC"
		if ch.CoupleType == genericps.Ac {
			coupleStr = "AC"
		}
		trigDir := formatTriggerDirection(ch.Trigger.TriggerDirection)

		msg.Channels = append(msg.Channels, MultiChannelSyncParams{
			Channel:           i,
			Enabled:           ch.Enabled,
			VRange:            vRangeStr,
			CoupleType:        coupleStr,
			X10:               ch.X10,
			Offset:            ch.Offset,
			Inverted:          ch.Inverted,
			TriggerSource:     ch.TriggerSource,
			TriggerDirection:  trigDir,
			TriggerMv:         ch.Trigger.Mv,
			TriggerHysteresis: ch.Trigger.Hysteresis,
		})
	}

	return msg
}

func (scp *ScpDesc) applyCommonTrigger(triggerName string) {
	if scp.Settings == nil {
		return
	}
	slog.Info("Applying common trigger", "trigger", triggerName)

	if triggerName == "None" || triggerName == "" {
		// Disable analog trigger
		scp.triggerSource = dontCare
		for i := range scp.Settings.Channels {
			scp.Settings.Channels[i].TriggerSource = false
			if i < len(scp.triggerCheck) && scp.triggerCheck[i] != nil {
				scp.triggerCheck[i].Checked = false
				scp.triggerCheck[i].Refresh()
			}
		}
		if scp.triggerDisplays != nil {
			scp.triggerDisplays.Hide()
		}
		scp.setTrigger(false, genericps.ChA, 0, genericps.TriggerRising, 1000, scp.Settings.Time.TriggerTimeOffset)

		// Disable digital trigger
		scp.Settings.Digital.Trigger.Enabled = false
		for i := 0; i < 16; i++ {
			scp.Settings.Digital.Trigger.Directions[i] = genericps.DigitalDontCare
		}
		scp.updateDigitalTrigger()
		return
	}

	if strings.HasPrefix(strings.ToUpper(triggerName), "D") {
		digIdx, err := strconv.Atoi(strings.TrimPrefix(strings.ToUpper(triggerName), "D"))
		if err == nil && digIdx >= 0 && digIdx < 16 {
			// Disable analog trigger
			scp.triggerSource = dontCare
			for i := range scp.Settings.Channels {
				scp.Settings.Channels[i].TriggerSource = false
				if i < len(scp.triggerCheck) && scp.triggerCheck[i] != nil {
					scp.triggerCheck[i].Checked = false
					scp.triggerCheck[i].Refresh()
				}
			}
			if scp.triggerDisplays != nil {
				scp.triggerDisplays.Hide()
			}
			scp.setTrigger(false, genericps.ChA, 0, genericps.TriggerRising, 1000, scp.Settings.Time.TriggerTimeOffset)

			// Enable digital trigger
			if genericps.DigitalDirectionRising == 0 {
				genericps.DigitalDirectionRising = 3
			}
			scp.Settings.Digital.Trigger.Enabled = true
			portIdx := digIdx / 8
			scp.Settings.Digital.Ports[portIdx].Enabled = true
			scp.Settings.Digital.ChannelsEnabled[digIdx] = true
			if scp.Settings.Digital.Trigger.Directions[digIdx] == genericps.DigitalDontCare {
				scp.Settings.Digital.Trigger.Directions[digIdx] = genericps.DigitalDirectionRising
			}
			for j := 0; j < 16; j++ {
				if j != digIdx {
					scp.Settings.Digital.Trigger.Directions[j] = genericps.DigitalDontCare
				}
			}
			scp.updateDigitalTrigger()
			return
		}
	}

	chIdx, err := parseChannelIdentifier(triggerName)
	if err == nil && chIdx >= 0 && chIdx < len(scp.Settings.Channels) {
		// Disable digital trigger
		scp.Settings.Digital.Trigger.Enabled = false
		for i := 0; i < 16; i++ {
			scp.Settings.Digital.Trigger.Directions[i] = genericps.DigitalDontCare
		}
		scp.updateDigitalTrigger()

		// Enable analog trigger
		scp.triggerSource = genericps.ChannelId(chIdx)
		for i := range scp.Settings.Channels {
			isSrc := (i == chIdx)
			scp.Settings.Channels[i].TriggerSource = isSrc
			if i < len(scp.triggerCheck) && scp.triggerCheck[i] != nil {
				scp.triggerCheck[i].Checked = isSrc
				scp.triggerCheck[i].Refresh()
			}
		}
		if scp.triggerDisplays != nil {
			scp.triggerDisplays.Show()
		}
		ch := &scp.Settings.Channels[chIdx]
		scp.setTrigger(true, genericps.ChannelId(chIdx), ch.Trigger.Mv, ch.Trigger.TriggerDirection, 1000, scp.Settings.Time.TriggerTimeOffset)
	}
}

func (scp *ScpDesc) applyMultiSyncParams(msg *MultiSyncMessage) {
	if msg == nil || scp.Settings == nil {
		return
	}

	// 1. Timebase
	if msg.TimeUnit != "" {
		if scp.timeUnitSelect != nil {
			scp.timeUnitSelect.SetSelected(msg.TimeUnit)
		} else {
			scp.Settings.Time.Unit = msg.TimeUnit
			scp.timeUnit = tu[msg.TimeUnit]
		}
	}
	if msg.TimeDiv != "" {
		if scp.timeSelect != nil {
			scp.timeSelect.SetSelected(msg.TimeDiv)
		} else {
			scp.Settings.Time.TimeDiv = msg.TimeDiv
			intDiv, _ := strconv.Atoi(msg.TimeDiv)
			scp.timeDiv = intDiv
		}
	}
	if msg.TriggerTimeOffset != 0 || scp.Settings.Time.TriggerTimeOffset != 0 {
		scp.Settings.Time.TriggerTimeOffset = msg.TriggerTimeOffset
		scp.setTriggerTime(scp.Settings.Time.TriggerTimeOffset)
	}

	// 2. Channels
	chCount := int(scp.channelCount)
	if chCount == 0 {
		chCount = len(scp.Settings.Channels)
	}
	for _, chParam := range msg.Channels {
		chIdx := chParam.Channel
		if chIdx < 0 || chIdx >= len(scp.Settings.Channels) || chIdx >= chCount {
			continue
		}
		chId := genericps.ChannelId(chIdx)

		// Range
		if chParam.VRange != "" {
			scp.changeChannelRange(chId, chParam.VRange)
		}
		// X10
		scp.changeChannelX10(chId, chParam.X10)

		// Coupling
		cType := genericps.Ac
		if chParam.CoupleType == "DC" {
			cType = genericps.Dc
		}
		scp.Settings.Channels[chIdx].CoupleType = cType
		if chIdx < len(scp.channelViewers) && scp.channelViewers[chIdx].acdcSelect != nil {
			scp.channelViewers[chIdx].acdcSelect.SilentSetSelected(chParam.CoupleType)
			scp.channelViewers[chIdx].acdcSelect.Refresh()
		}

		// Offset
		scp.Settings.Channels[chIdx].Offset = chParam.Offset
		if chIdx < len(scp.channelViewers) && scp.channelViewers[chIdx].offset != nil {
			scp.channelViewers[chIdx].offset.SilentSetFloatValue(float64(chParam.Offset), 3)
			scp.channelViewers[chIdx].offset.Refresh()
		}

		// Inverted
		scp.Settings.Channels[chIdx].Inverted = chParam.Inverted
		if chIdx < len(scp.channelViewers) && scp.channelViewers[chIdx].invertCheckbox != nil {
			scp.channelViewers[chIdx].invertCheckbox.Checked = chParam.Inverted
			scp.channelViewers[chIdx].invertCheckbox.Refresh()
		}

		// Channel trigger parameters
		scp.Settings.Channels[chIdx].Trigger.Mv = chParam.TriggerMv
		scp.Settings.Channels[chIdx].Trigger.Hysteresis = chParam.TriggerHysteresis
		if chParam.TriggerDirection != "" {
			dir := parseTriggerDirection(chParam.TriggerDirection)
			scp.Settings.Channels[chIdx].Trigger.TriggerDirection = dir
			if chIdx < len(scp.channelViewers) && scp.channelViewers[chIdx].triggerDirectionSelect != nil {
				scp.channelViewers[chIdx].triggerDirectionSelect.SilentSetSelected(chParam.TriggerDirection)
				scp.channelViewers[chIdx].triggerDirectionSelect.Refresh()
			}
		}

		// Enabled
		scp.EnableChannel(chId, chParam.Enabled)

		// Update device
		channelCopy := scp.Settings.Channels[chIdx]
		channelCopy.ID = chId
		if scp.psControl != nil && scp.psControl.SetChannelCh != nil {
			go func(c settings.ChSettings) {
				scp.psControl.SetChannelCh <- &c
			}(channelCopy)
		}
	}

	// 3. Trigger Mode
	if msg.TriggerMode != "" {
		scp.Settings.Trigger.Mode = msg.TriggerMode
		if scp.triggerModeSelect != nil {
			scp.triggerModeSelect.SetSelected(msg.TriggerMode)
		}
	}

	// 4. Digital Channel trigger direction if specified
	if msg.DigitalDirection != 0 && msg.DigitalChannel >= 0 && msg.DigitalChannel < 16 {
		scp.Settings.Digital.Trigger.Directions[msg.DigitalChannel] = msg.DigitalDirection
	}

	// 5. Common Trigger
	if msg.CommonTrigger != "" {
		scp.Settings.Multiscope.CommonTrigger = msg.CommonTrigger
		if scp.multiCommonTrigSelect != nil {
			scp.multiCommonTrigSelect.SetSelected(msg.CommonTrigger)
		}
		scp.applyCommonTrigger(msg.CommonTrigger)
	}

	// 6. Update Status Label
	if scp.multiSyncStatusLabel != nil {
		scp.multiSyncStatusLabel.SetText(fmt.Sprintf("Synced: Trig=%s, T/div=%s %s (%s)",
			msg.CommonTrigger, msg.TimeDiv, msg.TimeUnit, time.Now().Format("15:04:05")))
	}

	scp.clearAllFtPersistentLayers()
	scp.clearAllDftPersistentLayers()
	scp.refreshRasters()
	scp.SaveSettings()
}

func (scp *ScpDesc) publishMultiSync() {
	if scp.multiCommonTrigSelect != nil && scp.multiCommonTrigSelect.Selected != "" {
		scp.Settings.Multiscope.CommonTrigger = scp.multiCommonTrigSelect.Selected
		scp.applyCommonTrigger(scp.multiCommonTrigSelect.Selected)
		scp.SaveSettings()
	}

	msg := scp.buildMultiSyncMessage()
	data, err := json.Marshal(msg)
	if err != nil {
		slog.Error("Failed to marshal multiscope sync message", "err", err)
		return
	}
	data = append(data, '\n')

	scp.multiServerMu.Lock()
	defer scp.multiServerMu.Unlock()

	count := 0
	for _, client := range scp.multiClients {
		if client.Conn != nil {
			_, err := client.Conn.Write(data)
			if err == nil {
				count++
			} else {
				slog.Error("Failed to send sync message to client", "err", err)
			}
		}
	}
	slog.Info("Published multiscope sync settings", "clientsReached", count, "totalClients", len(scp.multiClients))

	if scp.multiServerStatusLabel != nil {
		fyne.Do(func() {
			scp.multiServerStatusLabel.SetText(fmt.Sprintf("Synced to %d client(s) at %s", count, time.Now().Format("15:04:05")))
		})
	}
}
