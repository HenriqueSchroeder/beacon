package show

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/HenriqueSchroeder/beacon/pkg/vault"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestViewer_Show_ByRelativePath(t *testing.T) {
	viewer := NewViewer(newShowVault(t, map[string]string{
		filepath.Join("notes", "Target Note.md"): "# Target Note\nBody\n",
	}))

	output, err := viewer.Show(context.Background(), filepath.Join("notes", "Target Note.md"), Options{})

	require.NoError(t, err)
	assert.Equal(t, filepath.Join("notes", "Target Note.md"), output.Path)
	assert.Equal(t, "# Target Note\nBody\n", output.Content)
}

func TestViewer_Show_ByBasename(t *testing.T) {
	viewer := NewViewer(newShowVault(t, map[string]string{
		filepath.Join("notes", "Target Note.md"): "# Target Note\nBody\n",
	}))

	output, err := viewer.Show(context.Background(), "Target Note", Options{})

	require.NoError(t, err)
	assert.Equal(t, filepath.Join("notes", "Target Note.md"), output.Path)
}

func TestViewer_Show_ByTitleFallback(t *testing.T) {
	viewer := NewViewer(newShowVault(t, map[string]string{
		"target-note.md": "# Target Note\nBody\n",
	}))

	output, err := viewer.Show(context.Background(), "Target Note", Options{})

	require.NoError(t, err)
	assert.Equal(t, "target-note.md", output.Path)
}

func TestViewer_Show_WithOptionalExtension(t *testing.T) {
	viewer := NewViewer(newShowVault(t, map[string]string{
		"Target Note.md": "# Target Note\nBody\n",
	}))

	output, err := viewer.Show(context.Background(), "Target Note.md", Options{})

	require.NoError(t, err)
	assert.Equal(t, "Target Note.md", output.Path)
}

func TestViewer_Show_NoFrontmatter(t *testing.T) {
	viewer := NewViewer(newShowVault(t, map[string]string{
		"note.md": "---\ntags:\n  - test\n---\n# Note\nBody\n",
	}))

	output, err := viewer.Show(context.Background(), "note", Options{NoFrontmatter: true})

	require.NoError(t, err)
	assert.Equal(t, "# Note\nBody\n", output.Content)
}

func TestViewer_Show_NoFrontmatterWithInvalidFrontmatterReturnsOriginalContent(t *testing.T) {
	raw := "---\ntags: [broken\n---\n# Broken\n"
	viewer := NewViewer(newShowVault(t, map[string]string{
		"broken.md": raw,
	}))

	output, err := viewer.Show(context.Background(), "broken", Options{NoFrontmatter: true})

	require.NoError(t, err)
	assert.Equal(t, raw, output.Content)
}

func TestViewer_Show_PreservesCRLFLineEndings(t *testing.T) {
	raw := "---\r\ntags:\r\n  - test\r\n---\r\n\r\n# Note\r\nBody\r\n"
	viewer := NewViewer(newShowVault(t, map[string]string{
		"note.md": raw,
	}))

	output, err := viewer.Show(context.Background(), "note", Options{})

	require.NoError(t, err)
	assert.Equal(t, raw, output.Content)

	output, err = viewer.Show(context.Background(), "note", Options{NoFrontmatter: true})
	require.NoError(t, err)
	assert.Equal(t, "\r\n# Note\r\nBody\r\n", output.Content)
}

