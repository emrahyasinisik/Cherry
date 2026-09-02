"use client";

import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";

import { AppShell } from "@/components/app-shell";
import { Button } from "@/components/ui/button";
import { getToken, graphql, type User } from "@/lib/api";
import { getLastProjectId } from "@/lib/last-project";

export default function MaestroIndexPage() {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);

  useEffect(() => {
    if (!getToken()) {
      router.replace("/");
      return;
    }
    const last = getLastProjectId();
    if (last) {
      router.replace(`/projects/${last}/maestro`);
      return;
    }
    const timer = window.setTimeout(() => {
      void graphql<{ me: User | null }>(`query Me { me { id email workspaceKind totpEnabled } }`).then(
        (data) => {
          if (!data.me) {
            router.replace("/");
            return;
          }
          setUser(data.me);
        },
      );
    }, 0);
    return () => {
      window.clearTimeout(timer);
    };
  }, [router]);

  if (!user) {
    return (
      <div className="flex min-h-full items-center justify-center text-muted-foreground">
        Yükleniyor…
      </div>
    );
  }

  return (
    <AppShell user={user} title="Maestro">
      <div className="cherry-enter mx-auto flex max-w-lg flex-col gap-4 p-8">
        <h1 className="text-base font-medium">Maestro</h1>
        <p className="text-muted-foreground">
          Önce bir uygulama tarif et. Test aşamasına gelince bu ekran kendiliğinden açılır; sen de
          her an yan menüden isteyebilirsin.
        </p>
        <Button type="button" onClick={() => router.push("/projects/new")}>
          Yeni proje
        </Button>
      </div>
    </AppShell>
  );
}
