#!/usr/bin/env bash
# Build Cherry desktop installers (Windows / macOS). Run from a Win or Mac machine
# for native installers; this script also prepares resources and can smoke-test
# an unpacked Linux dir on CI.
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

os="$(uname -s)"
arch="$(uname -m)"
goos=""
goarch=""
api_name="cherry-api"
eb_target=""

case "${os}-${arch}" in
  Linux-x86_64 | Linux-amd64)
    goos=linux
    goarch=amd64
    eb_target="--linux dir"
    ;;
  Darwin-arm64)
    goos=darwin
    goarch=arm64
    eb_target="--mac dmg zip"
    ;;
  Darwin-x86_64)
    goos=darwin
    goarch=amd64
    eb_target="--mac dmg zip"
    ;;
  MINGW*-* | MSYS*-* | CYGWIN*-* | Windows_NT-*)
    goos=windows
    goarch=amd64
    api_name="cherry-api.exe"
    eb_target="--win nsis zip"
    ;;
  *)
    echo "unsupported pack host: $os $arch"
    exit 1
    ;;
esac

if [[ ! -x vendor/bin/opencode && ! -f vendor/bin/opencode.exe ]]; then
  echo "vendoring sidecars…"
  ./scripts/vendor-sidecars.sh
fi

python3 scripts/generate-desktop-icon.py

echo "building web (Next standalone)…"
npm run build:web

standalone=""
for candidate in \
  "$root/apps/web/.next/standalone/apps/web" \
  "$root/apps/web/.next/standalone"; do
  if [[ -f "$candidate/server.js" ]]; then
    standalone="$candidate"
    break
  fi
done
if [[ -z "$standalone" ]]; then
  echo "Next standalone server.js not found under apps/web/.next/standalone"
  find "$root/apps/web/.next/standalone" -name server.js 2>/dev/null | head
  exit 1
fi

res="$root/apps/desktop/resources"
rm -rf "$res"
mkdir -p "$res/web" "$res/api" "$res/bin" "$res/colab"

echo "copying standalone from $standalone"
cp -R "$standalone"/. "$res/web/"
mkdir -p "$res/web/.next"
if [[ -d "$root/apps/web/.next/static" ]]; then
  mkdir -p "$res/web/apps/web/.next" 2>/dev/null || true
  if [[ -d "$res/web/apps/web" ]]; then
    mkdir -p "$res/web/apps/web/.next"
    cp -R "$root/apps/web/.next/static" "$res/web/apps/web/.next/static"
  else
    cp -R "$root/apps/web/.next/static" "$res/web/.next/static"
  fi
fi
if [[ -d "$root/apps/web/public" ]]; then
  if [[ -d "$res/web/apps/web" ]]; then
    cp -R "$root/apps/web/public" "$res/web/apps/web/public"
  else
    mkdir -p "$res/web/public"
    cp -R "$root/apps/web/public"/. "$res/web/public/"
  fi
fi

echo "building API ($goos/$goarch)…"
(cd "$root/services/api" && CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" go build -trimpath -ldflags="-s -w" -o "$res/api/$api_name" ./cmd/api)
chmod +x "$res/api/$api_name" 2>/dev/null || true

if [[ -f "$root/vendor/bin/opencode" ]]; then
  cp "$root/vendor/bin/opencode" "$res/bin/opencode"
  chmod +x "$res/bin/opencode"
elif [[ -f "$root/vendor/bin/opencode.exe" ]]; then
  cp "$root/vendor/bin/opencode.exe" "$res/bin/opencode.exe"
fi
if [[ -d "$root/vendor/maestro-dist" ]]; then
  cp -R "$root/vendor/maestro-dist" "$res/maestro-dist"
  if [[ -f "$res/maestro-dist/bin/maestro" ]]; then
    ln -sfn "../maestro-dist/bin/maestro" "$res/bin/maestro"
  fi
  if [[ -f "$res/maestro-dist/bin/maestro.bat" ]]; then
    ln -sfn "../maestro-dist/bin/maestro.bat" "$res/bin/maestro.bat"
  fi
fi
if [[ -d "$root/colab" ]]; then
  cp -R "$root/colab"/. "$res/colab/"
fi

echo "resources ready:"
find "$res" -maxdepth 3 -type f | head -40

echo "electron-builder $eb_target"
# shellcheck disable=SC2086
npx --workspace cherry-desktop electron-builder $eb_target

echo "out: apps/desktop/out"
ls -la "$root/apps/desktop/out" || true
