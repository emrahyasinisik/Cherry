"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useEffect, useState, type ReactNode } from "react";
import {
  FolderKanban,
  Bot,
  Smartphone,
  Shield,
  Building2,
  Cable,
  Cpu,
  Scale,
} from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { getLastProjectId } from "@/lib/last-project";
import { graphql, llmOccupancyLabel, setToken, workspaceLabel, type LlmStatus, type User } from "@/lib/api";
import { cn } from "@/lib/utils";

type NavItem = {
  href: string;
  label: string;
  icon: typeof FolderKanban;
  enabled: boolean;
};

export function AppShell({
  user,
  title,
  status,
  children,
}: {
  user: User;
  title?: string;
  status?: string;
  children: ReactNode;
}) {
  const pathname = usePathname();
  const router = useRouter();
  const [lastProject, setLastProject] = useState<string | null>(null);
  const [llm, setLlm] = useState<LlmStatus | null>(null);
  const [openCode, setOpenCode] = useState<string | null>(null);
  const [maestro, setMaestro] = useState<string | null>(null);

  useEffect(() => {
    setLastProject(getLastProjectId());
    let cancelled = false;
    function refresh() {
      void graphql<{ llmStatus: LlmStatus }>(
        `query Chip {
          llmStatus { slot versionName channel gdpr queued occupancyA occupancyB versionA versionB }
        }`,
      )
        .then((data) => {
          if (!cancelled) {
            setLlm(data.llmStatus);
          }
        })
        .catch(() => {
          if (!cancelled) {
            setLlm(null);
          }
        });
      void fetch("/health")
        .then((response) => (response.ok ? response.json() : null))
        .then((data: { opencode?: string; maestro?: string } | null) => {
          if (!cancelled) {
            setOpenCode(data?.opencode ?? null);
            setMaestro(data?.maestro ?? null);
          }
        })
        .catch(() => {
          if (!cancelled) {
            setOpenCode(null);
            setMaestro(null);
          }
        });
    }
    refresh();
    const timer = window.setInterval(refresh, 2000);
    return () => {
      cancelled = true;
      window.clearInterval(timer);
    };
  }, [pathname]);
  const maestroHref = lastProject ? `/projects/${lastProject}/maestro` : "/maestro";
  const agentHref = lastProject ? `/projects/${lastProject}` : "/projects";

  const nav: NavItem[] = [
    { href: "/projects", label: "Projeler", icon: FolderKanban, enabled: true },
    { href: agentHref, label: "Ajan", icon: Bot, enabled: true },
    { href: maestroHref, label: "Maestro", icon: Smartphone, enabled: true },
    { href: "/security", label: "Güvenlik", icon: Shield, enabled: true },
    { href: "/projects", label: "Organizasyon", icon: Building2, enabled: false },
    { href: "/connections", label: "Bağlantılar", icon: Cable, enabled: true },
    { href: "/llm", label: "LLM", icon: Cpu, enabled: true },
    { href: "/privacy", label: "Gizlilik", icon: Scale, enabled: true },
  ];

  async function handleLogout() {
    try {
      await graphql<{ logout: boolean }>("mutation { logout }");
    } catch {
      // Session is local; still leave.
    }
    setToken(null);
    router.replace("/");
  }

  return (
    <div className="flex h-full min-h-screen flex-col bg-background">
      <header className="flex h-12 shrink-0 items-center justify-between border-b border-border px-4">
        <div className="flex items-center gap-3">
          <span className="text-[15px] font-medium tracking-tight">Cherry</span>
          <span className="text-muted-foreground">/</span>
          <span className="text-muted-foreground">{title ?? headerTitle(pathname)}</span>
        </div>
          <div className="flex items-center gap-2">
            <div className="flex overflow-hidden rounded-md border border-border">
              <WorkerChip
                label="LLM A"
                version={llm?.versionA}
                occupancy={llm?.occupancyA ?? "IDLE"}
              />
              <WorkerChip
                label="LLM B"
                version={llm?.versionB}
                occupancy={llm?.occupancyB ?? "IDLE"}
              />
            </div>
            {llm && llm.queued > 0 ? (
              <span className="font-mono text-[11px] text-muted-foreground">kuyruk {llm.queued}</span>
            ) : null}
          </div>
      </header>

      <div className="flex min-h-0 flex-1">
        <aside className="flex w-[220px] shrink-0 flex-col border-r border-border bg-sidebar">
          <nav className="flex flex-1 flex-col gap-0.5 p-2">
            {nav.map((item) => {
              const Icon = item.icon;
              const active = navActive(pathname, item.label, item.href);
              if (!item.enabled) {
                return (
                  <span
                    key={item.label}
                    className="flex items-center gap-2 rounded-md px-2 py-1.5 text-muted-foreground/70"
                  >
                    <Icon className="size-4" />
                    {item.label}
                    <Badge variant="outline" className="ml-auto text-[10px]">
                      sonra
                    </Badge>
                  </span>
                );
              }
              return (
                <Link
                  key={item.label}
                  href={item.href}
                  data-active={active}
                  className={cn(
                    "cherry-nav-link flex items-center gap-2 rounded-md py-1.5 pr-2 pl-2.5",
                    active
                      ? "bg-sidebar-accent text-foreground"
                      : "text-muted-foreground hover:bg-sidebar-accent hover:text-foreground",
                  )}
                >
                  <Icon className="size-4 shrink-0" />
                  {item.label}
                </Link>
              );
            })}
          </nav>
          <div className="border-t border-border p-3">
            <p className="truncate text-foreground">{user.email}</p>
            <p className="text-muted-foreground">{workspaceLabel(user.workspaceKind)}</p>
            <Button
              variant="ghost"
              size="sm"
              className="mt-2 w-full justify-start px-2"
              onClick={() => {
                void handleLogout();
              }}
            >
              Çıkış
            </Button>
          </div>
        </aside>
        <main className="min-w-0 flex-1 overflow-auto">{children}</main>
      </div>

      <footer className="flex h-8 shrink-0 items-center gap-3 border-t border-border px-4 font-mono text-[11px] text-muted-foreground">
        <span>1 oturum</span>
        <span>·</span>
        <span>{status ?? "kayıtlı cihaz"}</span>
        <span>·</span>
        <span>{llm?.gdpr ? "GDPR katmanı bağlı" : "GDPR katmanı henüz bağlı değil"}</span>
        {llm ? (
          <>
            <span>·</span>
            <span>{llm.channel}</span>
          </>
        ) : null}
        <span>·</span>
        <span>{sidecarChip("OpenCode", openCode)}</span>
        <span>·</span>
        <span>{sidecarChip("Maestro", maestro)}</span>
      </footer>
    </div>
  );
}

