package llm

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestExportSeedPackFiles(t *testing.T) {
	if os.Getenv("CHERRY_EXPORT_SEED") != "1" {
		t.Skip("set CHERRY_EXPORT_SEED=1 to rewrite colab/examples")
	}
	root := os.Getenv("CHERRY_EXPORT_SEED_DIR")
	if root == "" {
		t.Fatal("CHERRY_EXPORT_SEED_DIR required")
	}
	examples := SeedExamples()
	if len(examples) < 24 {
		t.Fatalf("seed too small: %d", len(examples))
	}
	pack := Pack{
		Schema:     PackSchema,
		ExportedAt: time.Date(2026, 9, 5, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
		Recipe:     DefaultRecipe(),
		Examples:   examples,
		Stats:      PackStats{LiveExamples: 0, SeedExamples: len(examples)},
		Note:       "Seed corpus. Canlı iz yokken Colab bununla eğitilir. Stüdyoda proje üretince LLM sayfasından canlı paketi indir.",
	}
	raw, err := json.MarshalIndent(pack, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "cherry_training_pack.json"), append(raw, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(filepath.Join(root, "cherry_sft.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, ex := range examples {
		if err := enc.Encode(map[string]string{
			"instruction": ex.Instruction,
			"input":       ex.Input,
			"output":      ex.Output,
		}); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("exported %d seed examples to %s", len(examples), root)
}

func TestSeedCorpusSize(t *testing.T) {
	if n := len(SeedExamples()); n < 24 {
		t.Fatalf("seed corpus too small for QLoRA: %d", n)
	}
}
