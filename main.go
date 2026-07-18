package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"fynescope/genericps"
	"fynescope/gui"
	_ "fynescope/ps2000a"
	_ "fynescope/ps3000a"
	"fynescope/settings"
	"fynescope/sim"
	"fynescope/web"
	"image"
	"log/slog"
	"os"
	"path"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// settingFileName is the default name for the settings file.
// It will be modified to include device serial number after connection.
var (
	settingFileName = "scopesettings.yaml"
	GitUUID         = ""
	Version         = "0.0.1"
	BuildDate       = "04-06-2026"
)

//go:embed LICENSE
var license string

//go:embed THIRD_PARTY_LICENSES
var thirdPartyLicenses string

// debugOn lists source files that should have debug logging enabled.
// This is used for targeted debugging of specific components.
var (
	debugOn = map[string]bool{
		"gui/adv_trigger_point.go":           false,
		"control/block_mode.go":              false,
		"control/buffers.go":                 false,
		"ps2000a/c.go":                       false,
		"ps3000a/c.go":                       false,
		"ps2000a/callbacks.go":               false,
		"ps3000a/callbacks.go":               false,
		"control/channels.go":                false,
		"gui/channels.go":                    false,
		"checkcolorpick/check_color_pick.go": false,
		"genericps/connection.go":            false,
		"ps2000a/const.go":                   false,
		"ps3000a/const.go":                   false,
		"settings/consts.go":                 false,
		"control/control.go":                 false,
		"gui/dft_channel_label.go":           false,
		"gui/dft_raster.go":                  false,
		"disp7/disp7.go":                     false,
		"control/ets.go":                     false,
		"control/ext_gen.go":                 false,
		"gui/ff_raster.go":                   false,
		"gui/ft_channel_label.go":            false,
		"gui/ft_raster.go":                   false,
		"gui/fv_raster.go":                   false,
		"control/gen.go":                     false,
		"genericps/genericps.go":             false,
		"gui/gui.go":                         false,
		"control/interpolation.go":           false,
		"main/main.go":                       false,
		"gui/measure.go":                     false,
		"genericps/no_scope.go":              false,
		"genericps/open.go":                  false,
		"gui/params.go":                      false,
		"psc/ps_consts.go":                   false,
		"gui/raster.go":                      false,
		"control/scheduler.go":               false,
		"scpi/scpi.go":                       false,
		"gui/scratch.go":                     false,
		"gui/screen_draw.go":                 false,
		"control/screen_time.go":             false,
		"selectscroll/selectscroll.go":       false,
		"settings/settings.go":               false,
		"sim/sim.go":                         false,
		"sim/sim_gen.go":                     false,
		"sliderscroll/slider_scroll.go":      false,
		"gui/status.go":                      false,
		"control/stream.go":                  false,
		"sim/sweep.go":                       false,
		"gui/sync.go":                        false,
		"tastybutton/tasty_button.go":        false,
		"gui/test_proxy.go":                  false,
		"gui/theme.go":                       false,
		"gui/time_div.go":                    false,
		"control/timing.go":                  false,
		"control/trigger.go":                 false,
		"gui/trigger_point.go":               false,
		"ps2000a/types.go":                   false,
		"ps3000a/types.go":                   false,
		"genericps/types.go":                 false,
		"sim/waveforms.go":                   false,
	}
)

// FilterHandler wraps another slog.Handler and filters debug messages based on the source file.
type FilterHandler struct {
	handler slog.Handler
	level   *slog.LevelVar
}

// Enabled returns true if the log level is enabled globally, or if it's LevelDebug (to allow per-file filtering).
func (h *FilterHandler) Enabled(ctx context.Context, level slog.Level) bool {
	if level == slog.LevelDebug {
		return true
	}
	return level >= h.level.Level()
}

// Handle processes the log record, filtering debug messages if they are not from a file in debugOn.
func (h *FilterHandler) Handle(ctx context.Context, r slog.Record) error {
	// If the level is at or above the global level, always log it.
	if r.Level >= h.level.Level() {
		return h.handler.Handle(ctx, r)
	}

	// If the level is below the global level, only log it if it's DEBUG and the file is in debugOn.
	if r.Level == slog.LevelDebug {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		frame, _ := fs.Next()

		pkg := frame.Function
		if lastSlash := strings.LastIndexByte(pkg, '/'); lastSlash != -1 {
			pkg = pkg[lastSlash+1:]
		}
		if firstDot := strings.IndexByte(pkg, '.'); firstDot != -1 {
			pkg = pkg[:firstDot]
		}
		filename := path.Join(pkg, path.Base(frame.File))

		if debugOn[filename] {
			return h.handler.Handle(ctx, r)
		}
	}
	return nil
}

