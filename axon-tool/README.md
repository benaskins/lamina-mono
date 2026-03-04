# axon-tool

Primitives for defining and executing tools that can be used by LLM-powered agents.

Provider-agnostic — no dependency on any specific LLM backend.

## Install

```
go get github.com/benaskins/axon-tool@latest
```

Requires Go 1.24+.

## Usage

```go
tools := map[string]tool.ToolDef{
    "current_time":  tool.CurrentTimeTool(),
    "web_search":    tool.WebSearchTool(searcher),
    "fetch_page":    tool.FetchPageTool(fetcher),
    "check_weather": tool.CheckWeatherTool(weatherProvider),
}
```

### Key types

- `ToolDef` — tool definition with name, description, parameters, and execute function
- `ToolResult` — execution result (text, images, errors)
- `ToolContext` — request-scoped context (conversation ID, user ID)
- `ParameterSchema` — JSON Schema for tool parameters
- `Searcher`, `WeatherProvider` — integration interfaces
- `PageFetcher` — rate-limited web page fetcher with content extraction
- `ToolRouter` — LLM-based tool selection for routing requests

### Built-in tools

- `CurrentTimeTool()` — current date and time
- `WebSearchTool()` — web search via SearXNG
- `FetchPageTool()` — fetch and extract web page content
- `CheckWeatherTool()` — weather via Open-Meteo

## License

Apache 2.0 — see [LICENSE](LICENSE).
