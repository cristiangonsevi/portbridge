#!/bin/bash

# Set output directory for binaries
OUTPUT_DIR="build"
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
  GOOS="$OS" GOARCH="$ARCH" go build -o "$OUTPUT_DIR/$OUTPUT_NAME" ./cmd
done

echo "Builds completed. Binaries are in the '$OUTPUT_DIR' directory."