// WithAttrs returns a new FilterHandler with the given attributes.
func (h *FilterHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &FilterHandler{handler: h.handler.WithAttrs(attrs), level: h.level}
}

// WithGroup returns a new FilterHandler with the given group name.
func (h *FilterHandler) WithGroup(name string) slog.Handler {
	return &FilterHandler{handler: h.handler.WithGroup(name), level: h.level}
}

// setLogging configures the application's logging system with custom formatting.
// It sets up a text handler that:
// - Shows only the base filename (not full path) for source locations
// - Formats timestamps as time-only (HH:MM:SS)
// - Outputs to stderr
// The log level is set based on the provided loglevel string parameter.
func setLogging(loglevel *string) {
	// Create a dynamic log level that can be changed at runtime
	programLevel := new(slog.LevelVar)

	// Configure the text handler with custom attribute formatting
	// We set the base handler level to Debug so it doesn't filter records
	// that FilterHandler has already decided to allow.
	baseHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{AddSource: true,
		Level: slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			switch a.Key {
			case slog.SourceKey:
				// Simplify source file paths to just the filename
				s := a.Value.Any().(*slog.Source)
				s.File = path.Base(s.File)
			case slog.TimeKey:
				// Format time as HH:MM:SS instead of full timestamp
				t := a.Value.Time()
				a.Value = slog.StringValue(t.Format(time.TimeOnly))
			}
			return a
		}})

	handler := &FilterHandler{handler: baseHandler, level: programLevel}
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Set the log level based on command-line argument
	switch *loglevel {
	case "debug":
		programLevel.Set(slog.LevelDebug)
	case "info":
		programLevel.Set(slog.LevelInfo)
	case "warning":
		programLevel.Set(slog.LevelWarn)
	case "error":
		programLevel.Set(slog.LevelError)
	default:
		slog.Error("wrong", "loglevel", *loglevel)
		slog.Error("fallback to error level")
		programLevel.Set(slog.LevelError)
	}
}

// startProfile initializes CPU profiling and writes to a numbered profile file.
// The profile file will be named "fynescope_N.prof" where N is the provided number.
// This is useful for performance analysis using "go tool pprof".
func startProfile(n int) error {
	f, err := os.Create("fynescope_" + strconv.Itoa(n) + ".prof")
	if err != nil {
		return err
	}
	slog.Info("profiling is on")
	return pprof.StartCPUProfile(f)
}

// parseFlags processes command-line arguments and returns pointers to flag values.
// Supported flags:
//
//	-loglevel: Sets logging verbosity (debug, info, warning, error)
//	-profile:  Enables CPU profiling when set to true
//	-sim:      Runs in simulator-only mode when set to true
//	-screensize: Sets the screen size scaling (e.g. 1920x1080, 1366x768, 1280x720, 1024x768)
//	-webport:  Starts a read-only MJPEG stream of the GUI on the specified port
func parseFlags() (profile, simulator *bool, logLevel *string, chCount *int, chCountExplicit bool, extGenEnabled bool, screenSize *string, screenSizeExplicit bool, webPort *int, webPortNoVoice *int, webAuth, webAuthView *string) {
	logLevel = flag.String("loglevel", "warning", "-loglevel=info | debug | warning | error")
	profile = flag.Bool("profile", false, "-profile=true")
	simulator = flag.Bool("sim", false, "-sim=true")
	chCount = flag.Int("chcount", sim.DefaultChannels, fmt.Sprintf("-chcount=%d .. %d (simulator only)", sim.MinChannels, sim.MaxChannels))
	about := flag.Bool("about", false, "show version, build date and license")
	webPort = flag.Int("webport", 0, "-webport=8080 (starts web server on specified port, 0 to disable)")
	webPortNoVoice = flag.Int("webport-novoice", 0, "-webport-novoice=8081 (starts web server without voice control on specified port, 0 to disable)")
	webAuth = flag.String("webauth", "", "-webauth=user:pass (credentials for full access, voice + stream)")
	webAuthView = flag.String("webauth-view", "", "-webauth-view=user:pass (credentials for read-only stream access)")
	inTestMode := strings.HasSuffix(os.Args[0], ".test") || strings.Contains(os.Args[0], "/_test/")
	extGenFlag := registerExtGenFlag(inTestMode)
	screenSize = flag.String("screensize", settings.ScreenSize1920x1080, "-screensize=1920x1080 | 1366x768 | 1280x720 | 1024x768")

	flag.Parse()

	if *about {
		fmt.Printf("Version: %s\nBuild Date: %s\n\nGit UUID:%s\n\nLicense:\n%s\n\nThird-Party Licenses:\n%s\n", Version, BuildDate, GitUUID, license, thirdPartyLicenses)
		os.Exit(0)
	}

	flag.Visit(func(f *flag.Flag) {
		if f.Name == "chcount" {
			chCountExplicit = true
		}
		if f.Name == "screensize" {
			screenSizeExplicit = true
		}
	})

	if *chCount < sim.MinChannels || *chCount > sim.MaxChannels {
		fmt.Fprintf(os.Stderr, "invalid channel count %d: must be between %d and %d\n", *chCount, sim.MinChannels, sim.MaxChannels)
		flag.Usage()
		os.Exit(1)
	}

	extGenEnabled = *extGenFlag
	return
}

