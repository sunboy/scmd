# Remaining Commands Specification

This document provides complete specifications for the 22 remaining commands to complete the 60-command repository.

## Docker Commands (6)

### 1. docker-cleanup.yaml
**Category:** docker
**Description:** Clean up stopped containers, unused images, volumes
**Args:** `aggressive` (bool, default: false), `dry_run` (bool, default: true)
**Key features:**
- Show disk usage: `docker system df`
- Remove stopped containers: `docker container prune`
- Remove unused images: `docker image prune -a`
- Remove volumes: `docker volume prune`
- Network cleanup: `docker network prune`
- Show space freed

### 2. docker-logs-follow.yaml
**Category:** docker
**Description:** Follow logs from multiple containers with filtering
**Args:** `containers` (pattern or IDs), `since` (time), `follow` (bool)
**Key features:**
- `docker logs -f CONTAINER`
- Multi-container: `docker-compose logs -f`
- Filter by time: `--since`
- Search logs: `grep`

### 3. docker-resource-usage.yaml
**Category:** docker
**Description:** Show resource usage of all containers (CPU, memory, network)
**Args:** none
**Key features:**
- `docker stats --no-stream`
- Show CPU%, MEM%, NET I/O, BLOCK I/O
- Sort by resource usage
- Identify resource hogs

### 4. docker-shell.yaml
**Category:** docker
**Description:** Intelligently open shell in container (bash/sh/ash detection)
**Args:** `container` (name or ID), `user` (optional)
**Key features:**
- Try bash first: `docker exec -it CONTAINER bash`
- Fall back to sh: `docker exec -it CONTAINER sh`
- Try ash for Alpine
- Show available shells in container

### 5. docker-network-inspect.yaml
**Category:** docker
**Description:** Inspect and troubleshoot Docker networks
**Args:** `network` (name, optional - shows all if not specified)
**Key features:**
- `docker network ls`
- `docker network inspect NETWORK`
- Show connected containers
- Show IP addresses
- Test connectivity between containers

### 6. docker-compose-helper.yaml
**Category:** docker
**Description:** Generate docker-compose.yml from running containers
**Args:** `containers` (optional filter)
**Key features:**
- Inspect running containers
- Extract image, ports, volumes, env vars
- Generate valid docker-compose.yml
- Save to file

## Development Commands (8)

### 1. setup-project.yaml
**Category:** development
**Description:** Initialize new project (gitignore, README, license)
**Args:** `project_name`, `language` (js/py/go/rust), `license` (MIT/Apache)
**Key features:**
- `git init`
- Generate appropriate .gitignore
- Create README.md template
- Add LICENSE file
- Initialize package manager (npm/pip/go mod)

### 2. dependency-audit.yaml
**Category:** development
**Description:** Audit dependencies for vulnerabilities
**Args:** `fix` (bool, attempt to fix)
**Key features:**
- `npm audit` (Node.js)
- `pip-audit` (Python)
- `go list -m all && go mod verify` (Go)
- `cargo audit` (Rust)
- Show severity levels
- Suggest fixes

### 3. port-is-free.yaml
**Category:** development
**Description:** Check if port is available, kill process if needed
**Args:** `port`, `kill` (bool)
**Key features:**
- Check port: `lsof -i :PORT`
- Show process using port
- Optional kill with confirmation
- Suggest alternative ports

### 4. generate-env-template.yaml
**Category:** development
**Description:** Generate .env.example from .env file
**Args:** `env_file` (default: .env)
**Key features:**
- Read .env
- Replace values with placeholders
- Preserve structure and comments
- Save as .env.example
- Never commit secrets

### 5. check-outdated-deps.yaml
**Category:** development
**Description:** Check for outdated dependencies
**Args:** `update` (bool)
**Key features:**
- `npm outdated` (Node.js)
- `pip list --outdated` (Python)
- `go list -u -m all` (Go)
- `cargo outdated` (Rust)
- Show current vs latest versions
- Highlight breaking changes

### 6. run-all-tests.yaml
**Category:** development
**Description:** Run tests with coverage and timing
**Args:** `pattern` (test pattern filter)
**Key features:**
- Auto-detect test framework
- Run with coverage: `--coverage`
- Show timing for slow tests
- Generate coverage report
- Fail fast option

