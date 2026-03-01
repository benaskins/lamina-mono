package tool

import "context"

// Searcher abstracts web search functionality.
type Searcher interface {
	Search(ctx context.Context, query string, limit int) ([]SearchResult, error)
}

// SearchResult represents a single search result.
type SearchResult struct {
	Title   string
	URL     string
	Snippet string
}
