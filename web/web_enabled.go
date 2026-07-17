//go:build web
// +build web

package web

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"fyne.io/fyne/v2"
)

// StartServer launches an HTTP server providing a read-only MJPEG view of the GUI.
// The getCapture function must return the current canvas image; it will only be
// called from the Fyne main thread via fyne.Do.
func StartServer(port int, getCapture func() image.Image) {
	if port <= 0 {
		return
	}

	addr := fmt.Sprintf("0.0.0.0:%d", port)
	slog.Info("Starting read-only web server", "addr", addr)

	// latestFrame holds the most recently captured JPEG frame.
	// A mutex protects concurrent access between the capture goroutine
	// and HTTP handler goroutines.
	var (
		latestFrame []byte
		frameMu     sync.RWMutex
	)

	// Background goroutine: periodically captures frames on the Fyne main thread
	// and pre-encodes them to JPEG. This avoids calling Canvas().Capture() from
	// HTTP handler goroutines which would race with the Fyne render loop.
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond) // ~10 FPS
		defer ticker.Stop()
		for range ticker.C {
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

	// Root handler serves the HTML page
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

	// Stream handler serves MJPEG from the pre-captured frame buffer
	mux.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=frame")
		w.Header().Set("Cache-Control", "no-cache")

		for {
			// Check if client disconnected
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

	go func() {
		if err := http.ListenAndServe(addr, mux); err != nil {
			slog.Error("Web server error", "err", err)
		}
	}()
}
