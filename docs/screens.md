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
- Nav (hedef): Projeler, Ajan, Maestro, Güvenlik, Organizasyon, **Bağlantılar**, LLM, Gizlilik.
- Titlebar: wordmark + proje adı + **LLM A | B** (sağ) — iki kapasite işçisi; meşgul/boş, kod vs test değil.
- Statusbar: oturum, cihaz, GDPR, OpenCode/Maestro sidecar, yerel URL.

## 1. Giriş / Login

Referans: [design/cherry-login.png](design/cherry-login.png)

```mermaid
flowchart TB
  Canvas[full_ink_canvas]
  Canvas --> Card[center_card_360]
  Card --> Mark[Cherry]
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

Referans: [design/cherry-studio.png](design/cherry-studio.png)

```mermaid
flowchart LR
  Brief[brif_ve_yigin]
  Chat[cherry_sohbet_OpenCode_gizli]
  Side[yerel_aktif_plus_Maestro]
```

Üç sütun, min 1280px pencerede. Orta: sohbet. Kişi yalnızca buraya yazar; OpenCode TUI yok.

## 5. LLM yönetici

Referans: [design/cherry-llm-admin.png](design/cherry-llm-admin.png)

İki kart A/B: aynı iş, ikinci kart yoğunluk işçisi. Versiyon select, MCP kök, kuyruk/meşgul durumu. Versiyon değişince chip metni anında yeni pointer.

Colab: eğitim paketi indir (JSON + JSONL), notebook A/B, seed paket. **Colab tüneli:** aç / kapat, trycloudflare URL + köprü token (mono), boş = tünel kapalı (elle yükle), hata = `cloudflared` yok — yeşil “bağlı” yok. Adapter zip sonrası `registerLlmVersion`; aktif pointer ayrı. Tünel ≠ Bağlantılar Cloudflare.

## 6. Güvenlik

Cihaz tablosu + oturum listesi + iptal. X’e benzer sakin liste; kırmızı yalnızca iptal.

## 7. Bağlantılar / Connections

Plugin menüsü. Kartlarda sağlayıcı logosu. Birincil yol **OAuth 2.0 izin ekranı** (“Cherry hesabına erişmek istiyor”). Token yapıştırma gelişmiş yedek. Cherry host olmaz. Boş / iptal bağlı değildir.

```mermaid
flowchart TB
  List[baglanti_listesi] --> Empty[empty_henuz_yok]
  List --> Row[logo_ad_durum]
  Row --> OAuth[oauth_izin_ekrani]
  OAuth --> Callback[code_exchange]
```

Kartlar (ilk set): Supabase, Cloudflare, GitHub, Vercel, Render. Boş: “Hesap bağla”. Hata: yetki yok / token yok — yeşil tik yok.

Proje: yığın (Expo SDK 57 / Flutter 3.47 / SwiftUI) + Clean Architecture + **backend hedefi**. Zip o dilin kaynağıdır; `preview/` HTML teslim değildir.

## 8. Maestro

Test aşaması veya yan menü. Telefon maketi (üretilen ekranlar) + YAML akış listesi. SKIPPED kırmızı “geçti” değildir.

```mermaid
flowchart LR
  Phone[ekran_maketi] --> Flows[yaml_akislar]
  Flows --> Result[skipped_or_run]
```

## 9. Gizlilik / org

Üye tablosu; `Verilerimi dışa aktar` / `Hesabı sil` ayrı, tehlikeli aksiyon onay dialog (320ms panel).
