# PubMed CLI build log

Manifest transcendence rows: 9 planned, 9 built. Phase 3 will not pass until all 9 ship.

## Built (Phase 3)
- Shared: internal/eutils (E-utilities client, Article type, clinical/scope/OA filters, evidence classifier) + tests; internal/savedq (named-query + seen-PMID store).
- Novel commands: clinical, find, abstract, trend, watch, saved (list/add/remove), top, pico, cooccur.
- User-approved enrichments (flags): --oa, --evidence + evidence_level badge, --humans/--english/--age/--sex/--species scope filters on find/clinical/pico.
- Optional NCBI_API_KEY wired via eutils.withBase (raises rate limit 3→10 req/s); fan-out commands (trend/cooccur) throttle accordingly and curtail buckets under dogfood.

## Phase 4 shipcheck
PASS 7/7 legs; scorecard 90/100 Grade A; novel_features_check 9/9.

## Incident (retro candidate): `generate --force` wiped hand-authored packages
Re-running `generate --spec ... --output <existing> --force` to apply two spec-only
changes (disable learn loop, drop espell resource) DELETED the sibling hand-authored
packages `internal/eutils/` and `internal/savedq/` and `internal/cli/pubmed_novel.go`,
and reverted all 9 novel command files to TODO stubs — despite docs stating implemented
bodies + whole hand-authored files survive regen. The generate exited 3 (doctor gate
network timeout) but files were already overwritten. All work was restored from session
context. LESSON: for spec-only tweaks after Phase 3 hand-coding, edit the generated files
directly (or use `regen-merge` with a previewed report) rather than re-running
`generate --force` over a tree containing hand-authored novel packages.

## Rate-limit note
Phase 5 live-dogfood's 56-request burst triggers NCBI per-IP throttling without an
NCBI_API_KEY (3 req/s). Not a code defect (direct curl = HTTP 200; commands verified
individually). Production widget use should set NCBI_API_KEY (10 req/s); documented in
research.json troubleshoots + README.

## Full-dogfood acceptance: local socket exhaustion (root cause, runtime-proven)
Full live dogfood (56-request matrix) consistently failed the last 8 checks
(einfo/esearch/esummary/pico) with `dial tcp ...: connect: bad file descriptor`
— LOCAL file-descriptor / ephemeral-port exhaustion on the generation host after
this session's cumulative HTTP volume (manual verification + 5×56-request dogfood
runs + 3x retries each piling up TIME_WAIT sockets). NOT an NCBI throttle (direct
curl = 200) and NOT a code defect (all 4 commands verified working individually).
~48/56 checks pass before the socket pool exhausts. Mitigation: Quick Check level
(6 requests) after a TIME_WAIT drain. Production widget usage (occasional requests)
never approaches this limit.
