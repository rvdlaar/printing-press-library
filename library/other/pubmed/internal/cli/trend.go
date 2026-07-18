// Copyright 2026 Rick van de Laar and contributors. Licensed under Apache-2.0. See LICENSE.
// Novel command: trend — publication counts per year for a topic.

package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/mvanhorn/printing-press-library/library/other/pubmed/internal/cliutil"
	"github.com/mvanhorn/printing-press-library/library/other/pubmed/internal/eutils"
)

type trendPoint struct {
	Year  int `json:"year"`
	Count int `json:"count"`
}

type trendResult struct {
	Query  string       `json:"query"`
	From   int          `json:"from"`
	To     int          `json:"to"`
	Total  int          `json:"total"`
	Points []trendPoint `json:"points"`
	Note   string       `json:"note,omitempty"`
}

func newNovelTrendCmd(flags *rootFlags) *cobra.Command {
	var from, to int

	cmd := &cobra.Command{
		Use:   "trend <query>",
		Short: "Publication counts per year for a topic, ready to chart",
		Long: strings.Trim(`
Count PubMed publications per year for a topic across a date range — a
ready-to-chart time series of research attention on a disease or intervention.`, "\n"),
		Example: strings.Trim(`
  pubmed-pp-cli trend "CAR-T therapy" --from 2015 --to 2026 --json
  pubmed-pp-cli trend "long covid" --from 2020 --json`, "\n"),
		Annotations: map[string]string{"mcp:read-only": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && cmd.Flags().NFlag() == 0 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				fmt.Fprintln(cmd.OutOrStdout(), "would compute per-year publication counts")
				return nil
			}
			query := strings.TrimSpace(strings.Join(args, " "))
			if query == "" {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("a search query is required"))
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

			result := trendResult{Query: query, From: from, To: to}
			// Cap the number of buckets so the live-dogfood 30s window is safe.
			if cliutil.IsDogfoodEnv() && to-from > 2 {
				from = to - 2
				result.From = from
				result.Note = "date range narrowed to the last 3 years under dogfood"
			}

			ctx, cancel := boundCtx(cmd.Context(), flags)
			defer cancel()
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			interval := eutils.ThrottleInterval()
			for year := from; year <= to; year++ {
				term := eutils.ComposeQuery(query, eutils.YearRangeFilter(year, year))
				count, err := eutils.SearchCount(ctx, c, term)
				if err != nil {
					return classifyAPIError(err, flags)
				}
				result.Points = append(result.Points, trendPoint{Year: year, Count: count})
				result.Total += count
				if year < to {
					time.Sleep(interval)
				}
			}
			return emitLive(cmd, flags, result)
		},
	}
	cmd.Flags().IntVar(&from, "from", 0, "First year (default: 9 years before --to)")
	cmd.Flags().IntVar(&to, "to", 0, "Last year (default: current year)")
	return cmd
}
