package daily

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/HenriqueSchroeder/beacon/pkg/templates"
)

func newTestManager(t *testing.T, vaultDir string) *Manager {
	t.Helper()
	loader := templates.NewTemplateLoader(vaultDir, "templates")
	typePaths := map[string]string{"daily": "Diário"}
	return NewManager(vaultDir, "Diário", "2006-01-02", "daily", loader, typePaths)
}

func TestGetOrCreate_CreatesNewNote(t *testing.T) {
	vaultDir := t.TempDir()
	m := newTestManager(t, vaultDir)

	date := time.Date(2026, 3, 22, 0, 0, 0, 0, time.UTC)
	result, err := m.GetOrCreate(context.Background(), date)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Created {
		t.Error("expected Created=true for new note")
	}

	expectedPath := filepath.Join(vaultDir, "Diário", "2026-03-22.md")
	if result.Path != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, result.Path)
	}

	if _, err := os.Stat(result.Path); err != nil {
		t.Errorf("file not found at %s: %v", result.Path, err)
	}
}

func TestGetOrCreate_FindsExistingNote(t *testing.T) {
	vaultDir := t.TempDir()
	m := newTestManager(t, vaultDir)

	date := time.Date(2026, 3, 22, 0, 0, 0, 0, time.UTC)

	// Create on first call
	first, err := m.GetOrCreate(context.Background(), date)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}
	if !first.Created {
		t.Error("expected Created=true on first call")
	}

	// Find on second call
	second, err := m.GetOrCreate(context.Background(), date)
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}
	if second.Created {
		t.Error("expected Created=false on second call (note already exists)")
	}
	if second.Path != first.Path {
		t.Errorf("expected same path on second call: got %s, want %s", second.Path, first.Path)
	}
}

func TestGetOrCreate_CustomDateFormat(t *testing.T) {
	vaultDir := t.TempDir()
	loader := templates.NewTemplateLoader(vaultDir, "templates")
	m := NewManager(vaultDir, "Diário", "02-01-2006", "daily", loader, map[string]string{})

	date := time.Date(2026, 3, 22, 0, 0, 0, 0, time.UTC)
	result, err := m.GetOrCreate(context.Background(), date)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedPath := filepath.Join(vaultDir, "Diário", "22-03-2026.md")
	if result.Path != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, result.Path)
	}
}

func TestGetOrCreate_YesterdayAndTomorrow(t *testing.T) {
	vaultDir := t.TempDir()
	m := newTestManager(t, vaultDir)
	ctx := context.Background()

	today := time.Date(2026, 3, 22, 12, 0, 0, 0, time.UTC)
	yesterday := today.AddDate(0, 0, -1)
	tomorrow := today.AddDate(0, 0, 1)

	for _, tc := range []struct {
		name string
		date time.Time
		want string
	}{
		{"yesterday", yesterday, "2026-03-21.md"},
		{"tomorrow", tomorrow, "2026-03-23.md"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, err := m.GetOrCreate(ctx, tc.date)
			if err != nil {
				t.Fatalf("%s: unexpected error: %v", tc.name, err)
			}
			if filepath.Base(result.Path) != tc.want {
				t.Errorf("%s: expected filename %s, got %s", tc.name, tc.want, filepath.Base(result.Path))
			}
		})
	}
}

func TestGetOrCreate_SanitizesDateFormatPathSeparators(t *testing.T) {
	vaultDir := t.TempDir()
	loader := templates.NewTemplateLoader(vaultDir, "templates")
	// date_format with "/" would create subdirectories without sanitization
	m := NewManager(vaultDir, "Daily", "02/01/2006", "daily", loader, map[string]string{})

	date := time.Date(2026, 3, 22, 0, 0, 0, 0, time.UTC)
	result, err := m.GetOrCreate(context.Background(), date)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Slashes should be replaced by dashes — no subdirectory traversal
	base := filepath.Base(result.Path)
	if base != "22-03-2026.md" {
		t.Errorf("expected sanitized filename 22-03-2026.md, got %s", base)
	}

	// Verify the file is inside the vault (no path traversal)
	if !strings.HasPrefix(filepath.Clean(result.Path), filepath.Clean(vaultDir)) {
		t.Errorf("path %s escaped vault %s", result.Path, vaultDir)
	}
}

func TestGetOrCreate_CreatesDirectoryIfMissing(t *testing.T) {
	vaultDir := t.TempDir()
	// Use a nested folder that doesn't exist yet
	loader := templates.NewTemplateLoader(vaultDir, "templates")
	m := NewManager(vaultDir, "Notas/Diário/2026", "2006-01-02", "daily", loader, map[string]string{})

	date := time.Date(2026, 3, 22, 0, 0, 0, 0, time.UTC)
	result, err := m.GetOrCreate(context.Background(), date)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Created {
		t.Error("expected Created=true")
	}
	if _, err := os.Stat(result.Path); err != nil {
		t.Errorf("file not created: %v", err)
	}
}
