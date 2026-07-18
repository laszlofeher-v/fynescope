//go:build web
// +build web

package web

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log/slog"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
)

// parseAuth splits a "user:pass" credential string into its components.
func parseAuth(auth string) (user, pass string) {
	if auth == "" {
		return "", ""
	}
	parts := strings.SplitN(auth, ":", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return parts[0], ""
}

// generateSelfSignedCert creates a self-signed TLS certificate in memory.
func generateSelfSignedCert() (tls.Certificate, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Fynescope Web Interface"},
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

// newServerMux builds the HTTP handler mux for the full server (stream + voice control).
// It is intentionally separated from StartServer so it can be exercised directly
// in tests using net/http/httptest without needing a live TLS listener or CLI flags.
func newServerMux(authAdmin, authView string, getCapture func() image.Image, onCommand func(string)) *http.ServeMux {
	adminUser, adminPass := parseAuth(authAdmin)
	viewUser, viewPass := parseAuth(authView)

	// checkAuth returns (authenticated, isAdmin)
	checkAuth := func(r *http.Request) (bool, bool) {
		if adminUser == "" && viewUser == "" {
			return true, true // No auth configured — open access
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

	var (
		latestFrame   []byte
		frameMu       sync.RWMutex
		activeClients atomic.Int32
	)

	// Background goroutine: captures frames only while clients are connected.
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond) // ~10 FPS
		defer ticker.Stop()
		for range ticker.C {
			if activeClients.Load() <= 0 {
				continue
			}
			fyne.Do(func() {
				img := getCapture()
				if img == nil {
					return
				}
				var buf bytes.Buffer
				if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}); err != nil {
					return
				}
				frameMu.Lock()
				latestFrame = buf.Bytes()
				frameMu.Unlock()
			})
		}
	}()

	mux := http.NewServeMux()

	htmlAdmin := `<!DOCTYPE html>
<html>
<head>
	<title>Fynescope</title>
	<style>
		body { margin: 0; background-color: #111; color: white; display: flex; flex-direction: column; align-items: center; height: 100vh; overflow: hidden; font-family: sans-serif; }
		#stream-container { flex-grow: 1; display: flex; justify-content: center; align-items: center; width: 100%; overflow: hidden; }
		img { max-width: 100%; max-height: 100%; object-fit: contain; }
		#controls { padding: 10px; display: flex; gap: 15px; align-items: center; background: #222; width: 100%; justify-content: center; }
		button { padding: 10px 20px; font-size: 16px; cursor: pointer; border: none; border-radius: 5px; background: #007bff; color: white; }
		button.listening { background: #dc3545; animation: pulse 1.5s infinite; }
		@keyframes pulse { 0% { opacity: 1; } 50% { opacity: 0.5; } 100% { opacity: 1; } }
		#status { font-size: 14px; color: #aaa; }
	</style>
</head>
<body>
	<div id="controls">
		<button id="micBtn">Start Voice Control</button>
		<select id="langSelect">
			<option value="en-US" selected>English</option>
			<option value="es-ES">Español</option>
			<option value="fr-FR">Français</option>
			<option value="de-DE">Deutsch</option>
			<option value="hu-HU">Magyar</option>
		</select>
		<div id="status">Voice control inactive. Note: Requires Chrome/Edge. Click to allow microphone.</div>
	</div>
	<div id="stream-container">
		<img src="/stream" alt="Fynescope Stream" />
	</div>
	<script>
		const micBtn = document.getElementById('micBtn');
		const langSelect = document.getElementById('langSelect');
		const status = document.getElementById('status');
		let recognition = null;
		let isListening = false;

		if ('webkitSpeechRecognition' in window || 'SpeechRecognition' in window) {
			const SpeechRecognition = window.SpeechRecognition || window.webkitSpeechRecognition;
			recognition = new SpeechRecognition();
			recognition.continuous = true;
			recognition.interimResults = false;
			recognition.lang = langSelect.value;

			langSelect.onchange = () => {
				recognition.lang = langSelect.value;
				if (isListening) {
					recognition.stop();
				}
			};

			recognition.onstart = () => {
				isListening = true;
				micBtn.textContent = 'Stop Voice Control';
				micBtn.classList.add('listening');
				status.textContent = 'Listening for commands... (e.g. "Run", "Stop", "Enable channel A")';
			};

			recognition.onresult = (event) => {
				const transcript = event.results[event.results.length - 1][0].transcript.trim();
				status.textContent = 'Heard: "' + transcript + '"';
				fetch('/command', {
					method: 'POST',
					body: transcript
				}).catch(err => console.error("Error sending command", err));
			};

			recognition.onerror = (event) => {
				status.textContent = 'Speech recognition error: ' + event.error;
			};

			recognition.onend = () => {
				if (isListening) {
					recognition.start();
				} else {
					micBtn.textContent = 'Start Voice Control';
					micBtn.classList.remove('listening');
					status.textContent = 'Voice control stopped.';
				}
			};

			micBtn.onclick = () => {
				if (isListening) {
					isListening = false;
					recognition.stop();
				} else {
					recognition.start();
				}
			};
		} else {
			micBtn.disabled = true;
			micBtn.style.background = '#555';
			status.textContent = 'Web Speech API not supported in this browser. Please use Chrome or Edge.';
		}
	</script>
</body>
</html>`

	htmlView := `<!DOCTYPE html>
<html>
<head>
	<title>Fynescope (Read-Only)</title>
	<style>
		body { margin: 0; background-color: #111; display: flex; justify-content: center; align-items: center; height: 100vh; overflow: hidden; }
		img { max-width: 100%; max-height: 100%; object-fit: contain; }
	</style>
</head>
<body>
	<img src="/stream" alt="Fynescope Stream" />
</body>
</html>`

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		authOk, isAdmin := checkAuth(r)
		if !authOk {
			w.Header().Set("WWW-Authenticate", `Basic realm="Fynescope"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		if isAdmin {
			w.Write([]byte(htmlAdmin))
		} else {
			w.Write([]byte(htmlView))
		}
	})

	mux.HandleFunc("/command", func(w http.ResponseWriter, r *http.Request) {
		authOk, isAdmin := checkAuth(r)
		if !authOk {
			w.Header().Set("WWW-Authenticate", `Basic realm="Fynescope"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if !isAdmin {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		if r.Method == http.MethodPost {
			body, _ := io.ReadAll(r.Body)
			cmd := string(body)
			slog.Info("Web Voice Command Received", "cmd", cmd)
			if onCommand != nil {
				onCommand(cmd)
			}
		}
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		authOk, _ := checkAuth(r)
		if !authOk {
			w.Header().Set("WWW-Authenticate", `Basic realm="Fynescope"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=frame")
		w.Header().Set("Cache-Control", "no-cache")

		activeClients.Add(1)
		defer activeClients.Add(-1)

		for {
			select {
			case <-r.Context().Done():
				return
			default:
			}
			frameMu.RLock()
			frame := latestFrame
			frameMu.RUnlock()

			if frame != nil {
				_, err := fmt.Fprintf(w, "--frame\r\nContent-Type: image/jpeg\r\nContent-Length: %d\r\n\r\n", len(frame))
				if err != nil {
					return
				}
				if _, err = w.Write(frame); err != nil {
					return
				}
				if _, err = w.Write([]byte("\r\n")); err != nil {
					return
				}
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
	})

	return mux
}

// StartServer launches an HTTPS server providing a read-only MJPEG view of the GUI
// and a voice command interface.
func StartServer(port int, authAdmin, authView string, getCapture func() image.Image, onCommand func(string)) {
	if port <= 0 {
		return
	}
	addr := fmt.Sprintf("0.0.0.0:%d", port)
	slog.Info("Starting HTTPS web server with voice control", "addr", addr)

	cert, err := generateSelfSignedCert()
	if err != nil {
		slog.Error("Failed to generate self-signed certificate", "err", err)
		return
	}
	server := &http.Server{
		Addr:    addr,
		Handler: newServerMux(authAdmin, authView, getCapture, onCommand),
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
	}
	go func() {
		if err := server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTPS Web server error", "err", err)
		}
	}()
}

// newServerNoVoiceMux builds the HTTP handler mux for the read-only (no voice) server.
// It is intentionally separated from StartServerNoVoice so it can be exercised directly
// in tests using net/http/httptest without needing a live TLS listener or CLI flags.
func newServerNoVoiceMux(authAdmin, authView string, getCapture func() image.Image) *http.ServeMux {
	adminUser, adminPass := parseAuth(authAdmin)
	viewUser, viewPass := parseAuth(authView)

	checkAuth := func(r *http.Request) bool {
		if adminUser == "" && viewUser == "" {
			return true
		}
		u, p, ok := r.BasicAuth()
		if !ok {
			return false
		}
		if adminUser != "" && u == adminUser && p == adminPass {
			return true
		}
		if viewUser != "" && u == viewUser && p == viewPass {
			return true
		}
		return false
	}

	var (
		latestFrame   []byte
		frameMu       sync.RWMutex
		activeClients atomic.Int32
	)

	// Background goroutine: captures frames only while clients are connected.
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			if activeClients.Load() <= 0 {
				continue
			}
			fyne.Do(func() {
				img := getCapture()
				if img == nil {
					return
				}
				var buf bytes.Buffer
				if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}); err != nil {
					return
				}
				frameMu.Lock()
				latestFrame = buf.Bytes()
				frameMu.Unlock()
			})
		}
	}()

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		html := `<!DOCTYPE html>
<html>
<head>
	<title>Fynescope (Read-Only)</title>
	<style>
		body { margin: 0; background-color: #111; display: flex; justify-content: center; align-items: center; height: 100vh; overflow: hidden; }
		img { max-width: 100%; max-height: 100%; object-fit: contain; }
	</style>
</head>
<body>
	<img src="/stream" alt="Fynescope Stream" />
</body>
</html>`
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	})

	mux.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		if !checkAuth(r) {
			w.Header().Set("WWW-Authenticate", `Basic realm="Fynescope"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=frame")
		w.Header().Set("Cache-Control", "no-cache")

		activeClients.Add(1)
		defer activeClients.Add(-1)

		for {
			select {
			case <-r.Context().Done():
				return
			default:
			}
			frameMu.RLock()
			frame := latestFrame
			frameMu.RUnlock()

			if frame != nil {
				_, err := fmt.Fprintf(w, "--frame\r\nContent-Type: image/jpeg\r\nContent-Length: %d\r\n\r\n", len(frame))
				if err != nil {
					return
				}
				if _, err = w.Write(frame); err != nil {
					return
				}
				if _, err = w.Write([]byte("\r\n")); err != nil {
					return
				}
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
	})

	return mux
}

// StartServerNoVoice launches an HTTPS server providing a read-only MJPEG view of the GUI without voice control.
func StartServerNoVoice(port int, authAdmin, authView string, getCapture func() image.Image) {
	if port <= 0 {
		return
	}
	addr := fmt.Sprintf("0.0.0.0:%d", port)
	slog.Info("Starting HTTPS web server (read-only, no voice)", "addr", addr)

	cert, err := generateSelfSignedCert()
	if err != nil {
		slog.Error("Failed to generate self-signed certificate", "err", err)
		return
	}
	server := &http.Server{
		Addr:    addr,
		Handler: newServerNoVoiceMux(authAdmin, authView, getCapture),
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
	}
	go func() {
		if err := server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTPS Web server error", "err", err)
		}
	}()
}
