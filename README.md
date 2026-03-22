<div align="center">

# 🔦 Beacon

**Give your AI agent hands inside your Obsidian vault.**

A fast, single-binary CLI that lets AI tools like Claude Code, Codex, and OpenClaw\
search, create, and validate notes — on headless servers, with zero GUI.

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go&logoColor=white)](https://go.dev)
[![Release](https://img.shields.io/github/v/release/HenriqueSchroeder/beacon?style=flat&color=blue)](https://github.com/HenriqueSchroeder/beacon/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg?style=flat)](LICENSE)
[![Tests](https://img.shields.io/github/actions/workflow/status/HenriqueSchroeder/beacon/release.yml?label=tests&style=flat)](https://github.com/HenriqueSchroeder/beacon/actions)

</div>

---

## Why Beacon?

AI coding agents run on headless servers — no GUI, no Electron, no Obsidian app. But your knowledge lives in Obsidian. Beacon bridges that gap.

Instead of giving your agent raw file access and hoping it figures out wiki-links, frontmatter, and vault structure, you give it a **purpose-built CLI** that already understands Obsidian.

```bash
# Your AI agent can search your vault in milliseconds (powered by ripgrep)
beacon search "meeting action items"

# Find every note linking to a target — understand your knowledge graph
beacon search --related "Project Atlas"

# Create structured notes from templates
beacon create "Sprint Review" --type=meetings --tags="sprint,team"

# Validate wiki-links before broken references pile up
beacon validate --strict
```

### Built for agents, useful for humans

Beacon was born from a real setup: an [OpenClaw](https://github.com/HenriqueSchroeder/openclaw) agent running on a headless Linux server, managing an Obsidian vault with thousands of notes. No display server, no Electron — just a single Go binary and `--json` output that agents parse natively.

It works just as well when **you** run it from a terminal, a cron job, or a CI pipeline.

---

## Install

**One command:**

```bash
# Linux / macOS
curl -sL https://raw.githubusercontent.com/HenriqueSchroeder/beacon/main/install.sh | sh

# Windows (PowerShell)
irm https://raw.githubusercontent.com/HenriqueSchroeder/beacon/main/install.ps1 | iex
```

<details>
<summary><b>Other methods</b></summary>

#### From releases

Download the latest binary for your platform from [GitHub Releases](https://github.com/HenriqueSchroeder/beacon/releases/latest).

```bash
# Linux (amd64)
curl -sL https://github.com/HenriqueSchroeder/beacon/releases/latest/download/beacon_linux_amd64.tar.gz | tar xz
sudo mv beacon /usr/local/bin/

# macOS (Apple Silicon)
curl -sL https://github.com/HenriqueSchroeder/beacon/releases/latest/download/beacon_darwin_arm64.tar.gz | tar xz
sudo mv beacon /usr/local/bin/

# Windows — download beacon_windows_amd64.zip from Releases and add to PATH
```

#### Via GitHub CLI

```bash
gh release download --repo HenriqueSchroeder/beacon --pattern "*linux_amd64*"
tar xzf beacon_linux_amd64.tar.gz
sudo mv beacon /usr/local/bin/
```

#### From source

```bash
git clone https://github.com/HenriqueSchroeder/beacon.git
cd beacon && make build
./bin/beacon version
```

</details>

---

## What Can It Do?

### Search — five modes, one command

| Mode | Command | What it finds |
|------|---------|---------------|
| **Content** | `beacon search "TODO"` | Full-text search across all notes (via ripgrep) |
| **Tags** | `beacon search --tags "project,idea"` | Notes matching specific frontmatter tags |
| **Type** | `beacon search --type "daily"` | Notes filtered by frontmatter type |
| **Filename** | `beacon search --filename "Project Plan"` | Quick lookup by note name |
| **Backlinks** | `beacon search --related "Calian"` | Every note that wiki-links to a target |

Add `--json` to any search for machine-readable output — pipe it to `jq`, feed it to scripts.

### Daily Notes — one command for today's note

```bash
beacon daily            # create or open today's daily note
beacon daily --yesterday
beacon daily --tomorrow
```

Idempotent — running it twice prints "found" instead of "created". Configurable date format, folder, and template via `.beacon.yml`.

### Create — templated notes in one line

```bash
beacon create "Daily Standup" --type=daily
beacon create "Bug Report" --template=bug --tags="bug,critical"
beacon create "Q2 OKRs" --type=projects --path="Active/Q2 OKRs.md"
```

Notes land in the right directory based on `type_paths` config. Templates use Go's `text/template` syntax with access to title, tags, date, and type.

### Validate — find broken links before they rot

```bash
beacon validate                          # scan entire vault
beacon validate --file "Notes/Index.md"  # check a single file
beacon validate --strict                 # exit 1 on any broken link (CI-friendly)
beacon validate --json                   # structured output for tooling
```

The validator parses `[[wiki-links]]`, resolves them against your vault, and suggests fixes using fuzzy matching when a link is close but not exact.

---

## Configuration

```yaml
# .beacon.yml
vault_path: /home/user/obsidian-vault
editor: nvim
templates_dir: "Templates"
ignore:
  - ".obsidian"
  - ".trash"
type_paths:
  daily: "Journal/Daily"
  projects: "Projects"
  meetings: "Work/Meetings"
validation:
  fuzzy_threshold: 0.8
  strict_mode: false
  ignore_patterns:
    - "^http"
daily:
  date_format: "2006-01-02"   # Go reference time format
  folder: "Daily"             # vault-relative path (defaults to type_paths["daily"])
  template: "daily"           # template name to use when creating
```

Or just set the vault path:

```bash
export BEACON_VAULT_PATH="/path/to/vault"
```

<details>
<summary><b>All config options</b></summary>

| Option | Default | Description |
|--------|---------|-------------|
| `vault_path` | *(required)* | Path to Obsidian vault |
| `editor` | `vim` | Default editor |
| `ignore` | `.obsidian` | Directories to skip |
| `templates_dir` | `700 - Recursos/Templates` | Template directory (relative to vault) |
| `type_paths` | see defaults | Map of note type → subdirectory |
| `validation.fuzzy_threshold` | `0.8` | Similarity threshold for link suggestions |
| `validation.strict_mode` | `false` | Fail on any invalid link |
| `validation.ignore_patterns` | `[]` | Regex patterns for links to ignore |
| `daily.date_format` | `2006-01-02` | Go reference time format for daily note filenames |
| `daily.folder` | `Daily` | Vault-relative folder for daily notes |
| `daily.template` | `daily` | Template name to use when creating a daily note |

</details>

---

## Use Cases

**AI agents on headless servers**\
Give Claude Code, Codex, or OpenClaw structured access to your vault — search, create, and validate notes without a GUI.

**Server-side automation**\
Cron job that creates daily notes, validates links weekly, or searches for stale TODOs.

**CI/CD pipelines**\
Run `beacon validate --strict` in CI to catch broken links before merging docs.

**Docker containers**\
Single static binary — copy it in, point it at a volume-mounted vault, done.

**Scripts and tooling**\
`--json` output on search and validate makes Beacon a building block for your own workflows.

---

## Requirements

- **ripgrep** (`rg`) — for content search only. All other commands work without it.

That's it. Beacon is a single static binary with zero runtime dependencies.

---

## Development

```bash
make build      # compile binary
make test       # run tests
make coverage   # coverage report
make lint       # golangci-lint
make clean      # remove artifacts
make install    # install to $GOPATH/bin
```

<details>
<summary><b>Project structure</b></summary>

```
beacon/
├── cmd/beacon/        # CLI commands (Cobra)
│   ├── main.go        # root command & version
│   ├── list.go        # list notes
│   ├── search.go      # multi-mode search
│   ├── create.go      # note creation
│   ├── daily.go       # daily notes
│   ├── move.go        # move & rename with backlink updates
│   └── validate.go    # link validation
├── pkg/
│   ├── config/        # YAML config loading
│   ├── vault/         # vault interface & file implementation
│   ├── search/        # ripgrep + vault-based searchers
│   ├── create/        # note creation logic
│   ├── daily/         # daily note find-or-create logic
│   ├── templates/     # template loading & rendering
│   ├── validate/      # link validation with fuzzy matching
│   ├── move/          # note mover with backlink updates
│   └── links/         # wiki-style link parser
└── testdata/fixtures/ # test fixtures
```

</details>

---

## Roadmap

- [x] Multi-mode search (content, tags, type, filename, backlinks)
- [x] Note creation with templates
- [x] Wiki-link validation with fuzzy suggestions
- [x] Daily notes with `--yesterday`/`--tomorrow`
- [x] Move & rename with automatic backlink updates
- [ ] Content manipulation (append/prepend)
- [ ] Frontmatter/property management
- [ ] Git integration & auto-sync
- [ ] Interactive TUI mode

---

## Contributing

Contributions are welcome.

1. Fork the repo
2. Create a feature branch (`git checkout -b feat/your-feature`)
3. Write tests for your changes
4. Commit with a descriptive message (`git commit -m 'feat: add feature'`)
5. Open a Pull Request against `main`

---

## License

[MIT](LICENSE) — Henrique Schroeder

---

<div align="center">

**[Install](https://github.com/HenriqueSchroeder/beacon/releases/latest)** · **[Report Bug](https://github.com/HenriqueSchroeder/beacon/issues)** · **[Request Feature](https://github.com/HenriqueSchroeder/beacon/discussions)**

</div>
