# Ekran tasarımları / Screen designs

**Kural:** [15-design.mdc](../.cursor/rules/15-design.mdc) · [frontend.md](frontend.md)

Her ana ekranın iskeleti. Uygulama bu hiyerarşiyi bozmaz.

Each primary screen has a skeleton. Implementation must keep this hierarchy.

## Kabuk / App shell

```mermaid
flowchart TB
  subgraph window [Electron_window]
    Titlebar[titlebar_48_LLM_switch]
    subgraph body [body]
      Nav[sidebar_220]
      Main[main]
    end
    Status[statusbar_32_oturum_GDPR]
  end
```

- Sidebar 220px, collapse 64px (ikon). Mobil tab bar yok.
- Titlebar: wordmark + proje adı + **LLM A | B** (sağ) — iki kapasite işçisi; meşgul/boş, kod vs test değil.
- Statusbar: oturum, cihaz, GDPR, yerel URL.

## 1. Giriş / Login

Referans: [design/icerde-login.png](design/icerde-login.png)

```mermaid
flowchart TB
  Canvas[full_ink_canvas]
  Canvas --> Card[center_card_360]
  Card --> Mark[Icerde]
  Card --> Sub[Masaustu_stüdyo]
  Card --> Email[email]
  Card --> Pass[sifre]
  Card --> CTA[Giris_yap]
  Card --> Hint[yeni_cihaz_6_hane]
```

Boş/hata: kart içinde kırmızı satır, shake yok.

## 2. 6 hane + TOTP

Altı kutu, `font-mono`, 40×48, 8px gap. Yapıştırınca dağılır. TOTP: tek alan + “Authenticator uygulaması”.

## 3. Projeler / Projects

```mermaid
flowchart LR
  List[proje_listesi] --> Empty[empty_ilk_proje]
  List --> Row[ad_yigin_durum_tarih]
```

Empty: “Henüz proje yok” + “Yeni proje”. Lorem yok.

## 4. Stüdyo / Studio

Referans: [design/icerde-studio.png](design/icerde-studio.png)

```mermaid
flowchart LR
  Brief[brif_ve_yigin]
  Chat[icerde_sohbet_OpenCode_gizli]
  Side[yerel_aktif_plus_Maestro]
```

Üç sütun, min 1280px pencerede. Orta: sohbet. Kişi yalnızca buraya yazar; OpenCode TUI yok.

## 5. LLM yönetici

Referans: [design/icerde-llm-admin.png](design/icerde-llm-admin.png)

İki kart A/B: aynı iş, ikinci kart yoğunluk işçisi. Versiyon select, MCP kök, kuyruk/meşgul durumu. Versiyon değişince chip metni anında yeni pointer.

## 6. Güvenlik

Cihaz tablosu + oturum listesi + iptal. X’e benzer sakin liste; kırmızı yalnızca iptal.

## 8. Maestro

Test aşaması veya yan menü. Telefon maketi (üretilen ekranlar) + YAML akış listesi. SKIPPED kırmızı “geçti” değildir.

```mermaid
flowchart LR
  Phone[ekran_maketi] --> Flows[yaml_akislar]
  Flows --> Result[skipped_or_run]
```


Üye tablosu; `Verilerimi dışa aktar` / `Hesabı sil` ayrı, tehlikeli aksiyon onay dialog (320ms panel).
