package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/scmd/scmd/internal/backend"
)

// Executor handles tool execution for LLM tool calling
type Executor struct {
	registry *Registry
	backend  backend.Backend
	maxRounds int
}

// NewExecutor creates a new tool executor
func NewExecutor(registry *Registry, backend backend.Backend) *Executor {
	return &Executor{
		registry: registry,
		backend:  backend,
		maxRounds: 5, // Max iterations to prevent infinite loops
	}
}

// ExecuteWithTools runs a completion with tool calling support
func (e *Executor) ExecuteWithTools(
	ctx context.Context,
	prompt string,
	systemPrompt string,
) (string, error) {
	if !e.backend.SupportsToolCalling() {
		// Fall back to regular completion
		resp, err := e.backend.Complete(ctx, &backend.CompletionRequest{
			Prompt:       prompt,
			SystemPrompt: systemPrompt,
		})
		if err != nil {
			return "", err
		}
		return resp.Content, nil
	}

	// Build tool request
	tools := e.registry.ToBackendTools()
	var conversationHistory []string
	currentPrompt := prompt

	for round := 0; round < e.maxRounds; round++ {
		// Call LLM with tools
		toolReq := &backend.ToolRequest{
			CompletionRequest: backend.CompletionRequest{
				Prompt:       currentPrompt,
				SystemPrompt: systemPrompt,
				MaxTokens:    2048,
				Temperature:  0.7,
			},
			Tools: tools,
		}

		resp, err := e.backend.CompleteWithTools(ctx, toolReq)
		if err != nil {
			return "", fmt.Errorf("tool calling failed: %w", err)
		}

		// Add response to history
		conversationHistory = append(conversationHistory, resp.Content)

		// If no tool calls, we're done
		if len(resp.ToolCalls) == 0 {
			return e.formatFinalResponse(conversationHistory), nil
		}

		// Execute tool calls
		toolResults := make([]string, len(resp.ToolCalls))
		for i, toolCall := range resp.ToolCalls {
			result, err := e.registry.Execute(ctx, toolCall.Name, toolCall.Parameters)
			if err != nil {
				toolResults[i] = fmt.Sprintf("Error executing %s: %v", toolCall.Name, err)
			} else if !result.Success {
				toolResults[i] = fmt.Sprintf("%s failed: %s", toolCall.Name, result.Error)
			} else {
				toolResults[i] = fmt.Sprintf("%s result:\n%s", toolCall.Name, result.Output)
			}
		}

		// Build next prompt with tool results
		currentPrompt = e.buildToolResultPrompt(resp.Content, toolResults)

		// If all tools succeeded and we have a final answer, return it
		if e.hasFinalAnswer(resp.Content) {
			return e.formatFinalResponse(conversationHistory), nil
		}
	}

	// Max rounds reached
	return e.formatFinalResponse(conversationHistory), nil
}

// buildToolResultPrompt creates a prompt with tool execution results
func (e *Executor) buildToolResultPrompt(llmResponse string, toolResults []string) string {
	var parts []string

	if llmResponse != "" {
		parts = append(parts, "Previous response:")
		parts = append(parts, llmResponse)
	}

	parts = append(parts, "\nTool execution results:")
	for _, result := range toolResults {
		parts = append(parts, result)
	}

	parts = append(parts, "\nBased on these results, provide your final answer.")

	return strings.Join(parts, "\n")
}

// hasFinalAnswer checks if the LLM response contains a final answer
func (e *Executor) hasFinalAnswer(content string) bool {
	// Simple heuristic: if response doesn't contain tool call markers
	// and has substantial content, consider it a final answer
	return !strings.Contains(content, "<tool_call>") && len(strings.TrimSpace(content)) > 50
}

// formatFinalResponse formats the conversation history into final output
func (e *Executor) formatFinalResponse(history []string) string {
	if len(history) == 0 {
		return ""
	}

	// Return the last response (most complete answer)
	lastResponse := history[len(history)-1]

	// Remove any tool call markers from final output
	lastResponse = strings.ReplaceAll(lastResponse, "<tool_call>", "")
	lastResponse = strings.ReplaceAll(lastResponse, "</tool_call>", "")

	return strings.TrimSpace(lastResponse)
}

// DefaultRegistry creates a registry with all built-in tools
func DefaultRegistry(confirmUI ConfirmUI) *Registry {
	registry := NewRegistry(confirmUI)

	// Register all built-in tools
	registry.Register(NewShellTool(confirmUI))
	registry.Register(NewReadFileTool())
	registry.Register(NewWriteFileTool(confirmUI))
	registry.Register(NewHTTPGetTool())

	return registry
}
