package graph

import (
	"context"
	"time"

	"github.com/icerde/api/internal/store"
)

func (r *Resolver) llmAdminPayload(ctx context.Context, userID string) (*LlmAdmin, error) {
	state, err := r.LLM.Store.GetLlmState(ctx)
	if err != nil {
		return nil, err
	}
	versions, err := r.LLM.Store.ListLlmVersions(ctx, store.SlotA)
	if err != nil {
		return nil, err
	}
	cards := make([]*LlmVersion, 0, len(versions))
	for _, version := range versions {
		item := version
		cards = append(cards, &LlmVersion{ID: item.ID, Name: item.Name, Note: item.Note})
	}
	active := state.ActiveAID
	audits, err := r.LLM.Store.ListAudit(ctx, userID)
	if err != nil {
		return nil, err
	}
	completions := make([]*LlmCompletion, 0, len(audits))
	for i, event := range audits {
		if i >= 20 {
			break
		}
		item := event
		completions = append(completions, &LlmCompletion{
			At:               item.CreatedAt.UTC().Format(time.RFC3339),
			Purpose:          item.Purpose,
			VersionName:      item.VersionName,
			Channel:          item.Channel,
			InputRedactions:  item.InputRedactions,
			OutputRedactions: item.OutputRedactions,
			PromptPreview:    item.PromptPreview,
			OutputPreview:    item.OutputPreview,
		})
	}
	return &LlmAdmin{
		Gdpr:       true,
		ActiveSlot: "A",
		McpRoot:    state.McpRoot,
		SlotA: &LlmSlotCard{
			Slot:            "A",
			Wired:           true,
			Role:            "İşçi 1 — kapasite",
			ActiveVersionID: &active,
			Versions:        cards,
		},
		SlotB: &LlmSlotCard{
			Slot:     "B",
			Wired:    false,
			Role:     "İşçi 2 — yoğunluk, henüz yok",
			Versions: []*LlmVersion{},
		},
		Completions: completions,
	}, nil
}
