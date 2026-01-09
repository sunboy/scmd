# Stability-First Architecture

**Core Principle**: Users should never need to manually manage the LLM server. The system must be self-healing and provide clear feedback when issues occur.

## Design Tenets

### 1. Zero Manual Intervention
- Server automatically starts, stops, and restarts as needed
- No `pkill` commands required from users
- System detects and recovers from crashes automatically

### 2. Clear Feedback Loop
- Every error message includes actionable solutions
- Users know exactly what's happening and what to do
- No silent failures or confusing states

### 3. Intelligent Recovery
- Detect server health issues (crashes, context mismatches, OOM)
- Automatically retry with safer configurations (CPU-only, smaller context)
- Provide feedback about degraded performance modes

### 4. Graceful Degradation
- When GPU fails → Fall back to CPU mode
- When context too large → Suggest splitting input or cloud backend
- When server unreachable → Auto-restart with clear messaging

## Implementation

### Health Check System
```
┌─────────────────────────────────────────┐
│ User runs: scmd /explain                │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│ Pre-inference Health Check              │
│ • Is server running?                    │
│ • Does context size match expected?     │
│ • Is server responsive?                 │
└──────────────┬──────────────────────────┘
               │
        ┌──────┴──────┐
        │ Healthy?    │
        └──────┬──────┘
               │
        ┌──────┴──────────┐
        │ YES        NO   │
        │                 │
        ▼                 ▼
   ┌─────────┐    ┌──────────────┐
   │ Execute │    │ Auto-Restart │
   │ Request │    │ with Retry   │
   └─────────┘    └──────┬───────┘
                         │
                         ▼
                  ┌──────────────┐
                  │ Show Clear   │
                  │ Feedback     │
                  └──────────────┘
```

### Error Handling Hierarchy

1. **Detection**: Parse error to identify root cause
   - Context size exceeded
   - Out of memory
   - Server crashed
   - Connection failed

2. **Auto-Recovery**: Try to fix automatically
   - Restart server
   - Reduce context size
   - Switch to CPU mode
   - Use different model

3. **User Feedback**: When auto-recovery fails
   - Show what went wrong (in user terms)
   - Show what was tried (transparency)
   - Show what user can do (actionable steps)
   - Provide one-command fixes when possible

### Example Error Message Flow

**BAD** (Current state in some areas):
```
Error: inference failed
Try: scmd doctor
```

**GOOD** (Target state):
```
❌ Input too large for available GPU memory

What happened:
  Your input (5502 tokens) exceeds GPU memory capacity (4096 tokens)
  Metal GPU cannot allocate enough VRAM for the full context

What I tried:
  ✓ Started llama-server with 32K context
  ✗ GPU allocated only 4K (memory limitation)

What you can do:
  1. Use CPU mode (slower, supports full 32K):
     export SCMD_CPU_ONLY=1
     scmd /explain <your-input>

  2. Split your input into smaller files

  3. Use cloud backend (fastest):
     export OPENAI_API_KEY=your-key
     scmd -b openai /explain <your-input>
```

## Testing Strategy

### Unit Tests
- Test each error detection pattern
- Test auto-recovery logic
- Test health check functions

### Integration Tests
- Simulate server crashes
- Test recovery from OOM situations
- Validate error messages are helpful

### User Testing
- Can users recover from errors without documentation?
- Do error messages make sense to non-technical users?
- Are suggested commands copy-pasteable and working?

## Monitoring Points

1. **Server Startup**
   - Did server start successfully?
   - Is reported context size what we requested?
   - Is GPU/CPU mode correct?

2. **Inference Request**
   - Is server still running?
   - Does context size still match?
   - Is response time acceptable?

3. **Error Recovery**
   - Did auto-restart succeed?
   - Did fallback configuration work?
   - Was user informed of changes?

## Success Metrics

- **Zero manual server management**: Users never use `pkill` or manual restarts
- **Clear feedback**: 100% of errors include actionable next steps
- **Auto-recovery**: 90%+ of issues resolved without user intervention
- **User confidence**: Users trust the system will handle server issues

## Future Enhancements

1. **Persistent Server**: Keep server running between commands
2. **Smart Caching**: Cache model loads to reduce startup time
3. **Multi-Model Support**: Automatically switch to smaller model when needed
4. **Cloud Fallback**: Auto-suggest cloud when local fails repeatedly
