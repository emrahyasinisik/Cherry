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

## İki backend / Two backends

```mermaid
flowchart TB
  IcerdeAPI[Icerde_Go_GraphQL]
  CustAPI[Generated_customer_API]
  IcerdeAPI --> Jobs[job_metadata]
  CustAPI --> Disk[files_on_disk]
```

Müşteriye dosya verilir. Kaynak müşterinin.
