# scmd Architecture

## Overview

scmd uses a **repository-first architecture** that separates the core tool from the commands it runs. Think of it like a package manager (npm, pip, Homebrew) for AI-powered commands.

## Design Philosophy

### Core Principles

1. **Small Core, Big Ecosystem**
   - Minimal built-in functionality (~14MB binary)
   - Commands distributed through repositories
   - Community-driven command library

2. **Offline-First**
   - Works without network after initial setup
   - Local model inference with llama.cpp
   - Commands cached locally after installation

3. **Decentralized & Flexible**
   - Anyone can create command repositories
   - No central authority required
   - Teams can host private repositories

4. **Reproducible & Shareable**
   - Lockfiles ensure consistent command versions
   - Commands have explicit dependencies
   - Team members use identical setups

## Architecture Layers

```
┌─────────────────────────────────────────────────┐
│              User Interface                      │
│  - CLI flags and arguments                       │
│  - Slash command syntax (/explain, /review)      │
│  - Shell integration (Bash, Zsh, Fish)          │
└─────────────────────────────────────────────────┘
                      ▼
┌─────────────────────────────────────────────────┐
│            Command Layer                         │
│  - Built-in: /explain                            │
│  - Repository-based: /review, /commit, etc.      │
│  - Custom: user-defined commands                 │
└─────────────────────────────────────────────────┘
                      ▼
┌─────────────────────────────────────────────────┐
│         Repository System                        │
│  - Command discovery and search                  │
│  - Manifest parsing and validation               │
│  - Dependency resolution                         │
│  - Version management                            │
│  - Local caching                                 │
└─────────────────────────────────────────────────┘
                      ▼
┌─────────────────────────────────────────────────┐
│          Backend Layer                           │
│  - llama.cpp (default, offline)                  │
│  - Ollama (local, optional)                      │
│  - OpenAI, Groq, Together.ai (cloud)             │
└─────────────────────────────────────────────────┘
                      ▼
┌─────────────────────────────────────────────────┐
│              LLM Models                          │
│  - Local: Qwen models (GGUF format)              │
│  - Cloud: GPT-4, Claude, Llama, etc.             │
└─────────────────────────────────────────────────┘
```

## Why Repository-First?

### The Problem

Traditional approaches bundle all commands into the binary:
- ❌ Large binary size
- ❌ Slow updates (need to rebuild entire tool)
- ❌ Limited to developer-provided commands
- ❌ No community contributions without forking

### The Solution

Repository-first architecture:
- ✅ Small binary (~14MB vs 100MB+)
- ✅ Commands update independently
- ✅ Community can create and share commands
- ✅ Teams can create private repositories
- ✅ Reproducible with lockfiles

### Comparison

| Aspect | Monolithic | Repository-First (scmd) |
|--------|-----------|------------------------|
| **Binary size** | 100-500MB | 14MB |
| **Built-in commands** | 50-100 | 1 (explain) |
| **Update speed** | Rebuild entire tool | Update individual commands |
| **Community contributions** | Requires fork + PR | Publish to any repository |
| **Private commands** | Not possible | Host your own repository |
| **Version control** | Single version | Per-command versions |
| **Offline capability** | All bundled | Cached after install |

## Component Details

### 1. Built-in Commands

Only one command is built into the scmd binary:

**`/explain`** - Explain code or concepts
- Rationale: Core functionality needed for first-run experience
- Zero dependencies
- Demonstrates core capabilities
- Works offline immediately after model download

### 2. Repository System

Commands are distributed through Git repositories or HTTP endpoints.

**Repository Structure:**
```
my-commands/
├── scmd-repo.yaml          # Manifest
└── commands/
    ├── review.yaml
    ├── commit.yaml
    └── fix.yaml
```

**Manifest (scmd-repo.yaml):**
```yaml
name: my-commands
version: "1.0.0"
description: Custom AI commands
author: Your Team

commands:
  - name: review
    description: Review code for issues
    file: commands/review.yaml
    version: "1.2.0"

  - name: commit
    description: Generate commit messages
    file: commands/commit.yaml
    version: "1.0.0"
```

