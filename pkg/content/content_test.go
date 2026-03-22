package content

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- FindFrontmatterEnd ---

func TestFindFrontmatterEnd_NoFrontmatter(t *testing.T) {
	input := "# Title\n\nBody text here."
	if got := FindFrontmatterEnd(input); got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
}

func TestFindFrontmatterEnd_WithFrontmatter(t *testing.T) {
	input := "---\ntitle: Test\ndate: 2026-01-01\n---\n\n# Body"
	// Exact expected offset: len("---\n") + len("title: Test\ndate: 2026-01-01") + len("\n---\n")
	expected := len("---\ntitle: Test\ndate: 2026-01-01\n---\n")
	end := FindFrontmatterEnd(input)
	if end != expected {
		t.Errorf("expected offset %d, got %d", expected, end)
	}
	rest := input[end:]
	if !strings.HasPrefix(rest, "\n# Body") {
		t.Errorf("expected rest to start with '\\n# Body', got: %q", rest)
	}
}

func TestFindFrontmatterEnd_CRLF(t *testing.T) {
	input := "---\r\ntitle: Test\r\ndate: 2026-01-01\r\n---\r\n\r\n# Body"
	end := FindFrontmatterEnd(input)
	if end == 0 {
		t.Fatal("expected non-zero offset for CRLF frontmatter")
	}
	rest := input[end:]
	if !strings.HasPrefix(rest, "\r\n# Body") {
		t.Errorf("expected rest to start with '\\r\\n# Body', got: %q", rest)
	}
}

func TestFindFrontmatterEnd_FrontmatterOnly(t *testing.T) {
	input := "---\ntitle: Test\n---\n"
	end := FindFrontmatterEnd(input)
	if end != len(input) {
		t.Errorf("expected end=%d (full length), got %d", len(input), end)
	}
}

func TestFindFrontmatterEnd_MissingClosingDelimiter(t *testing.T) {
	input := "---\ntitle: Test\n\n# Body without closing"
	if got := FindFrontmatterEnd(input); got != 0 {
		t.Errorf("expected 0 for unclosed frontmatter, got %d", got)
	}
}

func TestFindFrontmatterEnd_EmptyFile(t *testing.T) {
	if got := FindFrontmatterEnd(""); got != 0 {
		t.Errorf("expected 0 for empty content, got %d", got)
	}
}

// --- Append ---

func TestAppend_AddsContentToEndOfFile(t *testing.T) {
	path := writeTempFile(t, "# Note\n\nExisting body.\n")
	m := New()

	if err := m.Append(path, "New line appended."); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readFile(t, path)
	if !strings.HasSuffix(got, "New line appended.\n") {
		t.Errorf("expected suffix 'New line appended.\\n', got: %q", got)
	}
	if !strings.Contains(got, "Existing body.") {
		t.Error("existing body should be preserved")
	}
}

func TestAppend_AddsNewlineBeforeContent(t *testing.T) {
	// File without trailing newline
	path := writeTempFile(t, "body")
	m := New()

	if err := m.Append(path, "appended"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readFile(t, path)
	if !strings.Contains(got, "body\nappended\n") {
		t.Errorf("expected newline separator, got: %q", got)
	}
}

func TestAppend_EmptyContent_ReturnsError(t *testing.T) {
	path := writeTempFile(t, "body\n")
	m := New()

	if err := m.Append(path, ""); err == nil {
		t.Error("expected error for empty content, got nil")
	}
}

func TestAppend_WhitespaceContent_ReturnsError(t *testing.T) {
	path := writeTempFile(t, "body\n")
	m := New()

	if err := m.Append(path, "   \n\t  "); err == nil {
		t.Error("expected error for whitespace-only content, got nil")
	}
}

func TestAppend_NonExistentFile_ReturnsError(t *testing.T) {
	m := New()
	err := m.Append("/nonexistent/path/note.md", "content")
	if err == nil {
		t.Error("expected error for non-existent file, got nil")
	}
}

// --- Prepend ---

func TestPrepend_NoFrontmatter_InsertsAtTop(t *testing.T) {
	path := writeTempFile(t, "# Body\n\nSome content.\n")
	m := New()

	if err := m.Prepend(path, "Prepended line."); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readFile(t, path)
	if !strings.HasPrefix(got, "Prepended line.\n") {
		t.Errorf("expected prepended content at top, got: %q", got)
	}
	if !strings.Contains(got, "# Body") {
		t.Error("body should be preserved")
	}
}

func TestPrepend_WithFrontmatter_InsertsAfterFrontmatter(t *testing.T) {
	initial := "---\ntitle: Test\ndate: 2026-01-01\n---\n\n# Body\n"
	path := writeTempFile(t, initial)
	m := New()

	if err := m.Prepend(path, "Prepended line."); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readFile(t, path)
	if !strings.HasPrefix(got, "---\n") {
		t.Error("frontmatter should be preserved at start")
	}
	// Frontmatter end followed by prepended content
	if !strings.Contains(got, "---\n\nPrepended line.\n") {
		t.Errorf("expected prepended content after frontmatter, got: %q", got)
	}
	if !strings.Contains(got, "# Body") {
		t.Error("body should be preserved")
	}
}

func TestPrepend_EmptyContent_ReturnsError(t *testing.T) {
	path := writeTempFile(t, "body\n")
	m := New()

	if err := m.Prepend(path, ""); err == nil {
		t.Error("expected error for empty content, got nil")
	}
}

func TestPrepend_NonExistentFile_ReturnsError(t *testing.T) {
	m := New()
	if err := m.Prepend("/nonexistent/note.md", "text"); err == nil {
		t.Error("expected error for non-existent file, got nil")
	}
}

// --- Snippet ---

func TestSnippet_ShortContent(t *testing.T) {
	got := Snippet("hello world", 50)
	if got != "hello world" {
		t.Errorf("expected full text, got: %q", got)
	}
}

func TestSnippet_LongContent_Truncated(t *testing.T) {
	content := strings.Repeat("a", 200)
	got := Snippet(content, 50)
	if len(got) > 53 { // 50 chars + "..."
		t.Errorf("expected truncated snippet, got length %d", len(got))
	}
	if !strings.HasSuffix(got, "...") {
		t.Errorf("expected '...' suffix, got: %q", got)
	}
}

func TestSnippet_MultilineContent_ShowsFirstLine(t *testing.T) {
	content := "First line.\nSecond line.\nThird line."
	got := Snippet(content, 50)
	if strings.Contains(got, "\n") {
		t.Errorf("expected no newlines in snippet, got: %q", got)
	}
}

// --- helpers ---

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "note.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	return path
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	return string(data)
}
