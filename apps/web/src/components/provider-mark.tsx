import type { ConnectionKind } from "@/lib/api";
import { cn } from "@/lib/utils";

type MarkProps = {
  kind: ConnectionKind;
  className?: string;
  title?: string;
};

export function ProviderMark({ kind, className, title }: MarkProps) {
  const label = title ?? kind;
  switch (kind) {
    case "GITHUB":
      return (
        <svg viewBox="0 0 24 24" className={cn("size-8", className)} role="img" aria-label={label}>
          <title>{label}</title>
          <path
            fill="currentColor"
            d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12"
          />
        </svg>
      );
    case "VERCEL":
      return (
        <svg viewBox="0 0 24 24" className={cn("size-8", className)} role="img" aria-label={label}>
          <title>{label}</title>
          <path fill="currentColor" d="M12 2 22 20H2L12 2z" />
        </svg>
      );
    case "SUPABASE":
      return (
        <svg viewBox="0 0 24 24" className={cn("size-8", className)} role="img" aria-label={label}>
          <title>{label}</title>
          <path
            fill="currentColor"
            d="M13.86 2.18a1.2 1.2 0 0 0-2.12.08L4.3 16.4A1.35 1.35 0 0 0 5.47 18.4h5.2v3.42a1.2 1.2 0 0 0 2.12-.08l7.44-14.14A1.35 1.35 0 0 0 19.06 5.6h-5.2V2.18Z"
          />
        </svg>
      );
    case "CLOUDFLARE":
      return (
        <svg viewBox="0 0 24 24" className={cn("size-8", className)} role="img" aria-label={label}>
          <title>{label}</title>
          <path
            fill="currentColor"
            d="M8.2 16.9h11.3c.7 0 1.3-.5 1.4-1.2.3-1.4-.8-2.6-2.2-2.6h-.3c.1-.4.2-.8.2-1.2 0-2.7-2.2-4.9-4.9-4.9-2.2 0-4.1 1.5-4.7 3.5-1.8.2-3.2 1.7-3.2 3.6 0 .5.1 1 .3 1.4H5.1c-1.4 0-2.6 1.2-2.6 2.6 0 .2 0 .3.1.5.2.8 1 1.3 1.8 1.3h4v-3Z"
          />
          <path
            fill="currentColor"
            opacity="0.55"
            d="M6.4 14.4c.4-1.6 1.8-2.8 3.5-2.8.3 0 .6 0 .9.1.7-1.6 2.3-2.7 4.1-2.7 1.9 0 3.5 1.1 4.2 2.8h.7c1.8 0 3.2 1.5 3.2 3.3 0 .2 0 .4-.1.6H8.2c-.7 0-1.4-.5-1.8-1.3Z"
          />
        </svg>
      );
    case "RENDER":
      return (
        <svg viewBox="0 0 24 24" className={cn("size-8", className)} role="img" aria-label={label}>
          <title>{label}</title>
          <rect width="24" height="24" rx="6" fill="currentColor" opacity="0.18" />
          <path
            fill="currentColor"
            d="M8 6.5h5.1c2.6 0 4.4 1.6 4.4 4 0 1.8-1 3.1-2.6 3.7L17.6 17.5h-2.7l-2.4-3.2H10.4v3.2H8V6.5Zm2.4 2v3.1h2.5c1.3 0 2.1-.7 2.1-1.6 0-.9-.8-1.5-2.1-1.5H10.4Z"
          />
        </svg>
      );
    default: {
      const exhaustive: never = kind;
      return exhaustive;
    }
  }
}

export function IcerdeMark({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 24 24" className={cn("size-8", className)} role="img" aria-label="İçerde">
      <title>İçerde</title>
      <rect width="24" height="24" rx="6" fill="#c4a574" />
      <rect x="7" y="6" width="10" height="12" rx="2" fill="none" stroke="#0e1114" strokeWidth="1.6" />
      <rect x="10.2" y="11" width="3.6" height="7" rx="0.8" fill="#0e1114" />
    </svg>
  );
}
