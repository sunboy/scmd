# Release Process

This guide explains how to create and publish new releases of scmd using our automated distribution pipeline.

## Overview

scmd uses [GoReleaser](https://goreleaser.com) for automated multi-platform releases. When you push a git tag, GitHub Actions automatically:

1. Runs tests
2. Builds binaries for all platforms
3. Creates Linux packages (deb, rpm, apk)
4. Updates Homebrew tap
5. Publishes to npm
6. Creates Docker images
7. Generates release notes

## Prerequisites

### Required Tools

```bash
# Install GoReleaser
make install-goreleaser

# Or manually
go install github.com/goreleaser/goreleaser/v2@latest
```

### Required GitHub Secrets

Set these in your repository settings (`Settings` → `Secrets and variables` → `Actions`):

| Secret | Purpose | How to Get |
|--------|---------|------------|
| `GITHUB_TOKEN` | Auto-generated | Automatically available in GitHub Actions |
| `HOMEBREW_TAP_GITHUB_TOKEN` | Update Homebrew tap | Personal Access Token with `repo` scope |
| `NPM_TOKEN` | Publish to npm | [Create at npmjs.com](https://www.npmjs.com/settings/~/tokens) |
| `DOCKERHUB_USERNAME` | Docker Hub login | Your Docker Hub username |
| `DOCKERHUB_TOKEN` | Docker Hub authentication | [Create at hub.docker.com](https://hub.docker.com/settings/security) |

## Release Workflow

### 1. Prepare the Release

#### Update Version Information

No manual version updates are needed. GoReleaser automatically uses the git tag as the version.

#### Update CHANGELOG.md

Document all changes since the last release:

```markdown
## [1.2.0] - 2025-01-15

### Added
- New feature X
- Support for Y

### Changed
- Updated Z for better performance

### Fixed
- Bug in component A
- Issue with B

### Security
- Patched vulnerability CVE-2025-XXXX
```

#### Commit Changes

```bash
git add CHANGELOG.md
git commit -m "docs: update changelog for v1.2.0"
git push origin main
```

### 2. Create a Git Tag

```bash
# Create and push tag using Makefile
make tag VERSION=v1.2.0

# Push the tag to trigger release
git push origin v1.2.0
```

Or manually:

```bash
# Create annotated tag
git tag -a v1.2.0 -m "Release v1.2.0"

# Push tag
git push origin v1.2.0
```

### 3. Monitor the Release

1. Go to the [Actions tab](https://github.com/scmd/scmd/actions)
2. Find the "Release" workflow
3. Monitor progress through each step:
   - ✅ Run Tests
   - ✅ Build with GoReleaser
   - ✅ Publish to npm
   - ✅ Create install script
   - ✅ Notify

### 4. Verify the Release

Once the workflow completes, verify:

#### GitHub Release

1. Go to [Releases](https://github.com/scmd/scmd/releases)
2. Check that the new release appears with:
   - Release notes
   - All platform binaries
   - Checksums
   - Linux packages

#### Homebrew

```bash
# Test Homebrew installation
brew uninstall scmd
brew tap scmd/tap
brew install scmd
scmd --version  # Should show v1.2.0
```

#### npm

```bash
# Test npm installation
npm uninstall -g scmd-cli
npm install -g scmd-cli
scmd --version  # Should show 1.2.0 (without 'v')
```

#### Docker

```bash
# Test Docker image
docker pull scmd/scmd:1.2.0
docker run scmd/scmd:1.2.0 --version
```

#### Install Script

```bash
# Test install script
curl -fsSL https://raw.githubusercontent.com/scmd/scmd/v1.2.0/scripts/install.sh | bash
scmd --version
```

## Testing Releases Locally

### Snapshot Release (No Publishing)

Test the full release process without publishing:

```bash
# Create a snapshot release
make release-snapshot

# Check dist/ directory for artifacts
ls -lh dist/
```

This creates:

- Binaries for all platforms in `dist/`
- Archives and checksums
- Linux packages
- Test the installation locally

### Dry Run (Test Without Publishing)

Test GoReleaser configuration:

```bash
# Validate configuration
make check-goreleaser

# Dry run (skip publishing)
make release-dry-run
```

### Local Testing Workflow

```bash
# 1. Validate GoReleaser config
make check-goreleaser

# 2. Run tests
make test

# 3. Create snapshot
make release-snapshot

# 4. Test binary locally
./dist/scmd-darwin-arm64 --version

# 5. Test install script locally
./scripts/install.sh
```

## Distribution Channels

### Homebrew Tap

**Repository**: [scmd/homebrew-tap](https://github.com/scmd/homebrew-tap)

GoReleaser automatically updates the tap with:

- New formula version
- Updated checksums
- Binary URLs

Users install with:

```bash
brew tap scmd/tap
brew install scmd
```

**To submit to homebrew-core** (future):

1. Fork [Homebrew/homebrew-core](https://github.com/Homebrew/homebrew-core)
2. Copy formula from tap to `Formula/scmd.rb`
3. Update URLs to point to main release
4. Submit PR

### npm Registry

**Package**: [scmd-cli](https://www.npmjs.com/package/scmd-cli)

The npm package is a wrapper that:

1. Downloads the correct binary for the user's platform
2. Installs it to npm's bin directory
3. Automatically adds to PATH

Publishing is automated via GitHub Actions.

### Linux Package Repositories

Packages are created for:

- **Debian/Ubuntu** (`.deb`)
- **Red Hat/Fedora/CentOS** (`.rpm`)
- **Alpine Linux** (`.apk`)

Packages include:

- Binary in `/usr/bin/scmd`
- Shell completions
- Man pages (future)
- Post-install/remove scripts

### Docker Hub

**Repository**: [scmd/scmd](https://hub.docker.com/r/scmd/scmd)

Multi-arch images are built for:

- `scmd/scmd:latest`
- `scmd/scmd:1.2.0`
- `scmd/scmd:1.2.0-amd64`
- `scmd/scmd:1.2.0-arm64`

## Version Strategy

### Semantic Versioning

We follow [Semantic Versioning 2.0.0](https://semver.org):

- **Major** (v2.0.0): Breaking changes
- **Minor** (v1.2.0): New features, backward compatible
- **Patch** (v1.2.1): Bug fixes, backward compatible

### Pre-release Versions

For testing:

```bash
# Alpha
git tag -a v1.3.0-alpha.1 -m "Alpha release"

# Beta
git tag -a v1.3.0-beta.1 -m "Beta release"

# Release candidate
git tag -a v1.3.0-rc.1 -m "Release candidate"
```

Pre-releases are marked as "pre-release" on GitHub and not promoted as "latest".

### Version Tags

- Use `v` prefix: `v1.2.0` (not `1.2.0`)
- Create annotated tags (not lightweight)
- Include meaningful tag message

## Release Checklist

Before creating a release:

- [ ] All tests pass (`make test`)
- [ ] Documentation is updated
- [ ] CHANGELOG.md is updated
- [ ] Breaking changes are documented
- [ ] Security advisories are published (if applicable)
- [ ] Migration guide exists (for breaking changes)

After creating a release:

- [ ] GitHub release created successfully
- [ ] All binaries are uploaded
- [ ] Checksums are correct
- [ ] Homebrew tap updated
- [ ] npm package published
- [ ] Docker images pushed
- [ ] Installation methods verified
- [ ] Release notes reviewed
- [ ] Announcements posted (if major release)

## Hotfix Releases

For urgent fixes:

```bash
# 1. Create hotfix branch from tag
git checkout -b hotfix/v1.2.1 v1.2.0

# 2. Make fix
git commit -m "fix: critical bug in X"

# 3. Create tag
git tag -a v1.2.1 -m "Hotfix: critical bug in X"

# 4. Push tag
git push origin v1.2.1

# 5. Merge back to main
git checkout main
git merge hotfix/v1.2.1
git push origin main
```

## Rollback

If a release has issues:

### Delete GitHub Release

```bash
# Delete tag locally
git tag -d v1.2.0

# Delete tag remotely
git push origin :refs/tags/v1.2.0

# Delete GitHub release via UI or API
gh release delete v1.2.0
```

### Revert npm Package

```bash
# Deprecate version
npm deprecate scmd-cli@1.2.0 "This version has been deprecated due to critical bug"

# Publish fixed version
# ... fix code ...
npm version patch
npm publish
```

### Update Homebrew

The tap will auto-update, but for homebrew-core:

```bash
# Submit PR to revert formula
```

## Troubleshooting

### GoReleaser Fails

**Check configuration**:

```bash
make check-goreleaser
```

**Common issues**:

- Missing GITHUB_TOKEN
- Invalid .goreleaser.yml syntax
- Build failures for specific platforms

**Debug**:

```bash
# Run with verbose output
goreleaser release --clean --verbose
```

### npm Publish Fails

**Check token**:

```bash
# Verify npm token is valid
npm whoami
```

**Manual publish** (if automation fails):

```bash
cd npm
npm version 1.2.0 --no-git-tag-version
npm publish --access public
```

### Homebrew Tap Not Updated

**Check**:

1. Verify `HOMEBREW_TAP_GITHUB_TOKEN` is set
2. Check token has `repo` scope
3. Verify repository exists: `scmd/homebrew-tap`

**Manual update**:

```bash
# Clone tap
git clone https://github.com/scmd/homebrew-tap
cd homebrew-tap

# Update formula
# Edit Formula/scmd.rb with new version and checksums

# Commit and push
git commit -am "Update to v1.2.0"
git push
```

### Docker Build Fails

**Check Docker Hub credentials**:

- Verify `DOCKERHUB_USERNAME` and `DOCKERHUB_TOKEN`

**Manual build and push**:

```bash
# Build locally
docker build -t scmd/scmd:1.2.0 .

# Push
docker push scmd/scmd:1.2.0
```

## Advanced Topics

### Custom Binary Names

Edit `.goreleaser.yml`:

```yaml
builds:
  - id: scmd
    binary: scmd
    # Add custom naming
```

### Platform-Specific Builds

Skip platforms:

```yaml
builds:
  - id: scmd
    ignore:
      - goos: windows
        goarch: arm64
```

### Signing Binaries

Add GPG signing:

```yaml
signs:
  - artifacts: checksum
    args:
      - "--batch"
      - "--local-user"
      - "{{ .Env.GPG_FINGERPRINT }}"
      - "--output"
      - "${signature}"
      - "--detach-sign"
      - "${artifact}"
```

### Custom Release Notes

Edit `.goreleaser.yml`:

```yaml
changelog:
  filters:
    exclude:
      - '^docs:'
      - '^test:'
  groups:
    - title: 'Features'
      regexp: '^.*?feat(\([[:word:]]+\))??!?:.+$'
      order: 0
    - title: 'Bug Fixes'
      regexp: '^.*?fix(\([[:word:]]+\))??!?:.+$'
      order: 1
```

## References

- [GoReleaser Documentation](https://goreleaser.com)
- [Semantic Versioning](https://semver.org)
- [GitHub Actions](https://docs.github.com/en/actions)
- [Homebrew Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)
- [npm Publishing Guide](https://docs.npmjs.com/cli/v8/commands/npm-publish)
