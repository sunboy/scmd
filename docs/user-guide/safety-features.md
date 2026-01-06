# Safety Features

scmd includes powerful safety features to protect you from accidentally executing destructive commands. When tool-calling commands use the shell tool, scmd automatically detects dangerous operations and provides an interactive preview.

!!! tip "Proactive Safety"
    Safety features work automatically - no configuration needed. Destructive commands are detected and previewed before execution.

## Command Preview

### Overview

When scmd detects a potentially destructive shell command, it pauses execution and shows you:

- **Warning banner** with severity level
- **Command breakdown** with detected risks
- **Impact estimate** (affected files, size, scope)
- **Interactive prompt** with actions

### Severity Levels

Commands are classified by severity:

| Severity | Icon | Examples | Behavior |
|----------|------|----------|----------|
| Low | ℹ️ | `kill`, basic `chmod` | Execute without preview |
| Medium | ⚠️ | `pkill`, `chmod 777` | Show preview |
| High | 🔥 | `rm files`, `git reset --hard` | Show preview |
| Critical | 💀 | `rm -rf`, `DROP TABLE`, `dd` | Show preview with emphasis |

!!! note "Threshold"
    Only Medium severity and above trigger the preview. Low severity commands execute normally.

### Interactive Actions

When a preview is shown, you have four options:

#### [E]dit - Modify the Command

Opens your `$EDITOR` (or `vi` as fallback) to modify the command before execution:

```
What would you like to do?
  [E]dit command
  [D]ry-run (show what would happen)
  [Enter] Execute anyway
  [Q]uit / Cancel

Choice: e
```

After editing, the modified command is re-validated and executed (or previewed again if still destructive).

**Example flow:**
```
Original: rm -rf /tmp/old-data
(Edit in vi)
Modified: rm -rf /tmp/old-data/cache-only
(Preview shown again if needed)
Execute
```

#### [D]ry-run - Preview Without Executing

Shows what would happen without making any changes:

```
Choice: d

[DRY RUN] Would execute: rm -rf node_modules

No actual changes made.
```

This is useful for:
- Verifying the command is correct
- Understanding the scope of changes
- Testing command composition

#### [Enter] - Execute Anyway

Proceeds with execution of the original command. Use this when:
- You've reviewed the command and it's correct
- You understand the impact
- You want to proceed with the operation

```
Choice: (press Enter)

Executing: rm -rf node_modules

(command executes normally)
```

#### [Q]uit - Cancel Operation

Cancels the command entirely without executing anything:

```
Choice: q

Command cancelled by user
```

The LLM will receive a message that the command was cancelled.

## Detected Dangerous Patterns

scmd detects 20+ categories of destructive operations:

### File Deletion

```bash
rm file.txt                    # High severity
rm -r directory                # Critical severity
rm -rf *                       # Critical severity
```

**Detected patterns:**
- Recursive deletion (`-r`, `-R`, `-rf`)
- Wildcard deletion (`*`, `?`)
- Force deletion (`-f`)

**Impact estimate:**
- File count (estimated from patterns)
- Directory vs. files
- Special cases (e.g., `node_modules` = ~10,000 files)

### Git Operations

```bash
git push --force               # Critical severity
git push -f origin main        # Critical severity
git reset --hard HEAD          # High severity
git clean -fd                  # High severity
git branch -D feature          # Medium severity
```

