"use client";

import { useRouter, useSearchParams } from "next/navigation";
import { useEffect, useState, type FormEvent } from "react";

import { CodeInputs } from "@/components/code-inputs";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  graphql,
  setToken,
  type EmailChannel,
  type LoginResult,
  type MailMessage,
} from "@/lib/api";
import { getDeviceFingerprint, getDeviceLabel } from "@/lib/device";

const LOGIN_RESULT = `
  next
  token
  challengeId
  emailSent
  emailChannel
  user { id email workspaceKind totpEnabled }
`;

type Step = "credentials" | "code" | "totp";

export function LoginForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [step, setStep] = useState<Step>("credentials");
  const [mode, setMode] = useState<"login" | "register">("login");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [code, setCode] = useState("");
  const [totp, setTotp] = useState("");
  const [trustDevice, setTrustDevice] = useState(true);
  const [challengeId, setChallengeId] = useState<string | null>(null);
  const [mailbox, setMailbox] = useState<MailMessage | null>(null);
  const [emailSent, setEmailSent] = useState(false);
  const [emailChannel, setEmailChannel] = useState<EmailChannel>("");
  const [error, setError] = useState<string | null>(null);
  const [pending, setPending] = useState(false);

  useEffect(() => {
    const link = searchParams.get("link");
    if (!link) {
      return;
    }
    let cancelled = false;
    async function run() {
      setPending(true);
      try {
        const data = await graphql<{ verifyLink: LoginResult }>(
          `mutation VerifyLink($token: String!, $deviceFingerprint: String!, $deviceLabel: String!) {
            verifyLink(token: $token, deviceFingerprint: $deviceFingerprint, deviceLabel: $deviceLabel) {
              ${LOGIN_RESULT}
            }
          }`,
          {
            token: link,
            deviceFingerprint: getDeviceFingerprint(),
            deviceLabel: getDeviceLabel(),
          },
        );
        if (!cancelled) {
          await applyResult(data.verifyLink);
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : "Link geçersiz.");
        }
      } finally {
        if (!cancelled) {
          setPending(false);
        }
      }
    }
    const timer = window.setTimeout(() => {
      void run();
    }, 0);
    return () => {
      cancelled = true;
      window.clearTimeout(timer);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps -- verify link once
  }, [searchParams, router]);

  async function applyResult(result: LoginResult) {
    switch (result.next) {
      case "SESSION":
        if (!result.token) {
          setError("Oturum token’ı gelmedi.");
          return;
        }
        setToken(result.token);
        router.push("/projects");
        return;
      case "DEVICE_CODE":
        if (!result.challengeId) {
          setError("Doğrulama başlatılamadı.");
          return;
        }
        setChallengeId(result.challengeId);
        setEmailSent(result.emailSent);
        setEmailChannel(normalizeEmailChannel(result.emailChannel));
        setStep("code");
        if (!result.emailSent) {
          await loadMailbox(result.challengeId);
        } else {
          setMailbox(null);
        }
        return;
      case "TOTP":
        if (!result.challengeId) {
          setError("TOTP başlatılamadı.");
          return;
        }
        setChallengeId(result.challengeId);
        setStep("totp");
        return;
      default: {
        const exhaustive: never = result.next;
        setError(String(exhaustive));
      }
    }
  }

  async function loadMailbox(id: string) {
    const data = await graphql<{ challengeMailbox: MailMessage | null }>(
      `query Box($challengeId: ID!) {
        challengeMailbox(challengeId: $challengeId) {
          id subject body purpose createdAt
        }
      }`,
      { challengeId: id },
    );
    setMailbox(data.challengeMailbox);
  }

  async function handleCredentials(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    if (!email.trim() || !password) {
      setError("E-posta ve şifre gerekli.");
      return;
    }
    if (mode === "register" && password.length < 8) {
      setError("Şifre en az 8 karakter.");
      return;
    }
    setPending(true);
    try {
      const mutation =
        mode === "register"
          ? `mutation Register($email: String!, $password: String!, $deviceFingerprint: String!, $deviceLabel: String!) {
              register(email: $email, password: $password, deviceFingerprint: $deviceFingerprint, deviceLabel: $deviceLabel) {
                ${LOGIN_RESULT}
              }
            }`
          : `mutation Login($email: String!, $password: String!, $deviceFingerprint: String!, $deviceLabel: String!) {
              login(email: $email, password: $password, deviceFingerprint: $deviceFingerprint, deviceLabel: $deviceLabel) {
                ${LOGIN_RESULT}
              }
            }`;
      const data = await graphql<Record<string, LoginResult>>(mutation, {
        email: email.trim(),
        password,
        deviceFingerprint: getDeviceFingerprint(),
        deviceLabel: getDeviceLabel(),
      });
      const result = mode === "register" ? data.register : data.login;
      await applyResult(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : "İşlem başarısız.");
    } finally {
      setPending(false);
    }
  }

  async function handleCode(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    if (!challengeId || code.replace(/\D/g, "").length !== 6) {
      setError("6 haneli kodu gir.");
      return;
    }
    setPending(true);
    try {
      const data = await graphql<{ verifyCode: LoginResult }>(
        `mutation VerifyCode($challengeId: ID!, $code: String!, $trustDevice: Boolean!) {
          verifyCode(challengeId: $challengeId, code: $code, trustDevice: $trustDevice) {
            ${LOGIN_RESULT}
          }
        }`,
        {
          challengeId,
          code: code.replace(/\D/g, ""),
          trustDevice,
        },
      );
      await applyResult(data.verifyCode);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Kod geçersiz.");
    } finally {
      setPending(false);
    }
  }

  async function handleTotp(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    if (!challengeId || totp.replace(/\D/g, "").length !== 6) {
      setError("Authenticator kodunu gir.");
      return;
    }
    setPending(true);
    try {
      const data = await graphql<{ verifyTotp: LoginResult }>(
        `mutation VerifyTotp($challengeId: ID!, $code: String!) {
          verifyTotp(challengeId: $challengeId, code: $code) {
            ${LOGIN_RESULT}
          }
        }`,
        { challengeId, code: totp.replace(/\D/g, "") },
      );
      await applyResult(data.verifyTotp);
    } catch (err) {
      setError(err instanceof Error ? err.message : "TOTP geçersiz.");
    } finally {
      setPending(false);
    }
  }

  if (step === "code") {
    return (
      <form onSubmit={(event) => void handleCode(event)} className="flex flex-col gap-4">
        <p className="text-muted-foreground">{codeStepCopy(emailSent, emailChannel)}</p>
        {emailSent ? (
          <p className="text-[13px] text-muted-foreground">
            Gelen kutusunu ve spam klasörünü kontrol et. Aynı kayıt için e-postadaki link de çalışır.
          </p>
        ) : (
          <p className="rounded-md border border-border bg-background/60 p-3 text-[13px] text-muted-foreground">
            Bu ortamda SMTP veya Resend yok — kod gerçek e-postaya gitmedi. Geliştirme kutusu aşağıda.
            Production için <code className="font-mono text-[11px]">SMTP_HOST</code> veya{" "}
            <code className="font-mono text-[11px]">RESEND_API_KEY</code> tanımla.
          </p>
        )}
        {!emailSent && mailbox ? (
          <pre className="overflow-auto rounded-md border border-border bg-background/60 p-3 font-mono text-[11px] leading-4 whitespace-pre-wrap">
            {mailbox.body}
          </pre>
        ) : null}
        <CodeInputs value={code} onChange={setCode} disabled={pending} />
        <label className="flex items-center gap-2 text-muted-foreground">
          <input
            type="checkbox"
            checked={trustDevice}
            onChange={(event) => setTrustDevice(event.target.checked)}
          />
          Bu cihaza güven
        </label>
        {error ? (
          <p className="text-destructive" role="alert">
            {error}
          </p>
        ) : null}
        <Button type="submit" size="lg" className="h-10 w-full" disabled={pending}>
          {pending ? "Doğrulanıyor…" : "Kodu doğrula"}
        </Button>
      </form>
    );
  }

  if (step === "totp") {
    return (
      <form onSubmit={(event) => void handleTotp(event)} className="flex flex-col gap-4">
        <p className="text-muted-foreground">Authenticator uygulamasındaki 6 hane.</p>
        <CodeInputs value={totp} onChange={setTotp} disabled={pending} />
        {error ? (
          <p className="text-destructive" role="alert">
            {error}
          </p>
        ) : null}
        <Button type="submit" size="lg" className="h-10 w-full" disabled={pending}>
          {pending ? "Doğrulanıyor…" : "TOTP doğrula"}
        </Button>
      </form>
    );
  }

  return (
    <form onSubmit={(event) => void handleCredentials(event)} className="flex flex-col gap-4">
      <div className="flex flex-col gap-1.5">
        <Label htmlFor="email">E-posta</Label>
        <Input
          id="email"
          type="email"
          autoComplete="username"
          value={email}
          onChange={(event) => setEmail(event.target.value)}
          className="icerde-focus h-9"
          placeholder="sen@icerde.dev"
          aria-invalid={Boolean(error)}
        />
      </div>
      <div className="flex flex-col gap-1.5">
        <Label htmlFor="password">Şifre</Label>
        <Input
          id="password"
          type="password"
          autoComplete={mode === "register" ? "new-password" : "current-password"}
          value={password}
          onChange={(event) => setPassword(event.target.value)}
          className="icerde-focus h-9"
          aria-invalid={Boolean(error)}
        />
      </div>
      {error ? (
        <p className="text-destructive" role="alert">
          {error}
        </p>
      ) : null}
      <Button type="submit" size="lg" className="h-10 w-full" disabled={pending}>
        {pending ? "Gönderiliyor…" : mode === "register" ? "Hesap oluştur" : "Giriş yap"}
      </Button>
      <button
        type="button"
        className="text-center text-muted-foreground underline-offset-4 hover:underline"
        onClick={() => {
          setMode(mode === "login" ? "register" : "login");
          setError(null);
        }}
      >
        {mode === "login" ? "Hesabın yok mu? Oluştur" : "Zaten hesabın var mı? Giriş yap"}
      </button>
    </form>
  );
}

function normalizeEmailChannel(channel: string): EmailChannel {
  switch (channel) {
    case "inbox":
    case "smtp":
    case "resend":
    case "":
      return channel;
    default:
      return "";
  }
}

function codeStepCopy(sent: boolean, channel: EmailChannel): string {
  if (!sent) {
    return "Yeni veya tanınmayan cihaz. 6 haneli kodu gir.";
  }
  switch (channel) {
    case "smtp":
    case "resend":
      return "Kod e-postana gönderildi (10 dakika, 5 deneme).";
    case "inbox":
    case "":
      return "Kod gönderildi. Gelen kutunu kontrol et.";
    default: {
      const exhaustive: never = channel;
      return exhaustive;
    }
  }
}
