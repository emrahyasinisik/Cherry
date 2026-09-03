# Sidecar binaries — OpenCode + Maestro

**TR:** Müşteri OpenCode, Maestro veya cloudflared kurmaz. `./scripts/vendor-sidecars.sh` resmi OpenCode binary’sini buraya indirir (git’e ikili atılmaz). Maestro / cloudflared PATH’teyse kopyalanır.

**EN:** The customer does not install OpenCode or Maestro. `./scripts/vendor-sidecars.sh` downloads the official OpenCode binary here (not committed). Maestro is copied from PATH when present.

## Look order / Arama sırası

1. `CHERRY_OPENCODE_BIN` / `CHERRY_MAESTRO_BIN` / `CHERRY_CLOUDFLARED_BIN`
2. `CHERRY_SIDECAR_DIR` only (when set — Electron `resources/bin`)
3. else `vendor/bin` walked from the API working directory, plus next-to-exe `resources/bin`
4. `PATH` (developer fallback)

Kaynak etiketleri: `env` | `bundled` | `path` | `missing`.

Yazmak için CLI yetmez; `CHERRY_LLM_API_KEY` (OpenCode’a `OPENAI_API_KEY`) veya bağlı Colab inferans URL gerekir. Anahtar yoksa CLI bulunur ama model çağrısı düşer — sahte yazım yok. `CHERRY_LLM_BASE_URL` / Colab URL → `OPENAI_BASE_URL` + `opencode.json` provider.

The CLI is not enough to write: `CHERRY_LLM_API_KEY` is forwarded as `OPENAI_API_KEY`, or a connected Colab inference URL is used. No key and no Colab URL → no fake write.
