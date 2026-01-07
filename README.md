# scmd

**AI-powered slash commands for any terminal. Works offline by default.**

scmd brings the power of LLM-based slash commands to your command line. Works offline by default with llama.cpp and Qwen models, or connect to Ollama, OpenAI, and more. Type `/gc` to generate commit messages, `/explain` to understand code, or install new commands from community repositories.

```bash
# Works immediately - no API keys or setup required:
./scmd /explain main.go        # Explain code
./scmd /gc                      # Generate commit message from staged changes
./scmd /review                  # Review code for issues
git diff | ./scmd /sum          # Summarize changes

# Or use the scmd command directly:
cat main.go | scmd explain
git diff | scmd review
```

## Features

- **Offline-First** - llama.cpp with local Qwen models, no API keys required
- **Auto-Download Models** - Qwen3-4B downloads automatically on first use (~2.6GB)
- **Real Slash Commands** - Type `/command` directly (with or without shell integration)
- **Repository-First Architecture** - Commands install from repos like npm packages
- **Multiple LLM Backends** - llama.cpp (default), Ollama, OpenAI, Together.ai, Groq
- **Command Composition** - Chain commands in pipelines, run in parallel, or use fallbacks
- **Shell Integration** - Bash, Zsh, and Fish support with tab completion
- **Local Caching** - Commands and manifests cached locally
- **Lockfiles** - Reproducible installations for teams

## Architecture

scmd uses a **repository-first architecture** similar to package managers like npm or Homebrew:

- **Small Core**: Only the `explain` command is built-in, keeping the binary lean (~14MB)
- **Repository-Based**: Most commands install from repositories (official or community)
- **Network Optional**: Core functionality works offline; network needed only for installing new commands
- **Decentralized**: Anyone can create and host command repositories
- **Version Management**: Commands have versions, dependencies, and lockfiles for reproducibility

**Example workflow:**
```bash
# Built-in command works immediately
scmd /explain main.go

# Install additional commands from repositories
scmd repo add official https://github.com/scmd/commands/raw/main
scmd repo install official/review
scmd /review code.py  # Now available
```

This design allows:
- ✅ Small binary size and fast installation
- ✅ Community-driven command ecosystem
- ✅ Easy command discovery and sharing
- ✅ Team-specific command repositories
- ✅ Reproducible environments with lockfiles

## Installation

### Quick Install

Choose the method that works best for you:

#### Homebrew (macOS/Linux)

```bash
brew install scmd/tap/scmd
```

#### npm (Cross-Platform)

```bash
npm install -g scmd-cli
```

#### Shell Script (wget/curl)

```bash
# Using curl
curl -fsSL https://scmd.sh/install.sh | bash

# Using wget
wget -qO- https://scmd.sh/install.sh | bash
```

#### Linux Packages

**Debian/Ubuntu:**
```bash
wget https://github.com/scmd/scmd/releases/latest/download/scmd_VERSION_linux_amd64.deb
sudo dpkg -i scmd_VERSION_linux_amd64.deb
```

**Fedora/RHEL:**
```bash
wget https://github.com/scmd/scmd/releases/latest/download/scmd_VERSION_linux_amd64.rpm
sudo rpm -i scmd_VERSION_linux_amd64.rpm
```

### Post-Installation

1. **Verify installation:**
   ```bash
   scmd --version
   ```

2. **Install llama.cpp for offline usage:**
   ```bash
   # macOS
   brew install llama.cpp

   # Linux - build from source
   # See: https://github.com/ggerganov/llama.cpp
   ```

3. **Try it out:**
   ```bash
   scmd /explain "what is a goroutine"
   ```

For detailed installation instructions, platform-specific guides, and troubleshooting, see [INSTALL.md](INSTALL.md).

### Build from Source

```bash
# Clone and build
git clone https://github.com/scmd/scmd
cd scmd
make build

# Install to /usr/local/bin
sudo make install

# Or build with Go directly
go install github.com/scmd/scmd/cmd/scmd@latest
```

## Model Management

scmd uses llama.cpp with efficient Qwen models for offline inference. Models are downloaded automatically on first use.

### Available Models

```bash
# List available models
scmd models list

# Output:
# NAME          SIZE      STATUS          DESCRIPTION
# qwen2.5-3b    1.9 GB    not downloaded  Qwen2.5 3B - Good balance
# qwen2.5-1.5b  940 MB    not downloaded  Qwen2.5 1.5B - Fast and lightweight
# qwen2.5-0.5b  379 MB    not downloaded  Qwen2.5 0.5B - Smallest, fastest
# qwen2.5-7b    4.4 GB    not downloaded  Qwen2.5 7B - Best quality
# qwen3-4b      2.5 GB    ✓ ready         Qwen3 4B - Default (tool calling)
```

### Managing Models

```bash
# Download a specific model
scmd models pull qwen2.5-3b

# Show model info
scmd models info qwen3-4b

# Set default model
scmd models default qwen2.5-3b

# Remove a downloaded model
scmd models remove qwen2.5-7b
```

