# Installation

scmd works offline by default using llama.cpp and local Qwen models. This guide covers installation on all supported platforms using multiple installation methods.

## Quick Install

Choose the installation method that works best for you:

=== "Homebrew (macOS/Linux)"

    The easiest way to install on macOS and Linux:

    ```bash
    # Add the scmd tap
    brew tap scmd/tap

    # Install scmd
    brew install scmd

    # Verify installation
    scmd --version

    # Install llama.cpp for offline usage
    brew install llama.cpp
    ```

    Homebrew automatically:

    - Installs the binary to your PATH
    - Adds shell completions for bash, zsh, and fish
    - Manages updates via `brew upgrade scmd`

=== "npm (Cross-Platform)"

    Works on any platform with Node.js:

    ```bash
    # Install globally
    npm install -g scmd-cli

    # Verify installation
    scmd --version

    # Install llama.cpp for offline usage
    # macOS:
    brew install llama.cpp
    # Linux: build from source (see below)
    ```

    The npm package:

    - Downloads the correct binary for your platform
    - Automatically adds scmd to your PATH
    - Works on macOS, Linux, and Windows

=== "Shell Script (wget/curl)"

    Universal installer for Unix-like systems:

    ```bash
    # Using curl (recommended)
    curl -fsSL https://scmd.sh/install.sh | bash

    # Using wget
    wget -qO- https://scmd.sh/install.sh | bash

    # Verify installation
    scmd --version
    ```

    The install script:

    - Auto-detects your OS and architecture
    - Verifies checksums for security
    - Installs to `/usr/local/bin` or `~/.local/bin`
    - Sets up shell completions

    **Custom installation:**

    ```bash
    # Install to custom directory
    curl -fsSL https://scmd.sh/install.sh | SCMD_INSTALL_DIR=$HOME/bin bash

    # Install without sudo (user-local)
    curl -fsSL https://scmd.sh/install.sh | SCMD_NO_SUDO=true bash

    # Install specific version
    curl -fsSL https://scmd.sh/install.sh | SCMD_VERSION=v1.0.0 bash
    ```

=== "Linux Packages"

    Native packages for Debian, Red Hat, and Alpine:

    **Debian/Ubuntu (apt):**
    ```bash
    # Download and install
    wget https://github.com/scmd/scmd/releases/latest/download/scmd_VERSION_linux_amd64.deb
    sudo dpkg -i scmd_VERSION_linux_amd64.deb

    # Or use apt for dependency resolution
    sudo apt install ./scmd_VERSION_linux_amd64.deb

    # Verify
    scmd --version
    ```

    **Red Hat/Fedora/CentOS (rpm):**
    ```bash
    # Download and install
    wget https://github.com/scmd/scmd/releases/latest/download/scmd_VERSION_linux_amd64.rpm

    # Fedora/RHEL 8+
    sudo dnf install scmd_VERSION_linux_amd64.rpm

    # CentOS 7/RHEL 7
    sudo yum install scmd_VERSION_linux_amd64.rpm

    # Verify
    scmd --version
    ```

    **Alpine Linux (apk):**
    ```bash
    wget https://github.com/scmd/scmd/releases/latest/download/scmd_VERSION_linux_amd64.apk
    sudo apk add --allow-untrusted scmd_VERSION_linux_amd64.apk
    ```

    Linux packages include:

    - Binary in `/usr/bin/scmd`
    - Shell completions (bash, zsh, fish)
    - Integration with system package manager

