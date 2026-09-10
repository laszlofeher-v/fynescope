package gui

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"fynescope/genericps"
	"fynescope/settings"
)

func newTestScpDesc() *ScpDesc {
	genericps.ChA = genericps.ChannelId(0)
	genericps.ChB = genericps.ChannelId(1)
	genericps.ChC = genericps.ChannelId(2)
	genericps.ChD = genericps.ChannelId(3)

	cfg := settings.NewDefaultSettings()
	cfg.Channels = make([]settings.ChSettings, 2)
	cfg.Channels[0] = settings.ChSettings{
		ID:            genericps.ChA,
		Enabled:       true,
		VRange:        genericps.Range_1v,
		CoupleType:    genericps.Dc,
		Inverted:      false,
		TriggerSource: true,
		Trigger: settings.ChTriggerSettings{
			TriggerDirection: genericps.TriggerRising,
			Mv:               100,
			Hysteresis:       10,
		},
	}
	cfg.Channels[1] = settings.ChSettings{
		ID:            genericps.ChB,
		Enabled:       false,
		VRange:        genericps.Range_2v,
		CoupleType:    genericps.Ac,
		Inverted:      true,
		TriggerSource: false,
		Trigger: settings.ChTriggerSettings{
			TriggerDirection: genericps.TriggerFalling,
			Mv:               200,
			Hysteresis:       20,
		},
	}
	cfg.Trigger.Mode = settings.TriggerModeAuto
	cfg.Trigger.Type = settings.TriggerTypeSimple

	scp := &ScpDesc{
		Settings:                    cfg,
		channelCount:                2,
		running:                     true,
		triggerSource:               genericps.ChA,
		triggerSources:              []string{"Ch A", "Ch B"},
		timeDiv:                     10,
		timeUnit:                    -6, // us
		controlSamplingTimeInterval: 1e-6,
		MaxValue:                    32767,
		MinValue:                    -32768,
		displayBuffers: [][]float32{
			{10.0, 20.0, 30.0, -10.0, -20.0, 10.0},
			{50.0, 100.0, 150.0},
		},
		channelViewers: make([]channelViewerDesc, 2),
	}
	return scp
}

func TestGenerateAPICert(t *testing.T) {
	cert, err := generateAPICert()
	if err != nil {
		t.Fatalf("generateAPICert failed: %v", err)
	}
	if len(cert.Certificate) == 0 {
		t.Fatalf("expected at least 1 certificate in chain")
	}
}

func TestParseAuthCredentials(t *testing.T) {
	u, p := parseAuthCredentials("admin:secret123")
	if u != "admin" || p != "secret123" {
		t.Errorf("expected admin:secret123, got %q:%q", u, p)
	}
	u, p = parseAuthCredentials("admin")
	if u != "admin" || p != "" {
		t.Errorf("expected admin:'', got %q:%q", u, p)
	}
	u, p = parseAuthCredentials("")
	if u != "" || p != "" {
		t.Errorf("expected empty strings, got %q:%q", u, p)
	}
}

func TestParseChannelIdentifier(t *testing.T) {
	tests := []struct {
		in      string
		wantIdx int
		wantErr bool
	}{
		{"A", 0, false},
		{"b", 1, false},
		{"CH C", 2, false},
		{"ch0", 0, false},
		{"1", 1, false},
		{"3", 3, false},
		{"X", -1, true},
		{"10", -1, true},
	}
	for _, tt := range tests {
		idx, err := parseChannelIdentifier(tt.in)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseChannelIdentifier(%q) error = %v, wantErr %v", tt.in, err, tt.wantErr)
		}
		if idx != tt.wantIdx {
			t.Errorf("parseChannelIdentifier(%q) = %d, want %d", tt.in, idx, tt.wantIdx)
		}
	}
}

func TestScpDesc_GetInfoAndStatus(t *testing.T) {
	scp := newTestScpDesc()
	status := scp.GetStatus()
	if status != "Running" {
		t.Errorf("expected 'Running', got %q", status)
	}
	info := scp.GetInfo()
	if info.ChannelCount != 2 || info.TimeDiv != 10 || info.TimeUnitStr != "us" {
		t.Errorf("unexpected info: %+v", info)
	}
}

