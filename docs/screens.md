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
- Titlebar: wordmark + proje adı + **LLM A | B** segmented control (sağ).
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
  Log[ajan_log_mono]
  Side[yerel_aktif_plus_Maestro]
```

Üç sütun, min 1280px pencerede. Daralınca log alta, side sheet.

## 5. LLM yönetici

Referans: [design/icerde-llm-admin.png](design/icerde-llm-admin.png)

İki kart A/B, versiyon select, MCP kök, tek tuş “Nöbet değiştir”. Switch sonrası chip metni anında yeni model adı.

## 6. Güvenlik

Cihaz tablosu + oturum listesi + iptal. X’e benzer sakin liste; kırmızı yalnızca iptal.

## 7. Org / gizlilik

Üye tablosu; `Verilerimi dışa aktar` / `Hesabı sil` ayrı, tehlikeli aksiyon onay dialog (320ms panel).
