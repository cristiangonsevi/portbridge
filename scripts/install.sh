#!/bin/bash
#
# PortBridge Installer
# Downloads the latest PortBridge binary from GitHub Releases for your platform.
#
# Usage:
#   ./scripts/install.sh                       # Install to ~/.local/bin
#   ./scripts/install.sh --prefix ~/.local/bin # Install to custom path
#   ./scripts/install.sh --version v1.0.0      # Install a specific version
#   ./scripts/install.sh --dry-run             # Show what would be done
#

set -euo pipefail

# ─── Config ────────────────────────────────────────────────────────────────
REPO_OWNER="${REPO_OWNER:-cristiangonsevi}"
REPO_NAME="${REPO_NAME:-portbridge}"
GITHUB_API="https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/releases"
GITHUB_DL="https://github.com/$REPO_OWNER/$REPO_NAME/releases/download"

# ─── Colors ────────────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# ─── Defaults ──────────────────────────────────────────────────────────────
INSTALL_PREFIX="${HOME}/.local/bin"
VERSION="latest"
DRY_RUN=false

# ─── Parse flags ───────────────────────────────────────────────────────────
while [[ $# -gt 0 ]]; do
  case "$1" in
    --prefix)
      INSTALL_PREFIX="$2"
      shift 2
      ;;
    --version)
      VERSION="$2"
      shift 2
      ;;
    --dry-run)
      DRY_RUN=true
      shift
      ;;
    --help|-h)
      echo "Usage: $0 [OPTIONS]"
      echo ""
      echo "Options:"
      echo "  --prefix <path>       Install binary to <path> (default: ~/.local/bin)"
      echo "  --version <tag>       Install a specific version/tag (default: latest)"
      echo "  --dry-run             Show what would be done without actually installing"
      echo "  --help, -h            Show this help message"
      echo ""
      echo "Environment variables:"
      echo "  REPO_OWNER            GitHub owner/org (default: yourusername)"
      echo "  REPO_NAME             GitHub repo name (default: portbridge)"
      echo ""
      echo "Examples:"
      echo "  REPO_OWNER=myuser REPO_NAME=portbridge $0"
      echo "  $0 --version v0.1.0"
      echo "  $0 --prefix /usr/local/bin"
      echo ""
      echo "Note: The default install prefix is ~/.local/bin so no sudo is needed."
      echo "      If you install to a system directory (e.g. /usr/local/bin),"
      echo "      you may need to run with sudo."
      exit 0
      ;;
    *)
      echo -e "${RED}Unknown option: $1${NC}"
      echo "Usage: $0 [--prefix <path>] [--version <tag>] [--dry-run]"
      exit 1
      ;;
  esac
done

# ─── Helper functions ──────────────────────────────────────────────────────
info()  { echo -e "${BLUE}[INFO]${NC} $*"; }
ok()    { echo -e "${GREEN}[OK]${NC}   $*"; }
warn()  { echo -e "${YELLOW}[WARN]${NC} $*"; }
fail()  { echo -e "${RED}[FAIL]${NC} $*"; }

# Detect platform
detect_platform() {
  local os arch

  case "$(uname -s)" in
    Linux)  os="linux" ;;
    Darwin) os="darwin" ;;
    *)
      fail "Unsupported OS: $(uname -s). Only Linux and macOS are supported."
      exit 1
      ;;
  esac

  case "$(uname -m)" in
    x86_64|amd64) arch="amd64" ;;
    aarch64|arm64) arch="arm64" ;;
    armv7l|armv8l) arch="arm64" ;;
    *)
      fail "Unsupported architecture: $(uname -m). Only amd64 and arm64 are supported."
      exit 1
      ;;
  esac

  echo "${os}-${arch}"
}

# ─── Banner ────────────────────────────────────────────────────────────────
echo ""
echo -e "${BLUE}╔══════════════════════════════════════╗${NC}"
echo -e "${BLUE}║      PortBridge Installer            ║${NC}"
echo -e "${BLUE}║   (downloads pre-built binaries)     ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════╝${NC}"
echo ""

# ─── Step 1: Detect platform ──────────────────────────────────────────────
echo -e "${YELLOW}━━━ Step 1: Detecting platform ─━━━${NC}"
echo ""

PLATFORM=$(detect_platform)
info "Detected platform: $PLATFORM"

# ─── Step 2: Check dependencies ───────────────────────────────────────────
echo ""
echo -e "${YELLOW}━━━ Step 2: Checking dependencies ─${NC}"
echo ""

ERRORS=0

if command -v curl &>/dev/null; then
  DOWNLOADER="curl"
  ok "Found curl: $(command -v curl)"
elif command -v wget &>/dev/null; then
  DOWNLOADER="wget"
  ok "Found wget: $(command -v wget)"
else
  fail "Neither curl nor wget is installed."
  ERRORS=$((ERRORS + 1))
fi

if [ "$ERRORS" -gt 0 ]; then
  echo ""
  fail "Please install curl or wget and try again."
  exit 1
fi

# ─── Step 3: Resolve release version ──────────────────────────────────────
echo ""
echo -e "${YELLOW}━━━ Step 3: Resolving release ─━━━━${NC}"
echo ""

