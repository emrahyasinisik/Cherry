package factory

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/icerde/api/internal/llm"
	"github.com/icerde/api/internal/store"
)

func TestPipelineWritesTreeAndSkipsMaestro(t *testing.T) {
	mem := store.NewMemory()
	root := t.TempDir()
	svc := New(mem, root)
	svc.StepDelay = 0
	svc.AutoRun = false
	ctx := context.Background()
	if err := llm.Seed(ctx, mem); err != nil {
		t.Fatal(err)
	}
	svc.LLM = &llm.Service{Store: mem, Completer: llm.MockCompleter{}}

	project, err := svc.Create(ctx, "u1", "Kahve sipariş", "Mahalle kahvesi için sipariş ve kuyruk uygulaması.", "EXPO")
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.RunSync(ctx, project.ID); err != nil {
		t.Fatal(err)
	}
	got, err := svc.Get(ctx, "u1", project.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != store.StatusReady {
		t.Fatalf("status %s", got.Status)
	}
	for _, rel := range []string{"README.md", "frontend/app/index.tsx", "backend/main.go", "maestro/login.yaml", "preview/home.html", "llm/plan.md", "icerde.zip"} {
		if _, err := os.Stat(filepath.Join(got.RootPath, rel)); err != nil {
			t.Fatalf("%s: %v", rel, err)
		}
	}
	studio, err := svc.Maestro(ctx, "u1", project.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !studio.Ready || len(studio.Flows) == 0 || len(studio.Screens) == 0 {
		t.Fatalf("maestro %#v", studio)
	}
	if studio.Flows[0].Result != store.MaestroSkipped {
		t.Fatalf("must skip without emulator, got %s", studio.Flows[0].Result)
	}
}

func TestStacksExhaustive(t *testing.T) {
	for _, stack := range []string{"EXPO", "FLUTTER", "NATIVE"} {
		if _, err := parseStack(stack); err != nil {
			t.Fatalf("%s: %v", stack, err)
		}
		if _, err := stackLabel(store.ProjectStack(stack)); err != nil {
			t.Fatalf("label %s: %v", stack, err)
		}
		if _, err := frontendKind(store.ProjectStack(stack)); err != nil {
			t.Fatalf("kind %s: %v", stack, err)
		}
	}
	if _, err := parseStack("UNITY"); err == nil {
		t.Fatal("expected validation")
	}
}

func TestForeignProjectHidden(t *testing.T) {
	mem := store.NewMemory()
	svc := New(mem, t.TempDir())
	svc.StepDelay = 0
	svc.AutoRun = false
	ctx := context.Background()
	project, err := svc.Create(ctx, "owner", "Uygulama", "Kısa bir brif metni burada.", "FLUTTER")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Get(ctx, "other", project.ID); err != store.ErrNotFound {
		t.Fatalf("got %v", err)
	}
}

func TestSlug(t *testing.T) {
	if slugify("Kahve Sipariş!") != "kahve-sipari" && !strings.Contains(slugify("Kahve Sipariş"), "kahve") {
		t.Fatalf("%q", slugify("Kahve Sipariş"))
	}
}
