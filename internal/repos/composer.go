package repos

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/scmd/scmd/internal/command"
)

// Composer handles command composition and chaining
type Composer struct {
	registry *command.Registry
	loader   *Loader
}

// NewComposer creates a new command composer
func NewComposer(registry *command.Registry, loader *Loader) *Composer {
	return &Composer{
		registry: registry,
		loader:   loader,
	}
}

// ExecuteComposed runs a composed command (pipeline, parallel, or fallback)
func (c *Composer) ExecuteComposed(
	ctx context.Context,
	spec *CommandSpec,
	args *command.Args,
	execCtx *command.ExecContext,
) (*command.Result, error) {
	if spec.Compose == nil {
		return nil, fmt.Errorf("command has no composition defined")
	}

	// Execute pipeline
	if len(spec.Compose.Pipeline) > 0 {
		return c.executePipeline(ctx, spec.Compose.Pipeline, args, execCtx)
	}

	// Execute parallel
	if len(spec.Compose.Parallel) > 0 {
		return c.executeParallel(ctx, spec.Compose.Parallel, args, execCtx)
	}

	// Execute fallback
	if len(spec.Compose.Fallback) > 0 {
		return c.executeFallback(ctx, spec.Compose.Fallback, args, execCtx)
	}

	return nil, fmt.Errorf("empty composition")
}

// executePipeline chains commands, passing output as input to next command
func (c *Composer) executePipeline(
	ctx context.Context,
	steps []PipelineStep,
	args *command.Args,
	execCtx *command.ExecContext,
) (*command.Result, error) {
	var lastOutput string

	// Get initial input from args
	if stdin, ok := args.Options["stdin"]; ok {
		lastOutput = stdin
	}

	for i, step := range steps {
		// Resolve command
		cmd, ok := c.registry.Get(step.Command)
		if !ok {
			// Try loading from installed plugins
			if err := c.loader.RegisterAll(c.registry); err == nil {
				cmd, ok = c.registry.Get(step.Command)
			}
		}

		if !ok {
			if step.OnError == "continue" {
				continue
			}
			return nil, fmt.Errorf("pipeline step %d: command '%s' not found", i, step.Command)
		}

		// Build args for this step
		stepArgs := command.NewArgs()
		stepArgs.Options["stdin"] = lastOutput

		// Apply step-specific args
		for k, v := range step.Args {
			stepArgs.Options[k] = v
		}

		// Execute step
		result, err := cmd.Execute(ctx, stepArgs, execCtx)
		if err != nil {
			if step.OnError == "continue" {
				continue
			}
			return nil, fmt.Errorf("pipeline step %d (%s): %w", i, step.Command, err)
		}

		if !result.Success {
			if step.OnError == "continue" {
				continue
			}
			return result, nil
		}

		// Apply transform if specified
		output := result.Output
		if step.Transform != "" {
			output = applyTransform(output, step.Transform)
		}

		lastOutput = output
	}

	return &command.Result{
		Success: true,
		Output:  lastOutput,
	}, nil
}

// executeParallel runs commands in parallel and merges results
func (c *Composer) executeParallel(
	ctx context.Context,
	commands []string,
	args *command.Args,
	execCtx *command.ExecContext,
) (*command.Result, error) {
	var wg sync.WaitGroup
	results := make([]*command.Result, len(commands))
	errors := make([]error, len(commands))

	for i, cmdName := range commands {
		wg.Add(1)
		go func(idx int, name string) {
			defer wg.Done()

			cmd, ok := c.registry.Get(name)
			if !ok {
				errors[idx] = fmt.Errorf("command '%s' not found", name)
				return
			}

			result, err := cmd.Execute(ctx, args, execCtx)
			if err != nil {
				errors[idx] = err
				return
			}
			results[idx] = result
		}(i, cmdName)
	}

	wg.Wait()

	// Merge results
	var outputs []string
	var allSuccess = true

	for i, result := range results {
		if errors[i] != nil {
			allSuccess = false
			outputs = append(outputs, fmt.Sprintf("[%s] Error: %v", commands[i], errors[i]))
		} else if result != nil {
			if !result.Success {
				allSuccess = false
			}
			outputs = append(outputs, fmt.Sprintf("## %s\n%s", commands[i], result.Output))
		}
	}

	return &command.Result{
		Success: allSuccess,
		Output:  strings.Join(outputs, "\n\n"),
	}, nil
}

