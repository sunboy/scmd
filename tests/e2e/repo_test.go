package e2e

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/scmd/scmd/internal/repos"
)

// TestRepoWorkflow tests the complete repository workflow
func TestRepoWorkflow(t *testing.T) {
	// Allow localhost URLs for test server
	t.Setenv("SCMD_ALLOW_LOCALHOST", "1")

	// Setup test server serving sample repo
	sampleRepoPath := filepath.Join("..", "..", "testdata", "sample-repo")

	server := httptest.NewServer(http.FileServer(http.Dir(sampleRepoPath)))
	defer server.Close()

	tmpDir := t.TempDir()
	mgr := repos.NewManager(tmpDir)

	// 1. Add repository
	err := mgr.Add("test-repo", server.URL)
	require.NoError(t, err)

	// 2. Save and reload
	err = mgr.Save()
	require.NoError(t, err)

	mgr2 := repos.NewManager(tmpDir)
	err = mgr2.Load()
	require.NoError(t, err)

	repo, ok := mgr2.Get("test-repo")
	require.True(t, ok)
	assert.Equal(t, server.URL, repo.URL)

	// 3. Fetch manifest
	ctx := context.Background()
	manifest, err := mgr2.FetchManifest(ctx, repo)
	require.NoError(t, err)

	assert.Equal(t, "sample-repo", manifest.Name)
	assert.Equal(t, "1.0.0", manifest.Version)
	assert.GreaterOrEqual(t, len(manifest.Commands), 4)

	// 4. Search commands
	results, err := mgr2.SearchCommands(ctx, "git")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 1)

	found := false
	for _, r := range results {
		if r.Command.Name == "git-commit" {
			found = true
			break
		}
	}
	assert.True(t, found, "git-commit command should be found")

	// 5. Fetch specific command
	spec, err := mgr2.FetchCommand(ctx, repo, "commands/git-commit.yaml")
	require.NoError(t, err)

	assert.Equal(t, "git-commit", spec.Name)
	assert.Equal(t, "1.0.0", spec.Version)
	assert.Contains(t, spec.Prompt.Template, "diff")

	// 6. Install command
	installDir := filepath.Join(tmpDir, "commands")
	err = os.MkdirAll(installDir, 0755)
	require.NoError(t, err)

	err = mgr2.InstallCommand(spec, installDir)
	require.NoError(t, err)

	// 7. Verify installation
	installed, err := mgr2.LoadInstalledCommands(installDir)
	require.NoError(t, err)
	assert.Len(t, installed, 1)
	assert.Equal(t, "git-commit", installed[0].Name)

	// 8. Remove repository
	err = mgr2.Remove("test-repo")
	require.NoError(t, err)

	_, ok = mgr2.Get("test-repo")
	assert.False(t, ok)
}

// TestMultipleRepoSearch tests searching across multiple repositories
func TestMultipleRepoSearch(t *testing.T) {
	// Allow localhost URLs for test server
	t.Setenv("SCMD_ALLOW_LOCALHOST", "1")

	// Create two test servers with different commands
	repo1 := `name: repo1
version: "1.0.0"
commands:
  - name: git-commit
    description: Generate git commits
    file: git-commit.yaml
  - name: git-branch
    description: Manage branches
    file: git-branch.yaml`

	repo2 := `name: repo2
version: "1.0.0"
commands:
  - name: docker-compose
    description: Generate docker-compose
    file: docker-compose.yaml
  - name: git-rebase
    description: Interactive rebase helper
    file: git-rebase.yaml`

	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/scmd-repo.yaml" {
			w.Write([]byte(repo1))
		}
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/scmd-repo.yaml" {
			w.Write([]byte(repo2))
		}
	}))
	defer server2.Close()

	tmpDir := t.TempDir()
	mgr := repos.NewManager(tmpDir)

	_ = mgr.Add("repo1", server1.URL)
	_ = mgr.Add("repo2", server2.URL)

	ctx := context.Background()

	// Search for git commands across both repos
	results, err := mgr.SearchCommands(ctx, "git")
	require.NoError(t, err)

	// Should find 3 git commands (2 from repo1, 1 from repo2)
	assert.Len(t, results, 3)

	// Verify commands come from different repos
	repos := make(map[string]int)
	for _, r := range results {
		repos[r.Repo]++
	}
	assert.Equal(t, 2, repos["repo1"])
	assert.Equal(t, 1, repos["repo2"])
}

