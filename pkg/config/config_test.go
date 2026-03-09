package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFromFile(t *testing.T) {
	cfg, err := LoadFrom("../../testdata/fixtures/config/valid.yml")

	require.NoError(t, err)
	assert.Equal(t, "/tmp/test-vault", cfg.VaultPath)
	assert.Equal(t, "nvim", cfg.Editor)
	assert.Equal(t, []string{".obsidian", "*.tmp"}, cfg.Ignore)
}

func TestLoadFromFile_Minimal(t *testing.T) {
	cfg, err := LoadFrom("../../testdata/fixtures/config/minimal.yml")

	require.NoError(t, err)
	assert.Equal(t, "/tmp/test-vault", cfg.VaultPath)
	assert.Equal(t, "vim", cfg.Editor, "should use default editor")
	assert.Equal(t, []string{".obsidian"}, cfg.Ignore, "should use default ignore")
}

func TestLoadFromFile_NotFound(t *testing.T) {
	_, err := LoadFrom("/nonexistent/path.yml")

	assert.Error(t, err)
}

func TestLoadFromFile_InvalidYAML(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "bad.yml")
	os.WriteFile(path, []byte(":\ninvalid: [\nyaml"), 0644)

	_, err := LoadFrom(path)

	assert.Error(t, err)
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("BEACON_VAULT_PATH", "/tmp/env-vault")

	cfg, err := LoadFrom("")

	require.NoError(t, err)
	assert.Equal(t, "/tmp/env-vault", cfg.VaultPath)
}

func TestLoadFromEnv_NoConfig(t *testing.T) {
	_, err := LoadFrom("")

	assert.Error(t, err, "should error when no vault_path is set")
}
