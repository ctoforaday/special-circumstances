# Pre-dry-run batch — path-free, poison-free, self-report retired

Written 2026-07-19. Approved scope (operator, this session): the prompt-only
tool-is-the-contract §VIII/§IV.C batch **and** the Go self-report/telemetry
retirement, all before the next controlled dry run. This overrides §III's
"re-measure first" sequencing by explicit operator call — the largest self-report
surface is retired BEFORE the run, not after.

Base: stacks on `feat/feov-tool-contract-migration` (PR #54, the render-shadow
migration) until it merges, then rebases onto main.

## I. Summary & Goals

After this batch, in the next run: no seat reads or writes a storage path (it
pulls the board through `show --view`), no chair prompt carries a cross-run
scorecard seed, and the board telemetry is COMPUTED BY THE TOOL from the record —
not hand-written by the merge seat and not independently re-derived as a trusted
self-report. The orchestrator keeps its own sandboxed detector recompute (it is
the orchestrator's own control signal over envelopes it already holds, not a
seat self-report — the distinction that matters).

## II. Design decisions (the forks, resolved)

1. **Tool computes telemetry authoritatively; seat stops hand-writing.** The Go
   tool already exposes a round-aware structured board (`record.BoardJSON`: per-gap
   `Round`, `Severity/Likelihood/Impact/ComplexityCost`, `ClosedRound`, open/closed,
   counts). Add a `MASS` mapping (mirror debate.js v2 exactly) + a telemetry
   computation over the board per round, emitted by a new `merge telemetry
   --round N` verb that WRITES the authoritative line to
   `records/render-shadow/board-telemetry.jsonl` (where the dashboard already
   reads, post-#54). The merge prompt's ~1,100 B hand-computed schema spec is
   deleted; the seat just runs the verb.
2. **The script keeps its detector recompute (648-670).** It is sandboxed (no
   tool access in-loop) and computes from envelope gap arrays it holds, for the
   convergence-vs-verdict DETECTOR and the FAIL-with-low-mass guard. This is the
   orchestrator's own derivation, not a trusted seat self-report. Capture audits
   the truth from disk post-hoc (§6.2 attestation ceiling, unchanged).
3. **Dead self-report counts removed.** `ledger_closure_lines`, `archive_blocks`
   leave the `RED_ENVELOPE` schema + the disabled-gate corpse (628-643). Their only
   consumers were the two gates disabled 2026-07-19.
4. **Path retirement is prompt-only** (§VIII). `show --view board|ledger|archive|
   debate|changelog|citation-ledger|lines-of-inquiry` is byte-identical to the
   files (enforced by `show_test.go`). Swap blue (749), judge (772, 829), assembly
   (842) file-reads → tool reads. Preserve blue's read-batching (composite read or
   accept one extra `show` call).
5. **priors-are-poison, both halves.** Half-1 (prompt-only): stop injecting the
   cross-run `scorecardClause` seed into chair prompts. Half-2 (Go): add `show
   --view scorecard` computing the CALLING role's scorecard live from THIS run's
   record (blank until in-run history exists); chairs read it before the open
   docket. Cross-run `scorecards` arg feeds operator analytics only, never prompts.
6. **Friction ①②** folded into the same prompt pass: ① blue citation-hygiene
   (GitHub refs `owner/repo#N` + a locating anchor per quoted span); ② proactive
   report-write path (write to scratchpad neutral name then `cp`, skip the
   Write-block stall). ③ verify W1.6 claim-count reaches the seat prompt.

## III. Phases

