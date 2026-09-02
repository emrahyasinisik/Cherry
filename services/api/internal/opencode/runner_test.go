package opencode

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFakeWritesMarker(t *testing.T) {
	dir := t.TempDir()
	fake := &Fake{}
	res, err := fake.Run(context.Background(), Request{Dir: dir, Prompt: "brif", Title: "t"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != StatusRan {
		t.Fatalf("status %s", res.Status)
	}
	if _, err := os.Stat(filepath.Join(dir, "llm", "opencode.ran")); err != nil {
		t.Fatal(err)
	}
	if fake.Prompt != "brif" {
		t.Fatalf("prompt %q", fake.Prompt)
	}
}

func TestCLIMissing(t *testing.T) {
	t.Setenv("PATH", "/nonexistent")
	cli := &CLI{Bin: "", Timeout: 0, Require: false}
	res, err := cli.Run(context.Background(), Request{Dir: t.TempDir(), Prompt: "x"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != StatusMissing {
		t.Fatalf("status %s", res.Status)
	}
}

func TestCLIRequireMissing(t *testing.T) {
	t.Setenv("PATH", "/nonexistent")
	cli := &CLI{Require: true}
	_, err := cli.Run(context.Background(), Request{Dir: t.TempDir(), Prompt: "x"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestWriteConfigAndLog(t *testing.T) {
	dir := t.TempDir()
	if err := WriteConfig(dir); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(dir, "opencode.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "opencode.ai/config.json") {
		t.Fatalf("%s", body)
	}
	if err := WriteAgents(dir, "Kahve", "Expo", "sipariş"); err != nil {
		t.Fatal(err)
	}
	if err := WriteLog(dir, LogBody(Result{Status: StatusRan, Bin: "opencode", Output: "ok"})); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "AGENTS.md")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "llm", "opencode.log")); err != nil {
		t.Fatal(err)
	}
}

func TestStatusLabelExhaustive(t *testing.T) {
	for _, status := range []Status{StatusRan, StatusMissing, StatusFailed, Status("other")} {
		if status.Label() == "" {
			t.Fatalf("empty label for %s", status)
		}
	}
}
