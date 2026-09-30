// Copyright 2026 Sameer Bajaj and contributors. Licensed under Apache-2.0. See LICENSE.


package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/sameerbajaj/ankiweb-cli/internal/ankiwebproto"
	"github.com/spf13/cobra"
)

func newCardNotetypesCmd(flags *rootFlags) *cobra.Command {
	var stdinBody bool

	cmd := &cobra.Command{
		Use:         "notetypes",
		Short:       "List available note types and decks in your collection",
		Example:     "  ankiweb card notetypes",
		Annotations: map[string]string{"pp:endpoint": "card.notetypes", "pp:method": "POST", "pp:path": "/svc/editor/get-info-for-adding", "mcp:read-only": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			info, err := ankiwebproto.FetchAddInfo(cmd.Context(), c)
			if err != nil {
				return classifyAPIError(cmd.OutOrStdout(), err, flags)
			}
			if wantsHumanTable(cmd.OutOrStdout(), flags) {
				fmt.Fprintln(cmd.OutOrStdout(), "NOTE TYPES:")
				for _, nt := range info.Notetypes {
					fmt.Fprintf(cmd.OutOrStdout(), "  • %-24s (ID: %d)\n", nt.Name, nt.ID)
				}
				fmt.Fprintln(cmd.OutOrStdout(), "\nDECKS:")
				for _, d := range info.Decks {
					fmt.Fprintf(cmd.OutOrStdout(), "  • %-24s (ID: %d)\n", d.Name, d.ID)
				}
				return nil
			}
			data, err := json.Marshal(info)
			if err != nil {
				return err
			}
			prov := attachFreshness(DataProvenance{Source: "live"}, flags)
			outputData := data
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
					filtered = compactFields(filtered, nil)
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
			return printOutputWithFlagsMeta(cmd.OutOrStdout(), formatData, flags, map[string]any{"source": "live"}, nil)
		},
	}
	cmd.Flags().BoolVar(&stdinBody, "stdin", false, "Read request body as JSON from stdin")

	return cmd
}
