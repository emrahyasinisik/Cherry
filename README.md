# Cherry

**TR:** Cherry bir mobil uygulama değildir. Windows ve Mac’te çalışan bir masaüstü stüdyosudur. Arka plandaki ajan mobil uygulamanın frontend ve backend kodunu yazar, yerelde ayağa kaldırır, **Maestro MCP** ile UI test eder. Müşteriye **klasör / zip / git** verilir. Barındırma yoktur.

**EN:** Cherry is not a mobile app. It is a Win/Mac desktop studio. A background agent writes a mobile app’s frontend and backend, boots it locally, and UI-tests it with **Maestro MCP**. The customer receives **folder / zip / git**. No hosting.

Ajanlar için: [AGENTS.md](AGENTS.md)

## Belgeler / Documents

Başlangıç noktası: [docs/README.md](docs/README.md)

| Bölüm / Section | Kurallar / Rules | Belge + çizim / Doc + diagram |
| --- | --- | --- |
| Mimari | [.cursor/rules/01-architecture.mdc](.cursor/rules/01-architecture.mdc) | [docs/architecture.md](docs/architecture.md) |
| GraphQL Go | [.cursor/rules/02-backend-graphql.mdc](.cursor/rules/02-backend-graphql.mdc) | [docs/backend-graphql.md](docs/backend-graphql.md) |
| MongoDB | [.cursor/rules/03-database.mdc](.cursor/rules/03-database.mdc) | [docs/database.md](docs/database.md) |
| Next.js UI | [.cursor/rules/04-frontend.mdc](.cursor/rules/04-frontend.mdc) | [docs/frontend.md](docs/frontend.md) |
| Tasarım | [.cursor/rules/15-design.mdc](.cursor/rules/15-design.mdc) | [docs/design-system.md](docs/design-system.md) · [docs/screens.md](docs/screens.md) |
| Hareket | [.cursor/rules/16-motion.mdc](.cursor/rules/16-motion.mdc) | [docs/motion.md](docs/motion.md) |
| MCP / CLI | [.cursor/rules/17-integrations.mdc](.cursor/rules/17-integrations.mdc) | [docs/integrations.md](docs/integrations.md) |
| E-posta / 6 hane | [.cursor/rules/18-email-verification.mdc](.cursor/rules/18-email-verification.mdc) | [docs/email-verification.md](docs/email-verification.md) |
| Yapım sırası | [.cursor/rules/19-build-order.mdc](.cursor/rules/19-build-order.mdc) | [docs/build-order.md](docs/build-order.md) |
| Electron | [.cursor/rules/05-desktop.mdc](.cursor/rules/05-desktop.mdc) | [docs/desktop.md](docs/desktop.md) |
| Güvenlik | [.cursor/rules/06-security.mdc](.cursor/rules/06-security.mdc) | [docs/security.md](docs/security.md) |
| Org / kişisel | [.cursor/rules/07-organizations.mdc](.cursor/rules/07-organizations.mdc) | [docs/organizations.md](docs/organizations.md) |
| Mobil fabrika | [.cursor/rules/08-mobile-factory.mdc](.cursor/rules/08-mobile-factory.mdc) | [docs/mobile-factory.md](docs/mobile-factory.md) |
| Yerel aktif | [.cursor/rules/09-local-activate.mdc](.cursor/rules/09-local-activate.mdc) | [docs/local-activate.md](docs/local-activate.md) |
| Maestro | [.cursor/rules/10-maestro.mdc](.cursor/rules/10-maestro.mdc) | [docs/maestro.md](docs/maestro.md) |
| LLMOps | [.cursor/rules/11-llmops.mdc](.cursor/rules/11-llmops.mdc) | [docs/llmops.md](docs/llmops.md) · [docs/colab.md](docs/colab.md) |
| KVKK / GDPR | [.cursor/rules/12-gdpr-kvkk.mdc](.cursor/rules/12-gdpr-kvkk.mdc) | [docs/gdpr-kvkk.md](docs/gdpr-kvkk.md) |

## Çalıştır / Run

Depo bir **npm workspaces monorepo**'sudur: bağımlılıklar kökten tek seferde kurulur, tek `package-lock.json` ve tek `node_modules` vardır. Alt klasörlerde `npm install` çalıştırma.

**TR:** İki süreç: GraphQL API (`43148`) ve Next.js UI (`43147`). SMTP veya Resend yoksa kod **in-app geçici kutuya** düşer (`emailSent: false`). Oturum bellekte (Docker yoksa).

```bash
npm install          # bir kez, kökte — apps/web + apps/desktop
npm run dev:api      # Go GraphQL, 43148 — .env yükler, scripts/ensure-mongo.sh
npm run dev:web      # Next.js, 43147
npm run dev:desktop  # Electron kabuk (dev sunucuya bağlanır)
npm run dist:desktop # Win NSIS/zip veya Mac dmg/zip — müşteri Next/Go kurmaz
```

`dev:api` `MONGO_URI` doluysa native `mongod`’u ayağa kaldırmaya çalışır (`scripts/ensure-mongo.sh`). Docker tercih: `docker compose up -d mongo`.

Tek bir workspace'e paket eklemek için: `npm i <paket> -w cherry-web`.

Gerçek e-posta için `.env.example` → `SMTP_*` (Gmail uygulama şifresi) **veya** `RESEND_API_KEY`. Ayrıntı: [docs/email-verification.md](docs/email-verification.md).

UI: http://127.0.0.1:43147 — **Hesap oluştur** (şifre ≥ 8). Kod e-postaya gittiyse kartta görünmez; gitmediyse geliştirme kutusu. Güvenlik ekranı: cihaz, oturum, TOTP, kutu. **Bağlantılar** = OAuth 2.0 izin ekranı (logolar + “hesabına erişmek istiyor”). GitHub client id yoksa yerel onay; varsa github.com.

## Dilim durumu / Slice status

Dilim 1–8 bitti. Colab dosyaları: [colab/](colab/) ve [docs/colab.md](docs/colab.md). Kalan: [docs/remaining.md](docs/remaining.md).

