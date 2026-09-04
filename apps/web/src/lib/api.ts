export type WorkspaceKind = "PERSONAL" | "ORGANIZATION";
export type LoginNext = "SESSION" | "DEVICE_CODE" | "TOTP";
export type EmailChannel = "inbox" | "smtp" | "resend" | "";
export type VerifyPurpose =
  | "NEW_DEVICE"
  | "LOGIN_CHALLENGE"
  | "EMAIL_VERIFY"
  | "SUSPICIOUS_LOGIN";

export type User = {
  id: string;
  email: string;
  workspaceKind: WorkspaceKind;
  totpEnabled: boolean;
};

export type LoginResult = {
  next: LoginNext;
  token?: string | null;
  challengeId?: string | null;
  user?: User | null;
  emailSent: boolean;
  emailChannel: EmailChannel;
};

export type Project = {
  id: string;
  name: string;
  brief: string;
  stack: ProjectStack;
  status: ProjectStatus;
  rootPath: string;
  backendTarget: BackendTarget;
  createdAt: string;
  logs: JobLog[];
  files: ProjectFile[];
  maestro: MaestroStudio;
  activate: LocalActivate;
};

export type ProjectStack = "EXPO" | "FLUTTER" | "NATIVE";
export type BackendTarget = "LOCAL" | "SUPABASE" | "CLOUDFLARE" | "RENDER";
export type ConnectionKind = "SUPABASE" | "CLOUDFLARE" | "GITHUB" | "VERCEL" | "RENDER";
export type ConnectionStatus = "DISCONNECTED" | "CONNECTED" | "FAILED";
export type ProjectStatus = "QUEUED" | "WRITING" | "TESTING" | "READY" | "FAILED";
export type MaestroResult = "SKIPPED" | "PASSED" | "FAILED";
export type ActivateStatus = "IDLE" | "STARTING" | "RUNNING" | "STOPPING" | "FAILED";

export type ChatRole = "USER" | "AGENT" | "SYSTEM";

export type JobLog = {
  at: string;
  message: string;
  role: ChatRole;
};

export type ProjectFile = {
  path: string;
  kind: string;
};

export type DesignScreen = {
  id: string;
  name: string;
  html: string;
};

export type MaestroFlow = {
  id: string;
  name: string;
  yaml: string;
  result: MaestroResult;
  note: string;
};

export type MaestroStudio = {
  ready: boolean;
  deviceStatus: string;
  screens: DesignScreen[];
  flows: MaestroFlow[];
};

export type LocalActivate = {
  status: ActivateStatus;
  url?: string | null;
  port?: number | null;
  pid?: number | null;
  note: string;
};

export const PROJECT_FIELDS = `
  id name brief stack backendTarget status rootPath createdAt
  logs { at message role }
  files { path kind }
  maestro {
    ready
    deviceStatus
    screens { id name html }
    flows { id name yaml result note }
  }
  activate { status url port pid note }
`;

export type Device = {
  id: string;
  label: string;
  trusted: boolean;
  current: boolean;
  lastSeenAt: string;
};

export type SessionInfo = {
  id: string;
  current: boolean;
  createdAt: string;
  deviceLabel: string;
};

export type MailMessage = {
  id: string;
  subject: string;
  body: string;
  purpose: VerifyPurpose;
  createdAt: string;
};

export type Health = {
  ok: boolean;
  store: string;
  version: string;
  mail: string;
  gdpr: boolean;
  llm: string;
  opencode?: string;
  maestro?: string;
};

export type LlmOccupancy = "IDLE" | "BUSY";

export type LlmStatus = {
  slot: string;
  versionName: string;
  channel: string;
  gdpr: boolean;
  queued: number;
  occupancyA: LlmOccupancy;
  occupancyB: LlmOccupancy;
  versionA: string;
  versionB: string;
};

export type LlmVersion = {
  id: string;
  name: string;
  note: string;
  checkpointRef: string;
};

export type LlmSlotCard = {
  slot: string;
  wired: boolean;
  role: string;
  occupancy: LlmOccupancy;
  activeVersionId?: string | null;
  versions: LlmVersion[];
};

export type LlmCompletion = {
  at: string;
  purpose: string;
  slot: string;
  versionName: string;
  channel: string;
  inputRedactions: number;
  outputRedactions: number;
  promptPreview: string;
  outputPreview: string;
};

export type LlmAdmin = {
  gdpr: boolean;
  activeSlot: string;
  mcpRoot: string;
  queued: number;
  slotA: LlmSlotCard;
  slotB: LlmSlotCard;
  completions: LlmCompletion[];
};

export type TrainingPack = {
  schema: string;
  filename: string;
  json: string;
  jsonl: string;
  liveExamples: number;
  seedExamples: number;
  note: string;
};

export type ColabBridgeStatus = "IDLE" | "STARTING" | "RUNNING" | "STOPPING" | "FAILED";

