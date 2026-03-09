package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/HenriqueSchroeder/beacon/pkg/config"
	"github.com/HenriqueSchroeder/beacon/pkg/search"
	"github.com/spf13/cobra"
)

var jsonOutput bool

func init() {
	searchCmd.Flags().BoolVar(&jsonOutput, "json", false, "output results as JSON")
	rootCmd.AddCommand(searchCmd)
}

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search notes by content",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]

		cfg, err := config.LoadFrom(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		s, err := search.NewRipgrepSearcher(cfg.VaultPath, cfg.Ignore)
		if err != nil {
			return fmt.Errorf("failed to create searcher: %w", err)
		}

		results, err := s.SearchContent(context.Background(), query)
		if err != nil {
			return fmt.Errorf("search failed: %w", err)
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
		fmt.Printf("%s:%d\n", r.Path, r.Line)

		for _, line := range r.ContextBefore {
			fmt.Printf("  %s\n", line)
		}
		fmt.Printf("> %s\n", r.Match)
		for _, line := range r.ContextAfter {
			fmt.Printf("  %s\n", line)
		}
	}
}
