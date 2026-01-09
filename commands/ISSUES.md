# SCMD Command Repository - Issues Tracking

**Date:** 2026-01-09
**Test Report:** TEST_REPORT.md

---

## High Priority Issues

### QA-004: Manifest Discrepancy

**Severity:** High
**Status:** Open
**Category:** Documentation

**Description:**
The manifest.yaml file lists 60 total commands, but only 38 commands currently exist in the repository. While the manifest includes a stats section noting 38 completed and 22 remaining, the commands are all listed together without clear differentiation between implemented and planned commands.

**Impact:**
- Users may expect 60 commands to be available
- Confusion about repository completeness
- Unclear roadmap for future development

**Location:**
- File: /Users/sandeep/Projects/scmd/commands/manifest.yaml
- Lines: 29-337

**Current State:**
```yaml
commands:
  # File Operations (15)
  - name: find-large-files
    file: file-ops/find-large-files.yaml
    category: file-ops
    description: Find files larger than specified size
  # ... all 60 commands listed together

stats:
  total_commands: 60
  completed: 38
  remaining: 22
```

**Recommended Fix:**

Option 1 - Add status field:
```yaml
commands:
  - name: find-large-files
    file: file-ops/find-large-files.yaml
    category: file-ops
    description: Find files larger than specified size
    status: implemented  # NEW FIELD
```

Option 2 - Separate sections:
```yaml
implemented_commands:
  # 38 implemented commands here

planned_commands:
  # 22 planned commands here
```

**Estimated Effort:** 30 minutes

---

## Medium Priority Issues

### QA-001: Missing Args Declaration

**Severity:** Medium
**Status:** Open
**Category:** Structure

**Description:**
The network-speed-test.yaml file does not have an explicit args section defined. While this command doesn't require arguments, consistency across all commands dictates it should have `args: []` or `args:` with an empty list.

**Impact:**
- Structural inconsistency
- Potential parsing issues with some YAML processors
- Deviation from command template standard

**Location:**
- File: /Users/sandeep/Projects/scmd/commands/network/network-speed-test.yaml
- Lines: 1-50

**Current State:**
```yaml
name: network-speed-test
version: 1.0.0
description: Test download/upload speed and latency
category: network
author: scmd team
license: MIT

prompt:  # args section missing
  system: |
    ...
```

**Recommended Fix:**
```yaml
name: network-speed-test
version: 1.0.0
description: Test download/upload speed and latency
category: network
author: scmd team
license: MIT

args: []  # ADD THIS LINE

prompt:
  system: |
    ...
```

**Estimated Effort:** 5 minutes

---

### QA-005: Insufficient Examples

**Severity:** Medium
**Status:** Open
**Category:** Documentation

**Description:**
The specification requests at least 2 examples per command, but 3 commands have only 1 example. While the single examples are sufficient for these simple commands, adding a second example improves consistency and demonstrates additional usage patterns.

**Impact:**
- Minor specification non-compliance
- Reduced learning opportunities for users
- Inconsistent documentation quality

**Affected Files:**
1. /Users/sandeep/Projects/scmd/commands/network/network-speed-test.yaml
2. /Users/sandeep/Projects/scmd/commands/network/find-local-ips.yaml
3. /Users/sandeep/Projects/scmd/commands/git/git-conflict-resolver.yaml

**Current State (network-speed-test.yaml):**
```yaml
examples:
  - scmd /network-speed-test
```

**Recommended Fix:**
```yaml
examples:
  - scmd /network-speed-test
  - scmd /network-speed-test  # Could add note about running multiple times for average
```

**Current State (find-local-ips.yaml):**
```yaml
examples:
  - scmd /find-local-ips
```

**Recommended Fix:**
```yaml
examples:
  - scmd /find-local-ips
  - scmd /find-local-ips  # Could note that it shows all interfaces each time
```

**Current State (git-conflict-resolver.yaml):**
```yaml
examples:
  - scmd /git-conflict-resolver
  - scmd /git-conflict-resolver src/main.go
```
Note: This file actually has 2 examples - can be marked as PASS.

**Estimated Effort:** 15 minutes

---

### QA-006: Temperature Consistency

**Severity:** Medium
**Status:** Open
**Category:** Configuration

**Description:**
Similar operation types have slightly different temperature settings. While the differences are minor, standardizing temperatures for operation types improves consistency and predictability.

**Impact:**
- Minor behavioral inconsistencies
- Less predictable output for similar operations
- Unclear temperature selection rationale

**Examples:**
- bulk-rename.yaml: 0.3 (destructive operation)
- find-and-replace-filename.yaml: 0.2 (destructive operation)
- safe-delete.yaml: 0.2 (destructive operation)

These three commands all perform destructive file operations with preview, but have different temperatures.

**Location:**
Multiple files in file-ops/ category

**Recommended Fix:**
Standardize destructive operations to temperature: 0.2
- Affects: bulk-rename.yaml (change 0.3 → 0.2)

Review analysis operations (currently 0.4):
- find-by-extension.yaml: 0.4
- git-conflict-resolver.yaml: 0.4
- analyze-logs.yaml: 0.4

Consider if 0.4 is appropriate or should be 0.3.

**Estimated Effort:** 45 minutes (includes review and testing)

---

### QA-007: Max Tokens Variation

**Severity:** Medium
**Status:** Open
**Category:** Configuration

**Description:**
Similar operations have different max_tokens settings. While this might be intentional based on expected output, reviewing for consistency could optimize resource usage.

