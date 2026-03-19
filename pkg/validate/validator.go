package validate

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/HenriqueSchroeder/beacon/pkg/links"
	"github.com/HenriqueSchroeder/beacon/pkg/vault"
)

// ValidationResult represents the result of validating a link
type ValidationResult struct {
	Link       links.Link
	Valid      bool
	Reason     string // Why the link is invalid (empty if valid)
	Suggestion string // Suggested fix (if available)
}

// DocumentValidation contains validation results for a document
type DocumentValidation struct {
	FilePath   string
	Results    []ValidationResult
	TotalLinks int
	ValidLinks int
}

// Validator validates wiki links in the vault
type Validator struct {
	vault          vault.Vault
	noteIndex      map[string]bool      // Map of valid note paths
	headingIndex   map[string][]string  // Map of note paths to headings
	cache          map[string]*DocumentValidation
	cacheMutex     sync.RWMutex
	parser         *links.Parser
	maxWorkers     int
	fuzzyThreshold float64 // Threshold for fuzzy matching (0.0-1.0)
}

// NewValidator creates a new link validator
func NewValidator(vault vault.Vault, maxWorkers int) *Validator {
	if maxWorkers <= 0 {
		maxWorkers = 4
	}
	return &Validator{
		vault:          vault,
		noteIndex:      make(map[string]bool),
		headingIndex:   make(map[string][]string),
		cache:          make(map[string]*DocumentValidation),
		parser:         links.NewParser(),
		maxWorkers:     maxWorkers,
		fuzzyThreshold: 0.8,
	}
}

// BuildIndex builds the index of all notes and their headings
func (v *Validator) BuildIndex(ctx context.Context) error {
	notes, err := v.vault.ListNotes(ctx)
	if err != nil {
		return fmt.Errorf("validate: failed to list notes: %w", err)
	}

	for _, note := range notes {
		// Index note by both full path and filename
		v.noteIndex[strings.ToLower(note.Path)] = true
		v.noteIndex[strings.ToLower(filepath.Base(note.Path))] = true
		v.noteIndex[strings.ToLower(strings.TrimSuffix(filepath.Base(note.Path), ".md"))] = true

		// Extract headings from content
		headings := extractHeadings(note.Content)
		if len(headings) > 0 {
			v.headingIndex[strings.ToLower(note.Path)] = headings
			v.headingIndex[strings.ToLower(filepath.Base(note.Path))] = headings
			v.headingIndex[strings.ToLower(strings.TrimSuffix(filepath.Base(note.Path), ".md"))] = headings
		}
	}

	return nil
}

// ValidateDocument validates all links in a document
func (v *Validator) ValidateDocument(ctx context.Context, note *vault.Note) *DocumentValidation {
	v.cacheMutex.RLock()
	if cached, ok := v.cache[note.Path]; ok {
		v.cacheMutex.RUnlock()
		return cached
	}
	v.cacheMutex.RUnlock()

	links := v.parser.Parse(note.Content, note.Path)
	results := make([]ValidationResult, 0, len(links))

	for _, link := range links {
		result := v.validateLink(link)
		results = append(results, result)
	}

	validation := &DocumentValidation{
		FilePath:   note.Path,
		Results:    results,
		TotalLinks: len(links),
		ValidLinks: countValidLinks(results),
	}

	v.cacheMutex.Lock()
	v.cache[note.Path] = validation
	v.cacheMutex.Unlock()

	return validation
}

// ValidateAll validates all documents in the vault (with parallel workers)
func (v *Validator) ValidateAll(ctx context.Context) ([]DocumentValidation, error) {
	notes, err := v.vault.ListNotes(ctx)
	if err != nil {
		return nil, fmt.Errorf("validate: failed to list notes: %w", err)
	}

	// Use worker pool pattern
	noteChan := make(chan vault.Note, len(notes))
	resultChan := make(chan DocumentValidation, len(notes))
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < v.maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for note := range noteChan {
				select {
				case <-ctx.Done():
					return
				default:
				}
				result := v.ValidateDocument(ctx, &note)
				resultChan <- *result
			}
		}()
	}

	// Send notes to workers
	go func() {
		for _, note := range notes {
			noteChan <- note
		}
		close(noteChan)
	}()

	// Collect results
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	var results []DocumentValidation
	for result := range resultChan {
		results = append(results, result)
	}

	return results, nil
}

