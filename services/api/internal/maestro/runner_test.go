package maestro

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cherry/api/internal/store"
)

func TestRunDirSkipsWithoutDevice(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "login.yaml"), []byte("appId: x\n---\n- launchApp\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	r := &Runner{Bin: "", Timeout: 0}
	t.Setenv("CHERRY_MAESTRO_BIN", "")
	t.Setenv("CHERRY_SIDECAR_DIR", t.TempDir())
	t.Setenv("PATH", "/nonexistent")
	report := r.RunDir(context.Background(), dir, "http://127.0.0.1:47001")
	if report.DeviceStatus != "no_cli" || len(report.Flows) != 1 {
		t.Fatalf("%+v", report)
	}
	if report.Flows[0].Result != store.MaestroSkipped {
		t.Fatalf("must skip, got %s", report.Flows[0].Result)
	}
	if report.Flows[0].Result == store.MaestroPassed {
		t.Fatal("never pass without a device")
	}
}

func TestRunDirSkipsWithoutDeviceWhenCLIPresent(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "login.yaml"), []byte("appId: x\n---\n- launchApp\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(t.TempDir(), "maestro")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\necho ok\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	r := &Runner{Bin: bin, Timeout: 0, ListDevices: func(context.Context) []Device { return nil }}
	t.Setenv("PATH", "/nonexistent")
	report := r.RunDir(context.Background(), dir, "")
	if report.DeviceStatus != "none" {
		t.Fatalf("want none (CLI ok, no device), got %+v", report)
	}
	if report.Flows[0].Result != store.MaestroSkipped {
		t.Fatalf("must skip, got %s", report.Flows[0].Result)
	}
}

func TestNeverPassOnEmptyDevicesEvenWithBin(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "home.yaml"), []byte("appId: x\n"), 0o644)
	r := &Runner{
		Bin:         "/nonexistent/maestro",
		ListDevices: func(context.Context) []Device { return nil },
	}
	t.Setenv("PATH", "/nonexistent")
	t.Setenv("CHERRY_SIDECAR_DIR", t.TempDir())
	report := r.RunDir(context.Background(), dir, "")
	for _, flow := range report.Flows {
		if flow.Result == store.MaestroPassed {
			t.Fatal("passed without device")
		}
	}
}

func TestRunDirPassesWithDeviceAndCLI(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "login.yaml"), []byte("appId: x\n---\n- launchApp\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(t.TempDir(), "maestro")
	script := "#!/bin/sh\n# args: --device ID test path\nif [ \"$1\" != \"--device\" ] || [ \"$3\" != \"test\" ]; then\n  echo bad args: \"$@\" >&2\n  exit 2\nfi\necho passed\nexit 0\n"
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	r := &Runner{
		Bin:     bin,
		Timeout: 5 * time.Second,
		ListDevices: func(context.Context) []Device {
			return []Device{{ID: "emulator-5554", Online: true, Platform: "android"}}
		},
	}
	report := r.RunDir(context.Background(), dir, "http://127.0.0.1:47001")
	if report.DeviceStatus != "device" {
		t.Fatalf("want device, got %+v", report)
	}
	if report.Flows[0].Result != store.MaestroPassed {
		t.Fatalf("want PASSED with device+CLI, got %s note=%s", report.Flows[0].Result, report.Flows[0].Note)
	}
}

func TestRunDirRetriesFailedThenPasses(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "home.yaml"), []byte("appId: x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(t.TempDir(), "maestro")
	counter := filepath.Join(t.TempDir(), "n")
	script := "#!/bin/sh\nn=0\nif [ -f '" + counter + "' ]; then n=$(cat '" + counter + "'); fi\nn=$((n+1))\necho \"$n\" > '" + counter + "'\nif [ \"$n\" -lt 2 ]; then echo fail; exit 1; fi\necho ok; exit 0\n"
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	r := &Runner{
		Bin:         bin,
		Timeout:     5 * time.Second,
		MaxAttempts: 3,
		ListDevices: func(context.Context) []Device {
			return []Device{{ID: "emu-1", Online: true, Platform: "android"}}
		},
	}
	report := r.RunDir(context.Background(), dir, "")
	if report.Flows[0].Result != store.MaestroPassed {
		t.Fatalf("want pass after retry, got %+v", report.Flows[0])
	}
}

func TestPickDeviceEnvOverride(t *testing.T) {
	t.Setenv("CHERRY_MAESTRO_DEVICE", "second")
	got := pickDevice([]Device{
		{ID: "first", Platform: "android"},
		{ID: "second", Platform: "android"},
	})
	if got.ID != "second" {
		t.Fatalf("got %s", got.ID)
	}
}
