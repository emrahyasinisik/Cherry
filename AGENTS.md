# Agent instructions — Cherry

Read this file before any implementation. Respond to the user in Turkish and English.

Kod yazmadan önce ilgili bölümün **kuralını ve belgesini** oku. Belge yoksa kod yazma.

Before writing code, read that section’s **rule and document**. If the document does not exist, do not code.

## Map

1. [docs/README.md](docs/README.md) — index and drawings overview
2. [docs/remaining.md](docs/remaining.md) — pause note: slices 1–8 done; Colab is files, not production inference
3. Matching `.cursor/rules/*.mdc` for the files you will touch
4. Matching `docs/*.md` for the drawing and invariants

## Hard product facts

- Cherry itself is **not** a mobile app. It is Electron on Windows and macOS.
- There are **two backends**. Never mix them:
  - Platform API: Go GraphQL + MongoDB (Cherry)
  - Generated customer backend: files on disk, activated locally for tests only
- Customer delivery is **files** (folder / zip / git). **No hosting** in v1. Zip is the chosen stack’s **language** and Clean Architecture (Expo TS, Flutter Dart, SwiftUI) — never `preview/` HTML.
- Mobile stack of the generated app is **chosen by the user** (Expo, Flutter, SwiftUI). Backend **target** is chosen via **Bağlantılar** (local / Supabase / Cloudflare / Render). GitHub push lives on that page.
- Security is **X-inspired**: password, 6-digit new-device code, TOTP, trusted devices, session revoke. **No SMS. No phone identity.** One active session per user.
- Every LLM call goes through the **KVKK/GDPR layer**.
- Two LLMs: **A and B are the same kind of worker**. They exist for **concurrent load** (e.g. 10 people at once), not for splitting job types (codegen vs test). The router assigns a free worker; if both are busy, jobs queue. Version pointer still changes subsequent answers on that worker.
- OpenCode is the code-writing engine; do not rewrite it.
- MCP/CLI: Maestro + OpenCode required in the product. Email and 6-digit codes are first-party (no AgentMail). Optional GitHub in this Cursor project. See [docs/integrations.md](docs/integrations.md).
- Follow [docs/build-order.md](docs/build-order.md): scaffold first, then one LLM slot (A). Colab notebooks live in `colab/` (slice 8).
- UI must look like a professional desktop atelier: [docs/design-system.md](docs/design-system.md), motion [docs/motion.md](docs/motion.md), screens [docs/screens.md](docs/screens.md). No slop gradients or spring animations.

## Intended layout (when coding starts)

```
package.json       npm workspaces root (apps/*) — single lockfile, single node_modules
apps/web/          Next.js (renderer UI)          workspace: cherry-web
apps/desktop/      Electron main + preload        workspace: cherry-desktop
services/api/      Go GraphQL (own Go module, outside the npm workspace)
docs/              this documentation
.cursor/rules/     cursor rules per section
```

Install from the repo root only (`npm install`). Never run `npm install` inside `apps/*`, and never commit a lockfile there. Add a dependency with `npm i <pkg> -w cherry-web`.

## Do not

- Do not add a mobile client for Cherry.
- Do not host customer backends.
- Do not send PII to an LLM without redaction.
- Do not implement a section without updating its doc if behavior changes.
- Do not ship default shadcn violet, bounce/spring motion, or mascot illustrations.
