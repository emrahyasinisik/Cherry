# Colab dosyaları / Colab files

**TR:** İki notebook, aynı QLoRA tarifi, iki ayrı 16GB T4 oturumu. Colab üretim inferansı değil.

**EN:** Two notebooks, same QLoRA recipe, two 16GB T4 sessions. Colab is not production inference.

| Dosya / File | Ne / What |
| --- | --- |
| [icerde_worker_a.ipynb](icerde_worker_a.ipynb) | İşçi A |
| [icerde_worker_b.ipynb](icerde_worker_b.ipynb) | İşçi B |
| [examples/icerde_training_pack.json](examples/icerde_training_pack.json) | Seed paket |
| [examples/icerde_sft.jsonl](examples/icerde_sft.jsonl) | Seed JSONL |
| [training_pack.schema.json](training_pack.schema.json) | Şema |

Nasıl: [docs/colab.md](../docs/colab.md)

```bash
python3 colab/generate_notebooks.py
```
