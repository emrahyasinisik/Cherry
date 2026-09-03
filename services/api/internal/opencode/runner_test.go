package opencode

import (
	"context"
	"fmt"
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
	t.Setenv("CHERRY_OPENCODE_BIN", "")
	t.Setenv("CHERRY_SIDECAR_DIR", t.TempDir())
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
	t.Setenv("CHERRY_OPENCODE_BIN", "")
	t.Setenv("CHERRY_SIDECAR_DIR", t.TempDir())
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
	if err := WriteAgents(dir, "Kahve", "Expo / React Native", "sipariş", "Uygulama kodu TypeScript / React Native (Expo)."); err != nil {
		t.Fatal(err)
	}
	agents, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(agents)
	if !strings.Contains(text, "TypeScript") {
		t.Fatalf("AGENTS.md must name the stack language: %s", text)
	}
	if !strings.Contains(text, "HTML") {
		t.Fatalf("AGENTS.md must forbid HTML as the app: %s", text)
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

func TestAbsDir(t *testing.T) {
	rel := filepath.Join(".", "opencode-abs-test")
	if err := os.MkdirAll(rel, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(rel) })
	got, err := absDir(rel)
	if err != nil {
		t.Fatal(err)
	}
	want, err := filepath.Abs(rel)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("got %s want %s", got, want)
	}
	if _, err := absDir(filepath.Join(rel, "missing")); err == nil {
		t.Fatal("expected missing dir error")
	}
}

func TestCLIMissingKeyFailsHonestly(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "opencode")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\necho should-not-run\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CHERRY_OPENCODE_BIN", bin)
	t.Setenv("CHERRY_LLM_API_KEY", "")
	t.Setenv("CHERRY_LLM_BASE_URL", "")
	t.Setenv("OPENAI_API_KEY", "")
	cli := NewCLI()
	res, err := cli.Run(context.Background(), Request{Dir: t.TempDir(), Prompt: "x"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != StatusFailed {
		t.Fatalf("status %s", res.Status)
	}
	if !strings.Contains(res.Err, "model anahtarı yok") {
		t.Fatalf("err %q", res.Err)
	}
}

func TestWriteConfigWithBaseURL(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CHERRY_SIDECAR_DIR", t.TempDir())
	if err := WriteConfig(dir, Endpoint{BaseURL: "https://cherry.visevent.com/v1", Model: "Qwen/Qwen2.5-1.5B-Instruct"}); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(dir, "opencode.json"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if !strings.Contains(text, "cherry.visevent.com/v1") {
		t.Fatalf("%s", text)
	}
	if !strings.Contains(text, "openai/") {
		t.Fatalf("model provider prefix missing: %s", text)
	}
}

func TestFailErrPrefersOpenCodeLine(t *testing.T) {
	raw := "\x1b[91m\x1b[1mError: \x1b[0mFailed to change directory to ../../var/projects/x\n"
	got := failErr(fmt.Errorf("exit status 1"), raw)
	if !strings.Contains(got, "Failed to change directory") {
		t.Fatalf("%q", got)
	}
	if strings.Contains(got, "exit status") {
		t.Fatalf("should not be generic: %q", got)
	}
}
