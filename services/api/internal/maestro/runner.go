package maestro

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/cherry/api/internal/sidecar"
	"github.com/cherry/api/internal/store"
)

type Device struct {
	ID       string
	Online   bool
	Platform string // android | ios
}

type FlowResult struct {
	ID     string
	Result store.MaestroResult
	Note   string
}

type Report struct {
	DeviceStatus string
	Flows        []FlowResult
	Note         string
}

type Runner struct {
	Bin         string
	Timeout     time.Duration
	MaxAttempts int
	// ListDevices overrides adb/simctl discovery (tests).
	ListDevices func(ctx context.Context) []Device
}

func New() *Runner {
	bin := ""
	if hit, err := sidecar.Look("maestro"); err == nil {
		bin = hit.Path
	}
	return &Runner{Bin: bin, Timeout: 90 * time.Second, MaxAttempts: 3}
}

func (r *Runner) Probe() (bin string, ok bool) {
	if strings.TrimSpace(r.Bin) != "" {
		return r.Bin, true
	}
	hit, err := sidecar.Look("maestro")
	if err != nil {
		return "", false
	}
	r.Bin = hit.Path
	return hit.Path, true
}

func (r *Runner) Devices(ctx context.Context) []Device {
	if r.ListDevices != nil {
		return r.ListDevices(ctx)
	}
	out := adbDevices(ctx)
	out = append(out, iosSimulators(ctx)...)
	return out
}

func (r *Runner) RunDir(ctx context.Context, maestroDir, localURL string) Report {
	report := Report{DeviceStatus: "none", Note: "Cihaz yok. SKIPPED — geçti sayılmaz."}
	entries, err := filepath.Glob(filepath.Join(maestroDir, "*.yaml"))
	if err != nil {
		report.Note = err.Error()
		return report
	}
	devices := r.Devices(ctx)
	bin, haveCLI := r.Probe()
	if !haveCLI {
		report.DeviceStatus = "no_cli"
		report.Note = "Maestro CLI yok. SKIPPED — passed uydurulmaz."
		for _, path := range entries {
			name := strings.TrimSuffix(filepath.Base(path), ".yaml")
			note := "Maestro CLI yok. SKIPPED — sahte geçiş yok."
			if len(devices) == 0 {
				note = "Maestro CLI yok; cihaz da yok. SKIPPED — sahte geçiş yok."
			} else {
				note = "Cihaz görüldü ama Maestro CLI yok. SKIPPED — passed uydurulmaz."
			}
			if localURL != "" {
				note += " Yerel API: " + localURL
			}
			report.Flows = append(report.Flows, FlowResult{
				ID:     name,
				Result: store.MaestroSkipped,
				Note:   note,
			})
		}
		return report
	}
	if len(devices) == 0 {
		if envTruthy("CHERRY_MAESTRO_START_DEVICE") {
			if started := r.tryStartDevice(ctx, bin); len(started) > 0 {
				devices = started
			}
		}
	}
	if len(devices) == 0 {
		report.DeviceStatus = "none"
		for _, path := range entries {
			name := strings.TrimSuffix(filepath.Base(path), ".yaml")
			note := "Emülatör yok. SKIPPED — geçti sayılmaz. Android Studio AVD veya iOS Simulator aç."
			if localURL != "" {
				note += " Yerel API: " + localURL
			}
			report.Flows = append(report.Flows, FlowResult{
				ID:     name,
				Result: store.MaestroSkipped,
				Note:   note,
			})
		}
		return report
	}
	device := pickDevice(devices)
	report.DeviceStatus = "device"
	report.Note = "Cihaz: " + device.ID + " (" + device.Platform + ")"
	for _, path := range entries {
		name := strings.TrimSuffix(filepath.Base(path), ".yaml")
		res, note := r.testFile(ctx, bin, path, device)
		if localURL != "" {
			note += " API " + localURL
		}
		report.Flows = append(report.Flows, FlowResult{ID: name, Result: res, Note: note})
	}
	return report
}

