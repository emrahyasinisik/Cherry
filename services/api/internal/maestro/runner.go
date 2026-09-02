package maestro

import (
	"bytes"
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/icerde/api/internal/sidecar"
	"github.com/icerde/api/internal/store"
)

type Device struct {
	ID     string
	Online bool
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
	Bin     string
	Timeout time.Duration
}

func New() *Runner {
	bin := ""
	if hit, err := sidecar.Look("maestro"); err == nil {
		bin = hit.Path
	}
	return &Runner{Bin: bin, Timeout: 90 * time.Second}
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
	out := adbDevices(ctx)
	if len(out) > 0 {
		return out
	}
	return nil
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
	if len(devices) == 0 {
		report.DeviceStatus = "none"
		for _, path := range entries {
			name := strings.TrimSuffix(filepath.Base(path), ".yaml")
			note := "Emülatör yok. SKIPPED — geçti sayılmaz."
			if localURL != "" {
				note += " Yerel API: " + localURL
			}
			if !haveCLI {
				note = "Maestro sidecar yok; cihaz da yok. SKIPPED — sahte geçiş yok."
			}
			report.Flows = append(report.Flows, FlowResult{
				ID:     name,
				Result: store.MaestroSkipped,
				Note:   note,
			})
		}
		return report
	}
	report.DeviceStatus = "device"
	report.Note = "Cihaz var."
	if !haveCLI {
		for _, path := range entries {
			name := strings.TrimSuffix(filepath.Base(path), ".yaml")
			report.Flows = append(report.Flows, FlowResult{
				ID:     name,
				Result: store.MaestroSkipped,
				Note:   "Cihaz görüldü ama Maestro CLI yok. SKIPPED — passed uydurulmaz.",
			})
		}
		report.Note = "Maestro CLI yok."
		return report
	}
	for _, path := range entries {
		name := strings.TrimSuffix(filepath.Base(path), ".yaml")
		res, note := r.testFile(ctx, bin, path)
		if localURL != "" {
			note += " API " + localURL
		}
		report.Flows = append(report.Flows, FlowResult{ID: name, Result: res, Note: note})
	}
	return report
}

func (r *Runner) testFile(ctx context.Context, bin, yamlPath string) (store.MaestroResult, string) {
	timeout := r.Timeout
	if timeout <= 0 {
		timeout = 90 * time.Second
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(runCtx, bin, "test", yamlPath)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	out := strings.TrimSpace(buf.String())
	if err == nil {
		return store.MaestroPassed, "Maestro geçti."
	}
	low := strings.ToLower(out + " " + err.Error())
	if strings.Contains(low, "no device") || strings.Contains(low, "device not found") || strings.Contains(low, "emulator") && strings.Contains(low, "not") {
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
		out = append(out, Device{ID: fields[0], Online: true})
	}
	return out
}
