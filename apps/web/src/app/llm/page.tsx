"use client";

import { useRouter } from "next/navigation";
import { useEffect, useRef, useState } from "react";

import { AppShell } from "@/components/app-shell";
import { Button, buttonVariants } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  LLM_ADMIN_FIELDS,
  getToken,
  graphql,
  llmOccupancyLabel,
  type LlmAdmin,
  type LlmSlotCard,
  type TrainingPack,
  type User,
} from "@/lib/api";
import { cn } from "@/lib/utils";

function downloadText(filename: string, body: string, mime: string) {
  const blob = new Blob([body], { type: mime });
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = filename;
  anchor.click();
  URL.revokeObjectURL(url);
}

type ColabSlot = "A" | "B";

export default function LlmAdminPage() {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [admin, setAdmin] = useState<LlmAdmin | null>(null);
  const [mcpRoot, setMcpRoot] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [pending, setPending] = useState(false);
  const [packNote, setPackNote] = useState<string | null>(null);
  const [regSlot, setRegSlot] = useState<ColabSlot>("A");
  const [regName, setRegName] = useState("v-colab");
  const [regNote, setRegNote] = useState("");
  const [regRef, setRegRef] = useState("cherry_adapter_worker_A.zip");
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
      <div className="cherry-enter mx-auto flex max-w-5xl flex-col gap-6 p-8">
        <div>
          <h1 className="text-base font-medium">LLM yönetici</h1>
          <p className="text-muted-foreground">
            A ve B aynı işi yapar. Boş olan alır; ikisi de meşgulse kuyruk. Fine-tune Colab’de: paketi indir,
            iki T4 oturumu, adapter zip’ini kaydet. Colab üretim inferansı değil.
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
        <section className="flex flex-col gap-3 rounded-[10px] border border-border p-4">
          <h2 className="text-sm font-medium">Colab eğitim dosyaları</h2>
          <p className="text-muted-foreground">
            İki notebook, aynı QLoRA tarifi. Her oturum 16GB T4. Ham PII yok — paket redakte. Canlı iz inceyse seed
            örnekler eklenir.
          </p>
          {packNote ? <p className="text-muted-foreground">{packNote}</p> : null}
          <div className="flex flex-wrap gap-2">
            <Button
              type="button"
              variant="outline"
              disabled={pending}
              onClick={() => {
                setPending(true);
                setError(null);
                void graphql<{ trainingPack: TrainingPack }>(
                  `query Pack {
                    trainingPack { schema filename json jsonl liveExamples seedExamples note }
                  }`,
                )
                  .then((data) => {
                    const pack = data.trainingPack;
                    downloadText(pack.filename, pack.json, "application/json");
                    downloadText(
                      pack.filename.replace(/\.json$/u, ".jsonl"),
                      pack.jsonl,
                      "application/x-ndjson",
                    );
                    setPackNote(
                      `${pack.liveExamples} canlı · ${pack.seedExamples} seed — ${pack.note}`,
                    );
                  })
                  .catch((err: unknown) => {
                    setError(err instanceof Error ? err.message : "Paket indirilemedi.");
                  })
                  .finally(() => setPending(false));
              }}
            >
              Eğitim paketini indir
            </Button>
            <a
              className={buttonVariants({ variant: "outline" })}
              href="/colab/cherry_worker_a.ipynb"
              download
            >
              Notebook A
            </a>
            <a
              className={buttonVariants({ variant: "outline" })}
              href="/colab/cherry_worker_b.ipynb"
              download
            >
              Notebook B
            </a>
            <a
              className={buttonVariants({ variant: "outline" })}
              href="/colab/examples/cherry_training_pack.json"
              download
            >
              Seed paket
            </a>
          </div>
          <p className="font-mono text-[11px] text-muted-foreground">
            colab/cherry_worker_a.ipynb · colab/cherry_worker_b.ipynb
          </p>
          <h3 className="mt-2 text-sm font-medium">Colab sürümü kaydet</h3>
          <p className="text-muted-foreground">
            Adapter zip indikten sonra pointer’ı sen çevirirsin. In-flight iş eski sürüme biter.
          </p>
          <div className="flex flex-wrap gap-2">
            {(["A", "B"] as const).map((slot) => (
              <Button
                key={slot}
                type="button"
                size="sm"
                variant={regSlot === slot ? "default" : "outline"}
                onClick={() => {
                  setRegSlot(slot);
                  setRegRef(`cherry_adapter_worker_${slot}.zip`);
                }}
              >
                İşçi {slot}
              </Button>
            ))}
          </div>
          <div className="grid gap-2 sm:grid-cols-2">
            <Input
              value={regName}
              onChange={(event) => setRegName(event.target.value)}
              placeholder="v-colab"
              className="cherry-focus font-mono text-[12px]"
            />
            <Input
              value={regRef}
              onChange={(event) => setRegRef(event.target.value)}
              placeholder="cherry_adapter_worker_A.zip"
              className="cherry-focus font-mono text-[12px]"
            />
          </div>
          <Input
            value={regNote}
            onChange={(event) => setRegNote(event.target.value)}
            placeholder="T4 QLoRA notu (isteğe bağlı)"
            className="cherry-focus text-[12px]"
          />
          <Button
            type="button"
            variant="outline"
            disabled={pending}
            onClick={() => {
              setPending(true);
              setError(null);
              void graphql<{ registerLlmVersion: LlmAdmin }>(
                `mutation Reg($slot: String!, $name: String!, $note: String!, $ref: String!) {
                  registerLlmVersion(slot: $slot, name: $name, note: $note, checkpointRef: $ref) {
                    ${LLM_ADMIN_FIELDS}
                  }
                }`,
                { slot: regSlot, name: regName, note: regNote, ref: regRef },
              )
                .then((data) => {
                  setAdmin(data.registerLlmVersion);
                  setPackNote(`Kayıt: işçi ${regSlot} · ${regName} · ${regRef}. Aktif yapmak için sürüm düğmesine bas.`);
                })
                .catch((err: unknown) => {
                  setError(err instanceof Error ? err.message : "Sürüm kaydedilemedi.");
                })
                .finally(() => setPending(false));
            }}
          >
            Sürümü kaydet
          </Button>
        </section>
        <section className="flex flex-col gap-2 rounded-[10px] border border-border p-4">
          <h2 className="text-sm font-medium">MCP kök</h2>
          <p className="text-muted-foreground">Model yalnızca bu kök altını okur. Yeni proje üretince otomatik o klasöre iner.</p>
          <div className="flex flex-col gap-2 sm:flex-row">
            <Input
              value={mcpRoot}
              onChange={(event) => setMcpRoot(event.target.value)}
              className="cherry-focus font-mono text-[12px]"
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
  const active = card.versions.find((item) => item.id === card.activeVersionId);
  return (
    <article
      className={cn(
        "cherry-surface rounded-[10px] border p-4",
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
      <p className="mt-3 font-mono text-[11px] text-muted-foreground">{active?.note}</p>
      {active?.checkpointRef ? (
        <p className="font-mono text-[11px] text-muted-foreground">{active.checkpointRef}</p>
      ) : null}
    </article>
  );
}
