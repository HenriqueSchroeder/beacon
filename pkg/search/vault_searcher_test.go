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

func newTestVault(t *testing.T) vault.Vault {
	t.Helper()
	v, err := vault.NewFileVault(fixturesPath, nil)
	require.NoError(t, err)
	return v
}

func TestVaultSearcher_SearchTags_SingleTag(t *testing.T) {
	s := NewVaultSearcher(newTestVault(t))

	results, err := s.SearchTags(context.Background(), []string{"golang"})

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "note1.md", results[0].Path)
	assert.Equal(t, "Go Tips", results[0].Match)
}

func TestVaultSearcher_SearchTags_MultipleTags(t *testing.T) {
	s := NewVaultSearcher(newTestVault(t))

	results, err := s.SearchTags(context.Background(), []string{"golang", "programming"})

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "note1.md", results[0].Path)
}

func TestVaultSearcher_SearchTags_NoMatch(t *testing.T) {
	s := NewVaultSearcher(newTestVault(t))

	results, err := s.SearchTags(context.Background(), []string{"nonexistent"})

	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestVaultSearcher_SearchTags_PartialMatch(t *testing.T) {
	s := NewVaultSearcher(newTestVault(t))

	// note1 has golang+programming, searching golang+cli should NOT match
	results, err := s.SearchTags(context.Background(), []string{"golang", "cli"})

	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestVaultSearcher_SearchByType(t *testing.T) {
	s := NewVaultSearcher(newTestVault(t))

	results, err := s.SearchByType(context.Background(), "daily")

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "note2.md", results[0].Path)
}

func TestVaultSearcher_SearchByType_Multiple(t *testing.T) {
	s := NewVaultSearcher(newTestVault(t))

	results, err := s.SearchByType(context.Background(), "note")

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "note1.md", results[0].Path)
}

func TestVaultSearcher_SearchByType_NoMatch(t *testing.T) {
	s := NewVaultSearcher(newTestVault(t))

	results, err := s.SearchByType(context.Background(), "nonexistent")

	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestVaultSearcher_SearchByType_EmptyVault(t *testing.T) {
	emptyVault, err := vault.NewFileVault(t.TempDir(), nil)
	require.NoError(t, err)

	s := NewVaultSearcher(emptyVault)

	results, err := s.SearchByType(context.Background(), "note")

	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestVaultSearcher_SearchByFilename_SubstringMatch(t *testing.T) {
	s := NewVaultSearcher(newTestVault(t))

	results, err := s.SearchByFilename(context.Background(), "note")

	require.NoError(t, err)
	require.Len(t, results, 3)
	assert.Equal(t, "note1.md", results[0].Path)
	assert.Equal(t, "note1", results[0].Match)
	assert.Equal(t, 0, results[0].Line)
	assert.Empty(t, results[0].ContextBefore)
	assert.Empty(t, results[0].ContextAfter)
}

func TestVaultSearcher_SearchByFilename_NormalizesQueryLikeCreate(t *testing.T) {
	v := newFilenameSearchVault(t)
	s := NewVaultSearcher(v)

	results, err := s.SearchByFilename(context.Background(), "Project Plan")

	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Contains(t, results, SearchResult{Path: "Project-Plan.md", Line: 0, Match: "Project-Plan"})
	assert.Contains(t, results, SearchResult{Path: "Project_Plan.md", Line: 0, Match: "Project_Plan"})
}

func TestVaultSearcher_SearchByFilename_MatchesDashedAndUnderscoredNames(t *testing.T) {
	v := newFilenameSearchVault(t)
	s := NewVaultSearcher(v)

	results, err := s.SearchByFilename(context.Background(), "Project Plan")

	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, "Project-Plan.md", results[0].Path)
	assert.Equal(t, "Project-Plan", results[0].Match)
	assert.Equal(t, "Project_Plan.md", results[1].Path)
	assert.Equal(t, "Project_Plan", results[1].Match)
}

func TestVaultSearcher_SearchByFilename_MatchesBasenameInSubdirectories(t *testing.T) {
	v := newFilenameSearchVault(t)
	s := NewVaultSearcher(v)

	results, err := s.SearchByFilename(context.Background(), "architecture")

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, filepath.Join("docs", "Beacon_Architecture.md"), results[0].Path)
	assert.Equal(t, "Beacon_Architecture", results[0].Match)
}

func TestVaultSearcher_SearchByFilename_NoMatch(t *testing.T) {
	v := newFilenameSearchVault(t)
	s := NewVaultSearcher(v)

	results, err := s.SearchByFilename(context.Background(), "missing")

	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestVaultSearcher_SearchByFilename_PrefersPathListingOverFullNoteParsing(t *testing.T) {
	s := NewVaultSearcher(filenameOnlyVault{
		paths: []string{"Project_Plan.md"},
	})

	results, err := s.SearchByFilename(context.Background(), "Project Plan")

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "Project_Plan.md", results[0].Path)
	assert.Equal(t, "Project_Plan", results[0].Match)
}

func newFilenameSearchVault(t *testing.T) vault.Vault {
	t.Helper()

	root := t.TempDir()
	paths := []string{
		"Project-Plan.md",
		"Project_Plan.md",
		filepath.Join("docs", "Beacon_Architecture.md"),
		filepath.Join(".obsidian", "Hidden.md"),
	}

	for _, path := range paths {
		fullPath := filepath.Join(root, path)
		require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0o755))
		require.NoError(t, os.WriteFile(fullPath, []byte("# Test\n"), 0o644))
	}

	v, err := vault.NewFileVault(root, []string{".obsidian"})
	require.NoError(t, err)
	return v
}

type filenameOnlyVault struct {
	paths []string
}

func (v filenameOnlyVault) ListNotes(context.Context) ([]vault.Note, error) {
	return nil, fmt.Errorf("ListNotes should not be used for filename search")
}

func (v filenameOnlyVault) GetNote(context.Context, string) (*vault.Note, error) {
	return nil, fmt.Errorf("GetNote not implemented")
}

func (v filenameOnlyVault) ListNotePaths(context.Context) ([]string, error) {
	return v.paths, nil
}
