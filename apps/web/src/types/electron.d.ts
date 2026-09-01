export {};

declare global {
  interface Window {
    icerdeDesktop?: {
      platform: string;
      deviceFingerprint: () => string;
      deviceLabel: () => string;
    };
  }
}
