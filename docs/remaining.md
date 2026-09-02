# Duraklama notu / Pause note

**Durum:** Dilim 1–6 bitti. Sıradaki kod **dilim 7 (işçi B + kuyruk)**.  
**Status:** Slices 1–6 done. Next code is **slice 7 (worker B + queue)**.

Kullanıcı notları (henüz dilim değil) aşağıda. Uygulama yokken spekülatif provider bağlama.

User notes below are not the current slice. Do not attach speculative providers.

## Kullanıcı notları / User notes (kayıt)

### 1. Zip HTML verdi, kaynak dil vermedi / Zip looked like HTML, not the stack language

**TR:** Teslim HTML değil. Expo SDK 57 / Flutter 3.47 / SwiftUI, Clean Architecture. `preview/` zip’e girmez. **Kodda düzeltildi.**

**EN:** Handoff is never an HTML site. Expo SDK 57 / Flutter 3.47 / SwiftUI with Clean Architecture. `preview/` is not in the zip. **Fixed in code.**

### 2. Bağlantılar menüsü / Connections menu

**TR:** Kişi backend’i istediği platforma koyabilmeli (Supabase, Cloudflare, benzeri). GitHub, Vercel, Render da bağlanmalı. Sidebar **Bağlantılar**. **Kodda: menü + token bağlama + GitHub push.** OAuth yok.

**EN:** Person attaches their own backends. **In code: menu + token connect + GitHub push.** No OAuth.

İçerde hâlâ barındırmaz. Ayrıntı: [connections.md](connections.md).

Sıra: dilim 7 → 8 Colab. Bağlantılar menüsü duruyor (token; OAuth sonra).

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
