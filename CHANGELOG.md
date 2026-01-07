# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- GoReleaser configuration for automated multi-platform releases
- Homebrew tap support for easy macOS/Linux installation
- npm package wrapper for cross-platform distribution
- Shell install script for wget/curl installation
- Native Linux packages (deb, rpm, apk) via nfpm
- GitHub Actions workflow for automated releases
- Shell completion support (bash, zsh, fish)
- Comprehensive installation documentation (INSTALL.md)
- Makefile targets for release management
- Docker image support (multi-arch)
- Checksum verification for downloads
- Post-install and pre-remove scripts for package managers

### Changed
- Updated Makefile with release and distribution targets
- Enhanced documentation with multiple installation methods

## [0.1.0] - 2025-01-06

### Added
- Initial release
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

### Security
- Input validation and sanitization
- Path traversal prevention
- Resource limits for LLM inference
- Secure model downloads with checksum verification
- Comprehensive security documentation

### Documentation
- README with quick start guide
- Architecture documentation
- Security documentation
- Troubleshooting guide
- API documentation
- Contributing guidelines

## Release Process

To create a new release:

1. Update version in relevant files
2. Update this CHANGELOG with release notes
3. Create and push a git tag:
   ```bash
   make tag VERSION=v1.0.0
   git push origin v1.0.0
   ```
4. GitHub Actions will automatically:
   - Build binaries for all platforms
   - Create GitHub release with notes
   - Publish to Homebrew tap
   - Publish to npm registry
   - Build and push Docker images

## Version History

[Unreleased]: https://github.com/scmd/scmd/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/scmd/scmd/releases/tag/v0.1.0
