---
name: pp-pubmed
description: "Every PubMed E-utility, plus clinical-query refinement, a local article store, and trend/what's-new analytics no other PubMed tool ships. Trigger phrases: `search pubmed for`, `find clinical evidence on`, `what's new in the literature on`, `publication trend for`, `refine this pubmed search`, `use pubmed`, `run pubmed`."
author: "Rick van de Laar"
license: "Apache-2.0"
argument-hint: "<command> [args] | install cli|mcp"
allowed-tools: "Read Bash"
metadata:
  openclaw:
    requires:
      bins:
        - pubmed-pp-cli
    install:
      - kind: go
        bins: [pubmed-pp-cli]
        module: github.com/mvanhorn/printing-press-library/library/other/pubmed/cmd/pubmed-pp-cli
---

# PubMed — Printing Press CLI

## Prerequisites: Install the CLI

This skill drives the `pubmed-pp-cli` binary. **You must verify the CLI is installed before invoking any command from this skill.** If it is missing, install it first:

1. Install via the Printing Press installer. It defaults binaries to `$HOME/.local/bin` on macOS/Linux and `%LOCALAPPDATA%\Programs\PrintingPress\bin` on Windows:
   ```bash
   npx -y @mvanhorn/printing-press-library install pubmed --cli-only
   ```
2. Verify: `pubmed-pp-cli --version`
3. Ensure the reported install directory is on `$PATH` for the agent/runtime that will invoke this skill.

If the `npx` install fails (no Node, offline, etc.), fall back to a direct Go install (requires Go 1.26.5 or newer). This installs into `$GOPATH/bin` (default `$HOME/go/bin`), so add that directory to `$PATH` instead:

```bash
go install github.com/mvanhorn/printing-press-library/library/other/pubmed/cmd/pubmed-pp-cli@latest
```

If `--version` reports "command not found" after install, the runtime cannot see the binary directory on `$PATH`. Do not proceed with skill commands until verification succeeds.

