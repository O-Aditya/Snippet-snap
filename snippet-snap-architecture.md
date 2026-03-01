# Snippet-Snap — Go Architecture & AI System Prompt

> Stack: Go · Cobra · Bubble Tea · SQLite · fzf-style fuzzy search

---

## 1. Project Folder Structure

```
snippet-snap/
├── cmd/                        # Cobra command definitions (one file per command)
│   ├── root.go                 # Root command + global flags (--config, --db)
│   ├── add.go                  # `snap add` — CRUD Create
│   ├── list.go                 # `snap list` — CRUD Read
│   ├── rm.go                   # `snap rm <id>` — CRUD Delete
│   ├── edit.go                 # `snap edit <id>` — opens $EDITOR
│   ├── find.go                 # `snap find` — launches Bubble Tea TUI
│   └── copy.go                 # `snap copy <id>` — clipboard + {{VAR}} injection
│
├── internal/
│   ├── db/
│   │   ├── db.go               # SQLite open/close, migrations runner
│   │   ├── migrations/
│   │   │   └── 001_init.sql    # Initial schema (snippets table)
│   │   └── queries.go          # All SQL queries as typed functions (no raw SQL outside here)
│   │
│   ├── models/
│   │   └── snippet.go          # Snippet struct + validation logic
│   │
│   ├── tui/
│   │   ├── finder.go           # Bubble Tea Model for `snap find` (search + list + preview)
│   │   ├── keymap.go           # Key bindings
│   │   └── styles.go           # Lip Gloss styles (colors, borders, layout)
│   │
│   ├── clipboard/
│   │   └── clipboard.go        # Cross-platform copy (pbcopy / xclip / wl-copy / Windows)
│   │
│   ├── inject/
│   │   └── inject.go           # {{VAR}} placeholder detection + prompt loop
│   │
│   └── highlight/
│       └── highlight.go        # Chroma-based syntax highlighting for terminal
│
├── config/
│   └── config.go               # Viper config loader (~/.config/snippet-snap/config.yaml)
│
├── scripts/
│   └── install.sh              # One-liner install (curl | bash)
│
├── .goreleaser.yaml            # Single-binary cross-platform release config
├── go.mod
├── go.sum
├── Makefile                    # make build / make test / make lint
└── README.md
```

---

## 2. Architecture Layers

```
┌─────────────────────────────────────────────────────────┐
│                     CLI Layer (cmd/)                    │
│    Cobra parses args → calls internal service funcs     │
└──────────────────────────┬──────────────────────────────┘
                           │
          ┌────────────────┼────────────────┐
          ▼                ▼                ▼
   ┌─────────────┐  ┌──────────┐   ┌──────────────┐
   │  db/queries │  │  tui/    │   │  inject/     │
   │  (SQLite)   │  │  finder  │   │  clipboard/  │
   └──────┬──────┘  └────┬─────┘   └──────────────┘
          │              │
          ▼              ▼
   ┌────────────────────────────────┐
   │        models/snippet.go       │
   │  (shared Snippet{} struct)     │
   └────────────────────────────────┘
```

**Key principles:**
- `cmd/` never talks to SQLite directly — always through `internal/db/queries.go`
- `tui/` is isolated and receives `[]models.Snippet` — it does NOT query the DB itself
- All SQL lives in `internal/db/queries.go` (single source of truth)
- `{{VAR}}` injection happens in `internal/inject/inject.go`, called by both `copy.go` and the TUI confirm action

---

## 3. Data Flow per Command

### `snap add --name docker-clean --lang bash --tags "docker,cleanup"`
```
cmd/add.go
  → prompts for content (opens $EDITOR or stdin)
  → models.Snippet{}.Validate()
  → db.InsertSnippet(snippet)
  → prints confirmation with ID
```

### `snap list`
```
cmd/list.go
  → db.ListSnippets(filter?)
  → highlight.Render(snippet.Content, snippet.Language)
  → tablewriter prints formatted table to stdout
```

### `snap find` (TUI)
```
cmd/find.go
  → db.ListSnippets() → []Snippet
  → tui.NewFinder(snippets).Run()       ← Bubble Tea event loop
      ├── KeyInput → fuzzy filter in-memory
      ├── ArrowKeys → select snippet
      ├── Preview pane → highlight.Render(selected.Content)
      └── Enter → inject.ResolveVars(content) → clipboard.Copy(result)
```

### `snap copy <id>`
```
cmd/copy.go
  → db.GetSnippetByID(id)
  → inject.ResolveVars(snippet.Content)   ← prompts for {{VAR}} if present
  → clipboard.Copy(resolved)
  → prints "✓ Copied to clipboard"
```

---

## 4. Database Schema

```sql
-- internal/db/migrations/001_init.sql
CREATE TABLE IF NOT EXISTS snippets (
    id          INTEGER  PRIMARY KEY AUTOINCREMENT,
    alias       TEXT     UNIQUE NOT NULL,
    content     TEXT     NOT NULL,
    language    TEXT     DEFAULT '',
    tags        TEXT     DEFAULT '',        -- comma-separated: "docker,devops"
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- FTS virtual table for snap find performance (<100ms on 1000+ rows)
CREATE VIRTUAL TABLE IF NOT EXISTS snippets_fts
    USING fts5(alias, content, tags, content=snippets, content_rowid=id);

-- Keep FTS in sync
CREATE TRIGGER IF NOT EXISTS snippets_ai AFTER INSERT ON snippets BEGIN
    INSERT INTO snippets_fts(rowid, alias, content, tags)
    VALUES (new.id, new.alias, new.content, new.tags);
END;
```

---

## 5. Key Go Dependencies