func TestAPI_AuthControls(t *testing.T) {
	scp := newTestScpDesc()
	mux := scp.NewAPIMux("admin:adminpass", "viewer:viewpass")

	// 1. No auth -> 401
	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %d", rec.Code)
	}

	// 2. Viewer auth -> GET allowed
	req = httptest.NewRequest(http.MethodGet, "/api/status", nil)
	req.SetBasicAuth("viewer", "viewpass")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 OK for viewer GET, got %d", rec.Code)
	}

	// 3. Viewer auth -> POST forbidden (403)
	req = httptest.NewRequest(http.MethodPost, "/api/run", nil)
	req.SetBasicAuth("viewer", "viewpass")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for viewer POST, got %d", rec.Code)
	}

	// 4. Admin auth -> POST allowed (200)
	req = httptest.NewRequest(http.MethodPost, "/api/run", nil)
	req.SetBasicAuth("admin", "adminpass")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 OK for admin POST, got %d", rec.Code)
	}
}

func TestAPI_Endpoints(t *testing.T) {
	scp := newTestScpDesc()
	mux := scp.NewAPIMux("", "") // open access

	// /api/help
	req := httptest.NewRequest(http.MethodGet, "/api/help", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("/api/help failed: code %d", rec.Code)
	}

	// /api/status
	req = httptest.NewRequest(http.MethodGet, "/api/status", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Running") {
		t.Errorf("/api/status failed: %s", rec.Body.String())
	}

	// /api/info
	req = httptest.NewRequest(http.MethodGet, "/api/info", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "channel_count") {
		t.Errorf("/api/info failed: %s", rec.Body.String())
	}

	// /api/settings GET
	req = httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("/api/settings GET failed: code %d", rec.Code)
	}

	// /api/settings POST
	patch := `{"trigger":{"mode":"Repeat"}}`
	req = httptest.NewRequest(http.MethodPost, "/api/settings", bytes.NewBufferString(patch))
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("/api/settings POST failed: %s", rec.Body.String())
	}
	if scp.Settings.Trigger.Mode != "Repeat" {
		t.Errorf("expected trigger mode Repeat, got %q", scp.Settings.Trigger.Mode)
	}

	// /api/run, /api/stop, /api/single, /api/autorange, /api/command
	for _, path := range []string{"/api/run", "/api/stop", "/api/single", "/api/autorange"} {
		req = httptest.NewRequest(http.MethodPost, path, nil)
		rec = httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("%s failed: code %d", path, rec.Code)
		}
	}

	cmdPayload := `{"command":"run"}`
	req = httptest.NewRequest(http.MethodPost, "/api/command", bytes.NewBufferString(cmdPayload))
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("/api/command failed: %s", rec.Body.String())
	}

	// /api/channels
	req = httptest.NewRequest(http.MethodGet, "/api/channels", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("/api/channels failed: code %d", rec.Code)
	}
	var chs []ChannelInfo
	if err := json.Unmarshal(rec.Body.Bytes(), &chs); err != nil || len(chs) != 2 {
		t.Errorf("failed to unmarshal channels: %v, len=%d", err, len(chs))
	}

	// /api/channels/A GET
	req = httptest.NewRequest(http.MethodGet, "/api/channels/A", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("/api/channels/A failed: code %d", rec.Code)
	}

	// /api/channels/A/disable POST
	req = httptest.NewRequest(http.MethodPost, "/api/channels/A/disable", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("/api/channels/A/disable failed: code %d", rec.Code)
	}
	if scp.Settings.Channels[0].Enabled {
		t.Errorf("expected channel A disabled")
	}

	// /api/channels/A/enable POST
	req = httptest.NewRequest(http.MethodPost, "/api/channels/A/enable", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("/api/channels/A/enable failed: code %d", rec.Code)
	}
	if !scp.Settings.Channels[0].Enabled {
		t.Errorf("expected channel A enabled")
	}

	// /api/channels/A POST config
	cfgBody := `{"coupling":"AC","range":"500mV","invert":true}`
	req = httptest.NewRequest(http.MethodPost, "/api/channels/A", bytes.NewBufferString(cfgBody))
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("/api/channels/A POST config failed: %s", rec.Body.String())
	}
	if scp.Settings.Channels[0].CoupleType != genericps.Ac || !scp.Settings.Channels[0].Inverted {
		t.Errorf("channel A update mismatch: %+v", scp.Settings.Channels[0])
	}

	// /api/timebase GET and POST
	req = httptest.NewRequest(http.MethodGet, "/api/timebase", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("/api/timebase GET failed: code %d", rec.Code)
	}

	tbBody := `{"time_div":20,"time_unit":-3}`
	req = httptest.NewRequest(http.MethodPost, "/api/timebase", bytes.NewBufferString(tbBody))
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("/api/timebase POST failed: code %d", rec.Code)
	}
	if scp.timeDiv != 20 || scp.timeUnit != -3 {
		t.Errorf("timebase update mismatch: %d, %d", scp.timeDiv, scp.timeUnit)
	}

	// /api/trigger GET and POST
	req = httptest.NewRequest(http.MethodGet, "/api/trigger", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("/api/trigger GET failed: code %d", rec.Code)
	}

	trigBody := `{"source":"B","mode":"Normal","direction":"Falling","threshold_mv":250}`
	req = httptest.NewRequest(http.MethodPost, "/api/trigger", bytes.NewBufferString(trigBody))
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("/api/trigger POST failed: code %d", rec.Code)
	}
	if scp.triggerSource != genericps.ChB || scp.Settings.Channels[1].Trigger.Mv != 250 {
		t.Errorf("trigger update mismatch: src=%d, mv=%d", scp.triggerSource, scp.Settings.Channels[1].Trigger.Mv)
	}

	// /api/waveform GET (JSON)
	req = httptest.NewRequest(http.MethodGet, "/api/waveform", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("/api/waveform GET failed: code %d", rec.Code)
	}
	var wfMap map[string]WaveformData
	if err := json.Unmarshal(rec.Body.Bytes(), &wfMap); err != nil {
		t.Errorf("waveform JSON unmarshal failed: %v", err)
	}

	// /api/waveform/A GET (CSV)
	req = httptest.NewRequest(http.MethodGet, "/api/waveform/A?format=csv", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Voltage_mV") {
		t.Errorf("/api/waveform/A CSV failed: %s", rec.Body.String())
	}

	// /api/measurements GET
	req = httptest.NewRequest(http.MethodGet, "/api/measurements", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("/api/measurements failed: code %d", rec.Code)
	}
	var m map[string]ChannelMeasurements
	if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
		t.Errorf("measurements JSON unmarshal failed: %v", err)
	}
	if mA, ok := m["A"]; ok {
		if mA.VPp == 0 {
			t.Errorf("expected non-zero VPp for channel A")
		}
	} else {
		t.Errorf("missing measurements for channel A")
	}
}

