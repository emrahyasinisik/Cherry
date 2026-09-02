package llm

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/icerde/api/internal/store"
)

func TestBuildPackRedactsAndSkipsSecrets(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "README.md"), "Kahve uygulaması. İletişim ada@icerde.dev")
	mustWrite(t, filepath.Join(root, "llm", "plan.md"), "Ekranları böl. PII yok.")
	mustWrite(t, filepath.Join(root, "frontend", "src", "domain", "entities", "item.ts"), "export type Item = { id: string };\n")
	mustWrite(t, filepath.Join(root, ".env"), "SECRET=sk-live-should-never-leave")
	mustWrite(t, filepath.Join(root, "preview", "home.html"), "<html>preview</html>")
	mustWrite(t, filepath.Join(root, "maestro", "login.yaml"), "appId: demo\n---\n- launchApp\n")

	pack := BuildPack(context.Background(), PackInput{
		Projects: []store.Project{{
			ID:       "p1",
			Name:     "Kahve",
			Brief:    "Ada için ada@icerde.dev ile kahve. Kod 123456.",
			Stack:    store.StackExpo,
			RootPath: root,
			Status:   store.StatusReady,
		}},
		Audits: []store.AuditEvent{{
			ID:            "a1",
			UserID:        "u1",
			Purpose:       "codegen",
			Slot:          store.SlotA,
			VersionName:   "v1.0",
			PromptPreview: "plan yaz",
			OutputPreview: "İşçi A v1.0 plan",
			CreatedAt:     time.Now().UTC(),
		}},
		Logs: map[string][]store.JobLog{
			"p1": {
				{ProjectID: "p1", Role: store.RoleUser, Message: "ana ekranı sade tut"},
				{ProjectID: "p1", Role: store.RoleAgent, Message: "HomeScreen listesini domain'den çek."},
			},
		},
		Maestro: []MaestroTrace{{
			ProjectID: "p1",
			Name:      "login.yaml",
			YAML:      "appId: demo\n---\n- launchApp\n",
			Result:    "SKIPPED",
			Note:      "Emülatör yok.",
		}},
	})
	blob, err := pack.JSON()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(blob, "ada@icerde.dev") || strings.Contains(blob, "123456") || strings.Contains(blob, "sk-live") {
		t.Fatalf("pii or secret leaked: %s", blob)
	}
	if strings.Contains(blob, "preview/home.html") || strings.Contains(blob, "<html>preview") {
		t.Fatal("preview must stay out of the pack")
	}
	if pack.Stats.LiveExamples == 0 {
		t.Fatal("expected live examples")
	}
	jsonl, err := pack.JSONL()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(jsonl, `"instruction"`) {
		t.Fatalf("jsonl: %s", jsonl)
	}
	foundMaestro := false
	for _, example := range pack.Examples {
		if example.Kind == "maestro" && example.Meta["result"] == "SKIPPED" {
			foundMaestro = true
		}
	}
	if !foundMaestro {
		t.Fatal("maestro SKIPPED missing")
	}
}

func TestBuildPackSeedsWhenEmpty(t *testing.T) {
	pack := BuildPack(context.Background(), PackInput{})
	if pack.Stats.LiveExamples != 0 {
		t.Fatalf("live=%d", pack.Stats.LiveExamples)
	}
	if pack.Stats.SeedExamples < 4 {
		t.Fatalf("seed=%d", pack.Stats.SeedExamples)
	}
	if pack.Schema != PackSchema {
		t.Fatalf("schema %s", pack.Schema)
	}
	raw, err := json.Marshal(pack.Examples[0])
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "instruction") {
		t.Fatal(string(raw))
	}
}

func TestRegisterVersionPointsAtCheckpoint(t *testing.T) {
	mem := store.NewMemory()
	ctx := context.Background()
	if err := Seed(ctx, mem); err != nil {
		t.Fatal(err)
	}
	svc := &Service{Store: mem, Completer: MockCompleter{}}
	version, err := svc.RegisterVersion(ctx, "A", "v1.2-colab", "T4 QLoRA", "icerde_adapter_worker_A.zip")
	if err != nil {
		t.Fatal(err)
	}
	if version.Slot != store.SlotA || version.CheckpointRef == "" {
		t.Fatalf("%#v", version)
	}
	if _, err := svc.RegisterVersion(ctx, "C", "x", "n", "ref"); err == nil {
		t.Fatal("invalid slot must fail")
	}
}

func mustWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
