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
MIRROR LIFECYCLE: created at first checkpoint, refreshed each round, and reaped
by AGE — mirrors untouched for 30 days are removed, by `run-setup` and by
`capture`. A live run rewrites its mirror every round, so what crosses that line
stopped writing weeks ago; the reap is orphan cleanup for crashed runs and
nothing else. (b) capture writes `records/` to `run-archive/<slug>.tar.gz`, and
a HUMAN commits it — capture invokes no git at all. So a post-capture copy-back
is still possible, and deliberately: at the moment capture returns, the archive
is an UNTRACKED file in the working tree, exposed to the same add -A / checkout
/ stash classes this paragraph opens with, and the mirror is still the recovery
path until someone commits. (c) the freeze-guard warning classes stand.

CORRECTED 2026-08-29. This paragraph read "DELETED by capture after `records/`
is committed", and neither half was ever true: capture does not commit, and
capture did not delete. Written as a completed design, it was read as one — the
stage looked implemented, and the actual reaper (`PurgeStaleMirrors`) was
reachable only from `run-setup`, which nobody runs between research runs, so a
crashed run's mirror sat until someone happened to start a new one. Building the
delete as specified would have removed the recovery path at the moment its
replacement was least durable. The reap was widened to `capture` instead. Window: at most one round's events —
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
  **RE-GROUNDED 2026-08-15 (#223): the detector survives, its host does not.**
  record-join was deleted — five independent ways to be wrong, one of which
  fabricated invocations and could mask a real orphan. Back-fill is now
  `BackfillAudit`, measured from the record's own `ts` (register → first write
  → last write) and needing no transcript at all. Different instrument, same
  phenomenon: the old one measured POSITION among a seat's tool calls, this one
  measures ELAPSED TIME, and neither can see what the other can.
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
- ~~The record-join audit runs in every capture from R2 on; its verdict line
  lands in run-record-audit.md (PASS / FLAGGED list / back-fill WARN).~~
  **RETIRED 2026-08-15 (#223).** Seat bypass by hand-appended events was the
  threat this answered; it was never the threat it could answer, because
  `feov-record` is the sole validated writer and an event exists only because
  the command ran. What replaced it in the audit list is `backfill`, which
  keeps the WARN tier and the verdict line.
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
    ~/.cache/feov/run-mirror/<runDirHash>/            (checkpoint mirror, age-reaped at 30d)

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

## Amendment (operator, 2026-07-18): lock-file defense on shared surfaces

The shards are single-writer, but two surfaces are shared and now LOCKED
(dependency-free — mkdir as the atomic primitive every lockfile package wraps):
the per-seat pointer (racing registers) and the PROJECTIONS (concurrent
render-on-mutation from parallel lenses; Windows rename-over-existing throws
under contention rather than last-writer-wins). Lock: mkdir-acquire, 10s
stale-steal by mtime (crashed holders never deadlock a seat), 5s bounded wait
then proceed-unlocked (a lost render self-heals on the next mutation — full-
state projections make this safe), retry-wrapped renames for the antivirus/
indexer EPERM class. Tested: 6 truly parallel lens processes with per-verb
renders (all events land, no lock or temp leaks, final render complete) and
stale-lock stealing inside the wait bound.

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

## R2g design notes (thought through 2026-07-18; execution deferred)

FAITHFUL-FIRST, IMPROVE-SECOND: the differential gate only means something if
the port is semantically exact. R2g.1 = faithful port, gate green; R2g.2 =
improvements as separate commits (each visible in the diff, each tested):
role/seat-id cross-check per subcommand namespace, board.json render,
tool_version stamped on register events, event schema_version field.

DIFFERENTIAL COMPARISON POLICY (the trap that would eat a naive port):
byte-identical output is IMPOSSIBLE naively — nonces are random, JS object-key
order != Go struct marshal order, and mtime drives winner selection. Policy:
(1) both implementations accept a test-only nonce seed (env var) OR the
harness normalizes nonces to canonical placeholders on both sides;
(2) EVENTS compare by structural (parsed deep-) equality, never bytes;
(3) RENDERS compare byte-identical AFTER nonce normalization — this is where
JS/Go number-formatting drift (String(46) vs strconv, toFixed vs %.2f) gets
caught, which is precisely the E0.5f aggregate-drift class at the language
boundary; (4) mtime-dependent tests control mtimes explicitly on both sides;
(5) lock artifacts are excluded from comparison (implementation detail).
HARNESS: the 29 oracle tests' scenarios exported as replayable command lists,
plus a random-valid-verb-sequence generator (differential fuzzing — cheap in
Go, and the highest-value test class for a port).

LOCKING: keep the mkdir-lock ALGORITHM in Go (portable Windows/Unix, stdlib-
only, semantics already tested) rather than flock/LockFileEx (x/sys dep,
platform fork). go test -race validates what the mjs could only spawn-test.

DISTRIBUTION: FEOV gets its own tools/ Go module (W1.15 ownership) —
plugins/frank-exchange-of-views/tools/cmd/feov-record/ — riding the existing
CI matrix; tag family frank-exchange-of-views--v* added to the release
workflow. Doctor: extend sibling aggregation to sibling BIN BOOTSTRAP (small
PC change) so --fix installs FEOV's binary too. EMPTY-BIN DEFENSE at the
right layer: setup-research-run.mjs preflights `feov-record --version`
against the plugin version BEFORE writing the run-live marker — a missing or
skewed binary fails at setup, never mid-round (the mid-run failure mode is
the one that costs a seat).

VERSION SKEW: register events carry tool_version; the never-update-mid-run
rule stands; skew becomes visible in the log rather than mysterious.

POST-PORT SIMPLIFICATIONS (R2g.2+): parity-check's shadow side reads
board.json instead of regexing rendered md; the dashboard's heuristic parses
retire against board.json (the descoped W-item returns with a real substrate);
mjs consumers (capture, dashboard) keep reading events/board.json directly —
no Go dependency for readers, which is the JSONL format paying off.

ORACLE FREEZE: at #35 merge the mjs write path is FROZEN — post-port changes
land in Go only; an intentional semantics change regenerates the oracle and
the differential together, never one side alone.

## R2g SHIPPED (2026-07-18) — and the amendments it forces

R2g.1 landed as a faithful port validated by the differential gate (20 scenarios
+ 12 fuzz sequences), which caught three port bugs no reading would have:
`${undefined}` interpolating as the literal "undefined"; JS slice() counting
UTF-16 code units where Go slices bytes; encoding/json escaping <, > and & where
JSON.stringify does not.

ORACLE RETIRED, EARLY. The plan sequenced retirement at R3, after the mjs write
path had been superseded in a live run. It went at R2g instead, because the
premise for waiting was false: the mjs tools were NEVER USED — no run directory
contains records/, and the engine's toolsDir defaulted to null throughout. An
oracle that has certified its port and never shipped has no remaining job. Its
validation is preserved in the golden transcripts, recorded while the gate was
green and verified to replay in the same tree.

BEYOND THE ORACLE (the port is no longer merely faithful):
- DURABILITY: every append and projection is write -> fsync -> rename ->
  fsync(dir). The oracle had none; a crash could leave a correctly-named empty
  projection or a zero-filled shard.
- ABORT SAFETY: writes run inside critical sections; SIGINT/SIGTERM lets
  in-flight writes finish, releases locks, then exits 130. Seats are killed
  routinely (quota aborts, timeouts, cancels), so this is the common path.
- COBRA: unknown flags are refused. `--anchor-set` for `--anchor-seat` used to
  produce an unanchored closure silently.
- FLOCK: replaces the mkdir lock, whose ten-second staleness steal could revoke a
  LIVE holder's lock and admit two writers to the critical section.
- SEAT-ROLE BINDING: seat identity is bound to its namespace. Before it, a lens
  could run `feov-record merge mint --seat-id red-lens-r1-L1` and the tool
  answered "minted R1-1" — the role boundary was a naming convention, not a
  boundary, which made the record evidence of nothing.
- HELP IS THE CONTRACT: every flag documented, and every help ends by naming the
  friction path. If a seat can see it, it may do it; if it cannot see it, the
  capability does not exist and the gap is a FINDING about the tooling, not a
  puzzle to route around by hand-writing the artifact.

## AMENDMENT — the mjs/Go boundary is re-litigated (operator, 2026-07-18)

The old doctrine line was "prompt-invoked orchestration = mjs IN GIT (no binary
bootstrap — FEOV stays tag-free); event-driven guards = Go hooks in PC." The
load-bearing clause was the parenthesis, and R2g killed it: FEOV now ships
feov-record as a CI release asset with its own tag family and doctor bootstrap.

