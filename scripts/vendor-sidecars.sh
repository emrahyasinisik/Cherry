#!/usr/bin/env bash
# Vendor OpenCode (and Maestro if on PATH) into vendor/bin for the Icerde installer.
# Binaries are gitignored. The customer never installs these CLIs separately.
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
dest="$root/vendor/bin"
mkdir -p "$dest"

copy_from_path() {
  local name="$1"
  local target="$dest/$name"
  if ! command -v "$name" >/dev/null 2>&1; then
    return 1
  fi
  local src
  src="$(command -v "$name")"
  if [[ "$name" == "maestro" ]]; then
    ln -sfn "$src" "$target"
    echo "linked $name -> $src"
    return 0
  fi
  cp "$src" "$target"
  chmod +x "$target"
  echo "copied $name -> $target"
  return 0
}

opencode_asset() {
  local os arch
  os="$(uname -s)"
  arch="$(uname -m)"
  case "${os}-${arch}" in
    Linux-x86_64 | Linux-amd64) echo "opencode-linux-x64.tar.gz" ;;
    Linux-aarch64 | Linux-arm64) echo "opencode-linux-arm64.tar.gz" ;;
    Darwin-arm64) echo "opencode-darwin-arm64.zip" ;;
    Darwin-x86_64) echo "opencode-darwin-x64.zip" ;;
    MINGW*-* | MSYS*-* | CYGWIN*-*)
      if [[ "$arch" == "ARM64" || "$arch" == "aarch64" ]]; then
        echo "opencode-windows-arm64.zip"
      else
        echo "opencode-windows-x64.zip"
      fi
      ;;
    *)
      echo ""
      ;;
  esac
}

download_opencode() {
  local asset url tmp
  asset="$(opencode_asset)"
  if [[ -z "$asset" ]]; then
    echo "no OpenCode asset for $(uname -s) $(uname -m)"
    return 1
  fi
  url="https://github.com/anomalyco/opencode/releases/latest/download/${asset}"
  tmp="$(mktemp -d)"
  echo "downloading $url"
  curl -fL --retry 3 --retry-delay 2 -o "$tmp/$asset" "$url"
  if [[ "$asset" == *.tar.gz ]]; then
    tar -xzf "$tmp/$asset" -C "$tmp"
  else
    python3 - "$tmp/$asset" "$tmp" <<'PY'
import sys, zipfile
zipfile.ZipFile(sys.argv[1]).extractall(sys.argv[2])
PY
  fi
  local found
  found="$(find "$tmp" -type f \( -name opencode -o -name opencode.exe \) | head -n 1)"
  if [[ -z "$found" ]]; then
    echo "opencode binary missing from archive"
    rm -rf "$tmp"
    return 1
  fi
  local name="opencode"
  if [[ "$found" == *.exe ]]; then
    name="opencode.exe"
  fi
  cp "$found" "$dest/$name"
  chmod +x "$dest/$name"
  rm -rf "$tmp"
  echo "vendored $dest/$name"
}

if copy_from_path opencode || download_opencode; then
  true
else
  echo "opencode not vendored. Set ICERDE_OPENCODE_BIN or install https://opencode.ai"
  echo "Customer still does not install OpenCode by hand — this script is the installer."
fi

if copy_from_path maestro; then
  true
else
  echo "maestro not on PATH. Optional: curl -fsSL \"https://get.maestro.mobile.dev\" | bash"
  echo "Java 17+ and JAVA_HOME required. Customer still does not install Maestro by hand."
fi

echo "look order: ICERDE_*_BIN -> ICERDE_SIDECAR_DIR -> $dest -> PATH"
ls -la "$dest"
