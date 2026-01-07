#!/usr/bin/env bash
#
# scmd uninstaller script
# Removes scmd binary and associated files
#
# Usage:
#   ./uninstall.sh
#   curl -fsSL https://scmd.sh/uninstall.sh | bash

set -e

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
}

# Confirm uninstallation
confirm() {
    read -p "Are you sure you want to uninstall scmd? [y/N] " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        info "Uninstallation cancelled."
        exit 0
    fi
}

# Find scmd binary
find_binary() {
    local binary_path

    # Check common locations
    for path in "/usr/local/bin/scmd" "$HOME/.local/bin/scmd" "$GOPATH/bin/scmd" "$HOME/go/bin/scmd"; do
        if [ -f "$path" ]; then
            echo "$path"
            return 0
        fi
    done

    # Check if scmd is in PATH
    if command -v scmd &> /dev/null; then
        which scmd
        return 0
    fi

    return 1
}

# Remove binary
remove_binary() {
    local binary="$1"
    local dir=$(dirname "$binary")

    info "Removing binary: $binary"

    if [ -w "$dir" ]; then
        rm -f "$binary"
    else
        sudo rm -f "$binary"
    fi

    success "Binary removed"
}

# Remove shell completions
remove_completions() {
    info "Removing shell completions..."

    # Bash
    rm -f "$HOME/.bash_completion.d/scmd" 2>/dev/null || true
    sudo rm -f /usr/share/bash-completion/completions/scmd 2>/dev/null || true
    sudo rm -f /etc/bash_completion.d/scmd 2>/dev/null || true

    # Zsh
    rm -f "$HOME/.zsh/completion/_scmd" 2>/dev/null || true
    if [ -n "$ZSH" ]; then
        rm -f "$ZSH/completions/_scmd" 2>/dev/null || true
    fi
    sudo rm -f /usr/share/zsh/site-functions/_scmd 2>/dev/null || true
    sudo rm -f /usr/local/share/zsh/site-functions/_scmd 2>/dev/null || true

    # Fish
    rm -f "$HOME/.config/fish/completions/scmd.fish" 2>/dev/null || true
    sudo rm -f /usr/share/fish/vendor_completions.d/scmd.fish 2>/dev/null || true

    success "Completions removed"
}

# Ask about data directory
remove_data() {
    echo ""
    read -p "Do you want to remove scmd data directory (~/.scmd)? This includes config, models, and cache. [y/N] " -n 1 -r
    echo

    if [[ $REPLY =~ ^[Yy]$ ]]; then
        if [ -d "$HOME/.scmd" ]; then
            info "Removing data directory..."
            du -sh "$HOME/.scmd" 2>/dev/null || true
            rm -rf "$HOME/.scmd"
            success "Data directory removed"
        else
            info "Data directory not found"
        fi
    else
        info "Keeping data directory at ~/.scmd"
        info "You can manually remove it later with: rm -rf ~/.scmd"
    fi
}

# Remove shell integration
remove_shell_integration() {
    info "Checking for shell integration..."

    local files_to_check=(
        "$HOME/.bashrc"
        "$HOME/.bash_profile"
        "$HOME/.zshrc"
        "$HOME/.config/fish/config.fish"
    )

    local found=false

    for file in "${files_to_check[@]}"; do
        if [ -f "$file" ] && grep -q "scmd slash init" "$file"; then
            warn "Found scmd shell integration in $file"
            found=true
        fi
    done

    if [ "$found" = true ]; then
        echo ""
        info "Please manually remove scmd shell integration from your shell config files."
        info "Look for lines containing: scmd slash init"
    fi
}

# Main uninstallation function
main() {
    info "scmd uninstaller"
    echo ""

    # Confirm
    confirm

    # Find and remove binary
    local binary=$(find_binary)
    if [ -n "$binary" ]; then
        remove_binary "$binary"
    else
        warn "scmd binary not found in common locations"
        info "If installed, you may need to remove it manually"
    fi

    # Remove completions
    remove_completions

    # Check for shell integration
    remove_shell_integration

    # Ask about data directory
    remove_data

    echo ""
    success "scmd has been uninstalled!"
    info "Thank you for using scmd. We'd love to hear your feedback:"
    info "https://github.com/scmd/scmd/issues"
    echo ""
}

# Run main function
main "$@"
