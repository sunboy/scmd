# Phase 1: Fresh Installation & Onboarding - Findings

## Test Date
January 6, 2026

## Tester Profile
Brand new user, excited to try scmd, no prior experience with the tool.

## Installation Experience

### ✅ What Worked Well

1. **Binary Build**: The `go build` command worked perfectly
   - Clean build, no errors
   - Binary size: 14.5MB (reasonable)

2. **Model Management CLI**: Excellent UX
   ```bash
   ./scmd models list
   ```
   - Clear table output with size, status, and descriptions
   - Shows which model is ready (qwen3-4b ✓)
   - Helpful descriptions ("Fast, efficient, tool calling support")
   - Clear instructions on how to download models

3. **Backend Discovery**: Clear and informative
   ```bash
   ./scmd backends
   ```
   - Shows available backends with checkmarks
   - Lists environment variables needed for cloud providers
   - Good visual feedback (✓ and ✗)

4. **Slash Commands Listing**: Well organized
   ```bash
   ./scmd slash list
   ```
   - Clean table with command, aliases, what it runs, and description
   - Shows multiple aliases (e.g., `/explain` → e, exp)
   - Usage instructions at bottom

5. **First Successful Command**: When it worked, it was impressive!
   - The qwen3-4b model gave clear, well-formatted explanations
   - Markdown formatting in output
   - Breakdown of code with examples

### ❌ Critical Issues

1. **llama-server Management** - BIGGEST ISSUE
   - **Problem**: User must manually start llama-server
   - **Impact**: Commands fail with unhelpful "Error" message
   - **What happened**:
     ```bash
     echo "code" | ./scmd /explain
     Error
     ```
   - **Debug showed**: `Post "http://127.0.0.1:8089/completion": EOF`
   - **User expectation**: Tool should "just work" offline as advertised
   - **Reality**: Need to manually run:
     ```bash
     llama-server -m ~/.scmd/models/qwen3-4b-Q4_K_M.gguf -c 4096 --port 8089 -ngl 99
     ```

2. **GPU Out of Memory Crashes**
   - **Problem**: llama-server crashed multiple times with OOM errors
   - **Error**: `Insufficient Memory (00000008:kIOGPUCommandBufferCallbackErrorOutOfMemory)`
   - **Context**: Apple M1 with 8GB unified memory
   - **Issue**: Multiple llama-server instances or high context sizes (-c 8192) exhaust GPU memory
   - **Impact**: Tool completely stops working, requires manual kill and restart
   - **User frustration**: Very high - happened repeatedly during testing

3. **Error Messages Are Not Helpful**
   - **Problem**: Failures show generic "Error" with no details
   - **Example**: `./scmd /review code.go` → `Error`
   - **Need**: User-friendly errors like:
     ```
     Error: Could not connect to llama-server

     Is llama-server running? Start it with:
     llama-server -m ~/.scmd/models/qwen3-4b-Q4_K_M.gguf --port 8089

     Or use a cloud provider:
     export OPENAI_API_KEY=your-key
     ./scmd -b openai /review code.go
     ```

4. **Slash Command Syntax Confusion**
   - **Problem**: Inconsistent syntax between direct vs backend flag usage
   - **What works**: `./scmd /explain code.go`
   - **What doesn't**: `./scmd -b mock /explain` → `Error: unknown command "/explain"`
   - **What does work**: `./scmd -b mock explain`
   - **User confusion**: When do I use `/` and when don't I?

5. **Performance: CPU Mode is VERY Slow**
   - **Problem**: Without GPU acceleration (ngl=0), inference is extremely slow
   - **Test**: Simple "Hello world" explanation took 30+ seconds (still running when I stopped waiting)
   - **README claims**: "~5 tokens/sec on CPU" for qwen3-4b
   - **Reality**: Much slower in practice, potentially unusable
   - **User expectation**: "Fast, efficient" - doesn't match reality on CPU

### ⚠️ Moderate Issues

1. **No Automatic llama-server Lifecycle**
   - Can't tell if llama-server is running
   - No `scmd server start/stop/status` commands
   - Multiple instances can run simultaneously causing conflicts

2. **No Health Checks**
   - Tool doesn't verify backend is responsive before attempting inference
   - Would be helpful to have: `./scmd doctor` to check configuration

3. **Context Size Not Configurable Per-Command**
   - llama-server started with fixed context size
   - No way to say "this command needs more context"
   - Impacts memory usage and crashes

4. **Model Auto-Download** (Not tested, but claimed)
   - README says model auto-downloads on first use
   - Couldn't verify because model was already downloaded
   - Unclear what happens on slow connection or partial download

### 📊 Comparison: Expectations vs Reality

| Feature | Documentation Says | Reality |
|---------|-------------------|----------|
| "Works offline by default" | ✅ Should just work | ❌ Requires manual llama-server setup |
| "No API keys or setup required" | ✅ True | ⚠️ True but setup still needed |
| "Fast, efficient inference" | ⚠️ ~5 tok/sec CPU | ❌ Much slower, potentially unusable |
| "GPU acceleration" | ✅ Metal on macOS | ❌ Crashes with OOM errors frequently |
| Error handling | N/A | ❌ Unhelpful generic "Error" messages |

## Recommendations for Developers

### P0 (Critical - Blocks Usage)

1. **Auto-manage llama-server process**
   - Start llama-server automatically on first command
   - Check if already running on correct port
   - Handle graceful shutdown
   - Or: provide `scmd server start/stop/status` commands

2. **Improve error messages**
   - Detect specific failure modes (server not running, OOM, network issues)
   - Provide actionable suggestions
   - Show how to switch to alternative backends

3. **Prevent GPU OOM crashes**
   - Detect available memory before starting
   - Auto-tune context size based on available RAM
   - Prevent multiple llama-server instances on same port
   - Fallback to CPU if GPU fails

### P1 (High - Significantly Impacts UX)

4. **Add `scmd doctor` command**
   - Check if llama-server is running
   - Verify model files exist and are valid
   - Test backend connectivity
   - Check system resources (RAM, GPU)
   - Provide configuration recommendations

5. **Unify slash command syntax**
   - Make `/command` work consistently everywhere
   - Or: make `command` work everywhere
   - Currently confusing when to use which

6. **Performance expectations**
   - Update documentation with realistic benchmarks
   - Warn users that CPU mode is slow
   - Recommend GPU usage more strongly
   - Consider smaller default model if targeting CPU users

### P2 (Nice to Have)

7. **Progress indicators**
   - Show "Loading model..." when starting llama-server
   - Show "Generating..." with spinner during inference
   - Show tokens/sec during generation

8. **Resource monitoring**
   - Show memory usage
   - Warn if approaching OOM
   - Auto-switch to smaller context or CPU if needed

## Overall Impression

**As a brand new excited user**: The first 15 minutes were frustrating and confusing. The tool has great potential, but the onboarding experience needs significant work. The biggest issue is that "works offline by default" creates an expectation that it just works, but in reality, there's hidden complexity around managing llama-server.

**Rating**: 4/10 for onboarding experience
- Would be 8/10 if llama-server was auto-managed
- Would be 9/10 with better error messages
- Would be 10/10 with `scmd doctor` command

The core concept is excellent, and when it works, it's impressive. But the technical hurdles prevent it from being the "just works" experience that's promised.
