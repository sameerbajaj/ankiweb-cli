# AnkiWeb CLI (`ankiweb`)

[![Go Reference](https://pkg.go.dev/badge/github.com/sameerbajaj/ankiweb-cli.svg)](https://pkg.go.dev/github.com/sameerbajaj/ankiweb-cli)
[![License: Apache-2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

A terminal-native Go CLI and MCP server for [AnkiWeb](https://ankiuser.net). 

Every existing Anki tool requires the desktop Anki GUI running with the AnkiConnect add-on bound to a local port. `ankiweb` talks directly to AnkiWeb's cloud services, enabling fast card creation with custom templates, batch Markdown importing, in-terminal flashcard reviews, deck triage, and community shared deck discovery — from any terminal or headless machine.

---

## ⚡ Highlights

- **Dynamic Custom Card Templates**: Add cards to any deck respecting custom notetypes with arbitrary fields (`--field "Additional Info=..."`, `--field "Source=..."`), not just basic `Front` and `Back`.
- **Markdown & Note Batch Importer (`card import`)**: Extract flashcards directly from Markdown reading notes, structured Q&A blocks, bullet pairs, and cloze syntax (`{{c1::...}}`).
- **Interactive Terminal Study (`study`)**: Review and grade due flashcards directly in the CLI with instant keyboard flips and 1–4 ease score recording (`Again`, `Hard`, `Good`, `Easy`).
- **Deck Backlog Triage (`deck triage`)**: Analyze review backlog debt ratios, detect bottleneck decks, and estimate clearance days across nested deck trees.
- **Cross-Deck Duplicate Detection (`card duplicates`)**: Identify duplicate notes and colliding prompts across multiple decks in your collection.
- **Community Deck Topic Briefing (`shared brief`)**: Surface top-rated community decks using Bayesian approval rate smoothing and audio/image media coverage filters.
- **Offline SQLite & FTS5**: Fast local caching with full-text search across your collection.
- **AI Agent Native (MCP Server)**: Ships with a built-in Model Context Protocol server (`ankiweb-mcp`) and `--agent` JSON envelopes for autonomous assistants.

---

## 📦 Installation

### From Source (Go 1.22+)

```bash
# Install the ankiweb CLI
go install github.com/sameerbajaj/ankiweb-cli/cmd/ankiweb@latest

# Optionally install the MCP server
go install github.com/sameerbajaj/ankiweb-cli/cmd/ankiweb-mcp@latest
```

### From Releases

Download precompiled binaries for macOS, Linux, and Windows from [GitHub Releases](https://github.com/sameerbajaj/ankiweb-cli/releases).

---

## 🔐 Authentication

AnkiWeb uses session cookie authentication. You can authenticate in one of two ways:

1. **Environment Variable (Recommended for CI / Headless)**:
   ```bash
   export ANKIWEB_COOKIES="ankiweb=YOUR_SESSION_COOKIE"
   ```

2. **Interactive Login**:
   ```bash
   ankiweb auth login
   ```

3. **Verify Auth Status**:
   ```bash
   ankiweb auth status
   ```

*Note: Public community shared deck searches (`ankiweb shared search`, `ankiweb shared brief`) require no authentication.*

---

## 🚀 Quick Start

### 1. Check health & environment
```bash
ankiweb doctor
```

### 2. List your decks & due review counts
```bash
ankiweb deck list
```

### 3. Add a flashcard with custom template fields
```bash
# Basic card
ankiweb card add --deck "Default" --front "What is Raft?" --back "A distributed consensus algorithm" --tags "cs"

# Custom notetype with arbitrary fields
ankiweb card add \
  --deck "Computer Science" \
  --notetype "Technical Concept" \
  --front "Byzantine Fault Tolerance" \
  --back "Safety in presence of arbitrary failing nodes" \
  --field "Additional Info=Used in Tendermint and PBFT" \
  --field "Source=Castro & Liskov 1999" \
  --tags "cs,distributed-systems"
```

### 4. Batch import flashcards from Markdown
```bash
ankiweb card import --file study-notes.md --deck "Medicine"
```
Supported Markdown patterns:
- **Heading Q&A**: `## Question` followed by answer paragraph
- **Explicit Q/A**: `Q: Question` / `A: Answer`
- **Bullet pairs**: `- Front :: Back`
- **Cloze syntax**: `Text with {{c1::cloze deletion}}`

### 5. Review flashcards in your terminal
```bash
ankiweb study --deck "Spanish" --limit 20
```

### 6. Triage deck review backlogs & pacing
```bash
ankiweb deck triage
```

### 7. Find duplicate prompts across decks
```bash
ankiweb card duplicates
```

### 8. Discover top-rated community shared decks
```bash
# Get a topic briefing with Bayesian approval ratings and audio coverage
ankiweb shared brief "Japanese"

# Search shared decks with media filters
ankiweb shared search "Spanish" --limit 20
```

---

## 📖 Command Reference

### `ankiweb card`
- `add` — Add a new note/card with support for custom fields (`--field Name=Value`)
- `get` — Inspect note details and field values by note ID
- `search` — Search your collection using Anki query syntax
- `notetypes` — List available notetypes, fields, and decks in your collection
- `import` — Batch parse and import cards from Markdown files or stdin
- `duplicates` — Detect colliding prompts and duplicate notes across decks

### `ankiweb deck`
- `list` — List all decks with new, learn, and review queue counts
- `create` — Create a new deck
- `rename` — Rename an existing deck
- `delete` — Delete a deck
- `limits` — Inspect daily new and review card study limits
- `triage` — Analyze backlog debt ratios and estimated clearance pacing

### `ankiweb study`
- `study` — Start an interactive or agent-automated review session

### `ankiweb shared`
- `search` — Search public community decks on AnkiWeb
- `info` — View detailed ratings, samples, and reviews for a shared deck
- `brief` — Generate a topic summary comparing highest-rated, freshest, and audio-rich decks

### Utility
- `doctor` — Run environment diagnostics and connectivity verification
- `sync` — Sync collection data to local SQLite database for offline fast queries
- `sql` — Run raw SQL queries against your local collection store
- `which` — Natural language command discovery

---

## 🤖 Agent & Automation Usage

Every command supports non-interactive execution with structured outputs:

- `--agent`: Enables non-interactive JSON envelopes with execution metadata.
- `--json`: Formats output as JSON.
- `--select field1,field2`: Projects only specific fields for minimal token usage.
- `--dry-run`: Previews the network mutation payload without executing it.

```bash
# Dry-run preview of a card creation
ankiweb card add --deck "Default" --front "Hello" --back "World" --dry-run

# Structured agent query
ankiweb deck triage --agent
```

### Model Context Protocol (MCP) Setup

Run the bundled MCP server in Claude Desktop, Cursor, or Codex:

```json
{
  "mcpServers": {
    "ankiweb": {
      "command": "ankiweb-mcp",
      "env": {
        "ANKIWEB_COOKIES": "ankiweb=YOUR_COOKIE"
      }
    }
  }
}
```

---

## 📄 License

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for details.
