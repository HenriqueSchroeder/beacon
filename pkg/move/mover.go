package move

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/HenriqueSchroeder/beacon/pkg/links"
	"github.com/HenriqueSchroeder/beacon/pkg/vault"
)

// LinkReplacement represents a single link text substitution within a file
type LinkReplacement struct {
	OldRaw string // e.g. "[[Old Name#Heading|Display]]"
	NewRaw string // e.g. "[[New Name#Heading|Display]]"
}

// FileUpdate represents all link replacements needed in a single file
type FileUpdate struct {
	Path         string            // relative path within vault
	Replacements []LinkReplacement // link substitutions to apply
}

// MoveResult holds the computed plan for a move operation
type MoveResult struct {
	Source      string       // original relative path
	Dest        string       // destination relative path
	OldStem     string       // filename stem before move
	NewStem     string       // filename stem after move
	NeedsRelink bool         // true when stem changed
	Updates     []FileUpdate // files that need link updates
}

// MoveSummary holds the outcome of applying a move
type MoveSummary struct {
	FilesMoved   int
	LinksUpdated int
	FilesUpdated int
	Errors       []error
}

// Mover orchestrates note rename/move with backlink updates
type Mover struct {
	vaultPath string
	vault     vault.Vault
	parser    *links.Parser
}

// NewMover creates a Mover
func NewMover(vaultPath string, v vault.Vault) *Mover {
	return &Mover{
		vaultPath: vaultPath,
		vault:     v,
		parser:    links.NewParser(),
	}
}

// Plan computes which files and links need updating without applying changes.
// source and dest are relative paths within the vault (e.g. "notes/My Note.md").
func (m *Mover) Plan(ctx context.Context, source, dest string) (*MoveResult, error) {
	// Prevent path traversal outside the vault
	if !m.isInsideVault(source) || !m.isInsideVault(dest) {
		return nil, fmt.Errorf("move: path must be within the vault")
	}

	// Validate source exists
	absSource := filepath.Join(m.vaultPath, source)
	if _, err := os.Stat(absSource); err != nil {
		return nil, fmt.Errorf("move: source not found: %s", source)
	}

	// Validate dest does not exist
	absDest := filepath.Join(m.vaultPath, dest)
	if _, err := os.Stat(absDest); err == nil {
		return nil, fmt.Errorf("move: destination already exists: %s", dest)
	}

	oldStem := stemFromPath(source)
	newStem := stemFromPath(dest)
	needsRelink := !strings.EqualFold(oldStem, newStem)

	result := &MoveResult{
		Source:      source,
		Dest:        dest,
		OldStem:     oldStem,
		NewStem:     newStem,
		NeedsRelink: needsRelink,
	}

	if !needsRelink {
		return result, nil
	}

	// Scan vault for notes containing links to old stem
	notes, err := m.vault.ListNotes(ctx)
	if err != nil {
		return nil, fmt.Errorf("move: failed to list notes: %w", err)
	}

	for _, note := range notes {
		fileLinks := m.parser.Parse(note.Content, note.Path)
		var replacements []LinkReplacement

		for _, link := range fileLinks {
			if !strings.EqualFold(link.Target, oldStem) {
				continue
			}

			newRaw := buildNewRaw(newStem, link.Heading, link.Alias)
			replacements = append(replacements, LinkReplacement{
				OldRaw: link.Raw,
				NewRaw: newRaw,
			})
		}

		if len(replacements) > 0 {
			result.Updates = append(result.Updates, FileUpdate{
				Path:         note.Path,
				Replacements: replacements,
			})
		}
	}

	return result, nil
}

// Apply executes the move: renames the file first, then updates all backlinks.
// Rename-first is safer: if it fails, zero changes have been made to the vault.
// If link updates fail after rename, the file is at the correct destination and
// broken links can be fixed with "beacon validate --fix".
func (m *Mover) Apply(result *MoveResult) (*MoveSummary, error) {
	summary := &MoveSummary{}

	// Step 1: create destination directory if needed
	absDest := filepath.Join(m.vaultPath, result.Dest)
	destDir := filepath.Dir(absDest)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return summary, fmt.Errorf("move: failed to create destination directory: %w", err)
	}

	// Step 2: rename/move the file (safest to do first — if this fails, nothing changed)
	absSource := filepath.Join(m.vaultPath, result.Source)
	if err := os.Rename(absSource, absDest); err != nil {
		return summary, fmt.Errorf("move: failed to rename file: %w", err)
	}
	summary.FilesMoved = 1

	// Step 3: update backlinks in referencing files
	for _, update := range result.Updates {
		n, err := m.applyFileUpdate(update)
		if err != nil {
			summary.Errors = append(summary.Errors, fmt.Errorf("%s: %w", update.Path, err))
			continue
		}
		summary.LinksUpdated += n
		summary.FilesUpdated++
	}

	return summary, nil
}

// isInsideVault checks that a relative path does not escape the vault root
func (m *Mover) isInsideVault(rel string) bool {
	abs := filepath.Clean(filepath.Join(m.vaultPath, rel))
	root := filepath.Clean(m.vaultPath) + string(filepath.Separator)
	return strings.HasPrefix(abs+string(filepath.Separator), root)
}

// applyFileUpdate rewrites a single file, replacing all matched link texts.
// Returns the number of replacements made.
func (m *Mover) applyFileUpdate(update FileUpdate) (int, error) {
	absPath := filepath.Join(m.vaultPath, update.Path)

	data, err := os.ReadFile(absPath)
	if err != nil {
		return 0, fmt.Errorf("failed to read file: %w", err)
	}

	content := string(data)
	count := 0

	// Deduplicate replacements: same OldRaw may appear multiple times in Plan output
	// but ReplaceAll handles all occurrences in one pass
	seen := make(map[string]bool)
	for _, rep := range update.Replacements {
		if seen[rep.OldRaw] {
			continue
		}
		seen[rep.OldRaw] = true
		n := strings.Count(content, rep.OldRaw)
		if n > 0 {
			content = strings.ReplaceAll(content, rep.OldRaw, rep.NewRaw)
			count += n
		}
	}

	if count == 0 {
		return 0, nil
	}

	// Atomic write: temp file + rename
	dir := filepath.Dir(absPath)
	tmp, err := os.CreateTemp(dir, ".beacon-move-*")
	if err != nil {
		return 0, fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.WriteString(content); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return 0, fmt.Errorf("failed to write temp file: %w", err)
	}
	tmp.Close()

	if err := os.Rename(tmpName, absPath); err != nil {
		os.Remove(tmpName)
		return 0, fmt.Errorf("failed to rename temp file: %w", err)
	}

	return count, nil
}

// buildNewRaw constructs a wiki-link string with the new target, preserving heading and alias
func buildNewRaw(target, heading, alias string) string {
	var b strings.Builder
	b.WriteString("[[")
	b.WriteString(target)
	if heading != "" {
		b.WriteString("#")
		b.WriteString(heading)
	}
	if alias != "" {
		b.WriteString("|")
		b.WriteString(alias)
	}
	b.WriteString("]]")
	return b.String()
}

// stemFromPath extracts the filename without extension from a path
func stemFromPath(p string) string {
	base := filepath.Base(p)
	return strings.TrimSuffix(base, filepath.Ext(base))
}
