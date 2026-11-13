// Package main implements the Fynescope Android client, a landscape-oriented
// mobile GUI with vertically scrollable control tabs that connects to a
// headless fynescope-server over TCP.
package main

import (
	"flag"
	"fmt"
	"fynescope/genericps"
	"fynescope/gui"
	"fynescope/netcontrol"
	"fynescope/settings"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	appID                  = "com.github.laszlofeher.fynescope.android"
	defaultServerAddr      = "192.168.1.10:9876"
	defaultServerPort      = "9876"
	emulatorHostAddr       = "10.0.2.2:9876"
	localLoopbackAddr      = "127.0.0.1:9876"
	prefLastServer         = "last_server"
	prefRecentServers      = "recent_servers"
	prefScreenSize         = "mobile_screen_size"
	prefLeftControl        = "mobile_left_control"
	maxRecentServers       = 5
	settingFileSuffix      = ".settings.json"
	defaultSettingFileName = "settings.json"
)

// normalizeServerAddress trims whitespace and appends defaultServerPort (9876)
// when no port is provided.
func normalizeServerAddress(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", fmt.Errorf("server address cannot be empty")
	}
	host, port, err := net.SplitHostPort(s)
	if err != nil {
		// If no port was specified, append the default port and re-validate.
		trimmed := strings.Trim(s, "[]")
		if trimmed == "" {
			return "", fmt.Errorf("invalid server host %q", raw)
		}
		candidate := net.JoinHostPort(trimmed, defaultServerPort)
		h, p, err2 := net.SplitHostPort(candidate)
		if err2 != nil || h == "" || p == "" {
			return "", fmt.Errorf("invalid server address %q: %w", raw, err)
		}
		return candidate, nil
	}
	if host == "" {
		return "", fmt.Errorf("server host cannot be empty")
	}
	if port == "" {
		return net.JoinHostPort(host, defaultServerPort), nil
	}
	return net.JoinHostPort(host, port), nil
}

// loadRecentServers reads the saved list of recent server addresses from preferences.
func loadRecentServers(prefs fyne.Preferences) []string {
	if prefs == nil {
		return nil
	}
	raw := prefs.StringWithFallback(prefRecentServers, "")
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, "\n")
	out := make([]string, 0, len(parts))
	seen := make(map[string]bool, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" || seen[p] {
			continue
		}
		seen[p] = true
		out = append(out, p)
		if len(out) >= maxRecentServers {
			break
		}
	}
	return out
}

// saveRecentServer records addr as the most recent server in preferences and
// returns the updated list of recent servers.
func saveRecentServer(prefs fyne.Preferences, addr string) []string {
	addr = strings.TrimSpace(addr)
	if prefs == nil || addr == "" {
		return nil
	}
	prefs.SetString(prefLastServer, addr)
	existing := loadRecentServers(prefs)
	updated := make([]string, 0, maxRecentServers)
	updated = append(updated, addr)
	for _, item := range existing {
		if item == addr {
			continue
		}
		updated = append(updated, item)
		if len(updated) >= maxRecentServers {
			break
		}
	}
	prefs.SetString(prefRecentServers, strings.Join(updated, "\n"))
	return updated
}

// resolveSettingsPath returns a writable path for settings persistence, using
// the Fyne app storage root when available (required on Android where CWD is read-only).
func resolveSettingsPath(a fyne.App, batchAndSerial string) string {
	sanitized := strings.TrimSpace(batchAndSerial)
	sanitized = strings.ReplaceAll(sanitized, "/", "_")
	sanitized = strings.ReplaceAll(sanitized, "\\", "_")
	fileName := defaultSettingFileName
	if sanitized != "" {
		fileName = sanitized + settingFileSuffix
	}
	if a != nil && a.Storage() != nil && a.Storage().RootURI() != nil {
		rootPath := a.Storage().RootURI().Path()
		if rootPath != "" {
			return filepath.Join(rootPath, fileName)
		}
	}
	return fileName
}

// clientApp manages the Android client lifecycle, connection screen, and active
// oscilloscope session.
type clientApp struct {
	app         fyne.App
	window      fyne.Window
	initialAddr string
	autoConnect bool
	dialFunc    func(string) (*netcontrol.Client, error)
	activeScp   *gui.ScpDesc
	activeCli   *netcontrol.Client
}

func newClientApp(a fyne.App, w fyne.Window, initialAddr string, autoConnect bool) *clientApp {
	return &clientApp{
		app:         a,
		window:      w,
		initialAddr: initialAddr,
		autoConnect: autoConnect,
		dialFunc:    netcontrol.Dial,
	}
}

// showConnectScreen displays the landscape server connection UI inside the master window.
func (c *clientApp) showConnectScreen(statusMsg string, isErr bool) {
	c.window.SetTitle("Fynescope Android Client")
	c.window.SetContent(c.buildConnectUI(statusMsg, isErr))
}

