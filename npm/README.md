# scmd-cli

AI-powered slash commands for any terminal. Works offline by default.

This is the npm distribution of scmd. For the main project, see [scmd on GitHub](https://github.com/scmd/scmd).

## Installation

```bash
npm install -g scmd-cli
```

## Quick Start

```bash
# Use the built-in explain command
scmd /explain "what is a goroutine"

# Generate commit messages from git diff
git diff --staged | scmd /gc

# Review code
cat myfile.go | scmd review
```

## Requirements

For offline functionality, install llama.cpp:

```bash
# macOS
brew install llama.cpp

# Linux
# Build from source: https://github.com/ggerganov/llama.cpp
```

## Documentation

For full documentation, visit: https://github.com/scmd/scmd

## Features

- **Offline-First** - Works with local models via llama.cpp, no API keys required
- **Auto-Download Models** - Qwen models download automatically on first use
- **Real Slash Commands** - Type `/command` directly in your terminal
- **Multiple LLM Backends** - Supports llama.cpp, Ollama, OpenAI, and more
- **Shell Integration** - Works with Bash, Zsh, and Fish

## Other Installation Methods

### Homebrew (macOS/Linux)

```bash
brew install scmd/tap/scmd
```

### Shell Script

```bash
curl -fsSL https://scmd.sh/install.sh | bash
```

### Binary Download

Download from [GitHub Releases](https://github.com/scmd/scmd/releases)

## License

MIT
