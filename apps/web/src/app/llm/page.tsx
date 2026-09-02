"use client";

import { useRouter } from "next/navigation";
import { useEffect, useRef, useState } from "react";

import { AppShell } from "@/components/app-shell";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  LLM_ADMIN_FIELDS,
  getToken,
  graphql,
  llmOccupancyLabel,
  type LlmAdmin,
  type LlmSlotCard,
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
  const rooted = useRef(false);

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
    if (!rooted.current) {
      setMcpRoot(data.llmAdmin.mcpRoot);
      rooted.current = true;
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
    const poll = window.setInterval(() => {
      void load().catch(() => {
        /* occupancy poll; keep last */
      });
    }, 1500);
    return () => {
      window.clearTimeout(timer);
      window.clearInterval(poll);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps -- mount
  }, [router]);

  if (error && !user) {
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
      <div className="icerde-enter mx-auto flex max-w-5xl flex-col gap-6 p-8">
        <div>
          <h1 className="text-base font-medium">LLM yönetici</h1>
          <p className="text-muted-foreground">
            A ve B aynı işi yapar. Boş olan alır; ikisi de meşgulse kuyruk. Kod/test ayrımı yok. Colab yok —
            fine-tune sonra.
          </p>
        </div>
        {error ? (
          <p className="text-destructive" role="alert">
            {error}
          </p>
        ) : null}
        <p className="font-mono text-[11px] text-muted-foreground">
          kuyruk {admin.queued} · son işçi {admin.activeSlot}
        </p>
        <div className="grid gap-4 md:grid-cols-2">
          <SlotCard
            card={admin.slotA}
            pending={pending}
            onSelect={(id) => {
              setPending(true);
              void graphql<{ setActiveVersion: LlmAdmin }>(
                `mutation Set($id: ID!) { setActiveVersion(id: $id) { ${LLM_ADMIN_FIELDS} } }`,
                { id },
              )
                .then((data) => setAdmin(data.setActiveVersion))
                .catch((err: unknown) => {
                  setError(err instanceof Error ? err.message : "Değiştirilemedi.");
                })
                .finally(() => setPending(false));
            }}
          />
          <SlotCard
            card={admin.slotB}
            pending={pending}
            onSelect={(id) => {
              setPending(true);
              void graphql<{ setActiveVersion: LlmAdmin }>(
                `mutation Set($id: ID!) { setActiveVersion(id: $id) { ${LLM_ADMIN_FIELDS} } }`,
                { id },
              )
                .then((data) => setAdmin(data.setActiveVersion))
                .catch((err: unknown) => {
                  setError(err instanceof Error ? err.message : "Değiştirilemedi.");
                })
                .finally(() => setPending(false));
            }}
          />
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
                    <th className="px-3 py-2 font-medium">İşçi</th>
                    <th className="px-3 py-2 font-medium">Sürüm</th>
                    <th className="px-3 py-2 font-medium">Redakt</th>
                    <th className="px-3 py-2 font-medium">Çıktı</th>
                  </tr>
                </thead>
                <tbody>
                  {admin.completions.map((row) => (
                    <tr key={row.at + row.slot + row.outputPreview} className="border-b border-border last:border-0">
                      <td className="px-3 py-2 font-mono text-[11px] whitespace-nowrap">{row.at}</td>
                      <td className="px-3 py-2 font-mono text-[11px]">{row.slot || "A"}</td>
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

function SlotCard({
  card,
  pending,
  onSelect,
}: {
  card: LlmSlotCard;
  pending: boolean;
  onSelect: (id: string) => void;
}) {
  const busy = card.occupancy === "BUSY";
  return (
    <article
      className={cn(
        "rounded-[10px] border bg-card/40 p-4",
        busy ? "border-primary" : "border-border",
      )}
    >
      <p className="font-mono text-[11px] text-muted-foreground">
        {busy ? "MEŞGUL" : "BOŞ"} · {llmOccupancyLabel(card.occupancy)}
      </p>
      <h2 className="text-base font-medium">LLM {card.slot}</h2>
      <p className="mt-2 text-muted-foreground">{card.role}</p>
      <div className="mt-3 flex flex-wrap gap-2">
        {card.versions.map((version) => (
          <Button
            key={version.id}
            type="button"
            size="sm"
            variant={card.activeVersionId === version.id ? "default" : "outline"}
            disabled={pending}
            onClick={() => onSelect(version.id)}
          >
            {version.name}
          </Button>
        ))}
      </div>
      <p className="mt-3 font-mono text-[11px] text-muted-foreground">
        {card.versions.find((item) => item.id === card.activeVersionId)?.note}
      </p>
    </article>
  );
}
