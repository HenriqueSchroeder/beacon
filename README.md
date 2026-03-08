# Beacon 🔦

A lightweight, fast CLI for managing Obsidian vaults on headless servers.

## Why Beacon?

The official Obsidian CLI requires the Obsidian app (Electron-based) to be installed. This makes it impossible to use on headless servers (Linux without GUI). **Beacon** is a standalone CLI that works directly with vault files.

### Features
- ⚡ **Fast** — Written in Go, compiles to a single binary
- 🖥️ **Headless** — Works on servers without GUI
- 🔍 **Powerful search** — Find notes by content, tags, type
- 📝 **Templates** — Create notes with automatic structure
- 🔗 **Link validation** — Check for broken backlinks
- 🎯 **Inbox management** — Organize and process inbox notes

## Installation

### From Source
```bash
git clone https://github.com/HenriqueSchroeder/beacon.git
cd beacon
make build
./bin/beacon --help
```

### From Releases
Download the latest binary from [Releases](https://github.com/HenriqueSchroeder/beacon/releases).

## Quick Start

Set your vault path:
```bash
export VAULT_PATH="/path/to/your/vault"
```

### Search notes
```bash
beacon search "golang"
```

### List inbox
```bash
beacon list inbox
```

### Create a note
```bash
beacon create "My Note"
```

## Commands

- `search <query>` — Search notes by content
- `list [inbox|tags]` — List notes
- `create <name>` — Create new note
- `edit <file>` — Open note in $EDITOR
- `sync` — Git pull/push vault
- `validate` — Check for broken links
- `version` — Show version

## Configuration

Create `.beacon.yml` in your vault root:

```yaml
vault_path: /home/user/obsidian-vault
editor: nvim
ignore_patterns:
  - "*.tmp"
  - ".obsidian/*"
```

## Development

```bash
make build    # Build binary
make test     # Run tests
make lint     # Run linter
make clean    # Clean build artifacts
```

## License

MIT

## Author

[Henrique Schroeder](https://github.com/HenriqueSchroeder)
