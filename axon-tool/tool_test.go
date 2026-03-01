package tool_test

import (
	"context"
	"testing"

	tool "github.com/benaskins/axon-tool"
)

func TestToolDefExecute(t *testing.T) {
	def := tool.ToolDef{
		Name:        "greet",
		Description: "Says hello",
		Execute: func(ctx *tool.ToolContext, args map[string]any) tool.ToolResult {
			name, _ := args["name"].(string)
			return tool.ToolResult{Content: "Hello, " + name + "!"}
		},
	}

	tc := &tool.ToolContext{
		Ctx:    context.Background(),
		UserID: "user-1",
	}

	result := def.Execute(tc, map[string]any{"name": "World"})

	if result.Content != "Hello, World!" {
		t.Errorf("got %q, want %q", result.Content, "Hello, World!")
	}
}

func TestToolContextCarriesMetadata(t *testing.T) {
	tc := &tool.ToolContext{
		Ctx:            context.Background(),
		UserID:         "user-1",
		Username:       "alice",
		AgentSlug:      "helper",
		ConversationID: "conv-1",
		SystemPrompt:   "You are helpful.",
	}

	if tc.UserID != "user-1" {
		t.Errorf("UserID = %q, want %q", tc.UserID, "user-1")
	}
	if tc.AgentSlug != "helper" {
		t.Errorf("AgentSlug = %q, want %q", tc.AgentSlug, "helper")
	}
}

func TestToolDefWithParameters(t *testing.T) {
	params := tool.ParameterSchema{
		Type:     "object",
		Required: []string{"query"},
		Properties: map[string]tool.PropertySchema{
			"query": {
				Type:        "string",
				Description: "The search query",
			},
		},
	}

	def := tool.ToolDef{
		Name:        "search",
		Description: "Search the web",
		Parameters:  params,
		Execute: func(ctx *tool.ToolContext, args map[string]any) tool.ToolResult {
			return tool.ToolResult{Content: "results"}
		},
	}

	if def.Parameters.Required[0] != "query" {
		t.Errorf("expected 'query' in required, got %v", def.Parameters.Required)
	}
	if def.Parameters.Properties["query"].Type != "string" {
		t.Errorf("expected string type for query property")
	}
}

// Verify interfaces are implementable with simple test doubles.

type stubSearcher struct {
	results []tool.SearchResult
}

func (s *stubSearcher) Search(_ context.Context, query string, limit int) ([]tool.SearchResult, error) {
	return s.results, nil
}

func TestSearcherInterface(t *testing.T) {
	s := &stubSearcher{
		results: []tool.SearchResult{
			{Title: "Go docs", URL: "https://go.dev", Snippet: "The Go programming language"},
		},
	}

	var searcher tool.Searcher = s
	results, err := searcher.Search(context.Background(), "go", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Title != "Go docs" {
		t.Errorf("got title %q, want %q", results[0].Title, "Go docs")
	}
}

type stubWeather struct {
	result *tool.WeatherResult
}

func (w *stubWeather) GetWeather(_ context.Context, location string) (*tool.WeatherResult, error) {
	return w.result, nil
}

func TestWeatherProviderInterface(t *testing.T) {
	w := &stubWeather{
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

	var provider tool.WeatherProvider = w
	result, err := provider.GetWeather(context.Background(), "Melbourne")
	if err != nil {
		t.Fatal(err)
	}
	if result.Location != "Melbourne" {
		t.Errorf("got location %q, want %q", result.Location, "Melbourne")
	}
	if result.Temperature != 22.5 {
		t.Errorf("got temp %f, want 22.5", result.Temperature)
	}
}

func TestTextGenerator(t *testing.T) {
	gen := tool.TextGenerator(func(ctx context.Context, prompt string) (string, error) {
		return "response to: " + prompt, nil
	})

	result, err := gen(context.Background(), "hello")
	if err != nil {
		t.Fatal(err)
	}
	if result != "response to: hello" {
		t.Errorf("got %q, want %q", result, "response to: hello")
	}
}
