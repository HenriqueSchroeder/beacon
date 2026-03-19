package templates

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TemplateLoader loads note templates from the vault or fallback hardcoded templates
type TemplateLoader struct {
	vaultPath string
	templatesDir string
}

// NewTemplateLoader creates a new TemplateLoader
// templatesDir is relative to vaultPath (typically "700 - Recursos/Templates")
func NewTemplateLoader(vaultPath, templatesDir string) *TemplateLoader {
	return &TemplateLoader{
		vaultPath:    vaultPath,
		templatesDir: templatesDir,
	}
}

// LoadTemplate loads a template by name. First tries vault, then fallback hardcoded templates
// ctx reserved for future use
func (tl *TemplateLoader) LoadTemplate(ctx context.Context, templateName string) (string, error) {
	// Try to load from vault first
	templatePath := filepath.Join(tl.vaultPath, tl.templatesDir, templateName+".md")
	content, err := os.ReadFile(templatePath)
	if err == nil {
		return string(content), nil
	}

	// Fallback to hardcoded templates
	if fallback, exists := hardcodedTemplates[templateName]; exists {
		return fallback, nil
	}

	return "", fmt.Errorf("template %q not found (vault: %s, hardcoded: %v)", templateName, templatePath, listHardcodedTemplates())
}

// ListTemplates returns available template names (vault + hardcoded)
func (tl *TemplateLoader) ListTemplates(ctx context.Context) ([]string, error) {
	templates := make(map[string]bool)

	// Add hardcoded templates
	for name := range hardcodedTemplates {
		templates[name] = true
	}

	// Try to read from vault
	templatesPath := filepath.Join(tl.vaultPath, tl.templatesDir)
	if entries, err := os.ReadDir(templatesPath); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
				name := strings.TrimSuffix(entry.Name(), ".md")
				templates[name] = true
			}
		}
	}

	// Convert map to slice
	var result []string
	for name := range templates {
		result = append(result, name)
	}
	return result, nil
}

// Hardcoded templates with sensible defaults
var hardcodedTemplates = map[string]string{
	"default": `# {{.Title}}

**Date:** {{.Date}}
**Created:** {{.Now.Format "2006-01-02 15:04:05"}}
{{if .Tags}}**Tags:** {{.Tags}}{{end}}

## Content

`,

	"daily": `# Daily Note - {{.Date}}

**Created:** {{.Now.Format "2006-01-02 15:04:05"}}
{{if .Tags}}**Tags:** {{.Tags}}{{end}}

## Summary

## To-Do

## Notes

`,

	"project": `# {{.Title}}

**Created:** {{.Date}}
**Created at:** {{.Now.Format "2006-01-02 15:04:05"}}
{{if .Tags}}**Tags:** {{.Tags}}{{end}}

## Overview

## Objectives

## Status

## Tasks

- [ ] 

## References

`,

	"meeting": `# Meeting: {{.Title}}

**Date:** {{.Date}}
**Time:** {{.Now.Format "2006-01-02 15:04:05"}}
{{if .Tags}}**Attendees:** {{.Tags}}{{end}}

## Agenda

## Discussion

## Action Items

- [ ] 

## Notes

`,

	"template": `# Template: {{.Title}}

**Version:** 1.0
**Created:** {{.Now.Format "2006-01-02 15:04:05"}}
{{if .Tags}}**Tags:** {{.Tags}}{{end}}

## Purpose

## Usage

## Content

`,
}

func listHardcodedTemplates() []string {
	var names []string
	for name := range hardcodedTemplates {
		names = append(names, name)
	}
	return names
}
