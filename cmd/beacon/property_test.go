package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPropertyGetCommand_PrintsScalarValue(t *testing.T) {
	vaultPath := t.TempDir()
	writePropertyCommandNote(t, vaultPath, "note.md", "---\nstatus: done\n---\n# Note\n")
	configPath := writePropertyConfig(t, vaultPath)

	output, err := executePropertyCommandWithError("--config", configPath, "property", "get", "status", "note.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.TrimSpace(output) != "done" {
		t.Fatalf("expected done, got %q", output)
	}
}

func TestPropertySetCommand_UpdatesNote(t *testing.T) {
	vaultPath := t.TempDir()
	writePropertyCommandNote(t, vaultPath, "note.md", "---\nstatus: todo\n---\n# Note\n")
	configPath := writePropertyConfig(t, vaultPath)

	output, err := executePropertyCommandWithError("--config", configPath, "property", "set", "status", "done", "note.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "status") {
		t.Fatalf("expected confirmation output, got %q", output)
	}
	if !strings.Contains(readPropertyCommandNote(t, vaultPath, "note.md"), "status: done") {
		t.Fatalf("expected note to be updated")
	}
}

func TestPropertyAddCommand_AppendsTagWithoutDuplicating(t *testing.T) {
	vaultPath := t.TempDir()
	writePropertyCommandNote(t, vaultPath, "note.md", "---\ntags:\n  - work\n---\n# Note\n")
	configPath := writePropertyConfig(t, vaultPath)

	if _, err := executePropertyCommandWithError("--config", configPath, "property", "add", "tags", "urgent", "note.md"); err != nil {
		t.Fatalf("unexpected error on first add: %v", err)
	}
	if _, err := executePropertyCommandWithError("--config", configPath, "property", "add", "tags", "urgent", "note.md"); err != nil {
		t.Fatalf("unexpected error on duplicate add: %v", err)
	}

	got := readPropertyCommandNote(t, vaultPath, "note.md")
	if strings.Count(got, "- urgent") != 1 {
		t.Fatalf("expected one urgent tag, got %q", got)
	}
}

func TestPropertyCommand_RejectsMissingNotePath(t *testing.T) {
	vaultPath := t.TempDir()
	configPath := writePropertyConfig(t, vaultPath)

	_, err := executePropertyCommandWithError("--config", configPath, "property", "get", "status")
	if err == nil {
		t.Fatal("expected error for missing note path")
	}
}

func TestPropertyAddCommand_RejectsNonListField(t *testing.T) {
	vaultPath := t.TempDir()
	writePropertyCommandNote(t, vaultPath, "note.md", "---\nstatus: todo\n---\n# Note\n")
	configPath := writePropertyConfig(t, vaultPath)

	_, err := executePropertyCommandWithError("--config", configPath, "property", "add", "status", "done", "note.md")
	if err == nil {
		t.Fatal("expected error when adding to non-list field")
	}
}

func TestPropertySetCommand_CreatesFrontmatterOnPlainMarkdownNote(t *testing.T) {
	vaultPath := t.TempDir()
	writePropertyCommandNote(t, vaultPath, "note.md", "# Note\nBody\n")
	configPath := writePropertyConfig(t, vaultPath)

	if _, err := executePropertyCommandWithError("--config", configPath, "property", "set", "status", "done", "note.md"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readPropertyCommandNote(t, vaultPath, "note.md")
	if !strings.HasPrefix(got, "---\nstatus: done\n---\n") {
		t.Fatalf("expected created frontmatter, got %q", got)
	}
}

func TestPropertyGetCommand_PrintsYAMLList(t *testing.T) {
	vaultPath := t.TempDir()
	writePropertyCommandNote(t, vaultPath, "note.md", "---\ntags:\n  - work\n  - urgent\n---\n# Note\n")
	configPath := writePropertyConfig(t, vaultPath)

	output, err := executePropertyCommandWithError("--config", configPath, "property", "get", "tags", "note.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.TrimSpace(output) != "- work\n- urgent" {
		t.Fatalf("expected yaml list output, got %q", output)
	}
}

func TestPropertyGetCommand_ErrorsWhenKeyMissing(t *testing.T) {
	vaultPath := t.TempDir()
	writePropertyCommandNote(t, vaultPath, "note.md", "---\nstatus: done\n---\n# Note\n")
	configPath := writePropertyConfig(t, vaultPath)

	_, err := executePropertyCommandWithError("--config", configPath, "property", "get", "owner", "note.md")
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestPropertyGetCommand_PrintsTimestampAsYAMLScalar(t *testing.T) {
	vaultPath := t.TempDir()
	writePropertyCommandNote(t, vaultPath, "note.md", "---\ndate: 2026-03-22\n---\n# Note\n")
	configPath := writePropertyConfig(t, vaultPath)

	output, err := executePropertyCommandWithError("--config", configPath, "property", "get", "date", "note.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.TrimSpace(output) != "2026-03-22" {
		t.Fatalf("expected yaml scalar date, got %q", output)
	}
}

func TestPropertySetCommand_RejectsPathOutsideVault(t *testing.T) {
	vaultPath := t.TempDir()
	configPath := writePropertyConfig(t, vaultPath)

	_, err := executePropertyCommandWithError("--config", configPath, "property", "set", "status", "done", "../outside.md")
	if err == nil {
		t.Fatal("expected error for path outside vault")
	}
}

func TestPropertySetCommand_RejectsNonMarkdownPath(t *testing.T) {
	vaultPath := t.TempDir()
	configPath := writePropertyConfig(t, vaultPath)

	_, err := executePropertyCommandWithError("--config", configPath, "property", "set", "status", "done", "note.txt")
	if err == nil {
		t.Fatal("expected error for non-markdown path")
	}
}

func writePropertyConfig(t *testing.T, vaultPath string) string {
	t.Helper()

	configPath := filepath.Join(t.TempDir(), "beacon.yml")
	if err := os.WriteFile(configPath, []byte("vault_path: "+vaultPath+"\n"), 0o644); err != nil {
		t.Fatalf("write config failed: %v", err)
	}
	return configPath
}

func writePropertyCommandNote(t *testing.T, vaultPath, relPath, content string) {
	t.Helper()

	fullPath := filepath.Join(vaultPath, relPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write note failed: %v", err)
	}
}

func readPropertyCommandNote(t *testing.T, vaultPath, relPath string) string {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(vaultPath, relPath))
	if err != nil {
		t.Fatalf("read note failed: %v", err)
	}
	return string(data)
}

func executePropertyCommandWithError(args ...string) (string, error) {
	resetPropertyCommandState()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}

	os.Stdout = w
	rootCmd.SetArgs(args)
	err = rootCmd.Execute()
	rootCmd.SetArgs(nil)
	_ = w.Close()
	os.Stdout = oldStdout

	var output bytes.Buffer
	_, readErr := output.ReadFrom(r)
	_ = r.Close()
	if readErr != nil {
		return "", readErr
	}

	return output.String(), err
}

func resetPropertyCommandState() {
	cfgFile = ""
	rootCmd.SetOut(bytes.NewBuffer(nil))
	rootCmd.SetErr(bytes.NewBuffer(nil))
}