// executeFallback tries commands in order until one succeeds
func (c *Composer) executeFallback(
	ctx context.Context,
	commands []string,
	args *command.Args,
	execCtx *command.ExecContext,
) (*command.Result, error) {
	var lastErr error

	for _, cmdName := range commands {
		cmd, ok := c.registry.Get(cmdName)
		if !ok {
			lastErr = fmt.Errorf("command '%s' not found", cmdName)
			continue
		}

		result, err := cmd.Execute(ctx, args, execCtx)
		if err != nil {
			lastErr = err
			continue
		}

		if result.Success {
			return result, nil
		}

		lastErr = fmt.Errorf("%s: %s", cmdName, result.Error)
	}

	return nil, fmt.Errorf("all fallback commands failed: %w", lastErr)
}

// applyTransform applies a simple transform to output
// Supports: trim, upper, lower, lines, first, last, json.field
func applyTransform(input, transform string) string {
	switch transform {
	case "trim":
		return strings.TrimSpace(input)
	case "upper":
		return strings.ToUpper(input)
	case "lower":
		return strings.ToLower(input)
	case "lines":
		// Return line count
		return fmt.Sprintf("%d", strings.Count(input, "\n")+1)
	case "first":
		// Return first line
		if idx := strings.Index(input, "\n"); idx >= 0 {
			return input[:idx]
		}
		return input
	case "last":
		// Return last line
		lines := strings.Split(input, "\n")
		for i := len(lines) - 1; i >= 0; i-- {
			if strings.TrimSpace(lines[i]) != "" {
				return lines[i]
			}
		}
		return input
	default:
		// Check for json.field pattern
		if strings.HasPrefix(transform, "json.") {
			// Simple JSON field extraction (basic implementation)
			field := transform[5:]
			// Look for "field": "value" pattern
			pattern := fmt.Sprintf(`"%s":\s*"([^"]*)"`, field)
			// This is a simplified version - production would use proper JSON parsing
			return extractPattern(input, pattern)
		}
		return input
	}
}

func extractPattern(input, pattern string) string {
	// Simple pattern matching for JSON-like extraction
	// In production, use proper regex or JSON parsing
	return input // Fallback
}

// ResolveDependencies resolves and installs command dependencies
func (c *Composer) ResolveDependencies(
	ctx context.Context,
	spec *CommandSpec,
	manager *Manager,
	installDir string,
) error {
	if len(spec.Dependencies) == 0 {
		return nil
	}

	for _, dep := range spec.Dependencies {
		// Check if already installed
		if cmd, ok := c.registry.Get(dep.Command); ok {
			_ = cmd // Already available
			continue
		}

		// Parse repo/command format
		parts := strings.SplitN(dep.Command, "/", 2)
		if len(parts) != 2 {
			if dep.Optional {
				continue
			}
			return fmt.Errorf("invalid dependency format: %s (expected repo/command)", dep.Command)
		}

		repoName, cmdName := parts[0], parts[1]

		// Get repo
		repo, ok := manager.Get(repoName)
		if !ok {
			if dep.Optional {
				continue
			}
			return fmt.Errorf("dependency repo '%s' not found", repoName)
		}

		// Fetch manifest
		manifest, err := manager.FetchManifest(ctx, repo)
		if err != nil {
			if dep.Optional {
				continue
			}
			return fmt.Errorf("fetch manifest for dependency: %w", err)
		}

		// Find command
		var cmdFile string
		for _, cmd := range manifest.Commands {
			if cmd.Name == cmdName {
				cmdFile = cmd.File
				break
			}
		}

		if cmdFile == "" {
			if dep.Optional {
				continue
			}
			return fmt.Errorf("dependency command '%s' not found in repo '%s'", cmdName, repoName)
		}

		// Fetch and install
		depSpec, err := manager.FetchCommand(ctx, repo, cmdFile)
		if err != nil {
			if dep.Optional {
				continue
			}
			return fmt.Errorf("fetch dependency command: %w", err)
		}

		// Check version constraint
		if dep.Version != "" && !checkVersionConstraint(depSpec.Version, dep.Version) {
			if dep.Optional {
				continue
			}
			return fmt.Errorf("dependency %s version %s does not satisfy %s",
				dep.Command, depSpec.Version, dep.Version)
		}

		// Install
		if err := manager.InstallCommand(depSpec, installDir); err != nil {
			if dep.Optional {
				continue
			}
			return fmt.Errorf("install dependency: %w", err)
		}

		// Register
		pluginCmd := NewPluginCommand(depSpec)
		_ = c.registry.Register(pluginCmd)
	}

	return nil
}

