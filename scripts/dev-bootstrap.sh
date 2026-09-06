#!/usr/bin/env bash
# Resolve Cherry sidecars + Mongo for local/dev API runs.
# Prefer PATH binaries (developer machine), then vendor/bin (bundled).
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

if [[ -f .env ]]; then
  set -a
  # shellcheck disable=SC1091
  . ./.env
  set +a
fi

resolve_bin() {
  local name="$1"
  local env_val="$2"
  if [[ -n "$env_val" && -x "$env_val" ]]; then
    printf '%s' "$env_val"
    return 0
  fi
  if command -v "$name" >/dev/null 2>&1; then
    command -v "$name"
    return 0
  fi
  if [[ -x "$root/vendor/bin/$name" ]]; then
    printf '%s' "$root/vendor/bin/$name"
    return 0
  fi
  return 1
}

OC_BIN="$(resolve_bin opencode "${CHERRY_OPENCODE_BIN:-}" || true)"
MA_BIN="$(resolve_bin maestro "${CHERRY_MAESTRO_BIN:-}" || true)"
CF_BIN="$(resolve_bin cloudflared "${CHERRY_CLOUDFLARED_BIN:-}" || true)"

export MONGO_URI="${MONGO_URI:-mongodb://127.0.0.1:27017/cherry}"
export CHERRY_API_ADDR="${CHERRY_API_ADDR:-127.0.0.1:43148}"
export CHERRY_WEB_ORIGIN="${CHERRY_WEB_ORIGIN:-http://127.0.0.1:43147}"
export CHERRY_WEB_URL="${CHERRY_WEB_URL:-http://127.0.0.1:43147}"

if [[ -n "$OC_BIN" ]]; then
  export CHERRY_OPENCODE_BIN="$OC_BIN"
fi
if [[ -n "$MA_BIN" ]]; then
  export CHERRY_MAESTRO_BIN="$MA_BIN"
fi
if [[ -n "$CF_BIN" ]]; then
  export CHERRY_CLOUDFLARED_BIN="$CF_BIN"
fi

# Keep resolved paths in .env (no secrets written here).
upsert_env() {
  local key="$1"
  local val="$2"
  [[ -z "$val" ]] && return 0
  touch .env
  if grep -q "^${key}=" .env 2>/dev/null; then
    # portable-ish in-place replace
    awk -v k="$key" -v v="$val" 'BEGIN{done=0} $0 ~ "^"k"=" {print k"="v; done=1; next} {print} END{if(!done) print k"="v}' .env > .env.tmp
    mv .env.tmp .env
  elif grep -q "^# ${key}=" .env 2>/dev/null; then
    awk -v k="$key" -v v="$val" 'BEGIN{done=0} $0 ~ "^# "k"=" {print k"="v; done=1; next} {print} END{if(!done) print k"="v}' .env > .env.tmp
    mv .env.tmp .env
  else
    printf '\n%s=%s\n' "$key" "$val" >> .env
  fi
}

upsert_env MONGO_URI "$MONGO_URI"
upsert_env CHERRY_OPENCODE_BIN "${CHERRY_OPENCODE_BIN:-}"
upsert_env CHERRY_MAESTRO_BIN "${CHERRY_MAESTRO_BIN:-}"
upsert_env CHERRY_CLOUDFLARED_BIN "${CHERRY_CLOUDFLARED_BIN:-}"

bash "$root/scripts/ensure-mongo.sh"

echo "bootstrap:"
echo "  MONGO_URI=$MONGO_URI"
echo "  CHERRY_OPENCODE_BIN=${CHERRY_OPENCODE_BIN:-missing}"
echo "  CHERRY_MAESTRO_BIN=${CHERRY_MAESTRO_BIN:-missing}"
echo "  CHERRY_CLOUDFLARED_BIN=${CHERRY_CLOUDFLARED_BIN:-missing}"
echo "  CHERRY_LLM_API_KEY=$([ -n "${CHERRY_LLM_API_KEY:-}" ] && echo set || echo empty)"

if [[ -z "${CHERRY_LLM_API_KEY:-}" && -z "${CHERRY_LLM_BASE_URL:-}" ]]; then
  echo "warn: CHERRY_LLM_API_KEY / CHERRY_LLM_BASE_URL empty → LLM mock; OpenCode will not write without a key or Colab URL" >&2
fi
