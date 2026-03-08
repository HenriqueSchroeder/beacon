# CLAUDE.md — Development with Claude

_Guide for working with Claude on Beacon development._

---

## 🤖 Claude's Role

Claude handles:
- ✅ Code generation & implementation
- ✅ Architecture design & refactoring
- ✅ Documentation & explanation
- ✅ Test writing & debugging logic
- ✅ Code review & suggestions

Claude doesn't (and shouldn't):
- ❌ Commit code without explicit approval
- ❌ Push to main without permission
- ❌ Make config changes without confirmation
- ❌ Delete or force-push history without asking
- ❌ Send external messages without context

---

## 📋 Development Workflow

### 1. Get Context

```bash
# Ask Claude to read ARCHITECTURE.md
# Ask Claude to check current state
git log --oneline -5
git status
```

### 2. Discuss Plan

Before implementing anything beyond small fixes:
- **Describe what you want**
- **Claude proposes approach**
- **Confirm approach is sound**
- **Then implement**

Example:
```
You: "I want to add a command to list notes by type"
Claude: "Here's the plan:
  1. Add FilterByType method to Vault interface
  2. Implement in FileVault
  3. Create TypeCommand in internal/cli
  4. Write tests for each layer"
You: "Looks good, go ahead"
Claude: *implements + tests*
```

### 3. Implement with Tests

```bash
# Claude writes test first
# Then implements code to pass test
make test          # Verify tests pass
make coverage      # Check coverage
make lint          # Verify code style
```

### 4. Review Before Push

```bash
# Show me the diff
git diff HEAD~1

# Ask Claude to explain changes
# Verify coverage targets met
# Confirm no breaking changes
```

### 5. Commit & Push

```bash
# Only after explicit approval:
git commit -m "feat: descriptive message"
git push origin <branch>
```

---

## 🧪 Testing Expectations

**Every feature MUST have tests.**

### Test Structure

```
// pkg/vault/vault_test.go

func TestListNotes(t *testing.T) {
    // Setup
    mock := &MockVault{...}
    
    // Execute
    notes, err := mock.ListNotes(ctx)
    
    // Assert
    if err != nil {
        t.Fatalf("expected nil error, got %v", err)
    }
    if len(notes) != expectedCount {
        t.Errorf("expected %d notes, got %d", expectedCount, len(notes))
    }
}
```

### Coverage Targets

| Scenario | Coverage |
|----------|----------|
| Happy path | Required |
| Error cases | Required |
| Edge cases (empty, nil, etc) | Required |
| Concurrent access | Required if relevant |

### Before Committing

```bash
make test           # All tests pass
make coverage       # Check report
# Ensure:
# - pkg/vault: 100%
# - pkg/search: 95%+
# - pkg/git: 85%+
# - pkg/config: 90%+
# - internal/cli: 80%+
```

---

## 📝 Code Style & Conventions

### Go Standards

```go
// 1. Error handling: explicit, wrapped
if err != nil {
    return fmt.Errorf("context: %w", err)
}

// 2. Interfaces over concrete types
func ProcessNotes(vault Vault) error {
    // vault is interface, easy to mock
}

// 3. Context in signatures
func (v *FileVault) ListNotes(ctx context.Context) ([]Note, error) {
    // Supports cancellation, timeout
}

// 4. Receiver names: v (vault), s (search), c (config), etc
func (v *FileVault) GetNote(...) { }
func (s *RegexSearcher) Search(...) { }

// 5. Error wrapping (don't lose context)
❌ return err
✅ return fmt.Errorf("vault: failed to read: %w", err)
```

### Naming

- **Functions:** `ListNotes()` — verb + noun
- **Interfaces:** `Vault`, `Searcher`, `GitSync` — noun form
- **Packages:** lowercase, short, `vault`, `search`, `git`
- **Tests:** `TestFunctionName()` or `TestFunctionName_EdgeCase()`

### Comments

```go
// Exported: describe what, not how
// ListNotes returns all notes in the vault
func (v *FileVault) ListNotes(ctx context.Context) ([]Note, error)

// Unexported: explain why, if non-obvious
// walkDir recursively finds markdown files
func (v *FileVault) walkDir(path string) []string

// Complex logic: explain intent
// Use regex with word boundaries to avoid partial matches
// e.g., "go" matches "golang" but "\\bgo\\b" doesn't
const queryRegex = `\b%s\b`
```