func TestViewer_Show_AmbiguousTarget(t *testing.T) {
	viewer := NewViewer(newShowVault(t, map[string]string{
		"Target.md":                     "# Target\n",
		filepath.Join("a", "Target.md"): "# Target\n",
	}))

	_, err := viewer.Show(context.Background(), "Target", Options{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ambiguous note target")
}

func TestViewer_Show_TargetNotFound(t *testing.T) {
	viewer := NewViewer(newShowVault(t, map[string]string{
		"Existing.md": "# Existing\n",
	}))

	_, err := viewer.Show(context.Background(), "Missing", Options{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "note target not found")
}

func TestViewer_Show_PrefersPathListingBeforeFullNoteParsing(t *testing.T) {
	viewer := NewViewer(showPathOnlyVault{
		paths: []string{"target-note.md"},
		notes: map[string]vault.Note{
			"target-note.md": {
				Path:       "target-note.md",
				Name:       "Target Note",
				RawContent: "# Target Note\n",
				Content:    "# Target Note\n",
			},
		},
	})

	output, err := viewer.Show(context.Background(), "target-note", Options{})

	require.NoError(t, err)
	assert.Equal(t, "target-note.md", output.Path)
	assert.Equal(t, "# Target Note\n", output.Content)
}

func TestViewer_Show_TitleLookupFallsBackAfterPathMiss(t *testing.T) {
	viewer := NewViewer(showTitleOnlyVault{
		paths: []string{},
		notes: []vault.Note{
			{
				Path:       "target-note.md",
				Name:       "Project Roadmap",
				RawContent: "# Project Roadmap\n",
				Content:    "# Project Roadmap\n",
			},
		},
	})

	output, err := viewer.Show(context.Background(), "Project Roadmap", Options{})

	require.NoError(t, err)
	assert.Equal(t, "target-note.md", output.Path)
	assert.Equal(t, "# Project Roadmap\n", output.Content)
}

func TestViewer_Show_SpacedBasenameStillUsesPathListing(t *testing.T) {
	viewer := NewViewer(showSpacedBasenamePathOnlyVault{
		paths: []string{"Project Roadmap.md"},
		notes: map[string]vault.Note{
			"Project Roadmap.md": {
				Path:       "Project Roadmap.md",
				Name:       "Different Title",
				RawContent: "# Different Title\n",
				Content:    "# Different Title\n",
			},
		},
	})

	output, err := viewer.Show(context.Background(), "Project Roadmap", Options{})

	require.NoError(t, err)
	assert.Equal(t, "Project Roadmap.md", output.Path)
	assert.Equal(t, "# Different Title\n", output.Content)
}

func TestViewer_Show_NoFrontmatterPlainMarkdownNote(t *testing.T) {
	viewer := NewViewer(newShowVault(t, map[string]string{
		"plain.md": "# Plain\nBody\n",
	}))

	output, err := viewer.Show(context.Background(), "plain", Options{NoFrontmatter: true})

	require.NoError(t, err)
	assert.Equal(t, "# Plain\nBody\n", output.Content)
}

func TestViewer_Show_DefaultOutputPreservesFrontmatterDelimiters(t *testing.T) {
	raw := "---\nstatus: active\n---\n# Note\n"
	viewer := NewViewer(newShowVault(t, map[string]string{
		"note.md": raw,
	}))

	output, err := viewer.Show(context.Background(), "note", Options{})

	require.NoError(t, err)
	assert.Equal(t, raw, output.Content)
}

func TestViewer_Show_AmbiguousBasenameAcrossDirectories(t *testing.T) {
	viewer := NewViewer(newShowVault(t, map[string]string{
		filepath.Join("a", "note.md"): "# One\n",
		filepath.Join("b", "note.md"): "# Two\n",
	}))

	_, err := viewer.Show(context.Background(), "note", Options{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ambiguous note target")
	assert.Contains(t, err.Error(), filepath.Join("a", "note.md"))
	assert.Contains(t, err.Error(), filepath.Join("b", "note.md"))
}

func TestViewer_Show_InvalidFrontmatterDoesNotTriggerHeuristicStripping(t *testing.T) {
	raw := "---\nstatus: [oops\n---\n# Note\nBody\n"
	viewer := NewViewer(newShowVault(t, map[string]string{
		"broken.md": raw,
	}))

	output, err := viewer.Show(context.Background(), "broken", Options{NoFrontmatter: true})

	require.NoError(t, err)
	assert.Equal(t, raw, output.Content)
}

func newShowVault(t *testing.T, notes map[string]string) vault.Vault {
	t.Helper()

	root := t.TempDir()
	for relPath, content := range notes {
		fullPath := filepath.Join(root, relPath)
		require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0o755))
		require.NoError(t, os.WriteFile(fullPath, []byte(content), 0o644))
	}

	v, err := vault.NewFileVault(root, nil)
	require.NoError(t, err)
	return v
}

type showPathOnlyVault struct {
	paths []string
	notes map[string]vault.Note
}

func (v showPathOnlyVault) ListNotes(context.Context) ([]vault.Note, error) {
	return nil, fmt.Errorf("ListNotes should not be used for path-based show resolution")
}

func (v showPathOnlyVault) GetNote(_ context.Context, path string) (*vault.Note, error) {
	note, ok := v.notes[path]
	if !ok {
		return nil, fmt.Errorf("note not found: %s", path)
	}
	return &note, nil
}

func (v showPathOnlyVault) ListNotePaths(context.Context) ([]string, error) {
	return v.paths, nil
}

type showTitleOnlyVault struct {
	paths []string
	notes []vault.Note
}

func (v showTitleOnlyVault) ListNotes(context.Context) ([]vault.Note, error) {
	return v.notes, nil
}

func (v showTitleOnlyVault) GetNote(_ context.Context, path string) (*vault.Note, error) {
	for _, note := range v.notes {
		if note.Path == path {
			return &note, nil
		}
	}

	return nil, fmt.Errorf("note not found: %s", path)
}

func (v showTitleOnlyVault) ListNotePaths(context.Context) ([]string, error) {
	return v.paths, nil
}

type showSpacedBasenamePathOnlyVault struct {
	paths []string
	notes map[string]vault.Note
}

func (v showSpacedBasenamePathOnlyVault) ListNotes(context.Context) ([]vault.Note, error) {
	return nil, fmt.Errorf("ListNotes should not be used for spaced basename resolution")
}

func (v showSpacedBasenamePathOnlyVault) GetNote(_ context.Context, path string) (*vault.Note, error) {
	note, ok := v.notes[path]
	if !ok {
		return nil, fmt.Errorf("note not found: %s", path)
	}
	return &note, nil
}

func (v showSpacedBasenamePathOnlyVault) ListNotePaths(context.Context) ([]string, error) {
	return v.paths, nil
}
