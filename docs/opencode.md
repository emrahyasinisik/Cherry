# OpenCode — kod yazma motoru

**Kural:** [.cursor/rules/08-mobile-factory.mdc](../.cursor/rules/08-mobile-factory.mdc) · [integrations.md](integrations.md)

**TR:** OpenCode’u yeniden yazmıyoruz. İçerde CLI’yi **çağırır**. LLM (işçi A) beynidir; OpenCode dosyayı yazar.

**EN:** We do not reimplement OpenCode. Icerde **invokes** the CLI. The LLM (worker A) is the brain; OpenCode writes files.

```mermaid
flowchart LR
  Brief[brif] --> GDPR[KVKK_redact]
  GDPR --> WorkerA[worker_A]
  WorkerA --> Plan[llm_plan.md]
  Plan --> OC[opencode_run]
  OC --> Tree[frontend_backend_maestro]
```

## Çağrı / Invoke

```bash
opencode run --dir <projectRoot> --auto
```

Prompt stdin’den, GDPR’den geçmiş brif + plan. Çalışma dizini yalnızca müşteri klasörü.

**Kişi OpenCode görmez.** Yalnızca İçerde sohbetine yazar. TUI açılmaz. Program `sendProjectMessage` → `opencode run --dir --auto` (sonraki mesajda `--continue`).

**The person never sees OpenCode.** They only write in Icerde. Do not launch the TUI. The program calls `run --auto` in the background.

The prompt is stdin; already redacted. Working directory is the customer folder only.

| Env | Anlam |
| --- | --- |
| `ICERDE_OPENCODE_BIN` | `opencode` yolu (boşsa PATH) |
| `ICERDE_OPENCODE_TIMEOUT_SEC` | tavan (varsayılan 480) |
| `ICERDE_OPENCODE_REQUIRE=1` | CLI yoksa iş başarısız |
| `ICERDE_LLM_API_KEY` | varsa `OPENAI_API_KEY` olarak OpenCode’a geçer |

CLI yoksa iskelet kalır; **sahte OpenCode yazımı yok**. Log: `llm/opencode.log`.

If the CLI is missing the scaffold stays. **No fake OpenCode write.** Log: `llm/opencode.log`.

Proje köküne `opencode.json` yazılır. `maestro` PATH’teyse Maestro MCP bağlanır (çalıştırma dilim 6).

Fine-tune (Colab, 16GB × 2 notebook) dilim 8 — OpenCode’un yerine geçmez.
