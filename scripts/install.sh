#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(dirname "$(realpath "$0")")"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"
BIN_NAME="o-mark"

DEFAULT_PREFIX="/usr/local"
PREFIX="$DEFAULT_PREFIX"

REGISTER_MIME=0
UNINSTALL=0
SKIP_DEPS=0
ASSUME_YES=0

# Required Qt6 pkg-config modules (mirrors the imports actually used in src/ui/*.qml + miqt's cgo directive)
REQUIRED_PKGCONFIG=(Qt6Widgets Qt6Quick Qt6QuickControls2Basic Qt6QuickLayouts Qt6WebEngineQuick)

usage() {
    cat <<EOF
Install O'Mark.

Usage:
  $0 [--prefix=DIR] [--register-mime] [--uninstall] [--skip-deps-check] [--yes]

Options:
  --prefix=DIR        Install under DIR/bin (default: $DEFAULT_PREFIX). Uses sudo when
                       DIR is not writable by the current user (e.g. the default
                       $DEFAULT_PREFIX). Pass --prefix="\$HOME/.local" for a user install
                       with no sudo.
  --register-mime     Also register O'Mark as the default handler for .md files
  --uninstall         Remove the binary (and, with --register-mime, the .desktop entry)
  --skip-deps-check   Skip the system dependency check before building
  --yes, -y           Skip the confirmation prompt (for scripted/non-interactive runs)
  -h, --help          Show this help
EOF
}

for arg in "$@"; do
    case "$arg" in
        --prefix=*) PREFIX="${arg#--prefix=}" ;;
        --register-mime) REGISTER_MIME=1 ;;
        --uninstall) UNINSTALL=1 ;;
        --skip-deps-check) SKIP_DEPS=1 ;;
        --yes|-y) ASSUME_YES=1 ;;
        -h|--help) usage; exit 0 ;;
        *) echo "✗ Unknown option: $arg"; usage; exit 1 ;;
    esac
done

BIN_DIR="$PREFIX/bin"
BIN_PATH="$BIN_DIR/$BIN_NAME"
DESKTOP_FILE="$HOME/.local/share/applications/o-mark.desktop"

# Walk up from BIN_DIR to the nearest existing ancestor and test if it's
# writable by the current user. Covers both "dir already exists" and
# "dir (and maybe its parent) still needs to be created".
needs_sudo() {
    local dir="$1"
    while [[ ! -d "$dir" && "$dir" != "/" ]]; do
        dir="$(dirname "$dir")"
    done
    [[ ! -w "$dir" ]]
}

SUDO=""
if needs_sudo "$BIN_DIR"; then
    SUDO="sudo"
fi

info() { echo "● $1"; }
ok()   { echo "  ✓ $1"; }
warn() { echo "  ○ $1"; }
fail() { echo "  ✗ $1"; DEPS_FAILED=1; }

# =====================================================================
#  Summary + confirmation
# =====================================================================
info "$([[ $UNINSTALL -eq 1 ]] && echo "Uninstall" || echo "Install") plan"
echo "    action          $([[ $UNINSTALL -eq 1 ]] && echo "uninstall" || echo "build + install")"
echo "    prefix          $PREFIX$([[ "$PREFIX" == "$DEFAULT_PREFIX" ]] && echo " (default — pass --prefix=DIR to change)")"
echo "    binary target   $BIN_PATH"
echo "    sudo required   $([[ -n "$SUDO" ]] && echo "yes ($BIN_DIR is not writable by $(whoami))" || echo "no")"
if [[ $UNINSTALL -eq 0 ]]; then
    echo "    register .md as default viewer   $([[ $REGISTER_MIME -eq 1 ]] && echo "yes" || echo "no (pass --register-mime to enable)")"
    echo "    dependency check                 $([[ $SKIP_DEPS -eq 1 ]] && echo "skipped (--skip-deps-check)" || echo "yes")"
fi
echo

