# E-posta ve 6 haneli kod — birinci parti

**Kural:** [.cursor/rules/18-email-verification.mdc](../.cursor/rules/18-email-verification.mdc)

**TR:** E-posta gönderimi, geçici kutu ve 6 haneli / link doğrulama **bizim kodumuz**. AgentMail, Resend-MCP, harici “inbox API” yok.

**EN:** Mail send, temp inbox, and 6-digit / link verification are **first-party**. No AgentMail, Resend-MCP, or third-party inbox API.

## Akış / Flow

```mermaid
sequenceDiagram
  participant U as User
  participant App as Electron
  participant API as GoGraphQL
  participant Mailer as Icerde_mailer
  participant Box as tempMailboxes
  participant SMTP as SMTP_or_dev_sink
  participant DB as verificationCodes
  U->>App: login_or_new_device
  App->>API: request_code
  API->>DB: store_codeHash_TTL
  API->>Box: insert_message_redacted_later
  API->>Mailer: send
  Mailer->>SMTP: message
  alt production
    SMTP-->>U: account_email
  else development
    SMTP-->>Box: same_row_visible_in_app
  end
  U->>App: six_digits_or_link
  App->>API: verify_code
  API->>DB: hash_compare_attempts
```

## Ne yazıyoruz / What we own

| Parça | Nerede | Not |
| --- | --- | --- |
| Mailer | `services/api` Go | SMTP env (`SMTP_HOST`…). Kütüphane (net/smtp) taşıma; ürün mantığı bizde. |
| Şablon | düz metin + basit HTML | İçerde wordmark, kod 6 hane mono, TTL metni. |
| `verificationCodes` | Mongo | `codeHash`, `purpose`, `expiresAt`, `attempts`, `userId`, `deviceId` |
| `tempMailboxes` | Mongo | Kullanıcının in-app kutusu (kod / şüpheli giriş). Gövde PII; silmede drop. |
| UI | Next.js | 6 kutu + “e-postadaki link”; güvenlik ekranında kutu listesi |
| Dev sink | mailer | SMTP yoksa kutu + log; sahte “gönderildi” yok. |

## Kod kuralları / Code rules

- 6 hane: `crypto/rand`, `000000–999999`, **yalnızca hash** saklanır (pepper + SHA-256). Düz kod Mongo’da yok.
- TTL ~10 dakika. Max 5 deneme sonra kilit; yeni kod eskiyi geçersiz kılar.
- Amaçlar (`purpose`): `new_device`, `login_challenge`, `email_verify`, `suspicious_login`. Switch exhaustive.
- Link: imzalı token (aynı doğrulama kaydı); UI hâlâ 6 haneyi de kabul eder.
- Rate limit: kullanıcı + IP. SMS yok.
- GDPR: kod ve kutu gövdesi LLM’e gitmez. `deleteMe` kod + kutu siler.

## Yasak / Forbidden

- AgentMail, harici geçici-posta SaaS, SMS, telefon.
- Cursor MCP ile mail göndermek.
- Dev’de kodu GraphQL error mesajında döndürmek (yalnızca `tempMailboxes` + sunucu log, asla production response).
