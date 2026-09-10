package gui

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/csv"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"image/jpeg"
	"image/png"
	"io"
	"log/slog"
	"math"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fynescope/genericps"
	"fynescope/settings"
)

// APIServerConfig contains options for configuring the HTTPS API server.
type APIServerConfig struct {
	Port      int    // TCP port to listen on (e.g. 8443)
	AuthAdmin string // "user:pass" credentials for full read-write admin access
	AuthView  string // "user:pass" credentials for read-only access
	CertFile  string // Optional path to custom TLS certificate
	KeyFile   string // Optional path to custom TLS private key
}

// APIServer manages the lifecycle and active state of the HTTPS remote API server.
type APIServer struct {
	server *http.Server
	addr   string
	mu     sync.Mutex
	active bool
}

// ScopeInfo provides high-level metadata about the oscilloscope state and hardware.
type ScopeInfo struct {
	Status        string `json:"status"`
	Running       bool   `json:"running"`
	ChannelCount  int    `json:"channel_count"`
	IsMSO         bool   `json:"is_mso"`
	MaxValue      int32  `json:"max_value"`
	MinValue      int32  `json:"min_value"`
	TimeDiv       int    `json:"time_div"`
	TimeUnit      int    `json:"time_unit"`
	TimeUnitStr   string `json:"time_unit_str"`
	ExtGenEnabled bool   `json:"ext_gen_enabled"`
}

// ChannelInfo describes the current runtime settings and properties of an oscilloscope channel.
type ChannelInfo struct {
	ID             string  `json:"id"`
	Index          int     `json:"index"`
	Enabled        bool    `json:"enabled"`
	Range          string  `json:"range"`
	RangeMV        int32   `json:"range_mv"`
	Coupling       string  `json:"coupling"`
	Invert         bool    `json:"invert"`
	DisplayVOffset int     `json:"display_v_offset"`
	TriggerSource  bool    `json:"trigger_source"`
}

// TriggerInfo describes the current trigger configuration.
type TriggerInfo struct {
	Source      string  `json:"source"`
	Mode        string  `json:"mode"`
	Direction   string  `json:"direction"`
	Type        string  `json:"type"`
	ThresholdMV float32 `json:"threshold_mv"`
	Hysteresis  int32   `json:"hysteresis"`
}

// TimebaseInfo describes horizontal time per division and sampling parameters.
type TimebaseInfo struct {
	TimeDiv        int     `json:"time_div"`
	TimeUnit       int     `json:"time_unit"`
	TimeUnitStr    string  `json:"time_unit_str"`
	SampleInterval float64 `json:"sample_interval_s"`
}

// WaveformData contains the sampled signal points for a channel.
type WaveformData struct {
	Channel        string    `json:"channel"`
	Unit           string    `json:"unit"`
	SampleCount    int       `json:"sample_count"`
	SampleInterval float64   `json:"sample_interval_s"`
	Data           []float32 `json:"data"`
}

// ChannelMeasurements holds calculated waveform parameters.
type ChannelMeasurements struct {
	Channel     string  `json:"channel"`
	VMin        float32 `json:"v_min"`
	VMax        float32 `json:"v_max"`
	VPp         float32 `json:"v_pp"`
	VMean       float32 `json:"v_mean"`
	VRms        float32 `json:"v_rms"`
	FrequencyHz float64 `json:"frequency_hz"`
	PeriodS     float64 `json:"period_s"`
}

// parseAuthCredentials splits a "user:pass" credential string into its components.
func parseAuthCredentials(auth string) (user, pass string) {
	if auth == "" {
		return "", ""
	}
	parts := strings.SplitN(auth, ":", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return parts[0], ""
}

// generateAPICert creates an in-memory self-signed ECDSA P-256 certificate for HTTPS.
func generateAPICert() (tls.Certificate, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Fynescope Remote API"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, err
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	b, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return tls.Certificate{}, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: b})
	return tls.X509KeyPair(certPEM, keyPEM)
}

// timeUnitToString converts numeric timeUnit to a readable unit string.
func timeUnitToString(u int) string {
	switch u {
	case -12:
		return "ps"
	case -9:
		return "ns"
	case -6:
		return "us"
	case -3:
		return "ms"
	case 0:
		return "s"
	default:
		return fmt.Sprintf("10^%d s", u)
	}
}

// parseChannelIdentifier parses channel strings like "A", "B", "0", "1", "ChA" to a channel index.
func parseChannelIdentifier(s string) (int, error) {
	clean := strings.ToUpper(strings.ReplaceAll(s, " ", ""))
	clean = strings.TrimPrefix(clean, "CH")
	if len(clean) == 1 && clean[0] >= 'A' && clean[0] <= 'D' {
		return int(clean[0] - 'A'), nil
	}
	idx, err := strconv.Atoi(clean)
	if err == nil && idx >= 0 && idx < 4 {
		return idx, nil
	}
	return -1, fmt.Errorf("invalid channel identifier: %q", s)
}

