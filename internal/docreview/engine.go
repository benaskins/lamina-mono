package docreview

import (
	"context"
	"fmt"

	loop "github.com/benaskins/axon-loop"
	talk "github.com/benaskins/axon-talk"
	tool "github.com/benaskins/axon-tool"
)

const systemPrompt = `You are a documentation reviewer for Go libraries. A new version has just been released and you need to ensure the documentation is accurate.

You have tools to inspect the codebase: read files, list exported symbols, check go.mod, and view the git log between versions.

Review these files and update them if needed:

1. **README.md** — Check that:
   - Code examples use current function signatures (run list_exports to verify)
   - The "Key types" list matches what's actually exported
   - The Go version requirement matches what go.mod declares
   - Install instructions are correct

2. **AGENTS.md** — Check that:
   - The "Key files" list includes all significant source files (use list_files)
   - Architecture descriptions match the current code
   - Dependencies listed match go.mod
   - Build/test commands are correct

Start by reading the current README.md and AGENTS.md, then use list_exports and list_files to compare against reality. Use git_log to understand what changed in this release. Only write files that need updates — if documentation is already accurate, say so.

Be surgical: preserve existing structure, voice, and formatting. Only change what is factually wrong or missing.`

// Engine runs LLM-powered documentation review for a repository.
type Engine struct {
	client talk.LLMClient
	model  string
	tools  map[string]tool.ToolDef
}

// NewEngine creates a doc review engine for the given repo directory.
func NewEngine(client talk.LLMClient, model, dir string, writer Writer) *Engine {
	return &Engine{
		client: client,
		model:  model,
		tools:  Tools(dir, writer),
	}
}

// Review runs a documentation review. The onToken callback is called with
// each streamed token for real-time output.
func (e *Engine) Review(ctx context.Context, name, oldTag, newTag string, onToken func(string)) (*loop.Result, error) {
	toolDefs := make([]tool.ToolDef, 0, len(e.tools))
	for _, t := range e.tools {
		toolDefs = append(toolDefs, t)
	}

	req := &talk.Request{
		Model: e.model,
		Messages: []talk.Message{
			{Role: talk.RoleSystem, Content: systemPrompt},
			{Role: talk.RoleUser, Content: userMessage(name, oldTag, newTag)},
		},
		Tools:         toolDefs,
		Stream:        true,
		MaxIterations: 15,
	}

	cfg := loop.RunConfig{
		Client:  e.client,
		Request: req,
		Tools:   e.tools,
		ToolCtx: &tool.ToolContext{Ctx: ctx},
		Callbacks: loop.Callbacks{
			OnToken: onToken,
		},
	}
	return loop.Run(ctx, cfg)
}

func userMessage(name, oldTag, newTag string) string {
	if oldTag != "" {
		return fmt.Sprintf("Review documentation for %s after the %s release (previous version: %s). Check the git log for %s..%s to understand what changed, then verify README.md and AGENTS.md are accurate.", name, newTag, oldTag, oldTag, newTag)
	}
	return fmt.Sprintf("Review documentation for %s after the %s release (first release). Verify README.md and AGENTS.md are accurate.", name, newTag)
}
