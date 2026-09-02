"use client";

import { useParams, useRouter } from "next/navigation";
import { useEffect, useState } from "react";

import { AppShell } from "@/components/app-shell";
import { Button } from "@/components/ui/button";
import {
  PROJECT_FIELDS,
  activateStatusLabel,
  getToken,
  graphql,
  maestroResultLabel,
  projectStatusLabel,
  type MaestroFlow,
  type Project,
  type User,
} from "@/lib/api";
import { setLastProjectId } from "@/lib/last-project";
import { cn } from "@/lib/utils";

export default function MaestroPage() {
  const params = useParams<{ id: string }>();
  const router = useRouter();
  const id = params.id;
  const [user, setUser] = useState<User | null>(null);
  const [project, setProject] = useState<Project | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [screenId, setScreenId] = useState<string | null>(null);
  const [flowId, setFlowId] = useState<string | null>(null);
  const [running, setRunning] = useState(false);
  const [runError, setRunError] = useState<string | null>(null);

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
          `query Maestro($id: ID!) {
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
        setScreenId((current) => current ?? data.project?.maestro.screens[0]?.id ?? null);
        setFlowId((current) => current ?? data.project?.maestro.flows[0]?.id ?? null);
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : "Yüklenemedi.");
        }
      }
    }
    void tick();
    const timer = window.setInterval(() => {
      void tick();
    }, 900);
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

  const screen = project.maestro.screens.find((item) => item.id === screenId) ?? project.maestro.screens[0];
  const flow = project.maestro.flows.find((item) => item.id === flowId) ?? project.maestro.flows[0];
  const ready = project.maestro.ready;

  async function handleRunMaestro() {
    setRunning(true);
    setRunError(null);
    try {
      const data = await graphql<{ runMaestro: Project }>(
        `mutation RunMaestro($id: ID!) {
          runMaestro(projectId: $id) { ${PROJECT_FIELDS} }
        }`,
        { id },
      );
      setProject(data.runMaestro);
    } catch (err) {
      setRunError(err instanceof Error ? err.message : "Maestro koşulamadı.");
    } finally {
      setRunning(false);
    }
  }

  return (
    <AppShell user={user} title={`${project.name} · Maestro`} status={projectStatusLabel(project.status)}>
      <div className="flex min-h-full flex-col gap-4 p-6 lg:flex-row">
        <div className="flex min-w-0 flex-1 flex-col gap-3">
          <div className="flex flex-wrap items-center justify-between gap-2">
            <div>
              <h1 className="text-base font-medium">Maestro</h1>
              <p className="text-muted-foreground">
                {project.status === "TESTING" || project.status === "READY"
                  ? "HTML maket stüdyo içindir, zip’e girmez. Asıl uygulama seçilen dil. Cihaz yoksa akış SKIPPED — sahte geçiş yok."
                  : "Ajan yazınca tasarım ve testler burada durur."}
              </p>
              {project.activate.url ? (
                <p className="mt-1 font-mono text-[11px] text-muted-foreground">
                  Yerel API {project.activate.url} · {activateStatusLabel(project.activate.status)}
                </p>
              ) : (
                <p className="mt-1 text-muted-foreground">
                  Yerel API kapalı. Stüdyodan başlat; Maestro Cherry GraphQL’e değil localhost’a konuşur.
                </p>
              )}
            </div>
            <div className="flex gap-2">
              <Button type="button" variant="outline" disabled={running} onClick={() => void handleRunMaestro()}>
                {running ? "Koşuyor…" : "Maestro’yu koş"}
              </Button>
              <Button type="button" variant="outline" onClick={() => router.push(`/projects/${id}`)}>
                Ajan günlüğüne dön
              </Button>
            </div>
          </div>
          {!ready ? (
            <div className="rounded-[10px] border border-dashed border-border px-4 py-10 text-center text-muted-foreground">
              Ajan hâlâ yazıyor. Ekranlar düşünce burada görünür.
            </div>
          ) : (
            <>
              <div className="flex gap-1">
                {project.maestro.screens.map((item) => (
                  <button
                    key={item.id}
                    type="button"
                    onClick={() => setScreenId(item.id)}
                    className={cn(
                      "rounded-md border px-2.5 py-1 text-[13px]",
                      screen?.id === item.id
                        ? "border-primary bg-primary/10"
                        : "border-border text-muted-foreground",
                    )}
                  >
                    {item.name}
                  </button>
                ))}
              </div>
              {screen ? (
                <div className="flex justify-center">
                  <div className="h-[560px] w-[280px] overflow-hidden rounded-[28px] border border-border bg-black shadow-none">
                    <iframe
                      title={screen.name}
                      srcDoc={screen.html}
                      sandbox="allow-same-origin"
                      className="h-full w-full border-0 bg-[#0E1114]"
                    />
                  </div>
                </div>
              ) : null}
            </>
          )}
        </div>
        <aside className="flex w-full flex-col gap-3 lg:w-[360px]">
          <h2 className="text-sm font-medium">Akışlar</h2>
          {project.maestro.flows.length === 0 ? (
            <p className="text-muted-foreground">YAML henüz yok.</p>
          ) : (
            <ul className="flex flex-col gap-2">
              {project.maestro.flows.map((item) => (
                <li key={item.id}>
                  <button
                    type="button"
                    onClick={() => setFlowId(item.id)}
                    className={cn(
                      "w-full rounded-[10px] border px-3 py-2 text-left",
                      flow?.id === item.id ? "border-primary" : "border-border",
                    )}
                  >
                    <span className="block">{item.name}</span>
                    <span className="font-mono text-[11px] text-muted-foreground">
                      {maestroResultLabel(item.result)}
                    </span>
                  </button>
                </li>
              ))}
            </ul>
          )}
          {runError ? <p className="text-destructive">{runError}</p> : null}
          {flow ? <FlowDetail flow={flow} /> : null}
        </aside>
      </div>
    </AppShell>
  );
}

function FlowDetail({ flow }: { flow: MaestroFlow }) {
  return (
    <div className="flex flex-col gap-2">
      <p className="text-muted-foreground">{flow.note}</p>
      <pre className="overflow-auto rounded-[10px] border border-border bg-card/40 p-3 font-mono text-[11px] leading-4 whitespace-pre-wrap">
        {flow.yaml}
      </pre>
    </div>
  );
}
