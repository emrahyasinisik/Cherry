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

Dilim 4: yalnızca **işçi A**. B kartı “bağlı değil — yoğunluk işçisi sonra”. Versiyon değiştirmek A’daki sonraki tamamlamaları değiştirir (in-flight eski pointer’da biter).

Slice 4: **worker A** only. B card is “not wired — capacity worker later”.

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
