package create

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/HenriqueSchroeder/beacon/pkg/templates"
	"github.com/HenriqueSchroeder/beacon/pkg/vault"
)

func TestNewCreator(t *testing.T) {
	typePaths := map[string]string{
		"daily": "100 - Diário",
	}
	loader := templates.NewTemplateLoader("/vault", "templates")
	creator := NewCreator("/vault", loader, typePaths)

	if creator.vaultPath != "/vault" {
		t.Errorf("expected vaultPath=/vault, got %s", creator.vaultPath)
	}
}

func TestResolvePathUsesSharedFilenameSanitization(t *testing.T) {
	tmpDir := t.TempDir()
	loader := templates.NewTemplateLoader(tmpDir, "templates")
	creator := NewCreator(tmpDir, loader, map[string]string{})

	path, err := creator.resolvePath(CreateNoteOptions{Title: "Meeting: Notes"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasSuffix(path, vault.SanitizeFilename("Meeting: Notes")+".md") {
		t.Fatalf("expected shared filename sanitization in path, got %s", path)
	}
}

func TestResolvePath(t *testing.T) {
	tmpDir := t.TempDir()
	typePaths := map[string]string{
		"daily":    "100 - Diário",
		"projects": "200 - Projetos",
	}

	loader := templates.NewTemplateLoader(tmpDir, "templates")
	creator := NewCreator(tmpDir, loader, typePaths)

	tests := []struct {
		name       string
		opts       CreateNoteOptions
		shouldFail bool
		check      func(path string) bool
	}{
		{
			name: "with type",
			opts: CreateNoteOptions{
				Type:  "daily",
				Title: "My Note",
			},
			shouldFail: false,
			check: func(path string) bool {
				return strings.Contains(path, "100 - Diário") &&
					strings.HasSuffix(path, "My_Note.md")
			},
		},
		{
			name: "with custom path",
			opts: CreateNoteOptions{
				Title:      "Test",
				CustomPath: "Custom/Path.md",
			},
			shouldFail: false,
			check: func(path string) bool {
				return strings.HasSuffix(path, "Custom/Path.md")
			},
		},
		{
			name: "invalid type",
			opts: CreateNoteOptions{
				Type:  "invalid",
				Title: "Test",
			},
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := creator.resolvePath(tt.opts)
			if tt.shouldFail && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.shouldFail && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.shouldFail && !tt.check(path) {
				t.Errorf("path check failed for: %s", path)
			}
		})
	}
}

func TestRenderTemplate(t *testing.T) {
	data := RenderData{
		Title: "Test Note",
		Date:  "2026-03-19",
		Tags:  "tag1, tag2",
		Now:   time.Now(),
	}

	templateStr := `# {{.Title}}

Date: {{.Date}}
Tags: {{.Tags}}`

	result, err := renderTemplate(templateStr, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "Test Note") {
		t.Errorf("expected rendered template to contain title")
	}

	if !strings.Contains(result, "2026-03-19") {
		t.Errorf("expected rendered template to contain date")
	}

	if !strings.Contains(result, "tag1, tag2") {
		t.Errorf("expected rendered template to contain tags")
	}
}

func TestCreateNote(t *testing.T) {
	tmpDir := t.TempDir()

	// Create templates directory
	templatesDir := filepath.Join(tmpDir, "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatalf("failed to create templates dir: %v", err)
	}

	typePaths := map[string]string{
		"daily": "100 - Diário",
	}

	loader := templates.NewTemplateLoader(tmpDir, "templates")
	creator := NewCreator(tmpDir, loader, typePaths)

	opts := CreateNoteOptions{
		Type:     "daily",
		Title:    "Test Note",
		Template: "default",
		Tags:     []string{"tag1", "tag2"},
	}

	ctx := context.Background()
	path, err := creator.CreateNote(ctx, opts)
	if err != nil {
		t.Fatalf("failed to create note: %v", err)
	}

	// Verify file was created
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("file was not created: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "Test Note") {
		t.Errorf("expected note to contain title, got: %s", contentStr)
	}
}

func TestCreateNoteWithCustomPath(t *testing.T) {
	tmpDir := t.TempDir()

	typePaths := map[string]string{}
	loader := templates.NewTemplateLoader(tmpDir, "templates")
	creator := NewCreator(tmpDir, loader, typePaths)

	opts := CreateNoteOptions{
		Title:      "Custom Note",
		Template:   "default",
		CustomPath: "Custom/Subfolder/Note.md",
	}

	ctx := context.Background()
	path, err := creator.CreateNote(ctx, opts)
	if err != nil {
		t.Fatalf("failed to create note: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file was not created: %v", err)
	}

	// Verify path structure
	if !strings.Contains(path, "Custom") {
		t.Errorf("expected custom path in result")
	}
}

func TestCreateNoteFileExists(t *testing.T) {
	tmpDir := t.TempDir()

	typePaths := map[string]string{}
	loader := templates.NewTemplateLoader(tmpDir, "templates")
	creator := NewCreator(tmpDir, loader, typePaths)

	// Create first note
	opts := CreateNoteOptions{
		Title:      "Duplicate",
		Template:   "default",
		CustomPath: "Note.md",
	}

	ctx := context.Background()
	_, err := creator.CreateNote(ctx, opts)
	if err != nil {
		t.Fatalf("failed to create first note: %v", err)
	}

	// Try to create duplicate without overwrite
	_, err = creator.CreateNote(ctx, opts)
	if err == nil {
		t.Errorf("expected error when file exists, got nil")
	}

	// Try with overwrite
	opts.Overwrite = true
	_, err = creator.CreateNote(ctx, opts)
	if err != nil {
		t.Errorf("expected no error with overwrite, got: %v", err)
	}
}

func TestCreateNoteNoTitle(t *testing.T) {
	tmpDir := t.TempDir()

	loader := templates.NewTemplateLoader(tmpDir, "templates")
	creator := NewCreator(tmpDir, loader, map[string]string{})

	opts := CreateNoteOptions{
		Title:    "",
		Template: "default",
	}

	ctx := context.Background()
	_, err := creator.CreateNote(ctx, opts)
	if err == nil {
		t.Errorf("expected error for empty title, got nil")
	}
}