// connectToDevice opens a connection to the specified device.
// It handles both real hardware devices and simulator connections.
// For simulators, only the device ID is needed.
// For real devices, both ID and serial number are required.
// Returns a Connection object or an error if the connection fails.
func connectToDevice(device *genericps.DeviceInfo) (*genericps.Connection, error) {
	con := genericps.NewConnection()
	var err error

	// Choose connection method based on device type
	if device.IsSimulator {
		con.Handle, err = genericps.OpenSimulator(con, device.Id)
	} else {
		con.Handle, err = genericps.OpenUnit(con, device.Id, device.Serial)
	}
	con.ID = device.Id
	if err != nil {
		return nil, fmt.Errorf("failed to open device: %w", err)
	}

	return con, nil
}

// openSimulator is a helper function for tests to open a simulator connection.
func openSimulator(id string) (*genericps.Connection, error) {
	return connectToDevice(&genericps.DeviceInfo{Id: id, IsSimulator: true})
}

// setupSettingsFile configures the settings filename based on device info.
// It retrieves the device's batch and serial number and uses it to create
// a unique settings filename for this specific device.
// Forward slashes in the serial are replaced with underscores to ensure
// the filename is valid across all operating systems.
func setupSettingsFile(con *genericps.Connection) error {
	info, err := con.GetUnitInfo(genericps.PicoBatchAndSerial)
	if err != nil {
		return err
	}

	// Sanitize the serial number for use in filename
	info = strings.ReplaceAll(info, "/", "_")
	settingFileName = "scopesettings_" + info + ".yaml"
	return nil
}

// initializeAndRunApp loads settings and runs the main application menu.
// This function:
// 1. Loads saved settings from the YAML file (or creates defaults if not found)
// 2. Launches the main GUI menu with the device connection and settings
// The Menu function will block until the user closes the application.
func initializeAndRunApp(con *genericps.Connection, scp *gui.ScpDesc, explicitScreenSize *string, isScreenSizeExplicit bool) error {
	var err error
	scp.Settings, err = settings.Load(settingFileName)
	if err != nil {
		slog.Error("failed to load settings, using defaults", "err", err)
		scp.Settings = settings.NewDefaultSettings()
	}

	if isScreenSizeExplicit {
		scp.Settings.ScreenSize = *explicitScreenSize
		if err := settings.Save(settingFileName, scp.Settings); err != nil {
			slog.Error("failed to save new default screen size", "err", err)
		}
	}

	return scp.Menu(con, scp.Settings, settingFileName)
}

