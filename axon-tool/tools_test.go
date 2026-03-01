package tool_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	tool "github.com/benaskins/axon-tool"
)

func TestCurrentTimeTool(t *testing.T) {
	def := tool.CurrentTimeTool()

	if def.Name != "current_time" {
		t.Errorf("Name = %q, want %q", def.Name, "current_time")
	}

	ctx := &tool.ToolContext{Ctx: context.Background()}
	result := def.Execute(ctx, nil)

	// Should contain the current year
	year := fmt.Sprintf("%d", time.Now().Year())
	if result.Content == "" {
		t.Error("expected non-empty content")
	}
	if !containsString(result.Content, year) {
		t.Errorf("expected content to contain year %s, got %q", year, result.Content)
	}
}

func TestWebSearchTool(t *testing.T) {
	called := false
	searcher := &stubSearcher{
		results: []tool.SearchResult{
			{Title: "Result 1", URL: "https://example.com", Snippet: "A snippet"},
		},
	}

	def := tool.WebSearchTool(searcher)

	if def.Name != "web_search" {
		t.Errorf("Name = %q, want %q", def.Name, "web_search")
	}

	ctx := &tool.ToolContext{Ctx: context.Background()}
	result := def.Execute(ctx, map[string]any{"query": "test query"})

	if result.Content == "" {
		t.Error("expected non-empty content")
	}
	if !containsString(result.Content, "Result 1") {
		t.Errorf("expected content to contain search result title, got %q", result.Content)
	}
	_ = called
}

func TestWebSearchToolNoSearcher(t *testing.T) {
	def := tool.WebSearchTool(nil)
	ctx := &tool.ToolContext{Ctx: context.Background()}
	result := def.Execute(ctx, map[string]any{"query": "test"})

	if !containsString(result.Content, "not available") {
		t.Errorf("expected 'not available' message, got %q", result.Content)
	}
}

func TestWebSearchToolEmptyQuery(t *testing.T) {
	searcher := &stubSearcher{}
	def := tool.WebSearchTool(searcher)
	ctx := &tool.ToolContext{Ctx: context.Background()}
	result := def.Execute(ctx, map[string]any{"query": ""})

	if !containsString(result.Content, "not available") {
		t.Errorf("expected 'not available' message for empty query, got %q", result.Content)
	}
}

func TestFetchPageTool(t *testing.T) {
	fetcher := tool.NewPageFetcher(nil)
	def := tool.FetchPageTool(fetcher)

	if def.Name != "fetch_page" {
		t.Errorf("Name = %q, want %q", def.Name, "fetch_page")
	}

	// Without a real URL we just test the nil/empty guards
	ctx := &tool.ToolContext{Ctx: context.Background()}

	result := def.Execute(ctx, map[string]any{})
	if !containsString(result.Content, "url is required") {
		t.Errorf("expected url required error, got %q", result.Content)
	}

	result = def.Execute(ctx, map[string]any{"url": "https://example.com"})
	if !containsString(result.Content, "question is required") {
		t.Errorf("expected question required error, got %q", result.Content)
	}
}

func TestFetchPageToolNilFetcher(t *testing.T) {
	def := tool.FetchPageTool(nil)
	ctx := &tool.ToolContext{Ctx: context.Background()}
	result := def.Execute(ctx, map[string]any{"url": "https://example.com", "question": "what?"})

	if !containsString(result.Content, "not available") {
		t.Errorf("expected 'not available' message, got %q", result.Content)
	}
}

func TestCheckWeatherTool(t *testing.T) {
	provider := &stubWeather{
		result: &tool.WeatherResult{
			Location:    "Melbourne",
			Description: "Clear sky",
			Temperature: 22.5,
			FeelsLike:   21.0,
			Humidity:    60,
			WindSpeed:   15.0,
			IsDay:       true,
		},
	}

	def := tool.CheckWeatherTool(provider)

	if def.Name != "check_weather" {
		t.Errorf("Name = %q, want %q", def.Name, "check_weather")
	}

	ctx := &tool.ToolContext{Ctx: context.Background()}
	result := def.Execute(ctx, map[string]any{"location": "Melbourne"})

	if !containsString(result.Content, "Melbourne") {
		t.Errorf("expected Melbourne in content, got %q", result.Content)
	}
	if !containsString(result.Content, "22.5") {
		t.Errorf("expected temperature in content, got %q", result.Content)
	}
}

func TestCheckWeatherToolNilProvider(t *testing.T) {
	def := tool.CheckWeatherTool(nil)
	ctx := &tool.ToolContext{Ctx: context.Background()}
	result := def.Execute(ctx, map[string]any{"location": "Melbourne"})

	if !containsString(result.Content, "not available") {
		t.Errorf("expected 'not available' message, got %q", result.Content)
	}
}

func TestCheckWeatherToolEmptyLocation(t *testing.T) {
	provider := &stubWeather{result: &tool.WeatherResult{}}
	def := tool.CheckWeatherTool(provider)
	ctx := &tool.ToolContext{Ctx: context.Background()}
	result := def.Execute(ctx, map[string]any{"location": ""})

	if !containsString(result.Content, "location is required") {
		t.Errorf("expected location required error, got %q", result.Content)
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
