# Agent instructions — İçerde

Read this file before any implementation. Respond to the user in Turkish and English.

Kod yazmadan önce ilgili bölümün **kuralını ve belgesini** oku. Belge yoksa kod yazma.

Before writing code, read that section’s **rule and document**. If the document does not exist, do not code.

## Map

1. [docs/README.md](docs/README.md) — index and drawings overview
2. [docs/remaining.md](docs/remaining.md) — pause note: next is slice 4 (LLM A + GDPR)
3. Matching `.cursor/rules/*.mdc` for the files you will touch
4. Matching `docs/*.md` for the drawing and invariants

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
- MCP/CLI: Maestro + OpenCode required in the product. Email and 6-digit codes are first-party (no AgentMail). Optional GitHub in this Cursor project. See [docs/integrations.md](docs/integrations.md).
- Follow [docs/build-order.md](docs/build-order.md): scaffold first, then one LLM slot (A), not Colab or LLM B.
- UI must look like a professional desktop atelier: [docs/design-system.md](docs/design-system.md), motion [docs/motion.md](docs/motion.md), screens [docs/screens.md](docs/screens.md). No slop gradients or spring animations.

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
- Do not ship default shadcn violet, bounce/spring motion, or mascot illustrations.
