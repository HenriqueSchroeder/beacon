package main

import (
	"context"
	"fmt"

	"github.com/HenriqueSchroeder/beacon/pkg/config"
	"github.com/HenriqueSchroeder/beacon/pkg/tasks"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(tasksCmd)
}

var tasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "List pending tasks across the vault",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadFrom(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		s, err := tasks.NewSearcher(cfg.VaultPath, cfg.Ignore)
		if err != nil {
			return fmt.Errorf("failed to create task searcher: %w", err)
		}

		results, err := s.ListPending(context.Background())
		if err != nil {
			return fmt.Errorf("failed to list tasks: %w", err)
		}

		if len(results) == 0 {
			fmt.Println("No tasks found.")
			return nil
		}

		for _, task := range results {
			fmt.Printf("%s:%d %s\n", task.Path, task.Line, task.Text)
		}

		return nil
	},
}
