// Copyright 2026 Sameer Bajaj and contributors. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"modernc.org/sqlite"
)

var registerCollationOnce sync.Once

func InitAnkiCollation() {
	registerCollationOnce.Do(func() {
		_ = sqlite.RegisterCollationUtf8("unicase", func(a, b string) int {
			return strings.Compare(strings.ToLower(a), strings.ToLower(b))
		})
	})
}

// FindCollectionPath locates collection.anki2 automatically on macOS and Linux VPS.
func FindCollectionPath(explicitPath string) string {
	if explicitPath != "" {
		if _, err := os.Stat(explicitPath); err == nil {
			return explicitPath
		}
	}
	if env := os.Getenv("ANKI_COLLECTION_PATH"); env != "" {
		if _, err := os.Stat(env); err == nil {
			return env
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	candidates := []string{
		filepath.Join(home, "anki", "collection.anki2"),
		filepath.Join(home, ".local", "share", "ankiweb", "collection.anki2"),
		filepath.Join(home, "Library", "Application Support", "Anki2", "User 1", "collection.anki2"),
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}

	// Glob in Library/Application Support/Anki2/*/collection.anki2
	macGlob := filepath.Join(home, "Library", "Application Support", "Anki2", "*", "collection.anki2")
	if matches, _ := filepath.Glob(macGlob); len(matches) > 0 {
		return matches[0]
	}

	// Glob in .local/share/Anki2/*/collection.anki2
	linuxGlob := filepath.Join(home, ".local", "share", "Anki2", "*", "collection.anki2")
	if matches, _ := filepath.Glob(linuxGlob); len(matches) > 0 {
		return matches[0]
	}

	return ""
}

type CollectionNote struct {
	ID     int64    `json:"id"`
	Deck   string   `json:"deck"`
	Front  string   `json:"front"`
	Back   string   `json:"back"`
	Fields []string `json:"fields,omitempty"`
	Tags   []string `json:"tags,omitempty"`
}

// SearchCollection queries notes in collection.anki2 with tag, deck, and text filtering.
func SearchCollection(ctx context.Context, colPath string, query string, limit int) ([]CollectionNote, error) {
	InitAnkiCollation()

	dsn := fmt.Sprintf("file:%s?mode=ro&_pragma=busy_timeout(5000)", colPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening collection database: %w", err)
	}
	defer db.Close()

	if limit <= 0 {
		limit = 50
	}

	var whereClauses []string
	var args []any

	tokens := strings.Fields(query)
	var textKeywords []string

	for _, token := range tokens {
		lower := strings.ToLower(token)
		if strings.HasPrefix(lower, "tag:") {
			tagVal := strings.Trim(token[4:], `"'`)
			whereClauses = append(whereClauses, "n.tags LIKE ?")
			args = append(args, "% "+tagVal+" %")
		} else if strings.HasPrefix(lower, "deck:") {
			deckVal := strings.Trim(token[5:], `"'`)
			whereClauses = append(whereClauses, "d.name LIKE ?")
			args = append(args, "%"+deckVal+"%")
		} else {
			cleanToken := strings.Trim(token, `"'`)
			if cleanToken != "" && cleanToken != "*" {
				textKeywords = append(textKeywords, cleanToken)
			}
		}
	}

	for _, kw := range textKeywords {
		whereClauses = append(whereClauses, "(n.sfld LIKE ? OR n.flds LIKE ? OR n.tags LIKE ?)")
		pattern := "%" + kw + "%"
		args = append(args, pattern, pattern, pattern)
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	querySQL := fmt.Sprintf(`
		SELECT n.id, group_concat(DISTINCT d.name) as deck_names, n.sfld, n.flds, n.tags
		FROM notes n
		JOIN cards c ON c.nid = n.id
		JOIN decks d ON d.id = c.did
		%s
		GROUP BY n.id
		ORDER BY n.id DESC
		LIMIT ?
	`, whereSQL)
	args = append(args, limit)

	rows, err := db.QueryContext(ctx, querySQL, args...)
	if err != nil {
		return nil, fmt.Errorf("executing collection search: %w", err)
	}
	defer rows.Close()

	var results []CollectionNote
	for rows.Next() {
		var id int64
		var deckNames, sfld, flds, tags string
		if err := rows.Scan(&id, &deckNames, &sfld, &flds, &tags); err != nil {
			return nil, fmt.Errorf("scanning note row: %w", err)
		}

		fields := strings.Split(flds, "\x1f")
		front := sfld
		if len(fields) > 0 && fields[0] != "" {
			front = fields[0]
		}
		back := ""
		if len(fields) > 1 {
			back = fields[1]
		}

		tagList := strings.Fields(tags)

		results = append(results, CollectionNote{
			ID:     id,
			Deck:   deckNames,
			Front:  front,
			Back:   back,
			Fields: fields,
			Tags:   tagList,
		})
	}

	return results, nil
}