function WorkerChip({
  label,
  version,
  occupancy,
}: {
  label: string;
  version?: string;
  occupancy: LlmStatus["occupancyA"];
}) {
  const busy = occupancy === "BUSY";
  return (
    <span
      className={cn(
        "cherry-chip px-2.5 py-1 text-[11px] font-medium",
        busy ? "bg-primary text-primary-foreground" : "text-muted-foreground",
      )}
    >
      {label}
      {version ? ` · ${version}` : ""}
      {` · ${llmOccupancyLabel(occupancy)}`}
    </span>
  );
}

function sidecarChip(name: string, source: string | null): string {
  switch (source) {
    case "env":
      return `${name} env`;
    case "bundled":
      return `${name} paket`;
    case "path":
      return `${name} PATH`;
    case "missing":
    case null:
      return `${name} yok`;
    default:
      return `${name} ${source}`;
  }
}

function headerTitle(pathname: string): string {
  if (pathname.startsWith("/connections")) {
    return "Bağlantılar";
  }
  if (pathname.startsWith("/llm")) {
    return "LLM yönetici";
  }
  if (pathname.startsWith("/privacy")) {
    return "Gizlilik";
  }
  if (pathname.includes("/maestro")) {
    return "Maestro";
  }
  if (pathname.startsWith("/security")) {
    return "Güvenlik";
  }
  if (pathname.startsWith("/projects/new")) {
    return "Yeni proje";
  }
  if (pathname.startsWith("/projects/") && pathname !== "/projects") {
    return "Ajan";
  }
  return "Projeler";
}

function navActive(pathname: string, label: string, href: string): boolean {
  switch (label) {
    case "Maestro":
      return pathname.includes("/maestro") || pathname === "/maestro";
    case "Ajan":
      return pathname.startsWith("/projects/") && pathname !== "/projects" && pathname !== "/projects/new" && !pathname.includes("/maestro");
    case "Projeler":
      return pathname === "/projects" || pathname === "/projects/new";
    case "Güvenlik":
      return pathname.startsWith("/security");
    case "LLM":
      return pathname.startsWith("/llm");
    case "Gizlilik":
      return pathname.startsWith("/privacy");
    case "Bağlantılar":
      return pathname.startsWith("/connections");
    default:
      return pathname.startsWith(href);
  }
}
