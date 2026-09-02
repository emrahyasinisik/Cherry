package graph

import (
	"context"
	"sort"
	"time"

	"github.com/icerde/api/internal/store"
)

func (r *Resolver) llmAdminPayload(ctx context.Context, userID string) (*LlmAdmin, error) {
	state, err := r.LLM.Store.GetLlmState(ctx)
	if err != nil {
		return nil, err
	}
	snap, err := r.LLM.Snapshot(ctx)
	if err != nil {
		return nil, err
	}
	slotA, err := r.llmSlotCard(ctx, store.SlotA, state.ActiveAID, snap.BusyA)
	if err != nil {
		return nil, err
	}
	slotB, err := r.llmSlotCard(ctx, store.SlotB, state.ActiveBID, snap.BusyB)
	if err != nil {
		return nil, err
	}
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
			Slot:             string(item.Slot),
			VersionName:      item.VersionName,
			Channel:          item.Channel,
			InputRedactions:  item.InputRedactions,
			OutputRedactions: item.OutputRedactions,
			PromptPreview:    item.PromptPreview,
			OutputPreview:    item.OutputPreview,
		})
	}
	last := string(snap.LastSlot)
	if last == "" {
		last = "A"
	}
	return &LlmAdmin{
		Gdpr:        true,
		ActiveSlot:  last,
		McpRoot:     state.McpRoot,
		Queued:      snap.Queued,
		SlotA:       slotA,
		SlotB:       slotB,
		Completions: completions,
	}, nil
}

func (r *Resolver) llmSlotCard(ctx context.Context, slot store.LlmSlot, activeID string, busy bool) (*LlmSlotCard, error) {
	versions, err := r.LLM.Store.ListLlmVersions(ctx, slot)
	if err != nil {
		return nil, err
	}
	sort.Slice(versions, func(i, j int) bool {
		if versions[i].Name == versions[j].Name {
			return versions[i].ID < versions[j].ID
		}
		return versions[i].Name < versions[j].Name
	})
	cards := make([]*LlmVersion, 0, len(versions))
	for _, version := range versions {
		item := version
		cards = append(cards, &LlmVersion{ID: item.ID, Name: item.Name, Note: item.Note})
	}
	active := activeID
	return &LlmSlotCard{
		Slot:            string(slot),
		Wired:           true,
		Role:            "Kapasite işçisi — aynı iş",
		Occupancy:       occupancyFromBusy(busy),
		ActiveVersionID: &active,
		Versions:        cards,
	}, nil
}

func occupancyFromBusy(busy bool) LlmOccupancy {
	if busy {
		return LlmOccupancyBusy
	}
	return LlmOccupancyIdle
}
