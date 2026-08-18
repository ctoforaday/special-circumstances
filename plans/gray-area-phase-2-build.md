# Gray Area Phase 2 — implementation spec

> Parent: [`gray-area-phase-2.md`](gray-area-phase-2.md) — the scope proposal and the two rulings.
> That document argued for a narrower scope; the operator chose **Phase 2 whole, gated on #189**,
> and #189 closed on 2026-08-18 (`gray-area.md` §11.11). **The gate is discharged. This is the
> build.**
>
> The proposal's §5 carrier census is short and its §6 loop is written against a scope that no
> longer applies. Both are restated here rather than referenced, because a spec that sends the
> reader to a superseded document for its own acceptance criteria has no acceptance criteria.

## 1. What Phase 2 whole is

From `gray-area.md` §7: *"The miner — Go CLI over the manifest: act-vs-claim, rework, stalls.
Provenance on every answer."*

Three inspections. One of them — act-vs-claim over **checkpoint notes** — shipped early under
Phase 5 as `internal/claims`, and its verdict model is reused unchanged rather than re-argued.

| # | verb | claim source | what it answers |
|---|---|---|---|
| P2-1 ✅ | `gray-area rework <transcript>` | none — pure trajectory | the same file written again and again; the same command re-run after a claimed fix |
| P2-2 ✅ | `gray-area stalls <transcript>` | none — pure trajectory | repeated identical invocations; the [[anti-spinning]] 3-strike rule adjudicated against what happened rather than against the hook counter |
| P2-3 ✅ | `gray-area pr <body.md> [transcript]` | a pull request body | act-vs-claim over the body, per the operator's ruling: **full post-hoc adjudication** |

## 2. The unsolved question, and the answer

`gray-area-phase-2.md` §0 left one thing open and named it the builder's problem:

> a false positive on an accurate body puts the operator between two of the agent's outputs.
> §6's check 4 covers the absence case; a PR-body reader will need its own answer for the
> false-positive case.

### 2.1 The boundary that answers it

**The trajectory records what was RUN. It does not record what the run SAID.**

`trajectory.Parse` emits `tool_result` events carrying a `tool_use_id` and nothing else — no
output, by construction, because the result bodies are conversation content and this plugin does
not copy conversation content (G6). So:

| claim | adjudicable? |
|---|---|
| "I ran `go run ./check`" | **yes** — a Bash invocation either contains the signature or does not |
| "`go run ./check` → 26 passed, 2 failed" | the ACT half yes; **the RESULT half is not in the record at all** |
| "CI green on all 11 jobs" | **no** — no act in this session produced it |
| "carriers swept: `plans/gray-area.md`" | **yes** — a write to that path either happened or did not |

So every claim decomposes into an **act** and a **result**, and *the miner may adjudicate only
the act*. A verdict on the result half would be manufactured from a record that does not contain
the fact — which is precisely how a false positive gets made, and it would be made confidently,
in the tool built to catch exactly that.

### 2.2 What the tool does instead

**The unmeasurable half is REPORTED, not dropped.** Each finding carries `Unmeasured` — the parts
of the claim this record cannot speak to, named. A row that says

> CITED · `go run ./check` · *not measured: the asserted outcome ("26 passed, 2 failed") — the
> trajectory records invocations, not their output*

cannot be read as vindicating the count, and cannot be read as convicting it either. Silence on
the result half would collapse into the CITED and read as a full endorsement.

That is `facts-are-fields` clause 3 turned on this tool's own output: **make the miss loud rather
than let it fold into the honest zero.** It is the same move `NO-EVIDENCE` already makes for the
act half — state the search, let the reader convict — applied to the half where no search is
possible at all.

### 2.3 No conviction verdict, and why that is a decision

There is no `CONTRADICTED`. It was considered and rejected: without result bodies, the only claims
the record could positively contradict are negative ones ("I did not touch X"), which PR bodies
essentially never make, and `STALE` already covers supersession. **A conviction verdict with
almost no safe trigger is a loaded gun with a hair trigger** — the first false positive would come
from the one exotic path nobody tested, and it would arrive wearing the tool's authority.

If result bodies are ever captured, this decision must be re-opened rather than inherited. Noted
here because a rejected design that leaves no trace gets re-proposed every six months.

