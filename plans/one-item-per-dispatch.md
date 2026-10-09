# The rig — one work item per dispatch

> STATUS 2026-09-27: proposed — a research document and a direction, not an implementation plan.
> Landed on `main` as #1185; audited and corrected against `b0ff6679` the day after (the audit's
> findings are listed at the end of this banner). It asks for one ruling (§I, "The question")
> before any step in §III is started. Every number in §II is from the run it names, and the ones
> marked *single run* have not been reproduced; §V.0 is the measurement that has to come first,
> and its item 4 has now been run (§II.4).
>
> Audit 2026-09-27, seven defects, all corrected in place: the wall-clock estimate mixed a haiku
> per-turn rate with opus turn counts and put the bookends at half their measured length; the
> dispatch read was described as seat-less when its verb writes dispatch rows and #1192 registers
> a writer; §III.5 described a flat "re-read the full report" order that 19a4ddef had already
> scoped; §III.6 still said "concurrency 2 on this box" after §II.4 was corrected; the cap row
> predated #1192; every `debate.js` line reference was off by one after #1185's own edits; and
> two m11 measurements cited a box-local note when commits 0b3715e7 and 2fe606cb carry them.
>
> Added 2026-10-02 (#1222): §III.6 gains a fifth channel — a plain process can post into a live
> session's inbox socket with no model call, and that channel addresses a process, not an
> in-process Workflow seat — and gblock's ruling that every host surface the rig drives sits
> behind a harness boundary so Antigravity can be swapped in (`plans/dual-target-claude-antigravity.md`).
>
> Corrected 2026-10-09 against `12879192` (179 commits since `44ce54d6`; EventSchema 23): the record
> now names the work item's parts — `kind` is the record's `Occasion`, the dispatch row carries
> seat, occasions, gap ids and blockers, and markers part 4 delivers the brief — so §III.1 is
> restated on those names; #1190/#1191 are closed by the record route and §III.2/§IV.1 no longer
> argue from envelopes; the petition-latency ruling and #1281's relay rule bound §III.2 and are
> argued against, not around; §III.2's schemas and §III.3's generator belong to slimming plans 09
> and 07; §III.6 states the vehicle question as relay versus read. Each correction is dated in
> place. Nothing in §III has started; §I's ruling is still open.

## I. Summary & Goals

**The relationship has inverted, and the tree is half-way through the inversion.** The engine
began as agents driving a shared tool: a seat sat for as long as it liked, chose its own verbs,
read whatever it wanted, and decided when it was done. What the tree does today is different in
every part that was measured to matter: the record computes who sits and on which gaps
(`dispatch next`), the hooks supply the seat's identity and run, the work list is pushed into the
seat's context at `SubagentStart`, the tool answers "may I stop" after every write, and a hard
call limit ends a sitting that does not end itself. Only the inside of a sitting is still the
agent's: what to read, in what order, how many times, and when to return.

This document takes that inversion to its end and asks what the engine looks like when a
dispatch carries **one work item** — the item on the list the seat is already handed, with its
brief delivered, and one structured decision returned — and when the harness, not the seat,
owns everything between one decision and the next.

**Why now.** The measured defects of the current shape are all one defect: the sitting is the
unit, and a sitting is a long, self-directed context.

- Orientation dominates: 34 of 38 calls in one lens sitting were orientation for 4 adjudications
  (§II.3). Before 19a4ddef scoped the re-read, seats re-read the whole report in 13 of 13 empty
  sittings; after it, on m11, 12 sittings handed their gaps read the board anyway and 6 looked a
  gap up by id while holding it (commit 0b3715e7).
- Context, not output, is the bill: 237M cache-read tokens against 0.24M output in one keeper
  run; a warm resume cut turns by 30% and cost by 6%, because the conversation it re-read grew
  every sitting (§II.3).
- The plan is computed at the start of an epoch and is stale by the middle of it: blue sits on
  gaps a lens already closed, which is why `found_closed` exists (§II.1).
- The chair is a judgment-tier sitting whose first and largest job is to run a deterministic read
  and relay it verbatim — a model call spent on a query (§II.1).
- Four facts the prompts depend on — holdings, rulings, relief, friction — live in the workflow's
  JavaScript memory and not on the record, so nothing can refuse them and a resume rebuilds them
  only if the agent cache replays (§II.1). That is [[facts-are-fields]] inside the engine itself.

**The question this document asks the human to rule on.** Is the debate engine a swarm of agents
that use a record, or a rig that drives a record and calls an agent at each point where judgment
is needed? The tree has been answering "rig" one mechanism at a time since #500, and it has now
used the word: commit 0b3715e7, "Speech is questioned; the rig is not", rules that the envelope
the harness delivers is the instrument reporting its own state and the contents it carries are
another party's words. Saying it out loud changes what the next ten changes are, and it changes
what the Workflow script is *for*.

### Goals

1. A dispatch carries one work item and its brief; the seat returns one structured decision;
   nothing in the seat's context outlives the item.
2. The engine's control flow reads the record and never a seat's self-report — the record is the
   only state, and every fact a prompt interpolates is a projection of it.
3. Work moves as soon as it is ready: no epoch, no barrier between lens and blue, no chair sitting
   as a global clock.
4. Discovery stays whole: finding what is *not* on the board is an item with a whole-area brief,
   never a casualty of quantization (§IV.1 is the objection this goal answers).

### Non-goals

- Replacing the Workflow tool. §III.6 is the conditional under which that becomes worth it;
  concurrency is not what it would buy (§II.4).
- Reducing judgment. The bench, the PASS/FAIL verdict and the lens's audit of an area remain
  whole reads; the change is to what surrounds a decision, not to the decision.

## II. Technical Context — what the tree does today, measured

### II.1 The engine (`debate.js`, 1148 lines at `b0ff6679`)

> Corrected 2026-10-09: 1262 lines at `12879192`; the epoch loop is at `:1076` (chair `:1078`,
> lenses `:1116`, blue `:1121`, bench `:1150`) with a petition bench sitting inserted before the
> lenses (`:1112`, #1239). Every other line number in this section is at `b0ff6679` and is read
> as a locator, not a fact. `takeFriction` and the envelope `log` are gone (#1282); `unruled_motions`
> is gone from the chair envelope, which is `{plan, verdict, notes}` (`:578`, #1239).

The script calls `agent()` and `parallel()` only; `pipeline()` is stubbed in the goja harness and
unused. The bookends are fixed: frontier (1 agent), lanes (N in parallel), synthesis (1),
terminal bench (0–1), assemble (1). The debate is `while (!halted)` at line 986, one **epoch**
per iteration, in fixed order:

1. The chair sits (judgment tier). It runs `dispatch next` and relays the JSON as `plan`.
2. Lenses sit in one `parallel()` batch — the only concurrency in the loop.
3. Petitions are heard serially. Blue sits, serially, on every gap in the plan. The bench sits.

Three structural facts follow from the script's own comments.

- **The script reads nothing.** "debate.js reads no record" (lines 49–51, 135, 367).
  Every fact the engine acts on is a seat's envelope relay: the chair's `plan`, `verdict` and
  `unruled_motions`; blue's `claim_count`, `found_closed`, `manifest`; the bench's
  `dispositions` and `holdings`. The Workflow tool has no side-effect primitive, so a
  deterministic read costs a model call, and today it costs a judgment-tier one.
- **The epoch is the quantum.** The plan is computed once, at the top; a lens made ready
  mid-epoch waits for the next chair; blue is dispatched on gaps that closed while it waited.
  The `for … of blues` loop runs at most once because `engage()` groups every gap by seat — a
  plan holds at most one blue party and one bench party, and that party carries all its gaps.
- **JS memory holds facts the record should.** `holdingsInEffect`, `rulingsInEffect`,
  `reliefInEffect`, `friction`, `petitionLog` (lines 674–784) are interpolated into later
  prompts and are lost on a throw unless the agent cache replays. Relief is described as
  "operative this sitting" and only ever grows.

Two defects found while mapping, filed as #1190 and #1191: lens envelopes are
unschematized strings, so `takeFriction` and `hearPetitions` read `undefined` and a lens's
petitions never reach the engine; and `JUDGE_ENVELOPE` has no `petitions` field while the bench
prompt grants the petition right.

### II.2 The rig that already exists (harness drives seat)

| Mechanism | Where | What it decides for the seat |
|---|---|---|
| Who sits, on which gaps | `record.PlanDispatch`, `dispatch.go:82` | readiness from the board; per dispute (roundless §III.B.3) |
| Identity and run | `feov-pretooluse` rewrites every Bash command with `FEOV_RUN`, `FEOV_AGENT_ID`; `seatenv` refuses a disagreeing `--run` or `--seat-id` | the seat never types either |
| The work list | `feov-subagentstart` → `sittingwrite.injectWorkList`, delivered as `additionalContext` | measured to land in the seat's own context three times (#500, #507, 2026-09-25) |
| May I stop | `standing.go`: the blocking items and `complete` ride every successful write | 4–5 KB list, ~1.2k tokens, too heavy to attach to every act |
| The hard stop | `sittingcap`, default 150 calls, counted from the seat's `register`; the hook refuses past it and lets only `register` through. Since #1192 a write by an unregistered agent registers it (`seat.requireBound`), so a sitting that writes is capped by construction; a sitting that writes nothing has no header and nothing to cap (`sittingcap.go:14–25`) | the only thing that forces a return |
| The surface | `agentgen` inlines the tool's `manual` into every constitution: ~10.7k generated words per lens on 0.5–0.7k hand-written; blue-researcher 11.1k on 3.0k | exists because fetching help cost 171 calls, 10% of a 69-seat run |

What is still the seat's: everything inside the sitting. The nearest prior statement of
per-item routing is "dispatch is per dispute" (`plans/roundless.md` §III.B.3, `chair/command.go:7`).
The nearest counter-principle is `plans/tool-is-the-contract.md` §IX, "active pull beats passive
accept", which the pushed work list already departs from on the ground that it is the
projection's own bytes (`worklist.go:23–28`). §IV.2 takes that argument to where it has to go.

### II.3 What a sitting costs (single runs unless stated; none re-run for this document)

| Measure | Value | Source |
|---|---|---|
| Turns per lens sitting, opus | ≈49–56 (two runs agree in magnitude) | `research/2026-08-23_*/cost.md` |
| Turns per blue-respond sitting | 71–136 | same |
| Median sitting, 2026-09-20 | 35 turns, 20 calls; head+tail 5%, turn-to-turn 95% | `cost.go:494` |
| One lens sitting, m11 epoch 3 | 38 calls, 75 turns, **4 adjudications; 34 calls orientation** | m11 investigation (box-local note; the run is `universe-m11`, haiku on both tiers, 1 lane, k-max 2, mint budget 1) |
| Chair read:write | ≈26:7 | same, `:97` |
| Cache read vs output, one keeper run | 237.48M read, 16.44M write, **0.24M output** | `cost.md:26` |
| Mean context per turn | ≈85k (lens), ≈180k (blue); peaks 294k, 256k | run-record-audit `:38,:44` |
| Warm vs cold, 8 matched pairs (#861) | −30% turns, −47% cache write, −36% time, **−6% cost**: run-level cache read +44% | #861 (eight matched pairs; arm B aborted) |
| Turn time decomposition, one run | 43% thinking, **41% round-trip tail** (1,780 small turns), 11% generation, 5% stalls; absolute latencies contaminated by a co-tenant | #684 F11/F14 |
| Empty sittings | 48% across eight runs, median 131.9 s each; 13/42 on 2026-09-22, all 13 re-read the whole report | `cost.go:468`; commit 19a4ddef |
| Board read despite holding gaps | 42 board reads; 12 sittings handed gaps read the board anyway; 6 re-fetched a gap by id while holding it | commit 0b3715e7 (m11) |
| Proofs re-run by hand | 12 `reproduce` calls cost 7 hand runs of the script, 22 proof-file reads, 14 filesystem hunts | commit 2fe606cb (m11) |
| m11 sitting durations, from its 47 seat transcripts (this audit) | lens 114 s median, n=26, max 231; chair 107 s, n=9, growing 53→171 s across epochs; blue 124 s, n=5; bench 163 s, n=2; synthesis 191 s. Summed 88 min; wall 55.7 min | `universe-m11` transcripts, `subagents/workflows/wf_e9af9482…` |
| Work list delivered to blue-researcher | 0 of 8 sittings (its agent type seats three seats) | `worklist.go`, `agentrole.go:35` |

Two things the numbers say together. **The seat pays for its context, not its work**: output
is a rounding error and cache-read is the bill, and #861's warm arm is the direct evidence that
a longer conversation costs what it saves in turns. **And the work the seat does with that
context is mostly finding out where it is**: the orientation share of one sitting was 89% of
calls, and half of all sittings had nothing to do and read the whole report to learn it.

### II.4 What the Workflow tool can and cannot do (verified 2026-09-26)

- `agent()` with `schema` returns a validated object; a seat that returns the wrong shape retries
  at the tool layer. Lens and lane envelopes do not use it today.
- `pipeline(items, …)` runs each item through its stages with no barrier; the script has never
  called it.
- **Concurrency defaults to `min(16, CPUs − 2)` per workflow and is overridable.** On this
  4-core box the default is 2, which is what the 2026-08-23 programme ran at (29 seats strictly
  sequential), what #753 repeats as settled, and what the `DEFAULT_AREAS` comment repeated until #1185. `scripts/universe.sh` sets
  `CLAUDE_CODE_WORKFLOW_MAX_CONCURRENT_AGENTS` (default 16) for every run it launches, and
  m11's own transcripts show it: 7 seats concurrent at peak, the whole lens batch in one wave.
  The box is idle during a run — the time is remote model time — so the cap is not the
  constraint on wall clock. **The epoch chain is**, and §V.0 item 4 has measured it: summing
  the chair, the longest lens, blue and the bench per epoch gives 45.8 of the 48.9 min m11
  spent after its 6.7 min of bookends — **94% of the run is the serial chain**, and the lens
  wave's width changes none of it. The remaining 6% is the gaps between one seat returning and
  the next being spawned.
- Lifetime cap 1000 agents per workflow; no filesystem; no `Date.now()`; resume replays the
  longest unchanged prefix of `(prompt, opts)` from cache. A prompt that interpolates a
  volatile brief misses on resume; a prompt keyed on an item id and a record sequence hits.
- `SubagentStart` delivers `additionalContext` to the seat (measured here; the public hook page
  does not list it). `SubagentStop` cannot speak to the seat and must stay silent (§10 of the
  hook-surface spike: nine firings, nothing delivered).
- Headless: `claude -p --output-format json --json-schema` returns a validated
  `structured_output`; `--agents <json>` and `--plugin-dir` load seats; hooks in settings and
  plugins run in a non-bare `-p` session; `SubagentStart`/`SubagentStop` do not fire for a
  top-level session, and `sittingcap` already defines a sitting by `register` for that reason.
  `--bare` never reads the subscription login — the trap the universe runs paid for in API credits.

## III. The design

### III.0 The one sentence

**The record is the state machine; the engine advances it; a seat is a function the engine calls
at a node that needs judgment, with the node's brief as its argument and a structured decision as
its result.**

Everything below is that sentence applied to one part of the tree. It is deliberately ordered so
that each step is a complete concept on its own ([[complete-the-concept]]), and so that the
measurement in §V.0 can stop the sequence after any step.

### III.1 The work item is the unit — and it already exists

The list `show work` renders is a list of `Item`s (`record/sitting.go`), each with what is owed
and whether it blocks the seat finishing. Today the whole list is one dispatch. The change: **the
engine dispatches one `Item` at a time**, and `Item` gains what a dispatch needs:

| field | what it carries | why a field |
|---|---|---|
| `kind` | the decision this item asks for — `audit-area`, `adjudicate` (a lens grading blue's movement on its gap), `verify-citation`, `respond` (blue on one gap), `propagate` (blue carrying a fix to every site of the claim), `closing`, `rule` (bench), `motion`, `verdict`, `spot-check` | the kind names the verb set and the constitution slice (§III.3) |
| `seat` | the role that decides it | dispatch already computes this |
| `subject` | gap id, claim anchor, citation id, motion id — one of them | the record can refuse a subject that does not exist |
| `as_of` | the record sequence the item was computed at | the premise the seat acts on; a write against a moved premise is refused (§III.4) |
| `brief` | the projection the seat needs: the gap's problem text and acceptance check, the anchored report text, the edits since the seat last sat, the sources cited — rendered by the record, delivered with the dispatch | replaces the orientation calls; it is `show` with the selector chosen by the engine |
| `budget` | calls allowed for this item | the cap becomes per item, small, and refusable |

**Corrected 2026-10-09 — the item's parts have names on the record, and the plan takes them.**
`kind` is the record's `Occasion` (`recordpb/record.proto:1799–1820`): "what a sitting was
convened to do, as against who was asked", today recorded only on the bench's register because
the bench is "the one seat whose id stopped determining its question", and the comment at
`:1934–1948` says where the next such seat's occasion goes — the same field. One item per
dispatch is that collapse applied to every seat: a lens convened to `adjudicate` one gap is not
told apart from a lens convened to `audit-area` by its id. So the kinds in the table above are
values of `Occasion`, the field is extended to every seat's register, and the refusal that today
rejects an occasion from any seat but the bench inverts into "every register states one". The
dispatch row already carries the rest of the item — `seat_id`, `gap_ids`, `occasions`, and the
party's `blockers` (`Dispatch` at `record.proto`, #1239) — so `subject` is `gap_ids` narrowed to
one, and the row is the item. `brief` exists in substance since markers part 4 (#1271): each
work-list gap carries the marker's sentence, passage, backing, `edited_since` and a
`location_state` of `marked`, `gone` or `unrendered`, and the lens prompt says the work list
tells it whether blue moved (`debate.js:1042`). Two consequences from the markers plan: every id
becomes `<LETTER>-<8 hex>` from `crypto/rand` at part 6 (epoch 24, in flight), so `subject` is
unique by construction; and "cut = answer" (its R-9) means an `adjudicate` item takes a `gone`
gap as a decision input, never as silence. Its non-goal "no per-seat view of markers; every seat
sees the same text" bounds `brief`: a window onto the report, never a seat-specific rendering.
Slimming plan 04 phase 2 widens `seat_of_agent` with `occasion`; that is the change the
extension rides on, and the register's "one occasion per sitting" refusal is its test (#1291).

The `kind` census above is derived from the duties the constitutions and prompts already hold
(§II.2, debate.js lines 948–980): it adds no duty. What it removes is the seat's freedom to
choose which of them to do next. Each kind maps to a verb set; a `respond` item may `edit`,
`motion grade file`, and `manifest`; an `adjudicate` item may `regrade`, `close`, `retire`; an
`audit-area` item may `mint` within its budget and `petition`. The record refuses the rest, as
it refuses `close` on a gap another lens minted today.

**Discovery is a kind, not a leftover.** `audit-area` is the whole-area read the lens does now:
it keeps its breadth, because the brief for that kind IS the area. Quantization is for the
lifecycle of a gap once it exists, where every act has one subject. §IV.1 is why this
distinction is load-bearing.

### III.2 The engine reads the record, and a seat's return is a receipt

The Workflow script cannot read a file. Today that constraint is paid with a judgment-tier chair
sitting per epoch. Two changes remove the epoch and most of the cost:

1. **Every envelope is schema'd** (`agent(…, {schema})`), for every role, and carries the event
   ids the seat wrote — a receipt. The engine does not trust the receipt as a fact about the
   board — red audits the record as if it were a claim, and the engine holds its seats to the same standard; it uses it only to decide that the item's
   pipeline may advance. The lens and lane envelopes are the two that are strings today, and the
   petitions and friction they lose are the cost of that.
2. **The dispatch read is a chair item of one kind, at the cheapest tier, with a schema.** Its
   only act is to run `dispatch next` and return the plan as the validated object. It is not
   seat-less: that verb records dispatch rows, it sits on the chair's surface, and since #1192
   any write registers its agent — so it sits under the chair's agent type, with a constitution
   generated for that one verb (§III.3) and no judgment asked of it. It is the engine asking the
   record a question through the only channel the Workflow tool has. The chair's judgment — the
   PASS/FAIL verdict, rulings, closings, the spot-check — becomes items of their own kinds.

**Corrected 2026-10-09.** (1) The envelope argument is stale: #1190 is closed, absorbed by
#1267; #1191 is closed by #1239. Petitions reach the bench from the record — `motion petition
file` readies a bench sitting with occasion `petition` at the next chair plan — and no envelope
carries a log (#1282). Lens and lane envelopes are still prose (`:961`, `:1116`); the chair,
blue, bench, petition, terminal and assemble envelopes are schema'd, and the chair's refuses
everything `requirePlan` refuses (#1281). The receipt principle stands; the cost it was argued
from does not. Envelope and plan schemas generated from Go are slimming plan 09 phase 5's; a
hand-written schema here would be its rival, so this step is that phase. (2) Two rulings bound
the dispatch read and the plan argues against them, not around them. #1281 rejected the engine
reading the record, twice, because the Workflow sandbox has no filesystem; so under vehicle 1
"the engine reads the record" means a seat relays the record's plan, which is what the chair
does today, and `dispatch next`'s prose form now prints the plan as JSON so the relaying seat
composes nothing. The petition ruling (slimming P06.F2, 2026-09-29) chose one epoch of latency
over "a chair re-plan after every wave", costed at two to three judgment sittings per epoch, and
rejected an envelope doorbell as a second channel. The read-after-every-receipt above is the
option that ruling declined. What it was priced against no longer exists: the chair rode the
judgment tier (`debate.js:28`) and composed the plan; since #1281 the relay is a verbatim print
of a record verb, and item 2 above puts it at the bulk tier with one act. The measurement that
reopens the ruling is in §V.0, item 5 (#1292). Until it is run, §III.2 is bound by the one-epoch
latency, and the loop below is an epoch of width one. (3) The no-progress valve is not deleted;
it is being made honest — plan 06 ruled it keys only on fields the relay audit compares against
the record, with `why` leaving the key (#1267, open). #1290 proposes the fold that makes the
audit an equality — store the plan the record rendered on the dispatch event — which is also
what makes a relay a receipt. Under an item loop the valve is per item:
an item re-readied with an unchanged premise is the no-progress condition, which the `as_of`
refusal of §III.4 states exactly.

With both, the loop is: read the ready items → dispatch each as it becomes ready (`pipeline`,
not `parallel`) → on each receipt, re-read. The epoch, `max_epochs`, and the no-progress valve
keyed on identical plans go; the mint budget (roundless §III.B.2.2) and a per-run item budget
bound the run instead. `holdingsInEffect`, `rulingsInEffect`, `reliefInEffect` and the petition
log move to the record and reach a prompt as part of the brief, never from JS memory.

### III.3 The constitution is generated per kind

`agentgen` already inlines the tool's `manual` into each seat's constitution, and `manual`
already renders only that seat's surface. A kind has a smaller surface than a seat: the
`respond` kind needs `edit`, `manifest`, `motion grade file` and the reads its brief did not
already deliver. The generated block shrinks from ~10.7k words to that kind's verbs, and the
hand-written duties split the same way — the craft principles (DEFEND FOCUS, COMPLEXITY MUST
PAY) go with `respond` and `propagate`, the stopping heuristics with `rule` and `verdict`. This
is the same generator with a finer key; `agentgen -check` gates staleness as it does now.

**Corrected 2026-10-09.** This section is slimming plan 07 phase C — a duty registry whose
fragments generate the constitutions, with a gate over the rendered dispatch — and the split
between brief and constitution is already ruled there (P07.F10): a duty conditional on this
sitting's state goes in the prompt, an unconditional one in the constitution. Under this plan
the brief is the conditional carrier and the per-occasion constitution the unconditional one;
nothing here adds a third. Plan 09 phase 3 generates the engine's seat block from Go. Two
generators writing agent-facing text from two declarations is the shape this plan exists to
remove; whether they are one is a question those plans own, and this document does not file it. The numbers this
section lacked: #1284 measured the fixed context per seat at 96–102 KB of definition plus
preloaded skills (the bench 80 KB, preloading nothing), with only the hook-delivered list
per-item; it also recorded that the run-directory map drew a seat to read `report.md`, which
`hookgate` refused — evidence that seats stop reading the whole report by design.

The cost model this buys, stated as a model and not a measurement: a lens sitting of 4
adjudications at ≈50 turns and ≈85k mean context is ≈4M cache-read tokens. Four `adjudicate`
items at ≈8 turns each on a ≈30k context (the kind's constitution plus the brief) is ≈1M. The
win is on context per turn and on the orientation turns the brief pre-empts; it is not on the
round-trip tail, which is per turn and unchanged. §V.0 measures whether the model holds.

### III.4 Premise refusal — concurrency lives at the record

Two `respond` items on gaps that anchor to the same section will run concurrently. Today that
cannot happen because blue sits alone. The rule that makes it safe is one the record already
half-holds: **a write names the `as_of` it was computed against, and the record refuses a write
whose subject moved since** — the anchored text was edited, the gap was regraded, the motion was
ruled. The half it holds today is by content: an anchor that is no longer in the report is
refused as "a stale reference or belongs to another run" (`anchor/window.go:93`). The half it
lacks is by sequence: `events.id` is the record's append-only order (`recordsql`), and `as_of`
is the highest id the brief was rendered from. The refusal names what moved and the engine
re-dispatches the item with a fresh brief.
This is the pushed-projection objection of `tool-is-the-contract` §IX answered structurally:
a delivered brief is allowed to be stale precisely because staleness is refused at the write
rather than discovered at the read. It is also the concurrency-namespace lesson of
[[facts-are-fields]] clause 4: the thing that makes parallel writes safe is a field, and it must
be one the record can see.

### III.5 What the seat stops being told

The seat prompt shrinks to the kind, the subject, the brief, and the return schema. "Read
`sitting.last_sitting` first and branch on its kind" and "pull your working set in one pass" —
the instructions that exist to orient a self-directed sitting — go, because the brief is the
orientation. The flat "re-read the FULL report" order already left with 19a4ddef, which scoped
the re-read by the last sitting's kind; what remained was the seats' distrust of the list
itself, which 0b3715e7 ruled on: the envelope is the rig and is not questioned, the contents
are speech and are. A brief inherits that ruling exactly — its `kind`, `subject`, `as_of` and
`budget` are the instrument's, and the gap text it carries is another lens's words, verified at
the leaf as before. `audit-area` keeps the full read, because there the full read is the item.

### III.6 The vehicle — Workflow first, a Go driver only on a condition

Four vehicles were considered; the order is a recommendation, not a survey.

1. **Workflow, restructured** (§III.1–III.5). Everything above fits inside it: `pipeline` over
   items, schema'd envelopes, the cheap dispatch read, `SubagentStart` delivering the brief.
   Cost: one small model call per engine tick; resume via the agent cache keyed on
   `(kind, subject, as_of)`. **Recommended first**, because every step is a diff to files that
   exist and the release surface does not move.
2. **A Go driver over `claude -p` per item** (`feov-record run`). Reads SQLite directly, needs
   no dispatch-read call, schedules on the record, resumes from the record (better than the
   agent cache: what is done is what the record says is done), and its concurrency is bounded
   by the box and the rate limit rather than by `CPUs − 2`. Cost: every seat is a top-level
   session, so the sitting span moves to `SessionStart`/`SessionEnd` (the code already defines a
   sitting by `register` for headless seats); gray-area's capture binding and the
   `subagents/workflows/wf_<id>/` transcript layout both change; each item pays a process start
   with plugins loaded; `--bare` is not usable on the subscription. The CLAUDE.md warning that a
   harness must not stand in for the engine was earned by a *test* rig; a driver that IS the
   engine does not dodge it but must re-earn every hook fact on the new event surface.
   **Condition to take it:** §V.0 shows the dispatch-read overhead is the binding constraint,
   or the subscription's rate limit is reached below the width the items could use.
3. **The Agent SDK** (TypeScript or Python): in-process hooks, structured output, `canUseTool`.
   Refused by the repository's rule that committed tooling is Go; there is no Go SDK. Noted so
   the refusal is on the record, not so it can be reopened silently.
4. **A `claude -p` pool spawned from inside a Workflow seat's Bash.** Rejected: it is exactly
   the rig the CLAUDE.md warning describes, fires no subagent hooks, and nests a session inside
   a seat.

On parallelism specifically: the Workflow tool already runs 16 seats wide when told to
(§II.4), and the box does nothing while a seat thinks, so vehicle 2 buys no concurrency
vehicle 1 lacks. What bounds wall clock today is the epoch chain, and that is what §III.2's
`pipeline` over items removes; §IV.5 is the estimate.

**Added 2026-10-02 (#1222) — a fifth channel, and the harness boundary.**

*The channel.* Every non-bare Claude Code session binds an inbox socket
(`~/.claude/sessions/<pid>.json` → `messagingSocketPath`; `$CLAUDE_CODE_MESSAGING_SOCKET` inside
its hooks and Bash). The wire is newline-delimited JSON: one optional auth line carrying
`$CLAUDE_CODE_MESSAGING_TOKEN`, then one stream-json user message. Proven 2026-09-28 from a
Python process: the probe arrived in the receiving session mid-turn as a peer message. A post
costs no model call, which is the cost the dispatch read above (§III.2) was paying. Measured
2026-10-02 in this session: a subagent spawned here reports the PARENT's socket path and no
agent-scoped one. So the socket addresses a **process-level session** — a `claude -p` seat — and
not an in-process Workflow seat, whose only inbox is the lead's. The channel therefore belongs
to vehicle 2 and does not create a hybrid of vehicle 1: "Workflow spawns and caps the seats, the
record drives them over the socket" has no address to drive. What it changes in the table above:
vehicle 2's "needs no dispatch-read call" is now a push into a seat that is already sitting, so a
`-p` seat can take item after item without a process start per item, and the brief for the next
item rides the same wire the first one did. Limits, measured: a post has no reply address (the
receipt is the seat's record write, which §III.2 already makes the return); an unattended `-p`
seat accepts an unknown sender only with `--settings '{"crossSessionInbound":"accept"}'`, while a
post from the seat's own child process is delivered without hold on Linux; `--bare` binds no
socket. The other ways in are a full turn (`--resume`), the spawned seat's own stdin
(`--input-format stream-json`), or a cloud session (`claude -p --cloud <id>`); none reaches a
running local session for less than the socket does.

*The ruling.* gblock, 2026-10-02: every host API the rig drives goes behind an abstraction, so
that the Antigravity port swaps the host without touching the rig. The vocabulary already exists
and is extended, not rivalled: `plans/dual-target-claude-antigravity.md` names the discriminator
(`Harness`, values `claude` and `antigravity`) and `hookcore` as the normaliser for hook
payloads; `record.HarnessSeat` is the pseudo-seat the host writes under today. The rig's
seat-control surface gets the same shape in FEOV's tools module — one interface, one
implementation per `Harness`, and above it only items, seats, briefs and receipts:

| Operation the rig needs | `claude` carrier today | `antigravity` carrier (from the dual-target plan's measurements) |
|---|---|---|
| Spawn a seat with a brief | Workflow `agent()`; or a `claude -p` process | `invoke_subagent`, or an `agy -p` process |
| Deliver context at seat start | `SubagentStart` → `additionalContext` (measured to land, §II.2) | `SessionStart` → `injectSteps` |
| Post into a sitting seat | the inbox socket (process-level only, above) | not measured — `PreInvocation` → `injectSteps` is pulled per turn, not pushed |
| Identify the seat to the record | `PreToolUse` env rewrite (`FEOV_RUN`, `FEOV_AGENT_ID`) | `conversationId` on every payload |
| Observe the sitting's end | `SubagentStop` / `Stop` | `Stop` (`terminationReason`) |
| Cap the sitting | `sittingcap` in `PreToolUse` | `PreToolUse` deny (`decision: "deny"`) |

Three consequences for the vehicles. First, nothing above the interface may name a socket path,
an environment variable, a hook event, or a JSON key — the gate is a test that greps the rig's
packages for the `claude` carrier's names, the same shape as `pluginparity`'s import gate.
Second, the interface is written from the carrier census, not from the table: today's direct
readers of Claude Code's hook and transcript shapes in FEOV's `internal/` (non-test) are
`record/recordsql/views.go` (31 sites), `sittinghook/sitting.go`, `record/seattiming.go`,
`record/seatturn.go`, `hookgate/hookgate.go`, `hookcmd/hookcmd.go`, `record/recordsql/schema.go`,
`record/migrate/sqlitesource.go`, `record/agentbinding.go` and `record/recordsql/store.go`, with
`CLAUDE_CONFIG_DIR` read at ten sites and `CLAUDE_PROJECT_DIR` at three; each is either moved
behind the boundary or recorded as host-neutral with the reason. Third, the Antigravity column
is a hole where it says *not measured*: the mid-sitting post has no proven carrier there, so a
rig that depends on it is a rig that runs on one host until that cell is measured the way the
Claude one was. The boundary is not a reason to build vehicle 2 sooner; §V.0 still gates it.

**Corrected 2026-10-09 — relay versus read.** #1281 settled that under the Workflow tool the
engine never reads the record: a seat relays its plan. Under vehicle 2 the rig reads the record
itself, and it is now also the only vehicle the inbox socket reaches. The vehicle question is
therefore not "which is faster" (§II.4 answered that) but "does the record drive the seats, or
does a seat relay the record to a script that drives them", and the petition-latency ruling
(§III.2) is the price of the second answer. Three open defects are rows of the boundary table's
`claude` column and are owned there, not here: a hook bracket opens a sitting and binds a seat
(#1262), `SubagentStart` does not fire on a resumed dispatch (#1111), and two open runs record
nothing (#1255). The slimming survey's D4 — stop injecting `FEOV_RUN` — would remove the identity
row's carrier; identity then comes from the dispatch through the boundary or not at all, which
is why D4 and #1255 are one decision (commented there 2026-10-09).

## IV. Risks & Objections

### IV.1 The punch list crowds out the survey — the strongest objection

`docs/why-a-seat-stops.md`: "from that moment my sitting had a shape — a punch list to zero out,
authored by the tool. Things off the list weren't declined. They were invisible." One item per
dispatch is the limit of that mechanism. Everything discretionary — the contradiction between two
sections nobody minted, the motion nobody filed — is off every item's list.

Answer, and it is the reason `audit-area` is a kind and not a residue: discovery is dispatched
with a whole-area brief and a mint budget, at the moments the record says an area moved. What
the record cannot say — "something is wrong that no item points at" — reaches the engine only
through the petition right — which since #1239 is a record route (`motion petition file`
readies the bench), so a lens's petition no longer depends on its envelope's shape [corrected
2026-10-09; the sentence it replaces argued from the string envelope #1190 described]. If §V.0's discovery measure falls
(new gaps minted per area per run), this objection was right and the sequence stops at §III.2.

### IV.2 A delivered brief is a pushed projection

`tool-is-the-contract` §IX. Answered in §III.4: the brief may be stale; the write may not. The
principle was written against summaries a seat must *trust*; a brief the record rendered and will
refuse a write against is not a summary, it is the projection's own bytes with a version.

### IV.3 Fixed cost multiplies

Every dispatch pays its constitution. §III.3 shrinks it per kind; the subagent cache holds the
prefix for five minutes, so a burst of same-kind items reads it from cache. Not measured; §V.0.

### IV.4 Cross-item judgment

A single `respond` cannot see that its fix contradicts another gap's fix. `propagate` is the
kind that exists for that, minted by every accepted `edit`; the bench's `rule` and the chair's
`verdict` keep whole-board briefs. What is lost is the incidental noticing a long sitting does
for free; §V.0 counts it (defects found outside any item, per run) so the loss is a number.

### IV.5 Item explosion, and the wall-clock estimate

The item count roughly triples — an m11-scale run of 47 sittings becomes ≈120 dispatches: ≈19
area audits, ≈48 gap items, ≈10 chair items, 6 bookends, ≈40 dispatch reads — and each item is
shorter. The baseline is m11 as measured from its transcripts (§II.3): haiku on both tiers,
88 min of summed sitting time, 55.7 min of wall clock, of which 6.7 min is the bookends and
45.8 min is the serial chain. The per-kind figures are a model §V.0 replaces: an area audit is a
whole lens sitting (114 s median); a response, an adjudication or a ruling on one subject
≈40–60 s at haiku (8–12 turns on a context a third the size); a read ≈10 s; a per-dispatch
floor ≈15 s. Summed, ≈90–100 min against today's 88.

Wall clock is max(summed ÷ concurrency, the longest serial chain). At 16 wide the sum is not the
bound; the chain is: the bookends, unchanged (6.7 min; they shorten only if audits may start on
lane drafts before synthesis, which the rig permits and today's script does not) + the longest
gap's lifecycle (mint inside an area audit ≈2 min → respond → adjudicate → respond →
adjudicate → close, ≈4 items × ≈50 s with a read between each, ≈6 min) + verdict and assemble
(≈3 min) ≈ **16–18 min at haiku, against the measured 55.7**. That ≈3× is the epoch leaving,
not the items getting faster; the 41% round-trip tail is per turn and unchanged. Every figure
here is at haiku's ≈3.4 s per turn; a sonnet or opus tier lengthens both columns by the same
factor, so the ratio is the claim and the minutes are not. At the un-overridden default of 2
the sum binds instead and the rig is no faster than today.

Two costs inside that estimate. **The dispatch read is a design variable**: one read per receipt
spends a slot and ≈10 s of chain per item; the estimate assumes one read per tick with every
ready item dispatched from it. **Rate limit**: sixteen seats thinking at once is sixteen
concurrent model calls on one subscription, and the limit has not been measured.

Batching same-kind items on one subject's neighbourhood is the fallback if the floor per
dispatch is higher than modelled, and it is a partial return to the sitting.

### IV.6 The record grows a state machine it only half has

`as_of` refusal, the item kinds, the per-item budget, holdings and relief as record facts: each
is a schema change and an epoch bump, each migrates archived runs through `migrate`, and none
of it is a compatibility path. That is the repository's rule and the cost is real.

## V. Verification — what proves any of this

### V.0 Measure before building (the gate on every step below)

Instrument the current engine, no design change, on one dev run and one smoke:

1. Per sitting: items handed, items completed, calls that were orientation vs act (a read of a
   thing the work list already carried is orientation). *Re-arms on any run.* Nothing measures
   "items handed vs completed" today (§II.3).
2. Per dispatch: the fixed prefix (constitution + skills) in tokens, and the brief a kind would
   need, in tokens — computable from the record for the `respond` and `adjudicate` kinds
   without building them.
3. Discovery baseline: gaps minted per `audit` sitting per area, and defects found outside any
   engaged gap.
4. **Done 2026-09-27**, from m11's 47 seat transcripts: peak concurrency 7; the chain — chair +
   longest lens + blue + bench per epoch — is 45.8 of the 48.9 min after the bookends. §IV.5
   stands on it. #1124 (open) is the chain defect stated as a defect: lenses are dispatched
   before blue responds in the same epoch and verify repairs that do not exist — the epoch as
   quantum, measured on a run. Method: per seat transcript under the workflow's directory, the first and last
   record timestamps are the sitting's span and the seat prompt's opening words name the role;
   chair starts delimit epochs; the chain is summed inside each. *Re-arms on any new run.*
5. **Added 2026-10-09 (#1292).** The relay-only chair sitting, priced: dispatch a chair at the
   bulk tier whose only permitted act is `dispatch next` and the relay, on the wave A archive
   (PR #1288) or m12; take its cost from the transcript's usage and its wall clock; compare to a
   full chair sitting on the same record. Of the order of one call: the petition-latency ruling
   was priced on a seat that no longer exists and §III.2's per-receipt read is priced, not
   refused. Otherwise the ruling stands and §III.2 is an epoch of width one. *Re-arms on any
   change to the chair's constitution or preloads.*

The decision rule: if orientation is under a third of calls, or the modelled per-kind context is
not under half the sitting's mean, the cost case fails and only §III.2 (the schema'd envelopes
and the cheap dispatch read) stands on its own merits.

### V.1 Per step

- §III.2: `TestNoPromptGrowsItsCommandCatalogue` and the seat-prompt goldens; a run whose
  petitions from a lens reach the engine (today: zero, structurally); no `holdingsInEffect` in
  `debate.js`. *Re-arms on debate.js, the envelope schemas.*
- §III.1/III.4: a record test where two `respond` writes race on one anchor and the second is
  refused naming the first; the `as_of` field on `Item` in `record.proto` with `migrate` on an
  archived run. *Re-arms on record.proto, dispatch.go, sitting.go.*
- §III.3: `agentgen -check` green per kind; the `manual` golden per kind read line by line.
- §III.6 (added 2026-10-02): a test that fails any rig package naming a `claude` carrier — a
  socket path, a `CLAUDE_*` or `FEOV_*` variable, a hook event name, a hook JSON key — outside the
  `claude` implementation; the Antigravity cell for the mid-sitting post is *not measured* until
  a probe posted from a plain process arrives inside a running `agy -p` seat, the way the socket
  probe did on 2026-09-28. *Re-arms on any new reader of a hook payload or `CLAUDE_*` variable.*
- The engine still runs: `universe.sh build` + `run` on `--smoke`, the installed copy grepped
  for a phrase from the change, `stream-json` watched, never `--output-format json`.
- `FEOV_RELEASE_GATE=1 go test ./releasegate/fuzz/` by hand before any record or hook-shape
  change is called done.

## VI. Deliberately not in this document

- Any change to what a lens, blue, or the bench judges. The kinds are the duties that exist.
- The sleeper-service loop and `/self-improve`.
- A move off the subscription, or a bigger box. §III.6 names the condition; the human sets it.
