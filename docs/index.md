# AnkiWeb CLI

[![Canonical Site](https://img.shields.io/badge/Canonical%20Site-sameerbajaj.com%2Ftools%2Fankiweb--cli-blue.svg)](https://sameerbajaj.com/tools/ankiweb-cli/)
[![Go Reference](https://pkg.go.dev/badge/github.com/sameerbajaj/ankiweb-cli.svg)](https://pkg.go.dev/github.com/sameerbajaj/ankiweb-cli)
[![License: Apache-2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/sameerbajaj/ankiweb-cli/blob/main/LICENSE)

A terminal-native Go command-line tool and Model Context Protocol (MCP) server for [AnkiWeb](https://ankiuser.net), built by [Sameer Bajaj](https://sameerbajaj.com).

Official canonical documentation and project hub:  
**[https://sameerbajaj.com/tools/ankiweb-cli/](https://sameerbajaj.com/tools/ankiweb-cli/)**

---

## Overview

`ankiweb` communicates directly with AnkiWeb cloud services without requiring a running desktop Anki instance or the AnkiConnect add-on. It enables:

- **Dynamic Custom Card Templates**: Add cards respecting custom notetypes with arbitrary fields (`--field "Additional Info=..."`).
- **Batch Markdown Importing**: Automatically extract flashcards from notes, Q&A blocks, and cloze syntax (`{{c1::...}}`).
- **Interactive Terminal Study**: Review and grade due flashcards directly in the terminal.
- **Deck Backlog Triage**: Analyze backlog debt ratios and calculate clearance velocity across decks.
- **Model Context Protocol (MCP)**: Run `ankiweb-mcp` to give AI agents access to flashcards and spaced repetition.

---

## Installation

### From Source

```bash
go install github.com/sameerbajaj/ankiweb-cli/cmd/ankiweb@latest
go install github.com/sameerbajaj/ankiweb-cli/cmd/ankiweb-mcp@latest
```

### Precompiled Binaries

Download precompiled binaries for macOS, Linux, and Windows from [GitHub Releases](https://github.com/sameerbajaj/ankiweb-cli/releases).

---

## Authentication

Set your AnkiWeb session cookies via environment variable:

```bash
export ANKIWEB_COOKIES="ankiweb=YOUR_SESSION_COOKIE"
```

Or run `ankiweb auth` for guided setup.

---

## Quick Reference

```bash
# Check connection and profile status
ankiweb doctor

# List decks and review backlog
ankiweb deck list
ankiweb deck triage

# Interactive terminal study session
ankiweb study "Default"

# Import flashcards from Markdown notes
ankiweb card import notes.md --deck "Computer Science"

# Start MCP server
ankiweb-mcp
```

For full guides and release details, visit [https://sameerbajaj.com/tools/ankiweb-cli/](https://sameerbajaj.com/tools/ankiweb-cli/).
