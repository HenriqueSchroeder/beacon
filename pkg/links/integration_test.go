package links

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration tests demonstrating real-world usage patterns

func TestRealWorldLinkPatterns(t *testing.T) {
	p := NewParser()

	tests := []struct {
		description string
		content     string
		expectedLinks int
		checks      func(t *testing.T, links []Link)
	}{
		{
			description:   "README with multiple link types",
			expectedLinks: 4,
			content: `# My Project

See the [[setup]] guide for installation.

## Architecture

This uses [[design#patterns|design patterns]] and [[architecture]].

For more, see [[docs#advanced#section|Advanced Topics]].
`,
			checks: func(t *testing.T, links []Link) {
				// Should have parsed all link types correctly
				assert.Equal(t, "setup", links[0].Target)
				assert.Equal(t, "design", links[1].Target)
				assert.Equal(t, "patterns", links[1].Heading)
				assert.Equal(t, "design patterns", links[1].Alias)
			},
		},
		{
			description:   "Code documentation with inline links",
			expectedLinks: 3,
			content: `
// Function description
// See [[stdlib#arrays]] for more info
// Related: [[utilities|utility functions]]
func process(data []string) {
	// Implementation
	// Based on [[algorithm]]
}
`,
			checks: func(t *testing.T, links []Link) {
				assert.True(t, links[0].IsValid())
				assert.Equal(t, "stdlib", links[0].Target)
				assert.Equal(t, "arrays", links[0].Heading)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			links := p.Parse(tt.content, "test.md")
			assert.Equal(t, tt.expectedLinks, len(links))
			if tt.checks != nil {
				tt.checks(t, links)
			}
		})
	}
}

func TestEdgeCases(t *testing.T) {
	p := NewParser()

	tests := []struct {
		name      string
		content   string
		expectLen int
	}{
		{
			name:      "Empty link",
			content:   "This [[]] has an empty link",
			expectLen: 0, // regex requires at least one char inside [[]]
		},
		{
			name:      "Link with only heading",
			content:   "This [[#section]] has only heading",
			expectLen: 1,
		},
		{
			name:      "Multiple consecutive links",
			content:   "[[a]][[b]][[c]]",
			expectLen: 3,
		},
		{
			name:      "Link at line start",
			content:   "[[start]] in the beginning",
			expectLen: 1,
		},
		{
			name:      "Link at line end",
			content:   "At the end [[end]]",
			expectLen: 1,
		},
		{
			name:      "Only link on line",
			content:   "[[only]]",
			expectLen: 1,
		},
		{
			name:      "Link with newlines around it",
			content:   "\n\n[[surrounded]]\n\n",
			expectLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			links := p.Parse(tt.content, "test.md")
			assert.Equal(t, tt.expectLen, len(links), "expected %d links in %q", tt.expectLen, tt.content)
		})
	}
}

func TestLinkTypeClassification(t *testing.T) {
	p := NewParser()

	tests := []struct {
		raw      string
		expected LinkType
	}{
		{"[[note]]", LinkTypeNote},
		{"[[note#heading]]", LinkTypeHeading},
		{"[[note|alias]]", LinkTypeAliasNote},
		{"[[note#heading|alias]]", LinkTypeHeadingAlias},
	}

	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			links := p.Parse(tt.raw, "test.md")
			require.Len(t, links, 1)
			assert.Equal(t, tt.expected, links[0].Type)
		})
	}
}

func TestLinkDisplayText(t *testing.T) {
	tests := []struct {
		raw             string
		expectedDisplay string
	}{
		{"[[note]]", "note"},
		{"[[note#heading]]", "note#heading"},
		{"[[note|alias]]", "alias"},
		{"[[note#heading|Click here]]", "Click here"},
	}

	p := NewParser()

	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			links := p.Parse(tt.raw, "test.md")
			require.Len(t, links, 1)
			assert.Equal(t, tt.expectedDisplay, links[0].GetDisplayText())
		})
	}
}

func TestMultilineMarkdownDocument(t *testing.T) {
	content := `# My Notes

## Section 1

See [[note1]] for details.

### Subsection

More information in [[note2#introduction|the introduction]].

## Section 2

[[note3]], [[note4#heading]], and [[note5|alias]] are related.

End of document.
`

	p := NewParser()
	links := p.Parse(content, "document.md")

	assert.Equal(t, 5, len(links))

	// Verify line numbers
	assert.Equal(t, 5, links[0].Line)
	assert.Equal(t, 9, links[1].Line)
	assert.Equal(t, 13, links[2].Line)
	assert.Equal(t, 13, links[3].Line)
	assert.Equal(t, 13, links[4].Line)

	// Verify targets
	targets := make([]string, len(links))
	for i, l := range links {
		targets[i] = l.Target
	}
	assert.Equal(t, []string{"note1", "note2", "note3", "note4", "note5"}, targets)
}

func TestParserConsistency(t *testing.T) {
	content := "[[link1]] and [[link2|alias]] plus [[link3#heading]]"

	p1 := NewParser()
	p2 := NewParser()

	links1 := p1.Parse(content, "test.md")
	links2 := p2.Parse(content, "test.md")

	assert.Equal(t, len(links1), len(links2))

	for i := range links1 {
		assert.Equal(t, links1[i].Raw, links2[i].Raw)
		assert.Equal(t, links1[i].Target, links2[i].Target)
		assert.Equal(t, links1[i].Heading, links2[i].Heading)
		assert.Equal(t, links1[i].Alias, links2[i].Alias)
	}
}

func TestLinkValidation(t *testing.T) {
	tests := []struct {
		name  string
		link  Link
		valid bool
	}{
		{
			name:  "Valid simple link",
			link:  Link{Target: "note", Heading: "", Alias: ""},
			valid: true,
		},
		{
			name:  "Valid with heading and alias",
			link:  Link{Target: "note", Heading: "section", Alias: "Click"},
			valid: true,
		},
		{
			name:  "Invalid - no target",
			link:  Link{Target: "", Heading: "section", Alias: ""},
			valid: false,
		},
		{
			name:  "Invalid - target has brackets",
			link:  Link{Target: "note[1]", Heading: "", Alias: ""},
			valid: false,
		},
		{
			name:  "Invalid - heading has brackets",
			link:  Link{Target: "note", Heading: "[section]", Alias: ""},
			valid: false,
		},
		{
			name:  "Invalid - alias has brackets",
			link:  Link{Target: "note", Heading: "", Alias: "[text]"},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, tt.link.IsValid())
		})
	}
}
