# FEOV_RUN injection — the seat stops typing the run directory

> STATUS 2026-09-02: shipped 2026-08-06 — historical record; twice superseded in part (banners below). `hookgate.PreOutcome` and the `setup`-written wrapper (`FEOV_RUN_FROM_WRAPPER`, `setup/wrapper.go`) are the live contract.

> **SUPERSEDED IN PART, 2026-08-22 (#510).** The position matcher this plan designs was removed:
> injection is now unconditional on every Bash command in a live run, and it carries `FEOV_AGENT_ID`
> as well as `FEOV_RUN`. The mention-vs-invocation reasoning below describes a harm the mechanism
> cannot produce — injection prepends — and the matcher's silent miss cost 21 of 65 registers their
> identity across six runs. `hookgate.PreOutcome` is the current contract.

> **SUPERSEDED AGAIN, 2026-08-23 (#526).** Injection is no longer the primary carrier of the run
> directory. `setup` now writes a per-run wrapper — `<runDir>/.bin/feov-record`, execing the real
> binary with `FEOV_RUN_FROM_WRAPPER` set — and prints its directory as the `binDir` the workflow
> hands every seat. The dispatcher bakes the run in **before the run starts**, where injection
> re-derives it per call from a marker that can move; the wrapper cannot go stale and the marker
> can. Injection remains, and is now the SECOND carrier with a live gate between them: hook and
> wrapper disagreeing is refused as *"the engine's live-run marker has moved since this run
> started"*, which is exactly the run-3 `assemble` incident (#500) that nothing could see while
> there was one carrier and it was the one that had moved.
>
> Measured on a live smoke run: 249 commands through the wrapper, **0** containing `--run`, 13/13
> seats identified. On a healthy run the wrapper is invisible — the hook sets `FEOV_RUN`, which
> outranks it.

> **Rev 6.** Rev 1/2 failed on scope; rev 3 failed on nine specifics. Fixed here: the goal is
> restated to what this scope can deliver, the version surface is decided, the census is
> re-run against the **real** API (`PreDecision`, not the `Decide` I named without checking),
> deny-vs-rewrite composition is specified, the marker resolution is named, the match rule is
> bounded to invocation position, and the test-env hygiene is enumerated.

## I. Summary & Goals

### The measured failure

- **First live run: 55 tool-call errors in 534 executions, TEN a dropped `--run`**
  (`cli/seat/seat.go:76`). Inference was added then and never fires — the prompt hands the
  seat an absolute path and *an explicit `--run` always wins*.
- **2026-08-05 smoke**: `blue-respond-r1` typed `special circumstances` (space, not hyphen)
  and got *"names gap R1-2, which no mint event created"*. It filed friction blaming a
  dangling-reference rule and abandoned the manifest rows for R1-3…R1-7. **One typo → five
  lost receipts and a false bug report.**

### Goals — what THIS scope delivers

1. **The injected run directory is authoritative, and no wrong one ever wins silently.**
   Measured as: given a live marker, the rewrite emits `export FEOV_RUN='<marker runDir>';`
   ahead of an invocation **at one of the positions enumerated in §2**, and a `--run` that
   disagrees is refused **naming both values** instead of being obeyed.
   **Coverage is bounded, and a miss is observed BY THE SEAT, not by the hook.** Shapes
   outside the enumerated set — `$(feov-record …)`, `if feov-record …; then`,
   `VAR=x feov-record …`, `time feov-record …`, `bash -lc "feov-record …"` — are not
   rewritten. Rev 5 proposed the hook logging friction on a miss; that was wrong twice over
   (see §3), and the fix is to observe it where identity exists: **when a seat verb resolves
   its run directory from `--run` or inference while a marker IS present, the tool logs
   friction as that seat.** The hook cannot classify a mention from a miss — they are the same
   observable to a matcher — but the tool never sees a mention at all, only real invocations.
2. **Nothing new appears in `--help`.**
3. **An operator running `feov-record` by hand is unaffected**, including against an archived
   run while another is live.

**Explicitly NOT a goal of this plan: "zero run-path tool errors in a live run."** Prompts
still emit `--run ${runDir}` at every seat call, so a typo lands on Goal 1's refusal — which
IS a tool error. That target belongs to the follow-on that stops `debate.js` emitting the
flag; **tracked as its own issue, not as a parenthesis here.** Rev 3 asserted the zero-error
target in §V Manual while conceding in §V-1 that it could not hold — that contradiction is
removed.

## II. Technical Context

- Go, `plugins/frank-exchange-of-views/tools`, cobra.
- **The real hook API** (census re-run — `grep -n "^func " internal/hookgate/hookgate.go`):
  `PreDenyReason() string`, `PreDecision(Input) (bool, string)`, `PostDropped(...)`,
  `DefaultAnchorIDs(...)`. **There is no `Decide`** — rev 3 named a symbol that does not
  exist, which means that census was not run.
  `Input` (hookgate.go:32-36) carries `agent_type`, `tool_name`, `tool_input`.
- **`cli/hook.go:61-76`** parses the payload, calls `PreDecision`, and on deny writes **one**
  `hookSpecificOutput` document via `emitPreDeny`.
- **Client capability**: `hookSpecificOutput.updatedInput` replaces a tool's arguments;
  **no hook output can set an environment variable**, so env travels inside the command string.
- `run-live.json`: written by `setup/setup.go:438`; read by `prosthetic/internal/runlive`,
  `pushfreezeguard`, `toolchainnudge`, `capture.go:1105`, `dashboard_serve.go:100`,
  `cli/seat/seat.go` (`InferRunDir`). **Schema unchanged by this plan.**

## III. Proposed Changes (the spec)

```
plugins/frank-exchange-of-views/
├── commands/research.md                 [MODIFY] launch passes --bin-dir (makes the preflight FATAL)
├── hooks/hooks.json                     [MODIFY] _comment — PreToolUse is no longer only the lockdown
├── .claude-plugin/plugin.json           [MODIFY] 0.72.1 -> 0.73.0
└── tools/
    ├── internal/seatenv/                [NEW] seatenv.go + seatenv_test.go
    ├── internal/hookgate/hookgate.go    [MODIFY] new PreRewrite(cmd, runDir) — Input UNCHANGED
    ├── internal/hookgate/hookgate_test.go [MODIFY] mkInput helper at :14 (not :37 — rev 4 cited a line it had not opened)
    ├── internal/cli/hook.go             [MODIFY] resolve the marker; emit rewrite; deny wins
    ├── internal/cli/seat/seat.go        [MODIFY] FEOV_RUN above the flag
    ├── internal/cli/root.go             [MODIFY] cli.Version 0.34.0 -> 0.35.0 + changelog
    ├── requirements.json                [MODIFY] recordToolVersion -> 0.35.0 (versionsync_test)
    ├── internal/fuzz/prerewrite_fuzz_test.go [NEW] Go fuzz over PreOutcome(command, runDir)
    ├── testdata/pretooluse-bash-feov.json [NEW] the real-data fixture (§V)
    └── feov-record                      [REBUILD] tracked 6 MB binary, Jul 29 — see below
```

**The tracked binary is a carrier.** `git ls-files` shows
`plugins/frank-exchange-of-views/tools/feov-record` — a committed 6 MB build dated Jul 29,
carrying the OLD contract. Moving `recordToolVersion` arms `setup`'s preflight against it, so
**`setup` refuses to start any run until it is rebuilt**. Rebuilding it is a step in §III, and
the post-merge `/plugin update` + `doctor --fix` refresh CLAUDE.md requires is sequenced in §V
Manual. Rev 4 graded only the risk the bump mitigates and not the one it creates.

### The version decision (rev 3 left it open; it is forced)

**`cli.Version` and `recordToolVersion` BOTH move to 0.35.0** — and the gate that is supposed
to enforce it **does not currently fire on the documented launch path**, which this plan must
fix or stop claiming.

`internal/setup/run.go:91` binds the record-binary preflight to INTENT: `--bin-dir` is what
`cli/setup.go:70` documents as making it FATAL. `grep -rn "bin-dir"` across the plugin's
markdown, JavaScript and JSON returns **nothing** — `commands/research.md:20` does not pass it,
so today a version-skewed binary produces a WARNING, not a refusal. (I passed `--bin-dir`
by hand launching the 2026-08-05 smoke, which is why it appeared to work.)

**So `commands/research.md` gains `--bin-dir` at the launch step** `[MODIFY]`. Without it the
version bump buys neither the protection R8 claims nor the blast radius R9 grades.
`research.md:31` already asserts this guarantee incorrectly — a pre-existing defect this plan
was leaning on, fixed here rather than inherited.

**Consequence, stated rather than discovered:** the version literal is stamped into every
register event's `tool_version` and appears in **18** files under `internal/difftest/testdata/`.
Those goldens **will** move, so §V regenerates them and verifies after — rev 3's "run `golden`
without `-update` and treat any change as a finding" is wrong once the version moves. The
plugin version alone appears in no golden (verified).

### 1. `internal/seatenv` `[NEW]`

`FEOV_RUN` only. For **seat** verbs (`cli/seat.Of`): `FEOV_RUN` → `--run` → `InferRunDir` →
error. A `--run` disagreeing with `FEOV_RUN` is **refused, naming both**.

**Scoped to seat verbs.** Operator commands (`verify`, `capture`, `dashboard`, `graph`,
`count-claims`, `scorecard`, `setup`) do not route through `seat.Of` and are never refused.

### 3. Miss observability — in the tool, not the hook `[MODIFY seatenv]`

Rev 5 used "log friction" three times as a hook-side mitigation. Both reasons it fails:

- **The hook has no identity.** Friction is a seat verb: `cli/seat/verbs.go:42` calls
  `record.Append(s.RunDir, s.SeatID, "friction", …)`, and `record.Append` (record.go:278) keys
  the shard and pointer on `seatID` and derives the round from it. The hook process has no
  seat id, and an invented one lands "under an identity no dispatch created", which
  `roles.go:70` refuses by name. (The record library itself works fine from a hook — the
  blocker is identity, not file access.)
- **A mention and a miss are the same observable.** `grep -rn "feov-record"` and
  `$(feov-record …)` are indistinguishable to a position matcher, so hook-side friction would
  fire on exactly the documentation writes and heredocs §2 exists to protect.

**So the seat observes it.** In `seatenv.Resolve`, when a seat verb resolves its run directory
from `--run` or inference **while `run-live.json` exists and `FEOV_RUN` does not**, append one
friction event as that seat: *"this invocation was not rewritten by the PreToolUse hook — the
command shape is outside the injected set (#281)"*. Identity and run directory are both in
hand, and it fires only on real invocations.

Bounded: one friction per seat per run (keyed like any idempotent append), so an uncovered
shape used forty times is one signal, not forty.

### 2. The rewrite `[MODIFY hookgate + cli/hook.go]`

**Marker resolution — the core stays pure, and `Input` stays wire-only.**

`hookgate.Input` is unmarshalled directly from stdin (`cli/hook.go:44-51`), so **every field on
it is wire-supplied**. Rev 4 proposed adding a CLI-computed `RunDir` to it — a reader could not
then tell a derived value from a payload key, and `PostDropped` (hookgate.go:120-129) already
derives its *own* run directory from the report path. Two unreconciled sources in one struct.

Instead: **one entry point, `PreOutcome(in Input, runDir string) (Outcome, string)`**, where
`Outcome` is `OutcomeDeny | OutcomeRewrite | OutcomeNone`. The directory is a parameter, exactly
as `PostDropped` takes injected `anchorIDs`/`readReport` "so the logic stays pure and testable";
`Input` is unchanged and `PostDropped` keeps its own derivation.

**Deny-first is structural, not a convention.** Rev 5 left `PreRewrite` an exported free
function with the ordering living in prose plus one call site — and R1's consequence is *the
blue-report lockdown opens*. `PreOutcome` consults the deny arm first and **cannot return
`OutcomeRewrite` when it fires**; there is no exported path that emits a rewrite without having
asked. `PreDecision` stays exported for its existing callers and tests.

`cli/hook.go` reads the payload's **`cwd`** — added as a `Cwd` field on the parse struct, which
is legitimate because `cwd` IS wire-supplied (it is a documented common input field) — and
resolves via the existing `seat.InferRunDir` walk. **Not** the hook process's `os.Getwd()`,
which is the hook's cwd, not the seat's. Absent marker, or a marker naming a directory that
does not exist → empty → **no rewrite**, matching `InferRunDir`'s "say nothing rather than
guess" (seat.go:99-113).

**Match on invocation position, not substring.** The command must contain `feov-record` as a
**command token** — at the start, or immediately after `&&`, `||`, `;`, `|`, or a newline,
optionally quoted and/or path-prefixed. A mention (`grep -rn "feov-record" …`, a heredoc, a
`--reason` quoting a failed command) is **not** rewritten. Rev 3's "contains" rule would have
mutated documentation writes and friction messages.

**Emission:** `export FEOV_RUN='<path>'; <original command>`.

- **`export …;`, never `VAR=x cmd`.** Measured: blue's real command was
  `cd C:/… && "…/feov-record" blue manifest-row …`; an inline prefix binds to `cd` and never
  crosses the `&&`.
- **Quoting is a security boundary**: single-quote, escape `'` as `'\''`, and **refuse**
  — return the command unmodified — for any value containing a control character. No
  hook-side friction (see §3): the seat logs the miss when it resolves without `FEOV_RUN`.
- **Idempotent**: a command already containing `export FEOV_RUN=` is returned unmodified.

**The unparseable-payload branch emits no rewrite.** `cli/hook.go:63-71` returns *before*
`PreDecision` when the payload will not parse, emitting a deny only if the raw bytes mention
`blue/report.md`. `PreRewrite` is **not** consulted there — there is no parsed `tool_input` to
rewrite, and inventing one from raw bytes is the guessing this design refuses. §V-10 covers it.

**DENY WINS — one document, never two.** `cli/hook.go` currently emits one
`hookSpecificOutput`; two would be invalid JSON, and emitting a rewrite in place of a deny
would **silently open the blue-report lockdown**. Order: `PreDecision` first; if it denies,
emit the deny and return. Only when it does not deny is `PreRewrite` consulted. A command
tripping both arms (`cp draft.md …/blue/report.md && "…/feov-record" blue edit …`) is denied.

## IV. Risk & Mitigation

| # | risk | L×I×Cx | mitigation |
|---|---|---|---|
| R1 | **Deny silently downgraded to a rewrite** — the lockdown opens | low × **high** × low | §2 deny-wins ordering; §V-7 drives a both-arms command |
| R2 | **Shell injection** via a crafted `runDir` | low × **high** × low | §2 quoting; hostile-value tests |
| R3 | **Compound-command miss** | **high** × high × low | `export …;`; §V-2 **must fail** against an inline prefix |
| R4 | **Rewriting a mention, not an invocation** | med × med × low | §2 invocation-position rule; §V-6 negative test |
| R5 | **`FEOV_RUN` leaking into the test suite** changes asserted output (`missing_required_flags.golden:4` asserts `--run <runDir> is required`) | **high** × med × low | seat-resolution and difftest harnesses `t.Setenv("FEOV_RUN", "")`; listed in §III |
| R6 | Operator regression | low × med × low | §1 scoping; §V-4 |
| R7 | Hook absent / marker missing / marker names a dead directory | med × med × low | No rewrite emitted; falls through to flag then inference — today's behaviour |
| R8 | Stale binary lacking the rewrite | med × med × low | Version bump + `setup` preflight + changelog entry |
| R9 | **The bump blocks runs until binaries are refreshed** — including the tracked Jul-29 `tools/feov-record` — but ONLY once `--bin-dir` reaches the launch path | med × med × low | Rebuild is a §III step; §V Manual sequences `/plugin update` + `doctor --fix`. **Regraded from high**: today the preflight WARNS rather than refuses, which is the defect §III fixes |
| R10 | An invocation shape outside §2's enumerated positions is missed | med × med × low | Goal 1 bounded to those positions; §V-11 drives the real `debate.js` shapes; **§3 observes the miss at the SEAT** — the hook cannot separate a mention from a miss; the tool never sees a mention |
| R11 | `PreOutcome`'s shell-shape parsing or quoting mishandles a hostile path | med × **high** × low | Go fuzz over `PreOutcome(command, runDir)` — a command parser plus an escaping path is the highest-value fuzz surface in this plan, and the repo's standing rule is that fuzzers track code |

## V. Verification Plan (the checklist)

### Automated — exact commands, one per line

```
cd plugins/frank-exchange-of-views/tools
gofmt -l .
go vet ./...
go test ./internal/seatenv/ -v
go test ./internal/hookgate/ -v
go test ./internal/cli/ -run "Seat|Run|Hook" -v
cd ../../scripts
go run ./golden -update          # MUST precede `go test ./...` — see below
go run ./golden
cd ../plugins/frank-exchange-of-views/tools
go test ./...
go test ./internal/fuzz -run TestFuzzDebate -count=1 -timeout 900s
go test ./internal/fuzz -run FuzzPreOutcome -fuzz FuzzPreOutcome -fuzztime 60s
cd ../ && node --test tests/simulator/*.mjs
cd ../../scripts
go run ./frontmatter
go run ./validatejson
go run ./mjsparity
go run ./pluginparity
go run ./versionguard
go run ./rulesweep
```

**Ordering is load-bearing, and rev 4 had it backwards.** `internal/difftest`'s comparison is
gated on `UPDATE_GOLDENS=1` (`golden_test.go:37`), and the ONLY thing that sets it is
`scripts/golden` (`main.go:78`). Since this plan states the 18 goldens **will** move,
`go test ./...` fails unless regeneration runs first — one-command-per-line only meant the
operator watched it fail instead of short-circuiting. So `golden -update` runs **before**
`go test ./...`.

The reviewable assertion is that `git diff` over `testdata/` contains **only** version-string
changes. Any other moved golden means this plan changed a contract it does not claim to — that
check is the human's, made possible by regenerating in isolation before the suite runs.

**Re-armed by**: `internal/seatenv`, `internal/hookgate`, `internal/cli/hook.go`,
`internal/cli/seat/seat.go`, `internal/cli/root.go`, `requirements.json`.

### Named regressions, with stated success

1. **The measured typo.** `PreRewrite` on the exact smoke command containing
   `special circumstances`, with a live marker. **Success: emitted command starts
   `export FEOV_RUN='<marker runDir>';`**, and `seat.Of` then refuses the disagreeing `--run`
   naming both values. (A refusal is a tool error under the *follow-on's* metric, not this
   plan's — see §I.)
2. **Compound command.** `cd X && "…/feov-record" blue edit …` — `FEOV_RUN` reaches the
   invocation after `&&`. **MUST FAIL against an inline `VAR=x` prefix.**
3. **Hostile quoting.** `runDir` with `'` → escaped, command still valid. With a newline or
   `$(…)` → refused, unmodified, friction logged.
4. **Operator untouched.** `verify --run <archived>` with a different run live → served.
5. **Idempotent.** Already-prefixed command returned unchanged.
6. **Mention, not invocation.** `grep -rn "feov-record" .` and a heredoc containing the
   string → **not** rewritten.
7. **Deny wins.** `cp draft.md …/blue/report.md && "…/feov-record" blue edit …` → one
   document, `permissionDecision: deny`, **no `updatedInput`**.
8. **Dead marker.** Marker present but its `runDir` absent → no rewrite (this is the state of
   the real marker on this machine today; see below).
9. **Degradation.** No marker → command unmodified; suite green.
10. **Unparseable payload.** Malformed stdin → the `!ok` branch emits at most a conservative
    deny and **never** a rewrite (`cli/hook.go:63-71`).
11. **The shapes `debate.js` actually emits.** Drive the concrete command strings from
    `recordClause` and `ledgerClause` — including the `cd … && "<binDir>/feov-record" …` form
    every seat is handed — and assert each IS rewritten. §V-6 proves mentions are skipped;
    this proves real invocations are not.
12. **A miss is observable AT THE SEAT.** A seat verb resolving its run dir from `--run` or
    inference, while a marker exists and `FEOV_RUN` is unset → one friction event under that
    seat's own id. Driven through the verb, not the hook. Bounded to one per seat per run.
13. **The preflight actually refuses.** With `--bin-dir` on the launch (§III), `setup` against
    a binary reporting 0.34.0 while `recordToolVersion` is 0.35.0 → **refusal**, not a
    warning. This is the assertion R8 rests on and which does not hold today.

### On real data, now

`plugins/frank-exchange-of-views/.claude/run-live.json` exists but its `runDir` points at a
scratchpad directory that **no longer exists** — so it is the §V-8 dead-marker fixture, not a
live one. The fixture `tools/testdata/pretooluse-bash-feov.json` `[NEW]` carries a real
PreToolUse payload; the test writes a marker whose `runDir` **is** a directory it created, and
asserts the emitted `updatedInput.command` verbatim for both cases.

```
cd plugins/frank-exchange-of-views/tools
go run ./cmd/feov-record hook pretooluse < testdata/pretooluse-bash-feov.json
```

### Manual

- **Rebuild and refresh first**: rebuild `tools/feov-record`, then `/plugin update` +
  `/prosthetic-conscience:doctor --fix`, then confirm `feov-record --version` reports 0.35.0
  and matches `recordToolVersion`. `setup` refuses until this is done (R9).
- `feov-record blue --help`: `--run` still described; no new flag.
- One live run: **no seat command resolves the wrong run directory** (Goal 1). The
  zero-tool-error target belongs to the follow-on.

### The auditor gate

`/plan-audit` on this file. Not approved until PASS.

---

## Appendix A — the identity concept (NOT in this plan)

**The principle** (gb): *all of the "frank exchange" in Frank Exchange of Views is
tool-mediated.* Failure classes: **markdown standing in for records**, and **string
concatenation / regexes as structured-data paths**.

**Class 2 instances**: `roundRe` (`record.go:66`) recovers round from the seat name; `roleRe`
(`findinglabel.go:12`) recovers the lens index; `Sprintf("R%d-%d")` (`replay.go:496`) encodes
round *into* the gap id; `view/changes.go:91` counts estoppel rejections by matching the prose
substring `"estoppel —"` — silently zero if reworded, and zero reads as "red behaved".

**Class 1 instances**: `blue/frontier.md` (no event backs it), `blue/CHANGELOG.md` (#251), and
**57 markdown files** of red's gap-pattern memory, 59 entries UNCLASSIFIED and never delivered.

**Solve first**: `NextFindingLabel` (`findinglabel.go:30`) is called at `cli/lens/finding.go:51`
**outside any lock** — scan, count prefix, return `n+1`. The per-seat prefix (`L1`,`L2`,`L3`)
is therefore **the uniqueness namespace that makes lock-free allocation safe** under parallel
dispatch (`debate.js:668`), not a comparability label.

**Resolution** (gb): **`agent_id` is a collision-free identifier — if a UUID is needed, it is
that.** A per-`agent_id` counter namespace is unique by construction; identity lives in the
record as a **field**, not concatenated into a name.

**Also unresolved there**: `debate.js:683` dispatches red-merge under the same `red-auditor`
type as the lenses, so a per-lens split must name a distinct merge agent; red's agent-memory
dir is keyed on that type (`setup/run.go:114`, `commands/research.md:25`); 20
`difftest/testdata` goldens embed raw event JSON; `Event` is lowerCamel (`seatId`, `ts`).

---

## Status: §1 and §2 SHIPPED 2026-08-06; §3 CUT

Implemented as the stable core only, per gb: **the injection, `export …;` over an inline
prefix, the quoting boundary, and refusal-on-disagreement** — the part unchallenged since
rev 3. **§3 (miss-observability) is CUT**, not deferred: it turned `seat.Of` into a record
writer on every read-only verb, it was the round-5 fix that introduced a round-6 defect, and
the idea was mine rather than the measurement's. R10 ships **unmitigated** and is stated as
such rather than quietly regraded.

How the six known-open items below resolved:

1. `PreRewrite`/`PreOutcome` — settled as **`PreOutcome`** at every site; there is no second name.
2. `seat.Of` as a record writer — **gone with §3.**
3. The `Append` idempotency claim — **gone with §3.**
4. R9 graded against the pre-change state — **overtaken:** #282 landed the unconditional
   preflight, so R9's premise ("today the preflight WARNS") is no longer true and the risk is
   the post-change one.
5. `--bin-dir` vs the legacy resume — **resolved by #282**, which split the two concerns:
   `--bin-dir` says WHERE, the Workflow's `binDir` says WHETHER.
6. The fuzz target — **built with stated invariants**, not crash-only: the value round-trips
   through an independent single-quote DECODER, the seat's command survives verbatim, the
   rewrite is idempotent, and a refused value is never emitted. 3.7M executions, zero
   failures.

Found during implementation, and not in any revision of this plan:

- **A heredoc is a blind spot.** A newline is a command separator, but inside a heredoc body
  it introduces DATA. The matcher rewrote a heredoc documenting a verb — which would have
  injected an export into a document a seat was writing. A command containing `<<` is now not
  rewritten at all. Caught by a test, not by review.
- **The DENY arm has the same mention-vs-invocation confusion the rewrite arm avoids.** Its
  write patterns match anywhere in a Bash command, so a heredoc *containing* `cp … blue/report.md`
  is refused as though it were a write. Measured by hitting it while writing these tests.
- **The tracked `tools/feov-record` binary reports 0.17.0** against a 0.36.0 manifest, and
  nothing in the repo references it. With #282's preflight it would now be refused rather than
  used. It is a stray build artifact, not a carrier — removal is a call for gb, not a rebuild.
  **RESOLVED 2026-09-02: deleted by #293 (`3e5ffa33`, "stop committing binaries — delete the
  stray one, and gate the class"); `git ls-files` no longer lists it.**

## Original hand-over note (superseded above)

Six revisions, six FAILs at the plan-auditor gate. Handed over rather than taken to a
seventh round, on the judgement that the remaining findings are cheaper to shake out in
execution than in review (#284 records the principles that call rests on).

**This plan is NOT approved.** The known-open items are listed in #281's comment thread and
must be resolved during implementation, not assumed away:

1. `PreRewrite` / `PreOutcome` — the rename is incomplete at four sites above, including §III's
   tree, which is the buildable list.
2. §3 turns `seat.Of` into a record writer on **every** seat verb including read-only `show`.
   Probably wrong: observe at `register` only, or cut §3 and grade R10 unmitigated.
3. §3's "one friction per seat per run" claims an idempotency `record.Append` does not have —
   it needs an explicit pre-read, the way `ExistingMintByKey` does.
4. R9 is graded against the pre-change state.
5. `--bin-dir` on the launch collides with the documented legacy/pre-record resume flow (#282).
6. The fuzz target needs a seed corpus and a stated oracle.

**What is stable and unchallenged since rev 3** — the part that answers the measured failure —
is §1 and §2: the injection, `export …;` over an inline prefix, the quoting boundary, and
refusal-on-disagreement.
