// Copyright 2026 Rick van de Laar and contributors. Licensed under Apache-2.0. See LICENSE.
// Novel command: abstract — fetch clean abstracts by PMID.

package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mvanhorn/printing-press-library/library/other/pubmed/internal/eutils"
)

func newNovelAbstractCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "abstract <pmid> [pmid...]",
		Short: "Fetch one or more abstracts by PMID as clean structured JSON",
		Long: strings.Trim(`
Fetch abstracts for one or more PMIDs. efetch only returns XML/MEDLINE upstream;
this parses it into clean article records (title, journal, authors, abstract).`, "\n"),
		Example: strings.Trim(`
  pubmed-pp-cli abstract 42467460 --json
  pubmed-pp-cli abstract 42467460 40012345 --json`, "\n"),
		Annotations: map[string]string{"mcp:read-only": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && cmd.Flags().NFlag() == 0 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				fmt.Fprintln(cmd.OutOrStdout(), "would fetch abstracts by PMID")
				return nil
			}
			pmids := normalizePMIDs(args)
			if len(pmids) == 0 {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("at least one numeric PMID is required"))
			}

			ctx, cancel := boundCtx(cmd.Context(), flags)
			defer cancel()
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			articles, err := eutils.Summaries(ctx, c, pmids)
			if err != nil {
				return classifyAPIError(err, flags)
			}
			if err := eutils.AttachAbstracts(ctx, c, articles); err != nil {
				return classifyAPIError(err, flags)
			}
			return emitLive(cmd, flags, articles)
		},
	}
	return cmd
}

// normalizePMIDs keeps only tokens that look like PMIDs (all digits),
// splitting comma-separated args too.
func normalizePMIDs(args []string) []string {
	var out []string
	for _, a := range args {
		for _, tok := range strings.Split(a, ",") {
			tok = strings.TrimSpace(tok)
			if tok == "" {
				continue
			}
			allDigits := true
			for _, r := range tok {
				if r < '0' || r > '9' {
					allDigits = false
					break
				}
			}
			if allDigits {
				out = append(out, tok)
			}
		}
	}
	return out
}
