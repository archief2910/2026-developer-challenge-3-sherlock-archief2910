#!/usr/bin/env bash
set -euo pipefail

###############################################################################
# cli.sh — Bitcoin chain analysis CLI
#
# Usage:
#   ./cli.sh --block <blk.dat> <rev.dat> <xor.dat>
###############################################################################

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

error_json() {
  local code="$1"
  local message="$2"
  printf '{"ok":false,"error":{"code":"%s","message":"%s"}}\n' "$code" "$message"
}

# --- Block mode ---
if [[ "${1:-}" != "--block" ]]; then
  error_json "INVALID_ARGS" "Usage: cli.sh --block <blk.dat> <rev.dat> <xor.dat>"
  echo "Error: This CLI only supports block mode. Use --block flag." >&2
  exit 1
fi

shift
if [[ $# -lt 3 ]]; then
  error_json "INVALID_ARGS" "Block mode requires: --block <blk.dat> <rev.dat> <xor.dat>"
  echo "Error: Block mode requires 3 file arguments: <blk.dat> <rev.dat> <xor.dat>" >&2
  exit 1
fi

BLK_FILE="$1"
REV_FILE="$2"
XOR_FILE="$3"

for f in "$BLK_FILE" "$REV_FILE" "$XOR_FILE"; do
  if [[ ! -f "$f" ]]; then
    error_json "FILE_NOT_FOUND" "File not found: $f"
    echo "Error: File not found: $f" >&2
    exit 1
  fi
done

# Create output directory
mkdir -p out

# Build and run CLI
cd "$SCRIPT_DIR"
if [[ ! -f bin/sherlock-cli ]]; then
  echo "Building sherlock-cli..." >&2
  go build -o bin/sherlock-cli ./cmd/cli
fi

exec "$SCRIPT_DIR/bin/sherlock-cli" --block "$BLK_FILE" "$REV_FILE" "$XOR_FILE"
