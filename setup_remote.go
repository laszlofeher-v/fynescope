//go:build remote

// The remote build tag produces a GUI-only binary that connects to a running
// fynescope-server over TCP.  Hardware access happens on the server; this
// binary only renders the oscilloscope UI.
//
// Build with:
//
//	go build -tags remote -o fynescope-remote .
//
// Run with:
//
//	./fynescope-remote -server 192.168.1.10:9876

package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"fyne.io/fyne/v2/app"
	"fynescope/genericps"
	"fynescope/gui"
	"fynescope/netcontrol"
	"fynescope/settings"
)

func main() {
	serverAddr := flag.String("server", "", "-server=host:port  address of the fynescope-server (required)")
	logLevel := flag.String("loglevel", "warning", "-loglevel=info | debug | warning | error")
	screenSize := flag.String("screensize", settings.ScreenSize1920x1080, "-screensize=1920x1080 | 1366x768 | 1280x720 | 1024x768")
	flag.Parse()

	if *serverAddr == "" {
		fmt.Fprintln(os.Stderr, "error: -server flag is required")
		flag.Usage()
		os.Exit(1)
	}

	setLogging(logLevel)

	// Connect to the remote fynescope-server.
	slog.Info("connecting to fynescope-server", "addr", *serverAddr)
	client, err := netcontrol.Dial(*serverAddr)
	if err != nil {
		slog.Error("failed to connect to server", "addr", *serverAddr, "err", err)
		os.Exit(1)
	}
	defer client.Close()
	slog.Info("connected to fynescope-server", "addr", *serverAddr)

	// Initialize default genericps constants for remote GUI.
	genericps.LoadDefaultConstants()

	// Initialize the Fyne application.
	scp := &gui.ScpDesc{}
	scp.App = app.New()

	// Load settings (use device serial if available, matching local behavior).
	activeSettingFileName := settingFileName
	if serial, err := client.UnitBatchAndSerialInfo(); err == nil && serial != "" {
		serialClean := strings.ReplaceAll(serial, "/", "_")
		activeSettingFileName = "scopesettings_" + serialClean + ".yaml"
		slog.Info("using device settings file", "file", activeSettingFileName)
	}

	cfg, err := settings.Load(activeSettingFileName)
	if err != nil {
		slog.Warn("failed to load settings, using defaults", "file", activeSettingFileName, "err", err)
		cfg = settings.NewDefaultSettings()
	}
	cfg.ScreenSize = *screenSize

	// Launch the main GUI using the remote TCP controller.
	if err := scp.MenuWithController(client, cfg, activeSettingFileName); err != nil {
		slog.Error("remote GUI init failed", "err", err)
		os.Exit(1)
	}

	scp.App.Run()

	// Save settings on exit.
	if err := settings.Save(activeSettingFileName, scp.Settings); err != nil {
		slog.Error("failed to save settings", "file", activeSettingFileName, "err", err)
	}
}
