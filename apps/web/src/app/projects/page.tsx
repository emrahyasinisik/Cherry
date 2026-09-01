"use client";

import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";

import { AppShell } from "@/components/app-shell";
import { Button } from "@/components/ui/button";
import { getToken, graphql, type Project, type User } from "@/lib/api";

export default function ProjectsPage() {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [projects, setProjects] = useState<Project[] | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [info, setInfo] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    async function load() {
      if (!getToken()) {
        router.replace("/");
        return;
      }
      try {
        const data = await graphql<{ me: User | null; projects: Project[] }>(
          `query ProjectsHome {
            me { id email workspaceKind }
            projects { id name stack status }
          }`,
        );
        if (cancelled) {
          return;
        }
        if (!data.me) {
          router.replace("/");
          return;
        }
        setUser(data.me);
        setProjects(data.projects);
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : "Yüklenemedi.");
        }
      }
    }
    void load();
    return () => {
      cancelled = true;
    };
  }, [router]);

  if (error) {
    return (
      <div className="flex min-h-full items-center justify-center text-destructive">
        {error}
      </div>
    );
  }

  if (!user || !projects) {
    return (
      <div className="flex min-h-full items-center justify-center text-muted-foreground">
        Yükleniyor…
      </div>
    );
  }

  return (
    <AppShell user={user}>
      <div className="mx-auto flex max-w-3xl flex-col gap-6 p-8">
        <div className="flex items-end justify-between">
          <div>
            <h1 className="text-base font-medium">Projeler</h1>
            <p className="text-muted-foreground">
              Üretilen mobil uygulamalar burada listelenir.
            </p>
          </div>
          <Button
            type="button"
            onClick={() => {
              setInfo("Proje oluşturma dilim 3’te (proje diski).");
            }}
          >
            Yeni proje
          </Button>
        </div>
        {info ? <p className="text-muted-foreground">{info}</p> : null}
        {projects.length === 0 ? (
          <div className="icerde-enter rounded-[10px] border border-dashed border-border bg-card/40 px-6 py-16 text-center">
            <p className="text-base font-medium">Henüz proje yok</p>
            <p className="mt-1 text-muted-foreground">
              İlk mobil uygulamayı ajan yazacak. Şimdilik kabuk hazır.
            </p>
          </div>
        ) : (
          <ul className="divide-y divide-border rounded-[10px] border border-border">
            {projects.map((project) => (
              <li
                key={project.id}
                className="flex items-center justify-between px-4 py-3"
              >
                <span>{project.name}</span>
                <span className="font-mono text-muted-foreground">
                  {project.stack}
                </span>
              </li>
            ))}
          </ul>
        )}
      </div>
    </AppShell>
  );
}
