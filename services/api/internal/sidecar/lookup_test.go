package sidecar

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLookPrefersEnv(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "opencode")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CHERRY_OPENCODE_BIN", bin)
	hit, err := Look("opencode")
	if err != nil {
		t.Fatal(err)
	}
	if hit.Source != "env" || hit.Path != bin {
		t.Fatalf("%+v", hit)
	}
}

func TestLookBundledDir(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "maestro")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CHERRY_OPENCODE_BIN", "")
	t.Setenv("CHERRY_MAESTRO_BIN", "")
	t.Setenv("CHERRY_SIDECAR_DIR", dir)
	t.Setenv("PATH", "/nonexistent")
	hit, err := Look("maestro")
	if err != nil {
		t.Fatal(err)
	}
	if hit.Source != "bundled" {
		t.Fatalf("source %s path %s", hit.Source, hit.Path)
	}
}

func TestLookMissing(t *testing.T) {
	t.Setenv("CHERRY_OPENCODE_BIN", "")
	t.Setenv("CHERRY_SIDECAR_DIR", t.TempDir())
	t.Setenv("PATH", "/nonexistent")
	if _, err := Look("opencode"); err == nil {
		t.Fatal("expected missing")
	}
}
