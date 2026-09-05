# Seat identity: role attested by the harness, instance bound at register, never re-derived (#290)

> [!IMPORTANT]
> **RETIRED 2026-09-03 — SUPERSEDED, NEVER IMPLEMENTED. Kept as historical record.**
>
> This plan was written 2026-08-15/16 against main `e7bf90e`, audited through thirteen
> rounds, and never merged. By `24bd55d1` (688 commits later) every defect it was written
> to fix was closed by a DIFFERENT design that arrived incrementally — #435, #500, #510,
> #512, #526 and the record→store migration — so the document below is a record of
> reasoning, not a description of the tree. **Nothing in §III is a live instruction.**
> The audit that retired it is §0.

## 0. Retirement audit (2026-09-03, against main `24bd55d1`)

**The design that won instead.** This plan's answer was *make the seat id opaque and carry
every fact as a field* (§I, the opaque-id invariant). Main's answer is the **roster gate**:
`record/roster.go` refuses any registered seat id that does not match the engine's own
grammar (`red-lens-r\d+-L\d+`, `red-merge-r\d+`, …). The string became a schema a writer can
REFUSE, which reaches facts-are-fields by the other route — and it is this plan's own §I
invariant that got dropped rather than the derivations. **The gate is weaker than its own
comment claims** — see open item 3.

**One premise was measured shut.** #290 (2026-08-23) registered `SubagentStart` for real and
captured its payload whole: `agent_type` names a ROLE covering up to four seats (four types,
thirteen seats), and nothing the dispatcher supplies reaches the hook. A hook cannot bind a
SEAT. This plan survives that measurement only because it bound at `register` rather than at
the hook — the fork resolved in §II, for a different reason.

| Plan item | Status on `24bd55d1` | How it actually landed |
|---|---|---|
| **A** — hook exports `FEOV_AGENT_ID` | **landed, wider** | `hookgate.go:287` injects per-variable, UNCONDITIONALLY and with no matcher. This plan's "when the command invokes feov-record" gate was itself the defect: #510 measured `RB="…/feov-record"` costing 21 of 65 registers their agent_id. |
| **A** — hook exports `FEOV_AGENT_TYPE` | **not landed** | `hookgate.Input.AgentType` (`hookgate.go:34`) is still read only by the report lockdown. |
| **B.1** — `agents/red-merge.md` fork | **not landed** | `debate.js:845` (lens) and `:871` (merge) both dispatch `frank-exchange-of-views:red-auditor`. |
| **B.2** — `register --round <n>` | **moot** | `RoundIn` (`round.go:60`) derives a terminal seat's round FROM RECORD FIELDS — the last round any seat stamped. `FEOV_ROUND` was deleted as a reader with no writer, which is what this plan proposed to give one. |
| **B.3** — petition sitting ids | landed #433 | already reconciled in the text below. |
| **C.1** — `record/agentrole.go` | **not landed** | the live work pulls the other way: #618 proposes deriving the four roles from a single `record.SeatRoles()`. |
| **C.2** — `record/binding.go` | **landed, better** | no per-`agent_id` side file: `agent_id` is a FIELD on the register event (`record.go:240`), read back by `record.SeatOfAgent`. The nonce is gone, shards are gone, and `DiscardedForSeat` is unrepresentable under the store rather than merely clean. |
| **C.3** — resolve order + refusal | **landed** | `requireBound` (`seat.go:191`) refuses every non-register verb from an unbound agent; `seatenv.ResolveSeat` refuses a `--seat-id` disagreeing with the binding (`identity.go:113`). |
| **C.4** — stamps from resolved identity | **round yes, role no** | `envelope()` (`record.go:377`) still calls `roleOfSeat(seatID)` at the write. The READ side is honest — `PartyOf` reads the stamped field — so #396's filed defect is closed; the write-side derivation now runs over a roster-refused id. |
| **D** — derivation census (18 hits) | **every anchor stale** | 17 hits today; the plan's line numbers no longer resolve. |
| **E** — fuzz regex-path counter | **not landed** | no counter in `internal/fuzz/`. |

**Issue disposition.** #396 and #394 closed. #290 stays open on doctrine, with #345 and #355
unmoved behind it.

**What the audit left genuinely open**, both carried out of this document as issues rather
than as a revived plan:

1. **Role is still self-asserted at `register`** — #674. The roster bounds SHAPE, never MEMBERSHIP —
   its own comment says so — so a `lead-judge` agent registering as `red-merge-r1` is
   accepted. Stage A's second export plus Stage C.1's table would refuse three of the four
   roles today and all four after the red fork. No measured failure of this class is on file:
   it is hardening, not a fix, and it is sized as one.
