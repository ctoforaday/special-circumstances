# The record tool: a record-management library + bespoke per-seat CLIs

Rev 5 (2026-07-18) — post plan-audit FAIL #3 (seven gaps: prose-verb dedup, shadow renders, render inventory, CHANGELOG fate, bench friction, setup accounting, render locus). Rev 4 was post FAIL #2 (six gaps, two empirically grounded
in the run-5 corpus: 8 duplicate seat dispatches in 58 starts; transcripts carry
no seat label). Rev 3 was post FAIL #1 (eight gaps). Origin:
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

## II. Technical Context & Design

SEAT IDENTITY (audit-2 gaps 1-3, grounded in the measured duplicate-dispatch
anomaly — 8 of run 5's 50 cache keys dispatched twice): the ENGINE assigns and
the PROMPT carries an explicit `SEAT_ID` — `<seat>-r<round>[-L<lens>]` (e.g.
`red-lens-r4-L3`, `blue-respond-r2`) — one line added to every seat prompt.
Every CLI invocation requires `--seat-id`; the tool suffixes the SHARD FILE with
a per-process nonce generated at first invocation
(`events-<seatId>-<nonce>.jsonl`), so a re-dispatched duplicate of the same seat
writes its OWN shard — uniqueness holds by process, not by name. Side benefit:
transcripts become self-identifying (the SEAT_ID is in the first user message
AND in every tool command), retiring the dashboard's regex classification.

STORAGE — per-process shards, contention-free by construction: O_APPEND, one
JSON line per event, line at most atomic-write size; prose payloads over 2KB go
via `--file <path>` or stdin — never inline shell args, per the documented
heredoc-mangling recurrence (which bit AGAIN while authoring this plan). seq is
PER-SHARD monotonic (re-read tail on append). Global order is a render-time
merge by (round, seatId, nonce, seq) — deterministic, no locks. DEDUP SEMANTICS
(audit-3 gap 1 — content hashing fails for prose verbs, where duplicate seats
write DIFFERENT text): keys are STRUCTURAL, never content hashes.
Singleton-per-round verbs (position, revision, verdict, spot-check) key on
`seatId + round + verb`; multi-instance verbs key on their stable labels
(closing → gap_id; finding → L-label; cite → reference; dispose → observation
id; manifest-row → gap_id; opinion → gap_id). Friction and observe key on
`seatId + round + verb + ordinal`. WINNER SELECTION under nonce multiplicity:
the nonce whose shard carries that seat's TERMINAL event (verdict / revision /
the last verb of the seat's contract) wins; ties fall to latest-mtime; EVERY
multi-nonce seatId is listed in the render's anomaly footer AND flagged by the
join audit — duplicate dispatch is never silently normalized, it is dedup'd
AND reported. Lens labels restart each round; keys are round-qualified through
seatId by construction, with a same-label-next-round collision test in R1's
suite. A torn round (mint without close) renders as open state; nothing is
lost, nothing blocks.

RENDER LOCUS (audit-3 gap 7, dissolving ordering contracts): every MUTATING
verb on every CLI triggers a re-render of its affected projections as a side
effect (local node process, zero tokens — cost cleared by audit-3), and every
CLI exposes read-only `render`. Readers therefore never depend on another
seat's verb order: the projection is current as of the last mutation, and the
bench's prompt FIRST ACTION is `bench.mjs render` as a belt.

LOSS DISCIPLINE — the log lives on the untracked live blackboard, so the W1.13
incident classes (add -A / checkout / stash) threaten it: (a) `red-merge.mjs
verdict` runs an automatic `checkpoint` — mirrors `records/` to
`~/.cache/feov/run-mirror/<runDirHash>/` (user cache, NOT OS temp — temp purge
must not void the sole recovery path) every round; recovery = copy back.
MIRROR LIFECYCLE: created at first checkpoint, refreshed each round, DELETED by
capture after `records/` is committed; mirrors older than 30 days are purged by
the next setup run (orphan cleanup for crashed runs). (b) capture commits
`records/` with the run record — git-tracked from then on; a post-capture
copy-back is impossible by construction (the mirror is gone). (c) the
freeze-guard warning classes stand. Window: at most one round's events —
BETTER than today, where a sweep loses every untracked round at once.

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
  W1.8 duty — moved here from lens; lenses report, the merge records), POSITION
  (the round's ### RED section) and CLOSING (### RED CLOSING entries — prose via
  --file), verdict (renders boards + telemetry, then runs checkpoint), friction,
  PETITION.
