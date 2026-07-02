package gui

import (
	"fynescope/control"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (scp *ScpDesc) initStatus() {
	scp.status = widget.NewLabel("                                                 ")
	scp.status.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	scp.statusChan = make(chan string, 1)
	scp.statusQuit = make(chan struct{})
	go func() {
		var (
			s         string
			count     int
			afterTime time.Duration
		)
		afterTime = time.Second
		for {
			select {
			case <-scp.statusQuit:
				return
			case s = <-scp.statusChan:
			}
			count = 0
			for count < errorDisplayTime {
				select {
				case <-scp.statusQuit:
					return
				case s = <-scp.statusChan:
					count = 0
					afterTime = time.Second
					scp.status.Text = s
					fyne.DoAndWait(scp.status.Refresh)
					scp.status.Importance = widget.DangerImportance
				case <-time.After(afterTime):
					if s != "" {
						count++
						scp.status.Importance = widget.SuccessImportance
						s = " " + s[:len(s)-1]
						scp.status.Text = s
						fyne.DoAndWait(scp.status.Refresh)
					}
				}
			}
			scp.status.Text = ""
			fyne.DoAndWait(scp.status.Refresh)
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
		scp.statusChan <- s
	}
}
