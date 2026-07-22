# FEOV handoff / memento — checkpoint 2026-07-21

You are resuming work on **frank-exchange-of-views** (the research-debate engine). This file
is the pickup point. Read it, then `gh issue list` for the live queue. Verify before trusting:
several claims this session were wrong on first telling and corrected only when checked.

## State right now

- **`main`** = `8b6148b` (PR #69 merged). Plugin **0.33.0**, recordToolVersion/cli.Version **0.8.0**.
- **In flight: PR #81** on branch `feat/feov-report-structure` — GREEN, ready to merge. Bumps
  plugin→**0.34.0**, recordToolVersion/cli.Version→**0.9.0**. Contains: report-composer fixes
  (#74 case-insensitive lift, #76 risk-matrix distill) + the **`--reason` vocabulary
  consolidation** (one required prose field; `--file/--text/--comment/--basis` retired).
- **Released:** git tag `frank-exchange-of-views--v0.33.0` (6 platform binaries). `sc-doctor -fix`
  installed feov-record from it into the plugin cache. Env is **READY** (qlty/jq now on PATH; gcc
  still off PATH — see [[tools-installed-but-off-path]]).
- No dry run has been done with the **current** engine code yet. The last real run was the
  2026-07-20 haiku smoke (`research/2026-07-20_record-model-soundness/`) — it ran code BEFORE
  #65/#67/#74/#76/#81, and it is the source of most open bugs below.

## The queue IS GitHub issues (single queue). issue=tracker+state, `plans/*.md`=design, PR diff=line review. NEVER close a bug until a RUN confirms it — see [[bug-state-tracking]].

State labels: `state:in-progress` / `needs-verify` (fixed+merged, awaiting a run) / `triage`.

**Immediate next work — the structural report boxes the user wants BEFORE the next run:**
- **#75** (box 3) — synthesis/orientation layer: surface the `certify` event as a top "read
  this first", rank open gaps, outcome framing — composed from the record, NOT re-authored.
- **#65** (box 4) — kill lift-AND-embed duplication in `assemble.go` (blue's sections are
  lifted and then re-embedded via "## Blue team report (in full)"). Union-safe: embed only
  the non-lifted remainder.
- **#66 / #77** (box 5) — the per-gap **argument thread** (the tool's "debate" section is
  empty). KEY INSIGHT: the argument is already atomic to the gap — `closing --id <gap>` exists,
  is wired to blue+merge, now requires `--reason`. So this is a RENDER + prompt-wiring job
  (regroup the debate BY GAP: mint.problem → manifest-row → dispute → dispute-respond → closing
  → opinion → close), not a schema change. `position` (round-level, no gap_id) is the wrong
  abstraction — skip it.

**Also open:** #62 (record = the ONLY inter-agent channel — retire candidates/debate.md/ledger
files; lenses emit findings; merge de-editorializes via `supersedes`; the umbrella — carries a
pinned parallel-safety constraint re the unguarded MintGapID counter), #63 (model tiers —
`plans/model-tier-flags.md`), #64 (red constitution: local pinned artifacts are primary source),
#68 (event schema versioning), #70 (compute claim_count in the tool — highest-value net-new),
#71 (citations_checked/finding-id derive), #72 (lens slice verb), #73 (agent .md files cite
retired paths), #79 (verdict contradiction — partly #65), #80 (test-coverage audit — green suites
over unprotected invariants; note the prompt goldens do NOT cover the `recordClause` binDir
branch). **needs-verify (fixed, run to confirm):** #67, #74, #76, #78.

## Design decisions settled this session (don't re-litigate)

- **The report is assembled from the RECORD** (event log), not projection `.md` files or seat
  prose. Blue authors only its own audited surfaces; the tool composes the rest.
- **One prose field, `--reason`**, required on claim/ruling acts. The "why" is atomic to the act,
  on the public record. No junk-drawer free-text (`--comment` retired entirely).
- **Storage:** keep JSONL now; **embedded SQLite is the leading migration target, coupled to
  #62's concurrency** (a SQL transaction fixes the MintGapID race for free). Reject
  Postgres/frameworks. `plans/storage.md`. Indexing is the real driver.
- **Model tiers** by recoverable-error (cheap) vs unrecoverable-absence (capability-bound):
  construction (blue lanes/synthesize, red L5/L6) = big; lookup (red L1–L4, blue-respond) = cheap.

## Pitfalls / lessons (earned this session)
- **Verify delegated subagent work at the leaf** — a subagent reported "all green" while a
  compile error was live; it was stale, but the check is non-negotiable. Never trust a report
  over `go build` / a driveable check.
- **Bump cli.Version BEFORE regenerating stamp goldens**, run `-count=1` (cached `go test` hid a
  stale-stamp failure once — the #57 lesson).
- **To debug branch code without merging:** build feov-record from `tools/` into a binDir (Git
  Bash resolves the extensionless name to `.exe`), run the WORKING-TREE `setup-research-run.mjs`
  + Workflow with `scriptPath`=working-tree `debate.js`, `--smoke` (lanes 1, maxRounds 2, haiku).
  Recipe detail in [[feov-projection-retirement-queued]].
- **Before a normal `/research` run with the new code:** merge #81, then `/plugin update` +
  `/reload-plugins` (cache is version-gated; it has the binary but stale skill/prompt content).

## Sequencing the user has signalled
Fix all structural report gaps (#75/#65/#66-77) up front, THEN a larger run, THEN a
`/deep-research` "what good research looks like" comparative sweep (a model can be pinned on it).
No point running the comparison while known report gaps remain.