// GetStatus returns the current run status of the oscilloscope.
func (scp *ScpDesc) GetStatus() string {
	if scp.running {
		return "Running"
	}
	return "Stopped"
}

// GetInfo returns high-level metadata about the oscilloscope.
func (scp *ScpDesc) GetInfo() ScopeInfo {
	return ScopeInfo{
		Status:        scp.GetStatus(),
		Running:       scp.running,
		ChannelCount:  int(scp.channelCount),
		IsMSO:         scp.IsMSO,
		MaxValue:      scp.MaxValue,
		MinValue:      scp.MinValue,
		TimeDiv:       scp.timeDiv,
		TimeUnit:      scp.timeUnit,
		TimeUnitStr:   timeUnitToString(scp.timeUnit),
		ExtGenEnabled: scp.ExtGenEnabled,
	}
}

// GetSettings returns a pointer to the current application settings.
func (scp *ScpDesc) GetSettings() *settings.PsSettings {
	return scp.Settings
}

// ApplySettingsJSON unmarshals a JSON payload over the existing settings struct (deep merge)
// and saves those settings to disk.
func (scp *ScpDesc) ApplySettingsJSON(data []byte) error {
	if err := json.Unmarshal(data, scp.Settings); err != nil {
		return err
	}

	scp.SaveSettings()
	return nil
}

// RunScope starts the oscilloscope capture.
func (scp *ScpDesc) RunScope() error {
	scp.ExecuteVoiceCommand("run")
	return nil
}

// StopScope stops the oscilloscope capture.
func (scp *ScpDesc) StopScope() error {
	scp.ExecuteVoiceCommand("stop")
	return nil
}

// SingleScope arms the oscilloscope for a single acquisition.
func (scp *ScpDesc) SingleScope() error {
	if scp.triggerModeSelect != nil {
		fyne.Do(func() {
			scp.triggerModeSelect.SetSelected(settings.TriggerModeSingle)
		})
	} else if scp.Settings != nil {
		scp.Settings.Trigger.Mode = settings.TriggerModeSingle
	}
	return scp.RunScope()
}

// AutoRange triggers the auto range feature.
func (scp *ScpDesc) AutoRange() error {
	scp.ExecuteVoiceCommand("auto range")
	return nil
}

// GetChannels returns information on all physical channels.
func (scp *ScpDesc) GetChannels() []ChannelInfo {
	channels := make([]ChannelInfo, 0, scp.channelCount)
	count := int(scp.channelCount)
	if count <= 0 && scp.Settings != nil {
		count = len(scp.Settings.Channels)
	}
	for i := 0; i < count; i++ {
		chName := string(rune('A' + i))
		var enabled bool
		var rangeStr string
		var rangeMV int32
		var coupling string
		var invert bool
		var offset int
		var isTrig bool

		if scp.Settings != nil && i < len(scp.Settings.Channels) {
			chs := scp.Settings.Channels[i]
			enabled = chs.Enabled
			invert = chs.Inverted
			offset = chs.DisplayVOffset
			isTrig = chs.TriggerSource
			if int(chs.CoupleType) == int(genericps.Ac) {
				coupling = "AC"
			} else {
				coupling = "DC"
			}
			if int(chs.VRange) >= 0 && int(chs.VRange) < len(genericps.InputRanges) {
				rangeMV = genericps.InputRanges[chs.VRange]
				rangeStr = fmt.Sprintf("%dmV", rangeMV)
				if rangeMV >= 1000 && rangeMV%1000 == 0 {
					rangeStr = fmt.Sprintf("%dV", rangeMV/1000)
				}
			}
		}

		channels = append(channels, ChannelInfo{
			ID:             chName,
			Index:          i,
			Enabled:        enabled,
			Range:          rangeStr,
			RangeMV:        rangeMV,
			Coupling:       coupling,
			Invert:         invert,
			DisplayVOffset: offset,
			TriggerSource:  isTrig,
		})
	}
	return channels
}

