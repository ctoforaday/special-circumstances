# Design review — the whole corpus, 2026-08-15

> STATUS 2026-09-02: shipped — historical record (review as written stands; §V guidance marked item-by-item below)

Inputs: all 851 commits, all 268 pull requests (261 merged), all 156 issues regardless of
state with their discussion threads, the plans/ and ideas/ corpus, the four plugins' code
(73,270 lines: frank-exchange-of-views 51,286 · prosthetic-conscience 13,290 · gray-area
3,328 · scripts 5,366 · sleeper-service 0), and the produced research artifacts. The
question held throughout is the one #380 already states for the tool and this review
applies to the whole system: **does each piece shape the output or expose dissent — and is
the complexity that remains the minimum the ambition needs?**

---

## I. How we got here — five eras in thirty-five days

### Era 1 — bootstrap and the rule corpus (Jul 11–13, PRs #1–#14)

The AgentOrange port: marketplace skeleton, the prosthetic-conscience skill corpus in three
chunks, and the FEOV debate engine as a single workflow script. Two design decisions from
this week still govern everything: **review rounds inside every PR** (each feature PR
carries its own adversarial re-read as follow-up commits), and **doubts as artifacts**
(ideas/doubts.md — hypotheses about our own design, checked against dogfood runs). All five
founding doubts were later CONFIRMED by run 3, which is worth remembering: the initial
design intuitions about our own weaknesses were right, and the discipline of writing them
down is what made them checkable.

### Era 2 — run-driven hardening (Jul 14–18, PRs #15–#45)

