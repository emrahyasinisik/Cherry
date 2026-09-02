import {
  connectionKindLabel,
  type ConnectionKind,
} from "@/lib/api";

export type ConsentTheme = {
  bg: string;
  card: string;
  text: string;
  muted: string;
  border: string;
  authorizeBg: string;
  authorizeFg: string;
  host: string;
};

export function consentTheme(kind: ConnectionKind): ConsentTheme {
  switch (kind) {
    case "GITHUB":
      return {
        bg: "#0d1117",
        card: "#161b22",
        text: "#e6edf3",
        muted: "#8b949e",
        border: "#30363d",
        authorizeBg: "#238636",
        authorizeFg: "#ffffff",
        host: "github.com",
      };
    case "VERCEL":
      return {
        bg: "#000000",
        card: "#111111",
        text: "#fafafa",
        muted: "#a1a1a1",
        border: "#333333",
        authorizeBg: "#ffffff",
        authorizeFg: "#000000",
        host: "vercel.com",
      };
    case "SUPABASE":
      return {
        bg: "#1c1c1c",
        card: "#171717",
        text: "#ededed",
        muted: "#8d8d8d",
        border: "#2e2e2e",
        authorizeBg: "#3ecf8e",
        authorizeFg: "#1c1c1c",
        host: "supabase.com",
      };
    case "CLOUDFLARE":
      return {
        bg: "#161616",
        card: "#1d1d1d",
        text: "#f3f3f3",
        muted: "#9a9a9a",
        border: "#3a3a3a",
        authorizeBg: "#f38020",
        authorizeFg: "#161616",
        host: "dash.cloudflare.com",
      };
    case "RENDER":
      return {
        bg: "#160c27",
        card: "#1e1233",
        text: "#f4f0ff",
        muted: "#b5a8cc",
        border: "#3b2a55",
        authorizeBg: "#46e3b6",
        authorizeFg: "#160c27",
        host: "dashboard.render.com",
      };
    default: {
      const exhaustive: never = kind;
      return exhaustive;
    }
  }
}

export function oauthPermissions(kind: ConnectionKind): string[] {
  switch (kind) {
    case "GITHUB":
      return [
        "Depolarını oku ve yaz (repo)",
        "Profilini oku (read:user)",
        "E-posta adresini oku (user:email)",
      ];
    case "VERCEL":
      return ["Hesap bilgini oku", "Projelerine deploy et"];
    case "SUPABASE":
      return ["Organizasyonlarını oku", "Projelerini oku"];
    case "CLOUDFLARE":
      return ["Workers scriptlerini düzenle", "D1 veritabanlarını düzenle", "R2 bucket’larını düzenle"];
    case "RENDER":
      return ["Servislerini oluştur ve güncelle"];
    default: {
      const exhaustive: never = kind;
      return exhaustive;
    }
  }
}

export function oauthHeadline(kind: ConnectionKind): string {
  return `İçerde, ${connectionKindLabel(kind)} hesabına erişmek istiyor`;
}

export function providerPurpose(kind: ConnectionKind): string {
  switch (kind) {
    case "SUPABASE":
      return "Müşteri backend’i bu projede durur";
    case "CLOUDFLARE":
      return "Workers, D1 ve R2 — senin hesabın";
    case "GITHUB":
      return "Gelişen projeyi kendi reposuna gönder";
    case "VERCEL":
      return "Frontend’i kendi hesabına deploy et";
    case "RENDER":
      return "Servisi kendi hesabına deploy et";
    default: {
      const exhaustive: never = kind;
      return exhaustive;
    }
  }
}

export function tileClass(kind: ConnectionKind): string {
  switch (kind) {
    case "GITHUB":
      return "border-[#30363d] bg-[#0d1117] text-[#e6edf3]";
    case "VERCEL":
      return "border-[#333] bg-[#000] text-[#fafafa]";
    case "SUPABASE":
      return "border-[#2e2e2e] bg-[#171717] text-[#3ecf8e]";
    case "CLOUDFLARE":
      return "border-[#3a2a18] bg-[#1a1510] text-[#f38020]";
    case "RENDER":
      return "border-[#3b2a55] bg-[#160c27] text-[#46e3b6]";
    default: {
      const exhaustive: never = kind;
      return exhaustive;
    }
  }
}

export function isConnectionKind(value: string | null): value is ConnectionKind {
  switch (value) {
    case "SUPABASE":
    case "CLOUDFLARE":
    case "GITHUB":
    case "VERCEL":
    case "RENDER":
      return true;
    default:
      return false;
  }
}

export function markClass(kind: ConnectionKind): string {
  switch (kind) {
    case "GITHUB":
      return "text-[#e6edf3]";
    case "VERCEL":
      return "text-white";
    case "SUPABASE":
      return "text-[#3ecf8e]";
    case "CLOUDFLARE":
      return "text-[#f38020]";
    case "RENDER":
      return "text-[#46e3b6]";
    default: {
      const exhaustive: never = kind;
      return exhaustive;
    }
  }
}
