package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FixAction represents the user's decision for a fix prompt
type FixAction int

const (
	FixActionApply FixAction = iota
	FixActionSkip
	FixActionApplyAll
	FixActionQuit
)

// Fix represents a fixable broken link with its suggested correction
type Fix struct {
	FilePath        string
	Line            int
	Column          int
	OriginalRaw     string // e.g. "[[Vini]]"
	OriginalTarget  string // e.g. "Vini"
	SuggestedTarget string // e.g. "Vinicius Dal Magro"
	Heading         string // preserved if present
	Alias           string // preserved if present
	Reason          string
}

// CorrectedRaw returns the corrected wiki-link string
func (f Fix) CorrectedRaw() string {
	var b strings.Builder
	b.WriteString("[[")
	b.WriteString(f.SuggestedTarget)
	if f.Heading != "" {
		b.WriteString("#")
		b.WriteString(f.Heading)
	}
	if f.Alias != "" {
		b.WriteString("|")
		b.WriteString(f.Alias)
	}
	b.WriteString("]]")
	return b.String()
}

// FixSummary contains the results of a fix operation
type FixSummary struct {
	Applied int
	Skipped int
	Errors  []error
}

// Prompter defines the interface for user interaction during fix
type Prompter interface {
	// Prompt presents a fix to the user and returns the chosen action
	Prompt(fix Fix, current, total int) FixAction
}

// Fixer applies suggested corrections to broken links in vault files
type Fixer struct {
	vaultPath string
	prompter  Prompter
}

// NewFixer creates a new Fixer
func NewFixer(vaultPath string, prompter Prompter) *Fixer {
	return &Fixer{
		vaultPath: vaultPath,
		prompter:  prompter,
	}
}

// CollectFixes extracts fixable items from validation results
func CollectFixes(results []DocumentValidation) []Fix {
	var fixes []Fix
	for _, doc := range results {
		for _, vr := range doc.Results {
			if !vr.Valid && vr.SuggestedTarget != "" {
				fixes = append(fixes, Fix{
					FilePath:        doc.FilePath,
					Line:            vr.Link.Line,
					Column:          vr.Link.Column,
					OriginalRaw:     vr.Link.Raw,
					OriginalTarget:  vr.Link.Target,
					SuggestedTarget: vr.SuggestedTarget,
					Heading:         vr.Link.Heading,
					Alias:           vr.Link.Alias,
					Reason:          vr.Reason,
				})
			}
		}
	}
	return fixes
}

// ApplyFixes prompts the user for each fix and applies accepted ones
func (f *Fixer) ApplyFixes(fixes []Fix) FixSummary {
	summary := FixSummary{}
	applyAll := false

	for i, fix := range fixes {
		action := FixActionApply
		if !applyAll {
			action = f.prompter.Prompt(fix, i+1, len(fixes))
		}

		switch action {
		case FixActionApply:
			if err := f.applyFix(fix); err != nil {
				summary.Errors = append(summary.Errors, fmt.Errorf("%s:%d: %w", fix.FilePath, fix.Line, err))
			} else {
				summary.Applied++
			}
		case FixActionSkip:
			summary.Skipped++
		case FixActionApplyAll:
			applyAll = true
			if err := f.applyFix(fix); err != nil {
				summary.Errors = append(summary.Errors, fmt.Errorf("%s:%d: %w", fix.FilePath, fix.Line, err))
			} else {
				summary.Applied++
			}
		case FixActionQuit:
			summary.Skipped += len(fixes) - i
			return summary
		}
	}

	return summary
}

// applyFix replaces a single broken link in its source file
func (f *Fixer) applyFix(fix Fix) error {
	absPath := filepath.Join(f.vaultPath, fix.FilePath)

	data, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	lineIdx := fix.Line - 1 // Line is 1-indexed
	if lineIdx < 0 || lineIdx >= len(lines) {
		return fmt.Errorf("line %d out of range (file has %d lines)", fix.Line, len(lines))
	}

	line := lines[lineIdx]
	corrected := fix.CorrectedRaw()

	// Use Column for precise positioning; fall back to string search
	idx := -1
	if fix.Column >= 0 && fix.Column+len(fix.OriginalRaw) <= len(line) &&
		line[fix.Column:fix.Column+len(fix.OriginalRaw)] == fix.OriginalRaw {
		idx = fix.Column
	} else {
		idx = strings.Index(line, fix.OriginalRaw)
	}
	if idx == -1 {
		return fmt.Errorf("link %s not found on line %d", fix.OriginalRaw, fix.Line)
	}

	lines[lineIdx] = line[:idx] + corrected + line[idx+len(fix.OriginalRaw):]

	// Atomic write: temp file + rename to avoid corruption on interruption
	dir := filepath.Dir(absPath)
	tmp, err := os.CreateTemp(dir, ".beacon-fix-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.WriteString(strings.Join(lines, "\n")); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	tmp.Close()

	if err := os.Rename(tmpName, absPath); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}
