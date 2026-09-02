"use client";

import { useRouter } from "next/navigation";
import { useEffect, useState, type FormEvent } from "react";

import { AppShell } from "@/components/app-shell";
import { ProviderMark } from "@/components/provider-mark";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  getToken,
  graphql,
  PROJECT_FIELDS,
  stackLabel,
  stackSourceHint,
  backendTargetLabel,
  type BackendTarget,
  type Connection,
  type Project,
  type ProjectStack,
  type User,
} from "@/lib/api";
import { setLastProjectId } from "@/lib/last-project";
import { cn } from "@/lib/utils";

const STACKS: ProjectStack[] = ["EXPO", "FLUTTER", "NATIVE"];

export default function NewProjectPage() {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [name, setName] = useState("");
  const [brief, setBrief] = useState("");
  const [stack, setStack] = useState<ProjectStack>("EXPO");
  const [backend, setBackend] = useState<BackendTarget>("LOCAL");
  const [connections, setConnections] = useState<Connection[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [pending, setPending] = useState(false);

  useEffect(() => {
    if (!getToken()) {
      router.replace("/");
      return;
    }
    const timer = window.setTimeout(() => {
      void graphql<{ me: User | null; connections: Connection[] }>(
        `query Me {
          me { id email workspaceKind totpEnabled }
          connections { kind status account note }
        }`,
      )
        .then((data) => {
          if (!data.me) {
            router.replace("/");
            return;
          }
          setUser(data.me);
          setConnections(data.connections);
        })
        .catch((err: unknown) => {
          setError(err instanceof Error ? err.message : "Yüklenemedi.");
        });
    }, 0);
    return () => {
      window.clearTimeout(timer);
    };
  }, [router]);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    if (!name.trim() || brief.trim().length < 8) {
      setError("Ad ve en az 8 karakterlik brif gerekli.");
      return;
    }
    setPending(true);
    try {
      const data = await graphql<{ createProject: Project }>(
        `mutation Create($name: String!, $brief: String!, $stack: ProjectStack!, $backendTarget: BackendTarget) {
          createProject(name: $name, brief: $brief, stack: $stack, backendTarget: $backendTarget) {
            ${PROJECT_FIELDS}
          }
        }`,
        { name: name.trim(), brief: brief.trim(), stack, backendTarget: backend },
      );
      setLastProjectId(data.createProject.id);
      router.push(`/projects/${data.createProject.id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Oluşturulamadı.");
    } finally {
      setPending(false);
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

  return (
    <AppShell user={user} title="Yeni proje">
      <form
        onSubmit={(event) => void handleSubmit(event)}
        className="icerde-enter mx-auto flex max-w-xl flex-col gap-5 p-8"
      >
        <div>
          <h1 className="text-base font-medium">Ne üretilsin?</h1>
          <p className="text-muted-foreground">
            Brifi yaz, yığını seç. Ajan o dilde, Clean Architecture ile yazar (Expo SDK 57, Flutter 3.47, SwiftUI) — HTML site değil.
            Stüdyo sohbetine düşersin; OpenCode açılmaz.
          </p>
        </div>
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="name">Uygulama adı</Label>
          <Input
            id="name"
            value={name}
            onChange={(event) => setName(event.target.value)}
            placeholder="Mahalle kahvesi"
            className="icerde-focus h-9"
          />
        </div>
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="brief">Ne yapsın?</Label>
          <Textarea
            id="brief"
            value={brief}
            onChange={(event) => setBrief(event.target.value)}
            placeholder="Sipariş kuyruğu, masa QR, kurye durumu…"
          />
        </div>
        <fieldset className="flex flex-col gap-2">
          <legend className="text-sm font-medium">Yığın</legend>
          <div className="grid gap-2 sm:grid-cols-3">
            {STACKS.map((item) => (
              <button
                key={item}
                type="button"
                onClick={() => setStack(item)}
                className={cn(
                  "rounded-[10px] border px-3 py-2 text-left text-[13px] transition-colors",
                  stack === item
                    ? "border-primary bg-primary/10 text-foreground"
                    : "border-border text-muted-foreground hover:bg-muted/40",
                )}
              >
                {stackLabel(item)}
                <span className="mt-1 block text-[11px] text-muted-foreground">{stackSourceHint(item)}</span>
              </button>
            ))}
          </div>
        </fieldset>
        <fieldset className="flex flex-col gap-2">
          <legend className="text-sm font-medium">Backend hedefi</legend>
          <p className="text-muted-foreground">
            İçerde host değil. Cloud için önce Bağlantılar’dan hesap bağla.
          </p>
          <div className="grid gap-2 sm:grid-cols-2">
            {(["LOCAL", "SUPABASE", "CLOUDFLARE", "RENDER"] as const).map((item) => {
              const enabled =
                item === "LOCAL" ||
                connections.some((conn) => conn.kind === item && conn.status === "CONNECTED");
              return (
                <button
                  key={item}
                  type="button"
                  disabled={!enabled}
                  onClick={() => setBackend(item)}
                  className={cn(
                    "rounded-[10px] border px-3 py-2 text-left text-[13px] transition-colors",
                    backend === item
                      ? "border-primary bg-primary/10 text-foreground"
                      : "border-border text-muted-foreground hover:bg-muted/40",
                    !enabled ? "opacity-50" : "",
                  )}
                >
                  <span className="flex items-center gap-2">
                    {item === "LOCAL" ? null : (
                      <ProviderMark kind={item} className="size-4 shrink-0" />
                    )}
                    {backendTargetLabel(item)}
                  </span>
                  <span className="mt-1 block text-[11px] text-muted-foreground">
                    {enabled ? (item === "LOCAL" ? "localhost 47000–47999" : "bağlı") : "bağlı değil"}
                  </span>
                </button>
              );
            })}
          </div>
        </fieldset>
        {error ? (
          <p className="text-destructive" role="alert">
            {error}
          </p>
        ) : null}
        <Button type="submit" size="lg" className="h-10" disabled={pending}>
          {pending ? "Kuyruğa alınıyor…" : "Arka planda üret"}
        </Button>
      </form>
    </AppShell>
  );
}
