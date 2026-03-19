# Link Validation Feature

## Overview

The Link Validation feature provides comprehensive checking for wiki links (`[[...]]`) in Obsidian vaults. It can:

- Detect invalid link targets (references to non-existent notes)
- Verify heading references are valid
- Suggest alternatives for broken links using fuzzy matching
- Generate detailed validation reports (text or JSON)
- Validate entire vault or specific files

## Architecture

The feature is implemented across three main components:

### 1. Link Parser (`pkg/links/parser.go`)

Parses wiki link syntax from markdown content:

```go
type Parser struct { ... }

parser := links.NewParser()
links := parser.Parse(content, "filename.md")
```

**Supported Link Types:**
- Simple: `[[note]]`
- With heading: `[[note#heading]]`
- With alias: `[[note|display text]]`
- Full: `[[note#heading|display text]]`

**Link struct:**
```go
type Link struct {
    Raw      string    // Original text: "[[note#heading|alias]]"
    Target   string    // Note name: "note"
    Heading  string    // Section: "heading"
    Alias    string    // Display text: "alias"
    Type     LinkType  // One of: Note, Heading, AliasNote, HeadingAlias
    Line     int       // Line number (1-indexed)
    Column   int       // Column position (0-indexed)
    FileName string    // Source file
}
```

### 2. Validator (`pkg/validate/validator.go`)

Validates links against the vault structure:

```go
vault := vault.NewFileVault(vaultPath, ignorePatterns)
validator := validate.NewValidator(vault, maxWorkers)

// Build index of all notes and headings
validator.BuildIndex(ctx)

// Validate single document
result := validator.ValidateDocument(ctx, note)

// Validate entire vault (parallel)
results, err := validator.ValidateAll(ctx)
```

**Features:**
- **Parallel validation**: Uses worker pool for concurrent document validation
- **Caching**: Results are cached to avoid re-validation
- **Fuzzy matching**: Suggests similar note names for broken links
- **Heading extraction**: Automatically extracts markdown headings for verification

### 3. CLI Command (`cmd/beacon/validate.go`)

Command-line interface for validation:

```bash
beacon validate [FLAGS]
```

**Flags:**
- `--json`: Output results as JSON
- `--file <path>`: Validate specific file only
- `--strict`: Exit with error if any invalid links found
- `--fix`: Attempt to fix broken links (planned)
- `--use-cache`: Use validation cache (planned)

## Usage Examples

### Basic validation of entire vault

```bash
beacon validate
```

Output shows summary and issues:
```
Validation Summary
===================
Total documents: 42
Total links: 156
Valid links: 152
Invalid links: 4
Documents with issues: 2

Issues by document:
-------------------

docs/architecture.md (2 invalid links)
  Line 5, Col 23: [[design-pattern]]
    Error: target note not found: design-pattern
    Hint: Did you mean 'design-patterns'?

  Line 12, Col 15: [[api#request-handler]]
    Error: heading not found in api: request-handler
    Hint: Available headings: requests, responses, errors
```

### Validate specific file

```bash
beacon validate --file docs/guide.md
```

### JSON output for programmatic use

```bash
beacon validate --json > validation-report.json
```

Output:
```json
[
  {
    "file": "docs/guide.md",
    "total_links": 5,
    "valid_links": 4,
    "invalid_links": 1,
    "invalid_links_details": [
      {
        "link": "[[nonexistent]]",
        "target": "nonexistent",
        "heading": "",
        "alias": "",
        "line": 10,
        "column": 8,
        "reason": "target note not found: nonexistent",
        "suggestion": "Did you mean 'next-steps'?"
      }
    ]
  }
]
```

### Strict mode (CI/CD integration)

```bash
beacon validate --strict
```

Exits with error code 1 if any invalid links found. Perfect for CI pipelines.

## Configuration

Add validation settings to `.beacon.yml`:

```yaml
vault_path: /path/to/vault
ignore:
  - .obsidian
  - node_modules

# Validation configuration
validation:
  enabled: true
  fuzzy_threshold: 0.8  # 0.0-1.0, higher = stricter matching
  strict_mode: false    # Default behavior for --strict flag
  ignore_patterns:
    - "*.tmp"
    - "draft-*"
```

