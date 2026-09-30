// Copyright 2026 Sameer Bajaj and contributors. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/sameerbajaj/ankiweb-cli/internal/ankiwebproto"
	"github.com/spf13/cobra"
)

type ParsedCard struct {
	Front    string            `json:"front"`
	Back     string            `json:"back"`
	Notetype string            `json:"notetype,omitempty"`
	Tags     []string          `json:"tags,omitempty"`
	Fields   map[string]string `json:"fields,omitempty"`
}

type CardImportResult struct {
	DeckName      string       `json:"deck"`
	TotalParsed   int          `json:"total_parsed"`
	TotalImported int          `json:"total_imported"`
	Cards         []ParsedCard `json:"cards"`
}

func newNovelCardImportCmd(flags *rootFlags) *cobra.Command {
	var flagFile string
	var flagDeck string
	var flagNotetype string
	var flagTags string

	cmd := &cobra.Command{
		Use:         "import",
		Short:       "Batch import flashcards from Markdown files, Q&A blocks, and cloze notes into any deck.",
		Example:     "  ankiweb card import --file notes.md --deck \"Medicine\" --agent",
		Annotations: map[string]string{"mcp:read-only": "false", "pp:data-source": "auto"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "card import")
			}

			var content []byte
			var err error

			if flagFile == "" || flagFile == "-" {
				if isTerminal(os.Stdin) && flagFile == "" {
					return usageErr(fmt.Errorf("missing required --file flag (or pass '-' to read from stdin)"))
				}
				content, err = io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("reading stdin: %w", err)
				}
			} else {
				content, err = os.ReadFile(flagFile)
				if err != nil {
					return fmt.Errorf("reading file %s: %w", flagFile, err)
				}
			}

			deckName := flagDeck
			if deckName == "" {
				deckName = "Default"
			}
			notetype := flagNotetype
			if notetype == "" {
				notetype = "Basic"
			}

			parsedCards := parseMarkdownToCards(string(content), notetype, flagTags)
			if len(parsedCards) == 0 {
				return fmt.Errorf("no flashcards found in input (supported formats: '## Question\\nAnswer', 'Q: ...\\nA: ...', '- Front :: Back', or '{{c1::cloze}}')")
			}

			result := CardImportResult{
				DeckName:      deckName,
				TotalParsed:   len(parsedCards),
				TotalImported: len(parsedCards),
				Cards:         parsedCards,
			}

			// In real mode, post each card to AnkiWeb via client if authenticated
			if !flags.dryRun {
				c, clientErr := flags.newClient()
				if clientErr != nil {
					return clientErr
				}
				importedCount := 0
				for _, card := range parsedCards {
					addReq := ankiwebproto.AddCardRequest{
						DeckName:     deckName,
						NotetypeName: card.Notetype,
						Front:        card.Front,
						Back:         card.Back,
						Fields:       card.Fields,
						Tags:         card.Tags,
					}
					if _, err := ankiwebproto.AddCard(cmd.Context(), c, addReq); err != nil {
						return fmt.Errorf("failed adding card %q: %w", card.Front, err)
					}
					importedCount++
				}
				result.TotalImported = importedCount
			}

			if flags.agent || flags.asJSON || !isTerminal(cmd.OutOrStdout()) {
				dataBytes, err := json.Marshal(result)
				if err != nil {
					return err
				}
				return printOutputWithFlags(cmd.OutOrStdout(), dataBytes, flags)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Successfully parsed and imported %d cards into deck %q!\n", len(parsedCards), deckName)
			for i, card := range parsedCards {
				fmt.Fprintf(cmd.OutOrStdout(), "  [%d] %s -> %s\n", i+1, card.Front, card.Back)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagFile, "file", "", "Markdown file path to import (or '-' for stdin)")
	cmd.Flags().StringVar(&flagDeck, "deck", "Default", "Target deck name or ID")
	cmd.Flags().StringVar(&flagNotetype, "notetype", "Basic", "Default notetype for imported cards")
	cmd.Flags().StringVar(&flagTags, "tags", "", "Default tags to apply to all imported cards")
	return cmd
}

func parseMarkdownToCards(text, defaultNotetype, defaultTags string) []ParsedCard {
	var cards []ParsedCard
	tagsList := []string{}
	if defaultTags != "" {
		for _, t := range strings.Fields(defaultTags) {
			tagsList = append(tagsList, strings.TrimPrefix(t, "#"))
		}
	}

	scanner := bufio.NewScanner(strings.NewReader(text))
	var currentFront string
	var currentBack strings.Builder
	inHeadingBlock := false

	clozeRe := regexp.MustCompile(`\{\{c\d+::([^}]+)\}\}`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			if inHeadingBlock && currentFront != "" {
				back := strings.TrimSpace(currentBack.String())
				if back != "" {
					cards = append(cards, makeParsedCard(currentFront, back, defaultNotetype, tagsList))
				}
				currentFront = ""
				currentBack.Reset()
				inHeadingBlock = false
			}
			continue
		}

		// Check for inline Q: and A:
		if strings.HasPrefix(strings.ToUpper(line), "Q:") {
			currentFront = strings.TrimSpace(line[2:])
			continue
		}
		if currentFront != "" && strings.HasPrefix(strings.ToUpper(line), "A:") {
			back := strings.TrimSpace(line[2:])
			cards = append(cards, makeParsedCard(currentFront, back, defaultNotetype, tagsList))
			currentFront = ""
			continue
		}

		// Check for bullet separator: - Front :: Back or - Front : Back
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			bullet := strings.TrimSpace(line[2:])
			if idx := strings.Index(bullet, "::"); idx != -1 {
				front := strings.TrimSpace(bullet[:idx])
				back := strings.TrimSpace(bullet[idx+2:])
				if front != "" && back != "" {
					cards = append(cards, makeParsedCard(front, back, defaultNotetype, tagsList))
					continue
				}
			}
		}

		// Check for Markdown headings: ## Question
		if strings.HasPrefix(line, "#") {
			if inHeadingBlock && currentFront != "" {
				back := strings.TrimSpace(currentBack.String())
				if back != "" {
					cards = append(cards, makeParsedCard(currentFront, back, defaultNotetype, tagsList))
				}
			}
			currentFront = strings.TrimSpace(strings.TrimLeft(line, "#"))
			currentBack.Reset()
			inHeadingBlock = true
			continue
		}

		if inHeadingBlock {
			currentBack.WriteString(line + "\n")
			continue
		}

		// Check for Cloze deletion in a single paragraph
		if clozeRe.MatchString(line) {
			cards = append(cards, makeParsedCard(line, "", "Cloze", tagsList))
			continue
		}
	}

	if inHeadingBlock && currentFront != "" {
		back := strings.TrimSpace(currentBack.String())
		if back != "" {
			cards = append(cards, makeParsedCard(currentFront, back, defaultNotetype, tagsList))
		}
	}

	return cards
}

func makeParsedCard(front, back, notetype string, tags []string) ParsedCard {
	fields := map[string]string{
		"Front": front,
	}
	if back != "" {
		fields["Back"] = back
	}
	if notetype == "Cloze" {
		fields["Text"] = front
	}
	return ParsedCard{
		Front:    front,
		Back:     back,
		Notetype: notetype,
		Tags:     tags,
		Fields:   fields,
	}
}
