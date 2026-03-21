package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/HenriqueSchroeder/beacon/pkg/config"
	"github.com/HenriqueSchroeder/beacon/pkg/search"
	"github.com/HenriqueSchroeder/beacon/pkg/vault"
	"github.com/spf13/cobra"
)

var (
	jsonOutput     bool
	searchTags     string
	searchType     string
	searchFilename bool
)

func init() {
	searchCmd.Flags().BoolVar(&jsonOutput, "json", false, "output results as JSON")
	searchCmd.Flags().StringVar(&searchTags, "tags", "", "search by tags (comma-separated)")
	searchCmd.Flags().StringVar(&searchType, "type", "", "search by note type")
	searchCmd.Flags().BoolVar(&searchFilename, "filename", false, "search by filename")
	rootCmd.AddCommand(searchCmd)
}

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search notes by content, tags, or type",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadFrom(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		modeCount := 0
		if searchTags != "" {
			modeCount++
		}
		if searchType != "" {
			modeCount++
		}
		if searchFilename {
			modeCount++
		}
		if len(args) > 0 {
			modeCount++
		}
		if modeCount > 1 && !searchFilename {
			return fmt.Errorf("only one search mode can be used at a time")
		}
		if searchFilename && (searchTags != "" || searchType != "") {
			return fmt.Errorf("only one search mode can be used at a time")
		}

		var results []search.SearchResult

		switch {
		case searchTags != "":
			tags := strings.Split(searchTags, ",")
			for i := range tags {
				tags[i] = strings.TrimSpace(tags[i])
			}
			v, err := vault.NewFileVault(cfg.VaultPath, cfg.Ignore)
			if err != nil {
				return fmt.Errorf("failed to open vault: %w", err)
			}
			s := search.NewVaultSearcher(v)
			results, err = s.SearchTags(context.Background(), tags)
			if err != nil {
				return fmt.Errorf("search failed: %w", err)
			}

		case searchType != "":
			v, err := vault.NewFileVault(cfg.VaultPath, cfg.Ignore)
			if err != nil {
				return fmt.Errorf("failed to open vault: %w", err)
			}
			s := search.NewVaultSearcher(v)
			results, err = s.SearchByType(context.Background(), searchType)
			if err != nil {
				return fmt.Errorf("search failed: %w", err)
			}

		case searchFilename:
			if len(args) != 1 {
				return fmt.Errorf("--filename requires a query argument")
			}
			v, err := vault.NewFileVault(cfg.VaultPath, cfg.Ignore)
			if err != nil {
				return fmt.Errorf("failed to open vault: %w", err)
			}
			s := search.NewVaultSearcher(v)
			results, err = s.SearchByFilename(context.Background(), args[0])
			if err != nil {
				return fmt.Errorf("search failed: %w", err)
			}

		case len(args) > 0:
			if len(args) != 1 {
				return fmt.Errorf("provide exactly one query")
			}
			s, err := search.NewRipgrepSearcher(cfg.VaultPath, cfg.Ignore)
			if err != nil {
				return fmt.Errorf("failed to create searcher: %w", err)
			}
			results, err = s.SearchContent(context.Background(), args[0])
			if err != nil {
				return fmt.Errorf("search failed: %w", err)
			}

		default:
			return fmt.Errorf("provide a query, --tags, --type, or --filename")
		}

		if len(results) == 0 {
			fmt.Println("No results found.")
			return nil
		}

		if jsonOutput {
			return printJSON(results)
		}

		printPlain(results)
		return nil
	},
}

func printJSON(results []search.SearchResult) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(results)
}

func printPlain(results []search.SearchResult) {
	for i, r := range results {
		if i > 0 {
			fmt.Println()
		}
		if r.Line > 0 {
			fmt.Printf("%s:%d\n", r.Path, r.Line)
		} else {
			fmt.Printf("%s\n", r.Path)
		}

		for _, line := range r.ContextBefore {
			fmt.Printf("  %s\n", line)
		}
		fmt.Printf("> %s\n", r.Match)
		for _, line := range r.ContextAfter {
			fmt.Printf("  %s\n", line)
		}
	}
}
