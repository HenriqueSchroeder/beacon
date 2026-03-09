package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/HenriqueSchroeder/beacon/pkg/config"
	"github.com/HenriqueSchroeder/beacon/pkg/vault"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List notes in the vault",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadFrom(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		v, err := vault.NewFileVault(cfg.VaultPath, cfg.Ignore)
		if err != nil {
			return fmt.Errorf("failed to open vault: %w", err)
		}

		notes, err := v.ListNotes(context.Background())
		if err != nil {
			return fmt.Errorf("failed to list notes: %w", err)
		}

		if len(notes) == 0 {
			fmt.Println("No notes found.")
			return nil
		}

		for _, n := range notes {
			tags := ""
			if len(n.Tags) > 0 {
				tags = fmt.Sprintf(" [%s]", strings.Join(n.Tags, ", "))
			}
			fmt.Fprintf(os.Stdout, "%-30s %s%s\n", n.Name, n.Path, tags)
		}

		return nil
	},
}
