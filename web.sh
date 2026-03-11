#!/usr/bin/env bash
set -euo pipefail

###############################################################################
# web.sh — Web visualizer
#
# Starts the web visualizer server.
###############################################################################

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PORT="${PORT:-3000}"

cd "$SCRIPT_DIR"

# Build web server if not already built
if [[ ! -f bin/sherlock-web ]]; then
  echo "Building sherlock-web..." >&2
  go build -o bin/sherlock-web ./cmd/web
fi

export PORT
exec "$SCRIPT_DIR/bin/sherlock-web"
