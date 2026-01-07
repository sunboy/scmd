# Changelog

For the complete, detailed changelog, see [CHANGELOG.md](../CHANGELOG.md) in the repository root.

## Recent Releases

### [Unreleased]

#### Production & Distribution Readiness

**Added:**

- **Multi-Platform Distribution**: Automated releases via GoReleaser for macOS, Linux, and Windows
- **Homebrew Support**: Official Homebrew tap at `scmd/tap`
- **npm Distribution**: Cross-platform installation via `scmd-cli` npm package
- **Linux Packages**: Native deb, rpm, and apk packages
- **Install Scripts**: Universal wget/curl installation script with checksum verification
- **Docker Images**: Multi-arch Docker images on Docker Hub
- **Shell Completions**: Auto-generated completions for bash, zsh, fish, and PowerShell
- **Release Automation**: GitHub Actions workflow for automated releases
- **Comprehensive Documentation**:
  - Detailed installation guide (INSTALL.md)
  - Release process documentation
  - Distribution infrastructure guide

**Installation Methods:**

```bash
# Homebrew (macOS/Linux)
brew install scmd/tap/scmd

# npm (Cross-platform)
npm install -g scmd-cli

# Shell script (wget/curl)
curl -fsSL https://scmd.sh/install.sh | bash

# Linux packages
# Debian/Ubuntu
sudo dpkg -i scmd_VERSION_linux_amd64.deb

# Red Hat/Fedora/CentOS
sudo rpm -i scmd_VERSION_linux_amd64.rpm
```

**Enhanced:**

- Updated README with all installation methods
- Makefile with release and distribution targets
- Shell completion generation command

### [0.1.0] - 2025-01-06

Initial release with core functionality.

**Added:**

- Offline-first AI-powered slash commands
- llama.cpp integration with Qwen models
- Auto-download of models on first use
- Built-in `/explain` command
- Repository system for command distribution
- Shell integration for bash, zsh, and fish
- Multi-backend support (llama.cpp, Ollama, OpenAI, Together.ai, Groq)
- Command composition and chaining
- Configuration management
- Model management (list, pull, remove, info)
- Slash command system
- Command lockfiles for reproducibility
- Context gathering and caching

**Security:**

- Input validation and sanitization
- Path traversal prevention
- Resource limits for LLM inference
- Secure model downloads with checksum verification
- Comprehensive security documentation

**Documentation:**

- README with quick start guide
- Architecture documentation
- Security documentation
- Troubleshooting guide
- API documentation
- Contributing guidelines

## Version History

For detailed version information and complete release notes, see:

- [CHANGELOG.md](https://github.com/scmd/scmd/blob/main/CHANGELOG.md) - Complete changelog
- [GitHub Releases](https://github.com/scmd/scmd/releases) - Release notes and downloads

## Release Schedule

scmd follows semantic versioning and releases on an as-needed basis:

- **Major releases** (v2.0.0): Breaking changes, significant new features
- **Minor releases** (v1.1.0): New features, backward compatible
- **Patch releases** (v1.0.1): Bug fixes, backward compatible

## Upgrade Guide

### Upgrading to Latest Version

=== "Homebrew"

    ```bash
    brew upgrade scmd
    ```

=== "npm"

    ```bash
    npm update -g scmd-cli
    ```

=== "Shell Script"

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

## Staying Updated

- **Watch Releases**: [GitHub Watch](https://github.com/scmd/scmd/subscription) → Releases only
- **Star the Repo**: Get updates in your GitHub feed
- **Follow Changelog**: Check this page regularly for updates

## Deprecation Policy

When we deprecate features:

1. **Advance notice**: Minimum 1 major version before removal
2. **Migration guide**: Clear documentation on alternatives
3. **Warnings**: Deprecation warnings in the application
4. **Support period**: Bug fixes for deprecated features during transition

## Security Updates

Security updates are released as patch versions as soon as fixes are available. For security advisories, see:

- [Security Policy](../SECURITY.md)
- [GitHub Security Advisories](https://github.com/scmd/scmd/security/advisories)

## Contributing to the Changelog

When contributing, please update CHANGELOG.md following the [Keep a Changelog](https://keepachangelog.com/) format:

```markdown
## [Unreleased]

### Added
- New feature X

### Changed
- Modified Y for better performance

### Fixed
- Bug in Z

### Security
- Patched vulnerability in A
```

See [Contributing Guidelines](../contributing/pull-requests.md) for more details.
