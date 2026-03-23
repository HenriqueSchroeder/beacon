package property

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Editor reads and mutates YAML frontmatter on a single note.
type Editor struct {
	vaultPath string
}

// NewEditor creates an Editor rooted at the given vault path.
func NewEditor(vaultPath string) *Editor {
	return &Editor{vaultPath: vaultPath}
}

// Get returns the raw frontmatter value for key.
func (e *Editor) Get(notePath, key string) (any, error) {
	fullPath, err := e.resolveNotePath(notePath)
	if err != nil {
		return nil, err
	}

	doc, err := readDocument(fullPath)
	if err != nil {
		return nil, err
	}

	value, ok := doc.frontmatter[key]
	if !ok {
		return nil, fmt.Errorf("property: key %q not found", key)
	}

	return value, nil
}

// Set sets or overwrites a frontmatter key.
func (e *Editor) Set(notePath, key string, value any) error {
	fullPath, err := e.resolveNotePath(notePath)
	if err != nil {
		return err
	}

	doc, err := readDocument(fullPath)
	if err != nil {
		return err
	}

	doc.frontmatter[key] = value
	return writeDocument(fullPath, doc)
}

// Add appends a unique value to a frontmatter list key.
func (e *Editor) Add(notePath, key string, value any) error {
	fullPath, err := e.resolveNotePath(notePath)
	if err != nil {
		return err
	}

	doc, err := readDocument(fullPath)
	if err != nil {
		return err
	}

	switch existing := doc.frontmatter[key].(type) {
	case nil:
		doc.frontmatter[key] = []any{value}
	case []any:
		if !containsValue(existing, value) {
			doc.frontmatter[key] = append(existing, value)
		}
	case []string:
		if !containsValueString(existing, value) {
			doc.frontmatter[key] = append(existing, fmt.Sprint(value))
		}
	default:
		return fmt.Errorf("property: key %q is not a list", key)
	}

	return writeDocument(fullPath, doc)
}

func (e *Editor) resolveNotePath(notePath string) (string, error) {
	if filepath.Ext(notePath) != ".md" {
		return "", fmt.Errorf("property: note path must end in .md")
	}
	if filepath.IsAbs(notePath) {
		return "", fmt.Errorf("property: note path must be relative to the vault")
	}

	fullPath := filepath.Clean(filepath.Join(e.vaultPath, notePath))
	vaultRoot := filepath.Clean(e.vaultPath) + string(filepath.Separator)
	if !strings.HasPrefix(fullPath+string(filepath.Separator), vaultRoot) {
		return "", fmt.Errorf("property: note path escapes vault root")
	}

	return fullPath, nil
}

type document struct {
	frontmatter map[string]any
	body        string
	lineEnding  string
}

func readDocument(path string) (*document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("property: failed to read file: %w", err)
	}

	content := string(data)
	nl := detectLineEnding(content)
	fm, body, err := splitFrontmatter(content, nl)
	if err != nil {
		return nil, err
	}

	return &document{
		frontmatter: fm,
		body:        body,
		lineEnding:  nl,
	}, nil
}

func splitFrontmatter(content, nl string) (map[string]any, string, error) {
	if !strings.HasPrefix(content, "---"+nl) {
		return map[string]any{}, content, nil
	}

	start := len("---" + nl)
	rest := content[start:]

	closePattern := nl + "---" + nl
	idx := strings.Index(rest, closePattern)
	bodyStart := -1
	if idx >= 0 {
		bodyStart = start + idx + len(closePattern)
	} else if strings.HasSuffix(content, nl+"---") {
		idx = strings.LastIndex(rest, nl+"---")
		if idx >= 0 && start+idx+len(nl+"---") == len(content) {
			bodyStart = len(content)
		}
	}

	if idx == -1 || bodyStart == -1 {
		return nil, "", fmt.Errorf("property: invalid frontmatter block")
	}

	raw := content[start : start+idx]
	body := content[bodyStart:]

	fm := map[string]any{}
	if strings.TrimSpace(raw) != "" {
		if err := yaml.Unmarshal([]byte(raw), &fm); err != nil {
			return nil, "", fmt.Errorf("property: failed to parse frontmatter: %w", err)
		}
	}

	return fm, body, nil
}

func writeDocument(path string, doc *document) error {
	rawYAML, err := yaml.Marshal(doc.frontmatter)
	if err != nil {
		return fmt.Errorf("property: failed to marshal frontmatter: %w", err)
	}

	nl := doc.lineEnding
	if nl == "" {
		nl = "\n"
	}

	serialized := strings.ReplaceAll(string(rawYAML), "\n", nl)
	content := "---" + nl + serialized + "---" + nl + doc.body
	return atomicWrite(path, []byte(content))
}

func detectLineEnding(content string) string {
	if strings.Contains(content, "\r\n") {
		return "\r\n"
	}
	return "\n"
}

func containsValue(values []any, target any) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func containsValueString(values []string, target any) bool {
	s := fmt.Sprint(target)
	for _, value := range values {
		if value == s {
			return true
		}
	}
	return false
}

func atomicWrite(path string, data []byte) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("property: failed to stat file: %w", err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), ".beacon-property-*")
	if err != nil {
		return fmt.Errorf("property: failed to create temp file: %w", err)
	}

	tmpPath := tmp.Name()
	success := false
	defer func() {
		_ = tmp.Close()
		if !success {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		return fmt.Errorf("property: failed to write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("property: failed to close temp file: %w", err)
	}
	if err := os.Chmod(tmpPath, info.Mode().Perm()); err != nil {
		return fmt.Errorf("property: failed to set file permissions: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("property: failed to replace file: %w", err)
	}

	success = true
	return nil
}
