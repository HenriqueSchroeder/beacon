package validate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockPrompter returns pre-configured actions for testing
type mockPrompter struct {
	actions []FixAction
	idx     int
}

func (m *mockPrompter) Prompt(fix Fix, current, total int) FixAction {
	if m.idx >= len(m.actions) {
		return FixActionSkip
	}
	action := m.actions[m.idx]
	m.idx++
	return action
}

func TestCorrectedRaw(t *testing.T) {
	tests := []struct {
		name string
		fix  Fix
		want string
	}{
		{
			name: "simple link",
			fix:  Fix{SuggestedTarget: "Vinicius Dal Magro"},
			want: "[[Vinicius Dal Magro]]",
		},
		{
			name: "link with heading",
			fix:  Fix{SuggestedTarget: "Vinicius Dal Magro", Heading: "Projects"},
			want: "[[Vinicius Dal Magro#Projects]]",
		},
		{
			name: "link with alias",
			fix:  Fix{SuggestedTarget: "Vinicius Dal Magro", Alias: "Vini"},
			want: "[[Vinicius Dal Magro|Vini]]",
		},
		{
			name: "link with heading and alias",
			fix:  Fix{SuggestedTarget: "Vinicius Dal Magro", Heading: "Projects", Alias: "Vini's projects"},
			want: "[[Vinicius Dal Magro#Projects|Vini's projects]]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.fix.CorrectedRaw())
		})
	}
}

func TestCollectFixes(t *testing.T) {
	results := []DocumentValidation{
		{
			FilePath: "note-a.md",
			Results: []ValidationResult{
				{Valid: true},
				{Valid: false, Reason: "not found", SuggestedTarget: "correct-name"},
				{Valid: false, Reason: "malformed link"}, // no suggestion
			},
		},
		{
			FilePath: "note-b.md",
			Results: []ValidationResult{
				{Valid: false, Reason: "not found", SuggestedTarget: "other-name"},
			},
		},
	}

	fixes := CollectFixes(results)
	assert.Len(t, fixes, 2)
	assert.Equal(t, "note-a.md", fixes[0].FilePath)
	assert.Equal(t, "note-b.md", fixes[1].FilePath)
}

func TestCollectFixes_NoFixable(t *testing.T) {
	results := []DocumentValidation{
		{
			FilePath: "note.md",
			Results: []ValidationResult{
				{Valid: true},
				{Valid: false, Reason: "malformed"}, // no SuggestedTarget
			},
		},
	}

	fixes := CollectFixes(results)
	assert.Empty(t, fixes)
}

func setupTempVault(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		path := filepath.Join(dir, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0755))
		require.NoError(t, os.WriteFile(path, []byte(content), 0644))
	}
	return dir
}

