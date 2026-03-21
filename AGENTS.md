# AGENTS.md

Instructions for Codex in this repository.

`CLAUDE.md` exists for Claude-specific collaboration. This file is the Codex-facing guide for working in Beacon.

## What Beacon Is

Beacon is a Go CLI for working with Obsidian vaults in headless environments. It uses Cobra for command wiring and keeps business logic in `pkg/...`.

Primary user-facing commands:

- `beacon list`
- `beacon search`
- `beacon create`
- `beacon validate`
- `beacon version`

## How Codex Should Work Here

Start by reading the code that actually implements the requested behavior. Do not infer architecture from old docs alone.

Default approach:

1. Inspect the relevant command under `cmd/beacon/`.
2. Inspect the supporting package under `pkg/...`.
3. Check the nearest tests before editing.
4. Make the smallest coherent change.
5. Run focused verification first, then broader verification if needed.

Prefer repository reality over generic patterns or stale documentation.

## Project Map

- `cmd/beacon/`: Cobra commands and CLI entrypoints
- `pkg/config/`: config loading, defaults, and validation
- `pkg/vault/`: filesystem-backed note access
- `pkg/search/`: content, tag, and type search
- `pkg/create/`: note creation workflow
- `pkg/templates/`: template loading and rendering
- `pkg/validate/`: wiki-link validation
- `pkg/links/`: wiki-link parsing
- `docs/`: architecture and feature notes
- `testdata/fixtures/`: test fixtures

Some older docs mention packages or layers that are not present anymore. Verify against the current tree before relying on them.

## Code Placement Rules

- Keep Cobra command files thin.
- Put business logic in `pkg/...`.
- Keep output formatting near the CLI layer unless the formatting is reusable domain behavior.
- Avoid introducing new abstractions unless the existing code is clearly blocking the task.

If a change affects command behavior, inspect both the command file and the package it calls.

## Editing Rules

- Prefer small, localized patches.
- Match existing naming and error-handling style.
- Wrap errors with useful context using `%w`.
- Keep interfaces narrow.
- Do not perform unrelated refactors while solving the current task.
- Preserve existing CLI output unless the task explicitly changes it.

Use `rg` and `rg --files` for searching whenever possible.

## Testing And Verification

Every behavior change should come with test coverage in the closest relevant package.

Useful commands:

```bash
go test ./...
make build
make test
make coverage
make lint
```

Verification expectations:

- run focused tests while iterating
- run `go test ./...` or `make test` before claiming a behavior change is complete
- run `make lint` when exported APIs, command wiring, or broader refactors are involved
- if verification was not run, say so explicitly

## Config And Runtime Notes

Current behavior to keep in mind:

- config is loaded via `config.LoadFrom`
- `vault_path` is required, either from config or `BEACON_VAULT_PATH`
- ripgrep is required for content search
- defaults for editor, templates directory, ignore list, and type paths live in `pkg/config/config.go`

When changing config behavior, update tests and relevant docs together.

## Documentation Rules

Treat code as the source of truth. Use `README.md`, `CLAUDE.md`, and `docs/ARCHITECTURE.md` as supporting context only.

When updating docs:

- keep command examples executable for the current repo
- avoid documenting unimplemented behavior as if it already exists
- align package references with the actual directory structure

If docs disagree with code, fix the docs.

## Git Hygiene

Before finishing:

```bash
git status --short
git diff --stat
git diff
```

Rules:

- do not commit unless the user explicitly asks
- do not push unless the user explicitly asks
- do not rewrite history unless the user explicitly asks
- do not revert unrelated user changes

## Practical Guardrails

- Prefer targeted fixes over broad rewrites.
- Do not invent missing subsystems from historical docs.
- Keep changes easy to review.
- Surface larger design issues before implementing a wide refactor.
