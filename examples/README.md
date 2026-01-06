# scmd Command Examples

This directory contains example command specifications demonstrating scmd's advanced features.

## Features Demonstrated

### 1. Hooks (Pre/Post Execution)

**File**: `commands/git-commit-with-hooks.yaml`

Demonstrates how to use pre and post-execution hooks to run shell commands before and after the main LLM completion.

```yaml
hooks:
  pre:
    - shell: git status --short
    - shell: echo "Generating commit message..."
  post:
    - shell: echo "Commit message generated!"
    - shell: git diff --stat
```

**Usage**:
```bash
scmd git-commit-with-hooks
```

### 2. Command Composition (Pipeline)

**File**: `commands/analyze-then-summarize.yaml`

Demonstrates how to chain commands together in a pipeline, where the output of one command feeds into the next.

```yaml
compose:
  pipeline:
    - command: analyze-code
      transform: trim
      on_error: continue
    - command: summarize
      args:
        format: bullet-points
      transform: trim
```

**Features**:
- Chain multiple commands
- Transform outputs between steps
- Error handling (continue, stop, fallback)

### 3. Tool Calling (Agentic Behavior)

**File**: `commands/project-analyzer.yaml`

Demonstrates a command designed to leverage LLM tool calling capabilities. When executed with a backend that supports tool calling (like llama.cpp with appropriate models), the LLM can autonomously:

- Execute shell commands (`ls`, `find`, `git status`)
- Read files (`package.json`, `README.md`, etc.)
- Fetch HTTP resources
- Make multiple tool calls iteratively to gather information

**Available Tools**:
- `shell`: Execute safe shell commands (whitelist enforced)
- `read_file`: Read file contents
- `write_file`: Write to files (with confirmation)
- `http_get`: Fetch URLs

**Usage**:
```bash
scmd project-analyzer ./my-project
```

## Installing Examples

To install these example commands:

```bash
# Copy to your scmd commands directory
cp examples/commands/*.yaml ~/.scmd/commands/

# Or create a local repository
scmd repo add examples file://$(pwd)/examples/commands
scmd repo install examples/git-commit-with-hooks
```

## Command Composition Types

scmd supports three types of composition:

### Pipeline
Chains commands sequentially, passing output as input:
```yaml
compose:
  pipeline:
    - command: step1
    - command: step2
    - command: step3
```

### Parallel
Runs commands concurrently and merges results:
```yaml
compose:
  parallel:
    - command1
    - command2
    - command3
```

### Fallback
Tries commands in order until one succeeds:
```yaml
compose:
  fallback:
    - primary-command
    - backup-command
    - last-resort-command
```

## Tool Security

Tool calling includes security measures:

1. **Shell Tool**: Only whitelisted commands are allowed (ls, git, cat, etc.)
2. **Write File**: Requires user confirmation before writing
3. **HTTP**: Limited to 10MB response size
4. **Timeout**: All operations have timeout protection

## Creating Your Own Commands

See the [scmd documentation](../README.md) for details on creating custom commands with:
- Custom prompts and templates
- Argument and flag definitions
- Model preferences (temperature, max_tokens)
- Dependencies on other commands
- Context requirements (files, git, env vars)

## Testing Commands

Test a command before installing:

```bash
# Test with scmd
scmd -f examples/commands/project-analyzer.yaml project-analyzer .

# Or install locally
cp examples/commands/project-analyzer.yaml ~/.scmd/commands/
scmd project-analyzer .
```
