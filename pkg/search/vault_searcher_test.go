package search

import (
	"context"
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
