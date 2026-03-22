package daily

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/HenriqueSchroeder/beacon/pkg/create"
	"github.com/HenriqueSchroeder/beacon/pkg/templates"
)

// Result holds the outcome of a GetOrCreate call.
type Result struct {
	Path    string
	Created bool // true if the note was newly created, false if it already existed
}

// Manager handles daily note creation and lookup.
type Manager struct {
	vaultPath  string
	folder     string
	dateFormat string
	template   string
	creator    *create.Creator
}

// NewManager creates a Manager.
// folder is the vault-relative directory for daily notes (e.g. "Daily").
// dateFormat is the Go reference time format used for the filename (e.g. "2006-01-02").
// template is the template name to use when creating a new note.
func NewManager(vaultPath, folder, dateFormat, template string, loader *templates.TemplateLoader, typePaths map[string]string) *Manager {
	return &Manager{
		vaultPath:  vaultPath,
		folder:     folder,
		dateFormat: dateFormat,
		template:   template,
		creator:    create.NewCreator(vaultPath, loader, typePaths),
	}
}

// GetOrCreate returns the daily note for the given date, creating it if it does not exist.
func (m *Manager) GetOrCreate(ctx context.Context, date time.Time) (Result, error) {
	// Sanitize title: replace path separators to prevent directory traversal via date_format.
	title := strings.ReplaceAll(date.Format(m.dateFormat), "/", "-")
	title = strings.ReplaceAll(title, string(filepath.Separator), "-")
	notePath := filepath.Join(m.vaultPath, m.folder, title+".md")

	if _, err := os.Stat(notePath); err == nil {
		return Result{Path: notePath, Created: false}, nil
	} else if !os.IsNotExist(err) {
		return Result{}, fmt.Errorf("daily: failed to stat note: %w", err)
	}

	opts := create.CreateNoteOptions{
		Title:      title,
		Template:   m.template,
		CustomPath: filepath.Join(m.folder, title+".md"),
		// Overwrite: true so concurrent calls don't get an unexpected error —
		// GetOrCreate is the existence guard; Creator should not re-block.
		Overwrite: true,
	}

	if _, err := m.creator.CreateNote(ctx, opts); err != nil {
		return Result{}, fmt.Errorf("daily: failed to create note: %w", err)
	}

	return Result{Path: notePath, Created: true}, nil
}
