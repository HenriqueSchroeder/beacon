package show

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/HenriqueSchroeder/beacon/pkg/vault"
)

type notePathLister interface {
	ListNotePaths(ctx context.Context) ([]string, error)
}

type Options struct {
	NoFrontmatter bool
}

type Output struct {
	Path    string
	Content string
}

type Viewer struct {
	vault vault.Vault
}

func NewViewer(v vault.Vault) *Viewer {
	return &Viewer{vault: v}
}

func (v *Viewer) Show(ctx context.Context, query string, opts Options) (Output, error) {
	path, err := v.resolvePath(ctx, query)
	if err != nil {
		return Output{}, err
	}

	note, err := v.vault.GetNote(ctx, path)
	if err != nil {
		return Output{}, fmt.Errorf("show: failed to load note %s: %w", path, err)
	}

	content := note.RawContent
	if opts.NoFrontmatter {
		content = note.Content
	}

	return Output{
		Path:    note.Path,
		Content: content,
	}, nil
}

func (v *Viewer) resolvePath(ctx context.Context, query string) (string, error) {
	normalizedQuery := normalizeQuery(query)
	pathMatches, err := v.findPathMatches(ctx, normalizedQuery)
	if err != nil {
		return "", err
	}

	switch len(pathMatches) {
	case 1:
		return pathMatches[0], nil
	case 0:
	default:
		return "", fmt.Errorf("ambiguous note target %q: %s", query, strings.Join(pathMatches, ", "))
	}

	return v.resolveFromNotes(ctx, query, normalizedQuery)
}

func (v *Viewer) resolveFromNotes(ctx context.Context, query, normalizedQuery string) (string, error) {
	notes, err := v.vault.ListNotes(ctx)
	if err != nil {
		return "", fmt.Errorf("show: failed to list notes: %w", err)
	}

	basenameMatches := make(map[string]struct{})
	titleMatches := make(map[string]struct{})
	for _, note := range notes {
		baseWithoutExt := strings.TrimSuffix(filepath.Base(note.Path), filepath.Ext(note.Path))
		if normalizeQuery(baseWithoutExt) == normalizedQuery {
			basenameMatches[note.Path] = struct{}{}
		}

		if normalizeQuery(note.Name) == normalizedQuery {
			titleMatches[note.Path] = struct{}{}
		}
	}

	switch len(basenameMatches) {
	case 0:
	case 1:
		for path := range basenameMatches {
			return path, nil
		}
	default:
		return "", ambiguousTargetError(query, basenameMatches)
	}

	switch len(titleMatches) {
	case 0:
		return "", fmt.Errorf("note target not found: %s", query)
	case 1:
		for path := range titleMatches {
			return path, nil
		}
	default:
		return "", ambiguousTargetError(query, titleMatches)
	}

	return "", fmt.Errorf("note target not found: %s", query)
}

func (v *Viewer) findPathMatches(ctx context.Context, normalizedQuery string) ([]string, error) {
	lister, ok := v.vault.(notePathLister)
	if !ok {
		return nil, nil
	}

	paths, err := lister.ListNotePaths(ctx)
	if err != nil {
		return nil, fmt.Errorf("show: failed to list note paths: %w", err)
	}

	var matches []string
	for _, path := range paths {
		if normalizeQuery(path) == normalizedQuery {
			matches = append(matches, path)
			continue
		}

		baseWithoutExt := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		if normalizeQuery(baseWithoutExt) == normalizedQuery {
			matches = append(matches, path)
		}
	}

	sort.Strings(matches)
	return matches, nil
}

func normalizeQuery(value string) string {
	normalized := filepath.ToSlash(strings.TrimSpace(value))
	normalized = strings.TrimSuffix(normalized, ".md")
	return strings.ToLower(normalized)
}

func ambiguousTargetError(query string, matches map[string]struct{}) error {
	paths := make([]string, 0, len(matches))
	for path := range matches {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return fmt.Errorf("ambiguous note target %q: %s", query, strings.Join(paths, ", "))
}
