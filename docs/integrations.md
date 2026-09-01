# Entegrasyonlar — MCP, CLI, bağlama

**Kural:** [.cursor/rules/17-integrations.mdc](../.cursor/rules/17-integrations.mdc)

**TR:** Buraya her MCP’yi bağlama. Üç katman var; karışırsa ajan yanlış yere yazar.

**EN:** Do not attach every MCP. Three layers; mixing them makes the agent write in the wrong place.

```mermaid
flowchart TB
  subgraph cursorIDE [Cursor_IDE_gelistirme]
    CursorMCP[istege_GitHub_Maestro]
  end
  subgraph icerdeApp [Icerde_Electron_urun]
    Host[MCP_host]
    Host --> Read[mcp_read_file]
    Host --> Maestro[maestro_mcp]
    Mailer[birinci_parti_mailer]
  end
  subgraph writer [OpenCode_CLI]
    OC[opencode]
    OC --> MaestroOC[maestro_mcp]
  end
  CursorIDE -.->|sadece_biz_kodlarken| icerdeApp
  icerdeApp --> OC
```

## 1. Zorunlu — ürünün kendisi / Required for the product

Bunlar **Cursor’a süs için değil**, İçerde ve OpenCode’un çalışması için.

| Ne | Neden | Nasıl |
| --- | --- | --- |
| **Maestro CLI + `maestro mcp`** | UI test; LLM B ekranı görür | `curl -fsSL "https://get.maestro.mobile.dev" \| bash` sonra `maestro mcp`. Java 17+, `JAVA_HOME`. |
| **OpenCode CLI** | Ajanın kod yazma motoru | OpenCode kurulumu; proje `opencode.json` içinde Maestro bağlı. |
| **Node.js LTS + Go 1.22+ + Docker** | Next.js, GraphQL, Mongo | Docker Compose ile Mongo. |
| **Emülatör** | Maestro’nun gerçek cihazı | Android Studio AVD (Win/Mac). iOS Simulator yalnızca Mac. Yoksa test `skipped`, asla sahte `passed`. |
| **SMTP (prod)** | Bizim e-posta + 6 hane | Env ile mailer. AgentMail yok. Dev: in-app kutu. [email-verification.md](email-verification.md) |

Maestro **İçerde Electron MCP host** ve **OpenCode** içine bağlanır. Sadece Cursor’a bağlayıp üründe unutmak yetmez.

Örnek: [../opencode.json.example](../opencode.json.example)

## 2. Bu Cursor projesinde / In this Cursor project

Geliştirirken (İçerde’yi biz yazarken):

| Ne | Durum | Ne işe yarar |
| --- | --- | --- |
| **GitHub MCP** | İsteğe bağlı | Müşteri teslimi git ise. Zip-only ise gerekmez. |
| **Maestro MCP** | Geliştirici makinede | Flow’ları sen elle denersin. Ürün host’unun yerine geçmez. |

Örnek: [../.cursor/mcp.json.example](../.cursor/mcp.json.example) → kopyala `.cursor/mcp.json` (secret koyma, git’e atma).

**E-posta / 6 haneli kod için MCP bağlama.** Birinci parti: [email-verification.md](email-verification.md).

## 3. Bağlama / Do not connect

| Ne | Neden |
| --- | --- |
| **AgentMail / inbox SaaS MCP** | Mail ve 6 haneyi biz yazıyoruz. |
| Figma / Canva MCP | İçerde Figma’dan üretmez (şimdilik). |
| Rastgele filesystem / shell MCP | Electron zaten `mcp_read_file` sunar; kök admin’de. |
| MongoDB MCP | Platform Docker ile konuşur; şart değil. |
| SMS / Twilio | Güvenlik modeli SMS yasak. |

Colab **MCP değil**: Google hesabı + notebook export. Cursor’a bağlanmaz.

## Kurulum sırası / Install order (senin makinen)

1. Java 17+, Node LTS, Go, Docker Desktop
2. Maestro CLI:

```bash
curl -fsSL "https://get.maestro.mobile.dev" | bash
maestro --help
```
3. OpenCode CLI → `opencode --help`
4. Android emulator (ve Mac’te Xcode sim)
5. İsteğe `.cursor/mcp.json` (GitHub / Maestro — **mail yok**)
6. OpenCode: `opencode.json` ile Maestro
7. Prod: `SMTP_*` env

Cloud agent ortamında emülatör ağır olabilir; orada Maestro mock/`skipped` kabul.
