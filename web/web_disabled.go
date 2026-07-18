//go:build !web
// +build !web

package web

import (
	"image"
	"log/slog"
)

// StartServer is a stub for when the web tag is not provided during compilation.
func StartServer(port int, authAdmin, authView string, getCapture func() image.Image, onCommand func(string)) {
	slog.Warn("Web server requested, but 'web' build tag was not provided during compilation. Web server will not start.")
}

// StartServerNoVoice is a stub for when the web tag is not provided during compilation.
func StartServerNoVoice(port int, authAdmin, authView string, getCapture func() image.Image) {
	slog.Warn("Web server requested, but 'web' build tag was not provided during compilation. Web server will not start.")
}
