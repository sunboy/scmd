# Phase 1: New User Experience Findings (Round 3)
## Fresh Install Testing - January 6, 2026

**Test Scenario**: Brand new excited user testing scmd with the new repository-first architecture

**Duration**: ~30 minutes
**Commands Tested**: 5 slash commands listed, 7 attempted

---

## 🎯 Executive Summary

**Critical Discovery**: **Severe disconnect between what's advertised and what actually works**

- ✅ Health check: **EXCELLENT** (10/10)
- ✅ Auto-start: **EXCELLENT** (10/10)
- ✅ Built-in commands: **WORK GREAT** (9/10)
- ❌ Repository system: **BROKEN/MISLEADING** (2/10)
- ❌ Command availability: **60% PHANTOM COMMANDS** (3/10)

**Overall Phase 1 Rating**: 6.5/10 (same as original, but different reasons)

---

## ✅ What Worked Perfectly

### 1. Health Check (10/10) 🎉

```bash
$ ./scmd doctor
🏥 scmd Health Check
✅ scmd binary: /path/to/scmd (13.9 MB)
✅ Downloaded models: 1 model(s), 2.3 GB total
✅ llama-server binary: /opt/homebrew/bin/llama-server
✅ llama-server status: Running on port 8089
✅ System memory: 8.0 GB total
✅ All checks passed! scmd is ready to use.
```

**Verdict**: Perfect. Exactly what a new user needs.

---

### 2. Auto-Start (10/10) 🎉

```bash
$ echo "print('Hello, World!')" | ./scmd /explain
⏳ Starting llama-server...
✅ GPU acceleration enabled (Apple Silicon (Metal))
✅ llama-server ready
[detailed explanation in 26 seconds]
```

**Verdict**: Flawless. From zero to working in under 30 seconds.

---

### 3. Built-in Commands Performance (9/10) 🎉

**`/explain` Test**:
- Input: `print('Hello, World!')`
- Time: 26 seconds
- Output: Comprehensive, markdown-formatted explanation
- Quality: Excellent
- **Verdict**: Fast, thorough, exactly what's needed

**`/review` Test**:
- Input: `def divide(a, b): return a / b`
- Time: ~30 seconds
- Output: Found division by zero bug, suggested type hints, documentation
- Quality: Production-ready code review
- **Verdict**: Impressive! Would actually use this daily.

---

## ❌ What's Broken

### 1. Phantom Commands (CRITICAL BUG) 🚨

**Issue**: `scmd slash list` shows 5 commands, but 3 don't work

```bash
$ ./scmd slash list

COMMAND     ALIASES       RUNS           DESCRIPTION
/explain    e, exp        explain        Explain code or text
/review     r, rev        review         Review code
/commit     gc, gitc      git-commit     Generate git commit message  ❌
/summarize  s, sum, tldr  summarize      Summarize text               ❌
/fix        f, err        explain-error  Explain and fix errors       ❌
```

**Testing**:
```bash
$ git diff --staged | ./scmd /commit
❌ Command 'commit' not found

$ echo "Long article..." | ./scmd /summarize
❌ Command 'summarize' not found

$ echo "Error: database timeout" | ./scmd /fix
❌ Command 'fix' not found
```

**Root Cause**:
- `~/.scmd/slash.yaml` has stale entries
- Maps to commands `git-commit`, `summarize`, `explain-error`
- But these YAML-based commands don't exist
- No `~/.scmd/commands/` directory
- Repository installation fails with 404

**Impact**: **SEVERE**
- 60% command failure rate (3/5)
- New users will be frustrated immediately
- Trust in the tool is broken
- Confusing error messages (suggests they're findable, but they're not)

---

### 2. Repository System Confusion (P0) 🚨

**README Claims**:
> "Only the `explain` command is built-in. Others install from repositories."

**Reality**:
- 2 commands are built-in: `explain`, `review`
- 3 commands are listed but missing: `commit`, `summarize`, `fix`
- Repository system returns 404 errors
- No working repository available
- Documentation doesn't match implementation

**Testing**:

```bash
$ ./scmd repo list
official    https://github.com/scmd/commands/raw/main

$ ./scmd repo search review
No commands found.

$ ./scmd repo show official/review
Error: fetch manifest: status 404

$ ./scmd repo install official/commit
Error: fetch manifest: status 404

$ ./scmd repo add local file://$(pwd)/testdata/sample-repo
Error: unsupported URL scheme 'file' (only http and https allowed)
```

**Confusion Matrix**:

| What User Sees | What User Expects | Reality | User Emotion |
|----------------|-------------------|---------|--------------|
| `scmd slash list` shows `/commit` | Command works | Doesn't exist | Frustrated |
| README says "install from repos" | `repo install` works | Returns 404 | Confused |
| `repo list` shows "official" | Repo has commands | Repo doesn't exist | Betrayed |
| `/review` works | It was installed from repo | Actually built-in | Uncertain |

---

