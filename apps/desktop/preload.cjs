const { contextBridge } = require("electron");

contextBridge.exposeInMainWorld("icerdeDesktop", {
  platform: process.platform,
});
