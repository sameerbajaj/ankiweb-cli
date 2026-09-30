# AnkiWeb CLI Developer & Agent Guide

`ankiweb` is a fast, terminal-native Go CLI and MCP server for AnkiWeb (`https://ankiuser.net`). It provides direct personal collection management, custom template card creation, Markdown import, interactive terminal study, and community shared deck intelligence.

## Local Operating Contract

Start by asking the CLI for current runtime truth:

```bash
ankiweb doctor --json
ankiweb agent-context --pretty
```

Use runtime discovery instead of relying on static lists:

```bash
ankiweb which "<capability>" --json
ankiweb <command> --help
```

Add `--agent` to any command for JSON, compact output, non-interactive defaults, and no color:

```bash
ankiweb <command> --agent
```

Before running an unfamiliar mutating command, inspect its help and run with `--dry-run`:

```bash
ankiweb <command> --help
ankiweb <command> --dry-run --agent
```

When a command requires confirmation in automated scripts, pass `--yes` explicitly.

## Command Surface

- **Card Management**:
  - `ankiweb card add --deck <name> --front <q> --back <a> [--field Name=Value...] [--tags <tags>]`
  - `ankiweb card get --note-id <id>`
  - `ankiweb card notetypes`
  - `ankiweb card search --query <query>`
  - `ankiweb card import --file <path> --deck <deck>`
  - `ankiweb card duplicates [--deck <deck>]`
- **Deck Management**:
  - `ankiweb deck list`
  - `ankiweb deck create --name <name>`
  - `ankiweb deck rename --deck-id <id> --name <new-name>`
  - `ankiweb deck delete --deck-id <id>`
  - `ankiweb deck limits --deck-id <id>`
  - `ankiweb deck triage`
- **Study & Review**:
  - `ankiweb study --deck <name> [--limit <n>]`
- **Community Shared Decks**:
  - `ankiweb shared search <query>`
  - `ankiweb shared info <id>`
  - `ankiweb shared brief <topic>`
- **Utility & Local Storage**:
  - `ankiweb doctor`
  - `ankiweb auth status` / `ankiweb auth login`
  - `ankiweb sync`
  - `ankiweb sql "<query>"`
  - `ankiweb context`
