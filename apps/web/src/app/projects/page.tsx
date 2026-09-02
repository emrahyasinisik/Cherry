"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";

import { AppShell } from "@/components/app-shell";
import { Button } from "@/components/ui/button";
import {
  getToken,
  graphql,
  projectStatusLabel,
  stackLabel,
  type ProjectStatus,
  type ProjectStack,
  type User,
} from "@/lib/api";

export default function ProjectsPage() {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [projects, setProjects] = useState<
    {
      id: string;
      name: string;
      brief: string;
      stack: ProjectStack;
      status: ProjectStatus;
      createdAt: string;
    }[] | null
  >(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    async function load() {
      if (!getToken()) {
        router.replace("/");
        return;
      }
      try {
        const data = await graphql<{
          me: User | null;
          projects: {
            id: string;
            name: string;
            brief: string;
            stack: ProjectStack;
            status: ProjectStatus;
            createdAt: string;
          }[];
        }>(
          `query ProjectsHome {
            me { id email workspaceKind totpEnabled }
            projects { id name brief stack status createdAt }
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
        <div className="flex items-end justify-between gap-4">
          <div>
            <h1 className="text-base font-medium">Projeler</h1>
            <p className="text-muted-foreground">
              Uygulamayı tarif et; ajan arka planda yazar. Test aşamasında Maestro açılır.
            </p>
          </div>
          <Button type="button" onClick={() => router.push("/projects/new")}>
            Yeni proje
          </Button>
        </div>
        {projects.length === 0 ? (
          <div className="cherry-enter cherry-surface rounded-[10px] border border-dashed border-border px-6 py-16 text-center">
            <p className="text-base font-medium">Henüz proje yok</p>
            <p className="mt-1 text-muted-foreground">
              İstediğin mobil uygulamayı anlat. Kodlama ve test arka planda yürür.
            </p>
          </div>
        ) : (
          <ul className="cherry-stagger divide-y divide-border overflow-hidden rounded-[10px] border border-border">
            {projects.map((project) => (
              <li key={project.id}>
                <Link
                  href={`/projects/${project.id}`}
                  className="cherry-row flex items-center justify-between gap-3 px-4 py-3 hover:bg-muted/40"
                >
                  <span>
                    <span className="block">{project.name}</span>
                    <span className="font-mono text-[11px] text-muted-foreground">
                      {stackLabel(project.stack)} · {projectStatusLabel(project.status)}
                    </span>
                  </span>
                </Link>
              </li>
            ))}
          </ul>
        )}
      </div>
    </AppShell>
  );
}
