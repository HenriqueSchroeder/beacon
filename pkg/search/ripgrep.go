package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// rgJSON represents a single line of ripgrep JSON output.
type rgJSON struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// rgMatch represents a match line from ripgrep JSON output.
type rgMatch struct {
	Path       rgPath       `json:"path"`
	Lines      rgText       `json:"lines"`
	LineNumber int          `json:"line_number"`
	Submatches []rgSubmatch `json:"submatches"`
}

// rgContext represents a context line from ripgrep JSON output.
type rgContext struct {
	Path       rgPath `json:"path"`
	Lines      rgText `json:"lines"`
	LineNumber int    `json:"line_number"`
}

// rgPath represents a file path in ripgrep JSON output.
type rgPath struct {
	Text string `json:"text"`
}

// rgText represents text content in ripgrep JSON output.
type rgText struct {
	Text string `json:"text"`
}

// rgSubmatch represents a submatch in ripgrep JSON output.
type rgSubmatch struct {
	Match rgText `json:"match"`
}

// RipgrepSearcher implements Searcher using ripgrep for fast content search.
type RipgrepSearcher struct {
	vaultPath string
	ignore    []string
	rgPath    string
}

// NewRipgrepSearcher creates a new RipgrepSearcher after validating that
// ripgrep is available in PATH and the vault path exists and is a directory.
func NewRipgrepSearcher(vaultPath string, ignore []string) (*RipgrepSearcher, error) {
	rgPath, err := exec.LookPath("rg")
	if err != nil {
		return nil, fmt.Errorf("ripgrep not found in PATH: %w", err)
	}

	absVaultPath, err := filepath.Abs(vaultPath)
	if err != nil {
		return nil, fmt.Errorf("vault path error: %w", err)
	}

	info, err := os.Stat(absVaultPath)
	if err != nil {
		return nil, fmt.Errorf("vault path error: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("vault path is not a directory: %s", absVaultPath)
	}

	return &RipgrepSearcher{
		vaultPath: absVaultPath,
		ignore:    ignore,
		rgPath:    rgPath,
	}, nil
}

// SearchContent searches for the given query across all markdown files in the vault.
func (s *RipgrepSearcher) SearchContent(ctx context.Context, query string) ([]SearchResult, error) {
	return s.runSearch(ctx, query, false)
}

// SearchRelated searches for wiki-link backlinks to the resolved note target.
func (s *RipgrepSearcher) SearchRelated(ctx context.Context, target ResolvedTarget) ([]SearchResult, error) {
	if len(target.Aliases) == 0 {
		return nil, fmt.Errorf("search: missing target aliases")
	}

	patterns := make([]string, 0, len(target.Aliases))
	for _, alias := range dedupeStrings(target.Aliases) {
		patterns = append(patterns, regexp.QuoteMeta(filepath.ToSlash(alias)))
	}

	query := fmt.Sprintf(`\[\[(?:%s)(?:#[^|\]]*)?(?:\|[^\]]*)?\]\]`, strings.Join(patterns, "|"))
	return s.runSearch(ctx, query, true)
}

func (s *RipgrepSearcher) runSearch(ctx context.Context, query string, caseInsensitive bool) ([]SearchResult, error) {
	args := []string{"--json", "-C", "2", "--type", "md"}

	for _, pattern := range s.ignore {
		args = append(args, "--glob", fmt.Sprintf("!%s", pattern))
	}

	if caseInsensitive {
		args = append(args, "-i")
	}

	args = append(args, "-e", query, s.vaultPath)

	cmd := exec.CommandContext(ctx, s.rgPath, args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// Check if context was cancelled
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// ripgrep exits with code 1 when no matches are found
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil, nil
		}

		return nil, fmt.Errorf("ripgrep execution failed: %w", err)
	}

	return s.parseOutput(stdout.Bytes())
}

// parseOutput parses ripgrep's JSON-per-line output into SearchResult slices.
// Ripgrep JSON output consists of lines with types: begin, match, context, end.
// Context lines before a match go into ContextBefore, after into ContextAfter.
func (s *RipgrepSearcher) parseOutput(data []byte) ([]SearchResult, error) {
	if len(data) == 0 {
		return nil, nil
	}

	var results []SearchResult
	var contextBuffer []string
	var lastMatch *SearchResult

	flushMatch := func() {
		if lastMatch != nil {
			results = append(results, *lastMatch)
			lastMatch = nil
		}
	}

	lines := bytes.Split(data, []byte("\n"))
	for _, line := range lines {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		var entry rgJSON
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}

		switch entry.Type {
		case "begin":
			flushMatch()
			contextBuffer = nil

		case "match":
			flushMatch()

			var m rgMatch
			if err := json.Unmarshal(entry.Data, &m); err != nil {
				continue
			}

			relPath, err := filepath.Rel(s.vaultPath, m.Path.Text)
			if err != nil {
				relPath = m.Path.Text
			}

			sr := SearchResult{
				Path:          relPath,
				Line:          m.LineNumber,
				Match:         strings.TrimRight(m.Lines.Text, "\n"),
				ContextBefore: contextBuffer,
			}
			lastMatch = &sr
			contextBuffer = nil

		case "context":
			var c rgContext
			if err := json.Unmarshal(entry.Data, &c); err != nil {
				continue
			}

			contextLine := strings.TrimRight(c.Lines.Text, "\n")

			if lastMatch != nil && c.LineNumber > lastMatch.Line {
				// Context after the current match
				lastMatch.ContextAfter = append(lastMatch.ContextAfter, contextLine)
			} else {
				// Context before the next match
				contextBuffer = append(contextBuffer, contextLine)
			}

		case "end":
			flushMatch()
			contextBuffer = nil
		}
	}

	flushMatch()

	return results, nil
}
