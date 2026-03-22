package main

import (
	"context"
	"fmt"
	"time"

	"github.com/HenriqueSchroeder/beacon/pkg/config"
	"github.com/HenriqueSchroeder/beacon/pkg/daily"
	"github.com/HenriqueSchroeder/beacon/pkg/templates"
	"github.com/spf13/cobra"
)

var (
	dailyYesterday bool
	dailyTomorrow  bool
)

func init() {
	dailyCmd.Flags().BoolVar(&dailyYesterday, "yesterday", false, "open yesterday's daily note")
	dailyCmd.Flags().BoolVar(&dailyTomorrow, "tomorrow", false, "open tomorrow's daily note")
	dailyCmd.MarkFlagsMutuallyExclusive("yesterday", "tomorrow")
	rootCmd.AddCommand(dailyCmd)
}

var dailyCmd = &cobra.Command{
	Use:   "daily",
	Short: "Create or open today's daily note",
	Long: `Create or open the daily note for today (or a relative date).

Examples:
  beacon daily
  beacon daily --yesterday
  beacon daily --tomorrow`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadFrom(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		loader := templates.NewTemplateLoader(cfg.VaultPath, cfg.TemplatesDir)
		m := daily.NewManager(cfg.VaultPath, cfg.Daily.Folder, cfg.Daily.DateFormat, cfg.Daily.Template, loader, cfg.TypePaths)

		date := targetDate()

		result, err := m.GetOrCreate(context.Background(), date)
		if err != nil {
			return err
		}

		if result.Created {
			fmt.Printf("✓ Daily note created: %s\n", result.Path)
		} else {
			fmt.Printf("✓ Daily note found: %s\n", result.Path)
		}
		return nil
	},
}

// targetDate returns the date to operate on based on CLI flags.
func targetDate() time.Time {
	now := time.Now()
	switch {
	case dailyYesterday:
		return now.AddDate(0, 0, -1)
	case dailyTomorrow:
		return now.AddDate(0, 0, 1)
	default:
		return now
	}
}
