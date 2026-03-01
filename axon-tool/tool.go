// Package tool provides primitives for defining and executing tools
// that can be used by LLM-powered agents. It is provider-agnostic —
// no dependency on any specific LLM backend.
package tool

import "context"

// ToolDef describes a tool that an agent can invoke.
type ToolDef struct {
	Name        string
	Description string
	Parameters  ParameterSchema
	Execute     func(ctx *ToolContext, args map[string]any) ToolResult
}

// ToolResult is the output of a tool execution.
type ToolResult struct {
	Content string
}

// ToolContext carries request-scoped state into tool execution.
type ToolContext struct {
	Ctx            context.Context
	UserID         string
	Username       string
	AgentSlug      string
	ConversationID string
	SystemPrompt   string
}

// ParameterSchema describes the parameters a tool accepts.
type ParameterSchema struct {
	Type       string                    `json:"type"`
	Required   []string                  `json:"required,omitempty"`
	Properties map[string]PropertySchema `json:"properties,omitempty"`
}

// PropertySchema describes a single parameter property.
type PropertySchema struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

// TextGenerator is a simple function that sends a prompt to an LLM
// and returns the response text. Used by components that need LLM
// capabilities without depending on a full ChatClient.
type TextGenerator func(ctx context.Context, prompt string) (string, error)
