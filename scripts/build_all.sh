#!/bin/bash
#
# PortBridge Cross-Platform Builder
# Builds PortBridge binaries for multiple OS/architecture targets.
#
# Usage:
#   ./scripts/build_all.sh              # Build all targets
#   GOOS=linux GOARCH=amd64 ./scripts/build_all.sh  # Build for a single target
#

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

# Set output directory for binaries (relative to project root)
OUTPUT_DIR="$PROJECT_ROOT/build"
mkdir -p "$OUTPUT_DIR"

# Define target OS and architectures
TARGETS=(
  "linux/amd64"
  "linux/arm64"
  "darwin/amd64"
  "darwin/arm64"
)

# Build for each target
for TARGET in "${TARGETS[@]}"; do
  OS=$(echo "$TARGET" | cut -d'/' -f1)
  ARCH=$(echo "$TARGET" | cut -d'/' -f2)
  OUTPUT_NAME="portbridge-$OS-$ARCH"
  [ "$OS" == "windows" ] && OUTPUT_NAME+=".exe"

  echo "Building for $OS/$ARCH..."
  cd "$PROJECT_ROOT"
  GOOS="$OS" GOARCH="$ARCH" go build -ldflags="-s -w" -o "$OUTPUT_DIR/$OUTPUT_NAME" .
done

echo ""
echo "Builds completed. Binaries are in the '$OUTPUT_DIR' directory."
ls -lh "$OUTPUT_DIR"
echo ""
