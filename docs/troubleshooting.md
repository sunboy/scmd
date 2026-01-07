# Troubleshooting Guide

This guide helps you diagnose and fix common issues with scmd.

## Quick Diagnosis

**First step**: Run the doctor command to check your setup:

```bash
scmd doctor
```

This will check:
- ✅ scmd installation
- ✅ Models downloaded
- ✅ llama-server availability
- ✅ System resources
- ✅ Backend connectivity

---

## Common Issues

### 1. "Cannot connect to llama-server"

**Problem**: Commands fail with connection errors.

**Solutions**:

1. **Let scmd auto-start the server** (recommended):
   ```bash
   # Just run your command - server will start automatically
   echo "Hello" | scmd /explain
   ```

2. **Manually start the server**:
   ```bash
   scmd server start
   ```

3. **Check server status**:
   ```bash
   scmd server status
   ```

4. **Use a cloud backend instead**:
   ```bash
   export OPENAI_API_KEY=your-key
   scmd -b openai /explain code.go
   ```

---

### 2. "llama-server not found"

**Problem**: llama-server binary is not installed.

**Solutions**:

**macOS**:
```bash
brew install llama.cpp
```

**Linux**:
```bash
# Build from source
git clone https://github.com/ggerganov/llama.cpp
cd llama.cpp
make llama-server
sudo cp llama-server /usr/local/bin/
```

**Alternative**: Use cloud backends (no local installation needed):
```bash
export OPENAI_API_KEY=your-key
scmd -b openai /explain
```

---

### 3. GPU Out of Memory (OOM) Crashes

**Problem**: Server crashes with "kIOGPUCommandBufferCallbackErrorOutOfMemory"

**Solutions**:

1. **Restart in CPU mode** (slower but stable):
   ```bash
   scmd server restart --cpu
   ```

2. **Use smaller context size**:
   ```bash
   scmd server restart -c 2048
   ```

3. **Switch to smaller model**:
   ```bash
   scmd server start -m qwen2.5-1.5b
   ```

4. **Close other applications** using GPU/memory

5. **Check memory with doctor**:
   ```bash
   scmd doctor
   ```

**Prevention**: scmd now auto-detects available memory and tunes configuration automatically!

---

### 4. CPU Mode is Very Slow

**Problem**: Queries take 30-60+ seconds.

**Explanation**: CPU-only inference is inherently slow. This is expected behavior.

**Solutions**:

1. **Enable GPU acceleration** (if you have a GPU):
   ```bash
   scmd server restart --gpu
   ```

2. **Use smaller model**:
   ```bash
   scmd server start -m qwen2.5-0.5b  # Fastest
   ```

3. **Use cloud backend** for faster results:
   ```bash
   export OPENAI_API_KEY=your-key
   scmd -b openai /explain
   ```

4. **Use Groq** (free tier, very fast):
   ```bash
   export GROQ_API_KEY=your-key
   scmd -b groq /explain
   ```

**Performance expectations**:
- **CPU mode**: 30-60 seconds per query (0.2-0.5 tokens/sec)
- **GPU mode (M1/M2)**: 2-5 seconds per query (~20 tokens/sec)
- **Cloud (OpenAI/Groq)**: 1-3 seconds per query

---

### 5. Model Not Downloaded

**Problem**: "Model 'xxx' not found"

**Solution**:

Models download automatically on first use, but you can also download manually:

```bash
# List available models
scmd models list

# Download specific model
scmd models download qwen3-4b

# Check downloaded models
scmd doctor
```

---

### 6. Port 8089 Already in Use

**Problem**: Another process is using port 8089.

**Solutions**:

1. **Let scmd use the existing server**:
   ```bash
   # scmd will automatically detect and use it
   scmd /explain code.go
   ```

2. **Stop the conflicting process**:
   ```bash
   # Find process on port 8089
   lsof -ti:8089

   # Kill it
   kill $(lsof -ti:8089)

   # Restart scmd server
   scmd server start
   ```

3. **Stop scmd's server**:
   ```bash
   scmd server stop
   ```

---

### 7. Commands Fail with Generic "Error"

**Problem**: Old behavior - should be fixed now!

**Solution**: Update to latest version with improved error messages:

```bash
# Build latest from source
go build -o scmd cmd/scmd/main.go
```

