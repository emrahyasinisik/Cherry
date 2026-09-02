"use client";

import { useRouter, useSearchParams } from "next/navigation";
import { Suspense, useEffect, useState } from "react";

import { AppShell } from "@/components/app-shell";
import { ProviderTile } from "@/components/provider-mark";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  connectionAuthLabel,
  connectionKindLabel,
  connectionStatusLabel,
  connectionTokenHint,
  getToken,
  graphql,
  type Connection,
  type ConnectionKind,
  type OAuthMode,
  type Project,
  type User,
} from "@/lib/api";
import { getLastProjectId } from "@/lib/last-project";
import { isConnectionKind, providerPurpose } from "@/lib/oauth";
import { cn } from "@/lib/utils";

const CONN_FIELDS = `kind status account tokenHint note authMethod scopes`;

function oauthBanner(status: string | null, kind: string | null): { tone: "ok" | "warn"; text: string } | null {
  const name = isConnectionKind(kind) ? connectionKindLabel(kind) : null;
  switch (status) {
    case "connected":
      return { tone: "ok", text: name ? `${name} bağlandı.` : "Hesap bağlandı." };
    case "denied":
      return { tone: "warn", text: "İzin verilmedi. Hesap bağlanmadı." };
    case "expired":
      return { tone: "warn", text: "OAuth isteğinin süresi doldu. Yeniden bağla." };
    case "error":
      return { tone: "warn", text: "OAuth tamamlanamadı." };
    default:
      return null;
  }
}

