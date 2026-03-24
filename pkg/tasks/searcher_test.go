package tasks

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func requireRipgrep(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("rg"); err != nil {
		t.Skip("ripgrep not installed, skipping test")
	}
}

func TestNewSearcher_RequiresRipgrep(t *testing.T) {
	t.Setenv("PATH", "")

	_, err := NewSearcher(t.TempDir(), nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ripgrep")
}

func TestSearcher_ListPendingTasks(t *testing.T) {
	requireRipgrep(t)

	root := t.TempDir()
	writeTaskFixture(t, root, "Inbox.md", "- [ ] first task\ntext\n- [ ] second task\n")

	s, err := NewSearcher(root, nil)
	require.NoError(t, err)

	results, err := s.ListPending(context.Background())
	require.NoError(t, err)
	require.Len(t, results, 2)

	assert.Equal(t, Task{Path: "Inbox.md", Line: 1, Text: "first task"}, results[0])
	assert.Equal(t, Task{Path: "Inbox.md", Line: 3, Text: "second task"}, results[1])
}

func TestSearcher_ListPendingTasks_IgnoresCompletedTasks(t *testing.T) {
	requireRipgrep(t)

	root := t.TempDir()
	writeTaskFixture(t, root, "Tasks.md", "- [x] done\n- [ ] open\n")

	s, err := NewSearcher(root, nil)
	require.NoError(t, err)

	results, err := s.ListPending(context.Background())
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "open", results[0].Text)
}

func TestSearcher_ListPendingTasks_RespectsIgnorePatterns(t *testing.T) {
	requireRipgrep(t)

	root := t.TempDir()
	writeTaskFixture(t, root, filepath.Join("ignored", "Skip.md"), "- [ ] hidden\n")
	writeTaskFixture(t, root, "Show.md", "- [ ] visible\n")

	s, err := NewSearcher(root, []string{"ignored"})
	require.NoError(t, err)

	results, err := s.ListPending(context.Background())
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "Show.md", results[0].Path)
	assert.Equal(t, "visible", results[0].Text)
}

func TestSearcher_ListPendingTasks_RespectsNestedPathIgnorePrefixes(t *testing.T) {
	requireRipgrep(t)

	root := t.TempDir()
	writeTaskFixture(t, root, filepath.Join("archive", "projects", "Skip.md"), "- [ ] hidden\n")
	writeTaskFixture(t, root, filepath.Join("archive", "Keep.md"), "- [ ] visible\n")

	s, err := NewSearcher(root, []string{filepath.Join("archive", "projects")})
	require.NoError(t, err)

	results, err := s.ListPending(context.Background())
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, filepath.Join("archive", "Keep.md"), results[0].Path)
	assert.Equal(t, "visible", results[0].Text)
}

func TestSearcher_ListPendingTasks_KeepsSlashIgnorePatternsAnchoredToVaultRoot(t *testing.T) {
	requireRipgrep(t)

	root := t.TempDir()
	writeTaskFixture(t, root, filepath.Join("notes", "template.md"), "- [ ] hidden\n")
	writeTaskFixture(t, root, filepath.Join("archive", "notes", "template.md"), "- [ ] still visible\n")

	s, err := NewSearcher(root, []string{filepath.Join("notes", "template.md")})
	require.NoError(t, err)

	results, err := s.ListPending(context.Background())
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, filepath.Join("archive", "notes", "template.md"), results[0].Path)
	assert.Equal(t, "still visible", results[0].Text)
}

func TestSearcher_ListPendingTasks_ReturnsLineNumbersAndRelativePaths(t *testing.T) {
	requireRipgrep(t)

	root := t.TempDir()
	writeTaskFixture(t, root, filepath.Join("projects", "Roadmap.md"), "# Roadmap\n\n- [ ] ship\n")

	s, err := NewSearcher(root, nil)
	require.NoError(t, err)

	results, err := s.ListPending(context.Background())
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, filepath.Join("projects", "Roadmap.md"), results[0].Path)
	assert.Equal(t, 3, results[0].Line)
}

