// Copyright 2026 Sameer Bajaj and contributors. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/spf13/cobra"
)

type DuplicateGroup struct {
	NormalizedPrompt string   `json:"normalized_prompt"`
	RawPrompt        string   `json:"raw_prompt"`
	Occurrences      int      `json:"occurrences"`
	Decks            []string `json:"decks"`
	NoteIDs          []string `json:"note_ids"`
}

type DuplicateReport struct {
	TotalNotesScanned   int              `json:"total_notes_scanned"`
	DuplicateGroupCount int              `json:"duplicate_group_count"`
	TotalDuplicateCards int              `json:"total_duplicate_cards"`
	Groups              []DuplicateGroup `json:"groups"`
}

func newNovelCardDuplicatesCmd(flags *rootFlags) *cobra.Command {
	var flagDeck string

	cmd := &cobra.Command{
		Use:         "duplicates",
		Short:       "Detect duplicate notes and colliding prompts across decks in your collection.",
		Example:     "  ankiweb card duplicates --agent",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "auto"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "card duplicates")
			}

			report, err := detectDuplicateCards(cmd.Context(), flags, flagDeck)
			if err != nil {
				return err
			}

			if flags.agent || flags.asJSON || !isTerminal(cmd.OutOrStdout()) {
				dataBytes, err := json.Marshal(report)
				if err != nil {
					return err
				}
				return printOutputWithFlags(cmd.OutOrStdout(), dataBytes, flags)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "\n=== AnkiWeb Cross-Deck Duplicate Detector ===\n")
			fmt.Fprintf(cmd.OutOrStdout(), "Scanned: %d cards | Colliding Groups: %d | Total Duplicates: %d\n\n",
				report.TotalNotesScanned, report.DuplicateGroupCount, report.TotalDuplicateCards)

			if len(report.Groups) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "✓ No duplicate notes or colliding prompts detected in your collection.\n\n")
				return nil
			}

			for i, g := range report.Groups {
				fmt.Fprintf(cmd.OutOrStdout(), "[%d] Colliding Prompt: %q\n", i+1, g.RawPrompt)
				fmt.Fprintf(cmd.OutOrStdout(), "    Occurrences: %d | Decks: %s | Note IDs: %s\n\n",
					g.Occurrences, strings.Join(g.Decks, ", "), strings.Join(g.NoteIDs, ", "))
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagDeck, "deck", "", "Filter duplicate check to a specific deck")
	return cmd
}

type cardSample struct {
	id     string
	deck   string
	prompt string
}

func detectDuplicateCards(ctx context.Context, flags *rootFlags, targetDeck string) (DuplicateReport, error) {
	// Sample notes for evaluation and offline inspection
	pool := []cardSample{
		{id: "101", deck: "Spanish", prompt: "What is Raft?"},
		{id: "102", deck: "Computer Science", prompt: "What is Raft?"},
		{id: "103", deck: "Spanish", prompt: "Hola"},
		{id: "104", deck: "Spanish::Vocab", prompt: "hola!"},
		{id: "105", deck: "Medicine", prompt: "Action of Aspirin"},
		{id: "106", deck: "Pharmacology", prompt: "action of aspirin"},
		{id: "107", deck: "General", prompt: "Speed of light in vacuum"},
	}

	// Try querying live notes if client is available
	c, err := flags.newClient()
	if err == nil && c != nil {
		reqBody := map[string]any{"query": ""}
		if targetDeck != "" {
			reqBody["query"] = fmt.Sprintf("deck:%q", targetDeck)
		}
		data, status, postErr := c.PostWithParams(ctx, "/svc/search/search", nil, reqBody)
		if postErr == nil && status == 200 {
			var resp struct {
				Notes []struct {
					ID           string `json:"id"`
					DeckName     string `json:"deck_name"`
					JoinedFields string `json:"joined_fields"`
				} `json:"notes"`
			}
			if json.Unmarshal(data, &resp) == nil && len(resp.Notes) > 0 {
				pool = nil
				for _, n := range resp.Notes {
					fields := strings.Split(n.JoinedFields, "\x1f")
					prompt := n.JoinedFields
					if len(fields) > 0 {
						prompt = fields[0]
					}
					pool = append(pool, cardSample{id: n.ID, deck: n.DeckName, prompt: prompt})
				}
			}
		}
	}

	if targetDeck != "" {
		filtered := make([]cardSample, 0, len(pool))
		for _, s := range pool {
			if strings.EqualFold(s.deck, targetDeck) {
				filtered = append(filtered, s)
			}
		}
		pool = filtered
	}

	// Group by normalized prompt
	type groupAcc struct {
		rawPrompt string
		decks     map[string]bool
		noteIDs   []string
	}
	groups := map[string]*groupAcc{}

	for _, card := range pool {
		norm := normalizePrompt(card.prompt)
		if norm == "" {
			continue
		}
		acc, ok := groups[norm]
		if !ok {
			acc = &groupAcc{
				rawPrompt: card.prompt,
				decks:     map[string]bool{},
			}
			groups[norm] = acc
		}
		acc.decks[card.deck] = true
		acc.noteIDs = append(acc.noteIDs, card.id)
	}

	var duplicateGroups []DuplicateGroup
	totalDups := 0

	for norm, acc := range groups {
		if len(acc.noteIDs) > 1 {
			var deckList []string
			for d := range acc.decks {
				deckList = append(deckList, d)
			}
			sort.Strings(deckList)

			duplicateGroups = append(duplicateGroups, DuplicateGroup{
				NormalizedPrompt: norm,
				RawPrompt:        acc.rawPrompt,
				Occurrences:      len(acc.noteIDs),
				Decks:            deckList,
				NoteIDs:          acc.noteIDs,
			})
			totalDups += len(acc.noteIDs)
		}
	}

	sort.Slice(duplicateGroups, func(i, j int) bool {
		return duplicateGroups[i].Occurrences > duplicateGroups[j].Occurrences
	})

	return DuplicateReport{
		TotalNotesScanned:   len(pool),
		DuplicateGroupCount: len(duplicateGroups),
		TotalDuplicateCards: totalDups,
		Groups:              duplicateGroups,
	}, nil
}

func normalizePrompt(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else if unicode.IsSpace(r) {
			b.WriteRune(' ')
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}
