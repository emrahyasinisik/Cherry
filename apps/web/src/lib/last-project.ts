const LAST_PROJECT_KEY = "icerde.lastProject";

export function setLastProjectId(id: string | null): void {
  if (typeof window === "undefined") {
    return;
  }
  if (id) {
    sessionStorage.setItem(LAST_PROJECT_KEY, id);
    return;
  }
  sessionStorage.removeItem(LAST_PROJECT_KEY);
}

export function getLastProjectId(): string | null {
  if (typeof window === "undefined") {
    return null;
  }
  return sessionStorage.getItem(LAST_PROJECT_KEY);
}
