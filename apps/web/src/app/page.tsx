import { Suspense } from "react";

import { LoginForm } from "@/components/login-form";

export default function HomePage() {
  return (
    <div className="cherry-canvas flex min-h-full flex-1 items-center justify-center px-4">
      <div className="cherry-enter cherry-auth-card w-full max-w-[360px] rounded-[10px] border border-border bg-card p-6">
        <p className="text-[22px] leading-7 font-medium tracking-tight">Cherry</p>
        <p className="mt-1 text-muted-foreground">
          Masaüstü stüdyo — mobil uygulama üretir ve dener
        </p>
        <div className="mt-6">
          <Suspense fallback={<p className="text-muted-foreground">Yükleniyor…</p>}>
            <LoginForm />
          </Suspense>
        </div>
        <p className="mt-6 text-center text-[11px] text-muted-foreground">
          KVKK / GDPR · SMS yok · 6 haneli kod birinci parti
        </p>
      </div>
    </div>
  );
}
