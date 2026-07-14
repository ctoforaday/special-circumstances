# Already shipped — winnow list for the efficiency investigation (run 4)

Staged input. Purpose: this run investigates the efficiency/termination levers the run-3
retrospective graded but did NOT ratify. The changes below are ALREADY BUILT AND MERGED —
do not re-recommend them; audit them only where run 4's own evidence shows a shipped fix
failing in practice. Anything absent from this list is fair game.

## Shipped between run 3 and this run

**PR #14 (merged 2026-07-13, engine + run hygiene)**
- Per-role model knobs: `model` (bulk) / `judgmentModel` (judgment seats, defaults to session tier)
- Citation ledger (`red/citation-ledger.md`) — verified citations don't un-verify
- Pre-created blackboard skeleton (write-block workaround; guard is filename-keyed, run-3 proof)
- Workflow simulator in CI (zero-token regression suite)

**PR #15 (merged 2026-07-14, FEOV 0.4.0 — the §3 engine docket)**
- Lineage-aware contested docket: `supersedes` arrays + `closures` records on the red envelope,
  whole-debate docket window, STRUCTURAL enforcement throw (a `closed_with_regression` with no
  successor naming it aborts the run) — answers R5-5's unenforced-good-faith critique
- Degenerate-FAIL guard (FAIL with zero gaps throws)
- Judge and blue-respond null-guards (skipped-agent resilience)
- `citationPasses` recomputed per round from current `claim_count` (was: computed once — the
  under-scaling defect red found live)
- Lane METHOD diversity roster + lane floor of 3 (`laneFloorOverride` for smoke runs)
- Minority-claim provenance tagging (consensus-vs-minority survives synthesis)
- Ledger drift triggers; friction-to-file from every seat (survives aborts); blue propagation
  clause (corrections propagate to ALL sites); `open_questions` given a home in assembly

**PR #16 (merged 2026-07-14, protocol + run record)**
- Corpus PINNING (`inputs/PINNED.md`; freeze pushes to cited paths mid-run) — answers the
  observer-effect / citation-drift class (R3-8, R5-2)
- Run-record capture: friction merge, `trajectories/journal.jsonl` tracked, transcripts
  tarball gitignored, `cost.md` via `scripts/cost-audit.mjs` (fixture-tested, in CI)
- `--smoke` mode; blue pre-flight self-audit (reads red's gap-pattern memory); red lineage
  discipline + lens-scoped labels (L2-F1); keeper runs OMIT `model` (row 16b)

**PR #17 (merged 2026-07-14, PC 0.6.1 — environment)**
- PDF-extraction MCPs vetted + pinned project-scope (`arxiv-latex`, `pdf-reader`) — the #1
  friction item of runs 2 and 3; protocol MUST-try clause before "unable to corroborate"
- node + uv in requirements.json (doctor reports them); gofmt gate + JSON validation in CI

**PR #18 (open at staging time — verify merged before citing as shipped)**
- qmd recall layer: `.mcp.json` server entry, `sc-recall-index` PostToolUse hook (FTS update
  on every markdown write, capability-gated), three-access-modes doctrine in research-protocol
  (retrieval for evidence / FULL READ for the document under audit / leaf-node fetch for
  verification), collection-per-corpus-root lifecycle policy

## NOT shipped — the subject matter of this run

Graded in the retrospective's §3 docket but deliberately deferred for debate: severity-floor
termination; risk-mass-proportional spend (lens count / audit scope tracking open risk mass);
the grade-dispute channel (can blue argue a likelihood/impact down — and does lone-voice
grading need best-of-N); sharded findings (open ledger vs closed archive) and the collator
stage (merge-seat cache-read structure); round-scoped audit (held: conflicts with the
full-re-read principle — needs the debate); sanctioned write path for red's living artifacts;
log()-per-transition heartbeats; doctor cross-plugin aggregation.