func TestSearcher_ListPendingTasks_AllowsIndentedTasks(t *testing.T) {
	requireRipgrep(t)

	root := t.TempDir()
	writeTaskFixture(t, root, "Nested.md", "  - [ ] nested task\n")

	s, err := NewSearcher(root, nil)
	require.NoError(t, err)

	results, err := s.ListPending(context.Background())
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "nested task", results[0].Text)
}

func TestSearcher_ListPendingTasks_HandlesTabsAndSpacesBeforeCheckbox(t *testing.T) {
	requireRipgrep(t)

	root := t.TempDir()
	writeTaskFixture(t, root, "Mixed.md", "\t- [ ] tabbed task\n    - [ ] spaced task\n")

	s, err := NewSearcher(root, nil)
	require.NoError(t, err)

	results, err := s.ListPending(context.Background())
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, "tabbed task", results[0].Text)
	assert.Equal(t, "spaced task", results[1].Text)
}

func TestSearcher_ListPendingTasks_SupportsAsteriskAndPlusBullets(t *testing.T) {
	requireRipgrep(t)

	root := t.TempDir()
	writeTaskFixture(t, root, "Bullets.md", "* [ ] first\n+ [ ] second\n")

	s, err := NewSearcher(root, nil)
	require.NoError(t, err)

	results, err := s.ListPending(context.Background())
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, "first", results[0].Text)
	assert.Equal(t, "second", results[1].Text)
}

func TestSearcher_ListPendingTasks_SupportsOrderedListCheckboxes(t *testing.T) {
	requireRipgrep(t)

	root := t.TempDir()
	writeTaskFixture(t, root, "Ordered.md", "1. [ ] first\n2. [ ] second\n")

	s, err := NewSearcher(root, nil)
	require.NoError(t, err)

	results, err := s.ListPending(context.Background())
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, "first", results[0].Text)
	assert.Equal(t, "second", results[1].Text)
}

func TestSearcher_ListPendingTasks_SupportsEmptyCheckboxes(t *testing.T) {
	requireRipgrep(t)

	root := t.TempDir()
	writeTaskFixture(t, root, "EmptyTasks.md", "- [ ]\n1. [ ]\n")

	s, err := NewSearcher(root, nil)
	require.NoError(t, err)

	results, err := s.ListPending(context.Background())
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, "", results[0].Text)
	assert.Equal(t, "", results[1].Text)
}

