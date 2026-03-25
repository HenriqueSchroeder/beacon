package tasks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

const pendingTaskPattern = `^[[:space:]]*(?:[-+*]|\d+\.)\s\[\s\](?:\s+.+)?$`

type rgJSON struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type rgMatch struct {
	Path       rgPath `json:"path"`
	Lines      rgText `json:"lines"`
	LineNumber int    `json:"line_number"`
}

type rgPath struct {
	Text string `json:"text"`
}

type rgText struct {
	Text string `json:"text"`
}

// Task represents a single pending markdown checkbox.
type Task struct {
	Path string
	Line int
	Text string
}

// Searcher lists pending tasks from a vault using ripgrep.
type Searcher struct {
	vaultPath string
	ignore    []string
	rgPath    string
}

// NewSearcher validates ripgrep availability and the target vault path.
func NewSearcher(vaultPath string, ignore []string) (*Searcher, error) {
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

	return &Searcher{
		vaultPath: absVaultPath,
		ignore:    ignore,
		rgPath:    rgPath,
	}, nil
}

// ListPending returns pending markdown checkbox items across the vault.
func (s *Searcher) ListPending(ctx context.Context) ([]Task, error) {
	args := []string{"--json", "--type", "md", "--hidden", "--no-ignore"}
	for _, pattern := range s.ignore {
		if hasGlobMeta(pattern) {
			continue
		}
		for _, glob := range ignoreGlobs(pattern) {
			args = append(args, "--glob", fmt.Sprintf("!%s", glob))
		}
	}
	args = append(args, "-e", pendingTaskPattern, ".")

	cmd := exec.CommandContext(ctx, s.rgPath, args...)
	cmd.Dir = s.vaultPath

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil, nil
		}
		return nil, fmt.Errorf("ripgrep execution failed: %w", err)
	}

	results, err := s.parseOutput(stdout.Bytes())
	if err != nil {
		return nil, err
	}
	results = s.filterIgnored(results)

	sort.Slice(results, func(i, j int) bool {
		if results[i].Path != results[j].Path {
			return results[i].Path < results[j].Path
		}
		return results[i].Line < results[j].Line
	})

	return results, nil
}

func ignoreGlobs(pattern string) []string {
	pattern = filepath.ToSlash(pattern)

	globs := []string{pattern}
	if strings.Contains(pattern, "/") {
		if hasGlobMeta(pattern) {
			return dedupeGlobs(globs)
		}
		globs = append(globs, pattern+"/**")
		return dedupeGlobs(globs)
	}

	globs = append(globs, "**/"+pattern, pattern+"/**", "**/"+pattern+"/**")
	return dedupeGlobs(globs)
}

func dedupeGlobs(globs []string) []string {
	seen := make(map[string]struct{}, len(globs))
	result := make([]string, 0, len(globs))
	for _, glob := range globs {
		if _, ok := seen[glob]; ok {
			continue
		}
		seen[glob] = struct{}{}
		result = append(result, glob)
	}
	return result
}

func hasGlobMeta(pattern string) bool {
	return strings.ContainsAny(pattern, "*?[")
}

func (s *Searcher) filterIgnored(results []Task) []Task {
	filtered := make([]Task, 0, len(results))
	for _, task := range results {
		if shouldIgnore(task.Path, s.ignore) {
			continue
		}
		filtered = append(filtered, task)
	}
	return filtered
}

func shouldIgnore(relPath string, ignore []string) bool {
	nativeRelPath := filepath.FromSlash(relPath)

	for _, pattern := range ignore {
		nativePattern := filepath.FromSlash(pattern)

		if strings.HasPrefix(nativeRelPath, nativePattern+string(filepath.Separator)) || nativeRelPath == nativePattern {
			return true
		}

		if matched, _ := filepath.Match(nativePattern, nativeRelPath); matched {
			return true
		}

		parts := strings.Split(nativeRelPath, string(filepath.Separator))
		for _, part := range parts {
			if matched, _ := filepath.Match(nativePattern, part); matched {
				return true
			}
		}
	}
	return false
}

func normalizeRipgrepPath(path string) string {
	path = filepath.ToSlash(path)
	path = strings.TrimPrefix(path, "./")
	return path
}

func (s *Searcher) parseOutput(data []byte) ([]Task, error) {
	if len(data) == 0 {
		return nil, nil
	}

	var results []Task
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
		if entry.Type != "match" {
			continue
		}

		var match rgMatch
		if err := json.Unmarshal(entry.Data, &match); err != nil {
			continue
		}

		results = append(results, Task{
			Path: normalizeRipgrepPath(match.Path.Text),
			Line: match.LineNumber,
			Text: extractTaskText(match.Lines.Text),
		})
	}

	return results, nil
}

func extractTaskText(line string) string {
	line = strings.TrimRight(line, "\r\n")

	index := strings.Index(line, "[ ]")
	if index == -1 {
		return strings.TrimSpace(line)
	}

	return strings.TrimSpace(line[index+3:])
}