func (r *Runner) testFile(ctx context.Context, bin, yamlPath string, device Device) (store.MaestroResult, string) {
	attempts := r.MaxAttempts
	if attempts <= 0 {
		attempts = 1
	}
	if attempts > 3 {
		attempts = 3
	}
	var lastNote string
	for try := 1; try <= attempts; try++ {
		res, note := r.runOnce(ctx, bin, yamlPath, device)
		if res == store.MaestroPassed {
			if try > 1 {
				note += " (deneme " + strconv.Itoa(try) + "/" + strconv.Itoa(attempts) + ")"
			}
			return res, note
		}
		if res == store.MaestroSkipped {
			return res, note
		}
		lastNote = note
		if try < attempts {
			lastNote += " — yeniden denenecek (" + strconv.Itoa(try) + "/" + strconv.Itoa(attempts) + ")"
		}
	}
	return store.MaestroFailed, lastNote
}

func (r *Runner) runOnce(ctx context.Context, bin, yamlPath string, device Device) (store.MaestroResult, string) {
	timeout := r.Timeout
	if timeout <= 0 {
		timeout = 90 * time.Second
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	args := []string{}
	if device.ID != "" {
		args = append(args, "--device", device.ID)
	}
	args = append(args, "test", yamlPath)
	cmd := exec.CommandContext(runCtx, bin, args...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	out := strings.TrimSpace(buf.String())
	if err == nil {
		return store.MaestroPassed, "Maestro geçti · " + device.ID
	}
	low := strings.ToLower(out + " " + err.Error())
	if strings.Contains(low, "no device") || strings.Contains(low, "device not found") ||
		(strings.Contains(low, "emulator") && strings.Contains(low, "not")) {
		return store.MaestroSkipped, "Cihaz kayboldu. SKIPPED — geçti sayılmaz."
	}
	if len(out) > 240 {
		out = out[:240] + "…"
	}
	if out == "" {
		out = err.Error()
	}
	return store.MaestroFailed, "Maestro kaldı: " + out
}

func (r *Runner) tryStartDevice(ctx context.Context, bin string) []Device {
	runCtx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	platform := "android"
	if runtime.GOOS == "darwin" {
		if v := strings.TrimSpace(os.Getenv("CHERRY_MAESTRO_PLATFORM")); v != "" {
			platform = v
		}
	}
	cmd := exec.CommandContext(runCtx, bin, "start-device", "--platform", platform)
	_ = cmd.Run()
	return r.Devices(ctx)
}

func pickDevice(devices []Device) Device {
	want := strings.TrimSpace(os.Getenv("CHERRY_MAESTRO_DEVICE"))
	if want != "" {
		for _, d := range devices {
			if d.ID == want {
				return d
			}
		}
	}
	return devices[0]
}

func adbDevices(ctx context.Context) []Device {
	bin, err := exec.LookPath("adb")
	if err != nil {
		return nil
	}
	runCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	cmd := exec.CommandContext(runCtx, bin, "devices")
	raw, err := cmd.CombinedOutput()
	if err != nil {
		return nil
	}
	var out []Device
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "List") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		if fields[1] != "device" {
			continue
		}
		out = append(out, Device{ID: fields[0], Online: true, Platform: "android"})
	}
	return out
}

func iosSimulators(ctx context.Context) []Device {
	if runtime.GOOS != "darwin" {
		return nil
	}
	bin, err := exec.LookPath("xcrun")
	if err != nil {
		return nil
	}
	runCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	cmd := exec.CommandContext(runCtx, bin, "simctl", "list", "devices", "booted")
	raw, err := cmd.CombinedOutput()
	if err != nil {
		return nil
	}
	var out []Device
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		// iPhone 15 (UUID) (Booted)
		if !strings.Contains(line, "(Booted)") {
			continue
		}
		start := strings.Index(line, "(")
		end := strings.Index(line, ")")
		if start < 0 || end <= start {
			continue
		}
		id := strings.TrimSpace(line[start+1 : end])
		if id == "" {
			continue
		}
		out = append(out, Device{ID: id, Online: true, Platform: "ios"})
	}
	return out
}

func envTruthy(key string) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	switch v {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