pubmed-pp-cli wraps NCBI E-utilities with agent-native JSON on every command, then adds a clinical-evidence layer (clinical, pico), widget-ready article cards (find), and a local SQLite store that powers trend, watch (what's new), and topic landscape. Built for dashboards and agents, not Unix one-liners.

## When to Use This CLI

Reach for pubmed-pp-cli when an agent or dashboard needs biomedical literature: clinically refined evidence on a disease or intervention, clean article metadata for a query, publication trends, or a 'what's new' feed on a standing topic. It returns structured JSON built for programmatic consumption.

## Anti-triggers

Do not use this CLI for:
- Do not use it to retrieve full-text article PDFs — E-utilities returns metadata and abstracts, not licensed full text (use the PMC links it surfaces).
- Do not use it for non-PubMed databases beyond what Entrez exposes; it is scoped to PubMed literature.
- Do not use it as a citation manager of record — export to BibTeX/RIS and manage citations in a dedicated tool.

## Unique Capabilities

These capabilities aren't available in any other tool for this API.

### Clinical evidence
- **`clinical`** — Refine a clinical or disease topic to therapy, diagnosis, etiology, prognosis, clinical-prediction, or systematic-review evidence, with open-access, evidence-level, and human/language/age scope filters.

  _When a clinician or agent needs evidence of a specific type on a disease topic, this returns the right slice without hand-crafting filter syntax._

  ```bash
  pubmed-pp-cli clinical "atrial fibrillation" --category therapy --scope narrow --humans --english --json
  ```
- **`pico`** — Compose a structured Patient/Intervention/Comparison/Outcome question into a precise PubMed query and run it.

  _Lets a clinician or agent ask a well-formed evidence question and get the matching literature directly._

  ```bash
  pubmed-pp-cli pico --patient "type 2 diabetes" --intervention "SGLT2 inhibitor" --outcome "cardiovascular mortality" --json
  ```

### Widget-ready reads
- **`find`** — Search and return clean article cards (title, authors, journal, date, DOI, PMID, evidence-level badge, optional abstract) with open-access and clinical scope filters, in one command.

  _The dashboard widget calls one command and renders ranked, evidence-tagged cards — no client-side orchestration of E-utilities._

  ```bash
  pubmed-pp-cli find "semaglutide heart failure" --limit 10 --with-abstract --oa --humans --json
  ```
- **`abstract`** — Fetch one or more abstracts by PMID as clean structured JSON or plain text.

  _Gives an agent the abstract text for a known PMID in one clean call, no XML parsing._

  ```bash
  pubmed-pp-cli abstract 42467460 --json
  ```

### Local analytics
- **`trend`** — Publication counts per year (or month) for a topic, ready to chart.

  _Gives the dashboard a time-series of research attention on any disease or intervention._

  ```bash
  pubmed-pp-cli trend "CAR-T therapy" --from 2015 --to 2026 --json
  ```
- **`top`** — Facet a query's top results by journal, author, year, or evidence level to show where the literature concentrates.

  _Surfaces the leading journals/authors on a disease topic for at-a-glance dashboard context._

  ```bash
  pubmed-pp-cli top "long covid" --by journal --limit 15 --json
  ```
- **`cooccur`** — Track how often two terms (e.g. a drug and an adverse effect, or a disease and an intervention) appear together in the literature over time.

  _Surfaces emerging drug-safety or disease-intervention signals as a time series for the dashboard._

  ```bash
  pubmed-pp-cli cooccur "semaglutide" "pancreatitis" --from 2018 --to 2026 --json
  ```

### Local state that compounds
- **`watch`** — Return only articles new since the last run of a saved query, using a local seen-PMID store.

  _Powers a live 'new research' panel and disease-surveillance alerts without re-showing seen results._

  ```bash
  pubmed-pp-cli watch "sepsis biomarkers" --json
  ```
- **`saved list`** — Persist named queries with their filters, re-run them, and see when each last ran.

  _The widget can drive a fixed panel of standing clinical questions by name._

  ```bash
  pubmed-pp-cli saved add af-anticoag "atrial fibrillation AND anticoagulation" --json
  ```

## Command Reference

**efetch** — Fetch full records (efetch) — abstracts, MeSH, full author lists (XML/text)

- `pubmed-pp-cli efetch` — Fetch full records for PMIDs (efetch returns XML or MEDLINE text, not JSON)

**einfo** — Database metadata (einfo) — searchable fields and link sets

- `pubmed-pp-cli einfo` — List Entrez databases, or describe fields/links for one db

**elink** — Related and cited-by links (elink) — related articles, PMC full text

- `pubmed-pp-cli elink` — Find related articles or links for PMIDs

**esearch** — Search PubMed (esearch) — returns matching PMIDs and total count

- `pubmed-pp-cli esearch` — Run a raw PubMed query and return PMIDs + count

**esummary** — Fetch article summaries (esummary) — title, authors, journal, date, DOI

- `pubmed-pp-cli esummary` — Get document summaries for one or more PMIDs


### Finding the right command

When you know what you want to do but not which command does it, ask the CLI directly:

```bash
pubmed-pp-cli which "<capability in your own words>"
```

`which` resolves a natural-language capability query to the best matching command from this CLI's curated feature index. Exit code `0` means at least one match; exit code `2` means no confident match — fall back to `--help` or use a narrower query.

## Recipes

### Widget: refined evidence cards

```bash
pubmed-pp-cli clinical "heart failure" --category therapy --scope broad --limit 8 --agent --select articles.title,articles.journal,articles.pmid,articles.doi
```

Narrow a topic to therapy evidence and return only the fields the widget renders.

### What's new on a standing topic

```bash
pubmed-pp-cli watch "sepsis biomarkers" --agent
```

Only articles not seen on the previous run, for a live 'new research' panel.

### Trend chart data

```bash
pubmed-pp-cli trend "long covid" --from 2020 --to 2026 --agent
```

Per-year publication counts ready to feed a time-series chart.

### PICO evidence question

```bash
pubmed-pp-cli pico --patient "type 2 diabetes" --intervention "SGLT2 inhibitor" --outcome "cardiovascular mortality" --agent
```

Compose a structured EBM question and return the matching literature.

## Auth Setup

No authentication required.

Run `pubmed-pp-cli doctor` to verify setup.

## Agent Mode

Add `--agent` to any command. Expands to: `--json --compact --no-input --no-color --yes`.

- **Pipeable** — JSON on stdout, errors on stderr
- **Filterable** — `--select` keeps a subset of fields. Dotted paths descend into nested structures; arrays traverse element-wise. Critical for keeping context small on verbose APIs:

  ```bash
  pubmed-pp-cli efetch --id 550e8400-e29b-41d4-a716-446655440000 --agent --select id,name,status
  ```
- **Previewable** — `--dry-run` shows the request without sending
- **Offline-friendly** — sync/search commands can use the local SQLite store when available
- **Non-interactive** — never prompts, every input is a flag
- **Read-only** — do not use this CLI for create, update, delete, publish, comment, upvote, invite, order, send, or other mutating requests

### Response envelope

Commands that read from the local store or the API wrap output in a provenance envelope:

```json
{
  "meta": {"source": "live" | "local", "synced_at": "...", "reason": "..."},
  "results": <data>
}
```

Parse `.results` for data and `.meta.source` to know whether it's live or local. A human-readable `N results (live)` summary is printed to stderr only when stdout is a terminal AND no machine-format flag (`--json`, `--csv`, `--compact`, `--quiet`, `--plain`, `--select`) is set — piped/agent consumers and explicit-format runs get pure JSON on stdout.

## Paths and state

Agents should treat the CLI's path resolver as part of the runtime contract:

- Use `--home <dir>` for one invocation, or set `PUBMED_HOME=<dir>` to relocate all four path kinds under one root.
- Use per-kind env vars only when a specific kind must diverge: `PUBMED_CONFIG_DIR`, `PUBMED_DATA_DIR`, `PUBMED_STATE_DIR`, `PUBMED_CACHE_DIR`.
- Resolution order is per-kind env var, `--home`, `PUBMED_HOME`, XDG (`XDG_CONFIG_HOME`, `XDG_DATA_HOME`, `XDG_STATE_HOME`, `XDG_CACHE_HOME`), then platform defaults.
- `config` contains settings like `config.toml` and profiles. `data` contains `credentials.toml`, `data.db`, cookies, and auth sidecars. `state` contains persisted queries, jobs, and `teach.log`. `cache` contains regenerable HTTP/cache files.
- Stored secrets live in `credentials.toml` under the data dir. Existing legacy `config.toml` secrets are read for compatibility and leave `config.toml` on the first auth write.
- Run `pubmed-pp-cli doctor --fail-on warn` to surface path and credential-location warnings. `agent-context` exposes a schema v4 `paths` block for agents that need the resolved dirs.
- For MCP, pass relocation through the MCP host config. The MCP binary does not inherit CLI flags:

  ```json
  {
    "mcpServers": {
      "pubmed": {
        "command": "pubmed-pp-mcp",
        "env": {
          "PUBMED_HOME": "/srv/pubmed"
        }
      }
    }
  }
  ```

Fleet precedence: an inherited per-kind env var overrides an explicit `--home` for that kind. Use `PUBMED_HOME` or per-kind vars as durable fleet levers, and use `--home` only for a single invocation. Relocation is not reversible by unsetting env vars; move files manually before clearing `PUBMED_HOME`, or `doctor` will not find credentials left under the former root.

## Agent Feedback

When you (or the agent) notice something off about this CLI, record it:

```
pubmed-pp-cli feedback "the --since flag is inclusive but docs say exclusive"
pubmed-pp-cli feedback --stdin < notes.txt
pubmed-pp-cli feedback list --json --limit 10
```

Entries are stored locally as `feedback.jsonl` under the resolved data dir. They are never POSTed unless `PUBMED_FEEDBACK_ENDPOINT` is set AND either `--send` is passed or `PUBMED_FEEDBACK_AUTO_SEND=true`. Default behavior is local-only.

Write what *surprised* you, not a bug report. Short, specific, one line: that is the part that compounds.

## Output Delivery

Every command accepts `--deliver <sink>`. The output goes to the named sink in addition to (or instead of) stdout, so agents can route command results without hand-piping. Three sinks are supported:

| Sink | Effect |
|------|--------|
| `stdout` | Default; write to stdout only |
| `file:<path>` | Atomically write output to `<path>` (tmp + rename) |
| `webhook:<url>` | POST the output body to the URL (`application/json` or `application/x-ndjson` when `--compact`) |

Unknown schemes are refused with a structured error naming the supported set. Webhook failures return non-zero and log the URL + HTTP status on stderr.

## Named Profiles

A profile is a saved set of flag values, reused across invocations. Use it when a scheduled or recurring agent reuses the same saved flags while providing different input each run.

```
pubmed-pp-cli profile save briefing --json
pubmed-pp-cli --profile briefing efetch --id 550e8400-e29b-41d4-a716-446655440000
pubmed-pp-cli profile list --json
pubmed-pp-cli profile show briefing
pubmed-pp-cli profile delete briefing --yes
```

Explicit flags always win over profile values; profile values win over defaults. `agent-context` lists all available profiles under `available_profiles` so introspecting agents discover them at runtime.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 2 | Usage error (wrong arguments) |
| 3 | Resource not found |
| 5 | API error (upstream issue) |
| 7 | Rate limited (wait and retry) |
| 10 | Config error |

## Argument Parsing

Parse `$ARGUMENTS`:

1. **Empty, `help`, or `--help`** → show `pubmed-pp-cli --help` output
2. **Starts with `install`** → ends with `mcp` → MCP installation; otherwise → see Prerequisites above
3. **Anything else** → Direct Use (execute as CLI command with `--agent`)

## MCP Server Installation

1. Install the MCP server:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/other/pubmed/cmd/pubmed-pp-mcp@latest
   ```
2. Register with Claude Code:
   ```bash
   claude mcp add pubmed-pp-mcp -- pubmed-pp-mcp
   ```
3. Verify: `claude mcp list`

## Direct Use

1. Check if installed: `which pubmed-pp-cli`
   If not found, offer to install (see Prerequisites at the top of this skill).
2. Match the user query to the best command from the Unique Capabilities and Command Reference above.
3. Execute with the `--agent` flag:
   ```bash
   pubmed-pp-cli <command> [subcommand] [args] --agent
   ```
4. If ambiguous, drill into subcommand help: `pubmed-pp-cli <command> --help`.
