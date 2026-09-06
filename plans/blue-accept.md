# Blue accepts red's proposal: the tool applies the text, blue retypes nothing

Written 2026-09-06. Third revision; the two prior drafts failed audit and are at
`plans/historical/gap-ownership-DRAFT.md`. **Narrowed on 2026-09-06: acceptance APPLIES the edit and does not close the gap** — closure stays
with the merge. **Shape decided on re-audit: a FLAG on `blue edit`, not a new verb.**

## I. Summary & Goals

**Today blue retypes red's words and the tool checks whether it got them right.** `blue edit` takes
`--old`/`--new` from blue; `ProposalAppliedVerbatim` (`estoppel.go:134-146`) then tests
`loc == old && fixNew == new` against the gap's mint row. If blue agrees with red's prescription
entirely it must still transcribe both the span and the replacement exactly — and a stray space, a
smart quote or a re-wrapped line means blue did what red asked while the record says it did not.

**This adds an acceptance path.** Blue names a gap; the tool reads `location` and `fix_new` from
the mint and performs that edit. Blue supplies no prose. `applied_verbatim` becomes true **by
construction rather than by comparison**.

**Why it matters.** Estoppel answers a measured pathology (`estoppel.go:11-24`): in the 2026-08-04
smoke blue made 26 edits, round 1 was nine, every one additive at red's instruction — and **3 of 3**
of round 2's gaps then targeted that new text. *"Blue's discipline was not the problem… It did what
it was told, carefully, and was penalised for it."* The protection attaches only where the tool can
see the text is red's, and today that turns on blue's transcription being byte-perfect.

### Goals
1. Blue can apply red's recorded proposal by naming the gap, supplying **no replacement text**.
   It still supplies its `--reason` — the argument red re-audits against — because accepting is
   an act blue should justify, and `edit.go:49-51` requires it of every edit.
2. On that path `applied_verbatim` is true by construction; no byte comparison decides it.
   The act is **legible as an acceptance** on the record, not inferred from empty prose.
3. The ordinary `blue edit` path and its comparison are unchanged.

### Non-goals — and one correction to the previous draft
Closure, ownership, minting, the `red-merge` rename, every-lens-every-round.

**The previous draft claimed this widens estoppel to short prescriptions. That was false and is
withdrawn.** `ProposalAppliedVerbatim` has **no length floor**; `minEstoppelOverlap = 40` lives only
in `EstoppelConflict` (`estoppel.go:75`), which folds over the *recorded* flag however it was set.
So an accepted 12-character prescription still does not estop a later short quote, exactly as today.
**This plan changes how the flag is SET, and nothing about how it is READ.**

## II. Technical Context

- `Mint.fix_new` holds red's prescribed replacement; the gap's `location` holds the span it
  replaces. `record.proto:604-607`: the span a proposal replaces and the gap's location *"were
  never two facts"*.
- `ProposalAppliedVerbatim` reads exactly that pair and compares in Go *"so it stays byte-exact;
  the record only hands back the pair."*
- `blue edit` sets `AppliedVerbatim` at `cli/blue/edit.go:112-118`, only when `--answers` names a
  gap, and the comment is explicit that the fact is *"computed here — never claimed by either seat"*.
  **That premise is preserved**: on the accept path the tool still computes it, from bytes it
  supplied itself.
- **Blue cannot close, and this plan does not make it able to.** `record.go`'s `Close` validator
  requires `--reason` prose unless `carried_from`; `roles.go:19-23` makes the verb set the role
  boundary and states blue *"is additive-only and never touches the ledger"*;
  `role_boundaries_and_help_contracts.golden` pins `close --seat-id blue-respond-r1` as a refusal.
  All of that stands.
- **A new verb has three mechanically-enforced carriers.** `cli/seat/help.go:175` **panics** without
  `cli/seat/help/<verb>.md` (33 exist today). `seatprobe/surfacecoverage_test.go`'s
  `TestEveryVerbHasABoardThatDemandsIt` reads the verb list from the cobra tree and fails unless
  `seatprobe/boards.go` gains a board demanding it. `difftest/testdata/help_contracts.golden` and
  `role_boundaries_and_help_contracts.golden` pin the role help listings.

## III. Proposed Changes (the spec)

### A. The acceptance path — `blue edit --accept <gap>` `[MODIFY: cli/blue]`

**A flag on `blue edit`, not a new verb.** `cli/seat/help/edit.md:3,7` states edit is *"the only
path into report.md"* and *"The ONLY path into blue/report.md — a raw write or a shell edit is
denied."* A second writing verb would falsify that sentence and require a help document, a
seat-probe board, two regenerated goldens and a fuzz command-path entry. A flag preserves the
invariant and needs none of them.

**What blue supplies:** `--accept <gap>` (which implies `--answers <gap>`), `--reason` (its
argument for accepting — required of every edit today and no less required here), and `--key`
for crash-retry idempotency, exactly as now.

**What the tool supplies:** `--old` and `--new`, read from the mint's `location` and `fix_new`.
Passing either alongside `--accept` is refused rather than merged — two sources for one fact.