### 3. Architecture Unclear (P1)

**Question**: What commands are actually available?

**Answer After 30 min of testing**:
- 2 Go built-in commands work: `/explain`, `/review`
- 3 YAML commands are listed but missing
- Repository system exists but is non-functional
- YAML command files exist in `testdata/sample-repo/` but can't be installed
- No clear way to add custom commands

**New User Mental Model**:
- ❓ "Is scmd repository-first or builtin-first?"
- ❓ "Which commands can I trust?"
- ❓ "How do I install commands?"
- ❓ "Is the official repo working?"

---

## 🔍 Technical Details

### Slash Command Registry

**File**: `~/.scmd/slash.yaml`

```yaml
commands:
  - name: explain      # ✅ WORKS (Go builtin)
    command: explain
  - name: review       # ✅ WORKS (Go builtin)
    command: review
  - name: commit       # ❌ PHANTOM (YAML missing)
    command: git-commit
  - name: summarize    # ❌ PHANTOM (YAML missing)
    command: summarize
  - name: fix          # ❌ PHANTOM (YAML missing)
    command: explain-error
```

### Builtin Commands

**Files** in `internal/command/builtin/`:
- ✅ `explain.go` - Works
- ✅ `review.go` - Works
- ⚠️ `help.go` - Utility
- ⚠️ `config.go` - Utility
- ⚠️ `kill_process.go` - Utility

**Registry** (`register.go`):
```go
commands := []command.Command{
    helpCmd,
    NewExplainCommand(),
    NewReviewCommand(),
    NewConfigCommand(),
    &KillProcessCmd{},
}
```

### Missing YAML Commands

**Found in** `testdata/sample-repo/commands/`:
- `git-commit.yaml` - Well-defined, looks production-ready
- `summarize.yaml` - Exists
- `explain-error.yaml` - Exists
- `code-review.yaml` - Exists (duplicate of builtin?)

**But**:
- Not installed in `~/.scmd/commands/`
- Not accessible from scmd
- Repository installation returns 404
- No local file:// installation support

---

## 📊 Comparison: Round 2 vs Round 3

| Aspect | Round 2 (Retest) | Round 3 (Fresh User) | Change |
|--------|------------------|----------------------|--------|
| **Auto-start** | ✅ Works (10/10) | ✅ Works (10/10) | **No change** |
| **Health check** | ✅ Excellent (10/10) | ✅ Excellent (10/10) | **No change** |
| **Builtin commands** | ✅ Work well | ✅ Work well | **No change** |
| **Command count** | Assumed 5 work | Only 2 work | **3 missing!** |
| **Repository** | Not tested | 404 errors | **NEW ISSUE** |
| **Documentation** | Accurate | Misleading | **NEW ISSUE** |
| **User trust** | High | Low | **REGRESSION** |

---

## 🎯 Impact on User Journey

### Excited New User Timeline

**0-1 min**: "Wow, this looks amazing!"
- Reads README promising offline slash commands
- ⭐⭐⭐⭐⭐

**1-2 min**: "Let me install it"
- Builds scmd, runs first command
- Auto-start works perfectly!
- ⭐⭐⭐⭐⭐

**2-5 min**: "Let me try more commands"
- `/explain` works → Happy!
- `/review` works → Impressed!
- ⭐⭐⭐⭐⭐

**5-8 min**: "Let me generate a commit message"
- `scmd slash list` shows `/commit`
- Tries it → "Command not found"
- ⭐⭐⭐ (confused)

**8-15 min**: "Maybe I need to install it?"
- Tries `scmd repo search commit` → "No commands found"
- Tries `scmd repo install official/commit` → 404 error
- ⭐⭐ (frustrated)

**15-25 min**: "Let me read the docs"
- README says "only explain is built-in"
- But `/review` works and isn't from a repo
- README says "install from official repo"
- But repo returns 404
- ⭐ (betrayed)

**25+ min**: "Is this tool ready?"
- 60% of listed commands don't work
- Repository system is broken
- Documentation doesn't match reality
- ⭐ (considering abandoning)

---

## 🚨 P0 Issues for New Users

### 1. Remove Phantom Commands from slash.yaml (URGENT)

**Current**:
```yaml
- name: commit
  command: git-commit  # Doesn't exist
```

**Fix**:
```yaml
# Remove these entries until commands actually exist:
# - commit, summarize, fix
```

**Impact**: Prevents 60% failure rate

---

### 2. Fix or Remove Repository References (URGENT)

**Option A**: Fix the official repo
- Deploy commands to https://github.com/scmd/commands/raw/main
- Make git-commit, summarize, explain-error available
- Test installation

**Option B**: Update README
- Remove references to repository-first architecture
- Clarify that only 2 commands are currently available
- Provide roadmap for repository system

**Impact**: Prevents documentation betrayal

---

### 3. Add Command Status to `slash list` (HIGH)

**Current**:
```
COMMAND     DESCRIPTION
/commit     Generate git commit message
```

