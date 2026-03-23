package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShowCommand_PrintsRawNoteContent(t *testing.T) {
	vaultPath := t.TempDir()
	writeShowCommandNote(t, vaultPath, "note.md", "---\nstatus: active\n---\n# Note\nBody\n")
	configPath := writeShowConfig(t, vaultPath)

	output := executeShowCommand(t, "--config", configPath, "show", "note")

	assert.Equal(t, "---\nstatus: active\n---\n# Note\nBody\n", output)
}

func TestShowCommand_NoFrontmatter(t *testing.T) {
	vaultPath := t.TempDir()
	writeShowCommandNote(t, vaultPath, "note.md", "---\nstatus: active\n---\n# Note\nBody\n")
	configPath := writeShowConfig(t, vaultPath)

	output := executeShowCommand(t, "--config", configPath, "show", "note", "--no-frontmatter")

	assert.Equal(t, "# Note\nBody\n", output)
}

func TestShowCommand_ResolvesNoteByTitle(t *testing.T) {
	vaultPath := t.TempDir()
	writeShowCommandNote(t, vaultPath, "target-note.md", "# Target Note\nBody\n")
	configPath := writeShowConfig(t, vaultPath)

	output := executeShowCommand(t, "--config", configPath, "show", "Target Note")

	assert.Equal(t, "# Target Note\nBody\n", output)
}

func TestShowCommand_RejectsMissingArgument(t *testing.T) {
	vaultPath := t.TempDir()
	configPath := writeShowConfig(t, vaultPath)

	err := executeShowCommandExpectError(t, "--config", configPath, "show")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}

func TestShowCommand_ReportsAmbiguousTarget(t *testing.T) {
	vaultPath := t.TempDir()
	writeShowCommandNote(t, vaultPath, filepath.Join("a", "note.md"), "# One\n")
	writeShowCommandNote(t, vaultPath, filepath.Join("b", "note.md"), "# Two\n")
	configPath := writeShowConfig(t, vaultPath)

	err := executeShowCommandExpectError(t, "--config", configPath, "show", "note")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ambiguous note target")
}

func TestShowCommand_NoTrailingBannerOrWrapperText(t *testing.T) {
	vaultPath := t.TempDir()
	writeShowCommandNote(t, vaultPath, "note.md", "# Note\nBody\n")
	configPath := writeShowConfig(t, vaultPath)

	output := executeShowCommand(t, "--config", configPath, "show", "note")

	assert.NotContains(t, output, "Note created")
	assert.NotContains(t, output, "show:")
	assert.Equal(t, "# Note\nBody\n", output)
}

func TestShowCommand_DoesNotAppendTrailingNewline(t *testing.T) {
	vaultPath := t.TempDir()
	writeShowCommandNote(t, vaultPath, "note.md", "# Note\nBody")
	configPath := writeShowConfig(t, vaultPath)

	output := executeShowCommand(t, "--config", configPath, "show", "note")

	assert.Equal(t, "# Note\nBody", output)
}

func TestShowCommand_PreservesBlankLines(t *testing.T) {
	vaultPath := t.TempDir()
	writeShowCommandNote(t, vaultPath, "note.md", "# Note\n\nBody\n\nTail\n")
	configPath := writeShowConfig(t, vaultPath)

	output := executeShowCommand(t, "--config", configPath, "show", "note")

	assert.Equal(t, "# Note\n\nBody\n\nTail\n", output)
}

func TestShowCommand_SupportsSubdirectoryPathWithoutExtension(t *testing.T) {
	vaultPath := t.TempDir()
	writeShowCommandNote(t, vaultPath, filepath.Join("notes", "daily.md"), "# Daily\n")
	configPath := writeShowConfig(t, vaultPath)

	output := executeShowCommand(t, "--config", configPath, "show", filepath.Join("notes", "daily"))

	assert.Equal(t, "# Daily\n", output)
}

func TestShowCommand_PreservesContentWithoutTerminalNewline(t *testing.T) {
	vaultPath := t.TempDir()
	writeShowCommandNote(t, vaultPath, "note.md", "# Note")
	configPath := writeShowConfig(t, vaultPath)

	output := executeShowCommand(t, "--config", configPath, "show", "note")

	assert.Equal(t, "# Note", output)
}

func TestShowCommand_PreservesCRLFOutput(t *testing.T) {
	vaultPath := t.TempDir()
	writeShowCommandNote(t, vaultPath, "note.md", "---\r\nstatus: active\r\n---\r\n\r\n# Note\r\nBody\r\n")
	configPath := writeShowConfig(t, vaultPath)

	output := executeShowCommand(t, "--config", configPath, "show", "note")

	assert.Equal(t, "---\r\nstatus: active\r\n---\r\n\r\n# Note\r\nBody\r\n", output)
}

func TestShowCommand_PreservesCRLFWithoutNormalization(t *testing.T) {
	vaultPath := t.TempDir()
	writeShowCommandNote(t, vaultPath, "note.md", "---\r\nstatus: active\r\n---\r\n\r\n# Note\r\nBody\r\n")
	configPath := writeShowConfig(t, vaultPath)

	output := executeShowCommand(t, "--config", configPath, "show", "note", "--no-frontmatter")

	assert.Equal(t, "\r\n# Note\r\nBody\r\n", output)
}

func writeShowConfig(t *testing.T, vaultPath string) string {
	t.Helper()

	configPath := filepath.Join(t.TempDir(), "beacon.yml")
	require.NoError(t, os.WriteFile(configPath, []byte("vault_path: "+vaultPath+"\n"), 0o644))
	return configPath
}

func writeShowCommandNote(t *testing.T, vaultPath, relPath, content string) {
	t.Helper()

	fullPath := filepath.Join(vaultPath, relPath)
	require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0o755))
	require.NoError(t, os.WriteFile(fullPath, []byte(content), 0o644))
}

func executeShowCommand(t *testing.T, args ...string) string {
	t.Helper()

	output, err := executeShowCommandWithError(args...)
	require.NoError(t, err)
	return output
}

func executeShowCommandExpectError(t *testing.T, args ...string) error {
	t.Helper()

	_, err := executeShowCommandWithError(args...)
	return err
}

func executeShowCommandWithError(args ...string) (string, error) {
	resetShowCommandState()

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

func resetShowCommandState() {
	cfgFile = ""
	showNoFrontmatter = false
	rootCmd.SetOut(bytes.NewBuffer(nil))
	rootCmd.SetErr(bytes.NewBuffer(nil))
}
