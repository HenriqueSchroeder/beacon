package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/HenriqueSchroeder/beacon/pkg/config"
	"github.com/HenriqueSchroeder/beacon/pkg/content"
	"github.com/HenriqueSchroeder/beacon/pkg/create"
	"github.com/HenriqueSchroeder/beacon/pkg/templates"
	"github.com/spf13/cobra"
)

var (
	createType      string
	createTemplate  string
	createTags      string
	createPath      string
	createOverwrite bool
	createContent   string
	createAppend    bool
	createPrepend   bool
)

func init() {
	createCmd.Flags().StringVar(&createType, "type", "", "note type (determines output directory)")
	createCmd.Flags().StringVar(&createTemplate, "template", "", "template name to use (default: default)")
	createCmd.Flags().StringVar(&createTags, "tags", "", "tags to include (comma-separated)")
	createCmd.Flags().StringVar(&createPath, "path", "", "custom output path (relative to vault root)")
	createCmd.Flags().BoolVar(&createOverwrite, "overwrite", false, "overwrite existing file")
	createCmd.Flags().StringVar(&createContent, "content", "", "content to add to the note")
	createCmd.Flags().BoolVar(&createAppend, "append", false, "append content to end of note")
	createCmd.Flags().BoolVar(&createPrepend, "prepend", false, "prepend content after frontmatter")
	createCmd.MarkFlagsMutuallyExclusive("append", "prepend")
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
  beacon create "Inbox" --content "- [ ] task" --append
  echo "- [ ] task" | beacon create "Inbox" --append`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadFrom(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Resolve content: --content flag takes precedence, then stdin if piped
		text, err := resolveContent(createContent)
		if err != nil {
			return fmt.Errorf("failed to read content: %w", err)
		}

		// Validate: --append/--prepend require content
		if (createAppend || createPrepend) && strings.TrimSpace(text) == "" {
			return fmt.Errorf("--append/--prepend require content via --content or stdin")
		}

		loader := templates.NewTemplateLoader(cfg.VaultPath, cfg.TemplatesDir)
		creator := create.NewCreator(cfg.VaultPath, loader, cfg.TypePaths)

		var tags []string
		if createTags != "" {
			tags = strings.Split(createTags, ",")
			for i := range tags {
				tags[i] = strings.TrimSpace(tags[i])
			}
		}

		opts := create.CreateNoteOptions{
			Type:       createType,
			Title:      args[0],
			Template:   createTemplate,
			Tags:       tags,
			CustomPath: createPath,
			Overwrite:  createOverwrite,
		}

		// Resolve path before any operation so we know whether the note exists
		notePath, err := creator.ResolvePath(opts)
		if err != nil {
			return err
		}

		isContentOp := createAppend || createPrepend

		if isContentOp {
			// "get or create" semantics: create the note if it doesn't exist
			if !fileExists(notePath) {
				if _, err := creator.CreateNote(context.Background(), opts); err != nil {
					return err
				}
				fmt.Printf("✓ Note created: %s\n", notePath)
			}
		} else {
			// Standard create semantics: error if file exists (unless --overwrite)
			if _, err := creator.CreateNote(context.Background(), opts); err != nil {
				return err
			}
			fmt.Printf("✓ Note created: %s\n", notePath)

			if strings.TrimSpace(text) == "" {
				return nil
			}
		}

		// Apply content manipulation
		if strings.TrimSpace(text) == "" {
			return nil
		}

		m := content.New()
		switch {
		case createPrepend:
			if err := m.Prepend(notePath, text); err != nil {
				return err
			}
			fmt.Printf("✓ Prepended to: %s\n  %s\n", notePath, content.Snippet(text, 60))
		default: // --append or --content alone both append
			if err := m.Append(notePath, text); err != nil {
				return err
			}
			fmt.Printf("✓ Appended to: %s\n  %s\n", notePath, content.Snippet(text, 60))
		}

		return nil
	},
}

// resolveContent returns the content string: flag value if set, or stdin if piped.
func resolveContent(flag string) (string, error) {
	if flag != "" {
		return flag, nil
	}
	stat, err := os.Stdin.Stat()
	if err != nil {
		return "", nil // can't stat stdin — treat as no input
	}
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return "", nil // interactive terminal — no piped content
	}
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("failed to read stdin: %w", err)
	}
	return string(data), nil
}

// fileExists reports whether path is an existing regular file.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