## 3. Verdict model — reused, not re-argued

`claims.Verdict` stands as shipped: `CITED` / `STALE` / `NO-EVIDENCE` / `UNCHECKABLE`. The
`NO-EVIDENCE` asymmetry was argued once (a matcher too narrow makes a confident false accusation;
one too broad makes a citation a reader can see is wrong — so state the search and let the reader
convict) and is not re-argued per inspection.

`Finding` gains exactly one field, `Unmeasured []string`, per §2.2.

## 4. Carrier census

[[complete-the-concept]] requires this in the spec, and requires the auditor to **re-run each
census rather than trust this list**.

| carrier | change |
|---|---|
| `internal/trajectory` | a grouping layer — repetition and ordering over `ToolUses()`. Nothing groups today |
| `internal/claims` | `Finding.Unmeasured`; the `Verdict` set is untouched |
| `internal/prbody` (new) | reads a pull request body into `claims.Claim`s, splitting act from result |
| `cmd/gray-area` | three new verbs beside `tools` and `checkpoint`; the usage text and its goldens move |
| `plugins/gray-area/README.md` | the shipped documentation |
| `commands/` | a command doc per new verb, on the model of `audit-checkpoint.md` |
| `.claude-plugin/plugin.json` | **no version bump** — versions move at a release boundary (#444) |
| `scripts/check/gates.go` | only if a new CI gate appears |
| `plans/gray-area.md` §7 | the phase table's Phase 2 row |
| `plans/gray-area-phase-2.md` | §1's "three inspections bundled" framing, once built |

**Seat coverage is NOT in scope and must not be claimed.** #469 is open: one subagent transcript
exists on disk that no manifest row ever named, so "every seat was inspected" is unproven in the
missed-seat direction. Any Phase 2 output that states seat coverage as a number is blocked on it.

## 5. Validation loop, written before implementation

Per [[validation-loop]], with what re-arms each:

1. `go test -C plugins/gray-area/tools ./...` — clean · re-armed by `plugins/gray-area/tools/`
2. `plugins/gray-area/bin/gray-area checkpoint .claude/checkpoints/CHECKPOINT.md` — still exits 0
   with every claim adjudicated · re-armed by `internal/claims/` — **the Phase 5 check must not
   regress while its engine is extended**
3. Every new inspection refuses to emit a row whose `Cited()` is false — asserted **per
   inspection**, not once · re-armed by `internal/trajectory/`
4. **The plausible-zero gate.** An inspection run against a transcript with zero matching events
   reports its search and its `EventsSeen`, and is distinguishable in output from one that found
   nothing wrong · re-armed by `internal/claims/`
5. **The act/result gate.** A PR-body claim carrying an asserted outcome emits a non-empty
   `Unmeasured`, and a CITED row with an unmeasured result half says so in the table output, not
   only in the JSON · re-armed by `internal/prbody/` — this is §2's answer, and it is the one a
   summary will drop first
6. `go run ./golden -review` — the help contract moves deliberately · re-armed by the goldens
7. `go run ./check` (from `scripts/`) — the whole gate set; standing baseline 26 passed, 2 failed
   (the `example.com` 403 pair, which passes in CI), 4 not run here

Checks 4 and 5 are the two worth stating loudly: 4 is what makes an empty result honest, and 5 is
what makes a full result honest.

## 6. Sequencing

P2-1 and P2-2 share the grouping layer and neither needs a claim source, so they land together.
P2-3 lands after, because it is the one that needs §2's boundary built into it.

**All three shipped 2026-08-18.** Two things the real data taught, both recorded where they will be
read again rather than only here:

- `CommandKey` was falsified TWICE by this repository's own session after passing every
  hand-written test. Truncating a command is what invents groups; the key now keeps the whole
  command and display truncation is the caller's job. See `internal/trajectory/repeat.go`.
- The PR-body reader attached each line's LAST outcome to every claim on that line — a caveat
  naming the wrong number, which is worse than no caveat. Fixed by bounding the tail at the next
  code span. See `internal/prbody/prbody.go`.

**A pull request boundary is not a completion boundary** ([[complete-the-concept]]). This document
is the thread: the concept is finished when all three verbs exist, the census in §4 is re-run, and
`gray-area.md` §7's phase table stops describing Phase 2 as unbuilt.