**Proposed**:
```
COMMAND     STATUS      DESCRIPTION
/explain    ready       Explain code or text
/review     ready       Review code
/commit     missing     Generate git commit message (install: scmd repo install official/commit)
```

**Impact**: Transparency about what works

---

### 4. Clean `scmd doctor` to Validate Commands (HIGH)

**Add to health check**:
```
🏥 scmd Health Check
✅ scmd binary: ready
✅ llama-server: running
✅ Slash commands: 2/5 available
   ✅ /explain - ready
   ✅ /review - ready
   ❌ /commit - not installed (run: scmd repo install official/commit)
   ❌ /summarize - not installed
   ❌ /fix - not installed
```

**Impact**: User knows what to expect

---

## 💡 Recommendations

### Immediate (This Week)

1. **Clean up slash.yaml** - Remove phantom command entries
2. **Update README** - Match documentation to reality
3. **Fix repository OR document it's not ready** - Don't promise what doesn't exist
4. **Add validation to `scmd doctor`** - Check command availability

### Short-term (1-2 Weeks)

5. **Deploy working official repository** - Make repo install actually work
6. **Add commands properly** - Either as builtins or via working repo
7. **Support local file:// repos** - For development and testing
8. **Improve error messages** - When command not found, explain why

### Medium-term (1 Month)

9. **First-run experience** - Detect missing commands, offer to install
10. **Command validation** - Prevent phantom entries in slash.yaml
11. **Repository health checks** - Validate repo URLs before adding
12. **Better onboarding docs** - Step-by-step guide matching actual behavior

---

## 📈 Ratings Breakdown

| Category | Rating | Reasoning |
|----------|--------|-----------|
| **Auto-start** | 10/10 | Perfect, no notes |
| **Health check** | 10/10 | Comprehensive and helpful |
| **Builtin commands** | 9/10 | Work great, good quality |
| **Command availability** | 3/10 | 60% failure rate |
| **Repository system** | 2/10 | Returns 404, doesn't work |
| **Documentation** | 4/10 | Misleading, doesn't match reality |
| **Error messages** | 7/10 | Good suggestions, but wrong context |
| **Overall Onboarding** | **6.5/10** | Great start, broken promises |

---

## 🎓 Key Insights

### What This Means for scmd

1. **Infrastructure is excellent** - Auto-start, health checks, server management all work perfectly
2. **Core commands are great** - `/explain` and `/review` are production-quality
3. **Architecture is unclear** - Is it builtin-first or repo-first? Pick one.
4. **Promises exceed delivery** - Don't list commands that don't work
5. **Repository system isn't ready** - Either finish it or don't advertise it

### Comparison to Previous Rounds

**Round 1 (Original)**: Infrastructure broken, commands worked
**Round 2 (Retest)**: Infrastructure fixed, assumed commands worked
**Round 3 (Fresh user)**: Infrastructure excellent, **command ecosystem broken**

**Rating trajectory**: 6.5 → 8.5 → 6.5 (for different reasons)

---

## 🔮 Next Steps for Evaluation

Since I can only use 2 working commands (`/explain`, `/review`), my testing plan adjusts:

### Phase 1.3: Test Working Commands
- ✅ Test `/explain` with various code snippets
- ✅ Test `/review` with real code
- 🔲 Evaluate quality and usefulness

### Phase 2: Test Custom Commands
- 🔲 Figure out how to manually install YAML commands
- 🔲 Test my 16 pre-created commands if possible
- 🔲 Or just document that custom commands don't work

### Phase 3: Real Workflows
- 🔲 Use `/explain` for learning complex unix commands
- 🔲 Use `/review` for actual code review tasks
- 🔲 Measure time savings vs traditional methods

### Phase 4: Create New Commands
- 🔲 Research more hard unix tasks
- 🔲 Create 5 new YAML command specs
- 🔲 Document how they SHOULD be installed (even if can't test)

### Phase 5: Recommendations
- 🔲 Feature requests based on limitations discovered
- 🔲 Repository system fixes needed
- 🔲 Command ecosystem improvements

### Phase 6: Final Report
- 🔲 NEW_USER_EXPERIENCE_REPORT.md with all findings
- 🔲 Honest assessment of current vs promised capabilities
- 🔲 Roadmap for reaching the vision

---

## 💬 Bottom Line

**For a brand new excited user:**

✅ **LOVE**:
- Health check is perfect
- Auto-start is magical
- Built-in commands are excellent
- Infrastructure is production-ready

❌ **HATE**:
- 60% of listed commands don't work
- Repository system is broken
- Documentation is misleading
- Can't tell what's real vs what's coming soon

**Overall**: "I'm impressed by what works, but frustrated by broken promises. Fix the phantom commands and I'll be back."

**Rating**: 6.5/10 (same as original, but now it's command ecosystem not infrastructure)

---

**Next**: Continue to Phase 1.3 - Deep testing of the 2 working commands with real workflows.
