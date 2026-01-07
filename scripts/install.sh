#!/usr/bin/env bash
#
# scmd installer script
# Detects OS and architecture, downloads the appropriate binary, and installs it
#
# Usage:
#   curl -fsSL https://scmd.sh/install.sh | bash
#   wget -qO- https://scmd.sh/install.sh | bash
#
# Options:
#   SCMD_INSTALL_DIR  - Installation directory (default: /usr/local/bin or ~/.local/bin)
#   SCMD_VERSION      - Version to install (default: latest)
#   SCMD_NO_SUDO      - Set to skip sudo (installs to ~/.local/bin)

set -e

# Configuration
REPO="scmd/scmd"
VERSION="${SCMD_VERSION:-latest}"
INSTALL_DIR="${SCMD_INSTALL_DIR:-}"
NO_SUDO="${SCMD_NO_SUDO:-false}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
info() {
    printf "${BLUE}==>${NC} %s\n" "$1"
}

success() {
    printf "${GREEN}==>${NC} %s\n" "$1"
}

warn() {
    printf "${YELLOW}Warning:${NC} %s\n" "$1"
}

error() {
    printf "${RED}Error:${NC} %s\n" "$1" >&2
    exit 1
}

# Detect OS
detect_os() {
    local os
    case "$(uname -s)" in
        Darwin*)
            os="macOS"
            ;;
        Linux*)
            os="linux"
            ;;
        MINGW*|MSYS*|CYGWIN*)
            os="windows"
            ;;
        *)
            error "Unsupported operating system: $(uname -s)"
            ;;
    esac
    echo "$os"
}

# Detect architecture
detect_arch() {
    local arch
    case "$(uname -m)" in
        x86_64|amd64)
            arch="amd64"
            ;;
        aarch64|arm64)
            arch="arm64"
            ;;
        armv7l|armv6l)
            arch="arm"
            ;;
        *)
            error "Unsupported architecture: $(uname -m)"
            ;;
    esac
    echo "$arch"
}

# Get latest version from GitHub
get_latest_version() {
    local latest_url="https://api.github.com/repos/${REPO}/releases/latest"

    if command -v curl &> /dev/null; then
        curl -sSfL "$latest_url" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
    elif command -v wget &> /dev/null; then
        wget -qO- "$latest_url" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
    else
        error "Neither curl nor wget found. Please install one of them."
    fi
}

# Download file
download() {
    local url="$1"
    local output="$2"

    if command -v curl &> /dev/null; then
        curl -sSfL "$url" -o "$output"
    elif command -v wget &> /dev/null; then
        wget -qO "$output" "$url"
    else
        error "Neither curl nor wget found. Please install one of them."
    fi
}

# Verify checksum
verify_checksum() {
    local archive="$1"
    local checksums_file="$2"
    local filename="$(basename "$archive")"

    info "Verifying checksum..."

    if command -v sha256sum &> /dev/null; then
        grep "$filename" "$checksums_file" | sha256sum -c - &> /dev/null
    elif command -v shasum &> /dev/null; then
        grep "$filename" "$checksums_file" | shasum -a 256 -c - &> /dev/null
    else
        warn "No checksum tool found (sha256sum or shasum). Skipping verification."
        return 0
    fi
}

# Determine installation directory
get_install_dir() {
    if [ -n "$INSTALL_DIR" ]; then
        echo "$INSTALL_DIR"
        return
    fi

    # Check if we should use sudo or user-local install
    if [ "$NO_SUDO" = "true" ] || [ ! -w "/usr/local/bin" ]; then
        mkdir -p "$HOME/.local/bin"
        echo "$HOME/.local/bin"
    else
        echo "/usr/local/bin"
    fi
}

# Check if directory is in PATH
check_path() {
    local dir="$1"

    if [[ ":$PATH:" != *":$dir:"* ]]; then
        warn "$dir is not in your PATH"
        info "Add it to your PATH by adding this line to your ~/.bashrc, ~/.zshrc, or ~/.profile:"
        echo ""
        echo "    export PATH=\"$dir:\$PATH\""
        echo ""
    fi
}

