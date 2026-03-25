package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/HenriqueSchroeder/beacon/pkg/config"
	prop "github.com/HenriqueSchroeder/beacon/pkg/property"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func init() {
	propertyCmd.AddCommand(propertyGetCmd)
	propertyCmd.AddCommand(propertySetCmd)
	propertyCmd.AddCommand(propertyAddCmd)
	propertyCmd.AddCommand(propertyRemoveCmd)
	rootCmd.AddCommand(propertyCmd)
}

var propertyCmd = &cobra.Command{
	Use:   "property",
	Short: "Read and modify note frontmatter properties",
}

var propertyGetCmd = &cobra.Command{
	Use:   "get <key> <note>",
	Short: "Get a frontmatter property from a note",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		editor, err := loadPropertyEditor()
		if err != nil {
			return err
		}

		value, err := editor.Get(args[1], args[0])
		if err != nil {
			return err
		}

		output, err := formatPropertyValue(value)
		if err != nil {
			return err
		}

		fmt.Println(output)
		return nil
	},
}

var propertySetCmd = &cobra.Command{
	Use:   "set <key> <value> <note>",
	Short: "Set a frontmatter property on a note",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		editor, err := loadPropertyEditor()
		if err != nil {
			return err
		}

		if err := editor.Set(args[2], args[0], args[1]); err != nil {
			return err
		}

		fmt.Printf("updated %s: %s\n", args[2], args[0])
		return nil
	},
}

var propertyAddCmd = &cobra.Command{
	Use:   "add <key> <value> <note>",
	Short: "Add a value to a list frontmatter property on a note",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		editor, err := loadPropertyEditor()
		if err != nil {
			return err
		}

		if err := editor.Add(args[2], args[0], args[1]); err != nil {
			return err
		}

		fmt.Printf("updated %s: %s\n", args[2], args[0])
		return nil
	},
}

var propertyRemoveCmd = &cobra.Command{
	Use:   "remove <key> <note>",
	Short: "Remove a frontmatter property from a note",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		editor, err := loadPropertyEditor()
		if err != nil {
			return err
		}

		if err := editor.Remove(args[1], args[0]); err != nil {
			return err
		}

		fmt.Printf("updated %s: %s\n", args[1], args[0])
		return nil
	},
}

func loadPropertyEditor() (*prop.Editor, error) {
	cfg, err := config.LoadFrom(cfgFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return prop.NewEditor(cfg.VaultPath), nil
}

func formatPropertyValue(value any) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case time.Time:
		if v.Hour() == 0 && v.Minute() == 0 && v.Second() == 0 && v.Nanosecond() == 0 {
			return v.Format("2006-01-02"), nil
		}
		return v.Format(time.RFC3339Nano), nil
	case nil:
		return "", nil
	default:
		data, err := yaml.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("failed to encode property value: %w", err)
		}
		return strings.TrimRight(string(data), "\n"), nil
	}
}
