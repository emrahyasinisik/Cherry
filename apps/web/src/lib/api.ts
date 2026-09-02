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
  createdAt: string;
  logs: JobLog[];
  files: ProjectFile[];
  maestro: MaestroStudio;
};

export type ProjectStack = "EXPO" | "FLUTTER" | "NATIVE";
export type ProjectStatus = "QUEUED" | "WRITING" | "TESTING" | "READY" | "FAILED";
export type MaestroResult = "SKIPPED" | "PASSED" | "FAILED";

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

export const PROJECT_FIELDS = `
  id name brief stack status rootPath createdAt
  logs { at message role }
  files { path kind }
  maestro {
    ready
    deviceStatus
    screens { id name html }
    flows { id name yaml result note }
  }
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
};

export type LlmStatus = {
  slot: string;
  versionName: string;
  channel: string;
  gdpr: boolean;
};

export type LlmVersion = {
  id: string;
  name: string;
  note: string;
};

export type LlmSlotCard = {
  slot: string;
  wired: boolean;
  role: string;
  activeVersionId?: string | null;
  versions: LlmVersion[];
};

export type LlmCompletion = {
  at: string;
  purpose: string;
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
  slotA: LlmSlotCard;
  slotB: LlmSlotCard;
  completions: LlmCompletion[];
};

export const LLM_ADMIN_FIELDS = `
  gdpr activeSlot mcpRoot
  slotA { slot wired role activeVersionId versions { id name note } }
  slotB { slot wired role activeVersionId versions { id name note } }
  completions { at purpose versionName channel inputRedactions outputRedactions promptPreview outputPreview }
`;

type GraphQLResponse<T> = {
  data?: T;
  errors?: { message: string }[];
};

const TOKEN_KEY = "icerde.token";

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
  try {
    response = await fetch(`${getApiBase()}/graphql`, {
      method: "POST",
      headers,
      body: JSON.stringify({ query, variables }),
    });
  } catch {
    throw new Error("API’ye ulaşılamadı. Go sunucusunun açık olduğundan emin ol.");
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
      return "Expo / React Native";
    case "FLUTTER":
      return "Flutter";
    case "NATIVE":
      return "Native iOS + Android";
    default: {
      const exhaustive: never = stack;
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
      return "Atlandı (cihaz yok)";
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