### 7. benchmark-code.yaml
**Category:** development
**Description:** Run and compare code benchmarks
**Args:** `benchmark_pattern`
**Key features:**
- `go test -bench=. -benchmem` (Go)
- `pytest --benchmark-only` (Python)
- `cargo bench` (Rust)
- Compare with baseline
- Show performance metrics

### 8. lint-and-format.yaml
**Category:** development
**Description:** Auto-detect and run linters/formatters
**Args:** `fix` (bool, auto-fix issues)
**Key features:**
- Detect language
- Run appropriate linter (eslint, pylint, golint)
- Run formatter (prettier, black, gofmt)
- Show issues with file:line
- Optional auto-fix

## Text Processing Commands (5)

### 1. grep-advanced.yaml
**Category:** text
**Description:** Advanced text search with context and highlighting
**Args:** `pattern`, `files`, `context_lines` (before/after)
**Key features:**
- `grep -rn PATTERN`
- Show context: `-A`, `-B`, `-C`
- Recursive search
- Ignore case: `-i`
- Count matches: `-c`
- Color output

### 2. csv-to-json.yaml
**Category:** text
**Description:** Convert CSV to JSON with column mapping
**Args:** `csv_file`, `output_file`
**Key features:**
- Parse CSV headers
- Convert rows to JSON objects
- Handle quoted fields
- Pretty print JSON
- Validate output

### 3. json-query.yaml
**Category:** text
**Description:** Query JSON files with jq-like syntax
**Args:** `json_file`, `query`
**Key features:**
- Use `jq` if available
- Parse JSON
- Filter by key/value
- Extract nested fields
- Format output

### 4. extract-urls.yaml
**Category:** text
**Description:** Extract all URLs from files or stdin
**Args:** `files` or stdin
**Key features:**
- Regex pattern for URLs
- Extract http/https URLs
- Deduplicate
- Validate URLs
- Show count

### 5. count-lines-by-type.yaml
**Category:** text
**Description:** Count lines of code by file type
**Args:** `directory`
**Key features:**
- Count by extension
- Exclude comments
- Exclude blank lines
- Show percentage breakdown
- Total LOC

## SSH Commands (3)

### 1. ssh-tunnel.yaml
**Category:** ssh
**Description:** Create SSH tunnel with port forwarding
**Args:** `local_port`, `remote_host`, `remote_port`, `ssh_host`
**Key features:**
- Local forward: `ssh -L`
- Remote forward: `ssh -R`
- Dynamic forward (SOCKS): `ssh -D`
- Show connection status
- Background process

### 2. ssh-copy-id-helper.yaml
**Category:** ssh
**Description:** Copy SSH key to remote server with troubleshooting
**Args:** `user@host`, `key_file` (optional)
**Key features:**
- Use `ssh-copy-id` if available
- Manual copy: `cat ~/.ssh/id_rsa.pub | ssh USER@HOST 'cat >> ~/.ssh/authorized_keys'`
- Verify key exists
- Test connection after
- Fix permissions if needed
- Troubleshoot common issues

### 3. remote-command.yaml
**Category:** ssh
**Description:** Execute command on multiple remote hosts
**Args:** `hosts` (comma-separated), `command`
**Key features:**
- Loop through hosts
- Execute command: `ssh HOST 'COMMAND'`
- Collect output
- Show which host output came from
- Parallel execution option
- Handle failures gracefully

## Template Structure

All commands should follow this YAML structure:

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

hooks:
  pre:
    - shell: validation commands
  post:
    - shell: cleanup commands

prompt:
  system: |
    System prompt with expert persona and guidelines

  template: |
    Template using {{.args}} with clear instructions

model:
  temperature: 0.3  # Lower for precise operations
  max_tokens: 2000

examples:
  - scmd /command-name example1
  - scmd /command-name example2
```

## Quality Standards

- Safety-first: Always confirm destructive operations
- Clear output: Use formatting (boxes, colors, symbols)
- Error handling: Helpful error messages with solutions
- Cross-platform: Support macOS and Linux where possible
- Examples: 2-3 practical usage examples
- Documentation: Explain what the command does and why

## Next Steps

1. Create YAML files for all 22 remaining commands using above specs
2. Test a few critical commands to verify format
3. Create comprehensive README.md
4. Create manifest.yaml for repository metadata
5. Verify all 60 commands are present and properly formatted
