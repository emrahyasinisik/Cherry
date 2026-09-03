package colabbridge

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParsePublicURL(t *testing.T) {
	log := `
INF Starting metrics server
INF |  Your quick Tunnel has been created! Visit it at (it may take some time to be reachable):  |
INF |  https://amber-coral-cherry.trycloudflare.com                                                     |
`
	got := ParsePublicURL(log)
	if got != "https://amber-coral-cherry.trycloudflare.com" {
		t.Fatalf("got %q", got)
	}
	if ParsePublicURL("https://trycloudflare.com") != "" {
		t.Fatal("bare host must not count")
	}
	if ParsePublicURL("no url here") != "" {
		t.Fatal("empty")
	}
}

func TestStartMissingCloudflaredNoFakeURL(t *testing.T) {
	t.Setenv("CHERRY_CLOUDFLARED_BIN", "")
	t.Setenv("CHERRY_SIDECAR_DIR", t.TempDir())
	t.Setenv("PATH", "/nonexistent")
	svc := New("127.0.0.1:0", t.TempDir())
	snap := svc.Start("user-1")
	if snap.Status != StatusFailed {
		t.Fatalf("status %s", snap.Status)
	}
	if snap.PublicURL != "" {
		t.Fatalf("fake url %q", snap.PublicURL)
	}
	if !strings.Contains(snap.Note, "cloudflared yok") {
		t.Fatalf("note %q", snap.Note)
	}
	if snap.Token != "" {
		t.Fatal("token must be empty when tunnel failed")
	}
}

func TestBridgePackAndCheckpoint(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "cloudflared")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CHERRY_CLOUDFLARED_BIN", bin)
	t.Setenv("CHERRY_SIDECAR_DIR", t.TempDir())

	checkDir := t.TempDir()
	var registered string
	svc := New("127.0.0.1:0", checkDir)
	svc.NewTunnel = func(string) Tunnel {
		return StaticTunnel{URL: "https://cherry-test.trycloudflare.com"}
	}
	svc.Pack = func(context.Context, string) (string, string, error) {
		return `{"schema":"cherry.training_pack.v1","examples":[]}`, `{"instruction":"x"}` + "\n", nil
	}
	svc.Register = func(_ context.Context, userID, slot, name, note, ref string) error {
		if userID != "user-1" || slot != "A" || name != "v-colab" {
			t.Fatalf("register %s %s %s", userID, slot, name)
		}
		registered = ref
		return nil
	}

	start := svc.Start("user-1")
	if start.Status != StatusStarting && start.Status != StatusRunning {
		t.Fatalf("start %s %s", start.Status, start.Note)
	}
	snap := waitRunning(t, svc)
	if snap.PublicURL != "https://cherry-test.trycloudflare.com" {
		t.Fatalf("url %q", snap.PublicURL)
	}
	if snap.Token == "" || snap.TokenHint != snap.Token[len(snap.Token)-4:] {
		t.Fatalf("token hint %+v", snap)
	}

	client := &http.Client{Timeout: 3 * time.Second}
	base := snap.LocalURL

	res, err := client.Get(base + "/pack.json")
	if err != nil {
		t.Fatal(err)
	}
	_ = res.Body.Close()
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unauth %d", res.StatusCode)
	}

	req, err := http.NewRequest(http.MethodGet, base+"/pack.json", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+snap.Token)
	res, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(res.Body)
	_ = res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("pack %d %s", res.StatusCode, body)
	}
	if !bytes.Contains(body, []byte("cherry.training_pack.v1")) {
		t.Fatalf("pack body %s", body)
	}

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	_ = w.WriteField("slot", "A")
	_ = w.WriteField("name", "v-colab")
	_ = w.WriteField("note", "T4")
	part, err := w.CreateFormFile("file", "cherry_adapter_worker_A.zip")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte("PK\x03\x04fakezip")); err != nil {
		t.Fatal(err)
	}
	_ = w.Close()
	req, err = http.NewRequest(http.MethodPost, base+"/checkpoint", &buf)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+snap.Token)
	req.Header.Set("Content-Type", w.FormDataContentType())
	res, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	payload, _ := io.ReadAll(res.Body)
	_ = res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("checkpoint %d %s", res.StatusCode, payload)
	}
	var out map[string]any
	if err := json.Unmarshal(payload, &out); err != nil {
		t.Fatal(err)
	}
	if out["ok"] != true || out["inference"] != false {
		t.Fatalf("payload %+v", out)
	}
	if registered == "" {
		t.Fatal("version not registered")
	}
	if _, err := os.Stat(filepath.Join(checkDir, registered)); err != nil {
		t.Fatal(err)
	}

	health, err := client.Get(base + "/health")
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := io.ReadAll(health.Body)
	_ = health.Body.Close()
	if !bytes.Contains(raw, []byte(`"inference":false`)) {
		t.Fatalf("health %s", raw)
	}

	stopped := svc.Stop()
	if stopped.Status != StatusIdle || stopped.PublicURL != "" || stopped.Token != "" {
		t.Fatalf("stop %+v", stopped)
	}
}

func waitRunning(t *testing.T, svc *Service) Snapshot {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		snap := svc.Snapshot()
		if snap.Status == StatusRunning {
			return snap
		}
		if snap.Status == StatusFailed {
			t.Fatalf("failed: %s", snap.Note)
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("not running")
	return Snapshot{}
}
