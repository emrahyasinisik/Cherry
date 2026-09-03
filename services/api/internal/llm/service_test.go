package llm

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cherry/api/internal/store"
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
		Prompt:     "Ada için ada@cherry.dev ile kahve uygulaması yaz. Kod 123456.",
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.Text, "ada@cherry.dev") || strings.Contains(out.Text, "123456") {
		t.Fatalf("pii leaked: %s", out.Text)
	}
	if out.InputN == 0 {
		t.Fatal("expected input redactions")
	}
	if !strings.Contains(out.Text, "v1.0") {
		t.Fatalf("version in output: %s", out.Text)
	}
	if out.Channel != "mock" {
		t.Fatalf("channel %s", out.Channel)
	}
	events, err := mem.ListAudit(ctx, "u1")
	if err != nil || len(events) != 1 {
		t.Fatalf("audit %v %#v", err, events)
	}
	if strings.Contains(events[0].PromptPreview, "ada@cherry.dev") {
		t.Fatalf("audit leaked %s", events[0].PromptPreview)
	}
}

func TestEffectiveChannelPrefersColabTunnel(t *testing.T) {
	mem := store.NewMemory()
	ctx := context.Background()
	if err := Seed(ctx, mem); err != nil {
		t.Fatal(err)
	}
	svc := &Service{Store: mem, Completer: MockCompleter{}}
	svc.SetColabInferenceURL("https://example.invalid/v1")
	svc.colabMu.Lock()
	if svc.colabStopHealth != nil {
		close(svc.colabStopHealth)
		svc.colabStopHealth = nil
	}
	svc.colabInferenceURL = "https://example.invalid/v1"
	svc.colabStatus = ColabInferenceConnected
	svc.colabMu.Unlock()
	snap, err := svc.Snapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if snap.Channel != "colab-tunnel" {
		t.Fatalf("channel %s", snap.Channel)
	}
	base, key := svc.OpenCodeEndpoint()
	if base != "https://example.invalid/v1" {
		t.Fatalf("base %s", base)
	}
	if key != "cherry-colab" {
		t.Fatalf("key %s", key)
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
	if first.Slot != store.SlotA {
		t.Fatalf("prefer A when idle, got %s", first.Slot)
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

func TestIdlePrefersAThenB(t *testing.T) {
	mem := store.NewMemory()
	ctx := context.Background()
	if err := Seed(ctx, mem); err != nil {
		t.Fatal(err)
	}
	hold := newHoldCompleter()
	svc := &Service{Store: mem, Completer: hold}
	started := make(chan store.LlmSlot, 2)
	errc := make(chan error, 2)
	go func() {
		out, err := svc.Complete(ctx, CompleteInput{UserID: "u1", Purpose: "codegen", LegalBasis: "contract", Prompt: "bir"})
		errc <- err
		started <- out.Slot
	}()
	go func() {
		out, err := svc.Complete(ctx, CompleteInput{UserID: "u2", Purpose: "codegen", LegalBasis: "contract", Prompt: "iki"})
		errc <- err
		started <- out.Slot
	}()
	<-hold.entered
	<-hold.entered
	snap, err := svc.Snapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !snap.BusyA || !snap.BusyB || snap.Queued != 0 {
		t.Fatalf("%+v", snap)
	}
	close(hold.release)
	slots := map[store.LlmSlot]int{}
	for i := 0; i < 2; i++ {
		if err := <-errc; err != nil {
			t.Fatal(err)
		}
		slots[<-started]++
	}
	if slots[store.SlotA] != 1 || slots[store.SlotB] != 1 {
		t.Fatalf("slots %+v", slots)
	}
}

func TestThirdJobQueuesUntilWorkerFrees(t *testing.T) {
	mem := store.NewMemory()
	ctx := context.Background()
	if err := Seed(ctx, mem); err != nil {
		t.Fatal(err)
	}
	hold := newHoldCompleter()
	svc := &Service{Store: mem, Completer: hold}
	errc := make(chan error, 3)
	for i := 0; i < 2; i++ {
		go func() {
			_, err := svc.Complete(ctx, CompleteInput{UserID: "u", Purpose: "codegen", LegalBasis: "contract", Prompt: "x"})
			errc <- err
		}()
	}
	<-hold.entered
	<-hold.entered
	go func() {
		_, err := svc.Complete(ctx, CompleteInput{UserID: "u", Purpose: "codegen", LegalBasis: "contract", Prompt: "üç"})
		errc <- err
	}()
	waitQueued(t, svc)
	hold.release <- struct{}{}
	<-hold.entered
	close(hold.release)
	for i := 0; i < 3; i++ {
		if err := <-errc; err != nil {
			t.Fatal(err)
		}
	}
}

func TestInFlightKeepsOldVersion(t *testing.T) {
	mem := store.NewMemory()
	ctx := context.Background()
	if err := Seed(ctx, mem); err != nil {
		t.Fatal(err)
	}
	hold := newHoldCompleter()
	svc := &Service{Store: mem, Completer: hold}
	done := make(chan CompleteResult, 1)
	go func() {
		out, err := svc.Complete(ctx, CompleteInput{UserID: "u1", Purpose: "codegen", LegalBasis: "contract", Prompt: "uçuş"})
		if err != nil {
			t.Errorf("complete: %v", err)
		}
		done <- out
	}()
	<-hold.entered
	if err := svc.SetActive(ctx, "ver-a-2"); err != nil {
		t.Fatal(err)
	}
	close(hold.release)
	out := <-done
	if !strings.Contains(out.Text, "v1.0") || strings.Contains(out.Text, "v1.1") {
		t.Fatalf("in-flight must keep v1.0: %s", out.Text)
	}
	next, err := svc.Complete(ctx, CompleteInput{UserID: "u1", Purpose: "codegen", LegalBasis: "contract", Prompt: "sonraki"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(next.Text, "v1.1") {
		t.Fatalf("queued/new must use new pointer: %s", next.Text)
	}
}

func TestBPointerDoesNotChangeA(t *testing.T) {
	mem := store.NewMemory()
	ctx := context.Background()
	if err := Seed(ctx, mem); err != nil {
		t.Fatal(err)
	}
	svc := &Service{Store: mem, Completer: MockCompleter{}}
	if err := svc.SetActive(ctx, "ver-b-2"); err != nil {
		t.Fatal(err)
	}
	out, err := svc.Complete(ctx, CompleteInput{UserID: "u1", Purpose: "codegen", LegalBasis: "contract", Prompt: "yalnız A"})
	if err != nil {
		t.Fatal(err)
	}
	if out.Slot != store.SlotA || !strings.Contains(out.Text, "v1.0") || !strings.Contains(out.Text, "LLM A") {
		t.Fatalf("%s %s", out.Slot, out.Text)
	}
}

type holdCompleter struct {
	entered chan int
	release chan struct{}
	n       atomic.Int32
}

func newHoldCompleter() *holdCompleter {
	return &holdCompleter{entered: make(chan int, 8), release: make(chan struct{})}
}

func (h *holdCompleter) Channel() string { return "mock" }

func (h *holdCompleter) Complete(ctx context.Context, version store.LlmVersion, prompt string) (string, error) {
	n := int(h.n.Add(1))
	h.entered <- n
	select {
	case <-h.release:
	case <-ctx.Done():
		return "", ctx.Err()
	}
	return MockCompleter{}.Complete(ctx, version, prompt)
}

func waitQueued(t *testing.T, svc *Service) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		snap, err := svc.Snapshot(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if snap.Queued >= 1 {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("queue never filled")
}
