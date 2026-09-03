#!/usr/bin/env python3
"""Generate Cherry Colab notebooks A and B from one recipe."""

from __future__ import annotations

import json
from pathlib import Path

ROOT = Path(__file__).resolve().parent


def cells(worker: str) -> list[dict]:
    other = "B" if worker == "A" else "A"
    md = lambda text: {
        "cell_type": "markdown",
        "metadata": {},
        "source": [line + "\n" for line in text.strip("\n").split("\n")],
    }
    py = lambda text: {
        "cell_type": "code",
        "metadata": {},
        "execution_count": None,
        "outputs": [],
        "source": [line + "\n" for line in text.strip("\n").split("\n")],
    }
    return [
        md(
            f"""# Cherry — Colab fine-tune (işçi {worker})

**TR:** Bu notebook **LLM {worker}** için. Aynı tarif **LLM {other}** notebook’unda da var. İki ayrı Colab oturumu, her biri **16GB GPU (T4)**. Tek kartta iki notebook yok. Colab üretim inferansı değildir — adapter’ı indir, stüdyoda sürüm olarak kaydet.

**EN:** Same QLoRA recipe as worker {other}. Two Colab sessions, **16GB GPU each**. Colab is not production inference.

Dosyalar / Files:
1. `cherry_training_pack.json` (stüdyodan veya `colab/examples/`)
2. Bu `.ipynb` — Runtime → GPU (T4)"""
        ),
        md("## 0. GPU kontrol / GPU check"),
        py(
            """import torch

print("cuda", torch.cuda.is_available())
if torch.cuda.is_available():
    print(torch.cuda.get_device_name(0))
    props = torch.cuda.get_device_properties(0)
    print("vram_gb", round(props.total_memory / 1024**3, 2))
else:
    raise SystemExit("GPU yok. Runtime → Change runtime type → T4 GPU.")"""
        ),
        md("## 1. Paketler / Packages"),
        py(
            """%pip -q install -U transformers==4.51.3 datasets==3.6.0 peft==0.15.2 accelerate==1.6.0 bitsandbytes==0.45.5"""
        ),
        md("## 2. İşçi ve tarif / Worker + recipe"),
        py(
            f"""WORKER = "{worker}"  # do not change; this file is worker {worker}
BASE_MODEL = "Qwen/Qwen2.5-1.5B-Instruct"
MAX_SEQ = 1024
LORA_R = 16
LORA_ALPHA = 32
BATCH = 1
GRAD_ACCUM = 8
EPOCHS = 2
LR = 2e-4
print("worker", WORKER, "base", BASE_MODEL, "gpu_budget_gb", 16)"""
        ),
        md(
            """## 3. Eğitim paketi / Training pack

Soldan **cherry_training_pack.json** yükle (stüdyo LLM sayfası veya `colab/examples/`). Yoksa hücredeki seed çalışır."""
        ),
        py(
            r'''from pathlib import Path
import json

PACK_PATH = Path("/content/cherry_training_pack.json")
MINI = {
  "schema": "cherry.training_pack.v1",
  "recipe": {"baseModel": BASE_MODEL, "method": "qlora", "gpuBudgetGb": 16},
  "examples": [
    {
      "instruction": "Cherry stüdyosu için mobil uygulama planı yaz. preview/ HTML site yazma. PII uydurma.",
      "input": "Proje: Kahve sipariş\nYığın: EXPO\nBrif: Mahalle kahvecisi. Giriş + ana ekran. Yerel backend.",
      "output": "Plan (Expo SDK 57):\n- frontend/ domain-data-presentation\n- backend/ yerel\n- maestro/ login.yaml + home.yaml"
    },
    {
      "instruction": "Maestro YAML yaz. Cihaz yoksa SKIPPED; PASSED uydurma.",
      "input": "Akış: login.yaml\nSonuç: SKIPPED",
      "output": "appId: com.cherry.demo\n---\n- launchApp\n- assertVisible: \"Giriş\"\n"
    },
    {
      "instruction": "Bu yola uygun kaynak dosyayı yaz. Seçilen dil. HTML site değil.",
      "input": "Yığın: EXPO\nYol: frontend/src/domain/entities/item.ts",
      "output": "export type Item = {\n  id: string;\n  title: string;\n};\n"
    },
  ],
}

if PACK_PATH.exists():
    pack = json.loads(PACK_PATH.read_text())
    print("loaded", PACK_PATH, "examples", len(pack.get("examples", [])))
else:
    pack = MINI
    print("no upload; using mini seed", len(pack["examples"]))

rows = []
for ex in pack.get("examples", []):
    instruction = (ex.get("instruction") or "").strip()
    output = (ex.get("output") or "").strip()
    if not instruction or not output:
        continue
    rows.append({
        "instruction": instruction,
        "input": (ex.get("input") or "").strip(),
        "output": output,
    })
print("sft_rows", len(rows))
if len(rows) < 2:
    raise SystemExit("Paket boş. Stüdyodan JSON indir veya examples/cherry_training_pack.json yükle.")'''
        ),
        md(
            """## 3b. Tünel ile paket çek / Fetch pack via tunnel (isteğe bağlı)

Stüdyoda **Tüneli aç** dediysen URL ve token'ı aşağıya yapıştır. Yoksa üstteki dosya yüklemeyi kullan."""
        ),
        py(
            r'''# Optional: fetch training pack from Cherry tunnel instead of file upload.
# Paste the tunnel URL and token from the LLM admin page.
TUNNEL_URL = ""   # e.g. "https://xxx.trycloudflare.com"
TUNNEL_TOKEN = "" # bearer token from studio

if TUNNEL_URL and TUNNEL_TOKEN:
    import urllib.request, json as _json
    req = urllib.request.Request(
        TUNNEL_URL.rstrip("/") + "/pack",
        headers={"Authorization": f"Bearer {TUNNEL_TOKEN}"},
    )
    with urllib.request.urlopen(req, timeout=30) as resp:
        pack = _json.loads(resp.read())
    PACK_PATH.write_text(_json.dumps(pack, ensure_ascii=False))
    print("tunnel: fetched pack", len(pack.get("examples", [])), "examples")
else:
    print("tunnel: skipped (no URL). Using file upload or seed.")'''
        ),
        md("## 4. Dataset"),
        py(
            r'''from datasets import Dataset

def format_row(ex):
    user = ex["instruction"]
    if ex["input"]:
        user = user + "\n\n" + ex["input"]
    return (
        f"<|im_start|>system\nCherry işçi {WORKER}. Mobil frontend/backend ve Maestro YAML yaz. PII yok. HTML site yazma.<|im_end|>\n"
        f"<|im_start|>user\n{user}<|im_end|>\n"
        f"<|im_start|>assistant\n{ex['output']}<|im_end|>"
    )

ds = Dataset.from_list(rows).map(lambda ex: {"text": format_row(ex)})
print(ds[0]["text"][:400])'''
        ),
        md("## 5. 4-bit QLoRA (16GB T4)"),
        py(
            r'''from transformers import AutoModelForCausalLM, AutoTokenizer, BitsAndBytesConfig
from peft import LoraConfig, get_peft_model, prepare_model_for_kbit_training
import torch

bnb = BitsAndBytesConfig(
    load_in_4bit=True,
    bnb_4bit_quant_type="nf4",
    bnb_4bit_use_double_quant=True,
    bnb_4bit_compute_dtype=torch.bfloat16 if torch.cuda.is_bf16_supported() else torch.float16,
)
tok = AutoTokenizer.from_pretrained(BASE_MODEL, trust_remote_code=True)
if tok.pad_token is None:
    tok.pad_token = tok.eos_token
model = AutoModelForCausalLM.from_pretrained(
    BASE_MODEL,
    quantization_config=bnb,
    device_map="auto",
    trust_remote_code=True,
)
model = prepare_model_for_kbit_training(model)
model = get_peft_model(model, LoraConfig(
    r=LORA_R,
    lora_alpha=LORA_ALPHA,
    lora_dropout=0.05,
    bias="none",
    task_type="CAUSAL_LM",
    target_modules=["q_proj", "k_proj", "v_proj", "o_proj", "gate_proj", "up_proj", "down_proj"],
))
model.print_trainable_parameters()'''
        ),
        md("## 6. Eğitim / Train"),
        py(
            r'''from transformers import DataCollatorForLanguageModeling, Trainer, TrainingArguments

def tokenize(batch):
    out = tok(batch["text"], truncation=True, max_length=MAX_SEQ, padding=False)
    out["labels"] = [ids[:] for ids in out["input_ids"]]
    return out

tokenized = ds.map(tokenize, batched=True, remove_columns=ds.column_names)
args = TrainingArguments(
    output_dir=f"/content/out_{WORKER}",
    per_device_train_batch_size=BATCH,
    gradient_accumulation_steps=GRAD_ACCUM,
    num_train_epochs=EPOCHS,
    learning_rate=LR,
    logging_steps=1,
    save_strategy="epoch",
    fp16=not torch.cuda.is_bf16_supported(),
    bf16=torch.cuda.is_bf16_supported(),
    report_to=[],
    remove_unused_columns=False,
)
collator = DataCollatorForLanguageModeling(tok, mlm=False)
trainer = Trainer(model=model, args=args, train_dataset=tokenized, data_collator=collator)
trainer.train()'''
        ),
        md("## 7. Adapter indir / Download adapter"),
        py(
            f'''from google.colab import files
import shutil

adapter_dir = f"/content/cherry_adapter_worker_{worker}"
zip_path = f"/content/cherry_adapter_worker_{worker}"
model.save_pretrained(adapter_dir)
tok.save_pretrained(adapter_dir)
shutil.make_archive(zip_path, "zip", adapter_dir)
print("zip", zip_path + ".zip")
print("Stüdyo LLM sayfasında kaydet / Register in Cherry:")
print(f'  slot={worker}  name=v-colab  checkpointRef=cherry_adapter_worker_{worker}.zip')
files.download(zip_path + ".zip")'''
        ),
        md(
            """## 7b. Adapter'ı tünele yükle / POST adapter to tunnel (isteğe bağlı)

Tünel açıksa adapter zip'i stüdyoya geri gönderir. Yoksa zip'i manuel indir."""
        ),
        py(
            f'''# Optional: POST adapter zip back to the Cherry tunnel.
import os

zip_file = zip_path + ".zip"
if TUNNEL_URL and TUNNEL_TOKEN and os.path.exists(zip_file):
    import urllib.request
    with open(zip_file, "rb") as f:
        body = f.read()
    req = urllib.request.Request(
        TUNNEL_URL.rstrip("/") + "/checkpoint",
        data=body,
        headers={{
            "Authorization": f"Bearer {{TUNNEL_TOKEN}}",
            "Content-Type": "application/zip",
            "X-Worker": "{worker}",
        }},
        method="POST",
    )
    with urllib.request.urlopen(req, timeout=120) as resp:
        print("tunnel: uploaded", len(body), "bytes", resp.status)
else:
    print("tunnel: skipped checkpoint upload (no URL or no zip).")
    print("Zip dosyasını manuel indir ve stüdyoda kaydet.")'''
        ),
        md(
            f"""## Sonra / Next

1. Zip’i makineye indir.
2. Cherry → LLM yönetici → **Colab sürümü kaydet** (işçi {worker}).
3. Pointer’ı o sürüme al. In-flight işler eski pointer’da biter.
4. Colab’i kapat. Üretim çağrıları stüdyo işçilerinde kalır."""
        ),
    ]


def notebook(worker: str) -> dict:
    return {
        "nbformat": 4,
        "nbformat_minor": 5,
        "metadata": {
            "kernelspec": {"display_name": "Python 3", "language": "python", "name": "python3"},
            "language_info": {"name": "python"},
            "accelerator": "GPU",
            "colab": {"provenance": [], "gpuType": "T4"},
        },
        "cells": cells(worker),
    }


def main() -> None:
    for worker in ("A", "B"):
        path = ROOT / f"cherry_worker_{worker.lower()}.ipynb"
        path.write_text(json.dumps(notebook(worker), indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
        print("wrote", path)


if __name__ == "__main__":
    main()