// SetChannel updates runtime and persistent settings for a specific channel.
func (scp *ScpDesc) SetChannel(chIdx int, enabled *bool, coupling *string, vrange *string, invert *bool) error {
	if chIdx < 0 || (scp.channelCount > 0 && chIdx >= int(scp.channelCount)) {
		return fmt.Errorf("channel index %d out of bounds (max %d)", chIdx, scp.channelCount)
	}

	if scp.Settings != nil && chIdx < len(scp.Settings.Channels) {
		ch := &scp.Settings.Channels[chIdx]
		if enabled != nil {
			ch.Enabled = *enabled
		}
		if invert != nil {
			ch.Inverted = *invert
		}
		if coupling != nil {
			switch strings.ToUpper(strings.TrimSpace(*coupling)) {
			case "AC":
				ch.CoupleType = genericps.Ac
			case "DC":
				ch.CoupleType = genericps.Dc
			}
		}
		if vrange != nil {
			vstr := strings.ToLower(strings.TrimSpace(*vrange))
			for idx, mv := range genericps.InputRanges {
				s1 := fmt.Sprintf("%dmv", mv)
				s2 := fmt.Sprintf("%dv", mv/1000)
				if vstr == s1 || (mv >= 1000 && vstr == s2) {
					ch.VRange = genericps.RangeEnum(idx)
					break
				}
			}
		}
		scp.SaveSettings()
	}

	// Update GUI widgets if active
	if chIdx < len(scp.channelViewers) {
		cv := &scp.channelViewers[chIdx]
		fyne.Do(func() {
			if enabled != nil && cv.enableCheckbox != nil {
				cv.enableCheckbox.SetVal(*enabled)
			}
			if invert != nil && cv.invertCheckbox != nil {
				cv.invertCheckbox.SetChecked(*invert)
			}
		})
	}

	return nil
}

// GetTrigger returns the active trigger configuration.
func (scp *ScpDesc) GetTrigger() TriggerInfo {
	var srcName string
	if int(scp.triggerSource) >= 0 && int(scp.triggerSource) < len(scp.triggerSources) {
		srcName = scp.triggerSources[scp.triggerSource]
	} else if int(scp.triggerSource) >= 0 {
		srcName = string(rune('A' + int(scp.triggerSource)))
	} else {
		srcName = "None"
	}

	var mode string
	var dir string
	var trigType string
	var thresholdMV float32
	var hysteresis int32

	if scp.Settings != nil {
		mode = scp.Settings.Trigger.Mode
		trigType = scp.Settings.Trigger.Type
		srcIdx := int(scp.triggerSource)
		if srcIdx >= 0 && srcIdx < len(scp.Settings.Channels) {
			thresholdMV = float32(scp.Settings.Channels[srcIdx].Trigger.Mv)
			hysteresis = scp.Settings.Channels[srcIdx].Trigger.Hysteresis
			if scp.Settings.Channels[srcIdx].Trigger.TriggerDirection == genericps.TriggerRising {
				dir = "Rising"
			} else if scp.Settings.Channels[srcIdx].Trigger.TriggerDirection == genericps.TriggerFalling {
				dir = "Falling"
			} else {
				dir = fmt.Sprintf("%d", scp.Settings.Channels[srcIdx].Trigger.TriggerDirection)
			}
		}
	}

	return TriggerInfo{
		Source:      srcName,
		Mode:        mode,
		Direction:   dir,
		Type:        trigType,
		ThresholdMV: thresholdMV,
		Hysteresis:  hysteresis,
	}
}

// SetTrigger configures trigger parameters.
func (scp *ScpDesc) SetTrigger(source string, mode string, direction string, thresholdMV *float32) error {
	if source != "" {
		chIdx, err := parseChannelIdentifier(source)
		if err == nil && chIdx < int(scp.channelCount) {
			scp.triggerSource = genericps.ChannelId(chIdx)
			if scp.Settings != nil && chIdx < len(scp.Settings.Channels) {
				for i := range scp.Settings.Channels {
					scp.Settings.Channels[i].TriggerSource = (i == chIdx)
				}
			}
		}
	}

	if mode != "" && scp.Settings != nil {
		scp.Settings.Trigger.Mode = mode
		if scp.triggerModeSelect != nil {
			fyne.Do(func() {
				scp.triggerModeSelect.SetSelected(mode)
			})
		}
	}

	if direction != "" && scp.Settings != nil {
		srcIdx := int(scp.triggerSource)
		if srcIdx >= 0 && srcIdx < len(scp.Settings.Channels) {
			switch strings.ToLower(strings.TrimSpace(direction)) {
			case "rising":
				scp.Settings.Channels[srcIdx].Trigger.TriggerDirection = genericps.TriggerRising
			case "falling":
				scp.Settings.Channels[srcIdx].Trigger.TriggerDirection = genericps.TriggerFalling
			}
		}
	}

	if thresholdMV != nil && scp.Settings != nil {
		srcIdx := int(scp.triggerSource)
		if srcIdx >= 0 && srcIdx < len(scp.Settings.Channels) {
			scp.Settings.Channels[srcIdx].Trigger.Mv = int32(*thresholdMV)
		}
	}

	scp.SaveSettings()
	return nil
}

// GetTimebase returns current horizontal time per division and sampling interval.
func (scp *ScpDesc) GetTimebase() TimebaseInfo {
	return TimebaseInfo{
		TimeDiv:        scp.timeDiv,
		TimeUnit:       scp.timeUnit,
		TimeUnitStr:    timeUnitToString(scp.timeUnit),
		SampleInterval: scp.controlSamplingTimeInterval,
	}
}

