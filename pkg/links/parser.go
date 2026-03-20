package links

import (
	"regexp"
	"strings"
)

// LinkType represents the type of wiki link
type LinkType int

const (
	// LinkTypeNote represents a simple note reference [[note]]
	LinkTypeNote LinkType = iota
	// LinkTypeHeading represents a note with heading reference [[note#heading]]
	LinkTypeHeading
	// LinkTypeAlias represents a note with alias reference [[note|alias]]
	LinkTypeAliasNote
	// LinkTypeHeadingAlias represents a note with heading and alias [[note#heading|alias]]
	LinkTypeHeadingAlias
)

// Link represents a parsed wiki link
type Link struct {
	Raw      string   // Raw text as it appears in the document (e.g., "[[note#heading|alias]]")
	Target   string   // The target note (e.g., "note")
	Heading  string   // The heading (e.g., "heading"), empty if not specified
	Alias    string   // The alias/display text (e.g., "alias"), empty if not specified
	Type     LinkType // Type of link
	Line     int      // Line number where the link appears (1-indexed)
	Column   int      // Column number where the link starts (0-indexed)
	FileName string   // Name of the file containing the link
}

// Parser parses wiki links from markdown content
type Parser struct {
	// Regex to match wiki links [[...]]
	linkRegex *regexp.Regexp
}

// NewParser creates a new link parser
func NewParser() *Parser {
	// This regex matches:
	// [[target#heading|alias]]
	// [[target#heading]]
	// [[target|alias]]
	// [[target]]
	linkRegex := regexp.MustCompile(`\[\[([^\[\]]+)\]\]`)
	return &Parser{
		linkRegex: linkRegex,
	}
}

// Parse extracts all wiki links from the given markdown content
func (p *Parser) Parse(content string, fileName string) []Link {
	var links []Link

	lines := strings.Split(content, "\n")
	for lineNum, line := range lines {
		matches := p.linkRegex.FindAllStringSubmatchIndex(line, -1)
		for _, match := range matches {
			// match[0] and match[1] are the indices of the full match ([[...]])
			// match[2] and match[3] are the indices of the captured group (content inside)
			fullStart := match[0]
			fullEnd := match[1]
			contentStart := match[2]
			contentEnd := match[3]

			raw := line[fullStart:fullEnd]
			content := line[contentStart:contentEnd]

			link := p.parseLink(content, raw, fileName, lineNum+1, fullStart)
			links = append(links, link)
		}
	}

	return links
}

// parseLink parses a single link content and returns a Link struct
func (p *Parser) parseLink(content, raw, fileName string, lineNum, column int) Link {
	link := Link{
		Raw:      raw,
		FileName: fileName,
		Line:     lineNum,
		Column:   column,
	}

	// Check for alias (|)
	var target string
	if pipeIdx := strings.LastIndex(content, "|"); pipeIdx != -1 {
		target = strings.TrimSpace(content[:pipeIdx])
		link.Alias = strings.TrimSpace(content[pipeIdx+1:])
	} else {
		target = content
	}

	// Check for heading (#)
	if hashIdx := strings.LastIndex(target, "#"); hashIdx != -1 {
		link.Target = strings.TrimSpace(target[:hashIdx])
		link.Heading = strings.TrimSpace(target[hashIdx+1:])

		if link.Alias != "" {
			link.Type = LinkTypeHeadingAlias
		} else {
			link.Type = LinkTypeHeading
		}
	} else {
		link.Target = strings.TrimSpace(target)

		if link.Alias != "" {
			link.Type = LinkTypeAliasNote
		} else {
			link.Type = LinkTypeNote
		}
	}

	return link
}

// IsValid checks if a link appears to be properly formatted
func (l *Link) IsValid() bool {
	if l.Target == "" {
		return false
	}

	// Target shouldn't contain invalid characters
	if strings.ContainsAny(l.Target, "[]|#") {
		return false
	}

	// Heading shouldn't contain invalid characters (except |, which we handle)
	if strings.ContainsAny(l.Heading, "[]") {
		return false
	}

	// Alias shouldn't contain invalid characters
	if strings.ContainsAny(l.Alias, "[]") {
		return false
	}

	return true
}

// GetDisplayText returns what the link should display as
func (l *Link) GetDisplayText() string {
	if l.Alias != "" {
		return l.Alias
	}
	if l.Heading != "" {
		return l.Target + "#" + l.Heading
	}
	return l.Target
}

// String returns a human-readable representation of the link
func (l *Link) String() string {
	return l.Raw
}
