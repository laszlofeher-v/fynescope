//go:build web

package web

import (
	"image"
	"image/color"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// blankCapture returns a tiny 1x1 white image — used as a stub for getCapture.
func blankCapture() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.White)
	return img
}

// --- parseAuth ---

func TestParseAuth_UserPass(t *testing.T) {
	u, p := parseAuth("admin:secret")
	if u != "admin" || p != "secret" {
		t.Errorf("expected admin:secret, got %q:%q", u, p)
	}
}

func TestParseAuth_UserOnly(t *testing.T) {
	u, p := parseAuth("admin")
	if u != "admin" || p != "" {
		t.Errorf("expected admin:'', got %q:%q", u, p)
	}
}

func TestParseAuth_Empty(t *testing.T) {
	u, p := parseAuth("")
	if u != "" || p != "" {
		t.Errorf("expected empty, got %q:%q", u, p)
	}
}

// --- newServerMux: no auth (open access) ---

func TestServerMux_NoAuth_Root(t *testing.T) {
	mux := newServerMux("", "", blankCapture, nil)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Fynescope") {
		t.Error("expected HTML body to contain 'Fynescope'")
	}
}

func TestServerMux_NoAuth_CommandPost(t *testing.T) {
	var received string
	mux := newServerMux("", "", blankCapture, func(cmd string) {
		received = cmd
	})

	req := httptest.NewRequest(http.MethodPost, "/command", strings.NewReader("start"))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if received != "start" {
		t.Errorf("expected command 'start', got %q", received)
	}
}

// --- newServerMux: with auth ---

func TestServerMux_WithAuth_NoCredentials_Returns401(t *testing.T) {
	mux := newServerMux("admin:secret", "viewer:hello", blankCapture, nil)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestServerMux_WithAuth_AdminCredentials_Returns200(t *testing.T) {
	mux := newServerMux("admin:secret", "viewer:hello", blankCapture, nil)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.SetBasicAuth("admin", "secret")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	// Admin should see the voice control UI
	if !strings.Contains(rec.Body.String(), "micBtn") {
		t.Error("expected admin HTML with voice control button")
	}
}

func TestServerMux_WithAuth_ViewerCredentials_Returns200_NoVoice(t *testing.T) {
	mux := newServerMux("admin:secret", "viewer:hello", blankCapture, nil)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.SetBasicAuth("viewer", "hello")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	// Viewer should NOT see the voice control button
	if strings.Contains(rec.Body.String(), "micBtn") {
		t.Error("viewer HTML should not contain voice control button")
	}
}

func TestServerMux_WithAuth_ViewerCannotPostCommand(t *testing.T) {
	mux := newServerMux("admin:secret", "viewer:hello", blankCapture, nil)

	req := httptest.NewRequest(http.MethodPost, "/command", strings.NewReader("stop"))
	req.SetBasicAuth("viewer", "hello")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestServerMux_WithAuth_WrongPassword_Returns401(t *testing.T) {
	mux := newServerMux("admin:secret", "", blankCapture, nil)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.SetBasicAuth("admin", "wrongpassword")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

// --- newServerNoVoiceMux ---

func TestServerNoVoiceMux_NoAuth_Root(t *testing.T) {
	mux := newServerNoVoiceMux("", "", blankCapture)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	// Should never contain voice control
	if strings.Contains(rec.Body.String(), "micBtn") {
		t.Error("no-voice mux should not contain voice control button")
	}
}

func TestServerNoVoiceMux_WithAuth_NoCredentials_StreamReturns401(t *testing.T) {
	mux := newServerNoVoiceMux("admin:secret", "", blankCapture)

	req := httptest.NewRequest(http.MethodGet, "/stream", nil)
	rec := httptest.NewRecorder()
	// Use a cancelled context so the streaming loop exits quickly
	ctx := req.Context()
	_ = ctx
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}
