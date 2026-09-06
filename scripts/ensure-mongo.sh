#!/usr/bin/env bash
# Ensure local MongoDB is running for Cherry (no Docker required).
# Safe to re-run. Does not start Docker Compose.
set -euo pipefail

DBPATH="${CHERRY_MONGO_DBPATH:-/var/lib/mongodb}"
LOGPATH="${CHERRY_MONGO_LOGPATH:-/var/log/mongodb/mongod.log}"
BIND="${CHERRY_MONGO_BIND:-127.0.0.1}"
PORT="${CHERRY_MONGO_PORT:-27017}"
URI="${MONGO_URI:-mongodb://${BIND}:${PORT}/cherry}"

if ! command -v mongod >/dev/null 2>&1; then
  echo "mongod not found. Install MongoDB 7 (mongodb-org) or set Docker:"
  echo "  docker compose up -d mongo"
  exit 1
fi

if mongosh --quiet --eval 'db.runCommand({ ping: 1 })' "$URI" >/dev/null 2>&1; then
  echo "mongo already up: $URI"
  exit 0
fi

# Prefer package paths; fall back to /data/db for manual installs.
if [[ ! -d "$DBPATH" ]]; then
  DBPATH=/data/db
fi
mkdir -p "$(dirname "$LOGPATH")" "$DBPATH" 2>/dev/null || true

if id mongodb >/dev/null 2>&1; then
  sudo chown -R mongodb:mongodb "$DBPATH" "$(dirname "$LOGPATH")" 2>/dev/null || true
  sudo -u mongodb mongod --dbpath "$DBPATH" --logpath "$LOGPATH" --bind_ip "$BIND" --port "$PORT" --fork
else
  mongod --dbpath "$DBPATH" --logpath "$LOGPATH" --bind_ip "$BIND" --port "$PORT" --fork
fi

sleep 1
if mongosh --quiet --eval 'db.runCommand({ ping: 1 })' "$URI" >/dev/null 2>&1; then
  echo "mongo started: $URI"
else
  echo "mongo failed to answer ping on $URI" >&2
  exit 1
fi
