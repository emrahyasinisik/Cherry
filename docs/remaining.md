# Duraklama notu / Pause note

**Durum:** Dilim 1–3 bitti. Sıradaki iş **dilim 4 (LLM A + GDPR)**.  
**Status:** Slices 1–3 done. Next is **slice 4 (LLM A + GDPR)**.

Repo (private): https://cursor.com/codebase/emrahisik/agent-forge

## Bitti / Done

| Dilim | Ne |
| --- | --- |
| 1 İskele | Electron + Next.js + Go GraphQL |
| 2 Auth + posta | Şifre, 6 hane, TOTP, SMTP/Resend |
| 3 Proje diski | Brif → arka plan stub yazıcı → `frontend/` `backend/` `maestro/` → zip. Test aşamasında veya talepte Maestro ekranı (tasarım + YAML, cihaz yoksa SKIPPED) |

## Ürün akışı (dilim 3) / Product flow

Kişi uygulamayı tarif eder → ajan arka planda yazar → **test aşamasında Maestro kendiliğinden açılır** veya yan menüden istenir. Ekran maketleri + flow YAML görünür. Emülatör yok = SKIPPED, sahte geçiş yok. Gerçek `maestro mcp` dilim 6.

## Kalan / Remaining

4 LLM A + GDPR → 5 OpenCode (stub yerine gerçek yazıcı) → 6 yerel aktif + Maestro MCP → 7 LLM B switch → 8 Colab.

Açık uç: Mongo bellek, gerçek mail `.env`, Electron installer yok.

```bash
npm run dev:api
npm run dev:web
```
