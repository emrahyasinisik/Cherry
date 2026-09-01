"use client";

import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";

import { AppShell } from "@/components/app-shell";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  getToken,
  graphql,
  purposeLabel,
  type Device,
  type MailMessage,
  type SessionInfo,
  type User,
} from "@/lib/api";

export default function SecurityPage() {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [devices, setDevices] = useState<Device[]>([]);
  const [sessions, setSessions] = useState<SessionInfo[]>([]);
  const [mailbox, setMailbox] = useState<MailMessage[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [info, setInfo] = useState<string | null>(null);
  const [totpUrl, setTotpUrl] = useState<string | null>(null);
  const [totpSecret, setTotpSecret] = useState<string | null>(null);
  const [totpCode, setTotpCode] = useState("");

  async function load() {
    const data = await graphql<{
      me: User | null;
      devices: Device[];
      sessions: SessionInfo[];
      mailbox: MailMessage[];
    }>(`query Security {
      me { id email workspaceKind totpEnabled }
      devices { id label trusted current lastSeenAt }
      sessions { id current createdAt deviceLabel }
      mailbox { id subject body purpose createdAt }
    }`);
    if (!data.me) {
      router.replace("/");
      return;
    }
    setUser(data.me);
    setDevices(data.devices);
    setSessions(data.sessions);
    setMailbox(data.mailbox);
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
    // eslint-disable-next-line react-hooks/exhaustive-deps -- fetch on mount
  }, [router]);

  if (error) {
    return (
      <div className="flex min-h-full items-center justify-center text-destructive">
        {error}
      </div>
    );
  }
  if (!user) {
    return (
      <div className="flex min-h-full items-center justify-center text-muted-foreground">
        Yükleniyor…
      </div>
    );
  }

  return (
    <AppShell user={user}>
      <div className="mx-auto flex max-w-3xl flex-col gap-8 p-8">
        <div>
          <h1 className="text-base font-medium">Güvenlik</h1>
          <p className="text-muted-foreground">
            X-inspired: cihaz, oturum, TOTP, geçici kutu. SMS yok.
          </p>
        </div>
        {info ? <p className="text-muted-foreground">{info}</p> : null}

        <section className="flex flex-col gap-3">
          <h2 className="font-medium">Authenticator (TOTP)</h2>
          <p className="text-muted-foreground">
            Durum: {user.totpEnabled ? "açık" : "kapalı"}
          </p>
          {!user.totpEnabled ? (
            <div className="flex flex-col gap-2">
              <Button
                type="button"
                variant="outline"
                onClick={() => {
                  void graphql<{ enableTotp: { secret: string; otpauthUrl: string } }>(
                    `mutation { enableTotp { secret otpauthUrl } }`,
                  ).then((data) => {
                    setTotpSecret(data.enableTotp.secret);
                    setTotpUrl(data.enableTotp.otpauthUrl);
                    setInfo("Authenticator’a ekle, sonra 6 haneyi onayla.");
                  }).catch((err: unknown) => {
                    setError(err instanceof Error ? err.message : "TOTP açılamadı.");
                  });
                }}
              >
                TOTP kur
              </Button>
              {totpSecret ? (
                <div className="rounded-md border border-border p-3 font-mono text-[11px]">
                  <p>{totpSecret}</p>
                  {totpUrl ? <p className="mt-2 break-all text-muted-foreground">{totpUrl}</p> : null}
                  <div className="mt-2 flex gap-2">
                    <Input
                      value={totpCode}
                      onChange={(event) => setTotpCode(event.target.value)}
                      className="h-9 font-mono"
                      placeholder="123456"
                    />
                    <Button
                      type="button"
                      onClick={() => {
                        void graphql<{ confirmTotp: boolean }>(
                          `mutation Confirm($code: String!) { confirmTotp(code: $code) }`,
                          { code: totpCode },
                        ).then(() => {
                          setInfo("TOTP açıldı.");
                          setTotpSecret(null);
                          setTotpUrl(null);
                          void load();
                        }).catch((err: unknown) => {
                          setError(err instanceof Error ? err.message : "Onay başarısız.");
                        });
                      }}
                    >
                      Onayla
                    </Button>
                  </div>
                </div>
              ) : null}
            </div>
          ) : (
            <div className="flex gap-2">
              <Input
                value={totpCode}
                onChange={(event) => setTotpCode(event.target.value)}
                className="h-9 max-w-40 font-mono"
                placeholder="kapatmak için kod"
              />
              <Button
                type="button"
                variant="destructive"
                onClick={() => {
                  void graphql<{ disableTotp: boolean }>(
                    `mutation Disable($code: String!) { disableTotp(code: $code) }`,
                    { code: totpCode },
                  ).then(() => {
                    setInfo("TOTP kapatıldı.");
                    setTotpCode("");
                    void load();
                  }).catch((err: unknown) => {
                    setError(err instanceof Error ? err.message : "Kapatılamadı.");
                  });
                }}
              >
                TOTP kapat
              </Button>
            </div>
          )}
        </section>

        <section className="flex flex-col gap-3">
          <h2 className="font-medium">Cihazlar</h2>
          <ul className="divide-y divide-border rounded-[10px] border border-border">
            {devices.map((device) => (
              <li key={device.id} className="flex items-center justify-between gap-3 px-4 py-3">
                <div>
                  <p>
                    {device.label}
                    {device.current ? " · bu cihaz" : ""}
                    {device.trusted ? " · güvenilir" : ""}
                  </p>
                  <p className="font-mono text-[11px] text-muted-foreground">{device.lastSeenAt}</p>
                </div>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  className="text-destructive"
                  onClick={() => {
                    void graphql<{ revokeDevice: boolean }>(
                      `mutation Revoke($id: ID!) { revokeDevice(id: $id) }`,
                      { id: device.id },
                    ).then(() => {
                      void load();
                    });
                  }}
                >
                  Güveni kaldır
                </Button>
              </li>
            ))}
          </ul>
        </section>

        <section className="flex flex-col gap-3">
          <h2 className="font-medium">Oturumlar</h2>
          <p className="text-muted-foreground">Aynı anda tek aktif oturum.</p>
          <ul className="divide-y divide-border rounded-[10px] border border-border">
            {sessions.map((session) => (
              <li key={session.id} className="flex items-center justify-between gap-3 px-4 py-3">
                <div>
                  <p>
                    {session.deviceLabel}
                    {session.current ? " · şimdi" : ""}
                  </p>
                  <p className="font-mono text-[11px] text-muted-foreground">{session.createdAt}</p>
                </div>
                {!session.current ? (
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    className="text-destructive"
                    onClick={() => {
                      void graphql<{ revokeSession: boolean }>(
                        `mutation Revoke($id: ID!) { revokeSession(id: $id) }`,
                        { id: session.id },
                      ).then(() => {
                        void load();
                      });
                    }}
                  >
                    İptal
                  </Button>
                ) : null}
              </li>
            ))}
          </ul>
        </section>

        <section className="flex flex-col gap-3">
          <h2 className="font-medium">Geçici kutu</h2>
          {mailbox.length === 0 ? (
            <p className="text-muted-foreground">Kutu boş.</p>
          ) : (
            <ul className="flex flex-col gap-3">
              {mailbox.map((item) => (
                <li key={item.id} className="rounded-[10px] border border-border p-4">
                  <p className="font-medium">
                    {item.subject} · {purposeLabel(item.purpose)}
                  </p>
                  <pre className="mt-2 overflow-auto font-mono text-[11px] whitespace-pre-wrap text-muted-foreground">
                    {item.body}
                  </pre>
                </li>
              ))}
            </ul>
          )}
        </section>
      </div>
    </AppShell>
  );
}