func TestSearcher_ListPendingTasks_NoResults(t *testing.T) {
	requireRipgrep(t)

	root := t.TempDir()
	writeTaskFixture(t, root, "Empty.md", "plain text\n- [x] complete\n")

	s, err := NewSearcher(root, nil)
	require.NoError(t, err)

	results, err := s.ListPending(context.Background())
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestSearcher_ListPendingTasks_IncludesHiddenNotesAndIgnoredByDefaultNotes(t *testing.T) {
	requireRipgrep(t)

	root := t.TempDir()
	writeTaskFixture(t, root, filepath.Join(".hidden", "Task.md"), "- [ ] hidden task\n")
	writeTaskFixture(t, root, "IgnoredByRipgrep.md", "- [ ] still included\n")
	require.NoError(t, os.WriteFile(filepath.Join(root, ".ignore"), []byte("IgnoredByRipgrep.md\n"), 0o644))

	s, err := NewSearcher(root, nil)
	require.NoError(t, err)

	results, err := s.ListPending(context.Background())
	require.NoError(t, err)
	require.Len(t, results, 2)

	paths := []string{results[0].Path, results[1].Path}
	assert.Contains(t, paths, filepath.Join(".hidden", "Task.md"))
	assert.Contains(t, paths, "IgnoredByRipgrep.md")
}

func TestSearcher_ListPendingTasks_DoesNotMatchPlainBullets(t *testing.T) {
	requireRipgrep(t)

	root := t.TempDir()
	writeTaskFixture(t, root, "Bullets.md", "- plain bullet\n* another bullet\n")

	s, err := NewSearcher(root, nil)
	require.NoError(t, err)

	results, err := s.ListPending(context.Background())
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestSearcher_ListPendingTasks_DoesNotMatchCheckedBoxes(t *testing.T) {
	requireRipgrep(t)

	root := t.TempDir()
	writeTaskFixture(t, root, "Done.md", "- [x] completed\n- [X] also completed\n")

	s, err := NewSearcher(root, nil)
	require.NoError(t, err)

	results, err := s.ListPending(context.Background())
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestSearcher_ListPendingTasks_SupportsMultipleMatchesPerFile(t *testing.T) {
	requireRipgrep(t)

	root := t.TempDir()
	writeTaskFixture(t, root, "Many.md", "- [ ] first\n- [ ] second\n- [ ] third\n")

	s, err := NewSearcher(root, nil)
	require.NoError(t, err)

	results, err := s.ListPending(context.Background())
	require.NoError(t, err)
	require.Len(t, results, 3)
	assert.Equal(t, 1, results[0].Line)
	assert.Equal(t, 2, results[1].Line)
	assert.Equal(t, 3, results[2].Line)
}

func TestSearcher_ListPendingTasks_PreservesColonsInsideTaskText(t *testing.T) {
	requireRipgrep(t)

	root := t.TempDir()
	writeTaskFixture(t, root, "Colon.md", "- [ ] call API: review payload\n")

	s, err := NewSearcher(root, nil)
	require.NoError(t, err)

	results, err := s.ListPending(context.Background())
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "call API: review payload", results[0].Text)
}

func TestSearcher_ListPendingTasks_KeepsWildcardPathIgnoresConsistentWithFileVault(t *testing.T) {
	requireRipgrep(t)

	root := t.TempDir()
	writeTaskFixture(t, root, filepath.Join("archive", "foo", "Task.md"), "- [ ] still visible\n")
	writeTaskFixture(t, root, filepath.Join("archive", "Task.md"), "- [ ] hidden\n")

	s, err := NewSearcher(root, []string{filepath.Join("archive", "*")})
	require.NoError(t, err)

	results, err := s.ListPending(context.Background())
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, filepath.Join("archive", "foo", "Task.md"), results[0].Path)
	assert.Equal(t, "still visible", results[0].Text)
}

func TestShouldIgnore_UsesNativePathMatchingWithSlashInput(t *testing.T) {
	assert.True(t, shouldIgnore("archive/task.md", []string{"archive"}))
	assert.True(t, shouldIgnore("archive/task.md", []string{filepath.Join("archive", "*.md")}))
	assert.False(t, shouldIgnore("archive/foo/task.md", []string{filepath.Join("archive", "*")}))
}

func TestSearcher_ListPendingTasks_FiltersByFileSubstring(t *testing.T) {
	requireRipgrep(t)

	root := t.TempDir()
	writeTaskFixture(t, root, filepath.Join("notes", "Inbox.md"), "- [ ] visible\n")
	writeTaskFixture(t, root, filepath.Join("archive", "notes", "Inbox.md"), "- [ ] hidden\n")

	s, err := NewSearcher(root, nil)
	require.NoError(t, err)

	results, err := s.ListPendingWithFileFilter(context.Background(), "archive/notes")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, filepath.Join("archive", "notes", "Inbox.md"), results[0].Path)
	assert.Equal(t, "hidden", results[0].Text)
}

func TestSearcher_ListPendingTasks_FiltersByBarePathSubstring(t *testing.T) {
	requireRipgrep(t)

	root := t.TempDir()
	writeTaskFixture(t, root, filepath.Join("archive", "notes", "Inbox.md"), "- [ ] visible\n")
	writeTaskFixture(t, root, filepath.Join("archive", "projects", "Inbox.md"), "- [ ] hidden\n")

	s, err := NewSearcher(root, nil)
	require.NoError(t, err)

	results, err := s.ListPendingWithFileFilter(context.Background(), "notes")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, filepath.Join("archive", "notes", "Inbox.md"), results[0].Path)
	assert.Equal(t, "visible", results[0].Text)
}

