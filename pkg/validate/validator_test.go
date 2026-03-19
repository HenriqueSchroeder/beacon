package validate

import (
	"context"
	"testing"

	"github.com/HenriqueSchroeder/beacon/pkg/links"
	"github.com/HenriqueSchroeder/beacon/pkg/vault"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockVault implements vault.Vault for testing
type MockVault struct {
	notes map[string]*vault.Note
}

func (m *MockVault) ListNotes(ctx context.Context) ([]vault.Note, error) {
	var notes []vault.Note
	for _, note := range m.notes {
		notes = append(notes, *note)
	}
	return notes, nil
}

func (m *MockVault) GetNote(ctx context.Context, path string) (*vault.Note, error) {
	return m.notes[path], nil
}

func newMockVault() *MockVault {
	return &MockVault{
		notes: make(map[string]*vault.Note),
	}
}

func addNote(vault *MockVault, path, content string) {
	vault.notes[path] = &vault.Note{
		Path:    path,
		Content: content,
	}
}

func TestNewValidator(t *testing.T) {
	v := NewValidator(newMockVault(), 0)
	assert.NotNil(t, v)
	assert.Equal(t, 4, v.maxWorkers) // Default
}

func TestValidator_BuildIndex(t *testing.T) {
	mockVault := newMockVault()
	addNote(mockVault, "note1.md", "# Heading 1\n## Heading 2\nContent")
	addNote(mockVault, "note2.md", "# Another\nContent")

	v := NewValidator(mockVault, 1)
	err := v.BuildIndex(context.Background())

	require.NoError(t, err)
	assert.True(t, v.noteIndex["note1.md"])
	assert.True(t, v.noteIndex["note1"])
	assert.True(t, v.noteIndex["note2.md"])
	assert.True(t, v.noteIndex["note2"])
}

func TestValidator_BuildIndex_ExtractsHeadings(t *testing.T) {
	mockVault := newMockVault()
	addNote(mockVault, "note1.md", "# Heading 1\n## Heading 2\nContent")

	v := NewValidator(mockVault, 1)
	err := v.BuildIndex(context.Background())

	require.NoError(t, err)
	headings, ok := v.headingIndex["note1.md"]
	assert.True(t, ok)
	assert.Contains(t, headings, "heading 1")
	assert.Contains(t, headings, "heading 2")
}

func TestValidator_ValidateDocument_ValidLink(t *testing.T) {
	mockVault := newMockVault()
	addNote(mockVault, "target.md", "# Introduction\nContent")
	addNote(mockVault, "source.md", "See [[target]] for details.")

	v := NewValidator(mockVault, 1)
	err := v.BuildIndex(context.Background())
	require.NoError(t, err)

	sourceNote := mockVault.notes["source.md"]
	result := v.ValidateDocument(context.Background(), sourceNote)

	assert.Equal(t, 1, result.TotalLinks)
	assert.Equal(t, 1, result.ValidLinks)
	assert.True(t, result.Results[0].Valid)
}

func TestValidator_ValidateDocument_InvalidTarget(t *testing.T) {
	mockVault := newMockVault()
	addNote(mockVault, "source.md", "See [[nonexistent]] for details.")

	v := NewValidator(mockVault, 1)
	err := v.BuildIndex(context.Background())
	require.NoError(t, err)

	sourceNote := mockVault.notes["source.md"]
	result := v.ValidateDocument(context.Background(), sourceNote)

	assert.Equal(t, 1, result.TotalLinks)
	assert.Equal(t, 0, result.ValidLinks)
	assert.False(t, result.Results[0].Valid)
	assert.Contains(t, result.Results[0].Reason, "target note not found")
}

func TestValidator_ValidateDocument_InvalidHeading(t *testing.T) {
	mockVault := newMockVault()
	addNote(mockVault, "target.md", "# Introduction\nContent")
	addNote(mockVault, "source.md", "See [[target#nonexistent]] for details.")

	v := NewValidator(mockVault, 1)
	err := v.BuildIndex(context.Background())
	require.NoError(t, err)

	sourceNote := mockVault.notes["source.md"]
	result := v.ValidateDocument(context.Background(), sourceNote)

	assert.Equal(t, 1, result.TotalLinks)
	assert.Equal(t, 0, result.ValidLinks)
	assert.False(t, result.Results[0].Valid)
	assert.Contains(t, result.Results[0].Reason, "heading not found")
}

func TestValidator_ValidateDocument_ValidHeading(t *testing.T) {
	mockVault := newMockVault()
	addNote(mockVault, "target.md", "# Introduction\n## Installation\nContent")
	addNote(mockVault, "source.md", "See [[target#installation]] for details.")

	v := NewValidator(mockVault, 1)
	err := v.BuildIndex(context.Background())
	require.NoError(t, err)

	sourceNote := mockVault.notes["source.md"]
	result := v.ValidateDocument(context.Background(), sourceNote)

	assert.Equal(t, 1, result.TotalLinks)
	assert.Equal(t, 1, result.ValidLinks)
	assert.True(t, result.Results[0].Valid)
}

func TestValidator_ValidateDocument_MultipleLinks(t *testing.T) {
	mockVault := newMockVault()
	addNote(mockVault, "note1.md", "Content")
	addNote(mockVault, "note2.md", "Content")
	addNote(mockVault, "source.md", "See [[note1]], [[note2]], and [[invalid]].")

	v := NewValidator(mockVault, 1)
	err := v.BuildIndex(context.Background())
	require.NoError(t, err)

	sourceNote := mockVault.notes["source.md"]
	result := v.ValidateDocument(context.Background(), sourceNote)

	assert.Equal(t, 3, result.TotalLinks)
	assert.Equal(t, 2, result.ValidLinks)
}

func TestValidator_Cache(t *testing.T) {
	mockVault := newMockVault()
	addNote(mockVault, "source.md", "See [[note1]] and [[note2]].")

	v := NewValidator(mockVault, 1)
	err := v.BuildIndex(context.Background())
	require.NoError(t, err)

	sourceNote := mockVault.notes["source.md"]
	result1 := v.ValidateDocument(context.Background(), sourceNote)
	result2 := v.ValidateDocument(context.Background(), sourceNote)

	// Should return the same cached result (pointer equality)
	assert.Equal(t, result1, result2)

	v.ClearCache()
	result3 := v.ValidateDocument(context.Background(), sourceNote)
	assert.NotEqual(t, result1, result3)
}

func TestValidator_FuzzySuggestion(t *testing.T) {
	mockVault := newMockVault()
	addNote(mockVault, "example.md", "Content")
	addNote(mockVault, "source.md", "See [[exmple]].")

	v := NewValidator(mockVault, 1)
	err := v.BuildIndex(context.Background())
	require.NoError(t, err)

	sourceNote := mockVault.notes["source.md"]
	result := v.ValidateDocument(context.Background(), sourceNote)

	assert.False(t, result.Results[0].Valid)
	assert.Contains(t, result.Results[0].Suggestion, "example")
}

func TestValidator_ValidateAll(t *testing.T) {
	mockVault := newMockVault()
	addNote(mockVault, "note1.md", "Content")
	addNote(mockVault, "note2.md", "Content")
	addNote(mockVault, "source1.md", "See [[note1]] and [[invalid1]].")
	addNote(mockVault, "source2.md", "See [[note2]] and [[invalid2]].")

	v := NewValidator(mockVault, 2)
	err := v.BuildIndex(context.Background())
	require.NoError(t, err)

	results, err := v.ValidateAll(context.Background())
	require.NoError(t, err)

	// We have 4 notes total
	assert.Equal(t, 4, len(results))

	// Count total validations
	totalLinks := 0
	validLinks := 0
	for _, r := range results {
		totalLinks += r.TotalLinks
		validLinks += r.ValidLinks
	}

	assert.Equal(t, 4, totalLinks)
	assert.Equal(t, 2, validLinks)
}

// Test helper functions

func TestExtractHeadings(t *testing.T) {
	content := `# Main Title
This is content.

## Section 1
More content.

### Subsection
Even more content.

# Another Title
Final content.`

	headings := extractHeadings(content)
	assert.Len(t, headings, 4)
	assert.Contains(t, headings, "main title")
	assert.Contains(t, headings, "section 1")
	assert.Contains(t, headings, "subsection")
	assert.Contains(t, headings, "another title")
}

func TestNormalizeTarget(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"note", "note"},
		{"Note", "note"},
		{"note.md", "note"},
		{"Note.md", "note"},
		{"  note  ", "note"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, normalizeTarget(tt.input))
	}
}

