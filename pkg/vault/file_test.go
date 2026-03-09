package vault

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const fixturesPath = "../../testdata/fixtures/vault"

func TestNewFileVault(t *testing.T) {
	v, err := NewFileVault(fixturesPath, nil)
	require.NoError(t, err)
	assert.NotNil(t, v)
	assert.Equal(t, fixturesPath, v.rootPath)
}

func TestNewFileVault_InvalidPath(t *testing.T) {
	v, err := NewFileVault("/nonexistent/path/does/not/exist", nil)
	assert.Error(t, err)
	assert.Nil(t, v)
	assert.Contains(t, err.Error(), "vault: invalid root path")
}

func TestFileVault_ListNotes(t *testing.T) {
	v, err := NewFileVault(fixturesPath, nil)
	require.NoError(t, err)

	notes, err := v.ListNotes(context.Background())
	require.NoError(t, err)
	assert.Len(t, notes, 4)

	// Collect all paths
	paths := make(map[string]bool)
	for _, n := range notes {
		paths[n.Path] = true
	}
	assert.True(t, paths["note1.md"])
	assert.True(t, paths["note2.md"])
	assert.True(t, paths["empty.md"])
	assert.True(t, paths["subdir/note3.md"])
}

func TestFileVault_ListNotes_WithIgnore(t *testing.T) {
	v, err := NewFileVault(fixturesPath, []string{"subdir"})
	require.NoError(t, err)

	notes, err := v.ListNotes(context.Background())
	require.NoError(t, err)
	assert.Len(t, notes, 3)

	for _, n := range notes {
		assert.NotContains(t, n.Path, "subdir")
	}
}

func TestFileVault_ListNotes_ParsesFrontmatter(t *testing.T) {
	v, err := NewFileVault(fixturesPath, nil)
	require.NoError(t, err)

	notes, err := v.ListNotes(context.Background())
	require.NoError(t, err)

	var note1 *Note
	for i := range notes {
		if notes[i].Path == "note1.md" {
			note1 = &notes[i]
			break
		}
	}
	require.NotNil(t, note1, "note1.md should be found")

	assert.Equal(t, "Go Tips", note1.Name)
	assert.Equal(t, []string{"golang", "programming"}, note1.Tags)
	assert.Equal(t, "note", note1.Frontmatter["type"])
	assert.Equal(t, "active", note1.Frontmatter["status"])
}

func TestFileVault_ListNotes_EmptyVault(t *testing.T) {
	dir := t.TempDir()

	v, err := NewFileVault(dir, nil)
	require.NoError(t, err)

	notes, err := v.ListNotes(context.Background())
	require.NoError(t, err)
	assert.Empty(t, notes)
}

func TestFileVault_ListNotes_NoFrontmatter(t *testing.T) {
	v, err := NewFileVault(fixturesPath, nil)
	require.NoError(t, err)

	notes, err := v.ListNotes(context.Background())
	require.NoError(t, err)

	var emptyNote *Note
	for i := range notes {
		if notes[i].Path == "empty.md" {
			emptyNote = &notes[i]
			break
		}
	}
	require.NotNil(t, emptyNote, "empty.md should be found")

	assert.Empty(t, emptyNote.Tags)
	assert.Empty(t, emptyNote.Frontmatter)
	// Name falls back to filename without extension
	assert.Equal(t, "empty", emptyNote.Name)
}

func TestFileVault_GetNote(t *testing.T) {
	v, err := NewFileVault(fixturesPath, nil)
	require.NoError(t, err)

	note, err := v.GetNote(context.Background(), "note1.md")
	require.NoError(t, err)
	require.NotNil(t, note)

	assert.Equal(t, "note1.md", note.Path)
	assert.Equal(t, "Go Tips", note.Name)
	assert.Equal(t, []string{"golang", "programming"}, note.Tags)
	assert.Contains(t, note.Content, "Go is a statically typed language")
}

func TestFileVault_GetNote_Subdir(t *testing.T) {
	v, err := NewFileVault(fixturesPath, nil)
	require.NoError(t, err)

	note, err := v.GetNote(context.Background(), "subdir/note3.md")
	require.NoError(t, err)
	require.NotNil(t, note)

	assert.Equal(t, "subdir/note3.md", note.Path)
	assert.Equal(t, "Beacon Project", note.Name)
	assert.Equal(t, []string{"beacon", "cli"}, note.Tags)
}

func TestFileVault_GetNote_NotFound(t *testing.T) {
	v, err := NewFileVault(fixturesPath, nil)
	require.NoError(t, err)

	note, err := v.GetNote(context.Background(), "nonexistent.md")
	assert.Error(t, err)
	assert.Nil(t, note)
	assert.Contains(t, err.Error(), "vault: note not found")
}
