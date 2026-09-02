# LLMOps — iki işçi, kuyruk, versiyon, Colab

**Kural:** [.cursor/rules/11-llmops.mdc](../.cursor/rules/11-llmops.mdc)

**TR:** A ve B **aynı işi** yapar. İki yuva, 10 kişi aynı anda üretirken kuyruğu paylaşmak içindir — kod vs test ayrımı değildir.

**EN:** A and B do **the same work**. Two slots exist so ~10 concurrent users can share load — not to split codegen vs test.

## Neden iki / Why two

Tek model bir anda bir tamamlamayı doldurur. Stüdyoda birden fazla kişi (veya bir kişinin birden fazla işi) beklemesin diye ikinci işçi var.

One model fills one completion at a time. The second worker exists so several people in the studio are not stuck behind a single in-flight call.

| Bu / This | Değil / Not |
| --- | --- |
| Kapasite: boş işçiyi al, doluysa kuyruğa yaz | A = kod, B = test |
| Aynı GDPR sarmalı, aynı OpenCode / Maestro araçları | Rol switch’i (“nöbet: test”) |
| Versiyon pointer’ı işçide sonraki cevapları değiştirir | İş türüne göre model kilidi |

## Yönlendirici / Router

```mermaid
flowchart TB
  Job[job_any_kind] --> GDPR[KVKK_layer]
  GDPR --> Router[router_pick_idle]
  Router -->|idle| SlotA[worker_A]
  Router -->|idle| SlotB[worker_B]
  Router -->|both_busy| Queue[queue]
  Queue --> Router
  SlotA --> VA[active_version_A]
  SlotB --> VB[active_version_B]
  Colab[Colab_checkpoint] --> Versions[llmVersions]
  Versions --> VA
  Versions --> VB
```

Dilim 4: yalnızca **işçi A**. Dilim 7: A ve B kuyruğu paylaşır. Versiyon değiştirmek o işçideki sonraki tamamlamaları değiştirir (in-flight eski pointer’da biter).

Slice 4: **worker A** only. Slice 7 wires B and the queue.

## Versiyon pointer / Version pointer

```mermaid
sequenceDiagram
  participant Admin
  participant API
  participant Jobs
  Admin->>API: set_active_version
  API->>API: point_active_on_worker
  API->>Jobs: new_jobs_use_new_pointer
  Note over Jobs: in_flight_keep_old
```

Default: in-flight jobs finish on the old assignment; queued/new jobs use the new one.

## MCP read-file

Admin sets allowed root (the generated project path). Models read files only under that root.

```mermaid
flowchart LR
  LLM --> MCP[mcp_read_file]
  MCP --> Root[allowed_root]
  Root --> Src[frontend_backend_maestro]
```

## Colab

Export training pack → Colab → checkpoint back → new immutable `llmVersion` → admin marks active. Colab is not the production server.

**Karar / Decision:** İki notebook, iki işçi. Aynı anda çalışabilir — **iki ayrı Colab oturumu**, her biri **16GB GPU** (T4 sınıfı). Tek 16GB kartta iki notebook yok.

Two notebooks, two workers. Parallel means **two Colab runtimes**, each with a **16GB GPU** budget (T4-class). One 16GB card does not host both.

```mermaid
flowchart LR
  Pack[training_pack] --> NbA[notebook_A_16GB]
  Pack --> NbB[notebook_B_16GB]
  NbA --> CkA[checkpoint_A]
  NbB --> CkB[checkpoint_B]
  CkA --> VerA[llmVersion_worker_A]
  CkB --> VerB[llmVersion_worker_B]
```

| Sabit / Invariant | Neden / Why |
| --- | --- |
| Aynı tarif (LoRA/QLoRA), aynı veri paketi | A ve B yük paylaşır; farklı iş uzmanı değiller |
| 16GB / oturum | Colab T4 bütçesi; tam ağırlık büyük model yok |
| İki oturum = iki GPU hakkı | Paralel fine-tune; tek kartı bölme |
| Colab üretim inferansı değil | Stüdyo işçileri İçerde’de çalışır |

Dilim 8: dosyalar `colab/`, stüdyodan paket indirme, checkpoint → `registerLlmVersion`. Colab üretim inferansı değil.

Slice 8: files in `colab/`, studio pack download, checkpoint → `registerLlmVersion`. Colab is not production inference.

Adım adım: [colab.md](colab.md)
