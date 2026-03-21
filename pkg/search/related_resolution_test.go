package search

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

func TestResolveRelatedTarget_ByBasename(t *testing.T) {
	s := NewVaultSearcher(newRelatedSearchVault(t, map[string]string{
		"Target Note.md": "# Target Note\n",
	}))

	target, err := s.ResolveRelatedTarget(context.Background(), "Target Note")

	require.NoError(t, err)
	assert.Equal(t, "Target Note.md", target.Path)
	assert.ElementsMatch(t, []string{"Target Note"}, target.Aliases)
}

func TestResolveRelatedTarget_ByRelativePath(t *testing.T) {
	s := NewVaultSearcher(newRelatedSearchVault(t, map[string]string{
		filepath.Join("notes", "Target Note.md"): "# Target Note\n",
	}))

	target, err := s.ResolveRelatedTarget(context.Background(), filepath.Join("notes", "Target Note"))

	require.NoError(t, err)
	assert.Equal(t, filepath.Join("notes", "Target Note.md"), target.Path)
	assert.Contains(t, target.Aliases, filepath.Join("notes", "Target Note"))
}

func TestResolveRelatedTarget_WithOptionalExtension(t *testing.T) {
	s := NewVaultSearcher(newRelatedSearchVault(t, map[string]string{
		"Target Note.md": "# Target Note\n",
	}))

	target, err := s.ResolveRelatedTarget(context.Background(), "Target Note.md")

	require.NoError(t, err)
	assert.Equal(t, "Target Note.md", target.Path)
}

func TestResolveRelatedTarget_ByTitle(t *testing.T) {
	s := NewVaultSearcher(newRelatedSearchVault(t, map[string]string{
		"target-note.md": "# Target Note\n",
	}))

	target, err := s.ResolveRelatedTarget(context.Background(), "Target Note")

	require.NoError(t, err)
	assert.Equal(t, "target-note.md", target.Path)
	assert.Contains(t, target.Aliases, "Target Note")
}

func TestResolveRelatedTarget_Ambiguous(t *testing.T) {
	s := NewVaultSearcher(newRelatedSearchVault(t, map[string]string{
		"Target.md":                     "# Target\n",
		filepath.Join("a", "Target.md"): "# Target\n",
	}))

	_, err := s.ResolveRelatedTarget(context.Background(), "Target")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ambiguous note target")
	assert.Contains(t, err.Error(), "Target.md")
	assert.Contains(t, err.Error(), filepath.Join("a", "Target.md"))
}

func TestResolveRelatedTarget_NotFound(t *testing.T) {
	s := NewVaultSearcher(newRelatedSearchVault(t, map[string]string{
		"Existing.md": "# Existing\n",
	}))

	_, err := s.ResolveRelatedTarget(context.Background(), "Missing")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "note target not found")
}

func TestResolveRelatedTarget_PrefersPathListingBeforeFullNoteParsing(t *testing.T) {
	s := NewVaultSearcher(relatedPathOnlyVault{
		paths: []string{"target-note.md"},
		notes: map[string]vault.Note{
			"target-note.md": {
				Path: "target-note.md",
				Name: "Target Note",
			},
		},
	})

	target, err := s.ResolveRelatedTarget(context.Background(), "target-note")

	require.NoError(t, err)
	assert.Equal(t, "target-note.md", target.Path)
	assert.ElementsMatch(t, []string{"target-note", "Target Note"}, target.Aliases)
}

func newRelatedSearchVault(t *testing.T, notes map[string]string) vault.Vault {
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

type relatedPathOnlyVault struct {
	paths []string
	notes map[string]vault.Note
}

func (v relatedPathOnlyVault) ListNotes(context.Context) ([]vault.Note, error) {
	return nil, fmt.Errorf("ListNotes should not be used for path-based related resolution")
}

func (v relatedPathOnlyVault) GetNote(_ context.Context, path string) (*vault.Note, error) {
	note, ok := v.notes[path]
	if !ok {
		return nil, fmt.Errorf("note not found: %s", path)
	}
	return &note, nil
}

func (v relatedPathOnlyVault) ListNotePaths(context.Context) ([]string, error) {
	return v.paths, nil
}