// buildConnectUI builds a landscape-friendly, vertically scrollable connection screen
// for entering or selecting a headless fynescope-server address and display options.
func (c *clientApp) buildConnectUI(statusMsg string, isErr bool) fyne.CanvasObject {
	prefs := c.app.Preferences()

	addrEntry := widget.NewEntry()
	addrEntry.SetPlaceHolder(defaultServerAddr)
	defaultAddr := prefs.StringWithFallback(prefLastServer, defaultServerAddr)
	if strings.TrimSpace(c.initialAddr) != "" {
		defaultAddr = strings.TrimSpace(c.initialAddr)
	}
	addrEntry.SetText(defaultAddr)

	statusLabel := widget.NewLabel(statusMsg)
	statusLabel.Wrapping = fyne.TextWrapWord
	if isErr {
		statusLabel.Importance = widget.DangerImportance
	} else {
		statusLabel.Importance = widget.MediumImportance
	}

	screenOptions := []string{
		settings.ScreenSize1024x768,
		settings.ScreenSize1280x720,
		settings.ScreenSize1366x768,
		settings.ScreenSize1920x1080,
	}
	savedScreen := prefs.StringWithFallback(prefScreenSize, settings.ScreenSize1280x720)
	screenSelect := widget.NewSelect(screenOptions, func(val string) {
		prefs.SetString(prefScreenSize, val)
	})
	screenSelect.SetSelected(savedScreen)

	leftControlCheck := widget.NewCheck("Place control tabs on the left side", func(checked bool) {
		prefs.SetBool(prefLeftControl, checked)
	})
	leftControlCheck.SetChecked(prefs.BoolWithFallback(prefLeftControl, false))

	var connectBtn *widget.Button
	triggerConnect := func(target string) {
		normalized, err := normalizeServerAddress(target)
		if err != nil {
			statusLabel.Importance = widget.DangerImportance
			statusLabel.SetText(err.Error())
			statusLabel.Refresh()
			return
		}
		addrEntry.SetText(normalized)
		statusLabel.Importance = widget.MediumImportance
		statusLabel.SetText("Connecting to " + normalized + " ...")
		statusLabel.Refresh()
		if connectBtn != nil {
			connectBtn.Disable()
		}

		go func() {
			cli, err := c.dialFunc(normalized)
			fyne.Do(func() {
				if connectBtn != nil {
					connectBtn.Enable()
				}
				if err != nil {
					slog.Error("failed to connect to fynescope-server", "addr", normalized, "err", err)
					statusLabel.Importance = widget.DangerImportance
					statusLabel.SetText(fmt.Sprintf("Connection failed: %v", err))
					statusLabel.Refresh()
					return
				}
				c.onConnected(normalized, cli)
			})
		}()
	}

	addrEntry.OnSubmitted = func(s string) {
		triggerConnect(s)
	}

	connectBtn = widget.NewButtonWithIcon("Connect to Server", theme.ConfirmIcon(), func() {
		triggerConnect(addrEntry.Text)
	})
	connectBtn.Importance = widget.HighImportance

	emulatorBtn := widget.NewButton("Android Emulator Host (10.0.2.2:9876)", func() {
		addrEntry.SetText(emulatorHostAddr)
		triggerConnect(emulatorHostAddr)
	})
	localBtn := widget.NewButton("Localhost (127.0.0.1:9876)", func() {
		addrEntry.SetText(localLoopbackAddr)
		triggerConnect(localLoopbackAddr)
	})

	recentBox := container.NewVBox()
	var rebuildRecent func()
	rebuildRecent = func() {
		recentBox.RemoveAll()
		recents := loadRecentServers(prefs)
		if len(recents) == 0 {
			recentBox.Add(widget.NewLabel("No recent servers saved."))
		} else {
			for _, srv := range recents {
				srvAddr := srv
				btn := widget.NewButtonWithIcon(srvAddr, theme.HistoryIcon(), func() {
					addrEntry.SetText(srvAddr)
					triggerConnect(srvAddr)
				})
				recentBox.Add(btn)
			}
			clearBtn := widget.NewButtonWithIcon("Clear Recent", theme.DeleteIcon(), func() {
				prefs.SetString(prefRecentServers, "")
				rebuildRecent()
			})
			recentBox.Add(clearBtn)
		}
		recentBox.Refresh()
	}
	rebuildRecent()

	title := widget.NewLabelWithStyle("Fynescope Android Client", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	subtitle := widget.NewLabel("Landscape mobile client with vertically scrollable control tabs")

	connectionCard := widget.NewCard(
		"Server Connection",
		"Enter headless fynescope-server host:port (default port 9876)",
		container.NewVBox(
			addrEntry,
			connectBtn,
			widget.NewSeparator(),
			widget.NewLabel("Quick Targets:"),
			container.NewGridWithColumns(2, emulatorBtn, localBtn),
			statusLabel,
		),
	)

	optionsCard := widget.NewCard(
		"Mobile UI & Recent Servers",
		"UI scaling and quick reconnect",
		container.NewVBox(
			container.NewBorder(nil, nil, widget.NewLabel("UI Scale Preset:"), nil, screenSelect),
			leftControlCheck,
			widget.NewSeparator(),
			widget.NewLabelWithStyle("Recent Servers", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			recentBox,
		),
	)

	columns := container.NewGridWithColumns(2, connectionCard, optionsCard)
	header := container.NewHBox(title, layout.NewSpacer(), subtitle)
	root := container.NewVBox(header, columns)
	return container.NewVScroll(container.NewPadded(root))
}

// onConnected transitions the master window from the connection screen to the
// oscilloscope UI configured for small landscape mobile screens.
func (c *clientApp) onConnected(addr string, cli *netcontrol.Client) {
	prefs := c.app.Preferences()
	saveRecentServer(prefs, addr)
	c.initialAddr = addr

	genericps.LoadDefaultConstants()

	sn, err := cli.UnitBatchAndSerialInfo()
	if err != nil {
		slog.Warn("UnitBatchAndSerialInfo returned error, using default settings file", "err", err)
	}
	settingsPath := resolveSettingsPath(c.app, sn)

	cfg, err := settings.Load(settingsPath)
	if err != nil {
		slog.Info("initializing default settings", "path", settingsPath, "err", err)
		cfg = settings.NewDefaultSettings()
		if saveErr := settings.Save(settingsPath, cfg); saveErr != nil {
			slog.Warn("could not save initial settings", "path", settingsPath, "err", saveErr)
		}
	}

	selectedScale := prefs.StringWithFallback(prefScreenSize, settings.ScreenSize1280x720)
	if selectedScale != "" {
		cfg.ScreenSize = selectedScale
	}
	cfg.Window.LeftControl = prefs.BoolWithFallback(prefLeftControl, cfg.Window.LeftControl)

	scp := &gui.ScpDesc{
		App:                c.app,
		Window:             c.window,
		MobileMode:         true,
		GifEnabled:         false,
		ExtGenEnabled:      true,
		MultiEnabled:       true,
		FfAutoRangeEnabled: true,
	}
	scp.OnDisconnect = func() {
		c.disconnectActiveSession("Disconnected from " + addr)
	}

	c.activeCli = cli
	c.activeScp = scp

	if err := scp.MenuWithController(cli, cfg, settingsPath); err != nil {
		slog.Error("failed to initialize scope GUI", "err", err)
		_ = cli.Close()
		c.activeCli = nil
		c.activeScp = nil
		c.showConnectScreen(fmt.Sprintf("Failed to initialize scope UI: %v", err), true)
	}
}

// disconnectActiveSession closes the current netcontrol client and returns the
// master window to the server connection screen.
func (c *clientApp) disconnectActiveSession(msg string) {
	if c.activeScp != nil {
		c.activeScp.SaveSettings()
		c.activeScp = nil
	}
	if c.activeCli != nil {
		_ = c.activeCli.Close()
		c.activeCli = nil
	}
	c.showConnectScreen(msg, false)
}

func main() {
	var (
		addrFlag    string
		autoConnect bool
		debug       bool
	)
	flag.StringVar(&addrFlag, "addr", "", "fynescope-server address (host:port)")
	flag.BoolVar(&autoConnect, "connect", false, "immediately connect to -addr on startup")
	flag.BoolVar(&debug, "d", false, "enable debug logs")
	flag.Parse()

	logLevel := slog.LevelInfo
	if debug {
		logLevel = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     logLevel,
	})))

	a := app.NewWithID(appID)
	a.Settings().SetTheme(gui.Theme(settings.DarkTheme))

	w := a.NewWindow("Fynescope Android Client")
	w.SetMaster()
	w.Resize(fyne.NewSize(960, 540))

	client := newClientApp(a, w, addrFlag, autoConnect)
	client.showConnectScreen("Ready to connect.", false)

	if autoConnect && strings.TrimSpace(addrFlag) != "" {
		go func() {
			normalized, err := normalizeServerAddress(addrFlag)
			if err != nil {
				fyne.Do(func() {
					client.showConnectScreen(err.Error(), true)
				})
				return
			}
			netCli, err := client.dialFunc(normalized)
			fyne.Do(func() {
				if err != nil {
					client.showConnectScreen(fmt.Sprintf("Connection failed: %v", err), true)
					return
				}
				client.onConnected(normalized, netCli)
			})
		}()
	}

	w.ShowAndRun()
}

