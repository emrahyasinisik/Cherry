# İçerde

**TR:** İçerde bir mobil uygulama değildir. Windows ve Mac’te çalışan bir masaüstü stüdyosudur. Arka plandaki ajan mobil uygulamanın frontend ve backend kodunu yazar, yerelde ayağa kaldırır, **Maestro MCP** ile UI test eder. Müşteriye **klasör / zip / git** verilir. Barındırma yoktur.

**EN:** Icerde is not a mobile app. It is a Win/Mac desktop studio. A background agent writes a mobile app’s frontend and backend, boots it locally, and UI-tests it with **Maestro MCP**. The customer receives **folder / zip / git**. No hosting.

Bu depoda henüz uygulama kodu yoktur. Önce kurallar ve belgeler yazıldı.

**This repo has no application code yet.** Rules and documents come first.

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
| LLMOps | [.cursor/rules/11-llmops.mdc](.cursor/rules/11-llmops.mdc) | [docs/llmops.md](docs/llmops.md) |
| KVKK / GDPR | [.cursor/rules/12-gdpr-kvkk.mdc](.cursor/rules/12-gdpr-kvkk.mdc) | [docs/gdpr-kvkk.md](docs/gdpr-kvkk.md) |

Ajanlar için: [AGENTS.md](AGENTS.md)

## Çalıştır / Run (dilim 1 — iskele)

İki süreç: GraphQL API (`43148`) ve Next.js UI (`43147`). Bu ortamda Docker yoksa oturum **bellekte** kalır. Mongo için `docker compose up -d` ve `MONGO_URI=mongodb://127.0.0.1:27017`.

```bash
npm run dev:api
npm run dev:web
```

Masaüstü (API + web açıkken):

```bash
npm --prefix apps/desktop install
npm run dev:desktop
```

UI: http://127.0.0.1:43147 — herhangi bir e-posta + şifre (iskele; gerçek MFA dilim 2).

## Dilim durumu / Slice status

İskele çalışıyor. Sıradaki: [docs/build-order.md](docs/build-order.md) dilim 2 (auth + e-posta).

The scaffold is running. Next: auth + first-party mail.
