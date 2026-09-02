# E-posta ve 6 haneli kod — birinci parti

**Kural:** [.cursor/rules/18-email-verification.mdc](../.cursor/rules/18-email-verification.mdc)

**TR:** E-posta gönderimi, geçici kutu ve 6 haneli / link doğrulama **bizim kodumuz**. AgentMail, Resend-MCP, harici “inbox API” yok. Resend yalnızca Go sürecinden HTTP API; SMTP `net/smtp`.

**EN:** Mail send, temp inbox, and 6-digit / link verification are **first-party**. No AgentMail, Resend-MCP, or third-party inbox API. Resend is our HTTP call; SMTP is `net/smtp`.

## Nasıl gerçek mail atılır / How to send real mail

Kod **her zaman** in-app `tempMailboxes` satırına yazılır. Gerçek çıkış için **birini** doldur:

1. **SMTP** (Gmail uygulama şifresi, Workspace, vs.)
2. **Resend** (`RESEND_API_KEY` + doğrulanmış `RESEND_FROM`)

İkisi de yoksa kanal `inbox`: geliştirmede kutu + log. GraphQL `emailSent: false`. Sahte “gönderildi” yok.

Production’da `CHERRY_MAIL_REQUIRE=1` — çıkış yoksa kayıt/giriş kod adımı hata verir.

```bash
# Gmail app password (not the account password)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=you@gmail.com
SMTP_PASSWORD=xxxx xxxx xxxx xxxx
SMTP_FROM=Cherry <you@gmail.com>

# or Resend
RESEND_API_KEY=re_...
RESEND_FROM=Cherry <hello@yourdomain.com>
```

`register` / `login` yanıtı: `emailSent`, `emailChannel` (`inbox` | `smtp` | `resend`). UI, gönderildiyse kodu kartta göstermez.

## Akış / Flow

```mermaid
sequenceDiagram
  participant U as User
  participant App as Electron
  participant API as GoGraphQL
  participant Mailer as Cherry_mailer
  participant Box as tempMailboxes
  participant Out as SMTP_or_Resend
  participant DB as verificationCodes
  U->>App: login_or_new_device
  App->>API: request_code
  API->>DB: store_codeHash_TTL
  API->>Box: insert_message
  alt RESEND_API_KEY
    Mailer->>Out: HTTPS Resend
    Out-->>U: account_email
  else SMTP_HOST
    Mailer->>Out: STARTTLS SMTP
    Out-->>U: account_email
  else no transport
    Mailer-->>Box: inbox_only_dev_sink
  end
  U->>App: six_digits_or_link
  App->>API: verify_code
  API->>DB: hash_compare_attempts
```

## Ne yazıyoruz / What we own

| Parça | Nerede | Not |
| --- | --- | --- |
| Mailer | `services/api/internal/mailer` | SMTP env veya Resend HTTP. Kütüphane taşıma; ürün mantığı bizde. |
| Şablon | düz metin + HTML | Cherry wordmark, kod 6 hane mono, TTL metni. |
| `verificationCodes` | store / Mongo | `codeHash`, `purpose`, `expiresAt`, `attempts`, `userId`, `deviceId` |
| `tempMailboxes` | store / Mongo | In-app kutu. Gövde PII; silmede drop. |
| UI | Next.js | 6 kutu + “e-postadaki link”; güvenlik ekranında kutu listesi |
| Dev sink | mailer | SMTP/Resend yoksa kutu + log; `emailSent=false`. |

## Kod kuralları / Code rules

- 6 hane: `crypto/rand`, `000000–999999`, **yalnızca hash** saklanır (pepper + SHA-256). Düz kod Mongo’da yok.
- TTL ~10 dakika. Max 5 deneme sonra kilit; yeni kod eskiyi geçersiz kılar.
- Amaçlar (`purpose`): `new_device`, `login_challenge`, `email_verify`, `suspicious_login`. Switch exhaustive.
- Link: imzalı token (aynı doğrulama kaydı); UI hâlâ 6 haneyi de kabul eder.
- Rate limit: kullanıcı + IP. SMS yok.
- GDPR: kod ve kutu gövdesi LLM’e gitmez. `deleteMe` kod + kutu siler.
- Login mutation payload’ında düz kod yok. Üretimde kart kodu dump etmez.

## Yasak / Forbidden

- AgentMail, harici geçici-posta SaaS, SMS, telefon.
- Cursor MCP ile mail göndermek.
- Dev’de kodu GraphQL error mesajında döndürmek (yalnızca `tempMailboxes` + sunucu log, asla production response).
