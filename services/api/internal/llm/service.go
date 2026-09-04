package llm

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/cherry/api/internal/gdpr"
	"github.com/cherry/api/internal/store"
)

type Completer interface {
	Channel() string
	Complete(ctx context.Context, version store.LlmVersion, prompt string) (string, error)
}

type ColabInferenceStatus string

const (
	ColabInferenceOff          ColabInferenceStatus = "OFF"
	ColabInferenceConnected    ColabInferenceStatus = "CONNECTED"
	ColabInferenceDisconnected ColabInferenceStatus = "DISCONNECTED"
	ColabInferenceChecking     ColabInferenceStatus = "CHECKING"
)

type colabSlot struct {
	url    string
	status ColabInferenceStatus
	stop   chan struct{}
}

type Service struct {
	Store     store.Store
	Completer Completer
	q         *queue

	colabMu sync.Mutex
	colab   map[store.LlmSlot]*colabSlot
}

func (s *Service) ensureQueue() *queue {
	if s.q == nil {
		s.q = newQueue()
	}
	return s.q
}

func (s *Service) ensureColab() {
	if s.colab == nil {
		s.colab = map[store.LlmSlot]*colabSlot{
			store.SlotA: {status: ColabInferenceOff},
			store.SlotB: {status: ColabInferenceOff},
		}
	}
}

// LoadColabFromStore restores A/B inference URLs after process start (Mongo/memory).
func (s *Service) LoadColabFromStore(ctx context.Context) error {
	if s.Store == nil {
		return nil
	}
	state, err := s.Store.GetLlmState(ctx)
	if err != nil {
		return err
	}
	if state.ColabURLA != "" {
		s.SetColabInferenceURL(store.SlotA, state.ColabURLA)
	}
	if state.ColabURLB != "" {
		s.SetColabInferenceURL(store.SlotB, state.ColabURLB)
	}
	return nil
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
	slot, err := s.ensureQueue().acquire(ctx)
	if err != nil {
		return CompleteResult{}, err
	}
	held := lease{Slot: slot, q: s.q}
	defer held.Release()
	comp := s.effectiveCompleter(slot)
	if comp == nil {
		return CompleteResult{}, fmt.Errorf("%w: completer yok", store.ErrLLMFailed)
	}
	version, err := s.versionFor(ctx, slot)
	if err != nil {
		return CompleteResult{}, err
	}
	redacted, inCounts := gdpr.Redact(in.Prompt)
	raw, err := comp.Complete(ctx, version, redacted)
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
		Channel:          comp.Channel(),
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
		Channel: comp.Channel(),
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
	if comp := s.effectiveCompleter(store.SlotA); comp != nil {
		channel = comp.Channel()
	}
	if channel != "colab-tunnel" {
		if comp := s.effectiveCompleter(store.SlotB); comp != nil && comp.Channel() == "colab-tunnel" {
			channel = "colab-tunnel"
		} else if channel == "none" {
			if comp := s.effectiveCompleter(store.SlotB); comp != nil {
				channel = comp.Channel()
			}
		}
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

func ParseSlot(raw string) (store.LlmSlot, error) {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case string(store.SlotA):
		return store.SlotA, nil
	case string(store.SlotB):
		return store.SlotB, nil
	default:
		return "", store.ErrValidation
	}
}

func (s *Service) SetColabInferenceURL(slot store.LlmSlot, url string) {
	url = strings.TrimSpace(url)
	s.colabMu.Lock()
	s.ensureColab()
	cur := s.colab[slot]
	if cur == nil {
		cur = &colabSlot{status: ColabInferenceOff}
		s.colab[slot] = cur
	}
	if cur.stop != nil {
		close(cur.stop)
		cur.stop = nil
	}
	if url == "" {
		cur.url = ""
		cur.status = ColabInferenceOff
		s.colabMu.Unlock()
		s.persistColabURL(slot, "")
		return
	}
	cur.url = strings.TrimRight(url, "/")
	cur.status = ColabInferenceChecking
	stop := make(chan struct{})
	cur.stop = stop
	base := cur.url
	s.colabMu.Unlock()
	s.persistColabURL(slot, base)
	go s.healthLoop(slot, base, stop)
}

func (s *Service) persistColabURL(slot store.LlmSlot, url string) {
	if s.Store == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	state, err := s.Store.GetLlmState(ctx)
	if err != nil {
		return
	}
	switch slot {
	case store.SlotA:
		state.ColabURLA = url
	case store.SlotB:
		state.ColabURLB = url
	default:
		return
	}
	_ = s.Store.SetLlmState(ctx, state)
}

func (s *Service) ColabInferenceState(slot store.LlmSlot) (string, ColabInferenceStatus) {
	s.colabMu.Lock()
	defer s.colabMu.Unlock()
	s.ensureColab()
	cur := s.colab[slot]
	if cur == nil {
		return "", ColabInferenceOff
	}
	return cur.url, cur.status
}

func (s *Service) healthLoop(slot store.LlmSlot, baseURL string, stop chan struct{}) {
	client := &http.Client{Timeout: 8 * time.Second}
	checkOnce := func() bool {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/models", nil)
		if err != nil {
			return false
		}
		resp, err := client.Do(req)
		if err != nil {
			return false
		}
		resp.Body.Close()
		return resp.StatusCode < 300
	}
	check := func() bool {
		for attempt := 0; attempt < 3; attempt++ {
			if checkOnce() {
				return true
			}
			if attempt < 2 {
				select {
				case <-stop:
					return false
				case <-time.After(time.Duration(attempt+1) * 400 * time.Millisecond):
				}
			}
		}
		return false
	}

	ok := check()
	s.colabMu.Lock()
	if cur := s.colab[slot]; cur != nil && cur.url == baseURL {
		if ok {
			cur.status = ColabInferenceConnected
		} else {
			cur.status = ColabInferenceDisconnected
		}
	}
	s.colabMu.Unlock()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			ok := check()
			s.colabMu.Lock()
			cur := s.colab[slot]
			if cur == nil || cur.url != baseURL {
				s.colabMu.Unlock()
				return
			}
			if ok {
				cur.status = ColabInferenceConnected
			} else {
				cur.status = ColabInferenceDisconnected
			}
			s.colabMu.Unlock()
		}
	}
}

