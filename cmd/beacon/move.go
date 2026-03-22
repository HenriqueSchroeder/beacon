package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/HenriqueSchroeder/beacon/pkg/config"
	"github.com/HenriqueSchroeder/beacon/pkg/move"
	"github.com/HenriqueSchroeder/beacon/pkg/vault"
	"github.com/spf13/cobra"
)

var moveDryRun bool

func init() {
	moveCmd.Flags().BoolVar(&moveDryRun, "dry-run", false, "show what would change without applying")
	rootCmd.AddCommand(moveCmd)
}

var moveCmd = &cobra.Command{
	Use:   "move <source> <destination>",
	Short: "Move or rename a note, updating all backlinks",
	Long: `Move or rename a note within the vault.

All wiki-links ([[...]]) across the vault that reference the old note name
are automatically updated to point to the new name. Headings and aliases
in links are preserved.

If only the folder changes (same filename), no link updates are needed
since Obsidian uses shortest-path resolution by default.

Examples:
  beacon move "My Note.md" "Renamed Note.md"
  beacon move "inbox/Draft.md" "published/Draft.md"
  beacon move "Old Name.md" "folder/New Name.md" --dry-run`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		source := ensureMdExtension(args[0])
		dest := ensureMdExtension(args[1])

		cfg, err := config.LoadFrom(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		v, err := vault.NewFileVault(cfg.VaultPath, cfg.Ignore)
		if err != nil {
			return fmt.Errorf("failed to open vault: %w", err)
		}

		ctx := context.Background()
		mover := move.NewMover(cfg.VaultPath, v)

		result, err := mover.Plan(ctx, source, dest)
		if err != nil {
			return err
		}

		if moveDryRun {
			printDryRun(result)
			return nil
		}

		summary, err := mover.Apply(result)
		if err != nil {
			return err
		}

		printSummary(result, summary)

		if len(summary.Errors) > 0 {
			for _, e := range summary.Errors {
				fmt.Fprintf(os.Stderr, "  error: %s\n", e)
			}
			return fmt.Errorf("move completed with %d error(s)", len(summary.Errors))
		}

		return nil
	},
}

func printDryRun(result *move.MoveResult) {
	fmt.Printf("Would rename: %s -> %s\n", result.Source, result.Dest)

	if !result.NeedsRelink {
		fmt.Println("\nNo link updates needed (filename unchanged).")
		return
	}

	if len(result.Updates) == 0 {
		fmt.Println("\nNo files reference this note.")
		return
	}

	totalLinks := 0
	fmt.Printf("\nFiles to update (%d):\n", len(result.Updates))
	for _, u := range result.Updates {
		n := len(u.Replacements)
		totalLinks += n
		fmt.Printf("  %-40s — %d link(s)\n", u.Path, n)
	}
	fmt.Printf("\nTotal: %d link(s) in %d file(s)\n", totalLinks, len(result.Updates))
}

func printSummary(result *move.MoveResult, summary *move.MoveSummary) {
	fmt.Printf("Moved: %s -> %s\n", result.Source, result.Dest)
	if summary.LinksUpdated > 0 {
		fmt.Printf("Updated: %d link(s) in %d file(s)\n", summary.LinksUpdated, summary.FilesUpdated)
	}
}

func ensureMdExtension(path string) string {
	if filepath.Ext(path) == "" {
		return path + ".md"
	}
	return path
}
