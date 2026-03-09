package search

import (
	"context"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const fixturesPath = "../../testdata/fixtures/vault"

func requireRipgrep(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("rg"); err != nil {
		t.Skip("ripgrep not installed, skipping test")
	}
}

func TestNewRipgrepSearcher(t *testing.T) {
	requireRipgrep(t)

	s, err := NewRipgrepSearcher(fixturesPath, nil)
	require.NoError(t, err)
	assert.NotNil(t, s)
	assert.NotEmpty(t, s.vaultPath)
	assert.True(t, filepath.IsAbs(s.vaultPath), "vaultPath should be absolute")
}

func TestNewRipgrepSearcher_InvalidPath(t *testing.T) {
	requireRipgrep(t)

	_, err := NewRipgrepSearcher("/nonexistent/path/that/does/not/exist", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "vault path error")
}

func TestNewRipgrepSearcher_NoRipgrep(t *testing.T) {
	t.Setenv("PATH", "")

	_, err := NewRipgrepSearcher(fixturesPath, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ripgrep")
}

func TestRipgrepSearcher_SearchContent(t *testing.T) {
	requireRipgrep(t)

	s, err := NewRipgrepSearcher(fixturesPath, nil)
	require.NoError(t, err)

	results, err := s.SearchContent(context.Background(), "golang")
	require.NoError(t, err)
	require.NotEmpty(t, results)

	found := false
	for _, r := range results {
		if r.Path == "note1.md" {
			found = true
			assert.Greater(t, r.Line, 0)
			assert.Contains(t, r.Match, "golang")
		}
	}
	assert.True(t, found, "expected to find match in note1.md")
}

func TestRipgrepSearcher_SearchContent_WithContext(t *testing.T) {
	requireRipgrep(t)

	s, err := NewRipgrepSearcher(fixturesPath, nil)
	require.NoError(t, err)

	results, err := s.SearchContent(context.Background(), "statically typed")
	require.NoError(t, err)
	require.NotEmpty(t, results)

	found := false
	for _, r := range results {
		if r.Path == "note1.md" {
			found = true
			assert.Contains(t, r.Match, "statically typed")
		}
	}
	assert.True(t, found, "expected to find 'statically typed' in note1.md")
}

func TestRipgrepSearcher_SearchContent_NoResults(t *testing.T) {
	requireRipgrep(t)

	s, err := NewRipgrepSearcher(fixturesPath, nil)
	require.NoError(t, err)

	results, err := s.SearchContent(context.Background(), "xyznonexistent123")
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestRipgrepSearcher_SearchContent_MultipleFiles(t *testing.T) {
	requireRipgrep(t)

	s, err := NewRipgrepSearcher(fixturesPath, nil)
	require.NoError(t, err)

	results, err := s.SearchContent(context.Background(), "Obsidian")
	require.NoError(t, err)
	require.NotEmpty(t, results)

	paths := make(map[string]bool)
	for _, r := range results {
		paths[r.Path] = true
	}
	assert.GreaterOrEqual(t, len(paths), 2, "expected matches in at least 2 files")
}

func TestRipgrepSearcher_SearchContent_WithIgnore(t *testing.T) {
	requireRipgrep(t)

	s, err := NewRipgrepSearcher(fixturesPath, []string{"subdir"})
	require.NoError(t, err)

	results, err := s.SearchContent(context.Background(), "Beacon")
	require.NoError(t, err)

	for _, r := range results {
		assert.False(t, hasPrefix(r.Path, "subdir"), "expected no results from subdir, got path: %s", r.Path)
	}
}

func TestRipgrepSearcher_SearchContent_ContextCancelled(t *testing.T) {
	requireRipgrep(t)

	s, err := NewRipgrepSearcher(fixturesPath, nil)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err = s.SearchContent(ctx, "golang")
	require.Error(t, err)
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