**Why dangerous:**
- `--force push`: Rewrites remote history, affects team
- `--hard reset`: Discards uncommitted changes permanently
- `clean -fd`: Removes untracked files (can't be recovered)
- `-D branch`: Force deletes branch without merge check

### Docker Operations

```bash
docker system prune -a         # High severity
docker rm -f container         # Medium severity
docker volume rm data-vol      # High severity (DATA LOSS)
```

**Impact:**
- `prune -a`: Removes all unused resources
- `volume rm`: Permanent data loss
- `-f`: Force removal without confirmation

### Kubernetes

```bash
kubectl delete pod my-pod      # High severity
kubectl delete deployment app  # High severity
kubectl delete namespace prod  # Critical severity
```

**Impact:**
- Service disruption
- Data loss if stateful
- Affects production systems

### Database Operations

```bash
DROP TABLE users;              # Critical severity
TRUNCATE TABLE logs;           # Critical severity
DELETE FROM customers;         # Critical severity
```

**Why critical:**
- Permanent data loss
- No undo (unless backups exist)
- Can affect production systems

### System Operations

```bash
shutdown now                   # Critical severity
reboot                         # Critical severity
dd if=/dev/zero of=/dev/sda    # Critical severity
mkfs.ext4 /dev/sdb1            # Critical severity
```

**Impact:**
- System downtime
- Disk formatting (irreversible)
- Data destruction

### Process Management

```bash
kill -9 12345                  # Medium severity
pkill node                     # Medium severity
killall python                 # Medium severity
```

**Risks:**
- Force kills (no graceful shutdown)
- May kill multiple processes
- Can disrupt services

### Permissions

```bash
chmod 777 secret.key           # Medium severity (security risk)
chown -R nobody:nobody /       # Medium severity
```

**Why flagged:**
- Security vulnerabilities (world-writable)
- Recursive changes can break system
- Permission errors can lock you out

### Package Management

```bash
npm uninstall -g typescript    # Medium severity
apt remove nginx               # High severity
pip uninstall -y pandas        # Medium severity
```

**Impact:**
- Global package removal
- System package removal (may break dependencies)
- Can affect other projects

## Examples

### Example 1: File Deletion with Preview

A command using the shell tool tries to delete files:

```
💀 CRITICAL DESTRUCTIVE COMMAND DETECTED
============================================================

Command:
  rm -rf node_modules dist build

Detected Risks:
  1. 💀 Recursive file deletion - CANNOT BE UNDONE
     Matched: 'rm -rf'

Estimated Impact:
  Affects: directories (recursive)
  Count: ~10000 items
  Size: ~500 MB

What would you like to do?
  [E]dit command
  [D]ry-run (show what would happen)
  [Enter] Execute anyway
  [Q]uit / Cancel

Choice:
```

**User chooses [D]ry-run:**

```
[DRY RUN] Would execute: rm -rf node_modules dist build

No actual changes made.
```

LLM receives: "Dry run completed. No changes made."

### Example 2: Git Force Push with Edit

```
💀 CRITICAL DESTRUCTIVE COMMAND DETECTED
============================================================

Command:
  git push --force origin main

Detected Risks:
  1. 💀 Force push - rewrites remote history
     Matched: 'git push --force'

Estimated Impact:
  Affects: git commits (remote)
  Count: Unknown (potentially many)

What would you like to do?
  [E]dit command
  [D]ry-run (show what would happen)
  [Enter] Execute anyway
  [Q]uit / Cancel

Choice: e
```

**User edits in $EDITOR:**

Original:
```bash
git push --force origin main
```

Modified to safer option:
```bash
git push --force-with-lease origin feature-branch
```

**Result:**
- Safer `--force-with-lease` instead of `--force`
- Target is feature branch, not main
- Re-validated and executed

### Example 3: Docker Cleanup

```
🔥 HIGH DESTRUCTIVE COMMAND DETECTED
============================================================

Command:
  docker system prune -a

Detected Risks:
  1. 🔥 Remove all unused Docker resources
     Matched: 'docker system prune -a'

Estimated Impact:
  Affects: Docker resources
  Count: Unknown (potentially many)

What would you like to do?
  [E]dit command
  [D]ry-run (show what would happen)
  [Enter] Execute anyway
  [Q]uit / Cancel

Choice: (Enter to execute)

Executing: docker system prune -a

WARNING! This will remove:
  - all stopped containers
  - all networks not used by at least one container
  - all images without at least one container associated to them
  - all build cache

Are you sure you want to continue? [y/N]
```

Docker's own confirmation provides second layer of safety.

## Configuration

### Adjusting Sensitivity

Currently, the severity threshold is hardcoded to Medium. Commands with Low severity execute without preview.

Future versions will support configuration:

```yaml
# .scmdrc (planned)
safety:
  preview_threshold: high  # Only preview high/critical
  auto_dry_run: true       # Always dry-run first
  require_confirmation: true
```

### Custom Patterns

Future support for custom destructive patterns:

```yaml
# .scmdrc (planned)
safety:
  custom_patterns:
    - pattern: "terraform destroy"
      severity: critical
      description: "Destroys infrastructure"
```

### Whitelisting Commands

Currently, the shell tool has a whitelist of allowed commands. Only whitelisted base commands can execute.

See [Tool Calling Guide](../command-authoring/tool-calling.md#shell-tool) for the full whitelist.

## Best Practices

### For End Users

1. **Read the Preview**: Don't just hit Enter - read what command will execute
2. **Use Dry-Run**: When in doubt, use dry-run mode first
3. **Edit for Safety**: Modify commands to be safer (e.g., add specific paths instead of wildcards)
4. **Cancel if Unsure**: It's always safe to quit and try again

### For Command Authors

1. **Trust the Safety Net**: Let users leverage the preview system
2. **Educate the LLM**: Include safety notes in system prompts
3. **Suggest Dry-Runs**: Prompt LLMs to suggest dry-run for risky operations
4. **Provide Alternatives**: LLMs should offer safer alternatives when possible

Example system prompt:

```yaml
prompt:
  system: |
    You are a helpful assistant that uses the shell tool.

    IMPORTANT: The shell tool has automatic safety features:
    - Destructive commands show interactive previews
    - Users can edit, dry-run, execute, or cancel

    When suggesting destructive operations:
    1. Explain what will be deleted/changed
    2. Suggest dry-run mode first
    3. Offer safer alternatives if available
    4. Use specific paths instead of wildcards when possible
```

### Example: Safe Cleanup Command

```yaml
name: safe-cleanup
prompt:
  system: |
    Help clean up build artifacts safely.

    SAFETY PROTOCOL:
    1. ALWAYS list what will be deleted first (ls or find)
    2. Use specific paths, avoid wildcards when possible
    3. Suggest dry-run for review
    4. The preview system will catch dangerous operations

  template: |
    Clean up: {{.artifacts}}

    Steps:
    1. Use 'find' or 'ls' to list what will be deleted
    2. Show the list to the user
    3. Use 'rm' to delete (preview will activate automatically)
    4. Confirm deletion completed
```

## Troubleshooting

### Preview Not Showing

**Symptom:** Destructive command executes without preview

**Possible causes:**
1. Command severity is Low (below threshold)
2. Command not in detected patterns
3. Shell tool not being used (direct execution)

**Solutions:**
- Check if command matches patterns in detector
- Verify tool-calling is enabled
- Report missing patterns as feature requests

### Can't Edit Command

**Symptom:** Edit mode fails with "editor not found"

**Cause:** `$EDITOR` environment variable not set

**Solution:**
```bash
# Set your preferred editor
export EDITOR=nano  # or vim, emacs, code, etc.

# Or set permanently in ~/.bashrc or ~/.zshrc
echo 'export EDITOR=nano' >> ~/.bashrc
```

Fallback: scmd uses `vi` if `$EDITOR` is not set.

### False Positives

**Symptom:** Safe commands trigger preview

**Example:** `mkdir /tmp/test-rm-backup` (contains "rm" substring)

**Current status:** Pattern matching is substring-based, may have false positives

**Workaround:** Just press Enter to execute if you know it's safe

**Planned fix:** Improve regex patterns to be word-boundary aware

## Limitations

### Current Limitations

1. **Pattern-Based Detection**: Uses regex, not semantic analysis
   - May have false positives (safe commands flagged)
   - May miss variations of dangerous commands

2. **No Undo**: Preview prevents execution, but can't undo if you execute
   - Backups are your responsibility
   - Consider `git commit` before risky operations

3. **Limited to Shell Tool**: Only works for shell tool calls
   - Direct file tool writes not previewed
   - HTTP requests not previewed

4. **English Only**: Descriptions and UI in English only

### Future Enhancements

- [ ] Semantic command analysis (understand intent, not just pattern)
- [ ] Preview for file writes (show diff before writing)
- [ ] Preview for HTTP requests (show what will be sent)
- [ ] Configurable sensitivity levels
- [ ] Custom pattern definitions
- [ ] Integration with `trash` for recoverable deletion
- [ ] Backup before destructive operations
- [ ] Multi-language support

## Related Documentation

- [Tool Calling Guide](../command-authoring/tool-calling.md) - How LLMs use tools
- [Shell Tool Reference](../command-authoring/tool-calling.md#shell-tool) - Shell tool details
- [Command Examples](../examples/tool-calling-examples.md) - Real-world usage

## FAQ

**Q: Can I disable the preview?**
A: Not currently. Safety features are always on. Future versions may add configuration.

**Q: What if I want to bypass the preview for a specific command?**
A: You can press Enter to execute anyway. The preview is informational, not blocking.

**Q: How do I report a missing dangerous pattern?**
A: Open an issue on GitHub with the command that should be detected.

**Q: Can I add custom patterns?**
A: Not yet, but it's planned for a future release.

**Q: Does this prevent all mistakes?**
A: No. It's a safety net, not a guarantee. Always review commands and maintain backups.