// checkVersionConstraint checks if version satisfies constraint
// Supports: =, >=, <=, >, <, ~> (pessimistic)
func checkVersionConstraint(version, constraint string) bool {
	if constraint == "" || constraint == "*" {
		return true
	}

	// Parse constraint
	var op string
	var target string

	if strings.HasPrefix(constraint, ">=") {
		op = ">="
		target = strings.TrimPrefix(constraint, ">=")
	} else if strings.HasPrefix(constraint, "<=") {
		op = "<="
		target = strings.TrimPrefix(constraint, "<=")
	} else if strings.HasPrefix(constraint, ">") {
		op = ">"
		target = strings.TrimPrefix(constraint, ">")
	} else if strings.HasPrefix(constraint, "<") {
		op = "<"
		target = strings.TrimPrefix(constraint, "<")
	} else if strings.HasPrefix(constraint, "~>") {
		op = "~>"
		target = strings.TrimPrefix(constraint, "~>")
	} else if strings.HasPrefix(constraint, "=") {
		op = "="
		target = strings.TrimPrefix(constraint, "=")
	} else {
		// Default to exact match
		op = "="
		target = constraint
	}

	target = strings.TrimSpace(target)

	// Compare versions (simplified semver comparison)
	cmp := compareVersions(version, target)

	switch op {
	case "=":
		return cmp == 0
	case ">=":
		return cmp >= 0
	case "<=":
		return cmp <= 0
	case ">":
		return cmp > 0
	case "<":
		return cmp < 0
	case "~>":
		// Pessimistic: >= target but < next major/minor
		return cmp >= 0 && isSameMajorMinor(version, target)
	default:
		return false
	}
}

// compareVersions compares two semver versions
// Returns: -1 if a < b, 0 if a == b, 1 if a > b
func compareVersions(a, b string) int {
	aParts := parseVersion(a)
	bParts := parseVersion(b)

	for i := 0; i < 3; i++ {
		if aParts[i] < bParts[i] {
			return -1
		}
		if aParts[i] > bParts[i] {
			return 1
		}
	}
	return 0
}

func parseVersion(v string) [3]int {
	var parts [3]int
	v = strings.TrimPrefix(v, "v")

	segments := strings.Split(v, ".")
	for i := 0; i < 3 && i < len(segments); i++ {
		// Parse only numeric part
		numStr := ""
		for _, c := range segments[i] {
			if c >= '0' && c <= '9' {
				numStr += string(c)
			} else {
				break
			}
		}
		if numStr != "" {
			var n int
			for _, c := range numStr {
				n = n*10 + int(c-'0')
			}
			parts[i] = n
		}
	}
	return parts
}

func isSameMajorMinor(a, b string) bool {
	aParts := parseVersion(a)
	bParts := parseVersion(b)
	return aParts[0] == bParts[0] && aParts[1] == bParts[1]
}

// ChainBuilder helps build command chains programmatically
type ChainBuilder struct {
	steps []PipelineStep
}

// NewChainBuilder creates a new chain builder
func NewChainBuilder() *ChainBuilder {
	return &ChainBuilder{}
}

// Add adds a step to the chain
func (b *ChainBuilder) Add(command string) *ChainBuilder {
	b.steps = append(b.steps, PipelineStep{Command: command})
	return b
}

// AddWithArgs adds a step with arguments
func (b *ChainBuilder) AddWithArgs(command string, args map[string]string) *ChainBuilder {
	b.steps = append(b.steps, PipelineStep{Command: command, Args: args})
	return b
}

// Transform adds a transform to the last step
func (b *ChainBuilder) Transform(transform string) *ChainBuilder {
	if len(b.steps) > 0 {
		b.steps[len(b.steps)-1].Transform = transform
	}
	return b
}

// OnError sets error handling for the last step
func (b *ChainBuilder) OnError(action string) *ChainBuilder {
	if len(b.steps) > 0 {
		b.steps[len(b.steps)-1].OnError = action
	}
	return b
}

// Build creates the ComposeSpec
func (b *ChainBuilder) Build() *ComposeSpec {
	return &ComposeSpec{Pipeline: b.steps}
}
