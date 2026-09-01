const { app, BrowserWindow } = require("electron");

const WEB_URL = process.env.ICERDE_WEB_URL || "http://127.0.0.1:43147";

function createWindow() {
  const win = new BrowserWindow({
    width: 1280,
    height: 800,
    backgroundColor: "#0E1114",
    title: "İçerde",
    webPreferences: {
      preload: require("path").join(__dirname, "preload.cjs"),
      nodeIntegration: false,
      contextIsolation: true,
    },
  });
  void win.loadURL(WEB_URL);
}

app.whenReady().then(() => {
  createWindow();
  app.on("activate", () => {
    if (BrowserWindow.getAllWindows().length === 0) {
      createWindow();
    }
  });
});

app.on("window-all-closed", () => {
  if (process.platform !== "darwin") {
    app.quit();
  }
});
