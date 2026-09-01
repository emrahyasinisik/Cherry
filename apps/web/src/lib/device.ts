export function getDeviceFingerprint(): string {
  if (typeof window === "undefined") {
    return "ssr";
  }
  const desktop = window.icerdeDesktop;
  if (desktop) {
    return desktop.deviceFingerprint();
  }
  const key = "icerde.device";
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
  const desktop = window.icerdeDesktop;
  if (desktop) {
    return desktop.deviceLabel();
  }
  return `Tarayıcı · ${navigator.platform}`;
}
