# Duraklama notu / Pause note

**Durum:** Dilim 1–5 bitti. Sıradaki iş **dilim 6 (yerel aktif + Maestro MCP)**.  
**Status:** Slices 1–5 done. Next is **slice 6 (local activate + Maestro MCP)**.

Repo (private): https://cursor.com/codebase/emrahisik/agent-forge

## Bitti / Done

| Dilim | Ne |
| --- | --- |
| 1–3 | İskele, auth, proje diski + Maestro viewer |
| 4 LLM A + GDPR | redact → işçi A → tarama → denetim |
| 5 OpenCode | `opencode run --dir` GDPR’li prompt ile. CLI yoksa iskelet, sahte yazım yok. |

LLM anahtarı yoksa `mock` kanal; `ICERDE_LLM_API_KEY` varsa HTTP + OpenCode’a `OPENAI_API_KEY`.

## Kalan / Remaining

6 yerel aktif + Maestro MCP → 7 ikinci işçi B + kuyruk → 8 Colab (2 notebook, 16GB × oturum).

```bash
npm run dev:api
npm run dev:web
```