func TestSearcher_ListPendingTasks_NormalizesFileSeparators(t *testing.T) {
	requireRipgrep(t)

	root := t.TempDir()
	writeTaskFixture(t, root, filepath.Join("archive", "projects", "Roadmap.md"), "- [ ] visible\n")
	writeTaskFixture(t, root, filepath.Join("archive", "notes", "Roadmap.md"), "- [ ] hidden\n")

	filter := strings.ReplaceAll(filepath.Join("archive", "projects"), "/", "\\")
	s, err := NewSearcher(root, nil)
	require.NoError(t, err)

	results, err := s.ListPendingWithFileFilter(context.Background(), filter)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, filepath.Join("archive", "projects", "Roadmap.md"), results[0].Path)
}

func TestSearcher_ListPendingTasks_EmptyFileFilterBehavesLikeUnfilteredListing(t *testing.T) {
	requireRipgrep(t)

	root := t.TempDir()
	writeTaskFixture(t, root, "A.md", "- [ ] first\n")
	writeTaskFixture(t, root, "B.md", "- [ ] second\n")

	unfiltered, err := NewSearcher(root, nil)
	require.NoError(t, err)
	filtered, err := NewSearcher(root, nil)
	require.NoError(t, err)

	want, err := unfiltered.ListPending(context.Background())
	require.NoError(t, err)
	got, err := filtered.ListPendingWithFileFilter(context.Background(), "")
	require.NoError(t, err)

	assert.Equal(t, want, got)
}

func TestSearcher_ListPendingTasks_PreservesDeterministicOrderingAfterFileFilter(t *testing.T) {
	requireRipgrep(t)

	root := t.TempDir()
	writeTaskFixture(t, root, filepath.Join("z", "Task.md"), "- [ ] third\n")
	writeTaskFixture(t, root, filepath.Join("a", "Task.md"), "- [ ] first\n- [ ] second\n")

	s, err := NewSearcher(root, nil)
	require.NoError(t, err)

	results, err := s.ListPendingWithFileFilter(context.Background(), "Task.md")
	require.NoError(t, err)
	require.Len(t, results, 3)
	assert.Equal(t, filepath.Join("a", "Task.md"), results[0].Path)
	assert.Equal(t, 1, results[0].Line)
	assert.Equal(t, filepath.Join("a", "Task.md"), results[1].Path)
	assert.Equal(t, 2, results[1].Line)
	assert.Equal(t, filepath.Join("z", "Task.md"), results[2].Path)
	assert.Equal(t, 1, results[2].Line)
}

func TestSearcher_ListPendingTasks_NoResultsAfterFileFilter(t *testing.T) {
	requireRipgrep(t)

	root := t.TempDir()
	writeTaskFixture(t, root, "Inbox.md", "- [ ] task\n")

	s, err := NewSearcher(root, nil)
	require.NoError(t, err)

	results, err := s.ListPendingWithFileFilter(context.Background(), "missing")
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestSearcher_ListPendingTasks_FileFilterIsCaseSensitive(t *testing.T) {
	requireRipgrep(t)

	root := t.TempDir()
	writeTaskFixture(t, root, filepath.Join("archive", "notes", "Inbox.md"), "- [ ] lower\n")
	writeTaskFixture(t, root, filepath.Join("Archive", "notes", "Inbox.md"), "- [ ] upper\n")

	s, err := NewSearcher(root, nil)
	require.NoError(t, err)

	results, err := s.ListPendingWithFileFilter(context.Background(), "archive/notes")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, filepath.Join("archive", "notes", "Inbox.md"), results[0].Path)
	assert.Equal(t, "lower", results[0].Text)
}

func writeTaskFixture(t *testing.T, root, relPath, content string) {
	t.Helper()

	fullPath := filepath.Join(root, relPath)
	require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0o755))
	require.NoError(t, os.WriteFile(fullPath, []byte(content), 0o644))
}