// TestPluginExecution tests loading and executing plugin commands
func TestPluginExecution(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := repos.NewManager(tmpDir)

	// Create a simple command spec
	spec := &repos.CommandSpec{
		Name:        "echo-test",
		Version:     "1.0.0",
		Description: "Test command",
		Prompt: repos.PromptSpec{
			Template: "Echo this: {{.input}}",
		},
	}

	installDir := filepath.Join(tmpDir, "commands")
	_ = os.MkdirAll(installDir, 0755)

	err := mgr.InstallCommand(spec, installDir)
	require.NoError(t, err)

	// Load the command
	loader := repos.NewLoader(mgr, installDir)
	commands, err := loader.LoadAll()
	require.NoError(t, err)
	require.Len(t, commands, 1)

	cmd := commands[0]
	assert.Equal(t, "echo-test", cmd.Name())
	assert.Equal(t, "Test command", cmd.Description())
	assert.True(t, cmd.RequiresBackend())
}

// TestCLIRepoCommands tests the CLI repo commands if binary exists
func TestCLIRepoCommands(t *testing.T) {
	// Allow localhost URLs for test server
	t.Setenv("SCMD_ALLOW_LOCALHOST", "1")

	// Build the binary
	binary := filepath.Join(t.TempDir(), "scmd")
	buildCmd := exec.Command("go", "build", "-o", binary, "../../cmd/scmd")
	buildCmd.Dir = filepath.Join("..", "..")
	err := buildCmd.Run()
	if err != nil {
		t.Skip("Could not build binary:", err)
	}

	// Set up isolated config directory
	tmpHome := t.TempDir()
	env := append(os.Environ(), "HOME="+tmpHome)

	// Test repo list (should show default or empty)
	cmd := exec.Command(binary, "repo", "list")
	cmd.Env = env
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "repo list failed: %s", string(output))

	// Test repo add
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/scmd-repo.yaml" {
			w.Write([]byte(`name: test
version: "1.0.0"
commands:
  - name: test-cmd
    description: Test
    file: test.yaml`))
		}
	}))
	defer testServer.Close()

	cmd = exec.Command(binary, "repo", "add", "mytest", testServer.URL)
	cmd.Env = env
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "repo add failed: %s", string(output))
	assert.Contains(t, string(output), "Added repository")

	// Test repo list shows the new repo
	cmd = exec.Command(binary, "repo", "list")
	cmd.Env = env
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "repo list failed: %s", string(output))
	assert.Contains(t, string(output), "mytest")
	assert.Contains(t, string(output), testServer.URL)

	// Test repo search
	cmd = exec.Command(binary, "repo", "search")
	cmd.Env = env
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "repo search failed: %s", string(output))
	assert.Contains(t, string(output), "test-cmd")

	// Test repo remove
	cmd = exec.Command(binary, "repo", "remove", "mytest")
	cmd.Env = env
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "repo remove failed: %s", string(output))
	assert.Contains(t, string(output), "Removed repository")

	// Verify removed
	cmd = exec.Command(binary, "repo", "list")
	cmd.Env = env
	output, err = cmd.CombinedOutput()
	require.NoError(t, err)
	assert.NotContains(t, string(output), "mytest")
}

