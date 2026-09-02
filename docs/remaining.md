# Duraklama notu / Pause note

**Durum:** Dilim 1–6 bitti. Sıradaki iş **dilim 7 (işçi B + kuyruk)**.  
**Status:** Slices 1–6 done. Next is **slice 7 (worker B + queue)**.

Repo (private): https://cursor.com/codebase/emrahisik/agent-forge

## Bitti / Done

| Dilim | Ne |
| --- | --- |
| 1–3 | İskele, auth, proje diski + Maestro viewer |
| 4 LLM A + GDPR | redact → işçi A → tarama → denetim |
| 5 OpenCode | `opencode run --dir` GDPR’li prompt ile. CLI yoksa iskelet, sahte yazım yok. |
| 6 Yerel aktif + Maestro | Çocuk `go run` 47000–47999. Sidecar `vendor/bin`. Cihaz yok → SKIPPED, PASSED yok. |

LLM anahtarı yoksa `mock` kanal; `ICERDE_LLM_API_KEY` varsa HTTP + OpenCode’a `OPENAI_API_KEY`.

## Kalan / Remaining

7 ikinci işçi B + kuyruk → 8 Colab.

```bash
npm run dev:api
npm run dev:web
```
