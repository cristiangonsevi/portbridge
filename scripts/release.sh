#!/bin/bash
#
# PortBridge Release Script
# Builds all cross-platform binaries and creates a GitHub Release.
#
# Prerequisites:
#   - gh CLI installed and authenticated (run: gh auth login)
#   - Git tag already pushed (or use --tag to create one)
#
# Usage:
#   ./scripts/release.sh v1.0.0                  # Build + create release with tag v1.0.0
#   ./scripts/release.sh v1.0.0 --dry-run        # Simulate without actually releasing
#   ./scripts/release.sh v1.0.0 --notes-file     # Prompt for release notes in editor
#   ./scripts/release.sh v1.0.0 --title "My Title" --notes "Release notes here"
#
# Environment:
#   REPO_OWNER        GitHub owner (default: detected from git remote)
#   REPO_NAME         GitHub repo name (default: detected from git remote)
#

set -euo pipefail

# ─── Colors ────────────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BUILD_DIR="$PROJECT_ROOT/build"

# ─── Defaults ──────────────────────────────────────────────────────────────
DRY_RUN=false
EDIT_NOTES=false
RELEASE_TITLE=""
RELEASE_NOTES=""

# ─── Parse args ────────────────────────────────────────────────────────────
if [ $# -lt 1 ]; then
  echo -e "${RED}Usage: $0 <tag> [options]${NC}"
  echo ""
  echo "Examples:"
  echo "  $0 v1.0.0"
  echo "  $0 v1.0.0 --dry-run"
  echo "  $0 v1.0.0 --notes-file"
  echo "  $0 v1.0.0 --title 'v1.0.0' --notes 'First stable release'"
  exit 1
fi

RELEASE_TAG="$1"
shift

while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run)
      DRY_RUN=true
      shift
      ;;
    --notes-file)
      EDIT_NOTES=true
      shift
      ;;
    --title)
      RELEASE_TITLE="$2"
      shift 2
      ;;
    --notes)
      RELEASE_NOTES="$2"
      shift 2
      ;;
    --help|-h)
      echo "Usage: $0 <tag> [options]"
      echo ""
      echo "Options:"
      echo "  --dry-run               Simulate without creating a release"
      echo "  --notes-file            Open editor to write release notes"
      echo "  --title <title>         Release title (default: tag name)"
      echo "  --notes <text>          Release notes text"
      echo "  --help, -h              Show this help"
      exit 0
      ;;
    *)
      echo -e "${RED}Unknown option: $1${NC}"
      exit 1
      ;;
  esac
done

# ─── Helper functions ──────────────────────────────────────────────────────
info()  { echo -e "${BLUE}[INFO]${NC} $*"; }
ok()    { echo -e "${GREEN}[OK]${NC}   $*"; }
warn()  { echo -e "${YELLOW}[WARN]${NC} $*"; }
fail()  { echo -e "${RED}[FAIL]${NC} $*"; }

# ─── Banner ────────────────────────────────────────────────────────────────
echo ""
echo -e "${BLUE}╔══════════════════════════════════════╗${NC}"
echo -e "${BLUE}║     PortBridge Release Script        ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════╝${NC}"
echo ""

# ─── Step 1: Check prerequisites ──────────────────────────────────────────
echo -e "${YELLOW}━━━ Step 1: Checking prerequisites ─${NC}"
echo ""

ERRORS=0

# Check gh CLI
if ! command -v gh &>/dev/null; then
  fail "gh CLI is not installed."
  fail "Install it from: https://cli.github.com/"
  ERRORS=$((ERRORS + 1))
else
  ok "Found gh: $(command -v gh)"
  # Check auth status
  if ! gh auth status 2>&1 | grep -q "Logged in"; then
    if [ "$DRY_RUN" = true ]; then
      warn "gh is not authenticated (--dry-run, continuing anyway). To release for real, run: gh auth login"
    else
      fail "gh is not authenticated. Run: gh auth login"
      ERRORS=$((ERRORS + 1))
    fi
  else
    ok "gh is authenticated"
  fi