**Command Definition (commands/review.yaml):**
```yaml
name: review
version: "1.2.0"
description: Review code for issues and improvements
category: code-quality
author: scmd team

dependencies:
  - command: official/explain
    version: ">=1.0.0"

args:
  - name: severity
    description: Minimum severity (low, medium, high)
    default: medium

prompt:
  system: |
    You are an expert code reviewer.
    Focus on bugs, security issues, and best practices.

  template: |
    Review this code for issues:
    {{.stdin}}

    Minimum severity: {{.severity}}

    Provide specific suggestions with line numbers.

model:
  temperature: 0.3
  max_tokens: 1024
```

### 3. Command Lifecycle

```
┌─────────────┐
│   Discover  │  scmd repo search "git"
└──────┬──────┘
       ▼
┌─────────────┐
│    Show     │  scmd repo show official/commit
└──────┬──────┘
       ▼
┌─────────────┐
│   Install   │  scmd repo install official/commit
└──────┬──────┘  - Downloads YAML
       │         - Resolves dependencies
       │         - Validates command
       │         - Caches locally (~/.scmd/commands/)
       ▼
┌─────────────┐
│     Use     │  scmd /commit
└──────┬──────┘
       ▼
┌─────────────┐
│   Update    │  scmd update --check
└──────┬──────┘  scmd update official/commit
       │
       ▼
┌─────────────┐
│   Remove    │  scmd repo uninstall official/commit
└─────────────┘
```

### 4. Backend Abstraction

scmd supports multiple LLM backends through a unified interface:

```go
type Backend interface {
    Name() string
    IsAvailable(ctx context.Context) (bool, error)
    Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)
    Stream(ctx context.Context, req *CompletionRequest) (<-chan StreamChunk, error)
    ModelInfo() ModelInfo
}
```

**Backend Priority:**
1. **llama.cpp** - Default, offline, auto-starts server
2. **Ollama** - If running on localhost
3. **Groq** - If GROQ_API_KEY set (fast, free tier)
4. **OpenAI** - If OPENAI_API_KEY set
5. **Together.ai** - If TOGETHER_API_KEY set

This allows users to:
- Start with offline llama.cpp
- Switch to cloud providers for better quality
- Fall back to alternative backends if one fails

### 5. Local Storage

```
~/.scmd/
├── config.yaml              # User configuration
├── scmd.lock               # Lockfile for reproducibility
├── models/                 # Downloaded LLM models
│   ├── qwen3-4b.gguf
│   └── qwen2.5-3b.gguf
├── commands/               # Installed commands
│   ├── official/
│   │   ├── review.yaml
│   │   └── commit.yaml
│   └── custom/
│       └── my-command.yaml
├── repos/                  # Repository manifests
│   ├── official.yaml
│   └── custom.yaml
├── cache/                  # Temporary cache
└── logs/                   # Server logs
    └── llama-server.log
```

## Advanced Features

### 1. Command Dependencies

Commands can depend on other commands:

```yaml
dependencies:
  - command: official/explain
    version: ">=1.0.0"
  - command: official/summarize
    version: "^1.2.0"
    optional: true
```

When installing a command, scmd:
1. Resolves all dependencies
2. Installs missing commands
3. Validates version constraints
4. Builds dependency graph
5. Detects circular dependencies

### 2. Command Composition

Chain commands together:

```yaml
compose:
  pipeline:
    - command: explain
      output: explanation
    - command: summarize
      input: "{{.explanation}}"
      args:
        length: short
```

### 3. Lockfiles

Ensure reproducible installations:

```yaml
# scmd.lock
version: "1.0.0"
commands:
  - name: official/review
    version: "1.2.3"
    checksum: sha256:abc123...
    dependencies:
      - name: official/explain
        version: "1.0.0"
        checksum: sha256:def456...
```

Team workflow:
```bash
# Developer 1: Install and lock
scmd repo install official/review
scmd lock generate

# Commit scmd.lock to git
git add scmd.lock
git commit -m "Lock scmd commands"

# Developer 2: Install from lockfile
scmd lock install  # Installs exact versions
```

