#!/bin/bash

set -e

# Configuration
REPO="blue-monads/potatoverse"
INSTALL_DIR="${HOME}/.local/bin"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored messages
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Fetch latest version from GitHub
get_latest_version() {
    print_info "Fetching latest version from GitHub..."
    
    if command -v curl &> /dev/null; then
        VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"v?([^"]+)".*/\1/')
    elif command -v wget &> /dev/null; then
        VERSION=$(wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"v?([^"]+)".*/\1/')
    else
        print_error "Neither curl nor wget found. Please install one of them."
        exit 1
    fi
    
    if [ -z "$VERSION" ]; then
        print_error "Could not fetch latest version from GitHub"
        exit 1
    fi
    
    print_info "Latest version: ${VERSION}"
}

# Detect OS and Architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    case "$OS" in
        linux*)
            OS="linux"
            ;;
        darwin*)
            OS="darwin"
            ;;
        mingw* | msys* | cygwin*)
            OS="windows"
            ;;
        *)
            print_error "Unsupported operating system: $OS"
            exit 1
            ;;
    esac
    
    case "$ARCH" in
        x86_64 | amd64)
            ARCH="amd64"
            ;;
        aarch64 | arm64)
            ARCH="arm64"
            ;;
        *)
            print_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac
    
    print_info "Detected platform: ${OS}_${ARCH}"
}

# Download and extract binary
download_and_install() {
    BINARY_NAME="potatoverse_${VERSION}_${OS}_${ARCH}.tar.gz"
    # Check if VERSION already has 'v' prefix
    
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY_NAME}"

    
    print_info "Downloading ${BINARY_NAME}..."
    print_info "URL: ${DOWNLOAD_URL}"
    
    # Create temporary directory
    TMP_DIR=$(mktemp -d)
    trap "rm -rf ${TMP_DIR}" EXIT
    
    # Download the archive
    if command -v curl &> /dev/null; then
        curl -fsSL "${DOWNLOAD_URL}" -o "${TMP_DIR}/${BINARY_NAME}"
    elif command -v wget &> /dev/null; then
        wget -q "${DOWNLOAD_URL}" -O "${TMP_DIR}/${BINARY_NAME}"
    else
        print_error "Neither curl nor wget found. Please install one of them."
        exit 1
    fi
    
    print_info "Extracting archive..."
    tar -xzf "${TMP_DIR}/${BINARY_NAME}" -C "${TMP_DIR}"
    
    # Create install directory if it doesn't exist
    mkdir -p "${INSTALL_DIR}"
    
    # Find and move the binary
    BINARY_FILE=$(find "${TMP_DIR}" -type f -name "potatoverse*" ! -name "*.tar.gz" | head -n 1)
    
    if [ -z "$BINARY_FILE" ]; then
        print_error "Could not find potatoverse binary in archive"
        exit 1
    fi
    
    print_info "Installing to ${INSTALL_DIR}/potatoverse..."
    mv "${BINARY_FILE}" "${INSTALL_DIR}/potatoverse"
    chmod +x "${INSTALL_DIR}/potatoverse"
    
    print_info "✓ Installation complete!"
}

# Check if binary is in PATH
check_path() {
    if [[ ":$PATH:" != *":${INSTALL_DIR}:"* ]]; then
        print_warning "${INSTALL_DIR} is not in your PATH"
        print_warning "Add this line to your ~/.bashrc or ~/.zshrc:"
        echo ""
        echo "    export PATH=\"${INSTALL_DIR}:\$PATH\""
        echo ""
    fi
}

# Main execution
main() {
    echo "========================================="
    echo "  Potatoverse Installer"
    echo "========================================="
    echo ""
    
    get_latest_version
    detect_platform
    download_and_install
    check_path
    
    echo ""
    print_info "Run 'potatoverse --help' to verify installation"
    
    # Ask if user wants to run it now
    read -p "Do you want to run potatoverse now? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        print_info "Executing potatoverse..."
        "${INSTALL_DIR}/potatoverse" server init-and-start
    else
        print_info "You can run potatoverse later with: ${INSTALL_DIR}/potatoverse server init-and-start"
    fi
}

main "$@"
