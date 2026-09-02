"use client";

import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";

import { AppShell } from "@/components/app-shell";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  LLM_ADMIN_FIELDS,
  getToken,
  graphql,
  type LlmAdmin,
  type User,
} from "@/lib/api";
import { cn } from "@/lib/utils";

export default function LlmAdminPage() {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [admin, setAdmin] = useState<LlmAdmin | null>(null);
  const [mcpRoot, setMcpRoot] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [pending, setPending] = useState(false);

  async function load() {
    const data = await graphql<{ me: User | null; llmAdmin: LlmAdmin }>(
      `query LlmAdmin {
        me { id email workspaceKind totpEnabled }
        llmAdmin { ${LLM_ADMIN_FIELDS} }
      }`,
    );
    if (!data.me) {
      router.replace("/");
      return;
    }
    setUser(data.me);
    setAdmin(data.llmAdmin);
    setMcpRoot(data.llmAdmin.mcpRoot);
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

  if (error) {
    return (
      <div className="flex min-h-full items-center justify-center text-destructive">{error}</div>
    );
  }
  if (!user || !admin) {
    return (
      <div className="flex min-h-full items-center justify-center text-muted-foreground">
        Yükleniyor…
      </div>
    );
  }

  return (
    <AppShell user={user} title="LLM yönetici" status="GDPR katmanı bağlı">
      <div className="mx-auto flex max-w-5xl flex-col gap-6 p-8">
        <div>
          <h1 className="text-base font-medium">LLM yönetici</h1>
          <p className="text-muted-foreground">
            Her çağrı redact → model → tarama → denetim. Yazıcı OpenCode (`opencode run`).
            A ve B aynı işi yapar; B ikinci kapasite işçisidir. B ve Colab henüz yok.
          </p>
        </div>
        <div className="grid gap-4 md:grid-cols-2">
          <article className="rounded-[10px] border border-primary bg-card/40 p-4">
            <p className="text-[11px] text-primary">AKTİF</p>
            <h2 className="text-base font-medium">LLM A — {admin.slotA.role}</h2>
            <p className="mt-2 text-muted-foreground">MCP read-file kökü proje klasörü. Versiyon pointer’ı sonraki cevapları değiştirir.</p>
            <div className="mt-3 flex flex-wrap gap-2">
              {admin.slotA.versions.map((version) => (
                <Button
                  key={version.id}
                  type="button"
                  size="sm"
                  variant={admin.slotA.activeVersionId === version.id ? "default" : "outline"}
                  disabled={pending}
                  onClick={() => {
                    setPending(true);
                    void graphql<{ setActiveVersion: LlmAdmin }>(
                      `mutation Set($id: ID!) { setActiveVersion(id: $id) { ${LLM_ADMIN_FIELDS} } }`,
                      { id: version.id },
                    )
                      .then((data) => {
                        setAdmin(data.setActiveVersion);
                      })
                      .catch((err: unknown) => {
                        setError(err instanceof Error ? err.message : "Değiştirilemedi.");
                      })
                      .finally(() => setPending(false));
                  }}
                >
                  {version.name}
                </Button>
              ))}
            </div>
            <p className="mt-3 font-mono text-[11px] text-muted-foreground">
              {admin.slotA.versions.find((item) => item.id === admin.slotA.activeVersionId)?.note}
            </p>
          </article>
          <article className="rounded-[10px] border border-border bg-card/40 p-4">
            <p className="text-[11px] text-muted-foreground">SONRA</p>
            <h2 className="text-base font-medium">LLM B — {admin.slotB.role}</h2>
            <p className="mt-2 text-muted-foreground">
              Dilim 7: ikinci işçi. On kişi aynı anda üretince boş olan alır; ikisi de
              meşgulse kuyruk. Sahte rol yok — B bağlı değil.
            </p>
          </article>
        </div>
        <section className="flex flex-col gap-2 rounded-[10px] border border-border p-4">
          <h2 className="text-sm font-medium">MCP kök</h2>
          <p className="text-muted-foreground">Model yalnızca bu kök altını okur. Yeni proje üretince otomatik o klasöre iner.</p>
          <div className="flex flex-col gap-2 sm:flex-row">
            <Input
              value={mcpRoot}
              onChange={(event) => setMcpRoot(event.target.value)}
              className="icerde-focus font-mono text-[12px]"
            />
            <Button
              type="button"
              variant="outline"
              disabled={pending}
              onClick={() => {
                setPending(true);
                void graphql<{ setMcpRoot: LlmAdmin }>(
                  `mutation Root($path: String!) { setMcpRoot(path: $path) { ${LLM_ADMIN_FIELDS} } }`,
                  { path: mcpRoot },
                )
                  .then((data) => setAdmin(data.setMcpRoot))
                  .catch((err: unknown) => {
                    setError(err instanceof Error ? err.message : "Kök yazılamadı.");
                  })
                  .finally(() => setPending(false));
              }}
            >
              Kaydet
            </Button>
          </div>
        </section>
        <section className="flex flex-col gap-2">
          <h2 className="text-sm font-medium">Son tamamlamalar</h2>
          {admin.completions.length === 0 ? (
            <p className="text-muted-foreground">Henüz GDPR sarmalı çağrı yok. Bir proje üret.</p>
          ) : (
            <div className="overflow-x-auto rounded-[10px] border border-border">
              <table className="w-full text-left text-[13px]">
                <thead className="border-b border-border text-muted-foreground">
                  <tr>
                    <th className="px-3 py-2 font-medium">Zaman</th>
                    <th className="px-3 py-2 font-medium">Sürüm</th>
                    <th className="px-3 py-2 font-medium">Redakt</th>
                    <th className="px-3 py-2 font-medium">Çıktı</th>
                  </tr>
                </thead>
                <tbody>
                  {admin.completions.map((row) => (
                    <tr key={row.at + row.outputPreview} className="border-b border-border last:border-0">
                      <td className="px-3 py-2 font-mono text-[11px] whitespace-nowrap">{row.at}</td>
                      <td className="px-3 py-2">
                        {row.versionName}
                        <span className="block font-mono text-[11px] text-muted-foreground">{row.channel}</span>
                      </td>
                      <td className="px-3 py-2 font-mono text-[11px]">
                        {row.inputRedactions}/{row.outputRedactions}
                      </td>
                      <td className={cn("px-3 py-2 text-muted-foreground")}>{row.outputPreview}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </section>
      </div>
    </AppShell>
  );
}