### 4. Private Repositories

Host your own command repository:

```bash
# Self-hosted (GitHub, GitLab, etc.)
scmd repo add mycompany https://github.com/mycompany/scmd-commands/raw/main

# HTTP server
scmd repo add internal http://internal.company.com/scmd-repo.yaml

# Local filesystem (for development)
scmd repo add local file:///path/to/repo
```

## Design Decisions

### Why Only One Built-in Command?

**Rationale:**
- Keeps binary small and focused
- Forces dogfooding of repository system
- Prevents feature creep
- Encourages community contributions

**Why `/explain`?**
- Core use case: understanding code
- Demonstrates core capabilities
- Useful immediately after install
- No external dependencies
- Good first-time user experience

### Why YAML for Commands?

**Considered alternatives:**
- JSON: Too verbose, no comments
- TOML: Less familiar, limited nesting
- Custom DSL: Learning curve, tooling overhead

**YAML chosen for:**
- Human-readable and editable
- Good for templates (multiline strings)
- Familiar to developers
- Rich ecosystem of parsers
- Comments and anchors support

### Why Not Plugin System?

**Plugins (compiled binaries) rejected because:**
- Security concerns (arbitrary code execution)
- Platform-specific binaries needed
- Larger download sizes
- Complex build process
- Harder to audit

**YAML commands preferred:**
- Safe (declarative, no code execution)
- Cross-platform
- Small (few KB vs MB)
- Easy to audit and review
- Shareable as text

## Migration from Earlier Versions

If you used scmd before the repository-first architecture:

### Before (Monolithic)
```bash
scmd /review code.py    # Built-in command
scmd /commit            # Built-in command
scmd /fix error.log     # Built-in command
```

### After (Repository-First)
```bash
# Only /explain is built-in
scmd /explain code.py   # ✓ Works immediately

# Install other commands
scmd repo add official https://github.com/scmd/commands/raw/main
scmd repo install official/review
scmd repo install official/commit
scmd repo install official/fix

# Now use them
scmd /review code.py    # ✓ Works after install
scmd /commit            # ✓ Works after install
scmd /fix error.log     # ✓ Works after install
```

### Benefits of Migration
- ✅ Smaller binary size (500MB → 14MB)
- ✅ Faster updates (update individual commands)
- ✅ Better version control (per-command versions)
- ✅ Community can contribute commands
- ✅ Team-specific private commands

## Future Directions

### Planned Enhancements

1. **Command Marketplace**
   - Central registry of commands
   - Ratings and reviews
   - Usage statistics
   - Featured commands

2. **Command Playground**
   - Test commands before installing
   - Interactive prompt editor
   - Preview results

3. **Enhanced Discovery**
   - Semantic search (embeddings)
   - Command recommendations
   - Trending commands
   - Category browsing

4. **Enterprise Features**
   - Private registry support
   - Access control
   - Audit logs
   - Compliance checks

## Contributing

### Creating Commands

1. Fork the [official commands repository](https://github.com/scmd/commands)
2. Add your command YAML file
3. Update the manifest
4. Test locally: `scmd repo add local file:///path/to/repo`
5. Submit a PR

### Creating Repositories

1. Create repository structure (scmd-repo.yaml + commands/)
2. Host on GitHub, GitLab, or any HTTP server
3. Share the repository URL
4. Submit to [scmd registry](https://github.com/scmd/registry) for discovery

## Conclusion

The repository-first architecture enables:
- **Small core tool** with focused functionality
- **Vibrant ecosystem** of community commands
- **Flexible deployment** (public, private, local)
- **Reproducible environments** with lockfiles
- **Independent updates** without tool rebuilds

This design scales from individual developers to large teams, from hobbyist projects to enterprise deployments.

---

**Last Updated**: January 2026
**scmd Version**: 1.0.0

For questions or suggestions, please [open an issue](https://github.com/scmd/scmd/issues).
