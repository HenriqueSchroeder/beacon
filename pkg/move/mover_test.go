package move

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/HenriqueSchroeder/beacon/pkg/vault"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helper to create a temp vault with files
func setupVault(t *testing.T, files map[string]string) (string, *vault.FileVault) {
	t.Helper()
	dir := t.TempDir()

	for path, content := range files {
		abs := filepath.Join(dir, path)
		require.NoError(t, os.MkdirAll(filepath.Dir(abs), 0o755))
		require.NoError(t, os.WriteFile(abs, []byte(content), 0o644))
	}

	v, err := vault.NewFileVault(dir, nil)
	require.NoError(t, err)
	return dir, v
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(data)
}

func TestPlan_BasicRename(t *testing.T) {
	dir, v := setupVault(t, map[string]string{
		"Old Note.md": "# Old Note\nSome content",
		"Index.md":    "# Index\nSee [[Old Note]] for details",
	})

	mover := NewMover(dir, v)
	result, err := mover.Plan(context.Background(), "Old Note.md", "New Note.md")

	require.NoError(t, err)
	assert.Equal(t, "Old Note", result.OldStem)
	assert.Equal(t, "New Note", result.NewStem)
	assert.True(t, result.NeedsRelink)
	require.Len(t, result.Updates, 1)
	assert.Equal(t, "Index.md", result.Updates[0].Path)
	require.Len(t, result.Updates[0].Replacements, 1)
	assert.Equal(t, "[[Old Note]]", result.Updates[0].Replacements[0].OldRaw)
	assert.Equal(t, "[[New Note]]", result.Updates[0].Replacements[0].NewRaw)
}

func TestPlan_FolderMoveNoRename(t *testing.T) {
	dir, v := setupVault(t, map[string]string{
		"notes/My Note.md": "# My Note\nContent",
		"Index.md":         "# Index\n[[My Note]] link here",
	})

	mover := NewMover(dir, v)
	result, err := mover.Plan(context.Background(), "notes/My Note.md", "archive/My Note.md")

	require.NoError(t, err)
	assert.False(t, result.NeedsRelink, "folder-only move should not need relinking")
	assert.Empty(t, result.Updates)
}

func TestPlan_SourceNotFound(t *testing.T) {
	dir, v := setupVault(t, map[string]string{
		"Existing.md": "content",
	})

	mover := NewMover(dir, v)
	_, err := mover.Plan(context.Background(), "Missing.md", "New.md")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "source not found")
}

func TestPlan_DestinationAlreadyExists(t *testing.T) {
	dir, v := setupVault(t, map[string]string{
		"Source.md": "content",
		"Target.md": "existing content",
	})

	mover := NewMover(dir, v)
	_, err := mover.Plan(context.Background(), "Source.md", "Target.md")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "destination already exists")
}

func TestPlan_PreservesHeadingAndAlias(t *testing.T) {
	dir, v := setupVault(t, map[string]string{
		"Note.md":  "# Note\nContent",
		"Index.md": "See [[Note#Section|click here]] for details",
	})

	mover := NewMover(dir, v)
	result, err := mover.Plan(context.Background(), "Note.md", "Renamed.md")

	require.NoError(t, err)
	require.Len(t, result.Updates, 1)
	rep := result.Updates[0].Replacements[0]
	assert.Equal(t, "[[Note#Section|click here]]", rep.OldRaw)
	assert.Equal(t, "[[Renamed#Section|click here]]", rep.NewRaw)
}

func TestPlan_CaseInsensitiveMatching(t *testing.T) {
	dir, v := setupVault(t, map[string]string{
		"My Note.md": "# My Note\nContent",
		"Index.md":   "Link to [[my note]] here",
	})

	mover := NewMover(dir, v)
	result, err := mover.Plan(context.Background(), "My Note.md", "Renamed.md")

	require.NoError(t, err)
	require.Len(t, result.Updates, 1)
	rep := result.Updates[0].Replacements[0]
	assert.Equal(t, "[[my note]]", rep.OldRaw)
	assert.Equal(t, "[[Renamed]]", rep.NewRaw)
}

func TestPlan_SelfReference(t *testing.T) {
	dir, v := setupVault(t, map[string]string{
		"Note.md": "# Note\nSee also [[Note#Section]]",
	})

	mover := NewMover(dir, v)
	result, err := mover.Plan(context.Background(), "Note.md", "New Name.md")

	require.NoError(t, err)
	require.Len(t, result.Updates, 1)
	assert.Equal(t, "Note.md", result.Updates[0].Path)
	assert.Equal(t, "[[Note#Section]]", result.Updates[0].Replacements[0].OldRaw)
	assert.Equal(t, "[[New Name#Section]]", result.Updates[0].Replacements[0].NewRaw)
}

func TestPlan_MultipleLinksInSameFile(t *testing.T) {
	dir, v := setupVault(t, map[string]string{
		"Note.md":  "# Note\nContent",
		"Index.md": "First [[Note]] and second [[Note|alias]] and third [[Note#h1]]",
	})

	mover := NewMover(dir, v)
	result, err := mover.Plan(context.Background(), "Note.md", "Renamed.md")

	require.NoError(t, err)
	require.Len(t, result.Updates, 1)
	assert.Len(t, result.Updates[0].Replacements, 3)
}

