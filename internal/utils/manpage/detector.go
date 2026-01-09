package manpage

import (
	"strings"
)

// CommonCommands lists frequently used CLI commands
var CommonCommands = []string{
	// File operations
	"find", "ls", "cp", "mv", "rm", "mkdir", "touch", "cat", "less", "more",
	"head", "tail", "grep", "sed", "awk", "sort", "uniq", "wc", "diff",
	"tar", "gzip", "gunzip", "zip", "unzip",

	// File permissions
	"chmod", "chown", "chgrp",

	// Process management
	"ps", "top", "kill", "pkill", "killall", "htop", "nice", "renice",

	// Network
	"curl", "wget", "ping", "netstat", "ss", "ip", "ifconfig", "route",
	"nslookup", "dig", "traceroute", "ssh", "scp", "rsync",

	// Git
	"git",

	// Docker
	"docker", "docker-compose",

	// Kubernetes
	"kubectl", "helm",

	// System
	"systemctl", "service", "journalctl", "dmesg", "uname", "df", "du",
	"free", "date", "cal", "uptime", "whoami", "hostname",

	// Text processing
	"cut", "paste", "join", "tr", "expand", "unexpand", "fmt",

	// Archiving
	"xargs", "parallel",
}

// DetectCommands detects which commands might be relevant to a query
func DetectCommands(query string) []string {
	query = strings.ToLower(query)
	words := strings.Fields(query)

	detected := make(map[string]bool)
	var result []string

	// Check for exact command names in query
	for _, cmd := range CommonCommands {
		// Check if command appears as a word
		for _, word := range words {
			cleaned := strings.Trim(word, ".,;:!?()")
			if cleaned == cmd {
				if !detected[cmd] {
					detected[cmd] = true
					result = append(result, cmd)
				}
			}
		}

		// Also check for mentions of the command
		if strings.Contains(query, " "+cmd+" ") ||
			strings.HasPrefix(query, cmd+" ") ||
			strings.HasSuffix(query, " "+cmd) ||
			query == cmd {
			if !detected[cmd] {
				detected[cmd] = true
				result = append(result, cmd)
			}
		}
	}

	// Keyword-based detection for common use cases
	keywordMap := map[string][]string{
		"find":   {"find", "search", "locate", "files", "directories"},
		"grep":   {"search", "pattern", "text", "match", "filter"},
		"sed":    {"replace", "substitute", "edit", "modify"},
		"awk":    {"column", "field", "extract", "process"},
		"tar":    {"archive", "compress", "extract", "unpack"},
		"chmod":  {"permission", "executable", "rights", "access"},
		"ps":     {"process", "running", "pid"},
		"kill":   {"terminate", "stop", "end"},
		"curl":   {"download", "http", "api", "request", "fetch"},
		"git":    {"commit", "push", "pull", "branch", "merge", "clone", "repository"},
		"docker": {"container", "image", "dockerfile"},
		"df":     {"disk space", "storage", "capacity"},
		"du":     {"directory size", "folder size"},
	}

	for cmd, keywords := range keywordMap {
		if detected[cmd] {
			continue // Already detected
		}

		for _, keyword := range keywords {
			if strings.Contains(query, keyword) {
				detected[cmd] = true
				result = append(result, cmd)
				break
			}
		}
	}

	// If no commands detected, try to infer from context
	if len(result) == 0 {
		result = inferFromContext(query)
	}

	// Limit to top 3 most relevant commands to avoid overwhelming context
	if len(result) > 3 {
		result = result[:3]
	}

	return result
}

// inferFromContext tries to infer commands from query context
func inferFromContext(query string) []string {
	query = strings.ToLower(query)

	// File/directory operations
	if containsAny(query, []string{"file", "files", "directory", "folder", "path"}) {
		if containsAny(query, []string{"find", "search", "locate", "modified", "created"}) {
			return []string{"find"}
		}
		if containsAny(query, []string{"list", "show", "display"}) {
			return []string{"ls"}
		}
	}

	// Text processing
	if containsAny(query, []string{"text", "content", "line", "lines"}) {
		if containsAny(query, []string{"search", "match", "pattern", "filter"}) {
			return []string{"grep"}
		}
		if containsAny(query, []string{"replace", "substitute", "change"}) {
			return []string{"sed"}
		}
	}

	// Process management
	if containsAny(query, []string{"process", "running", "pid"}) {
		return []string{"ps"}
	}

	// Network operations
	if containsAny(query, []string{"download", "fetch", "http", "https", "api", "url"}) {
		return []string{"curl"}
	}

	// Default fallback - no specific command detected
	return nil
}

// containsAny checks if string contains any of the substrings
func containsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}
