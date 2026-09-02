package llm

import (
	"context"
	"strings"
	"testing"

	"github.com/icerde/api/internal/store"
)

func TestCompleteRedactsAndAudits(t *testing.T) {
	mem := store.NewMemory()
	ctx := context.Background()
	if err := Seed(ctx, mem); err != nil {
		t.Fatal(err)
	}
	svc := &Service{Store: mem, Completer: MockCompleter{}}
	out, err := svc.Complete(ctx, CompleteInput{
		UserID:     "u1",
		ProjectID:  "p1",
		Purpose:    "codegen",
		LegalBasis: "contract",
		Prompt:     "Ada için ada@icerde.dev ile kahve uygulaması yaz. Kod 123456.",
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.Text, "ada@icerde.dev") || strings.Contains(out.Text, "123456") {
		t.Fatalf("pii leaked: %s", out.Text)
	}
	if out.InputN == 0 {
		t.Fatal("expected input redactions")
	}
	if !strings.Contains(out.Text, "v1.0") {
		t.Fatalf("version in output: %s", out.Text)
	}
	events, err := mem.ListAudit(ctx, "u1")
	if err != nil || len(events) != 1 {
		t.Fatalf("audit %v %#v", err, events)
	}
	if strings.Contains(events[0].PromptPreview, "ada@icerde.dev") {
		t.Fatalf("audit leaked %s", events[0].PromptPreview)
	}
}

func TestVersionSwitchChangesLaterAnswers(t *testing.T) {
	mem := store.NewMemory()
	ctx := context.Background()
	if err := Seed(ctx, mem); err != nil {
		t.Fatal(err)
	}
	svc := &Service{Store: mem, Completer: MockCompleter{}}
	first, err := svc.Complete(ctx, CompleteInput{UserID: "u1", Purpose: "codegen", LegalBasis: "contract", Prompt: "ekran planı"})
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.SetActiveA(ctx, "ver-a-2"); err != nil {
		t.Fatal(err)
	}
	second, err := svc.Complete(ctx, CompleteInput{UserID: "u1", Purpose: "codegen", LegalBasis: "contract", Prompt: "ekran planı"})
	if err != nil {
		t.Fatal(err)
	}
	if first.Text == second.Text {
		t.Fatal("switch must change subsequent answers")
	}
	if !strings.Contains(second.Text, "v1.1") {
		t.Fatalf("got %s", second.Text)
	}
}
