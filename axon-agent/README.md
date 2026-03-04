# axon-agent

A provider-agnostic conversation loop for LLM-powered agents.

Handles message exchange, tool call dispatch, and streaming — with no HTTP, persistence, or UI concerns.

## Install

```
go get github.com/benaskins/axon-agent@latest
```

Requires Go 1.24+.

## Usage

Implement the `ChatClient` interface for your LLM backend, define tools, and run:

```go
result, err := agent.Run(ctx, agent.RunConfig{
    Client:   myLLMClient,
    Messages: messages,
    Tools:    toolDefs,
    Callbacks: agent.Callbacks{
        OnToken: func(token string) { fmt.Print(token) },
    },
})
```

### Key types

- `ChatClient` — interface abstracting communication with any LLM backend
- `ChatRequest` / `ChatResponse` — provider-agnostic request and streamed response
- `Message` — a single message in a conversation
- `ToolCall` — an LLM's decision to invoke a tool
- `Run()` — executes the agent conversation loop with tool dispatch

## License

Apache 2.0 — see [LICENSE](LICENSE).
