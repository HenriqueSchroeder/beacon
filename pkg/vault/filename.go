package vault

import "strings"

// SanitizeFilename converts a title to a filesystem-safe markdown basename.
func SanitizeFilename(title string) string {
	name := strings.ReplaceAll(title, " ", "_")
	replacer := strings.NewReplacer(
		"/", "-",
		"\\", "-",
		":", "-",
		"*", "-",
		"?", "-",
		"\"", "-",
		"<", "-",
		">", "-",
		"|", "-",
	)
	return replacer.Replace(name)
}
