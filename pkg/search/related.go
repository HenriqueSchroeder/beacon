package search

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/HenriqueSchroeder/beacon/pkg/vault"
)

// ResolvedTarget is a canonical note target plus the wiki-link aliases that
// should be considered equivalent backlinks for that note.
type ResolvedTarget struct {
	Path    string
	Aliases []string
}

// ResolveRelatedTarget resolves a user-provided note reference to a single note.
func (s *VaultSearcher) ResolveRelatedTarget(ctx context.Context, query string) (ResolvedTarget, error) {
	normalizedQuery := normalizeRelatedValue(query)
	pathMatches, err := s.findPathMatches(ctx, normalizedQuery)
	if err != nil {
		return ResolvedTarget{}, err
	}
	switch len(pathMatches) {
	case 1:
		return s.resolvedTargetFromPath(ctx, pathMatches[0])
	case 0:
		// Fall back to full note parsing only when basename/path matching did not resolve.
	default:
		return ResolvedTarget{}, fmt.Errorf("ambiguous note target %q: %s", query, strings.Join(pathMatches, ", "))
	}

	notes, err := s.vault.ListNotes(ctx)
	if err != nil {
		return ResolvedTarget{}, fmt.Errorf("search: failed to list notes: %w", err)
	}

	matches := make(map[string]vault.Note)

	for _, note := range notes {
		for _, candidate := range relatedCandidates(note) {
			if normalizeRelatedValue(candidate) == normalizedQuery {
				matches[note.Path] = note
				break
			}
		}
	}

	switch len(matches) {
	case 0:
		return ResolvedTarget{}, fmt.Errorf("note target not found: %s", query)
	case 1:
		for _, note := range matches {
			return ResolvedTarget{
				Path:    note.Path,
				Aliases: relatedAliases(note),
			}, nil
		}
	}

	paths := make([]string, 0, len(matches))
	for path := range matches {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	return ResolvedTarget{}, fmt.Errorf("ambiguous note target %q: %s", query, strings.Join(paths, ", "))
}

func (s *VaultSearcher) findPathMatches(ctx context.Context, normalizedQuery string) ([]string, error) {
	paths, err := s.listNotePaths(ctx)
	if err != nil {
		return nil, err
	}

	var matches []string
	for _, path := range paths {
		if normalizeRelatedValue(path) == normalizedQuery {
			matches = append(matches, path)
			continue
		}

		baseWithoutExt := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		if normalizeRelatedValue(baseWithoutExt) == normalizedQuery {
			matches = append(matches, path)
		}
	}

	sort.Strings(matches)
	return matches, nil
}

func (s *VaultSearcher) resolvedTargetFromPath(ctx context.Context, path string) (ResolvedTarget, error) {
	note, err := s.vault.GetNote(ctx, path)
	if err != nil {
		return ResolvedTarget{}, fmt.Errorf("search: failed to load note %s: %w", path, err)
	}

	return ResolvedTarget{
		Path:    note.Path,
		Aliases: relatedAliases(*note),
	}, nil
}

func relatedCandidates(note vault.Note) []string {
	candidates := relatedAliases(note)
	if note.Name != "" {
		candidates = append(candidates, note.Name)
	}
	return dedupeStrings(candidates)
}

func relatedAliases(note vault.Note) []string {
	pathWithoutExt := strings.TrimSuffix(filepath.ToSlash(note.Path), filepath.Ext(note.Path))
	baseWithoutExt := strings.TrimSuffix(filepath.Base(note.Path), filepath.Ext(note.Path))
	aliases := []string{baseWithoutExt, pathWithoutExt}
	if note.Name != "" {
		aliases = append(aliases, note.Name)
	}
	return dedupeStrings(aliases)
}

func normalizeRelatedValue(value string) string {
	normalized := filepath.ToSlash(strings.TrimSpace(value))
	normalized = strings.TrimSuffix(normalized, ".md")
	return strings.ToLower(normalized)
}

func dedupeStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
