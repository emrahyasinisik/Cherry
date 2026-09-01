# LLMOps — iki model, switch, versiyon, Colab

**Kural:** [.cursor/rules/11-llmops.mdc](../.cursor/rules/11-llmops.mdc)

**TR:** İki çalışma yuvası. Yönetici değiştirince **sonraki cevaplar** değişir.

**EN:** Two runtime slots. After the admin switches, **later answers** change.

## Yönlendirici / Router

```mermaid
flowchart TB
  Job[job] --> GDPR[KVKK_layer]
  GDPR --> Router[router]
  Router --> SlotA[slot_A_codegen]
  Router --> SlotB[slot_B_test]
  SlotA --> VA[active_version_A]
  SlotB --> VB[active_version_B]
  Admin[admin_switch] --> Router
  Colab[Colab_checkpoint] --> Versions[llmVersions]
  Versions --> VA
  Versions --> VB
```

## Switch

```mermaid
sequenceDiagram
  participant Admin
  participant API
  participant Jobs
  Admin->>API: switch_slots_or_version
  API->>API: point_active
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
