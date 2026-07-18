# The record tool: a record-management library + bespoke per-seat CLIs

Rev 3 (2026-07-18) — post plan-audit FAIL #1; all eight gaps addressed. Origin:
the E0.5 convergence (four audits acquitted the seats, indicted the record) +
the operator's proposal (tooling for lossless records, automatic metrics,
structured tagging, constitutional deletion) + the per-seat refinement (verb
sets ARE role boundaries).

## I. Summary & Goals

Replace hand-maintained run records with an append-only EVENT LOG owned by a
shared library (`lib/record.mjs`) and written only through bespoke per-seat CLIs
whose verb sets encode each chair's role. Markdown boards, telemetry, scorecards,
and the judicial record become PROJECTIONS (`render`). Goals: (1) metrics as
projections, never self-reports; (2) losslessness by construction at the
sanctioned path, with a stated loss-window discipline for the unsanctioned ones;
(3) structured tagging (class/grade/lineage/anchor) validated at append;
(4) measured problem classes prevented structurally; (5) constitutional
DELETION of record mechanics — judgment stays, procedure leaves. Non-goals:
verifying claim truth (vacuity remains post-hoc behavior audit); replacing
prose payloads (the tool structures envelopes, never arguments).

## II. Design

STORAGE — per-seat shards, contention-free by construction: each seat appends
ONLY to `<runDir>/records/events-<seatId>.jsonl` (O_APPEND, one JSON line per
event, line at most atomic-write size; prose payloads over 2KB go via
`--file <path>` or stdin — never inline shell args, per the documented
heredoc-mangling recurrence, which bit AGAIN while writing this very plan).
seq is PER-SHARD monotonic; global order is a render-time merge by
(round, seat, seq) — deterministic, no locks, 6 parallel lenses cannot race.
Every mint-class verb takes `--idempotency-key` (seat + stable local label); a
duplicate key returns the existing id instead of double-minting — crash-retry
safe. A torn round (mint without close) renders as open state; nothing is lost,
nothing blocks.

LOSS DISCIPLINE — the log lives on the untracked live blackboard, so the W1.13
incident classes (add -A / checkout / stash) threaten it: (a) `red-merge.mjs
verdict` runs an automatic `checkpoint` — mirrors `records/` to an out-of-repo
session location (OS temp keyed by runDir hash) every round; recovery = copy
back; (b) capture commits `records/` with the run record (git-tracked from then
on); (c) the freeze-guard warning classes stand. Window: at most one round's
events — the same exposure as today, now with a stated recovery procedure.

VERB SETS (complete against everything the engine currently records — audit gap
3 closed):
- red-lens.mjs: finding (auto L*-F* ids), observe (note / checked-held), CITE
  (citation-ledger event: claim, reference, confidence, round, access-date —
  the cross-round re-fetch gate), friction, PETITION.
- red-merge.mjs: mint (--class validated; --class-new SLUG requires
  --definition --neighbor EXISTING --distinguisher — appended as a run-local
  registry-extension event, promoted to the seed registry only at post-run human
  review, two-tier like law), close (anchors required; --carried-from for
  re-attestations), dispose (every lens observation demands one), regrade
  (--basis), DISPUTE-RESPOND (accepted/rejected + rationale), SPOT-CHECK (the
  W1.8 duty — moved here from lens; lenses report, the merge records), verdict
  (renders boards + telemetry, then runs checkpoint), friction, PETITION.
- blue.mjs: revision (the round-record event), manifest-row, dispute,
  confidence, friction, PETITION.
- bench.mjs: opinion (disposition/principle/tension/review-flag as required
  args), petition-rule, halt, certify. No mint — the bench rules, never
  originates.
- Capture + dashboard consume the library read-only. `--help` on each CLI is the
  seat's record contract (prompt prose retires to it).

ENFORCEMENT — stated at its true tier (audit gap 7): INTERFACE (no sanctioned
path to out-of-role verbs) + DETERRENCE + POST-HOC JOIN AUDIT, per the
attestation ceiling. The join audit is specified: capture check `record-join` —
for every event, find the emitting seat's transcript tool_use whose Bash command
invokes the event's CLI verb with the event's id or idempotency-key (payloads
via --file match on key, not text); events with no matching invocation, or
invocations from a transcript whose seat differs from the event's claimed seat,
are FLAGGED. Bulk back-fill (hand-writing boards then transcribing at round
end) is visible as a tail-clustered invocation pattern and flagged as a
parity-vacuity WARN for human review. Hand-appended well-formed lines are
caught by the same join. "Physically impossible" applies only to the sanctioned
path; a PC-side seat-to-tool PreToolUse guard is a deferred hardening rung,
priced on first observed misuse.

