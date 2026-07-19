# Handoff — pick up here in a fresh context

Written 2026-07-19, revised at the end of the second long session. Everything below is
either committed or explicitly flagged as not. Verify before trusting: several claims in
these sessions were wrong on first telling and corrected only when checked — twice by the
human asking one more question, once by CI, once by a subagent refusing an instruction.

## 0. STATE RIGHT NOW

| | |
|---|---|
| branch | `fix/command-surface`, pushed |
| open PR | **#53** — command surface, vocabulary, timestamps, read surface, cross-seat tests |
| versions | frank-exchange-of-views **0.14.0**, prosthetic-conscience **0.10.1** |
| CI on #53 | goldens / feov-record / test / debate-sim **PASS**; **rule-sweep FAILS** (see §1) |
| plugin update dance | **STILL NOT DONE** — owed across two sessions now |
| the run | 2026-07-18_gray-area-telemetry — COMPLETE, CEILING, and now COMMITTED as evidence |

## 1. BLOCKING — the one thing standing between #53 and merge

**`rule-sweep` fails: four commits carry `Rule-Class: protocol-surface`, which is not a
registry slug.** I invented a catch-all instead of using the real classes — the exact
behaviour that registry exists to prevent. Each maps to a slug that already exists:

| commit | class |
|---|---|
| `0.12.0 — one word per intent` | `policy-without-mechanism` |
| `test(feov): cross-seat sequences` | `adjacent-seat-omission` |
| `0.13.0 — show` | `lossy-channel-substitution` |
| `0.14.0 — board JSON` | `format-selects-audit-surface` |

The fix is a message rewrite over `origin/main..HEAD` plus a force-push. **A permission
classifier blocked the rewrite in-session and it was NOT worked around.** The human must
allow it, or decide to register `protocol-surface` instead — the sweep's own error message
permits that, but it is a surface name rather than a failure class, and adding a catch-all
to a registry of failure modes defeats the sweep. Recommend the rewrite.

**Then the plugin update dance** (owed twice over): `/plugin update` + `/reload-plugins`,
then `/prosthetic-conscience:doctor`.

**While reloading, settle the plugin-bin question.** A probe is armed in both active plugin
`bin/` directories. After the reload run `command -v sc-path-probe`:

- resolves → a plugin's `bin/` IS on the Bash PATH at enable time, and shims for off-PATH
  tools belong in the plugin's own `bin/`, not the user's home. `plans/tool-resolution.md`
  adopts the user-bin approach and MUST then be amended.
- silent → the mechanism is unavailable and the user-bin plan stands.

## 2. QUEUE, in priority order

1. **Wire stdin payloads through the verbs.** `internal/flags` already has
   `ReadPayload`/`ReadComment` supporting `--file -` and `--comment-stdin`; the verbs do
   not call them yet. This is the escaping tax: the run's transcripts carried 68 commands
   with escaped quotes, 9 heredocs, and 37 staging a temp file first — twice failing
   because the staged file was not there.
2. **Free-text on the nine bare verbs.** `--comment` now reaches every verb, but
   `cite`, `dispose`, `dispute-respond`, `regrade`, `confidence`, `dispute` have no
   payload channel of their own. See `plans/command-surface.md`.
3. **Round-parity guard.** red-merge can still be spawned with no preceding blue turn,
   and the binary verdict forces FAIL, which telemetry then reports as failed repairs.
   The seat asked for either an engine precondition or an envelope field distinguishing
   an evidentiary FAIL from a no-response FAIL.
4. **Four verbs the seats asked for by hitting their absence**: `mint --amend` (a first
   mint with a placeholder payload is permanent), `amends_prior` usable by a seat that
   did not enter the original closure, `withdraw` for an archive block written in error,
   `closed_with_residue` (or `--residue` on `bench opinion`) for "check passes, residue
   of the same class survives at an unnamed site".
5. **Positional `--seat-id`.** 5 failures per run. It CANNOT be inferred (see §4). Its
   own PR: it breaks every seat prompt and 22 goldens at once.
6. **Assembly stops being an LLM.** 18 minutes of wall clock performing a union-copy the
   model has been measured CORRUPTING (run-5: 6/7 catechism answers regressed). A script
   does the copy; a small model writes only the TL;DR and synopsis. Correctness lever,
   not a cost one — it is ~$2 of a $54 run.
7. **Flip the ledger to the rendered projection.** NOW GOVERNED BY
   `plans/tool-is-the-contract.md`, which supersedes this line. The 34,086-vs-7,527
   archive metric is CONTAMINATED — see §3. Re-measure on the first post-timestamp run.
