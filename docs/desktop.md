# Electron — masaüstü / desktop

**Kural:** [.cursor/rules/05-desktop.mdc](../.cursor/rules/05-desktop.mdc)

**TR:** Windows ve macOS. Main süreç yerel işleri; renderer yalnızca UI.

**EN:** Windows and macOS. Main owns local work; renderer is UI only.

## Süreç modeli / Process model

```mermaid
flowchart TB
  Main[main]
  Preload[preload]
  Renderer[Nextjs_renderer]
  Tray[tray]
  Runner[agent_runner]
  Activate[local_activate]
  McpHost[MCP_host]
  Main --> Preload
  Preload --> Renderer
  Main --> Tray
  Main --> Runner
  Main --> Activate
  Main --> McpHost
  McpHost --> ReadFile[mcp_read_file]
  McpHost --> Maestro[maestro_mcp]
```

## Güvenlik sınırı / Trust boundary

```mermaid
flowchart LR
  Renderer -->|ipc_whitelist| Preload
  Preload --> Main
  Main -->|GraphQL| API
  Main --> Child[child_processes]
```

- `nodeIntegration: false`, `contextIsolation: true`.
- Device fingerprint in main only.
- Generated project path stays in userData or a chosen workspace folder.
