package command

import (
	"strings"
)

// Parser parses command arguments
type Parser struct{}

// NewParser creates a new argument parser
func NewParser() *Parser {
	return &Parser{}
}

// Parse parses a command string into Args
func (p *Parser) Parse(input string) *Args {
	args := NewArgs()
	args.Raw = input

	parts := tokenize(input)
	if len(parts) == 0 {
		return args
	}

	for i := 0; i < len(parts); i++ {
		part := parts[i]

		// Long option: --key=value or --key value
		if strings.HasPrefix(part, "--") {
			key := strings.TrimPrefix(part, "--")

			// Check for --key=value format
			if idx := strings.Index(key, "="); idx > 0 {
				args.Options[key[:idx]] = key[idx+1:]
				continue
			}

			// Check if next part is the value
			if i+1 < len(parts) && !strings.HasPrefix(parts[i+1], "-") {
				args.Options[key] = parts[i+1]
				i++
			} else {
				// It's a flag
				args.Flags[key] = true
			}
			continue
		}

		// Short flag: -f or -f value
		if strings.HasPrefix(part, "-") && len(part) > 1 {
			key := strings.TrimPrefix(part, "-")

			// Multiple flags: -abc
			if len(key) > 1 && !containsEquals(key) {
				for _, c := range key {
					args.Flags[string(c)] = true
				}
				continue
			}

			// Single flag with optional value
			if i+1 < len(parts) && !strings.HasPrefix(parts[i+1], "-") {
				args.Options[key] = parts[i+1]
				i++
			} else {
				args.Flags[key] = true
			}
			continue
		}

		// Positional argument
		args.Positional = append(args.Positional, part)
	}

	return args
}

// tokenize splits input into tokens, respecting quotes
func tokenize(input string) []string {
	var tokens []string
	var current strings.Builder
	inQuote := false
	quoteChar := rune(0)

	for _, r := range input {
		switch {
		case r == '"' || r == '\'':
			if inQuote && r == quoteChar {
				inQuote = false
				quoteChar = 0
			} else if !inQuote {
				inQuote = true
				quoteChar = r
			} else {
				current.WriteRune(r)
			}
		case r == ' ' || r == '\t':
			if inQuote {
				current.WriteRune(r)
			} else if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}

func containsEquals(s string) bool {
	return strings.Contains(s, "=")
}
