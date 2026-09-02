const { contextBridge } = require("electron");
const os = require("os");
const crypto = require("crypto");

function deviceFingerprint() {
  const raw = [os.hostname(), os.userInfo().username, process.platform, os.arch()].join("|");
  return crypto.createHash("sha256").update(raw).digest("hex");
}

function deviceLabel() {
  return `${os.hostname()} (${process.platform})`;
}

contextBridge.exposeInMainWorld("cherryDesktop", {
  platform: process.platform,
  deviceFingerprint,
  deviceLabel,
});
