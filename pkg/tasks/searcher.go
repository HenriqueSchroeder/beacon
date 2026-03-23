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

const pendingTaskPattern = `^[[:space:]]*-\s\[\s\]\s+.+`

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
	args := []string{"--json", "--type", "md"}
	for _, pattern := range s.ignore {
		args = append(args, "--glob", fmt.Sprintf("!%s", pattern))
	}
	args = append(args, "-e", pendingTaskPattern, s.vaultPath)

	cmd := exec.CommandContext(ctx, s.rgPath, args...)

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

	sort.Slice(results, func(i, j int) bool {
		if results[i].Path != results[j].Path {
			return results[i].Path < results[j].Path
		}
		return results[i].Line < results[j].Line
	})

	return results, nil
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

		relPath, err := filepath.Rel(s.vaultPath, match.Path.Text)
		if err != nil {
			relPath = match.Path.Text
		}

		results = append(results, Task{
			Path: relPath,
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
