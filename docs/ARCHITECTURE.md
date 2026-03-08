# Beacon — Architecture

_Living document. Design decisions, patterns, and logic behind Beacon._

---

## 🎯 Philosophy

**Beacon is a headless CLI for Obsidian — it's not Obsidian.**

This means:
- ✅ Reads/writes directly to filesystem
- ✅ Understands frontmatter, markdown, wikilinks
- ❌ Doesn't reimplements all of Obsidian's logic
- ❌ Doesn't compete with UI

**Core Principles:**
1. **Modular** — Each package does one thing well
2. **Testable** — Everything mockable, zero required external dependencies
3. **Extensible** — Easy to add plugins, commands, storage engines
4. **Fast** — Go single binary, zero overhead

---

## 🏗️ Structure & Responsibilities

### Top Level: `cmd/beacon/main.go`

**What it does:** Pure entrypoint. Orchestrates CLI.

```go
func main() {
    cfg := config.Load()          // Loads .beacon.yml
    v := vault.New(cfg.VaultPath) // Opens vault
    
    cmd := os.Args[1]
    switch cmd {
    case "list":
        handleList(v)
    case "search":
        handleSearch(v)
    // ...
    }
}
```

**What it doesn't do:** Business logic, complex error handling.

---

### Layer 1: `pkg/vault/` — The Heart

**Responsibility:** Filesystem operations on vault.

```
pkg/vault/
├── vault.go       # Vault interface — storage abstraction
├── file.go        # File operations (read, write, list)
└── types.go       # Note, NoteMetadata, etc
```

#### `vault.go` — Main Interface

```go
type Vault interface {
    ListNotes(ctx context.Context) ([]Note, error)
    GetNote(ctx context.Context, path string) (*Note, error)
    CreateNote(ctx context.Context, name string, content string) (*Note, error)
    DeleteNote(ctx context.Context, path string) error
    Search(ctx context.Context, query string) ([]Note, error)
    Move(ctx context.Context, oldPath, newPath string) error
}

type FileVault struct {
    rootPath string
}
```

**Why interface?** Easy to mock for tests, easy to add other storage engines (cloud, sqlite, etc) later.

#### `file.go` — Concrete Implementation

```go
func (fv *FileVault) ListNotes(ctx context.Context) ([]Note, error) {
    // Walk filesystem recursively
    // Parse frontmatter of each .md
    // Return []Note
}

func (fv *FileVault) GetNote(ctx context.Context, path string) (*Note, error) {
    // Read file
    // Parse frontmatter + content
    // Return *Note
}
```

**Why separate?** Easy to test without touching filesystem (mock in `*_test.go`).

#### Struct `Note`

```go
type Note struct {
    Path         string                 // /path/to/file.md
    Name         string                 // "File Title"
    Content      string                 // Full content
    Frontmatter  map[string]interface{} // tags, type, etc
    Modified     time.Time
    Tags         []string               // Parsed from frontmatter
    Links        []string               // Found wikilinks
}

type NoteMetadata struct {
    Type   string   `yaml:"type"`      // note, daily, project, etc
    Tags   []string `yaml:"tags"`
    Status string   `yaml:"status"`    // todo, done, archived
}
```

---

### Layer 2: `pkg/search/` — Query Engine

**Responsibility:** Search notes with various filters.

```
pkg/search/
├── search.go    # Search interface
├── regex.go     # Regex implementation
├── parser.go    # Parse queries
└── index.go     # (Future) In-memory indexing
```

#### `search.go` — Interface

```go
type Searcher interface {
    SearchContent(ctx context.Context, query string) ([]Note, error)
    SearchTags(ctx context.Context, tags []string) ([]Note, error)
    SearchFrontmatter(ctx context.Context, key, value string) ([]Note, error)
    SearchByType(ctx context.Context, t string) ([]Note, error)
}

type RegexSearcher struct {
    vault Vault
}
```

#### Flow

```
User: beacon search "golang tips"
  ↓
parser.Parse("golang tips")
  ↓
[query="golang", tags=[], type=""]
  ↓
searcher.SearchContent() + SearchFrontmatter()
  ↓
[]Note (results)
```

---

### Layer 3: `pkg/git/` — Synchronization

**Responsibility:** Git operations (pull, push, commit).

```
pkg/git/
├── git.go   # GitSync interface
└── cmd.go   # Git CLI wrapper
```

#### `git.go`

```go
type GitSync interface {
    Commit(ctx context.Context, message string) error
    Push(ctx context.Context) error
    Pull(ctx context.Context) error
    Status(ctx context.Context) ([]string, error)
}

type GitClient struct {
    repoPath string
}
```

**Why wrapper?** Git CLI is source of truth. We don't reimplement Git — we delegate.

#### Usage

```go
if err := git.Commit(ctx, "feat: add new note"); err != nil {
    return fmt.Errorf("failed to commit: %w", err)
}
git.Push(ctx) // Auto-push if configured
```

---

### Layer 4: `pkg/config/` — Configuration

**Responsibility:** Load `.beacon.yml`, defaults, env vars.

```
pkg/config/
├── config.go   # Load, defaults, validation
└── types.go    # Config struct
```

#### `config.go`

```go
type Config struct {
    VaultPath string            `yaml:"vault_path"`
    Editor    string            `yaml:"editor"`
    Ignore    []string          `yaml:"ignore"`
    Git       GitConfig         `yaml:"git"`
    Templates map[string]string `yaml:"templates"`
}

type GitConfig struct {
    AutoCommit bool   `yaml:"auto_commit"`
    AutoPush   bool   `yaml:"auto_push"`
    Author     Author `yaml:"author"`
}

func Load() (*Config, error) {
    // 1. Check ~/.beacon.yml
    // 2. Check ./.beacon.yml
    // 3. Validate required fields
    // 4. Apply env var overrides
    // 5. Return Config
}
```

