# Colab fine-tune — dosyalar ve sıra / files and order

**Kural:** [.cursor/rules/11-llmops.mdc](../.cursor/rules/11-llmops.mdc) · [llmops.md](llmops.md)

**TR:** Colab’de eğitim **senin Google hesabında** olur. Cherry Colab barındırmaz ve Colab MCP değildir. Notebook’ları Colab’e yükle; **seed pack notebook içinde gömülü** (JSON yüklemek zorunlu değil). Adapter zip’ini indir, stüdyoda yeni `llmVersion` kaydet.

**EN:** Training runs on **your** Colab runtime. Cherry does not host Colab. Upload the notebooks; the **seed pack is embedded** (JSON upload optional). Download the adapter zip, register a new immutable version in the studio.

## Dosyalar / Files

Repo kökü `colab/`:

| Dosya | İş |
| --- | --- |
| `colab/cherry_worker_a.ipynb` | İşçi A, 16GB T4, QLoRA |
| `colab/cherry_worker_b.ipynb` | İşçi B, **aynı tarif** |
| `colab/examples/cherry_training_pack.json` | Seed paket (canlı iz yokken) |
| `colab/examples/cherry_sft.jsonl` | Aynı seed, satır satır |
| `colab/training_pack.schema.json` | JSON şema |

Stüdyo **LLM** sayfasından da indirilir: canlı paket (JSON + JSONL), notebook’lar, seed.

## Sıra / Order

```mermaid
flowchart LR
  Studio[LLM_yonetici] --> Tunnel[cloudflared_quick_tunnel]
  Tunnel --> Pack[training_pack_json]
  Pack --> NbA[Colab_A_T4]
  Pack --> NbB[Colab_B_T4]
  NbA --> ZipA[adapter_A_zip]
  NbB --> ZipB[adapter_B_zip]
  ZipA --> VerA[registerLlmVersion_A]
  ZipB --> VerB[registerLlmVersion_B]
  VerA --> Pointer[setActiveVersion]
  VerB --> Pointer
```

