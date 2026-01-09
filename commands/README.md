# SCMD Command Repository

A comprehensive collection of 60 AI-powered slash commands for common Unix workflows, system administration, development tasks, and more.

## Overview

This repository contains production-ready slash commands that help users with:
- **File Operations** - Finding, managing, and organizing files
- **Git Workflows** - Complex git operations made simple
- **System Administration** - Monitoring, diagnostics, and maintenance
- **Network Tools** - Connectivity testing, debugging, and monitoring
- **Docker Management** - Container and image lifecycle
- **Development** - Testing, linting, dependency management
- **Text Processing** - Data transformation and analysis
- **SSH Operations** - Remote access and tunneling

## Installation

```bash
# Copy commands to your local scmd directory
cp -r commands/* ~/.scmd/commands/

# Or install from repository (when published)
scmd repo install scmd-team/commands
```

## Command Index

### File Operations (15 commands)

| Command | Description |
|---------|-------------|
| `/find-large-files` | Find files larger than specified size, sorted by size |
| `/find-recent` | Find recently modified files in last N hours/days |
| `/find-duplicates` | Find duplicate files by comparing content hash |
| `/bulk-rename` | Rename multiple files with patterns (prefix/suffix/replace) |
| `/safe-delete` | Move files to trash instead of permanent deletion |
| `/disk-usage` | Show disk usage with visual tree and largest directories |
| `/find-empty` | Find and optionally delete empty files/directories |
| `/change-permissions-recursive` | Safely change permissions recursively with preview |
| `/find-broken-symlinks` | Find and fix or remove broken symbolic links |
| `/archive-old-files` | Archive files older than N days to tar.gz |
| `/sync-directories` | Sync two directories with rsync (with dry-run) |
| `/find-by-extension` | Find files by extension and show statistics |
| `/batch-convert-images` | Convert images between formats (jpg↔png, resize) |
| `/find-and-replace-filename` | Find/replace text in filenames |
| `/check-file-encoding` | Detect and convert file encodings |

### Git Operations (8 commands)

| Command | Description |
|---------|-------------|
| `/git-undo` | Intelligently undo last commit/push/merge with safety checks |
| `/git-cleanup-branches` | Delete merged branches (local and remote) |
| `/git-find-commit` | Search commits by message/author/date/file/content |
| `/git-interactive-rebase` | AI-assisted interactive rebase with explanations |
| `/git-bisect-helper` | Guide through git bisect to find bug-introducing commit |
| `/git-stash-manager` | List, apply, pop, and manage git stashes |
| `/git-conflict-resolver` | Help resolve merge conflicts with context |
| `/git-blame-analysis` | Analyze git blame with context and author stats |

### System Administration (8 commands)

| Command | Description |
|---------|-------------|
| `/monitor-system` | Real-time system monitoring (CPU, RAM, disk, network) |
| `/find-port-user` | Find which process is using a specific port |
| `/check-service-status` | Check status of system services (systemd/launchd) |
| `/analyze-logs` | Search and analyze log files for errors/patterns |
| `/disk-cleanup` | Find and remove caches, temp files, old logs |
| `/process-tree` | Show process tree with resource usage |
| `/check-startup-programs` | List and manage programs that start on boot |
| `/system-health-check` | Comprehensive system health diagnostic |

### Network Tools (7 commands)

| Command | Description |
|---------|-------------|
| `/test-connectivity` | Test network connectivity (ping, traceroute, DNS) |
| `/port-scan` | Scan ports on localhost or remote host |
| `/network-speed-test` | Test download/upload speed and latency |
| `/dns-lookup` | Comprehensive DNS lookup with all record types |
| `/http-request-debug` | Debug HTTP requests with headers and timing |
| `/find-local-ips` | Show all local IP addresses and interfaces |
| `/network-bandwidth-monitor` | Monitor bandwidth usage by process |

### Docker Management (6 commands)

| Command | Description |
|---------|-------------|
| `/docker-cleanup` | Clean up stopped containers, unused images, volumes |
| `/docker-logs-follow` | Follow logs from multiple containers |
| `/docker-resource-usage` | Show resource usage of all containers |
| `/docker-shell` | Intelligently open shell in container (bash/sh/ash) |
| `/docker-network-inspect` | Inspect and troubleshoot Docker networks |
| `/docker-compose-helper` | Generate docker-compose.yml from running containers |

