// fynescope-server is a headless binary that exposes a PicoScope device over
// TCP using the netcontrol protocol. A remote fynescope GUI client (built with
// -tags remote) can connect to it and operate the scope from another machine.
//
// Build:
//
//	go build -o fynescope-server ./cmd/fynescope-server
//
// Run:
//
//	./fynescope-server -port 9876 -demo
//	./fynescope-server -port 9876          # auto-selects the first real device
//
// The server does NOT import fyne or the gui package, keeping it CGO-light
// (only the PicoScope SDK libraries are needed).
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strings"
	"syscall"
	"time"

	// Load device-specific drivers.  Mirror the set from main.go as needed;
	// comment out any that are not installed on the server machine.
	// _ "fynescope/ps2000"
	_ "fynescope/ps2000a"
	// _ "fynescope/ps3000"
	// _ "fynescope/ps3000a"
	// _ "fynescope/ps4000"
	// _ "fynescope/ps4000a"
	// _ "fynescope/ps5000"
	// _ "fynescope/ps5000a"
	// _ "fynescope/ps6000"
	// _ "fynescope/ps6000a"

	"fynescope/control"
	"fynescope/genericps"
	"fynescope/netcontrol"
)

var (
	GitUUID   = ""
	Version   = "1.4.0"
	BuildDate = ""
)

// FilterHandler is a copy of the one in main.go — keeping server self-contained.
type FilterHandler struct {
	handler slog.Handler
	level   *slog.LevelVar
}

func (h *FilterHandler) Enabled(_ context.Context, level slog.Level) bool {
	if level == slog.LevelDebug {
		return true
	}
	return level >= h.level.Level()
}

func (h *FilterHandler) Handle(ctx context.Context, r slog.Record) error {
	if r.Level >= h.level.Level() {
		return h.handler.Handle(ctx, r)
	}
	if r.Level == slog.LevelDebug {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		frame, _ := fs.Next()
		pkg := frame.Function
		if i := strings.LastIndexByte(pkg, '/'); i != -1 {
			pkg = pkg[i+1:]
		}
		if i := strings.IndexByte(pkg, '.'); i != -1 {
			pkg = pkg[:i]
		}
		filename := path.Join(pkg, path.Base(frame.File))
		_ = filename // debug-on map not used here; always suppress debug
	}
	return nil
}

func (h *FilterHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &FilterHandler{handler: h.handler.WithAttrs(attrs), level: h.level}
}

func (h *FilterHandler) WithGroup(name string) slog.Handler {
	return &FilterHandler{handler: h.handler.WithGroup(name), level: h.level}
}

func setLogging(loglevel string) {
	programLevel := new(slog.LevelVar)
	baseHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			switch a.Key {
			case slog.SourceKey:
				s := a.Value.Any().(*slog.Source)
				s.File = path.Base(s.File)
			case slog.TimeKey:
				t := a.Value.Time()
				a.Value = slog.StringValue(t.Format(time.TimeOnly))
			}
			return a
		},
	})
	slog.SetDefault(slog.New(&FilterHandler{handler: baseHandler, level: programLevel}))

	switch loglevel {
	case "debug":
		programLevel.Set(slog.LevelDebug)
	case "info":
		programLevel.Set(slog.LevelInfo)
	case "warning":
		programLevel.Set(slog.LevelWarn)
	case "error":
		programLevel.Set(slog.LevelError)
	default:
		programLevel.Set(slog.LevelWarn)
	}
}

func main() {
	port := flag.Int("port", 9876, "-port=9876  TCP port to listen on")
	demoOnly := flag.Bool("demo", false, "-demo=true  use demo scope instead of real hardware")
	logLevel := flag.String("loglevel", "warning", "-loglevel=info | debug | warning | error")
	about := flag.Bool("about", false, "print version and exit")
	flag.Parse()

	if *about {
		fmt.Printf("fynescope-server %s (built %s, git %s)\n", Version, BuildDate, GitUUID)
		os.Exit(0)
	}

	setLogging(*logLevel)

	// --- Enumerate / select device ------------------------------------------
	var device *genericps.DeviceInfo

	if *demoOnly {
		allDevices, _ := genericps.EnumerateAllDevices(256)
		for _, d := range allDevices {
			if d.IsDemo {
				d := d
				device = &d
				break
			}
		}
		if device == nil {
			device = &genericps.DeviceInfo{Id: genericps.DemoId, IsDemo: true}
		}
	} else {
		devices, err := genericps.EnumerateAllDevices(256)
		if err != nil || len(devices) == 0 {
			slog.Error("no PicoScope devices found", "err", err)
			os.Exit(1)
		}
		// Pick the first non-demo device; fall back to demo if all are demo.
		for i := range devices {
			if !devices[i].IsDemo {
				device = &devices[i]
				break
			}
		}
		if device == nil {
			device = &devices[0]
		}
		slog.Info("selected device", "id", device.Id, "serial", device.Serial)
	}

	// --- Connect to device --------------------------------------------------
	con := genericps.NewConnection()
	var connErr error
	if device.IsDemo {
		con.Handle, connErr = genericps.OpenDemo(con, device.Id)
	} else {
		con.Handle, connErr = genericps.OpenUnit(con, device.Id, device.Serial, 0)
	}
	con.ID = device.Id
	if connErr != nil {
		slog.Error("failed to open device", "err", connErr)
		os.Exit(1)
	}
	defer con.CloseUnit()

	info, _ := con.GetUnitInfo(genericps.PicoVariantInfo)
	slog.Info("device opened", "info", info)

	// --- Create controller & start netcontrol server ------------------------
	ctrl := control.NewControl(con)

	listenAddr := fmt.Sprintf(":%d", *port)
	srv := netcontrol.NewServer(ctrl)
	if err := srv.Start(listenAddr); err != nil {
		slog.Error("failed to start netcontrol server", "addr", listenAddr, "err", err)
		os.Exit(1)
	}
	slog.Info("fynescope-server listening", "addr", srv.Addr())
	fmt.Printf("fynescope-server listening on %s\n", srv.Addr())

	// --- Block until SIGINT / SIGTERM ---------------------------------------
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	slog.Info("shutting down")
	ctrl.Shutdown()
	_ = srv.Close()
	slog.Info("done")
}
