package main

import (
	"fynescope/control"
	"fynescope/genericps"
	"fynescope/netcontrol"
	"os"
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
)

func TestNormalizeServerAddress(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "host and port",
			input: "192.168.1.42:9876",
			want:  "192.168.1.42:9876",
		},
		{
			name:  "host without port uses default 9876",
			input: "192.168.1.42",
			want:  "192.168.1.42:9876",
		},
		{
			name:  "hostname without port",
			input: "  scope.local  ",
			want:  "scope.local:9876",
		},
		{
			name:  "custom port preserved",
			input: "10.0.2.2:9999",
			want:  "10.0.2.2:9999",
		},
		{
			name:  "ipv6 without port",
			input: "[::1]",
			want:  "[::1]:9876",
		},
		{
			name:    "empty string",
			input:   "   ",
			wantErr: true,
		},
		{
			name:    "missing host with colon",
			input:   ":9876",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := normalizeServerAddress(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for input %q, got %q", tc.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for input %q: %v", tc.input, err)
			}
			if got != tc.want {
				t.Fatalf("normalizeServerAddress(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestRecentServersPersistence(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	prefs := a.Preferences()

	if got := loadRecentServers(prefs); len(got) != 0 {
		t.Fatalf("expected empty initial recent servers, got %v", got)
	}

	saveRecentServer(prefs, "192.168.1.10:9876")
	saveRecentServer(prefs, "192.168.1.11:9876")
	saveRecentServer(prefs, "192.168.1.10:9876") // duplicate should move to front

	got := loadRecentServers(prefs)
	if len(got) != 2 || got[0] != "192.168.1.10:9876" || got[1] != "192.168.1.11:9876" {
		t.Fatalf("unexpected recent servers order: %v", got)
	}
	if last := prefs.String(prefLastServer); last != "192.168.1.10:9876" {
		t.Fatalf("prefLastServer = %q, want 192.168.1.10:9876", last)
	}
}

func hasVerticalScroll(obj fyne.CanvasObject) bool {
	if obj == nil {
		return false
	}
	switch v := obj.(type) {
	case *container.Scroll:
		return v.Direction == container.ScrollVerticalOnly || v.Direction == container.ScrollBoth
	case *container.Split:
		return hasVerticalScroll(v.Leading) || hasVerticalScroll(v.Trailing)
	case *container.AppTabs:
		for _, item := range v.Items {
			if !hasVerticalScroll(item.Content) {
				return false
			}
		}
		return len(v.Items) > 0
	case *fyne.Container:
		for _, child := range v.Objects {
			if hasVerticalScroll(child) {
				return true
			}
		}
	}
	return false
}

func TestConnectAndMobileScrollableTabs(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err == nil {
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()
	}

	genericps.LoadDefaultConstants()

	con := genericps.NewConnection()
	var connErr error
	con.Handle, connErr = genericps.OpenDemo(con, genericps.DemoId)
	if connErr != nil {
		t.Fatalf("OpenDemo failed: %v", connErr)
	}
	con.ID = genericps.DemoId
	defer con.CloseUnit()

	ctrl := control.NewControl(con)
	defer ctrl.Shutdown()

	srv := netcontrol.NewServer(ctrl)
	if err := srv.Start("127.0.0.1:0"); err != nil {
		t.Fatalf("srv.Start failed: %v", err)
	}
	defer srv.Close()

	a := test.NewApp()
	defer a.Quit()
	w := a.NewWindow("Test Android Client")
	w.Resize(fyne.NewSize(800, 400))

	c := newClientApp(a, w, srv.Addr().String(), false)
	c.showConnectScreen("Ready", false)

	// Verify connection screen is vertically scrollable.
	if _, ok := w.Content().(*container.Scroll); !ok {
		t.Fatalf("expected connection screen root to be *container.Scroll, got %T", w.Content())
	}

	netCli, err := netcontrol.Dial(srv.Addr().String())
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	c.onConnected(srv.Addr().String(), netCli)

	if c.activeScp == nil {
		t.Fatalf("expected activeScp to be initialized after onConnected")
	}
	if !c.activeScp.MobileMode {
		t.Fatalf("expected activeScp.MobileMode to be true")
	}

	// Verify every visible control tab in the oscilloscope UI has vertical scroll support.
	tabs := c.activeScp.ControlTabs()
	if len(tabs) == 0 {
		t.Fatalf("expected non-empty control tabs")
	}
	for _, item := range tabs {
		if !hasVerticalScroll(item.Content) {
			t.Errorf("control tab %q is not vertically scrollable (content type %T)", item.Text, item.Content)
		}
	}

	// Verify resolveSettingsPath sanitizes serial strings.
	sp := resolveSettingsPath(a, "BATCH/123\\456")
	if filepath.Base(sp) != "BATCH_123_456.settings.json" {
		t.Errorf("unexpected settings filename: %s", sp)
	}

	// Verify disconnecting restores the connection screen in the same master window.
	c.disconnectActiveSession("Disconnected")
	if c.activeScp != nil || c.activeCli != nil {
		t.Fatalf("expected active session to be cleared after disconnect")
	}
	if _, ok := w.Content().(*container.Scroll); !ok {
		t.Fatalf("expected connection screen (*container.Scroll) after disconnect, got %T", w.Content())
	}
}
