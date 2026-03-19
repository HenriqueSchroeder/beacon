# Create Notes Feature

The `create` command allows you to create new notes in your Obsidian vault from templates with automatic variable substitution.

## Basic Usage

```bash
beacon create "My Note Title"
```

This creates a note using the default template in the root vault directory.

## Options

- `--type <type>` - Note type, determines output directory (maps to `type_paths` in config)
- `--template <name>` - Template to use (default: "default")
- `--tags <tags>` - Comma-separated tags to include
- `--path <path>` - Custom output path relative to vault root
- `--overwrite` - Overwrite existing file

## Examples

### Create a daily note with tags

```bash
beacon create "Daily Standup" --type=daily --template=daily --tags="work,standup"
```

### Create a project note with custom path

```bash
beacon create "Project X" --type=projects --path="Active/Q1 2026/Project_X.md"
```

### Create a meeting note

```bash
beacon create "Team Sync" --template=meeting --tags="team,important"
```

## Templates

### Built-in Templates

Beacon includes several built-in templates:

- **default** - Basic note template
- **daily** - Daily note with summary and to-do sections
- **project** - Project template with objectives and status
- **meeting** - Meeting notes with agenda and action items
- **template** - Generic template structure

### Custom Templates

Place custom templates in `{vault_root}/700 - Recursos/Templates/` as `.md` files:

```
/path/to/vault/
  └── 700 - Recursos/
      └── Templates/
          ├── custom.md
          └── project-2026.md
```

## Template Variables

Templates support the following variables:

- `{{title}}` - Note title
- `{{date}}` - Current date in YYYY-MM-DD format
- `{{tags}}` - Comma-separated tags
- `{{now}}` - Current timestamp with time

## Configuration

Configure type mappings in `.beacon.yml`:

```yaml
type_paths:
  daily: "100 - Diário"
  projects: "200 - Projetos"
  work: "300 - Trabalho"
  personal: "400 - Pessoal"
  resources: "700 - Recursos"
```

## Architecture

### Components

1. **TemplateLoader** (`pkg/templates/loader.go`)
   - Loads templates from vault or fallback hardcoded templates
   - Searches in `{vault_path}/700 - Recursos/Templates/`

2. **Creator** (`pkg/create/creator.go`)
   - Renders templates with variable substitution
   - Handles path resolution and file creation
   - Sanitizes filenames

3. **CLI Command** (`cmd/beacon/create.go`)
   - Exposes create functionality via CLI
   - Integrates with config system

4. **Config** (`pkg/config/config.go`)
   - Extended with `TypePaths` mapping
   - Provides defaults if not configured

### Data Flow

```
CLI Input
  ↓
Config Loading (type_paths)
  ↓
Template Loading (vault → hardcoded fallback)
  ↓
Template Rendering (variable substitution)
  ↓
Path Resolution (type → directory → filename)
  ↓
File Creation (with mkdir -p)
```

## Testing

Run tests with:

```bash
make test
```

Tests cover:

- Template loading (vault and hardcoded)
- Path resolution
- Template rendering
- File creation
- Error handling

## Error Handling

- **Template not found** - Shows available templates
- **File exists** - Requires `--overwrite` flag
- **Invalid type** - Lists available types
- **Missing title** - Returns error
- **Directory creation** - Auto-creates parent directories