// showDeviceSelectionDialog displays a dialog for selecting a device.
// This function creates a GUI window with a radio button list of available devices.
// The user can select a device and click "Select" to connect, or "Cancel" to exit.
// After device selection:
// 1. Connects to the chosen device
// 2. Sets up a device-specific settings file
// 3. Loads settings and launches the main application
// 4. On exit, closes the connection and saves settings
func showDeviceSelectionDialog(scp *gui.ScpDesc, devices []genericps.DeviceInfo, explicitScreenSize *string, isScreenSizeExplicit bool) error {
	w := scp.App.NewWindow("Select Device")
	w.Resize(fyne.NewSize(400, 300))

	// Build display strings for each device
	options := make([]string, len(devices))
	for i, dev := range devices {
		if dev.IsSimulator {
			options[i] = "Simulator"
		} else {
			options[i] = dev.Id + " - " + dev.Serial
		}
	}

	// Track which device is currently selected
	var selectedIndex int
	var radio *widget.RadioGroup
	radio = widget.NewRadioGroup(options, func(value string) {
		if value == "" {
			// User clicked the already-selected item — revert to keep it selected
			radio.SetSelected(options[selectedIndex])
			return
		}
		// Update selectedIndex when user changes selection
		for i, opt := range options {
			if opt == value {
				selectedIndex = i
				break
			}
		}
	})
	radio.SetSelected(options[0]) // Default to first device

	// Connection will be stored here after successful device selection
	var con *genericps.Connection

	var selectButton *widget.Button
	var cancelButton *widget.Button

	selectButton = widget.NewButton("Select", func() {
		selectButton.Disable()
		cancelButton.Disable()
		go func() {
			selectedDevice := &devices[selectedIndex]
			var err error

			// Attempt to connect to the selected device
			con, err = connectToDevice(selectedDevice)
			if err != nil {
				slog.Error("failed to connect to device", "err", err)
				fyne.Do(func() {
					selectButton.Enable()
					cancelButton.Enable()
				})
				return
			}

			// Create device-specific settings filename
			if err := setupSettingsFile(con); err != nil {
				slog.Warn("failed to setup settings file", "err", err)
			}

			fyne.Do(func() {
				// Launch the main application
				if err := initializeAndRunApp(con, scp, explicitScreenSize, isScreenSizeExplicit); err != nil {
					slog.Error("Menu", "err", err)
					selectButton.Enable()
					cancelButton.Enable()
					return
				}

				w.Hide()
			})
		}()
	})
	selectButton.Importance = widget.HighImportance

	cancelButton = widget.NewButton("Cancel", func() {
		w.Close()
	})

	// Layout: label at top, buttons at bottom, scrollable radio list in center
	content := container.NewBorder(
		widget.NewLabel("Please select a device"),
		container.NewHBox(cancelButton, selectButton),
		nil,
		nil,
		container.NewVScroll(radio),
	)

	w.SetContent(content)
	if len(devices) == 1 { // Obvious choice
		selectButton.OnTapped()
	} else {
		w.Show()
	}
	scp.App.Run() // Blocks until window is closed

	// Cleanup: close connection and save settings if a device was connected
	if con != nil {
		con.CloseUnit()
		if err := settings.Save(settingFileName, scp.Settings); err != nil {
			slog.Error("failed to save settings", "err", err)
		}
	}
	slog.Info("Unit closed")
	return nil
}

// main is the entry point for the PicoScope GUI application.
// Application flow:
// 1. Parse command-line flags (log level, profiling, simulator mode)
// 2. Configure logging system
// 3. Optionally start CPU profiling
// 4. Initialize Fyne GUI application
// 5. Enumerate available devices (or use simulator only)
// 6. Show device selection dialog
// 7. User selects device, connects, and uses the application
// 8. On exit, cleanup and save settings
func main() {
	var (
		devices []genericps.DeviceInfo
		err     error
	)

	// Process command-line arguments
	profile, simulatorOnly, logLevel, chCount, chCountExplicit, extGenEnabled, explicitScreenSize, isScreenSizeExplicit, webPort, webPortNoVoice, webAuth, webAuthView := parseFlags()
	setLogging(logLevel)

	err = sim.SetChannelCount(*chCount, chCountExplicit)

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
		ExtGenEnabled: extGenEnabled,
	}
	scp.App = app.New()

	if *webPort > 0 {
		web.StartServer(*webPort, *webAuth, *webAuthView, func() image.Image {
			if scp.Window == nil || scp.Window.Canvas() == nil {
				return nil
			}
			return scp.Window.Canvas().Capture()
		}, scp.ExecuteVoiceCommand)
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
	if *simulatorOnly {
		// Simulator mode: enumerate devices but filter for only simulators
		// This ensures we get the correct serial number (e.g., SIM/CH2)
		allDevices, _ := genericps.EnumerateAllDevices(256)
		for _, dev := range allDevices {
			if dev.IsSimulator {
				devices = append(devices, dev)
			}
		}
		if len(devices) == 0 {
			// Fallback if enumeration failed
			devices = []genericps.DeviceInfo{
				{
					Id:          genericps.SimId,
					Serial:      "",
					IsSimulator: true,
				},
			}
		}
	} else {
		// Normal mode: enumerate all connected PicoScope devices
		devices, err = genericps.EnumerateAllDevices(256)
		if err != nil {
			slog.Error("no devices found", "err", err)
			return
		}
	}

	// Show device selection dialog and run the application
	if err = showDeviceSelectionDialog(scp, devices, explicitScreenSize, isScreenSizeExplicit); err != nil {
		slog.Error("device selection", "err", err)
	}
}
