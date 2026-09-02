package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/icerde/api/internal/gdpr"
	"github.com/icerde/api/internal/store"
)

const PackSchema = "icerde.training_pack.v1"

const (
	packBaseModel = "Qwen/Qwen2.5-1.5B-Instruct"
	packMethod    = "qlora"
	packGPUGB     = 16
	packMaxSeq    = 1024
	packLoraR     = 16
	maxFileBytes  = 6000
	maxFiles      = 10
	maxLogs       = 12
)

type Pack struct {
	Schema     string        `json:"schema"`
	ExportedAt string        `json:"exportedAt"`
	Recipe     PackRecipe    `json:"recipe"`
	Examples   []PackExample `json:"examples"`
	Stats      PackStats     `json:"stats"`
	Note       string        `json:"note"`
}

type PackRecipe struct {
	BaseModel   string `json:"baseModel"`
	Method      string `json:"method"`
	GpuBudgetGb int    `json:"gpuBudgetGb"`
	MaxSeqLen   int    `json:"maxSeqLen"`
	LoraR       int    `json:"loraR"`
}

type PackExample struct {
	ID          string            `json:"id"`
	Kind        string            `json:"kind"`
	Source      string            `json:"source"`
	Instruction string            `json:"instruction"`
	Input       string            `json:"input"`
	Output      string            `json:"output"`
	Meta        map[string]string `json:"meta,omitempty"`
}

type PackStats struct {
	LiveExamples int `json:"liveExamples"`
	SeedExamples int `json:"seedExamples"`
}

type MaestroTrace struct {
	ProjectID string
	Name      string
	YAML      string
	Result    string
	Note      string
}

type PackInput struct {
	Projects []store.Project
	Audits   []store.AuditEvent
	Logs     map[string][]store.JobLog
	Maestro  []MaestroTrace
}

func DefaultRecipe() PackRecipe {
	return PackRecipe{
		BaseModel:   packBaseModel,
		Method:      packMethod,
		GpuBudgetGb: packGPUGB,
		MaxSeqLen:   packMaxSeq,
		LoraR:       packLoraR,
	}
}

func BuildPack(_ context.Context, in PackInput) Pack {
	pack := Pack{
		Schema:     PackSchema,
		ExportedAt: time.Now().UTC().Format(time.RFC3339),
		Recipe:     DefaultRecipe(),
		Examples:   make([]PackExample, 0),
		Note:       "KVKK/GDPR redakte. Ham PII, token, .env yok. Colab üretim inferansı değildir.",
	}
	for _, project := range in.Projects {
		pack.Examples = append(pack.Examples, examplesFromProject(project, in.Logs[project.ID])...)
	}
	for _, trace := range in.Maestro {
		pack.Examples = append(pack.Examples, exampleFromMaestro(trace))
	}
	for i, event := range in.Audits {
		if i >= 40 {
			break
		}
		pack.Examples = append(pack.Examples, exampleFromAudit(event))
	}
	live := 0
	kept := make([]PackExample, 0, len(pack.Examples))
	for _, example := range pack.Examples {
		if strings.TrimSpace(example.Instruction) == "" || strings.TrimSpace(example.Output) == "" {
			continue
		}
		example.Instruction, _ = gdpr.Redact(example.Instruction)
		example.Input, _ = gdpr.Redact(example.Input)
		example.Output, _ = gdpr.Redact(example.Output)
		kept = append(kept, example)
		if example.Source == "live" {
			live++
		}
	}
	pack.Examples = kept
	if live < 4 {
		seed := SeedExamples()
		pack.Examples = append(pack.Examples, seed...)
		pack.Stats.SeedExamples = len(seed)
		if live == 0 {
			pack.Note = "Canlı iz yok veya ince. Seed örnekleri eklendi — Colab yine çalışır. Gerçek brif üretince paketi yeniden indir."
		} else {
			pack.Note = "Canlı iz ince. Seed örnekleri doldurma olarak eklendi. Yeni proje üretince paketi yeniden indir."
		}
	}
	pack.Stats.LiveExamples = live
	return pack
}

