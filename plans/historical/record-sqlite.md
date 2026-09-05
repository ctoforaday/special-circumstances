# The record moves to SQLite — archaeology

> Archaeology. The live design is [`../record-sqlite.md`](../record-sqlite.md), which holds the design as built and what is still open. This file is the record of what CHANGED: the JSONL storage this migration replaced, the defects the cutover surfaced and fixed, the decisions that were reversed, and the measurements that justified them. Past tense throughout. Nothing here describes the tree as it stands.

Superseded BY the SQLite record itself: `internal/record/recordsql` (landed on main via #556, `record-protobuf-pr2`), the views in `recordsql/views.go`, and the step-4 conversions (#700, #701, #703, then `e9c94e36`). The sections below are kept as they were written, at the dates they carry.

**What this document is.** The storage layer, not the schema — `record.Append(Identity,
proto.Message)` did not change signature, "so the ~500 fixture conversions already done on this
branch stand" (as the plan put it while the protobuf branch was in flight).

## I. Why, in one paragraph

The protobuf migration fixed the TYPING. It does not fix the three problems that come from
sharded append-only JSONL, and those are where this branch's bugs actually live:

- **Ordering.** Each seat writes its own shard and replay merges by timestamp, so a ruling can
  replay before its filing. `record/motion.go:182-189` says so in capitals, the same bug shipped
  once in the function `compat.go` existed to be the legacy twin of, and `record.Motions()` exists
  only to paper over it. A foreign key makes the state unwritable.
- **The join, re-implemented per reader.** Eight readers key a bench disposition on `gap_id`, which
  `motion-rule` does not carry. Eight hand-written joins, each able to drift. In SQL it is one.
- **Concurrency.** Nonces, torn lines, `ReadShard` anomalies — all of it is multiple writers
  appending to files. WAL and a transaction delete the category.

Two more fall out: `NOT NULL` is declarative rather than a hand-kept `required.go` table (which had
already omitted `Outcome.verdict`), and derived counts stop being Go loops that can read zero
forever (`filed > ruled` computing `0 > 0`).

> **As of 2026-09-05 none of that machinery exists.** `compat.go`, `shardPath`, `ReadShard`,
> `ClassifyLine`, `nextStamp` and `appendLine` are gone; `motion.go`'s capitalised ordering
> warning is gone; ordering is `events.id`. `record.Motions()` survives, but over a board
> replayed from an ordered record rather than as a repair for shard order.

## The cutover (done — see status at the end of this file)

**The seam is `MergedEvents`.** 28 call sites inside `internal/record` take `Merged` and
never touch a file. If `MergedEvents` returns events from SQLite, all 28 are untouched.
That is the whole reason this is one change and not thirty.

*(What each piece became, and the two channels that must not become plausible zeros, are in the
clean plan: they describe the storage as it stands.)*

### Validation loop, as it stood during the cutover

Re-arms on any edit under `internal/record/`:

    (cd plugins/frank-exchange-of-views/tools && GOTOOLCHAIN=go1.25.0 go build -gcflags=-e ./... 2>&1 | grep -c '\.go:[0-9]')
    (cd plugins/frank-exchange-of-views/tools && GOTOOLCHAIN=go1.25.0 go test ./internal/record/... ./internal/cli/...)
    (cd plugins/frank-exchange-of-views/tools && UPDATE_GOLDENS=1 GOTOOLCHAIN=go1.25.0 go test ./internal/record/recordsql && git diff --stat)

Known-red and NOT a regression: `internal/difftest`, `internal/flags`
(`TestGradeValueSetRejections`), `internal/proof` — all three fail identically at HEAD,
verified by stashing. The remaining build failures are the in-flight test-fixture
conversion (capture, verify, view, fuzz, report, cli, record).

> **The loop above no longer runs as written.** `go.mod` asks for `go 1.25.13`, so
> `GOTOOLCHAIN=go1.25.0` is refused before a test executes (measured 2026-09-05:
> `go: go.mod requires go >= 1.25.13`). The known-red list is also spent — `./internal/record/...`
> and `./internal/cli/...` were 10/10 green on 2026-09-05. The current loop is in the clean plan.

## What the cutover actually found (2026-08-22)

Nine production defects, none of which the pre-migration suite could have found. Recorded
here because the pattern matters more than the list: **every one was invisible because the
thing that would have caught it was itself broken, absent, or reading a fact out of the
wrong shape.**

| Defect | Why nothing caught it |
|---|---|
| `merge close` wrote `successor = ''` for an absent flag — every close in the tool failed once successor became a reference | Before the reference, the row was written and read as a closure whose successor was the empty gap. "Never said" and "said nothing" were the same bytes. |
| A re-dispatched seat could not record at all: the ordinal was per-sitting, `events.key` is global | Impossible under shards — the retry wrote the same keys into a NEW FILE and replay picked a winner. The storage change turned a tolerated duplicate into a refusal. |
| The sqlite driver's blank import was in `schema_test.go` | Every test had a driver; the binary had none. A blank import is invisible to the unused check. |
| 8 concurrent seat processes lost ~half their writes to `SQLITE_BUSY` | A hazard the storage change INTRODUCED. The first regression test used goroutines and passed with the fix reverted. |
| The debate view and the report printed `DISPOSITION_CLOSED` | Typing the enum made `%s` silently wrong — no type error, no test failure. |
| The fuzz prose gate was inert for nine event types (`reason` vs `text`/`rationale`/`basis`) | A payload map returned `""` for a wrong key and the rule was skipped. The gate reported coverage over rules that could not fire. |
| The coverage census grepped `Append(..., "type")`, a call shape that no longer exists | An empty set of ungated types reads exactly like full coverage. |
| `--as supports-with-bridge` was advertised in `--help` and refused by the write path | Nothing compared the advertised set against the schema. A comment called it a caveat "not mine to fix". |
| A required prose field was satisfied by `""` | The annotation collapsed two flavours of requiredness (present, present-and-non-empty) into one. |

Plus one gate reporting a **false** failure: the mass-parity regex matched `'low_medium'`
as `medium`, carried a wrong value, and reported two keys absent from a file that declares
both.

### The method that found them

Not the migration itself — the migration only made them *reachable*. What surfaced them:

1. **Driving the real binary**, not the library. Four of the nine only appear in
   `cmd/feov-record`.
2. **Making a miss LOUD.** `fieldStr` fails on a field name the schema does not carry;
   the enum census fails on a declared word the record cannot hold. Both found their
   defect within one run of being written.
3. **Reading generated output with eyes** — the golden schema, which found the arm tables
   with no foreign keys, and the rendered `--help`, which found the doubled `REQUIRED —`
   and cobra eating a backquoted value as the flag's placeholder.
4. **Asking whether a test can fail.** Two of my own passes produced assertions that
   could not: a Grade comparison against an `any`-typed string field (always true), and a
   blanket fixture rewrite that filled in the very fields four cases existed to omit.

## Two open decisions that have since been settled

Both were written 2026-08-22 as the operator's call, and both were answered the same day by
`b37232cb` ("a ruling states its rule, and a grade motion asks for a change"). Kept verbatim,
because each states the argument against the answer that was taken.

**A grade dispute may propose the grade already on the board, and is accepted.** Measured
2026-08-22: with `severity: high` on `R1-1`, `motion grade file --dimension severity
--proposed high` files cleanly. The motion then asks the merge to rule on a change that is
not a change, and a `rejected` ruling on it is indistinguishable from a rejection on the
merits — which is the shape the whole grade-dispute exchange exists to make legible.

Refusing it is a one-line check against the gap's current grade at that dimension. It is
NOT done here, deliberately: this is a judgement about what a dispute means, not about what
the record can hold, and a storage migration is the wrong place to narrow a verb's contract.
The counter-argument is real too — a seat re-affirming a contested grade at its current value
is a legible act, and refusing it would force that argument into prose.

**`--check ""` and friends.** Requiredness is present-and-non-empty again, with
`allow_empty` on the three fields the old Go table treated as presence-only
(`review_flag`, `principle`, `tension`). Whether `principle` and `tension` should be
tightened — a ruling with no stated rule is what `bench opinion` exists to prevent — is a
live question, deliberately not answered inside a storage migration.

**What was decided.** `requireGradeMotionAsksForAChange` (record.go) now refuses a grade motion
proposing the grade already on the board, and the refusal names the alternative: *"Propose the
grade you think it should be, or if you mean to argue the current grade is right, that belongs in
your closing rather than on the docket"*. It is pinned twice — `replay_test.go` on the write path
and `queries_parity_test.go` on the regrade overlay, the arm a naive query misses. `principle` lost
its `allow_empty` in the same commit, so the tightening question above is answered for that field;
`tension`, `review_flag` and a later `settled` still carry it.

## Status (2026-08-22)

**The cutover is done and driven.** The record is one SQLite database per run, verified
through `cmd/feov-record` rather than only through the library: register → record →
re-register → record, plus a once-per-sitting refusal and a full close.

**Green:** `record`, `recordsql`, `recordpb`, `capture`, `verify`, `view`, `report`,
`cli/bench`, `cli/merge`, `cli`. Production build at 0 errors.

**Known-red, and NOT a regression** — both fail identically before this work, on this
machine only, because `python` is not on PATH: `TestEveryProbeBoardStillBuilds`,
`TestEveryRequiredFlagIsActuallyRefused`. `internal/difftest`, `internal/flags` and
`internal/proof` were also red at the session's start (verified by stashing); difftest's
goldens will need regenerating against the new envelope, which is a deliberate act and not
this change's to make silently.

### Still owed

- **The fuzz coverage gate was right about every zero it reported, and the causes were
  all different.** Resolved: five drives named commands that do not exist (`spot_check` and
  `manifest_row` name the EVENT TYPE, whose word is underscored, where the command is
  hyphenated; three named `avenue` where the command is `line-of-inquiry move`).
  `certify` was refused at the flag layer, because `satisfiedByAnyOf` was keyed on the
  payload field (`statement`) rather than the flag, so `--reason` did not satisfy it.
  `close` was the absent-flag/`successor` bug. Tally after: spot_check 0 -> 54,
  manifest_row 0 -> 24, certify 0 -> 5, close 0 -> 3, avenue 8 -> 23, mint 6 -> 12.
- **`observe` has no verb at all.** The event type is in the schema and nothing in the
  command tree writes it — the only Appends of a `recordpb.Observe` are in record's own
  tests. Exempted with that stated, rather than driven (a drive would have to invent a verb)
  and rather than dropped from the list (the type's homelessness should be a line somebody
  reads). The `Observation` a board carries is built from FINDING events.
- **`motion_appeal` is unreachable, not broken.** Driven by hand the full cycle works —
  file, rule, appeal. Blue can only appeal a ruling from a PRIOR round, and 46 of 60 runs
  are single-round, so the path is starved by the scenario distribution rather than by a bad
  drive. Changing that distribution is a fuzz-design decision, not a bug fix.

  **Do not read a low tally as a quiet run** — that is the exact mistake the gate's own
  comment records from the last time, and every zero above had a real cause.
- **The re-verification decision** (above) is the operator's.

## What the second cutover pass found (2026-08-22)

Fifteen defects, all in production code, none findable before the migration made the
schema the authority. They fall into two families that keep recurring, so they are grouped
by CAUSE rather than by the order they surfaced.

### Family one: one fact, two spellings, one side moved

Six instances. Every one reads perfectly well and every one refuses a seat that obeys it.

| Where | The two spellings |
|---|---|
| `merge inquiry-support` help | advertised `--id Q1 --as supported\|weakened\|unsupported\|absent`; the schema collapsed the event to prose and the flags were removed, the sentence was not |
| fuzz `rulingFor` | returned `out-of-scope` / `too-thin`; DirectionRuling spells `out_of_scope` / `too_thin` |
| `validate`'s Retire arm | said `retire requires --claim`; the annotation declares the flag `--quote` |
| `internal/cli/motion` | typed the gavel as `subject("petition", …, "bench")`; the PASS gate in `internal/record` needed it and could not see it |
| `interpreterFor` | mapped `.py` to `python`; Debian, Ubuntu and macOS 12.3+ ship only `python3` |
| `--as supports-with-bridge` | advertised in help, refused by the write path (found in the first pass) |

The fuzz one is the instructive case: `motion inquiry rule` was refused on 17 of 27
invocations, the drive discarded the error, and the only thing that noticed was a coverage
gate reporting a MISSING drive for `motion inquiry appeal` — whose drive was correct and
whose precondition never happened. **A discarded refusal turns a broken write into a
coverage report about something else.**

Two gates now hold this class, and they must both exist because they see different messages:

- `TestNoRecordRefusalNamesAFlagASeatCannotType` (internal/record) drives an empty body of
  every event type plus the contract table. Its first draft ran the table alone and covered
  fourteen types — measured, `closing requires --id` could be renamed to a flag that does
  not exist and it stayed green.
- `TestNoRefusalNamesAFlagThatDoesNotExist` + the check folded into
  `TestEveryRequiredFlagIsActuallyRefused` (internal/cli) cover what cobra renders. The
  record layer's own messages are unreachable from the command line, because cobra marks
  the same fields required and refuses first.

### Family two: a rule stated twice, where the unreachable copy is the one that drifts

`recordpb.CheckRequired` runs BEFORE validate's type switch and refuses unconditionally.
Eleven hand-written guards restated a `required` annotation below it. Ten were harmless
dead code. **The eleventh was not a restatement**: `close` exempts a carry from stating the
closure argument, `prose` was annotated required anyway, and so `merge carry --id R2-3
--carried-from 2` — the invocation the verb's own help documents — was refused outright.

The general rule, which is the part worth carrying forward: **anything an annotation makes
unconditional is decided before a single line of the verb's own validation runs, so a
`required` marking silently deletes every exemption below it.** Conditional requirements
belong on the field as a comment and in validate as code; two-field rules belong in a
table-level `(check)`, which is where the carry rule went so the database keeps the wall
the nullable column gave up.

### The wedge: a round with no legal verdict

Not a spelling problem, and the most serious thing here. `verdict --as PASS` is refused
over an unruled motion; the refusal said "rule it"; for a PETITION that is refused in turn
by `requireRuler`, because the bench holds that gavel. With a clean gap board, debate.js
also rejects a FAIL that names no gaps. **There was no verdict the seat could give.**

Fixed by putting `(ruled_by)` on the MotionSubject enum — one declaration, two readers,
where before there was one declaration and one reader that needed it and could not import
it — and by making the rule-it instruction conditional on the gavel being yours, with FAIL
named as the act that ends the round for a seat that cannot rule.

Surfaced only because the fuzz stopped discarding the PASS refusal, and because the
per-round inquiry review moved off a 40% coin. It is a per-round DUTY and a hard
precondition for PASS: three rounds in five could not pass, and the drive told the harness
`PASS` anyway, so the run recorded `outcome: verified` with no verdict event and basis
`asserted`. **What varies between runs is what a review finds, never whether it happened.**

### One report defect

A minted finding's evidence had no reader. `surfaced by: L5-F1` and nothing in the document
defines that label — `unmintedFindings` renders a finding's text only when NO gap claims
it, so the instant the merge acts on a finding its leaf-level evidence leaves the report.
Runs exist where every finding was minted and red's words appeared nowhere at all. The gap's
`problem` is the merge's RESTATEMENT; both must be in front of a reader or a restatement
drifting from its evidence is invisible, which is the adversarial point of having two seats.

### Method note

Three candidate gates were written and one was DELETED for being unable to fail. A test
that asserts a hand-written arm cannot shadow an annotation can never fail, because the
annotation sweep always wins — the residue is dead code, not wrong behaviour. Each gate here
was mutation-tested against the defect it was written for before being kept; the flag-word
gate was broadened only after a mutation proved the narrow version green.

### The one that was not a spelling: no granted petition could be recorded

Found by chasing a fuzz refusal rate rather than a failure. The execution tally read
`motion petition file=40(22 refused)` and `motion petition rule=18(8 refused)`, and the
drives discarded both errors, so the only visible symptom was a scenario oracle three layers
away reporting a missing event.

`RulingBinds` carried `all | filer | none`. The flag's help, the table it renders from, the
report's prose and debate.js's `PETITION_RULING` schema all say `blue | red | both`. Nothing
overlapped — and `--binds` is set exactly when a petition is GRANTED. **The constitutional
short-circuit's grant path could not be recorded at all.** `PetitionClass` was the same
defect half-landed: `integrity | safety | process | scope` against an advertised
`ethical | safety | integrity | constitutional`, so two classes worked and two were refused
for words the seat had just read in `--help`.

The engine was right and the schema was wrong — every other carrier agreed with every other
carrier, and no Go code referenced `process`, `scope`, `all`, `filer` or `none`. The enums
move to the vocabulary the system speaks, retired numbers RESERVED rather than renumbered:
nothing could write a 3 or a 4, but a number back in service reinterprets any record that
did, and that rule does not bend for values one is confident are unused.

**The test that should have caught it asserted the opposite.**
`TestTheAdjudicationVocabulariesHaveExactlyOneSourceEach` said *"After #344 there is ONE
table … the drift is not detected — it is unrepresentable"*, and it was counting the two
tables it knew about while the write path resolved against a third. A test that names its
own completeness is only as good as its census. `MotionFields` derives from the descriptors
now, and the test checks both directions: nothing advertised that cannot be written, and
nothing in the schema that no surface can produce — the arm that would have found `process`
and `scope` sitting unreachable.

### The check nobody ran, and what it was hiding

The simulator (`tests/simulator`, `node --test`) sits on the plan's "owed before PR2 closes"
list, so it had not run once during the migration. Running it produced two more instances of
the same family — and both were **green by absence**, which is worse than red, because an
unrun gate reads as coverage:

- `debate.test.mjs` asserted the grade enum as `low-medium | medium-high`. `debate.js` and the
  tool both spell them with underscores. The engine and the tool agreed with each other; the
  test between them did not.
- `prompt-frontier.golden` and `prompt-red-merge-r1.golden` recorded prompts telling seats to
  type `too-thin` and `out-of-scope`. The prompts themselves say `too_thin` and `out_of_scope`.
  The golden that exists to catch exactly this drift was the thing carrying it.

**A gate's value is its last run, not its existence.** Six months of "we have a golden for that"
is worth nothing if the golden has not been compared. Where a check is expensive enough to end
up on an "owed" list, that is the check most likely to be stale when it finally runs.

### Standing pattern, stated once

Eight instances now, and six of the defects below them were found the same way: **a refusal
somebody discarded**. `_, _ = r.exec(...)` in the fuzz, five times; a help string that
outlived its flags; an annotation that pre-empted an exemption. The refusal was always
correct and always right there. What made each invisible was that nothing kept it.

Where a drive discards an error, the failure surfaces somewhere else entirely — as a
coverage report about a path whose drive was fine (`motion inquiry appeal`), as a scenario
oracle naming an absent event (`R1-1 has no dispute event`), as a derivation blamed for an
`asserted` verdict it was never given the chance to derive. **The distance between the
discarded refusal and the reported symptom is the whole cost.**


## The merge with `origin/main` (2026-08-23)

147 commits ahead, 110 conflicts, and **almost none of the risk was in them**. Six things merged
CLEANLY and were wrong. Every one is the same shape: two halves of one change arriving from
opposite sides, each half correct beside its own partner. `git` cannot see that class at all.

| what merged clean | what it did |
|---|---|
| the hand-written `(check)` expr | compared `closure_class` against `'closed_with_regression'` — a word no writer can now produce, so the expression is always true and "a regression closure must name its successor" never fires again |
| the gavel | main dropped `requireRuler` because its constructor scoped the verb; our constructor added the verb unconditionally. Together: **no gavel at all** — blue ruled its own motion and was told it succeeded |
| three coverage gates | walked `newRoot()` and asked `isSeatRole(path[0])`. Against a scoped surface no path qualifies, so they traversed NOTHING and reported success. Their floors (`< 20 flags`, `< 10 verbs`) are the only reason this surfaced |
| the class registry | main's strict arm landed beside our test asserting the opposite ("no registry staged is advisory") |
| the fuzz's PASS path | `r.exec("merge", "verdict", …)` — a path the scoped surface no longer has. The refusal reads as "refused over something that is not a gap", which is a REAL production case, so the drive recorded FAIL with an empty gaps array and the engine rejected the round. 21 of 60 seeds |
| the golden harness | `cmd{role: "merge", args: {"show", …}}` with no `--seat-id`. Every scenario silently lost its RENDERS and REPORT halves |

**The two worth remembering fail silently BY DESIGN.** The fuzz reads a refusal as a legitimate
FAIL; the golden harness appends only when the command succeeds, because a degenerate scenario
really does render nothing. In both, total failure and the honest empty case are the same bytes —
[[facts-are-fields]] clause 3, in a harness rather than in a parser. The golden one would have been
committed: regenerating first produced **2278 deletions against 323 insertions** across 20 files,
with one golden losing 113 lines and gaining zero. A vocabulary rename does not delete two thousand
lines, and that ratio is the only thing that prompted the look.

### The vocabulary, resolved

Main's `a1e8e26` renamed the closure VALUES; this branch renamed the TYPE. Orthogonal, and both
land: **main's words, our type name, our annotation, main's derived axis.** Main's numbering is
preserved exactly, so `carried` keeps slot 7 — a slot main has no value for, because it carries
deferral on a separate field. `ArtifactStateOf` is absorbed minus its retired-class arm: under
no-back-compat a retired class is unwritable, because the CHECK is generated from the same enum.

Values take the schema's spelling (`too_thin`, `defect_accepted`, `low_medium`); flags take dashes.
The enum refusal now names BOTH near-miss classes — it detected case and said nothing for
separators, which is the mistake those two conventions invite on one command line.

### The defect the merge surfaced, which ships and predates it

`recordsql.Open` built `"file:" + path + "?_txlock=…"`. The `file:` prefix makes that a URI, so a
run directory containing `#`, `?` or `%` was TRUNCATED at that character — silently, to a path that
still opens. Two runs then share one database, or a run opens one that is not its own, and every
write reports success. Reachable in production: run directories are named from the topic, and
"C# concurrency" produces exactly this path.

Found by accident. Two identically-named Go subtests get `#01` appended, `t.TempDir` builds from the
test name, and the second opened the FIRST one's database — surfacing three verbs away as "this
seat has already recorded a mint this sitting" in a run that had recorded nothing.

Fixed with `net/url`, **not** a hand-rolled escape. The hand-rolled version was the first attempt and
is instructive: it got the three reserved characters and could not answer the `//`-authority case at
all, because `/` is not escapable without changing the path. I had written that limitation into the
test as "not measured" and shipped it — a stated limit that was really the fix being wrong. The
driver has no DSN builder and neither does `mattn/go-sqlite3`; the format is a cross-driver string
convention with the escaping obligation pushed onto every caller.

### Prompts carry the JOB; the tool carries the contract

Main's workstream had one goal: move instructions out of prose and into the CLI, so a subagent
discovers paths from the tool rather than from a description of it. **The help is the instruction;
prose about the tool is stutter, and its partial information produces satisficing.**

`promptCatalogue` pinned at 0 is that rule's mechanism, not a lint. Red's merge prompt went to
main's — 20315 characters against a 13000 ceiling, 56% over, while every other seat sat under — and
what left was contract: the near-match rule naming the screening verb, `citations_checked` naming a
projection, `found_by` naming a mint field. `SKILL.md` had grown to 5 named commands on THIS branch
while every other surface was going to 0, and is now pinned explicitly so the carrier is covered by
a decision rather than by an untracked file's default.

### Verified

tools `./...` 35 packages 0 failures (fuzz included, 60/60) · simulator 94/94 · scripts 9/9 ·
difftest and prompt goldens re-recorded AND read · build, vet, `qlty` clean. The `go` directive
moved to 1.25.14, which is where 46 stdlib CVEs were; the suite was re-run under it rather than
assumed, because the goldens pin error text that can originate in stdlib. No golden moved.

## Two defects Windows CI found that Linux structurally cannot (2026-08-24)

The first CI run this branch ever had. Both are real, both predate the PR, and neither is
observable on Linux — which is the reason they survived a green suite for the whole migration.

### `Open` composed a URI and did not escape the path

`"file:" + path + "?_txlock=…"` is a URI, and `?` and `#` are reserved in one. A run directory
containing either was TRUNCATED there, silently, to a path that still opens: two runs share one
database, or a run opens one that is not its own, and every write reports success. Run directories
are named from the topic, so "C# concurrency" produces exactly this path.

Fixed by building the DSN with `net/url`. The instructive part is the FIRST fix, a three-character
`strings.Replacer` — a hand-kept list of what the grammar reserves, written while fixing an
instance of exactly that class. It also could not answer the case percent-encoding cannot: a path
beginning `//` parses as an authority and `/` is not escapable. I had written that limitation into
the test as a "not measured" comment and treated writing it down as discharging it.

Then `net/url` broke Windows on its own: `C:\Users\x` has no leading slash, so `String()` writes
`file://` and puts `C:` where the AUTHORITY goes — `invalid uri authority`, 60 of 60 fuzz runs,
zero events, round 0. SQLite documents the form: `file:///C:/Users/x`. `dsnFor` is a PURE FUNCTION,
so the platform it is wrong on is not the platform the test has to run on — the rooting is asserted
directly, and the comment that stood in for it is gone.

### The handle cache had no release

`Open` caches a `*sql.DB` per database and nothing ever closed one. Correct in production — a seat
is a process per command, and the dashboard wants the handle held — and invisible on Linux, where
an open file can still be unlinked. On Windows `t.TempDir` cleanup cannot remove an open file, so
ten packages failed at once on a line that has nothing to do with what they assert.

The cache moved from `record` down into `recordsql`, which is what made it releasable: `recordtest`
cannot import `record`, because `record`'s own tests import `recordtest`. Owned by the thing that
opens the database, both caching and release reach every caller with no cycle.

Two invariants a cache imposes, both found by the suite rather than by reasoning:

- **No caller may close.** `recordtest.Seed`'s `defer db.Close()` stopped releasing a private
  connection and started poisoning the cache — the next `Open` returns the same closed handle.
- **A fixture releases the run it CREATED.** A global `CloseAll` from a per-test cleanup takes the
  parent's and the siblings' handles with it.

### `Open` races on schema creation

**FIXED 2026-09-02:** commit 0fa2ce93 ("Fix #557: decide the schema INSIDE the transaction that would create it", merged via #632) closes the check-then-create gap; `recordsql/openrace_test.go` pins it, and `store.go` records why `CREATE TABLE IF NOT EXISTS` was rejected. The paragraphs below describe the defect as found.

Removing the cache to measure its cost exposed it. Two openers of a FRESH database each see no
`events` table and each apply the schema; the second fails with "table already exists". `Open` does
check-then-create with nothing between.

Unexercised today for two separate reasons, which is why it has never been seen: the in-process
cache means one opener per database per process, and the cross-process concurrency test SEEDS the
database before spawning its eight children, so the schema exists by the time they open. Production
is protected by the same accident — `setup` creates the run before any seat is dispatched.

NOT fixed here. `CREATE TABLE IF NOT EXISTS` and a transaction-with-retry differ in what they do to
a HALF-created schema, and that is the decision, not the mechanism.

## Step 4's next list, landed (2026-09-05)

The 2026-09-04 status closed with a named next step: *"the named questions that landed as SQL
strings inside Go functions move into ViewsDDL as views (stranded ancestors, gaps awaiting proof,
motion state with the row semantics the guards need)"*. That landed in `e9c94e36` ("Answer whole
question families in views, not per-function SQL"): the `gap` view grew `current_*` grades,
`proof_answered`, `awaiting_proof`, `superseded_by`, `stranded` and `minted_event`, and two views
were added — `motion_answers` (one row per answered motion, first-wins stated once) and
`line_of_inquiry`.

The same commit superseded one sentence of that status: *"The motion guards read BASE tables, not
`motion_state`"* was true when written and is not now — `motion.go` reads `motion_answers`, and
`motion_state` reads it too rather than carrying its own rule join, which multiplied rows on a
record holding an illegal second ruling.