func TestNormalizeHeading(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Introduction", "introduction"},
		{"  Heading Text  ", "heading text"},
		{"My Heading", "my heading"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, normalizeHeading(tt.input))
	}
}

func TestStringInSlice(t *testing.T) {
	slice := []string{"apple", "banana", "cherry"}

	assert.True(t, stringInSlice("apple", slice))
	assert.True(t, stringInSlice("APPLE", slice)) // Case insensitive
	assert.False(t, stringInSlice("grape", slice))
	assert.False(t, stringInSlice("", slice))
}

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		a        string
		b        string
		distance int
	}{
		{"", "", 0},
		{"a", "", 1},
		{"", "b", 1},
		{"abc", "abc", 0},
		{"abc", "bbc", 1},
		{"abc", "def", 3},
		{"kitten", "sitting", 3},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.distance, levenshteinDistance(tt.a, tt.b))
	}
}

func TestLevenshteinSimilarity(t *testing.T) {
	// Same strings should have similarity of 1.0
	assert.Equal(t, 1.0, levenshteinSimilarity("abc", "abc"))

	// Completely different strings should have low similarity
	sim := levenshteinSimilarity("abc", "def")
	assert.Less(t, sim, 0.5)

	// Similar strings should have high similarity
	sim = levenshteinSimilarity("example", "exmple")
	assert.Greater(t, sim, 0.8)
}
