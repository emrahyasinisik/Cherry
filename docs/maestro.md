# Maestro MCP

**Kural:** [.cursor/rules/10-maestro.mdc](../.cursor/rules/10-maestro.mdc)

**TR:** UI denemesi Maestro MCP ile. Flow’lar müşteri dosyalarına yazılır.

**EN:** UI trials go through Maestro MCP. Flows are written into the customer files.

## Kurulum / Install

Resmi CLI (Java 17+, `JAVA_HOME`):

```bash
curl -fsSL "https://get.maestro.mobile.dev" | bash
```

MCP sunucusu CLI’nin içinde: `maestro mcp`. Doğrulama: `maestro --help`.

## Test döngüsü / Test loop

```mermaid
flowchart TB
  Start[LLM_B_test] --> List[list_devices]
  List --> Boot[emulator_or_skip]
  Boot --> Inspect[inspect_screen]
  Inspect --> Write[write_yaml_flow]
  Write --> Run[run_flow]
  Run --> Pass[pass_save_flow]
  Run --> Fail[fail_screenshot]
  Fail --> Fix[agent_fix_code_or_flow]
  Fix --> Inspect
  Pass --> Report[job_report]
```

## MCP araçları / MCP tools

```mermaid
flowchart LR
  Host[Electron_MCP_host] --> LD[list_devices]
  Host --> IS[inspect_screen]
  Host --> SS[take_screenshot]
  Host --> Run[run]
  Host --> View[open_maestro_viewer]
```

Dilim 3 viewer: stüdyo UI (telefon maketi + YAML). Dilim 6: gerçek `maestro mcp`.


## Kurallar / Rules

- YAML under `maestro/` ships with zip/git.
- No emulator: status `skipped`, never `passed`.
- Bound repair attempts (suggested max 3 per flow) then fail the job.
- Maestro talks to the **locally activated** customer API, not Icerde GraphQL.
