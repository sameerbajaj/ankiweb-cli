# Changelog

All notable changes to `ankiweb` will be documented in this file.

## [0.1.0] - 2026-09-30

### Features
- Direct cloud API integration with AnkiWeb (`https://ankiuser.net`)
- Terminal-native card creation with dynamic custom fields (`--field Name=Value`, `--front`, `--back`, `--deck`, `--tags`)
- Batch Markdown flashcard importer (`card import`) supporting Q&A blocks, headers, bullet pairs, and cloze syntax
- Interactive in-terminal flashcard study and grading (`study`)
- Collection deck triage and review debt backlog pacing calculator (`deck triage`)
- Cross-deck duplicate card and colliding prompt detector (`card duplicates`)
- Community shared deck topic briefing with Bayesian approval ratings (`shared brief`)
- Fast local SQLite database with full-text search (FTS5) and offline caching
- MCP server (`ankiweb-mcp`) providing tools for AI coding agents
