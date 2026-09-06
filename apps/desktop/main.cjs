const { app, BrowserWindow, Tray, Menu, dialog, nativeImage } = require("electron");
const fs = require("fs");
const http = require("http");
const path = require("path");
const { spawn } = require("child_process");

const WEB_PORT = process.env.CHERRY_WEB_PORT || "43147";
const API_ADDR = process.env.CHERRY_API_ADDR || "127.0.0.1:43148";

let maestroMcp = null;
let apiChild = null;
let webChild = null;
let tray = null;
let quitting = false;

function isPackaged() {
  return app.isPackaged;
}

function resource(...parts) {
  const base = isPackaged() ? process.resourcesPath : path.resolve(__dirname, "resources");
  return path.join(base, ...parts);
}

function sidecarDir() {
  if (process.env.CHERRY_SIDECAR_DIR) {
    return process.env.CHERRY_SIDECAR_DIR;
  }
  const packed = path.join(process.resourcesPath || "", "bin");
  const localResources = path.resolve(__dirname, "resources", "bin");
  const repoVendor = path.resolve(__dirname, "..", "..", "vendor", "bin");
  for (const dir of [packed, localResources, repoVendor]) {
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
  const maestroDist = path.join(
    isPackaged() ? process.resourcesPath : path.resolve(__dirname, "resources"),
    "maestro-dist",
    "bin",
    process.platform === "win32" ? "maestro.bat" : "maestro",
  );
  if (name === "maestro" && fs.existsSync(maestroDist)) {
    return maestroDist;
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

function stopChild(child) {
  if (!child) {
    return;
  }
  try {
    if (process.platform === "win32") {
      spawn("taskkill", ["/pid", String(child.pid), "/f", "/t"], { windowsHide: true });
    } else {
      child.kill("SIGTERM");
    }
  } catch {
    // already gone
  }
}

function stopAll() {
  stopChild(maestroMcp);
  maestroMcp = null;
  stopChild(webChild);
  webChild = null;
  stopChild(apiChild);
  apiChild = null;
}

function waitHttp(url, tries) {
  return new Promise((resolve, reject) => {
    let left = tries;
    const tick = () => {
      const req = http.get(url, (res) => {
        res.resume();
        if (res.statusCode && res.statusCode < 500) {
          resolve();
          return;
        }
        retry();
      });
      req.on("error", retry);
      req.setTimeout(1500, () => {
        req.destroy();
        retry();
      });
    };
    const retry = () => {
      left -= 1;
      if (left <= 0) {
        reject(new Error(`timeout waiting for ${url}`));
        return;
      }
      setTimeout(tick, 400);
    };
    tick();
  });
}

function spawnLogged(bin, args, opts, label) {
  const child = spawn(bin, args, opts);
  child.stdout?.on("data", (buf) => {
    console.log(`[${label}] ${buf.toString().trimEnd()}`);
  });
  child.stderr?.on("data", (buf) => {
    console.log(`[${label}] ${buf.toString().trimEnd()}`);
  });
  child.on("error", (err) => {
    console.log(`[${label}] spawn error: ${err.message}`);
  });
  child.on("exit", (code, signal) => {
    console.log(`[${label}] exit code=${code} signal=${signal}`);
    if (!quitting && isPackaged() && (label === "api" || label === "web")) {
      dialog.showErrorBox("Cherry", `${label} kapandı (${code ?? signal}). Stüdyo durdu.`);
      app.quit();
    }
  });
  return child;
}

function startPackagedServices() {
  const userData = app.getPath("userData");
  const projectsRoot = path.join(userData, "projects");
  fs.mkdirSync(projectsRoot, { recursive: true });

  const oc = binPath("opencode");
  const ma = binPath("maestro");
  const env = {
    ...process.env,
    CHERRY_API_ADDR: API_ADDR,
    CHERRY_WEB_ORIGIN: `http://127.0.0.1:${WEB_PORT}`,
    CHERRY_WEB_URL: `http://127.0.0.1:${WEB_PORT}`,
    CHERRY_SIDECAR_DIR: sidecarDir(),
    CHERRY_PROJECTS_ROOT: projectsRoot,
    CHERRY_COLAB_DIR: resource("colab"),
  };
  if (oc) {
    env.CHERRY_OPENCODE_BIN = oc;
  }
  if (ma) {
    env.CHERRY_MAESTRO_BIN = ma;
  }

  const apiName = process.platform === "win32" ? "cherry-api.exe" : "cherry-api";
  const apiBin = resource("api", apiName);
  if (!fs.existsSync(apiBin)) {
    throw new Error(`API binary missing: ${apiBin}`);
  }
  apiChild = spawnLogged(apiBin, [], { env, windowsHide: true, stdio: ["ignore", "pipe", "pipe"] }, "api");

  const serverJs = resource("web", "server.js");
  if (!fs.existsSync(serverJs)) {
    throw new Error(`Web server missing: ${serverJs}`);
  }
  webChild = spawnLogged(
    process.execPath,
    [serverJs],
    {
      cwd: path.dirname(serverJs),
      env: {
        ...env,
        ELECTRON_RUN_AS_NODE: "1",
        PORT: WEB_PORT,
        HOSTNAME: "127.0.0.1",
      },
      windowsHide: true,
      stdio: ["ignore", "pipe", "pipe"],
    },
    "web",
  );
}

function createWindow() {
  const win = new BrowserWindow({
    width: 1280,
    height: 800,
    minWidth: 960,
    minHeight: 640,
    backgroundColor: "#0E1114",
    title: "Cherry",
    show: false,
    webPreferences: {
      preload: path.join(__dirname, "preload.cjs"),
      nodeIntegration: false,
      contextIsolation: true,
      sandbox: false,
    },
  });
  win.once("ready-to-show", () => {
    win.show();
  });
  void win.loadURL(`http://127.0.0.1:${WEB_PORT}`);
}

function createTray() {
  const iconFile = path.join(__dirname, "build", "icon.png");
  let image = nativeImage.createEmpty();
  if (fs.existsSync(iconFile)) {
    image = nativeImage.createFromPath(iconFile).resize({ width: 16, height: 16 });
  }
  tray = new Tray(image);
  tray.setToolTip("Cherry");
  tray.setContextMenu(
    Menu.buildFromTemplate([
      {
        label: "Cherry",
        click: () => {
          const win = BrowserWindow.getAllWindows()[0];
          if (win) {
            win.show();
          } else {
            createWindow();
          }
        },
      },
      { type: "separator" },
      {
        label: "Çıkış",
        click: () => {
          app.quit();
        },
      },
    ]),
  );
}

app.whenReady().then(async () => {
  process.env.CHERRY_SIDECAR_DIR = sidecarDir();
  try {
    if (isPackaged()) {
      startPackagedServices();
      await waitHttp(`http://127.0.0.1:${WEB_PORT}`, 80);
    } else if (process.env.CHERRY_PACKAGED_TEST === "1") {
      startPackagedServices();
      await waitHttp(`http://127.0.0.1:${WEB_PORT}`, 80);
    }
  } catch (err) {
    dialog.showErrorBox("Cherry", err instanceof Error ? err.message : String(err));
    app.quit();
    return;
  }
  startMaestroMcp();
  createTray();
  createWindow();
  app.on("activate", () => {
    if (BrowserWindow.getAllWindows().length === 0) {
      createWindow();
    }
  });
});

app.on("before-quit", () => {
  quitting = true;
  stopAll();
});

app.on("window-all-closed", () => {
  if (process.platform !== "darwin") {
    app.quit();
  }
});
