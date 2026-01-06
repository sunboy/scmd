package validation

import (
	"net"
	"testing"
)

func TestValidateCommandName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		// Valid names
		{"valid simple", "explain", false},
		{"valid with dash", "my-command", false},
		{"valid with underscore", "my_command", false},
		{"valid mixed", "cmd-name_123", false},
		{"valid single char", "e", false},
		{"valid numbers", "cmd123", false},

		// Invalid - path traversal
		{"path traversal 1", "../../../etc/passwd", true},
		{"path traversal 2", "cmd../bad", true},
		{"path traversal 3", "..cmd", true},
		{"path with slash", "path/to/cmd", true},
		{"path with backslash", "path\\to\\cmd", true},

		// Invalid - shell metacharacters
		{"semicolon", "test;rm -rf /", true},
		{"pipe", "test|whoami", true},
		{"ampersand", "test&", true},
		{"dollar", "test$var", true},
		{"backtick", "test`whoami`", true},
		{"command substitution", "test$(id)", true},
		{"parentheses", "test()", true},
		{"braces", "test{}", true},
		{"less than", "test<file", true},
		{"greater than", "test>file", true},
		{"newline", "test\nrm", true},
		{"null byte", "test\x00", true},

		// Invalid - other
		{"empty", "", true},
		{"too long", "this-is-a-very-long-command-name-that-exceeds-fifty-characters-limit", true},
		{"space", "my command", true},
		{"special chars", "cmd@home", true},
		{"dot", "cmd.txt", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCommandName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCommandName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateRepoURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		// Valid URLs
		{"valid http", "http://example.com/repo", false},
		{"valid https", "https://example.com/repo", false},
		{"valid with path", "https://github.com/user/repo", false},
		{"valid with port", "https://example.com:8080/repo", false},
		{"valid subdomain", "https://api.example.com/v1", false},

		// Invalid - dangerous schemes
		{"file scheme", "file:///etc/passwd", true},
		{"javascript scheme", "javascript:alert(1)", true},
		{"data scheme", "data:text/html,<script>alert(1)</script>", true},
		{"ftp scheme", "ftp://example.com", true},
		{"no scheme", "example.com", true},

		// Invalid - localhost
		{"localhost", "http://localhost:8080", true},
		{"127.0.0.1", "http://127.0.0.1", true},
		{"ipv6 localhost", "http://[::1]/repo", true},
		{"0.0.0.0", "http://0.0.0.0", true},
		{"subdomain localhost", "http://test.localhost", true},

		// Invalid - private IPs
		{"private 10.x", "http://10.0.0.1", true},
		{"private 192.168.x", "http://192.168.1.1", true},
		{"private 172.16.x", "http://172.16.0.1", true},
		{"private 172.31.x", "http://172.31.255.255", true},
		{"loopback", "http://127.1.1.1", true},

		// Invalid - link-local
		{"link-local", "http://169.254.1.1", true},
		{"aws metadata", "http://169.254.169.254/latest/meta-data", true},

		// Invalid - malformed
		{"empty", "", true},
		{"no hostname", "http://", true},
		{"invalid URL", "not a url", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRepoURL(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRepoURL(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateAliases(t *testing.T) {
	tests := []struct {
		name    string
		aliases []string
		wantErr bool
	}{
		{"valid aliases", []string{"e", "exp", "explain-cmd"}, false},
		{"empty list", []string{}, false},
		{"one invalid", []string{"e", "bad;cmd", "exp"}, true},
		{"path traversal", []string{"../../etc/passwd"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAliases(tt.aliases)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAliases(%v) error = %v, wantErr %v", tt.aliases, err, tt.wantErr)
			}
		})
	}
}

func TestIsPrivateIP(t *testing.T) {
	tests := []struct {
		ip      string
		want    bool
		comment string
	}{
		{"10.0.0.1", true, "10.x.x.x range"},
		{"10.255.255.255", true, "10.x.x.x range end"},
		{"172.16.0.1", true, "172.16.x.x range start"},
		{"172.31.255.255", true, "172.16-31.x.x range end"},
		{"172.15.0.1", false, "just outside 172.16.x.x"},
		{"172.32.0.1", false, "just outside 172.31.x.x"},
		{"192.168.0.1", true, "192.168.x.x range"},
		{"192.168.255.255", true, "192.168.x.x range end"},
		{"192.167.0.1", false, "just outside 192.168.x.x"},
		{"127.0.0.1", true, "loopback"},
		{"127.1.1.1", true, "loopback range"},
		{"8.8.8.8", false, "public IP (Google DNS)"},
		{"1.1.1.1", false, "public IP (Cloudflare DNS)"},
		{"169.254.169.254", true, "link-local (AWS metadata)"},
		{"::1", true, "IPv6 loopback"},
		{"fc00::1", true, "IPv6 unique local"},
	}

	for _, tt := range tests {
		t.Run(tt.comment, func(t *testing.T) {
			ip := parseIP(t, tt.ip)
			got := isPrivateIP(ip)
			if got != tt.want {
				t.Errorf("isPrivateIP(%s) = %v, want %v (%s)", tt.ip, got, tt.want, tt.comment)
			}
		})
	}
}

func TestIsLocalhost(t *testing.T) {
	tests := []struct {
		hostname string
		want     bool
	}{
		{"localhost", true},
		{"LOCALHOST", true},
		{"127.0.0.1", true},
		{"::1", true},
		{"0.0.0.0", true},
		{"::", true},
		{"test.localhost", true},
		{"api.localhost", true},
		{"example.com", false},
		{"google.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.hostname, func(t *testing.T) {
			got := isLocalhost(tt.hostname)
			if got != tt.want {
				t.Errorf("isLocalhost(%s) = %v, want %v", tt.hostname, got, tt.want)
			}
		})
	}
}

// Helper to parse IP for tests
func parseIP(t *testing.T, s string) net.IP {
	ip := net.ParseIP(s)
	if ip == nil {
		t.Fatalf("failed to parse IP: %s", s)
	}
	return ip
}
