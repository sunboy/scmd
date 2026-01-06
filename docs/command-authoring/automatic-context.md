# Automatic Context Gathering

Give your commands contextual awareness by automatically gathering files, git status, and environment variables before LLM execution. Context gathering happens transparently - no manual piping or file reading needed.

!!! tip "Why Context Matters"
    LLMs make better decisions with relevant context. Instead of manually piping files or explaining your environment, let scmd gather context automatically.

## Overview

The `context` field in command YAML enables automatic gathering of:

- **Files**: Match patterns and include file contents
- **Git**: Branch, status, recent commits, remote URL
- **Environment**: Specific environment variables
- **Token Management**: Automatic truncation to fit token budgets

## Basic Usage

### Minimal Example

```yaml
name: review-changes
description: Review uncommitted changes with git context

context:
  git: true  # Include git branch, status, recent commits

prompt:
  template: |
    Review my uncommitted changes.
    Focus on code quality and potential issues.
```

When executed, scmd automatically prepends:

```markdown
## Git Context

**Branch:** feature/new-feature

**Status:**
```
M  internal/repos/executor.go
M  internal/tools/shell.go
?? new-file.go
```

**Recent Commits:**
```
a1b2c3d feat: add new feature
e4f5g6h fix: bug in processor
i7j8k9l docs: update README
```

---

(Your prompt follows here)
```

### File Context Example

```yaml
name: analyze-go-project
description: Analyze Go project structure and dependencies

context:
  files:
    - "go.mod"
    - "go.sum"
    - "*.go"
    - "internal/**/*.go"
  max_tokens: 8000

prompt:
  template: |
    Analyze this Go project's structure and dependencies.
    {{.input}}
```

Gathered context includes:

```markdown
## Files

### go.mod
```
module github.com/user/project

go 1.21

require (
    github.com/pkg/errors v0.9.1
)
```

### main.go
```
package main

func main() {
    // ...
}
```

### internal/app/app.go
```
package app

// ...
```

---

(Your prompt follows)
```

## Context Specification

### Full Schema

```yaml
context:
  files:     []string  # Glob patterns
  git:       bool      # Include git context
  env:       []string  # Environment variable names
  max_tokens: int      # Maximum context tokens
```

### Files

Glob patterns to match files for inclusion:

```yaml
context:
  files:
    - "*.go"              # All Go files in current dir
    - "src/**/*.ts"       # All TypeScript files in src/ recursively
    - "package.json"      # Specific file
    - "config/*.yaml"     # All YAML files in config/
    - "README.md"         # Documentation
```

**Features:**
- Standard glob syntax (`*`, `**`, `?`)
- Relative to working directory
- Absolute paths supported
- Directories are skipped automatically
- Files >1MB show size warning instead of content

**Examples:**

| Pattern | Matches |
|---------|---------|
| `*.js` | All JS files in current directory |
| `**/*.py` | All Python files recursively |
| `src/**/*.{ts,tsx}` | All TypeScript files in src/ |
| `test/**/*_test.go` | All Go test files in test/ |

### Git Context

Enable with `git: true`:

```yaml
context:
  git: true
```

**Includes:**
- Current branch name
- Working tree status (`git status --short`)
- Recent 5 commits (`git log --oneline -5`)
- Remote origin URL

**Example output:**

```markdown
## Git Context

**Branch:** main

**Status:**
```
M  src/app.ts
A  src/new-feature.ts
D  src/old-file.ts
```

**Recent Commits:**
```
abc123 feat: add new feature
def456 fix: resolve bug in handler
ghi789 docs: update API documentation
jkl012 refactor: simplify error handling
mno345 test: add unit tests
```
```

**Non-blocking:** If not a git repository, context gathering continues without error.

### Environment Variables

Specify environment variables to include:

```yaml
context:
  env:
    - "GOPATH"
    - "GO111MODULE"
    - "NODE_ENV"
    - "DATABASE_URL"
    - "API_KEY"
```

**Example output:**

```markdown
## Environment

- `GOPATH`: /Users/user/go
- `GO111MODULE`: on
- `NODE_ENV`: development
```

**Security note:** Be careful with sensitive variables. Consider:
- Only include variables needed for the task
- Don't include secrets if sharing LLM conversations
- Use environment-specific commands

### Token Limits

Control context size with `max_tokens`:

```yaml
context:
  files:
    - "**/*.go"
  max_tokens: 8000  # ~32KB of text
```

**Token estimation:**
- Rough estimate: 1 token ≈ 4 characters
- Files contribute most tokens
- Git context: ~50-200 tokens
- Environment: minimal

**Truncation behavior:**
When context exceeds `max_tokens`:

1. Files are removed largest-first
2. Git context is preserved
3. Environment variables are preserved
4. Truncation continues until under limit

**Example:**

If you have:
- 100 Go files totaling 50,000 tokens
- `max_tokens: 8000`

