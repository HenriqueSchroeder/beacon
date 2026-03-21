package main

import (
	"bytes"
	"encoding/json"
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

func TestSearchCommand_RelatedSearch(t *testing.T) {
	vaultPath := t.TempDir()
	writeNote(t, vaultPath, "Target Note.md", "# Target Note\n")
	writeNote(t, vaultPath, "Source.md", "Mentions [[Target Note]] here.\n")

	configPath := writeSearchConfig(t, vaultPath)
	output := executeRootCommand(t, "--config", configPath, "search", "--related", "Target Note")

	assert.Contains(t, output, "Source.md:1")
	assert.Contains(t, output, "[[Target Note]]")
}

func TestSearchCommand_RelatedSearchJSON(t *testing.T) {
	vaultPath := t.TempDir()
	writeNote(t, vaultPath, "Target Note.md", "# Target Note\n")
	writeNote(t, vaultPath, "Source.md", "Mentions [[Target Note|target]].\n")

	configPath := writeSearchConfig(t, vaultPath)
	output := executeRootCommand(t, "--config", configPath, "search", "--related", "--json", "Target Note")

	var results []map[string]any
	require.NoError(t, json.Unmarshal([]byte(output), &results))
	require.Len(t, results, 1)
	assert.Equal(t, "Source.md", results[0]["path"])
	assert.Equal(t, float64(1), results[0]["line"])
}

func TestSearchCommand_RelatedSearchRejectsCombinedModes(t *testing.T) {
	vaultPath := t.TempDir()
	writeNote(t, vaultPath, "Target Note.md", "# Target Note\n")
	configPath := writeSearchConfig(t, vaultPath)

	err := executeRootCommandExpectError(t, "--config", configPath, "search", "--related", "--tags", "work", "Target Note")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "only one search mode")
}

func TestSearchCommand_RelatedSearchRequiresArgument(t *testing.T) {
	vaultPath := t.TempDir()
	configPath := writeSearchConfig(t, vaultPath)

	err := executeRootCommandExpectError(t, "--config", configPath, "search", "--related")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--related requires a note argument")
}

func TestSearchCommand_RelatedSearchAmbiguousTarget(t *testing.T) {
	vaultPath := t.TempDir()
	writeNote(t, vaultPath, "Target.md", "# Target\n")
	writeNote(t, vaultPath, filepath.Join("subdir", "Target.md"), "# Target\n")
	configPath := writeSearchConfig(t, vaultPath)

	err := executeRootCommandExpectError(t, "--config", configPath, "search", "--related", "Target")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ambiguous note target")
	assert.Contains(t, err.Error(), "Target.md")
	assert.Contains(t, err.Error(), filepath.Join("subdir", "Target.md"))
}

func writeSearchConfig(t *testing.T, vaultPath string) string {
	t.Helper()

	configPath := filepath.Join(t.TempDir(), "beacon.yml")
	content := []byte("vault_path: " + vaultPath + "\n")
	require.NoError(t, os.WriteFile(configPath, content, 0o644))
	return configPath
}

func writeNote(t *testing.T, vaultPath, relPath, content string) {
	t.Helper()

	fullPath := filepath.Join(vaultPath, relPath)
	require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0o755))
	require.NoError(t, os.WriteFile(fullPath, []byte(content), 0o644))
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
	searchRelated = false
	rootCmd.SetOut(bytes.NewBuffer(nil))
	rootCmd.SetErr(bytes.NewBuffer(nil))
}
