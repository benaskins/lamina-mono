package tool

import (
	"fmt"
	"log/slog"
	"strings"
	"time"
)

// CurrentTimeTool returns a ToolDef that reports the current date and time.
func CurrentTimeTool() ToolDef {
	return ToolDef{
		Name:        "current_time",
		Description: "Get the current date and time. Use when the user asks what time it is, the current date, or anything requiring knowledge of the current moment.",
		Parameters: ParameterSchema{
			Type:       "object",
			Required:   []string{},
			Properties: map[string]PropertySchema{},
		},
		Execute: func(ctx *ToolContext, args map[string]any) ToolResult {
			now := time.Now()
			return ToolResult{Content: fmt.Sprintf("Current time: %s", now.Format("Monday, 2 January 2006 3:04 PM MST"))}
		},
	}
}

// WebSearchTool returns a ToolDef that searches the web using the provided Searcher.
func WebSearchTool(searcher Searcher) ToolDef {
	return ToolDef{
		Name:        "web_search",
		Description: "Search the web for current information, news, facts, or anything not in your training data. Use when the user asks about recent events, specific facts you're unsure about, or when they explicitly ask you to look something up.",
		Parameters: ParameterSchema{
			Type:     "object",
			Required: []string{"query"},
			Properties: map[string]PropertySchema{
				"query": {
					Type:        "string",
					Description: "The search query to look up on the web.",
				},
			},
		},
		Execute: func(ctx *ToolContext, args map[string]any) ToolResult {
			query, _ := args["query"].(string)
			if searcher == nil || query == "" {
				return ToolResult{Content: "Search is not available."}
			}

			results, err := searcher.Search(ctx.Ctx, query, 5)
			if err != nil {
				slog.Error("web search failed", "error", err, "query", query)
				return ToolResult{Content: fmt.Sprintf("Search failed: %v", err)}
			}
			if len(results) == 0 {
				return ToolResult{Content: fmt.Sprintf("No results found for %q.", query)}
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("Search results for %q:\n\n", query))
			for i, r := range results {
				sb.WriteString(fmt.Sprintf("%d. %s\n   URL: %s\n   %s\n\n", i+1, r.Title, r.URL, r.Snippet))
			}
			return ToolResult{Content: sb.String()}
		},
	}
}

// FetchPageTool returns a ToolDef that fetches and extracts content from web pages.
func FetchPageTool(fetcher *PageFetcher) ToolDef {
	return ToolDef{
		Name:        "fetch_page",
		Description: "Fetch a web page and extract content relevant to a specific question. Use after web_search to read promising results in detail. Returns a focused summary of the page's relevant content.",
		Parameters: ParameterSchema{
			Type:     "object",
			Required: []string{"url", "question"},
			Properties: map[string]PropertySchema{
				"url": {
					Type:        "string",
					Description: "The URL of the web page to fetch and read.",
				},
				"question": {
					Type:        "string",
					Description: "What you are looking for on this page. Guides extraction of relevant content.",
				},
			},
		},
		Execute: func(ctx *ToolContext, args map[string]any) ToolResult {
			if fetcher == nil {
				return ToolResult{Content: "Page fetching is not available."}
			}

			urlStr, _ := args["url"].(string)
			question, _ := args["question"].(string)

			if urlStr == "" {
				return ToolResult{Content: "Error: url is required."}
			}
			if question == "" {
				return ToolResult{Content: "Error: question is required."}
			}

			result, err := fetcher.FetchAndExtract(ctx.Ctx, urlStr, question)
			if err != nil {
				slog.Error("fetch_page failed", "error", err, "url", urlStr)
				return ToolResult{Content: err.Error()}
			}

			return ToolResult{Content: result}
		},
	}
}

// CheckWeatherTool returns a ToolDef that checks weather using the provided WeatherProvider.
func CheckWeatherTool(provider WeatherProvider) ToolDef {
	return ToolDef{
		Name:        "check_weather",
		Description: "Check the current weather conditions for a given location. Use when the user asks about the weather, temperature, or conditions somewhere.",
		Parameters: ParameterSchema{
			Type:     "object",
			Required: []string{"location"},
			Properties: map[string]PropertySchema{
				"location": {
					Type:        "string",
					Description: "The city or location to check weather for, e.g. 'Melbourne', 'Tokyo', 'New York'.",
				},
			},
		},
		Execute: func(ctx *ToolContext, args map[string]any) ToolResult {
			if provider == nil {
				return ToolResult{Content: "Weather checking is not available."}
			}

			location, _ := args["location"].(string)
			if location == "" {
				return ToolResult{Content: "Error: location is required."}
			}

			result, err := provider.GetWeather(ctx.Ctx, location)
			if err != nil {
				slog.Error("weather lookup failed", "error", err, "location", location)
				return ToolResult{Content: fmt.Sprintf("Weather lookup failed: %v", err)}
			}

			dayNight := "Nighttime"
			if result.IsDay {
				dayNight = "Daytime"
			}

			return ToolResult{Content: fmt.Sprintf("Weather for %s:\nConditions: %s\nTemperature: %.1f°C (feels like %.1f°C)\nHumidity: %d%%\nWind: %.1f km/h\nTime of day: %s",
				result.Location, result.Description, result.Temperature, result.FeelsLike,
				result.Humidity, result.WindSpeed, dayNight)}
		},
	}
}
