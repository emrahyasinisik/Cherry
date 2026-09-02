#!/usr/bin/env bash
# Copy OpenCode and Maestro CLIs into vendor/bin for the Icerde installer.
# Does not download blindly. PATH copies only; prints install hints if missing.
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
  cp "$src" "$target"
  chmod +x "$target"
  echo "copied $name -> $target"
  return 0
}

if copy_from_path opencode; then
  true
else
  echo "opencode not on PATH. Developer fallback: install the CLI, then re-run."
  echo "  https://opencode.ai  — customer never installs this separately."
fi

if copy_from_path maestro; then
  true
else
  echo "maestro not on PATH. Install then re-run:"
  echo "  curl -fsSL \"https://get.maestro.mobile.dev\" | bash"
  echo "  Java 17+ and JAVA_HOME required. Customer still does not install Maestro by hand."
fi

echo "look order: ICERDE_*_BIN -> ICERDE_SIDECAR_DIR -> $dest -> PATH"
ls -la "$dest"
