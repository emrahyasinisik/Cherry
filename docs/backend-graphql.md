# GraphQL Go — platform API

**Kural:** [.cursor/rules/02-backend-graphql.mdc](../.cursor/rules/02-backend-graphql.mdc) · [14-go.mdc](../.cursor/rules/14-go.mdc)

**TR:** Bu, İçerde’nin kendi backend’idir. Müşteriye verilen mobil uygulamanın API’si değildir.

**EN:** This is Icerde’s own backend. It is not the API of the generated mobile app.

## İstek yolu / Request path

```mermaid
sequenceDiagram
  participant UI as Nextjs
  participant Pre as Electron_preload
  participant GQL as gqlgen
  participant Auth as AuthZ
  participant GDPR as GDPR_layer
  participant Store as Mongo
  UI->>Pre: GraphQL_operation
  Pre->>GQL: HTTPS_or_local
  GQL->>Auth: session_and_device
  alt llm_operation
    GQL->>GDPR: redact_and_route
    GDPR->>Store: audit
  else data_operation
    GQL->>Store: query_or_mutate
  end
  GQL-->>UI: typed_payload
```

## Modül sınırları / Module bounds

```mermaid
flowchart TB
  Schema[schema_graphqls] --> Resolvers[resolvers]
  Resolvers --> Auth[auth]
  Resolvers --> Mail[mailer_verify]
  Resolvers --> Orgs[orgs]
  Resolvers --> Projects[projects_jobs]
  Resolvers --> LLM[llm_router]
  Resolvers --> Privacy[gdpr_export_delete]
  Auth --> Store[mongo_store]
  Mail --> Store
  Orgs --> Store
  Projects --> Store
  LLM --> Store
  Privacy --> Store
```

## İlk şema grupları / First schema groups

- `Auth`: register, login, verifyCode, totp, devices, sessions, logout, mailbox (first-party; no AgentMail)
- `Workspace`: personal profile, organizations, members
- `Factory`: projects, jobs, artifacts path metadata (not file blobs in Mongo)
- `LLMOps`: versions, active pointer, switch, mcpRoot, trainingPack, registerLlmVersion
- `Privacy`: exportMe, deleteMe, audit (admin)

Health: `GET /health`. GraphQL otherwise, plus `/export/:id` (zip), `/colab/*` (notebook + seed files), `/oauth/*`.