export type ColabBridge = {
  status: ColabBridgeStatus;
  publicUrl?: string | null;
  localUrl?: string | null;
  token?: string | null;
  tokenHint: string;
  cloudflared: string;
  startedAt?: string | null;
  note: string;
};

export type ColabInferenceStatus = "OFF" | "CONNECTED" | "DISCONNECTED" | "CHECKING";

export type ColabInference = {
  url: string;
  status: ColabInferenceStatus;
  note: string;
};

export const COLAB_INFERENCE_FIELDS = `
  url status note
`;

export function colabInferenceStatusLabel(status: ColabInferenceStatus): string {
  switch (status) {
    case "OFF":
      return "kapalı";
    case "CONNECTED":
      return "bağlı";
    case "DISCONNECTED":
      return "koptu";
    case "CHECKING":
      return "kontrol";
    default: {
      const exhaustive: never = status;
      return exhaustive;
    }
  }
}

export const COLAB_BRIDGE_FIELDS = `
  status publicUrl localUrl token tokenHint cloudflared startedAt note
`;

export function colabBridgeStatusLabel(status: ColabBridgeStatus): string {
  switch (status) {
    case "IDLE":
      return "kapalı";
    case "STARTING":
      return "açılıyor";
    case "RUNNING":
      return "açık";
    case "STOPPING":
      return "kapanıyor";
    case "FAILED":
      return "hata";
    default: {
      const exhaustive: never = status;
      return exhaustive;
    }
  }
}

export const LLM_ADMIN_FIELDS = `
  gdpr activeSlot mcpRoot queued
  slotA { slot wired role occupancy activeVersionId versions { id name note checkpointRef } }
  slotB { slot wired role occupancy activeVersionId versions { id name note checkpointRef } }
  completions { at purpose slot versionName channel inputRedactions outputRedactions promptPreview outputPreview }
`;

export function llmOccupancyLabel(occupancy: LlmOccupancy): string {
  switch (occupancy) {
    case "IDLE":
      return "boş";
    case "BUSY":
      return "meşgul";
    default: {
      const exhaustive: never = occupancy;
      return exhaustive;
    }
  }
}

type GraphQLResponse<T> = {
  data?: T;
  errors?: { message: string }[];
};

const TOKEN_KEY = "cherry.token";

export function getApiBase(): string {
  if (process.env.NEXT_PUBLIC_API_URL) {
    return process.env.NEXT_PUBLIC_API_URL;
  }
  if (typeof window !== "undefined") {
    return "";
  }
  return "http://127.0.0.1:43148";
}

export function getToken(): string | null {
  if (typeof window === "undefined") {
    return null;
  }
  return sessionStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string | null): void {
  if (typeof window === "undefined") {
    return;
  }
  if (token) {
    sessionStorage.setItem(TOKEN_KEY, token);
    return;
  }
  sessionStorage.removeItem(TOKEN_KEY);
}

export async function graphql<T>(
  query: string,
  variables?: Record<string, unknown>,
): Promise<T> {
  const token = getToken();
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }

  let response: Response;
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), 20000);
  try {
    response = await fetch(`${getApiBase()}/graphql`, {
      method: "POST",
      headers,
      body: JSON.stringify({ query, variables }),
      signal: controller.signal,
    });
  } catch (err) {
    if (err instanceof DOMException && err.name === "AbortError") {
      throw new Error("API yanıt vermedi. Biraz bekle veya kodu yeniden iste.");
    }
    throw new Error("API’ye ulaşılamadı. Go sunucusunun açık olduğundan emin ol.");
  } finally {
    clearTimeout(timer);
  }

  const payload = (await response.json()) as GraphQLResponse<T>;
  if (payload.errors?.length) {
    throw new Error(payload.errors[0]?.message ?? "İstek başarısız.");
  }
  if (!payload.data) {
    throw new Error("Boş yanıt.");
  }
  return payload.data;
}

export function workspaceLabel(kind: WorkspaceKind): string {
  switch (kind) {
    case "PERSONAL":
      return "Kişisel";
    case "ORGANIZATION":
      return "Organizasyon";
    default: {
      const exhaustive: never = kind;
      return exhaustive;
    }
  }
}

export function purposeLabel(purpose: VerifyPurpose): string {
  switch (purpose) {
    case "NEW_DEVICE":
      return "Yeni cihaz";
    case "LOGIN_CHALLENGE":
      return "Giriş doğrulama";
    case "EMAIL_VERIFY":
      return "E-posta doğrulama";
    case "SUSPICIOUS_LOGIN":
      return "Şüpheli giriş";
    default: {
      const exhaustive: never = purpose;
      return exhaustive;
    }
  }
}

export function stackLabel(stack: ProjectStack): string {
  switch (stack) {
    case "EXPO":
      return "Expo";
    case "FLUTTER":
      return "Flutter";
    case "NATIVE":
      return "SwiftUI";
    default: {
      const exhaustive: never = stack;
      return exhaustive;
    }
  }
}

