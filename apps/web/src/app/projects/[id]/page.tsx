"use client";

import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { useEffect, useRef, useState } from "react";

import { AppShell } from "@/components/app-shell";
import { Button } from "@/components/ui/button";
import {
  PROJECT_FIELDS,
  getToken,
  graphql,
  projectStatusLabel,
  stackLabel,
  type Project,
  type User,
} from "@/lib/api";
import { setLastProjectId } from "@/lib/last-project";

export default function StudioPage() {
  const params = useParams<{ id: string }>();
  const router = useRouter();
  const id = params.id;
  const [user, setUser] = useState<User | null>(null);
  const [project, setProject] = useState<Project | null>(null);
  const [error, setError] = useState<string | null>(null);
  const openedTest = useRef(false);

  useEffect(() => {
    if (!getToken()) {
      router.replace("/");
      return;
    }
    setLastProjectId(id);
    let cancelled = false;
    async function tick() {
      try {
        const data = await graphql<{ me: User | null; project: Project | null }>(
          `query Studio($id: ID!) {
            me { id email workspaceKind totpEnabled }
            project(id: $id) { ${PROJECT_FIELDS} }
          }`,
          { id },
        );
        if (cancelled) {
          return;
        }
        if (!data.me) {
          router.replace("/");
          return;
        }
        setUser(data.me);
        if (!data.project) {
          setError("Proje bulunamadı.");
          return;
        }
        setProject(data.project);
        const watch = sessionStorage.getItem(`icerde.watch.${id}`) === "1";
        if (
          watch &&
          !openedTest.current &&
          (data.project.status === "TESTING" || data.project.status === "READY")
        ) {
          openedTest.current = true;
          sessionStorage.removeItem(`icerde.watch.${id}`);
          router.push(`/projects/${id}/maestro`);
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : "Yüklenemedi.");
        }
      }
    }
    void tick();
    const timer = window.setInterval(() => {
      void tick();
    }, 700);
    return () => {
      cancelled = true;
      window.clearInterval(timer);
    };
  }, [id, router]);

  if (error) {
    return (
      <div className="flex min-h-full items-center justify-center text-destructive">{error}</div>
    );
  }
  if (!user || !project) {
    return (
      <div className="flex min-h-full items-center justify-center text-muted-foreground">
        Yükleniyor…
      </div>
    );
  }

  const working = project.status === "QUEUED" || project.status === "WRITING";

  return (
    <AppShell
      user={user}
      title={project.name}
      status={projectStatusLabel(project.status)}
    >
      <div className="icerde-enter grid min-h-full gap-px bg-border lg:grid-cols-[minmax(240px,1fr)_minmax(320px,1.4fr)_minmax(240px,1fr)]">
        <section className="flex flex-col gap-4 bg-background p-6">
          <h1 className="text-base font-medium">{project.name}</h1>
          <p className="text-muted-foreground">{project.brief}</p>
          <p className="font-mono text-[11px] text-muted-foreground">
            {stackLabel(project.stack)} · {projectStatusLabel(project.status)}
          </p>
          <p className="break-all font-mono text-[11px] text-muted-foreground">{project.rootPath}</p>
          <Button
            type="button"
            variant="outline"
            onClick={() => router.push(`/projects/${id}/maestro`)}
          >
            Maestro ekranını aç
          </Button>
          {project.status === "READY" ? (
            <Button
              type="button"
              onClick={() => {
                void downloadZip(id);
              }}
            >
              Zip indir
            </Button>
          ) : null}
          <ul className="font-mono text-[11px] text-muted-foreground">
            {project.files.length === 0 ? (
              <li>{working ? "Dosyalar yazılıyor…" : "Dosya yok"}</li>
            ) : (
              project.files.slice(0, 12).map((file) => (
                <li key={file.path}>{file.path}</li>
              ))
            )}
          </ul>
        </section>
        <section className="flex min-h-[280px] flex-col bg-background p-6">
          <h2 className="mb-3 text-sm font-medium">Ajan günlüğü</h2>
          <pre className="min-h-0 flex-1 overflow-auto rounded-[10px] border border-border bg-card/40 p-3 font-mono text-[12px] leading-[18px] whitespace-pre-wrap">
            {project.logs.length === 0
              ? "Bekleniyor…"
              : project.logs.map((line) => line.message).join("\n")}
          </pre>
        </section>
        <section className="flex flex-col gap-3 bg-background p-6">
          <h2 className="text-sm font-medium">Maestro</h2>
          {project.maestro.ready ? (
            <>
              <p className="text-muted-foreground">
                Test aşamasına gelindi. Tasarım ve YAML akışları hazır. Cihaz yok — sonuç SKIPPED.
              </p>
              <Button type="button" onClick={() => router.push(`/projects/${id}/maestro?from=test`)}>
                Tasarımı ve testleri göster
              </Button>
            </>
          ) : (
            <p className="text-muted-foreground">
              Ajan yazmayı bitirince bu sütun ve Maestro ekranı açılır. İstediğin anda da açabilirsin.
            </p>
          )}
          <Link href="/projects" className="text-[13px] text-muted-foreground underline-offset-4 hover:underline">
            Projelere dön
          </Link>
        </section>
      </div>
    </AppShell>
  );
}

async function downloadZip(id: string) {
  const token = getToken();
  const response = await fetch(`/export/${id}`, {
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  });
  if (!response.ok) {
    return;
  }
  const blob = await response.blob();
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = "icerde.zip";
  anchor.click();
  URL.revokeObjectURL(url);
}
