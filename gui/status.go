package gui

import (
	"fynescope/control"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type StatusCode int

const (
	StatusNone StatusCode = iota
	StatusFrequencyCannotBeDetected
	StatusWrongFfTrigger
	StatusChannelNoActiveGen
	StatusCannotZoomOut
	StatusGeneralError
)

type InitStatus struct {
	label      *widget.Label
	code       StatusCode
	statusChan chan statusMessage
	statusQuit chan struct{}
}

func (is *InitStatus) Code() StatusCode {
	return is.code
}

type statusMessage struct {
	text string
	code StatusCode
}

func (scp *ScpDesc) initStatus() {
	scp.status = &InitStatus{
		label:      widget.NewLabel("                                                 "),
		statusChan: make(chan statusMessage, 1),
		statusQuit: make(chan struct{}),
	}
	scp.status.label.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}

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

	scp.psControl.DisplayStatus = func(s string, errorType control.ScopeError) {
		if errorType == control.Fatal {
			if scp.running {
				scp.running = false
				fyne.Do(func() {
					scp.runblockButton.SetIcon(theme.MediaPlayIcon())
				})
			}
		}

		code := StatusGeneralError
		if s == ErrFrequencyCannotBeDetected {
			code = StatusFrequencyCannotBeDetected
		} else if s == ErrWrongFfTrigger {
			code = StatusWrongFfTrigger
		} else if strings.HasPrefix(s, "Error: Channel ") && strings.HasSuffix(s, " has no active generator input") {
			code = StatusChannelNoActiveGen
		} else if s == "Cannot zoom out beyond Time Zoom snapshot" {
			code = StatusCannotZoomOut
		} else if s == "" {
			code = StatusNone
		}

		msg := statusMessage{text: s, code: code}
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
	}
}
