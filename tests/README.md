# scmd Test Suite

Comprehensive test suite for the scmd CLI tool covering all features from multiple user perspectives.

## Test Organization

```
tests/
├── e2e/                          # End-to-end tests
│   ├── cli_scenarios_test.go     # CLI usage scenarios (170+ tests)
│   ├── models_test.go            # Model management tests
│   ├── slash_test.go             # Slash command tests
│   ├── repo_test.go              # Repository system tests
│   ├── stress_test.go            # Performance and stress tests
│   └── tools_e2e_test.go         # (pending) Tool calling E2E tests
├── security/                     # Security tests
│   └── security_test.go          # Security vulnerability tests
├── integration/                  # (pending) Integration tests
│   └── workflows_test.go         # End-to-end workflow tests
└── testutil/                     # Test utilities
    ├── mock_backend.go
    └── mock_ui.go
```

## Test Coverage

### 1. E2E CLI Scenarios Tests (`cli_scenarios_test.go`)
**170+ comprehensive test scenarios covering:**

#### Basic Commands
- Help, version, config display
- All built-in commands (explain, review, etc.)
- Command aliases and variations

#### Flags and Options
- Backend selection (`-b`)
- Model selection (`-m`)
- Prompt input (`-p`)
- Output file (`-o`)
- Output formats (`-f`: json, markdown, text)
- Quiet mode (`-q`)
- Verbose mode (`-v`)
- Flag combinations and conflicts

#### Input Handling
- Stdin piping
- File input
- Empty input
- Large inputs (1MB, 10MB, 100MB)
- Multiline input
- Unicode and special characters
- Binary-like input
- Control characters
- Very long lines (100K chars)

#### Real-World Code Examples
- Complex Go code (goroutines, channels, context)
- Python async/await code
- TypeScript with generics
- Java streams and optionals
- C linked lists
- Kubernetes YAML
- Docker configurations
- Shell scripts
- SQL queries
- CSS and HTML templates
- Markdown documents
- Error logs
- API responses (JSON)
- Git diffs

#### Environment Variables
- `SCMD_DEBUG`
- `SCMD_DATA_DIR`
- `SCMD_CONFIG`

#### Error Scenarios
- Invalid commands
- Missing required flags
- Nonexistent backends
- Invalid formats
- Command failures

### 2. Model Management Tests (`models_test.go`)
**Tests covering:**

- List available models
- Model information display
- Set default model
- Model manager operations
- Backend initialization
- Model path resolution
- Model deletion
- Downloaded model tracking
- Model validation (URLs, sizes, variants)
- Default model verification
- No duplicate models
- Tool calling support
- Token estimation

### 3. Slash Command Tests (`slash_test.go`)
**Comprehensive slash command testing:**

#### CLI Tests
- Slash command listing
- Shell integration generation (Bash, Zsh, Fish)
- Direct invocation (`/explain`, `/review`)
- Command execution with aliases
- Arguments passing

#### Runner Unit Tests
- Config loading and saving
- Command finding (by name/alias)
- Command parsing
- Adding/removing commands
- Alias management
- Config persistence
- Shell integration generation
- Command execution with backends

#### Edge Cases
- Empty configs
- Corrupt configs
- Commands with special characters
- Case-insensitive matching

### 4. Repository Tests (`repo_test.go`)
**Existing tests for:**

- Repository workflow (add, list, update, remove)
- Manifest fetching
- Command search
- Command installation
- Multi-repository search
- Plugin execution
- CLI repo commands

### 5. Stress & Performance Tests (`stress_test.go`)
**Comprehensive stress testing:**

#### Concurrency Tests
- 10/50/100 concurrent commands
- Concurrent with stdin
- Sustained load testing (10 second duration)
- Mixed operations

#### Load Tests
- Rapid fire sequential (100 requests)
- Sustained load with multiple workers
- Request per second (RPS) measurements

#### Large Input Tests
- 1MB, 10MB inputs
- 10,000 small inputs
- Very long lines (100K chars)
- Random input sizes

#### Memory & Resource Tests
- Memory pressure testing
- Goroutine leak detection
- Resource exhaustion
- Many output files (100+)

#### Recovery Tests
- Error recovery
- Mixed success/failure scenarios
- Alternating backends

#### Performance Benchmarks
- Simple command benchmark
- Command with stdin
- Large input processing
- Parallel execution

#### Scaling Tests
- 1, 2, 5, 10, 20 worker scaling
- Realistic usage patterns

### 6. Security Tests (`security/security_test.go`)
**Comprehensive security testing:**

#### Command Injection
- Prompt injection attempts
- Stdin injection attempts
- Shell metacharacter handling

#### Path Traversal
- Output file path traversal
- Config directory traversal
- Malicious file paths

#### Environment Variable Injection
- PATH manipulation
- LD_PRELOAD attempts
- DYLD_INSERT_LIBRARIES attempts