=== "Binary Download"

    Download pre-built binaries from GitHub:

    1. Visit [GitHub Releases](https://github.com/scmd/scmd/releases/latest)
    2. Download the archive for your platform:
       - macOS (Intel): `scmd_VERSION_macOS_amd64.tar.gz`
       - macOS (Apple Silicon): `scmd_VERSION_macOS_arm64.tar.gz`
       - Linux (x64): `scmd_VERSION_linux_amd64.tar.gz`
       - Linux (ARM64): `scmd_VERSION_linux_arm64.tar.gz`
       - Windows (x64): `scmd_VERSION_windows_amd64.zip`

    3. Extract and install:
       ```bash
       # macOS/Linux
       tar -xzf scmd_VERSION_macOS_arm64.tar.gz
       sudo mv scmd /usr/local/bin/

       # Windows (PowerShell)
       Expand-Archive scmd_VERSION_windows_amd64.zip
       # Add to PATH
       ```

    4. Verify checksums (recommended):
       ```bash
       wget https://github.com/scmd/scmd/releases/download/v1.0.0/checksums.txt
       shasum -a 256 -c checksums.txt 2>&1 | grep scmd
       ```

=== "Build from Source"

    For developers or custom builds:

    ```bash
    # Prerequisites: Go 1.24 or later

    # Clone the repository
    git clone https://github.com/scmd/scmd
    cd scmd

    # Build using Makefile
    make build

    # Install to /usr/local/bin
    sudo make install

    # Or install to $GOPATH/bin
    make install-go

    # Or build with Go directly
    go build -o scmd ./cmd/scmd

    # Verify
    ./scmd --version
    ```

## Prerequisites

### llama.cpp (for offline usage)

scmd requires llama.cpp for offline inference:

=== "macOS"

    ```bash
    brew install llama.cpp

    # Verify
    which llama-server
    llama-server --version
    ```

=== "Linux"

    ```bash
    # Ubuntu/Debian - from package manager (if available)
    sudo apt install llama-cpp

    # Or build from source (recommended for latest version)
    git clone https://github.com/ggerganov/llama.cpp
    cd llama.cpp
    mkdir build && cd build

    # For NVIDIA GPU support
    cmake .. -DLLAMA_CUDA=ON

    # For CPU only
    # cmake ..

    cmake --build . --config Release
    sudo cp bin/llama-server /usr/local/bin/

    # Verify
    which llama-server
    llama-server --version
    ```

=== "Windows"

    Download pre-built binaries or build from source:

    1. Visit [llama.cpp releases](https://github.com/ggerganov/llama.cpp/releases)
    2. Download Windows binaries
    3. Add to PATH

    Or build with CMake:
    ```powershell
    git clone https://github.com/ggerganov/llama.cpp
    cd llama.cpp
    mkdir build
    cd build
    cmake ..
    cmake --build . --config Release
    ```

## Post-Installation

### 1. Verify Installation

```bash
# Check scmd version
scmd --version

# Check available backends
scmd backends
```

Expected output:
```
Available backends:

✓ llamacpp     qwen3-4b
✗ ollama       (not running)
✗ openai       (not configured)
```

### 2. First Run (Model Download)

On first use, scmd will automatically download the default model (~2.6GB):

```bash
scmd /explain "what is a channel in Go?"
```

Output:
```
[INFO] First run detected
[INFO] Downloading qwen3-4b model (2.6 GB)...
[INFO] Progress: ████████████████████ 100%
[INFO] Model downloaded to ~/.scmd/models/qwen3-4b-Q4_K_M.gguf
[INFO] Starting llama-server...

A channel in Go is a typed conduit through which you can send
and receive values with the channel operator <-...
```

### 3. Set Up Shell Completions (Optional)

Enable tab completion for scmd commands:

=== "Bash"

    ```bash
    # Generate completion script
    scmd completion bash > /tmp/scmd-completion.bash

    # Install for current user
    mkdir -p ~/.bash_completion.d
    mv /tmp/scmd-completion.bash ~/.bash_completion.d/scmd

    # Or install system-wide (requires sudo)
    sudo scmd completion bash > /etc/bash_completion.d/scmd

    # Reload
    source ~/.bashrc
    ```

=== "Zsh"

    ```bash
    # Enable completion system
    echo "autoload -U compinit; compinit" >> ~/.zshrc

    # Install completion
    scmd completion zsh > "${fpath[1]}/_scmd"

    # Reload
    source ~/.zshrc
    ```

=== "Fish"

    ```bash
    # Install completion
    scmd completion fish > ~/.config/fish/completions/scmd.fish

    # Reload
    source ~/.config/fish/config.fish
    ```

## Directory Structure

scmd uses the following directory structure:

```
~/.scmd/
├── config.yaml          # Configuration file
├── slash.yaml           # Slash command mappings
├── repos.json           # Repository list
├── models/              # Downloaded GGUF models
│   ├── qwen3-4b-Q4_K_M.gguf
│   └── qwen2.5-3b-Q4_K_M.gguf
├── commands/            # Installed command specs
│   ├── git-commit.yaml
│   └── explain.yaml
└── cache/               # Cached manifests
    └── official/
        └── manifest.yaml
```

### XDG Base Directory Support

scmd respects XDG environment variables if set:

```bash
export XDG_CONFIG_HOME=~/.config
export XDG_DATA_HOME=~/.local/share
export XDG_CACHE_HOME=~/.cache

# scmd will use:
# - $XDG_CONFIG_HOME/scmd/ for config
# - $XDG_DATA_HOME/scmd/ for models and data
# - $XDG_CACHE_HOME/scmd/ for cache
```

Or customize the data directory:

```bash
export SCMD_DATA_DIR=/path/to/custom/dir
scmd /explain "test"
```

## Updating scmd

=== "Homebrew"

    ```bash
    brew upgrade scmd
    ```

=== "npm"

    ```bash
    npm update -g scmd-cli
    ```

=== "Shell Script"

    Re-run the install script:
    ```bash
    curl -fsSL https://scmd.sh/install.sh | bash
    ```

=== "Linux Packages"

    ```bash
    # Debian/Ubuntu
    sudo apt upgrade scmd

    # Fedora/RHEL
    sudo dnf upgrade scmd
    ```

=== "Source"

    ```bash
    cd scmd
    git pull origin main
    make build
    sudo make install
    ```

## Uninstalling scmd

=== "Homebrew"

    ```bash
    brew uninstall scmd

    # Remove data (optional)
    rm -rf ~/.scmd
    ```

=== "npm"

    ```bash
    npm uninstall -g scmd-cli

    # Remove data (optional)
    rm -rf ~/.scmd
    ```

=== "Shell Script"

    ```bash
    # Use the uninstall script
    curl -fsSL https://scmd.sh/uninstall.sh | bash

    # Or manually
    sudo rm /usr/local/bin/scmd
    rm -rf ~/.scmd
    ```

=== "Linux Packages"

    ```bash
    # Debian/Ubuntu
    sudo apt remove scmd

    # Fedora/RHEL
    sudo dnf remove scmd

    # Remove data (optional)
    rm -rf ~/.scmd
    ```

## Troubleshooting

For detailed troubleshooting, see the [Troubleshooting Guide](../user-guide/troubleshooting.md).

### Common Issues

#### Command not found: scmd

**Issue**: Shell can't find the scmd binary.

**Solution**: Add scmd's installation directory to PATH:

```bash
# For ~/.local/bin
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# For Homebrew on Apple Silicon
echo 'eval "$(/opt/homebrew/bin/brew shellenv)"' >> ~/.zprofile
eval "$(/opt/homebrew/bin/brew shellenv)"
```

#### llama-server not found

**Issue**: Offline functionality requires llama.cpp.

**Solution**: Install llama.cpp (see Prerequisites section above)

#### Model download failed

**Issue**: Network issues or firewall restrictions.

**Solution**:

```bash
# Check network
curl -I https://huggingface.co

# Manually download model
scmd models pull qwen3-4b

# Use smaller model
scmd models pull qwen2.5-1.5b
```

#### Permission denied

**Issue**: Binary is not executable.

**Solution**:

```bash
chmod +x /path/to/scmd
```

## Optional: Additional LLM Backends

### Ollama

```bash
# Install Ollama
curl -fsSL https://ollama.com/install.sh | sh

# Pull a model
ollama pull qwen2.5-coder:1.5b

# Start Ollama server
ollama serve

# Use with scmd
scmd -b ollama /explain main.go
```

### OpenAI

```bash
# Set API key
export OPENAI_API_KEY=sk-...

# Use with scmd
scmd -b openai -m gpt-4o-mini /review code.py
```

### Together.ai (Free Tier Available)

```bash
# Get API key from https://together.ai
export TOGETHER_API_KEY=...

# Use with scmd
scmd -b together /explain main.go
```

### Groq (Free Tier Available)

```bash
# Get API key from https://groq.com
export GROQ_API_KEY=gsk_...

# Use with scmd
scmd -b groq -m llama-3.1-8b-instant /review code.py
```

## Next Steps

- [Quick Start Tutorial](quick-start.md) - Learn basic usage in 5 minutes
- [Your First Command](first-command.md) - Create a custom command
- [Shell Integration](shell-integration.md) - Set up `/command` shortcuts
- [Model Management](../user-guide/models.md) - Download and manage models

## Getting Help

- **Documentation**: [Full documentation](https://scmd.github.io/scmd/)
- **Issues**: [GitHub Issues](https://github.com/scmd/scmd/issues)
- **Discussions**: [GitHub Discussions](https://github.com/scmd/scmd/discussions)

For a detailed installation guide with platform-specific instructions and advanced options, see [INSTALL.md](https://github.com/scmd/scmd/blob/main/INSTALL.md) in the repository.