New error messages include:
- ❌ Clear description of the problem
- 💡 2-4 actionable solutions
- 🔗 Links to relevant documentation

---

### 8. Server Won't Start

**Problem**: `scmd server start` fails.

**Debug steps**:

1. **Check logs**:
   ```bash
   scmd server logs
   ```

2. **Run doctor**:
   ```bash
   scmd doctor
   ```

3. **Try manual start with debug**:
   ```bash
   SCMD_DEBUG=1 scmd server start
   ```

4. **Check disk space**:
   ```bash
   df -h ~/.scmd
   ```

5. **Check llama-server installation**:
   ```bash
   which llama-server
   llama-server --help
   ```

---

## Environment Variables

Control scmd behavior with environment variables:

```bash
# Disable auto-start (for debugging)
export SCMD_NO_AUTOSTART=1

# Enable debug output
export SCMD_DEBUG=1

# Set custom data directory
export SCMD_DATA_DIR=~/custom/path

# Suppress progress messages
export SCMD_QUIET=1
```

---

## Getting More Help

1. **Check logs**:
   ```bash
   scmd server logs
   tail -f ~/.scmd/logs/llama-server.log
   ```

2. **Run diagnostics**:
   ```bash
   scmd doctor
   ```

3. **Enable debug mode**:
   ```bash
   SCMD_DEBUG=1 scmd /explain test.go
   ```

4. **Report issue**:
   - GitHub: https://github.com/scmd/scmd/issues
   - Include output from `scmd doctor`
   - Include relevant error messages

---

## Performance Tuning

### Memory-Constrained Systems (< 8GB RAM)

```bash
# Use smallest model
scmd server start -m qwen2.5-0.5b --cpu

# Or use cloud backend
export GROQ_API_KEY=your-key
scmd -b groq
```

### High-Performance Systems (16+ GB RAM)

```bash
# Use larger model with more context
scmd server start -m qwen2.5-7b -c 8192 --gpu
```

### M1/M2 Macs (8GB)

```bash
# Recommended: Medium model with auto-tuned settings
scmd server start -m qwen2.5-3b
# scmd will auto-tune context size and GPU layers
```

---

## Verifying Installation

After installation, verify everything works:

```bash
# 1. Check installation
scmd doctor

# 2. Start server (should auto-start, but let's be explicit)
scmd server start

# 3. Test with simple query
echo "Hello world" | scmd /explain

# 4. Check status
scmd server status

# 5. View logs
scmd server logs --tail 20
```

Expected output:
- ✅ All `scmd doctor` checks pass (or have helpful recommendations)
- ✅ Server starts within 10-30 seconds
- ✅ Test query completes successfully
- ✅ No errors in logs

---

## Uninstalling

To completely remove scmd:

```bash
# 1. Stop server
scmd server stop

# 2. Remove data directory
rm -rf ~/.scmd

# 3. Remove binary
rm $(which scmd)

# 4. (Optional) Uninstall llama.cpp
brew uninstall llama.cpp
```

---

## FAQ

**Q: Do I need to manually start llama-server?**

A: No! As of the latest version, scmd automatically starts llama-server when needed.

**Q: Can I use scmd without installing llama.cpp?**

A: Yes! Use cloud backends:
```bash
export OPENAI_API_KEY=your-key
scmd -b openai /explain
```

**Q: Why is CPU mode so slow?**

A: CPU inference is inherently slow (30-60s per query). Use GPU mode or cloud backends for better performance.

**Q: How much disk space do I need?**

A: Models range from 400MB to 5GB:
- qwen2.5-0.5b: ~400MB
- qwen2.5-1.5b: ~1GB
- qwen2.5-3b: ~2GB
- qwen3-4b: ~2.6GB
- qwen2.5-7b: ~4.7GB

**Q: Can I use multiple models?**

A: Yes! Download multiple models and switch between them:
```bash
scmd server start -m qwen2.5-3b
scmd server restart -m qwen3-4b
```

**Q: Is my data sent to the cloud?**

A: When using llama.cpp backend: **No** - everything runs locally.
When using cloud backends (OpenAI, Groq, etc): **Yes** - data is sent to their servers.

---

**Last Updated**: January 2026
**Version**: 1.0.0
