// Copyright 2026 Sameer Bajaj and contributors. Licensed under Apache-2.0. See LICENSE.


package cli

import (
	"github.com/spf13/cobra"
)

func newDeckCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "deck",
		Short:       "Manage decks, limits, and review hierarchies",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:parent-group": "true", "pp:api-resource": "true", "pp:typed-exit-codes": "0,2"},
		RunE:        parentNoSubcommandRunE(flags),
	}

	cmd.AddCommand(newDeckCreateCmd(flags))
	cmd.AddCommand(newDeckDeleteCmd(flags))
	cmd.AddCommand(newDeckLimitsCmd(flags))
	cmd.AddCommand(newDeckListCmd(flags))
	cmd.AddCommand(newDeckRenameCmd(flags))
	addNovelCommandIfAbsent(cmd, newNovelDeckTriageCmd(flags))
	return cmd
}
