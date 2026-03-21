package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchCommand_FilenameSearch(t *testing.T) {
	vaultPath := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(vaultPath, "Project_Plan.md"), []byte("# Test\n"), 0o644))

	configPath := writeSearchConfig(t, vaultPath)
	output := executeRootCommand(t, "--config", configPath, "search", "--filename", "Project Plan")

	assert.Contains(t, output, "Project_Plan.md")
	assert.Contains(t, output, "> Project_Plan")
}

func TestSearchCommand_FilenameSearchRejectsCombinedModes(t *testing.T) {
	vaultPath := t.TempDir()
	configPath := writeSearchConfig(t, vaultPath)

	err := executeRootCommandExpectError(t, "--config", configPath, "search", "--filename", "--tags", "work", "plan")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "only one search mode")
}

func writeSearchConfig(t *testing.T, vaultPath string) string {
	t.Helper()

	configPath := filepath.Join(t.TempDir(), "beacon.yml")
	content := []byte("vault_path: " + vaultPath + "\n")
	require.NoError(t, os.WriteFile(configPath, content, 0o644))
	return configPath
}

func executeRootCommand(t *testing.T, args ...string) string {
	t.Helper()

	output, err := executeRootCommandWithError(args...)
	require.NoError(t, err)
	return output
}

func executeRootCommandExpectError(t *testing.T, args ...string) error {
	t.Helper()

	_, err := executeRootCommandWithError(args...)
	return err
}

func executeRootCommandWithError(args ...string) (string, error) {
	resetSearchCommandState()

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

func resetSearchCommandState() {
	cfgFile = ""
	jsonOutput = false
	searchTags = ""
	searchType = ""
	searchFilename = false
	rootCmd.SetOut(bytes.NewBuffer(nil))
	rootCmd.SetErr(bytes.NewBuffer(nil))
}
