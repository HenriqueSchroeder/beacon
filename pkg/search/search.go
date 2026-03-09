package search

import "context"

// SearchResult represents a single match found during content search.
type SearchResult struct {
	Path          string   `json:"path"`
	Line          int      `json:"line"`
	Match         string   `json:"match"`
	ContextBefore []string `json:"context_before"`
	ContextAfter  []string `json:"context_after"`
}

// Searcher defines the interface for searching note content.
type Searcher interface {
	SearchContent(ctx context.Context, query string) ([]SearchResult, error)
	SearchTags(ctx context.Context, tags []string) ([]SearchResult, error)
	SearchByType(ctx context.Context, noteType string) ([]SearchResult, error)
}
