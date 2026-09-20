package gui

import (
	"reflect"
	"sync"
	"unsafe"

	"fyne.io/fyne/v2"
	"github.com/go-gl/glfw/v3.4/glfw"
)

var (
	hookedWindows   = make(map[*glfw.Window]bool)
	hookedWindowsMu sync.Mutex
)

// getGLFWWindow extracts the underlying *glfw.Window from a fyne.Window if present.
func getGLFWWindow(w fyne.Window) *glfw.Window {
	if w == nil {
		return nil
	}
	val := reflect.ValueOf(w)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return nil
	}
	elem := val.Elem()
	if elem.Kind() != reflect.Struct {
		return nil
	}
	field, ok := elem.Type().FieldByName("viewport")
	if !ok || field.Type != reflect.TypeOf((*glfw.Window)(nil)) {
		return nil
	}
	ptr := unsafe.Pointer(uintptr(val.UnsafePointer()) + field.Offset)
	return *(**glfw.Window)(ptr)
}

// onMouseLeftWindow removes focus from all windows and immediately dismisses
// any visible contextual help popup or card.
func (scp *ScpDesc) onMouseLeftWindow() {
	if scp == nil {
		return
	}
	if scp.Window != nil && scp.Window.Canvas() != nil {
		scp.Window.Canvas().Unfocus()
	}
	if scp.App != nil && scp.App.Driver() != nil {
		for _, w := range scp.App.Driver().AllWindows() {
			if w != nil && w.Canvas() != nil {
				w.Canvas().Unfocus()
			}
		}
	}
	scp.currentFocused = nil
}

// hookWindowMouseLeave hooks GLFW's cursor enter/leave callback to detect
// when the mouse exits the window, even if the movement is very fast.
func (scp *ScpDesc) hookWindowMouseLeave(w fyne.Window) {
	gw := getGLFWWindow(w)
	if gw == nil {
		return
	}
	hookedWindowsMu.Lock()
	defer hookedWindowsMu.Unlock()
	if hookedWindows[gw] {
		return
	}
	hookedWindows[gw] = true

	gw.SetCursorEnterCallback(func(_ *glfw.Window, entered bool) {
		if !entered {
			fyne.Do(func() {
				scp.onMouseLeftWindow()
			})
		}
	})
}

// hookAllWindows hooks cursor leave events on all active windows.
func (scp *ScpDesc) hookAllWindows() {
	if scp == nil {
		return
	}
	if scp.Window != nil {
		scp.hookWindowMouseLeave(scp.Window)
	}
	if scp.App != nil && scp.App.Driver() != nil {
		for _, w := range scp.App.Driver().AllWindows() {
			scp.hookWindowMouseLeave(w)
		}
	}
}

func isCursorInGLFWWindow(gw *glfw.Window) bool {
	if gw == nil {
		return false
	}
	xpos, ypos := gw.GetCursorPos()
	w, h := gw.GetSize()
	if w <= 0 || h <= 0 {
		return false
	}
	return xpos >= 0 && ypos >= 0 && xpos < float64(w) && ypos < float64(h)
}

// isMouseInsideAppWindow checks if the mouse cursor is currently positioned
// inside any active GLFW window of the application.
func (scp *ScpDesc) isMouseInsideAppWindow() bool {
	if scp == nil {
		return true
	}
	scp.hookAllWindows()

	var checkedAny bool
	if scp.Window != nil {
		gw := getGLFWWindow(scp.Window)
		if gw != nil {
			checkedAny = true
			if isCursorInGLFWWindow(gw) {
				return true
			}
		}
	}
	if scp.App != nil && scp.App.Driver() != nil {
		for _, w := range scp.App.Driver().AllWindows() {
			if w == scp.Window {
				continue
			}
			gw := getGLFWWindow(w)
			if gw != nil {
				checkedAny = true
				if isCursorInGLFWWindow(gw) {
					return true
				}
			}
		}
	}
	if !checkedAny {
		// Non-GLFW driver (e.g. mock test environment)
		return true
	}
	return false
}
