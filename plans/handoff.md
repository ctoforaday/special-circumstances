# FEOV handoff / memento — checkpoint 2026-07-21

You are resuming work on **frank-exchange-of-views** (the research-debate engine). This file
is the pickup point. Read it, then `gh issue list` for the live queue. Verify before trusting:
several claims this session were wrong on first telling and corrected only when checked.

## State right now

- **`main`** = `90f9347` (PR #81 merged). Plugin **0.34.0**, recordToolVersion/cli.Version **0.9.0**.
- **In flight: PR #82** on branch `feat/feov-report-structure-v2` — GREEN, ready to merge. Bumps
  plugin→**0.35.0**, recordToolVersion/cli.Version→**0.10.0**. The deterministic structural report
  boxes: **#75** "## Read this first" orientation (open gaps ranked from the board + the bench's
  `certify`/`halt` voice promoted); **#65/#79** blue-embed dedup (embed carries only blue's
  non-composed remainder — lifted + tool-owned sections dropped, killing the lift-AND-embed
  duplication and the stale-verdict contradiction; **footnotes KEPT** as blue's citations) + the
  prompt now forbids blue a verdict line; the **#77 un-minted-findings slice** (lens findings
  credited by no gap's `found_by` are surfaced). Verified against the 2026-07-20 record.
- **DECIDED this session (don't re-litigate):** box 5 (#66/#77 per-gap debate THREAD — seats emit
  `position`/`closing`/`dispute`/`opinion`, render by gap) is **DEFERRED to post-run**. Leaf-check
  that forced it: `debate()` composes from event types the real record NEVER emits (zero
  position/closing/dispute/opinion/certify/halt); only `finding`/`mint`/`avenue` carry voice. So
  box 5 is prompt-wiring-FIRST and only a run can verify it. Un-minted findings: SURFACE (done).
- **PRIOR: PR #81** (merged) — report fixes (#74/#76) + the `--reason` vocabulary consolidation.
- **Released:** git tag `frank-exchange-of-views--v0.33.0` (6 platform binaries; a **v0.35.0** tag
  is owed after #82 merges). `sc-doctor -fix` installed feov-record into the plugin cache. Env is
  **READY** (qlty/jq on PATH; gcc still off PATH — see [[tools-installed-but-off-path]]).
- No dry run has been done with the **current** engine code yet. The last real run was the
  2026-07-20 haiku smoke (`research/2026-07-20_record-model-soundness/`) — it ran code BEFORE
  #65/#67/#74/#76/#81, and it is the source of most open bugs below.

## The queue IS GitHub issues (single queue). issue=tracker+state, `plans/*.md`=design, PR diff=line review. NEVER close a bug until a RUN confirms it — see [[bug-state-tracking]].

State labels: `state:in-progress` / `needs-verify` (fixed+merged, awaiting a run) / `triage`.

**Immediate next work:**
1. **Merge PR #82** (user's call), then `/plugin update` + `/reload-plugins` (cache is
   version-gated — it has the binary but stale skill/prompt content). Tag `v0.35.0` is owed.
2. **THE RUN.** The deterministic report boxes are done (#75 / #65 / #79 / #77-slice, all in #82).
   A run now (a) verifies them at the leaf → flips #65/#75/#77/#79 to closed, and (b) is the
   PREREQUISITE for box 5 — you cannot render a debate thread the seats never evented.
- **DONE in #82 (was box 3/4):** #75 orientation layer + #65/#79 embed dedup + prompt verdict-forbid
  + #77 un-minted-findings slice. `needs-verify` on merge; close only after the run.
- **DEFERRED to post-run (box 5, #66/#77 core):** the per-gap **argument thread**. The record
  carries ZERO `position`/`closing`/`dispute`/`opinion`/`certify`/`halt` events — `debate()`
  renders event types the seats never emit. So this is prompt-wiring-FIRST (make seats emit onto
  the record), render-second (regroup BY GAP: mint.problem → dispute → dispute-respond → closing
  → opinion → close). `position` (round-level, no gap_id) is the wrong abstraction — skip it.
  Let the run reveal exactly what the thread needs before investing.

**Also open:** #62 (record = the ONLY inter-agent channel — retire candidates/debate.md/ledger
files; lenses emit findings; merge de-editorializes via `supersedes`; the umbrella — carries a
pinned parallel-safety constraint re the unguarded MintGapID counter), #63 (model tiers —
`plans/model-tier-flags.md`), #64 (red constitution: local pinned artifacts are primary source),
#68 (event schema versioning), #70 (compute claim_count in the tool — highest-value net-new),
#71 (citations_checked/finding-id derive), #72 (lens slice verb), #73 (agent .md files cite
retired paths), #80 (test-coverage audit — green suites over unprotected invariants; note the
prompt goldens do NOT cover the `recordClause` binDir branch). **needs-verify (fixed, run to
confirm):** #67, #74, #76, #78 (in #81); **#65, #75, #77(slice), #79 (in #82 — flip to
needs-verify on merge).**

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
