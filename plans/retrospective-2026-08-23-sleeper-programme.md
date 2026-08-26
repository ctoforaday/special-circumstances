# Retrospective: the 2026-08-23 sleeper-service research programme

Performance review of the two FEOV runs merged in PR #540 (`2026-08-23_sleeper-service-plan`,
`2026-08-23_research-loop-counterparts`), from three evidence streams: the record tool's own
audits (`verify`, capture's nine checks, chair scorecards, cost.md), a full sweep of all 61
seat transcripts (33 MB, agent-by-agent), and leaf re-verification of the reports (arithmetic
re-derived, proofs re-run, anchor/footnote integrity checked). Findings are graded; every one
carries its evidence. Issues filed from this document: #589–#593; evidence added to #552.

## Headline

Both debates finished under their own power, every structural invariant holds, all 61 seats
returned envelopes, nothing spun, nothing was laundered — and the deliverables are honest
about their own debts. That is the good half, and it is real. The bad half concentrates in
three places: **the model tier the operator asked for was never served** (silent substitution,
caught only at capture); **assembly stripped the machine-checkable evidence layer** from both
final reports (every tool-placed anchor gone, proof footnotes dangling); and **two of run A's
six recorded proofs are defective as recorded** (wrong working directory — one enumerated
nothing and called it ABSENT, one measured 0 and printed a hard-coded sentence asserting the
opposite). None of the three was caught by red, the bench, or the in-run gates; two were
caught by capture's audits and one only by this retrospective's re-runs.

## What went right (verified, not assumed)

- **Completion and integrity.** 61/61 seats returned results (journal parity clean both runs);
  `verify` passes every applicable invariant on both records; no discarded events, no stray
  shards, record-parity PASS.
- **Convergence.** A: board mass 49 → 8 across four rounds, 21 of 23 gaps closed. B: 22.5 → 0,
  16/16 closed. Red raised fresh material every round it sat; no recycled arguments.
- **Discipline under a degraded harness.** With the identity-binding hook absent run-wide,
  every seat carried `--seat-id` by hand, correctly, for the entire programme; zero
  cross-identity writes.
- **Shim compliance.** The session's constitution-by-path shim held: 32/32 seats (A) and 23/29
  (B) read their constitution first and obeyed it; the six B seats that skipped later skill
  reads all skipped small ones (mostly `terse-communication`) — noted below, but no seat acted
  outside its role.
- **No spinning.** No agent repeated a byte-identical failing call three times in either run;
  the two failure chains (class-new bootstrap) changed approach and recovered.
- **Honest verdicts.** Both runs stamped CEILING, derived from the record, with the unaudited
  debt stated in the report head rather than smoothed. Both benches wrote explicit
  "a human should re-examine" sections.
- **Context safety.** Peak context 26% (A) / 29% (B) of the window; zero seats over the 50%
  tripwire; no compaction events; no truncated turns.
- **Report substance survives adversarial re-checking.** Every numeric claim re-derived here
  checks out: the M/M/1 figures (749 items / 250.0 weeks at ρ=749/750; 13.3/4.76 at ρ=0.93;
  the 3/7≈43% stability fraction; μ−λ≈0.004), the repo counts (336 Go files, 3 scaffold files
  at 2ce929f), and R4-1/R4-5's repairs are present in the shipped text. Assembly text parity:
  24/25 sampled blue sentences verbatim in the final report.
- **Friction as designed.** 83 friction entries across the programme, specific and actionable —
  they are what produced issues #550–#555.

## Findings, graded

### F1 — Silent model substitution: `fable` configured, `claude-opus-4-8` served (HIGH)

All 44 bulk seats across both runs requested `claude-fable-5` and were served
`claude-opus-4-8` on every turn (verified in the API responses; e.g. blue-lane-1: 1 request
line for fable, 75 responses opus-4-8). Judgment seats got `claude-sonnet-5` as configured.
Consequences: (a) the programme the operator asked for — fable research tier — never ran;
(b) run B's certified report **describes its own methodology wrong** ("the blue/red pairing
this run used is `fable` and `sonnet`") because seats and assembly only see config; (c) the
per-seat `meta.json` `model` field records the *request*, so the harness's own metadata cannot
distinguish substitution from service. Only capture's model-tier audit caught it, hours after
the money was spent (~$379 total: $197.67 + $181.18, cost.md list rates).
**Fix (#589):** the engine asserts served==configured at dispatch — the first API response
names the serving model; log it on the record at register and refuse-or-warn per a flag; the
report template states served models, not configured ones.

### F2 — Assembly stripped the anchor layer; proof footnotes dangle (HIGH)

Blue's living reports carry the full tool-placed layer (A: 11 `cite:`, 27 `fx:`, 6 `proof:`
anchors; B: 14/20/2). The assembled `report.md` in **both** runs carries **zero** anchors of
any kind, and references proof footnotes it never defines (A: `[^P1]`–`[^P6]`, 12 references,
0 definitions; B: `[^P1]`–`[^P2]`, 4 references, 0 definitions — GitHub renders them broken).
The protocol calls dropping one anchor "a hard refusal"; assembly dropped them all, silently,
in both runs. The numeric `[^1..N]` bibliography survived. This is the roadmap's
"assembly stops being an LLM" defect class with a newly measured corruption mode.
**Fix (#590):** mechanical union-copy for assembly; until then, a capture audit that fails on
(a) any anchor present in blue/report.md and absent from the assembled report, and (b) any
footnote referenced but undefined. Note `blue_scorecard`'s existing `unbacked_citations`
detector reads blue's surface, so it reported 0 while the deliverable shipped broken.

### F3 — Proof integrity: wrong-cwd proofs, and a "measured" figure no proof backs (HIGH, run A)

Re-running all six archived proofs: four reproduce byte-exact; two diverge because they were
**recorded from the wrong working directory**:
- `lane3_buildstate.sh` — recorded output contains `find: './plugins/sleeper-service': No such
  file or directory`; every ABSENT verdict below it is vacuous (`[ -e ]` on a path that wasn't
  visible). Run correctly from repo root it produces the true enumeration. The conclusions are
  independently true (confirmed here via `git ls-tree` at the pin) — the recorded *evidence*
  does not establish them.
- `lane3_accumulation.py` — recorded `0 files, 0 bytes` for `ideas/`+`research/` (same cwd
  error), immediately followed by the script's **hard-coded** narrative line "the store
  already accumulates" — a printed conclusion its own measurement just contradicted.
- The report's "measured, re-run at synthesis time, at 30 files / ~508KB" cites `[^P5]` —
  dangling per F2 — and **no archived proof contains that figure**; it was measured in-context
  and never recorded. This is red's own `pattern_ephemeral_instrument` gap-pattern, staged
  into the run's inputs, and neither red's four archive spot-checks nor the bench caught any
  of the three instances.
**Fix (#591):** `prove` records cwd and exit code and refuses an output carrying shell error
signatures; red's archive spot-check re-RUNS a sample rather than only reconciling; a
"measured" claim in the report must resolve to a proof id or be labeled unrecorded.

### F4 — `prove --quote` silent mis-anchor (MEDIUM-HIGH, run B — already #552, now with repro)

Pinned in the transcript: proof `p-efab2cde`'s long, exact, uniquely-matching quote (single
occurrence, byte offset 31633) was anchored at the unrelated shorter sentence "Status:
Proposal (for discussion)" in a different section, with the tool reporting success; the
shorter-quote retry `p-b66d1912` anchored correctly. Repro data added to #552.

### F5 — Adversary form regressions (MEDIUM)

- Red's `anchored_closures_pct`: **57 (A)** and **81 (B)** against target 100, baseline 89 —
  in A, 43% of closures carried no tool-anchored attestation.
- A's bench: **11 rulings without opinion form** (detector).
- B closed **all 16 gaps with zero formal closings** — red discharged the closing duty inside
  `mint --reason`, so at the ceiling the bench had no closings-based ruling basis, which is
  exactly the posture #554 addresses. Where the constitution demands ceremony, the tool should
  enforce it (the standing "enforce ceremony instead of asking for it" roadmap line).

### F6 — Discovery tax (MEDIUM)

~33–34% of all record-tool traffic in both runs was `--help` surface-walking (A: 198 help vs
~410 writes; B: 209 of 622 calls), repeated per seat because the constitution demands a
whole-tree read and nothing persists it. Worst: judge-terminal (A) 21 help vs 16 writes;
red-merge-r1 (B) 25 help in 54 calls. Round-4 seats needed ~3 — the knowledge exists, it just
dies with each seat. **Fix (#593):** setup stages each role's *generated* help tree as an
input file (generated from the binary, so it cannot drift — the facts-are-fields "generate the
derived carrier" preference); the constitution's read-first duty becomes one Read.

### F7 — A shared wrong idiom hit six seats (MEDIUM-LOW)

Six different seats across both runs piped `show work --json` into python expecting a
`sitting` key that does not exist (`KeyError: 'sitting'`) — same guess, six independent seats,
which means the schema is being inferred from somewhere common and wrong. Publish the actual
schema in the verb's help (folded into #593), and note three of A's refusals were **swallowed
inside exit-0 batched scripts** — per-call invocation keeps refusals visible to `is_error`
accounting and the strike counter.

### F8 — Environment gaps burned real tokens (MEDIUM-LOW)

- OCR unconfigured (`MCP_PDF_OCR_COMMAND`/`PRESET` unset) and `pdftotext` absent: red-lens-r1-L1
  (A) fell back to page-image renders — a **4.29 MB transcript, 3.61 MB of it base64 images**,
  6× the next-largest seat, for one lens slice. Friction has ranked lossy PDF handling the #1
  capability gap across prior runs; this run paid it again.
- The identity-binding hook never reached either run (plugins installed mid-session; the
  documented Setup-script field is the fix) — beyond ergonomics, it **blinded friction-parity**:
  60 envelope entries (33+27) could not be joined to seats at capture.
- `openai.com` 403 turned a leaf check into a permanently "open" question in B; the egress
  proxy's reach is part of the evidence surface. **Fix (#592, area:dev):** provision OCR +
  pdftotext, install plugins via the Setup script, and document the proxy's known-blocked hosts.

### F9 — Citation-lens economics (MEDIUM-LOW)

Per-seat citation-lens yield: A — 2 findings in round 1, then 0, 0, 0; B — 0, 0, 1, 0. The
dark-side and logic lenses stayed productive every round. W2i already down-weights citation
lenses; the evidence supports going further: dispatch the citation lens only when citations
were added or changed since the last audited round.

### F10 — Harness observations (LOW, for awareness)

- A's blue-respond-r4 found stale scratchpad from "a prior interrupted attempt" of its own
  sitting; the journal shows a clean 32/32 — an agent retry happened below the journal's
  visibility. A sitting-attempt counter on the record would make this visible.
- B's 29 seats ran strictly sequentially (two parallel workflows on a 4-CPU box; per-workflow
  concurrency cap = 2). The programme's wall clock (~4.6 h overlapped) was acceptable, but
  lane/lens parallelism inside each run largely did not materialize.

### F11 — Orchestration self-review (the session driving the runs)

- Launched both workflows before probing whether plugin agent types resolve mid-session — two
  dead launches (cheap, ~2 s each, but a probe-first order was available and used only after).
- The PR body understated F1 as "several bulk seats were served opus" when it was **all** of
  them; corrected here and in #589. The PR merged before this retrospective; the record is
  this document.
- The 40 KB gap-pattern index was pasted verbatim into both Workflow arg payloads (~50 KB of
  operator context) when setup had already staged the identical file in each run's `inputs/`;
  the engine could take a path (seats already read staged inputs).
- `maxRounds 4` bought both CEILING exits with repairs unaudited. The standing stop-and-resume
  practice only endorses *reduced* ceilings, so the honest recovery is not a +1 resume but
  either #554's grace-round mechanism or the cheap human path both benches explicitly invited:
  read `show changes --id R4-1` / `--id R4-5` (A) and confirm or reject the bench's
  leaf-verification. For future keeper runs on topics this size: ceiling 5–6.
- What held up: worktree isolation (zero cross-run contamination), the constitution-by-path
  shim (F: two compliance profiles above), disclosure discipline (every deviation is on the PR
  and the record), and capture-per-protocol on both runs.

## Report quality verdicts

**Run A (the plan): substance sound, evidence layer partially broken.** The plan's six
recommendations, the queueing analysis, and the build-state findings all survive independent
re-derivation; the ship/withdraw reasoning is grounded and the CEILING debt honestly stated.
Its defects are F2 (no anchors, dangling `[^P1..P6]`), F3 (two bad proofs + one unbacked
"measured" figure), and the disclosed mis-titled bibliography entry (`[^2]`, the #551 debt).
Recommendations are safe to act on — the specific *recorded evidence* for the build-state and
accumulation figures is not the artifact to cite; this retrospective's re-runs are.

**Run B (the positioning): argument sound, self-description wrong in one place.** The
compare/contrast holds up (the absence-claim is stated as absence, the kernel finding is
gated on its own prototype), but it describes its bulk tier as fable (F1), ships dangling
`[^P1..P2]` (F2), and closed its entire board without formal closings (F5) — which is why its
bench had to substitute at the ceiling.

Neither run is red-PASSed. Both are CEILING. Every consumer of these reports should read the
verdict stamp first, as the reports themselves insist.

## Addendum (2026-08-26): what the fix work itself turned up

Three things surfaced while implementing the fixes, and each changes something above.

**F12 — the record's storage format moved, and every archived run now reads as a clean empty
board (HIGH, #595).** Building the binary from `ac04072` and pointing it at the 2026-08-23 runs
returns 0 gaps, 0 findings, 0 events, verdict unrecorded, **every invariant `[ok]`, exit 0,
empty stderr**. The invariants pass vacuously over zero events. This is the sharpest instance
yet of the shape this whole document is about — the miss and the honest zero as the same bytes —
and it lands on `run-archive/`, which CLAUDE.md calls the only part of a run that outlives its
container and says every audit re-reads. Fixed in the stack: the read path now refuses a record
in a format it does not own, and `setup` records the writing binary's version so the refusal's
advice is answerable.

**F2 is narrower than stated above.** Assembly is *designed* to weave anchors into visible
footnotes rather than preserve them — `AssemblyScreen`'s own comment says so, and the 11
citation anchors did convert correctly into `[^1]`–`[^11]`. The defect is that the weave is
**lossy per class**: the proof anchors became references with no definitions in the same pass.
So the audit that ships is footnote *definedness* on the assembled artifact, not anchor
survival — a rule markdown itself supplies, needing no heuristic. Re-run against the two
reports as they shipped it fails both (6 and 2 dangling) and passes both as repaired, which is
the strongest available evidence both that the gate works and that the repair is complete.

**F3's remedy is one ask short of what the issue asked, deliberately.** #591 wanted `prove` to
record cwd. It is not recorded: `cmd.Dir` is always the run directory, so the field would
restate a constant, and a fact nothing can disagree with is not worth a field by this repo's
own standard. The refusal names the working directory instead, where the seat that assumed
otherwise will actually read it.

**F12's fix met a change that landed mid-flight, and the collision is instructive.** While the
stack was open, #597 introduced an event-schema **epoch** that explicitly *replaces* the record
binary's version. The first cut of the F12 fix recorded `recordToolVersion` in the run config —
reintroducing, under the retired name, the exact concept #597 had just removed on the grounds
that "a release number moves for reasons the schema does not care about". The rebase turned that
contradiction into a build error rather than a shipped inconsistency, and the fix now records
`eventSchema`. The two mechanisms divide cleanly and neither subsumes the other: #597 guards the
WRITE side (a stale binary cannot start a run), F12 the READ side (an archived run predating both
still reads as a clean empty board without it).

**Process note.** Two of these fixes were caught or corrected by the repo's own gates rather
than by me: `TestMergedEventsOnAnEmptyOrAbsentRun` rejected my first shard-matching rule as too
broad (it would have refused a run over a stray file), and
`TestEveryRegisteredFlagIsInTheDeclaredVocabulary` caught that declaring a flag constant is not
the same as registering the word; the compiler caught the third, above. All three are gates doing
exactly their job, and they are worth naming here because the retrospective's own complaint is
that gates passed over nothing.

## Addendum 2 (2026-08-26): what building the served-model gate turned up

Three corrections to what is written above, all measured while implementing #589.

**F1 overstates the environment. `fable` is served — it just was not served to those runs.**
F1 says the configured research tier "never ran". That is true of the 2026-08-23 programme and
was read as a property of the environment. It is not: two subagents dispatched on 2026-08-26
from this same container were served `claude-fable-5` for 52 and 30 turns respectively, with no
fallback notice. So the substitution is INTERMITTENT, which is worse than a standing environment
limitation — a run cannot be cleared once and trusted thereafter, and a check that fires only at
setup would have passed on both readings.

**The audit that "noticed" did not fail. Capture exited 0.**
F1 credits capture's model-tier audit with catching the substitution "hours later". It caught it
as a **WARN**, and capture exits 2 only on a FAIL — so the run's own mechanical auditor observed
44 substituted seats and returned success. The grading was inherited from the cost guard, where
DEARER-than-configured is the trap (money) and cheaper is "verification may be discounted".
`claude-opus-4-8` is CHEAPER than `claude-fable-5`, so the measured incident took the soft
branch. For a research debate the axis is wrong: what a tier buys is the strength of the party
arguing each side, and a silently WEAKER adversary is the worse of the two failures, not the
softer one. The tier check now fails in either direction.

**It was measurable all along, on the seat's own first turn.**
The retrospective established the substitution by reading `message.model` across whole
transcripts — which is why it could only be done post-hoc. The harness in fact DECLARES it, as
fields, on the seat's opening assistant line, before that seat's first tool call:

```json
{"type":"fallback","from":{"model":"claude-fable-5"},"to":{"model":"claude-opus-4-8"}}
```

and the trajectory is at `agent-<agent_id>.jsonl`, keyed by the same `agent_id` the PreToolUse
hook already injects as `FEOV_AGENT_ID`. So `register` — every seat's first act, already
engine-observed — can read it. Replaying all 63 trajectories through the new reader reproduces
the retrospective's own numbers exactly: 44 declared `claude-fable-5 -> claude-opus-4-8`, 17
judgment seats served `claude-sonnet-5` as configured, 2 unrelated (they are this retrospective's
own diagnostic subagents, and they are the fable ones).

**Process note, continued.** Two more gates corrected me, and the second is the more instructive:

- `TestEverySetShapedFlagIsEitherDeclaredOrExempt` refused a new `--format` flag advertising a
  closed set with no declared enforcement.
- My own first shape check on the agent id was `^[0-9a-f]{6,64}$` — which matches every real id
  ever measured, and would have silently and permanently refused every id the day the harness
  changed the format, reporting each one as a seat with no trajectory. That is precisely the
  defect class this retrospective is about, written by the person writing the fix for it. It was
  caught by an existing test fixture using an id of a different shape. The check now refuses only
  what would widen the glob, and an unrecognised id is SEARCHED FOR and honestly not found.

**Still open on #589 after this change.** The report's own methodology self-description is
sourced from config, not from the record — the certified run-B sentence naming its pairing as
`fable` and `sonnet` would still be written that way if an operator consented to a substitution.
The measured fact and its projection (`show tiers`) now exist for a seat to quote; instructing
the assembler to quote them is a prompt change, and it is not in this PR.

## Addendum 3 (2026-08-26): re-running the archived proofs says something different

**Correction to F3.** The retrospective records "re-ran all 6 archived proofs: 4 reproduce
byte-exact, 2 diverge". Re-measured while building the proof re-run audit: **all six reproduce
byte-exact**, from the run directory — which is the working directory `prove` actually uses.

Both readings are of real runs; they differ in *where from*. The earlier "2 diverge" was measured
from the REPO ROOT, where `./plugins/sleeper-service` exists and the script therefore prints a
different tree. That divergence was evidence of the cwd bug, and describing it as
non-reproducibility reads as a different defect than the one that happened.

The distinction matters because it decides what a re-run audit can do:

- `lane3_buildstate.sh` carries `# Re-arms: run from repo root` in its own header and takes
  `root="${1:-.}"`. Executed with the run directory as cwd, its `find` fails **identically every
  time**. The recorded output — error line, then a column of vacuous `ABSENT` verdicts — is
  perfectly reproducible.
- So **a re-run sample cannot catch the wrong-cwd class.** #591's ask 2 states its acceptance as
  "a capture re-run sample catches a seeded wrong-cwd proof"; that acceptance is not achievable by
  re-running, and the class it names is already closed by ask 1 — the error-signature refusal that
  landed in #599 and does refuse this exact script.

What a re-run does catch is the class the issue names in its own third bullet:
`pattern_ephemeral_instrument` — a measurement of state that has since moved, recorded as though it
were a computation. Also a script edited after it was recorded, and a proof whose stored artifact is
gone. That is what the new `proof-rerun` capture audit is scoped to, and the PR says so rather than
claiming the acceptance sentence.

**A layered-defence note worth keeping.** Pointed at the 2026-08-23 archive, the new audit returns
`SKIP — the record could not be read, so no proof could be re-run — which is NOT a run whose proofs
reproduce`. That is #598 doing its job: without the legacy-format refusal, this audit would have
enumerated zero proofs from an unreadable record and reported "this run recorded no proofs" — a
clean-looking zero over six real ones, which is the defect this whole retrospective is about,
rebuilt one layer up.

## Priority order for fixes

1. **#589 served-model assertion** — cheap, prevents the whole F1 class, and makes cost/tier
   decisions mean what they say.
2. **#590 assembly anchor survival** — the deliverable currently loses its machine-checkable
   layer every run; mechanical assembly is already on the roadmap, the capture audit is the
   stopgap.
3. **#591 proof cwd/exit integrity + re-running spot-checks** — the prototyping-and-testing
   discipline is the operator's standing requirement; this run shows it can record artifacts
   that look like proofs and aren't.
4. **#552/#553/#554/#555** — already filed from the runs' own friction; this review adds repro
   evidence and independent confirmation.
5. **#592 environment provisioning**, **#593 help-tree staging + schema docs** — token
   economics; together they address roughly a third of the record-tool traffic and the single
   largest transcript in the programme.
