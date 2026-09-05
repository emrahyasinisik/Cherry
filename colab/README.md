# Colab dosyaları / Colab files

**TR:** İki notebook, aynı QLoRA tarifi, iki ayrı 16GB T4 oturumu. Colab üretim inferansı değil.

**EN:** Two notebooks, same QLoRA recipe, two 16GB T4 sessions. Colab is not production inference.

| Dosya / File | Ne / What |
| --- | --- |
| [cherry_worker_a.ipynb](cherry_worker_a.ipynb) | İşçi A |
| [cherry_worker_b.ipynb](cherry_worker_b.ipynb) | İşçi B |
| [examples/cherry_training_pack.json](examples/cherry_training_pack.json) | Seed paket (~30+ satır; mini 3’lük yok) |
| [examples/cherry_sft.jsonl](examples/cherry_sft.jsonl) | Seed JSONL |
| [training_pack.schema.json](training_pack.schema.json) | Şema |

Nasıl: [docs/colab.md](../docs/colab.md)

Notebook `sft_rows < 24` ise train etmez. Seed kaynağı: `services/api/internal/llm/seed_pack.go`.

```bash
python3 colab/generate_notebooks.py
cd services/api && CHERRY_EXPORT_SEED=1 CHERRY_EXPORT_SEED_DIR=../../colab/examples \
  go test ./internal/llm -run TestExportSeedPackFiles
```
