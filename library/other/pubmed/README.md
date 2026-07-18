# PubMed CLI

**Every PubMed E-utility, plus clinical-query refinement, a local article store, and trend/what's-new analytics no other PubMed tool ships.**

pubmed-pp-cli wraps NCBI E-utilities with agent-native JSON on every command, then adds a clinical-evidence layer (clinical, pico), widget-ready article cards (find), and a local SQLite store that powers trend, watch (what's new), and topic landscape. Built for dashboards and agents, not Unix one-liners.

## Install

The recommended path installs both the `pubmed-pp-cli` binary and the `pp-pubmed` agent skill (Claude Code, Codex, Cursor, Gemini CLI, GitHub Copilot, and other agents supported by the upstream [`skills`](https://github.com/vercel-labs/skills) CLI) in one shot:

```bash
npx -y @mvanhorn/printing-press-library install pubmed
```

For CLI only (no skill):

```bash
npx -y @mvanhorn/printing-press-library install pubmed --cli-only
```

For skill only — installs the skill into the same agents as the default command above, but skips the CLI binary (use this to update or reinstall just the skill):

```bash
npx -y @mvanhorn/printing-press-library install pubmed --skill-only
```

To constrain the skill install to one or more specific agents (repeatable — agent names match the [`skills`](https://github.com/vercel-labs/skills) CLI):

```bash
npx -y @mvanhorn/printing-press-library install pubmed --agent claude-code
npx -y @mvanhorn/printing-press-library install pubmed --agent claude-code --agent codex
```

### Without Node (Go fallback)

If `npx` isn't available (no Node, offline), install the CLI directly via Go (requires Go 1.26.5 or newer):

```bash
go install github.com/mvanhorn/printing-press-library/library/other/pubmed/cmd/pubmed-pp-cli@latest
```

This installs the CLI only — no skill.

### Pre-built binary

Download a pre-built binary for your platform from the [latest release](https://github.com/mvanhorn/printing-press-library/releases/tag/pubmed-current). On macOS, clear the Gatekeeper quarantine: `xattr -d com.apple.quarantine <binary>`. On Unix, mark it executable: `chmod +x <binary>`.

<!-- pp-hermes-install-anchor -->
## Install for Hermes

Install the CLI binary first. The installer writes binaries to a per-user managed bin directory by default: `$HOME/.local/bin` on macOS/Linux and `%LOCALAPPDATA%\Programs\PrintingPress\bin` on Windows.

```bash
npx -y @mvanhorn/printing-press-library install pubmed --cli-only
```

Then install the focused Hermes skill.

From the Hermes CLI:

```bash
hermes skills install mvanhorn/printing-press-library/cli-skills/pp-pubmed --force
```

Inside a Hermes chat session:

```bash
/skills install mvanhorn/printing-press-library/cli-skills/pp-pubmed --force
```

Restart the Hermes session or gateway if the newly installed skill is not visible immediately.

## Install for OpenClaw
Install both the CLI binary and the focused OpenClaw skill. The installer defaults binaries to a per-user bin directory (`$HOME/.local/bin` on macOS/Linux, `%LOCALAPPDATA%\Programs\PrintingPress\bin` on Windows):

```bash
npx -y @mvanhorn/printing-press-library install pubmed --agent openclaw
```

Restart the OpenClaw session or gateway if the newly installed skill is not visible immediately.

## Use with Claude Desktop

This CLI ships an [MCPB](https://github.com/modelcontextprotocol/mcpb) bundle — Claude Desktop's standard format for one-click MCP extension installs (no JSON config required).

To install:

1. Download the `.mcpb` for your platform from the [latest release](https://github.com/mvanhorn/printing-press-library/releases/tag/pubmed-current).
2. Double-click the `.mcpb` file. Claude Desktop opens and walks you through the install.

Requires Claude Desktop 1.0.0 or later. Pre-built bundles ship for macOS Apple Silicon (`darwin-arm64`) and Windows (`amd64`, `arm64`); for other platforms, use the manual config below.

<details>
<summary>Manual JSON config (advanced)</summary>

If you can't use the MCPB bundle (older Claude Desktop, unsupported platform), install the MCP binary and configure it manually.


```bash
go install github.com/mvanhorn/printing-press-library/library/other/pubmed/cmd/pubmed-pp-mcp@latest
```

Add to your Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "pubmed": {
      "command": "pubmed-pp-mcp"
    }
  }
}
```

</details>

## Quick Start

```bash
# confirm the tool is wired and E-utilities is reachable
pubmed-pp-cli doctor --dry-run

# clean article cards for a topic in one call
pubmed-pp-cli find "atrial fibrillation anticoagulation" --limit 5 --json

# refine to high-specificity therapy evidence
pubmed-pp-cli clinical "atrial fibrillation" --category therapy --scope narrow --json

# publication counts per year for a chart
pubmed-pp-cli trend "CAR-T therapy" --from 2018 --to 2026 --json

```

## Unique Features

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

## Usage

Run `pubmed-pp-cli --help` for the full command reference and flag list.

## Paths & environment variables

This CLI separates local files into four path kinds:

| Kind | Contents |
|------|----------|
| `config` | User-editable settings such as `config.toml` and saved profiles |
| `data` | Durable local data such as `data.db` |
| `state` | Runtime state such as persisted queries, jobs, and `teach.log` |
| `cache` | Regenerable HTTP/cache files |

Each kind resolves independently. The ladder is:

1. Per-kind env var: `PUBMED_CONFIG_DIR`, `PUBMED_DATA_DIR`, `PUBMED_STATE_DIR`, or `PUBMED_CACHE_DIR`
2. `--home <dir>` for this invocation
3. `PUBMED_HOME` for a flat relocated root
4. XDG env vars: `XDG_CONFIG_HOME`, `XDG_DATA_HOME`, `XDG_STATE_HOME`, `XDG_CACHE_HOME`
5. Platform defaults matching existing installs

For containers and agent sandboxes, prefer a single relocated root:

```bash
export PUBMED_HOME=/srv/pubmed
pubmed-pp-cli doctor
```

Under `PUBMED_HOME=/srv/pubmed`, the four dirs resolve to `/srv/pubmed/config`, `/srv/pubmed/data`, `/srv/pubmed/state`, and `/srv/pubmed/cache`.

MCP servers do not receive CLI flags from the host. Put relocation in the host `env` block:

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

Precedence matters in fleets: an ambient per-kind variable such as `PUBMED_DATA_DIR` overrides an explicit `--home` for that kind. Use `PUBMED_HOME` or the per-kind variables for durable fleet relocation; treat `--home` as the weaker per-invocation lever.

Relocation is one-way. Unsetting `PUBMED_HOME` does not move files back to platform defaults, and `doctor` cannot find files left under a former root. Move the files manually before unsetting relocation variables.

Existing installs keep working because the platform-default rung matches the legacy layout. Run `pubmed-pp-cli doctor --fail-on warn` to check path warnings in automation.

## Commands

### efetch

Fetch full records (efetch) — abstracts, MeSH, full author lists (XML/text)

- **`pubmed-pp-cli efetch`** - Fetch full records for PMIDs (efetch returns XML or MEDLINE text, not JSON)

### einfo

Database metadata (einfo) — searchable fields and link sets

- **`pubmed-pp-cli einfo`** - List Entrez databases, or describe fields/links for one db

### elink

Related and cited-by links (elink) — related articles, PMC full text

- **`pubmed-pp-cli elink`** - Find related articles or links for PMIDs

### esearch

Search PubMed (esearch) — returns matching PMIDs and total count

- **`pubmed-pp-cli esearch`** - Run a raw PubMed query and return PMIDs + count

### esummary

Fetch article summaries (esummary) — title, authors, journal, date, DOI

- **`pubmed-pp-cli esummary`** - Get document summaries for one or more PMIDs


## Output Formats

```bash
# Human-readable table (default in terminal, JSON when piped)
pubmed-pp-cli efetch --id 550e8400-e29b-41d4-a716-446655440000

# JSON for scripting and agents
pubmed-pp-cli efetch --id 550e8400-e29b-41d4-a716-446655440000 --json

# Filter to specific fields
pubmed-pp-cli efetch --id 550e8400-e29b-41d4-a716-446655440000 --json --select id,name,status

# Dry run — show the request without sending
pubmed-pp-cli efetch --id 550e8400-e29b-41d4-a716-446655440000 --dry-run

# Agent mode — JSON + compact + no prompts in one flag
pubmed-pp-cli efetch --id 550e8400-e29b-41d4-a716-446655440000 --agent
```

## Agent Usage

This CLI is designed for AI agent consumption:

- **Non-interactive** - never prompts, every input is a flag
- **Pipeable** - `--json` output to stdout, errors to stderr
- **Filterable** - `--select id,name` returns only fields you need
- **Previewable** - `--dry-run` shows the request without sending
- **Read-only by default** - this CLI does not create, update, delete, publish, send, or mutate remote resources
- **Offline-friendly** - sync/search commands can use the local SQLite store when available
- **Agent-safe by default** - no colors or formatting unless `--human-friendly` is set

Exit codes: `0` success, `2` usage error, `3` not found, `5` API error, `7` rate limited, `10` config error.

## Health Check

```bash
pubmed-pp-cli doctor
```

Verifies configuration and connectivity to the API.

## Configuration

Run `pubmed-pp-cli doctor` to see the resolved config, data, state, and cache directories. The platform-default config path is `~/.config/pubmed-pp-cli/config.toml`; `--home`, `PUBMED_HOME`, and per-kind env vars can relocate it.

Static request headers can be configured under `headers`; per-command header overrides take precedence.

## Troubleshooting
**Not found errors (exit code 3)**
- Check the resource ID is correct
- Run the `list` command to see available items

### API-specific
- **HTTP 429 / rate-limited on bursts** — Set NCBI_API_KEY to raise the limit from 3 to 10 requests/sec; the CLI adds it automatically when present.
- **efetch returns XML/text, not JSON** — efetch has no JSON mode upstream; use 'find --with-abstract' or 'abstract <pmid>' for clean JSON, or 'fetch get' for the raw record.
- **clinical returns very few results** — Use --scope broad for a sensitive search; --scope narrow maximizes specificity and intentionally returns fewer, higher-precision hits.

## Sources & Inspiration

This CLI was built by studying these projects and resources:

- [**biopython**](https://github.com/biopython/biopython) — Python (4400 stars)
- [**edirect**](https://github.com/NCBI-Hackathons/EDirect) — Perl (300 stars)
- [**pymed**](https://github.com/gijswobben/pymed) — Python (300 stars)
- [**metapub**](https://github.com/metapub/metapub) — Python (150 stars)

Generated by [CLI Printing Press](https://github.com/mvanhorn/cli-printing-press)
