package llm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cherry/api/internal/gdpr"
	"github.com/cherry/api/internal/store"
)

type Completer interface {
	Channel() string
	Complete(ctx context.Context, version store.LlmVersion, prompt string) (string, error)
}

type Service struct {
	Store     store.Store
	Completer Completer
	q         *queue
}

func (s *Service) ensureQueue() *queue {
	if s.q == nil {
		s.q = newQueue()
	}
	return s.q
}

func Seed(ctx context.Context, st store.Store) error {
	versions := []store.LlmVersion{
		{ID: "ver-a-1", Slot: store.SlotA, Name: "v1.0", Note: "İşçi A — stub / yerel", CreatedAt: time.Now().UTC()},
		{ID: "ver-a-2", Slot: store.SlotA, Name: "v1.1", Note: "İşçi A — alternatif pointer (cevap değişir)", CreatedAt: time.Now().UTC()},
		{ID: "ver-b-1", Slot: store.SlotB, Name: "v1.0", Note: "İşçi B — aynı tarif, yoğunluk", CreatedAt: time.Now().UTC()},
		{ID: "ver-b-2", Slot: store.SlotB, Name: "v1.1", Note: "İşçi B — alternatif pointer (cevap değişir)", CreatedAt: time.Now().UTC()},
	}
	for _, version := range versions {
		if err := st.PutLlmVersion(ctx, version); err != nil {
			return err
		}
	}
	state, err := st.GetLlmState(ctx)
	if err != nil {
		return err
	}
	if state.ActiveAID == "" {
		state.ActiveAID = "ver-a-1"
	}
	if state.ActiveBID == "" {
		state.ActiveBID = "ver-b-1"
	}
	return st.SetLlmState(ctx, state)
}

func (s *Service) versionFor(ctx context.Context, slot store.LlmSlot) (store.LlmVersion, error) {
	state, err := s.Store.GetLlmState(ctx)
	if err != nil {
		return store.LlmVersion{}, err
	}
	id := state.ActiveAID
	if slot == store.SlotB {
		id = state.ActiveBID
	}
	if id == "" {
		return store.LlmVersion{}, store.ErrNotFound
	}
	version, err := s.Store.GetLlmVersion(ctx, id)
	if err != nil {
		return store.LlmVersion{}, err
	}
	return *version, nil
}

func (s *Service) ActiveVersion(ctx context.Context) (store.LlmVersion, error) {
	return s.versionFor(ctx, store.SlotA)
}

func (s *Service) SetActive(ctx context.Context, versionID string) error {
	version, err := s.Store.GetLlmVersion(ctx, versionID)
	if err != nil {
		return err
	}
	state, err := s.Store.GetLlmState(ctx)
	if err != nil {
		return err
	}
	switch version.Slot {
	case store.SlotA:
		state.ActiveAID = version.ID
	case store.SlotB:
		state.ActiveBID = version.ID
	default:
		return store.ErrValidation
	}
	return s.Store.SetLlmState(ctx, state)
}

func (s *Service) SetActiveA(ctx context.Context, versionID string) error {
	return s.SetActive(ctx, versionID)
}

func (s *Service) RegisterVersion(ctx context.Context, slotRaw, name, note, checkpointRef string) (store.LlmVersion, error) {
	name = strings.TrimSpace(name)
	note = strings.TrimSpace(note)
	checkpointRef = strings.TrimSpace(checkpointRef)
	if name == "" || checkpointRef == "" {
		return store.LlmVersion{}, store.ErrValidation
	}
	var slot store.LlmSlot
	switch strings.ToUpper(strings.TrimSpace(slotRaw)) {
	case string(store.SlotA):
		slot = store.SlotA
	case string(store.SlotB):
		slot = store.SlotB
	default:
		return store.LlmVersion{}, store.ErrValidation
	}
	version := store.LlmVersion{
		ID:            "ver-colab-" + store.NewID()[:12],
		Slot:          slot,
		Name:          name,
		Note:          note,
		CheckpointRef: checkpointRef,
		CreatedAt:     time.Now().UTC(),
	}
	if version.Note == "" {
		version.Note = "Colab QLoRA — " + checkpointRef
	}
	if err := s.Store.PutLlmVersion(ctx, version); err != nil {
		return store.LlmVersion{}, err
	}
	return version, nil
}

