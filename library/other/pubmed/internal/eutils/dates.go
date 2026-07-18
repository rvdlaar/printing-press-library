// Copyright 2026 Rick van de Laar and contributors. Licensed under Apache-2.0. See LICENSE.

package eutils

import "fmt"

// YearRangeFilter returns a PubMed publication-date range clause covering the
// inclusive [minYear, maxYear] span. Callers AND this onto a base term via
// ComposeQuery to bucket counts by time.
func YearRangeFilter(minYear, maxYear int) string {
	return fmt.Sprintf(`"%d/01/01"[Date - Publication] : "%d/12/31"[Date - Publication]`, minYear, maxYear)
}