// validateLink validates a single link
func (v *Validator) validateLink(link links.Link) ValidationResult {
	result := ValidationResult{
		Link:  link,
		Valid: true,
	}

	if !link.IsValid() {
		result.Valid = false
		result.Reason = "malformed link"
		return result
	}

	// Check if target exists
	targetPath := normalizeTarget(link.Target)
	if _, exists := v.noteIndex[targetPath]; !exists {
		result.Valid = false
		result.Reason = fmt.Sprintf("target note not found: %s", link.Target)

		// Try to find a suggestion via fuzzy matching
		suggestion := v.findFuzzySuggestion(link.Target)
		if suggestion != "" {
			result.Suggestion = fmt.Sprintf("Did you mean '%s'?", suggestion)
		}
		return result
	}

	// If heading is specified, check if it exists
	if link.Heading != "" {
		headings, ok := v.headingIndex[targetPath]
		if !ok || !stringInSlice(normalizeHeading(link.Heading), headings) {
			result.Valid = false
			result.Reason = fmt.Sprintf("heading not found in %s: %s", link.Target, link.Heading)

			// Suggest available headings
			if ok && len(headings) > 0 {
				result.Suggestion = fmt.Sprintf("Available headings: %s", strings.Join(headings[:min(3, len(headings))], ", "))
			}
			return result
		}
	}

	return result
}

// findFuzzySuggestion tries to find a similar note name using fuzzy matching
func (v *Validator) findFuzzySuggestion(target string) string {
	bestMatch := ""
	bestScore := 0.0

	for notePath := range v.noteIndex {
		score := levenshteinSimilarity(strings.ToLower(target), strings.ToLower(notePath))
		if score > v.fuzzyThreshold && score > bestScore {
			bestScore = score
			bestMatch = notePath
		}
	}

	return bestMatch
}

// ClearCache clears the validation cache
func (v *Validator) ClearCache() {
	v.cacheMutex.Lock()
	defer v.cacheMutex.Unlock()
	v.cache = make(map[string]*DocumentValidation)
}

// Helper functions

// extractHeadings extracts all heading names from markdown content
func extractHeadings(content string) []string {
	var headings []string
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			// Extract heading text
			heading := strings.TrimLeft(trimmed, "#")
			heading = strings.TrimSpace(heading)
			if heading != "" {
				headings = append(headings, normalizeHeading(heading))
			}
		}
	}

	return headings
}

// normalizeTarget normalizes a link target for comparison
func normalizeTarget(target string) string {
	// Remove .md extension if present, normalize case for comparison
	base := strings.ToLower(strings.TrimSuffix(target, ".md"))
	return strings.TrimSpace(base)
}

// normalizeHeading normalizes a heading for comparison
func normalizeHeading(heading string) string {
	return strings.ToLower(strings.TrimSpace(heading))
}

// stringInSlice checks if a string is in a slice
func stringInSlice(s string, slice []string) bool {
	for _, item := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}

// countValidLinks counts how many links are valid
func countValidLinks(results []ValidationResult) int {
	count := 0
	for _, r := range results {
		if r.Valid {
			count++
		}
	}
	return count
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// levenshteinSimilarity calculates similarity between two strings (0.0 to 1.0)
// using normalized Levenshtein distance
func levenshteinSimilarity(a, b string) float64 {
	distance := levenshteinDistance(a, b)
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}

	if maxLen == 0 {
		return 1.0
	}

	return 1.0 - (float64(distance) / float64(maxLen))
}

// levenshteinDistance calculates the Levenshtein distance between two strings
func levenshteinDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	aLen := len(a)
	bLen := len(b)

	if aLen > bLen {
		a, b = b, a
		aLen, bLen = bLen, aLen
	}

	prevRow := make([]int, aLen+1)
	for i := range prevRow {
		prevRow[i] = i
	}

	for j := 1; j <= bLen; j++ {
		currRow := make([]int, aLen+1)
		currRow[0] = j

		for i := 1; i <= aLen; i++ {
			add := prevRow[i] + 1
			del := currRow[i-1] + 1
			sub := prevRow[i-1]

			if a[i-1] != b[j-1] {
				sub++
			}

			currRow[i] = min(add, min(del, sub))
		}

		prevRow = currRow
	}

	return prevRow[aLen]
}
