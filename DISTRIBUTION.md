# Distribution Quick Reference

This document provides a quick reference for all scmd distribution channels and release procedures.

## Installation Methods

### Homebrew (macOS/Linux)

```bash
brew install scmd/tap/scmd
```

**Features**: Auto-updates, shell completions, dependency management

### npm (Cross-Platform)

```bash
npm install -g scmd-cli
```

**Features**: Works on all platforms, automatic PATH setup, easy updates

### Shell Script

```bash
curl -fsSL https://scmd.sh/install.sh | bash
```

**Features**: Platform detection, checksum verification, no dependencies

### Linux Packages

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

**Features**: System integration, package manager updates, completions included

### Binary Download

Download from [GitHub Releases](https://github.com/scmd/scmd/releases/latest)

**Features**: Direct download, all platforms, offline installation

## Creating a Release

### Quick Steps

```bash
# 1. Update CHANGELOG.md
git add CHANGELOG.md
git commit -m "docs: update changelog for v1.0.0"

# 2. Create and push tag
make tag VERSION=v1.0.0
git push origin v1.0.0

# 3. GitHub Actions automatically:
#    - Builds all platforms
#    - Creates GitHub release
#    - Updates Homebrew tap
#    - Publishes to npm
#    - Pushes Docker images
```

### Testing Locally

```bash
# Validate configuration
make check-goreleaser

# Test build without publishing
make release-snapshot

# Test specific binary
./dist/scmd-darwin-arm64 --version
```

## Distribution Channels

| Channel | URL | Auto-Updated | Manual Steps |
|---------|-----|--------------|--------------|
| GitHub Releases | [Releases](https://github.com/scmd/scmd/releases) | ✅ Yes | None |
| Homebrew Tap | [scmd/tap](https://github.com/scmd/homebrew-tap) | ✅ Yes | None |
| npm Registry | [scmd-cli](https://www.npmjs.com/package/scmd-cli) | ✅ Yes | None |
| Docker Hub | [scmd/scmd](https://hub.docker.com/r/scmd/scmd) | ✅ Yes | None |
| Install Script | [install.sh](https://scmd.sh/install.sh) | ✅ Yes | Host on CDN |

## Version Numbering

We follow [Semantic Versioning](https://semver.org/):

- **Major** (v2.0.0): Breaking changes
- **Minor** (v1.2.0): New features, backward compatible
- **Patch** (v1.2.1): Bug fixes, backward compatible

## Supported Platforms

### Operating Systems

- macOS (Intel and Apple Silicon)
- Linux (x64 and ARM64)
- Windows (x64)

### Architectures

- amd64 (x86_64)
- arm64 (Apple Silicon, ARM servers)

### Package Formats

- tar.gz (macOS, Linux)
- zip (Windows)
- deb (Debian, Ubuntu)
- rpm (Red Hat, Fedora, CentOS)
- apk (Alpine)

## Required Secrets

Configure in GitHub repository settings:

| Secret | Purpose |
|--------|---------|
| `HOMEBREW_TAP_GITHUB_TOKEN` | Update Homebrew tap |
| `NPM_TOKEN` | Publish to npm |
| `DOCKERHUB_USERNAME` | Docker Hub login |
| `DOCKERHUB_TOKEN` | Docker Hub auth |

## File Locations

```
Repository Structure:
├── .goreleaser.yml          # Release configuration
├── .github/workflows/
│   └── release.yml          # Release automation
├── scripts/
│   ├── install.sh           # Universal installer
│   ├── uninstall.sh         # Uninstaller
│   ├── postinstall.sh       # Package post-install
│   └── preremove.sh         # Package pre-remove
├── npm/                     # npm package
│   ├── package.json
│   ├── install.js
│   └── uninstall.js
├── Makefile                 # Build targets
├── INSTALL.md              # Installation guide
└── CHANGELOG.md            # Version history

User Installation:
├── /usr/local/bin/scmd              # Binary (Homebrew)
├── /usr/bin/scmd                    # Binary (Linux packages)
├── ~/.local/bin/scmd                # Binary (user install)
├── ~/.scmd/                         # Data directory
│   ├── config.yaml
│   ├── models/
│   └── cache/
└── Completions:
    ├── /usr/local/share/bash-completion/completions/scmd
    ├── /usr/local/share/zsh/site-functions/_scmd
    └── /usr/share/fish/vendor_completions.d/scmd.fish
```

## Makefile Targets

```bash
# Building
make build              # Build binary
make build-all          # Build for all platforms
make completions        # Generate shell completions

# Testing
make test               # Run tests
make lint               # Run linters

# Releasing
make tag VERSION=v1.0.0        # Create git tag
make release                   # Run GoReleaser (requires tag)
make release-snapshot          # Test release locally
make release-dry-run          # Test without publishing
make check-goreleaser         # Validate config

# Installation
make install            # Install to /usr/local/bin
make install-go         # Install to $GOPATH/bin

# Maintenance
make clean              # Clean build artifacts
make deps               # Update dependencies
```

## Common Tasks

### Create a Release

```bash
# 1. Update version info
vim CHANGELOG.md

# 2. Commit changes
git add CHANGELOG.md
git commit -m "docs: update changelog for v1.0.0"
git push origin main

# 3. Create tag
make tag VERSION=v1.0.0

# 4. Push tag (triggers release)
git push origin v1.0.0
```

### Test a Release Locally

```bash
# Create snapshot (no publishing)
make release-snapshot

# Check artifacts
ls -lh dist/

# Test binary
./dist/scmd-darwin-arm64 --version

# Test install script
./scripts/install.sh
```

### Update Homebrew Tap Manually

```bash
# Clone tap
git clone https://github.com/scmd/homebrew-tap
cd homebrew-tap

# Update formula (Formula/scmd.rb)
# - Update version
# - Update URLs
# - Update checksums

# Commit and push
git commit -am "Update to v1.0.0"
git push
```

### Publish to npm Manually

```bash
# Update version
cd npm
npm version 1.0.0 --no-git-tag-version

# Publish
npm publish --access public
```

## Troubleshooting

### Release Failed

1. Check GitHub Actions logs
2. Verify all secrets are set
3. Run `make check-goreleaser` locally
4. Test with `make release-snapshot`

### Homebrew Not Updated

1. Check `HOMEBREW_TAP_GITHUB_TOKEN` is set
2. Verify token has `repo` scope
3. Check homebrew-tap repository exists
4. Update manually (see above)

### npm Publish Failed

1. Check `NPM_TOKEN` is valid
2. Verify you have publish access
3. Publish manually (see above)

### Binary Not Working

1. Check platform/architecture match
2. Verify checksums
3. Check binary permissions
4. Test with `./scmd --version`

## Documentation

- **User Docs**: [Installation Guide](INSTALL.md)
- **Dev Docs**: [Release Process](docs/contributing/release-process.md)
- **Infrastructure**: [Distribution Guide](docs/contributing/distribution.md)
- **Full Docs**: https://scmd.github.io/scmd/

## Quick Links

- [GitHub Repository](https://github.com/scmd/scmd)
- [GitHub Releases](https://github.com/scmd/scmd/releases)
- [Homebrew Tap](https://github.com/scmd/homebrew-tap)
- [npm Package](https://www.npmjs.com/package/scmd-cli)
- [Docker Hub](https://hub.docker.com/r/scmd/scmd)
- [Documentation](https://scmd.github.io/scmd/)

## Support

- **Issues**: [GitHub Issues](https://github.com/scmd/scmd/issues)
- **Discussions**: [GitHub Discussions](https://github.com/scmd/scmd/discussions)
- **Security**: [Security Policy](SECURITY.md)
