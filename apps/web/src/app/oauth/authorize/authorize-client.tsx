"use client";

import { useRouter, useSearchParams } from "next/navigation";
import { useEffect, useState } from "react";

import { IcerdeMark, ProviderMark } from "@/components/provider-mark";
import { connectionKindLabel, getToken, graphql, type User } from "@/lib/api";
import {
  consentTheme,
  isConnectionKind,
  markClass,
  oauthHeadline,
  oauthPermissions,
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
      <div className="flex min-h-full items-center justify-center bg-[#0d1117] px-6 text-center text-[#e6edf3]">
        OAuth isteği eksik. Bağlantılar’dan yeniden dene.
      </div>
    );
  }

  if (loadError) {
    return (
      <div className="flex min-h-full items-center justify-center bg-[#0d1117] text-[#f85149]">{loadError}</div>
    );
  }

  if (!user) {
    return (
      <div className="flex min-h-full items-center justify-center bg-[#0d1117] text-[#8b949e]">Yükleniyor…</div>
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
    <div className="flex min-h-full items-center justify-center px-4 py-10" style={{ background: theme.bg, color: theme.text }}>
      <div className="w-full max-w-[480px]">
        <p className="mb-6 text-center font-mono text-[11px] tracking-wide uppercase" style={{ color: theme.muted }}>
          {theme.host} · OAuth 2.0 · Authorization Code
        </p>
        <div className="rounded-xl border p-6 shadow-2xl" style={{ background: theme.card, borderColor: theme.border }}>
          <div className="mb-6 flex items-center justify-center gap-4">
            <IcerdeMark className="size-10" />
            <span className="text-2xl" style={{ color: theme.muted }} aria-hidden>
              →
            </span>
            <ProviderMark kind={kind} className={`size-10 ${markClass(kind)}`} />
          </div>
          <h1 className="text-center text-lg font-semibold leading-snug">{oauthHeadline(kind)}</h1>
          <p className="mt-2 text-center text-sm" style={{ color: theme.muted }}>
            {user.email} olarak İçerde’den geliyorsun. Onay, {name} hesabındaki izinleri İçerde’ye verir. İçerde
            uygulamanı barındırmaz.
          </p>

          <label className="mt-5 block text-[12px] font-medium" htmlFor="oauth-account">
            {name} hesabı
          </label>
          <input
            id="oauth-account"
            value={account}
            autoComplete="off"
            onChange={(event) => setAccount(event.target.value)}
            className="mt-1 h-10 w-full rounded-md border bg-transparent px-3 text-sm outline-none"
            style={{ borderColor: theme.border, color: theme.text }}
          />
          {error ? (
            <p className="mt-2 text-sm text-[#f85149]" role="alert">
              {error}
            </p>
          ) : null}

          <p className="mt-5 text-[12px] font-medium">Bu uygulama şunları yapabilecek</p>
          <ul className="mt-2 flex flex-col gap-2 text-sm" style={{ color: theme.muted }}>
            {oauthPermissions(kind).map((item) => (
              <li key={item} className="flex gap-2">
                <span aria-hidden>✓</span>
                <span>{item}</span>
              </li>
            ))}
          </ul>

          <div className="mt-6 flex flex-col gap-2 sm:flex-row-reverse">
            <button
              type="button"
              className="h-10 flex-1 rounded-md text-sm font-semibold"
              style={{ background: theme.authorizeBg, color: theme.authorizeFg }}
              onClick={() => {
                if (account.trim().length < 2) {
                  setError("Hesap adı gerekli.");
                  return;
                }
                go("allow");
              }}
            >
              İçerde’yi yetkilendir
            </button>
            <button
              type="button"
              className="h-10 flex-1 rounded-md border text-sm"
              style={{ borderColor: theme.border, color: theme.text, background: "transparent" }}
              onClick={() => go("deny")}
            >
              İptal
            </button>
          </div>
        </div>
        <p className="mt-4 text-center text-[11px]" style={{ color: theme.muted }}>
          İzin vermezsen bağlantı kurulmaz. Client id yoksa bu ekran yerel OAuth izin ekranıdır; GitHub/Vercel client
          id tanımlandıysa gerçek siteye gidilir.
        </p>
      </div>
    </div>
  );
}
