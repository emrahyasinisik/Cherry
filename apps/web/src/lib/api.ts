export type WorkspaceKind = "PERSONAL" | "ORGANIZATION";

export type User = {
  id: string;
  email: string;
  workspaceKind: WorkspaceKind;
};

export type Project = {
  id: string;
  name: string;
  stack: string;
  status: string;
};

export type Health = {
  ok: boolean;
  store: string;
  version: string;
};

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