func getFreePort() (int, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer ln.Close()
	return ln.Addr().(*net.TCPAddr).Port, nil
}

func TestAPIServer_Lifecycle(t *testing.T) {
	scp := newTestScpDesc()
	port, err := getFreePort()
	if err != nil {
		t.Fatalf("failed to find free port: %v", err)
	}

	cfg := APIServerConfig{
		Port:      port,
		AuthAdmin: "admin:secret",
	}

	err = scp.StartAPIServer(cfg)
	if err != nil {
		t.Fatalf("StartAPIServer failed: %v", err)
	}
	defer scp.StopAPIServer()

	if !scp.IsAPIServerRunning() {
		t.Errorf("expected IsAPIServerRunning to be true")
	}

	// Attempting to start again while running should fail
	if err := scp.StartAPIServer(cfg); err == nil {
		t.Errorf("expected error when starting already running server")
	}

	// Connect over HTTPS with TLS
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr, Timeout: 2 * time.Second}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://127.0.0.1:%d/api/status", port), nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.SetBasicAuth("admin", "secret")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("HTTPS request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("expected 200 OK, got %d: %s", resp.StatusCode, string(body))
	}

	// Stop server
	if err := scp.StopAPIServer(); err != nil {
		t.Errorf("StopAPIServer failed: %v", err)
	}
	if scp.IsAPIServerRunning() {
		t.Errorf("expected IsAPIServerRunning to be false after stop")
	}
}
