# Plan: one generated constitution per seat, and registration reduced to informational

> **Status:** Design under review. Not implemented. Arrives as a PR per the `plans/` convention.
>
> **Origin:** measured on the 2026-09-19 universe smoke — a completed, `verified`, 69-seat run.

## The measurement this rests on

Across one completed run (69 seat transcripts, 1,702 tool calls):

| | |
|---|---|
| tool calls spent obtaining the seat manual | **171** (91 `manual` invocations + 80 reads) = **10% of all calls** |
| bytes of manual read into seat context | **4,280,752** (~1,070,188 tokens) = **92% of all Read traffic** |
| seats making a `--help` call after reading the manual | **7 of 69**; 8 calls total |
| `feov-record` calls relying on hook-injected `--run` | **83%** (1,097 of 1,329) |
| tool error rate | 6.4% (109 / 1,702) |

**The manual mechanism works and must be preserved.** Only 8 `--help` calls in a whole run is the
evidence: per-verb help was ~16 calls per seat, `manual` costs 2.48. A "surface listing plus
on-demand help" design was considered and **rejected** — it re-creates the problem `manual` was
built to solve, and the 62 seats currently asking for nothing would start asking.

What is wasteful is not the content but the **fetch**: a document identical for every seat of a role,
unchanged between runs, obtained at runtime in 2–7 tool calls, each of which is a turn that re-reads
the seat's whole context. Some seats run `manual` two or three times; several read one window past
EOF and get 4 bytes back.

## The shape

**One generated constitution per seat**, with as much stated statically as possible.

Today's agent definitions do not map 1:1 to seats, which is the reason the naive fix fails:

| definition | seats it serves |
|---|---|
| `blue-researcher` | `frontier`, `blue-lane-N`, `blue-respond` |
| `lead-judge` | `judge`, `judge-terminal`, `judge-petition-*`, `assemble` |
| `blue-synthesizer` | `blue-synthesize` |
| 7 lenses, `red-chair` | 1:1 already |

`--seat-id` selects the surface, so these seats have **different verb sets under one definition**.
Appending the manual to the existing files would hand `judge` the whole of `assemble`'s surface as
dead weight. So the mapping is flattened: **agent definition ≡ seat**, one file each, generated.

### Sources, all of which already exist and are already gated

- `record/roster.go` `seatShapes` — the canonical seat grammar, with a drift test
  (`TestTheRosterMatchesWhatTheEngineActuallyDispatches`) that checks it against `debate.js`'s real
  dispatch sites in both directions.
- `feov-record --seat-id <seat> manual` — the surface, rendered live by the binary that owns it.
- A hand-authored **role template** per role (lens / chair / blue / bench), carrying the duties that
  are genuinely authored prose.

### The generator

`scripts/agentgen`, following the idiom of `vocabdoc`, `marketplacegen`, `schemagen`, `massgen`,
`protogen` — generate by default, `-check` fails a stale commit, and a gate in `scripts/check/gates.go`
with a `why:` line. This is the closest thing the repository has to a Blaze genfile: the artifact is
committed so every reader and every install sees it, and CI refuses a copy that the generator would
not reproduce.

**The generated file must carry the `DO NOT EDIT` banner naming its generator and its sources**, as
`vocabdoc` does.

## Registration becomes informational

Once the definition *is* the seat, `register` no longer has to bind identity or select a surface —
both are static. It reduces to recording the sitting and its conversation id.

**It cannot simply be deleted, and this is the part most likely to be got wrong.** `register` is
load-bearing for at least:

- `internal/sittingcap` — "the sitting IS the register": the per-sitting tool-call cap counts in the
  window a register opens, and every call after it counts against that sitting until the next
  register replaces the header. This is the one mechanism that survived the old top-level-process
  harness, precisely because it keys on the register event rather than on a subagent span.
- `record/sitting.go`, `record/dispatch.go`, `record/impasse.go`, `record/manifestowed.go`,
  `internal/capture` — all read registration.

So the change is **"register stops gating, keeps opening the sitting"**, not "register goes away".
Each of those five readers needs checking against the new meaning before the gate is relaxed.

## Carriers to sweep (complete-the-concept)

1. `plugins/frank-exchange-of-views/agents/*.md` — 11 hand-written → ~15 generated
2. `skills/research-protocol/scripts/debate.js` — every `agentType:` reference, and the clause that
   mandates the `manual` read (it must go in the same change, or the concept half-lands)
3. `tools/internal/record/agentrole.go` — the `agentTypeRoles` table
4. `tools/internal/seatprobe/naming.go`
5. `TestEveryDispatchedAgentTypeIsAttestable`, `TestTheRosterMatchesWhatTheEngineActuallyDispatches`
6. `scripts/check/gates.go` — the new `agentgen -check` gate
7. The prompt goldens — every seat prompt changes when the manual duty is removed

## Verification

1. `go -C scripts run ./agentgen -check` — fails on a stale committed constitution
2. Delete one generated section and re-run: the gate must fail (mutation check, not coverage)
3. `node --test plugins/frank-exchange-of-views/tests/simulator/debate.test.mjs`
4. `go -C scripts run ./check` — read the FAIL COUNT
5. **A universe run, judged on the record**: `manual` invocations must be **0**, `--help` calls should
   stay at or near today's 8, and the tool error rate must not rise above 6.4%. Those three numbers
   are the acceptance test, and they are all already measurable from the seat transcripts.

## What would falsify this

If `--help` calls rise materially once the manual is static, the constitution is not delivering what
the runtime read delivered, and the change has moved cost rather than removed it. The measurement in
step 5 is what says so.
