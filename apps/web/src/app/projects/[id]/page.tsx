"use client";

import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { useEffect, useRef, useState, type FormEvent } from "react";

import { AppShell } from "@/components/app-shell";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  PROJECT_FIELDS,
  getToken,
  graphql,
  projectStatusLabel,
  stackLabel,
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

  return (
    <AppShell user={user} title={project.name} status={projectStatusLabel(project.status)}>
      <div className="icerde-enter grid min-h-full gap-px bg-border lg:grid-cols-[minmax(240px,1fr)_minmax(320px,1.4fr)_minmax(240px,1fr)]">
        <section className="flex flex-col gap-4 bg-background p-6">
          <h1 className="text-base font-medium">{project.name}</h1>
          <p className="text-muted-foreground">{project.brief}</p>
          <p className="font-mono text-[11px] text-muted-foreground">
            {stackLabel(project.stack)} · {projectStatusLabel(project.status)}
          </p>
          <p className="break-all font-mono text-[11px] text-muted-foreground">{project.rootPath}</p>
          <Button type="button" variant="outline" onClick={() => router.push(`/projects/${id}/maestro`)}>
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
              project.files.slice(0, 12).map((file) => <li key={file.path}>{file.path}</li>)
            )}
          </ul>
        </section>
        <section className="flex min-h-[280px] flex-col bg-background p-6">
          <h2 className="mb-3 text-sm font-medium">Sohbet</h2>
          <p className="mb-3 text-muted-foreground">
            Buraya yaz. OpenCode ayrı açılmaz; ajan İçerde içinde, arka planda çalışır.
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
              className="icerde-focus"
            />
            <Button type="submit" disabled={working || sending || draft.trim().length === 0}>
              Gönder
            </Button>
          </form>
          {sendError ? <p className="mt-2 text-destructive">{sendError}</p> : null}
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
