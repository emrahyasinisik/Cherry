const { app, BrowserWindow } = require("electron");
const fs = require("fs");
const path = require("path");
const { spawn } = require("child_process");

const WEB_URL = process.env.CHERRY_WEB_URL || "http://127.0.0.1:43147";

let maestroMcp = null;

function sidecarDir() {
  if (process.env.CHERRY_SIDECAR_DIR) {
    return process.env.CHERRY_SIDECAR_DIR;
  }
  const packed = process.resourcesPath ? path.join(process.resourcesPath, "bin") : "";
  const repoVendor = path.resolve(__dirname, "..", "..", "vendor", "bin");
  const nextToApp = path.join(__dirname, "vendor", "bin");
  for (const dir of [packed, repoVendor, nextToApp]) {
    if (dir && fs.existsSync(dir)) {
      return dir;
    }
  }
  return repoVendor;
}

function binPath(name) {
  const file = process.platform === "win32" ? `${name}.exe` : name;
  const candidate = path.join(sidecarDir(), file);
  if (fs.existsSync(candidate)) {
    return candidate;
  }
  return null;
}

function startMaestroMcp() {
  const fromEnv = (process.env.CHERRY_MAESTRO_BIN || "").trim();
  const bin = fromEnv || binPath("maestro");
  if (!bin) {
    console.log("maestro sidecar missing — MCP host idle (SKIPPED without a device)");
    return;
  }
  maestroMcp = spawn(bin, ["mcp"], {
    stdio: ["pipe", "pipe", "pipe"],
    env: { ...process.env, CHERRY_SIDECAR_DIR: sidecarDir() },
    windowsHide: true,
  });
  maestroMcp.on("error", (err) => {
    console.log("maestro mcp:", err.message);
  });
  maestroMcp.on("exit", () => {
    maestroMcp = null;
  });
}

function stopMaestroMcp() {
  if (!maestroMcp) {
    return;
  }
  try {
    maestroMcp.kill();
  } catch {
    // already gone
  }
  maestroMcp = null;
}

function createWindow() {
  const win = new BrowserWindow({
    width: 1280,
    height: 800,
    backgroundColor: "#0E1114",
    title: "Cherry",
    webPreferences: {
      preload: path.join(__dirname, "preload.cjs"),
      nodeIntegration: false,
      contextIsolation: true,
      sandbox: false,
    },
  });
  void win.loadURL(WEB_URL);
}

app.whenReady().then(() => {
  process.env.CHERRY_SIDECAR_DIR = sidecarDir();
  startMaestroMcp();
  createWindow();
  app.on("activate", () => {
    if (BrowserWindow.getAllWindows().length === 0) {
      createWindow();
    }
  });
});

app.on("before-quit", () => {
  stopMaestroMcp();
});

app.on("window-all-closed", () => {
  if (process.platform !== "darwin") {
    app.quit();
  }
});
