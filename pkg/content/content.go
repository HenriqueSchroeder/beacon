package content

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Manipulator handles content insertion into existing markdown files.
type Manipulator struct{}

// New creates a Manipulator.
func New() *Manipulator {
	return &Manipulator{}
}

// Append adds text to the end of the file at path.
// A newline is added before the text if the file does not already end with one.
// Returns an error if text is empty or whitespace-only.
func (m *Manipulator) Append(path, text string) error {
	if strings.TrimSpace(text) == "" {
		return fmt.Errorf("content: text must not be empty")
	}

	existing, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("content: failed to read file: %w", err)
	}

	var b strings.Builder
	b.Write(existing)

	// Ensure separator newline
	if len(existing) > 0 && existing[len(existing)-1] != '\n' {
		b.WriteByte('\n')
	}

	b.WriteString(text)
	if !strings.HasSuffix(text, "\n") {
		b.WriteByte('\n')
	}

	return atomicWrite(path, []byte(b.String()))
}

// Prepend inserts text after the YAML frontmatter block (if any), or at the
// beginning of the file when no frontmatter is present.
// Returns an error if text is empty or whitespace-only.
func (m *Manipulator) Prepend(path, text string) error {
	if strings.TrimSpace(text) == "" {
		return fmt.Errorf("content: text must not be empty")
	}

	existing, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("content: failed to read file: %w", err)
	}

	content := string(existing)
	insertAt := FindFrontmatterEnd(content)

	// Skip the blank line immediately after closing --- so that the prepended
	// content lands after the separator blank line, not before it.
	if insertAt > 0 && insertAt < len(content) && content[insertAt] == '\n' {
		insertAt++
	}

	// Ensure the inserted text ends with a newline
	if !strings.HasSuffix(text, "\n") {
		text += "\n"
	}

	result := content[:insertAt] + text + content[insertAt:]
	return atomicWrite(path, []byte(result))
}

// FindFrontmatterEnd returns the byte offset in content immediately after the
// closing --- of a YAML frontmatter block (including its trailing newline).
// Returns 0 if the content does not start with a valid frontmatter block.
// Both LF (\n) and CRLF (\r\n) line endings are supported.
func FindFrontmatterEnd(content string) int {
	// Detect line ending style from the opening delimiter
	var nl string
	switch {
	case strings.HasPrefix(content, "---\r\n"):
		nl = "\r\n"
	case strings.HasPrefix(content, "---\n"):
		nl = "\n"
	default:
		return 0
	}

	openDelim := "---" + nl
	closePattern := nl + "---" + nl

	rest := content[len(openDelim):]
	idx := strings.Index(rest, closePattern)
	if idx == -1 {
		return 0
	}

	return len(openDelim) + idx + len(closePattern)
}

// Snippet returns a single-line preview of text, truncated to maxLen characters.
func Snippet(text string, maxLen int) string {
	// Take the first line only
	line := text
	if nl := strings.IndexByte(text, '\n'); nl >= 0 {
		line = text[:nl]
	}

	line = strings.TrimSpace(line)
	if len(line) <= maxLen {
		return line
	}
	return line[:maxLen] + "..."
}

// atomicWrite writes data to path using a temporary file and rename to avoid
// partial writes on crash. The original file's permissions are preserved.
func atomicWrite(path string, data []byte) error {
	// Read original permissions before touching the file
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("content: failed to stat file: %w", err)
	}
	perm := info.Mode().Perm()

	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".beacon-tmp-*")
	if err != nil {
		return fmt.Errorf("content: failed to create temp file: %w", err)
	}

	tmpPath := tmp.Name()
	success := false
	defer func() {
		tmp.Close()
		if !success {
			os.Remove(tmpPath)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		return fmt.Errorf("content: failed to write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("content: failed to close temp file: %w", err)
	}

	if err := os.Chmod(tmpPath, perm); err != nil {
		return fmt.Errorf("content: failed to set file permissions: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("content: failed to replace file: %w", err)
	}

	success = true
	return nil
}
