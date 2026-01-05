// Package backend provides LLM backend interfaces and implementations
package backend

import (
	"context"
)

// Backend defines the LLM backend interface
type Backend interface {
	// Identity
	Name() string
	Type() Type

	// Lifecycle
	Initialize(ctx context.Context) error
	IsAvailable(ctx context.Context) (bool, error)
	Shutdown(ctx context.Context) error

	// Inference
	Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)
	Stream(ctx context.Context, req *CompletionRequest) (<-chan StreamChunk, error)

	// Tool calling (optional)
	SupportsToolCalling() bool
	CompleteWithTools(ctx context.Context, req *ToolRequest) (*ToolResponse, error)

	// Info
	ModelInfo() *ModelInfo
	EstimateTokens(text string) int
}

// Type identifies backend type
type Type string

const (
	TypeLocal  Type = "local"
	TypeOllama Type = "ollama"
	TypeClaude Type = "claude"
	TypeOpenAI Type = "openai"
	TypeMock   Type = "mock"
)

// CompletionRequest for inference
type CompletionRequest struct {
	Prompt        string
	SystemPrompt  string
	MaxTokens     int
	Temperature   float64
	StopSequences []string
}

// CompletionResponse from inference
type CompletionResponse struct {
	Content      string
	TokensUsed   int
	FinishReason FinishReason
	Timing       *Timing
}

// StreamChunk for streaming responses
type StreamChunk struct {
	Content string
	Done    bool
	Error   error
}

// FinishReason why generation stopped
type FinishReason string

const (
	FinishComplete FinishReason = "complete"
	FinishLength   FinishReason = "length"
	FinishStop     FinishReason = "stop"
	FinishError    FinishReason = "error"
)

// Timing information
type Timing struct {
	PromptMS     int64
	CompletionMS int64
	TokensPerSec float64
}

// ModelInfo describes the loaded model
type ModelInfo struct {
	Name          string
	Size          string
	Quantization  string
	ContextLength int
	Capabilities  []string
}

// ToolRequest for tool-calling inference
type ToolRequest struct {
	CompletionRequest
	Tools []ToolDefinition
}

// ToolDefinition defines a tool for the LLM
type ToolDefinition struct {
	Name        string
	Description string
	Parameters  map[string]ToolParameter
}

// ToolParameter defines a tool parameter
type ToolParameter struct {
	Type        string
	Description string
	Required    bool
	Enum        []string
}

// ToolResponse from tool-calling inference
type ToolResponse struct {
	Content   string
	ToolCalls []ToolCall
}

// ToolCall represents an LLM's request to call a tool
type ToolCall struct {
	Name       string
	Parameters map[string]interface{}
}