func (p Pack) JSON() (string, error) {
	raw, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func (p Pack) JSONL() (string, error) {
	var buf bytes.Buffer
	for _, example := range p.Examples {
		row := map[string]string{
			"instruction": example.Instruction,
			"input":       example.Input,
			"output":      example.Output,
		}
		raw, err := json.Marshal(row)
		if err != nil {
			return "", err
		}
		buf.Write(raw)
		buf.WriteByte('\n')
	}
	return buf.String(), nil
}

func examplesFromProject(project store.Project, logs []store.JobLog) []PackExample {
	out := make([]PackExample, 0)
	brief, _ := gdpr.Redact(strings.TrimSpace(project.Brief))
	plan := readCapped(project.RootPath, "llm/plan.md")
	if plan == "" {
		plan = firstAgentLog(logs)
	}
	if brief != "" && plan != "" {
		out = append(out, PackExample{
			ID:          "brief-" + project.ID,
			Kind:        "brief",
			Source:      "live",
			Instruction: "İçerde stüdyosu için mobil uygulama planı yaz. Seçilen yığın ve Clean Architecture. preview/ HTML site yazma. PII uydurma.",
			Input:       "Proje: " + project.Name + "\nYığın: " + string(project.Stack) + "\nBrif:\n" + clip(brief, 1200),
			Output:      clip(plan, 1800),
			Meta:        map[string]string{"projectId": project.ID, "stack": string(project.Stack), "kind": "brief"},
		})
	}
	readme := readCapped(project.RootPath, "README.md")
	if brief != "" && readme != "" {
		out = append(out, PackExample{
			ID:          "readme-" + project.ID,
			Kind:        "source",
			Source:      "live",
			Instruction: "Müşteri teslim README yaz. Barındırma yok; klasör/zip/git. Yığın dilini koru.",
			Input:       "Proje: " + project.Name + "\nYığın: " + string(project.Stack) + "\nBrif:\n" + clip(brief, 800),
			Output:      clip(readme, 1600),
			Meta:        map[string]string{"projectId": project.ID, "path": "README.md"},
		})
	}
	arch := firstExisting(project.RootPath, []string{
		"frontend/ARCHITECTURE.md",
		"frontend/Presentation/ARCHITECTURE.md",
	})
	if arch.body != "" {
		out = append(out, PackExample{
			ID:          "arch-" + project.ID,
			Kind:        "source",
			Source:      "live",
			Instruction: "Bu yığın için Clean Architecture özetini yaz. Katmanları karıştırma.",
			Input:       "Yığın: " + string(project.Stack) + "\nDosya: " + arch.path,
			Output:      clip(arch.body, 1600),
			Meta:        map[string]string{"projectId": project.ID, "path": arch.path},
		})
	}
	pairs := userAgentPairs(logs)
	for i, pair := range pairs {
		if i >= 6 {
			break
		}
		out = append(out, PackExample{
			ID:          "chat-" + project.ID + "-" + strconv.Itoa(i),
			Kind:        "completion",
			Source:      "live",
			Instruction: "İçerde ajanı: brife ve sohbet mesajına göre dosya planı veya yama öner. PII yok.",
			Input:       clip(pair.user, 1000),
			Output:      clip(pair.agent, 1400),
			Meta:        map[string]string{"projectId": project.ID, "kind": "chat"},
		})
	}
	_ = filepath.WalkDir(project.RootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "node_modules", ".opencode", "preview":
				return fs.SkipDir
			default:
				return nil
			}
		}
		rel, relErr := filepath.Rel(project.RootPath, path)
		if relErr != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if skipPackRel(rel) || !packSourceRel(rel) {
			return nil
		}
		if len(out) > maxFiles+8 {
			return fs.SkipAll
		}
		body := readCapped(project.RootPath, rel)
		if body == "" {
			return nil
		}
		out = append(out, PackExample{
			ID:          "file-" + project.ID + "-" + sanitizeID(rel),
			Kind:        "source",
			Source:      "live",
			Instruction: "Bu yola uygun kaynak dosyayı yaz. Seçilen dil. HTML site değil.",
			Input:       "Yığın: " + string(project.Stack) + "\nYol: " + rel + "\nBrif: " + clip(brief, 400),
			Output:      body,
			Meta:        map[string]string{"projectId": project.ID, "path": rel},
		})
		return nil
	})
	return out
}

type diskFile struct {
	path string
	body string
}