func TestApply_BasicRenameAndRelink(t *testing.T) {
	dir, v := setupVault(t, map[string]string{
		"Old.md":   "# Old\nSome content",
		"Index.md": "# Index\nSee [[Old]] for details\nAlso [[Old#Section]]",
	})

	mover := NewMover(dir, v)
	result, err := mover.Plan(context.Background(), "Old.md", "New.md")
	require.NoError(t, err)

	summary, err := mover.Apply(result)
	require.NoError(t, err)

	assert.Equal(t, 1, summary.FilesMoved)
	assert.Equal(t, 2, summary.LinksUpdated)
	assert.Equal(t, 1, summary.FilesUpdated)
	assert.Empty(t, summary.Errors)

	// Source should be gone
	_, err = os.Stat(filepath.Join(dir, "Old.md"))
	assert.True(t, os.IsNotExist(err))

	// Dest should exist
	destContent := readFile(t, filepath.Join(dir, "New.md"))
	assert.Contains(t, destContent, "# Old")

	// Index should have updated links
	indexContent := readFile(t, filepath.Join(dir, "Index.md"))
	assert.Contains(t, indexContent, "[[New]]")
	assert.Contains(t, indexContent, "[[New#Section]]")
	assert.NotContains(t, indexContent, "[[Old]]")
}

func TestApply_CreatesDestinationDirectory(t *testing.T) {
	dir, v := setupVault(t, map[string]string{
		"Note.md": "# Note\nContent",
	})

	mover := NewMover(dir, v)
	result, err := mover.Plan(context.Background(), "Note.md", "subdir/deep/Note.md")
	require.NoError(t, err)

	summary, err := mover.Apply(result)
	require.NoError(t, err)

	assert.Equal(t, 1, summary.FilesMoved)

	destContent := readFile(t, filepath.Join(dir, "subdir", "deep", "Note.md"))
	assert.Contains(t, destContent, "# Note")
}

func TestApply_FolderMoveNoLinkChanges(t *testing.T) {
	dir, v := setupVault(t, map[string]string{
		"notes/Note.md": "# Note\nContent",
		"Index.md":      "See [[Note]]",
	})

	mover := NewMover(dir, v)
	result, err := mover.Plan(context.Background(), "notes/Note.md", "archive/Note.md")
	require.NoError(t, err)

	summary, err := mover.Apply(result)
	require.NoError(t, err)

	assert.Equal(t, 1, summary.FilesMoved)
	assert.Equal(t, 0, summary.LinksUpdated)

	// Index should be untouched
	indexContent := readFile(t, filepath.Join(dir, "Index.md"))
	assert.Contains(t, indexContent, "[[Note]]")
}

func TestApply_MultipleLinksSameLine(t *testing.T) {
	dir, v := setupVault(t, map[string]string{
		"Note.md":  "# Note\nContent",
		"Index.md": "Both [[Note]] and [[Note|display]] on same line",
	})

	mover := NewMover(dir, v)
	result, err := mover.Plan(context.Background(), "Note.md", "Renamed.md")
	require.NoError(t, err)

	summary, err := mover.Apply(result)
	require.NoError(t, err)

	assert.Equal(t, 2, summary.LinksUpdated)

	indexContent := readFile(t, filepath.Join(dir, "Index.md"))
	assert.Contains(t, indexContent, "[[Renamed]]")
	assert.Contains(t, indexContent, "[[Renamed|display]]")
	assert.NotContains(t, indexContent, "[[Note]]")
}

func TestApply_PreservesEmbedPrefix(t *testing.T) {
	dir, v := setupVault(t, map[string]string{
		"Image Note.md": "# Image Note\nContent",
		"Page.md":       "Embedded: ![[Image Note]]",
	})

	mover := NewMover(dir, v)
	result, err := mover.Plan(context.Background(), "Image Note.md", "Photo.md")
	require.NoError(t, err)

	summary, err := mover.Apply(result)
	require.NoError(t, err)

	assert.Equal(t, 1, summary.LinksUpdated)

	pageContent := readFile(t, filepath.Join(dir, "Page.md"))
	assert.Contains(t, pageContent, "![[Photo]]")
}

func TestPlan_PathTraversalBlocked(t *testing.T) {
	dir, v := setupVault(t, map[string]string{
		"Note.md": "content",
	})

	mover := NewMover(dir, v)

	_, err := mover.Plan(context.Background(), "../../etc/passwd", "Note.md")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path must be within the vault")

	_, err = mover.Plan(context.Background(), "Note.md", "../outside.md")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path must be within the vault")
}

func TestApply_IdenticalLinksMultipleOccurrences(t *testing.T) {
	dir, v := setupVault(t, map[string]string{
		"Note.md":  "# Note\nContent",
		"Index.md": "First [[Note]] then [[Note]] and [[Note]] again",
	})

	mover := NewMover(dir, v)
	result, err := mover.Plan(context.Background(), "Note.md", "Renamed.md")
	require.NoError(t, err)

	summary, err := mover.Apply(result)
	require.NoError(t, err)

	// All 3 occurrences should be updated
	assert.Equal(t, 3, summary.LinksUpdated)

	indexContent := readFile(t, filepath.Join(dir, "Index.md"))
	assert.NotContains(t, indexContent, "[[Note]]")
	assert.Equal(t, "First [[Renamed]] then [[Renamed]] and [[Renamed]] again", indexContent)
}

func TestStemFromPath(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"Note.md", "Note"},
		{"path/to/My Note.md", "My Note"},
		{"no-ext", "no-ext"},
		{"deep/path/file.txt", "file"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, stemFromPath(tt.path), "stemFromPath(%q)", tt.path)
	}
}

func TestBuildNewRaw(t *testing.T) {
	tests := []struct {
		target, heading, alias, want string
	}{
		{"Note", "", "", "[[Note]]"},
		{"Note", "Section", "", "[[Note#Section]]"},
		{"Note", "", "display", "[[Note|display]]"},
		{"Note", "H1", "alias", "[[Note#H1|alias]]"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, buildNewRaw(tt.target, tt.heading, tt.alias))
	}
}
