package factory

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/icerde/api/internal/activate"
	"github.com/icerde/api/internal/llm"
	"github.com/icerde/api/internal/maestro"
	"github.com/icerde/api/internal/opencode"
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
	fake := &opencode.Fake{}
	svc.OpenCode = fake
	svc.MaestroRun = skipMaestro{}
	svc.Activator = &fakeActivate{}

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
	for _, rel := range []string{"README.md", "frontend/app/index.tsx", "backend/main.go", "maestro/login.yaml", "preview/home.html", "llm/plan.md", "llm/opencode.ran", "opencode.json", "AGENTS.md", "icerde.zip"} {
		if _, err := os.Stat(filepath.Join(got.RootPath, rel)); err != nil {
			t.Fatalf("%s: %v", rel, err)
		}
	}
	if fake.Ran != 1 {
		t.Fatalf("opencode ran %d", fake.Ran)
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

func TestSendMessageContinuesHeadlessOpenCode(t *testing.T) {
	mem := store.NewMemory()
	svc := New(mem, t.TempDir())
	svc.StepDelay = 0
	svc.AutoRun = false
	ctx := context.Background()
	if err := llm.Seed(ctx, mem); err != nil {
		t.Fatal(err)
	}
	svc.LLM = &llm.Service{Store: mem, Completer: llm.MockCompleter{}}
	fake := &opencode.Fake{}
	svc.OpenCode = fake
	project, err := svc.Create(ctx, "u1", "Kahve sipariş", "Mahalle kahvesi için sipariş ve kuyruk uygulaması.", "EXPO")
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.RunSync(ctx, project.ID); err != nil {
		t.Fatal(err)
	}
	if fake.Continue {
		t.Fatal("first write must not continue")
	}
	if _, err := svc.SendMessage(ctx, "u1", project.ID, "Girişe ada@icerde.dev QR ekle"); err != nil {
		t.Fatal(err)
	}
	if !fake.Continue {
		t.Fatal("chat follow-up must continue headless OpenCode")
	}
	if strings.Contains(fake.Prompt, "ada@icerde.dev") {
		t.Fatalf("chat leaked email: %s", fake.Prompt)
	}
	logs, err := svc.Logs(ctx, "u1", project.ID)
	if err != nil {
		t.Fatal(err)
	}
	var sawUser, sawAgent bool
	for _, line := range logs {
		if line.Role == store.RoleUser && strings.Contains(line.Message, "QR") {
			sawUser = true
		}
		if line.Role == store.RoleAgent {
			sawAgent = true
		}
	}
	if !sawUser || !sawAgent {
		t.Fatalf("chat roles user=%v agent=%v", sawUser, sawAgent)
	}
}

func TestPipelineWithoutOpenCodeKeepsScaffold(t *testing.T) {
	mem := store.NewMemory()
	svc := New(mem, t.TempDir())
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
	logs, err := svc.Logs(ctx, "u1", project.ID)
	if err != nil {
		t.Fatal(err)
	}
	joined := ""
	for _, line := range logs {
		joined += line.Message
	}
	if !strings.Contains(joined, "OpenCode bağlı değil") {
		t.Fatalf("logs %s", joined)
	}
}

func TestOpenCodePromptRedactsEmail(t *testing.T) {
	mem := store.NewMemory()
	svc := New(mem, t.TempDir())
	svc.StepDelay = 0
	svc.AutoRun = false
	ctx := context.Background()
	if err := llm.Seed(ctx, mem); err != nil {
		t.Fatal(err)
	}
	svc.LLM = &llm.Service{Store: mem, Completer: llm.MockCompleter{}}
	fake := &opencode.Fake{}
	svc.OpenCode = fake
	project, err := svc.Create(ctx, "u1", "Kahve", "ada@icerde.dev için sipariş kuyruğu uygulaması.", "EXPO")
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.RunSync(ctx, project.ID); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(fake.Prompt, "ada@icerde.dev") {
		t.Fatalf("prompt leaked email: %s", fake.Prompt)
	}
	if !strings.Contains(fake.Prompt, "[REDACTED_EMAIL]") {
		t.Fatalf("expected redaction in %s", fake.Prompt)
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

func TestActivateAndRunMaestroNeverPassWithoutDevice(t *testing.T) {
	mem := store.NewMemory()
	svc := New(mem, t.TempDir())
	svc.StepDelay = 0
	svc.AutoRun = false
	ctx := context.Background()
	if err := llm.Seed(ctx, mem); err != nil {
		t.Fatal(err)
	}
	svc.LLM = &llm.Service{Store: mem, Completer: llm.MockCompleter{}}
	svc.OpenCode = &opencode.Fake{}
	act := &fakeActivate{}
	svc.Activator = act
	svc.MaestroRun = skipMaestro{}
	project, err := svc.Create(ctx, "u1", "Kahve sipariş", "Mahalle kahvesi için sipariş ve kuyruk uygulaması.", "EXPO")
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.RunSync(ctx, project.ID); err != nil {
		t.Fatal(err)
	}
	if act.snap.Status != activate.StatusIdle {
		t.Fatalf("pipeline must stop child, got %s", act.snap.Status)
	}
	if _, err := svc.Activate(ctx, "u1", project.ID); err != nil {
		t.Fatal(err)
	}
	if act.snap.Status != activate.StatusRunning || act.snap.Port != 47001 {
		t.Fatalf("%+v", act.snap)
	}
	if _, err := svc.RunMaestro(ctx, "u1", project.ID); err != nil {
		t.Fatal(err)
	}
	studio, err := svc.Maestro(ctx, "u1", project.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(studio.Flows) == 0 {
		t.Fatal("expected flows")
	}
	for _, flow := range studio.Flows {
		if flow.Result == store.MaestroPassed {
			t.Fatal("passed without a device")
		}
		if flow.Result != store.MaestroSkipped {
			t.Fatalf("got %s", flow.Result)
		}
		if !strings.Contains(flow.Note, "127.0.0.1:47001") {
			t.Fatalf("local url missing: %s", flow.Note)
		}
	}
	if _, err := svc.Deactivate(ctx, "u1", project.ID); err != nil {
		t.Fatal(err)
	}
	if act.snap.Status != activate.StatusIdle {
		t.Fatalf("%s", act.snap.Status)
	}
}

type skipMaestro struct{}

func (skipMaestro) RunDir(_ context.Context, maestroDir, localURL string) maestro.Report {
	report := maestro.Report{DeviceStatus: "none", Note: "Cihaz yok. SKIPPED — geçti sayılmaz."}
	entries, err := filepath.Glob(filepath.Join(maestroDir, "*.yaml"))
	if err != nil {
		report.Note = err.Error()
		return report
	}
	for _, path := range entries {
		name := strings.TrimSuffix(filepath.Base(path), ".yaml")
		note := "Emülatör yok. SKIPPED — geçti sayılmaz."
		if localURL != "" {
			note += " Yerel API: " + localURL
		}
		report.Flows = append(report.Flows, maestro.FlowResult{
			ID:     name,
			Result: store.MaestroSkipped,
			Note:   note,
		})
	}
	return report
}

type fakeActivate struct {
	snap activate.Snapshot
}

func (f *fakeActivate) Start(_ context.Context, _, _ string) (activate.Snapshot, error) {
	f.snap = activate.Snapshot{
		Status: activate.StatusRunning,
		URL:    "http://127.0.0.1:47001",
		Port:   47001,
		PID:    1,
		Note:   "fake running",
	}
	return f.snap, nil
}

func (f *fakeActivate) Stop(_ string) activate.Snapshot {
	f.snap = activate.Snapshot{Status: activate.StatusIdle, Note: "Yerel API durdu."}
	return f.snap
}

func (f *fakeActivate) Snapshot(_ string) activate.Snapshot {
	if f.snap.Status == "" {
		return activate.Snapshot{Status: activate.StatusIdle, Note: "Yerel API kapalı."}
	}
	return f.snap
}
