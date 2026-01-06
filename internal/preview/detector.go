// Package preview provides command preview and safety features
package preview

import (
	"regexp"
	"strings"
)

// DestructivePattern represents a potentially dangerous command pattern
type DestructivePattern struct {
	Pattern     *regexp.Regexp
	Severity    Severity
	Description string
	Examples    []string
}

// Severity levels for destructive operations
type Severity int

const (
	SeverityLow Severity = iota
	SeverityMedium
	SeverityHigh
	SeverityCritical
)

func (s Severity) String() string {
	switch s {
	case SeverityLow:
		return "low"
	case SeverityMedium:
		return "medium"
	case SeverityHigh:
		return "high"
	case SeverityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

func (s Severity) Icon() string {
	switch s {
	case SeverityLow:
		return "ℹ️"
	case SeverityMedium:
		return "⚠️"
	case SeverityHigh:
		return "🔥"
	case SeverityCritical:
		return "💀"
	default:
		return "❓"
	}
}

// DestructiveCommands contains patterns for dangerous operations
var DestructiveCommands = []DestructivePattern{
	// File deletion
	{
		Pattern:     regexp.MustCompile(`\brm\s+.*-rf?\b`),
		Severity:    SeverityCritical,
		Description: "Recursive file deletion - CANNOT BE UNDONE",
		Examples:    []string{"rm -rf /", "rm -rf *", "rm -r node_modules"},
	},
	{
		Pattern:     regexp.MustCompile(`\brm\s+[^-]`),
		Severity:    SeverityHigh,
		Description: "File deletion",
		Examples:    []string{"rm file.txt", "rm *.log"},
	},

	// Git operations
	{
		Pattern:     regexp.MustCompile(`git\s+push\s+.*--force`),
		Severity:    SeverityCritical,
		Description: "Force push - rewrites remote history",
		Examples:    []string{"git push --force", "git push -f origin main"},
	},
	{
		Pattern:     regexp.MustCompile(`git\s+reset\s+--hard`),
		Severity:    SeverityHigh,
		Description: "Hard reset - discards uncommitted changes",
		Examples:    []string{"git reset --hard HEAD", "git reset --hard origin/main"},
	},
	{
		Pattern:     regexp.MustCompile(`git\s+clean\s+-[fd]+`),
		Severity:    SeverityHigh,
		Description: "Removes untracked files",
		Examples:    []string{"git clean -fd", "git clean -fdx"},
	},
	{
		Pattern:     regexp.MustCompile(`git\s+branch\s+-D`),
		Severity:    SeverityMedium,
		Description: "Force delete branch",
		Examples:    []string{"git branch -D feature-branch"},
	},

	// Docker operations
	{
		Pattern:     regexp.MustCompile(`docker\s+(system\s+)?prune\s+.*-a`),
		Severity:    SeverityHigh,
		Description: "Remove all unused Docker resources",
		Examples:    []string{"docker system prune -a", "docker prune -a"},
	},
	{
		Pattern:     regexp.MustCompile(`docker\s+rm\s+.*-f`),
		Severity:    SeverityMedium,
		Description: "Force remove Docker containers",
		Examples:    []string{"docker rm -f container", "docker rm -fv container"},
	},
	{
		Pattern:     regexp.MustCompile(`docker\s+rmi\s+.*-f`),
		Severity:    SeverityMedium,
		Description: "Force remove Docker images",
		Examples:    []string{"docker rmi -f image"},
	},
	{
		Pattern:     regexp.MustCompile(`docker\s+volume\s+rm`),
		Severity:    SeverityHigh,
		Description: "Remove Docker volumes - DATA LOSS possible",
		Examples:    []string{"docker volume rm my-volume"},
	},

	// Kubernetes operations
	{
		Pattern:     regexp.MustCompile(`kubectl\s+delete`),
		Severity:    SeverityHigh,
		Description: "Delete Kubernetes resources",
		Examples:    []string{"kubectl delete pod", "kubectl delete deployment"},
	},

	// Database operations
	{
		Pattern:     regexp.MustCompile(`(DROP|TRUNCATE|DELETE\s+FROM).*\b(TABLE|DATABASE)\b`),
		Severity:    SeverityCritical,
		Description: "Database destruction - PERMANENT DATA LOSS",
		Examples:    []string{"DROP TABLE users", "TRUNCATE TABLE logs"},
	},

	// Process management
	{
		Pattern:     regexp.MustCompile(`kill\s+-9`),
		Severity:    SeverityMedium,
		Description: "Force kill process (SIGKILL)",
		Examples:    []string{"kill -9 1234", "killall -9 nginx"},
	},
	{
		Pattern:     regexp.MustCompile(`pkill|killall`),
		Severity:    SeverityMedium,
		Description: "Kill multiple processes",
		Examples:    []string{"pkill node", "killall python"},
	},

	// Disk operations
	{
		Pattern:     regexp.MustCompile(`dd\s+if=.*of=`),
		Severity:    SeverityCritical,
		Description: "Direct disk write - can destroy data/partitions",
		Examples:    []string{"dd if=/dev/zero of=/dev/sda"},
	},
	{
		Pattern:     regexp.MustCompile(`mkfs\.|format`),
		Severity:    SeverityCritical,
		Description: "Format filesystem - ERASES ALL DATA",
		Examples:    []string{"mkfs.ext4 /dev/sdb1"},
	},

	// Permission changes
	{
		Pattern:     regexp.MustCompile(`chmod\s+777`),
		Severity:    SeverityMedium,
		Description: "Makes file/directory world-writable (security risk)",
		Examples:    []string{"chmod 777 file.txt"},
	},
	{
		Pattern:     regexp.MustCompile(`chown\s+-R`),
		Severity:    SeverityMedium,
		Description: "Recursive ownership change",
		Examples:    []string{"chown -R user:group /"},
	},

	// Package management
	{
		Pattern:     regexp.MustCompile(`(npm|yarn|pip)\s+uninstall\s+.*--global`),
		Severity:    SeverityMedium,
		Description: "Remove global package",
		Examples:    []string{"npm uninstall -g package"},
	},
	{
		Pattern:     regexp.MustCompile(`(apt|yum|dnf)\s+remove`),
		Severity:    SeverityHigh,
		Description: "Remove system packages",
		Examples:    []string{"apt remove nginx"},
	},

	// System operations
	{
		Pattern:     regexp.MustCompile(`shutdown|reboot|poweroff`),
		Severity:    SeverityCritical,
		Description: "System shutdown/reboot",
		Examples:    []string{"shutdown now", "reboot"},
	},
	{
		Pattern:     regexp.MustCompile(`:\(\)\{\s*:\|:&\s*\};:`),
		Severity:    SeverityCritical,
		Description: "Fork bomb - will hang system",
		Examples:    []string{":(){ :|:& };:"},
	},
}

// DetectResult contains information about a detected destructive command
type DetectResult struct {
	IsDestructive bool
	Matches       []Match
	HighestSeverity Severity
}

// Match represents a single pattern match
type Match struct {
	Pattern     DestructivePattern
	MatchedText string
	Position    int
}

// Detect analyzes a command for destructive patterns
func Detect(command string) *DetectResult {
	result := &DetectResult{
		Matches: make([]Match, 0),
		HighestSeverity: SeverityLow,
	}

	// Normalize command
	cmd := strings.TrimSpace(command)

	// Check each pattern
	for _, pattern := range DestructiveCommands {
		if matches := pattern.Pattern.FindAllStringIndex(cmd, -1); len(matches) > 0 {
			result.IsDestructive = true

			for _, match := range matches {
				result.Matches = append(result.Matches, Match{
					Pattern:     pattern,
					MatchedText: cmd[match[0]:match[1]],
					Position:    match[0],
				})

				if pattern.Severity > result.HighestSeverity {
					result.HighestSeverity = pattern.Severity
				}
			}
		}
	}

	return result
}

// EstimateImpact tries to estimate the impact of a command
func EstimateImpact(command string) *Impact {
	impact := &Impact{
		Command: command,
	}

	// Estimate affected files
	if strings.Contains(command, "rm") {
		if strings.Contains(command, "-r") || strings.Contains(command, "-R") {
			impact.AffectedType = "directories (recursive)"
			impact.EstimatedCount = estimateFileCount(command)
		} else {
			impact.AffectedType = "files"
			impact.EstimatedCount = estimateFileCount(command)
		}
	} else if strings.Contains(command, "docker") && strings.Contains(command, "prune") {
		impact.AffectedType = "Docker resources"
		impact.EstimatedCount = -1 // Unknown
	} else if strings.Contains(command, "git push --force") {
		impact.AffectedType = "git commits (remote)"
		impact.EstimatedCount = -1
	}

	return impact
}

// Impact describes the estimated impact of a command
type Impact struct {
	Command        string
	AffectedType   string // "files", "directories", "containers", etc.
	EstimatedCount int    // -1 if unknown
	EstimatedSize  int64  // in bytes, -1 if unknown
}

func estimateFileCount(command string) int {
	// This is a simple heuristic
	// In practice, you'd want to actually check the filesystem
	if strings.Contains(command, "*") {
		return -1 // Unknown, could be many
	}
	if strings.Contains(command, "node_modules") {
		return 10000 // Typical node_modules size
	}
	return 1
}

// ShouldPreview determines if a command should require preview
func ShouldPreview(command string) bool {
	result := Detect(command)
	return result.IsDestructive && result.HighestSeverity >= SeverityMedium
}