function ConnectionsBody() {
  const router = useRouter();
  const params = useSearchParams();
  const [user, setUser] = useState<User | null>(null);
  const [list, setList] = useState<Connection[]>([]);
  const [projects, setProjects] = useState<Project[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [info, setInfo] = useState<{ tone: "ok" | "warn"; text: string } | null>(
    oauthBanner(params.get("oauth"), params.get("kind")),
  );
  const [busy, setBusy] = useState<ConnectionKind | "push" | null>(null);
  const [drafts, setDrafts] = useState<Record<string, { account: string; token: string }>>({});
  const [repo, setRepo] = useState("");
  const [projectId, setProjectId] = useState("");

  async function load() {
    const data = await graphql<{
      me: User | null;
      connections: Connection[];
      projects: Project[];
    }>(`query Connections {
      me { id email workspaceKind totpEnabled }
      connections { ${CONN_FIELDS} }
      projects { id name status }
    }`);
    if (!data.me) {
      router.replace("/");
      return;
    }
    setUser(data.me);
    setList(data.connections);
    setProjects(data.projects);
    const last = getLastProjectId();
    if (last && data.projects.some((item) => item.id === last)) {
      setProjectId(last);
    } else if (data.projects[0]) {
      setProjectId(data.projects[0].id);
    }
  }

  useEffect(() => {
    if (!getToken()) {
      router.replace("/");
      return;
    }
    const timer = window.setTimeout(() => {
      void load().catch((err: unknown) => {
        setError(err instanceof Error ? err.message : "Yüklenemedi.");
      });
    }, 0);
    return () => {
      window.clearTimeout(timer);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps -- mount
  }, [router]);

  function draft(kind: ConnectionKind): { account: string; token: string } {
    return drafts[kind] ?? { account: "", token: "" };
  }

  async function handleOAuth(kind: ConnectionKind) {
    setError(null);
    setInfo(null);
    setBusy(kind);
    try {
      const data = await graphql<{
        startConnectionOAuth: { authorizeUrl: string; state: string; mode: OAuthMode };
      }>(
        `mutation Start($kind: ConnectionKind!) {
          startConnectionOAuth(kind: $kind) { authorizeUrl state mode }
        }`,
        { kind },
      );
      window.location.assign(data.startConnectionOAuth.authorizeUrl);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "OAuth başlatılamadı.");
      setBusy(null);
    }
  }

  async function handleConnect(kind: ConnectionKind) {
    setError(null);
    setInfo(null);
    setBusy(kind);
    const form = draft(kind);
    try {
      await graphql<{ connectProvider: Connection }>(
        `mutation Connect($kind: ConnectionKind!, $account: String!, $token: String!) {
          connectProvider(kind: $kind, account: $account, token: $token) { ${CONN_FIELDS} }
        }`,
        { kind, account: form.account.trim(), token: form.token.trim() },
      );
      setDrafts((prev) => ({ ...prev, [kind]: { account: "", token: "" } }));
      await load();
      setInfo({ tone: "ok", text: `${connectionKindLabel(kind)} token ile bağlandı.` });
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Bağlanamadı.");
    } finally {
      setBusy(null);
    }
  }

  async function handleDisconnect(kind: ConnectionKind) {
    setError(null);
    setInfo(null);
    setBusy(kind);
    try {
      await graphql<{ disconnectProvider: Connection }>(
        `mutation Disconnect($kind: ConnectionKind!) {
          disconnectProvider(kind: $kind) { ${CONN_FIELDS} }
        }`,
        { kind },
      );
      await load();
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Koparılamadı.");
    } finally {
      setBusy(null);
    }
  }

  async function handlePush() {
    setError(null);
    setInfo(null);
    if (!projectId || !repo.trim()) {
      setError("Proje ve owner/ad repo gerekli.");
      return;
    }
    setBusy("push");
    try {
      const data = await graphql<{ pushProjectGithub: { ok: boolean; note: string } }>(
        `mutation Push($projectId: ID!, $repo: String!) {
          pushProjectGithub(projectId: $projectId, repo: $repo) { ok note }
        }`,
        { projectId, repo: repo.trim() },
      );
      if (data.pushProjectGithub.ok) {
        setInfo({ tone: "ok", text: data.pushProjectGithub.note });
      } else {
        setError(data.pushProjectGithub.note);
      }
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Push olmadı.");
    } finally {
      setBusy(null);
    }
  }

  if (error && !user) {
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

  const github = list.find((item) => item.kind === "GITHUB");
  const connectedCount = list.filter((item) => item.status === "CONNECTED").length;

  return (
    <AppShell user={user} title="Bağlantılar">
      <div className="cherry-enter mx-auto flex max-w-[720px] flex-col gap-8 p-8">
        <div>
          <h1 className="text-base font-medium">Bağlantılar</h1>
          <p className="text-muted-foreground">
            Kendi hesapların. Bağla deyince OAuth 2.0 izin ekranı açılır — Cherry barındırmaz.
          </p>
        </div>

        {error ? (
          <p className="rounded-[8px] border border-destructive/40 bg-destructive/10 px-3 py-2 text-destructive" role="alert">
            {error}
          </p>
        ) : null}
        {info ? (
          <p
            className={cn(
              "rounded-[8px] border px-3 py-2",
              info.tone === "ok"
                ? "border-success/40 bg-success/10 text-success"
                : "border-border text-muted-foreground",
            )}
          >
            {info.text}
          </p>
        ) : null}

        <section className="flex flex-col gap-3">
          <div className="flex items-baseline justify-between gap-3">
            <h2 className="font-medium">Hesaplar</h2>
            <p className="font-mono text-[11px] text-muted-foreground">
              {connectedCount === 0 ? "Hiç bağlı yok" : `${connectedCount} bağlı`}
            </p>
          </div>
          <ul className="cherry-stagger divide-y divide-border overflow-hidden rounded-[10px] border border-border">
            {list.map((item) => {
              const form = draft(item.kind);
              const connected = item.status === "CONNECTED";
              const name = connectionKindLabel(item.kind);
              return (
                <li key={item.kind} className="bg-card px-4 py-3 transition-colors hover:bg-muted/40">
                  <div className="flex items-center gap-3">
                    <ProviderTile kind={item.kind} />
                    <div className="min-w-0 flex-1">
                      <p className="flex flex-wrap items-center gap-2">
                        {name}
                        <StatusDot connected={connected} label={connectionStatusLabel(item.status)} />
                      </p>
                      <p className="text-[11px] leading-4 text-muted-foreground">
                        {connected
                          ? `${item.account}${item.authMethod !== "NONE" ? ` · ${connectionAuthLabel(item.authMethod)}` : ""}`
                          : providerPurpose(item.kind)}
                      </p>
                    </div>
                    {connected ? (
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        className="text-destructive"
                        disabled={busy === item.kind}
                        onClick={() => {
                          void handleDisconnect(item.kind);
                        }}
                      >
                        {busy === item.kind ? "Kopuyor…" : "Kopar"}
                      </Button>
                    ) : (
                      <Button
                        type="button"
                        size="sm"
                        disabled={busy === item.kind}
                        onClick={() => {
                          void handleOAuth(item.kind);
                        }}
                      >
                        {busy === item.kind ? "…" : "Bağlan"}
                      </Button>
                    )}
                  </div>
                  {connected ? null : (
                    <details className="mt-2 pl-[52px]">
                      <summary className="cursor-pointer text-[11px] text-muted-foreground">
                        Token yapıştır
                      </summary>
                      <form
                        className="mt-2 grid gap-2 sm:grid-cols-[1fr_1fr_auto] sm:items-end"
                        onSubmit={(event) => {
                          event.preventDefault();
                          void handleConnect(item.kind);
                        }}
                      >
                        <div className="flex flex-col gap-1">
                          <Label htmlFor={`${item.kind}-account`} className="text-[11px]">
                            Hesap
                          </Label>
                          <Input
                            id={`${item.kind}-account`}
                            value={form.account}
                            autoComplete="off"
                            className="h-8"
                            onChange={(event) => {
                              const value = event.target.value;
                              setDrafts((prev) => ({
                                ...prev,
                                [item.kind]: { ...form, account: value },
                              }));
                            }}
                          />
                        </div>
                        <div className="flex flex-col gap-1">
                          <Label htmlFor={`${item.kind}-token`} className="text-[11px]">
                            {connectionTokenHint(item.kind)}
                          </Label>
                          <Input
                            id={`${item.kind}-token`}
                            type="password"
                            value={form.token}
                            autoComplete="off"
                            className="h-8"
                            onChange={(event) => {
                              const value = event.target.value;
                              setDrafts((prev) => ({
                                ...prev,
                                [item.kind]: { ...form, token: value },
                              }));
                            }}
                          />
                        </div>
                        <Button type="submit" variant="outline" size="sm" disabled={busy === item.kind}>
                          Yapıştır
                        </Button>
                      </form>
                    </details>
                  )}
                </li>
              );
            })}
          </ul>
        </section>

        <section className="flex flex-col gap-3">
          <h2 className="font-medium">GitHub’a gönder</h2>
          <p className="text-muted-foreground">
            Sahte başarı yok. Yerel izin ekranı gerçek push yapmaz.
          </p>
          {github?.status !== "CONNECTED" ? (
            <p className="rounded-[10px] border border-dashed border-border px-4 py-6 text-muted-foreground">
              Önce GitHub’ı bağla.
            </p>
          ) : projects.length === 0 ? (
            <p className="rounded-[10px] border border-dashed border-border px-4 py-6 text-muted-foreground">
              Gönderilecek proje yok.
            </p>
          ) : (
            <div className="grid gap-3 rounded-[10px] border border-border bg-card p-4 sm:grid-cols-[1fr_1fr_auto] sm:items-end">
              <div className="flex flex-col gap-1">
                <Label htmlFor="push-project">Proje</Label>
                <select
                  id="push-project"
                  className="h-8 rounded-md border border-border bg-background px-2 text-[13px]"
                  value={projectId}
                  onChange={(event) => setProjectId(event.target.value)}
                >
                  {projects.map((project) => (
                    <option key={project.id} value={project.id}>
                      {project.name}
                    </option>
                  ))}
                </select>
              </div>
              <div className="flex flex-col gap-1">
                <Label htmlFor="push-repo">Repo (owner/ad)</Label>
                <Input
                  id="push-repo"
                  value={repo}
                  className="h-8"
                  placeholder="emrah/kahve"
                  onChange={(event) => setRepo(event.target.value)}
                />
              </div>
              <Button
                type="button"
                disabled={busy === "push"}
                onClick={() => {
                  void handlePush();
                }}
              >
                {busy === "push" ? "Gönderiliyor…" : "Gönder"}
              </Button>
            </div>
          )}
        </section>
      </div>
    </AppShell>
  );
}

function StatusDot({ connected, label }: { connected: boolean; label: string }) {
  return (
    <span className="inline-flex items-center gap-1.5 font-mono text-[11px] font-normal text-muted-foreground">
      <span className={cn("size-1.5 rounded-full", connected ? "bg-success" : "bg-muted-foreground/40")} />
      {label}
    </span>
  );
}

export default function ConnectionsPage() {
  return (
    <Suspense
      fallback={
        <div className="flex min-h-full items-center justify-center text-muted-foreground">Yükleniyor…</div>
      }
    >
      <ConnectionsBody />
    </Suspense>
  );
}