if [[ $ASSUME_YES -eq 0 ]]; then
    if ! read -r -p "Continue? [y/N] " reply; then
        echo
        echo "✗ No interactive terminal detected — pass --yes to run non-interactively."
        exit 1
    fi
    case "$reply" in
        y|Y|yes|YES) ;;
        *) echo "Aborted."; exit 1 ;;
    esac
    echo
fi

# =====================================================================
#  0. Dependency check
# =====================================================================
check_dependencies() {
    DEPS_FAILED=0
    info "Checking system dependencies"

    DISTRO_ID=""
    DISTRO_VERSION=""
    if [[ -f /etc/os-release ]]; then
        # shellcheck disable=SC1091
        . /etc/os-release
        DISTRO_ID="${ID:-}"
        DISTRO_VERSION="${VERSION_ID:-}"
    fi

    case "$DISTRO_ID" in
        arch)
            ok "Arch Linux detected"
            warn "Arch is rolling release — package versions aren't pinned, but the system must be up to date:"
            echo "      sudo pacman -Syu"
            PKG_CHECK_CMD="pacman -Qi"
            MISSING_PKG_INSTALL="sudo pacman -S"
            REQUIRED_PACKAGES=(pkg-config gcc qt6-base qt6-declarative qt6-webengine)
            ;;
        ubuntu)
            if [[ "$DISTRO_VERSION" != "24.04" ]]; then
                fail "Ubuntu $DISTRO_VERSION detected — only Ubuntu 24.04 LTS is supported"
                echo "    → install on Ubuntu 24.04 LTS, or on Arch (kept up to date)"
                PKG_CHECK_CMD=""
                REQUIRED_PACKAGES=()
            else
                ok "Ubuntu 24.04 LTS detected"
                warn "Ubuntu 24.04's 'golang-go' package is usually older than this project's go.mod requires."
                warn "Install Go from https://go.dev/dl/ instead of apt if 'go version' below looks too old."
                PKG_CHECK_CMD="dpkg -s"
                MISSING_PKG_INSTALL="sudo apt install"
                REQUIRED_PACKAGES=(pkg-config build-essential qt6-base-dev qt6-declarative-dev qt6-webengine-dev)
            fi
            ;;
        *)
            warn "Unrecognized distro (${DISTRO_ID:-unknown}) — only Arch (rolling, kept up to date) and Ubuntu 24.04 LTS are supported."
            warn "Continuing best-effort; skipping package-manager checks."
            PKG_CHECK_CMD=""
            REQUIRED_PACKAGES=()
            ;;
    esac

    if [[ -n "$PKG_CHECK_CMD" ]]; then
        local missing=()
        for pkg in "${REQUIRED_PACKAGES[@]}"; do
            if $PKG_CHECK_CMD "$pkg" >/dev/null 2>&1; then
                ok "$pkg"
            else
                fail "$pkg not found"
                missing+=("$pkg")
            fi
        done
        if [[ ${#missing[@]} -gt 0 ]]; then
            echo "    → $MISSING_PKG_INSTALL ${missing[*]}"
        fi
    fi

    if ! command -v pkg-config >/dev/null 2>&1; then
        fail "pkg-config not found on PATH"
    else
        for mod in "${REQUIRED_PKGCONFIG[@]}"; do
            if pkg-config --exists "$mod" 2>/dev/null; then
                ok "pkg-config: $mod ($(pkg-config --modversion "$mod"))"
            else
                fail "pkg-config: $mod not found (Qt6 WebEngine/Quick/Widgets dev files missing)"
            fi
        done
    fi

    if ! command -v go >/dev/null 2>&1; then
        fail "go not found on PATH"
    else
        ok "go ($(go version | awk '{print $3}'))"
        if [[ "$(go env CGO_ENABLED)" != "1" ]]; then
            fail "CGO_ENABLED=0 — miqt requires CGO; run with CGO_ENABLED=1"
        fi
    fi

    echo
    if [[ $DEPS_FAILED -ne 0 ]]; then
        echo "✗ Missing dependencies — install the packages above, then re-run this script"
        exit 1
    fi
    ok "All dependencies present"
    echo
}

if [[ $UNINSTALL -eq 0 && $SKIP_DEPS -eq 0 ]]; then
    check_dependencies
fi

# =====================================================================
#  Uninstall
# =====================================================================
if [[ $UNINSTALL -eq 1 ]]; then
    info "Uninstalling O'Mark"
    $SUDO rm -f "$BIN_PATH"
    ok "Removed $BIN_PATH"
    if [[ -f "$DESKTOP_FILE" ]]; then
        rm -f "$DESKTOP_FILE"
        update-desktop-database "$HOME/.local/share/applications/" >/dev/null 2>&1 || true
        ok "Removed $DESKTOP_FILE"
    fi
    exit 0
fi

# =====================================================================
#  1. Build
# =====================================================================
info "Building O'Mark"
    export CGO_CXXFLAGS="${CGO_CXXFLAGS:+$CGO_CXXFLAGS }-Wno-sfinae-incomplete"
    ( cd "$REPO_ROOT/src" && go build -o target/"$BIN_NAME" . )
ok "Built $REPO_ROOT/src/target/$BIN_NAME"
echo

# =====================================================================
#  2. Install binary
# =====================================================================
info "Installing binary to $BIN_DIR"
$SUDO mkdir -p "$BIN_DIR"
$SUDO cp "$REPO_ROOT/src/target/$BIN_NAME" "$BIN_PATH"
ok "Installed $BIN_PATH"

if [[ ":$PATH:" != *":$BIN_DIR:"* ]]; then
    echo "  ○ $BIN_DIR is not on your PATH — add to ~/.bashrc or ~/.zshrc:"
    echo "      export PATH=\"$BIN_DIR:\$PATH\""
fi
echo

# =====================================================================
#  3. Register as default .md viewer (optional)
# =====================================================================
if [[ $REGISTER_MIME -eq 1 ]]; then
    info "Registering O'Mark as default .md handler"

    mkdir -p "$HOME/.local/share/mime/packages/"
    cat > "$HOME/.local/share/mime/packages/text-markdown.xml" << 'EOF'
<?xml version="1.0" encoding="utf-8"?>
<mime-info xmlns="http://www.freedesktop.org/standards/shared-mime-info">
  <mime-type type="text/markdown">
    <comment>Markdown document</comment>
    <glob pattern="*.md"/>
    <glob pattern="*.markdown"/>
  </mime-type>
</mime-info>
EOF
    update-mime-database "$HOME/.local/share/mime/"
    ok "Declared text/markdown MIME type"

    mkdir -p "$HOME/.local/share/applications/"
    cat > "$DESKTOP_FILE" << EOF
[Desktop Entry]
Name=O'Mark
Comment=Minimal Markdown Viewer
Exec=$BIN_PATH %f
Icon=text-x-markdown
Type=Application
MimeType=text/markdown;text/x-markdown;
Categories=Utility;
EOF
    ok "Created $DESKTOP_FILE"

    xdg-mime default o-mark.desktop text/markdown
    xdg-mime default o-mark.desktop text/x-markdown
    update-desktop-database "$HOME/.local/share/applications/"
    ok "Registered as default handler"

    detected=$(xdg-mime query default text/markdown || true)
    if [[ "$detected" == "o-mark.desktop" ]]; then
        ok "Verified: xdg-mime query default text/markdown → $detected"
    else
        echo "  ✗ xdg-mime query default text/markdown → ${detected:-'(none)'} (expected o-mark.desktop)"
        echo "    → check that .md files resolve to text/markdown, not text/plain:"
        echo "      xdg-mime query filetype /path/to/file.md"
    fi
    echo
fi

# =====================================================================
#  4. Verify
# =====================================================================
info "Verifying installation"
if command -v "$BIN_NAME" >/dev/null 2>&1; then
    ok "$("$BIN_NAME" --version 2>&1 || true)"
else
    echo "  ○ $BIN_NAME not found on PATH yet — restart your shell or update PATH as noted above"
fi

echo
echo "✓ O'Mark installed"