---

### Layer 5: `internal/cli/` — Commands

**Responsibility:** Orchestrate commands (list, search, create, etc).

```
internal/cli/
├── commands.go      # Command interface
├── list.go          # ListCommand impl
├── search.go        # SearchCommand impl
└── create.go        # CreateCommand impl
```

#### Pattern

```go
type Command interface {
    Execute(ctx context.Context) error
}

type ListCommand struct {
    vault Vault
    query string
}

func (lc *ListCommand) Execute(ctx context.Context) error {
    notes, err := lc.vault.ListNotes(ctx)
    if err != nil {
        return fmt.Errorf("list: %w", err)
    }
    // Format + print
    return nil
}
```

---

## 🔄 Data Flow

### Example: `beacon search "golang"`

```
main.go
  ↓
config.Load()  [.beacon.yml]
  ↓
vault.New(cfg.VaultPath)
  ↓
SearchCommand{vault, query="golang"}
  ↓
search.RegexSearcher.SearchContent("golang")
  ↓
vault.ListNotes()  [Reads filesystem]
  ↓
Filters by regex
  ↓
[]Note {matched results}
  ↓
Format + Print
  ↓
[STDOUT]
```

---

## 🧪 Testing Strategy

### Structure

```
pkg/vault/
├── vault.go
├── vault_test.go    ← Interface tests
├── file.go
├── file_test.go     ← Implementation tests
└── types.go

testdata/
└── fixtures/
    ├── vault/       ← Fake vault for tests
    │   ├── note1.md
    │   ├── note2.md
    │   └── subdir/note3.md
    └── config.yml
```

### Pattern: Mocks

```go
// vault_test.go
type MockVault struct {
    notes []Note
}

func (m *MockVault) ListNotes(ctx context.Context) ([]Note, error) {
    return m.notes, nil
}

// Usage in search_test.go
func TestSearchContent(t *testing.T) {
    mock := &MockVault{
        notes: []Note{
            {Name: "Go Tips", Content: "golang is fast"},
            {Name: "Rust Tips", Content: "rust is safe"},
        },
    }
    
    searcher := search.New(mock)
    results, _ := searcher.SearchContent(ctx, "golang")
    
    if len(results) != 1 {
        t.Fatalf("expected 1 result, got %d", len(results))
    }
}
```

### Coverage Targets

| Package | Target |
|---------|--------|
| `pkg/vault` | 100% |
| `pkg/search` | 95% |
| `pkg/git` | 85% |
| `pkg/config` | 90% |
| `internal/cli` | 80% |

---

## 🔌 Extensibility

### Example: Add new storage engine

```go
// pkg/vault/sql.go — New implementation
type SQLVault struct {
    db *sql.DB
}

func (sv *SQLVault) ListNotes(ctx context.Context) ([]Note, error) {
    // Query SQL
}

// Just swap in main.go:
// v := vault.New(cfg.VaultPath)  // File
// v := sqlvault.New(cfg.DBConn)  // SQL
// Interface is the same!
```

### Example: Add new command

```go
// internal/cli/backup.go
type BackupCommand struct {
    vault Vault
}

func (bc *BackupCommand) Execute(ctx context.Context) error {
    // Implements backup
}

// In main.go:
case "backup":
    cmd := cli.NewBackupCommand(v)
    return cmd.Execute(ctx)
```

---

## ⚡ Design Decisions

### Why `pkg/` instead of `internal/`?

- `pkg/` = public, reusable (Vault, Search, Git are libraries)
- `internal/` = private to project (CLI, commands are orchestration)

If someone wants to use Beacon as a lib in another project:
```go
import "github.com/HenriqueSchroeder/beacon/pkg/vault"

v := vault.New("/path/to/vault")
notes, _ := v.ListNotes(ctx)
```

### Why `Vault` interface?

1. **Testability:** Mock without filesystem
2. **Extensibility:** Multiple storage backends
3. **Inversion of Control:** Caller decides implementation

### Why `context.Context`?

1. **Cancellation:** `SIGINT` kills everything gracefully
2. **Timeout:** `context.WithTimeout()` on heavy operations
3. **Go Standard:** Expected in any modern lib

---

## 🚀 Path Forward

### Next Additions (Don't Break Architecture)

- **Indexing:** `pkg/index/` — In-memory note index (performance)
- **Plugins:** `pkg/plugins/` — Hook system for customization
- **Sync:** `pkg/sync/` — Multi-directional vault sync
- **Graph:** `pkg/graph/` — Build knowledge graph (backlinks, etc)
- **HTTP API:** `internal/api/` — Server mode (HTTP REST)

---

## 📊 Dependency Diagram

```
main.go
  ├── config/         (Load config)
  ├── vault/          (Core operations)
  │   └── types/
  ├── cli/            (Commands)
  │   ├── vault/
  │   ├── search/
  │   └── git/
  ├── search/         (Query engine)
  │   └── vault/
  └── git/            (Sync)
      └── (external: git CLI)

testdata/
  └── fixtures/       (Test vaults)
```

---

## 🎓 Applied Principles

| Principle | Implementation |
|-----------|-----------------|
| **SOLID** | S: SearchCommand does search. O: Extensible interfaces. L: Mock replaces real. I: Specific interfaces. D: Depend on interfaces, not concrete. |
| **DRY** | Logic centralized in packages. |
| **KISS** | No premature optimization. Simple first. |
| **TDD** | Tests before code. Coverage enforced. |
| **Clean Code** | Clear names, small functions, explicit error handling. |

---

_This document evolves with the project. Update as you add features._
