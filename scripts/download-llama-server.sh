#!/bin/bash
# Download llama-server binaries for bundling with scmd
# This runs during goreleaser build
# Compatible with bash 3.2+ (macOS default)

set -e

VERSION="b7688"  # llama.cpp release tag
BASE_URL="https://github.com/ggml-org/llama.cpp/releases/download/${VERSION}"

mkdir -p dist/llama-server

# Platform and file arrays (bash 3.2 compatible)
PLATFORMS=(darwin-amd64 darwin-arm64 linux-amd64 windows-amd64)
FILES=(
    "llama-${VERSION}-bin-macos-x64.tar.gz"
    "llama-${VERSION}-bin-macos-arm64.tar.gz"
    "llama-${VERSION}-bin-ubuntu-x64.tar.gz"
    "llama-${VERSION}-bin-win-cpu-x64.zip"
)

# Download for each platform
for i in "${!PLATFORMS[@]}"; do
    platform="${PLATFORMS[$i]}"
    file="${FILES[$i]}"

    echo "Downloading llama-server for $platform..."

    # Create platform-specific directory
    mkdir -p "dist/llama-server/$platform"

    # Determine file extension
    if [[ "$file" == *.tar.gz ]]; then
        archive_ext="tar.gz"
    else
        archive_ext="zip"
    fi

    # Download archive
    if command -v curl &> /dev/null; then
        curl -fsSL "${BASE_URL}/${file}" -o "dist/llama-server/${platform}.${archive_ext}"
    else
        wget -q "${BASE_URL}/${file}" -O "dist/llama-server/${platform}.${archive_ext}"
    fi

    # Extract based on file type
    if [[ "$archive_ext" == "tar.gz" ]]; then
        tar -xzf "dist/llama-server/${platform}.${archive_ext}" -C "dist/llama-server/$platform"
    else
        unzip -q "dist/llama-server/${platform}.${archive_ext}" -d "dist/llama-server/$platform"
    fi

    # Find and rename llama-server binary
    if [[ "$platform" == windows-* ]]; then
        find "dist/llama-server/$platform" -name "llama-server.exe" -exec mv {} "dist/llama-server/$platform/" \;
    else
        find "dist/llama-server/$platform" -name "llama-server" -exec mv {} "dist/llama-server/$platform/" \;
        chmod +x "dist/llama-server/$platform/llama-server"
    fi

    # Clean up extracted files (keep only llama-server)
    find "dist/llama-server/$platform" -mindepth 1 -maxdepth 1 ! -name "llama-server*" -exec rm -rf {} \;
    rm "dist/llama-server/${platform}.${archive_ext}"

    echo "✓ Downloaded llama-server for $platform"
done

echo ""
echo "All llama-server binaries downloaded successfully!"
echo "Location: dist/llama-server/"
