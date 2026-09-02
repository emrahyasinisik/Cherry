package llm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/icerde/api/internal/gdpr"
	"github.com/icerde/api/internal/store"
)

type Completer interface {
	Channel() string
	Complete(ctx context.Context, version store.LlmVersion, prompt string) (string, error)
}

type Service struct {
	Store     store.Store
	Completer Completer
}

func Seed(ctx context.Context, st store.Store) error {
	v1 := store.LlmVersion{
		ID:        "ver-a-1",
		Slot:      store.SlotA,
		Name:      "v1.0",
		Note:      "İşçi A — stub / yerel",
		CreatedAt: time.Now().UTC(),
	}
	v2 := store.LlmVersion{
		ID:        "ver-a-2",
		Slot:      store.SlotA,
		Name:      "v1.1",
		Note:      "İşçi A — alternatif pointer (cevap değişir)",
		CreatedAt: time.Now().UTC(),
	}
	if err := st.PutLlmVersion(ctx, v1); err != nil {
		return err
	}
	if err := st.PutLlmVersion(ctx, v2); err != nil {
		return err
	}
	state, err := st.GetLlmState(ctx)
	if err != nil {
		return err
	}
	if state.ActiveAID == "" {
		state.ActiveAID = v1.ID
		return st.SetLlmState(ctx, state)
	}
	return nil
}

func (s *Service) ActiveVersion(ctx context.Context) (store.LlmVersion, error) {
	state, err := s.Store.GetLlmState(ctx)
	if err != nil {
		return store.LlmVersion{}, err
	}
	if state.ActiveAID == "" {
		return store.LlmVersion{}, store.ErrNotFound
	}
	version, err := s.Store.GetLlmVersion(ctx, state.ActiveAID)
	if err != nil {
		return store.LlmVersion{}, err
	}
	return *version, nil
}

func (s *Service) SetActiveA(ctx context.Context, versionID string) error {
	version, err := s.Store.GetLlmVersion(ctx, versionID)
	if err != nil {
		return err
	}
	if version.Slot != store.SlotA {
		return store.ErrValidation
	}
	state, err := s.Store.GetLlmState(ctx)
	if err != nil {
		return err
	}
	state.ActiveAID = version.ID
	return s.Store.SetLlmState(ctx, state)
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
	Text     string
	Version  store.LlmVersion
	Channel  string
	InputN   int
	OutputN  int
	AuditID  string
}

func (s *Service) Complete(ctx context.Context, in CompleteInput) (CompleteResult, error) {
	if s.Completer == nil {
		return CompleteResult{}, fmt.Errorf("%w: completer yok", store.ErrLLMFailed)
	}
	version, err := s.ActiveVersion(ctx)
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
		Slot:             store.SlotA,
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

func (s *Service) Status(ctx context.Context) (store.LlmVersion, string, error) {
	version, err := s.ActiveVersion(ctx)
	if err != nil {
		return store.LlmVersion{}, "", err
	}
	channel := "none"
	if s.Completer != nil {
		channel = s.Completer.Channel()
	}
	return version, channel, nil
}