### Development Workflows (8 commands)

| Command | Description |
|---------|-------------|
| `/setup-project` | Initialize new project (gitignore, README, license) |
| `/dependency-audit` | Audit dependencies for vulnerabilities |
| `/port-is-free` | Check if port is available, kill process if needed |
| `/generate-env-template` | Generate .env.example from .env |
| `/check-outdated-deps` | Check for outdated dependencies |
| `/run-all-tests` | Run tests with coverage and timing analysis |
| `/benchmark-code` | Run and compare code benchmarks |
| `/lint-and-format` | Auto-detect and run linters/formatters |

### Text Processing (5 commands)

| Command | Description |
|---------|-------------|
| `/grep-advanced` | Advanced text search with context and highlighting |
| `/csv-to-json` | Convert CSV to JSON with column mapping |
| `/json-query` | Query JSON files with jq-like syntax |
| `/extract-urls` | Extract all URLs from files or stdin |
| `/count-lines-by-type` | Count lines of code by file type |

### SSH Operations (3 commands)

| Command | Description |
|---------|-------------|
| `/ssh-tunnel` | Create SSH tunnel with port forwarding |
| `/ssh-copy-id-helper` | Copy SSH key to remote server with troubleshooting |
| `/remote-command` | Execute command on multiple remote hosts |

## Usage Examples

### Find and clean up large files
```bash
# Find files larger than 1GB
scmd /find-large-files 1G

# Archive old log files
scmd /archive-old-files 90 ./logs "*.log"
```

### Git workflow automation
```bash
# Clean up merged branches
scmd /git-cleanup-branches both

# Find commits by author
scmd /git-find-commit author "john@example.com"

# Interactive rebase with guidance
scmd /git-interactive-rebase HEAD~5
```

### System monitoring and diagnostics
```bash
# Monitor system in real-time
scmd /monitor-system 2

# Run comprehensive health check
scmd /system-health-check

# Find what's using port 8080
scmd /find-port-user 8080
```

### Docker management
```bash
# Clean up Docker resources
scmd /docker-cleanup false false

# Show container resource usage
scmd /docker-resource-usage

# Open shell in running container
scmd /docker-shell mycontainer
```

## Features

### Safety First
- **Confirmation prompts** for destructive operations
- **Dry-run mode** for file operations (preview before execution)
- **Detailed explanations** of what each operation will do
- **Rollback instructions** when things go wrong

### Intelligent Assistance
- **Context-aware suggestions** based on error patterns
- **Auto-detection** of tools, languages, and environments
- **Alternative approaches** when primary method fails
- **Best practices** and optimization tips

### Clear Output
- **Formatted tables** and visual indicators
- **Progress tracking** for long operations
- **Categorized results** for easy scanning
- **Actionable next steps** after each command

## Command Structure

All commands follow a consistent YAML structure:

```yaml
name: command-name
version: 1.0.0
description: Brief description
category: category-name
author: scmd team
license: MIT

args:
  - name: arg_name
    description: Argument description
    required: true/false
    default: default_value

prompt:
  system: |
    System prompt defining expert persona
  template: |
    Command template with {{.args}} placeholders

model:
  temperature: 0.2-0.4  # Lower for precise operations
  max_tokens: 1500-4000  # Based on command complexity

examples:
  - scmd /command-name example1
  - scmd /command-name arg1 arg2
```

## Contributing

To add new commands to this repository:

1. Follow the command structure template
2. Include comprehensive error handling
3. Provide 2-3 usage examples
4. Add safety confirmations for destructive operations
5. Test on both macOS and Linux (when applicable)
6. Update this README with new command entry

## Requirements

- **scmd** - Main scmd CLI tool
- **Common Unix tools** - Most commands use standard Unix utilities
- **Optional tools** - Some commands benefit from: jq, nmap, docker, git

## License

MIT License - see individual command files for details

## Support

- **Documentation**: See individual command files for detailed usage
- **Issues**: Report problems or request features at the scmd repository
- **Examples**: Check `examples/` directory for advanced usage patterns

## Version

Repository version: 1.0.0
Last updated: 2024-01-09
Total commands: 60
