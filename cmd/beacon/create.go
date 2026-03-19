package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/HenriqueSchroeder/beacon/pkg/config"
	"github.com/HenriqueSchroeder/beacon/pkg/create"
	"github.com/HenriqueSchroeder/beacon/pkg/templates"
	"github.com/spf13/cobra"
)

var (
	createType       string
	createTemplate   string
	createTags       string
	createPath       string
	createOverwrite  bool
)

func init() {
	createCmd.Flags().StringVar(&createType, "type", "", "note type (determines output directory)")
	createCmd.Flags().StringVar(&createTemplate, "template", "", "template name to use (default: default)")
	createCmd.Flags().StringVar(&createTags, "tags", "", "tags to include (comma-separated)")
	createCmd.Flags().StringVar(&createPath, "path", "", "custom output path (relative to vault root)")
	createCmd.Flags().BoolVar(&createOverwrite, "overwrite", false, "overwrite existing file")
	rootCmd.AddCommand(createCmd)
}

var createCmd = &cobra.Command{
	Use:   "create [title]",
	Short: "Create a new note from a template",
	Long: `Create a new note from a template.

Examples:
  beacon create "My Note" --type=daily --template=daily
  beacon create "Meeting Notes" --tags="work,urgent" --template=meeting
  beacon create "Project X" --type=projects --path="Active/Project X.md"
  beacon create "Untitled" --overwrite`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadFrom(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Build type_paths from config (will add to Config struct in next step)
		typePaths := getTypePaths(cfg)

		// Create template loader
		loader := templates.NewTemplateLoader(cfg.VaultPath, "700 - Recursos/Templates")

		// Create creator
		creator := create.NewCreator(cfg.VaultPath, loader, typePaths)

		// Parse tags
		var tags []string
		if createTags != "" {
			tags = strings.Split(createTags, ",")
			for i := range tags {
				tags[i] = strings.TrimSpace(tags[i])
			}
		}

		// Create note
		opts := create.CreateNoteOptions{
			Type:       createType,
			Title:      args[0],
			Template:   createTemplate,
			Tags:       tags,
			CustomPath: createPath,
			Overwrite:  createOverwrite,
		}

		outputPath, err := creator.CreateNote(context.Background(), opts)
		if err != nil {
			return err
		}

		fmt.Printf("✓ Note created: %s\n", outputPath)
		return nil
	},
}

// getTypePaths returns the type_paths mapping from config
func getTypePaths(cfg *config.Config) map[string]string {
	if cfg.TypePaths != nil && len(cfg.TypePaths) > 0 {
		return cfg.TypePaths
	}

	// Fallback defaults if not configured
	return map[string]string{
		"daily":     "100 - Diário",
		"projects":  "200 - Projetos",
		"resources": "700 - Recursos",
		"work":      "300 - Trabalho",
		"personal":  "400 - Pessoal",
	}
}
