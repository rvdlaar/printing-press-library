// Copyright 2026 Rick van de Laar and contributors. Licensed under Apache-2.0. See LICENSE.
// Novel command: clinical — refine a topic to a Clinical Queries evidence category.

package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mvanhorn/printing-press-library/library/other/pubmed/internal/eutils"
)

type clinicalResult struct {
	Topic         string           `json:"topic"`
	Category      string           `json:"category"`
	Scope         string           `json:"scope"`
	ResolvedQuery string           `json:"resolved_query"`
	TotalCount    int              `json:"total_count"`
	Returned      int              `json:"returned"`
	Articles      []eutils.Article `json:"articles"`
}

func newNovelClinicalCmd(flags *rootFlags) *cobra.Command {
	var category, scope string
	var limit int
	var withAbstract, oa, humans, english bool
	var age, sex, species, evidence string

	cmd := &cobra.Command{
		Use:   "clinical <topic>",
		Short: "Refine a clinical or disease topic to a specific evidence category",
		Long: strings.Trim(`
Refine a clinical or disease topic to a specific type of evidence using PubMed's
Clinical Queries filters:

  --category therapy      treatment effectiveness (RCTs, trials)
  --category diagnosis    diagnostic test accuracy
  --category etiology     causation / risk factors
  --category prognosis    disease course / outcomes
  --category prediction   clinical prediction guides
  --category reviews      systematic reviews (scope is ignored)

  --scope broad   sensitive: more results, catches most relevant studies
  --scope narrow  specific: fewer, higher-precision results (default: broad)

For a general (non-typed) search use 'find' instead.`, "\n"),
		Example: strings.Trim(`
  pubmed-pp-cli clinical "atrial fibrillation" --category therapy --scope narrow --json
  pubmed-pp-cli clinical "chest pain" --category diagnosis --humans --english --json
  pubmed-pp-cli clinical "heart failure" --category reviews --limit 10 --json`, "\n"),
		Annotations: map[string]string{"mcp:read-only": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && cmd.Flags().NFlag() == 0 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				fmt.Fprintln(cmd.OutOrStdout(), "would run a Clinical Queries search")
				return nil
			}
			topic := strings.TrimSpace(strings.Join(args, " "))
			if topic == "" {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("a topic is required, e.g. clinical \"atrial fibrillation\" --category therapy"))
			}
			if strings.TrimSpace(category) == "" {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("--category is required (%s)", strings.Join(eutils.ClinicalCategories(), ", ")))
			}
			clinFilter, err := eutils.ClinicalFilter(category, scope)
			if err != nil {
				return usageErr(err)
			}
			scopeFilters, err := eutils.ScopeFilters(eutils.ScopeOpts{
				Humans: humans, English: english, Age: age, Sex: sex, Species: species,
			})
			if err != nil {
				return usageErr(err)
			}
			filters := []string{clinFilter}
			filters = append(filters, scopeFilters...)
			if oa {
				filters = append(filters, eutils.OpenAccessFilter())
			}
			term := eutils.ComposeQuery(topic, filters...)

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
			articles, err = filterByEvidence(articles, evidence)
			if err != nil {
				return err
			}
			if withAbstract {
				if err := eutils.AttachAbstracts(ctx, c, articles); err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "warning: could not fetch abstracts: %v\n", err)
				}
			}
			resolvedScope := strings.ToLower(strings.TrimSpace(scope))
			if resolvedScope == "" {
				resolvedScope = "broad"
			}
			if strings.EqualFold(category, "reviews") {
				resolvedScope = "n/a"
			}
			return emitLive(cmd, flags, clinicalResult{
				Topic:         topic,
				Category:      strings.ToLower(category),
				Scope:         resolvedScope,
				ResolvedQuery: term,
				TotalCount:    sr.Count,
				Returned:      len(articles),
				Articles:      articles,
			})
		},
	}
	cmd.Flags().StringVar(&category, "category", "", "Evidence category: "+strings.Join(eutils.ClinicalCategories(), ", "))
	cmd.Flags().StringVar(&scope, "scope", "broad", "Search scope: broad (sensitive) | narrow (specific)")
	cmd.Flags().IntVar(&limit, "limit", 20, "Max article cards to return")
	cmd.Flags().BoolVar(&withAbstract, "with-abstract", false, "Include each article's abstract text")
	cmd.Flags().BoolVar(&oa, "oa", false, "Restrict to free full-text (open-access) articles")
	cmd.Flags().BoolVar(&humans, "humans", false, "Restrict to human studies")
	cmd.Flags().BoolVar(&english, "english", false, "Restrict to English-language articles")
	cmd.Flags().StringVar(&age, "age", "", "Restrict to an age group: "+strings.Join(eutils.AgeGroups(), ", "))
	cmd.Flags().StringVar(&sex, "sex", "", "Restrict to sex: male | female")
	cmd.Flags().StringVar(&species, "species", "", "Restrict to species: humans | animals")
	cmd.Flags().StringVar(&evidence, "evidence", "", "Keep only a study type: "+strings.Join(eutils.EvidenceLevels(), ", "))
	return cmd
}
