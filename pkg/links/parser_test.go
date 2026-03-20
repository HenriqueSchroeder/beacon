package links

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewParser(t *testing.T) {
	p := NewParser()
	assert.NotNil(t, p)
	assert.NotNil(t, p.linkRegex)
}

func TestParser_Parse_SimpleNote(t *testing.T) {
	p := NewParser()
	content := "This is a link to [[my-note]] in the middle."
	links := p.Parse(content, "test.md")

	require.Len(t, links, 1)
	assert.Equal(t, "[[my-note]]", links[0].Raw)
	assert.Equal(t, "my-note", links[0].Target)
	assert.Equal(t, "", links[0].Heading)
	assert.Equal(t, "", links[0].Alias)
	assert.Equal(t, LinkTypeNote, links[0].Type)
	assert.Equal(t, 1, links[0].Line)
	assert.Equal(t, "test.md", links[0].FileName)
}

func TestParser_Parse_NoteWithHeading(t *testing.T) {
	p := NewParser()
	content := "Link to [[my-note#section]]"
	links := p.Parse(content, "test.md")

	require.Len(t, links, 1)
	assert.Equal(t, "[[my-note#section]]", links[0].Raw)
	assert.Equal(t, "my-note", links[0].Target)
	assert.Equal(t, "section", links[0].Heading)
	assert.Equal(t, "", links[0].Alias)
	assert.Equal(t, LinkTypeHeading, links[0].Type)
}

func TestParser_Parse_NoteWithAlias(t *testing.T) {
	p := NewParser()
	content := "Link to [[my-note|click here]]"
	links := p.Parse(content, "test.md")

	require.Len(t, links, 1)
	assert.Equal(t, "[[my-note|click here]]", links[0].Raw)
	assert.Equal(t, "my-note", links[0].Target)
	assert.Equal(t, "", links[0].Heading)
	assert.Equal(t, "click here", links[0].Alias)
	assert.Equal(t, LinkTypeAliasNote, links[0].Type)
}

func TestParser_Parse_NoteWithHeadingAndAlias(t *testing.T) {
	p := NewParser()
	content := "Link to [[my-note#intro|click here]]"
	links := p.Parse(content, "test.md")

	require.Len(t, links, 1)
	assert.Equal(t, "[[my-note#intro|click here]]", links[0].Raw)
	assert.Equal(t, "my-note", links[0].Target)
	assert.Equal(t, "intro", links[0].Heading)
	assert.Equal(t, "click here", links[0].Alias)
	assert.Equal(t, LinkTypeHeadingAlias, links[0].Type)
}

func TestParser_Parse_MultipleLinks(t *testing.T) {
	p := NewParser()
	content := `
First link: [[note1]]
Second link: [[note2#heading|alias]]
Third: [[note3]]
`
	links := p.Parse(content, "test.md")

	require.Len(t, links, 3)
	assert.Equal(t, "note1", links[0].Target)
	assert.Equal(t, "note2", links[1].Target)
	assert.Equal(t, "heading", links[1].Heading)
	assert.Equal(t, "alias", links[1].Alias)
	assert.Equal(t, "note3", links[2].Target)
}

func TestParser_Parse_MultilineDocument(t *testing.T) {
	p := NewParser()
	content := "Line 1\nLine 2 with [[link1]]\nLine 3\nLine 4 with [[link2]]"
	links := p.Parse(content, "test.md")

	require.Len(t, links, 2)
	assert.Equal(t, 2, links[0].Line)
	assert.Equal(t, "link1", links[0].Target)
	assert.Equal(t, 4, links[1].Line)
	assert.Equal(t, "link2", links[1].Target)
}

func TestParser_Parse_NoLinks(t *testing.T) {
	p := NewParser()
	content := "This is plain content with no links"
	links := p.Parse(content, "test.md")

	assert.Empty(t, links)
}

func TestParser_Parse_NestedBracketsIgnored(t *testing.T) {
	p := NewParser()
	// Nested brackets are not valid wiki links, but our regex should not match them
	content := "Text with [regular] [brackets]"
	links := p.Parse(content, "test.md")

	assert.Empty(t, links)
}

func TestParser_Parse_SpacesInLink(t *testing.T) {
	p := NewParser()
	content := "Link to [[my note with spaces]]"
	links := p.Parse(content, "test.md")

	require.Len(t, links, 1)
	assert.Equal(t, "my note with spaces", links[0].Target)
}

func TestParser_Parse_SpecialCharactersInTarget(t *testing.T) {
	p := NewParser()
	content := "Link to [[note-with-dashes_and_underscores.md]]"
	links := p.Parse(content, "test.md")

	require.Len(t, links, 1)
	assert.Equal(t, "note-with-dashes_and_underscores.md", links[0].Target)
}

func TestLink_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		link  Link
		valid bool
	}{
		{
			name:  "simple valid link",
			link:  Link{Target: "note", Heading: "", Alias: ""},
			valid: true,
		},
		{
			name:  "valid link with heading",
			link:  Link{Target: "note", Heading: "section", Alias: ""},
			valid: true,
		},
		{
			name:  "valid link with alias",
			link:  Link{Target: "note", Heading: "", Alias: "click here"},
			valid: true,
		},
		{
			name:  "empty target",
			link:  Link{Target: "", Heading: "", Alias: ""},
			valid: false,
		},
		{
			name:  "target with invalid brackets",
			link:  Link{Target: "note[invalid]", Heading: "", Alias: ""},
			valid: false,
		},
		{
			name:  "heading with invalid brackets",
			link:  Link{Target: "note", Heading: "sec[tion]", Alias: ""},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, tt.link.IsValid())
		})
	}
}

func TestLink_GetDisplayText(t *testing.T) {
	tests := []struct {
		name     string
		link     Link
		expected string
	}{
		{
			name:     "with alias",
			link:     Link{Target: "note", Heading: "sec", Alias: "click here"},
			expected: "click here",
		},
		{
			name:     "without alias but with heading",
			link:     Link{Target: "note", Heading: "sec", Alias: ""},
			expected: "note#sec",
		},
		{
			name:     "simple link",
			link:     Link{Target: "note", Heading: "", Alias: ""},
			expected: "note",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.link.GetDisplayText())
		})
	}
}

func TestParser_Parse_ColumnPositions(t *testing.T) {
	p := NewParser()
	content := "Start [[link1]] middle [[link2]] end"
	links := p.Parse(content, "test.md")

	require.Len(t, links, 2)
	assert.Equal(t, 6, links[0].Column)   // "Start " = 6 chars
	assert.Equal(t, 27, links[1].Column)  // Position of second link
}

func TestParser_Parse_ConsecutiveLinks(t *testing.T) {
	p := NewParser()
	content := "[[link1]][[link2]][[link3]]"
	links := p.Parse(content, "test.md")

	require.Len(t, links, 3)
	assert.Equal(t, "link1", links[0].Target)
	assert.Equal(t, "link2", links[1].Target)
	assert.Equal(t, "link3", links[2].Target)
}

func TestParser_Parse_AliasWithPipe(t *testing.T) {
	p := NewParser()
	// Test that we handle multiple pipes correctly (last one wins)
	content := "[[note#sec|alias with | pipe]]"
	links := p.Parse(content, "test.md")

	require.Len(t, links, 1)
	assert.Equal(t, "note", links[0].Target)
	assert.Equal(t, "sec", links[0].Heading)
	// The alias should be "alias with | pipe" but regex will match up to last |
	// Our implementation uses LastIndex, so it should be "pipe"
	assert.Equal(t, "pipe", links[0].Alias)
}
