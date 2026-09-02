# Güvenlik / Security — X-inspired

**Kural:** [.cursor/rules/06-security.mdc](../.cursor/rules/06-security.mdc) · posta: [email-verification.md](email-verification.md)

**TR:** X’in görünen akışı. SMS yok, telefon kimliği yok, tek oturum. E-posta ve 6 hane **birinci parti**.

**EN:** X’s visible flow. No SMS, no phone identity, one session. Mail and 6-digit codes are **first-party**.

## Giriş sırası / Login order

```mermaid
sequenceDiagram
  participant U as User
  participant App as Electron
  participant API as GoGraphQL
  participant Mail as Cherry_mailer
  participant DB as MongoDB
  U->>App: email_plus_password
  App->>API: login
  API->>DB: verify_password
  alt trusted_device_and_session_ok
    API->>DB: revoke_other_sessions
    API-->>App: session
  else new_or_untrusted_device
    API->>Mail: six_digit_code
    Mail-->>U: code
    U->>App: enter_code
    App->>API: verify_code
  end
  alt MFA_enabled
    U->>App: TOTP
    App->>API: verify_totp
  end
  API->>DB: revoke_other_sessions
  API-->>App: single_active_session
```

## Cihaz ve oturum / Device and session

```mermaid
stateDiagram-v2
  [*] --> unknownDevice
  unknownDevice --> pendingCode: six_digit_sent
  pendingCode --> trusted: code_ok_user_trusts
  pendingCode --> unknownDevice: expire_or_lock
  trusted --> active: totp_if_needed
  active --> revoked: new_login_or_user_revoke
  revoked --> unknownDevice: login_again
```

## Alınan / Taken from X

- Password
- 6-digit challenge on new device
- TOTP authenticator
- Trusted devices
- Session list + revoke
- Suspicious login notice

## Alınmayan / Not taken

- SMS, phone number
- Many concurrent sessions (we keep one)
- Blocking temporary-email *providers* as a product feature (we run our own inbox for codes)
- AgentMail or any third-party inbox API
- Passkeys (later)

Codes: ~10 minute TTL, hashed, attempt lock. Details: [email-verification.md](email-verification.md).