fi

# Check git
if ! command -v git &>/dev/null; then
  fail "git is not installed."
  ERRORS=$((ERRORS + 1))
else
  ok "Found git: $(command -v git)"
fi

if [ "$ERRORS" -gt 0 ]; then
  exit 1
fi

# ─── Step 2: Validate git state ────────────────────────────────────────────
echo ""
echo -e "${YELLOW}━━━ Step 2: Validating git state ─━━${NC}"
echo ""

cd "$PROJECT_ROOT"

# Check for uncommitted changes
if ! git diff --quiet HEAD 2>/dev/null; then
  warn "You have uncommitted changes. It's recommended to commit before releasing."
  warn "Continue anyway? (y/N)"
  if [ "$DRY_RUN" = false ]; then
    read -r response
    if [[ ! "$response" =~ ^[yY]$ ]]; then
      info "Aborted."
      exit 1
    fi
  else
    info "[DRY RUN] Would prompt to continue with uncommitted changes"
  fi
fi

# Check tag doesn't exist already
if git rev-parse "$RELEASE_TAG" &>/dev/null 2>&1; then
  fail "Tag $RELEASE_TAG already exists locally."
  fail "Delete it with: git tag -d $RELEASE_TAG && git push origin :$RELEASE_TAG"
  exit 1
else
  ok "Tag $RELEASE_TAG is available"
fi

# Detect repo from git remote
REMOTE_URL=$(git config --get remote.origin.url 2>/dev/null || echo "")
if [[ "$REMOTE_URL" =~ github\.com[:/](.+)/(.+)\.git ]]; then
  REPO_OWNER="${BASH_REMATCH[1]}"
  REPO_NAME="${BASH_REMATCH[2]}"
  ok "Detected repo: $REPO_OWNER/$REPO_NAME"
else
  REPO_OWNER="${REPO_OWNER:-unknown}"
  REPO_NAME="${REPO_NAME:-portbridge}"
  warn "Could not detect repo from git remote. Using: $REPO_OWNER/$REPO_NAME"
fi

# ─── Step 3: Build all binaries ───────────────────────────────────────────
echo ""
echo -e "${YELLOW}━━━ Step 3: Building binaries ─━━━${NC}"
echo ""

if [ "$DRY_RUN" = true ]; then
  info "[DRY RUN] Would run: ./scripts/build_all.sh"
else
  # Clean build dir first
  rm -rf "$BUILD_DIR"
  mkdir -p "$BUILD_DIR"

  # Source build_all.sh to build everything
  # (We inline the logic so we have full control)
  TARGETS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
  )

  for TARGET in "${TARGETS[@]}"; do
    OS=$(echo "$TARGET" | cut -d'/' -f1)
    ARCH=$(echo "$TARGET" | cut -d'/' -f2)
    OUTPUT_NAME="portbridge-$OS-$ARCH"

    echo "  Building for $OS/$ARCH..."
    cd "$PROJECT_ROOT"
    GOOS="$OS" GOARCH="$ARCH" go build -ldflags="-s -w" -o "$BUILD_DIR/$OUTPUT_NAME" .
  done

  ok "All binaries built successfully:"
  ls -lh "$BUILD_DIR" | sed 's/^/    /'
fi

# ─── Step 4: Create release assets (compress) ──────────────────────────────
echo ""
echo -e "${YELLOW}━━━ Step 4: Preparing assets ─━━━━${NC}"
echo ""

ASSETS_DIR=$(mktemp -d)
ASSETS=()

if [ "$DRY_RUN" = true ]; then
  info "[DRY RUN] Would compress each binary as .tar.gz"
else
  for BIN in "$BUILD_DIR"/portbridge-*; do
    [ -f "$BIN" ] || continue
    BASENAME=$(basename "$BIN")
    TARBALL="$ASSETS_DIR/${BASENAME}.tar.gz"

    info "Compressing: $BASENAME -> $BASENAME.tar.gz"
    cd "$BUILD_DIR"
    tar czf "$TARBALL" "$BASENAME" 2>/dev/null

    ASSETS+=("$TARBALL")
    ok "Created: $(basename "$TARBALL") ($(du -h "$TARBALL" | cut -f1))"
  done
