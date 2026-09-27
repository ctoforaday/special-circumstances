# The rig — one work item per dispatch

> STATUS 2026-09-26: proposed — a research document and a direction, not an implementation plan.
> Written against `d4fea590`. It asks for one ruling (§I, "The question") before any step in §III
> is started. Every number in §II is from the run it names, and the ones marked *single run* have
> not been reproduced; §V.0 is the measurement that has to come first.

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
  (§II.2). Seats re-read the whole report in 13 of 13 empty sittings.
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
is needed? The tree has been answering "rig" one mechanism at a time since #500. Saying it out
loud changes what the next ten changes are, and it changes what the Workflow script is *for*.

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

### II.1 The engine (`debate.js`, 1147 lines)

The script calls `agent()` and `parallel()` only; `pipeline()` is stubbed in the goja harness and
unused. The bookends are fixed: frontier (1 agent), lanes (N in parallel), synthesis (1),
terminal bench (0–1), assemble (1). The debate is `while (!halted)` at line 985, one **epoch**
per iteration, in fixed order:

1. The chair sits (judgment tier). It runs `dispatch next` and relays the JSON as `plan`.
2. Lenses sit in one `parallel()` batch — the only concurrency in the loop.
3. Petitions are heard serially. Blue sits, serially, on every gap in the plan. The bench sits.

Three structural facts follow from the script's own comments.

- **The script reads nothing.** "debate.js reads no record" (lines 49–51, 134–135, 367–368).
  Every fact the engine acts on is a seat's envelope relay: the chair's `plan`, `verdict` and
  `unruled_motions`; blue's `claim_count`, `found_closed`, `manifest`; the bench's
  `dispositions` and `holdings`. The Workflow tool has no side-effect primitive, so a
  deterministic read costs a model call, and today it costs a judgment-tier one.
- **The epoch is the quantum.** The plan is computed once, at the top; a lens made ready
  mid-epoch waits for the next chair; blue is dispatched on gaps that closed while it waited.
  The `for … of blues` loop runs at most once because `engage()` groups every gap by seat — a
  plan holds at most one blue party and one bench party, and that party carries all its gaps.
- **JS memory holds facts the record should.** `holdingsInEffect`, `rulingsInEffect`,
  `reliefInEffect`, `friction`, `petitionLog` (lines 673–783) are interpolated into later
  prompts and are lost on a throw unless the agent cache replays. Relief is described as
  "operative this sitting" and only ever grows.

