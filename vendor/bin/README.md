# Sidecar binaries — OpenCode + Maestro + cloudflared

**TR:** Müşteri bu CLI’leri kurmaz. `./scripts/vendor-sidecars.sh` resmi OpenCode + Maestro zip’ini indirir (git’e ikili atılmaz). cloudflared PATH’teyse kopyalanır.

**EN:** The customer does not install these CLIs. `./scripts/vendor-sidecars.sh` downloads OpenCode and Maestro (not committed). cloudflared is copied from PATH when present.

## Look order / Arama sırası

1. `CHERRY_OPENCODE_BIN` / `CHERRY_MAESTRO_BIN` / `CHERRY_CLOUDFLARED_BIN`
2. `CHERRY_SIDECAR_DIR` only (when set — Electron `resources/bin`)
3. else `vendor/bin` walked from the API working directory, plus next-to-exe `resources/bin`
4. `PATH` (developer fallback)

Kaynak etiketleri: `env` | `bundled` | `path` | `missing`.

Maestro needs **Java 17+** (`JAVA_HOME`). Without an emulator/device, runs stay **SKIPPED** — never fake PASSED.

Yazmak için CLI yetmez; `CHERRY_LLM_API_KEY` (OpenCode’a `OPENAI_API_KEY`) veya bağlı Colab inferans URL gerekir. Anahtar yoksa CLI bulunur ama model çağrısı düşer — sahte yazım yok. `CHERRY_LLM_BASE_URL` / Colab URL → `OPENAI_BASE_URL` + `opencode.json` provider.
