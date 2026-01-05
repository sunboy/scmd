package version

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShort(t *testing.T) {
	result := Short()
	assert.Equal(t, Version, result)
}

func TestInfo(t *testing.T) {
	result := Info()
	assert.Contains(t, result, "scmd")
	assert.Contains(t, result, Version)
}

func TestFull(t *testing.T) {
	result := Full()
	assert.True(t, strings.HasPrefix(result, Version))
}

func TestShortCommit(t *testing.T) {
	// Test with long commit
	original := Commit
	Commit = "abc1234567890"
	result := shortCommit()
	assert.Equal(t, "abc1234", result)

	// Test with short commit
	Commit = "abc"
	result = shortCommit()
	assert.Equal(t, "abc", result)

	// Restore original
	Commit = original
}
