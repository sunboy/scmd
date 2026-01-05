package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParser_Parse_Positional(t *testing.T) {
	p := NewParser()

	args := p.Parse("file.go src/main.go")
	assert.Equal(t, []string{"file.go", "src/main.go"}, args.Positional)
}

func TestParser_Parse_Flags(t *testing.T) {
	p := NewParser()

	args := p.Parse("--verbose -q")
	assert.True(t, args.Flags["verbose"])
	assert.True(t, args.Flags["q"])
}

func TestParser_Parse_Options(t *testing.T) {
	p := NewParser()

	args := p.Parse("--output=file.txt -f json")
	assert.Equal(t, "file.txt", args.Options["output"])
	assert.Equal(t, "json", args.Options["f"])
}

func TestParser_Parse_Mixed(t *testing.T) {
	p := NewParser()

	args := p.Parse("explain main.go --verbose -o output.txt")
	assert.Equal(t, []string{"explain", "main.go"}, args.Positional)
	assert.True(t, args.Flags["verbose"])
	assert.Equal(t, "output.txt", args.Options["o"])
}

func TestParser_Parse_QuotedStrings(t *testing.T) {
	p := NewParser()

	args := p.Parse(`explain "file with spaces.go" --msg "hello world"`)
	assert.Equal(t, []string{"explain", "file with spaces.go"}, args.Positional)
	assert.Equal(t, "hello world", args.Options["msg"])
}

func TestParser_Parse_SingleQuotes(t *testing.T) {
	p := NewParser()

	args := p.Parse(`explain 'single quoted'`)
	assert.Equal(t, []string{"explain", "single quoted"}, args.Positional)
}

func TestParser_Parse_Empty(t *testing.T) {
	p := NewParser()

	args := p.Parse("")
	assert.Empty(t, args.Positional)
	assert.Empty(t, args.Flags)
	assert.Empty(t, args.Options)
}

func TestParser_Parse_MultipleShortFlags(t *testing.T) {
	p := NewParser()

	args := p.Parse("-abc")
	assert.True(t, args.Flags["a"])
	assert.True(t, args.Flags["b"])
	assert.True(t, args.Flags["c"])
}

func TestParser_Parse_Raw(t *testing.T) {
	p := NewParser()

	input := "explain main.go --verbose"
	args := p.Parse(input)
	assert.Equal(t, input, args.Raw)
}

func TestArgs_HasFlag(t *testing.T) {
	args := NewArgs()
	args.Flags["verbose"] = true

	assert.True(t, args.HasFlag("verbose"))
	assert.False(t, args.HasFlag("quiet"))
}

func TestArgs_GetOption(t *testing.T) {
	args := NewArgs()
	args.Options["output"] = "file.txt"

	assert.Equal(t, "file.txt", args.GetOption("output"))
	assert.Equal(t, "", args.GetOption("nonexistent"))
}

func TestArgs_GetOptionOrDefault(t *testing.T) {
	args := NewArgs()
	args.Options["output"] = "file.txt"

	assert.Equal(t, "file.txt", args.GetOptionOrDefault("output", "default.txt"))
	assert.Equal(t, "default.txt", args.GetOptionOrDefault("nonexistent", "default.txt"))
}
