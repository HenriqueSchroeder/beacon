package create

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/HenriqueSchroeder/beacon/pkg/templates"
	"github.com/HenriqueSchroeder/beacon/pkg/vault"
)

// Creator renders and creates notes from templates
type Creator struct {
	vaultPath      string
	templateLoader *templates.TemplateLoader
	typePaths      map[string]string // Maps note type to directory path
}

// NewCreator creates a new Creator
func NewCreator(vaultPath string, templateLoader *templates.TemplateLoader, typePaths map[string]string) *Creator {
	return &Creator{
		vaultPath:      vaultPath,
		templateLoader: templateLoader,
		typePaths:      typePaths,
	}
}

// CreateNoteOptions holds options for creating a note
type CreateNoteOptions struct {
	Type       string   // Note type (determines output directory)
	Title      string   // Note title
	Template   string   // Template name to use
	Tags       []string // Tags to include in frontmatter
	CustomPath string   // Optional custom path (overrides type-based path)
	Overwrite  bool     // Whether to overwrite existing file
}

// RenderData holds data for template rendering
type RenderData struct {
	Title string
	Date  string    // YYYY-MM-DD format
	Tags  string    // Comma-separated tags
	Now   time.Time // Current time
}

// CreateNote creates a new note using a template
func (c *Creator) CreateNote(ctx context.Context, opts CreateNoteOptions) (string, error) {
	if opts.Title == "" {
		return "", fmt.Errorf("title is required")
	}

	if opts.Template == "" {
		opts.Template = "default"
	}

	// Load template
	templateContent, err := c.templateLoader.LoadTemplate(ctx, opts.Template)
	if err != nil {
		return "", fmt.Errorf("failed to load template: %w", err)
	}

	// Determine output path
	outputPath, err := c.resolvePath(opts)
	if err != nil {
		return "", fmt.Errorf("failed to resolve output path: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(outputPath); err == nil {
		if !opts.Overwrite {
			return "", fmt.Errorf("file already exists: %s (use --overwrite to force)", outputPath)
		}
	}

	// Create render data
	now := time.Now()
	renderData := RenderData{
		Title: opts.Title,
		Date:  now.Format("2006-01-02"),
		Tags:  strings.Join(opts.Tags, ", "),
		Now:   now,
	}

	// Render template
	renderedContent, err := renderTemplate(templateContent, renderData)
	if err != nil {
		return "", fmt.Errorf("failed to render template: %w", err)
	}

	// Ensure directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(outputPath, []byte(renderedContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return outputPath, nil
}

// resolvePath determines the output file path based on options
func (c *Creator) resolvePath(opts CreateNoteOptions) (string, error) {
	// If custom path provided, use it
	if opts.CustomPath != "" {
		// Make path absolute relative to vault
		if !filepath.IsAbs(opts.CustomPath) {
			return filepath.Join(c.vaultPath, opts.CustomPath), nil
		}
		return opts.CustomPath, nil
	}

	// Otherwise use type-based path
	var baseDir string
	if opts.Type != "" {
		var exists bool
		baseDir, exists = c.typePaths[opts.Type]
		if !exists {
			return "", fmt.Errorf("unknown note type: %q (available: %v)", opts.Type, c.listTypes())
		}
	} else {
		baseDir = "" // Root of vault
	}

	// Construct filename from title (replace spaces with underscores)
	filename := vault.SanitizeFilename(opts.Title) + ".md"

	fullPath := filepath.Join(c.vaultPath, baseDir, filename)
	return fullPath, nil
}

// renderTemplate renders a template string with data
func renderTemplate(templateStr string, data RenderData) (string, error) {
	// Create a text/template using Go struct notation
	tmpl, err := template.New("note").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var result strings.Builder
	if err := tmpl.Execute(&result, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return result.String(), nil
}

// listTypes returns available note types
func (c *Creator) listTypes() []string {
	var types []string
	for t := range c.typePaths {
		types = append(types, t)
	}
	return types
}