RETAINED JUDGMENT CLAUSES (audit gap 8 — load the tool does not replace):
(a) the near-match / reopen-vs-new judgment stays in red's constitution — the
tool validates id EXISTENCE; deciding whether a candidate is a reopen remains
judgment (the tool assists: `mint` prints closure-index near-matches for the
gap's location/class before committing); (b) the recompute-or-cite gate stays
in the bench's constitution for any ad-hoc aggregate no projection carries —
narrowed, not deleted: standard aggregates come from renders; inventing a
number outside them requires naming and recomputing its source.

## III. Implementation plan (file-by-file)

R1 [NEW] plugins/frank-exchange-of-views/skills/research-protocol/scripts/lib/record.mjs
   (append / validate / id-mint / replay / shard-merge / render: ledger,
   archive, telemetry); [NEW] scripts/tools/red-lens.mjs and red-merge.mjs;
   [NEW] tests/simulator/record.test.mjs (append atomicity, shard-merge
   determinism property test, idempotency, dangling-supersedes refusal, class
   validation incl. --class-new, render fixtures).
R2 [NEW] scripts/tools/blue.mjs, bench.mjs; [MODIFY] debate.js — DUAL-MODE:
   seats write boards AND events (prompts instruct both); engine checks
   unchanged. [MODIFY] capture-research-run.mjs: add record-parity check and
   record-join audit. No deletions yet.
R2.5 PARITY RUN (the one-run parallel period, scheduled here — audit gap 2):
   the first live run post-R2 runs dual-mode; capture executes
   [NEW] scripts/record-parity-check.mjs — normalizes the hand-written
   ledger.md / archive.md / board-telemetry.jsonl (whitespace-collapsed,
   section-order-insensitive, id-sorted rows) against fresh renders; FAIL = any
   divergence in gap ids, grades, closure classes, dispositions, counts, or
   telemetry fields; prose bodies compare presence-not-text. Verdict lines land
   in run-record-audit.md. Zero-FAIL is the gate to R3.
R3 [MODIFY] debate.js: prompts shrink to tool contracts; RETIRE
   round_record_appended (superseded by revision events), the empty-manifest
   throw (superseded by a manifest-row event check), the count-consistency
   throw at the merge (counts are renders), and the ledger/archive/telemetry
   prompt paragraphs; KEEP the spot-check floor (now reading spot-check events
   via the envelope summary), the null-guards, the lane floor, the dispute
   machinery. [MODIFY] capture: RETIRE the shards audit, friction-parity +
   harvest (friction is events), and the W1.7-form record-parity; KEEP
   telemetry-presence (as render-match), context-use, assembly-screen, and
   record-join. [MODIFY] the three constitutions: the deletion list minus the
   two retained judgment clauses; every deleted paragraph ships in the same
   diff as the tool clause replacing it.
R4 [MODIFY] lib/record.mjs: live class registry + within-run recurrence
   escalator (recalibrated per the seed data: counts reset when a class reaches
   zero open instances; cross-run class pressure routes to craft memory, never
   the docket) + scorecard and judicial-record renders (absorbs W2h and W2c's
   record surfaces).

## IV. Risk & Mitigation (likelihood x impact x complexity-to-mitigate)

- Shard-merge ordering bug renders a wrong board: med x high x low — the merge
  is a pure function, property-tested (same events in any shard arrival order
  render identically).
- Loss window (untracked log swept by an incident-class git op): low-med x
  high x low — per-round out-of-repo checkpoint + freeze-guard warnings + the
  named recovery procedure.
- Seat bypass (hand-appended events): low x med x low — record-join audit +
  interface deterrence; PC guard deferred until first observed misuse.
- Migration divergence (tool record vs prose record during dual-mode): med x
  med x low — the parity gate exists for exactly this; R3 is blocked on
  zero-FAIL.
- Back-fill vacuity (transcribe-at-end defeats parity): med x med x med — the
  tail-cluster detector in record-join, WARN tier, human-reviewed.
- Tool-arg mangling on prose payloads: med (documented recurrence, reproduced
  during this plan's own authoring) x low x low — --file/stdin mandatory over
  2KB.
- Registry poisoning via --class-new: low x med x low — run-local until human
  promotion (two-tier authority).
- Prompt/tool-help drift: med x low x low — help text lives beside the code; a
  simulator test asserts every exposed verb appears in its own help.

## V. Verification plan

- `node --test plugins/frank-exchange-of-views/tests/simulator/record.test.mjs`
  (R1 gates: atomicity, merge-determinism property test, idempotency,
  validation refusals, render fixtures) and the full simulator suite green.
- `node plugins/frank-exchange-of-views/skills/research-protocol/scripts/record-parity-check.mjs <runDir>`
  — spec in III/R2.5; exit 2 on divergence; runs at the parity run's capture.
- The record-join audit runs in every capture from R2 on; its verdict line
  lands in run-record-audit.md (PASS / FLAGGED list / back-fill WARN).
- Smoke (`/research --smoke`) after R2 and after R3: one round exercising every
  seat CLI end-to-end; capture audits green.
- The R3 constitutional-deletion PR diff IS the review surface: every deleted
  paragraph must show its replacing tool clause in the same diff (reviewer
  checklist in the PR body).
