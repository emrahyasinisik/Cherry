package maestro

import (
	"context"
	"os"
	"path/filepath"
	"testing"

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
	r := &Runner{Bin: bin, Timeout: 0}
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
	r := &Runner{Bin: "/nonexistent/maestro"}
	t.Setenv("PATH", "/nonexistent")
	t.Setenv("CHERRY_SIDECAR_DIR", t.TempDir())
	report := r.RunDir(context.Background(), dir, "")
	for _, flow := range report.Flows {
		if flow.Result == store.MaestroPassed {
			t.Fatal("passed without device")
		}
	}
}