# Install shell completions
install_completions() {
    local binary="$1"

    info "Installing shell completions..."

    # Bash completion
    if [ -d "$HOME/.bash_completion.d" ]; then
        "$binary" completion bash > "$HOME/.bash_completion.d/scmd" 2>/dev/null || true
    elif [ -d "/usr/share/bash-completion/completions" ] && [ -w "/usr/share/bash-completion/completions" ]; then
        "$binary" completion bash | sudo tee /usr/share/bash-completion/completions/scmd > /dev/null 2>&1 || true
    fi

    # Zsh completion
    if [ -d "$HOME/.zsh/completion" ]; then
        "$binary" completion zsh > "$HOME/.zsh/completion/_scmd" 2>/dev/null || true
    elif [ -n "$ZSH" ] && [ -d "$ZSH/completions" ]; then
        "$binary" completion zsh > "$ZSH/completions/_scmd" 2>/dev/null || true
    fi

    # Fish completion
    if [ -d "$HOME/.config/fish/completions" ]; then
        "$binary" completion fish > "$HOME/.config/fish/completions/scmd.fish" 2>/dev/null || true
    fi
}

# Main installation function
main() {
    info "Installing scmd..."

    # Detect system
    local os=$(detect_os)
    local arch=$(detect_arch)
    info "Detected: $os $arch"

    # Get version
    if [ "$VERSION" = "latest" ]; then
        VERSION=$(get_latest_version)
        info "Latest version: $VERSION"
    fi

    # Construct download URLs
    local archive_name="scmd_${VERSION#v}_${os}_${arch}.tar.gz"
    if [ "$os" = "windows" ]; then
        archive_name="scmd_${VERSION#v}_${os}_${arch}.zip"
    fi

    local base_url="https://github.com/${REPO}/releases/download/${VERSION}"
    local archive_url="${base_url}/${archive_name}"
    local checksums_url="${base_url}/checksums.txt"

    # Create temporary directory
    local tmp_dir=$(mktemp -d)
    trap "rm -rf '$tmp_dir'" EXIT

    # Download archive
    info "Downloading scmd ${VERSION}..."
    download "$archive_url" "$tmp_dir/$archive_name"

    # Download and verify checksums
    download "$checksums_url" "$tmp_dir/checksums.txt"
    verify_checksum "$tmp_dir/$archive_name" "$tmp_dir/checksums.txt" || \
        error "Checksum verification failed"

    success "Checksum verified"

    # Extract archive
    info "Extracting..."
    if [ "$os" = "windows" ]; then
        unzip -q "$tmp_dir/$archive_name" -d "$tmp_dir"
    else
        tar -xzf "$tmp_dir/$archive_name" -C "$tmp_dir"
    fi

    # Determine installation directory
    local install_dir=$(get_install_dir)
    info "Installing to $install_dir..."

    # Install binary
    local binary_name="scmd"
    if [ "$os" = "windows" ]; then
        binary_name="scmd.exe"
    fi

    if [ "$install_dir" = "/usr/local/bin" ] && [ -w "/usr/local/bin" ]; then
        cp "$tmp_dir/$binary_name" "$install_dir/"
        chmod +x "$install_dir/$binary_name"
    elif [ "$install_dir" = "/usr/local/bin" ]; then
        sudo cp "$tmp_dir/$binary_name" "$install_dir/"
        sudo chmod +x "$install_dir/$binary_name"
    else
        mkdir -p "$install_dir"
        cp "$tmp_dir/$binary_name" "$install_dir/"
        chmod +x "$install_dir/$binary_name"
    fi

    success "scmd installed successfully!"

    # Check PATH
    check_path "$install_dir"

    # Install completions (if completion command exists)
    if "$install_dir/$binary_name" completion bash --help &> /dev/null; then
        install_completions "$install_dir/$binary_name"
    fi

    # Display next steps
    echo ""
    info "Next steps:"
    echo "  1. Verify installation: scmd --version"
    echo "  2. Install llama.cpp (for offline usage): brew install llama.cpp"
    echo "  3. Try it out: scmd /explain \"what is a goroutine\""
    echo ""
    info "Documentation: https://github.com/${REPO}"
    echo ""

    # Check for llama.cpp
    if ! command -v llama-server &> /dev/null; then
        warn "llama-server not found. For offline functionality, install llama.cpp:"
        echo "  - macOS: brew install llama.cpp"
        echo "  - Linux: Build from source at https://github.com/ggerganov/llama.cpp"
    fi
}

# Run main function
main "$@"
