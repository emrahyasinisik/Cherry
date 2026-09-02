package gdpr

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRedactEmailAndSecrets(t *testing.T) {
	in := "Mail ada@icerde.dev ve anahtar sk-testABCDEFGH. Kod 052533. Bearer abcdef0123456789."
	out, counts := Redact(in)
	if counts.Total() < 3 {
		t.Fatalf("counts %#v out=%s", counts, out)
	}
	if stringsContains(out, "ada@icerde.dev") || stringsContains(out, "sk-test") || stringsContains(out, "052533") {
		t.Fatalf("leaked: %s", out)
	}
}

func TestScanOutput(t *testing.T) {
	out, counts := Scan("yaz bana user@x.com")
	if counts["email"] != 1 || stringsContains(out, "user@x.com") {
		t.Fatalf("%s %#v", out, counts)
	}
}

func TestReadFileStaysInRoot(t *testing.T) {
	root := t.TempDir()
	inner := filepath.Join(root, "frontend")
	if err := os.MkdirAll(inner, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inner, "app.tsx"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	data, err := ReadFile(root, "frontend/app.tsx")
	if err != nil || string(data) != "ok" {
		t.Fatalf("%v %s", err, data)
	}
	if _, err := ReadFile(root, "../secret"); err == nil {
		t.Fatal("expected path error")
	}
}

func stringsContains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})())
}
