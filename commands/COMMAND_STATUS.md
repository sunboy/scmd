# SCMD Command Repository - Individual Command Status

**Test Date:** 2026-01-09
**Total Commands:** 38

---

## Status Legend

- ✓ PASS - All tests passed
- ~ PASS* - Passed with minor recommendations
- ✗ FAIL - Critical issues found

---

## File Operations (15 commands)

| # | Command | Status | Structure | Content | Args | Prompt | Model | Issues |
|---|---------|--------|-----------|---------|------|--------|-------|--------|
| 1 | archive-old-files | ✓ PASS | ✓ | ✓ | ✓ | ✓ | ✓ | None |
| 2 | batch-convert-images | ✓ PASS | ✓ | ✓ | ✓ | ✓ | ✓ | None |
| 3 | bulk-rename | ~ PASS* | ✓ | ✓ | ✓ | ✓ | ✓ | QA-006 (temp) |
| 4 | change-permissions-recursive | ✓ PASS | ✓ | ✓ | ✓ | ✓ | ✓ | None |
| 5 | check-file-encoding | ✓ PASS | ✓ | ✓ | ✓ | ✓ | ✓ | None |
| 6 | disk-usage | ✓ PASS | ✓ | ✓ | ✓ | ✓ | ✓ | None |
| 7 | find-and-replace-filename | ✓ PASS | ✓ | ✓ | ✓ | ✓ | ✓ | None |
| 8 | find-broken-symlinks | ✓ PASS | ✓ | ✓ | ✓ | ✓ | ✓ | None |
| 9 | find-by-extension | ✓ PASS | ✓ | ✓ | ✓ | ✓ | ✓ | None |
| 10 | find-duplicates | ✓ PASS | ✓ | ✓ | ✓ | ✓ | ✓ | None |
| 11 | find-empty | ✓ PASS | ✓ | ✓ | ✓ | ✓ | ✓ | None |
| 12 | find-large-files | ✓ PASS | ✓ | ✓ | ✓ | ✓ | ✓ | None |
| 13 | find-recent | ✓ PASS | ✓ | ✓ | ✓ | ✓ | ✓ | None |
| 14 | safe-delete | ✓ PASS | ✓ | ✓ | ✓ | ✓ | ✓ | None |
| 15 | sync-directories | ✓ PASS | ✓ | ✓ | ✓ | ✓ | ✓ | None |

**Category Summary:** 14 PASS, 1 PASS* (93% clean)

---

## Git Operations (8 commands)

| # | Command | Status | Structure | Content | Args | Prompt | Model | Issues |
|---|---------|--------|-----------|---------|------|--------|-------|--------|
| 1 | git-bisect-helper | ✓ PASS | ✓ | ✓ | ✓ | ✓ | ✓ | None |
| 2 | git-blame-analysis | ✓ PASS | ✓ | ✓ | ✓ | ✓ | ✓ | None |
| 3 | git-cleanup-branches | ✓ PASS | ✓ | ✓ | ✓ | ✓ | ✓ | None |
| 4 | git-conflict-resolver | ✓ PASS | ✓ | ✓ | ✓ | ✓ | ✓ | None |
| 5 | git-find-commit | ✓ PASS | ✓ | ✓ | ✓ | ✓ | ✓ | None |
| 6 | git-interactive-rebase | ✓ PASS | ✓ | ✓ | ✓ | ✓ | ✓ | None |
| 7 | git-stash-manager | ✓ PASS | ✓ | ✓ | ✓ | ✓ | ✓ | None |
| 8 | git-undo | ✓ PASS | ✓ | ✓ | ✓ | ✓ | ✓ | None |

**Category Summary:** 8 PASS (100% clean) - Excellent!

---

## System Administration (8 commands)

| # | Command | Status | Structure | Content | Args | Prompt | Model | Issues |
|---|---------|--------|-----------|---------|------|--------|-------|--------|
| 1 | analyze-logs | ✓ PASS | ✓ | ✓ | ✓ | ✓ | ✓ | None |
| 2 | check-service-status | ✓ PASS | ✓ | ✓ | ✓ | ✓ | ✓ | None |
| 3 | check-startup-programs | ✓ PASS | ✓ | ✓ | ✓ | ✓ | ✓ | None |
| 4 | disk-cleanup | ✓ PASS | ✓ | ✓ | ✓ | ✓ | ✓ | None |
| 5 | find-port-user | ✓ PASS | ✓ | ✓ | ✓ | ✓ | ✓ | None |
| 6 | monitor-system | ✓ PASS | ✓ | ✓ | ✓ | ✓ | ✓ | None |
| 7 | process-tree | ✓ PASS | ✓ | ✓ | ✓ | ✓ | ✓ | None |
| 8 | system-health-check | ✓ PASS | ✓ | ✓ | ✓ | ✓ | ✓ | None |

**Category Summary:** 8 PASS (100% clean) - Excellent!