func (s *Service) effectiveCompleter(slot store.LlmSlot) Completer {
	s.colabMu.Lock()
	s.ensureColab()
	cur := s.colab[slot]
	var url string
	var status ColabInferenceStatus
	if cur != nil {
		url = cur.url
		status = cur.status
	}
	s.colabMu.Unlock()

	if url != "" && status == ColabInferenceConnected {
		return ColabTunnelCompleter{
			BaseURL: url,
			Client:  &http.Client{Timeout: 90 * time.Second},
		}
	}
	return s.Completer
}

// OpenCodeEndpoint returns base URL, API key, and model for the OpenCode CLI for a worker slot.
// Empty slot prefers A if connected, else B, else env defaults.
func (s *Service) OpenCodeEndpoint(slot store.LlmSlot) (baseURL, apiKey, model string) {
	apiKey = strings.TrimSpace(os.Getenv("CHERRY_LLM_API_KEY"))
	baseURL = strings.TrimSpace(os.Getenv("CHERRY_LLM_BASE_URL"))
	model = strings.TrimSpace(firstNonEmptyEnv("CHERRY_OPENCODE_MODEL", "CHERRY_LLM_MODEL"))
	pick := slot
	if pick != store.SlotA && pick != store.SlotB {
		if _, st := s.ColabInferenceState(store.SlotA); st == ColabInferenceConnected {
			pick = store.SlotA
		} else if _, st := s.ColabInferenceState(store.SlotB); st == ColabInferenceConnected {
			pick = store.SlotB
		} else {
			pick = store.SlotA
		}
	}
	s.colabMu.Lock()
	s.ensureColab()
	cur := s.colab[pick]
	var url string
	var status ColabInferenceStatus
	if cur != nil {
		url = cur.url
		status = cur.status
	}
	s.colabMu.Unlock()
	if url != "" && status == ColabInferenceConnected {
		baseURL = url
		if apiKey == "" {
			apiKey = "cherry-colab"
		}
		if model == "" {
			model = "Qwen/Qwen2.5-1.5B-Instruct"
		}
	}
	return baseURL, apiKey, model
}

func firstNonEmptyEnv(keys ...string) string {
	for _, key := range keys {
		if v := strings.TrimSpace(os.Getenv(key)); v != "" {
			return v
		}
	}
	return ""
}

func (s *Service) Status(ctx context.Context) (store.LlmVersion, string, error) {
	version, err := s.versionFor(ctx, store.SlotA)
	if err != nil {
		return store.LlmVersion{}, "", err
	}
	channel := "none"
	comp := s.effectiveCompleter(store.SlotA)
	if comp != nil {
		channel = comp.Channel()
	}
	return version, channel, nil
}
