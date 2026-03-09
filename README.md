# Beacon

A lightweight, fast CLI for managing Obsidian vaults on headless servers.

## Why Beacon?

The official Obsidian CLI requires the Obsidian app (Electron-based) to be installed. This makes it impossible to use on headless servers (Linux without GUI).

Beacon is a standalone CLI that works directly with vault files — no Electron, no GUI needed.

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

Both scripts detect your architecture automatically and install the latest version.

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

## Commands

```
beacon list                    List all notes in the vault
beacon search <query>          Search notes by content (uses ripgrep)
beacon search --tags <t1,t2>   Search notes by tags
beacon search --type <type>    Search notes by frontmatter type
beacon version                 Show version info
```

### Search flags

| Flag     | Description                        |
|----------|------------------------------------|
| `--json` | Output results as JSON             |
| `--tags` | Filter by tags (comma-separated)   |
| `--type` | Filter by frontmatter type         |

## Configuration

Create `.beacon.yml`:

```yaml
vault_path: /home/user/obsidian-vault
editor: nvim
ignore:
  - ".obsidian"
```

Or set the vault path via environment variable:

```bash
export BEACON_VAULT_PATH="/path/to/vault"
```

### Config options

| Option       | Default      | Description                  |
|--------------|--------------|------------------------------|
| `vault_path` | (required)   | Path to Obsidian vault       |
| `editor`     | `vim`        | Default editor               |
| `ignore`     | `.obsidian`  | Directories to ignore        |

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
│   └── search.go      # Search command
├── pkg/
│   ├── config/        # YAML configuration loading
│   ├── vault/         # Vault interface & FileVault implementation
│   └── search/        # Search: ripgrep + vault-based searchers
└── testdata/fixtures/ # Test fixtures
```

## Roadmap

- [ ] Git integration & auto-sync
- [ ] Note creation with templates
- [ ] Link validation (broken backlinks)
- [ ] Inbox workflows

## Contributing

Contributions welcome! Please:
1. Fork the repo
2. Create a feature branch (`git checkout -b feat/your-feature`)
3. Commit changes (`git commit -m 'feat: add feature'`)
4. Push to branch (`git push origin feat/your-feature`)
5. Open a Pull Request

## License

MIT License — see [LICENSE](LICENSE) for details.

## Author

[Henrique Schroeder](https://github.com/HenriqueSchroeder)

---

**Found a bug?** [Open an issue](https://github.com/HenriqueSchroeder/beacon/issues)

**Have an idea?** [Discussions](https://github.com/HenriqueSchroeder/beacon/discussions)
