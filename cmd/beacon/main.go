package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "beacon",
	Short: "Headless Obsidian Vault CLI",
	Long:  "Beacon — A lightweight CLI for managing Obsidian vaults on headless servers.",
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: .beacon.yml)")
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("beacon %s\n", version)
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