## Technical Details

### Link Parsing

The parser uses regex to find wiki links and parses their components:

```
Input: "[[my-note#section|Click here]]"
↓
Regex: \[\[([^\[\]]+)\]\]
↓
Captured: "my-note#section|Click here"
↓
Parsed:
- Target: "my-note"
- Heading: "section"
- Alias: "Click here"
- Type: LinkTypeHeadingAlias
```

### Validation Index

To efficiently validate links, the validator builds a two-part index:

1. **noteIndex**: Map of all valid note names (by full path, filename, and stem)
2. **headingIndex**: Map of note names to their markdown headings

Example:
```
noteIndex: {
  "guide.md": true,
  "guide": true,
  "docs/api.md": true,
  "api.md": true,
  "api": true,
}

headingIndex: {
  "api.md": ["installation", "configuration", "usage", "examples"],
  "api": ["installation", "configuration", "usage", "examples"],
}
```

### Fuzzy Matching

Broken link suggestions use Levenshtein distance similarity:

```
Input: [[desgin-pattern]]
↓
Similarity scores:
- "design-pattern": 0.96 ✓ (suggestion)
- "pattern": 0.71
- "design": 0.64
↓
Suggestion: "Did you mean 'design-pattern'?"
```

### Parallel Validation

The validator uses a worker pool pattern for concurrent validation:

```
vault.ListNotes()
    ↓
Create channel of notes
    ↓
Spawn N workers
    ↓
Each worker:
  - Receives note from channel
  - Parses links
  - Validates each link
  - Sends results
    ↓
Collect all results
```

This makes vault-wide validation fast even for large vaults with thousands of notes.

## Testing

### Run all tests

```bash
go test -v ./...
```

### Run specific test suites

```bash
# Parser tests
go test -v ./pkg/links/...

# Validator tests
go test -v ./pkg/validate/...

# Integration tests
go test -v -run Integration ./pkg/links/...
```

### Test coverage

```bash
go test -cover ./...
```

## Performance Characteristics

- **Link Parsing**: O(n) where n = document length
- **Index Building**: O(m) where m = total vault size
- **Single Document Validation**: O(l) where l = number of links in document
- **Vault-wide Validation**: O(m/w) with parallel workers, where w = number of workers

For a typical vault (1000 notes, 5000 links total):
- Index building: ~50-100ms
- Single document validation: ~1-5ms
- Vault-wide validation: ~100-200ms (with 4 workers)

## Future Enhancements

### Planned Features

1. **Link Fixing** (`--fix` flag)
   - Auto-rename broken links to matching notes
   - Handle heading changes in referenced notes

2. **Link Statistics**
   - Count internal link density
   - Identify orphaned notes
   - Find most-referenced notes

3. **Link Cleanup**
   - Remove duplicate links
   - Fix inconsistent link formats

4. **Integration**
   - Watch mode for real-time validation
   - Git hooks for pre-commit validation
   - GitHub Actions integration

## Troubleshooting

### "target note not found" errors

Ensure the note exists and is not ignored:

```bash
# Check if note file exists
find /path/to/vault -name "note.md"

# Check ignore patterns in .beacon.yml
# Links to notes in ignored directories will fail validation
```

### Performance issues on large vaults

Use `--file` to validate specific files instead of entire vault:

```bash
beacon validate --file docs/section/specific-note.md
```

Or reduce the fuzzy matching threshold in config to skip expensive similarity calculations:

```yaml
validation:
  fuzzy_threshold: 0.95  # Higher = less fuzzy matching
```

### Heading validation failing unexpectedly

Remember that headings are case-insensitive but must match exactly (after normalization):

Valid:
- Markdown: `## Getting Started`
- Link: `[[guide#getting started]]` ✓
- Link: `[[guide#Getting Started]]` ✓

Invalid:
- Link: `[[guide#starting]]` ✗ (doesn't match)

## Contributing

To extend the validator:

1. **Add new link type**: Update `LinkType` enum and parser
2. **Add validation rule**: Implement in `validateLink()`
3. **Add tests**: Create test cases in `*_test.go` files
4. **Update CLI**: Modify command flags and output formatting

## License

See main README for license information.
