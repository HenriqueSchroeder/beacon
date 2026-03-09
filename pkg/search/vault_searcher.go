package search

import (
	"context"
	"fmt"

	"github.com/HenriqueSchroeder/beacon/pkg/vault"
)

// VaultSearcher implements Searcher using the Vault interface for tag and type searches.
type VaultSearcher struct {
	vault vault.Vault
}

// NewVaultSearcher creates a new VaultSearcher backed by the given Vault.
func NewVaultSearcher(v vault.Vault) *VaultSearcher {
	return &VaultSearcher{vault: v}
}

// SearchTags returns notes that contain all of the specified tags.
func (s *VaultSearcher) SearchTags(ctx context.Context, tags []string) ([]SearchResult, error) {
	notes, err := s.vault.ListNotes(ctx)
	if err != nil {
		return nil, fmt.Errorf("search: failed to list notes: %w", err)
	}

	var results []SearchResult
	for _, note := range notes {
		if containsAllTags(note.Tags, tags) {
			results = append(results, noteToResult(note))
		}
	}

	return results, nil
}

// SearchByType returns notes whose frontmatter "type" field matches the given noteType.
func (s *VaultSearcher) SearchByType(ctx context.Context, noteType string) ([]SearchResult, error) {
	notes, err := s.vault.ListNotes(ctx)
	if err != nil {
		return nil, fmt.Errorf("search: failed to list notes: %w", err)
	}

	var results []SearchResult
	for _, note := range notes {
		if t, ok := note.Frontmatter["type"]; ok && t == noteType {
			results = append(results, noteToResult(note))
		}
	}

	return results, nil
}

// noteToResult converts a Note into a SearchResult.
func noteToResult(note vault.Note) SearchResult {
	return SearchResult{
		Path:  note.Path,
		Line:  0,
		Match: note.Name,
	}
}

// containsAllTags checks whether noteTags contains every tag in searchTags.
func containsAllTags(noteTags, searchTags []string) bool {
	tagSet := make(map[string]bool, len(noteTags))
	for _, t := range noteTags {
		tagSet[t] = true
	}
	for _, t := range searchTags {
		if !tagSet[t] {
			return false
		}
	}
	return true
}
