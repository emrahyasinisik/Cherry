"use client";

import { useRouter, useSearchParams } from "next/navigation";
import { useEffect, useState } from "react";

import { IcerdeMark } from "@/components/provider-mark";
import { getToken, graphql, type Connection } from "@/lib/api";

export function CallbackClient() {
  const router = useRouter();
  const params = useSearchParams();
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!getToken()) {
      router.replace("/");
      return;
    }
    const code = params.get("code") ?? "";
    const state = params.get("state") ?? "";
    if (!code || !state) {
      router.replace("/connections?oauth=error");
      return;
    }
    const timer = window.setTimeout(() => {
      void graphql<{ completeConnectionOAuth: Connection }>(
        `mutation Complete($code: String!, $state: String!) {
          completeConnectionOAuth(code: $code, state: $state) {
            kind status account tokenHint authMethod
          }
        }`,
        { code, state },
      )
        .then((data) => {
          router.replace(`/connections?oauth=connected&kind=${data.completeConnectionOAuth.kind}`);
        })
        .catch((err: unknown) => {
          setError(err instanceof Error ? err.message : "OAuth tamamlanamadı.");
        });
    }, 0);
    return () => {
      window.clearTimeout(timer);
    };
  }, [params, router]);

  if (error) {
    return (
      <div className="flex min-h-full flex-col items-center justify-center gap-3 px-6 text-center">
        <p className="text-destructive" role="alert">
          {error}
        </p>
        <button type="button" className="text-[13px] text-primary" onClick={() => router.replace("/connections")}>
          Bağlantılar’a dön
        </button>
      </div>
    );
  }

  return (
    <div className="flex min-h-full flex-col items-center justify-center gap-3">
      <IcerdeMark className="size-8" />
      <p className="text-muted-foreground">OAuth 2.0 kodu değiş tokuş…</p>
    </div>
  );
}