Models are stored in `~/.scmd/models/` and use GPU acceleration when available (Metal on macOS, CUDA on Linux).

## Quick Start

```bash
# Use the built-in explain command (model downloads on first run)
cat myfile.go | scmd explain
scmd /explain "what is a goroutine"

# Install additional commands from the official repository
scmd repo add official https://github.com/scmd/commands/raw/main
scmd repo install official/review
scmd repo install official/commit

# Now use the installed commands
git diff | scmd review
git diff --staged | scmd /gc  # Generate commit message

# Use with inline prompt
echo "SELECT * FROM users" | scmd -p "optimize this SQL query"

# Save output to file
git diff | scmd review -o review.md

# Use specific backend/model
scmd -b openai -m gpt-4 explain main.go

# Discover more commands
scmd repo search git
scmd repo search docker
```

## Slash Commands

The core feature of scmd is slash commands that work directly in your terminal.

### Direct Usage (No Setup Required)

You can use slash commands immediately without any shell integration:

```bash
# Direct invocation
./scmd /explain main.go
./scmd /review code.py
./scmd /gc
./scmd /e "what are channels?"

# With pipes
cat error.log | ./scmd /fix
git diff | ./scmd /gc
curl api.com/data | ./scmd /sum
```

### Setup Shell Integration (Optional)

For even better ergonomics, set up shell integration to use `/command` without the `./scmd` prefix:

```bash
# For Bash/Zsh - add to your ~/.bashrc or ~/.zshrc:
eval "$(scmd slash init bash)"

# For Fish - add to ~/.config/fish/config.fish:
scmd slash init fish | source
```

After setup, use slash commands directly:

```bash
/explain main.go           # Explain code
/gc                        # Generate commit message
/review                    # Review code
/sum article.md            # Summarize
/fix                       # Explain errors

# Pipe input to commands
cat error.log | /fix
git diff | /gc
curl api.com/data | /sum
```

### Built-in Commands

Only one command is built-in with scmd - others install from repositories:

| Command | Aliases | Description | Source |
|---------|---------|-------------|--------|
| `/explain` | `/e`, `/exp` | Explain code or concepts | Built-in ✓ |

### Popular Community Commands

Install additional commands from the official repository:

```bash
# Install popular commands
scmd repo add official https://raw.githubusercontent.com/scmd/commands/main
scmd repo install official/review        # Code review
scmd repo install official/commit        # Git commit messages
scmd repo install official/summarize     # Summarize text
scmd repo install official/fix           # Explain and fix errors
```

This repository-first architecture keeps the scmd binary small while allowing the community to build and share commands.

### Managing Slash Commands

```bash
# List all slash commands
scmd slash list

# Add a new slash command
scmd slash add doc generate-docs --alias=d,docs

# Add an alias to existing command
scmd slash alias commit c

# Remove a slash command
scmd slash remove doc

# Interactive mode (REPL)
scmd slash interactive
```

## Repository System

scmd's repository system lets you distribute and install AI commands. Think Homebrew taps, but for AI prompts.

### Installing Commands

```bash
# Add a repository
scmd repo add community https://raw.githubusercontent.com/scmd-community/commands/main

# Search for commands
scmd repo search git

# Show command details
scmd repo show community/git-commit

# Install a command
scmd repo install community/git-commit

# Use the installed command
git diff | scmd git-commit
```

### Managing Repositories

```bash
# List configured repos
scmd repo list

# Update repo manifests
scmd repo update

# Remove a repo
scmd repo remove community
```

### Central Registry

Discover commands from the central scmd registry:

```bash
# Search the registry
scmd registry search docker

# Browse by category
scmd registry categories

# Show trending commands
scmd registry featured
```

## Command Specification

Commands are defined in YAML files with a powerful specification:

```yaml
name: git-commit
version: "1.0.0"
description: Generate commit messages from diffs
category: git
author: scmd team

args:
  - name: style
    description: Commit style (conventional, simple)
    default: conventional

prompt:
  system: |
    You are a git commit message expert.
    Use conventional commits format.
  template: |
    Generate a commit message for:
    {{.stdin}}

    Style: {{.style}}

model:
  temperature: 0.3
  max_tokens: 256
```

### Advanced Features

**Dependencies** - Commands can depend on other commands:
```yaml
dependencies:
  - command: official/explain
    version: ">=1.0.0"
  - command: official/summarize
    optional: true
```

**Composition** - Chain commands together:
```yaml
compose:
  pipeline:
    - command: explain
    - command: summarize
      args:
        length: short
```

**Hooks** - Run shell commands before/after:
```yaml
hooks:
  pre:
    - shell: "git status --porcelain"
      if: "{{.git}}"
  post:
    - shell: "echo 'Done!'"
```

**Context** - Auto-include files and environment:
```yaml
context:
  files:
    - "*.go"
    - "go.mod"
  git: true
  env:
    - GOPATH
```

