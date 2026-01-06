// Package context provides automatic context gathering for commands
package context

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ContextSpec defines context requirements
// This mirrors repos.ContextSpec to avoid import cycles
type ContextSpec struct {
	Files     []string // File patterns to include
	Git       bool     // Include git context
	Env       []string // Environment variables
	MaxTokens int      // Max context tokens
}

// Gatherer collects context for commands
type Gatherer struct {
	workDir string
}

// NewGatherer creates a new context gatherer
func NewGatherer(workDir string) *Gatherer {
	if workDir == "" {
		workDir, _ = os.Getwd()
	}
	return &Gatherer{
		workDir: workDir,
	}
}

// Context represents gathered context
type Context struct {
	Files       map[string]string // filename -> content
	GitInfo     *GitInfo
	Environment map[string]string
	TotalTokens int // Approximate token count
}

// GitInfo contains git repository context
type GitInfo struct {
	Branch        string
	Status        string
	RecentCommits []string
	RemoteURL     string
}

// Gather collects context based on the specification
func (g *Gatherer) Gather(ctx context.Context, spec *ContextSpec) (*Context, error) {
	if spec == nil {
		return &Context{
			Files:       make(map[string]string),
			Environment: make(map[string]string),
		}, nil
	}

	result := &Context{
		Files:       make(map[string]string),
		Environment: make(map[string]string),
	}

	// Gather files
	if len(spec.Files) > 0 {
		if err := g.gatherFiles(spec.Files, result); err != nil {
			return nil, fmt.Errorf("gather files: %w", err)
		}
	}

	// Gather git context
	if spec.Git {
		gitInfo, err := g.gatherGitInfo(ctx)
		if err == nil { // Don't fail if not a git repo
			result.GitInfo = gitInfo
		}
	}

	// Gather environment variables
	if len(spec.Env) > 0 {
		g.gatherEnv(spec.Env, result)
	}

	// Estimate tokens and truncate if needed
	result.TotalTokens = g.estimateTokens(result)
	if spec.MaxTokens > 0 && result.TotalTokens > spec.MaxTokens {
		g.truncateContext(result, spec.MaxTokens)
	}

	return result, nil
}

// gatherFiles collects files matching the patterns
func (g *Gatherer) gatherFiles(patterns []string, result *Context) error {
	for _, pattern := range patterns {
		// Handle absolute and relative paths
		searchPattern := pattern
		if !filepath.IsAbs(pattern) {
			searchPattern = filepath.Join(g.workDir, pattern)
		}

		// Match files
		matches, err := filepath.Glob(searchPattern)
		if err != nil {
			return fmt.Errorf("glob pattern %s: %w", pattern, err)
		}

		for _, match := range matches {
			// Skip directories
			info, err := os.Stat(match)
			if err != nil || info.IsDir() {
				continue
			}

			// Read file (with size limit)
			if info.Size() > 1024*1024 { // 1MB limit
				result.Files[match] = fmt.Sprintf("[File too large: %d bytes]", info.Size())
				continue
			}

			content, err := os.ReadFile(match)
			if err != nil {
				result.Files[match] = fmt.Sprintf("[Error reading file: %v]", err)
				continue
			}

			// Store relative path if possible
			relPath := match
			if rel, err := filepath.Rel(g.workDir, match); err == nil {
				relPath = rel
			}

			result.Files[relPath] = string(content)
		}
	}

	return nil
}

// gatherGitInfo collects git repository information
func (g *Gatherer) gatherGitInfo(ctx context.Context) (*GitInfo, error) {
	info := &GitInfo{}

	// Get current branch
	if branch, err := g.runGitCommand(ctx, "rev-parse", "--abbrev-ref", "HEAD"); err == nil {
		info.Branch = strings.TrimSpace(branch)
	}

	// Get status
	if status, err := g.runGitCommand(ctx, "status", "--short"); err == nil {
		info.Status = strings.TrimSpace(status)
	}

	// Get recent commits (last 5)
	if commits, err := g.runGitCommand(ctx, "log", "--oneline", "-5"); err == nil {
		info.RecentCommits = strings.Split(strings.TrimSpace(commits), "\n")
	}

	// Get remote URL
	if remote, err := g.runGitCommand(ctx, "config", "--get", "remote.origin.url"); err == nil {
		info.RemoteURL = strings.TrimSpace(remote)
	}

	return info, nil
}

