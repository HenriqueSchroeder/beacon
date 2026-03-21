# Beacon

A lightweight, fast CLI for managing Obsidian vaults on headless servers, with performance as a first-class concern for large vaults.

## Why Beacon?

The official Obsidian CLI requires the Obsidian app (Electron-based) to be installed. This makes it impossible to use on headless servers (Linux without GUI).

Beacon is a standalone CLI that works directly with vault files — no Electron, no GUI needed.

The project optimizes for real Obsidian vaults that can grow large over time. New features should preserve that bias toward predictable, efficient behavior.

Perfect for:
- Server-side vault automation
- CI/CD pipelines
- Docker containers
- Scheduled tasks & cron jobs

## Installation

### Quick Install (Linux/macOS)

```bash
curl -sL https://raw.githubusercontent.com/HenriqueSchroeder/beacon/main/install.sh | sh
```

### Quick Install (Windows PowerShell)

```powershell
irm https://raw.githubusercontent.com/HenriqueSchroeder/beacon/main/install.ps1 | iex
```

Both scripts detect your architecture automatically and install the latest version. To update, run the same command again.

### From Source

```bash
git clone https://github.com/HenriqueSchroeder/beacon.git
cd beacon
make build
./bin/beacon version
```

### From Releases

Download the latest binary for your platform from [GitHub Releases](https://github.com/HenriqueSchroeder/beacon/releases/latest).

#### Linux (amd64)

```bash
curl -sL https://github.com/HenriqueSchroeder/beacon/releases/latest/download/beacon_linux_amd64.tar.gz | tar xz
sudo mv beacon /usr/local/bin/
```

#### macOS (Apple Silicon)

```bash
curl -sL https://github.com/HenriqueSchroeder/beacon/releases/latest/download/beacon_darwin_arm64.tar.gz | tar xz
sudo mv beacon /usr/local/bin/
```

#### Windows

Download `beacon_windows_amd64.zip` from [Releases](https://github.com/HenriqueSchroeder/beacon/releases/latest) and add to your PATH.

#### Via GitHub CLI

```bash
gh release download --repo HenriqueSchroeder/beacon --pattern "*linux_amd64*"
tar xzf beacon_linux_amd64.tar.gz
sudo mv beacon /usr/local/bin/
```

## Quick Start

```bash
# Set your vault path
export BEACON_VAULT_PATH="/path/to/your/obsidian-vault"

# Or use a config file
beacon --config .beacon.yml list
```

### List notes

```bash
beacon list
```

### Search by content (requires ripgrep)

```bash
beacon search "golang tips"
beacon search "TODO" --json
```

### Search by tags

```bash
beacon search --tags "project,idea"
```

### Search by note type

```bash
beacon search --type "daily"
```

### Search by filename

```bash
beacon search --filename "Project Plan"
```

### Create a note from template

```bash
beacon create "My Note" --type=daily
beacon create "Meeting Notes" --tags="work,urgent" --template=meeting
beacon create "Project X" --type=projects --path="Active/Project X.md"
```

### Validate wiki links

```bash
beacon validate
beacon validate --file "Notes/My Note.md"
beacon validate --json
beacon validate --strict   # exit 1 if any broken link found
```

## Commands

```
beacon list                    List all notes in the vault
beacon search <query>          Search notes by content (uses ripgrep)
beacon search --filename <q>   Search notes by filename
beacon search --tags <t1,t2>   Search notes by tags
beacon search --type <type>    Search notes by frontmatter type
beacon create <title>          Create a new note from a template
beacon validate                Validate wiki links in vault notes
beacon version                 Show version info
```

### Search flags

| Flag     | Description                        |
|----------|------------------------------------|
| `--json` | Output results as JSON             |
| `--filename` | Search by filename basename; query is normalized like `create` |
| `--tags` | Filter by tags (comma-separated)   |
| `--type` | Filter by frontmatter type         |

### Create flags

| Flag          | Description                                      |
|---------------|--------------------------------------------------|
| `--type`      | Note type (determines output directory)          |
| `--template`  | Template name to use (default: default)          |
| `--tags`      | Tags to include (comma-separated)                |
| `--path`      | Custom output path (relative to vault root)      |
| `--overwrite` | Overwrite existing file                          |

### Validate flags

| Flag          | Description                                      |
|---------------|--------------------------------------------------|
| `--json`      | Output results as JSON                           |
| `--strict`    | Exit with error if any invalid links found       |
| `--file`      | Validate a specific file only                    |

## Configuration

Create `.beacon.yml`:

```yaml
vault_path: /home/user/obsidian-vault
editor: nvim
templates_dir: "700 - Recursos/Templates"
ignore:
  - ".obsidian"
type_paths:
  daily: "100 - Diário"
  projects: "200 - Projetos"
  work: "300 - Trabalho"
validation:
  fuzzy_threshold: 0.8
  strict_mode: false
  ignore_patterns:
    - "^http"
```

Or set the vault path via environment variable:

```bash
export BEACON_VAULT_PATH="/path/to/vault"
```

### Config options

| Option                          | Default                          | Description                                  |
|---------------------------------|----------------------------------|----------------------------------------------|
| `vault_path`                    | (required)                       | Path to Obsidian vault                       |
| `editor`                        | `vim`                            | Default editor                               |
| `ignore`                        | `.obsidian`                      | Directories to ignore                        |
| `templates_dir`                 | `700 - Recursos/Templates`       | Directory with note templates (in vault)     |
| `type_paths`                    | see defaults                     | Map of note type → subdirectory              |
| `validation.fuzzy_threshold`    | `0.8`                            | Similarity threshold for link suggestions    |
| `validation.strict_mode`        | `false`                          | Fail if any invalid links found              |
| `validation.ignore_patterns`    | `[]`                             | Regex patterns for links to skip             |

## Dependencies

- **Go 1.21+** for building
- **ripgrep** (`rg`) for content search

## Development

```bash
make build      # Compile binary
make test       # Run tests
make coverage   # Test coverage report
make lint       # Run linter (requires golangci-lint)
make clean      # Remove artifacts
make install    # Install locally
```

### Project Structure

```
beacon/
├── cmd/beacon/        # CLI entry point (Cobra commands)
│   ├── main.go        # Root command & version
│   ├── list.go        # List command
│   ├── search.go      # Search command
│   ├── create.go      # Create command
│   └── validate.go    # Validate command
├── pkg/
│   ├── config/        # YAML configuration loading
│   ├── vault/         # Vault interface & FileVault implementation
│   ├── search/        # Search: ripgrep + vault-based searchers
│   ├── create/        # Note creation logic
│   ├── templates/     # Template loading and rendering
│   ├── validate/      # Link validation with fuzzy matching
│   └── links/         # Wiki-style link parser
└── testdata/fixtures/ # Test fixtures
```

## Roadmap

- [ ] Git integration & auto-sync
- [x] Note creation with templates
- [x] Link validation (broken backlinks)
- [ ] Inbox workflows

## Contributing

Contributions welcome! Please:
1. Fork the repo
2. Create a feature branch (`git checkout -b feat/your-feature`)
3. Commit changes (`git commit -m 'feat: add feature'`)
4. Push to branch (`git push origin feat/your-feature`)
5. Open a Pull Request

All changes intended for `main` should go through a Pull Request. Do not push or merge directly into `main`.

## License

MIT License — see [LICENSE](LICENSE) for details.

## Author

[Henrique Schroeder](https://github.com/HenriqueSchroeder)

---

**Found a bug?** [Open an issue](https://github.com/HenriqueSchroeder/beacon/issues)

**Have an idea?** [Discussions](https://github.com/HenriqueSchroeder/beacon/discussions)