1. Cherry’de bir proje üret (brif + kod + Maestro). Yoksa seed paketi kullan.
2. LLM yönetici → **Colab tünelini aç** (`cloudflared` quick tunnel). URL + köprü token’ı notebook’a yapıştır. `cloudflared` yoksa tünel açılmaz — sahte URL yok; o zaman paketi elle indir.
3. [colab.research.google.com](https://colab.research.google.com) — **iki oturum**. Her birinde Runtime → GPU → **T4**.
4. Oturum A’ya `cherry_worker_a.ipynb` yükle. Seed gömülü — JSON şart değil. Tünel/upload varsa canlı paket öncelikli. Oturum B’ye `cherry_worker_b.ipynb`.
5. Hücreleri sırayla çalıştır. Adapter zip tünelden stüdyoya döner; yoksa makineye iner.
6. Tünel kaydı immutable `llmVersion` üretir. Pointer’ı **sen** çevirirsin. In-flight iş eski pointer’da biter. Elle yol: LLM yönetici → **Colab sürümü kaydet**.

## Cloudflare tüneli / Cloudflare tunnel

**TR:** Colab Google’da çalışır; stüdyo yerelde. El sıkışması `cloudflared tunnel --url` (trycloudflare quick tunnel). Cherry Colab barındırmaz. Bu, **Bağlantılar → Cloudflare Workers** değildir.

**EN:** Colab runs on Google; the studio is local. Handshake is `cloudflared tunnel --url` (trycloudflare quick tunnel). Cherry does not host Colab. This is **not** Connections → Cloudflare Workers.

```mermaid
flowchart TB
  subgraph studio [Cherry_yerel]
    UI[LLM_yonetici]
    API[Go_API_43148]
    Bridge[bridge_127_0_0_1]
    UI --> API
    API --> Bridge
  end
  CF[cloudflared_quick_tunnel]
  subgraph colab [Colab_T4]
    Nb[notebook_A_veya_B]
  end
  Bridge --> CF
  Nb -->|Bearer_token_pack_ve_zip| CF
  CF --> Bridge
```

| Bu / This | Değil / Not |
| --- | --- |
| Yerel köprü + public HTTPS (trycloudflare) | Cherry’de Colab runtime |
| Token’lı `GET /pack.json` + `POST /checkpoint` | Tüm GraphQL’i internete açmak |
| Sidecar `cloudflared` (vendor/bin veya PATH) | Bağlantılar’daki kişi Cloudflare hesabı |
| Colab paket çeker / adapter gönderir | Colab 24/7 üretim inferansı |

Köprü yalnızca `127.0.0.1` üzerinde dinler. Tünel o porta bakar — platform GraphQL, auth ve mailbox public olmaz. Token oturum token’ı değildir; tünel her açılışta yenilenir. GraphQL stüdyo oturumuna tam token’ı gösterir (yapıştırmak için); `tokenHint` son 4.

`cloudflared` yoksa durum `FAILED`, `publicUrl` boş. Sahte “bağlı” yok.

Env: `CHERRY_CLOUDFLARED_BIN`, `CHERRY_COLAB_BRIDGE_ADDR` (varsayılan `127.0.0.1:43149`).

## Sabitler / Invariants

| Bu / This | Değil / Not |
| --- | --- |
| QLoRA / LoRA, ~1.5B 4-bit, 16GB / oturum | Tam ağırlık büyük model |
| İki Colab = iki GPU hakkı | Tek T4’te A+B |
| Paket: brif + kaynak + Maestro + redakte tamamlamalar | Ham e-posta, token, `.env`, `preview/` |
| Checkpoint → immutable `llmVersion` | Colab 24/7 üretim API |
| Cloudflare quick tunnel = Colab el sıkışması | Müşteri Workers / D1 / R2 |

## Paket içeriği / Pack contents

`schema`: `cherry.training_pack.v1`

Her örnek: `instruction`, `input`, `output`.

### Seed vs canlı / Seed vs live

| Kural / Rule | Detay / Detail |
| --- | --- |
| Seed corpus | `services/api/internal/llm/seed_pack.go` — ~30+ somut satır (brief, kaynak, Maestro, completion). `colab/examples/` ile senkron. |
| Canlı pad | `liveExamples < 48` ise stüdyo seed’i pakete ekler. |
| Kısa audit | UI preview (~180 karakter) SFT’ye **girmez**; gürültü. |
| Notebook guard | `sft_rows < 24` → eğitim durur. 3 satırlık mini seed **yok**. |
| Embedded seed | `cherry_worker_{a,b}.ipynb` içinde base64 seed (~34 satır). Log: `no upload; using EMBEDDED seed pack 34`. |

**TR:** JSON yükleyemiyorsan sorun değil — notebook gömülü corpus ile `sft_rows=34` basmalı. Canlı paket varsa upload/tünel onu ezer.

**EN:** JSON upload is optional — notebooks print `sft_rows=34` from the embedded corpus. A live upload/tunnel pack overrides it.

Seed dosyalarını Go’dan yenilemek / Refresh examples from Go:

```bash
cd services/api
CHERRY_EXPORT_SEED=1 CHERRY_EXPORT_SEED_DIR=../../colab/examples \
  go test ./internal/llm -run TestExportSeedPackFiles
```

JSONL satırı:

```json
{"instruction":"...","input":"...","output":"..."}
```

## Geçici inferans / Temporary inference

Fine-tune sonrası model, aynı Colab oturumunda OpenAI uyumlu API olarak sunulabilir. Bu **geçici** bir çözümdür — Colab oturumu kapanınca bağlantı kopar. Üretim için kalıcı endpoint kullan.

### Named tunnel vs quick tunnel

| | Named tunnel (önerilen / preferred) | Quick tunnel (yedek / fallback) |
| --- | --- | --- |
| Komut | `cloudflared tunnel --no-autoupdate run --token …` | `cloudflared tunnel --url http://localhost:8000` |
| URL | Sabit alt alan (Zero Trust DNS), örn. `https://colab.yourdomain.com` | Rastgele `https://xxx.trycloudflare.com` |
| Token | Colab **secret** / env `CLOUDFLARE_TUNNEL_TOKEN` — **asla git’e yazma** | Yok |
| Cherry | İşçi A/B ayrı public HTTPS URL (`setColabInferenceUrl(slot, url)`) | Aynı — URL’ler yuvaya göre |
| Ne zaman | Sabit hostname istiyorsan | Token yoksa / hızlı deneme |

**TR:** Token **yalnızca Colab içinde** çalışır. Cherry Go API’sine token koyma. Stüdyo yalnızca public URL saklar. Sohbette veya ekran görüntüsünde token göründüyse Zero Trust → Tunnels’te **döndür (rotate)**.

**EN:** The tunnel token runs **inside Colab only**. Do not put it in the Cherry Go API process. The studio stores only the public HTTPS URL. If the token appeared in chat or a screenshot, **rotate** it in Zero Trust → Tunnels.

### Named tunnel adımları / Steps

1. Cloudflare Zero Trust → **Networks → Tunnels** → Create → named tunnel.
2. Public hostname ekle (örn. `colab.yourdomain.com`) → service `http://localhost:8000` (Colab’deki FastAPI).
3. DNS CNAME’i Cloudflare’in verdiği tunnel hedefine bağla.
4. Token’ı kopyala → Colab **Secrets** adı `CLOUDFLARE_TUNNEL_TOKEN` (veya oturumda `os.environ` / hücre değişkeni). Notebook’a, Drive’a, `.env` örnek dosyasına gerçek değer yazma.
5. İsteğe bağlı: `CHERRY_COLAB_PUBLIC_URL=https://colab.yourdomain.com` (Colab env veya hücre).
6. Notebook hücre 9’u çalıştır. Yazdırılan sabit URL’yi Cherry LLM yönetici → **Colab inferans · işçi A veya B** kartına yapıştır (`setColabInferenceUrl(slot, url)`). A ve B ayrı tünel kullanabilir.

Placeholder’lar (gerçek değer değil): `.env.example` içinde yorum satırı olarak `CHERRY_COLAB_PUBLIC_URL` ve `CLOUDFLARE_TUNNEL_TOKEN`.

### Akış / Flow

```mermaid
flowchart LR
  Train[Fine_tune_bitir] --> Serve[FastAPI_8000]
  Serve --> CF[cloudflared_named_or_quick]
  CF --> URL[fixed_or_trycloudflare_URL]
  URL --> Studio[Cherry_stüdyo]
  Studio --> Complete[/v1/chat/completions]
```

1. Notebook hücre 8: Model `model.eval()` ile inferans moduna geçer. FastAPI + uvicorn `0.0.0.0:8000`'de dinler.
2. Notebook hücre 9: Token varsa named tunnel (`run --token`); yoksa quick tunnel. Named: `CHERRY_COLAB_PUBLIC_URL` yazdırılır. Quick: log’dan `trycloudflare.com` URL parse edilir.
3. Notebook hücre 10: Oturumu canlı tutar (60s döngü).
4. Cherry stüdyoda: LLM yönetici → Colab inferans → URL'yi yapıştır. `setColabInferenceUrl` mutation'ı çağrılır (sabit veya trycloudflare — kısıtlama yok).
5. Cherry completer, `colab-tunnel` kanalıyla istekleri tünele yönlendirir. GDPR katmanı hâlâ aktif (redact → complete → scan → audit). OpenCode aynı URL’yi alır ama **`@ai-sdk/openai-compatible`** provider ile (`cherry-colab`) — built-in `openai` `/v1/responses` istediği için Colab’de 404 verir. Model: `Qwen/Qwen2.5-1.5B-Instruct` (`GET /v1/models` ile aynı).
6. Sağlık kontrolü 30 saniyede bir `/models` endpoint’ini yoklar (8s timeout, 3 deneme). Yanıt gelmezse durum `DISCONNECTED` olur ve varsayılan completer kullanılır. Tunnel `Complete` çağrıları 90s HTTP timeout kullanır.
7. LLM yönetici UI’da **Colab inferans** alanı `setColabInferenceUrl` çağırır — curl gerekmez.

**Not:** Stüdyo ↔ Colab **paket/adapter** köprüsü hâlâ quick tunnel kullanabilir (`startColabBridge`). Named tunnel öncelikle **Colab → public inferans** içindir.

### Sabitler / Invariants

| Bu / This | Değil / Not |
| --- | --- |
| Geçici inferans — Colab açıkken | 24/7 üretim API |
| Aynı 16GB T4, aynı QLoRA model | Tam ağırlık büyük model |
| `colab-tunnel` kanal, GDPR aktif | GDPR'sız doğrudan çağrı |
| Kullanıcı URL'yi elle yapıştırır | Otomatik keşif |
| Token yalnızca Colab secret/env | Token’ı Cherry API veya git’e yazmak |
| Bağlantı kopunca varsayılan endpoint | Yeniden bağlanma |

## GPU

Colab ücretsiz T4 ≈ 16GB. Notebook 4-bit `Qwen/Qwen2.5-1.5B-Instruct` + LoRA r=16. GPU yoksa hücre durur; CPU’ya düşmez.

## Gizlilik / Privacy

Paket `exportMe` değildir. Yalnızca redakte eğitim satırları. Başka kiracı yok. Colab’e yüklemeden JSON’u açıp `[REDACTED_` ve secret tarayın.