// runGitCommand runs a git command and returns output
func (g *Gatherer) runGitCommand(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = g.workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", err
	}

	return stdout.String(), nil
}

// gatherEnv collects specified environment variables
func (g *Gatherer) gatherEnv(envVars []string, result *Context) {
	for _, envVar := range envVars {
		if value := os.Getenv(envVar); value != "" {
			result.Environment[envVar] = value
		}
	}
}

// estimateTokens provides a rough token count estimate
func (g *Gatherer) estimateTokens(ctx *Context) int {
	tokens := 0

	// Files: ~4 chars per token
	for _, content := range ctx.Files {
		tokens += len(content) / 4
	}

	// Git info: rough estimate
	if ctx.GitInfo != nil {
		tokens += 50 // branch, status, etc.
		tokens += len(ctx.GitInfo.RecentCommits) * 20
	}

	// Environment variables
	for k, v := range ctx.Environment {
		tokens += (len(k) + len(v)) / 4
	}

	return tokens
}

// truncateContext reduces context to fit within token limit
func (g *Gatherer) truncateContext(ctx *Context, maxTokens int) {
	currentTokens := ctx.TotalTokens

	// Strategy: remove files first (largest), starting with largest files
	if currentTokens > maxTokens && len(ctx.Files) > 0 {
		type fileSize struct {
			path string
			size int
		}

		var files []fileSize
		for path, content := range ctx.Files {
			files = append(files, fileSize{path, len(content)})
		}

		// Sort by size (largest first)
		for i := 0; i < len(files); i++ {
			for j := i + 1; j < len(files); j++ {
				if files[j].size > files[i].size {
					files[i], files[j] = files[j], files[i]
				}
			}
		}

		// Remove files until we're under limit
		for _, f := range files {
			if currentTokens <= maxTokens {
				break
			}
			tokensFreed := f.size / 4
			delete(ctx.Files, f.path)
			currentTokens -= tokensFreed
		}
	}

	ctx.TotalTokens = currentTokens
}

// Format formats context for inclusion in LLM prompts
func (ctx *Context) Format() string {
	var buf strings.Builder

	// Add file contents
	if len(ctx.Files) > 0 {
		buf.WriteString("## Files\n\n")
		for path, content := range ctx.Files {
			buf.WriteString(fmt.Sprintf("### %s\n", path))
			buf.WriteString("```\n")
			buf.WriteString(content)
			buf.WriteString("\n```\n\n")
		}
	}

	// Add git info
	if ctx.GitInfo != nil {
		buf.WriteString("## Git Context\n\n")
		if ctx.GitInfo.Branch != "" {
			buf.WriteString(fmt.Sprintf("**Branch:** %s\n", ctx.GitInfo.Branch))
		}
		if ctx.GitInfo.Status != "" {
			buf.WriteString(fmt.Sprintf("\n**Status:**\n```\n%s\n```\n", ctx.GitInfo.Status))
		}
		if len(ctx.GitInfo.RecentCommits) > 0 {
			buf.WriteString("\n**Recent Commits:**\n```\n")
			for _, commit := range ctx.GitInfo.RecentCommits {
				buf.WriteString(commit + "\n")
			}
			buf.WriteString("```\n")
		}
		buf.WriteString("\n")
	}

	// Add environment variables
	if len(ctx.Environment) > 0 {
		buf.WriteString("## Environment\n\n")
		for key, value := range ctx.Environment {
			buf.WriteString(fmt.Sprintf("- `%s`: %s\n", key, value))
		}
		buf.WriteString("\n")
	}

	return buf.String()
}
