# Mobil fabrika / Mobile factory

**Kural:** [.cursor/rules/08-mobile-factory.mdc](../.cursor/rules/08-mobile-factory.mdc)

**TR:** İçerde mobil uygulama **yazar**. Kendisi mobil değildir. Yığını kullanıcı seçer.

**EN:** Icerde **writes** mobile apps. It is not a mobile app. The user picks the stack.

## Boru hattı / Pipeline

```mermaid
flowchart LR
  Brief[Brif] --> Stack[Yigin_secimi]
  Stack --> Agent[Ajan_OpenCode]
  Agent --> FE[frontend]
  Agent --> BE[backend]
  Agent --> Flows[maestro_yaml]
  FE --> Layout[proje_klasoru]
  BE --> Layout
  Flows --> Layout
  Layout --> Handoff[zip_git]
```

## Çıktı ağacı / Output tree

```mermaid
flowchart TB
  Root[projectRoot]
  Root --> FE[frontend]
  Root --> BE[backend]
  Root --> M[maestro]
  Root --> R[README]
  Root --> D[docker_compose_optional]
```

## Yığın / Stack

Adapters behind one interface, for example:

- Expo / React Native
- Flutter
- Native iOS + Android

v1 may implement one adapter fully; others stay as explicit “not wired” rather than fake success.

## Stüdyo akışı / Studio flow

Kişi brifi yazar. Ajan **arka planda** `frontend/`, `backend/`, `maestro/` üretir. Test aşamasına gelince (veya kişi isterse) Maestro ekranı açılır: üretilen tasarım maketleri + YAML. Cihaz yoksa sonuç `SKIPPED`.

The user writes a brief. **OpenCode** writes `frontend/`, `backend/`, `maestro/` **in the background** (`opencode run --dir`, GDPR-wrapped prompt). At the test stage (or on demand) the Maestro screen opens: design mocks + YAML. No device → `SKIPPED`. Missing OpenCode CLI keeps the scaffold — never a fake write.


## İki backend / Two backends

```mermaid
flowchart TB
  IcerdeAPI[Icerde_Go_GraphQL]
  CustAPI[Generated_customer_API]
  IcerdeAPI --> Jobs[job_metadata]
  CustAPI --> Disk[files_on_disk]
```

Müşteriye dosya verilir. Kaynak müşterinin.
