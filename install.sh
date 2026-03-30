#!/bin/bash
set -euo pipefail

CORRIDOR_VERBOSE="${CORRIDOR_VERBOSE:-0}"

log_verbose() {
    if [ "$CORRIDOR_VERBOSE" = "1" ]; then
        echo "$@"
    fi
}

log_error() {
    echo "Error: $*" >&2
}

# --- Detect platform ---

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
    x86_64)       ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *)
        log_error "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

case "$OS" in
    linux|darwin) ;;
    *)
        log_error "Unsupported OS: $OS"
        exit 1
        ;;
esac

log_verbose "Installing Corridor CLI..."
log_verbose "Detected platform: ${OS}/${ARCH}"

# --- Resolve latest version ---

log_verbose "Fetching latest version..."
VERSION=$(curl -fsSL "https://app.corridor.dev/cli/latest-version" 2>/dev/null || echo "")
if [ -z "$VERSION" ]; then
    log_error "Could not fetch latest version"
    exit 1
fi
log_verbose "Latest version: ${VERSION}"

# --- Download ---

FILENAME="corridor_${OS}_${ARCH}.tar.gz"
DOWNLOAD_URL="https://releases.corridor.dev/cli/${VERSION}/${FILENAME}"
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

log_verbose "Downloading ${FILENAME}..."
if ! curl -fsSL "$DOWNLOAD_URL" -o "${TMP_DIR}/${FILENAME}"; then
    log_error "Download failed"
    exit 1
fi

# --- Verify checksum ---

log_verbose "Verifying checksum..."
CHECKSUM_URL="https://releases.corridor.dev/cli/${VERSION}/checksums.txt"
if curl -fsSL "$CHECKSUM_URL" -o "${TMP_DIR}/checksums.txt" 2>/dev/null; then
    if ! (cd "$TMP_DIR" && grep "$FILENAME" checksums.txt | sha256sum -c --quiet); then
        log_error "Checksum verification failed"
        exit 1
    fi
    log_verbose "Checksum verified."
fi

# --- Extract ---

INSTALL_DIR="${HOME}/.corridor/bin"
BIN_PATH="${INSTALL_DIR}/corridor"

log_verbose "Extracting..."
mkdir -p "$INSTALL_DIR"
tar -xzf "${TMP_DIR}/${FILENAME}" -C "$INSTALL_DIR"
chmod +x "$BIN_PATH"

# --- Symlink (if ~/.local/bin exists) ---

if [ -d "${HOME}/.local/bin" ]; then
    ln -sf "$BIN_PATH" "${HOME}/.local/bin/corridor"
    log_verbose "Symlinked ~/.local/bin/corridor -> ${BIN_PATH}"
fi

# --- Default output ---

echo "Corridor CLI ${VERSION} installed to ${BIN_PATH}"

# PATH check
case ":${PATH}:" in
    *":${INSTALL_DIR}:"*) ;;
    *)
        echo ""
        echo "Warning: ${INSTALL_DIR} is not in your PATH."
        echo "Add this to your shell profile:"
        echo ""
        echo "  export PATH=\"\${HOME}/.corridor/bin:\$PATH\""
        echo ""
        ;;
esac

# --- Run corridor install ---

echo "Running 'corridor install'..."
export CORRIDOR_VERBOSE
"$BIN_PATH" install