- blue.mjs: revision (the round-record event), manifest-row, dispute,
  confidence, POSITION (### BLUE) and CLOSING (### BLUE CLOSING), friction,
  PETITION.
- DEBATE.MD IS A PROJECTION (audit-2 gap 5): round sections assemble from
  position/closing/opinion events at render — otherwise it remains the central
  hand-maintained record the plan claims to replace. The judge's confined
  ruling basis is unchanged in content (closings + transcript + artifacts);
  only the transcript's write path changes. Any seat may invoke read-only
  `render` on demand; blue's `revision` triggers one automatically so
  projections are current for the next reader mid-round (audit-2 gap 4's
  render-locus requirement).
- bench.mjs: opinion (disposition/principle/tension/review-flag as required
  args), petition-rule, halt, certify, FRICTION (audit-3 gap 5 — judge-r,
  judge-terminal, and assemble are all bench seats whose friction previously
  rode the file dual-write; post-R3 the event IS the abort-surviving copy).
  No mint — the bench rules, never originates. The ASSEMBLE seat uses bench.mjs.
- CHANGELOG.MD IS A PROJECTION of revision events (audit-3 gap 4): `revision
  --file` carries the round's entry body; `render changelog` assembles it. The
  round_record_appended attestation is REDEFINED at R3 to attest "the revision
  event was emitted and the rendered CHANGELOG carries this round" (same in-run
  envelope gate, new referent); the citation-ledger re-fetch trigger reads the
  RENDERED changelog — reader unchanged.
- RENDER INVENTORY (audit-3 gap 3, complete): R1 renders ledger.md, archive.md,
  board-telemetry.jsonl; R2 adds debate.md, CHANGELOG.md, citation-ledger.md
  (all three needed for the parity gate); R4 adds scorecards + the judicial
  record. NOTHING cuts over at R3 without appearing in the R2.5 parity set.
- Capture + dashboard consume the library read-only. `--help` on each CLI is the
  seat's record contract (prompt prose retires to it).

