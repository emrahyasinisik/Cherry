"use client";

import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";

import { AppShell } from "@/components/app-shell";
import { Button } from "@/components/ui/button";
import { getToken, graphql, setToken, type User } from "@/lib/api";

export default function PrivacyPage() {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [bundle, setBundle] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [info, setInfo] = useState<string | null>(null);

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
        })
        .catch((err: unknown) => {
          setError(err instanceof Error ? err.message : "Yüklenemedi.");
        });
    }, 0);
    return () => {
      window.clearTimeout(timer);
    };
  }, [router]);

  if (error) {
    return (
      <div className="flex min-h-full items-center justify-center text-destructive">{error}</div>
    );
  }
  if (!user) {
    return (
      <div className="flex min-h-full items-center justify-center text-muted-foreground">
        Yükleniyor…
      </div>
    );
  }

  return (
    <AppShell user={user} title="Gizlilik" status="KVKK / GDPR">
      <div className="mx-auto flex max-w-lg flex-col gap-5 p-8">
        <h1 className="text-base font-medium">KVKK / GDPR</h1>
        <p className="text-muted-foreground">
          LLM’e giden her çağrı redakte edilir. Dışa aktarma yalnızca senin platform verini verir;
          başka kiracı yok.
        </p>
        {info ? <p className="text-muted-foreground">{info}</p> : null}
        <Button
          type="button"
          variant="outline"
          onClick={() => {
            void graphql<{ exportMe: { json: string } }>(`query Ex { exportMe { json } }`)
              .then((data) => setBundle(data.exportMe.json))
              .catch((err: unknown) => {
                setError(err instanceof Error ? err.message : "Dışa aktarılamadı.");
              });
          }}
        >
          Verilerimi dışa aktar
        </Button>
        {bundle ? (
          <pre className="overflow-auto rounded-[10px] border border-border p-3 font-mono text-[11px] whitespace-pre-wrap">
            {bundle}
          </pre>
        ) : null}
        <Button
          type="button"
          variant="destructive"
          onClick={() => {
            if (!window.confirm("Hesap ve LLM denetim kayıtları silinsin mi?")) {
              return;
            }
            void graphql<{ deleteMe: boolean }>(
              `mutation Del($wipe: Boolean!) { deleteMe(wipeProjects: $wipe) }`,
              { wipe: true },
            )
              .then(() => {
                setToken(null);
                router.replace("/");
              })
              .catch((err: unknown) => {
                setError(err instanceof Error ? err.message : "Silinemedi.");
              });
          }}
        >
          Hesabı sil
        </Button>
        <p className="text-[11px] text-muted-foreground">
          Silme: platform PII, kutu, oturum, denetim. Proje klasörleri de silinir (wipe).
        </p>
      </div>
    </AppShell>
  );
}
