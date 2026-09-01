# Yerel aktif / Local activate

**Kural:** [.cursor/rules/09-local-activate.mdc](../.cursor/rules/09-local-activate.mdc)

**TR:** Test için üretilen backend **localhost**’ta ayağa kalkar. Public barındırma yoktur.

**EN:** For tests, the generated backend boots on **localhost**. No public hosting.

## Aktivasyon / Activation

```mermaid
sequenceDiagram
  participant U as User
  participant UI as Nextjs
  participant Main as Electron_main
  participant Child as backend_process
  participant Maestro as Maestro_MCP
  U->>UI: activate
  UI->>Main: start_project
  Main->>Child: docker_or_run
  Child-->>Main: local_url_and_pid
  Main-->>UI: status_running
  Maestro->>Child: HTTP_to_localhost
  U->>UI: stop
  UI->>Main: stop_project
  Main->>Child: terminate
```

## Durum makinesi / Status

```mermaid
stateDiagram-v2
  [*] --> idle
  idle --> starting
  starting --> running
  starting --> failed
  running --> stopping
  failed --> idle
  stopping --> idle
```

## Portlar / Ports

- Icerde UI/API: uncommon ports (avoid 3000 / 5173 / 8080 when possible).
- Generated app: isolated range, recorded on the job (example `127.0.0.1:47000+`).
- Child process must not be the platform `services/api` binary.
