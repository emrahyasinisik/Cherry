# MongoDB

**Kural:** [.cursor/rules/03-database.mdc](../.cursor/rules/03-database.mdc)

**TR:** Tek platform veritabanı. Müşteri uygulamasının runtime verisi burada tutulmaz.

**EN:** One platform database. The generated app’s runtime data does not live here.

## Koleksiyon ilişkisi / Collection map

```mermaid
flowchart TB
  users[users]
  orgs[organizations]
  members[memberships]
  devices[devices]
  sessions[sessions]
  codes[verificationCodes]
  mail[tempMailboxes]
  projects[projects]
  jobs[jobs]
  conns[connections]
  models[llmModels]
  versions[llmVersions]
  audit[auditEvents]
  users --> members
  orgs --> members
  users --> devices
  users --> sessions
  users --> codes
  users --> mail
  users --> projects
  users --> conns
  orgs --> projects
  projects --> jobs
  models --> versions
  users --> audit
  orgs --> audit
```

## Silme / Delete

```mermaid
flowchart LR
  Req[deleteMe] --> User[users_anonymize_or_drop]
  User --> Sess[sessions_revoke]
  User --> Dev[devices_drop]
  User --> Codes[codes_drop]
  User --> Mail[mail_drop]
  User --> Audit[audit_strip_PII]
  User --> Jobs[jobs_detach_or_drop]
  User --> Files[optional_project_dir_wipe]
```

## Alan kuralları / Field rules

- Passwords: argon2id or bcrypt. TOTP secret encrypted at rest.
- `verificationCodes.codeHash`, TTL 10 minutes, max attempts. Plain code is never stored. First-party mailer: [email-verification.md](email-verification.md).
- `tempMailboxes` is **our** collection (in-app inbox), not AgentMail.
- `sessions.tokenHash`, unique; at most one `active` per user.
- `projects.diskPath` is a pointer on the machine, not the source tree in GridFS.
- `auditEvents` immutable insert; redacted payload only.

## Adapter / Runtime

**TR:** `store.Store` arayüzü: `Memory` (varsayılan) veya `Mongo` (`MONGO_URI`). API ping başarısızsa memory’ye düşer. Koleksiyonlar: `users`, `devices`, `sessions`, `verificationCodes`, `tempMailboxes`, `projects`, `jobs`, `connections`, `llmVersions`, `llmState`, `auditEvents`.

**EN:** `store.Store` is either `Memory` (default) or `Mongo` (`MONGO_URI`). Failed ping falls back to memory. Collections: `users`, `devices`, `sessions`, `verificationCodes`, `tempMailboxes`, `projects`, `jobs`, `connections`, `llmVersions`, `llmState`, `auditEvents`.

```bash
# Option A — native mongod (no Docker): scripts/ensure-mongo.sh
# Option B — Docker:
docker compose up -d mongo

export MONGO_URI=mongodb://127.0.0.1:27017/cherry   # also in .env; npm run dev:api loads it
npm run dev:api
# /health → "store":"mongo"
```

Indexes: unique `users.email`, unique `sessions.tokenHash`, unique `(userId,fpHash)` on devices, unique `(userId,kind)` on connections, TTL on `verificationCodes.expiresAt`.