func firstExisting(root string, rels []string) diskFile {
	for _, rel := range rels {
		body := readCapped(root, rel)
		if body != "" {
			return diskFile{path: rel, body: body}
		}
	}
	return diskFile{}
}

func exampleFromMaestro(trace MaestroTrace) PackExample {
	result := strings.TrimSpace(trace.Result)
	if result == "" {
		result = "SKIPPED"
	}
	note := strings.TrimSpace(trace.Note)
	if note == "" {
		note = "Cihaz yoksa SKIPPED. PASSED uydurma."
	}
	output := strings.TrimSpace(trace.YAML)
	if output == "" {
		output = "# Maestro flow\nappId: demo\n---\n- launchApp"
	}
	return PackExample{
		ID:          "maestro-" + trace.ProjectID + "-" + sanitizeID(trace.Name),
		Kind:        "maestro",
		Source:      "live",
		Instruction: "Maestro YAML yaz. Cihaz yoksa sonuç SKIPPED olur; PASSED uydurma.",
		Input:       "Akış: " + trace.Name + "\nSonuç: " + result + "\nNot: " + note,
		Output:      clip(output, 1600),
		Meta:        map[string]string{"projectId": trace.ProjectID, "flow": trace.Name, "result": result},
	}
}

func exampleFromAudit(event store.AuditEvent) PackExample {
	return PackExample{
		ID:          "audit-" + event.ID,
		Kind:        "completion",
		Source:      "live",
		Instruction: "GDPR sarmalı tamamla. PII yok. İşçi " + string(event.Slot) + ", sürüm " + event.VersionName + ".",
		Input:       event.PromptPreview,
		Output:      event.OutputPreview,
		Meta: map[string]string{
			"slot":    string(event.Slot),
			"version": event.VersionName,
			"purpose": event.Purpose,
		},
	}
}

type chatPair struct {
	user  string
	agent string
}

func userAgentPairs(logs []store.JobLog) []chatPair {
	out := make([]chatPair, 0)
	var pending string
	for _, item := range logs {
		switch item.Role {
		case store.RoleUser:
			pending = item.Message
		case store.RoleAgent:
			if pending != "" && item.Message != "" {
				out = append(out, chatPair{user: pending, agent: item.Message})
				pending = ""
			}
		}
	}
	if len(out) > maxLogs {
		return out[len(out)-maxLogs:]
	}
	return out
}

func firstAgentLog(logs []store.JobLog) string {
	for _, item := range logs {
		if item.Role == store.RoleAgent && strings.TrimSpace(item.Message) != "" {
			return item.Message
		}
	}
	return ""
}

func readCapped(root, rel string) string {
	rel = strings.TrimSpace(rel)
	if root == "" || rel == "" {
		return ""
	}
	data, err := gdpr.ReadFile(root, rel)
	if err != nil {
		return ""
	}
	safe, _ := gdpr.Redact(string(data))
	return clip(safe, maxFileBytes)
}

func skipPackRel(rel string) bool {
	rel = filepath.ToSlash(rel)
	base := filepath.Base(rel)
	switch base {
	case "icerde.zip", "opencode.json", "AGENTS.md", ".env":
		return true
	}
	if strings.HasPrefix(base, ".env") {
		return true
	}
	first, _, _ := strings.Cut(rel, "/")
	switch first {
	case ".git", "node_modules", ".opencode", "preview":
		return true
	}
	return false
}

func packSourceRel(rel string) bool {
	rel = filepath.ToSlash(rel)
	if strings.HasPrefix(rel, "maestro/") && strings.HasSuffix(rel, ".yaml") {
		return true
	}
	if rel == "frontend/ARCHITECTURE.md" {
		return false
	}
	switch {
	case strings.HasPrefix(rel, "frontend/src/"),
		strings.HasPrefix(rel, "frontend/lib/"),
		strings.HasPrefix(rel, "frontend/Presentation/"),
		strings.HasPrefix(rel, "frontend/Domain/"),
		strings.HasPrefix(rel, "backend/"):
		ext := strings.ToLower(filepath.Ext(rel))
		switch ext {
		case ".ts", ".tsx", ".dart", ".swift", ".go", ".kt":
			return true
		}
	}
	return false
}

func sanitizeID(rel string) string {
	rel = strings.ReplaceAll(rel, "/", "-")
	rel = strings.ReplaceAll(rel, ".", "-")
	return rel
}