fi

# ─── Step 5: Create GitHub Release ─────────────────────────────────────────
echo ""
echo -e "${YELLOW}━━━ Step 5: Creating GitHub Release ─${NC}"
echo ""

if [ -z "$RELEASE_TITLE" ]; then
  RELEASE_TITLE="$RELEASE_TAG"
fi

if [ "$DRY_RUN" = true ]; then
  info "[DRY RUN] Would create release:"
  info "  Tag:    $RELEASE_TAG"
  info "  Title:  $RELEASE_TITLE"
  info "  Repo:   $REPO_OWNER/$REPO_NAME"
  info "  Assets: ${ASSETS[*]:-(none)}"

  # Clean up
  rm -rf "$ASSETS_DIR"
else
  # Write release notes if --notes-file
  if [ "$EDIT_NOTES" = true ]; then
    TEMP_NOTES=$(mktemp)
    cat > "$TEMP_NOTES" <<- EOF
# PortBridge $RELEASE_TAG

## Changes

- TODO: list changes here

## Installation

\`\`\`bash
curl -fsSL https://raw.githubusercontent.com/$REPO_OWNER/$REPO_NAME/main/scripts/install.sh | sh
\`\`\`

## Checksums

EOF
    ${EDITOR:-vim} "$TEMP_NOTES"
    RELEASE_NOTES=$(cat "$TEMP_NOTES")
    rm -f "$TEMP_NOTES"
  fi

  # Prepare gh args
  GH_ARGS=(--repo "$REPO_OWNER/$REPO_NAME" --title "$RELEASE_TITLE" --tag "$RELEASE_TAG")

  if [ -n "$RELEASE_NOTES" ]; then
    GH_ARGS+=(--notes "$RELEASE_NOTES")
  fi

  # Add assets
  for ASSET in "${ASSETS[@]}"; do
    GH_ARGS+=("$ASSET")
  done

  info "Creating release $RELEASE_TAG on $REPO_OWNER/$REPO_NAME..."

  if gh release create "${GH_ARGS[@]}"; then
    ok "Release created: https://github.com/$REPO_OWNER/$REPO_NAME/releases/tag/$RELEASE_TAG"
  else
    fail "Failed to create release."
    rm -rf "$ASSETS_DIR"
    exit 1
  fi

  rm -rf "$ASSETS_DIR"
fi

# ─── Step 6: Push tag (optional prompt) ────────────────────────────────────
echo ""
echo -e "${YELLOW}━━━ Step 6: Tag management ─━━━━━━${NC}"
echo ""

if [ "$DRY_RUN" = false ]; then
  info "Do you want to push the tag to remote? (Y/n)"
  read -r response
  if [[ ! "$response" =~ ^[nN]$ ]]; then
    git tag "$RELEASE_TAG"
    git push origin "$RELEASE_TAG"
    ok "Tag $RELEASE_TAG pushed to origin"
  else
    info "Tag not pushed. Push manually when ready:"
    info "  git tag $RELEASE_TAG && git push origin $RELEASE_TAG"
  fi
else
  info "[DRY RUN] Would push tag: git push origin $RELEASE_TAG"
fi

# ─── Done ──────────────────────────────────────────────────────────────────
echo ""
echo -e "${GREEN}╔══════════════════════════════════════╗${NC}"
echo -e "${GREEN}║      Release completed!              ║${NC}"
echo -e "${GREEN}╚══════════════════════════════════════╝${NC}"
echo ""
info "Tag:    $RELEASE_TAG"
info "Repo:   https://github.com/$REPO_OWNER/$REPO_NAME"
if [ "$DRY_RUN" = false ]; then
  info "Release: https://github.com/$REPO_OWNER/$REPO_NAME/releases/tag/$RELEASE_TAG"
fi
echo ""