Two defects found while mapping, filed here so they are not lost: lens envelopes are
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
| The hard stop | `sittingcap`, default 150 calls; the hook refuses past it and lets only `register` through | the only thing that forces a return |
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
| One lens sitting, m11 epoch 3 | 38 calls, 75 turns, **4 adjudications; 34 calls orientation** | m11 investigation, box-local scratch note `:21` |
| Chair read:write | ≈26:7 | same, `:97` |
| Cache read vs output, one keeper run | 237.48M read, 16.44M write, **0.24M output** | `cost.md:26` |
| Mean context per turn | ≈85k (lens), ≈180k (blue); peaks 294k, 256k | run-record-audit `:38,:44` |
| Warm vs cold, 8 matched pairs (#861) | −30% turns, −47% cache write, −36% time, **−6% cost**: run-level cache read +44% | #861 (eight matched pairs; arm B aborted) |
| Turn time decomposition, one run | 43% thinking, **41% round-trip tail** (1,780 small turns), 11% generation, 5% stalls; absolute latencies contaminated by a co-tenant | #684 F11/F14 |
| Empty sittings | 48% across eight runs, median 131.9 s each; 13/42 on 2026-09-22, all 13 re-read the whole report | `cost.go:468`; commit 19a4ddef |
| Board read despite holding gaps | 12 sittings; 6 re-fetched a gap they held | m11 investigation, box-local scratch note (PR #1178 lineage) |
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
  sequential) and what #753 and `debate.js:630` repeat as settled. `scripts/universe.sh` sets
  `CLAUDE_CODE_WORKFLOW_MAX_CONCURRENT_AGENTS` (default 16) for every run it launches, so m10,
  m11 and the smoke ran with the whole lens batch in one wave; the trajectory catalogue shows 8
  subagents acting in the same minute on this box. The box is idle during a run — the time is
  remote model time — so the cap is not the constraint on wall clock. **The epoch chain is**:
  chair → lens wave → blue → bench is serial by construction, and m11's 56 min over 9 epochs is
  about 6 min per epoch however wide the lens wave is.
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

The `kind` census above is derived from the duties the constitutions and prompts already hold
(§II.2, debate.js lines 946–980): it adds no duty. What it removes is the seat's freedom to
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
2. **The dispatch read is its own dispatch, at the cheapest tier, with a schema.** Its only act
   is to run `dispatch next` and return the plan as the validated object. It has no duties, no
   constitution and no judgment; it is the engine asking the record a question through the only
   channel the Workflow tool has. The chair keeps its judgment — the PASS/FAIL verdict, rulings,
   closings, the spot-check — and each of those becomes an item of its own kind.

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
ruled. The refusal names what moved and the engine re-dispatches the item with a fresh brief.
This is the pushed-projection objection of `tool-is-the-contract` §IX answered structurally:
a delivered brief is allowed to be stale precisely because staleness is refused at the write
rather than discovered at the read. It is also the concurrency-namespace lesson of
[[facts-are-fields]] clause 4: the thing that makes parallel writes safe is a field, and it must
be one the record can see.

### III.5 What the seat stops being told

The seat prompt shrinks to the kind, the subject, the brief, and the return schema. "Read
`sitting.last_sitting` first and branch on its kind", "pull your working set in one pass",
"re-read the FULL report" — the instructions that exist to orient a self-directed sitting — go,
because the brief is the orientation. The measured contradiction between `critical-stance`'s
"re-read the full report" and the work item's "audit what moved" (m11) dissolves for every kind
but `audit-area`, where the full read is the item.

### III.6 The vehicle — Workflow first, a Go driver only on a condition

Four vehicles were considered; the order is a recommendation, not a survey.

1. **Workflow, restructured** (§III.1–III.5). Everything above fits inside it: `pipeline` over
   items, schema'd envelopes, the cheap dispatch read, `SubagentStart` delivering the brief.
   Cost: one small model call per engine tick; concurrency 2 on this box; resume via the agent
   cache keyed on `(kind, subject, as_of)`. **Recommended first**, because every step is a
   diff to files that exist and the release surface does not move.
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
   **Condition to take it:** §V.0 shows the dispatch-read overhead or the concurrency cap is the
   binding constraint on a box where more than 2 seats can actually run.
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

## IV. Risks & Objections

### IV.1 The punch list crowds out the survey — the strongest objection

`docs/why-a-seat-stops.md`: "from that moment my sitting had a shape — a punch list to zero out,
authored by the tool. Things off the list weren't declined. They were invisible." One item per
dispatch is the limit of that mechanism. Everything discretionary — the contradiction between two
sections nobody minted, the motion nobody filed — is off every item's list.

Answer, and it is the reason `audit-area` is a kind and not a residue: discovery is dispatched
with a whole-area brief and a mint budget, at the moments the record says an area moved. What
the record cannot say — "something is wrong that no item points at" — reaches the engine only
through the petition right, which is why every envelope carries `petitions` as a field (§III.2)
and why losing them in a string envelope is a defect today. If §V.0's discovery measure falls
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
shorter. A model, not a measurement (the per-kind turn counts are guesses §V.0 replaces): an
area audit ≈25 turns, a response ≈10, an adjudication ≈6, a read ≈10 s, a per-dispatch floor of
≈15 s for spawn plus two turns, 3.4–6.6 s per turn. Summed, ≈105 min against today's ≈100 min
of summed sitting time.

Wall clock is max(summed ÷ concurrency, the longest serial chain). At 16 wide the sum is not the
bound; the chain is: frontier → lanes → synthesis (≈10 min) + the longest gap's lifecycle
(mint → respond → adjudicate → respond → adjudicate → close, ≈5–6 min with a read between each)
+ verdict and assemble (≈3 min) ≈ **18–20 min**, against m11's measured 56 min, which is
9 epochs of a serial chair → lens → blue → bench. That ≈3× is the epoch leaving, not the items
getting faster; the 41% round-trip tail is per turn and unchanged. At the un-overridden default
of 2 the sum binds instead and the rig is no faster than today.

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
4. From m11's record: achieved concurrency per epoch, and the critical chain — the sum of the
   chair, the longest lens, blue and bench per epoch — against the run's wall clock. If the
   chain is not most of the 56 min, §IV.5's estimate is wrong.

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
- The engine still runs: `universe.sh build` + `run` on `--smoke`, the installed copy grepped
  for a phrase from the change, `stream-json` watched, never `--output-format json`.
- `FEOV_RELEASE_GATE=1 go test ./releasegate/fuzz/` by hand before any record or hook-shape
  change is called done.

## VI. Deliberately not in this document

- Any change to what a lens, blue, or the bench judges. The kinds are the duties that exist.
- The sleeper-service loop and `/self-improve`.
- A move off the subscription, or a bigger box. §III.6 names the condition; the human sets it.
