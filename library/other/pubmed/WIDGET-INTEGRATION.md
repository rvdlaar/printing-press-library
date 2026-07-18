# PDSP widget integration — pubmed-pp-cli

Turnkey spec for calling this CLI from a dashboard widget. Every command/flag/path
below was verified against the built binary on 2026-07-17.

## Install
```bash
cd <this dir> && go build -o ~/.local/bin/pubmed-pp-cli ./cmd/pubmed-pp-cli
# already installed at ~/.local/bin/pubmed-pp-cli (on PATH)
```
**macOS / Little Snitch:** a freshly-built binary is an unapproved process — its
outbound is silently dropped (`connect: bad file descriptor`, while `curl` works).
Approve it **once** in Little Snitch **Network Monitor** → find `pubmed-pp-cli` →
Allow → `eutils.ncbi.nlm.nih.gov`. Re-approve after any rebuild.

## Auth & rate
No credential required (public NCBI E-utilities, 3 req/s). For a busier widget set
`NCBI_API_KEY` (10 req/s); the CLI self-throttles via `--rate-limit` (default 3).

## Calling mechanism
- **Subprocess (recommended for a widget):** exec `pubmed-pp-cli <cmd> … --json`,
  read stdout as JSON, check the exit code. Stateless, one process per call.
- **MCP:** `go build -o ~/.local/bin/pubmed-pp-mcp ./cmd/pubmed-pp-mcp` exposes every
  command as an MCP tool for agent-driven dashboards.

## The widget calls (verified)
Each returns a **flat** JSON object on stdout. Use `--select` to trim to the fields
the widget renders. `--select` paths are **flat** (`articles.title`, NOT
`results.articles.title`); add `--agent` only if you want a `{meta, results:{…}}`
envelope (then the same `--select articles.X` paths nest under `.results`).

**1. Refined clinical evidence** (headline)
```bash
pubmed-pp-cli clinical "atrial fibrillation" --category therapy --scope broad \
  --limit 8 --json --select articles.pmid,articles.title,articles.journal,articles.evidence_level,articles.url
```
`--category`: therapy | diagnosis | etiology | prognosis | prediction | reviews.
`--scope`: broad (sensitive) | narrow (specific). Optional scope filters:
`--humans --english --age aged --sex female --oa`.

**2. Article cards for a query**
```bash
pubmed-pp-cli find "long covid" --limit 10 --oa --json \
  --select articles.pmid,articles.title,articles.journal,articles.doi,articles.evidence_level,articles.url
```
Add `--with-abstract` to include abstract text.

**3. Publication trend (chart data)**
```bash
pubmed-pp-cli trend "CAR-T therapy" --from 2016 --to 2026 --json
# → { query, from, to, total, points:[{year, count}, …] }
```

**4. What's new on a standing topic**
```bash
pubmed-pp-cli watch "sepsis biomarkers" --json   # → { new_count, new_articles:[…] }
```

## Response shape (flat, plain --json)
`clinical`/`find`/`pico`: `{ topic|query, category?, scope?, resolved_query,
total_count, returned, articles:[ {pmid, title, authors:[…], journal, pub_date,
year, doi?, pmcid?, pmc_url?, evidence_level, pub_types:[…], url, abstract?} ] }`.
`authors` is always a JSON array (never null). `evidence_level` ∈ {meta-analysis,
systematic-review, guideline, randomized-controlled-trial, clinical-trial, review,
case-report, comparative-study, observational-study, article}.

## Exit codes
0 ok · 2 usage error · 3 not found · 4 auth · 5 API error · 7 rate-limited ·
10 config. A widget should treat 7 as "retry later" and 5 as "NCBI unavailable".