**Idempotency is unchanged**: the accept path takes `--key` and reconciles through
`record.ExistingBlueEditByKey` like any edit, so a crash-retry does not append a second stack op.
An accept with no key would, which is why the flag does not exempt it.

**Refusals**, all three:
1. `fix_new` empty — red prescribed no concrete text; there is nothing to accept.
2. `location` no longer matching the document — the prescription is stale; the ordinary path is correct.
3. **The anchor-in-span case.** `planEdit` runs `bluedoc.AnchorsTransitUnchanged` and
   `settleAbuttingAnchor`; `normalizeQuote` skips annotation spans, so a `location` can still match
   after blue placed a citation anchor inside it, while the frozen `fix_new` cannot carry that
   anchor token. The ordinary path's message — *copy the anchor into `--new`* — is unactionable for
   a caller that supplies no `--new`, so the accept path must refuse with its own message naming the
   anchor and directing blue to the ordinary edit. `merge mint`'s `ValidateProposal`
   (`cli/merge/mint.go:138`) only proves the pair was legal at MINT time.

### A2. One schema field, and why "no schema change" is withdrawn `[MODIFY: recordpb]`

`BlueEdit.accepted` (optional bool). **This is a correction, not scope creep.** The previous
revision claimed no schema change and simultaneously claimed §V.6 could count accept-vs-authored
edits — those are incompatible: after this change `applied_verbatim` is byte-identical whether blue
accepted or transcribed perfectly, so provenance is recoverable only by inferring it from empty
prose, which is not a fact.

R1's whole mitigation is that rubber-stamping drift stays **visible**, and Goal 2 asks the act to be
legible. Both need the field. It is set by the tool on the accept path and by nothing else.

### B. Nothing else moves
No disposition, no closure, no role change, no `estoppel.go` logic change, and **no new verb**. The
gap remains open, still owes a manifest row (`record/available.go:77`), and the merge closes it as
it does today.

### Consumer census — run 2026-09-06, commands and full output

```
$ grep -rln "applied_verbatim\|AppliedVerbatim" --include=*.go --include=*.proto tools/     → 11
$ grep -rln "fix_new\|FixNew" --include=*.go --include=*.proto tools/ | grep -v _test        → 9
```

**Census 1 — `applied_verbatim` (11):**

| file | disposition |
|---|---|
| `cli/blue/edit.go` | **changes** — the accept path shares this write; the comparison branch stays |
| `record/estoppel.go` | **comment changes, logic unchanged** — `:38-41` says the fact is *"recorded at edit time by the tool comparing bytes"*, which becomes one of two ways it is set |
| `record/verdict.go:17` | **comment changes** — same stale phrasing |
| `record/recordpb/record.proto` `:1117` + `record.pb.go` | **comment changes** — field doc names comparison as the sole mechanism |
| `view/changes.go` | **unchanged** — renders the flag, indifferent to how it was set |
| `cli/blueedit_test.go`, `record/estoppel_test.go`, `record/queries_parity_test.go`, `recordpb/correspondence_test.go`, `releasegate/fuzz/fuzz_test.go` | **change** — add accept-path cases; existing comparison cases must keep passing |

**Census 2 — `fix_new` non-test (9):**

| file | disposition |
|---|---|
| `cli/merge/mint.go` | **verify** — the writer/validator; §III.A's empty-`fix_new` refusal depends on whether mint can emit one. Assert, do not assume |
| `flags/names.go` | **changes** — the gap flag for the new path |
| `record/estoppel.go` | as above |
| `recordpb/record.proto`, `record.pb.go` | comment only |
| `record/recordsql/views.go`, `record/viewjson.go`, `view/changes.go`, `report/assemble.go` | **unchanged** — read `fix_new` for display; acceptance does not change its meaning |

**Census 3 — carriers of the FLAG (the new-verb set is moot; kept visible so the saving is checkable):**

Because the shape is a flag, these four mechanical gates do **not** fire, and each is named so a
future reviewer can see the decision rather than re-derive it: `cli/seat/help.go:175` (panics
without `help/<verb>.md`), `seatprobe/surfacecoverage_test.go`'s `TestEveryVerbHasABoardThatDemandsIt`,
`difftest/testdata/{help_contracts,role_boundaries_and_help_contracts}.golden`, and
`releasegate/fuzz/trajectory_test.go`'s `unreachedSurfaces()` over `cli.CommandPaths()` — whose only
escape is `exemptSurfaces` with a stated reason. **A new verb would have owed all four.**

What the flag does owe:

| carrier | disposition |
|---|---|
| `cli/seat/help/edit.md` | **changes** — the flag is documented on the verb that owns it; the *"ONLY path"* sentences stay TRUE and unedited |
| `flags/names.go` | **changes** — the `--accept` flag name |
| `recordpb/record.proto`, `record.pb.go`, schema golden | **change** — `BlueEdit.accepted`; regenerate with `go run ./protogen`, diff read |
| `recordpb/correspondence_test.go` | **verify** — a field addition, not a body/enum addition; the census count is unchanged |
| `releasegate/fuzz/fuzz_test.go` | **changes** — the harness drives `--accept` so `trajectory_test.go` still sees the path exercised; `verbsWithEvents` is unchanged because the event type is still `blue_edit` |
| `agents/blue-researcher.md` | **changes** — blue's constitution instructs repair against `required_fix` and names "closure-shopping" among dodge patterns; it must name the accept path or blue keeps retyping |
| `debate.js` blue prompt, `tests/simulator/prompts.test.mjs`, prompt goldens | **change** |
| `docs/seat-command-triggers.md`, `docs/record-flow.md` | **change** |

**Correction to the previous revision's census 1:** `record.proto:1117` was listed as a stale
comment. It is not — it says *"text taken unchanged from red's own `--fix-new`"*, which names no
comparison and stays true. The two genuinely stale sites are `verdict.go:17` and
`estoppel.go:38-41`.

## IV. Risk & Mitigation

| # | Risk | L | I | Mitigation |
|---|---|---|---|---|
| R1 | **Accepting becomes the path of least resistance and blue stops arguing** | med | high | Reachable only where red prescribed concrete text — red's own gated act. Accepting is the outcome red asked for, and the measured pathology is the reverse. `DeclineStats` already counts counter-edits as blue exercising disagreement, and `BlueEdit.accepted` (§III.A2) makes accept-vs-transcribed countable per run — which is why the field exists and why 'no schema change' was withdrawn rather than the claim quietly dropped |
| R2 | **The prescription is stale** — blue has edited that span since red wrote it | med | med | Refused: acceptance requires `location` to still match. Blue takes the ordinary path and the comparison backstop still applies |
| R3 | A doc comment keeps saying the flag comes from comparison | high | low | Three sites, all in census 1, all named. `estoppel.go:38-41` is the load-bearing one because it states the *"never claimed by either seat"* premise this plan preserves |
| R4 | The new verb trips a mechanical gate at build time | high | low | All three are in census 3 and two fail loudly (a panic and a test); the goldens are regenerated and diffed |
| R5 | `mint` can emit an empty `fix_new`, making the refusal unreachable or the path inapplicable | med | med | §V.3 asserts the refusal against a real empty-`fix_new` mint rather than assuming mint forbids it |

## V. Verification Plan

1. `cd plugins/frank-exchange-of-views/tools && go test -race -count=1 ./...`
2. **Non-vacuous gate.** `go test -run` exits 0 when nothing matches, and `-v` emits `=== RUN` per
   subtest, so counting lines is brittle. Use the JSON stream and count DISTINCT top-level tests:
   ```
   go test -count=1 -json -run 'TestAcceptAppliesTheRecordedFix|TestAcceptStillRequiresBluesReason|TestAcceptRefusedWithoutFixNew|TestAcceptRefusedOnStaleLocation|TestAcceptRefusedOnAnchorInSpan|TestAcceptIsIdempotentOnRetry|TestOrdinaryEditStillComparesBytes' ./internal/cli/... ./internal/record/ \
     | jq -r 'select(.Action=="run" and (.Test|test("/")|not)) | .Test' | sort -u | wc -l
   ```
   must equal **7**, checked in CI rather than read by a human.
3. `go test -count=1 -json -run 'TestAcceptRefusedWithoutFixNew' ./internal/cli/...` — the fixture
   mints with an empty `fix_new` through the real `merge mint`, so R5 is tested against what mint
   actually permits.
4. `go test -count=1 ./internal/seatprobe/ ./internal/difftest/` — the verb-board and help-contract
   gates; goldens regenerated and **the diff read**, confirming blue gains one verb and the `close`
   role-boundary refusal row is unchanged.
5. `node --test .../tests/simulator/{debate,prompts}.test.mjs` — blue's prompt names the path.
6. **Seat-probe, real agents.** Invocation per `docs/seat-surface-naming.md:12-13`:
   ```
   go run ./cmd/seatprobe -board all -bin <feov-record> -dir <out> -constitutions ../agents
   ```
   Three assertions on the blue-respond board: an accept sets `applied_verbatim` AND `accepted`
   and leaves the gap OPEN; a counter-edit sets neither; and a red board relitigating accepted text
   over 40 characters is refused by `EstoppelConflict` — unchanged behaviour, asserted so the
   narrowing did not disturb it.
7. **Replay, sized honestly.** `2026-08-23_research-loop-counterparts` carries **one**
   `applied_verbatim` event and **2 of 16** mints with a non-empty `fix_new`, so it cannot show how
   often acceptance would fire. What it CAN answer, and what §V.7 measures instead: for every
   `blue_edit` naming a gap, how close its `old`/`new` came to that gap's `location`/`fix_new`
   without matching — the near-miss population this path removes. Command:
   `go run ./cmd/feov-record show changes --run <extracted> --json` joined against the mint rows.
   A zero near-miss count is a real answer (the path saves transcription effort, not correctness)
   and must be reported as such rather than treated as a failed measurement.
8. `/plan-audit` PASS before implementation.
