// Copyright 2026 Sameer Bajaj and contributors. Licensed under Apache-2.0. See LICENSE.


package cli

import (
	"github.com/spf13/cobra"
)

func newCardCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "card",
		Short:       "Manage and create notes and cards in your AnkiWeb collection",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:parent-group": "true", "pp:api-resource": "true", "pp:typed-exit-codes": "0,2"},
		RunE:        parentNoSubcommandRunE(flags),
	}

	cmd.AddCommand(newCardAddCmd(flags))
	cmd.AddCommand(newCardGetCmd(flags))
	cmd.AddCommand(newCardNotetypesCmd(flags))
	cmd.AddCommand(newCardSearchCmd(flags))
	addNovelCommandIfAbsent(cmd, newNovelCardDuplicatesCmd(flags))
	addNovelCommandIfAbsent(cmd, newNovelCardImportCmd(flags))
	return cmd
}
