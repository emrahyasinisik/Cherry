# Duraklama notu / Pause note

**Durum:** Dilim 1–4 bitti. Sıradaki iş **dilim 5 (OpenCode)**.  
**Status:** Slices 1–4 done. Next is **slice 5 (OpenCode)**.

Repo (private): https://cursor.com/codebase/emrahisik/agent-forge

## Bitti / Done

| Dilim | Ne |
| --- | --- |
| 1–3 | İskele, auth, proje diski + Maestro viewer |
| 4 LLM A + GDPR | redact → LLM A → tarama → denetim. Versiyon pointer’ı sonraki cevapları değiştirir. MCP kök = proje klasörü. B ve Colab yok. |

LLM anahtarı yoksa `mock` kanal; `ICERDE_LLM_API_KEY` varsa HTTP. Her iki yol da GDPR’den geçer.

## Kalan / Remaining

5 OpenCode → 6 yerel aktif + Maestro MCP → 7 LLM B switch → 8 Colab.

```bash
npm run dev:api
npm run dev:web
```
