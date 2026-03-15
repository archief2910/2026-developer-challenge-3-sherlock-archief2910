#!/usr/bin/env bash
set -euo pipefail

###############################################################################
# setup.sh — Install project dependencies
###############################################################################

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

echo "=== Sherlock Setup ==="

# Decompress block fixtures if not already present
for gz in fixtures/*.dat.gz; do
  dat="${gz%.gz}"
  if [[ ! -f "$dat" ]]; then
    echo "Decompressing $(basename "$gz")..."
    gunzip -k "$gz"
  fi
done

# Download Go dependencies
echo "Downloading Go dependencies..."
go mod tidy 2>&1 || go mod download 2>&1 || true

# Build CLI binary
echo "Building CLI..."
mkdir -p bin
go build -o bin/sherlock-cli ./cmd/cli

# Build web server binary
echo "Building web server..."
go build -o bin/sherlock-web ./cmd/web

# Build React frontend
echo "Building React frontend..."
cd web/ui
npm install
node node_modules/webpack/bin/webpack.js --mode production
mkdir -p ../dist
cp index.html ../dist/index.html
cd ../..

echo ""
echo "Setup complete!"