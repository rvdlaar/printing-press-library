// Copyright 2026 Rick van de Laar and contributors. Licensed under Apache-2.0. See LICENSE.
// Novel command: saved list — list named PubMed queries.

package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mvanhorn/printing-press-library/library/other/pubmed/internal/savedq"
)

func newNovelSavedListCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List saved queries and when each last ran",
		Example:     `  pubmed-pp-cli saved list --json`,
		Annotations: map[string]string{"mcp:read-only": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				fmt.Fprintln(cmd.OutOrStdout(), "would list saved queries")
				return nil
			}
			store, err := savedq.Open()
			if err != nil {
				return configErr(err)
			}
			// make([]..., 0) so an empty store marshals as [] not null.
			queries := store.List()
			if queries == nil {
				queries = []savedq.Saved{}
			}
			return printJSONFiltered(cmd.OutOrStdout(), queries, flags)
		},
	}
	return cmd
}
