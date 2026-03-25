package property

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEditorGet_PrintsScalarValue(t *testing.T) {
	vaultPath := t.TempDir()
	writePropertyNote(t, vaultPath, "note.md", "---\nstatus: done\n---\n# Note\n")

	editor := NewEditor(vaultPath)

	value, err := editor.Get("note.md", "status")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != "done" {
		t.Fatalf("expected done, got %#v", value)
	}
}

func TestEditorSet_CreatesFrontmatterWhenMissing(t *testing.T) {
	vaultPath := t.TempDir()
	writePropertyNote(t, vaultPath, "note.md", "# Note\nBody\n")

	editor := NewEditor(vaultPath)

	if err := editor.Set("note.md", "status", "done"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readPropertyNote(t, vaultPath, "note.md")
	if !strings.HasPrefix(got, "---\nstatus: done\n---\n") {
		t.Fatalf("expected frontmatter to be created, got %q", got)
	}
	if !strings.Contains(got, "# Note\nBody\n") {
		t.Fatalf("expected body to be preserved, got %q", got)
	}
}

func TestEditorSet_PreservesMarkdownBody(t *testing.T) {
	vaultPath := t.TempDir()
	originalBody := "# Note\n\nParagraph\n- item\n"
	writePropertyNote(t, vaultPath, "note.md", "---\nstatus: todo\n---\n"+originalBody)

	editor := NewEditor(vaultPath)

	if err := editor.Set("note.md", "status", "done"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readPropertyNote(t, vaultPath, "note.md")
	if !strings.HasSuffix(got, originalBody) {
		t.Fatalf("expected markdown body to be preserved, got %q", got)
	}
}

func TestEditorSet_PreservesCRLFLineEndings(t *testing.T) {
	vaultPath := t.TempDir()
	writePropertyNote(t, vaultPath, "note.md", "---\r\nstatus: todo\r\n---\r\n# Note\r\n")

	editor := NewEditor(vaultPath)

	if err := editor.Set("note.md", "status", "done"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readPropertyNote(t, vaultPath, "note.md")
	if strings.Contains(got, "---\n") {
		t.Fatalf("expected CRLF output, got %q", got)
	}
	if !strings.Contains(got, "---\r\nstatus: done\r\n---\r\n") {
		t.Fatalf("expected CRLF frontmatter, got %q", got)
	}
}

func TestEditorSet_UsesOpeningFenceLineEndingWhenBodyContainsCRLF(t *testing.T) {
	vaultPath := t.TempDir()
	writePropertyNote(t, vaultPath, "note.md", "---\nstatus: todo\n---\n# Note\r\n")

	editor := NewEditor(vaultPath)

	if err := editor.Set("note.md", "status", "done"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readPropertyNote(t, vaultPath, "note.md")
	if strings.Count(got, "---") != 2 {
		t.Fatalf("expected a single frontmatter block, got %q", got)
	}
	if !strings.HasPrefix(got, "---\nstatus: done\n---\n") {
		t.Fatalf("expected existing LF frontmatter to be updated, got %q", got)
	}
}

func TestEditorAdd_AppendsUniqueListValue(t *testing.T) {
	vaultPath := t.TempDir()
	writePropertyNote(t, vaultPath, "note.md", "---\ntags:\n  - work\n---\n# Note\n")

	editor := NewEditor(vaultPath)

	if err := editor.Add("note.md", "tags", "urgent"); err != nil {
		t.Fatalf("unexpected error adding new tag: %v", err)
	}
	if err := editor.Add("note.md", "tags", "urgent"); err != nil {
		t.Fatalf("unexpected error re-adding existing tag: %v", err)
	}

	got := readPropertyNote(t, vaultPath, "note.md")
	if strings.Count(got, "- urgent") != 1 {
		t.Fatalf("expected urgent to be added once, got %q", got)
	}
}

func TestEditorAdd_RejectsScalarField(t *testing.T) {
	vaultPath := t.TempDir()
	writePropertyNote(t, vaultPath, "note.md", "---\nstatus: todo\n---\n# Note\n")

	editor := NewEditor(vaultPath)

	err := editor.Add("note.md", "status", "done")
	if err == nil {
		t.Fatal("expected error when adding to scalar field")
	}
}

func TestEditorGet_YAMLEncodesLists(t *testing.T) {
	vaultPath := t.TempDir()
	writePropertyNote(t, vaultPath, "note.md", "---\ntags:\n  - work\n  - urgent\n---\n# Note\n")

	editor := NewEditor(vaultPath)

	value, err := editor.Get("note.md", "tags")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	list, ok := value.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", value)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(list))
	}
}

func TestEditorSet_RejectsNoteOutsideVault(t *testing.T) {
	vaultPath := t.TempDir()
	editor := NewEditor(vaultPath)

	err := editor.Set("../outside.md", "status", "done")
	if err == nil {
		t.Fatal("expected error for path outside vault")
	}
}

func TestEditorSet_RejectsNonMarkdownPath(t *testing.T) {
	vaultPath := t.TempDir()
	writePropertyNote(t, vaultPath, "note.txt", "plain text")

	editor := NewEditor(vaultPath)

	err := editor.Set("note.txt", "status", "done")
	if err == nil {
		t.Fatal("expected error for non-markdown path")
	}
}

func TestEditorSet_RewritesExistingFrontmatterKey(t *testing.T) {
	vaultPath := t.TempDir()
	writePropertyNote(t, vaultPath, "note.md", "---\nstatus: todo\nowner: henrique\n---\n# Note\n")

	editor := NewEditor(vaultPath)

	if err := editor.Set("note.md", "status", "done"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readPropertyNote(t, vaultPath, "note.md")
	if !strings.Contains(got, "status: done") {
		t.Fatalf("expected rewritten key, got %q", got)
	}
	if !strings.Contains(got, "owner: henrique") {
		t.Fatalf("expected other keys to remain, got %q", got)
	}
}

func TestEditorGet_ErrorsWhenKeyMissing(t *testing.T) {
	vaultPath := t.TempDir()
	writePropertyNote(t, vaultPath, "note.md", "---\nstatus: done\n---\n# Note\n")

	editor := NewEditor(vaultPath)

	_, err := editor.Get("note.md", "owner")
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestEditorSet_PreservesBlankLineBetweenFrontmatterAndBody(t *testing.T) {
	vaultPath := t.TempDir()
	writePropertyNote(t, vaultPath, "note.md", "---\nstatus: todo\n---\n\n# Note\n")

	editor := NewEditor(vaultPath)

	if err := editor.Set("note.md", "status", "done"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readPropertyNote(t, vaultPath, "note.md")
	if !strings.Contains(got, "---\n\n# Note\n") {
		t.Fatalf("expected blank line between frontmatter and body, got %q", got)
	}
}

func TestEditorGet_AllowsFrontmatterClosingDelimiterAtEOF(t *testing.T) {
	vaultPath := t.TempDir()
	writePropertyNote(t, vaultPath, "note.md", "---\nstatus: done\n---")

	editor := NewEditor(vaultPath)

	value, err := editor.Get("note.md", "status")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != "done" {
		t.Fatalf("expected done, got %#v", value)
	}
}

func TestEditorSet_TreatsUnterminatedFenceAsPlainMarkdown(t *testing.T) {
	vaultPath := t.TempDir()
	writePropertyNote(t, vaultPath, "note.md", "---\n# Note\nBody\n")

	editor := NewEditor(vaultPath)

	if err := editor.Set("note.md", "status", "done"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readPropertyNote(t, vaultPath, "note.md")
	if !strings.HasPrefix(got, "---\nstatus: done\n---\n") {
		t.Fatalf("expected a new frontmatter block to be prepended, got %q", got)
	}
	if !strings.Contains(got, "---\n# Note\nBody\n") {
		t.Fatalf("expected original markdown body to be preserved, got %q", got)
	}
}

func TestEditorSet_NormalizesFrontmatterFormatting(t *testing.T) {
	vaultPath := t.TempDir()
	writePropertyNote(t, vaultPath, "note.md", "---\n# comment\nstatus: \"todo\"\ntags: [work]\n---\n# Note\n")

	editor := NewEditor(vaultPath)

	if err := editor.Set("note.md", "status", "done"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readPropertyNote(t, vaultPath, "note.md")
	if strings.Contains(got, "# comment") {
		t.Fatalf("expected frontmatter comments to be normalized away, got %q", got)
	}
	if strings.Contains(got, "\"todo\"") {
		t.Fatalf("expected yaml formatting to be normalized, got %q", got)
	}
	if !strings.Contains(got, "tags:\n    - work") && !strings.Contains(got, "tags:\n  - work") {
		t.Fatalf("expected tags to be rewritten in block yaml form, got %q", got)
	}
}

func TestEditorRemove_DeletesExistingKey(t *testing.T) {
	vaultPath := t.TempDir()
	writePropertyNote(t, vaultPath, "note.md", "---\nstatus: todo\nowner: henrique\n---\n# Note\nBody\n")

	editor := NewEditor(vaultPath)

	if err := editor.Remove("note.md", "status"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readPropertyNote(t, vaultPath, "note.md")
	if strings.Contains(got, "status:") {
		t.Fatalf("expected status key to be removed, got %q", got)
	}
	if !strings.Contains(got, "owner: henrique") {
		t.Fatalf("expected remaining keys to stay intact, got %q", got)
	}
	if !strings.HasSuffix(got, "# Note\nBody\n") {
		t.Fatalf("expected markdown body to be preserved, got %q", got)
	}
}

func TestEditorRemove_PreservesCRLFLineEndingsWhenFrontmatterRemains(t *testing.T) {
	vaultPath := t.TempDir()
	originalBody := "# Note\r\n\r\nParagraph\r\n"
	writePropertyNote(t, vaultPath, "note.md", "---\r\nstatus: todo\r\nowner: henrique\r\n---\r\n"+originalBody)

	editor := NewEditor(vaultPath)

	if err := editor.Remove("note.md", "status"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readPropertyNote(t, vaultPath, "note.md")
	if strings.Contains(got, "---\n") {
		t.Fatalf("expected CRLF frontmatter to be preserved, got %q", got)
	}
	if !strings.Contains(got, "---\r\nowner: henrique\r\n---\r\n") {
		t.Fatalf("expected remaining frontmatter to preserve CRLF, got %q", got)
	}
	if !strings.HasSuffix(got, originalBody) {
		t.Fatalf("expected body to be preserved, got %q", got)
	}
}

func TestEditorRemove_RemovesFrontmatterBlockWhenLastKeyDeleted(t *testing.T) {
	vaultPath := t.TempDir()
	writePropertyNote(t, vaultPath, "note.md", "---\nstatus: todo\n---\n# Note\nBody\n")

	editor := NewEditor(vaultPath)

	if err := editor.Remove("note.md", "status"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readPropertyNote(t, vaultPath, "note.md")
	if strings.Contains(got, "---") {
		t.Fatalf("expected frontmatter block to be removed, got %q", got)
	}
	if got != "# Note\nBody\n" {
		t.Fatalf("expected body only after removing last key, got %q", got)
	}
}

func TestEditorRemove_RemovesLastKeyAndPreservesCRLFBody(t *testing.T) {
	vaultPath := t.TempDir()
	originalBody := "# Note\r\nBody\r\n"
	writePropertyNote(t, vaultPath, "note.md", "---\r\nstatus: todo\r\n---\r\n"+originalBody)

	editor := NewEditor(vaultPath)

	if err := editor.Remove("note.md", "status"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readPropertyNote(t, vaultPath, "note.md")
	if strings.Contains(got, "---") {
		t.Fatalf("expected frontmatter block to be removed, got %q", got)
	}
	if got != originalBody {
		t.Fatalf("expected CRLF body to be preserved, got %q", got)
	}
}

func TestEditorRemove_ErrorsWhenKeyMissing(t *testing.T) {
	vaultPath := t.TempDir()
	writePropertyNote(t, vaultPath, "note.md", "---\nstatus: todo\n---\n# Note\n")

	editor := NewEditor(vaultPath)

	err := editor.Remove("note.md", "owner")
	if err == nil {
		t.Fatal("expected error when removing missing key")
	}
	if !strings.Contains(err.Error(), `property: key "owner" not found`) {
		t.Fatalf("expected missing-key error, got %v", err)
	}
}

func TestEditorRemove_TreatsPlainMarkdownAsMissingKey(t *testing.T) {
	vaultPath := t.TempDir()
	writePropertyNote(t, vaultPath, "note.md", "# Note\nBody\n")

	editor := NewEditor(vaultPath)

	err := editor.Remove("note.md", "status")
	if err == nil {
		t.Fatal("expected error when removing missing key from plain markdown")
	}
	if !strings.Contains(err.Error(), `property: key "status" not found`) {
		t.Fatalf("expected missing-key error, got %v", err)
	}
}

func TestEditorRemove_PreservesMarkdownBodyAndPathValidation(t *testing.T) {
	t.Run("preserves markdown body", func(t *testing.T) {
		vaultPath := t.TempDir()
		originalBody := "# Note\n\nParagraph\n- item\n"
		writePropertyNote(t, vaultPath, "note.md", "---\nstatus: todo\n---\n"+originalBody)

		editor := NewEditor(vaultPath)

		if err := editor.Remove("note.md", "status"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got := readPropertyNote(t, vaultPath, "note.md")
		if got != originalBody {
			t.Fatalf("expected markdown body to be preserved, got %q", got)
		}
	})

	t.Run("rejects note outside vault", func(t *testing.T) {
		vaultPath := t.TempDir()
		editor := NewEditor(vaultPath)

		err := editor.Remove("../outside.md", "status")
		if err == nil {
			t.Fatal("expected error for path outside vault")
		}
	})

	t.Run("rejects non-markdown path", func(t *testing.T) {
		vaultPath := t.TempDir()
		editor := NewEditor(vaultPath)

		err := editor.Remove("note.txt", "status")
		if err == nil {
			t.Fatal("expected error for non-markdown path")
		}
	})
}

func writePropertyNote(t *testing.T, vaultPath, relPath, content string) {
	t.Helper()

	fullPath := filepath.Join(vaultPath, relPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
}

func readPropertyNote(t *testing.T, vaultPath, relPath string) string {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(vaultPath, relPath))
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	return string(data)
}
