# Next.js UI

**Kural:** [.cursor/rules/04-frontend.mdc](../.cursor/rules/04-frontend.mdc) · [13-typescript.mdc](../.cursor/rules/13-typescript.mdc)

**TR:** Renderer, Electron içinde Next.js’tir. İçerde’nin mobil arayüzü yoktur.

**EN:** The renderer is Next.js inside Electron. Icerde has no mobile UI.

## Ekran haritası / Screen map

```mermaid
flowchart TB
  Login[Giris] --> Code[6_haneli_kod]
  Code --> TOTP[TOTP]
  TOTP --> Home[Projeler]
  Home --> Project[Proje]
  Project --> Agent[Ajan_log]
  Project --> Activate[Yerel_aktif]
  Project --> Maestro[Maestro_sonuclar]
  Home --> Security[Guvenlik_cihaz_oturum]
  Home --> Org[Org_uyeler]
  Home --> LLMAdmin[LLM_yonetici]
  Home --> Privacy[KVKK_sil_export]
```

## Durumlar / States

Her ekran: `empty` | `loading` | `error` | `ready`.

```mermaid
stateDiagram-v2
  [*] --> loading
  loading --> ready
  loading --> error
  loading --> empty
  error --> loading
  empty --> loading
  ready --> loading
```

## UI kuralları / UI rules

- shadcn/ui + Tailwind only. Theme = [design/tokens.css](design/tokens.css), not default violet.
- Visual system: [design-system.md](design-system.md). Motion: [motion.md](motion.md). Layouts: [screens.md](screens.md).
- LLM admin switch is one control; it must call the API and then show the new active model on the next completion.
- Do not fetch Mongo or filesystem from the browser; desktop preload handles files.
