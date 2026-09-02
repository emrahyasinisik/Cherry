import { Suspense } from "react";

import { CallbackClient } from "./callback-client";

export default function OAuthCallbackPage() {
  return (
    <Suspense
      fallback={
        <div className="flex min-h-full items-center justify-center text-muted-foreground">
          OAuth tamamlanıyor…
        </div>
      }
    >
      <CallbackClient />
    </Suspense>
  );
}
