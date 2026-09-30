// Copyright 2026 Sameer Bajaj and contributors. Licensed under Apache-2.0. See LICENSE.


package cli

import (
	"github.com/spf13/cobra"
)

func newSharedCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "shared",
		Short:       "Search and inspect community shared decks",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:parent-group": "true", "pp:api-resource": "true", "pp:typed-exit-codes": "0,2"},
		RunE:        parentNoSubcommandRunE(flags),
	}

	cmd.AddCommand(newSharedInfoCmd(flags))
	cmd.AddCommand(newSharedSearchCmd(flags))
	addNovelCommandIfAbsent(cmd, newNovelSharedBriefCmd(flags))
	return cmd
}
