package vault

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/frontmatter"
	"gopkg.in/yaml.v3"
)

// FileVault implements the Vault interface using the local filesystem.
type FileVault struct {
	rootPath string
	ignore   []string
}

// NewFileVault creates a new FileVault rooted at the given path.
// It validates that the path exists and is a directory.
func NewFileVault(rootPath string, ignore []string) (*FileVault, error) {
	info, err := os.Stat(rootPath)
	if err != nil {
		return nil, fmt.Errorf("vault: invalid root path: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("vault: root path is not a directory: %s", rootPath)
	}

	return &FileVault{
		rootPath: rootPath,
		ignore:   ignore,
	}, nil
}

// ListNotes walks the filesystem for .md files and returns all notes found.
func (v *FileVault) ListNotes(ctx context.Context) ([]Note, error) {
	var notes []Note

	err := filepath.Walk(v.rootPath, func(fullPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if info.IsDir() {
			return nil
		}

		if filepath.Ext(fullPath) != ".md" {
			return nil
		}

		relPath, err := filepath.Rel(v.rootPath, fullPath)
		if err != nil {
			return fmt.Errorf("vault: failed to compute relative path: %w", err)
		}

		if v.shouldIgnore(relPath) {
			return nil
		}

		note, err := v.parseNote(fullPath, relPath, false)
		if err != nil {
			return fmt.Errorf("vault: failed to parse %s: %w", relPath, err)
		}

		notes = append(notes, *note)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return notes, nil
}

// ListNotePaths walks the filesystem for .md files and returns relative note paths.
func (v *FileVault) ListNotePaths(ctx context.Context) ([]string, error) {
	var paths []string

	err := filepath.Walk(v.rootPath, func(fullPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if info.IsDir() {
			return nil
		}

		if filepath.Ext(fullPath) != ".md" {
			return nil
		}

		relPath, err := filepath.Rel(v.rootPath, fullPath)
		if err != nil {
			return fmt.Errorf("vault: failed to compute relative path: %w", err)
		}

		if v.shouldIgnore(relPath) {
			return nil
		}

		paths = append(paths, relPath)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return paths, nil
}

// GetNote reads a specific file by its relative path.
func (v *FileVault) GetNote(ctx context.Context, path string) (*Note, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	fullPath := filepath.Join(v.rootPath, path)

	if _, err := os.Stat(fullPath); err != nil {
		return nil, fmt.Errorf("vault: note not found: %w", err)
	}

	return v.parseNote(fullPath, path, true)
}

// parseNote reads a file and parses its frontmatter and content.
func (v *FileVault) parseNote(fullPath, relPath string, includeRaw bool) (*Note, error) {
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("vault: failed to read file: %w", err)
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("vault: failed to stat file: %w", err)
	}

	var fm map[string]any
	formats := []*frontmatter.Format{
		frontmatter.NewFormat("---", "---", yaml.Unmarshal),
	}

	content, err := frontmatter.Parse(bytes.NewReader(data), &fm, formats...)
	if err != nil {
		// If frontmatter parsing fails, treat whole file as content
		fm = make(map[string]any)
		content = data
	}

	if fm == nil {
		fm = make(map[string]any)
	}

	contentStr := string(content)
	name := v.extractName(contentStr, relPath)
	tags := v.extractTags(fm)

	rawContent := ""
	if includeRaw {
		rawContent = string(data)
	}

	return &Note{
		Path:        relPath,
		Name:        name,
		RawContent:  rawContent,
		Content:     contentStr,
		Frontmatter: fm,
		Tags:        tags,
		Modified:    info.ModTime(),
	}, nil
}

// shouldIgnore checks whether a relative path matches any ignore pattern.
func (v *FileVault) shouldIgnore(relPath string) bool {
	for _, pattern := range v.ignore {
		// Check prefix matching (directory-level ignore)
		if strings.HasPrefix(relPath, pattern+string(filepath.Separator)) || relPath == pattern {
			return true
		}

		// Check filepath.Match against each path component and the full path
		if matched, _ := filepath.Match(pattern, relPath); matched {
			return true
		}

		// Check against individual path components
		parts := strings.Split(relPath, string(filepath.Separator))
		for _, part := range parts {
			if matched, _ := filepath.Match(pattern, part); matched {
				return true
			}
		}
	}
	return false
}

// extractName gets the title from the first # heading line, or falls back to filename.
func (v *FileVault) extractName(content, relPath string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if name, found := strings.CutPrefix(trimmed, "# "); found {
			return name
		}
	}

	// Fallback to filename without extension
	base := filepath.Base(relPath)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

// extractTags extracts the tags array from a frontmatter map.
func (v *FileVault) extractTags(fm map[string]any) []string {
	tagsRaw, ok := fm["tags"]
	if !ok {
		return nil
	}

	switch tags := tagsRaw.(type) {
	case []any:
		result := make([]string, 0, len(tags))
		for _, t := range tags {
			if s, ok := t.(string); ok {
				result = append(result, s)
			}
		}
		return result
	case []string:
		return tags
	default:
		return nil
	}
}