export function stackSourceHint(stack: ProjectStack): string {
  switch (stack) {
    case "EXPO":
      return "SDK 57 · TypeScript · Clean Architecture";
    case "FLUTTER":
      return "3.47 · Dart 3.13 · Clean Architecture";
    case "NATIVE":
      return "Swift 6 · iOS 18 · Clean Architecture";
    default: {
      const exhaustive: never = stack;
      return exhaustive;
    }
  }
}

export type ConnectionAuth = "NONE" | "OAUTH" | "TOKEN";
export type OAuthMode = "CONSENT" | "PROVIDER";

export type Connection = {
  kind: ConnectionKind;
  status: ConnectionStatus;
  account: string;
  tokenHint: string;
  note: string;
  authMethod: ConnectionAuth;
  scopes: string[];
  oauthMode: OAuthMode;
};

export function connectionKindLabel(kind: ConnectionKind): string {
  switch (kind) {
    case "SUPABASE":
      return "Supabase";
    case "CLOUDFLARE":
      return "Cloudflare";
    case "GITHUB":
      return "GitHub";
    case "VERCEL":
      return "Vercel";
    case "RENDER":
      return "Render";
    default: {
      const exhaustive: never = kind;
      return exhaustive;
    }
  }
}

export function connectionStatusLabel(status: ConnectionStatus): string {
  switch (status) {
    case "DISCONNECTED":
      return "Bağlı değil";
    case "CONNECTED":
      return "Bağlı";
    case "FAILED":
      return "Hata";
    default: {
      const exhaustive: never = status;
      return exhaustive;
    }
  }
}

export function backendTargetLabel(target: BackendTarget): string {
  switch (target) {
    case "LOCAL":
      return "Yerel API";
    case "SUPABASE":
      return "Supabase";
    case "CLOUDFLARE":
      return "Cloudflare";
    case "RENDER":
      return "Render";
    default: {
      const exhaustive: never = target;
      return exhaustive;
    }
  }
}

export function connectionAuthLabel(method: ConnectionAuth): string {
  switch (method) {
    case "NONE":
      return "";
    case "OAUTH":
      return "OAuth 2.0";
    case "TOKEN":
      return "Token";
    default: {
      const exhaustive: never = method;
      return exhaustive;
    }
  }
}

export function oauthModeLabel(mode: OAuthMode): string {
  switch (mode) {
    case "CONSENT":
      return "Yerel izin ekranı";
    case "PROVIDER":
      return "Gerçek OAuth";
    default: {
      const exhaustive: never = mode;
      return exhaustive;
    }
  }
}

export function connectionTokenHint(kind: ConnectionKind): string {
  switch (kind) {
    case "SUPABASE":
      return "Service role veya anon key";
    case "CLOUDFLARE":
      return "API token";
    case "GITHUB":
      return "repo kapsamlı PAT";
    case "VERCEL":
      return "Deploy token";
    case "RENDER":
      return "API key";
    default: {
      const exhaustive: never = kind;
      return exhaustive;
    }
  }
}

export function projectStatusLabel(status: ProjectStatus): string {
  switch (status) {
    case "QUEUED":
      return "Kuyrukta";
    case "WRITING":
      return "Arka planda yazılıyor";
    case "TESTING":
      return "Test aşaması";
    case "READY":
      return "Hazır";
    case "FAILED":
      return "Durdu";
    default: {
      const exhaustive: never = status;
      return exhaustive;
    }
  }
}

export function maestroResultLabel(result: MaestroResult): string {
  switch (result) {
    case "SKIPPED":
      return "Atlandı";
    case "PASSED":
      return "Geçti";
    case "FAILED":
      return "Kaldı";
    default: {
      const exhaustive: never = result;
      return exhaustive;
    }
  }
}

export function maestroDeviceLabel(deviceStatus: string): string {
  switch (deviceStatus) {
    case "device":
      return "Cihaz görüldü. Koşu gerçek sonuç yazar; PASSED uydurulmaz.";
    case "no_cli":
      return "Maestro CLI yok. Koşu SKIPPED — cihaz olsa da CLI olmadan test yok.";
    case "none":
      return "Cihaz yok. Koşu SKIPPED olur — geçti sayılmaz.";
    default:
      return "Maestro durumu bilinmiyor. SKIPPED olabilir.";
  }
}

export function activateStatusLabel(status: ActivateStatus): string {
  switch (status) {
    case "IDLE":
      return "Kapalı";
    case "STARTING":
      return "Kalkıyor";
    case "RUNNING":
      return "Çalışıyor";
    case "STOPPING":
      return "Durduruluyor";
    case "FAILED":
      return "Hata";
    default: {
      const exhaustive: never = status;
      return exhaustive;
    }
  }
}
