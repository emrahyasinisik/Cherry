#!/usr/bin/env python3
"""Generate Cherry Colab notebooks A and B from one recipe.

The seed training pack (colab/examples/cherry_training_pack.json) is base64-embedded
into each notebook so Colab runs without a manual JSON upload.
"""

from __future__ import annotations

import base64
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parent
SEED_PACK_PATH = ROOT / "examples" / "cherry_training_pack.json"


def embedded_pack_b64() -> str:
    pack = json.loads(SEED_PACK_PATH.read_text(encoding="utf-8"))
    n = len(pack.get("examples") or [])
    if n < 24:
        raise SystemExit(f"seed pack too thin for embed: {SEED_PACK_PATH} has {n} examples")
    return base64.b64encode(json.dumps(pack, ensure_ascii=False).encode("utf-8")).decode("ascii")


def md(text: str) -> dict:
    return {
        "cell_type": "markdown",
        "metadata": {},
        "source": [line + "\n" for line in text.strip("\n").split("\n")],
    }


def py(text: str) -> dict:
    return {
        "cell_type": "code",
        "metadata": {},
        "execution_count": None,
        "outputs": [],
        "source": [line + "\n" for line in text.strip("\n").split("\n")],
    }


def pack_cell() -> str:
    b64 = embedded_pack_b64()
    b64_literal = "\n".join(b64[i : i + 100] for i in range(0, len(b64), 100))
    return f'''from pathlib import Path
import base64
import json

MIN_SFT_ROWS = 24
PACK_PATH = Path("/content/cherry_training_pack.json")

# Full seed corpus baked into this notebook (from colab/examples/cherry_training_pack.json).
_EMBEDDED_B64 = """
{b64_literal}
""".replace("\\n", "").strip()

EMBEDDED_PACK = json.loads(base64.b64decode(_EMBEDDED_B64).decode("utf-8"))

def rows_from_pack(pack):
    out = []
    for ex in pack.get("examples", []):
        instruction = (ex.get("instruction") or "").strip()
        output = (ex.get("output") or "").strip()
        if not instruction or not output:
            continue
        out.append({{
            "instruction": instruction,
            "input": (ex.get("input") or "").strip(),
            "output": output,
        }})
    return out

CANDIDATES = [
    PACK_PATH,
    Path("/content/examples/cherry_training_pack.json"),
]

pack = None
loaded_from = None
for path in CANDIDATES:
    if path.exists():
        pack = json.loads(path.read_text())
        loaded_from = str(path)
        break

if pack is None:
    pack = EMBEDDED_PACK
    loaded_from = "embedded_seed_pack"
    print("no upload; using EMBEDDED seed pack", len(pack.get("examples", [])))
else:
    print("loaded upload/file", loaded_from, "examples", len(pack.get("examples", [])))

rows = rows_from_pack(pack)
stats = pack.get("stats") or {{}}
print("source", loaded_from)
print("examples_in_pack", len(pack.get("examples", [])))
print("liveExamples", stats.get("liveExamples"), "seedExamples", stats.get("seedExamples"))
print("sft_rows", len(rows))
if len(rows) < MIN_SFT_ROWS:
    raise SystemExit(
        f"sft_rows={{len(rows)}} < {{MIN_SFT_ROWS}}. Paket fine-tune için çok ince."
    )'''