#### File Permissions
- Output file permissions
- Config file permissions
- Directory permissions

#### Input Validation
- Oversized input (100MB)
- Null bytes
- Control characters
- YAML bombs (billion laughs)

#### Configuration Security
- Malicious YAML configs
- Config validation
- Safe defaults

#### Backend Security
- Backend isolation
- File access restrictions

#### Repository Security
- Malicious repository URLs
- SSRF (Server-Side Request Forgery)
- file://, javascript:, data: URL schemes
- Internal URL access (localhost, 127.0.0.1, metadata endpoints)

#### Tool Calling Security
- Tool execution restrictions
- Dangerous operation prevention

#### Data Sanitization
- Sensitive data in output
- Error message disclosure
- Version information leakage

#### Resource Limits
- Resource exhaustion prevention
- Timeout enforcement

## Running Tests

### Run All Tests
```bash
go test ./tests/...
```

### Run Specific Test Suite
```bash
# E2E tests
go test ./tests/e2e/

# Security tests
go test ./tests/security/

# Specific test file
go test ./tests/e2e/models_test.go
```

### Run With Coverage
```bash
go test ./tests/... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Run Stress Tests
```bash
# Skip stress tests (default for quick runs)
go test ./tests/e2e/ -short

# Run stress tests
go test ./tests/e2e/stress_test.go -v
```

### Run Specific Test
```bash
go test ./tests/e2e/cli_scenarios_test.go -run TestScenario_ExplainGoFunction
```

### Run Benchmarks
```bash
go test ./tests/e2e/stress_test.go -bench=. -benchmem
```

## Test Philosophy

### Test from Different User Perspectives

1. **Beginner User**
   - First run experience
   - Help discovery
   - Basic commands
   - Error message clarity

2. **Power User**
   - Complex piping
   - Backend switching
   - Custom commands
   - Repository management

3. **Pro Developer**
   - CI/CD integration
   - Scriptability
   - Exit codes
   - JSON output

4. **Attacker/Security Tester**
   - Command injection
   - Path traversal
   - SSRF
   - Resource exhaustion

### Test Principles

1. **Comprehensive Coverage**: Test every feature, flag, and command
2. **Real-World Scenarios**: Use realistic code examples and use cases
3. **Edge Cases**: Test boundaries, limits, and unusual inputs
4. **Error Handling**: Verify graceful failure and helpful messages
5. **Performance**: Ensure scalability and resource efficiency
6. **Security**: Verify protection against common vulnerabilities
7. **Isolation**: Each test is independent and can run in parallel
8. **Clarity**: Test names clearly describe what is being tested

## Test Data

### Mock Backend
The `mock` backend is used for most tests to avoid dependencies on:
- External LLM APIs
- Model downloads
- Network availability
- llama-server installation

### Test Fixtures
- Sample repository in `testdata/sample-repo/`
- Sample command specifications
- Mock command implementations

## Future Test Additions

### Pending Test Files
1. **tools_e2e_test.go**: Tool calling end-to-end tests
2. **integration/workflows_test.go**: Complete workflow integration tests
3. **Unit tests**: For tools, composition, hooks, context modules

### Areas for Expansion
- More repository system tests
- Hook execution tests
- Context gathering tests
- Command composition tests
- Interactive mode tests
- Shell integration tests (requires shell environment)

## CI/CD Integration

### GitHub Actions Example
```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: '1.21'
      - name: Run tests
        run: go test ./tests/... -v -short
      - name: Run stress tests
        run: go test ./tests/e2e/stress_test.go -v
      - name: Run security tests
        run: go test ./tests/security/... -v
```

## Test Maintenance

### Adding New Tests
1. Identify the feature or scenario to test
2. Choose appropriate test file or create new one
3. Write clear test name: `TestFeature_Scenario`
4. Include both happy path and error cases
5. Use table-driven tests for variations
6. Clean up resources (use `t.TempDir()`)

### Test Naming Convention
- `TestScenario_*`: E2E scenario tests
- `TestStress_*`: Stress and performance tests
- `TestSecurity_*`: Security tests
- `Test*Manager_*`: Unit tests for managers
- `Benchmark*`: Performance benchmarks

### Best Practices
- Use `t.Helper()` for test helpers
- Use `t.TempDir()` for temporary files/directories
- Use `t.Parallel()` when tests can run concurrently
- Use `testing.Short()` to skip long-running tests
- Provide clear error messages with context
- Log important information with `t.Logf()`

## Coverage Goals

- **Line Coverage**: > 70%
- **Feature Coverage**: 100% of documented features
- **Critical Paths**: 100% coverage
- **Security**: All common vulnerabilities tested

## Reporting Issues

If you find bugs or missing test coverage:
1. Check if test already exists
2. Add test that reproduces the issue
3. Fix the issue
4. Verify test passes
5. Submit PR with test and fix

## License

Tests are part of the scmd project and follow the same MIT License.
# Test comment
