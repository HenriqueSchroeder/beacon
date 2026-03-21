package search

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/HenriqueSchroeder/beacon/pkg/vault"
)

// VaultSearcher implements Searcher using the Vault interface for tag and type searches.
type VaultSearcher struct {
	vault vault.Vault
}

type notePathLister interface {
	ListNotePaths(ctx context.Context) ([]string, error)
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

// SearchByFilename returns notes whose basename contains the normalized query.
func (s *VaultSearcher) SearchByFilename(ctx context.Context, query string) ([]SearchResult, error) {
	paths, err := s.listNotePaths(ctx)
	if err != nil {
		return nil, err
	}

	normalizedQuery := normalizeFilenameSearchTerm(query)
	var results []SearchResult
	for _, path := range paths {
		baseName := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		normalizedBaseName := normalizeFilenameSearchTerm(baseName)
		if strings.Contains(normalizedBaseName, normalizedQuery) {
			results = append(results, SearchResult{
				Path:  path,
				Line:  0,
				Match: baseName,
			})
		}
	}

	return results, nil
}

func (s *VaultSearcher) listNotePaths(ctx context.Context) ([]string, error) {
	if lister, ok := s.vault.(notePathLister); ok {
		paths, err := lister.ListNotePaths(ctx)
		if err != nil {
			return nil, fmt.Errorf("search: failed to list note paths: %w", err)
		}
		return paths, nil
	}

	notes, err := s.vault.ListNotes(ctx)
	if err != nil {
		return nil, fmt.Errorf("search: failed to list notes: %w", err)
	}

	paths := make([]string, 0, len(notes))
	for _, note := range notes {
		paths = append(paths, note.Path)
	}
	return paths, nil
}

func normalizeFilenameSearchTerm(value string) string {
	normalized := strings.ToLower(vault.SanitizeFilename(value))
	normalized = strings.ReplaceAll(normalized, "_", "")
	normalized = strings.ReplaceAll(normalized, "-", "")
	return normalized
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
