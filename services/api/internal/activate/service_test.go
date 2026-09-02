package activate

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStartHealthAndStop(t *testing.T) {
	root := t.TempDir()
	backend := filepath.Join(root, "backend")
	if err := os.MkdirAll(backend, 0o755); err != nil {
		t.Fatal(err)
	}
	src := `package main
import (
  "encoding/json"
  "log"
  "net/http"
  "os"
)
func main() {
  addr := os.Getenv("ICERDE_CUSTOMER_ADDR")
  if addr == "" { addr = "127.0.0.1:18080" }
  http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
    _ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
  })
  log.Fatal(http.ListenAndServe(addr, nil))
}
`
	if err := os.WriteFile(filepath.Join(backend, "main.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	svc := New()
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	snap, err := svc.Start(ctx, "p1", root)
	if err != nil {
		t.Fatal(err)
	}
	if snap.Status != StatusRunning || snap.Port < 47000 || snap.Port > 47999 {
		t.Fatalf("%+v", snap)
	}
	res, err := http.Get(snap.URL + "/health")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(res.Body)
	_ = res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatalf("health %d %s", res.StatusCode, body)
	}
	idle := svc.Stop("p1")
	if idle.Status != StatusIdle {
		t.Fatalf("%+v", idle)
	}
}

func TestStatusLabelExhaustive(t *testing.T) {
	for _, status := range []Status{StatusIdle, StatusStarting, StatusRunning, StatusStopping, StatusFailed, Status("x")} {
		if status.Label() == "" {
			t.Fatalf("empty %s", status)
		}
	}
}
