package graph

import (
	"context"
	"sort"
	"time"

	"github.com/cherry/api/internal/colabbridge"
	"github.com/cherry/api/internal/llm"
	"github.com/cherry/api/internal/store"
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
		cards = append(cards, &LlmVersion{ID: item.ID, Name: item.Name, Note: item.Note, CheckpointRef: item.CheckpointRef})
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

func (r *Resolver) buildTrainingPack(ctx context.Context, userID string) (llm.Pack, error) {
	projects, err := r.Factory.List(ctx, userID)
	if err != nil {
		return llm.Pack{}, err
	}
	audits, err := r.LLM.Store.ListAudit(ctx, userID)
	if err != nil {
		return llm.Pack{}, err
	}
	logs := map[string][]store.JobLog{}
	traces := make([]llm.MaestroTrace, 0)
	for _, project := range projects {
		items, logErr := r.Factory.Logs(ctx, userID, project.ID)
		if logErr == nil {
			logs[project.ID] = items
		}
		studio, studioErr := r.Factory.Maestro(ctx, userID, project.ID)
		if studioErr != nil {
			continue
		}
		for _, flow := range studio.Flows {
			traces = append(traces, llm.MaestroTrace{
				ProjectID: project.ID,
				Name:      flow.Name,
				YAML:      flow.YAML,
				Result:    string(flow.Result),
				Note:      flow.Note,
			})
		}
	}
	return llm.BuildPack(ctx, llm.PackInput{
		Projects: projects,
		Audits:   audits,
		Logs:     logs,
		Maestro:  traces,
	}), nil
}

func (r *Resolver) PackBodies(ctx context.Context, userID string) (string, string, error) {
	pack, err := r.buildTrainingPack(ctx, userID)
	if err != nil {
		return "", "", err
	}
	body, err := pack.JSON()
	if err != nil {
		return "", "", err
	}
	jsonl, err := pack.JSONL()
	if err != nil {
		return "", "", err
	}
	return body, jsonl, nil
}

func (r *Resolver) colabBridgePayload() *ColabBridge {
	if r.Bridge == nil {
		missing := "missing"
		note := "Colab köprüsü bağlı değil."
		return &ColabBridge{
			Status:      ColabBridgeStatusFailed,
			Cloudflared: missing,
			TokenHint:   "",
			Note:        note,
		}
	}
	snap := r.Bridge.Snapshot()
	status := ColabBridgeStatusIdle
	switch snap.Status {
	case colabbridge.StatusIdle:
		status = ColabBridgeStatusIdle
	case colabbridge.StatusStarting:
		status = ColabBridgeStatusStarting
	case colabbridge.StatusRunning:
		status = ColabBridgeStatusRunning
	case colabbridge.StatusStopping:
		status = ColabBridgeStatusStopping
	case colabbridge.StatusFailed:
		status = ColabBridgeStatusFailed
	default:
		status = ColabBridgeStatusFailed
	}
	out := &ColabBridge{
		Status:      status,
		Cloudflared: snap.Cloudflared,
		TokenHint:   snap.TokenHint,
		Note:        snap.Note,
	}
	if snap.PublicURL != "" {
		out.PublicURL = &snap.PublicURL
	}
	if snap.LocalURL != "" {
		out.LocalURL = &snap.LocalURL
	}
	if snap.Token != "" {
		out.Token = &snap.Token
	}
	if snap.StartedAt != "" {
		out.StartedAt = &snap.StartedAt
	}
	return out
}

func (r *Resolver) colabInferencePayload() *ColabInference {
	url, status := r.LLM.ColabInferenceState()
	gqlStatus := ColabInferenceStatusOff
	switch status {
	case llm.ColabInferenceOff:
		gqlStatus = ColabInferenceStatusOff
	case llm.ColabInferenceConnected:
		gqlStatus = ColabInferenceStatusConnected
	case llm.ColabInferenceDisconnected:
		gqlStatus = ColabInferenceStatusDisconnected
	case llm.ColabInferenceChecking:
		gqlStatus = ColabInferenceStatusChecking
	default:
		gqlStatus = ColabInferenceStatusOff
	}
	note := "Colab inferans kapalı."
	switch gqlStatus {
	case ColabInferenceStatusOff:
		note = "Colab inferans kapalı. Colab notebook'unda tünel aç ve URL'yi yapıştır."
	case ColabInferenceStatusConnected:
		note = "Colab inferans bağlı. İstekler Colab tüneline gidiyor. Colab oturumu kapanınca bağlantı kopar."
	case ColabInferenceStatusDisconnected:
		note = "Colab inferans bağlantı kesildi. Colab oturumu kapanmış olabilir. Varsayılan endpoint kullanılıyor."
	case ColabInferenceStatusChecking:
		note = "Colab inferans kontrol ediliyor…"
	}
	return &ColabInference{
		URL:    url,
		Status: gqlStatus,
		Note:   note,
	}
}

func trainingPackPayload(pack llm.Pack) (*TrainingPack, error) {
	body, err := pack.JSON()
	if err != nil {
		return nil, err
	}
	jsonl, err := pack.JSONL()
	if err != nil {
		return nil, err
	}
	return &TrainingPack{
		Schema:       pack.Schema,
		Filename:     "cherry-training-pack.json",
		JSON:         body,
		Jsonl:        jsonl,
		LiveExamples: pack.Stats.LiveExamples,
		SeedExamples: pack.Stats.SeedExamples,
		Note:         pack.Note,
	}, nil
}
