package cli

import (
	"fmt"
	"sort"
	"strings"
)

// ErrorType represents the type of CLI error
type ErrorType string

const (
	ErrorCommandNotFound ErrorType = "command_not_found"
	ErrorBackendNotFound ErrorType = "backend_not_found"
	ErrorNoBackend       ErrorType = "no_backend"
	ErrorInvalidInput    ErrorType = "invalid_input"
)

// CLIError represents a user-facing CLI error with helpful suggestions
type CLIError struct {
	Type        ErrorType
	Message     string
	Suggestions []string
}

// Error implements the error interface
func (e *CLIError) Error() string {
	var b strings.Builder

	// Main error message
	b.WriteString(fmt.Sprintf("❌ %s\n", e.Message))

	// Suggestions
	if len(e.Suggestions) > 0 {
		b.WriteString("\n💡 Suggestions:\n")
		for _, s := range e.Suggestions {
			b.WriteString(fmt.Sprintf("   %s\n", s))
		}
	}

	return b.String()
}

// NewCommandNotFoundError creates a helpful error for unknown commands
func NewCommandNotFoundError(cmdName string, availableCommands []string) error {
	suggestions := []string{}

	// Find similar commands using fuzzy matching
	similar := findSimilarCommands(cmdName, availableCommands, 3)
	if len(similar) > 0 {
		suggestions = append(suggestions, fmt.Sprintf("Did you mean: %s", strings.Join(similar, ", ")))
	}

	// Add helpful commands with specific keyword
	suggestions = append(suggestions,
		"Run 'scmd slash list' to see all installed commands",
		fmt.Sprintf("Run 'scmd repo search %s' to find this command in repositories", cmdName),
		"Run 'scmd repo add official <url>' to add the official repository",
		"Run 'scmd --help' to see built-in commands",
	)

	return &CLIError{
		Type:        ErrorCommandNotFound,
		Message:     fmt.Sprintf("Command '%s' not found", cmdName),
		Suggestions: suggestions,
	}
}

// NewBackendNotFoundError creates a helpful error for unknown backends
func NewBackendNotFoundError(backendName string, availableBackends []string) error {
	suggestions := []string{}

	// Find similar backends
	similar := findSimilarCommands(backendName, availableBackends, 2)
	if len(similar) > 0 {
		suggestions = append(suggestions, fmt.Sprintf("Did you mean: %s", strings.Join(similar, ", ")))
	}

	// Add available backends
	if len(availableBackends) > 0 {
		suggestions = append(suggestions, fmt.Sprintf("Available backends: %s", strings.Join(availableBackends, ", ")))
	}

	suggestions = append(suggestions, "Run 'scmd backends' to see all available backends")

	return &CLIError{
		Type:        ErrorBackendNotFound,
		Message:     fmt.Sprintf("Backend '%s' not found", backendName),
		Suggestions: suggestions,
	}
}

// NewNoBackendError creates a helpful error when no backend is available
func NewNoBackendError() error {
	return &CLIError{
		Type:    ErrorNoBackend,
		Message: "No backend available",
		Suggestions: []string{
			"Install llama.cpp: brew install llama.cpp",
			"Install Ollama: curl -fsSL https://ollama.com/install.sh | sh",
			"Set an API key: export OPENAI_API_KEY=your-key",
			"Set an API key: export GROQ_API_KEY=your-key (free tier available)",
			"Run 'scmd backends' to check backend status",
			"Run 'scmd doctor' to diagnose issues",
		},
	}
}

// findSimilarCommands finds commands similar to the input using fuzzy matching
// Returns up to maxResults commands sorted by similarity
func findSimilarCommands(input string, candidates []string, maxResults int) []string {
	if len(candidates) == 0 {
		return nil
	}

	type match struct {
		name  string
		score int
	}

	matches := []match{}
	input = strings.ToLower(input)

	for _, candidate := range candidates {
		score := similarity(input, strings.ToLower(candidate))
		// Only include if similarity is reasonably high
		if score > 50 {
			matches = append(matches, match{name: candidate, score: score})
		}
	}

	// Sort by score descending
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].score > matches[j].score
	})

	// Return top N results
	result := []string{}
	for i := 0; i < len(matches) && i < maxResults; i++ {
		result = append(result, matches[i].name)
	}

	return result
}

// similarity calculates a similarity score between two strings (0-100)
// Uses a combination of:
// - Exact match (100)
// - Prefix match (80-90)
// - Substring match (60-70)
// - Levenshtein distance (0-100)
func similarity(s1, s2 string) int {
	// Exact match
	if s1 == s2 {
		return 100
	}

	// Prefix match
	if strings.HasPrefix(s2, s1) || strings.HasPrefix(s1, s2) {
		shorter := len(s1)
		if len(s2) < shorter {
			shorter = len(s2)
		}
		longer := len(s1)
		if len(s2) > longer {
			longer = len(s2)
		}
		return 80 + (shorter * 20 / longer)
	}

	// Substring match
	if strings.Contains(s2, s1) || strings.Contains(s1, s2) {
		return 65
	}

	// Levenshtein distance
	dist := levenshteinDistance(s1, s2)
	maxLen := len(s1)
	if len(s2) > maxLen {
		maxLen = len(s2)
	}

	// Convert distance to similarity score
	if maxLen == 0 {
		return 0
	}

	similarity := 100 - (dist * 100 / maxLen)
	if similarity < 0 {
		similarity = 0
	}

	return similarity
}

// levenshteinDistance calculates the Levenshtein distance between two strings
func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Create matrix
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
	}

	// Initialize first row and column
	for i := 0; i <= len(s1); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}

	// Fill matrix
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 1
			if s1[i-1] == s2[j-1] {
				cost = 0
			}

			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
