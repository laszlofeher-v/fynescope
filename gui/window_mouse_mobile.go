//go:build android || ios || mobile

package gui

import "fyne.io/fyne/v2"

// onMouseLeftWindow is a no-op on mobile platforms.
func (scp *ScpDesc) onMouseLeftWindow() {
}

// hookWindowMouseLeave is a no-op on mobile platforms where GLFW is not used.
func (scp *ScpDesc) hookWindowMouseLeave(w fyne.Window) {
}

// hookAllWindows is a no-op on mobile platforms where GLFW is not used.
func (scp *ScpDesc) hookAllWindows() {
}

// isMouseInsideAppWindow always returns true on mobile platforms.
func (scp *ScpDesc) isMouseInsideAppWindow() bool {
	return true
}

