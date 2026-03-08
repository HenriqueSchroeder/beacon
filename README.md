# Beacon 🔦

A lightweight, fast CLI for managing Obsidian vaults on headless servers.

## Why Beacon?

The official Obsidian CLI requires the Obsidian app (Electron-based) to be installed. This makes it **impossible to use on headless servers** (Linux without GUI).

**Beacon** is a standalone CLI that works directly with vault files — **no Electron, no GUI needed**.

Perfect for:
- 🖥️ Server-side vault automation
- 🔧 CI/CD pipelines
- 🐋 Docker containers
- ⚙️ Scheduled tasks & cron jobs

## Features

- ⚡ **Fast** — Single Go binary, instant startup
- 🖥️ **Headless** — Works on any Linux server
- 🔍 **Powerful search** — Find notes by content, tags, frontmatter
- 📝 **Smart templates** — Create notes with automatic structure
- 🔗 **Link validation** — Detect broken backlinks
- 🎯 **Inbox workflows** — Organize & process inbox efficiently
- ⚙️ **Git integration** — Auto commit & push changes

## Installation

### Quick Install (Homebrew — coming soon)
```bash
brew install henrique/tap/beacon
```

### From Source
```bash
git clone https://github.com/HenriqueSchroeder/beacon.git
cd beacon
make build
./bin/beacon --version
```

### From Releases
Download pre-built binaries from [Releases](https://github.com/HenriqueSchroeder/beacon/releases).

## Quick Start

### Setup
```bash
# Export your vault path (or set in ~/.beacon.yml)
export VAULT_PATH="/home/user/obsidian-vault"
```

### Examples
```bash
# Search notes
beacon search "golang tips"

# List inbox (unprocessed notes)
beacon list inbox

# Create a note
beacon create "My new idea"

# Show all available tags
beacon list tags

# Validate broken links
beacon validate

# Sync vault (git pull + push)
beacon sync
```

## Commands

```
beacon search <query>      Search notes by content
beacon list [type]         List notes (inbox, tags, all)
beacon create <name>       Create new note with template
beacon edit <file>         Open note in $EDITOR
beacon sync                Git pull & push vault
beacon validate            Check for broken backlinks
beacon config              Show/edit configuration
beacon version             Show version info
```

## Configuration

Create `~/.beacon.yml` or `.beacon.yml` in your vault:

```yaml
# Vault configuration
vault_path: /home/user/obsidian-vault
editor: nvim

# Ignore patterns
ignore:
  - "*.tmp"
  - ".obsidian/*"
  - "node_modules/*"

# Git settings
git:
  auto_commit: true
  auto_push: true
  author:
    name: "Your Name"
    email: "your@email.com"

# Note templates
templates:
  default: "templates/default.md"
  daily: "templates/daily.md"
```

## Examples

### Daily Workflow
```bash
# Check inbox every morning
beacon list inbox

# Create daily note
beacon create "$(date +%Y-%m-%d)"

# Sync before sleeping
beacon sync
```

### Automation (Cron)
```bash
# Every 2 hours: validate links
0 */2 * * * beacon validate

# Every day at 9 AM: commit changes
0 9 * * * beacon sync
```

## Development

```bash
make build    # Compile binary
make test     # Run tests
make lint     # Run linter (requires golangci-lint)
make clean    # Remove artifacts
make install  # Install locally
```

### Project Structure
```
beacon/
├── cmd/beacon/          # CLI entry point
├── internal/vault/      # Vault operations
├── internal/search/     # Search engine
├── internal/git/        # Git integration
└── tests/               # Test suite
```

## Roadmap

- [ ] v0.2 — Full search implementation
- [ ] v0.3 — Git integration & auto-sync
- [ ] v0.4 — Link validation
- [ ] v0.5 — Configuration system
- [ ] v1.0 — Stable release

## Contributing

Contributions welcome! Please:
1. Fork the repo
2. Create a feature branch (`git checkout -b feature/your-feature`)
3. Commit changes (`git commit -am 'Add feature'`)
4. Push to branch (`git push origin feature/your-feature`)
5. Open a Pull Request

## License

MIT License — see [LICENSE](LICENSE) for details.

## Author

[Henrique Schroeder](https://github.com/HenriqueSchroeder)

---

**Found a bug?** [Open an issue](https://github.com/HenriqueSchroeder/beacon/issues)

**Have an idea?** [Discussions](https://github.com/HenriqueSchroeder/beacon/discussions)
