"use client";

import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { Cable } from "lucide-react";

import { AppShell } from "@/components/app-shell";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  connectionKindLabel,
  connectionStatusLabel,
  connectionTokenHint,
  getToken,
  graphql,
  type Connection,
  type ConnectionKind,
  type Project,
  type User,
} from "@/lib/api";
import { getLastProjectId } from "@/lib/last-project";

const CONN_FIELDS = `kind status account tokenHint note`;

export default function ConnectionsPage() {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [list, setList] = useState<Connection[]>([]);
  const [projects, setProjects] = useState<Project[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [info, setInfo] = useState<string | null>(null);
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
      setInfo(`${connectionKindLabel(kind)} bağlandı. Token ekranda durmaz.`);
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
        setInfo(data.pushProjectGithub.note);
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
  const noneConnected = list.every((item) => item.status !== "CONNECTED");

  return (
    <AppShell user={user} title="Bağlantılar">
      <div className="mx-auto flex max-w-3xl flex-col gap-6 p-8">
        <div>
          <h1 className="flex items-center gap-2 text-base font-medium">
            <Cable className="size-4" />
            Bağlantılar
          </h1>
          <p className="text-muted-foreground">
            Kendi hesapların. İçerde barındırmaz. Boş token bağlı sayılmaz. Token sohbete ve zip’e gitmez.
          </p>
        </div>
        {error ? (
          <p className="text-destructive" role="alert">
            {error}
          </p>
        ) : null}
        {info ? <p className="text-muted-foreground">{info}</p> : null}
        {noneConnected ? (
          <p className="text-muted-foreground">Henüz bağlı hesap yok. Aşağıdan ekle.</p>
        ) : null}

        <div className="grid gap-3">
          {list.map((item) => {
            const form = draft(item.kind);
            const connected = item.status === "CONNECTED";
            return (
              <Card key={item.kind} className="py-4">
                <CardHeader>
                  <CardTitle className="flex items-center justify-between gap-2">
                    <span>{connectionKindLabel(item.kind)}</span>
                    <span className="font-mono text-[11px] font-normal text-muted-foreground">
                      {connectionStatusLabel(item.status)}
                    </span>
                  </CardTitle>
                  <CardDescription>{item.note}</CardDescription>
                </CardHeader>
                <CardContent className="flex flex-col gap-3">
                  {connected ? (
                    <div className="flex flex-wrap items-center justify-between gap-2">
                      <p className="font-mono text-[12px]">
                        {item.account}
                        {item.tokenHint ? ` · ${item.tokenHint}` : ""}
                      </p>
                      <Button
                        type="button"
                        variant="outline"
                        disabled={busy === item.kind}
                        onClick={() => {
                          void handleDisconnect(item.kind);
                        }}
                      >
                        {busy === item.kind ? "Kopuyor…" : "Kopar"}
                      </Button>
                    </div>
                  ) : (
                    <form
                      className="flex flex-col gap-2"
                      onSubmit={(event) => {
                        event.preventDefault();
                        void handleConnect(item.kind);
                      }}
                    >
                      <Label htmlFor={`${item.kind}-account`}>Hesap</Label>
                      <Input
                        id={`${item.kind}-account`}
                        value={form.account}
                        autoComplete="off"
                        placeholder="kullanıcı veya proje"
                        onChange={(event) => {
                          const value = event.target.value;
                          setDrafts((prev) => ({
                            ...prev,
                            [item.kind]: { ...form, account: value },
                          }));
                        }}
                      />
                      <Label htmlFor={`${item.kind}-token`}>{connectionTokenHint(item.kind)}</Label>
                      <Input
                        id={`${item.kind}-token`}
                        type="password"
                        value={form.token}
                        autoComplete="off"
                        onChange={(event) => {
                          const value = event.target.value;
                          setDrafts((prev) => ({
                            ...prev,
                            [item.kind]: { ...form, token: value },
                          }));
                        }}
                      />
                      <Button type="submit" disabled={busy === item.kind}>
                        {busy === item.kind ? "Bağlanıyor…" : "Bağla"}
                      </Button>
                    </form>
                  )}
                </CardContent>
              </Card>
            );
          })}
        </div>

        <section className="flex flex-col gap-3 rounded-[10px] border border-border p-4">
          <h2 className="text-sm font-medium">Projeyi GitHub’a gönder</h2>
          <p className="text-muted-foreground">
            GitHub bağlı olmalı. Sahte başarı yok; push reddedilirse hata görünür.
          </p>
          {github?.status !== "CONNECTED" ? (
            <p className="text-muted-foreground">Önce GitHub bağla.</p>
          ) : projects.length === 0 ? (
            <p className="text-muted-foreground">Gönderilecek proje yok.</p>
          ) : (
            <>
              <Label htmlFor="push-project">Proje</Label>
              <select
                id="push-project"
                className="h-9 rounded-md border border-border bg-background px-2 text-sm"
                value={projectId}
                onChange={(event) => setProjectId(event.target.value)}
              >
                {projects.map((project) => (
                  <option key={project.id} value={project.id}>
                    {project.name}
                  </option>
                ))}
              </select>
              <Label htmlFor="push-repo">Repo (owner/ad)</Label>
              <Input
                id="push-repo"
                value={repo}
                placeholder="emrah/kahve"
                onChange={(event) => setRepo(event.target.value)}
              />
              <Button
                type="button"
                disabled={busy === "push"}
                onClick={() => {
                  void handlePush();
                }}
              >
                {busy === "push" ? "Gönderiliyor…" : "GitHub’a gönder"}
              </Button>
            </>
          )}
        </section>
      </div>
    </AppShell>
  );
}
