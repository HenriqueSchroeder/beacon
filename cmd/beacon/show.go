package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/HenriqueSchroeder/beacon/pkg/config"
	showpkg "github.com/HenriqueSchroeder/beacon/pkg/show"
	"github.com/HenriqueSchroeder/beacon/pkg/vault"
	"github.com/spf13/cobra"
)

var showNoFrontmatter bool

func init() {
	showCmd.Flags().BoolVar(&showNoFrontmatter, "no-frontmatter", false, "hide YAML frontmatter")
	rootCmd.AddCommand(showCmd)
}

var showCmd = &cobra.Command{
	Use:   "show <note>",
	Short: "Show note content in the terminal",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadFrom(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		v, err := vault.NewFileVault(cfg.VaultPath, cfg.Ignore)
		if err != nil {
			return fmt.Errorf("failed to open vault: %w", err)
		}

		output, err := showpkg.NewViewer(v).Show(context.Background(), args[0], showpkg.Options{
			NoFrontmatter: showNoFrontmatter,
		})
		if err != nil {
			return fmt.Errorf("show failed: %w", err)
		}

		_, err = io.WriteString(os.Stdout, output.Content)
		return err
	},
}
