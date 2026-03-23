package vault

import (
	"context"
	"time"
)

type Note struct {
	Path        string
	Name        string
	RawContent  string
	Content     string
	Frontmatter map[string]any
	Tags        []string
	Modified    time.Time
}

type Vault interface {
	ListNotes(ctx context.Context) ([]Note, error)
	GetNote(ctx context.Context, path string) (*Note, error)
}