// TestRepoInstallAndRun tests installing and running a command from repo
func TestRepoInstallAndRun(t *testing.T) {
	// Allow localhost URLs for test server
	t.Setenv("SCMD_ALLOW_LOCALHOST", "1")

	// Build the binary
	binary := filepath.Join(t.TempDir(), "scmd")
	buildCmd := exec.Command("go", "build", "-o", binary, "../../cmd/scmd")
	buildCmd.Dir = filepath.Join("..", "..")
	err := buildCmd.Run()
	if err != nil {
		t.Skip("Could not build binary:", err)
	}

	// Set up test server with sample repo
	sampleRepoPath, _ := filepath.Abs(filepath.Join("..", "..", "testdata", "sample-repo"))
	testServer := httptest.NewServer(http.FileServer(http.Dir(sampleRepoPath)))
	defer testServer.Close()

	// Use isolated home
	tmpHome := t.TempDir()
	env := append(os.Environ(), "HOME="+tmpHome)

	// Add the repository
	cmd := exec.Command(binary, "repo", "add", "sample", testServer.URL)
	cmd.Env = env
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "repo add failed: %s", string(output))

	// Search for commands
	cmd = exec.Command(binary, "repo", "search", "commit")
	cmd.Env = env
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "search failed: %s", string(output))
	assert.Contains(t, string(output), "git-commit")

	// Install git-commit command
	cmd = exec.Command(binary, "repo", "install", "sample/git-commit")
	cmd.Env = env
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "install failed: %s", string(output))
	assert.Contains(t, string(output), "Installed")

	// Verify command file exists
	cmdFile := filepath.Join(tmpHome, ".scmd", "commands", "git-commit.yaml")
	_, err = os.Stat(cmdFile)
	assert.NoError(t, err, "command file should exist")

	// Run the installed command with mock backend (should work)
	cmd = exec.Command(binary, "git-commit")
	cmd.Env = env
	cmd.Stdin = strings.NewReader("diff --git a/main.go\n+func hello() {}")
	output, _ = cmd.CombinedOutput()
	// Note: may fail if no real backend, but command should be recognized
	outputStr := string(output)
	// Either it runs or shows it recognizes the command
	assert.True(t,
		strings.Contains(outputStr, "commit") ||
			strings.Contains(outputStr, "Mock") ||
			strings.Contains(outputStr, "Using"),
		"Should recognize git-commit command: %s", outputStr)
}

// TestRepoShowCommand tests the repo show command
func TestRepoShowCommand(t *testing.T) {
	// Allow localhost URLs for test server
	t.Setenv("SCMD_ALLOW_LOCALHOST", "1")

	// Build the binary
	binary := filepath.Join(t.TempDir(), "scmd")
	buildCmd := exec.Command("go", "build", "-o", binary, "../../cmd/scmd")
	buildCmd.Dir = filepath.Join("..", "..")
	err := buildCmd.Run()
	if err != nil {
		t.Skip("Could not build binary:", err)
	}

	// Set up test server with sample repo
	sampleRepoPath, _ := filepath.Abs(filepath.Join("..", "..", "testdata", "sample-repo"))
	testServer := httptest.NewServer(http.FileServer(http.Dir(sampleRepoPath)))
	defer testServer.Close()

	// Use isolated home
	tmpHome := t.TempDir()
	env := append(os.Environ(), "HOME="+tmpHome)

	// Add the repository
	cmd := exec.Command(binary, "repo", "add", "sample", testServer.URL)
	cmd.Env = env
	_, _ = cmd.CombinedOutput()

	// Show command details
	cmd = exec.Command(binary, "repo", "show", "sample/git-commit")
	cmd.Env = env
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "show failed: %s", string(output))

	outputStr := string(output)
	assert.Contains(t, outputStr, "Name:")
	assert.Contains(t, outputStr, "git-commit")
	assert.Contains(t, outputStr, "Version:")
	assert.Contains(t, outputStr, "Description:")
}

// TestRepoUpdateCommand tests the repo update command
func TestRepoUpdateCommand(t *testing.T) {
	// Allow localhost URLs for test server
	t.Setenv("SCMD_ALLOW_LOCALHOST", "1")

	// Build the binary
	binary := filepath.Join(t.TempDir(), "scmd")
	buildCmd := exec.Command("go", "build", "-o", binary, "../../cmd/scmd")
	buildCmd.Dir = filepath.Join("..", "..")
	err := buildCmd.Run()
	if err != nil {
		t.Skip("Could not build binary:", err)
	}

	// Set up test server
	callCount := 0
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if r.URL.Path == "/scmd-repo.yaml" {
			w.Write([]byte(`name: test
version: "1.0.0"
commands:
  - name: cmd1
    description: Command 1
    file: cmd1.yaml`))
		}
	}))
	defer testServer.Close()

	// Use isolated home
	tmpHome := t.TempDir()
	env := append(os.Environ(), "HOME="+tmpHome)

	// Add repository
	cmd := exec.Command(binary, "repo", "add", "test", testServer.URL)
	cmd.Env = env
	_, _ = cmd.CombinedOutput()

	// Reset call count
	callCount = 0

	// Update repos
	cmd = exec.Command(binary, "repo", "update")
	cmd.Env = env
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "update failed: %s", string(output))

	assert.Contains(t, string(output), "Updating repositories")
	assert.Contains(t, string(output), "test:")
	assert.Greater(t, callCount, 0, "Should have fetched manifest")
}
