# PubMed CLI Brief

## API Identity
- Domain: NCBI Entrez **E-utilities** (`https://eutils.ncbi.nlm.nih.gov/entrez/eutils`) — the programmatic gateway to PubMed's 37M+ biomedical citations.
- Users: clinicians, researchers, evidence-synthesis teams — here specifically the **PDSP dashboard** (Chinese AI hospital) consuming JSON via a widget.
- Data profile: article metadata (title, authors, journal/source, pubdate, PMID, DOI, MeSH terms, abstract). Search returns PMID lists + counts; summaries and full records are fetched by PMID.
- Auth: **none required** (public, 3 req/s). Optional `NCBI_API_KEY` raises the cap to 10 req/s; wired into hand-built commands, never mandatory.

## Reachability Risk
- None. Live-verified 2026-07-17: `esearch` (HTTP 200, count 516k for covid-19), `esummary` (200, clean JSON fields), `efetch` (200, MEDLINE abstract text). Clinical-filter mechanism verified: `atrial fibrillation` 132,524 → therapy/broad 53,664 → systematic-review `[sb]` 3,716 → diagnosis/narrow 2,998.
- Probe-safe endpoint used: `GET /esearch.fcgi?db=pubmed&term=covid-19&retmax=1`.

## Top Workflows
1. **Search + refine on a clinical / disease topic** (Rick's stated need) — narrow a broad topic to therapy/diagnosis/etiology/prognosis evidence or systematic reviews. This is PubMed's *Clinical Queries* feature, reproducible via canonical Haynes filter strings on esearch.
2. **Fetch clean article cards for a query** — one call that returns title/authors/journal/date/DOI/PMID (+ optional abstract) ready to render in a widget.
3. **"What's new" on a saved topic** — new articles since last check for a standing query (disease surveillance / journal club).
4. **Publication trend over time** — counts per year/month for a topic (widget chart data).
5. **Pull an abstract / full record by PMID** and its related + cited-by links.

## Table Stakes (from incumbents: EDirect, Biopython Entrez, pymed, metapub, pubmed MCP servers)
- esearch / esummary / efetch / elink / einfo / espell passthroughs.
- Query by term, author, journal, date range, article type; paging (retmax/retstart).
- Export to BibTeX / RIS / CSV / NDJSON.
- Related-article + cited-by traversal; PMC full-text link resolution.

## Data Layer
- Primary entities: `article` (PMID keyed: title, authors[], journal, pubdate, doi, mesh[], abstract), `saved_query` (name, term, filters, last_run, seen_pmids).
- Sync cursor: per-saved-query `seen_pmids` set + `last_run` timestamp → powers `watch`/`since`.
- FTS/search: full-text index over synced article title+abstract+mesh for offline query.

## User Vision
- Primary consumer is a **dashboard widget** → every command must emit clean, stable `--json` with `--select` field narrowing, typed exit codes, and bounded responses.
- Explicit ask (mid-session): "use it for clinical queries — quickly refine PubMed searches on clinical or disease-specific topics." → `clinical` is the headline transcendence command.

## Product Thesis
- Name: **pubmed-pp-cli**
- Why it should exist: incumbents (EDirect, Biopython) are power-user toolkits that emit XML and assume a Unix pipeline. None give a widget/agent a single command that returns *clinically refined, ranked, clean-JSON article cards* with an offline store for trend + "what's new" queries. This CLI absorbs the E-utilities surface and adds a clinical-refinement + local-analytics layer purpose-built for a dashboard.

## Build Priorities
1. Data layer (article, saved_query) + sync + FTS + SQL.
2. Absorbed E-utilities surface (search/summary/fetch/links/info/spell) with clean JSON.
3. Transcendence: `clinical`, `find` (search→cards), `abstract`, `trend`, `watch`/`since`, `saved`, `top`/`landscape`, `pico`.