**Impact:**
- Potential resource over-allocation
- Potential response truncation
- Unclear token allocation strategy

**Examples:**
Find operations:
- find-large-files.yaml: 2000 tokens
- find-recent.yaml: 2000 tokens
- find-by-extension.yaml: 3000 tokens (more complex output)
- find-duplicates.yaml: 3000 tokens (more complex output)

Git operations:
- git-find-commit.yaml: 3000 tokens
- git-undo.yaml: 3000 tokens
- git-bisect-helper.yaml: 3500 tokens (educational content)
- git-interactive-rebase.yaml: 3500 tokens (step-by-step guide)

**Recommended Action:**
Review each command's expected output and verify token allocation is appropriate. Document the rationale for token limits in a separate guide.

**Estimated Effort:** 1 hour (includes documentation)

---

## Low Priority Issues

### QA-002: Missing Output Format Examples

**Severity:** Low
**Status:** Open
**Category:** Documentation

**Description:**
2 commands lack explicit output format examples in their templates, though the expected output is intuitive.

**Affected Files:**
1. network-speed-test.yaml - has format but minimal
2. find-local-ips.yaml - has format but minimal

**Recommendation:**
Add more detailed output format examples for consistency.

**Estimated Effort:** 20 minutes

---

### QA-003: Safety Notes for Low-Risk Commands

**Severity:** Low
**Status:** Open
**Category:** Best Practices

**Description:**
Some low-risk read-only commands could benefit from brief safety notes, even if just to explain they are non-destructive.

**Examples:**
- find-large-files.yaml (read-only)
- find-recent.yaml (read-only)
- monitor-system.yaml (read-only)

**Recommendation:**
Add brief note like: "This command is read-only and does not modify any files."

**Estimated Effort:** 30 minutes

---

### QA-008: Hooks Standardization

**Severity:** Low
**Status:** Open
**Category:** User Experience

**Description:**
Some similar commands use pre-execution hooks for progress indication, others don't. Consider adding hooks to more commands for consistent user feedback.

**Current Hook Usage:**
- bulk-rename.yaml: Has hook
- find-and-replace-filename.yaml: No hook (similar operation)
- change-permissions-recursive.yaml: Has hook
- sync-directories.yaml: Has hook

**Recommendation:**
Review commands that could benefit from progress indicators.

**Estimated Effort:** 1 hour

---

### QA-009: Variable Naming Consistency

**Severity:** Low
**Status:** Open
**Category:** User Experience

**Description:**
Some arguments use "path" while others use "directory" for the same concept, and some use "file" vs "files" inconsistently.

**Examples:**
- find-large-files.yaml: uses "path"
- bulk-rename.yaml: uses "path"
- sync-directories.yaml: uses "source" and "destination"
- batch-convert-images.yaml: uses "path"

**Recommendation:**
Standardize on:
- "path" for single file or directory paths
- "source"/"destination" for sync operations
- "files" for multiple file specifications

**Estimated Effort:** 1 hour (includes update and testing)

---

### QA-010: Description Length Variation

**Severity:** Low
**Status:** Open
**Category:** Documentation

**Description:**
Command descriptions vary significantly in length (5-60 words), affecting consistency in listings and help text.

**Examples:**
- Short: "Test network connectivity" (3 words)
- Long: "Analyze git blame with context, commit history, and author stats" (10 words)

**Recommendation:**
Target 8-15 words for descriptions. Be descriptive but concise.

**Estimated Effort:** 45 minutes

---

### QA-011: Example Parameter Ordering

**Severity:** Low
**Status:** Open
**Category:** Documentation

**Description:**
Some examples show arguments in a different order than they are defined in the args section.

**Recommendation:**
Match example parameter order to args definition order for clarity.

**Estimated Effort:** 30 minutes

---

### QA-012: Template Comment Style

**Severity:** Low
**Status:** Open
**Category:** Formatting

**Description:**
Some templates use bash comments (#) while others use markdown or plain text comments. Standardize for readability.

**Recommendation:**
Use bash comments for bash code blocks, markdown for explanatory text.

**Estimated Effort:** 30 minutes

---

### QA-013: Safety Warning Formatting

**Severity:** Low
**Status:** Open
**Category:** Formatting

**Description:**
Safety warnings use different formats:
- "**CRITICAL SAFETY:**"
- "**SAFETY:**"
- "**IMPORTANT SAFETY:**"
- "SAFETY CRITICAL:"

**Recommendation:**
Standardize to:
- "**CRITICAL SAFETY:**" for potentially destructive operations
- "**SAFETY:**" for general safety notes

**Estimated Effort:** 30 minutes

---

## Issue Priority Summary

**Immediate Actions (Before Release):**
- QA-004 (High): Fix manifest documentation
- QA-001 (Medium): Add args to network-speed-test

**Short-term (Next Sprint):**
- QA-005 (Medium): Add examples
- QA-006 (Medium): Review temperature settings
- QA-007 (Medium): Review token limits

**Long-term (Future Releases):**
- All Low priority issues (QA-002, QA-003, QA-008 through QA-013)

---

## Total Estimated Effort

- High Priority: 35 minutes
- Medium Priority: 2.5 hours
- Low Priority: 4.5 hours
- **Total: ~7.5 hours**

---

**Status Legend:**
- Open: Not yet addressed
- In Progress: Being worked on
- Fixed: Completed and verified
- Won't Fix: Issue accepted as-is with documented reason

**Last Updated:** 2026-01-09
