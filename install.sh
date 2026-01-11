#!/bin/bash
set -e

# shippost installer
# Usage: curl -fsSL https://raw.githubusercontent.com/tomswokowski/shippost/main/install.sh | bash

REPO="tomswokowski/shippost"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="shippost"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

info() {
    echo -e "${GREEN}==>${NC} $1"
}

warn() {
    echo -e "${YELLOW}warning:${NC} $1"
}

error() {
    echo -e "${RED}error:${NC} $1"
    exit 1
}

# Detect OS
detect_os() {
    case "$(uname -s)" in
        Linux*)  echo "Linux" ;;
        Darwin*) echo "Darwin" ;;
        MINGW*|MSYS*|CYGWIN*) echo "Windows" ;;
        *) error "Unsupported operating system: $(uname -s)" ;;
    esac
}

# Detect architecture
detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64) echo "x86_64" ;;
        arm64|aarch64) echo "arm64" ;;
        *) error "Unsupported architecture: $(uname -m)" ;;
    esac
}

# Get latest release version from GitHub
get_latest_version() {
    curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" |
        grep '"tag_name":' |
        sed -E 's/.*"([^"]+)".*/\1/'
}

main() {
    echo ""
    echo "  shippost installer"
    echo "  ===================="
    echo ""

    # Check for required tools
    if ! command -v curl &> /dev/null; then
        error "curl is required but not installed"
    fi

    OS=$(detect_os)
    ARCH=$(detect_arch)

    info "Detected OS: $OS, Arch: $ARCH"

    # Get latest version
    info "Fetching latest version..."
    VERSION=$(get_latest_version)

    if [ -z "$VERSION" ]; then
        error "Could not determine latest version"
    fi

    info "Latest version: $VERSION"

    # Build download URL
    if [ "$OS" = "Windows" ]; then
        FILENAME="${BINARY_NAME}_${VERSION#v}_${OS}_${ARCH}.zip"
    else
        FILENAME="${BINARY_NAME}_${VERSION#v}_${OS}_${ARCH}.tar.gz"
    fi

    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILENAME}"

    # Create temp directory
    TMP_DIR=$(mktemp -d)
    trap "rm -rf $TMP_DIR" EXIT

    # Download
    info "Downloading ${FILENAME}..."
    if ! curl -fsSL "$DOWNLOAD_URL" -o "${TMP_DIR}/${FILENAME}"; then
        error "Failed to download from ${DOWNLOAD_URL}"
    fi

    # Extract
    info "Extracting..."
    cd "$TMP_DIR"
    if [ "$OS" = "Windows" ]; then
        unzip -q "$FILENAME"
    else
        tar -xzf "$FILENAME"
    fi

    # Install
    info "Installing to ${INSTALL_DIR}..."
    if [ -w "$INSTALL_DIR" ]; then
        mv "$BINARY_NAME" "$INSTALL_DIR/"
    else
        sudo mv "$BINARY_NAME" "$INSTALL_DIR/"
    fi

    # Verify installation
    if command -v "$BINARY_NAME" &> /dev/null; then
        echo ""
        info "Successfully installed shippost!"
        echo ""
        echo "  Run 'shippost --setup' to configure your X API credentials"
        echo "  Run 'shippost' to start posting"
        echo ""
    else
        warn "Installed to ${INSTALL_DIR}/${BINARY_NAME} but it's not in your PATH"
        echo "  Add ${INSTALL_DIR} to your PATH, or run directly:"
        echo "  ${INSTALL_DIR}/${BINARY_NAME}"
    fi
}

main
