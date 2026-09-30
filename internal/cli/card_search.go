// Copyright 2026 Sameer Bajaj and contributors. Licensed under Apache-2.0. See LICENSE.


package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newCardSearchCmd(flags *rootFlags) *cobra.Command {
	var bodyQuery string
	var flagCollection string
	var limit int

	cmd := &cobra.Command{
		Use:         "search [query]",
		Short:       "Search notes and cards in your Anki collection",
		Example:     "  ankiweb card search \"What is Raft?\"\n  ankiweb card search --query \"tag:dev_bot\"\n  ankiweb card search --query \"deck:Default\"",
		Annotations: map[string]string{"pp:endpoint": "card.search", "mcp:read-only": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			query := bodyQuery
			if query == "" && len(args) > 0 {
				query = strings.Join(args, " ")
			}

			colPath := FindCollectionPath(flagCollection)
			if colPath == "" {
				return fmt.Errorf("local Anki collection not found; set ANKI_COLLECTION_PATH or pass --collection <path/to/collection.anki2>")
			}

			notes, err := SearchCollection(cmd.Context(), colPath, query, limit)
			if err != nil {
				return err
			}

			if wantsHumanTable(cmd.OutOrStdout(), flags) {
				if len(notes) == 0 {
					fmt.Fprintf(cmd.OutOrStdout(), "No matching cards found for query %q in %s.\n", query, colPath)
					return nil
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Found %d matching card(s) in %s:\n\n", len(notes), colPath)
				for i, n := range notes {
					fmt.Fprintf(cmd.OutOrStdout(), "[%d] %s (ID: %d | Deck: %s)\n", i+1, n.Front, n.ID, n.Deck)
					if n.Back != "" {
						backPreview := strings.ReplaceAll(n.Back, "\n", " ")
						if len(backPreview) > 80 {
							backPreview = backPreview[:80] + "..."
						}
						fmt.Fprintf(cmd.OutOrStdout(), "    %s\n", backPreview)
					}
					if len(n.Tags) > 0 {
						fmt.Fprintf(cmd.OutOrStdout(), "    Tags: %s\n", strings.Join(n.Tags, ", "))
					}
				}
				return nil
			}

			data, err := json.Marshal(notes)
			if err != nil {
				return err
			}

			prov := attachFreshness(DataProvenance{Source: "local", Reason: "collection_anki2"}, flags)
			if flags.asJSON || (!isTerminal(cmd.OutOrStdout()) && !flags.csv && !flags.quiet && !flags.plain) {
				var selectErr error
				filtered := data
				if flags.selectFields != "" {
					filtered, selectErr = filterFieldsChecked(filtered, flags.selectFields)
					selectErr = selectErrorForDryRun(selectErr, flags, data)
				} else if flags.compact {
					filtered = compactFields(filtered, map[string]bool{"id": true, "deck": true, "front": true})
				}
				wrapped, wrapErr := wrapWithProvenance(filtered, prov)
				if wrapErr != nil {
					return wrapErr
				}
				wrapped, wrapErr = wrapPlatformStructuredOutput(wrapped, flags, "results", true)
				if wrapErr != nil {
					return wrapErr
				}
				if err := printOutput(cmd.OutOrStdout(), wrapped, true); err != nil {
					return err
				}
				return selectErr
			}

			return printOutputWithFlagsMeta(cmd.OutOrStdout(), data, flags, map[string]any{"source": "local", "path": colPath}, map[string]bool{"id": true, "deck": true, "front": true})
		},
	}
	cmd.Flags().StringVar(&bodyQuery, "query", "", "Search query (supports tag:name, deck:name, and keywords)")
	cmd.Flags().StringVar(&flagCollection, "collection", "", "Path to collection.anki2 (auto-discovered if omitted)")
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum results to return")

	return cmd
}