## Lockfiles

Share exact command versions with your team:

```bash
# Generate lockfile from installed commands
scmd lock generate

# Install from lockfile
scmd lock install

# Check for updates
scmd update --check

# Update all commands
scmd update --all
```

## LLM Backends

scmd supports multiple LLM backends. llama.cpp is used by default for offline inference.

| Backend | Local | Free | Default | Setup |
|---------|-------|------|---------|-------|
| **llama.cpp** | ✓ | ✓ | ✓ | `brew install llama.cpp` |
| **Ollama** | ✓ | ✓ | | `ollama serve` |
| **OpenAI** | | | | `export OPENAI_API_KEY=...` |
| **Together.ai** | | Free tier | | `export TOGETHER_API_KEY=...` |
| **Groq** | | Free tier | | `export GROQ_API_KEY=...` |

### Backend Priority

Backends are tried in this order:
1. **llama.cpp** - Local, offline, no setup required (default)
2. **Ollama** - Local, if running
3. **OpenAI** - If API key set
4. **Together.ai** - If API key set
5. **Groq** - If API key set

### Using Backends

```bash
# Use specific backend
scmd -b ollama explain main.go

# Use specific model
scmd -b openai -m gpt-4 review code.py

# List available backends
scmd backends

# Example output:
#   ✓ llamacpp     qwen3-4b
#   ✗ ollama       qwen2.5-coder-1.5b
#   ✗ openai       (not configured)
```

## Creating a Repository

Create your own command repository:

```
my-commands/
├── scmd-repo.yaml          # Repository manifest
└── commands/
    ├── my-command.yaml
    └── another-command.yaml
```

**scmd-repo.yaml:**
```yaml
name: my-commands
version: "1.0.0"
description: My custom scmd commands
author: Your Name

commands:
  - name: my-command
    description: Does something useful
    file: commands/my-command.yaml
```

Host on GitHub, GitLab, or any HTTP server, then:
```bash
scmd repo add myrepo https://raw.githubusercontent.com/you/my-commands/main
```

## Configuration

Configuration is stored in `~/.scmd/config.yaml`:

```yaml
default_backend: llamacpp
default_model: qwen3-4b

backends:
  llamacpp:
    model: qwen3-4b
  ollama:
    host: http://localhost:11434
  openai:
    model: gpt-4o-mini

ui:
  color: true
  spinner: true
```

## CLI Reference

```
scmd [command] [flags]

Commands:
  explain     Explain code or concepts
  review      Review code for issues
  config      View/modify configuration
  backends    List available backends

  models      Manage local LLM models
    list      List available models
    pull      Download a model
    remove    Remove a model
    info      Show model information
    default   Set default model

  slash       Slash command management
    run       Run a slash command
    list      List slash commands
    add       Add a slash command
    remove    Remove a slash command
    alias     Add an alias
    init      Generate shell integration
    interactive  Start REPL mode

  repo        Manage repositories
    add       Add a repository
    remove    Remove a repository
    list      List repositories
    update    Update manifests
    search    Search for commands
    show      Show command details
    install   Install a command

  registry    Central registry
    search    Search registry
    featured  Trending commands
    categories List categories

  update      Check for updates
  lock        Manage lockfiles
  cache       Manage local cache

Flags:
  -b, --backend   Backend to use
  -m, --model     Model to use
  -p, --prompt    Inline prompt
  -o, --output    Output file
  -f, --format    Output format (text, json, markdown)
  -q, --quiet     Suppress progress
  -v, --verbose   Verbose output
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `OLLAMA_HOST` | Ollama server URL (default: http://localhost:11434) |
| `OPENAI_API_KEY` | OpenAI API key |
| `TOGETHER_API_KEY` | Together.ai API key |
| `GROQ_API_KEY` | Groq API key |
| `SCMD_CONFIG` | Config file path (default: ~/.scmd/config.yaml) |
| `SCMD_DATA_DIR` | Data directory (default: ~/.scmd) |
| `SCMD_DEBUG` | Enable debug logging (set to 1) |

## Performance

llama.cpp with Qwen models provides fast, efficient inference:

- **Qwen2.5-0.5B**: ~10 tokens/sec on CPU, ~50 tokens/sec on GPU
- **Qwen3-4B**: ~5 tokens/sec on CPU, ~20 tokens/sec on GPU (M1 Mac)
- **Qwen2.5-7B**: ~2 tokens/sec on CPU, ~10 tokens/sec on GPU

Models use 4-bit quantization (Q4_K_M) for optimal size/quality tradeoff.

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Creating Commands

1. Fork the [scmd-community/commands](https://github.com/scmd-community/commands) repo
2. Add your command YAML file
3. Update the manifest
4. Submit a PR

## License

MIT License - see [LICENSE](LICENSE) for details.

---

Built with Go. Inspired by the Unix philosophy and modern AI tooling.