ENFORCEMENT — stated at its true tier (audit-1 gap 7): INTERFACE (no sanctioned
path to out-of-role verbs) + DETERRENCE + POST-HOC JOIN AUDIT, per the
attestation ceiling. The join audit's DATA SOURCE is the SEAT_ID itself
(audit-2 gap 3 — transcripts carry no label field, verified): every tool
command embeds `--seat-id`, so each transcript self-identifies through its own
Bash calls, and each shard names the seatId+nonce that wrote it. The join:
for every event, find a transcript whose commands carry that shard's seatId and
the event's verb + idempotency key (payloads via --file match on key, not
text). Events with no matching invocation anywhere, or one seatId claimed by
commands in two transcripts with distinct nonces both ACTIVE in the same round
(beyond the known duplicate-dispatch pattern, which dedups at render), are
FLAGGED. Bulk back-fill (hand-writing boards then transcribing at round
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
   unchanged. SHADOW RENDERS (audit-3 gap 2): during dual-mode, ALL renders
   write to `<runDir>/records/render-shadow/` — the hand-written artifacts are
   never touched, so the parity gate compares hand vs shadow, never
   render-vs-render. [MODIFY] capture-research-run.mjs: add record-parity check
   and record-join audit. [MODIFY] setup-research-run.mjs (audit-3 gap 6):
   creates records/, runs the 30-day mirror orphan purge. No deletions yet.
R2.5 PARITY RUN (the one-run parallel period): the first live run post-R2 runs
   dual-mode; capture executes [NEW] scripts/record-parity-check.mjs —
   normalizes the hand-written ledger.md / archive.md / board-telemetry.jsonl /
   debate.md / CHANGELOG.md / citation-ledger.md (whitespace-collapsed,
   section-order-insensitive, id-sorted rows) against the SHADOW renders;
   FAIL = any divergence in gap ids, grades, closure classes, dispositions,
   counts, telemetry fields, round-section presence, or ledger rows; prose
   bodies compare presence-not-text. Verdict lines land in run-record-audit.md.
   Zero-FAIL is the gate to R3, at which point renders switch from shadow to
   the real paths.
R3 [MODIFY] debate.js: prompts shrink to tool contracts; RETIRE the
   count-consistency throw at the merge (counts are renders) and the
   ledger/archive/telemetry prompt paragraphs; KEEP — corrected per audit-2
   gap 4 — the round_record_appended and manifest envelope attestations AS
   IN-RUN GATES (the engine has no filesystem and events are post-hoc to it;
   these envelope throws are the only mid-round desync brakes and they are
   cheap — the events ADD the mechanical recount at capture, they do not
   replace the brake), plus the spot-check floor, the null-guards, the lane
   floor, and the dispute machinery. [MODIFY] capture: RETIRE the shards audit, friction-parity +
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

## Appendix: directory tree (formal-deviation fix)

    plugins/frank-exchange-of-views/skills/research-protocol/scripts/
      lib/record.mjs            [NEW R1]  append/validate/mint/replay/merge/render
      tools/red-lens.mjs        [NEW R1]
      tools/red-merge.mjs       [NEW R1]
      tools/blue.mjs            [NEW R2]
      tools/bench.mjs           [NEW R2]
      record-parity-check.mjs   [NEW R2]
      debate.js                 [MODIFY R2 dual-mode; R3 shrink]
      setup-research-run.mjs    [MODIFY R2: records/ + mirror orphan purge]
      capture-research-run.mjs  [MODIFY R2 +join/parity; R3 audit collapse]
    plugins/frank-exchange-of-views/tests/simulator/record.test.mjs  [NEW R1]
    <runDir>/records/events-<seatId>-<nonce>.jsonl   (per-process shards)
    <runDir>/records/registry-extensions.jsonl        (run-local --class-new)
    ~/.cache/feov/run-mirror/<runDirHash>/            (checkpoint mirror, capture-deleted)

## Plan-audit disposition (sitting 4, 2026-07-18): PASS with notes — all folded

Three sittings FAILed revs 2-4 (21 specification defects caught pre-implementation,
two grounded empirically in the run-5 corpus). Sitting 4: PASS. Notes, binding on R1:
1. ATOMIC RENDERS: record.mjs renders write temp + atomic-rename — concurrent lens
   mutations rewrite shared projections (cite -> citation-ledger.md) and
   "current as of last mutation" is false under the race otherwise. Last-writer-wins
   acceptable: every render is full-state, self-healing on next mutation.
2. Winner selection: the neither-nonce-has-terminal case falls EXPLICITLY to the
   latest-mtime tiebreak; R1's suite gains a multi-nonce winner-selection fixture
   (the 8/50 duplicate-dispatch anomaly gets its named test).
3. The dashboard-regex retirement claim is DESCOPED from this plan (inert here —
   render-run-dashboard.mjs and cost-audit.mjs are not in the modify inventory);
   scheduled instead as a small follow-up once SEAT_ID ships in transcripts.

## Amendment (operator challenge, 2026-07-18): Go for seat-side runtime; markdown never parsed

Two foundational reversals, both conceded on the merits:

1. LANGUAGE. The "node is already installed / shares code with debate.js"
   rationale is DEAD: debate.js is sandboxed (zero imports — sharing is
   impossible), and the record tool grew from glue into SEAT-SIDE RUNTIME
   (per-seat invocation frequency, locks, atomicity, role enforcement) where
   the mjs constraint set forced hand-rolled locks, untyped schemas, and
   un-race-tested concurrency. Doctrine line (matching PC's hooks precedent):
   LEAD-SIDE GLUE = mjs (setup/capture/dashboard/parity: run-per-run
   orchestration); SEAT-SIDE RUNTIME = compiled Go. The record CLIs move to Go:
   ONE binary, roles as subcommands (feov-record lens|merge|blue|bench <verb>),
   which strengthens role boundaries (a namespace refuses a --seat-id outside
   its role prefix — impossible with separate loose scripts), ships as one CI
   release asset, uses real file locks and typed event structs, and gets
   `go test -race` on the actual concurrency. FEOV owns its own tools/ Go
   module (W1.15 each-plugin-owns-its-deps) riding the existing CI release +
   doctor-bootstrap machinery; the empty-bin window is covered by the same
   guard pattern PC shipped.
2. MARKDOWN. JSON is the record AND the machine interface; markdown is a
   WRITE-ONLY reading view that nothing ever parses again. Renders emit
   board.json (versioned, structured — gaps, closures, observations,
   anomalies) beside the md views; dashboard, capture audits, and the parity
   checker's shadow side consume events or board.json. The only surviving
   markdown parsing is the parity checker's HAND side during dual-mode — it
   dies with the hand-written artifacts at R3.

MIGRATION — the mjs implementation is DEMOTED TO ORACLE, not discarded: its 29
tests encode the audited semantics, and the Go port is validated by
DIFFERENTIAL TESTING (same command sequences through both → byte-identical
events and renders) before the mjs write path retires. mjs consumers keep
reading the language-neutral JSONL directly — the format was the point.
Revised sequence: R2 (PR #35) merges as ORACLE + parity harness, not for seat
adoption; R2g = the Go port with differential suite; dual-mode prompts point
at the binary (toolsDir becomes binDir) only after R2g's differential gate is
green; R2.5 parity run follows on the Go tool.
