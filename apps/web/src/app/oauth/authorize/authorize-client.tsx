"use client";

import { Check, Lock } from "lucide-react";
import { useRouter, useSearchParams } from "next/navigation";
import { useEffect, useState } from "react";

import { CherryMark, ProviderMark } from "@/components/provider-mark";
import { connectionKindLabel, getToken, graphql, type User } from "@/lib/api";
import {
  consentTheme,
  isConnectionKind,
  markClass,
  oauthHeadline,
  oauthPermissions,
  tileClass,
} from "@/lib/oauth";

export function AuthorizeClient() {
  const router = useRouter();
  const params = useSearchParams();
  const kindRaw = params.get("kind");
  const state = params.get("state") ?? "";
  const accountError = params.get("error") === "account";
  const [user, setUser] = useState<User | null>(null);
  const [account, setAccount] = useState("");
  const [error, setError] = useState<string | null>(accountError ? "Hesap adı gerekli." : null);
  const [loadError, setLoadError] = useState<string | null>(null);

  useEffect(() => {
    if (!getToken()) {
      router.replace("/");
      return;
    }
    const timer = window.setTimeout(() => {
      void graphql<{ me: User | null }>(`query Me { me { id email workspaceKind totpEnabled } }`)
        .then((data) => {
          if (!data.me) {
            router.replace("/");
            return;
          }
          setUser(data.me);
          const local = data.me.email.split("@")[0] ?? data.me.email;
          setAccount((prev) => prev || local);
        })
        .catch((err: unknown) => {
          setLoadError(err instanceof Error ? err.message : "Yüklenemedi.");
        });
    }, 0);
    return () => {
      window.clearTimeout(timer);
    };
  }, [router]);

  if (!isConnectionKind(kindRaw) || !state) {
    return (
      <div className="flex min-h-full items-center justify-center bg-background px-6 text-center text-muted-foreground">
        OAuth isteği eksik. Bağlantılar’dan yeniden dene.
      </div>
    );
  }

  if (loadError) {
    return (
      <div className="flex min-h-full items-center justify-center bg-background text-destructive">{loadError}</div>
    );
  }

  if (!user) {
    return (
      <div className="flex min-h-full items-center justify-center bg-background text-muted-foreground">
        Yükleniyor…
      </div>
    );
  }

  const kind = kindRaw;
  const theme = consentTheme(kind);
  const name = connectionKindLabel(kind);
  const decisionBase = "/oauth/decision";

  function go(decision: "allow" | "deny") {
    const query = new URLSearchParams({
      state,
      decision,
      account: account.trim(),
    });
    window.location.assign(`${decisionBase}?${query.toString()}`);
  }

  return (
    <div className="flex min-h-full flex-col" style={{ background: theme.bg, color: theme.text }}>
      <header
        className="flex h-12 shrink-0 items-center justify-between border-b px-4"
        style={{ borderColor: theme.border }}
      >
        <span className="flex items-center gap-2 text-[13px]">
          <ProviderMark kind={kind} className={`size-4 ${markClass(kind)}`} />
          {theme.host}
        </span>
        <span className="text-[12px]" style={{ color: theme.muted }}>
          OAuth 2.0
        </span>
      </header>

      <div className="flex flex-1 items-center justify-center px-4 py-10">
        <div className="cherry-enter w-full max-w-[440px]">
          <div
            className="rounded-[10px] border p-6"
            style={{ background: theme.card, borderColor: theme.border }}
          >
            <div className="mb-5 flex items-center justify-center gap-3">
              <span
                className="flex size-12 items-center justify-center rounded-[8px] border"
                style={{ borderColor: theme.border, background: theme.bg }}
              >
                <CherryMark className="size-7" />
              </span>
              <span className="h-px w-8" style={{ background: theme.border }} aria-hidden />
              <span
                className={`flex size-12 items-center justify-center rounded-[8px] border ${tileClass(kind)}`}
              >
                <ProviderMark kind={kind} className="size-6" />
              </span>
            </div>

            <h1 className="text-center text-base leading-6 font-medium">{oauthHeadline(kind)}</h1>
            <p className="mt-2 text-center text-[13px] leading-5" style={{ color: theme.muted }}>
              {user.email} olarak devam ediyorsun. Onay, {name} izinlerini Cherry’ye verir. Cherry uygulamanı
              barındırmaz.
            </p>

            <label className="mt-5 block text-[12px] font-medium" htmlFor="oauth-account">
              {name} hesabı
            </label>
            <input
              id="oauth-account"
              value={account}
              autoComplete="off"
              onChange={(event) => setAccount(event.target.value)}
              className="mt-1 h-9 w-full rounded-[8px] border bg-transparent px-3 text-[13px] outline-none"
              style={{ borderColor: theme.border, color: theme.text }}
            />
            {error ? (
              <p className="mt-2 text-[13px] text-[#c45c4a]" role="alert">
                {error}
              </p>
            ) : null}

            <p className="mt-5 text-[12px] font-medium">Bu uygulama şunları yapabilecek</p>
            <ul
              className="mt-2 divide-y rounded-[8px] border"
              style={{ borderColor: theme.border, color: theme.muted }}
            >
              {oauthPermissions(kind).map((item) => (
                <li key={item} className="flex items-start gap-2 px-3 py-2 text-[13px] leading-5">
                  <Check className="mt-0.5 size-4 shrink-0" strokeWidth={1.5} aria-hidden />
                  <span>{item}</span>
                </li>
              ))}
            </ul>

            <div className="mt-5 flex flex-col-reverse gap-2 sm:flex-row">
              <button
                type="button"
                className="h-9 flex-1 rounded-[8px] border text-[13px]"
                style={{ borderColor: theme.border, color: theme.text, background: "transparent" }}
                onClick={() => go("deny")}
              >
                İptal
              </button>
              <button
                type="button"
                className="h-9 flex-1 rounded-[8px] text-[13px] font-medium"
                style={{ background: theme.authorizeBg, color: theme.authorizeFg }}
                onClick={() => {
                  if (account.trim().length < 2) {
                    setError("Hesap adı gerekli.");
                    return;
                  }
                  go("allow");
                }}
              >
                Cherry’yi yetkilendir
              </button>
            </div>
          </div>
          <p className="mt-4 flex items-center justify-center gap-1.5 text-center text-[11px]" style={{ color: theme.muted }}>
            <Lock className="size-3" strokeWidth={1.5} aria-hidden />
            İzin vermezsen bağlantı kurulmaz. Client id varsa gerçek {theme.host} açılır.
          </p>
        </div>
      </div>
    </div>
  );
}
