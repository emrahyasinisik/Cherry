"use client";

import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { useEffect, useRef, useState, type FormEvent } from "react";

import { AppShell } from "@/components/app-shell";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  PROJECT_FIELDS,
  activateStatusLabel,
  getToken,
  graphql,
  maestroResultLabel,
  projectStatusLabel,
  stackLabel,
  stackSourceHint,
  backendTargetLabel,
  type ActivateStatus,
  type ChatRole,
  type JobLog,
  type Project,
  type User,
} from "@/lib/api";
import { setLastProjectId } from "@/lib/last-project";
import { cn } from "@/lib/utils";

export default function StudioPage() {
  const params = useParams<{ id: string }>();
  const router = useRouter();
  const id = params.id;
  const [user, setUser] = useState<User | null>(null);
  const [project, setProject] = useState<Project | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [sendError, setSendError] = useState<string | null>(null);
  const [draft, setDraft] = useState("");
  const [sending, setSending] = useState(false);
  const [busy, setBusy] = useState<"activate" | "stop" | "maestro" | null>(null);
  const [sideError, setSideError] = useState<string | null>(null);
  const scroller = useRef<HTMLDivElement | null>(null);

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

  useEffect(() => {
    const node = scroller.current;
    if (!node) {
      return;
    }
    node.scrollTop = node.scrollHeight;
  }, [project?.logs.length]);

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

  async function handleSend(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const body = draft.trim();
    if (!body || working || sending) {
      return;
    }
    setSending(true);
    setSendError(null);
    try {
      const data = await graphql<{ sendProjectMessage: Project }>(
        `mutation Send($id: ID!, $body: String!) {
          sendProjectMessage(projectId: $id, body: $body) { ${PROJECT_FIELDS} }
        }`,
        { id, body },
      );
      setProject(data.sendProjectMessage);
      setDraft("");
    } catch (err) {
      setSendError(err instanceof Error ? err.message : "Gönderilemedi.");
    } finally {
      setSending(false);
    }
  }

  async function runMutation(
    kind: "activate" | "stop" | "maestro",
    query: string,
    variables: Record<string, unknown>,
    key: "activateProject" | "deactivateProject" | "runMaestro",
  ) {
    setBusy(kind);
    setSideError(null);
    try {
      const data = await graphql<Record<string, Project>>(query, variables);
      setProject(data[key]);
    } catch (err) {
      setSideError(err instanceof Error ? err.message : "İşlem yapılamadı.");
    } finally {
      setBusy(null);
    }
  }

  async function handleActivate() {
    await runMutation(
      "activate",
      `mutation Activate($id: ID!) {
        activateProject(id: $id) { ${PROJECT_FIELDS} }
      }`,
      { id },
      "activateProject",
    );
  }

  async function handleDeactivate() {
    await runMutation(
      "stop",
      `mutation Deactivate($id: ID!) {
        deactivateProject(id: $id) { ${PROJECT_FIELDS} }
      }`,
      { id },
      "deactivateProject",
    );
  }

  async function handleRunMaestro() {
    await runMutation(
      "maestro",
      `mutation RunMaestro($id: ID!) {
        runMaestro(projectId: $id) { ${PROJECT_FIELDS} }
      }`,
      { id },
      "runMaestro",
    );
  }

  return (
    <AppShell user={user} title={project.name} status={projectStatusLabel(project.status)}>
      <div className="cherry-enter grid min-h-full gap-px bg-border lg:grid-cols-[minmax(240px,1fr)_minmax(320px,1.4fr)_minmax(240px,1fr)]">
        <section className="flex flex-col gap-4 bg-background p-6">
          <h1 className="text-base font-medium">{project.name}</h1>
          <p className="text-muted-foreground">{project.brief}</p>
          <p className="font-mono text-[11px] text-muted-foreground">
            {stackLabel(project.stack)} · {stackSourceHint(project.stack)} · {backendTargetLabel(project.backendTarget)} · {projectStatusLabel(project.status)}
          </p>
          <p className="text-muted-foreground">
            Teslim bu dilin kaynağıdır (Clean Architecture). HTML maket yalnızca Maestro ekranında; zip’e girmez.
          </p>
          <p className="break-all font-mono text-[11px] text-muted-foreground">{project.rootPath}</p>
          <Button type="button" variant="outline" onClick={() => router.push(`/projects/${id}/maestro`)}>
            Maestro ekranını aç
          </Button>
          <Button type="button" variant="outline" onClick={() => router.push("/connections")}>
            Bağlantılar
          </Button>
          {project.status === "READY" ? (
            <Button
              type="button"
              onClick={() => {
                void downloadZip(id);
              }}
            >
              Zip indir — {stackSourceHint(project.stack)}
            </Button>
          ) : null}
          <ul className="font-mono text-[11px] text-muted-foreground">
            {project.files.length === 0 ? (
              <li>{working ? "Dosyalar yazılıyor…" : "Dosya yok"}</li>
            ) : (
              project.files.slice(0, 16).map((file) => <li key={file.path}>{file.path}</li>)
            )}
          </ul>
        </section>
        <section className="flex min-h-[280px] flex-col bg-background p-6">
          <h2 className="mb-3 text-sm font-medium">Sohbet</h2>
          <p className="mb-3 text-muted-foreground">
            Buraya yaz. OpenCode ayrı açılmaz; ajan Cherry içinde, arka planda çalışır.
          </p>
          <div
            ref={scroller}
            className="min-h-0 flex-1 space-y-2 overflow-auto rounded-[10px] border border-border bg-card/40 p-3"
          >
            {project.logs.length === 0 ? (
              <p className="text-muted-foreground">Bekleniyor…</p>
            ) : (
              project.logs.map((line, index) => <ChatLine key={line.at + index} line={line} />)
            )}
          </div>
          <form onSubmit={(event) => void handleSend(event)} className="mt-3 flex gap-2">
            <Input
              value={draft}
              onChange={(event) => setDraft(event.target.value)}
              placeholder={working ? "Ajan yazıyor…" : "Ne değişsin?"}
              disabled={working || sending}
              className="cherry-focus"
            />
            <Button type="submit" disabled={working || sending || draft.trim().length === 0}>
              Gönder
            </Button>
          </form>
          {sendError ? <p className="mt-2 text-destructive">{sendError}</p> : null}
        </section>
        <section className="flex flex-col gap-4 bg-background p-6">
          <h2 className="text-sm font-medium">Yerel aktif</h2>
          <p className="font-mono text-[11px] text-muted-foreground">
            {activateStatusLabel(project.activate.status)}
            {project.activate.port ? ` · 127.0.0.1:${project.activate.port}` : ""}
          </p>
          {project.activate.url ? (
            <p className="break-all font-mono text-[12px] text-foreground">{project.activate.url}</p>
          ) : (
            <p className="text-muted-foreground">
              Müşteri API’si bu makinede 47000–47999 aralığında kalkar. Public barındırma yok.
            </p>
          )}
          <p className="text-muted-foreground">{project.activate.note}</p>
          <div className="flex flex-wrap gap-2">
            {project.activate.status === "RUNNING" ? (
              <Button
                type="button"
                variant="outline"
                disabled={busy !== null}
                onClick={() => {
                  void handleDeactivate();
                }}
              >
                {busy === "stop" ? "Duruyor…" : "Durdur"}
              </Button>
            ) : (
              <Button
                type="button"
                disabled={busy !== null || project.activate.status === "STARTING" || project.activate.status === "STOPPING"}
                onClick={() => {
                  void handleActivate();
                }}
              >
                {activateButtonLabel(project.activate.status, busy)}
              </Button>
            )}
          </div>

          <h2 className="mt-4 text-sm font-medium">Maestro</h2>
          {project.maestro.ready ? (
            <>
              <p className="text-muted-foreground">
                {project.maestro.deviceStatus === "device"
                  ? "Cihaz görüldü. Koşu gerçek sonuç yazar; PASSED uydurulmaz."
                  : "Cihaz yok. Koşu SKIPPED olur — geçti sayılmaz."}
              </p>
              {project.maestro.flows[0] ? (
                <p className="font-mono text-[11px] text-muted-foreground">
                  {project.maestro.flows[0].name} · {maestroResultLabel(project.maestro.flows[0].result)}
                </p>
              ) : (
                <p className="text-muted-foreground">YAML henüz yok.</p>
              )}
              <Button
                type="button"
                variant="outline"
                disabled={busy !== null}
                onClick={() => {
                  void handleRunMaestro();
                }}
              >
                {busy === "maestro" ? "Maestro koşuyor…" : "Maestro’yu koş"}
              </Button>
              <Button type="button" variant="ghost" onClick={() => router.push(`/projects/${id}/maestro`)}>
                Tasarımı ve YAML’i göster
              </Button>
            </>
          ) : (
            <p className="text-muted-foreground">
              Ajan yazmayı bitirince bu sütun ve Maestro ekranı açılır. İstediğin anda da açabilirsin.
            </p>
          )}
          {sideError ? <p className="text-destructive">{sideError}</p> : null}
          <Link href="/projects" className="text-[13px] text-muted-foreground underline-offset-4 hover:underline">
            Projelere dön
          </Link>
        </section>
      </div>
    </AppShell>
  );
}

