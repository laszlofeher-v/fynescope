//go:build !remote

package main

import (
	"flag"
	"fmt"
	"image"
	"log/slog"
	"os"
	"runtime/pprof"

	"fyne.io/fyne/v2/app"
	"fynescope/demo"
	"fynescope/genericps"
	"fynescope/gui"
	"fynescope/web"
)

// main is the entry point for the PicoScope GUI application.
// Application flow:
// 1. Parse command-line flags (log level, profiling, demo mode)
// 2. Configure logging system
// 3. Optionally start CPU profiling
// 4. Initialize Fyne GUI application
// 5. Enumerate available devices (or use demo only)
// 6. Show device selection dialog
// 7. User selects device, connects, and uses the application
// 8. On exit, cleanup and save settings
func main() {
	var (
		devices []genericps.DeviceInfo
		err     error
	)

	// Process command-line arguments
	profile, demoOnly, logLevel, chCount, chCountExplicit, extGenEnabled, multiEnabled, explicitScreenSize, isScreenSizeExplicit, webPort, webPortNoVoice, webAuth, webAuthView, gifEnabled, ffAutoRange, simName, apiPort, apiAuth, apiAuthView := parseFlags()
	setLogging(logLevel)

	err = demo.SetChannelCount(*chCount, chCountExplicit)

	// Start CPU profiling if requested
	if *profile {
		if err := startProfile(0); err != nil {
			slog.Error("unable to start profiling", "err", err)
		}
		slog.Info("profiling is on", "open result", "go tool pprof fynescope fynescope.prof")
		defer pprof.StopCPUProfile()
	}

	// Initialize the GUI application
	scp := &gui.ScpDesc{
		ExtGenEnabled:      extGenEnabled,
		MultiEnabled:       multiEnabled,
		GifEnabled:         *gifEnabled,
		FfAutoRangeEnabled: *ffAutoRange,
	}
	scp.App = app.New()

	if *apiPort > 0 {
		if err := scp.StartAPIServer(gui.APIServerConfig{
			Port:      *apiPort,
			AuthAdmin: *apiAuth,
			AuthView:  *apiAuthView,
		}); err != nil {
			slog.Error("Failed to start remote HTTPS API server", "err", err)
		}
	}

	if *webPort > 0 {
		web.StartServer(*webPort, *webAuth, *webAuthView, func() image.Image {
			if scp.Window == nil || scp.Window.Canvas() == nil {
				return nil
			}
			return scp.Window.Canvas().Capture()
		}, scp)
	}

	if *webPortNoVoice > 0 {
		web.StartServerNoVoice(*webPortNoVoice, *webAuth, *webAuthView, func() image.Image {
			if scp.Window == nil || scp.Window.Canvas() == nil {
				return nil
			}
			return scp.Window.Canvas().Capture()
		})
	}

	// Determine which devices to show in the selection dialog
	if *simName != "" && *demoOnly {
		fmt.Fprintf(os.Stderr, "cannot specify both -sim and -demo=true\n")
		flag.Usage()
		os.Exit(1)
	}

	if *demoOnly {
		// Demo mode: enumerate devices but filter for only demo devices
		// This ensures we get the correct serial number (e.g., SIM/CH2)
		allDevices, _ := genericps.EnumerateAllDevices(256)
		for _, dev := range allDevices {
			if dev.IsDemo {
				devices = append(devices, dev)
			}
		}
		if len(devices) == 0 {
			// Fallback if enumeration failed
			devices = []genericps.DeviceInfo{
				{
					Id:     genericps.DemoId,
					Serial: "",
					IsDemo: true,
				},
			}
		}
	} else if *simName != "" {
		// Simulator mode: use the ps2000a simulator
		demo.ScopeSimVariantInfo = *simName + "SIM"
		devices = []genericps.DeviceInfo{
			{
				Id:     "ps2000a",
				Serial: *simName + "SIM",
				IsDemo: false,
			},
		}
	} else {
		// Normal mode: enumerate all connected PicoScope devices
		devices, err = genericps.EnumerateAllDevices(256)
		if err != nil {
			slog.Warn("no devices found", "err", err)
			return
		}
	}

	// Show device selection dialog and run the application
	if err = showDeviceSelectionDialog(scp, devices, explicitScreenSize, isScreenSizeExplicit); err != nil {
		slog.Error("device selection", "err", err)
	}
}