8. **Gray Area** (4th plugin, PR #3). Its acceptance test already exists: a seat logged a
   failure it had ALREADY RECOVERED FROM, and only the trajectory can show that. See §5.

## 2b. WHAT SHIPPED THIS SESSION (do not rebuild)

- **One vocabulary, enforced.** 179 literal→constant substitutions; 55 flag names resolve
  through `internal/flags`. Four collisions fixed: `--class`/`--petition-class`,
  `--label`/`--claim`, `--grade`/`--confidence`, `--rationale`→`--basis`. The
  `--reason` vs `--basis` line is now STATED: reason = why a thing happened, to a reader
  not contesting it; basis = the grounds you argue FROM where another seat may answer.
- **`vocabulary_test.go` walks the real command tree BOTH ways** — every registered flag
  must be declared, every declared flag must be registered. The second direction is what
  would have caught the original orphan state. Both proven to fail on purpose.
- **`show`** — reads through the tool, six views, byte-identical to the projection files.
- **`show --view board`** — the board as structured JSON, merge's default view. Closure
  anchors as FIELDS, `disposed` stated per observation, counts, and anomalies surfaced.
- **Nine cross-seat sequence tests** (`crossseat_test.go`), plus `close` gaining `--text`
  and `opinion --as halt` being refused — both defects found by writing the tests.
- **The run committed as evidence**, minus the gitignored transcript tarball.

## 3. WHAT THE RUN PROVED (do not re-derive)

- **The bench reform works.** `carried_share` 0.11 (1 of 9) against a 76/77 baseline. The
  judge stopped being "a router with robes".
- **W2i's assumption holds.** Citation yield PER DISPATCHED SEAT: 1.5 → 1.0 → 0.0. It
  collapses to nothing by round 3, so the consolidation cap is justified. L5 logic went
  9 → 3 → 9 — it RECOVERED, so the two-round reading that "L5 collapsed" was an artifact.
- **Cost is not where it looks.** Run cost $54 (not the $190 the dashboard claimed).
  red-merge is 47% and grows every round — $6.65 → $9.04 → $9.40 — tracking cache-read
  5.36M → 7.70M → 9.10M as the board accumulates. **Cost scales with board size, not
  round count.**
- **Wall clock is the real constraint**: 68% of 136 minutes was seven serial judgment
  seats. Fixed in #52 (`--smoke` now sets `judgmentModel`), unverified in practice.
- **THE PARITY METRIC IS CONTAMINATED — this correction matters most.** A parity analysis
  concluded the board misclassifies six gaps because `BoardState` honours only red's
  `close` and ignores bench closures. That is FALSE; a cross-seat test disproves it.
  Re-rendering the real events with today's binary still gives 9 open, for a different
  reason: **zero of the run's 271 events carry `ts`**. The run predates the timestamp
  schema, so ordering falls back to `(SeatID, Seq)`, `"judge-r2"` sorts before
  `"red-merge-r1"`, and every judge opinion replays before the mint it references and is
  dropped. Legacy-ordering damage from a defect ALREADY FIXED — not missing verbs. Do not
  build verbs against that byte gap. The anomaly machinery names all 12 cases; nobody had
  looked.
- **Scorecard defects still open**: `blue_sections_citing_direction` reports 3/2 (a
  benchmark over 100%); `anchored_closures_pct` reads 0 against an 89 baseline because
  the metric parses hand-written prose while anchors go to events;
  `lines_of_inquiry` reads "not computed" while capture logged 5 avenue events;
  `round_parity_failures` reads 0 for a run that HAD one.

## 4. ENVIRONMENT FACTS, MEASURED (do not re-litigate)

- `export` does **not** persist between Bash tool calls. The working directory is
  **reset** between calls. The shell is effectively stateless per invocation.
- A subagent's `CLAUDE_CODE_SESSION_ID` is the **parent's**, shared by all siblings.
  There is no per-agent identifier anywhere in the environment, so a state file keyed on
  session id would have every seat colliding on one key.
- Therefore `--seat-id` cannot come from env, cwd, or a hook. It is irreducibly
  per-seat and must come from the prompt or a positional argument.
- `SubagentStart` reportedly carries `agent_id`/`agent_type`, but `PreToolUse` does NOT
  receive `agent_id` — the chain breaks in the middle. UNVERIFIED; the same source was
  wrong about plugin `bin/` on PATH.
- gcc IS installed: `%LOCALAPPDATA%\Microsoft\WinGet\Packages\BrechtSanders.WinLibs.POSIX.UCRT_*\mingw64\bin`.
  Prepend it and set `CGO_ENABLED=1` for `-race`. A depth-5 search from `C:\` misses it.
- qlty (`~/.qlty/bin`) and jq (winget) are installed and off PATH.

## 5. GRAY AREA'S ACCEPTANCE TEST, ALREADY CAPTURED

blue-synthesize logged: *"PINNED.md names two evidence files that do not exist on
disk"*. The trajectory shows it ran `git show <pin>:inputs/x.md` from inside the run
directory, hit git's error naming the correct repo-root-relative path, **retried from the
root and succeeded**. It logged a failure it had already recovered from.

Both the claim and its refutation live in the same run, and only the trajectory can
adjudicate. **Friction is unverified self-report** — the exact defect the closure records
were rebuilt to remove, still sitting untouched one channel over. A second instance: red
reported `spot-check` "cannot record an honestly-empty round"; it always could.

Constraint for the miner: **~11% of tool calls are invisible to a naive matcher** because
seats alias the binary to a shell variable (`B=`, `R=`, `REC=`). Any analysis that greps
for a binary name misses them. Parse shell structure, not strings.

## 6. DISCIPLINE (unchanged, still binding)

- `gh auth switch --user gblock-agent` for writes, **ALWAYS restore `ctoforaday`**.
- Commits authored **Ethics Gradient**, with `Co-Authored-By` and `Claude-Session`.
- Pull before branching. Plan amendments land on main.
- Never change `model`/`judgmentModel` on a resume — it busts the agent cache.
- During a live run: no `git add -A`, no checkout, no stash over the untracked blackboard.
- Protocol-surface commits need `Rule-Class:` and `Sibling-Sweep:` trailers, and
  **continuation lines must be indented** or the whole trailer block is disqualified.
- New test files must be named in `.github/workflows/hooks.yml` — that job enumerates
  files because `node --test <dir>` fails on Windows, so an unlisted test runs NOWHERE.
- Goldens: regenerate deliberately, in their OWN commit, and read the diff line by line.

## 7. HOW I WENT WRONG THIS SESSION (repeat none of these)

- **Truncated output read as absence** — three times. `tail -12` hid a flag that sorted
  above the window; a depth-limited `find` missed gcc; a mid-token diff split looked like
  a payload key vanishing. Widen the window before concluding something is missing.
- **A test that could not fail for the right reason.** The bench-closure test first
  SKIPPED (excusing its own subject) and then PASSED on an assertion too weak to tell
  "closed" from "the word closure appears somewhere". Only the third form indicted the
  code. Assert on the specific state, and make the test fail once on purpose.
- **Truthiness on a legitimately-falsy value** — three instances in one day
  (`review_flag: false`, `!text` for an empty file, `startedMs: 0`). Use explicit
  `!= null` / `Number.isFinite` / presence checks.
- **Patching the instance and declaring the class fixed.** The two-pass replay ordered
  the two tiers I had a test for and left mutation-vs-mutation order filename-driven.
- **Quoting a number without auditing it.** The $190 cost came from a tile I had not
  checked; it was wrong by 3.5x and was steering a spend decision.
- **A test that reimplements its subject.** `TestValidateOpinionNamesEachMissingField`
  computed the expected flag spelling with the same underscore-to-hyphen transform the
  code used. It asserted the code agreed with itself, passed through a rename that broke
  the message, and read as coverage the whole time. Expectations must be LITERALS
  wherever the thing under test is itself a derivation.
- **Building a chokepoint and routing nothing through it.** `internal/flags/names.go`
  was written to make flag-name divergence require a deliberate act, but 20 of its 24
  constants were never referenced — and it had already drifted (`Supersedes` held
  `"supersedes"` while every call site registered `superseded-by`) without anything
  noticing. An unused single-source-of-truth is a comment, not a constraint.
- **A normalizer that was reproducible on one machine only.** Goldens and the determinism
  fuzz ranked DISTINCT CLOCK VALUES; when two events share a tick there is one fewer
  distinct instant and every later rank shifts. Green locally, red in CI, no semantic
  difference. It fixed machine-PATH dependence and left machine-SPEED dependence, and it
  made the determinism test itself non-deterministic — the property it exists to assert,
  broken in its own harness. Now ranked by POSITION in the canonical order.
- **Inventing a class instead of reading the registry.** Four commits shipped
  `Rule-Class: protocol-surface`, a surface name where a failure class was required. All
  four had exact matches already in the registry. This is what blocks #53.
- **Trusting a subagent's headline over my own test.** The parity report's top finding
  contradicted a passing cross-seat test; checking took one command and reversed the
  conclusion. When a report and a green test disagree, run the experiment.
- **Instructing a subagent to make a change I had not verified.** I told one to "fix" the
  `Supersedes` constant as a mismatch; it refused, correctly — `--supersedes` and
  `--superseded-by` are different relations on different objects, and the change would
  have broken a CSV invariant. The instruction was confident and wrong.
