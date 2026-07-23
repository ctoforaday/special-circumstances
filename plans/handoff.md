# FEOV handoff / memento — checkpoint 2026-07-23

Resuming **frank-exchange-of-views** (the research-debate engine). This is the pickup point.
Read it, then `gh issue list` for the live queue. **Verify before trusting** — many claims this
run were wrong on first telling and caught only by checking (a subagent's "all green" over a live
compile error; a green fuzz that tested nothing; verify/graph each flagging their own first bug).

## State right now
- **`main`** = `dd6b9c4` (PR #98 merged). Plugin **0.38.0**, recordToolVersion/cli.Version **0.12.0**
  (they move together — versionsync_test enforces it; the plugin version is decoupled).
- Env READY (qlty/jq on PATH; gcc still off PATH — [[tools-installed-but-off-path]]). Released tag
  is old (`v0.33.0`); a fresh tag is owed if consumers need the 0.12.0 binary.
- **To drive a run off local build (no plugin cache):** build feov-record from `tools/` to a
  binDir, run the working-tree `setup-research-run.mjs --bin-dir <winpath>`, then Workflow with
  `scriptPath`=working-tree `debate.js`. Pass WINDOWS paths (C:/...) to node scripts, not /c/...
  (MSYS paths make node/spawn ENOENT). Recipe + smoke config in [[feov-projection-retirement-queued]].

## What SHIPPED this run (all merged)
- **verify** (`feov-record verify --run <dir>`) — read-only invariant cross-check (gaps disposed,
  refs resolve, PASS closed everything, register-before-append) + authoritative tally. #92.
- **graph** (`feov-record graph --run <dir> [--format mermaid|dot]`) — renders a run's ACTUAL
  behaviour from the record (seat flow + gap lifecycle, holes flagged). #96.
- **Stage 1 of record-only-channel** (#93): seats EMIT position/closing/opinion events and READ
  via `show --view debate` instead of hand-writing debate.md (binDir mode; no-tool falls back to
  the file). render.go already composes the transcript from those events.
- **A1-A3 render fix + friction in report** (#96): assemble.go was reading dispute/dispute-respond/
  petition-rule prose under the WRONG payload key (wrote evidence/response/rationale/opinion, read
  basis/as/rationale) — rendered EMPTY. Fixed. friction events now render (were write-only).
- **gaps_outstanding truth** (#82): the completion envelope read the board's open count, not red's
  docket. **Fuzz** (#97) + **expansion + prose-renders oracle + speedup** (#98): goja runs the real
  debate.js against the real binary; agents make coherent random tool calls; oracles = verify +
  "every dialectic event's prose renders". 1000 runs, 0 fail. Default N=60 (~15s CI); FUZZ_N=1000
  for the sweep; FUZZ_C overrides concurrency (default NumCPU*3).

## The record-only-channel push (#62 umbrella) — plans/record-only-channel.md
The clean one-way answer: events are the ONLY inter-agent CONTENT channel; debate.md becomes a
read-only rendered view; envelopes carry orchestration refs, not a second copy. Staged:
- **verify** ✓ (the harness) · **Stage 1 closings/positions** ✓ (#93)
- **Stage 2** — disputes onto the record + docket derives from refs (envelope grade_disputes today)
- **Stage 2.5** — the ORCHESTRATOR on the record: debate.js's mechanics decisions (docket, round,
  deadlock, verdict) are ephemeral; record them via a lead-mechanics channel emitted by an agent
  proxy (the script is sandboxed, can't call the tool). The user's own insight.
- **Stage 3** — retire debate.md as a write AND read target; update the parity-audit (capture).

## Queue = GitHub issues. NEVER close a bug until a RUN confirms ([[bug-state-tracking]]).
- **Fixed this run, ripe to close after a real run confirms:** #83 (gaps_outstanding), #94 (A1-A3 —
  now guarded by the fuzz prose oracle), #66 (empty debate — Stage 1), and the report-structure set
  #65/#74/#75/#76/#79 (merged in #82). Re-verify with a `/research` run + `verify`, then close.
- **Schema hygiene (from the 2026-07-23 audit):** #95 — prose-key fan-out (--reason → 8 payload
  keys; the A1-A3/#83 root; collapse single-prose verbs to `reason`), --as → 5 keys, key
  collisions (verdict/class/label/disposition), write-only events (petition/retire/confidence/
  manifest-row/register.tool_version — DECIDE per item: wire a reader or drop). friction now wired.
- **Tool-contract basics (mostly deferred/corrected):** #84 (identity — NOT a bug in normal runs;
  --run already inferred, --seat-id can't be env-injected), #85 (avenue prompt fixed; NO --gap-id
  alias — one canonical flag, [[one-way-no-aliases]]), #86 (heredoc → a HOOK, not prose; the
  registry names heredoc-prose an antipattern), #87 (friction verb bypassed for the envelope),
  #88 (manifest word cap — real cap is in blue's constitution, not debate.js).
- **Bigger:** #62 (umbrella), #63 (model tiers), #64 (red constitution), #68 (schema versioning),
  #70 (compute claim_count in tool — ties #83), #73 (agent .md stale paths — worsens with Stage 3),
  #80 (test-coverage audit).

## Pending DECISION (not a build)
`confidence` surfacing: recommend NON-authoritative — surface as blue's SELF-assessment, visible to
red as a targeting signal + the judge as context, but NEVER feeding the red-audited risk matrix
(else blue grades its own exam). User asked "shouldn't it be in a risks table?" — pushed back, no
confirmation yet.

## Fuzz — next expansion (deferred paths, where edge cases still hide)
supersedes-lineage via the envelope, petitions, deadlock, the envelope-driven grade-dispute DOCKET
machinery (debate.js's gnarliest logic). Also: a per-run graph is a good debugger for a failing
fuzz seed. Add the "prose renders" style oracle to more event types as coverage grows.

## Earned pitfalls (this run)
- **Run go vet + gofmt + go test locally, not just go test** — the feov-record CI job runs all
  three; gofmt-unformatted passes `go test` but fails CI ([[ci-checks-and-local-loop]]). WATCH every
  PR's checks after pushing.
- **NEVER add flag aliases/synonyms** — one canonical way ([[one-way-no-aliases]]); the --gap-id
  alias got merged then torn out.
- **Don't force-push a PR without checking its merge state** — #89 was already merged when I
  force-pushed; had to redo as #91.
- **Instrument coverage in any fuzz/generator** — a green run that exercised nothing (the seat-id
  trailing-`.` bug) is worse than a red one. verify/graph/fuzz each caught their OWN first bug at
  the leaf; keep that discipline.
- **The report/debate render reads per-verb payload keys** — the `--reason` fan-out means the reader
  must know all 8; getting one wrong renders prose empty invisibly (A1-A3, #83 class).
- Bump cli.Version BEFORE regenerating stamp goldens, `-count=1` (the #57 lesson).
