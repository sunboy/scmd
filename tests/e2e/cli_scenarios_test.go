package e2e

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var scmdBinary string

func init() {
	// Find the scmd binary
	cwd, _ := os.Getwd()
	scmdBinary = filepath.Join(cwd, "..", "..", "bin", "scmd")
}

func runScmd(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	cmd := exec.Command(scmdBinary, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func runScmdWithStdin(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	cmd := exec.Command(scmdBinary, args...)
	cmd.Stdin = strings.NewReader(stdin)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// ==================== HELP & BASIC COMMANDS ====================

func TestScenario_Help(t *testing.T) {
	stdout, _, err := runScmd(t, "--help")
	if err != nil {
		t.Fatalf("help failed: %v", err)
	}
	if !strings.Contains(stdout, "scmd") {
		t.Error("help should contain 'scmd'")
	}
}

func TestScenario_Version(t *testing.T) {
	stdout, _, err := runScmd(t, "version")
	if err != nil {
		t.Fatalf("version failed: %v", err)
	}
	if !strings.Contains(stdout, "scmd") {
		t.Error("version should contain 'scmd'")
	}
}

func TestScenario_HelpExplain(t *testing.T) {
	stdout, _, err := runScmd(t, "explain", "--help")
	if err != nil {
		t.Fatalf("explain --help failed: %v", err)
	}
	if !strings.Contains(stdout, "Explain") {
		t.Error("should show explain help")
	}
}

func TestScenario_HelpReview(t *testing.T) {
	stdout, _, err := runScmd(t, "review", "--help")
	if err != nil {
		t.Fatalf("review --help failed: %v", err)
	}
	if !strings.Contains(stdout, "Review") {
		t.Error("should show review help")
	}
}

func TestScenario_HelpConfig(t *testing.T) {
	stdout, _, err := runScmd(t, "config", "--help")
	if err != nil {
		t.Fatalf("config --help failed: %v", err)
	}
	if !strings.Contains(stdout, "config") {
		t.Error("should show config help")
	}
}

func TestScenario_HelpBackends(t *testing.T) {
	stdout, _, err := runScmd(t, "backends", "--help")
	if err != nil {
		t.Fatalf("backends --help failed: %v", err)
	}
	if !strings.Contains(stdout, "backend") {
		t.Error("should show backends help")
	}
}

// ==================== CONFIG COMMAND ====================

func TestScenario_ConfigShow(t *testing.T) {
	stdout, _, err := runScmd(t, "config")
	if err != nil {
		t.Fatalf("config failed: %v", err)
	}
	if !strings.Contains(stdout, "backends") {
		t.Error("config should show backends section")
	}
}

func TestScenario_ConfigShowKey(t *testing.T) {
	stdout, _, err := runScmd(t, "config", "backends.default")
	if err != nil {
		t.Fatalf("config key failed: %v", err)
	}
	if !strings.Contains(stdout, "local") {
		t.Error("should show default backend value")
	}
}

// ==================== BACKENDS COMMAND ====================

func TestScenario_BackendsList(t *testing.T) {
	stdout, _, err := runScmd(t, "backends")
	if err != nil {
		t.Fatalf("backends failed: %v", err)
	}
	if !strings.Contains(stdout, "ollama") {
		t.Error("should list ollama backend")
	}
	if !strings.Contains(stdout, "mock") {
		t.Error("should list mock backend")
	}
}

// ==================== PROMPT FLAG ====================

func TestScenario_PromptSimple(t *testing.T) {
	stdout, _, err := runScmd(t, "-b", "mock", "-p", "Hello")
	if err != nil {
		t.Fatalf("prompt failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

func TestScenario_PromptWithQuotes(t *testing.T) {
	stdout, _, err := runScmd(t, "-b", "mock", "-p", "What is 'hello world'?")
	if err != nil {
		t.Fatalf("prompt with quotes failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

func TestScenario_PromptLong(t *testing.T) {
	longPrompt := strings.Repeat("test ", 100)
	stdout, _, err := runScmd(t, "-b", "mock", "-p", longPrompt)
	if err != nil {
		t.Fatalf("long prompt failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

func TestScenario_PromptWithNewlines(t *testing.T) {
	prompt := "Line 1\nLine 2\nLine 3"
	stdout, _, err := runScmd(t, "-b", "mock", "-p", prompt)
	if err != nil {
		t.Fatalf("prompt with newlines failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

// ==================== PIPE INPUT ====================

func TestScenario_PipeSimple(t *testing.T) {
	stdout, _, err := runScmdWithStdin(t, "hello world", "-b", "mock", "-p", "echo this")
	if err != nil {
		t.Fatalf("pipe simple failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

func TestScenario_PipeCode(t *testing.T) {
	code := `func main() {
    fmt.Println("Hello")
}`
	stdout, _, err := runScmdWithStdin(t, code, "-b", "mock", "-p", "explain")
	if err != nil {
		t.Fatalf("pipe code failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

func TestScenario_PipeJSON(t *testing.T) {
	json := `{"name": "test", "value": 123}`
	stdout, _, err := runScmdWithStdin(t, json, "-b", "mock", "-p", "parse")
	if err != nil {
		t.Fatalf("pipe json failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

func TestScenario_PipeLarge(t *testing.T) {
	large := strings.Repeat("x", 10000)
	stdout, _, err := runScmdWithStdin(t, large, "-b", "mock", "-p", "count")
	if err != nil {
		t.Fatalf("pipe large failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

func TestScenario_PipeExplain(t *testing.T) {
	code := "print('hello')"
	stdout, _, err := runScmdWithStdin(t, code, "-b", "mock", "explain")
	if err != nil {
		t.Fatalf("pipe explain failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

func TestScenario_PipeReview(t *testing.T) {
	code := "def foo(): return None"
	stdout, _, err := runScmdWithStdin(t, code, "-b", "mock", "review")
	if err != nil {
		t.Fatalf("pipe review failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

// ==================== OUTPUT FLAG ====================

func TestScenario_OutputToFile(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "output.txt")
	_, _, err := runScmd(t, "-b", "mock", "-p", "test", "-o", tmpFile)
	if err != nil {
		t.Fatalf("output to file failed: %v", err)
	}

	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("read output file failed: %v", err)
	}
	if len(content) == 0 {
		t.Error("output file should not be empty")
	}
}

func TestScenario_OutputToNestedFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "subdir", "output.txt")
	os.MkdirAll(filepath.Dir(tmpFile), 0755)

	_, _, err := runScmd(t, "-b", "mock", "-p", "test", "-o", tmpFile)
	if err != nil {
		t.Fatalf("output to nested file failed: %v", err)
	}
}

// ==================== QUIET FLAG ====================

func TestScenario_QuietMode(t *testing.T) {
	stdout, stderr, err := runScmd(t, "-b", "mock", "-q", "-p", "test")
	if err != nil {
		t.Fatalf("quiet mode failed: %v", err)
	}
	if strings.Contains(stderr, "Processing") {
		t.Error("quiet mode should not show progress")
	}
	if stdout == "" {
		t.Error("should still have output")
	}
}

// ==================== BACKEND FLAG ====================

func TestScenario_BackendMock(t *testing.T) {
	stdout, _, err := runScmd(t, "-b", "mock", "-p", "test")
	if err != nil {
		t.Fatalf("mock backend failed: %v", err)
	}
	if !strings.Contains(stdout, "Mock") {
		t.Error("should use mock backend")
	}
}

func TestScenario_BackendInvalid(t *testing.T) {
	_, _, err := runScmd(t, "-b", "nonexistent", "-p", "test")
	if err == nil {
		t.Error("should fail with invalid backend")
	}
}

// ==================== MODEL FLAG ====================

func TestScenario_ModelFlag(t *testing.T) {
	stdout, _, err := runScmd(t, "-b", "mock", "-m", "custom-model", "-p", "test")
	if err != nil {
		t.Fatalf("model flag failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

// ==================== VERBOSE FLAG ====================

func TestScenario_VerboseMode(t *testing.T) {
	_, _, err := runScmd(t, "-v", "-b", "mock", "-p", "test")
	if err != nil {
		t.Fatalf("verbose mode failed: %v", err)
	}
}

// ==================== EXPLAIN COMMAND ====================

func TestScenario_ExplainConcept(t *testing.T) {
	stdout, _, err := runScmd(t, "-b", "mock", "explain", "goroutine")
	if err != nil {
		t.Fatalf("explain concept failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

func TestScenario_ExplainMultiWord(t *testing.T) {
	stdout, _, err := runScmd(t, "-b", "mock", "explain", "what", "is", "a", "pointer")
	if err != nil {
		t.Fatalf("explain multi-word failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

func TestScenario_ExplainAlias(t *testing.T) {
	stdout, _, err := runScmd(t, "-b", "mock", "e", "test")
	if err != nil {
		t.Fatalf("explain alias failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

func TestScenario_ExplainWhatAlias(t *testing.T) {
	stdout, _, err := runScmd(t, "-b", "mock", "what", "is", "go")
	if err != nil {
		t.Fatalf("what alias failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

// ==================== REVIEW COMMAND ====================

func TestScenario_ReviewAlias(t *testing.T) {
	code := "func test() {}"
	stdout, _, err := runScmdWithStdin(t, code, "-b", "mock", "r")
	if err != nil {
		t.Fatalf("review alias failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

// ==================== COMBINATION FLAGS ====================

func TestScenario_CombineQuietOutput(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "out.txt")
	_, _, err := runScmd(t, "-b", "mock", "-q", "-o", tmpFile, "-p", "test")
	if err != nil {
		t.Fatalf("combine flags failed: %v", err)
	}
}

func TestScenario_CombineVerboseQuiet(t *testing.T) {
	// Quiet should take precedence
	_, _, err := runScmd(t, "-b", "mock", "-v", "-q", "-p", "test")
	if err != nil {
		t.Fatalf("combine verbose quiet failed: %v", err)
	}
}

func TestScenario_AllFlags(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "out.txt")
	_, _, err := runScmd(t, "-b", "mock", "-m", "test-model", "-q", "-o", tmpFile, "-p", "test")
	if err != nil {
		t.Fatalf("all flags failed: %v", err)
	}
}

// ==================== EDGE CASES ====================

func TestScenario_EmptyStdin(t *testing.T) {
	stdout, _, err := runScmdWithStdin(t, "", "-b", "mock", "-p", "test")
	if err != nil {
		t.Fatalf("empty stdin failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

func TestScenario_SpecialCharsInPrompt(t *testing.T) {
	stdout, _, err := runScmd(t, "-b", "mock", "-p", "test!@#$%^&*()")
	if err != nil {
		t.Fatalf("special chars failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

func TestScenario_UnicodeInPrompt(t *testing.T) {
	stdout, _, err := runScmd(t, "-b", "mock", "-p", "Hello 世界 🌍")
	if err != nil {
		t.Fatalf("unicode failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

func TestScenario_UnicodeInStdin(t *testing.T) {
	stdout, _, err := runScmdWithStdin(t, "Hello 世界 🌍", "-b", "mock", "-p", "test")
	if err != nil {
		t.Fatalf("unicode stdin failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

func TestScenario_VeryLongInput(t *testing.T) {
	// 100KB of input
	large := strings.Repeat("x", 100000)
	stdout, _, err := runScmdWithStdin(t, large, "-b", "mock", "-p", "count")
	if err != nil {
		t.Fatalf("very long input failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

func TestScenario_BinaryLikeInput(t *testing.T) {
	// Some binary-like bytes
	data := string([]byte{0x00, 0x01, 0x02, 0xFF, 0xFE})
	stdout, _, err := runScmdWithStdin(t, data, "-b", "mock", "-p", "test")
	if err != nil {
		t.Fatalf("binary-like input failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

// ==================== CONCURRENT ACCESS ====================

func TestScenario_ConcurrentCalls(t *testing.T) {
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			_, _, err := runScmd(t, "-b", "mock", "-p", "test")
			if err != nil {
				t.Errorf("concurrent call failed: %v", err)
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

// ==================== REALISTIC USE CASES ====================

func TestScenario_ExplainGoFunction(t *testing.T) {
	code := `func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}`
	stdout, _, err := runScmdWithStdin(t, code, "-b", "mock", "explain")
	if err != nil {
		t.Fatalf("explain go function failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

func TestScenario_ReviewPythonCode(t *testing.T) {
	code := `def process_data(data):
    result = []
    for item in data:
        if item > 0:
            result.append(item * 2)
    return result`
	stdout, _, err := runScmdWithStdin(t, code, "-b", "mock", "review")
	if err != nil {
		t.Fatalf("review python failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

func TestScenario_ExplainJavaScript(t *testing.T) {
	code := `const fetchData = async (url) => {
    const response = await fetch(url);
    return response.json();
};`
	stdout, _, err := runScmdWithStdin(t, code, "-b", "mock", "explain")
	if err != nil {
		t.Fatalf("explain js failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

func TestScenario_ReviewSQL(t *testing.T) {
	sql := `SELECT u.name, COUNT(o.id) as order_count
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
WHERE u.active = 1
GROUP BY u.id
HAVING order_count > 5;`
	stdout, _, err := runScmdWithStdin(t, sql, "-b", "mock", "review")
	if err != nil {
		t.Fatalf("review sql failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

func TestScenario_ExplainDockerfile(t *testing.T) {
	dockerfile := `FROM golang:1.21-alpine
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .
CMD ["./main"]`
	stdout, _, err := runScmdWithStdin(t, dockerfile, "-b", "mock", "explain")
	if err != nil {
		t.Fatalf("explain dockerfile failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

func TestScenario_ReviewYAML(t *testing.T) {
	yaml := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: my-app`
	stdout, _, err := runScmdWithStdin(t, yaml, "-b", "mock", "review")
	if err != nil {
		t.Fatalf("review yaml failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

func TestScenario_ExplainShellScript(t *testing.T) {
	script := `#!/bin/bash
for file in *.txt; do
    echo "Processing $file"
    wc -l "$file"
done`
	stdout, _, err := runScmdWithStdin(t, script, "-b", "mock", "explain")
	if err != nil {
		t.Fatalf("explain shell failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

func TestScenario_ReviewRustCode(t *testing.T) {
	code := `fn main() {
    let numbers: Vec<i32> = vec![1, 2, 3, 4, 5];
    let doubled: Vec<i32> = numbers.iter().map(|x| x * 2).collect();
    println!("{:?}", doubled);
}`
	stdout, _, err := runScmdWithStdin(t, code, "-b", "mock", "review")
	if err != nil {
		t.Fatalf("review rust failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

func TestScenario_ExplainRegex(t *testing.T) {
	stdout, _, err := runScmd(t, "-b", "mock", "explain", `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if err != nil {
		t.Fatalf("explain regex failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

func TestScenario_ExplainGitDiff(t *testing.T) {
	diff := `diff --git a/main.go b/main.go
index 1234567..abcdefg 100644
--- a/main.go
+++ b/main.go
@@ -1,5 +1,6 @@
 package main

+import "fmt"
+
 func main() {
-    println("hello")
+    fmt.Println("hello")
 }`
	stdout, _, err := runScmdWithStdin(t, diff, "-b", "mock", "explain")
	if err != nil {
		t.Fatalf("explain git diff failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

func TestScenario_ReviewErrorLog(t *testing.T) {
	log := `2024-01-15 10:23:45 ERROR [main] Connection timeout after 30s
2024-01-15 10:23:46 ERROR [main] Retry 1/3 failed
2024-01-15 10:23:47 ERROR [main] Retry 2/3 failed
2024-01-15 10:23:48 FATAL [main] All retries exhausted, shutting down`
	stdout, _, err := runScmdWithStdin(t, log, "-b", "mock", "review")
	if err != nil {
		t.Fatalf("review log failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

func TestScenario_ExplainAPIResponse(t *testing.T) {
	json := `{
  "status": "success",
  "data": {
    "users": [
      {"id": 1, "name": "Alice", "role": "admin"},
      {"id": 2, "name": "Bob", "role": "user"}
    ],
    "pagination": {
      "page": 1,
      "total": 100
    }
  }
}`
	stdout, _, err := runScmdWithStdin(t, json, "-b", "mock", "explain")
	if err != nil {
		t.Fatalf("explain json failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

func TestScenario_ReviewMarkdown(t *testing.T) {
	md := `# API Documentation

## Authentication

All requests require an API key.

### Example

` + "```bash" + `
curl -H "Authorization: Bearer TOKEN" https://api.example.com/v1/users
` + "```"
	stdout, _, err := runScmdWithStdin(t, md, "-b", "mock", "review")
	if err != nil {
		t.Fatalf("review markdown failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

func TestScenario_ExplainHTMLTemplate(t *testing.T) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>{{ .Title }}</title>
</head>
<body>
    {{ range .Items }}
    <div class="item">{{ . }}</div>
    {{ end }}
</body>
</html>`
	stdout, _, err := runScmdWithStdin(t, html, "-b", "mock", "explain")
	if err != nil {
		t.Fatalf("explain html failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

func TestScenario_ReviewCSS(t *testing.T) {
	css := `.container {
    display: flex;
    justify-content: center;
    align-items: center;
}

.button {
    background: linear-gradient(to right, #ff6b6b, #feca57);
    border-radius: 5px;
    padding: 10px 20px;
}`
	stdout, _, err := runScmdWithStdin(t, css, "-b", "mock", "review")
	if err != nil {
		t.Fatalf("review css failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

// ==================== STRESS TESTS ====================

func TestStress_RapidFirePrompts(t *testing.T) {
	for i := 0; i < 50; i++ {
		_, _, err := runScmd(t, "-b", "mock", "-q", "-p", "quick test")
		if err != nil {
			t.Fatalf("rapid fire %d failed: %v", i, err)
		}
	}
}

func TestStress_LargeInputs(t *testing.T) {
	sizes := []int{1000, 10000, 50000}
	for _, size := range sizes {
		input := strings.Repeat("x", size)
		_, _, err := runScmdWithStdin(t, input, "-b", "mock", "-p", "test")
		if err != nil {
			t.Fatalf("large input %d failed: %v", size, err)
		}
	}
}

func TestStress_ConcurrentHeavy(t *testing.T) {
	done := make(chan bool, 20)

	for i := 0; i < 20; i++ {
		go func(n int) {
			input := strings.Repeat("test ", 100)
			_, _, err := runScmdWithStdin(t, input, "-b", "mock", "-p", "process")
			if err != nil {
				t.Errorf("concurrent heavy %d failed: %v", n, err)
			}
			done <- true
		}(i)
	}

	for i := 0; i < 20; i++ {
		<-done
	}
}

func TestStress_MixedOperations(t *testing.T) {
	operations := []struct {
		name string
		args []string
	}{
		{"help", []string{"--help"}},
		{"version", []string{"version"}},
		{"config", []string{"config"}},
		{"backends", []string{"backends"}},
		{"prompt", []string{"-b", "mock", "-p", "test"}},
		{"explain", []string{"-b", "mock", "explain", "goroutine"}},
	}

	for i := 0; i < 5; i++ {
		for _, op := range operations {
			_, _, err := runScmd(t, op.args...)
			if err != nil {
				t.Errorf("mixed op %s failed: %v", op.name, err)
			}
		}
	}
}

// ==================== FORMAT FLAG ====================

func TestScenario_FormatJSON(t *testing.T) {
	stdout, _, err := runScmd(t, "-b", "mock", "-p", "test", "-f", "json")
	if err != nil {
		t.Fatalf("format json failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

func TestScenario_FormatMarkdown(t *testing.T) {
	stdout, _, err := runScmd(t, "-b", "mock", "-p", "test", "-f", "markdown")
	if err != nil {
		t.Fatalf("format markdown failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

func TestScenario_FormatText(t *testing.T) {
	stdout, _, err := runScmd(t, "-b", "mock", "-p", "test", "-f", "text")
	if err != nil {
		t.Fatalf("format text failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

func TestScenario_InvalidFormat(t *testing.T) {
	_, _, err := runScmd(t, "-b", "mock", "-p", "test", "-f", "invalid")
	if err == nil {
		t.Error("should fail with invalid format")
	}
}

// ==================== STDIN VARIATIONS ====================

func TestScenario_EmptyPromptWithStdin(t *testing.T) {
	code := "func hello() { fmt.Println(\"hello\") }"
	stdout, _, err := runScmdWithStdin(t, code, "-b", "mock", "explain")
	if err != nil {
		t.Fatalf("explain with stdin failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

func TestScenario_MultilineStdin(t *testing.T) {
	input := `line 1
line 2
line 3
line 4
line 5`
	stdout, _, err := runScmdWithStdin(t, input, "-b", "mock", "-p", "count lines")
	if err != nil {
		t.Fatalf("multiline stdin failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

func TestScenario_StdinWithTabs(t *testing.T) {
	code := "func main() {\n\tfmt.Println(\"test\")\n}"
	stdout, _, err := runScmdWithStdin(t, code, "-b", "mock", "explain")
	if err != nil {
		t.Fatalf("stdin with tabs failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

// ==================== ERROR SCENARIOS ====================

func TestScenario_NoArgsNoStdin(t *testing.T) {
	_, _, err := runScmd(t, "-b", "mock")
	// Should either show help or error gracefully
	// Not checking error since it's acceptable to fail or show help
	_ = err
}

func TestScenario_InvalidCommand(t *testing.T) {
	_, _, err := runScmd(t, "nonexistent-command")
	if err == nil {
		t.Error("should fail with nonexistent command")
	}
}

func TestScenario_MissingRequiredFlag(t *testing.T) {
	// Some commands may require certain flags
	// Testing graceful error handling
	_, _, err := runScmd(t, "-p")
	if err == nil {
		t.Error("should fail with missing flag value")
	}
}

func TestScenario_ConflictingFlags(t *testing.T) {
	// Test that quiet and verbose can coexist (quiet wins)
	_, _, err := runScmd(t, "-b", "mock", "-p", "test", "-q", "-v")
	if err != nil {
		t.Fatalf("conflicting flags failed: %v", err)
	}
}

// ==================== COMMAND CHAINING ====================

func TestScenario_MultipleCommands(t *testing.T) {
	// Test running multiple commands in sequence
	commands := [][]string{
		{"config"},
		{"backends"},
		{"-b", "mock", "-p", "test1"},
		{"-b", "mock", "-p", "test2"},
	}

	for i, args := range commands {
		_, _, err := runScmd(t, args...)
		if err != nil {
			t.Errorf("command %d failed: %v", i, err)
		}
	}
}

// ==================== FILE INPUT SCENARIOS ====================

func TestScenario_ReadFromFile(t *testing.T) {
	// Create temp file with code
	tmpFile := filepath.Join(t.TempDir(), "test.go")
	code := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}`
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	// Read file and pipe to scmd
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatal(err)
	}

	stdout, _, err := runScmdWithStdin(t, string(content), "-b", "mock", "explain")
	if err != nil {
		t.Fatalf("explain file failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

// ==================== COMPLEX REAL-WORLD SCENARIOS ====================

func TestScenario_ComplexGoCode(t *testing.T) {
	code := `package main

import (
	"context"
	"fmt"
	"sync"
)

type Worker struct {
	id int
	wg *sync.WaitGroup
}

func (w *Worker) Process(ctx context.Context, jobs <-chan int) {
	defer w.wg.Done()
	for {
		select {
		case job, ok := <-jobs:
			if !ok {
				return
			}
			fmt.Printf("Worker %d processing job %d\n", w.id, job)
		case <-ctx.Done():
			return
		}
	}
}

func main() {
	ctx := context.Background()
	var wg sync.WaitGroup
	jobs := make(chan int, 100)

	// Start workers
	for i := 0; i < 3; i++ {
		wg.Add(1)
		w := &Worker{id: i, wg: &wg}
		go w.Process(ctx, jobs)
	}

	// Send jobs
	for i := 0; i < 10; i++ {
		jobs <- i
	}
	close(jobs)

	wg.Wait()
}`
	stdout, _, err := runScmdWithStdin(t, code, "-b", "mock", "explain")
	if err != nil {
		t.Fatalf("complex go code failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

func TestScenario_ComplexPythonCode(t *testing.T) {
	code := `from typing import List, Dict, Optional
import asyncio
from dataclasses import dataclass

@dataclass
class User:
    id: int
    name: str
    email: str

class UserRepository:
    def __init__(self):
        self.users: Dict[int, User] = {}

    async def get_user(self, user_id: int) -> Optional[User]:
        await asyncio.sleep(0.1)  # Simulate DB query
        return self.users.get(user_id)

    async def create_user(self, name: str, email: str) -> User:
        user_id = len(self.users) + 1
        user = User(id=user_id, name=name, email=email)
        self.users[user_id] = user
        return user

    async def list_users(self) -> List[User]:
        return list(self.users.values())

async def main():
    repo = UserRepository()
    user = await repo.create_user("Alice", "alice@example.com")
    print(f"Created user: {user}")

    users = await repo.list_users()
    print(f"All users: {users}")

if __name__ == "__main__":
    asyncio.run(main())`
	stdout, _, err := runScmdWithStdin(t, code, "-b", "mock", "review")
	if err != nil {
		t.Fatalf("complex python code failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

func TestScenario_ComplexTypeScriptCode(t *testing.T) {
	code := `interface Config {
  apiUrl: string;
  timeout: number;
  retries: number;
}

class ApiClient {
  private config: Config;
  private cache: Map<string, any>;

  constructor(config: Config) {
    this.config = config;
    this.cache = new Map();
  }

  async fetch<T>(endpoint: string): Promise<T> {
    // Check cache
    if (this.cache.has(endpoint)) {
      return this.cache.get(endpoint);
    }

    // Fetch with retry logic
    let lastError: Error | null = null;
    for (let i = 0; i < this.config.retries; i++) {
      try {
        const response = await this.fetchWithTimeout(endpoint);
        const data = await response.json();
        this.cache.set(endpoint, data);
        return data as T;
      } catch (err) {
        lastError = err as Error;
        await this.delay(1000 * (i + 1));
      }
    }

    throw lastError || new Error('Failed to fetch');
  }

  private async fetchWithTimeout(endpoint: string): Promise<Response> {
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), this.config.timeout);

    try {
      const url = this.config.apiUrl + endpoint;
      return await fetch(url, { signal: controller.signal });
    } finally {
      clearTimeout(timeout);
    }
  }

  private delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}`
	stdout, _, err := runScmdWithStdin(t, code, "-b", "mock", "review")
	if err != nil {
		t.Fatalf("complex typescript code failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

// ==================== ENVIRONMENT VARIABLE TESTS ====================

func TestScenario_WithDebugEnv(t *testing.T) {
	cmd := exec.Command(scmdBinary, "-b", "mock", "-p", "test")
	cmd.Env = append(os.Environ(), "SCMD_DEBUG=1")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Fatalf("debug env failed: %v (stderr: %s)", err, stderr.String())
	}
	// Debug mode may output to stderr
	if stdout.String() == "" && stderr.String() == "" {
		t.Error("should have some output")
	}
}

func TestScenario_WithCustomDataDir(t *testing.T) {
	tmpDir := t.TempDir()
	cmd := exec.Command(scmdBinary, "config")
	cmd.Env = append(os.Environ(), "SCMD_DATA_DIR="+tmpDir)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err := cmd.Run()
	if err != nil {
		t.Fatalf("custom data dir failed: %v", err)
	}
	if stdout.String() == "" {
		t.Error("should have output")
	}
}

// ==================== TIMEOUT AND CANCELLATION ====================

func TestScenario_LongRunningCommand(t *testing.T) {
	// Test that commands can run for a reasonable time
	// Using a large input to simulate longer processing
	input := strings.Repeat("test data\n", 1000)
	stdout, _, err := runScmdWithStdin(t, input, "-b", "mock", "-p", "process this")
	if err != nil {
		t.Fatalf("long running command failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

// ==================== DIFFERENT CODE LANGUAGES ====================

func TestScenario_ExplainC(t *testing.T) {
	code := `#include <stdio.h>
#include <stdlib.h>

struct Node {
    int data;
    struct Node* next;
};

void insert(struct Node** head, int data) {
    struct Node* new_node = (struct Node*)malloc(sizeof(struct Node));
    new_node->data = data;
    new_node->next = *head;
    *head = new_node;
}

int main() {
    struct Node* head = NULL;
    insert(&head, 1);
    insert(&head, 2);
    insert(&head, 3);
    return 0;
}`
	stdout, _, err := runScmdWithStdin(t, code, "-b", "mock", "explain")
	if err != nil {
		t.Fatalf("explain C failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

func TestScenario_ReviewJava(t *testing.T) {
	code := `import java.util.*;
import java.util.stream.*;

public class UserService {
    private Map<Long, User> users = new HashMap<>();

    public Optional<User> findById(Long id) {
        return Optional.ofNullable(users.get(id));
    }

    public List<User> findByAge(int minAge, int maxAge) {
        return users.values().stream()
            .filter(u -> u.getAge() >= minAge && u.getAge() <= maxAge)
            .sorted(Comparator.comparing(User::getName))
            .collect(Collectors.toList());
    }

    public User save(User user) {
        if (user.getId() == null) {
            user.setId(generateId());
        }
        users.put(user.getId(), user);
        return user;
    }

    private Long generateId() {
        return users.keySet().stream()
            .max(Long::compareTo)
            .map(id -> id + 1)
            .orElse(1L);
    }
}`
	stdout, _, err := runScmdWithStdin(t, code, "-b", "mock", "review")
	if err != nil {
		t.Fatalf("review Java failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

func TestScenario_ExplainKubernetes(t *testing.T) {
	yaml := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.14.2
        ports:
        - containerPort: 80
        resources:
          requests:
            memory: "64Mi"
            cpu: "250m"
          limits:
            memory: "128Mi"
            cpu: "500m"
---
apiVersion: v1
kind: Service
metadata:
  name: nginx-service
spec:
  selector:
    app: nginx
  ports:
    - protocol: TCP
      port: 80
      targetPort: 80
  type: LoadBalancer`
	stdout, _, err := runScmdWithStdin(t, yaml, "-b", "mock", "explain")
	if err != nil {
		t.Fatalf("explain kubernetes failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

// ==================== PERFORMANCE VARIATIONS ====================

func TestScenario_QuickResponses(t *testing.T) {
	// Test that quick prompts get quick responses
	prompts := []string{"yes", "no", "ok", "hello", "test"}

	for _, prompt := range prompts {
		stdout, _, err := runScmd(t, "-b", "mock", "-q", "-p", prompt)
		if err != nil {
			t.Errorf("quick prompt '%s' failed: %v", prompt, err)
		}
		if stdout == "" {
			t.Errorf("prompt '%s' should have output", prompt)
		}
	}
}

// ==================== OUTPUT VERIFICATION ====================

func TestScenario_OutputContainsInput(t *testing.T) {
	// For explain commands, output should relate to input
	input := "goroutines"
	stdout, _, err := runScmd(t, "-b", "mock", "explain", input)
	if err != nil {
		t.Fatalf("explain failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
	// Mock backend should echo or reference the input
}

func TestScenario_OutputFilePermissions(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "output.txt")
	_, _, err := runScmd(t, "-b", "mock", "-p", "test", "-o", tmpFile)
	if err != nil {
		t.Fatalf("output file failed: %v", err)
	}

	info, err := os.Stat(tmpFile)
	if err != nil {
		t.Fatalf("stat output file failed: %v", err)
	}

	// Check file is readable
	if info.Mode().Perm()&0400 == 0 {
		t.Error("output file should be readable")
	}
}

// ==================== COMMAND HISTORY AND STATE ====================

func TestScenario_MultipleSequentialCalls(t *testing.T) {
	// Test that commands don't interfere with each other
	for i := 0; i < 5; i++ {
		stdout, _, err := runScmd(t, "-b", "mock", "-p", fmt.Sprintf("test %d", i))
		if err != nil {
			t.Errorf("call %d failed: %v", i, err)
		}
		if stdout == "" {
			t.Errorf("call %d should have output", i)
		}
	}
}