function ChatLine({ line }: { line: JobLog }) {
  return (
    <div className={cn("flex", alignFor(line.role))}>
      <div className={cn("max-w-[90%] rounded-[10px] px-3 py-2 text-[13px] leading-5", bubbleFor(line.role))}>
        {line.role !== "SYSTEM" ? (
          <span className="mb-1 block font-mono text-[10px] text-muted-foreground">{labelFor(line.role)}</span>
        ) : null}
        <span className={line.role === "SYSTEM" ? "font-mono text-[12px] leading-[18px] whitespace-pre-wrap" : "whitespace-pre-wrap"}>
          {line.message}
        </span>
      </div>
    </div>
  );
}

function alignFor(role: ChatRole): string {
  switch (role) {
    case "USER":
      return "justify-end";
    case "AGENT":
      return "justify-start";
    case "SYSTEM":
      return "justify-start";
    default: {
      const _never: never = role;
      return _never;
    }
  }
}

function bubbleFor(role: ChatRole): string {
  switch (role) {
    case "USER":
      return "bg-primary/15 text-foreground";
    case "AGENT":
      return "border border-border bg-background";
    case "SYSTEM":
      return "text-muted-foreground";
    default: {
      const _never: never = role;
      return _never;
    }
  }
}

function labelFor(role: ChatRole): string {
  switch (role) {
    case "USER":
      return "Sen";
    case "AGENT":
      return "Ajan";
    case "SYSTEM":
      return "Sistem";
    default: {
      const _never: never = role;
      return _never;
    }
  }
}

function activateButtonLabel(status: ActivateStatus, busy: "activate" | "stop" | "maestro" | null): string {
  if (busy === "activate") {
    return "Kalkıyor…";
  }
  if (busy === "stop") {
    return "Duruyor…";
  }
  switch (status) {
    case "IDLE":
      return "Yerelde başlat";
    case "STARTING":
      return "Kalkıyor…";
    case "RUNNING":
      return "Çalışıyor";
    case "STOPPING":
      return "Duruyor…";
    case "FAILED":
      return "Yeniden dene";
    default: {
      const _never: never = status;
      return _never;
    }
  }
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
  anchor.download = "cherry.zip";
  anchor.click();
  URL.revokeObjectURL(url);
}