What survives: debate.js MUST stay mjs — it is the workflow script the harness
loads and executes, a constraint rather than a preference.

What should move, and WHEN: capture-research-run, setup-research-run,
render-run-dashboard, record-parity-check and cost-audit do real parsing and are
where we have actually been bitten (the GNU-tar-C:-as-remote-host bug, the
heuristic ledger parse that failed twice live, the dashboard regexing prompts for
seat identity). They should become Go — but NOT YET, and the reason is
sequencing, not taste: the record layer exists to DELETE their parsing (capture
audits move to records/, the dashboard's heuristics retire against board.json,
parity-check's shadow side stops regexing markdown). Porting a parser we are
about to delete is the worst ordering available. Port after the record layer
shrinks them, when the port is mostly deletion.

## AMENDMENT — golden management (operator-driven)

Goldens are the contract now that the oracle is gone, and they will proliferate
(seat prompt goldens land one per seat class, and every wave edits prompts). The
risk is not diff quality but REVIEW EROSION: a wave regenerates everything, the
diff runs to hundreds of lines, it is rubber-stamped, and real drift ships inside
the noise. Countermeasures shipped in scripts/golden.mjs: one command for both
languages, a change report at update time, staleness failing in CI, and orphan
detection. Convention: a testdata change rides its OWN commit.

Libraries evaluated and declined, with the trigger stated so this is not
re-litigated from scratch: google/golden is ARCHIVED (read-only since 2022);
autogold formats through a Go AST and warns that golden formatting can change
across its own minor versions and across Go versions, which would inject
non-determinism into a gate whose entire job is detecting change; goldie is
active and well-made but its value-adds (templates, JSON/XML asserts) do not
reach us, and go-cmp already supplies the diff quality that was the real pain.
TRIGGER for revisiting: if fixture management becomes real work — many suites,
subtest-scoped fixtures, cleanup churn — goldie is the one to reach for.
