// DO NOT run test on real hardware
package main

import (
	"flag"
	"fmt"
	"fynescope/gui"
	"fynescope/settings"
	"fynescope/web"
	"image"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

const (
	defaultSimScope  = "2207BMSO"
	minFuzzerTimeout = 20 * time.Minute
)

var (
	scp           *gui.ScpDesc
	fuzzTimeFlag  *time.Duration
	fuzzerWebOnce sync.Once
)

func TestMain(m *testing.M) {
	s := "debug"
	setLogging(&s)

	log.SetFlags(log.Ltime | log.Lshortfile)

	if flag.Lookup("sim") == nil {
		_ = registerSimFlag()
	}
	if flag.Lookup("fuzzduration") == nil {
		fuzzTimeFlag = flag.Duration("fuzzduration", 0, "duration to run fuzzer test even when -timeout < 20m")
	}
	if !flag.Parsed() {
		flag.Parse()
	}

	simName := ""
	if f := flag.Lookup("sim"); f != nil {
		simName = f.Value.String()
	}
	if simName == "" {
		simName = os.Getenv("FUZZER_SIM_NAME")
	}
	if isSimBuild && simName == "" {
		simName = defaultSimScope
	}

	con, err := openSim(simName)
	if err != nil {
		log.Fatalf("failed to open simulator: %v", err)
	}

	_, hasExplicitDur, err := explicitFuzzerDuration()
	if err != nil {
		log.Fatalf("invalid fuzzer duration: %v", err)
	}
	isFuzzRun := os.Getenv("FUZZER_COMMIT_ID") != "" || hasExplicitDur
	if !isFuzzRun {
		if tf := flag.Lookup("test.timeout"); tf != nil {
			if d, err := time.ParseDuration(tf.Value.String()); err == nil && d >= minFuzzerTimeout {
				isFuzzRun = true
			}
		}
	}
	if isFuzzRun {
		settingFileName = fmt.Sprintf("scopesettings_fuzzer_%d.yaml", time.Now().UnixNano())
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
	done := make(chan struct{})
	go func() {
		time.Sleep(2 * time.Second)
		exitCode = m.Run()
		close(done)
		go func() {
			time.Sleep(5 * time.Second)
			log.Println("WARN: App shutdown timed out after m.Run(); forcing exit")
			os.Exit(exitCode)
		}()
		fyne.Do(scp.App.Quit)
	}()

	scp.App.Run()

	// Wait for the test goroutine to finish its defers (up to 2 seconds)
	// This ensures that if the app is closed via Ctrl+C, the fuzzer has time to write its log file.
	select {
	case <-done:
	case <-time.After(2 * time.Second):
	}

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

func TestFunctional(t *testing.T) {
	if scp == nil {
		t.Fatal("scp is nil — app failed to initialize")
	}
	err := scp.FunctionalTestUnit()
	if err != nil {
		t.Fatal(err)
	}
}

func explicitFuzzerDuration() (time.Duration, bool, error) {
	if fuzzTimeFlag != nil && *fuzzTimeFlag > 0 {
		return *fuzzTimeFlag, true, nil
	}
	if tf := flag.Lookup("test.fuzztime"); tf != nil && tf.Value.String() != "" {
		if d, err := time.ParseDuration(tf.Value.String()); err == nil && d > 0 {
			return d, true, nil
		}
	}
	if envDur := os.Getenv("FUZZER_DURATION"); envDur != "" {
		d, err := time.ParseDuration(envDur)
		if err != nil || d <= 0 {
			return 0, false, fmt.Errorf("FUZZER_DURATION %q: %v", envDur, err)
		}
		return d, true, nil
	}
	return 0, false, nil
}

func resolveFuzzerTimeout(t *testing.T) (time.Duration, bool) {
	t.Helper()
	if d, ok, err := explicitFuzzerDuration(); err != nil {
		t.Fatal(err)
	} else if ok {
		return d, true
	}
	deadline, ok := t.Deadline()
	if !ok {
		return 0, false
	}
	if time.Until(deadline) < minFuzzerTimeout {
		return 0, false
	}
	log.Printf("deadline: %v", deadline)
	timeout := time.Until(deadline) - 45*time.Second
	if timeout < 0 {
		timeout = 0
	}
	return timeout, true
}

func ensureFuzzerCommitID() {
	if gui.FuzzerCommitID != "" || os.Getenv("FUZZER_COMMIT_ID") != "" {
		return
	}
	out, err := exec.Command("git", "rev-parse", "HEAD").Output()
	if err == nil {
		gui.FuzzerCommitID = strings.TrimSpace(string(out))
	}
	if gui.FuzzerCommitID == "" {
		gui.FuzzerCommitID = "unknown"
	}
}

func runFuzzerTest(t *testing.T, defaultWebPort int, requireSim bool) {
	t.Helper()
	timeout, ok := resolveFuzzerTimeout(t)
	if !ok {
		t.Skip("Skipping fuzzer test")
	}
	if scp == nil {
		t.Fatal("scp is nil — app failed to initialize")
	}
	if requireSim {
		if !isSimBuild {
			t.Fatal("sim fuzzer test requires -tags=\"sim\" (got non-sim build)")
		}
		info, isDemo := scp.DeviceInfo()
		if info == "" || isDemo || !strings.Contains(info, "SIM") {
			t.Fatalf("sim fuzzer test must run on a simulated scope, not demo mode (device=%q, isDemo=%v)", info, isDemo)
		}
	}

	log.Printf("timeout: %v", timeout)
	ensureFuzzerCommitID()

	port := defaultWebPort
	if portStr := os.Getenv("FUZZER_WEBPORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}
	webportStr := ""
	if port > 0 {
		webportStr = fmt.Sprintf("%d", port)
		fuzzerWebOnce.Do(func() {
			web.StartServerNoVoice(port, "", "", func() image.Image {
				if scp.Window == nil || scp.Window.Canvas() == nil {
					return nil
				}
				return scp.Window.Canvas().Capture()
			})
		})
	}

	scp.Random(timeout, Version, BuildDate, webportStr, settingFileName)
}

// TestSimFuzzer runs the GUI fuzzer using -sim (default 2207BMSO) and refuses demo mode.
// Run with: go test -tags="sim,web" -v -run TestSimFuzzer -timeout 105m -sim=2207BMSO
func TestSimFuzzer(t *testing.T) {
	runFuzzerTest(t, 0, true)
}

// Test0 runs the GUI fuzzer for the duration set by the -timeout or -fuzztime flag.
// Run with: go test -tags="sim,web" -v -run Test0 -timeout 105m -sim=2207BMSO
func Test0(t *testing.T) {
	runFuzzerTest(t, 0, false)
}

// Test1 runs the GUI fuzzer with webport=8080 for the duration set by the -timeout or -fuzztime flag.
// Run with: go test -tags="sim,web" -v -run Test1 -timeout 105m -sim=2207BMSO
func Test1(t *testing.T) {
	runFuzzerTest(t, 8080, false)
}