def cells(worker: str) -> list[dict]:
    other = "B" if worker == "A" else "A"
    return [
        md(
            f"""# Cherry — Colab fine-tune (işçi {worker})

**TR:** Bu notebook **LLM {worker}** için. Aynı tarif **LLM {other}** notebook’unda. İki ayrı Colab oturumu, her biri **16GB GPU (T4)**. Colab üretim inferansı değil.

**EN:** Same QLoRA recipe as worker {other}. Two Colab sessions, **16GB GPU each**.

**Seed pack notebook içinde gömülü (~30+ satır).** JSON yüklemek **zorunlu değil**. Upload/tünel varsa o öncelikli.

Runtime → GPU (T4)"""
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

Gömülü seed corpus kullanılır (JSON yüklemeden). İstersen `/content/cherry_training_pack.json` yükle veya tünel aç — o zaman upload/tünel öncelikli.

Mini 3’lük seed **yok**. `sft_rows` en az 24 olmalı."""
        ),
        py(pack_cell()),
        md(
            """## 3b. Tünel ile paket çek / Fetch pack via tunnel (isteğe bağlı)

Stüdyoda tünel açtıysan URL + token yapıştır. Boş bırakırsan gömülü/yüklenen paket kullanılır."""
        ),
        py(
            '''# Optional: fetch training pack from Cherry tunnel.
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
    PACK_PATH.write_text(_json.dumps(pack, ensure_ascii=False), encoding="utf-8")
    rows = rows_from_pack(pack)
    print("tunnel: fetched pack", len(pack.get("examples", [])), "examples", "sft_rows", len(rows))
    if len(rows) < MIN_SFT_ROWS:
        raise SystemExit(f"tunnel pack too thin: sft_rows={len(rows)} < {MIN_SFT_ROWS}")
else:
    print("tunnel: skipped (no URL). Using pack source above; sft_rows=", len(rows))'''
        ),
        md("## 4. Dataset"),
        py(
            '''from datasets import Dataset

def format_row(ex):
    user = ex["instruction"]
    if ex["input"]:
        user = user + "\\n\\n" + ex["input"]
    return (
        f"<|im_start|>system\\nCherry işçi {WORKER}. Mobil frontend/backend ve Maestro YAML yaz. PII yok. HTML site yazma.<|im_end|>\\n"
        f"<|im_start|>user\\n{user}<|im_end|>\\n"
        f"<|im_start|>assistant\\n{ex['output']}<|im_end|>"
    )

ds = Dataset.from_list(rows).map(lambda ex: {"text": format_row(ex)})
print("dataset_rows", len(ds))
print(ds[0]["text"][:400])'''
        ),
        md("## 5. 4-bit QLoRA (16GB T4)"),
        py(
            '''from transformers import AutoModelForCausalLM, AutoTokenizer, BitsAndBytesConfig
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
            '''from transformers import DataCollatorForLanguageModeling, Trainer, TrainingArguments

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
            """## 8. İnferans sunucusu / Inference server (isteğe bağlı)

Fine-tune bitince modeli OpenAI uyumlu API olarak sun. **Geçici** — Colab kapanınca kopar."""
        ),
        py(
            '''%pip -q install fastapi uvicorn

from threading import Thread
import uvicorn
from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse, StreamingResponse
import torch, json, time as _time, uuid

app = FastAPI()
model.eval()

@app.get("/v1/models")
async def list_models():
    return {"object": "list", "data": [{"id": BASE_MODEL, "object": "model"}]}

@app.post("/v1/chat/completions")
async def chat_completions(request: Request):
    body = await request.json()
    messages = body.get("messages", [])
    stream = body.get("stream", False)
    max_tokens = min(body.get("max_tokens", 512), MAX_SEQ)

    prompt_parts = []
    for msg in messages:
        role = msg.get("role", "user")
        content = msg.get("content", "")
        prompt_parts.append(f"<|im_start|>{role}\\n{content}<|im_end|>")
    prompt_parts.append("<|im_start|>assistant\\n")
    prompt_text = "\\n".join(prompt_parts)

    inputs = tok(prompt_text, return_tensors="pt", truncation=True, max_length=MAX_SEQ).to(model.device)
    with torch.no_grad():
        outputs = model.generate(
            **inputs,
            max_new_tokens=max_tokens,
            do_sample=True,
            temperature=body.get("temperature", 0.7),
            top_p=body.get("top_p", 0.9),
            pad_token_id=tok.pad_token_id,
        )
    new_tokens = outputs[0][inputs["input_ids"].shape[1]:]
    text = tok.decode(new_tokens, skip_special_tokens=True)

    completion_id = "chatcmpl-" + uuid.uuid4().hex[:12]
    created = int(_time.time())

    if stream:
        async def generate():
            chunk = {
                "id": completion_id, "object": "chat.completion.chunk",
                "created": created, "model": BASE_MODEL,
                "choices": [{"index": 0, "delta": {"role": "assistant", "content": text}, "finish_reason": "stop"}],
            }
            yield f"data: {json.dumps(chunk)}\\n\\n"
            yield "data: [DONE]\\n\\n"
        return StreamingResponse(generate(), media_type="text/event-stream")

    return JSONResponse({
        "id": completion_id, "object": "chat.completion",
        "created": created, "model": BASE_MODEL,
        "choices": [{"index": 0, "message": {"role": "assistant", "content": text}, "finish_reason": "stop"}],
        "usage": {"prompt_tokens": inputs["input_ids"].shape[1], "completion_tokens": len(new_tokens), "total_tokens": inputs["input_ids"].shape[1] + len(new_tokens)},
    })

server_thread = Thread(target=lambda: uvicorn.run(app, host="0.0.0.0", port=8000, log_level="info"), daemon=True)
server_thread.start()
_time.sleep(3)
print("Inference server running on 0.0.0.0:8000")
print("POST /v1/chat/completions — OpenAI uyumlu")
print("GET  /v1/models — id:", BASE_MODEL)'''
        ),
        md(
            """## 9. Cloudflare tüneli / Cloudflare tunnel (inferans)

**Öncelik:** named tunnel (token). Yoksa quick tunnel (`trycloudflare`). Token’ı Colab secret yap — notebook’a yazma."""
        ),
        py(
            '''!wget -q https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb
!dpkg -i cloudflared-linux-amd64.deb

import os, subprocess, re, time as _time

CLOUDFLARE_TUNNEL_TOKEN = os.environ.get("CLOUDFLARE_TUNNEL_TOKEN", "").strip()
try:
    from google.colab import userdata
    if not CLOUDFLARE_TUNNEL_TOKEN:
        CLOUDFLARE_TUNNEL_TOKEN = (userdata.get("CLOUDFLARE_TUNNEL_TOKEN") or "").strip()
except Exception:
    pass

if CLOUDFLARE_TUNNEL_TOKEN:
    COLAB_PUBLIC_BASE = os.environ.get(
        "CHERRY_COLAB_PUBLIC_URL", "https://YOUR_SUBDOMAIN.example.com"
    ).rstrip("/")
    proc = subprocess.Popen(
        ["cloudflared", "tunnel", "--no-autoupdate", "run", "--token", CLOUDFLARE_TUNNEL_TOKEN],
        stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True,
    )
    deadline = _time.time() + 8
    while _time.time() < deadline:
        line = proc.stdout.readline() if proc.stdout else ""
        if not line:
            break
        print(line, end="")
    print("=" * 60)
    print(f"Cherry URL: {COLAB_PUBLIC_BASE}/v1")
    print("LLM yönetici → Colab inferans alanına yapıştır (setColabInferenceUrl).")
    print("=" * 60)
else:
    print("CLOUDFLARE_TUNNEL_TOKEN yok — quick tunnel (trycloudflare).")
    proc = subprocess.Popen(
        ["cloudflared", "tunnel", "--url", "http://localhost:8000", "--no-autoupdate"],
        stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True,
    )
    url = ""
    deadline = _time.time() + 30
    for line in proc.stdout:
        print(line, end="")
        m = re.search(r"https://[a-zA-Z0-9-]+\\.trycloudflare\\.com", line)
        if m:
            url = m.group()
            break
        if _time.time() > deadline:
            break
    if url:
        print(f"\\nCherry'de kullan: {url}/v1")
    else:
        print("Tunnel URL not found.")'''
        ),
        md("## 10. Canlı tut / Keep alive"),
        py(
            '''import time as _time
print("Oturum canlı tutuluyor… Stop cell to disconnect.")
while True:
    _time.sleep(60)
    print(".", end="", flush=True)'''
        ),
        md(
            f"""## Sonra / Next

1. Zip’i indir (`adapter_model.safetensors` içinde olmalı).
2. Cherry → LLM yönetici → **Colab sürümü kaydet** (işçi {worker}).
3. Pointer’ı o sürüme al.
4. Colab’ı kapat."""
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
