export function getDeviceFingerprint(): string {
  if (typeof window === "undefined") {
    return "ssr";
  }
  const desktop = window.cherryDesktop;
  if (desktop) {
    return desktop.deviceFingerprint();
  }
  const key = "cherry.device";
  const existing = localStorage.getItem(key);
  if (existing) {
    return existing;
  }
  const created = crypto.randomUUID();
  localStorage.setItem(key, created);
  return created;
}

export function getDeviceLabel(): string {
  if (typeof window === "undefined") {
    return "web";
  }
  const desktop = window.cherryDesktop;
  if (desktop) {
    return desktop.deviceLabel();
  }
  return `Tarayıcı · ${navigator.platform}`;
}
