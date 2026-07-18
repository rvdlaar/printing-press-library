// Copyright 2026 Rick van de Laar and contributors. Licensed under Apache-2.0. See LICENSE.
// Novel command: pico — compose a structured clinical question into a query.

package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mvanhorn/printing-press-library/library/other/pubmed/internal/eutils"
)

type picoComponents struct {
	Patient      string `json:"patient,omitempty"`
	Intervention string `json:"intervention,omitempty"`
	Comparison   string `json:"comparison,omitempty"`
	Outcome      string `json:"outcome,omitempty"`
}

type picoResult struct {
	PICO          picoComponents   `json:"pico"`
	ResolvedQuery string           `json:"resolved_query"`
	TotalCount    int              `json:"total_count"`
	Returned      int              `json:"returned"`
	Articles      []eutils.Article `json:"articles"`
}

func newNovelPicoCmd(flags *rootFlags) *cobra.Command {
	var patient, intervention, comparison, outcome string
	var limit int
	var withAbstract, humans, english bool

	cmd := &cobra.Command{
		Use:   "pico",
		Short: "Compose a P/I/C/O clinical question into a precise PubMed query and run it",
		Long: strings.Trim(`
Build an evidence-based-medicine question from its components and search PubMed:

  --patient       the population or problem (e.g. "type 2 diabetes")
  --intervention  the intervention or exposure (e.g. "SGLT2 inhibitor")
  --comparison    the comparator (optional, e.g. "metformin")
  --outcome       the outcome (e.g. "cardiovascular mortality")

At least a patient or intervention is required.`, "\n"),
		Example: strings.Trim(`
  pubmed-pp-cli pico --patient "type 2 diabetes" --intervention "SGLT2 inhibitor" --outcome "cardiovascular mortality" --json
  pubmed-pp-cli pico --patient "heart failure" --intervention "sacubitril valsartan" --comparison "enalapril" --json`, "\n"),
		Annotations: map[string]string{"mcp:read-only": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().NFlag() == 0 && len(args) == 0 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				fmt.Fprintln(cmd.OutOrStdout(), "would run a PICO-composed search")
				return nil
			}
			comps := picoComponents{
				Patient:      strings.TrimSpace(patient),
				Intervention: strings.TrimSpace(intervention),
				Comparison:   strings.TrimSpace(comparison),
				Outcome:      strings.TrimSpace(outcome),
			}
			var parts []string
			for _, p := range []string{comps.Patient, comps.Intervention, comps.Comparison, comps.Outcome} {
				if p != "" {
					parts = append(parts, p)
				}
			}
			if comps.Patient == "" && comps.Intervention == "" {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("at least --patient or --intervention is required"))
			}

			scopeFilters, err := eutils.ScopeFilters(eutils.ScopeOpts{Humans: humans, English: english})
			if err != nil {
				return usageErr(err)
			}
			base := parts[0]
			extra := append(parts[1:], scopeFilters...)
			term := eutils.ComposeQuery(base, extra...)

			ctx, cancel := boundCtx(cmd.Context(), flags)
			defer cancel()
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			sr, err := eutils.Search(ctx, c, term, eutils.SearchOpts{Retmax: limit})
			if err != nil {
				return classifyAPIError(err, flags)
			}
			articles, err := eutils.Summaries(ctx, c, sr.PMIDs)
			if err != nil {
				return classifyAPIError(err, flags)
			}
			if withAbstract {
				if err := eutils.AttachAbstracts(ctx, c, articles); err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "warning: could not fetch abstracts: %v\n", err)
				}
			}
			return emitLive(cmd, flags, picoResult{
				PICO:          comps,
				ResolvedQuery: term,
				TotalCount:    sr.Count,
				Returned:      len(articles),
				Articles:      articles,
			})
		},
	}
	cmd.Flags().StringVar(&patient, "patient", "", "Population or problem (P)")
	cmd.Flags().StringVar(&intervention, "intervention", "", "Intervention or exposure (I)")
	cmd.Flags().StringVar(&comparison, "comparison", "", "Comparator (C, optional)")
	cmd.Flags().StringVar(&outcome, "outcome", "", "Outcome (O)")
	cmd.Flags().IntVar(&limit, "limit", 20, "Max article cards to return")
	cmd.Flags().BoolVar(&withAbstract, "with-abstract", false, "Include each article's abstract text")
	cmd.Flags().BoolVar(&humans, "humans", false, "Restrict to human studies")
	cmd.Flags().BoolVar(&english, "english", false, "Restrict to English-language articles")
	return cmd
}