// SetTimebase updates the horizontal timebase settings.
func (scp *ScpDesc) SetTimebase(timeDiv int, timeUnit int) error {
	scp.timeDiv = timeDiv
	scp.timeUnit = timeUnit
	if scp.Settings != nil {
		scp.Settings.Time.TimeDiv = fmt.Sprintf("%d", timeDiv)
		scp.Settings.Time.Unit = timeUnitToString(timeUnit)
		scp.SaveSettings()
	}
	return nil
}

// GetWaveformData extracts the latest acquired float32 waveform voltage samples for all active channels.
func (scp *ScpDesc) GetWaveformData() map[string]WaveformData {
	scp.screenLocker.Lock()
	defer scp.screenLocker.Unlock()

	result := make(map[string]WaveformData)
	for i := 0; i < len(scp.displayBuffers) && i < int(scp.channelCount); i++ {
		if scp.Settings != nil && i < len(scp.Settings.Channels) && !scp.Settings.Channels[i].Enabled {
			continue
		}
		chName := string(rune('A' + i))
		buf := scp.displayBuffers[i]
		data := make([]float32, len(buf))
		copy(data, buf)

		result[chName] = WaveformData{
			Channel:        chName,
			Unit:           "mV",
			SampleCount:    len(data),
			SampleInterval: scp.controlSamplingTimeInterval,
			Data:           data,
		}
	}
	return result
}

// GetMeasurements calculates and returns current metrics (Vmin, Vmax, Vpp, Vmean, Vrms, Frequency, Period)
// for each active channel based on the latest acquired waveform buffer.
func (scp *ScpDesc) GetMeasurements() map[string]ChannelMeasurements {
	scp.screenLocker.Lock()
	defer scp.screenLocker.Unlock()

	result := make(map[string]ChannelMeasurements)
	for i := 0; i < len(scp.displayBuffers) && i < int(scp.channelCount); i++ {
		if scp.Settings != nil && i < len(scp.Settings.Channels) && !scp.Settings.Channels[i].Enabled {
			continue
		}
		chName := string(rune('A' + i))
		buf := scp.displayBuffers[i]
		if len(buf) == 0 {
			continue
		}

		var minVal float32 = float32(math.MaxFloat32)
		var maxVal float32 = -float32(math.MaxFloat32)
		var sum float64 = 0
		var sumSq float64 = 0

		for _, v := range buf {
			if v < minVal {
				minVal = v
			}
			if v > maxVal {
				maxVal = v
			}
			sum += float64(v)
			sumSq += float64(v) * float64(v)
		}

		n := float64(len(buf))
		vMean := float32(sum / n)
		vRms := float32(math.Sqrt(sumSq / n))
		vPp := maxVal - minVal

		var frq, period float64
		if len(buf) > 1 && scp.controlSamplingTimeInterval > 0 {
			mean := (minVal + maxVal) / 2
			frq, period = measureFrq(buf, mean, mean+0.8*(maxVal-mean), mean-0.8*(mean-minVal), scp.controlSamplingTimeInterval)
			if math.IsNaN(frq) || math.IsInf(frq, 0) || frq < 0 {
				frq = 0
			}
			if math.IsNaN(period) || math.IsInf(period, 0) || period < 0 {
				period = 0
			}
		}

		result[chName] = ChannelMeasurements{
			Channel:     chName,
			VMin:        minVal,
			VMax:        maxVal,
			VPp:         vPp,
			VMean:       vMean,
			VRms:        vRms,
			FrequencyHz: frq,
			PeriodS:     period,
		}
	}
	return result
}

// GetScreenshot captures the oscilloscope GUI screen as an encoded image byte slice.
func (scp *ScpDesc) GetScreenshot(format string) ([]byte, string, error) {
	if scp.Window == nil || scp.Window.Canvas() == nil {
		return nil, "", fmt.Errorf("screen capture unavailable: no active window canvas")
	}

	img := scp.Window.Canvas().Capture()
	if img == nil {
		return nil, "", fmt.Errorf("screen capture returned empty image")
	}

	var buf bytes.Buffer
	format = strings.ToLower(strings.TrimSpace(format))
	if format == "jpeg" || format == "jpg" {
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
			return nil, "", err
		}
		return buf.Bytes(), "image/jpeg", nil
	}

	if err := png.Encode(&buf, img); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), "image/png", nil
}