---

## 🔄 Branch Strategy

### Branch Names

```
feat/add-search-by-type
fix/handle-empty-vault
docs/improve-readme
refactor/simplify-config-parsing
test/add-git-integration-tests
chore/update-dependencies
```

### Commits

```
# Good
feat: add type filter to search
fix: prevent panic on missing frontmatter
docs: clarify architecture layers

# Bad
update code
fixed stuff
new features
wip
```

### PR/Merge Process

1. **Create feature branch**
   ```bash
   git checkout -b feat/something
   ```

2. **Develop & test**
   ```bash
   make test
   make coverage
   ```

3. **Describe changes**
   ```
   What: Add command X
   Why: Needed for Y
   How: Implemented via Z
   Tests: Coverage is 94%
   ```

4. **Merge to main**
   ```bash
   git checkout main
   git pull origin main
   git merge --no-ff feat/something
   git push origin main
   ```

---

## 🚨 What NOT to Do

### ❌ Force Pushes

Don't `git push --force` without asking first. Force rewrites history.

**Only force push if:**
- Explicitly told to fix commit author/message
- Fixing serious mistakes on non-main branch
- Even then: ask first

### ❌ Direct main Commits

Never commit directly to main. Always use feature branches:
```bash
❌ git commit -m "..." && git push origin main
✅ git checkout -b feat/X && commit && push && merge via branch
```

### ❌ Skipping Tests

No "quick fixes" without tests.

```bash
❌ "This is so small it doesn't need a test"
✅ Small or not, if it changes behavior, it gets a test
```

### ❌ Config Changes Without Confirmation

Never modify `.beacon.yml` defaults or internal config without asking.

```bash
❌ Change default vault path
✅ Ask first: "Should I update default vault path from X to Y?"
```

### ❌ External Communication

Don't send Discord/Telegram messages without context or approval.

```bash
❌ Automatically announcing "feature done"
✅ Reporting status when explicitly asked
```

---

## 💬 Communication Style

### With Henrique

- **Direct & concise** — say what was done, why, what's next
- **Ask before big changes** — architecture, dependencies, structure
- **Show diffs** — let him review before pushing
- **Explain trade-offs** — if there are multiple approaches
- **When stuck** — ask earlier, not after wasting hours

### In Code Reviews

- **Explain why** — not just what
- **Suggest, don't demand** — "Consider X because Y" vs "Do X"
- **Point to docs** — reference ARCHITECTURE.md, standards
- **Link to tests** — "See TestX for coverage"

---

## 📚 Reference

### Always Check First

Before implementing:
1. Read `ARCHITECTURE.md` — does it already fit a pattern?
2. Check existing code — similar features already done?
3. Verify tests — what's the coverage target?
4. Ask — if uncertain

### Useful Commands

```bash
# See current test coverage
make coverage

# Run specific test
go test -v -run TestListNotes ./pkg/vault

# Run with race detector (concurrent bugs)
go test -race ./...

# See git diff before committing
git diff

# Revert last commit (keep changes)
git reset --soft HEAD~1

# View specific commit
git show <hash>

# Check code since last push
git diff origin/main
```

---

## 🎯 Checklist Before Pushing

- [ ] All tests pass (`make test`)
- [ ] Coverage targets met (`make coverage`)
- [ ] Code lints clean (`make lint`)
- [ ] Commit message is descriptive
- [ ] No `// TODO` or `// FIXME` without issue
- [ ] Documentation updated if needed
- [ ] Feature branch, not main commit
- [ ] Henrique reviewed diffs

---

## 🚀 Example: Full Feature Cycle

### Scenario: Add `list by tag` command

```bash
# 1. Create branch
git checkout -b feat/list-by-tag

# 2. Write tests first (TDD)
# Create pkg/search/search_test.go with TestSearchByTag()

# 3. Implement code
# Create pkg/search/tag.go with SearchByTag() implementation

# 4. Verify
make test       # ✓ TestSearchByTag passes
make coverage   # ✓ pkg/search: 96% (>95% target)

# 5. Show diffs to Henrique
git diff origin/main

# 6. Get approval: "Looks good, merge"

# 7. Merge
git checkout main
git pull origin main
git merge --no-ff feat/list-by-tag
git push origin main

# 8. Done! Clean up
git branch -d feat/list-by-tag
```

---

_This document is the contract between Claude and human. Follow it._
