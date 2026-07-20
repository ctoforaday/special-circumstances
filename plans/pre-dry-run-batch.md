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

### Phase B — Go telemetry + scorecard
- `internal/record` (or a new `internal/telemetry`): `MASS` map + `Mass(gap)` +
  `Telemetry(runDir, round)` computing open_count, max_severity, new_mint,
  mass, realized_open, repair_regression{closures,lineage_mints,ratio},
  edge_deltas{down,up} over `BoardJSON` at rounds N and N-1.
- `merge telemetry --round N` verb: compute + append the line to
  render-shadow/board-telemetry.jsonl.
- `show --view scorecard`: the calling role's live in-run scorecard.
- Go tests: a golden record → assert the tool's telemetry equals the debate.js
  computation for the same gaps (parity test, the anti-drift gate).
- debate.js: replace the seat's `cat >>` telemetry instruction with the verb;
  delete the hand-computed schema spec; wire chairs to `show --view scorecard`.
- Bump plugin.json + recordToolVersion; versionsync.

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
