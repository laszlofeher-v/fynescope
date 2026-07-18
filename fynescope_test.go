// DO NOT run test on real hardware
package main

import (
	"fynescope/genericps"
	"fynescope/gui"
	"fynescope/settings"
	"log"
	"os"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

var (
	scp *gui.ScpDesc
)

func TestMain(m *testing.M) {
	s := "debug"
	setLogging(&s)

	log.SetFlags(log.Ltime | log.Lshortfile)

	con, err := openSimulator(genericps.SimId)
	if err != nil {
		log.Fatalf("failed to open simulator: %v", err)
	}

	// Ensure tests start with a clean state by removing the existing settings file
	_ = os.Remove(settingFileName)
	cfg, err := settings.Load(settingFileName)
	if err != nil {
		cfg = settings.NewDefaultSettings()
	}
	defer func() {
		_ = settings.Save(settingFileName, cfg)
	}()

	scp = &gui.ScpDesc{ExtGenEnabled: true}
	scp.App = app.New()

	if err := scp.Menu(con, cfg, settingFileName); err != nil {
		log.Printf("Menu returned error: %v", err)
	}

	var exitCode int
	go func() {
		time.Sleep(2 * time.Second)
		exitCode = m.Run()
		fyne.Do(func() {
			scp.App.Quit()
		})
	}()

	scp.App.Run()

	if err := con.CloseUnit(); err != nil {
		log.Printf("CloseUnit error: %v", err)
	}

	os.Exit(exitCode)
}

func TestSmoke(t *testing.T) {
	if scp == nil {
		t.Fatal("scp is nil — app failed to initialize")
	}
}

func TestGui(t *testing.T) {
	if scp == nil {
		t.Fatal("scp is nil — app failed to initialize")
	}
	scp.Test()
}

// Test0 runs the GUI fuzzer for the duration set by the -timeout flag.
// Run with: go test -tags=noscope -tags=testsw -v -run Test0 -timeout 105m
func Test0(t *testing.T) {
	if deadline, ok := t.Deadline(); ok {
		if time.Until(deadline) < 20*time.Minute {
			t.Skip("Skipping fuzzer test")
		}
	}
	if scp == nil {
		t.Fatal("scp is nil — app failed to initialize")
	}
	var timeout time.Duration
	if deadline, ok := t.Deadline(); ok {
		log.Printf("deadline: %v", deadline)
		timeout = time.Until(deadline) - 10*time.Second
		if timeout < 0 {
			timeout = 0
		}
	}
	log.Printf("timeout: %v", timeout)
	scp.Random(timeout)
}