---

## Network Tools (7 commands)

| # | Command | Status | Structure | Content | Args | Prompt | Model | Issues |
|---|---------|--------|-----------|---------|------|--------|-------|--------|
| 1 | dns-lookup | ✓ PASS | ✓ | ✓ | ✓ | ✓ | ✓ | None |
| 2 | find-local-ips | ~ PASS* | ✓ | ✓ | ✓ | ✓ | ✓ | QA-005 (examples) |
| 3 | http-request-debug | ✓ PASS | ✓ | ✓ | ✓ | ✓ | ✓ | None |
| 4 | network-bandwidth-monitor | ✓ PASS | ✓ | ✓ | ✓ | ✓ | ✓ | None |
| 5 | network-speed-test | ~ PASS* | ~ | ✓ | ~ | ✓ | ✓ | QA-001 (args), QA-005 (examples) |
| 6 | port-scan | ✓ PASS | ✓ | ✓ | ✓ | ✓ | ✓ | None |
| 7 | test-connectivity | ✓ PASS | ✓ | ✓ | ✓ | ✓ | ✓ | None |

**Category Summary:** 5 PASS, 2 PASS* (71% clean)

---

## Overall Statistics

### By Status
- ✓ PASS: 35 commands (92%)
- ~ PASS*: 3 commands (8%)
- ✗ FAIL: 0 commands (0%)

### By Test Category
| Test | Passed | Total | Percentage |
|------|--------|-------|------------|
| Structure | 37 | 38 | 97% |
| Content | 38 | 38 | 100% |
| Args | 37 | 38 | 97% |
| Prompt | 38 | 38 | 100% |
| Model | 38 | 38 | 100% |

---

## Detailed Results

### Commands with Perfect Score (35)

All tests passed with no issues:

**File Operations (14):**
1. archive-old-files
2. batch-convert-images
3. change-permissions-recursive
4. check-file-encoding
5. disk-usage
6. find-and-replace-filename
7. find-broken-symlinks
8. find-by-extension
9. find-duplicates
10. find-empty
11. find-large-files
12. find-recent
13. safe-delete
14. sync-directories

**Git Operations (8):**
1. git-bisect-helper
2. git-blame-analysis
3. git-cleanup-branches
4. git-conflict-resolver
5. git-find-commit
6. git-interactive-rebase
7. git-stash-manager
8. git-undo

**System Administration (8):**
1. analyze-logs
2. check-service-status
3. check-startup-programs
4. disk-cleanup
5. find-port-user
6. monitor-system
7. process-tree
8. system-health-check

**Network Tools (5):**
1. dns-lookup
2. http-request-debug
3. network-bandwidth-monitor
4. port-scan
5. test-connectivity

---

### Commands with Minor Issues (3)

#### 1. bulk-rename.yaml
**Status:** PASS*
**Category:** file-ops
**Issues:**
- QA-006: Temperature is 0.3, recommend 0.2 for destructive operations
**Impact:** Low - functional but inconsistent with similar commands
**Priority:** Medium

#### 2. find-local-ips.yaml
**Status:** PASS*
**Category:** network
**Issues:**
- QA-005: Only 1 example, recommend adding second example
**Impact:** Low - documentation completeness
**Priority:** Medium

#### 3. network-speed-test.yaml
**Status:** PASS*
**Category:** network
**Issues:**
- QA-001: Missing explicit `args: []` declaration
- QA-005: Only 1 example, recommend adding second example
**Impact:** Low - structural consistency
**Priority:** Medium

---

## Category Performance Ranking

1. **Git Operations:** 100% clean (8/8) 🏆
2. **System Administration:** 100% clean (8/8) 🏆
3. **File Operations:** 93% clean (14/15)
4. **Network Tools:** 71% clean (5/7)

---

## Quality Highlights

### Best Overall Commands (Perfect Implementation)

These commands exemplify best practices:

1. **git-bisect-helper** - Educational, comprehensive, excellent step-by-step guide
2. **safe-delete** - Outstanding safety focus, clear warnings
3. **system-health-check** - Comprehensive diagnostics, excellent output format
4. **git-interactive-rebase** - Perfect balance of guidance and safety
5. **bulk-rename** - Excellent preview-before-execute pattern
6. **sync-directories** - Great dry-run implementation
7. **git-undo** - Superb safety warnings and recovery instructions

### Best Safety Implementation

Commands with exemplary safety practices:

1. **safe-delete** - Never uses rm -rf, moves to trash
2. **git-undo** - Multiple confirmation levels, explains consequences
3. **change-permissions-recursive** - Preview, warnings, system path protection
4. **disk-cleanup** - Dry-run by default, clear space calculations
5. **git-interactive-rebase** - Checks for uncommitted changes, pushed commits

### Best Documentation

Commands with outstanding documentation:

1. **git-bisect-helper** - Explains process, provides automation tips
2. **git-conflict-resolver** - Clear conflict explanation, resolution strategies
3. **system-health-check** - Comprehensive health assessment, actionable recommendations
4. **archive-old-files** - Complete workflow with verification and restore instructions

---

## Test Execution Details

### Structure Tests (YAML Validation)
- Parser: Python yaml.safe_load() equivalent (manual validation)
- Required fields: name, version, description, category, author, license, args, prompt, model, examples
- Prompt structure: system, template
- Model structure: temperature, max_tokens
- **Result:** 97% (37/38) - 1 minor issue

### Content Tests
- Name-to-filename matching
- Version consistency
- Description quality
- Category consistency
- Metadata accuracy
- **Result:** 100% (38/38)

### Argument Tests
- Structure validation (name, description, required fields)
- Default value presence for non-required args
- Type consistency
- **Result:** 97% (37/38) - 1 missing explicit declaration

### Prompt Tests
- System prompt clarity
- Template syntax
- Variable usage
- Safety guidelines
- Example command formatting
- **Result:** 100% (38/38)

### Model Configuration Tests
- Temperature range (0-1)
- Temperature appropriateness for operation type
- Max tokens reasonableness
- **Result:** 100% (38/38)

---

## Recommendations by Command

### Immediate Actions

**network-speed-test.yaml**
- Add: `args: []` after license field
- Add: Second example
- Time: 10 minutes

**bulk-rename.yaml**
- Review: Temperature 0.3 → 0.2 for consistency
- Time: 5 minutes

**find-local-ips.yaml**
- Add: Second example
- Time: 5 minutes

---

## Test Coverage Matrix

| Command | YAML | Structure | Content | Args | Prompt | Model | Examples | Hooks | Safety |
|---------|------|-----------|---------|------|--------|-------|----------|-------|--------|
| archive-old-files | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓✓✓ | ✗ | ✓ |
| batch-convert-images | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓✓✓ | ✓ | ✓ |
| bulk-rename | ✓ | ✓ | ✓ | ✓ | ✓ | ~ | ✓✓✓✓ | ✓ | ✓ |
| change-permissions-recursive | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓✓✓ | ✓ | ✓✓ |
| check-file-encoding | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓✓✓ | ✗ | ✓ |
| disk-usage | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓✓✓ | ✗ | ✓ |
| find-and-replace-filename | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓✓ | ✗ | ✓✓ |
| find-broken-symlinks | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓✓✓✓ | ✗ | ✓ |
| find-by-extension | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓✓✓ | ✗ | ✗ |
| find-duplicates | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓✓✓ | ✗ | ✓ |
| find-empty | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓✓✓ | ✗ | ✓ |
| find-large-files | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓✓✓ | ✗ | ✓ |
| find-recent | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓✓✓ | ✗ | ✗ |
| safe-delete | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓✓✓ | ✗ | ✓✓✓ |
| sync-directories | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓✓✓ | ✓ | ✓✓ |
| git-bisect-helper | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓✓✓ | ✗ | ✓ |
| git-blame-analysis | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓✓✓ | ✗ | ✗ |
| git-cleanup-branches | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓✓✓✓ | ✓ | ✓✓ |
| git-conflict-resolver | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓✓ | ✗ | ✓ |
| git-find-commit | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓✓✓✓✓ | ✗ | ✗ |
| git-interactive-rebase | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓✓✓ | ✓ | ✓✓✓ |
| git-stash-manager | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓✓✓✓✓ | ✗ | ✓ |
| git-undo | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓✓✓✓ | ✓ | ✓✓✓ |
| analyze-logs | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓✓✓ | ✗ | ✗ |
| check-service-status | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓✓✓ | ✗ | ✗ |
| check-startup-programs | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓✓✓ | ✗ | ✓ |
| disk-cleanup | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓✓✓ | ✗ | ✓✓ |
| find-port-user | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓✓✓ | ✗ | ✓ |
| monitor-system | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓✓✓ | ✗ | ✗ |
| process-tree | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓✓✓ | ✗ | ✗ |
| system-health-check | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓✓ | ✗ | ✗ |
| dns-lookup | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓✓ | ✗ | ✗ |
| find-local-ips | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✗ | ✗ |
| http-request-debug | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓✓ | ✗ | ✗ |
| network-bandwidth-monitor | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓✓ | ✗ | ✗ |
| network-speed-test | ✓ | ✓ | ✓ | ~ | ✓ | ✓ | ✓ | ✗ | ✗ |
| port-scan | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓✓✓ | ✗ | ✓ |
| test-connectivity | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓✓ | ✗ | ✗ |

Legend:
- ✓ = Pass
- ~ = Pass with minor issue
- ✗ = Not applicable / Not present
- ✓✓ = Excellent
- ✓✓✓ = Outstanding

---

**Generated:** 2026-01-09
**Report Version:** 1.0
