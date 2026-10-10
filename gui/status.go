package gui

import (
	"fynescope/control"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// StatusCode identifies categories of oscilloscope runtime status and error conditions.
type StatusCode int

const (
	// StatusNone indicates no active error or message.
	StatusNone StatusCode = iota
	// StatusFrequencyCannotBeDetected indicates the hardware could not lock onto or detect the signal frequency.
	StatusFrequencyCannotBeDetected
	// StatusWrongFfTrigger indicates an incompatible trigger configuration for fast-frequency mode.
	StatusWrongFfTrigger
	// StatusChannelNoActiveGen indicates the requested channel has no connected signal generator source.
	StatusChannelNoActiveGen
	// StatusCannotZoomOut indicates time-zoom reached the maximum outer snapshot window bounds.
	StatusCannotZoomOut
	// StatusGeneralError indicates an unclassified general or fatal scope error.
	StatusGeneralError
)

// InitStatus manages the GUI status bar display label, current status code,
// and goroutine channels for displaying and fading out status notifications.
type InitStatus struct {
	label      *widget.Label
	code       StatusCode
	statusChan chan statusMessage
	statusQuit chan struct{}
}

// Code returns the current active StatusCode.
func (is *InitStatus) Code() StatusCode {
	return is.code
}

// statusMessage encapsulates a status bar update message and its associated status code.
type statusMessage struct {
	text string
	code StatusCode
}

// initStatus initializes the status bar widget and background animation loop on ScpDesc.
// It also binds the scope controller's SetDisplayStatus callback to receive error events
// from the hardware acquisition engine.
func (scp *ScpDesc) initStatus() {
	scp.status = &InitStatus{
		label:      widget.NewLabel("                                                 "),
		statusChan: make(chan statusMessage, 1),
		statusQuit: make(chan struct{}),
	}
	scp.status.label.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}

	// Background goroutine that receives status messages, renders them with Danger importance,
	// and gradually fades/clears the message after errorDisplayTime seconds.
	go func() {
		var (
			msg       statusMessage
			count     int
			afterTime time.Duration
		)
		afterTime = time.Second
		for {
			select {
			case <-scp.status.statusQuit:
				return
			case msg = <-scp.status.statusChan:
			}
			count = 0
			for count < errorDisplayTime {
				select {
				case <-scp.status.statusQuit:
					return
				case msg = <-scp.status.statusChan:
					count = 0
					afterTime = time.Second
					scp.status.label.Text = msg.text
					scp.status.code = msg.code
					fyne.Do(scp.status.label.Refresh)
					scp.status.label.Importance = widget.DangerImportance
				case <-time.After(afterTime):
					if msg.text != "" {
						count++
						scp.status.label.Importance = widget.SuccessImportance
						// Trim trailing character to create a disappearing text animation
						msg.text = " " + msg.text[:len(msg.text)-1]
						scp.status.label.Text = msg.text
						fyne.Do(scp.status.label.Refresh)
					}
				}
			}
			scp.status.label.Text = ""
			scp.status.code = StatusNone
			fyne.Do(scp.status.label.Refresh)
		}
	}()

	// Map known error message strings to their corresponding StatusCode values
	statusMap := map[string]StatusCode{
		ErrFrequencyCannotBeDetected:                StatusFrequencyCannotBeDetected,
		ErrWrongFfTrigger:                           StatusWrongFfTrigger,
		"Cannot zoom out beyond Time Zoom snapshot": StatusCannotZoomOut,
		"": StatusNone,
	}

	// Register callback with psControl to receive status and error messages from acquisition engine
	scp.psControl.SetDisplayStatus(func(s string, errorType control.ScopeError) {
		if errorType == control.Fatal {
			if scp.running {
				scp.running = false
				fyne.Do(func() {
					scp.runblockButton.SetIcon(theme.MediaPlayIcon())
				})
			}
		}

		code := StatusGeneralError
		if val, ok := statusMap[s]; ok {
			code = val
		} else if strings.HasPrefix(s, "Error: Channel ") && strings.HasSuffix(s,
			" has no active generator input") {
			code = StatusChannelNoActiveGen
		}

		dispText := s
		runes := []rune(dispText)
		if len(runes) > 60 {
			dispText = string(runes[:57]) + "..."
		}
		msg := statusMessage{text: dispText, code: code}
		// Send message non-blockingly, replacing any unread message in channel buffer
		select {
		case scp.status.statusChan <- msg:
		default:
			// If full, drain the old message and send the new one
			select {
			case <-scp.status.statusChan:
			default:
			}
			select {
			case scp.status.statusChan <- msg:
			default:
			}
		}
	})
}
