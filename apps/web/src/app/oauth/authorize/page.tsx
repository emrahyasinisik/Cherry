import { Suspense } from "react";

import { AuthorizeClient } from "./authorize-client";

export default function OAuthAuthorizePage() {
  return (
    <Suspense
      fallback={
        <div className="flex min-h-full items-center justify-center bg-[#0d1117] text-[#8b949e]">
          OAuth 2.0 izin ekranı yükleniyor…
        </div>
      }
    >
      <AuthorizeClient />
    </Suspense>
  );
}