### Phase A — prompt-only (self-contained, lands first)
- §VIII path retirement: blue/judge/assembly → `show --view`.
- §IV.C dead-text deletions (frictionClause dual-write, "MUST NOT write
  debate.md", duplicated lineage paragraph 596, speedClause path carve-out).
- Remove dead self-report counts from schema + disabled-gate corpse.
- priors half-1: remove scorecardClause seed injection.
- friction ①②③.
- Regen seat-prompt goldens; review diff line-by-line; simulator suite green.

### CORRECTION (2026-07-19, during Phase B scoping)

**`show --view scorecard` must NOT be a Go view.** The scorecard is computed by
`scorecards.mjs` — a rich JS module that reads telemetry, envelopes, archive.md,
debate.md and the candidate files, and whose OWN doctrine (its lines 362-373) is
that the module writing a format is the only place allowed to read it: "three
readers of one artifact, disagreeing about it, is the defect." A Go re-implementation
is exactly that defect. So the in-run scorecard read (priors half-2) is a **thin JS
CLI on scorecards.mjs** (`node scorecards.mjs --run <dir> --chair <role>`), which the
seat invokes via Bash — reusing the single implementation. B1 (telemetry-in-tool) is
unaffected: the tool becomes the ONE computer of the telemetry LINE; scorecards.mjs
still only READS it. Split Phase B into B1 (telemetry, Go) and B2 (scorecard CLI, JS).

### Phase B1 — Go telemetry (the self-report retirement)

**Field taxonomy (from the 2026-07-19 consumer read).** Two kinds of field:

- **Board-derived (tool computes authoritatively — THIS is the self-report being
  retired):** `open_count`, `mass` (Σ over open gaps of MASS[likelihood]×MASS[impact];
  realized→0), `max_severity` (the grade STRING of the highest-severity open gap —
  dashboard 430/443 renders it raw, not a number), `realized_open` (count of open
  gaps graded `realized`), `new_mint{count, by_severity}` (gaps with Round==N),
  `repair_regression{closures = gaps with ClosedRound==N, lineage_mints = Round==N
  gaps whose supersedes names a ClosedRound==N gap, ratio}`, `edge_deltas{down_mass,
  up_mass}` (Σ|mass(succ)−mass(anc)| over Round==N supersedes edges, split by sign),
  `mapping_version` = "v2", `round`.
- **Flow-derived (NOT board-derivable — passed IN, not a made-up metric):**
  `accepted_deltas` is the orchestrator's record of grade-disputes ACCEPTED that
  round (debate.js 677-692). The tool cannot see it in the record. `found_by_summary`
  and `excluded_mass_memo` are consumed by NO script — drop them.

**MASS (mirror debate.js v2 exactly):** trivial 0.5, low 1, low-medium 1.5, medium 2,
medium-high 2.5, high 3, certain 3.5, realized 0. mass(g)=MASS[L]×MASS[I].

**Design:** `merge telemetry --round N [--accepted-deltas <json>]` — the tool computes
every board-derived field from `BoardJSONOf(BoardState(runDir))` (per-round state via
Round/ClosedRound filtering), merges the passed-in flow fields, and appends the line to
`records/render-shadow/board-telemetry.jsonl` (where the dashboard already reads post-#54).
`scorecards.mjs` readTelemetry (line 44) must be pointed at render-shadow with a
trajectories/ fallback. `internal/record/telemetry.go`: pure `computeTelemetry(BoardJSON,
round)` (unit-testable with a hand-built board) + `Telemetry(runDir, round)`.

- Go tests: hand-built board → assert each field; a parity fixture vs the debate.js
  formulas. The pure function is the anti-drift gate.
- debate.js: replace the seat's `cat >>` telemetry instruction (line 606, ~1,100 B) with
  `feov-record merge telemetry --round N --accepted-deltas <the deltas the script/seat
  holds>`; delete the hand-computed schema spec; remove the TELEMETRY const. The
  script-side detector (645-670) STAYS (sandboxed, orchestrator's own control signal).
- Bump plugin.json + recordToolVersion; versionsync.

### B1 IMPLEMENTED (2026-07-19) — key finding: render already computed it

The tool ALREADY computed the per-round telemetry inline in `render.go` (the `render`
verb writes `records/render-shadow/board-telemetry.jsonl`, with all the JS-parity nuances
— insertion-ordered by_severity, round2, SliceStable max-severity). A standalone
`telemetry.go` module + `merge telemetry` verb would have been a SECOND computer of the
same fact — the exact defect this codebase forbids — so I deleted the module I'd started
and EXTENDED the canonical render computation instead:

- render.go telemetry now also emits `realized_open` (board-derivable). Difftest goldens
  regenerated (only that field added).
- debate.js: the merge seat's hand-written BOARD TELEMETRY line (~1,100 B) is replaced
  with a note that the tool computes it on `merge render` (which the seat already runs).
  The `TELEMETRY` const is gone.
- Consumers repointed to render-shadow with a trajectories/ fallback: `scorecards.mjs`
  readTelemetry and `cost-audit.mjs`.
- `accepted_deltas` is DEFERRED: it is dispute-flow state (which grade-disputes the
  orchestrator accepted that round), not board-derivable without threading dispute events
  into render. render omits it; the dashboard's deltas column reads 0 until it is added.
  `excluded_mass_memo`/`found_by_summary` are dropped (no consumer).
- Versions: cli.Version + recordToolVersion 0.2.0 → 0.3.0 (tool behavior changed);
  plugin 0.25.0 → 0.26.0.

### Phase B2 — scorecard in-run read (JS, not Go)
Thin CLI on scorecards.mjs: `node scorecards.mjs --run <dir> --chair <role>` prints the
chair's computed rows; the chair's prompt (recordClause / scorecardClause) instructs it to
run that before the open docket. Reuses the ONE scorecard implementation (§CORRECTION).

## IV. §V Verification loop (written, run every phase)

1. `cd plugins/frank-exchange-of-views/tools && go test -race -count=1 ./...`
2. `node --test plugins/frank-exchange-of-views/tests/simulator/*.test.mjs` (185+)
3. `node scripts/golden.mjs` (no --update in CI) — regen deliberately, review diff.
4. `node scripts/rule-sweep.mjs --base origin/main` — a protocol patch needs
   Rule-Class + Sibling-Sweep with REAL registry slugs.
5. **`TestSeatNeverNeedsAPath`** (§V.2, write first / let it fail): drive a full
   round per seat kind through ONLY `feov-record`; assert no path outside `--run`;
   assert debate.js prompt strings contain no `${runDir}/` outside `--run`.
6. The dry run itself: rerun the gray-area topic in haiku, compare to
   `research/2026-07-18_gray-area-telemetry/` (the Opus baseline) and the
   smoke-c haiku baseline. This run also produces the first clean timestamped
   parity measurement §III of tool-is-the-contract wanted.

## V. Deferred past this batch (still real, not in scope here)

- §IV.B enforcement gates: near-match as a tool return (`mint` requires
  `--supersedes`/`--not-a-reopen`), verdict-refuses-on-outstanding-duties,
  `found_by` foreign key (`mint --from-finding`), register-as-precondition,
  pattern-duty flag (`manifest-row --pattern-checked`), round-record derived.
- JS-renderer for SPA citations via claude-in-chrome MCP on red/blue/judge
  allowlists (red-verbatim-citations §follow-up).
- Bench sharpening: escalate leaf-node contradictions over rigor-asymmetry
  (from the smoke-c R1-4 vacuous-conditional miss).

## VI. Rule-sweep classes for this batch's commits

- Path retirement + telemetry-to-tool: `policy-without-mechanism` (a computation
  the prompt asked for, now enforced by the tool) + `format-selects-audit-surface`
  if the scorecard view changes what's audited.
- priors-are-poison: `staged-not-delivered` (the seed reached the prompt but was
  the wrong run's) — confirm against the registry at commit time, never invent.