2. **`consistency.go:438` (#676) cross-checks each event's stamped `Round` against
   `RoundOf(seatID)`** — precisely the regex-over-a-composed-string check §I argued out of
   existence. As a post-hoc audit rather than a write-path derivation it is defensible, but
   it is this plan's REJECTED design running in a corner nobody reconciled.
3. **The roster gate cites a verification that does not exist** — #675. `roster.go:27` — "Every id
   below is verified against debate.js's own dispatch sites; see
   `TestTheRosterMatchesWhatTheEngineActuallyDispatches`" — names a test present nowhere in
   the tree. The only reconciliation that exists, `TestTheRosterAndTheRoleTableAgree`
   (`durability_test.go:212`), compares `seatShapes` to `roleSeats`: two hand-written Go
   tables stating one vocabulary, neither of them the engine. Nothing anywhere compares
   `seatShapes` to `debate.js`. The harm is already measured in the other direction, by a
   sibling test's own comment: "The tool's roster accepts `judge-r\d+`, so nothing refused it
   — the seat was valid, registered, dispatched and scored, and the orchestrator has never
   seated it" (`seatprobe/fidelity_test.go:138`). This matters to the retirement because the
   roster is the whole reason this plan's opaque-id invariant was safe to drop.

---

## The plan as written, 2026-08-16 — unchanged below this line

Written 2026-08-15, revised 2026-08-16 through eleven auditor rounds and two gb
directions: **provide both the agent and the role from the agent configuration — only the
instance discriminators are engine-composed** (config forking accepted), and **the
opaque-id invariant** ("beware things you construct into strings only to later parse with
regexes"). The mechanism half of #290; the doctrine half shipped as facts-are-fields
(#285/#287). Sequenced by plans/design-review-2026-08-15.md §V.4. #345 (entity-grouped
tree) stays blocked on this and OUT of this plan.

**Re-baselined 2026-08-16 against main `e7bf90e`** after #433/#421/#424/#430 landed
mid-plan: #433 closed #394 (petitioner-derived sitting ids + the discarded-events
audit), so Stage B shrinks to the fork and `--round`; f3e8c63 added
`RequireDispatchedSeat`, a new Stage D census row; ffc4bf4 wiped the July/early-August
research corpus, so §V.6 repoints to the surviving real run. Line anchors cited from the
pre-#433 tree drift by a few lines (e.g. viewjson.go:433→:456, seat/seat.go:256→:296,
roles.go:66/:84→:71/:126); §V.7's re-run re-baselines every census at implementation
start.

## I. Summary & Goals

**Problem.** Every seat asserts its own identity: debate.js interpolates a SEAT_ID into
the prompt, the seat types it back on every call, and the tool believes it — then
re-derives round and role from the string at every append (`record.go:226`, `:394`).
On the PRODUCTION HOOK PATH nothing sets FEOV_SEAT/FEOV_ROUND (#396: seatenv's readers
have no production writer — the seatprobe harness sets both in its own env,
`cmd/seatprobe/main.go:312/:316`, which is a test rig, not the run path), so the regex
fallback is production's only path. Measured failures on file:

- `judge-terminal` stamps round 0 on every append — the phantom-archive defect #348 was
  filed for, still live (#396).
- Seven petition sittings share the literal seat id `judge-petition`; nonce rotation plus
  the terminal-event replay winner discards **every sitting but the last** from replay —
  live data loss of granted rulings (#394).
- A typo'd `--seat-id` files events under a phantom seat with nothing to refuse it — the
  class `--run` injection closed (#281).

**The design, in identity's three parts:**

| Fact | Source after this plan | How it arrives |
|---|---|---|
| **Role** (lens/merge/blue/bench) | the AGENT CONFIGURATION — `agent_type`, harness-attested, unforgeable by the seat | PreToolUse payload → `FEOV_AGENT_TYPE` env → one tool-side table |
| **Round** | the ENGINE — a fact it holds at every dispatch | `register --round <n>`, stored in the binding |
| **Instance** (lens slice, lane number, sitting seq) | the ENGINE — composed into the seat id | bound once at register to `agent_id`, injected thereafter |

The seat types nothing twice and can forge nothing: role never passes through the seat at
all; seat id and round pass through once, at register, where the tool validates and binds
them to `agent_id` (collision-free by construction — one file per agent, no shared-file
race under parallel dispatch). Everything after register is injection and
refusal-on-disagreement. One config fork makes the role column total: red-merge today
shares `red-auditor` with the lenses, so it gets its own agent type (gb: config forking
accepted where it simplifies).

**The opaque-id invariant (gb, 2026-08-16 — "beware things you construct into strings
only to later parse with regexes").** On the bound path, NO VALUE IS RECOVERED from the
seat id into a record. Every fact a machine stamps or stores travels as a field — role
from the attestation, round from the binding, instance from `agent_id`. The
`-r<N>`/`-s<K>` fragments in engine-composed ids are for the human reading a dashboard or
a shard filename; no code this plan adds may read a value out of them, and Stage E's
regex-path counter is the gate holding the bound path's Round/Role derivations to zero.

Exactly FOUR id-string reads survive, each with a stated reason:
1. `RoundOf`'s inference tail — UNBOUND callers only;
2. `PartyOf`'s prefix fallback — pre-`Role` RECORDS only;
3. `CheckSeatRole`'s prefix match — active on the bound path as a REFUSAL layer, and the
   asymmetry with the dropped round cross-check is deliberate: the id's role prefix is
   the shard/concurrency NAMESPACE (the same string keys the shard, so a namespace
   disagreeing with the attestation is a dispatch fault worth stopping), while the
   `-r<N>` fragment keys nothing and drift against the binding is merely cosmetic;
4. `RoleOf`'s lens-index regex (`findinglabel.go:16`), whose caller `NextFindingLabel`
   stamps `L<lens>-F<N>` into finding events — the ONE recovery-into-a-record that
   survives, DELIBERATELY: the per-seat prefix is what makes that lock-free allocation
   safe under parallel dispatch, and collapsing it once made 39 of 60 disposals ambiguous
   (facts-are-fields clause 3 exists because of this exact case; #290's appendix says a
   per-`agent_id` counter namespace is the eventual replacement). Replacing it is #345-
   era work over the binding this plan creates — OUT of this plan, counted rather than
   hidden. `viewjson.go:433` reads the same regex into a projection field for the same
   namespace.

**Success criteria (quantitative, checked in §V):**
1. Post-change smoke: 100% of engine-dispatched seats' events stamp Round/Role from
   binding/attestation — zero appends through the regex path, asserted by a test-hook
   counter (§III.E).
2. `judge-terminal` and petition-sitting events carry their real round, not 0.
3. A synthetic run with ≥2 petition sittings replays ALL sittings' rulings — zero hits
   in #433's `discarded-events` audit (the mechanism landed with #394's fix; this plan
   keeps the criterion as a regression check).
4. A `--seat-id` disagreeing with the binding, and a role tree disagreeing with the
   attested role, are refused naming both values.
5. Zero behaviour change with no `FEOV_AGENT_ID`/`FEOV_AGENT_TYPE` present (operator, CI,
   tests, hookless bootstrap window) — and an unbound engine run is *visible* (capture
   coverage row), never silently degraded.

## II. Technical Context

- **Language/deps:** Go 1.25 (tools/), cobra/pflag; debate.js (sandboxed Workflow JS — no
  fs/env; facts reach it as args, leave only in prompts); node --test simulator suite
  (CI: `.github/workflows/hooks.yml:368` runs `tests/simulator/debate.test.mjs` +
  `prompts.test.mjs`).
- **Measured harness facts this stands on:**
  - PreToolUse carries `agent_id` (plans/hook-surface-spike.md §7, recorded by #399):
    9/9 across Bash/Read/Grep, stable within an agent, distinct across concurrent agents,
    byte-identical at SubagentStop. `session_id`/`prompt_id` do NOT discriminate.
  - PreToolUse carries `agent_type` **in production today**: `hookgate.Input.AgentType`
    (hookgate.go:33) is how the blue-report lockdown allowlists
    `frank-exchange-of-views:blue-synthesizer` on every write. The role attestation reuses
    a field the lockdown has been reading since 0.27.0 — no new harness claim.
- **Existing machinery reused:** the hookgate rewrite arm (quoting boundary, heredoc
  exclusion, mention-vs-invocation, deny-wins — fuzzed, 3.7M executions);
  `seatenv.ResolveSeat` (injected → flag → inference with refusal-on-disagreement,
  `Round == -1` for unknown — readers shipped, writer missing); `record.RegisterSeat`
  nonce rotation (crash re-dispatch, 8/50 measured — untouched: #394's defect is one seat
  id worn by seven contracts, not the rotation).
- **Storage:** one JSON file per `agent_id` under the run's `RecordsDir` (bindings travel
  with the shards under record separation). Durable-written like the pointer files. NOT an
  event: run-scoped resolve-time state like `.active-<seat>` and `.clock` — replay never
  reads it, so old records replay bit-identically (a record is permanent).
- **Resolved design forks:**
  - *Bind at register, not SubagentStart* (gb, 2026-08-15): SubagentStart carries
    agent_id + agent_type but the seat id is engine-assigned in the prompt; recovering it
    hook-side is a regex over prose. The hook ferries what only it holds; the tool binds
    where the seat names itself with a full parser.
  - *Role from agent config, forking configs where needed* (gb, 2026-08-16): role is
    attested per call from `agent_type`, so it is never bound, never typed, and never
    guessed; only instance discriminators remain engine-composed strings. red-merge gets
    its own agent type; `red-auditor` remains the lens type (keeping the memory-dir key
    and the 40+ doc mentions stable — the merge is the mover).

## III. Proposed Changes (the spec)

Proposed structure (new and touched files):

```
plugins/frank-exchange-of-views/
├── agents/
│   ├── red-auditor.md                       [MODIFY]  lens duties + shared telos remain
│   └── red-merge.md                         [NEW]     merge duties move here
├── skills/research-protocol/scripts/
│   └── debate.js                            [MODIFY]  --round, sitting ids, :761 agentType
├── hooks/hooks.json                         [MODIFY]  _comment prefix prose
├── tests/simulator/{debate,prompts}.test.mjs [MODIFY] dispatch ids, register wording
├── tools/
│   ├── cmd/seatprobe/main.go                [MODIFY]  env block + constitution map
│   └── internal/
│       ├── hookgate/                        [MODIFY]  three-export prefix (+ inject_test)
│       ├── record/
│       │   ├── agentrole.go                 [NEW]     agent_type → role table
│       │   ├── binding.go                   [NEW]     per-agent_id binding
│       │   ├── record.go / roles.go         [MODIFY]  stamps from resolved identity
│       │   └── (replay untouched — bindings are never read at replay)
│       ├── cli/seat/                        [MODIFY]  Begin/Context resolve order
│       ├── setup/run.go                     [MODIFY]  memory-dir union
│       ├── seatenv/identity.go              [MODIFY]  doc contract rewritten
│       └── fuzz/                            [MODIFY]  env injection + regex counter + gates
└── docs/{seat-command-triggers,record-flow}.md [MODIFY]
```

### Stage A — the hook ferries the two harness facts `[MODIFY: hookgate]`

The PreToolUse rewrite prefix grows from `export FEOV_RUN='…'; ` to
`export FEOV_RUN='…'; export FEOV_AGENT_ID='…'; export FEOV_AGENT_TYPE='…'; ` when the
payload carries them and the command invokes feov-record in command position. Same
single-quote boundary, heredoc exclusion, deny-wins ordering. Independently shippable
(readers arrive in C); contract version moves at C where behaviour does.

**Consumer census** (`grep -rn "FEOV_RUN" plugins/frank-exchange-of-views --include="*.go"
--include="*.json" --include="*.md"`, re-run 2026-08-16 — widened past `*.go` after the
round-2 audit caught the prose carriers):
- emitter `hookgate.go` — changes;
- `hookgate/inject_test.go` (:33/:82/:104/:111 — unit tests asserting the exact prefix
  shape) — change with the emitter;
- rewrite fuzz target — extends round-trip/idempotency/refusal invariants to all three
  exports — changes;
- `hooks/hooks.json` `_comment` — documents the prefix verbatim
  (`export FEOV_RUN='<runDir>'; `) — prose carrier, updated in-PR;
- `seatenv/identity.go:40-42` — the doc contract "these constants have readers and no
  writer" becomes FALSE at Stage C — the comment is a carrier of the old model and is
  rewritten in the C PR;
- reader `seatenv/seatenv.go`, comment mentions in `cli/seat/seat_test.go`, doc-comment
  carriers `cli/root.go:205-209`, `cli/seat/verbs.go:252` — updated in-PR;
- `cmd/seatprobe/main.go:302-316` — the probe's per-seat env block already injects
  `seatenv.Var` (FEOV_RUN), FEOV_SEAT (:312) and FEOV_ROUND=1 (:316): a carrier of the
  injection model, and a `[MODIFY]` in Stage E (it must inject FEOV_AGENT_ID and
  FEOV_AGENT_TYPE per board seat, or every post-change probe board runs the legacy env
  path and §V.4's refusal goldens are undriveable) — changes;
- gray-area's shell resolver reads transcripts, not env — unaffected.

### Stage B — the engine states the facts it holds `[MODIFY: debate.js, agents/, setup]`

1. **Config fork: `agents/red-merge.md` `[NEW]`.** The merge duties (coalesce, board acts,
   verdict, closing discipline) move out of `red-auditor.md`, which keeps the lens duties
   (leaf verification, findings, evidence) and the shared red telos (duplicated initially;
   factoring into a skill is #255's pattern and out of scope). debate.js:761 dispatches
   red-merge as `frank-exchange-of-views:red-merge`; debate.js:748 (lenses) stays
   `red-auditor`.
2. **`register --round <n>` in every register instruction** — 0 for frontier/lanes/
   synthesis, `${round}` in-loop, the final round for judge-terminal/assemble, the
   sitting's round for petitions. The engine knows each as a fact.
3. ~~Unique petition sitting ids~~ **LANDED by #433 (2026-08-16), differently and
   better**: the sitting id derives from its petitioner (`judge-petition-red-merge-r1`) —
   unique by construction, not by a counter — plus a `discarded-events` capture audit
   that FAILS when a losing shard held work no surviving shard carries. #394 is closed.
   One interim note this plan supersedes: #433 fixed the round stamp by leaning on
   `RoundOf`'s first `-r<N>` match — the regex path. Under Stage C the binding's
   `--round` takes over as the round fact for these sittings, and the id goes opaque like
   every other (the invariant, §I).

**Consumer censuses, run 2026-08-16:**

*agent_type consumers* (`grep -rn "red-auditor\|blue-synthesizer\|blue-researcher\|
lead-judge\|agent_type" --include="*.go" --include="*.js" --include="*.mjs"
--include="*.md" --include="*.json"`, corrected after the round-2 audit):
- debate.js — **12** `agentType:` dispatch sites (:628, :655, :661, :665, :669, :748,
  :761, :934, :938, :962, :1019, :1046 — :938 is ensureRoundRecord's respond-retry); ONE
  changes (:761 red-merge). Lenses (:748), lead-judge sites, blue-researcher sites
  (respond-retry included), blue-synthesizer sites unchanged.
- `hookgate.AuthorAgentType` (hookgate.go:21) — blue-synthesizer; unaffected by the red
  fork. Lockdown-mention carriers that also do not change: `hooks/hooks.json` `_comment`,
  `cli/root.go:304`, `hookgate_test.go:37`, `docs/finding-markers.md`.
- `record/enums.go:40` and `cli/enum_help_test.go:50` — comments citing "the red-auditor
  prompt" as the carrier of enum-value prose: if those clauses move into `red-merge.md`
  in the fork (they are merge duties — closure classes), both comments go stale and are
  updated in the fork PR — change.
- The sibling agent constitutions `agents/blue-researcher.md`, `blue-synthesizer.md`,
  `lead-judge.md` — census hits by self-name only; none carries a red-auditor reference
  the fork would stale — unchanged (`red-auditor.md` itself is the §III-tree `[MODIFY]`).
- Repo-root remainder, dispositioned as classes WITHOUT counts (counts falsify as
  siblings grow — the gray-area line was written at 7 files and was 9 by the next audit;
  the class is the disposition, so §V.7's re-run reconciles by class membership):
  `plugins/gray-area/` (`red-auditor` as opaque fixture values, `agent_type` as an opaque
  manifest/JSON field) — unchanged;
  `plugins/prosthetic-conscience/tools/internal/checkpointseal/` (2 files — opaque
  `agent_type` JSON field) — unchanged; `scripts/rulesweep/sweep_test.go:35` (fixture
  path) — unchanged; root `README.md` (no red-fork string) — unchanged; `research/`,
  `plans/`, `ideas/` — historical and design corpus, records of their moment, untouched.
- `cmd/seatprobe/main.go:361-364` — constitution map (`lens`/`merge` both →
  `red-auditor.md`): merge row repoints to `red-merge.md` — changes.
- `internal/seatprobe/naming.go:28-31` and `docs/seat-surface-naming.md:17-20` (landed
  post-round-11 by 29eb10b) — both state `merge → red-auditor.md` / `lens →
  red-auditor.md`: the merge rows change with the fork or ship speaking the old model —
  change.
- `internal/seatprobe/naming_test.go` — their sibling, and the fork BREAKS it two ways:
  `TestRedactionKeepsTheSituationItWithholdsTheVerbFor` (:52-62) asserts red-auditor.md's
  "A grade moves through `regrade`" clause survives redaction — a MERGE duty the fork
  moves to `red-merge.md`, so the assertion re-anchors to the file that keeps the clause;
  and `TestRedactionRemovesTheNamesFromTheRealConstitutions` (:19) enumerates exactly
  four constitution files, so `red-merge.md` must join the list or ship outside the
  redaction/vacuous-pass gate — changes.
- `setup/run.go:141` — red memory dir keyed `frank-exchange-of-views-red-auditor`:
  becomes the UNION of `-red-auditor` and `-red-merge` agent-memory dirs (the merge is a
  gap-pattern writer; a forked type is a forked memory key) — changes, with
  `commands/research.md:26` (the --memory-dir guidance) and `setup_test`/`run_test`.
- `feov-memory/red-gap-patterns/README.md:26` — a REPO-level carrier of the one-key
  model (documents raw accrual at `…-red-auditor/` as the single red memory dir), which
  the plugin-scoped census could not see: the agent_type census runs from the REPO ROOT,
  and this file changes with the union work — changes.
- `docs/seat-command-triggers.md`, README agent list — prose carriers, updated in-PR.
- `record/testdata/pre-motion-real-run/PROVENANCE.md` — historical record, untouched.

*register-instruction / seat-id consumers* (`grep -n "recordClause(\|register"
debate.js`; `grep -n "judge-petition\|register" tests/simulator/*.mjs`):
- debate.js — `recordClause` is interpolated at 10 sites and carries the register
  instruction text; the `--round` value threads through `recordClause(seatId, tool,
  round)` — one function signature, 10 call sites, all changing mechanically.
- `tests/simulator/debate.test.mjs:885` — `dispatches()` matches
  `feov-record"? <role> register` — tolerant of a trailing `--round` (regex is
  unanchored); asserted, not assumed: the suite runs in §V.5.
- `tests/simulator/debate.test.mjs` petition matchers and `harness.mjs:90` — already
  moved to #433's petitioner-derived shape (the suite asserts
  `/^judge-petition-(red-merge|blue-respond)-r1$/` at :1270) — unchanged by this plan.
- `tests/simulator/prompts.test.mjs` — prompt-wording assertions; re-run in §V.5, updated
  where register wording moved.
- `fuzz/promptverbs_test.go:183-186` — this is the EXEMPT map, which today excuses
  `register` from prompt-naming ("issued by the harness prefix rather than asked for in
  prose"). With `--round` required in the register instruction, the exemption is
  REVISITED: register IS named in recordClause prose, so it leaves the exempt map, and a
  new `TestRegisterInstructionCarriesRound` (same file, beside
  `TestEveryRecordingVerbIsNamedInAPrompt`) asserts every register instruction in
  debate.js carries `--round` — the gate that keeps the flag from silently vanishing from
  a prompt.
- Seat-prompt goldens — regenerate; accepting them is a decision (#401), diffed in review.
- `docs/seat-command-triggers.md` — register row gains `--round`.

### Stage C — attest, bind, resolve `[MODIFY: record, cli/seat; NEW: record/binding.go, record/agentrole.go]`

1. `[NEW]` `record/agentrole.go` — ONE table `agent_type → role`, keyed by the FULL
   PREFIXED string exactly as the harness delivers it and as
   `hookgate.AuthorAgentType` already spells it
   (`frank-exchange-of-views:blue-researcher`→blue, `…:blue-synthesizer`→blue,
   `…:red-auditor`→lens, `…:red-merge`→merge, `…:lead-judge`→bench; anything else,
   or absent → unattested — a bare-keyed table would map every real seat to unattested).
   The same table is what #345 will later generate scoped help from — one source, two
   readers, per enums.go's rule.
2. `[NEW]` `record/binding.go` — `Binding{SeatID, Round, ToolVersion, BoundAt}` written
   durably to `<records>/seats/<agent_id>.json` at register when `FEOV_AGENT_ID` is
   present. Refusals at the write:
   - same agent_id, different seat id → refused (one agent is one seat for its life;
     same-id re-register rebinds the nonce exactly as today);
   - `--round` is the round FACT, stored as a field; it is NOT cross-checked against a
     `-r<N>` fragment in the seat id — an earlier draft refused that disagreement, and the
     check was itself a regex over a string the engine composed, re-legitimizing the id
     as a round-carrier at the moment the binding retires it (gb's opaque-id invariant,
     §I). If the engine's id label and its --round ever drift, the label is merely
     cosmetic and the binding is right;
   - seat id's role tree disagreeing with the ATTESTED role → refused naming both (a
     lens structurally cannot register — or act — as merge);
   - no `FEOV_AGENT_ID` → no binding; register proceeds as today and the register event
     carries `bound: false`, so capture counts unbound sittings (§V.6 coverage row).
3. `[MODIFY]` `cli/seat` (`Begin`/`Context`): role resolves **attested
   (FEOV_AGENT_TYPE) → tree-implied** (refusing disagreement); seat+round resolve
   **binding (FEOV_AGENT_ID) → FEOV_SEAT/FEOV_ROUND env → flags → inference**, refusing
   any disagreement between present sources (seatenv's shipped contract, now with its
   writer). `Context` carries the resolved identity every verb hands down.
4. `[MODIFY]` `record.Append`/`RegisterSeat` stamp `Round:`/`Role:` from the resolved
   identity passed in — the signature takes it; no derivation inside.

**Consumer census** (`grep -rn "record.Append(" --include="*.go" | grep -v _test`, run
2026-08-15, re-verified by the auditor): **32 sites, 22 files**, all under `cli/` inside
`seat.Begin`'s resolved context — mechanical threading. `RegisterSeat`: 2 non-test
callers (`seat/verbs.go`; `record.go` activeNonce's implicit register, which passes the
identity it was invoked under). Census re-runs post-change; a surviving derivation site
fails the PR (§V.7).

### Stage D — the string-derivation boundary, enumerated `[MODIFY: record.RoundOf, roles.go]`

Full caller census — the command reproduces its own paste verbatim (round-5 fix: the
declaration lines are excluded IN the command, so §V.7's re-run cannot trip on its own
baseline; round-4 fix before it: `-E`, because a literal-`|` grep reports empty forever,
the plausible zero). From `tools/`:

```
grep -rnE "RoundOf\(|roleOfSeat\(|PartyOf\(|CheckSeatRole\(|RoleOf\(" --include="*.go" internal \
  | grep -v _test | grep -vE "func (RoundOf|roleOfSeat|PartyOf|CheckSeatRole|RoleOf)\("
```

Re-run 2026-08-16 against post-#433 main (widened to `RoleOf\(` after the round-10 audit
caught the pattern structurally blind to the lens-index read): **18 hits** —
seatenv/identity.go:16 (comment), view.go:495, seatprobe.go:179,
assemble.go:870/872/874/876, viewjson.go:625, roles.go:71/:81/:126, spotcheck.go:99,
record.go:226/:394, motion/command.go:112, seat/seat.go:296, findinglabel.go:31,
viewjson.go:456. The excluded declarations are the functions whose fates the table
states. The 18th hit is NEW on main (f3e8c63): `RequireDispatchedSeat` at roles.go:81 —
see its table row. The fate of each hit:

| Site | What it does | Fate |
|---|---|---|
| `record.go:226`, `:394` (RegisterSeat/Append stamps) | re-derive at every write | **DELETED** — stamps come from resolved identity (C.4) |
| `RoundOf` env branch + regex | write-path fallback | survives at ONE place only: the unattested-caller inference tail (C.3) — the bind-time cross-check an earlier draft kept is dropped per the opaque-id invariant (§I) |
| `roleOfSeat` at `roles.go:66` (`PartyOf`) | READ-path fallback for pre-`Role` events | **SURVIVES UNTOUCHED** — old records must keep reading (§II constraint, R5); callers `view.go:495`, `viewjson.go:602`, `assemble.go:870-876`, `spotcheck.go:99` unchanged |
| `roleOfSeat` at `roles.go:126` (`CheckSeatRole`) | refusal text naming the owning role | survives — callers `seat/seat.go:296`, `motion/command.go:112`, `seatprobe.go:179`; its check is now the tree-vs-attested cross-check's second layer |
| `roleOfSeat` at `roles.go:81` (`RequireDispatchedSeat`, NEW in f3e8c63) | a prefix-SHAPE heuristic for "the engine created this seat" (a hand-invented `red-lens-r9-L9` passes it) | on the bound path this UPGRADES to a binding lookup — dispatched iff bound, exact rather than shape; the prefix heuristic survives as the unbound fallback |
| `RoleOf` at `findinglabel.go:31` (`NextFindingLabel`) and `viewjson.go:433` | lens-index namespace — the one surviving recovery-into-a-record | **SURVIVES UNTOUCHED**, counted not hidden (§I invariant read 4): the per-seat prefix is the lock-free allocation's safety under parallel dispatch; replacement is a per-`agent_id` counter namespace over this plan's binding, #345-era |
| `Event.Role` doc comment ("never re-derived") | currently false (#396) | becomes true; `RoundOf`'s 30-line danger comment shrinks to the two surviving facts |

### Stage E — the test harnesses exercise the injection `[MODIFY: internal/fuzz, cmd/seatprobe]`

`cmd/seatprobe/main.go:302-316` — the probe's per-seat env block (already injecting
FEOV_RUN/FEOV_SEAT/FEOV_ROUND) gains FEOV_AGENT_ID (deterministic per board seat) and
FEOV_AGENT_TYPE (the seat's constitution row's type), so probed seats run the
binding/attestation path and §V.4's refusal goldens are driveable.

The goja harness today execs the binary with `--run` as a flag and sets no env
(fuzz_test.go:289) — it cannot see this plan's machinery. It gains: per-simulated-seat
`FEOV_AGENT_ID` (deterministic per seat per seed) and `FEOV_AGENT_TYPE` (from the
dispatch's agentType) in the exec env, mirroring the rewrite's effect; and a test hook
counting regex-path ROUND/ROLE derivations (`RoundOf`'s regex, `roleOfSeat` at a write),
asserted ZERO for engine-dispatched seats — the counter's scope deliberately EXCLUDES
`RoleOf`'s lens-index read, which is the surviving namespace recovery (§I invariant read
4) and would make zero unattainable on any run with lens findings. The petition-seat
literals already carry #433's petitioner-derived shape (`judge-petition-${who}`,
simulator-asserted at debate.test.mjs:1270) — nothing for this plan to rename.

**Go-side seat-id literal census** (`grep -rn "judge-petition" --include="*.go"`,
re-run 2026-08-16 against post-#433 main — #433 landed the petitioner-derived id shape,
so this census DISPOSITIONS against the landed model rather than renaming anything):
`record/motion.go:189` — comment, no join — unchanged; `roles.go:30-35` — the roleSeats
doc comment ALREADY speaks `judge-petition-<petitioner>` (#433 rewrote it) — unchanged;
`cli/root.go:882/:888/:893` — #433's changelog/case-law comment, a record of its moment
— untouched; `capture/capture_test.go:620-631` — deliberately uses the BARE
`judge-petition` id as the collision-regression fixture and MUST NOT be updated to the
new shape (the test exists to prove the old shape's data loss is detected) — untouched;
`seatclass/seatclass.go:43/:91`, `seatclass_test.go:20`, `dispatch_bind_test.go:26` —
prompt-head and label-prefix keyed, tolerant — unchanged; the remaining Go test hits
(`assemble_test.go:191`, `verbs_test.go:157`, `roles_test.go:13`, `replay_test.go:324`,
`difftest/scenarios_test.go:194-197`, `fuzz_test.go` comments/:979) — deliberately carry
the bare id: they drive the tool directly, where the id is just a seat string, so they
are shape-tolerant fixtures, not carriers #433 "moved" — unchanged; and
`replay_test.go:1096-1111` is a SECOND bare-id discard-regression fixture with the same
must-NOT-move property as `capture_test.go:620-631` — both exist to prove the old
shape's data loss is detected, and modernizing either deletes its reason — untouched.

### Version & docs (same PR as C/D/E)

recordToolVersion bump (new register flag, binding + attestation refusals — a stale
binary ignores all three env vars and accepts a disagreeing --seat-id); plugin version
bump; changelog entry; `docs/seat-command-triggers.md`, `docs/record-flow.md` updated.
Closes #396 (#394 was closed by #433); #290 closes when #345 lands on top.

### Out of scope, stated

#345 (entity tree, identity-scoped help — unblocked, own plan; agentrole.go is built to
be its table). #355 (register's survival — decided at stage 6; register is the binding
moment here, evidence FOR keeping it). Factoring the shared red telos into a skill
(#255's pattern). Deleting `--seat-id` (stays as the cross-check surface).

## IV. Risk & Mitigation

| # | Risk | L | I | Cx | Mitigation |
|---|---|---|---|---|---|
| R1 | Hookless window: no env, seats self-asserted | med | med | low | `bound: false` on register + capture coverage row (§V.6) — visible, never silent; status quo is the floor |
| R2 | Append signature ripple (32 sites) — mechanical slip | med | high | low | Compiler enumerates; `-race` + difftest + goja fuzz with injection (§V.1–3); census re-run gates (§V.7) |
| R3 | Constitution fork degrades red (merge loses lens context or vice versa) | med | med | med | The fork MOVES text, deletes none; seat-probe boards for lens and merge run before/after (§V.4); shared telos duplicated verbatim until #255 factors it |
| R4 | Memory-dir split loses gap patterns | med | high | low | setup reads the UNION of both dirs + the promoted corpus; run_test gains the two-dir case; the mostly-undeliverable-corpus guard (#300) already fails setup loudly |
| R5 | Old runs stop replaying | low | high | low | Binding/attestation are resolve-time only; PartyOf fallback untouched (Stage D table); difftest's recorded pre-binding scenarios + §V.6's real-records re-run |
| R6 | Prompt gates green while register wording regresses (the 0.59.0 class) | med | med | med | promptverbs gains the `--round` assertion; simulator + prompts.test.mjs run in CI already (hooks.yml:368); goldens diffed as a decision (#401); live probe (§V.4) |
| R7 | Petition id shape breaks a literal join | low | high | low | Largely RETIRED by #433, which landed the shape with its own gates (simulator regex assert, discarded-events audit, dispatch-bind literal-prefix gate); §III.E's census dispositions every remaining Go literal against the landed model, including the bare-id collision-regression fixture that must not move |

## V. Verification Plan

1. `cd plugins/frank-exchange-of-views/tools && go test -race -count=1 ./...` — unit,
   difftest (pre-binding records replay unchanged), hookgate fuzz extended to the three
   exports. Includes the UNBOUND path's check (success criterion 5's second half): a
   register with no `FEOV_AGENT_ID` writes `bound: false`, and capture's coverage row
   counts that sitting as unbound — the unhooked run is visible in the row, not folded
   into the green.
2. `go test -race -count=1 -run 'TestFuzzDebate|TestFuzzHaltPath' ./internal/fuzz/` — the
   real debate.js against the real binary WITH Stage E's env injection; asserts the
   regex-path counter reads zero for engine seats (success criterion 1).
3. `go test -race -count=1 -run
   'TestEveryVerbNamedInAPromptExists|TestEveryRecordingVerbIsNamedInAPrompt|TestRegisterInstructionCarriesRound'
   ./internal/fuzz/` — the two existing prompt gates plus the new register-`--round`
   assertion (§III.B); the command names the real tests, and CI's gate set runs the same
   package so a renamed test cannot leave a vacuous `-run` behind unnoticed.
4. **Seat-probe (real agents):** `go run ./cmd/seatprobe` on lens AND merge boards under
   the forked constitutions — a forged `--seat-id` against the binding, and a wrong role
   tree under an attested type, are each refused naming both values (goldens for both
   refusals; the earlier draft's third golden — a `--round` contradicting the id — is
   gone with the check, per the opaque-id invariant); board outcomes compared against the
   pre-fork probe runs (R3).
5. `node --test plugins/frank-exchange-of-views/tests/simulator/debate.test.mjs
   plugins/frank-exchange-of-views/tests/simulator/prompts.test.mjs` — the zero-token
   engine suite over the new dispatch ids, register wording, and petition routing.
6. **Real data, driveable:** `feov-record capture` + `verify` re-run against
   `plugins/frank-exchange-of-views/tools/research/2026-08-10_dual-read-vs-migration/records/`
   (full path from the repo root — root `research/` is empty post-ffc4bf4; this run
   lives under tools/ and survived the wipe) — real pre-binding records, byte-identical
   board, zero new anomalies. Then one live haiku smoke with the new engine: `judge-terminal` events
   carry the final round (not 0); petition sittings replay in full (the #433
   discarded-events audit reads zero); capture's binding-coverage row reads 100% for
   engine seats.
7. Re-run every §III census after the change; a surfaced site not on a census fails the
   PR.
8. **Auditor gate:** `/plan-audit` PASS on this document before implementation.
