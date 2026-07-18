// Copyright 2026 Rick van de Laar and contributors. Licensed under Apache-2.0. See LICENSE.
// Novel command: top — facet a query's results by journal, author, year, or evidence.

package cli

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mvanhorn/printing-press-library/library/other/pubmed/internal/cliutil"
	"github.com/mvanhorn/printing-press-library/library/other/pubmed/internal/eutils"
)

type facetBucket struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}

type topResult struct {
	Query   string        `json:"query"`
	By      string        `json:"by"`
	Total   int           `json:"total_count"`
	Sampled int           `json:"sampled"`
	Buckets []facetBucket `json:"buckets"`
}

var topDimensions = map[string]bool{"journal": true, "author": true, "year": true, "evidence": true}

func newNovelTopCmd(flags *rootFlags) *cobra.Command {
	var by, query string
	var limit, sample int

	cmd := &cobra.Command{
		Use:   "top <query>",
		Short: "Facet a query's results by journal, author, year, or evidence level",
		Long: strings.Trim(`
Run a query, sample the top results, and count them by a chosen dimension to
show where the literature concentrates:

  --by journal    leading journals
  --by author     most-published authors
  --by year       counts per publication year
  --by evidence   distribution of study types

The query can be a positional argument or the --query flag.`, "\n"),
		Example: strings.Trim(`
  pubmed-pp-cli top "long covid" --by journal --limit 15 --json
  pubmed-pp-cli top "CAR-T therapy" --by author --json`, "\n"),
		Annotations: map[string]string{"mcp:read-only": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && cmd.Flags().NFlag() == 0 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				fmt.Fprintln(cmd.OutOrStdout(), "would facet query results")
				return nil
			}
			q := strings.TrimSpace(query)
			if q == "" {
				q = strings.TrimSpace(strings.Join(args, " "))
			}
			if q == "" {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("a query is required (positional or --query)"))
			}
			by = strings.ToLower(strings.TrimSpace(by))
			if by == "" {
				by = "journal"
			}
			if !topDimensions[by] {
				return usageErr(fmt.Errorf("unknown --by %q (valid: journal, author, year, evidence)", by))
			}
			if sample <= 0 {
				sample = 100
			}
			if cliutil.IsDogfoodEnv() && sample > 25 {
				sample = 25
			}

			ctx, cancel := boundCtx(cmd.Context(), flags)
			defer cancel()
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			sr, err := eutils.Search(ctx, c, q, eutils.SearchOpts{Retmax: sample})
			if err != nil {
				return classifyAPIError(err, flags)
			}
			articles, err := eutils.Summaries(ctx, c, sr.PMIDs)
			if err != nil {
				return classifyAPIError(err, flags)
			}
			counts := map[string]int{}
			for _, a := range articles {
				for _, v := range facetValues(a, by) {
					if v != "" {
						counts[v]++
					}
				}
			}
			buckets := make([]facetBucket, 0, len(counts))
			for v, n := range counts {
				buckets = append(buckets, facetBucket{Value: v, Count: n})
			}
			sort.Slice(buckets, func(i, j int) bool {
				if buckets[i].Count != buckets[j].Count {
					return buckets[i].Count > buckets[j].Count
				}
				return buckets[i].Value < buckets[j].Value
			})
			if limit > 0 && len(buckets) > limit {
				buckets = buckets[:limit]
			}
			return emitLive(cmd, flags, topResult{
				Query:   q,
				By:      by,
				Total:   sr.Count,
				Sampled: len(articles),
				Buckets: buckets,
			})
		},
	}
	cmd.Flags().StringVar(&by, "by", "journal", "Facet dimension: journal | author | year | evidence")
	cmd.Flags().StringVar(&query, "query", "", "Query to facet (alternative to the positional argument)")
	cmd.Flags().IntVar(&limit, "limit", 15, "Max buckets to return")
	cmd.Flags().IntVar(&sample, "sample", 100, "How many top results to sample for faceting")
	return cmd
}

func facetValues(a eutils.Article, by string) []string {
	switch by {
	case "journal":
		return []string{a.Journal}
	case "author":
		return a.Authors
	case "year":
		if a.Year > 0 {
			return []string{strconv.Itoa(a.Year)}
		}
		return nil
	case "evidence":
		if a.EvidenceLevel != "" {
			return []string{a.EvidenceLevel}
		}
		return nil
	}
	return nil
}
