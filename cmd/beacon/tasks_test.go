package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTasksCommand_ListsPendingTasks(t *testing.T) {
	requireTasksRipgrep(t)

	vaultPath := t.TempDir()
	writeTasksNote(t, vaultPath, "Inbox.md", "- [ ] first\n- [ ] second\n")
	configPath := writeTasksConfig(t, vaultPath)

	output := executeTasksCommand(t, "--config", configPath, "tasks")

	assert.Contains(t, output, "Inbox.md:1 first")
	assert.Contains(t, output, "Inbox.md:2 second")
}

func TestTasksCommand_PrintsRelativePathAndLine(t *testing.T) {
	requireTasksRipgrep(t)

	vaultPath := t.TempDir()
	writeTasksNote(t, vaultPath, filepath.Join("notes", "Roadmap.md"), "# Title\n- [ ] ship\n")
	configPath := writeTasksConfig(t, vaultPath)

	output := executeTasksCommand(t, "--config", configPath, "tasks")

	assert.Contains(t, output, filepath.Join("notes", "Roadmap.md")+":2 ship")
}

func TestTasksCommand_NoResults(t *testing.T) {
	requireTasksRipgrep(t)

	vaultPath := t.TempDir()
	writeTasksNote(t, vaultPath, "Empty.md", "plain text\n- [x] done\n")
	configPath := writeTasksConfig(t, vaultPath)

	output := executeTasksCommand(t, "--config", configPath, "tasks")

	assert.Equal(t, "No tasks found.\n", output)
}

func TestTasksCommand_UsesIgnorePatternsFromConfig(t *testing.T) {
	requireTasksRipgrep(t)

	vaultPath := t.TempDir()
	writeTasksNote(t, vaultPath, filepath.Join("ignored", "Skip.md"), "- [ ] hidden\n")
	writeTasksNote(t, vaultPath, "Show.md", "- [ ] visible\n")
	configPath := writeTasksConfigWithIgnore(t, vaultPath, []string{"ignored"})

	output := executeTasksCommand(t, "--config", configPath, "tasks")

	assert.Contains(t, output, "Show.md:1 visible")
	assert.NotContains(t, output, "Skip.md")
}

func TestTasksCommand_FailsWhenRipgrepIsUnavailable(t *testing.T) {
	vaultPath := t.TempDir()
	writeTasksNote(t, vaultPath, "Inbox.md", "- [ ] first\n")
	configPath := writeTasksConfig(t, vaultPath)
	t.Setenv("PATH", "")

	err := executeTasksCommandExpectError(t, "--config", configPath, "tasks")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ripgrep")
}

func TestTasksCommand_ListsTasksInDeterministicOrder(t *testing.T) {
	requireTasksRipgrep(t)

	vaultPath := t.TempDir()
	writeTasksNote(t, vaultPath, "b.md", "- [ ] second file\n")
	writeTasksNote(t, vaultPath, "a.md", "- [ ] first file\n")
	configPath := writeTasksConfig(t, vaultPath)

	output := executeTasksCommand(t, "--config", configPath, "tasks")

	assert.Equal(t, "a.md:1 first file\nb.md:1 second file\n", output)
}

func TestTasksCommand_PreservesTaskTextAfterCheckbox(t *testing.T) {
	requireTasksRipgrep(t)

	vaultPath := t.TempDir()
	writeTasksNote(t, vaultPath, "Colon.md", "- [ ] call API: review payload\n")
	configPath := writeTasksConfig(t, vaultPath)

	output := executeTasksCommand(t, "--config", configPath, "tasks")

	assert.Contains(t, output, "Colon.md:1 call API: review payload")
}

func TestTasksCommand_EmitsOneLinePerTask(t *testing.T) {
	requireTasksRipgrep(t)

	vaultPath := t.TempDir()
	writeTasksNote(t, vaultPath, "Inbox.md", "- [ ] one\n- [ ] two\n")
	configPath := writeTasksConfig(t, vaultPath)

	output := executeTasksCommand(t, "--config", configPath, "tasks")

	assert.Equal(t, "Inbox.md:1 one\nInbox.md:2 two\n", output)
}

func TestTasksCommand_RejectsUnexpectedArguments(t *testing.T) {
	requireTasksRipgrep(t)

	vaultPath := t.TempDir()
	writeTasksNote(t, vaultPath, "Inbox.md", "- [ ] one\n")
	configPath := writeTasksConfig(t, vaultPath)

	err := executeTasksCommandExpectError(t, "--config", configPath, "tasks", "extra")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown command")
}

func requireTasksRipgrep(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("rg"); err != nil {
		t.Skip("ripgrep not installed, skipping test")
	}
}

func writeTasksConfig(t *testing.T, vaultPath string) string {
	t.Helper()

	configPath := filepath.Join(t.TempDir(), "beacon.yml")
	content := []byte("vault_path: " + vaultPath + "\n")
	require.NoError(t, os.WriteFile(configPath, content, 0o644))
	return configPath
}

func writeTasksConfigWithIgnore(t *testing.T, vaultPath string, ignore []string) string {
	t.Helper()

	configPath := filepath.Join(t.TempDir(), "beacon.yml")
	content := "vault_path: " + vaultPath + "\n"
	if len(ignore) > 0 {
		content += "ignore:\n"
		for _, pattern := range ignore {
			content += "  - " + pattern + "\n"
		}
	}
	require.NoError(t, os.WriteFile(configPath, []byte(content), 0o644))
	return configPath
}

func writeTasksNote(t *testing.T, vaultPath, relPath, content string) {
	t.Helper()

	fullPath := filepath.Join(vaultPath, relPath)
	require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0o755))
	require.NoError(t, os.WriteFile(fullPath, []byte(content), 0o644))
}

func executeTasksCommand(t *testing.T, args ...string) string {
	t.Helper()

	output, err := executeTasksCommandWithError(args...)
	require.NoError(t, err)
	return output
}

func executeTasksCommandExpectError(t *testing.T, args ...string) error {
	t.Helper()

	_, err := executeTasksCommandWithError(args...)
	return err
}

func executeTasksCommandWithError(args ...string) (string, error) {
	resetTasksCommandState()

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

func resetTasksCommandState() {
	cfgFile = ""
	rootCmd.SetOut(bytes.NewBuffer(nil))
	rootCmd.SetErr(bytes.NewBuffer(nil))
}