func TestFixer_ApplyFixes_ApplyAll(t *testing.T) {
	vaultDir := setupTempVault(t, map[string]string{
		"daily/2024-01-15.md": "Today I talked to [[Vini]] about the project.\nAlso mentioned [[API desing]].\n",
	})

	fixes := []Fix{
		{
			FilePath:        "daily/2024-01-15.md",
			Line:            1,
			OriginalRaw:     "[[Vini]]",
			OriginalTarget:  "Vini",
			SuggestedTarget: "Vinicius Dal Magro",
		},
		{
			FilePath:        "daily/2024-01-15.md",
			Line:            2,
			OriginalRaw:     "[[API desing]]",
			OriginalTarget:  "API desing",
			SuggestedTarget: "API Design",
		},
	}

	prompter := &mockPrompter{actions: []FixAction{FixActionApplyAll}}
	fixer := NewFixer(vaultDir, prompter)

	summary := fixer.ApplyFixes(fixes)

	assert.Equal(t, 2, summary.Applied)
	assert.Equal(t, 0, summary.Skipped)
	assert.Empty(t, summary.Errors)

	content, err := os.ReadFile(filepath.Join(vaultDir, "daily/2024-01-15.md"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "[[Vinicius Dal Magro]]")
	assert.Contains(t, string(content), "[[API Design]]")
}

func TestFixer_ApplyFixes_SkipSome(t *testing.T) {
	vaultDir := setupTempVault(t, map[string]string{
		"note.md": "See [[Vini]] and [[JD]].\n",
	})

	fixes := []Fix{
		{
			FilePath:        "note.md",
			Line:            1,
			OriginalRaw:     "[[Vini]]",
			OriginalTarget:  "Vini",
			SuggestedTarget: "Vinicius",
		},
		{
			FilePath:        "note.md",
			Line:            1,
			OriginalRaw:     "[[JD]]",
			OriginalTarget:  "JD",
			SuggestedTarget: "John Doe",
		},
	}

	prompter := &mockPrompter{actions: []FixAction{FixActionSkip, FixActionApply}}
	fixer := NewFixer(vaultDir, prompter)

	summary := fixer.ApplyFixes(fixes)

	assert.Equal(t, 1, summary.Applied)
	assert.Equal(t, 1, summary.Skipped)

	content, err := os.ReadFile(filepath.Join(vaultDir, "note.md"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "[[Vini]]")    // skipped
	assert.Contains(t, string(content), "[[John Doe]]") // applied
}

func TestFixer_ApplyFixes_Quit(t *testing.T) {
	vaultDir := setupTempVault(t, map[string]string{
		"note.md": "See [[A]] and [[B]] and [[C]].\n",
	})

	fixes := []Fix{
		{FilePath: "note.md", Line: 1, OriginalRaw: "[[A]]", OriginalTarget: "A", SuggestedTarget: "Alpha"},
		{FilePath: "note.md", Line: 1, OriginalRaw: "[[B]]", OriginalTarget: "B", SuggestedTarget: "Beta"},
		{FilePath: "note.md", Line: 1, OriginalRaw: "[[C]]", OriginalTarget: "C", SuggestedTarget: "Charlie"},
	}

	prompter := &mockPrompter{actions: []FixAction{FixActionApply, FixActionQuit}}
	fixer := NewFixer(vaultDir, prompter)

	summary := fixer.ApplyFixes(fixes)

	assert.Equal(t, 1, summary.Applied)
	assert.Equal(t, 2, summary.Skipped) // B and C skipped via quit
}

func TestFixer_ApplyFixes_PreservesHeadingAndAlias(t *testing.T) {
	vaultDir := setupTempVault(t, map[string]string{
		"note.md": "See [[Vini#Projects|my friend]] for details.\n",
	})

	fixes := []Fix{
		{
			FilePath:        "note.md",
			Line:            1,
			OriginalRaw:     "[[Vini#Projects|my friend]]",
			OriginalTarget:  "Vini",
			SuggestedTarget: "Vinicius Dal Magro",
			Heading:         "Projects",
			Alias:           "my friend",
		},
	}

	prompter := &mockPrompter{actions: []FixAction{FixActionApply}}
	fixer := NewFixer(vaultDir, prompter)

	summary := fixer.ApplyFixes(fixes)
	assert.Equal(t, 1, summary.Applied)

	content, err := os.ReadFile(filepath.Join(vaultDir, "note.md"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "[[Vinicius Dal Magro#Projects|my friend]]")
}

func TestFixer_ApplyFixes_FileNotFound(t *testing.T) {
	vaultDir := t.TempDir()

	fixes := []Fix{
		{FilePath: "nonexistent.md", Line: 1, OriginalRaw: "[[X]]", SuggestedTarget: "Y"},
	}

	prompter := &mockPrompter{actions: []FixAction{FixActionApply}}
	fixer := NewFixer(vaultDir, prompter)

	summary := fixer.ApplyFixes(fixes)
	assert.Equal(t, 0, summary.Applied)
	assert.Len(t, summary.Errors, 1)
}

func TestFixer_ApplyFixes_LinkNotOnLine(t *testing.T) {
	vaultDir := setupTempVault(t, map[string]string{
		"note.md": "No links here.\n",
	})

	fixes := []Fix{
		{FilePath: "note.md", Line: 1, OriginalRaw: "[[Ghost]]", SuggestedTarget: "Real"},
	}

	prompter := &mockPrompter{actions: []FixAction{FixActionApply}}
	fixer := NewFixer(vaultDir, prompter)

	summary := fixer.ApplyFixes(fixes)
	assert.Equal(t, 0, summary.Applied)
	assert.Len(t, summary.Errors, 1)
}

func TestFixer_ApplyFixes_DuplicateLinksOnSameLine(t *testing.T) {
	vaultDir := setupTempVault(t, map[string]string{
		"note.md": "See [[A]] and [[A]] here.\n",
	})

	fixes := []Fix{
		{FilePath: "note.md", Line: 1, Column: 4, OriginalRaw: "[[A]]", OriginalTarget: "A", SuggestedTarget: "Alpha"},
		{FilePath: "note.md", Line: 1, Column: 14, OriginalRaw: "[[A]]", OriginalTarget: "A", SuggestedTarget: "Beta"},
	}

	prompter := &mockPrompter{actions: []FixAction{FixActionApplyAll}}
	fixer := NewFixer(vaultDir, prompter)

	summary := fixer.ApplyFixes(fixes)
	assert.Equal(t, 2, summary.Applied)
	assert.Empty(t, summary.Errors)

	content, err := os.ReadFile(filepath.Join(vaultDir, "note.md"))
	require.NoError(t, err)
	assert.Equal(t, "See [[Alpha]] and [[Beta]] here.\n", string(content))
}

func TestFixer_ApplyFixes_EmptyFixes(t *testing.T) {
	fixer := NewFixer(t.TempDir(), &mockPrompter{})
	summary := fixer.ApplyFixes(nil)
	assert.Equal(t, 0, summary.Applied)
	assert.Equal(t, 0, summary.Skipped)
	assert.Empty(t, summary.Errors)
}
