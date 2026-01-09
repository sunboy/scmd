// Package manpage provides utilities for reading and parsing man pages
package manpage

import (
	"fmt"
	"os/exec"
	"strings"
)

// ManPage represents a parsed man page
type ManPage struct {
	Command     string
	Name        string
	Synopsis    string
	Description string
	Options     string
	Examples    string
	FullText    string
}

// Read reads and parses a man page for the given command
func Read(command string) (*ManPage, error) {
	// Read the man page using 'man' command
	cmd := exec.Command("man", command)
	output, err := cmd.Output()
	if err != nil {
		// Man page doesn't exist or man command failed
		return nil, fmt.Errorf("man page not found for '%s': %w", command, err)
	}

	fullText := string(output)

	// Parse sections
	mp := &ManPage{
		Command:  command,
		FullText: fullText,
	}

	mp.Name = extractSection(fullText, "NAME")
	mp.Synopsis = extractSection(fullText, "SYNOPSIS")
	mp.Description = extractSection(fullText, "DESCRIPTION")
	mp.Options = extractSection(fullText, "OPTIONS")
	mp.Examples = extractSection(fullText, "EXAMPLES", "EXAMPLE")

	return mp, nil
}

// ReadMultiple reads man pages for multiple commands
func ReadMultiple(commands []string) map[string]*ManPage {
	result := make(map[string]*ManPage)

	for _, cmd := range commands {
		mp, err := Read(cmd)
		if err != nil {
			// Skip commands without man pages
			continue
		}
		result[cmd] = mp
	}

	return result
}

// extractSection extracts a section from a man page
// Supports multiple section names (e.g., "EXAMPLES" or "EXAMPLE")
func extractSection(text string, sectionNames ...string) string {
	lines := strings.Split(text, "\n")
	var sectionLines []string
	inSection := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check if we're starting the desired section
		for _, name := range sectionNames {
			if trimmed == name || strings.HasPrefix(trimmed, name) {
				inSection = true
				break
			}
		}

		// If in section, collect lines until next section
		if inSection {
			// Stop at next section (all caps line)
			if len(trimmed) > 0 && isAllCaps(trimmed) && !startsWithSectionName(trimmed, sectionNames) {
				break
			}

			// Skip the section header itself
			skipHeader := false
			for _, name := range sectionNames {
				if trimmed == name || strings.HasPrefix(trimmed, name) {
					skipHeader = true
					break
				}
			}

			if !skipHeader && len(trimmed) > 0 {
				sectionLines = append(sectionLines, line)
			}
		}
	}

	result := strings.Join(sectionLines, "\n")
	// Limit to first 100 lines to avoid overwhelming context
	lines = strings.Split(result, "\n")
	if len(lines) > 100 {
		lines = lines[:100]
		result = strings.Join(lines, "\n") + "\n... (truncated)"
	}

	return strings.TrimSpace(result)
}

// isAllCaps checks if a string is all uppercase (section headers)
func isAllCaps(s string) bool {
	if len(s) == 0 {
		return false
	}

	for _, r := range s {
		if r >= 'a' && r <= 'z' {
			return false
		}
	}

	// Must contain at least one letter
	for _, r := range s {
		if (r >= 'A' && r <= 'Z') {
			return true
		}
	}

	return false
}

// startsWithSectionName checks if a line starts with any of the section names
func startsWithSectionName(line string, names []string) bool {
	for _, name := range names {
		if strings.HasPrefix(line, name) {
			return true
		}
	}
	return false
}

// FormatForLLM formats man pages for LLM context
func FormatForLLM(manPages map[string]*ManPage) string {
	if len(manPages) == 0 {
		return ""
	}

	var sections []string

	for cmd, mp := range manPages {
		var cmdSections []string
		cmdSections = append(cmdSections, fmt.Sprintf("=== MAN PAGE: %s ===\n", cmd))

		if mp.Name != "" {
			cmdSections = append(cmdSections, fmt.Sprintf("NAME:\n%s\n", mp.Name))
		}

		if mp.Synopsis != "" {
			cmdSections = append(cmdSections, fmt.Sprintf("SYNOPSIS:\n%s\n", mp.Synopsis))
		}

		if mp.Description != "" {
			cmdSections = append(cmdSections, fmt.Sprintf("DESCRIPTION:\n%s\n", mp.Description))
		}

		if mp.Options != "" {
			cmdSections = append(cmdSections, fmt.Sprintf("OPTIONS:\n%s\n", mp.Options))
		}

		if mp.Examples != "" {
			cmdSections = append(cmdSections, fmt.Sprintf("EXAMPLES:\n%s\n", mp.Examples))
		}

		sections = append(sections, strings.Join(cmdSections, "\n"))
	}

	return strings.Join(sections, "\n\n")
}