Result:
- Largest files removed first
- ~20 most important files kept
- Git and env context preserved

!!! tip "Choosing Token Limits"
    - Small models (7B): 2000-4000 tokens
    - Medium models (13B): 4000-8000 tokens
    - Large models (70B): 8000-16000 tokens
    - Consider leaving room for response (~2000 tokens)

## Complete Examples

### Example 1: Code Review Command

```yaml
name: contextual-code-review
version: 1.0.0
description: Review code with full project context

context:
  files:
    - "*.go"
    - "go.mod"
    - "internal/**/*.go"
  git: true
  env:
    - "GOPATH"
    - "GO111MODULE"
  max_tokens: 10000

prompt:
  system: |
    You are an expert Go code reviewer.

    You have access to:
    - Project Go files
    - Go module dependencies
    - Git status and recent commits
    - Go environment configuration

    Review for:
    - Code quality
    - Go best practices
    - Potential bugs
    - Performance issues

  template: |
    Review the following changes: {{.input}}

    Consider the project context provided above.

args:
  - name: focus
    description: "Focus area (quality, security, performance)"
    required: false
    default: "all"

examples:
  - "git diff | scmd /contextual-code-review"
  - "cat new-feature.go | scmd /contextual-code-review security"
```

**Usage:**

```bash
git diff | scmd /contextual-code-review

# Context automatically gathered:
# - All .go files
# - go.mod
# - Git branch: feature/new-api
# - Git status: 3 modified files
# - GOPATH and GO111MODULE
# Then: LLM reviews diff with full context
```

### Example 2: Documentation Generator

```yaml
name: auto-docs
version: 1.0.0
description: Generate documentation from code

context:
  files:
    - "src/**/*.ts"
    - "package.json"
    - "README.md"
  max_tokens: 12000

prompt:
  system: |
    You are a technical writer creating API documentation.

    You have access to:
    - All source code
    - Package dependencies
    - Existing README

    Generate:
    - API reference
    - Usage examples
    - Installation instructions

  template: |
    Generate comprehensive documentation for this project.

    {{.format | default "markdown"}}

hooks:
  post:
    - shell: cat $OUTPUT > docs/API.md

examples:
  - "scmd /auto-docs > docs/API.md"
```

### Example 3: Test Generator

```yaml
name: generate-tests
version: 1.0.0
description: Generate tests based on code context

context:
  files:
    - "{{.file}}"
    - "*_test.go"
  git: true
  max_tokens: 6000

prompt:
  system: |
    You are an expert at writing Go tests.

    Context provided:
    - Source file to test
    - Existing test files (patterns to follow)
    - Git history (understand recent changes)

    Generate:
    - Table-driven tests
    - Edge cases
    - Error conditions
    - Follow existing test patterns

  template: |
    Generate comprehensive tests for: {{.file}}

args:
  - name: file
    description: "File to generate tests for"
    required: true

examples:
  - "scmd /generate-tests handler.go > handler_test.go"
```

### Example 4: Environment-Aware Deployment

```yaml
name: deploy-check
version: 1.0.0
description: Pre-deployment checks with environment awareness

context:
  files:
    - "Dockerfile"
    - "docker-compose.yml"
    - ".env.example"
  git: true
  env:
    - "NODE_ENV"
    - "DATABASE_URL"
    - "REDIS_URL"
    - "API_VERSION"
  max_tokens: 5000

prompt:
  system: |
    You are a DevOps expert performing pre-deployment checks.

    Context provided:
    - Docker configuration
    - Environment variables
    - Git status and branch

    Verify:
    - No uncommitted changes
    - Environment variables set correctly
    - Docker config matches environment
    - Not deploying from wrong branch

  template: |
    Perform pre-deployment check for {{.environment}}.

    Report any issues or confirm ready to deploy.

args:
  - name: environment
    description: "Target environment (staging, production)"
    required: true

examples:
  - "scmd /deploy-check production"
```

## How It Works

### Execution Flow

1. **Command Invoked**
   ```bash
   scmd /my-command
   ```

2. **Pre-Processing**
   - Parse command spec
   - Check if `context` field exists

3. **Context Gathering** (if specified)
   - Gather files matching patterns
   - Collect git information
   - Read environment variables
   - Estimate token count
   - Truncate if exceeds `max_tokens`

4. **Context Formatting**
   - Format as markdown
   - Structure: Files → Git → Environment

5. **Prompt Construction**
   - Prepend formatted context
   - Add separator (`---`)
   - Add user prompt

6. **LLM Execution**
   - Send combined prompt to LLM
   - LLM has full context
   - Generate response

7. **Result**
   - Return LLM output to user

### Context Format

Formatted context structure:

```markdown
## Files

### path/to/file1.go
```
(file contents)
```

### path/to/file2.go
```
(file contents)
```

## Git Context

**Branch:** feature-branch

**Status:**
```
(git status --short output)
```

**Recent Commits:**
```
(git log --oneline -5 output)
```

## Environment

- `VAR1`: value1
- `VAR2`: value2

---

(Your prompt here)
```

## Performance Considerations

### Token Budget

Context gathering can be expensive in tokens:

| Context Type | Typical Tokens | Notes |
|--------------|----------------|-------|
| Small file (100 lines) | ~500 | 1 token ≈ 4 chars |
| Large file (1000 lines) | ~5000 | Consider truncation |
| Git context | ~100-200 | Minimal impact |
| Environment (5 vars) | ~50 | Negligible |
| **Total for 10 files** | ~5000-10000 | Set `max_tokens` appropriately |

### File Reading

- Files are read synchronously
- Large projects may have slight delay
- Use specific patterns to reduce files read
- Consider `max_tokens` to auto-limit

### Git Commands

- Git operations are fast (< 100ms typically)
- Run in parallel where possible
- Fail gracefully if not a git repo

## Best Practices

### 1. Be Specific with File Patterns

**Good:**
```yaml
context:
  files:
    - "src/handlers/*.go"  # Only handlers
    - "config/app.yaml"    # Specific config
```

**Bad:**
```yaml
context:
  files:
    - "**/*"  # Everything (may hit token limit)
```

### 2. Set Appropriate Token Limits

```yaml
context:
  files:
    - "**/*.go"
  max_tokens: 8000  # Explicit limit
```

Without `max_tokens`, large projects may send too much context.

### 3. Use Environment Variables Wisely

**Good:**
```yaml
context:
  env:
    - "NODE_ENV"  # Relevant to task
```

**Bad:**
```yaml
context:
  env:
    - "DATABASE_PASSWORD"  # Security risk!
```

### 4. Combine Context Types

```yaml
context:
  files:
    - "*.go"
  git: true  # Git provides extra context
  env:
    - "GOPATH"
```

More context = better decisions, but watch token limits.

### 5. Document Context Usage

```yaml
name: my-command
description: |
  Does X with automatic context.

  Context gathered:
  - All Go files
  - Git status
  - GOPATH environment

context:
  files: ["*.go"]
  git: true
  env: ["GOPATH"]
```

Helps users understand what information is used.

## Troubleshooting

### Context Not Appearing

**Symptom:** LLM doesn't seem to have context

**Debug:**
1. Check command spec has `context:` field
2. Verify files match patterns
3. Check `max_tokens` not too restrictive
4. Look for errors in output

**Solution:**
```bash
# Run with debug to see context gathering
SCMD_DEBUG=1 scmd /my-command
```

### Token Limit Exceeded

**Symptom:** Context warning about truncation

**Cause:** Files exceed `max_tokens`

**Solution:**
- Increase `max_tokens`
- Use more specific file patterns
- Exclude large generated files

```yaml
context:
  files:
    - "*.go"
    - "!*_generated.go"  # Exclude generated
  max_tokens: 12000  # Increase limit
```

### Sensitive Data Exposure

**Symptom:** Environment variables include secrets

**Solution:**
Only include non-sensitive variables:

```yaml
context:
  env:
    - "NODE_ENV"     # Safe
    - "LOG_LEVEL"    # Safe
    # NOT: DATABASE_PASSWORD, API_SECRET
```

### File Too Large

**Symptom:** `[File too large: X bytes]` in context

**Cause:** File >1MB (safety limit)

**Solution:**
- Exclude large files with patterns
- Split large files into smaller modules
- Use more specific patterns

## Limitations

### Current Limitations

1. **No Recursive Limits**: `**/*.go` may match thousands of files
   - Mitigate with `max_tokens`

2. **No File Filtering**: Can't filter file contents
   - Workaround: Use hooks to preprocess

3. **Static Patterns**: Can't use dynamic patterns
   - Planned: Template support in patterns

4. **No Caching**: Context re-gathered each execution
   - Planned: Context caching

5. **English Only**: Context headers in English
   - Planned: Internationalization

### Future Enhancements

- [ ] Dynamic file patterns with templates
- [ ] File content filtering (e.g., only public functions)
- [ ] Context caching for repeated executions
- [ ] Semantic file selection (most relevant files)
- [ ] Custom context sources (databases, APIs)
- [ ] Context compression strategies

## Related Documentation

- [Command YAML Specification](yaml-specification.md) - Full spec reference
- [Prompts and Templates](prompts-and-templates.md) - Writing effective prompts
- [Tool Calling Guide](tool-calling.md) - Combining with tools
- [Best Practices](best-practices.md) - Command design patterns

## Examples

See [full examples](../examples/) for:
- [code-review-with-context.yaml](../../examples/commands/code-review-with-context.yaml) - Real command using context
- [Tool calling examples](../examples/tool-calling-examples.md) - Context + tools
