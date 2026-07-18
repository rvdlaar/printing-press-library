# PubMed CLI Absorb Manifest

## Absorbed (match or beat everything that exists)
| # | Feature | Best Source | Our Implementation | Added Value |
|---|---------|-------------|--------------------|-------------|
| 1 | Search PubMed → PMIDs + count | EDirect `esearch`, Biopython `esearch` | (generated endpoint) search run | Clean JSON, date/type/sort filters, paging |
| 2 | Article summaries (title/authors/journal/date/DOI) | EDirect `esummary`, pymed | (generated endpoint) summary get | JSON by default, `--select` field narrowing |
| 3 | Full record / abstract fetch | EDirect `efetch`, metapub | (generated endpoint) fetch get | Raw MEDLINE/XML passthrough for power users |
| 4 | Related + cited-by links | EDirect `elink` | (generated endpoint) links get | Related articles + PMC full-text link resolution |
| 5 | Database / field metadata | EDirect `einfo` | (generated endpoint) info get | Discover searchable fields + link sets |
| 6 | Spelling suggestions | EDirect `espell` | (generated endpoint) spell get | Query correction before searching |
| 7 | Search + auto-summary in one step | Biopython (manual chain) | (behavior in pubmed-pp-cli find) | One call returns article cards, no PMID round-trip |
| 8 | Abstract by PMID as clean text/JSON | metapub, pubmed-lookup | (behavior in pubmed-pp-cli abstract) | efetch XML parsed to clean JSON/text |
| 9 | Filter by author / journal / date / article-type | EDirect field syntax | (behavior in pubmed-pp-cli find) | Ergonomic flags compile to PubMed field tags |
| 10 | Export BibTeX / RIS / CSV / NDJSON | pymed, RISmed | (behavior in pubmed-pp-cli find) | `--format` on read commands + offline store export |
| 11 | Offline full-text search over pulled records | (none — incumbents are stateless) | (behavior in pubmed-pp-cli search/sql) | FTS + SQL over local SQLite article store |
| 12 | Sync a query's results into a local store | (none) | (behavior in pubmed-pp-cli sync) | Persist articles for offline + analytics |
| 13 | Health / reachability check | (none) | (generated endpoint) doctor | einfo probe, works without a key |

## Transcendence (only possible with our approach)
| # | Feature | Command | Buildability | Why Only We Can Do This | Long Description |
|---|---------|---------|--------------|-------------------------|------------------|
| 1 | Clinical Queries refinement | clinical | hand-code | Layers canonical Haynes therapy/diagnosis/etiology/prognosis/prediction + systematic-review filter strategies onto esearch; no CLI exposes this. | Use for evidence of a specific type (therapy/diagnosis/etiology/prognosis) on a clinical or disease topic. For a plain search use 'find'. |
| 2 | Article cards in one call | find | hand-code | Composes esearch + esummary (+ optional efetch abstract) into one widget-ready JSON payload. | Use for a clean list of article cards for a query. For type-specific clinical evidence use 'clinical'. |
| 3 | Abstract by PMID (clean) | abstract | hand-code | Parses efetch XML/MEDLINE into structured JSON the API never returns directly. | none |
| 4 | Publication trend over time | trend | hand-code | Fans esearch across date buckets and aggregates counts locally. | none |
| 5 | What's new since last check | watch | hand-code | Diffs against a local seen-PMID store per saved query. | Use for only-new results on a standing topic. For a full re-run use 'find'. |
| 6 | Saved named queries | saved | hand-code | Persists named queries + filters + last-run in SQLite. | none |
| 7 | Topic landscape (facets) | top | hand-code | Facets the local article store by journal/author/year/MeSH. | none |
| 8 | PICO question builder | pico | hand-code | Compiles a structured P/I/C/O clinical question into a precise query. | none |
| 9 | Two-term co-occurrence lens | cooccur | hand-code | Fans esearch across term A, term B, and A AND B per year; computes association series locally. | Use to track drug+adverse-effect or disease+intervention signals over time. For single-topic counts use 'trend'. |

### User additions (approved 2026-07-17, "add all") — implemented as flags/enrichments on find + clinical + pico
| # | Feature | Command | Buildability | Notes |
|---|---------|---------|--------------|-------|
| A1 | Open-access / PMC filter | (behavior in pubmed-pp-cli find) | hand-code | `--oa` appends free-full-text/PMC subset filter and resolves the PMC full-text link into each card; also on `clinical`. |
| A2 | Evidence-level badges | (behavior in pubmed-pp-cli find) | hand-code | Classifies each result's publication types into RCT / systematic-review / meta-analysis / guideline / review / case-report; adds `evidence_level` field and `--evidence <type>` filter; also on `clinical`. |
| A3 | Clinical scope filters | (behavior in pubmed-pp-cli find) | hand-code | `--humans`, `--english`, `--age <group>`, `--sex`, `--species` compile to PubMed field tags; also on `clinical` and `pico`. |

## Stubs
- None. All transcendence rows are shipping scope.

## Scope summary
- **13 absorbed** features across EDirect / Biopython / pymed / metapub (the full public E-utilities surface + widget/offline value-adds).
- **9 transcendence commands** (all hand-code): `clinical`, `find`, `abstract`, `trend`, `watch`, `saved`, `top`, `pico`, `cooccur` — headlined by `clinical`/`pico` for clinical refinement.
- **3 user-approved enrichments** (A1–A3) implemented as flags/behaviors on `find`/`clinical`/`pico`: open-access filter, evidence-level badges, clinical scope filters.
- No incumbent combines the E-utilities surface with clinical-query refinement + a local analytics store; that gap is this CLI's reason to exist.
