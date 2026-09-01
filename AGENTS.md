# Agent instructions — İçerde

Read this file before any implementation. Respond to the user in Turkish and English.

Kod yazmadan önce ilgili bölümün **kuralını ve belgesini** oku. Belge yoksa kod yazma.

Before writing code, read that section’s **rule and document**. If the document does not exist, do not code.

## Map

1. [docs/README.md](docs/README.md) — index and drawings overview
2. Matching `.cursor/rules/*.mdc` for the files you will touch
3. Matching `docs/*.md` for the drawing and invariants

## Hard product facts

- Icerde itself is **not** a mobile app. It is Electron on Windows and macOS.
- There are **two backends**. Never mix them:
  - Platform API: Go GraphQL + MongoDB (Icerde)
  - Generated customer backend: files on disk, activated locally for tests only
- Customer delivery is **files** (folder / zip / git). **No hosting** in v1.
- Mobile stack of the generated app is **chosen by the user**.
- Security is **X-inspired**: password, 6-digit new-device code, TOTP, trusted devices, session revoke. **No SMS. No phone identity.** One active session per user.
- Every LLM call goes through the **KVKK/GDPR layer**.
- Two LLMs: A = codegen, B = test/review. One-click switch changes subsequent answers.
- OpenCode is the code-writing engine; do not rewrite it.
- No inline imports. Exhaustive TypeScript switches with a `never` default.

## Intended layout (when coding starts)

```
apps/web/          Next.js (renderer UI)
apps/desktop/      Electron main + preload
services/api/      Go GraphQL
docs/              this documentation
.cursor/rules/     cursor rules per section
```

## Do not

- Do not add a mobile client for Icerde.
- Do not host customer backends.
- Do not send PII to an LLM without redaction.
- Do not implement a section without updating its doc if behavior changes.
