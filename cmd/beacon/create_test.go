package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateCommand_ContentAloneStillFailsIfNoteExists(t *testing.T) {
	vaultPath := t.TempDir()
	writeCreateNote(t, vaultPath, "Inbox.md", "existing\n")
	configPath := writeCreateConfig(t, vaultPath)

	err := executeCreateCommandExpectError(t, "--config", configPath, "create", "Inbox", "--content", "new line")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "file already exists")
	assert.Equal(t, "existing\n", readCreateNote(t, vaultPath, "Inbox.md"))
}

func TestCreateCommand_PrependCreatesAndAddsContent(t *testing.T) {
	vaultPath := t.TempDir()
	configPath := writeCreateConfig(t, vaultPath)

	output := executeCreateCommand(t, "--config", configPath, "create", "Inbox", "--content", "top", "--prepend")

	assert.Contains(t, output, "Note created")
	assert.Contains(t, output, "Prepended to")
	assert.Contains(t, readCreateNote(t, vaultPath, "Inbox.md"), "top\n# Inbox\n")
}

func writeCreateConfig(t *testing.T, vaultPath string) string {
	t.Helper()

	configPath := filepath.Join(t.TempDir(), "beacon.yml")
	content := []byte("vault_path: " + vaultPath + "\n")
	require.NoError(t, os.WriteFile(configPath, content, 0o644))
	return configPath
}

func writeCreateNote(t *testing.T, vaultPath, relPath, content string) {
	t.Helper()

	fullPath := filepath.Join(vaultPath, relPath)
	require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0o755))
	require.NoError(t, os.WriteFile(fullPath, []byte(content), 0o644))
}

func readCreateNote(t *testing.T, vaultPath, relPath string) string {
	t.Helper()

	fullPath := filepath.Join(vaultPath, relPath)
	data, err := os.ReadFile(fullPath)
	require.NoError(t, err)
	return string(data)
}

func executeCreateCommand(t *testing.T, args ...string) string {
	t.Helper()

	output, err := executeCreateCommandWithError(args...)
	require.NoError(t, err)
	return output
}

func executeCreateCommandExpectError(t *testing.T, args ...string) error {
	t.Helper()

	_, err := executeCreateCommandWithError(args...)
	return err
}

func executeCreateCommandWithError(args ...string) (string, error) {
	resetCreateCommandState()

	oldStdout := os.Stdout
	oldStdin := os.Stdin
	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		return "", err
	}
	stdinR, stdinW, err := os.Pipe()
	if err != nil {
		_ = stdoutR.Close()
		_ = stdoutW.Close()
		return "", err
	}

	_ = stdinW.Close()
	os.Stdout = stdoutW
	os.Stdin = stdinR
	rootCmd.SetArgs(args)
	err = rootCmd.Execute()
	rootCmd.SetArgs(nil)
	_ = stdoutW.Close()
	os.Stdout = oldStdout
	os.Stdin = oldStdin

	var output bytes.Buffer
	_, readErr := output.ReadFrom(stdoutR)
	_ = stdoutR.Close()
	_ = stdinR.Close()
	if readErr != nil {
		return "", readErr
	}

	return output.String(), err
}

func resetCreateCommandState() {
	cfgFile = ""
	createType = ""
	createTemplate = ""
	createTags = ""
	createPath = ""
	createOverwrite = false
	createContent = ""
	createAppend = false
	createPrepend = false
	rootCmd.SetOut(bytes.NewBuffer(nil))
	rootCmd.SetErr(bytes.NewBuffer(nil))
}
