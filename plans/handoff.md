# Handoff — pick up here in a fresh context

Written 2026-07-19 at the end of a very long session. Everything below is either
committed or explicitly flagged as not. Verify before trusting: several claims in this
session were wrong on first telling and corrected only when checked.

## 0. STATE RIGHT NOW

| | |
|---|---|
| branch | `fix/command-surface`, pushed |
| open PR | **#53** — command-surface audit, flags library, timestamps, integration tests |
| merged today | #50, #51, #52 |
| versions | frank-exchange-of-views **0.11.0**, prosthetic-conscience **0.10.1** |
| plugin update dance | **NOT DONE** — deliberately held to the end and never reached |
| the run | 2026-07-18_gray-area-telemetry — COMPLETE, verdict CEILING, captured, marker lifted |

## 1. BLOCKING — do this before any run

~~**`debate.js` prompts may still teach flag spellings that no longer exist.**~~
**SWEPT 2026-07-19. The prompts were clean** — no agent doc, skill, or `debate.js`
prompt named a deleted spelling. The prediction was wrong.

**But the sweep found the same defect one layer down, in the error messages.**
`record.go` built the flag name in its `opinion requires --%s` errors by replacing the
first underscore of the payload key with a hyphen. True until the audit renamed
`--gap-id`→`--id` and `--disposition`→`--as`; the payload keys never moved, so the
message kept teaching two flags the parser rejects. **The error string is the seat's only
teacher, so a wrong one costs a turn.** Fixed by `flags.ForPayloadKey` — a stated map,
not a derivation.

Two tests were holding it in place: one asserted the literal stale string, and one
computed its expectation with *the same transform the code used*, so it asserted the code
agreed with itself and sailed through the rename. Both now assert literals.
**A test that reimplements its subject cannot indict it** — see §7.

**Then the plugin update dance** (owed, never run): `/plugin update` + `/reload-plugins`,
then `/prosthetic-conscience:doctor`.

**While reloading, settle the plugin-bin question.** A probe is armed in both active
plugin `bin/` directories. After the reload run:

    command -v sc-path-probe

- resolves → a plugin's `bin/` IS injected onto the Bash PATH at enable time, and the
  `.gitignore` fix (`plugins/*/bin/*` + a tracked `.gitkeep`, merged in #51) is what made
  it possible. Shims for off-PATH tools then belong in the plugin's own `bin/`, NOT in
  the user's home — scoped, no shadowing, uninstall-clean. `plans/tool-resolution.md`
  currently adopts the user-bin approach and MUST be amended.
- silent → the mechanism is unavailable in this CLI and the user-bin plan stands.

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
7. **Flip the ledger to the rendered projection.** Blocked until record-parity closes;
   the archive gap is the metric (34,086 bytes hand-written vs 7,527 rendered).
8. **Gray Area** (4th plugin, PR #3). Its acceptance test already exists: a seat logged a
   failure it had ALREADY RECOVERED FROM, and only the trajectory can show that. See §5.

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
