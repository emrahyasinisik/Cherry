# Sidecar binaries — OpenCode + Maestro

**TR:** Müşteri OpenCode veya Maestro kurmaz. İçerde kurucusu bu klasöre CLI’leri koyar. Git’e ikili atılmaz.

**EN:** The customer does not install OpenCode or Maestro. The Icerde installer drops the CLIs here. Binaries are not committed.

## Look order / Arama sırası

1. `ICERDE_OPENCODE_BIN` / `ICERDE_MAESTRO_BIN`
2. `ICERDE_SIDECAR_DIR` (Electron sets this to `vendor/bin` or `resources/bin`)
3. `vendor/bin` walked from the API working directory
4. `PATH` (developer fallback only)

Kaynak etiketleri: `env` | `bundled` | `path` | `missing`.

## Vendor locally / Yerelde kopyala

```bash
./scripts/vendor-sidecars.sh
```

PATH’te `opencode` ve `maestro` varsa buraya kopyalar. Yoksa kurulum komutlarını yazar; uzaktan installer çalıştırmaz.
