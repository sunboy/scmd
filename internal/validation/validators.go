// Package validation provides input validation functions for security
package validation

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
)

var (
	// ErrInvalidCommandName indicates the command name contains invalid characters
	ErrInvalidCommandName = errors.New("invalid command name")

	// ErrInvalidURL indicates the URL is malformed or dangerous
	ErrInvalidURL = errors.New("invalid URL")

	// commandNamePattern matches valid command names (alphanumeric, dash, underscore)
	commandNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

	// Shell metacharacters that could enable command injection
	dangerousChars = []string{";", "|", "&", "$", "`", "(", ")", "{", "}", "<", ">", "\n", "\r", "\x00"}
)

// ValidateCommandName validates a slash command name
// Command names must:
// - Be 1-50 characters long
// - Contain only alphanumeric characters, dashes, and underscores
// - Not contain path separators (/, \, ..)
// - Not contain shell metacharacters
func ValidateCommandName(name string) error {
	if name == "" {
		return fmt.Errorf("%w: name cannot be empty", ErrInvalidCommandName)
	}

	// Check length
	if len(name) > 50 {
		return fmt.Errorf("%w: name too long (max 50 characters)", ErrInvalidCommandName)
	}

	// Reject path traversal patterns
	if strings.Contains(name, "..") {
		return fmt.Errorf("%w: cannot contain '..' (path traversal)", ErrInvalidCommandName)
	}

	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return fmt.Errorf("%w: cannot contain path separators", ErrInvalidCommandName)
	}

	// Reject shell metacharacters
	for _, char := range dangerousChars {
		if strings.Contains(name, char) {
			return fmt.Errorf("%w: cannot contain shell metacharacter '%s'", ErrInvalidCommandName, char)
		}
	}

	// Only allow alphanumeric, dash, underscore
	if !commandNamePattern.MatchString(name) {
		return fmt.Errorf("%w: must contain only letters, numbers, dashes, and underscores", ErrInvalidCommandName)
	}

	return nil
}

// ValidateRepoURL validates a repository URL
// URLs must:
// - Use http or https scheme only
// - Not point to localhost or private IPs (SSRF protection)
// - Not point to AWS metadata endpoint
// - Have a valid hostname
func ValidateRepoURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("%w: URL cannot be empty", ErrInvalidURL)
	}

	// Parse URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidURL, err)
	}

	// Validate scheme (only http/https allowed)
	scheme := strings.ToLower(parsedURL.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("%w: unsupported URL scheme '%s' (only http and https allowed)", ErrInvalidURL, parsedURL.Scheme)
	}

	// Get hostname
	hostname := parsedURL.Hostname()
	if hostname == "" {
		return fmt.Errorf("%w: missing hostname", ErrInvalidURL)
	}

	// SSRF Protection: Reject localhost
	if isLocalhost(hostname) {
		return fmt.Errorf("%w: localhost URLs not allowed (SSRF protection)", ErrInvalidURL)
	}

	// SSRF Protection: Reject private IPs
	if ip := net.ParseIP(hostname); ip != nil {
		if isPrivateIP(ip) {
			return fmt.Errorf("%w: private IP addresses not allowed (SSRF protection)", ErrInvalidURL)
		}

		// Reject link-local addresses (169.254.0.0/16)
		if isLinkLocal(ip) {
			return fmt.Errorf("%w: link-local addresses not allowed (SSRF protection)", ErrInvalidURL)
		}
	}

	// SSRF Protection: Reject AWS metadata endpoint specifically
	if hostname == "169.254.169.254" {
		return fmt.Errorf("%w: AWS metadata endpoint not allowed (SSRF protection)", ErrInvalidURL)
	}

	return nil
}

// isLocalhost checks if the hostname is localhost
func isLocalhost(hostname string) bool {
	hostname = strings.ToLower(hostname)
	return hostname == "localhost" ||
		hostname == "127.0.0.1" ||
		hostname == "::1" ||
		hostname == "0.0.0.0" ||
		hostname == "::" ||
		strings.HasSuffix(hostname, ".localhost")
}

// isPrivateIP checks if an IP is in a private range
// For security purposes, this includes:
// - RFC 1918 private ranges (10.x, 172.16-31.x, 192.168.x)
// - Loopback addresses (127.x)
// - Link-local addresses (169.254.x - including AWS metadata endpoint)
func isPrivateIP(ip net.IP) bool {
	// Check for IPv4 private ranges
	if ip4 := ip.To4(); ip4 != nil {
		// 10.0.0.0/8
		if ip4[0] == 10 {
			return true
		}

		// 172.16.0.0/12
		if ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31 {
			return true
		}

		// 192.168.0.0/16
		if ip4[0] == 192 && ip4[1] == 168 {
			return true
		}

		// 127.0.0.0/8 (loopback)
		if ip4[0] == 127 {
			return true
		}

		// 169.254.0.0/16 (link-local, includes AWS metadata endpoint)
		if ip4[0] == 169 && ip4[1] == 254 {
			return true
		}

		return false
	}

	// Check for IPv6 private ranges
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}

	// fc00::/7 (Unique Local Addresses)
	if len(ip) == 16 && (ip[0]&0xfe) == 0xfc {
		return true
	}

	return false
}

// isLinkLocal checks if an IP is in the link-local range (169.254.0.0/16)
func isLinkLocal(ip net.IP) bool {
	if ip4 := ip.To4(); ip4 != nil {
		return ip4[0] == 169 && ip4[1] == 254
	}
	return ip.IsLinkLocalUnicast()
}

// ValidateAliases validates a list of command aliases
func ValidateAliases(aliases []string) error {
	for _, alias := range aliases {
		if err := ValidateCommandName(alias); err != nil {
			return fmt.Errorf("invalid alias '%s': %w", alias, err)
		}
	}
	return nil
}
