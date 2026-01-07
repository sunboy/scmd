# Installation Guide for scmd

This guide covers all available installation methods for scmd, along with platform-specific instructions and troubleshooting tips.

## Table of Contents

- [Quick Start](#quick-start)
- [Installation Methods](#installation-methods)
  - [Homebrew (macOS/Linux)](#homebrew-macoslinux)
  - [npm (Cross-Platform)](#npm-cross-platform)
  - [Shell Script](#shell-script)
  - [Linux Package Managers](#linux-package-managers)
  - [Binary Download](#binary-download)
  - [Build from Source](#build-from-source)
- [Post-Installation](#post-installation)
- [Shell Integration](#shell-integration)
- [Troubleshooting](#troubleshooting)

## Quick Start

The fastest way to get started depends on your platform:

**macOS:**
```bash
brew install scmd/tap/scmd
```

**Linux (Debian/Ubuntu):**
```bash
curl -fsSL https://scmd.sh/install.sh | bash
```

**Any platform with npm:**
```bash
npm install -g scmd-cli
```

## Installation Methods

### Homebrew (macOS/Linux)

Homebrew is the recommended method for macOS and Linux users.

#### From Tap (Recommended)

```bash
# Add the scmd tap
brew tap scmd/tap

# Install scmd
brew install scmd

# Update scmd
brew upgrade scmd
```

#### Future: From Homebrew Core

Once scmd is added to Homebrew core, you can install directly:

```bash
brew install scmd
```

#### Includes

The Homebrew formula automatically:
- Installs the binary to `/usr/local/bin` (Intel) or `/opt/homebrew/bin` (Apple Silicon)
- Adds shell completions for bash, zsh, and fish
- Adds scmd to your PATH
- Suggests installing llama.cpp as an optional dependency

### npm (Cross-Platform)

npm provides cross-platform installation and automatic PATH configuration.

```bash
# Install globally
npm install -g scmd-cli

# Verify installation
scmd --version

# Update
npm update -g scmd-cli

# Uninstall
npm uninstall -g scmd-cli
```

#### How it works

The npm package:
1. Detects your OS and architecture
2. Downloads the appropriate binary from GitHub releases
3. Verifies checksums for security
4. Installs to npm's global bin directory (automatically in PATH)
5. Works on macOS, Linux, and Windows

#### Requirements

- Node.js 14.0.0 or later
- npm (comes with Node.js)

### Shell Script

The install script works on any Unix-like system with curl or wget.

#### Basic Installation

```bash
# Using curl (recommended)
curl -fsSL https://scmd.sh/install.sh | bash

# Using wget
wget -qO- https://scmd.sh/install.sh | bash
```

#### Custom Installation

```bash
# Install to custom directory
curl -fsSL https://scmd.sh/install.sh | SCMD_INSTALL_DIR=$HOME/bin bash

# Install without sudo (user-local install)
curl -fsSL https://scmd.sh/install.sh | SCMD_NO_SUDO=true bash

# Install specific version
curl -fsSL https://scmd.sh/install.sh | SCMD_VERSION=v1.0.0 bash
```

#### What the script does

1. Detects your OS and architecture
2. Downloads the appropriate binary and checksums
3. Verifies SHA256 checksums
4. Installs to `/usr/local/bin` (with sudo) or `~/.local/bin` (without sudo)
5. Makes the binary executable
6. Optionally installs shell completions

#### Uninstallation

```bash
curl -fsSL https://scmd.sh/uninstall.sh | bash
```

### Linux Package Managers

Native Linux packages provide integration with system package managers.

#### Debian/Ubuntu (apt)

```bash
# Download the .deb package
wget https://github.com/scmd/scmd/releases/download/v1.0.0/scmd_1.0.0_linux_amd64.deb

# Install
sudo dpkg -i scmd_1.0.0_linux_amd64.deb

# Or use apt
sudo apt install ./scmd_1.0.0_linux_amd64.deb

# Update (download new version and install)
sudo apt install ./scmd_1.1.0_linux_amd64.deb

# Uninstall
sudo apt remove scmd
```

#### Red Hat/Fedora/CentOS (rpm)

```bash
# Download the .rpm package
wget https://github.com/scmd/scmd/releases/download/v1.0.0/scmd_1.0.0_linux_amd64.rpm

# Install (Fedora/RHEL 8+)
sudo dnf install scmd_1.0.0_linux_amd64.rpm

# Install (CentOS 7/RHEL 7)
sudo yum install scmd_1.0.0_linux_amd64.rpm

# Update
sudo dnf upgrade scmd_1.1.0_linux_amd64.rpm

# Uninstall
sudo dnf remove scmd
```

#### Alpine Linux (apk)

```bash
# Download the .apk package
wget https://github.com/scmd/scmd/releases/download/v1.0.0/scmd_1.0.0_linux_amd64.apk

# Install
sudo apk add --allow-untrusted scmd_1.0.0_linux_amd64.apk

# Uninstall
sudo apk del scmd
```

#### What's included

Linux packages include:
- Binary installed to `/usr/bin/scmd`
- Shell completions for bash, zsh, and fish
- Man pages (future)
- Post-install scripts that preserve user data

### Binary Download

Download pre-built binaries from [GitHub Releases](https://github.com/scmd/scmd/releases).

#### Steps

1. Go to the [latest release](https://github.com/scmd/scmd/releases/latest)
2. Download the appropriate archive for your platform:
   - **macOS (Intel)**: `scmd_VERSION_macOS_amd64.tar.gz`
   - **macOS (Apple Silicon)**: `scmd_VERSION_macOS_arm64.tar.gz`
   - **Linux (x64)**: `scmd_VERSION_linux_amd64.tar.gz`
   - **Linux (ARM64)**: `scmd_VERSION_linux_arm64.tar.gz`
   - **Windows (x64)**: `scmd_VERSION_windows_amd64.zip`

3. Verify checksum (optional but recommended):
   ```bash
   # Download checksums
   wget https://github.com/scmd/scmd/releases/download/v1.0.0/checksums.txt

   # Verify (macOS/Linux)
   shasum -a 256 -c checksums.txt 2>&1 | grep scmd_1.0.0_macOS_arm64.tar.gz
   ```

4. Extract the archive:
   ```bash
   # macOS/Linux
   tar -xzf scmd_VERSION_macOS_arm64.tar.gz

   # Windows
   unzip scmd_VERSION_windows_amd64.zip
   ```

5. Move to a directory in your PATH:
   ```bash
   # macOS/Linux
   sudo mv scmd /usr/local/bin/

   # Or user-local
   mkdir -p ~/.local/bin
   mv scmd ~/.local/bin/
   export PATH="$HOME/.local/bin:$PATH"  # Add to ~/.bashrc or ~/.zshrc
   ```

### Build from Source

For development or if pre-built binaries aren't available for your platform.

#### Requirements

- Go 1.24 or later
- Git
- Make (optional but recommended)

#### Steps

```bash
# Clone the repository
git clone https://github.com/scmd/scmd.git
cd scmd

# Build using Make
make build

# Or build directly with Go
go build -o scmd ./cmd/scmd

# Install to /usr/local/bin
sudo make install

# Or install to GOPATH/bin
make install-go
```

#### Build for all platforms

```bash
make build-all
```

This creates binaries in `dist/` for:
- darwin/amd64, darwin/arm64
- linux/amd64, linux/arm64
- windows/amd64

## Post-Installation

### Verify Installation

```bash
# Check version
scmd --version

# Check available backends
scmd backends
```

### Install llama.cpp (for offline usage)

scmd works offline by default with llama.cpp:

**macOS:**
```bash
brew install llama.cpp
```

**Linux (Ubuntu/Debian):**
```bash
# Build from source
git clone https://github.com/ggerganov/llama.cpp
cd llama.cpp
make
sudo make install
```

**Verify llama.cpp:**
```bash
which llama-server
llama-server --version
```

### First Run

On first run, scmd will automatically download the default model (Qwen3-4B, ~2.6GB):

```bash
scmd /explain "what is a goroutine"
```

## Shell Integration

Enable slash commands directly in your shell (optional but recommended).

### Bash

Add to `~/.bashrc`:

```bash
eval "$(scmd slash init bash)"
```

### Zsh

Add to `~/.zshrc`:

```bash
eval "$(scmd slash init zsh)"
```

### Fish

Add to `~/.config/fish/config.fish`:

```fish
scmd slash init fish | source
```

### Reload Shell

```bash
# Bash/Zsh
source ~/.bashrc  # or ~/.zshrc

# Fish
source ~/.config/fish/config.fish
```

### Test Shell Integration

```bash
/explain "what is a goroutine"
git diff | /gc
```

## Troubleshooting

### Command not found: scmd

**Issue**: Shell can't find the scmd binary.

**Solutions**:

1. Check if scmd is installed:
   ```bash
   which scmd
   ```

2. If installed but not in PATH, add the installation directory to PATH:
   ```bash
   # For ~/.local/bin
   echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
   source ~/.bashrc
   ```

3. For Homebrew on Apple Silicon:
   ```bash
   echo 'eval "$(/opt/homebrew/bin/brew shellenv)"' >> ~/.zprofile
   eval "$(/opt/homebrew/bin/brew shellenv)"
   ```

### Permission denied

**Issue**: Binary is not executable.

**Solution**:
```bash
chmod +x /path/to/scmd
```

### llama-server not found

**Issue**: Offline functionality requires llama.cpp.

**Solution**: Install llama.cpp (see [Post-Installation](#install-llamacpp-for-offline-usage))

### Models not downloading

**Issue**: Firewall or network restrictions.

**Solutions**:

1. Check network connection:
   ```bash
   curl -I https://huggingface.co
   ```

2. Manually download models:
   ```bash
   scmd models pull qwen3-4b
   ```

3. Use alternative model:
   ```bash
   scmd models list
   scmd models pull qwen2.5-1.5b  # Smaller model
   ```

### npm installation fails

**Issue**: npm can't download binaries from GitHub.

**Solutions**:

1. Check GitHub connectivity:
   ```bash
   curl -I https://github.com
   ```

2. Try with verbose logging:
   ```bash
   npm install -g scmd-cli --verbose
   ```

3. Install from binary instead (see [Binary Download](#binary-download))

### Shell completion not working

**Issue**: Completions not loaded.

**Solutions**:

1. Check if completions are installed:
   ```bash
   # Bash
   ls /usr/share/bash-completion/completions/scmd

   # Zsh
   ls /usr/local/share/zsh/site-functions/_scmd
   ```

2. Manually source completions:
   ```bash
   # Bash
   source <(scmd completion bash)

   # Zsh
   source <(scmd completion zsh)
   ```

3. Regenerate completions:
   ```bash
   scmd completion bash > ~/.bash_completion.d/scmd
   ```

## Directory Structure

scmd uses the following directories:

```
~/.scmd/
├── config.yaml          # Configuration file
├── models/             # Downloaded LLM models
│   ├── qwen3-4b/
│   └── ...
├── cache/              # Cached responses and data
├── repos/              # Command repositories
└── logs/               # Application logs
```

### XDG Base Directory Support

scmd respects XDG environment variables if set:

- `$XDG_CONFIG_HOME/scmd/` - Configuration
- `$XDG_DATA_HOME/scmd/` - Models and data
- `$XDG_CACHE_HOME/scmd/` - Cache

## Updating scmd

### Homebrew

```bash
brew upgrade scmd
```

### npm

```bash
npm update -g scmd-cli
```

### Shell Script

Re-run the install script:

```bash
curl -fsSL https://scmd.sh/install.sh | bash
```

### Manual

Download the new version and replace the old binary.

## Uninstalling scmd

### Homebrew

```bash
brew uninstall scmd
```

### npm

```bash
npm uninstall -g scmd-cli

# Remove data (optional)
rm -rf ~/.scmd
```

### Shell Script

```bash
curl -fsSL https://scmd.sh/uninstall.sh | bash
```

### Linux Packages

```bash
# Debian/Ubuntu
sudo apt remove scmd

# Fedora/RHEL
sudo dnf remove scmd

# Remove data (optional)
rm -rf ~/.scmd
```

### Manual

```bash
# Remove binary
sudo rm /usr/local/bin/scmd

# Remove completions
sudo rm /usr/share/bash-completion/completions/scmd
sudo rm /usr/local/share/zsh/site-functions/_scmd
sudo rm /usr/share/fish/vendor_completions.d/scmd.fish

# Remove data (optional)
rm -rf ~/.scmd
```

## Getting Help

- **Documentation**: https://github.com/scmd/scmd
- **Issues**: https://github.com/scmd/scmd/issues
- **Discussions**: https://github.com/scmd/scmd/discussions

---

For more information, see the [README](README.md) or visit the [project homepage](https://github.com/scmd/scmd).