Runs 2–5, ~5M tokens of tuition. The efficiency phase (#19–#22), the constitutions gaining
telos and win conditions (#32), and the pivotal measurement: **the hand-written markdown
lied**. A board disagreed with its ledger 9-open/9-closed against 3-open/15-closed; a haiku
smoke fabricated `archive_blocks: 22` in a round whose true count was 0. The record tool
(R1/R2, #34/#35/#41/#42) was born on Jul 18 — 92 commits in one day — as the answer:
an append-only event log, seats writing only through role-scoped verbs.

### Era 3 — the record becomes the only channel (Jul 19–27, PRs #46–#130)

The defining arc. #62 ("make the record the ONLY inter-agent channel") plus
plans/tool-is-the-contract.md, which measured that **~44% of the seat-prompt corpus was
tool-absorbable**. The ledger, archive, candidates/, debate.md write-paths all retired; the
report becomes `f(event log, blue/report.md)` with the seat authoring nothing (#61);
`claim_count` and `citations_checked` move from self-report to derivation (#70/#71); every
operator .mjs is ported to Go (#121, five slices), leaving debate.js as the only JavaScript.
The goja fuzzer (#97/#98) starts driving the *real* debate.js against the *real* binary.

### Era 4 — continuity and the fourth plugin (Jul 28–Aug 2, PRs #131–#243)

Gray-area ships (trajectory capture with mandatory provenance; "exploration may summarize,
adjudication must cite"). Prosthetic-conscience gains the continuity loop (checkpoint seal
at every seam, restore, /resume) and the hook consolidation (#201: one binary per event,
units as libraries). CI grows the quality gate, mutation testing exposes that 100%
coverage protected nothing (2 of 8 secret patterns deletable with the suite green), and the
always-on import list gets its guard (#214).

### Era 5 — derived-not-asserted to the leaf, then deletion (Aug 3–15, PRs #252–#422)

The evidence model completes: cached fetch, tool-inserted immortal citations (#258), proofs
the engine runs twice and red re-executes (#279), estoppel (#275), the motion collapse — three
adjudication exchanges become one mechanism with an id (#344) — and identity arriving as
fields (#348). Then the tone shifts, and this is the healthiest signal in the whole corpus:
**the repo starts deleting**. `existence` deleted (0.65.0), `blue confidence` deleted
(0.54.0), `observe`/`dispose` retired (#327), the hand-written CHANGELOG retired (#402),
ledger/archive collapsed into `show board` (#379), dead renderers deleted (#386), the
legacy prompt set deleted (#406). The seat-probe elicitation (#363, #380) replaces "which
verbs get used" with "what does a seat perceive, weigh, and decline" — measurement of the
surface itself.

### What the eras say about trajectory

Each era ended by turning the previous era's workaround into a first-class mechanism, then
deleting the workaround. Prompts asked for behaviour → the tool enforced it → the prompt
clause became teaching text → (currently mid-flight) the teaching text should now shrink.
The system is one era away from its own stated destination and the remaining distance is
mostly *removal*, not construction.

---

## II. Where we are — the architecture as built

Three layers, one contract:

| Layer | Carrier | State |
|---|---|---|
| Orchestration | debate.js (sandboxed Workflow JS, no fs by design) | 1,067 lines; control flow driven by **envelopes** |
| Contract | feov-record (Go, one binary, role-scoped verb trees) | 0.68.0; validation-at-write; views just-in-time from the record; refusals teach |
| Judgment | four constitutions + seat prompts interpolated by the engine | constitutions ~48KB; single seat prompts up to ~4,500 words |

Around them: assembly (report = f(record, blue/report.md), raw-sliced never re-rendered),
capture (post-hoc audits off the record), scorecards (in-run self-read, priors-are-poison),
gray-area (trajectory evidence), and a gate suite (goja fuzzer, goldens, flag/enum coverage
gates, prompt-vocabulary gates, mutation testing, seat-probe harness).

**The strongest abstractions, named:**

1. **Validation-at-write with teaching refusals.** record.go's `validate` is the best code
   in the repo: every requirement carries its measured incident, every refusal tells the
   seat what to do instead, and the write is the one chokepoint so no caller can skip it.
   This is "the verb set IS the role boundary" made real.
2. **Derived, not asserted.** The verdict, claim_count, citations_checked, fix_basis,
   reproduced, awaiting_proof — every consequential number moved from a seat's word to a
   computation, each move justified by a measured fabrication. This is the system's core
   idea and it has been executed to near-completion.
3. **Just-in-time views.** No materialized projections, no staleness window, one replay
   feeding both JSON and markdown renderers. The render-shadow waypoint was correctly
   recognized as a waypoint and removed.
4. **The immortal-anchor mechanism.** One machinery (`lens.InsertAnchor` + the lockdown)
   carrying three classes — findings, citations, proofs — with the bijection guarded at the
   edit. Genuine reuse, not similarity-shaped duplication.
5. **Measured clauses.** Nearly every rule in the codebase cites the incident that created
   it. This is the repo's signature idiom and its most valuable property for a future
   maintainer: nothing is doctrine without provenance.

---

## III. The complexity audit — what earns its place and what does not

Held to #380's criterion. Four findings, ordered by cost.

### 1. The seat prompts are the system's largest unearned complexity

The single blue-respond prompt is ~4,500 words carrying ~30 imperative clauses; the
red-merge prompt is comparable. tool-is-the-contract measured 44% of the prompt corpus as
tool-absorbable in July — and since then the tool absorbed most of that *while the prompts
grew anyway*, because each incident added a clause narrating the fix. The prompts now
restate, at dispatch time, guarantees the tool already enforces at the write:

- anchor preservation is tool-enforced — and the prompt still explains the enforcement
  at length;
- `--answers` refusal, estoppel, verdict derivation, `--run` injection — all enforced,
  all still narrated;
- measured incidents ("a seat once summed twelve integers in its head…") appear in the
  prompt for every seat on every round, teaching history to an audience that needs only
  the duty.

The tool's own trajectory (worklist `sitting` block, enum help, refusals that name the
verb you meant, `show report`) exists precisely so that the surface teaches itself at the
moment of need. The prompt clauses are the scaffolding of that migration and the migration
is far enough along to strike the scaffolding. **This is the highest-leverage
simplification available**, and the probe harness (#363) is the instrument to do it
safely: measure seat behaviour on a dieted prompt against the current one, board by board.

The engine's own numbers say the same thing at a smaller scale: `#421` notes the injected
duty list has been one line since `show` became a group. The direction is set; it wants a
deliberate pass rather than accretion in reverse.

### 2. The envelope is a second channel the record was built to remove

The engine cannot read files (sandbox, by design) so control flow rides structured agent
returns: verdict, gaps, claim_count, dispute refs, `round_record_appended`. The
envelope→refs migration correctly shrank them to routing data — but the dual channel keeps
producing the exact defect class the record exists to kill, and the record of that is long:
#83 (gaps_outstanding disagreed with the board), 0.42.0 (the ephemeral channel was more
precise than the permanent one), the lineage guard that killed a 723k-token run whose
record was intact, #289 (a judged deadlock's determination lives ONLY in the envelope —
DeriveVerdict itself says the determination "is not on the record"), #394 (every petition
sitting reuses the seat id `judge-petition`, so sittings silently collapse).

The engine is the last consumer stuck on the lossy channel. Two honest paths exist:

- **(a) Clerk pattern, incremental:** every control-flow read goes through a seat that
  pulls the record and relays a number (the assemble seat's `open_gaps` already does
  this). Cheap, keeps the sandbox, but adds seats whose only job is ferrying facts.
- **(b) Move orchestration into the tool:** a `feov-record run`-class driver in Go that
  reads the record directly, dispatching seats via the harness. This deletes the envelope
  concept, makes #289 recordable, and turns debate.js's 1,067 lines into data (a round
  script). It is the larger surgery and the honest end-state of "the tool is the
  contract" — the contract should bind the orchestrator too.

The fork does not need deciding today, but it should be *decided rather than drifted into*:
every new envelope field deepens (a) by default.

### 3. Identity is the last unmediated fact, and its fix is measured and unbuilt

#348 closed with the measured failure still live (#396): nothing sets FEOV_ROUND/FEOV_SEAT,
so the regex fallback is in production the only path, and `judge-terminal` still stamps
round 0. The chain #290 → #345 (per-agent_id settings file, seats named by the engine,
entity-grouped command tree scoped to detected identity) is the one remaining structural
change with compounding returns: it deletes `RoundOf`'s documented danger, fixes #394,
unblocks the entity-tree restructure, and closes the `--seat-id` typo class the same way
`--run` injection closed its ten-errors-per-run class. The gating measurement
(PreToolUse carries agent_id, 9/9, stable, distinct across concurrent agents) is done.
Sequence this **before** any further surface work — it changes what the surface hangs off.

### 4. The changelog lives in the wrong carrier

root.go is 96% changelog (#407, correctly filed): ~850 of 1,063 lines are version
narration with no programmatic reader, in the file every CLI change touches. The
discipline — every contract change explains what a stale binary would do wrong — is
exactly right; the carrier is prose in a compilation unit. The same instinct that retired
blue/CHANGELOG.md ("the event is the record") applies to the tool's own history: a
versions.go table (version → entry) or docs/versions.md with a staleness gate keeps the
discipline and returns root.go to being readable code. Likewise the versionguard family
(#405, #423, #424): carriers typed by hand, diff-scoped gates that pass when they did not
run — the class fix is the tri-state (#411: pass / fail / **did-not-apply**) made uniform
across scripts/check, because half of the recent defect intake (#385, #409, #416, #338,
#296) is exactly "a gate that did not fire is not a gate that passed."

### What is NOT excess — called out to prevent over-correction

- **`merge mint`'s 23 flags.** The #380 static audit already adjudicated this: minting IS
  the act that carries the most, and the four `--class-new` flags are the one sub-form to
  extract. Do not diet mint.
- **The motion group's zero usage.** Measured to be a property of the corpus (red has
  never graded a fix above medium cost — #418), not of the surface. Seats perceive, weigh,
  and correctly decline. Keep it; watch red's `complexity_cost` distribution instead.
- **The fuzz/golden/probe suite** (2,539-line fuzzer, scenario boards, goldens). This is
  where the correctness comes from; the 0.59.0 incident — every `show` in every prompt
  refused for a release while three gates stayed green — argues for *more* of exactly this
  kind of end-to-end instrument, not less.
- **The dual-read of retired vocabularies** (0.48.0/0.49.0). Permanent by design because a
  record is permanent and consumers' runs are invisible. Correct, and correctly bounded:
  never-used vocabularies (confidence, existence) were deleted outright instead.

---

## IV. The report — the actual product, honestly graded

The report architecture is right: blue-authored surfaces audited in place and lifted
verbatim; everything else composed from the record; both sides of every exchange rendered;
regrades shown as arguments, not final numbers; the revision history and cost on the page.
The 2026-08-05 smoke's report demonstrates the composition working end to end.

It also demonstrates the gap between *auditable* and *readable*:

- **Rendering seams reach the reader — and the code has outrun the evidence.** The smoke's
  report shows avenue MOVES as empty-bold `**** —` bullets and risk-matrix cells cut
  mid-word; both are FIXED in the current assembler (`record.Avenues` renders one lifecycle
  row per avenue; `concise` takes the first sentence and the matrix names itself a scan
  surface). No run has exercised the fixed renderer — which is itself the finding: the
  report's face has been reworked repeatedly against artifacts no current binary produced.
- **The transcript dominates by volume.** The full debate belongs in the artifact; whether
  it belongs in the *reader's* document or in an appendix file the report links is a
  curation decision nobody has yet taken deliberately.
- **The quality loop has not closed.** #257 is explicit: haiku runs are the proving
  ground, **the sonnet run is the real validation** — and it has not happened. The
  seven-thread review (report quality, efficiency, relevance, discovery, struggle, reading
  behaviour, source trust) is designed and unrun. #247 (source *trustworthiness* — today
  any source counts) is still open. Meanwhile roughly 180 PRs have landed since the last
  keeper-grade research run (2026-07-20). Every run since has been a smoke or a probe.

That last point is the review's sharpest finding: **tool sophistication is now several
releases ahead of validated report quality.** The engine can record, derive, estop, prove,
and adjudicate — and the last document it produced under real load answered whether 7 is
prime. The ratio (51k lines of plugin to a corpus of smoke reports) is only justified by
the validation run that has been queued since #257 was filed.

---

## V. Guidance — ordered, with the reasoning attached

1. **Stop building; run the sonnet keeper run; run the seven-thread review.** The next
   material design information does not exist in this repo — it exists in the report a
   full-strength run produces. Freeze the verb surface for it (the version gate makes this
   cheap to honour). Let the review's findings, not incident accretion, set the next arc.
   **NOT DONE (2026-09-02):** #257 still open; run-archive's latest capture is 2026-08-23
   and none is keeper-grade (the 2026-08-23 programme runs are CEILING, not red-PASSed).
2. **Then the prompt diet, measured by the probe harness.** Strip enforcement narration
   and measured-incident history from the seat prompts; leave role, duty, and pointers
   into the self-teaching surface (`--help`, worklist sitting block, refusals). Target the
   blue-respond and red-merge prompts first; A/B against current prompts on the probe
   boards before trusting it live. The incidents move to where they already live (code
   comments, plans, law/) — the repo needs the history; the seat needs the duty.
   **IN PROGRESS (2026-09-02):** incident-history strips landed on main ("Keep the rule,
   drop the obituary", 413c21d2/8f443e47; "Trim the judge prompt", 586cd6ec) — but no
   probe-harness A/B is on record, and debate.js plus three of four constitutions have
   net-grown since 2026-08-15.
3. **Decide the envelope's future explicitly** — clerk pattern vs orchestration-in-tool.
   Either resolves #289/#394; drifting resolves neither. If choosing the tool-side driver,
   debate.js becomes data and the biggest remaining JS surface retires, which is the
   direction #121 and #177 have pointed all along.
   **DECIDED (2026-09-02):** orchestration-in-tool — the record-run driver is landing on
   main (#605, #612, #621, #631, #641; plans/record-run-migration.md).
4. **Build the identity chain (#290 → #345) before any other surface work.** The
   measurement is done, the design is written, and half a dozen open defects (#394, #396,
   #419's vocabulary collision surface, the RoundOf fallback) collapse into it.
   **DONE (2026-09-02):** identity now resolves from the record and an unregistered agent
   cannot act (tools/internal/seatenv/identity.go, commit 3b27efea on main); the
   petition-sitting collapse fixed via #433.
5. **Uniform tri-state gates.** Generalize #411's `n/a` across scripts/check and
   versionguard: a gate reports pass, fail, or did-not-run — never a pass it did not earn.
   This closes the largest active defect class in one move.
   **DONE (2026-09-02):** scripts/check reports a distinct not-measured SKIP
   (scripts/check/main.go `notMeasured`); versionguard's not-measured fix merged via #424.
6. **Relocate the tool changelog (#407)** to a queryable carrier with a staleness gate;
   keep the discipline, free the file.
   **RESOLVED DIFFERENTLY (2026-09-02):** #597's event-schema epoch deleted the version
   narration outright (cli.Version, the 49-entry capabilityDeltas table — commit 869b6d3b)
   rather than relocating it; root.go is now ~500 lines.
7. **Reader-tier the report.** The smoke-visible seams (empty-bold avenue bullets,
   mid-word matrix truncation) are already fixed in the assembler; what remains is the
   curation decision — what the reader's document carries versus what it links (the
   transcript above all) — and it belongs to the seven-thread review's report-quality
   thread rather than to more pre-validation rendering work.
   **NOT DONE (2026-09-02):** gated on the seven-thread review; #257 still open.
8. **Ship or shelve sleeper-service.** Its README honestly declares "scaffold only
   (Phase 0)" — but it has said so since Jul 11 while the marketplace manifest lists it
   beside three real plugins. Thirty-five days of scaffold is a decision by drift: either
   its first increment (/self-improve consuming the friction corpus the runs already
   produce) or withdrawal from the manifest until it exists.
   **NOT DONE (2026-09-02):** README still says "scaffold only (Phase 0)" and
   marketplace.json still lists it; the 2026-08-23 programme produced its plan
   (run-archive/2026-08-23_sleeper-service-plan) but neither increment nor withdrawal
   has landed.
9. **Reconcile plans/storage.md with reality.** #107's trigger (#62) has fired, but the
   hand-rolled locking + monotonic clock the decision record argued against building is
   now built, tested, and fuzzed. Either migrate to SQLite on the recorded rationale or
   amend the record to say the JSONL path won on evidence — a decision record
   contradicting shipped reality is the exact class this repo polices elsewhere. #68's
   schema-version field is cheap insurance either way and should land before the next
   event-schema change, not after.
   **DONE (2026-09-02):** plans/storage.md marked LANDED 2026-08-28 — the record moved to
   embedded SQLite (internal/record/recordsql); the schema-version insurance landed as
   #597's event-schema epoch.

## VI. Verdict

The system's spine — the record, validation-at-write, derived-not-asserted, teaching
refusals — is genuinely minimum-necessary for the ambition, and the recent deletion era
shows the discipline to keep it that way. The complexity that fails the #380 criterion
sits in exactly two places: the seat prompts (scaffolding for a migration that has largely
completed) and the envelope channel (a sandbox constraint promoted into architecture). The
evolution's own logic — every era turns the prior era's workaround into a mechanism, then
deletes the workaround — names both as the next deletions. But the sequencing matters more
than the list: the validation run comes first, because "the best possible report from our
process" is currently an untested claim, and every other line of this review is downstream
of what that run shows.
