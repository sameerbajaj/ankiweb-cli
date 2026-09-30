// Copyright 2026 Sameer Bajaj and contributors. Licensed under Apache-2.0. See LICENSE.


package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sameerbajaj/ankiweb-cli/internal/ankiwebproto"
	"github.com/spf13/cobra"
)

func newCardAddCmd(flags *rootFlags) *cobra.Command {
	var bodyDeck string
	var bodyNotetype string
	var bodyFront string
	var bodyBack string
	var bodyTags string
	var customFields []string
	var stdinBody bool

	cmd := &cobra.Command{
		Use:         "add",
		Short:       "Add or update a note/card in your AnkiWeb collection",
		Example:     "  ankiweb card add --deck \"Default\" --front \"What is Raft?\" --back \"A consensus algorithm\" --tags cs",
		Annotations: map[string]string{"pp:endpoint": "card.add", "pp:method": "POST", "pp:path": "/svc/editor/add-or-update"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if !stdinBody {
			}
			path := "/svc/editor/add-or-update"
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			fieldsDict := map[string]string{}
			if bodyFront != "" {
				fieldsDict["Front"] = bodyFront
			}
			if bodyBack != "" {
				fieldsDict["Back"] = bodyBack
			}
			for _, f := range customFields {
				parts := strings.SplitN(f, "=", 2)
				if len(parts) == 2 {
					fieldsDict[strings.TrimSpace(parts[0])] = parts[1]
				} else {
					fieldsDict[strings.TrimSpace(parts[0])] = ""
				}
			}

			if stdinBody {
				stdinData, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("reading stdin: %w", err)
				}
				var jsonBody map[string]any
				if err := json.Unmarshal(stdinData, &jsonBody); err == nil {
					if f, ok := jsonBody["front"].(string); ok && f != "" {
						bodyFront = f
						fieldsDict["Front"] = f
					}
					if b, ok := jsonBody["back"].(string); ok && b != "" {
						bodyBack = b
						fieldsDict["Back"] = b
					}
					if d, ok := jsonBody["deck"].(string); ok && d != "" {
						bodyDeck = d
					}
					if nt, ok := jsonBody["notetype"].(string); ok && nt != "" {
						bodyNotetype = nt
					}
					if t, ok := jsonBody["tags"].(string); ok && t != "" {
						bodyTags = t
					}
					if fMap, ok := jsonBody["fields"].(map[string]any); ok {
						for k, v := range fMap {
							fieldsDict[k] = fmt.Sprint(v)
						}
					}
				}
			}

			tagsList := strings.Fields(strings.ReplaceAll(bodyTags, ",", " "))

			var data json.RawMessage
			var statusCode int
			if flags.dryRun {
				statusCode = 0
				dryData, _ := json.Marshal(map[string]any{
					"dry_run":  true,
					"deck":     bodyDeck,
					"notetype": bodyNotetype,
					"front":    bodyFront,
					"back":     bodyBack,
					"fields":   fieldsDict,
					"tags":     tagsList,
				})
				data = dryData
			} else {
				addResp, addErr := ankiwebproto.AddCard(cmd.Context(), c, ankiwebproto.AddCardRequest{
					DeckName:     bodyDeck,
					NotetypeName: bodyNotetype,
					Front:        bodyFront,
					Back:         bodyBack,
					Fields:       fieldsDict,
					Tags:         tagsList,
				})
				if addErr != nil {
					return classifyAPIError(cmd.OutOrStdout(), addErr, flags)
				}
				statusCode = 200
				respBytes, _ := json.Marshal(addResp)
				data = respBytes
			}

			// Inspect the mutate response body for a partial-failure-shaped
			// field (e.g. Google Ads `partialFailureError`).
			var partialFailure *partialFailureReport
			if !flags.dryRun && statusCode >= 200 && statusCode < 300 && (partialFailure == nil || flags.allowPartialFailure) {
				writeMutationResponseToStore(cmd.Context(), "card", data, "")
			}
			if wantsHumanTable(cmd.OutOrStdout(), flags) {
				if !flags.dryRun && statusCode == 200 {
					targetDeck := bodyDeck
					if targetDeck == "" {
						targetDeck = "Default"
					}
					targetNT := bodyNotetype
					if targetNT == "" {
						targetNT = "Basic"
					}
					fmt.Fprintf(cmd.OutOrStdout(), "✓ Successfully added card to AnkiWeb deck %q (%s)\n", targetDeck, targetNT)
					return nil
				}
			}
			if flags.asJSON || (!isTerminal(cmd.OutOrStdout()) && !flags.csv && !flags.quiet && !flags.plain) {
				if flags.quiet {
					if partialFailure != nil && !flags.allowPartialFailure {
						return partialFailureErr(fmt.Errorf("partial failure in %s response: %s", "card", partialFailure.Message))
					}
					return nil
				}
				envelope := map[string]any{
					"action":   "post",
					"resource": "card",
					"path":     path,
					"status":   statusCode,
					"success":  statusCode >= 200 && statusCode < 300 && (partialFailure == nil || flags.allowPartialFailure),
				}
				if flags.agent {
					envelope["meta"] = map[string]any{"source": "live"}
				}
				if partialFailure != nil {
					envelope["partial_failure"] = partialFailure
				}
				if flags.dryRun {
					envelope["dry_run"] = true
					envelope["status"] = 0
					envelope["success"] = false
				}
				// Verify-mode synthetic envelope detection runs against RAW data
				// (before --compact/--select filtering) so the sentinel field is
				// guaranteed to be visible even if the operator passes a filter
				// flag that would otherwise strip it. Surfaces a top-level
				// verify_noop signal + flips success to false. Mirrors the dry_run
				// shape above.
				if len(data) > 0 {
					var rawParsed any
					if err := json.Unmarshal(data, &rawParsed); err == nil {
						if m, ok := rawParsed.(map[string]any); ok {
							if v, ok := m["__ankiweb_verify_synthetic__"].(bool); ok && v {
								envelope["verify_noop"] = true
								envelope["success"] = false
							}
						}
					}
				}
				// Mutation-riding reads (POST search, RPC-over-POST lists) return
				// the same single-key collection envelopes as GET reads. Unwrap
				// before filtering so rows nest once under the result key and
				// --select filters rows, not envelope keys; plain created-object
				// responses pass through unwrapSingleKeyArray untouched.
				// Apply --compact and --select to the API response before wrapping.
				// --select wins when both are set: explicit field choice trumps the
				// generic high-gravity allow-list. Otherwise --compact still applies
				// when --agent is on but the user did not name fields.
				var selectErr error
				filtered := unwrapSingleKeyArray(data)
				if flags.selectFields != "" {
					filtered, selectErr = filterFieldsChecked(filtered, flags.selectFields)
					selectErr = selectErrorForDryRun(selectErr, flags, data)
				} else if flags.compact {
					filtered = compactFields(filtered, map[string]bool{"id": true, "notetype_id": true, "deck_id": true})
				}
				if len(filtered) > 0 {
					var parsed any
					if err := json.Unmarshal(filtered, &parsed); err == nil {
						if flags.agent {
							envelope["results"] = parsed
						} else {
							envelope["data"] = parsed
						}
					}
				}
				envelopeJSON, err := json.Marshal(envelope)
				if err != nil {
					return err
				}
				resultKey := "data"
				if flags.agent {
					resultKey = "results"
				}
				structured, err := wrapPlatformStructuredOutput(json.RawMessage(envelopeJSON), flags, resultKey, true)
				if err != nil {
					return err
				}
				if perr := printOutput(cmd.OutOrStdout(), structured, true); perr != nil {
					return perr
				}
				if partialFailure != nil && !flags.allowPartialFailure {
					return partialFailureErr(fmt.Errorf("partial failure in %s response: %s", "card", partialFailure.Message))
				}
				return selectErr
			}
			// Fall-through for mutate paths that did not hit the table or
			// asJSON branches: --quiet, --csv, --plain, and default terminal
			// raw output. printOutputWithFlagsMeta renders the body with live
			// provenance, then the typed partial-failure exit fires unless
			// --allow-partial-failure downgrades it. Without this guard a
			// partial failure would exit 0 for these output modes — the exact
			// silent-swallow regression the surrounding patch is preventing
			// for asJSON / piped output.
			printErr := printOutputWithFlagsMeta(cmd.OutOrStdout(), data, flags, map[string]any{"source": "live"}, map[string]bool{"id": true, "notetype_id": true, "deck_id": true})
			if partialFailure != nil && !flags.allowPartialFailure {
				return partialFailureErr(fmt.Errorf("partial failure in %s response: %s", "card", partialFailure.Message))
			}
			return printErr
		},
	}
	cmd.Flags().StringVar(&bodyDeck, "deck", "", "Target deck name or ID")
	cmd.Flags().StringVar(&bodyNotetype, "notetype", "", "Note type name or ID (defaults to Basic)")
	cmd.Flags().StringVar(&bodyFront, "front", "", "Front field content (or primary field)")
	cmd.Flags().StringVar(&bodyBack, "back", "", "Back field content")
	cmd.Flags().StringArrayVar(&customFields, "field", nil, "Custom field in Name=Value format (can be specified multiple times)")
	cmd.Flags().StringVar(&bodyTags, "tags", "", "Space-separated tags")
	cmd.Flags().BoolVar(&stdinBody, "stdin", false, "Read request body as JSON from stdin")

	return cmd
}
