# Duraklama notu / Pause note

**Durum:** Dilim 1–8 bitti. Colab **dosya** olarak var; üretim inferansı değil.  
**Status:** Slices 1–8 done. Colab ships as **files**, not production inference.

Kullanıcı notları (henüz dilim değil) aşağıda.

User notes below are not the current slice.

## Kullanıcı notları / User notes (kayıt)

### 1. Zip HTML verdi, kaynak dil vermedi / Zip looked like HTML, not the stack language

**TR:** Teslim HTML değil. Expo SDK 57 / Flutter 3.47 / SwiftUI, Clean Architecture. `preview/` zip’e girmez. **Kodda düzeltildi.**

**EN:** Handoff is never an HTML site. Expo SDK 57 / Flutter 3.47 / SwiftUI with Clean Architecture. `preview/` is not in the zip. **Fixed in code.**

### 2. Bağlantılar menüsü / Connections menu

**TR:** Sidebar **Bağlantılar**. OAuth 2.0 izin ekranı + logolar. Cherry host değil.

**EN:** Connections: OAuth 2.0 consent + marks. Cherry does not host.

### 3. Colab belgeleri / Colab files

**TR:** `colab/` notebook + seed paket + LLM sayfasından indirme. İki T4 oturumu. Cloudflare quick tunnel ile stüdyo ↔ Colab. **Kodda.**

**EN:** `colab/` notebooks + seed pack + LLM admin download. Two T4 sessions. Studio ↔ Colab via Cloudflare quick tunnel. **In code.**

### 4. Marka / Brand

**TR:** Stüdyo adı **Cherry**. İçerde adı kalktı.

**EN:** The studio brand is **Cherry**. The Icerde name is gone.

## Bitti / Done

| Dilim | Ne |
| --- | --- |
| 1–3 | İskele, auth, proje diski + Maestro viewer |
| 4 LLM A + GDPR | redact → işçi A → tarama → denetim |
| 5 OpenCode | `opencode run --dir` GDPR’li prompt ile. CLI yoksa iskelet, sahte yazım yok. |
| 6 Yerel aktif + Maestro | Çocuk `go run` 47000–47999. Sidecar `vendor/bin`. Cihaz yok → SKIPPED, PASSED yok. |
| 7 LLM B + kuyruk | A ve B aynı iş. Boş olan alır; ikisi meşgulse kuyruk. Versiyon pointer’ı sonraki cevabı değiştirir; in-flight eski pointer’da biter. |
| 8 Colab fine-tune | `colab/cherry_worker_{a,b}.ipynb`, seed paket, `trainingPack` + `registerLlmVersion`. Colab MCP/üretim değil. |

LLM anahtarı yoksa `mock` kanal; `CHERRY_LLM_API_KEY` varsa HTTP + OpenCode’a `OPENAI_API_KEY`. Bağlı Colab inferans (`setColabInferenceUrl`) → `colab-tunnel` (health `/llm` alanı da bunu gösterir). OpenCode CLI `vendor/bin` ( `./scripts/vendor-sidecars.sh` ).

## Kalan / Remaining

Colab stüdyo işçisi değildir. Mongo adapter’ları, gerçek OAuth app, emülatörlü Maestro PASSED — ürün kararı, zorunlu dilim yok.

```bash
npm run dev:api
npm run dev:web
```
