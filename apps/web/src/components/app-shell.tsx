"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import type { ReactNode } from "react";
import {
  FolderKanban,
  Bot,
  Smartphone,
  Shield,
  Building2,
  Cpu,
  Scale,
} from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { graphql, setToken, workspaceLabel, type User } from "@/lib/api";
import { cn } from "@/lib/utils";

type NavItem = {
  href: string;
  label: string;
  icon: typeof FolderKanban;
  enabled: boolean;
};

const NAV: NavItem[] = [
  { href: "/projects", label: "Projeler", icon: FolderKanban, enabled: true },
  { href: "/projects", label: "Ajan", icon: Bot, enabled: false },
  { href: "/projects", label: "Maestro", icon: Smartphone, enabled: false },
  { href: "/security", label: "Güvenlik", icon: Shield, enabled: true },
  { href: "/projects", label: "Organizasyon", icon: Building2, enabled: false },
  { href: "/projects", label: "LLM", icon: Cpu, enabled: false },
  { href: "/projects", label: "Gizlilik", icon: Scale, enabled: false },
];

export function AppShell({
  user,
  children,
}: {
  user: User;
  children: ReactNode;
}) {
  const pathname = usePathname();
  const router = useRouter();

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
          <span className="text-[15px] font-medium tracking-tight">İçerde</span>
          <span className="text-muted-foreground">/</span>
          <span className="text-muted-foreground">
            {pathname.startsWith("/security") ? "Güvenlik" : "Projeler"}
          </span>
        </div>
        <div className="flex items-center gap-2">
          <div className="flex overflow-hidden rounded-md border border-border">
            <span className="bg-primary px-2.5 py-1 text-[11px] font-medium text-primary-foreground">
              LLM A
            </span>
            <span className="px-2.5 py-1 text-[11px] text-muted-foreground">
              LLM B
            </span>
          </div>
        </div>
      </header>

      <div className="flex min-h-0 flex-1">
        <aside className="flex w-[220px] shrink-0 flex-col border-r border-border bg-sidebar">
          <nav className="flex flex-1 flex-col gap-0.5 p-2">
            {NAV.map((item) => {
              const Icon = item.icon;
              const active = item.enabled && pathname.startsWith(item.href);
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
                  className={cn(
                    "flex items-center gap-2 rounded-md px-2 py-1.5 transition-colors",
                    active
                      ? "bg-sidebar-accent text-foreground"
                      : "text-muted-foreground hover:bg-sidebar-accent hover:text-foreground",
                  )}
                >
                  <Icon className="size-4" />
                  {item.label}
                </Link>
              );
            })}
          </nav>
          <div className="border-t border-border p-3">
            <p className="truncate text-foreground">{user.email}</p>
            <p className="text-muted-foreground">
              {workspaceLabel(user.workspaceKind)}
            </p>
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
        <span>kayıtlı cihaz</span>
        <span>·</span>
        <span>GDPR katmanı henüz bağlı değil</span>
      </footer>
    </div>
  );
}