// NewAPIMux creates and registers all HTTP handlers for the general remote API.
func (scp *ScpDesc) NewAPIMux(authAdmin, authView string) *http.ServeMux {
	adminUser, adminPass := parseAuthCredentials(authAdmin)
	viewUser, viewPass := parseAuthCredentials(authView)

	checkAuth := func(r *http.Request) (allowed bool, isAdmin bool) {
		if adminUser == "" && viewUser == "" {
			return true, true // Open access
		}
		u, p, ok := r.BasicAuth()
		if !ok {
			return false, false
		}
		if adminUser != "" && u == adminUser && p == adminPass {
			return true, true
		}
		if viewUser != "" && u == viewUser && p == viewPass {
			return true, false
		}
		return false, false
	}

	writeJSON := func(w http.ResponseWriter, status int, data any) {
		body, err := json.Marshal(data)
		if err != nil {
			http.Error(w, `{"error":"internal json marshal error"}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(status)
		_, _ = w.Write(body)
	}

	writeErr := func(w http.ResponseWriter, status int, msg string) {
		writeJSON(w, status, map[string]string{"error": msg})
	}

	mux := http.NewServeMux()

	// GET /api/help
	mux.HandleFunc("/api/help", func(w http.ResponseWriter, r *http.Request) {
		allowed, _ := checkAuth(r)
		if !allowed {
			w.Header().Set("WWW-Authenticate", `Basic realm="Fynescope Remote API"`)
			writeErr(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		endpoints := map[string]string{
			"GET /api/status":           "Returns current run state and basic metrics",
			"GET /api/info":             "Returns hardware and configuration information",
			"GET /api/settings":         "Returns full application settings JSON",
			"POST /api/settings":        "Applies deep-merged JSON payload to settings",
			"POST /api/run":             "Starts continuous oscilloscope capture",
			"POST /api/stop":            "Stops oscilloscope capture",
			"POST /api/single":          "Arms single acquisition",
			"POST /api/autorange":       "Triggers auto-range",
			"POST /api/command":         "Executes natural language/voice command (JSON {\"command\": \"...\"} or text body)",
			"GET /api/channels":         "Lists all channels with configuration and status",
			"GET /api/channels/{id}":    "Gets settings for specific channel (A-D or 0-3)",
			"POST /api/channels/{id}":   "Updates channel settings (enabled, coupling, range, invert)",
			"POST /api/channels/{id}/enable":  "Enables channel",
			"POST /api/channels/{id}/disable": "Disables channel",
			"GET /api/timebase":         "Returns horizontal timebase and sampling interval",
			"POST /api/timebase":        "Configures timebase (time_div, time_unit)",
			"GET /api/trigger":          "Returns trigger settings",
			"POST /api/trigger":         "Configures trigger (source, mode, direction, threshold_mv)",
			"GET /api/waveform":         "Returns latest acquired waveform samples in mV (?format=json|csv)",
			"GET /api/waveform/{id}":    "Returns waveform samples for specific channel (?format=json|csv)",
			"GET /api/measurements":     "Returns computed measurements (Vpp, Vmin, Vmax, Vmean, Vrms, frequency, period)",
			"GET /api/screenshot":       "Captures GUI screen (?format=png|jpeg)",
		}
		writeJSON(w, http.StatusOK, map[string]any{"endpoints": endpoints})
	})

	// GET /api/status
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		allowed, _ := checkAuth(r)
		if !allowed {
			w.Header().Set("WWW-Authenticate", `Basic realm="Fynescope Remote API"`)
			writeErr(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		if r.Method != http.MethodGet {
			writeErr(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"status":   scp.GetStatus(),
			"running":  scp.running,
			"channels": scp.channelCount,
		})
	})

	// GET /api/info
	mux.HandleFunc("/api/info", func(w http.ResponseWriter, r *http.Request) {
		allowed, _ := checkAuth(r)
		if !allowed {
			w.Header().Set("WWW-Authenticate", `Basic realm="Fynescope Remote API"`)
			writeErr(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		if r.Method != http.MethodGet {
			writeErr(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		writeJSON(w, http.StatusOK, scp.GetInfo())
	})

	// GET, POST /api/settings
	mux.HandleFunc("/api/settings", func(w http.ResponseWriter, r *http.Request) {
		allowed, isAdmin := checkAuth(r)
		if !allowed {
			w.Header().Set("WWW-Authenticate", `Basic realm="Fynescope Remote API"`)
			writeErr(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		if r.Method == http.MethodGet {
			writeJSON(w, http.StatusOK, scp.GetSettings())
			return
		}
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			if !isAdmin {
				writeErr(w, http.StatusForbidden, "Admin access required")
				return
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				writeErr(w, http.StatusBadRequest, "Failed to read request body")
				return
			}
			if err := scp.ApplySettingsJSON(body); err != nil {
				writeErr(w, http.StatusBadRequest, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, map[string]string{"result": "settings applied"})
			return
		}
		writeErr(w, http.StatusMethodNotAllowed, "Method not allowed")
	})

	// POST /api/run
	mux.HandleFunc("/api/run", func(w http.ResponseWriter, r *http.Request) {
		allowed, isAdmin := checkAuth(r)
		if !allowed {
			w.Header().Set("WWW-Authenticate", `Basic realm="Fynescope Remote API"`)
			writeErr(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		if !isAdmin {
			writeErr(w, http.StatusForbidden, "Admin access required")
			return
		}
		if r.Method != http.MethodPost {
			writeErr(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		_ = scp.RunScope()
		writeJSON(w, http.StatusOK, map[string]string{"result": "running"})
	})

	// POST /api/stop
	mux.HandleFunc("/api/stop", func(w http.ResponseWriter, r *http.Request) {
		allowed, isAdmin := checkAuth(r)
		if !allowed {
			w.Header().Set("WWW-Authenticate", `Basic realm="Fynescope Remote API"`)
			writeErr(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		if !isAdmin {
			writeErr(w, http.StatusForbidden, "Admin access required")
			return
		}
		if r.Method != http.MethodPost {
			writeErr(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		_ = scp.StopScope()
		writeJSON(w, http.StatusOK, map[string]string{"result": "stopped"})
	})

	// POST /api/single
	mux.HandleFunc("/api/single", func(w http.ResponseWriter, r *http.Request) {
		allowed, isAdmin := checkAuth(r)
		if !allowed {
			w.Header().Set("WWW-Authenticate", `Basic realm="Fynescope Remote API"`)
			writeErr(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		if !isAdmin {
			writeErr(w, http.StatusForbidden, "Admin access required")
			return
		}
		if r.Method != http.MethodPost {
			writeErr(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		_ = scp.SingleScope()
		writeJSON(w, http.StatusOK, map[string]string{"result": "single armed"})
	})

	// POST /api/autorange
	mux.HandleFunc("/api/autorange", func(w http.ResponseWriter, r *http.Request) {
		allowed, isAdmin := checkAuth(r)
		if !allowed {
			w.Header().Set("WWW-Authenticate", `Basic realm="Fynescope Remote API"`)
			writeErr(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		if !isAdmin {
			writeErr(w, http.StatusForbidden, "Admin access required")
			return
		}
		if r.Method != http.MethodPost {
			writeErr(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		_ = scp.AutoRange()
		writeJSON(w, http.StatusOK, map[string]string{"result": "autorange triggered"})
	})

	// POST /api/command
	mux.HandleFunc("/api/command", func(w http.ResponseWriter, r *http.Request) {
		allowed, isAdmin := checkAuth(r)
		if !allowed {
			w.Header().Set("WWW-Authenticate", `Basic realm="Fynescope Remote API"`)
			writeErr(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		if !isAdmin {
			writeErr(w, http.StatusForbidden, "Admin access required")
			return
		}
		if r.Method != http.MethodPost {
			writeErr(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		body, _ := io.ReadAll(r.Body)
		cmd := strings.TrimSpace(string(body))
		var jsonCmd struct {
			Command string `json:"command"`
		}
		if json.Unmarshal(body, &jsonCmd) == nil && jsonCmd.Command != "" {
			cmd = jsonCmd.Command
		}
		if cmd == "" {
			writeErr(w, http.StatusBadRequest, "Missing command string")
			return
		}
		scp.ExecuteVoiceCommand(cmd)
		writeJSON(w, http.StatusOK, map[string]string{"result": "command executed", "command": cmd})
	})

	// Channels routing: /api/channels and /api/channels/
	mux.HandleFunc("/api/channels", func(w http.ResponseWriter, r *http.Request) {
		allowed, _ := checkAuth(r)
		if !allowed {
			w.Header().Set("WWW-Authenticate", `Basic realm="Fynescope Remote API"`)
			writeErr(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		if r.Method != http.MethodGet {
			writeErr(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		writeJSON(w, http.StatusOK, scp.GetChannels())
	})

	mux.HandleFunc("/api/channels/", func(w http.ResponseWriter, r *http.Request) {
		allowed, isAdmin := checkAuth(r)
		if !allowed {
			w.Header().Set("WWW-Authenticate", `Basic realm="Fynescope Remote API"`)
			writeErr(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		path := strings.TrimPrefix(r.URL.Path, "/api/channels/")
		parts := strings.Split(strings.Trim(path, "/"), "/")
		if len(parts) == 0 || parts[0] == "" {
			writeErr(w, http.StatusBadRequest, "Missing channel identifier")
			return
		}
		chIdx, err := parseChannelIdentifier(parts[0])
		if err != nil {
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}

		// Check sub-actions: /enable or /disable
		if len(parts) > 1 {
			if !isAdmin {
				writeErr(w, http.StatusForbidden, "Admin access required")
				return
			}
			if r.Method != http.MethodPost {
				writeErr(w, http.StatusMethodNotAllowed, "Method not allowed")
				return
			}
			subAction := strings.ToLower(parts[1])
			if subAction == "enable" {
				t := true
				_ = scp.SetChannel(chIdx, &t, nil, nil, nil)
				writeJSON(w, http.StatusOK, map[string]any{"result": "channel enabled", "channel": parts[0]})
				return
			} else if subAction == "disable" {
				f := false
				_ = scp.SetChannel(chIdx, &f, nil, nil, nil)
				writeJSON(w, http.StatusOK, map[string]any{"result": "channel disabled", "channel": parts[0]})
				return
			}
			writeErr(w, http.StatusNotFound, "Unknown channel action")
			return
		}

		if r.Method == http.MethodGet {
			channels := scp.GetChannels()
			if chIdx >= 0 && chIdx < len(channels) {
				writeJSON(w, http.StatusOK, channels[chIdx])
				return
			}
			writeErr(w, http.StatusNotFound, "Channel not found")
			return
		}

		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			if !isAdmin {
				writeErr(w, http.StatusForbidden, "Admin access required")
				return
			}
			var payload struct {
				Enabled  *bool   `json:"enabled"`
				Coupling *string `json:"coupling"`
				Range    *string `json:"range"`
				Invert   *bool   `json:"invert"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				writeErr(w, http.StatusBadRequest, "Invalid JSON body")
				return
			}
			if err := scp.SetChannel(chIdx, payload.Enabled, payload.Coupling, payload.Range, payload.Invert); err != nil {
				writeErr(w, http.StatusBadRequest, err.Error())
				return
			}
			channels := scp.GetChannels()
			writeJSON(w, http.StatusOK, channels[chIdx])
			return
		}

		writeErr(w, http.StatusMethodNotAllowed, "Method not allowed")
	})

	// Timebase: /api/timebase
	mux.HandleFunc("/api/timebase", func(w http.ResponseWriter, r *http.Request) {
		allowed, isAdmin := checkAuth(r)
		if !allowed {
			w.Header().Set("WWW-Authenticate", `Basic realm="Fynescope Remote API"`)
			writeErr(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		if r.Method == http.MethodGet {
			writeJSON(w, http.StatusOK, scp.GetTimebase())
			return
		}
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			if !isAdmin {
				writeErr(w, http.StatusForbidden, "Admin access required")
				return
			}
			var payload struct {
				TimeDiv  int `json:"time_div"`
				TimeUnit int `json:"time_unit"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				writeErr(w, http.StatusBadRequest, "Invalid JSON body")
				return
			}
			_ = scp.SetTimebase(payload.TimeDiv, payload.TimeUnit)
			writeJSON(w, http.StatusOK, scp.GetTimebase())
			return
		}
		writeErr(w, http.StatusMethodNotAllowed, "Method not allowed")
	})

	// Trigger: /api/trigger
	mux.HandleFunc("/api/trigger", func(w http.ResponseWriter, r *http.Request) {
		allowed, isAdmin := checkAuth(r)
		if !allowed {
			w.Header().Set("WWW-Authenticate", `Basic realm="Fynescope Remote API"`)
			writeErr(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		if r.Method == http.MethodGet {
			writeJSON(w, http.StatusOK, scp.GetTrigger())
			return
		}
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			if !isAdmin {
				writeErr(w, http.StatusForbidden, "Admin access required")
				return
			}
			var payload struct {
				Source      string   `json:"source"`
				Mode        string   `json:"mode"`
				Direction   string   `json:"direction"`
				ThresholdMV *float32 `json:"threshold_mv"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				writeErr(w, http.StatusBadRequest, "Invalid JSON body")
				return
			}
			if err := scp.SetTrigger(payload.Source, payload.Mode, payload.Direction, payload.ThresholdMV); err != nil {
				writeErr(w, http.StatusBadRequest, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, scp.GetTrigger())
			return
		}
		writeErr(w, http.StatusMethodNotAllowed, "Method not allowed")
	})

	// Waveform: /api/waveform and /api/waveform/
	writeWaveformCSV := func(w http.ResponseWriter, chName string, wf WaveformData) {
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"waveform_%s.csv\"", chName))
		writer := csv.NewWriter(w)
		_ = writer.Write([]string{"SampleIndex", "Time_s", "Voltage_mV"})
		dt := wf.SampleInterval
		for idx, v := range wf.Data {
			t := float64(idx) * dt
			_ = writer.Write([]string{strconv.Itoa(idx), fmt.Sprintf("%.9e", t), fmt.Sprintf("%.4f", v)})
		}
		writer.Flush()
	}

	mux.HandleFunc("/api/waveform", func(w http.ResponseWriter, r *http.Request) {
		allowed, _ := checkAuth(r)
		if !allowed {
			w.Header().Set("WWW-Authenticate", `Basic realm="Fynescope Remote API"`)
			writeErr(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		if r.Method != http.MethodGet {
			writeErr(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		wf := scp.GetWaveformData()
		format := strings.ToLower(r.URL.Query().Get("format"))
		if format == "csv" {
			// Output first active channel as CSV if multiple
			for chName, data := range wf {
				writeWaveformCSV(w, chName, data)
				return
			}
			writeErr(w, http.StatusNotFound, "No waveform data available")
			return
		}
		writeJSON(w, http.StatusOK, wf)
	})

	mux.HandleFunc("/api/waveform/", func(w http.ResponseWriter, r *http.Request) {
		allowed, _ := checkAuth(r)
		if !allowed {
			w.Header().Set("WWW-Authenticate", `Basic realm="Fynescope Remote API"`)
			writeErr(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		if r.Method != http.MethodGet {
			writeErr(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		chStr := strings.TrimPrefix(r.URL.Path, "/api/waveform/")
		chIdx, err := parseChannelIdentifier(chStr)
		if err != nil {
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}
		chName := string(rune('A' + chIdx))
		wfAll := scp.GetWaveformData()
		wf, exists := wfAll[chName]
		if !exists {
			writeErr(w, http.StatusNotFound, fmt.Sprintf("Waveform data for channel %s not found or disabled", chName))
			return
		}
		format := strings.ToLower(r.URL.Query().Get("format"))
		if format == "csv" {
			writeWaveformCSV(w, chName, wf)
			return
		}
		writeJSON(w, http.StatusOK, wf)
	})

	// Measurements: /api/measurements
	mux.HandleFunc("/api/measurements", func(w http.ResponseWriter, r *http.Request) {
		allowed, _ := checkAuth(r)
		if !allowed {
			w.Header().Set("WWW-Authenticate", `Basic realm="Fynescope Remote API"`)
			writeErr(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		if r.Method != http.MethodGet {
			writeErr(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		writeJSON(w, http.StatusOK, scp.GetMeasurements())
	})

	// Screenshot: /api/screenshot
	mux.HandleFunc("/api/screenshot", func(w http.ResponseWriter, r *http.Request) {
		allowed, _ := checkAuth(r)
		if !allowed {
			w.Header().Set("WWW-Authenticate", `Basic realm="Fynescope Remote API"`)
			writeErr(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		if r.Method != http.MethodGet {
			writeErr(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		format := r.URL.Query().Get("format")
		imgBytes, contentType, err := scp.GetScreenshot(format)
		if err != nil {
			writeErr(w, http.StatusServiceUnavailable, err.Error())
			return
		}
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(imgBytes)
	})

	return mux
}

// StartAPIServer launches an HTTPS remote API server on the configured port.
func (scp *ScpDesc) StartAPIServer(cfg APIServerConfig) error {
	scp.apiServerMu.Lock()
	defer scp.apiServerMu.Unlock()

	if scp.apiServer != nil && scp.apiServer.active {
		return fmt.Errorf("API server is already running on %s", scp.apiServer.addr)
	}

	if cfg.Port <= 0 {
		return fmt.Errorf("invalid API server port: %d", cfg.Port)
	}

	addr := fmt.Sprintf("0.0.0.0:%d", cfg.Port)
	mux := scp.NewAPIMux(cfg.AuthAdmin, cfg.AuthView)

	var tlsCert tls.Certificate
	var err error
	if cfg.CertFile != "" && cfg.KeyFile != "" {
		tlsCert, err = tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			return fmt.Errorf("failed to load TLS key pair: %w", err)
		}
	} else {
		tlsCert, err = generateAPICert()
		if err != nil {
			return fmt.Errorf("failed to generate self-signed TLS cert: %w", err)
		}
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		MinVersion:   tls.VersionTLS12,
	}

	ln, err := tls.Listen("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	srv := &http.Server{
		Addr:      addr,
		Handler:   mux,
		TLSConfig: tlsConfig,
	}

	apiSrv := &APIServer{
		server: srv,
		addr:   addr,
		active: true,
	}
	scp.apiServer = apiSrv

	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTPS remote API server error", "err", err)
		}
		apiSrv.mu.Lock()
		apiSrv.active = false
		apiSrv.mu.Unlock()
	}()

	slog.Info("HTTPS Remote API server started", "addr", addr)
	return nil
}

// StopAPIServer gracefully stops the running HTTPS remote API server.
func (scp *ScpDesc) StopAPIServer() error {
	scp.apiServerMu.Lock()
	defer scp.apiServerMu.Unlock()

	if scp.apiServer == nil || !scp.apiServer.active {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := scp.apiServer.server.Shutdown(ctx)
	scp.apiServer.active = false
	scp.apiServer = nil
	return err
}

// IsAPIServerRunning returns true if the HTTPS remote API server is currently active.
func (scp *ScpDesc) IsAPIServerRunning() bool {
	scp.apiServerMu.Lock()
	defer scp.apiServerMu.Unlock()
	return scp.apiServer != nil && scp.apiServer.active
}
