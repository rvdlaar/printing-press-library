// Copyright 2026 Rick van de Laar and contributors. Licensed under Apache-2.0. See LICENSE.
// Novel command: watch — only articles new since the last run of a query.

package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mvanhorn/printing-press-library/library/other/pubmed/internal/eutils"
	"github.com/mvanhorn/printing-press-library/library/other/pubmed/internal/savedq"
)

type watchResult struct {
	Name        string           `json:"name"`
	Term        string           `json:"term"`
	FirstRun    bool             `json:"first_run"`
	TotalCount  int              `json:"total_count"`
	NewCount    int              `json:"new_count"`
	NewArticles []eutils.Article `json:"new_articles"`
}

func newNovelWatchCmd(flags *rootFlags) *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "watch <query-or-saved-name>",
		Short: "Return only articles new since the last watch run for a query",
		Long: strings.Trim(`
Run a query and return only the articles not seen on the previous watch run,
tracking seen PMIDs in a local store. Pass a saved-query name (see 'saved') or
an ad-hoc query string. The first run establishes the baseline and reports the
current top results as new.`, "\n"),
		Example: strings.Trim(`
  pubmed-pp-cli watch "sepsis biomarkers" --json
  pubmed-pp-cli watch af-anticoag --json`, "\n"),
		// Any non-empty string is a valid query, so there is no "invalid
		// argument" to reject; opt out of the error-path probe.
		Annotations: map[string]string{"mcp:read-only": "false", "pp:no-error-path-probe": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && cmd.Flags().NFlag() == 0 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				fmt.Fprintln(cmd.OutOrStdout(), "would return articles new since the last watch run")
				return nil
			}
			arg := strings.TrimSpace(strings.Join(args, " "))
			if arg == "" {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("a query or saved-query name is required"))
			}

			store, err := savedq.Open()
			if err != nil {
				return configErr(err)
			}
			name := savedq.NormalizeName(arg)
			term := arg
			seen := map[string]bool{}
			firstRun := true
			if saved, ok := store.Get(arg); ok {
				name = saved.Name
				term = saved.Term
				for _, p := range saved.SeenPMIDs {
					seen[p] = true
				}
				firstRun = len(saved.SeenPMIDs) == 0
			}

			ctx, cancel := boundCtx(cmd.Context(), flags)
			defer cancel()
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			sr, err := eutils.Search(ctx, c, term, eutils.SearchOpts{Retmax: limit, Sort: "pub_date"})
			if err != nil {
				return classifyAPIError(err, flags)
			}
			var newPMIDs []string
			for _, p := range sr.PMIDs {
				if !seen[p] {
					newPMIDs = append(newPMIDs, p)
				}
			}
			articles, err := eutils.Summaries(ctx, c, newPMIDs)
			if err != nil {
				return classifyAPIError(err, flags)
			}
			// Mark as seen only the PMIDs we actually surfaced as articles (plus
			// whatever was already seen, unioned inside RecordRun). A new PMID
			// whose esummary was momentarily dropped stays unseen and is retried
			// next run, rather than being silently marked seen and lost forever.
			surfaced := make([]string, 0, len(articles))
			for _, a := range articles {
				surfaced = append(surfaced, a.PMID)
			}
			if err := store.RecordRun(name, term, surfaced, sr.Count); err != nil {
				return configErr(err)
			}
			return emitLive(cmd, flags, watchResult{
				Name:        name,
				Term:        term,
				FirstRun:    firstRun,
				TotalCount:  sr.Count,
				NewCount:    len(articles),
				NewArticles: articles,
			})
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 50, "Max recent articles to check for new entries")
	return cmd
}
