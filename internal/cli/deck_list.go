// Copyright 2026 Sameer Bajaj and contributors. Licensed under Apache-2.0. See LICENSE.


package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/sameerbajaj/ankiweb-cli/internal/ankiwebproto"
	"github.com/spf13/cobra"
)

func newDeckListCmd(flags *rootFlags) *cobra.Command {
	var stdinBody bool

	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List all decks in your AnkiWeb collection",
		Example:     "  ankiweb deck list",
		Annotations: map[string]string{"pp:endpoint": "deck.list", "pp:method": "POST", "pp:path": "/svc/editor/get-info-for-adding", "mcp:read-only": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			info, err := ankiwebproto.FetchAddInfo(cmd.Context(), c)
			if err != nil {
				return classifyAPIError(cmd.OutOrStdout(), err, flags)
			}
			data, err := json.Marshal(info.Decks)
			if err != nil {
				return err
			}
			prov := attachFreshness(DataProvenance{Source: "live"}, flags)
			outputData := data
			if wantsHumanTable(cmd.OutOrStdout(), flags) {
				printProvenance(cmd, len(info.Decks), prov)
			}
			// For JSON output, wrap with provenance envelope before passing through flags.
			// --select wins over --compact when both are set; --compact only runs when
			// no explicit fields were requested. Explicit format flags (--csv, --quiet,
			// --plain) opt out of the auto-JSON path so piped consumers that asked for
			// a non-JSON format reach the standard pipeline below.
			if flags.asJSON || (!isTerminal(cmd.OutOrStdout()) && !flags.csv && !flags.quiet && !flags.plain) {
				var selectErr error
				filtered := data
				if flags.selectFields != "" {
					filtered, selectErr = filterFieldsChecked(filtered, flags.selectFields)
					selectErr = selectErrorForDryRun(selectErr, flags, data)
				} else if flags.compact {
					filtered = compactFields(filtered, map[string]bool{"id": true, "name": true, "new_count": true, "learn_count": true, "review_count": true})
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
			// For all other output modes (table, csv, plain, quiet), use the standard pipeline
			if wantsHumanTable(cmd.OutOrStdout(), flags) {
				var items []map[string]any
				if json.Unmarshal(outputData, &items) == nil && len(items) > 0 {
					if err := printAutoTable(cmd.OutOrStdout(), items); err != nil {
						return err
					}
					if len(items) >= 25 {
						fmt.Fprintf(os.Stderr, "\nShowing %d results. To narrow: add --limit, --json --select, or filter flags.\n", len(items))
					}
					return nil
				}
			}
			formatData := data
			if flags.csv || flags.plain {
				formatData = outputData
			}
			return printOutputWithFlagsMeta(cmd.OutOrStdout(), formatData, flags, map[string]any{"source": "live"}, map[string]bool{"id": true, "name": true, "new_count": true, "learn_count": true, "review_count": true})
		},
	}
	cmd.Flags().BoolVar(&stdinBody, "stdin", false, "Read request body as JSON from stdin")

	return cmd
}
