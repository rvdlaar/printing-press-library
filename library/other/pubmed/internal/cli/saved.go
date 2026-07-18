// Copyright 2026 Rick van de Laar and contributors. Licensed under Apache-2.0. See LICENSE.
// Novel command: saved — manage named PubMed queries (parent + add/remove).

package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mvanhorn/printing-press-library/library/other/pubmed/internal/savedq"
)

func newNovelSavedCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "saved",
		Short:       "Manage named PubMed queries (list, add, remove)",
		Long:        "Persist named queries so 'watch' and repeated searches can be driven by name.",
		Annotations: map[string]string{"mcp:read-only": "true"},
		RunE:        parentNoSubcommandRunE(flags),
	}
	cmd.AddCommand(newNovelSavedListCmd(flags))
	cmd.AddCommand(newSavedAddCmd(flags))
	cmd.AddCommand(newSavedRemoveCmd(flags))
	return cmd
}

func newSavedAddCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <name> <query>",
		Short: "Save a named query",
		Long: strings.Trim(`
Save a named query. The name is a short handle; the query is the full PubMed
term (quote it if it contains spaces or boolean operators). Changing the term
resets the watch baseline.`, "\n"),
		Example:     `  pubmed-pp-cli saved add af-anticoag "atrial fibrillation AND anticoagulation" --json`,
		Annotations: map[string]string{"mcp:read-only": "false"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && cmd.Flags().NFlag() == 0 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				fmt.Fprintln(cmd.OutOrStdout(), "would save a named query")
				return nil
			}
			if len(args) < 2 {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("both <name> and <query> are required"))
			}
			name := args[0]
			term := strings.TrimSpace(strings.Join(args[1:], " "))
			store, err := savedq.Open()
			if err != nil {
				return configErr(err)
			}
			saved, err := store.Add(name, term)
			if err != nil {
				return usageErr(err)
			}
			return printJSONFiltered(cmd.OutOrStdout(), saved, flags)
		},
	}
	return cmd
}

func newSavedRemoveCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "remove <name>",
		Short:       "Delete a saved query",
		Example:     `  pubmed-pp-cli saved remove af-anticoag --json`,
		Annotations: map[string]string{"mcp:read-only": "false"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && cmd.Flags().NFlag() == 0 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				fmt.Fprintln(cmd.OutOrStdout(), "would delete a saved query")
				return nil
			}
			if len(args) < 1 {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("a query <name> is required"))
			}
			store, err := savedq.Open()
			if err != nil {
				return configErr(err)
			}
			removed, err := store.Remove(args[0])
			if err != nil {
				return configErr(err)
			}
			if !removed {
				return notFoundErr(fmt.Errorf("no saved query named %q", args[0]))
			}
			return printJSONFiltered(cmd.OutOrStdout(), map[string]any{"removed": savedq.NormalizeName(args[0])}, flags)
		},
	}
	return cmd
}
