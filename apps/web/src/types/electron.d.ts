export {};

declare global {
  interface Window {
    cherryDesktop?: {
      platform: string;
      deviceFingerprint: () => string;
      deviceLabel: () => string;
    };
  }
}
