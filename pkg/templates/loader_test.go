package templates

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestNewTemplateLoader(t *testing.T) {
	loader := NewTemplateLoader("/vault", "templates")
	if loader.vaultPath != "/vault" {
		t.Errorf("expected vaultPath=/vault, got %s", loader.vaultPath)
	}
	if loader.templatesDir != "templates" {
		t.Errorf("expected templatesDir=templates, got %s", loader.templatesDir)
	}
}

func TestLoadTemplateFromHardcoded(t *testing.T) {
	loader := NewTemplateLoader("/nonexistent", "templates")
	ctx := context.Background()

	content, err := loader.LoadTemplate(ctx, "default")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if content == "" {
		t.Errorf("expected non-empty template, got empty")
	}

	if !contains(content, "{{.Title}}") {
		t.Errorf("expected template to contain {{.Title}}, got: %s", content)
	}
}

func TestLoadTemplateNotFound(t *testing.T) {
	loader := NewTemplateLoader("/nonexistent", "templates")
	ctx := context.Background()

	_, err := loader.LoadTemplate(ctx, "nonexistent_template_xyz")
	if err == nil {
		t.Errorf("expected error for nonexistent template, got nil")
	}
}

func TestLoadTemplateFromVault(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	templatesDir := filepath.Join(tmpDir, "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	// Write test template
	testTemplate := "# Test\n{{title}}"
	templatePath := filepath.Join(templatesDir, "test.md")
	if err := os.WriteFile(templatePath, []byte(testTemplate), 0644); err != nil {
		t.Fatalf("failed to write test template: %v", err)
	}

	loader := NewTemplateLoader(tmpDir, "templates")
	ctx := context.Background()

	content, err := loader.LoadTemplate(ctx, "test")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if content != testTemplate {
		t.Errorf("expected %q, got %q", testTemplate, content)
	}
}

func TestListTemplates(t *testing.T) {
	loader := NewTemplateLoader("/nonexistent", "templates")
	ctx := context.Background()

	templates, err := loader.ListTemplates(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(templates) == 0 {
		t.Errorf("expected templates to include hardcoded templates")
	}

	// Check that at least "default" template exists
	found := false
	for _, name := range templates {
		if name == "default" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'default' in templates list, got %v", templates)
	}
}

func TestHardcodedTemplates(t *testing.T) {
	tests := []struct {
		name       string
		templateKey string
		wantContent string
	}{
		{"default", "default", "{{.Title}}"},
		{"daily", "daily", "Daily Note"},
		{"project", "project", "Objectives"},
		{"meeting", "meeting", "Agenda"},
		{"template", "template", "Purpose"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, exists := hardcodedTemplates[tt.templateKey]
			if !exists {
				t.Errorf("expected template %q to exist", tt.templateKey)
			}
			if !contains(content, tt.wantContent) {
				t.Errorf("expected template to contain %q, got: %s", tt.wantContent, content)
			}
		})
	}
}

func contains(s, substring string) bool {
	for i := 0; i < len(s); i++ {
		if i+len(substring) > len(s) {
			break
		}
		if s[i:i+len(substring)] == substring {
			return true
		}
	}
	return false
}