if [ "$VERSION" = "latest" ]; then
  info "Fetching latest release from $REPO_OWNER/$REPO_NAME..."
  if [ "$DRY_RUN" = true ]; then
    RELEASE_TAG="latest"
    info "[DRY RUN] Would fetch latest release tag from $GITHUB_API/latest"
  else
    if [ "$DOWNLOADER" = "curl" ]; then
      RELEASE_TAG=$(curl -sSf "$GITHUB_API/latest" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": "\(.*\)",/\1/')
    else
      RELEASE_TAG=$(wget -qO- "$GITHUB_API/latest" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": "\(.*\)",/\1/')
    fi

    if [ -z "$RELEASE_TAG" ]; then
      fail "Could not determine the latest release."
      fail "Make sure REPO_OWNER and REPO_NAME are set correctly:"
      fail "  REPO_OWNER=$REPO_OWNER REPO_NAME=$REPO_NAME $0"
      fail ""
      fail "Or check the releases page:"
      fail "  https://github.com/$REPO_OWNER/$REPO_NAME/releases"
      exit 1
    fi
  fi
else
  RELEASE_TAG="$VERSION"
fi

info "Release tag: $RELEASE_TAG"

# ─── Step 4: Download the binary ──────────────────────────────────────────
echo ""
echo -e "${YELLOW}━━━ Step 4: Downloading binary ─━━${NC}"
echo ""

# Normalize platform name: portbridge-<os>-<arch>
OS="${PLATFORM%-*}"
ARCH="${PLATFORM#*-}"
BINARY_NAME="portbridge-${OS}-${ARCH}"
DOWNLOAD_URL="$GITHUB_DL/$RELEASE_TAG/$BINARY_NAME"

# Destination
TEMP_DIR=$(mktemp -d)
TEMP_BIN="$TEMP_DIR/portbridge"

info "Download URL: $DOWNLOAD_URL"

if [ "$DRY_RUN" = true ]; then
  info "[DRY RUN] Would download: $DOWNLOAD_URL"
  info "[DRY RUN] Would install to: $INSTALL_PREFIX/portbridge"
  rm -rf "$TEMP_DIR"
else
  if [ "$DOWNLOADER" = "curl" ]; then
    HTTP_CODE=$(curl -sSfL -w "%{http_code}" -o "$TEMP_BIN" "$DOWNLOAD_URL" 2>&1) || true
  else
    HTTP_CODE=$(wget -q -O "$TEMP_BIN" "$DOWNLOAD_URL" 2>&1 && echo "200") || true
  fi

  # Check if download succeeded
  if [ ! -f "$TEMP_BIN" ] || [ ! -s "$TEMP_BIN" ]; then
    fail "Download failed (HTTP $HTTP_CODE)."
    fail ""
    fail "Possible reasons:"
    fail "  1. The release $RELEASE_TAG doesn't exist"
    fail "  2. The binary for $PLATFORM wasn't built for that release"
    fail "  3. REPO_OWNER/REPO_NAME is incorrect"
    fail ""
    fail "Check: https://github.com/$REPO_OWNER/$REPO_NAME/releases"
    rm -rf "$TEMP_DIR"
    exit 1
  fi

  chmod +x "$TEMP_BIN"
  ok "Downloaded: $BINARY_NAME ($(du -h "$TEMP_BIN" | cut -f1))"
fi

# ─── Step 5: Install ──────────────────────────────────────────────────────
echo ""
echo -e "${YELLOW}━━━ Step 5: Installing ─━━━━━━━━━━━${NC}"
echo ""

if [ "$DRY_RUN" = true ]; then
  info "[DRY RUN] Skipping installation."
  rm -rf "$TEMP_DIR" 2>/dev/null || true
else
  INSTALL_DIR="$INSTALL_PREFIX"

  if [ ! -d "$INSTALL_DIR" ]; then
    info "Creating directory $INSTALL_DIR..."
    mkdir -p "$INSTALL_DIR"
  fi

  if cp "$TEMP_BIN" "$INSTALL_DIR/portbridge" 2>/dev/null; then
    chmod +x "$INSTALL_DIR/portbridge"
    ok "Installed to: $INSTALL_DIR/portbridge"
  else
    fail "Failed to install to $INSTALL_DIR. Check write permissions."
    fail "You can specify a different directory with --prefix:"
    fail "  $0 --prefix ~/bin"
    rm -rf "$TEMP_DIR"
    exit 1
  fi

  rm -rf "$TEMP_DIR"

  # Check if install dir is in PATH
  case ":$PATH:" in
    *:"$INSTALL_DIR":*)
      ok "$INSTALL_DIR is in your PATH"
      ;;
    *)
      warn "$INSTALL_DIR is NOT in your PATH. Add it to your shell config:"
      warn ""
      warn "  For bash:"
      warn "    echo 'export PATH=\"\$PATH:$INSTALL_DIR\"' >> ~/.bashrc"
      warn ""
      warn "  For zsh (macOS default):"
      warn "    echo 'export PATH=\"\$PATH:$INSTALL_DIR\"' >> ~/.zshrc"
      warn ""
      warn "  Then reload:"
      warn "    source ~/.bashrc   # or ~/.zshrc"
      ;;
  esac
fi

# ─── Done ──────────────────────────────────────────────────────────────────
echo ""
echo -e "${GREEN}╔══════════════════════════════════════╗${NC}"
echo -e "${GREEN}║   PortBridge installed successfully  ║${NC}"
echo -e "${GREEN}╚══════════════════════════════════════╝${NC}"
echo ""
info "Run the following to verify:"
info "  portbridge --help"
echo ""
info "Installed: $RELEASE_TAG ($PLATFORM)"
info "Source:    https://github.com/$REPO_OWNER/$REPO_NAME"
echo ""
