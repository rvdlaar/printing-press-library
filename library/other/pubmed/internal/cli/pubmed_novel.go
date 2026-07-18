// Copyright 2026 Rick van de Laar and contributors. Licensed under Apache-2.0. See LICENSE.
// Hand-authored shared helpers for the PubMed novel commands.

package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mvanhorn/printing-press-library/library/other/pubmed/internal/eutils"
)

// emitLive marshals a Go value and routes it through the standard output
// pipeline with a live-source provenance tag, so --json/--select/--compact/--csv
// and the --agent envelope all behave like generated endpoint commands.
func emitLive(cmd *cobra.Command, flags *rootFlags, v any) error {
	raw, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return printOutputWithFlagsMeta(cmd.OutOrStdout(), json.RawMessage(raw), flags, map[string]any{"source": "live"})
}

// filterByEvidence keeps only articles whose evidence level matches want.
// An empty want returns the slice unchanged. Returns an error for an
// unrecognized level so a typo fails loudly.
func filterByEvidence(articles []eutils.Article, want string) ([]eutils.Article, error) {
	want = strings.ToLower(strings.TrimSpace(want))
	if want == "" {
		return articles, nil
	}
	valid := false
	for _, lvl := range eutils.EvidenceLevels() {
		if lvl == want {
			valid = true
			break
		}
	}
	if !valid {
		return nil, usageErr(fmt.Errorf("unknown --evidence %q (valid: %s)", want, strings.Join(eutils.EvidenceLevels(), ", ")))
	}
	out := make([]eutils.Article, 0, len(articles))
	for _, a := range articles {
		if a.EvidenceLevel == want {
			out = append(out, a)
		}
	}
	return out, nil
}
