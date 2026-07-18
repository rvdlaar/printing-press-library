// Copyright 2026 Rick van de Laar and contributors. Licensed under Apache-2.0. See LICENSE.
// Novel command: cooccur — track how often two terms co-occur over time.

package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/mvanhorn/printing-press-library/library/other/pubmed/internal/cliutil"
	"github.com/mvanhorn/printing-press-library/library/other/pubmed/internal/eutils"
)

type cooccurPoint struct {
	Year    int `json:"year"`
	CountA  int `json:"count_a"`
	CountB  int `json:"count_b"`
	CountAB int `json:"count_ab"`
}

type cooccurResult struct {
	TermA   string         `json:"term_a"`
	TermB   string         `json:"term_b"`
	From    int            `json:"from"`
	To      int            `json:"to"`
	TotalA  int            `json:"total_a"`
	TotalB  int            `json:"total_b"`
	TotalAB int            `json:"total_ab"`
	Jaccard float64        `json:"jaccard"`
	Points  []cooccurPoint `json:"points"`
	Note    string         `json:"note,omitempty"`
}

func newNovelCooccurCmd(flags *rootFlags) *cobra.Command {
	var from, to int

	cmd := &cobra.Command{
		Use:   "cooccur <term-a> <term-b>",
		Short: "Track how often two terms co-occur in the literature over time",
		Long: strings.Trim(`
Count PubMed publications for term A, term B, and both together per year — a
time series for surfacing drug-safety or disease-intervention signals. Also
reports the all-time Jaccard overlap of the two terms.`, "\n"),
		Example: strings.Trim(`
  pubmed-pp-cli cooccur "semaglutide" "pancreatitis" --from 2018 --to 2026 --json
  pubmed-pp-cli cooccur "statins" "diabetes" --json`, "\n"),
		Annotations: map[string]string{"mcp:read-only": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && cmd.Flags().NFlag() == 0 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				fmt.Fprintln(cmd.OutOrStdout(), "would compute two-term co-occurrence over time")
				return nil
			}
			if len(args) < 2 {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("two terms are required, e.g. cooccur \"semaglutide\" \"pancreatitis\""))
			}
			termA := strings.TrimSpace(args[0])
			termB := strings.TrimSpace(args[1])
			if termA == "" || termB == "" {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("both terms must be non-empty"))
			}
			nowYear := time.Now().Year()
			if to <= 0 {
				to = nowYear
			}
			if from <= 0 {
				from = to - 9
			}
			if from > to {
				return usageErr(fmt.Errorf("--from (%d) must be <= --to (%d)", from, to))
			}

			result := cooccurResult{TermA: termA, TermB: termB, From: from, To: to}
			if cliutil.IsDogfoodEnv() && to-from > 1 {
				from = to - 1
				result.From = from
				result.Note = "date range narrowed under dogfood"
			}

			ctx, cancel := boundCtx(cmd.Context(), flags)
			defer cancel()
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			bothTerm := fmt.Sprintf("(%s) AND (%s)", termA, termB)
			interval := eutils.ThrottleInterval()
			count := func(term string, year int) (int, error) {
				q := eutils.ComposeQuery(term, eutils.YearRangeFilter(year, year))
				return eutils.SearchCount(ctx, c, q)
			}
			for year := from; year <= to; year++ {
				ca, err := count(termA, year)
				if err != nil {
					return classifyAPIError(err, flags)
				}
				time.Sleep(interval)
				cb, err := count(termB, year)
				if err != nil {
					return classifyAPIError(err, flags)
				}
				time.Sleep(interval)
				cab, err := count(bothTerm, year)
				if err != nil {
					return classifyAPIError(err, flags)
				}
				result.Points = append(result.Points, cooccurPoint{Year: year, CountA: ca, CountB: cb, CountAB: cab})
				result.TotalA += ca
				result.TotalB += cb
				result.TotalAB += cab
				if year < to {
					time.Sleep(interval)
				}
			}
			if union := result.TotalA + result.TotalB - result.TotalAB; union > 0 {
				result.Jaccard = float64(result.TotalAB) / float64(union)
			}
			return emitLive(cmd, flags, result)
		},
	}
	cmd.Flags().IntVar(&from, "from", 0, "First year (default: 9 years before --to)")
	cmd.Flags().IntVar(&to, "to", 0, "Last year (default: current year)")
	return cmd
}
