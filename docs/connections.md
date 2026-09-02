# Bağlantılar / eklentiler — müşteri hesapları

**Kural:** [.cursor/rules/20-connections.mdc](../.cursor/rules/20-connections.mdc)

**TR:** Kişi, ürettiği uygulamanın backend’inin **nerede** duracağını seçer. İçerde barındırmaz. Hesaplar **Bağlantılar** (eklenti) menüsünden kişinin kendisine aittir.

**EN:** The person chooses **where** the generated app’s backend lives. Icerde does not host it. Accounts under **Connections** (plugins) belong to the person.

Bu dilim **kodlandı** (sidebar + sayfa + token kaydı + GitHub push). OAuth yok; kişi kendi token’ını yapıştırır. Sahte bağlı yok.

This slice **is coded** (sidebar + page + token store + GitHub push). No OAuth; the person pastes their token. No fake connected state.

## İki backend kuralı durur / Two-backend rule still holds

```mermaid
flowchart TB
  subgraph icerde [Icerde_platform]
    GQL[Go_GraphQL]
    Mongo[(MongoDB)]
  end
  subgraph customer [Musteri_uygulamasi]
    FE[frontend_secilen_yigin]
    BE[backend_secilen_hedef]
  end
  subgraph accounts [Kisinin_hesaplari]
    SB[Supabase]
    CF[Cloudflare]
    GH[GitHub]
    VE[Vercel]
    RE[Render]
  end
  GQL -.->|asla_karisma| BE
  BE --> SB
  BE --> CF
  FE --> GH
  BE --> GH
  BE --> VE
  BE --> RE
```

- İçerde API’si (Go GraphQL + Mongo) müşteri backend’i **değildir**.
- Müşteri backend’i dosyadır; hedef **yerel Go iskeleti**, **Supabase**, **Cloudflare** veya benzeri olabilir — kişi seçer.
- İçerde bu servislerde uygulama **barındırmaz**. Token ve proje kişinin hesabındadır.
- OpenCode, bağlı hedefe göre `backend/` (ve gerekirse frontend env) yazar.

## Menü / Menu

Sidebar: **Bağlantılar** (plugin / connections). Organizasyon’un altında, LLM’den önce.

Kişi oradan kendi bağlantılarını yapar. Proje oluştururken veya stüdyoda “backend hedefi” bu listeden seçilir.

## Backend hedefleri / Backend targets

Kişi “backend hangi platformda olsun” der. Örnek ilk set (kapalı liste değil; arayüz eklenti gibi büyür):

| Hedef | Ne işe yarar |
| --- | --- |
| **Yerel** (varsayılan v1) | `backend/` + localhost 47000–47999. Maestro burayı dener. |
| **Supabase** | Kişinin projesi: Auth, Postgres, Storage. Ajan müşteri API’sini buna yazar. |
| **Cloudflare** | Workers / D1 / R2 — kişi hesabı. |
| Benzeri (eklenti) | Aynı sözleşme: bağla → seç → ajan o hedefe yazar. Sahte “bağlandı” yok. |

Bağlı değilse hedef seçilemez; ajan uydurma URL yazmaz.

## Teslim / git / deploy

| Bağlantı | Ne |
| --- | --- |
| **GitHub** | Geliştirilen proje kişinin reposuna **çekilir / push**. Teslim zip’e ek, yerine geçmez. |
| **Vercel** | Kişi hesabına frontend (veya seçilen parça) deploy — İçerde host değil. |
| **Render** | Kişi hesabına backend/servis deploy — İçerde host değil. |

GitHub push: `frontend/` + `backend/` + `maestro/` + README — seçilen dil. **`preview/*.html` zip/git teslimine girmez.**

## Sırlar / Secrets

- Token’lar Interior sohbete ve LLM’e düz gitmez; KVKK redact.
- Secret’ı Cursor MCP’sine veya İçerde platform Mongo’suna “ürün sırrı” diye gömme.
- Bağlantı kopuk / yetki yok → hata; sessiz başarı yok.

## Yapılmayacak / Do not

- İçerde’nin kendi GraphQL’ine Supabase/Cloudflare karıştırma.
- Müşteri uygulamasını İçerde bulutunda host etme.
- AgentMail, SMS, Figma’yı bu menüye alma.
- Dilim 7–8’i (işçi B, Colab) bu iş için erteleme — sıra [build-order.md](build-order.md).
