package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/HenriqueSchroeder/beacon/pkg/config"
	"github.com/HenriqueSchroeder/beacon/pkg/validate"
	"github.com/HenriqueSchroeder/beacon/pkg/vault"
	"github.com/spf13/cobra"
)

var (
	validateJSON     bool
	validateFix      bool
	validateStrict   bool
	validateUseCache bool
	validateFile     string
)

func init() {
	validateCmd.Flags().BoolVar(&validateJSON, "json", false, "output results as JSON")
	validateCmd.Flags().BoolVar(&validateFix, "fix", false, "attempt to fix broken links (WIP)")
	validateCmd.Flags().BoolVar(&validateStrict, "strict", false, "fail if any invalid links found")
	validateCmd.Flags().BoolVar(&validateUseCache, "use-cache", false, "use validation cache")
	validateCmd.Flags().StringVar(&validateFile, "file", "", "validate specific file only")
	rootCmd.AddCommand(validateCmd)
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate wiki links in vault notes",
	Long:  "Validate wiki links in vault notes. Checks if [[links]] point to valid notes and headings.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadFrom(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		v, err := vault.NewFileVault(cfg.VaultPath, cfg.Ignore)
		if err != nil {
			return fmt.Errorf("failed to open vault: %w", err)
		}

		ctx := context.Background()
		validator := validate.NewValidator(v, 4)

		// Build index of all notes and headings
		if err := validator.BuildIndex(ctx); err != nil {
			return fmt.Errorf("failed to build validation index: %w", err)
		}

		var results []validate.DocumentValidation
		var err2 error

		// Validate either a specific file or all files
		if validateFile != "" {
			results, err2 = validateSingleFile(ctx, v, validator, validateFile)
		} else {
			results, err2 = validator.ValidateAll(ctx)
		}

		if err2 != nil {
			return fmt.Errorf("validation failed: %w", err2)
		}

		// Sort results by filepath for consistent output
		sort.Slice(results, func(i, j int) bool {
			return results[i].FilePath < results[j].FilePath
		})

		if validateJSON {
			return outputJSON(results)
		}

		outputText(results)

		// Check strict mode
		if validateStrict {
			totalInvalid := 0
			for _, r := range results {
				for _, vr := range r.Results {
					if !vr.Valid {
						totalInvalid++
					}
				}
			}
			if totalInvalid > 0 {
				return fmt.Errorf("validation failed: %d invalid links found", totalInvalid)
			}
		}

		return nil
	},
}

func validateSingleFile(ctx context.Context, v vault.Vault, validator *validate.Validator, filePath string) ([]validate.DocumentValidation, error) {
	note, err := v.GetNote(ctx, filePath)
	if err != nil {
		return nil, fmt.Errorf("file not found: %s", filePath)
	}

	result := validator.ValidateDocument(ctx, note)
	return []validate.DocumentValidation{*result}, nil
}

func outputJSON(results []validate.DocumentValidation) error {
	// Transform results for JSON output
	jsonResults := make([]map[string]interface{}, 0, len(results))

	for _, docVal := range results {
		docJSON := map[string]interface{}{
			"file":        docVal.FilePath,
			"total_links": docVal.TotalLinks,
			"valid_links": docVal.ValidLinks,
			"invalid_links": docVal.TotalLinks - docVal.ValidLinks,
		}

		invalidResults := make([]map[string]interface{}, 0)
		for _, vr := range docVal.Results {
			if !vr.Valid {
				invalidResults = append(invalidResults, map[string]interface{}{
					"link":       vr.Link.Raw,
					"target":     vr.Link.Target,
					"heading":    vr.Link.Heading,
					"alias":      vr.Link.Alias,
					"line":       vr.Link.Line,
					"column":     vr.Link.Column,
					"reason":     vr.Reason,
					"suggestion": vr.Suggestion,
				})
			}
		}

		if len(invalidResults) > 0 {
			docJSON["invalid_links_details"] = invalidResults
		}

		jsonResults = append(jsonResults, docJSON)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(jsonResults)
}

func outputText(results []validate.DocumentValidation) {
	totalDocs := len(results)
	totalLinks := 0
	totalValid := 0
	totalInvalid := 0
	docsWithIssues := 0

	for _, r := range results {
		totalLinks += r.TotalLinks
		totalValid += r.ValidLinks
		invalid := r.TotalLinks - r.ValidLinks
		totalInvalid += invalid

		if invalid > 0 {
			docsWithIssues++
		}
	}

	// Summary line
	fmt.Printf("Validation Summary\n")
	fmt.Printf("===================\n")
	fmt.Printf("Total documents: %d\n", totalDocs)
	fmt.Printf("Total links: %d\n", totalLinks)
	fmt.Printf("Valid links: %d\n", totalValid)
	fmt.Printf("Invalid links: %d\n", totalInvalid)
	fmt.Printf("Documents with issues: %d\n\n", docsWithIssues)

	if totalInvalid == 0 {
		fmt.Println("✓ All links are valid!")
		return
	}

	// Detailed results
	fmt.Println("Issues by document:")
	fmt.Println("-------------------")

	for _, r := range results {
		invalid := r.TotalLinks - r.ValidLinks
		if invalid == 0 {
			continue
		}

		fmt.Printf("\n%s (%d invalid links)\n", r.FilePath, invalid)

		for _, vr := range r.Results {
			if !vr.Valid {
				fmt.Printf("  Line %d, Col %d: %s\n", vr.Link.Line, vr.Link.Column, vr.Link.Raw)
				fmt.Printf("    Error: %s\n", vr.Reason)

				if vr.Suggestion != "" {
					fmt.Printf("    Hint: %s\n", vr.Suggestion)
				}
			}
		}
	}
}

// Config validation extensions
// These would be added to the Config struct in the future
type ValidationConfig struct {
	Enabled       bool     `yaml:"enabled"`
	IgnorePatterns []string `yaml:"ignore_patterns"`
	FuzzyThreshold float64  `yaml:"fuzzy_threshold"`
	StrictMode    bool     `yaml:"strict_mode"`
}

// Example config structure for documentation
var exampleValidationConfig = `
# Validation configuration example
# Add this to .beacon.yml:

validation:
  enabled: true
  fuzzy_threshold: 0.8
  strict_mode: false
  ignore_patterns:
    - "*.tmp"
    - "draft-*"
`

func init() {
	// This init runs when the package is imported
	_ = exampleValidationConfig
}
