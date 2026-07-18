// Copyright 2026 Rick van de Laar and contributors. Licensed under Apache-2.0. See LICENSE.
// Novel command: find — search PubMed and return clean, evidence-tagged article cards.

package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mvanhorn/printing-press-library/library/other/pubmed/internal/eutils"
)

type findResult struct {
	Query         string           `json:"query"`
	ResolvedQuery string           `json:"resolved_query"`
	TotalCount    int              `json:"total_count"`
	Returned      int              `json:"returned"`
	Articles      []eutils.Article `json:"articles"`
}

func newNovelFindCmd(flags *rootFlags) *cobra.Command {
	var limit int
	var withAbstract, oa, humans, english bool
	var age, sex, species, evidence, sort string

	cmd := &cobra.Command{
		Use:   "find <query>",
		Short: "Search PubMed and return clean, evidence-tagged article cards in one call",
		Long: strings.Trim(`
Search PubMed and return ready-to-render article cards (title, authors, journal,
date, DOI, PMID, evidence-level badge, optional abstract) in a single command.

Use this for a general literature search. For evidence of a specific clinical
type (therapy, diagnosis, etc.) use 'clinical' instead.`, "\n"),
		Example: strings.Trim(`
  pubmed-pp-cli find "semaglutide heart failure" --limit 10 --json
  pubmed-pp-cli find "long covid" --oa --humans --with-abstract --json
  pubmed-pp-cli find "atrial fibrillation" --evidence randomized-controlled-trial --json`, "\n"),
		Annotations: map[string]string{"mcp:read-only": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && cmd.Flags().NFlag() == 0 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				fmt.Fprintln(cmd.OutOrStdout(), "would search PubMed and return article cards")
				return nil
			}
			query := strings.TrimSpace(strings.Join(args, " "))
			if query == "" {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("a search query is required"))
			}

			scopeFilters, err := eutils.ScopeFilters(eutils.ScopeOpts{
				Humans: humans, English: english, Age: age, Sex: sex, Species: species,
			})
			if err != nil {
				return usageErr(err)
			}
			var filters []string
			filters = append(filters, scopeFilters...)
			if oa {
				filters = append(filters, eutils.OpenAccessFilter())
			}
			term := query
			if len(filters) > 0 {
				term = eutils.ComposeQuery(query, filters...)
			}

			ctx, cancel := boundCtx(cmd.Context(), flags)
			defer cancel()
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			sr, err := eutils.Search(ctx, c, term, eutils.SearchOpts{Retmax: limit, Sort: sort})
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
					// Abstract enrichment is best-effort; surface a warning but
					// still return the cards.
					fmt.Fprintf(cmd.ErrOrStderr(), "warning: could not fetch abstracts: %v\n", err)
				}
			}
			return emitLive(cmd, flags, findResult{
				Query:         query,
				ResolvedQuery: term,
				TotalCount:    sr.Count,
				Returned:      len(articles),
				Articles:      articles,
			})
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 20, "Max article cards to return")
	cmd.Flags().BoolVar(&withAbstract, "with-abstract", false, "Include each article's abstract text")
	cmd.Flags().BoolVar(&oa, "oa", false, "Restrict to free full-text (open-access) articles")
	cmd.Flags().BoolVar(&humans, "humans", false, "Restrict to human studies")
	cmd.Flags().BoolVar(&english, "english", false, "Restrict to English-language articles")
	cmd.Flags().StringVar(&age, "age", "", "Restrict to an age group: "+strings.Join(eutils.AgeGroups(), ", "))
	cmd.Flags().StringVar(&sex, "sex", "", "Restrict to sex: male | female")
	cmd.Flags().StringVar(&species, "species", "", "Restrict to species: humans | animals")
	cmd.Flags().StringVar(&evidence, "evidence", "", "Keep only a study type: "+strings.Join(eutils.EvidenceLevels(), ", "))
	cmd.Flags().StringVar(&sort, "sort", "", "Sort order: relevance | pub_date | Author | JournalName")
	return cmd
}