func (s *Service) SetMcpRoot(ctx context.Context, root string) error {
	state, err := s.Store.GetLlmState(ctx)
	if err != nil {
		return err
	}
	state.McpRoot = strings.TrimSpace(root)
	return s.Store.SetLlmState(ctx, state)
}

type CompleteInput struct {
	UserID     string
	ProjectID  string
	Purpose    string
	LegalBasis string
	Prompt     string
}

type CompleteResult struct {
	Text    string
	Version store.LlmVersion
	Slot    store.LlmSlot
	Channel string
	InputN  int
	OutputN int
	AuditID string
}

func (s *Service) Complete(ctx context.Context, in CompleteInput) (CompleteResult, error) {
	if s.Completer == nil {
		return CompleteResult{}, fmt.Errorf("%w: completer yok", store.ErrLLMFailed)
	}
	slot, err := s.ensureQueue().acquire(ctx)
	if err != nil {
		return CompleteResult{}, err
	}
	held := lease{Slot: slot, q: s.q}
	defer held.Release()
	version, err := s.versionFor(ctx, slot)
	if err != nil {
		return CompleteResult{}, err
	}
	redacted, inCounts := gdpr.Redact(in.Prompt)
	raw, err := s.Completer.Complete(ctx, version, redacted)
	if err != nil {
		return CompleteResult{}, fmt.Errorf("%w: %v", store.ErrLLMFailed, err)
	}
	safe, outCounts := gdpr.Scan(raw)
	event := store.AuditEvent{
		UserID:           in.UserID,
		ProjectID:        in.ProjectID,
		Purpose:          in.Purpose,
		LegalBasis:       in.LegalBasis,
		Slot:             slot,
		VersionID:        version.ID,
		VersionName:      version.Name,
		Channel:          s.Completer.Channel(),
		InputRedactions:  inCounts.Total(),
		OutputRedactions: outCounts.Total(),
		PromptPreview:    gdpr.Preview(redacted, 180),
		OutputPreview:    gdpr.Preview(safe, 180),
		CreatedAt:        time.Now().UTC(),
	}
	if err := s.Store.AddAudit(ctx, event); err != nil {
		return CompleteResult{}, err
	}
	return CompleteResult{
		Text:    safe,
		Version: version,
		Slot:    slot,
		Channel: s.Completer.Channel(),
		InputN:  inCounts.Total(),
		OutputN: outCounts.Total(),
		AuditID: event.ID,
	}, nil
}

func (s *Service) ReadProjectFile(ctx context.Context, rel string) (string, error) {
	state, err := s.Store.GetLlmState(ctx)
	if err != nil {
		return "", err
	}
	data, err := gdpr.ReadFile(state.McpRoot, rel)
	if err != nil {
		return "", err
	}
	safe, _ := gdpr.Redact(string(data))
	return safe, nil
}

type StatusView struct {
	Channel  string
	Queued   int
	BusyA    bool
	BusyB    bool
	VersionA string
	VersionB string
	LastSlot store.LlmSlot
}

func (s *Service) Snapshot(ctx context.Context) (StatusView, error) {
	occ := s.ensureQueue().snapshot()
	a, err := s.versionFor(ctx, store.SlotA)
	if err != nil {
		return StatusView{}, err
	}
	b, err := s.versionFor(ctx, store.SlotB)
	if err != nil {
		return StatusView{}, err
	}
	channel := "none"
	if s.Completer != nil {
		channel = s.Completer.Channel()
	}
	return StatusView{
		Channel:  channel,
		Queued:   occ.Queued,
		BusyA:    occ.BusyA,
		BusyB:    occ.BusyB,
		VersionA: a.Name,
		VersionB: b.Name,
		LastSlot: occ.Last,
	}, nil
}

func (s *Service) Status(ctx context.Context) (store.LlmVersion, string, error) {
	version, err := s.versionFor(ctx, store.SlotA)
	if err != nil {
		return store.LlmVersion{}, "", err
	}
	channel := "none"
	if s.Completer != nil {
		channel = s.Completer.Channel()
	}
	return version, channel, nil
}