| Package | Role |
|---|---|
| `github.com/spf13/cobra` | CLI command router |
| `github.com/charmbracelet/bubbletea` | TUI event loop |
| `github.com/charmbracelet/lipgloss` | TUI styling |
| `github.com/charmbracelet/bubbles` | Textinput, list, viewport components |
| `github.com/mattn/go-sqlite3` | SQLite driver (CGO) |
| `github.com/alecthomas/chroma/v2` | Syntax highlighting |
| `github.com/atotto/clipboard` | Cross-platform clipboard |
| `github.com/spf13/viper` | Config file management |

---

## 6. v1.0 Build Order (your ranked priority)

```
Phase 1 — Core CRUD (ship first, works without TUI)
  ✅ db.go + migrations
  ✅ models/snippet.go
  ✅ cmd/add.go
  ✅ cmd/list.go
  ✅ cmd/rm.go
  ✅ cmd/edit.go

Phase 2 — Fuzzy Search TUI
  ✅ tui/styles.go (Lip Gloss layout)
  ✅ tui/finder.go (Bubble Tea model)
  ✅ cmd/find.go

Phase 3 — Variable Placeholder Injection
  ✅ inject/inject.go  (regex: {{[A-Z_]+}})
  ✅ Wire into cmd/copy.go + TUI confirm

Phase 4 — Clipboard Copy
  ✅ clipboard/clipboard.go
  ✅ cmd/copy.go
```

---

---

# AI System Prompt for Snippet-Snap Development

> Copy this into Cursor, Claude Code, Copilot Chat, or any AI coding assistant.

---

```
You are a senior Go engineer pair-programming on "Snippet-Snap" — a CLI tool for
managing code snippets. You have full context of the project. Follow these rules strictly.

## Project Identity
- Binary name: `snap`
- Language: Go 1.22+
- CLI framework: Cobra (github.com/spf13/cobra)
- TUI framework: Bubble Tea + Lip Gloss + Bubbles (charmbracelet suite)
- DB: SQLite via mattn/go-sqlite3 (CGO enabled)
- Config: Viper (~/.config/snippet-snap/config.yaml)
- Clipboard: atotto/clipboard (cross-platform)
- Syntax highlighting: alecthomas/chroma/v2

## Folder Structure Rules
- All Cobra command definitions live in cmd/ (one file per subcommand)
- cmd/ files NEVER import database/sql directly — always call internal/db/queries.go
- All SQL queries live ONLY in internal/db/queries.go
- The Bubble Tea TUI lives in internal/tui/ — it receives []models.Snippet as input,
  never queries the DB itself
- {{VAR}} placeholder logic lives ONLY in internal/inject/inject.go

## Code Style Rules
- Use named return values only when it genuinely clarifies error paths
- Errors always bubble up with fmt.Errorf("context: %w", err) — never log.Fatal in lib code
- Use log.Fatal only in main() or cmd/ layer
- Prefer table-driven tests in _test.go files
- All exported functions must have a godoc comment
- Never use init() for side effects; use explicit constructors

## Database Rules
- Always run migrations on startup via db.RunMigrations()
- Use FTS5 virtual table (snippets_fts) for all search queries — never LIKE '%query%'
- The snippets.tags column is comma-separated plain text (no JSON)
- updated_at must be manually set on UPDATE (SQLite has no auto-update trigger without one)

## TUI (Bubble Tea) Rules
- The Finder model must implement bubbletea.Model: Init(), Update(), View()
- Fuzzy filter runs in-memory on []models.Snippet — no DB calls during keystrokes
- The preview pane uses bubbles/viewport and renders chroma-highlighted content
- Key bindings are defined in internal/tui/keymap.go using bubbles/key
- Layout: top = textinput, middle = list, bottom = viewport (preview)

## Snippet Model
type Snippet struct {
    ID        int64
    Alias     string    // unique short handle, e.g. "docker-clean"
    Content   string    // actual code/command
    Language  string    // for syntax highlighting, e.g. "bash", "python"
    Tags      string    // comma-separated, e.g. "docker,cleanup"
    CreatedAt time.Time
    UpdatedAt time.Time
}

## {{VAR}} Injection Rules
- Pattern: {{VAR_NAME}} where VAR_NAME is [A-Z][A-Z0-9_]*
- Before copying, scan content with regexp, collect unique var names
- Prompt user for each var interactively via bufio.Scanner on os.Stdin
- Replace all occurrences before writing to clipboard

## When I ask you to implement a feature:
1. State which files you will create or modify
2. Show the complete file (no truncation with "// ... rest unchanged")
3. After code, list any new go.mod dependencies needed
4. Point out any migration or config change required

## When I ask you to fix a bug:
1. Identify the root cause in one sentence
2. Show the minimal diff (not whole file unless <50 lines)
3. Explain why this fix is correct

## Things you must never do:
- Add global mutable state (no package-level vars except db handle via singleton)
- Use CGO beyond what mattn/go-sqlite3 requires
- Suggest replacing SQLite with a file-based store (JSON/YAML) — the schema is final
- Generate code that calls os.Exit() outside of cmd/ layer
- Use interface{} or any{} — always use concrete types or generics
```

---

## Quick Start Commands

```bash
# Bootstrap the project
go mod init github.com/yourname/snippet-snap
go get github.com/spf13/cobra
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
go get github.com/charmbracelet/bubbles
go get github.com/mattn/go-sqlite3
go get github.com/alecthomas/chroma/v2
go get github.com/atonto/clipboard
go get github.com/spf13/viper

# First thing to build
mkdir -p cmd internal/{db/migrations,models,tui,clipboard,inject,highlight} config
touch cmd/root.go internal/db/db.go internal/db/queries.go internal/models/snippet.go
```
