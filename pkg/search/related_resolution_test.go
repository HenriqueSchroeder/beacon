package search

import (
	"context"
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
