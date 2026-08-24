# The record moves to SQLite

The storage layer, not the schema. `record.Append(Identity, proto.Message)` does not change
signature, so the ~500 fixture conversions already done on this branch stand.

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

## II. Decisions

1. **Driver: `modernc.org/sqlite`** — pure Go, no cgo, so cross-compilation and the release build
   stay as they are. Matches [[committed-tooling-is-go]].
2. **The schema is DERIVED FROM THE DESCRIPTORS AT DATABASE CREATION.** Not a committed `.sql`, not
   struct tags, not `protoc-gen-go-tags`. A DDL file is a second copy of the proto and would need a
   staleness gate; tags put the mapping on generated structs and pull in an ORM. Deriving it in
   process means the schema CANNOT drift from the schema, and there is no gate to write.
3. **No migrations, because a run directory IS the database.** Every run creates its own, so a
   schema change never meets an old database. This is the same ground `a12362c` stood on: a project
   in building mode whose every record is a test run.
4. **Class-table layout**: an `events` envelope table plus one table per body type, keyed on the
   event id. NOT a serialized body column — a blob would buy transactions and ordering while
   forfeiting joins, constraints and aggregate queries, which are the whole reason to move.
5. **`STRICT` tables, WAL, `foreign_keys=ON`.** SQLite's default affinity is loose enough to undo
   the point of typing the record at all.
6. **Append-only is enforced, not merely intended**: `BEFORE UPDATE`/`BEFORE DELETE` triggers that
   raise. The record is evidence in an adversarial process; nothing may rewrite it.
7. **Enums become `TEXT` with a `CHECK` over the schema's own spellings**; repeated fields become
   child tables so they stay joinable; oneofs become nullable columns with a `CHECK` that exactly
   one arm is present.

## III. Order of work

1. Derive the DDL from descriptors; test that every message and field is covered. ← here
2. The store: open, append in a transaction, read back.
3. Rewire `record.Append` and `BoardState` onto it; delete the shard/nonce/anomaly machinery.
4. Replace the hand-written projections with queries, one at a time, each with its test.

## The cutover (done — see status at the end of this file)

**The seam is `MergedEvents`.** 28 call sites inside `internal/record` take `Merged` and
never touch a file. If `MergedEvents` returns events from SQLite, all 28 are untouched.
That is the whole reason this is one change and not thirty.

### What each piece becomes

| Today | After | Why |
|---|---|---|
| `shardPath` + one JSONL file per (seat, nonce) | one `record.db` per run | The shard-per-seat layout existed to avoid write contention. WAL plus `busy_timeout` is the same guarantee, held by the storage rather than by the naming scheme. |
| pointer file + `withLock` + `durableWrite` | the nonce of the seat's latest register row | A pointer file is a record standing outside the record. It needed a lock because two registers race; a query does not. |
| `seq` from `len(ReadShard(shard))` | `count(*)` for (seat_id, nonce) inside the insert transaction | Read-then-write was two steps with a gap; now it is one statement under SQLite's writer serialization. |
| `nextStamp` monotonic clock file | wall clock, informational | Ordering is `events.id`. The clock file existed because filename order was not event order — the bug that dropped a whole sitting's bench closures. |
| `appendLine` torn-line healing | — | A torn line is a JSONL failure mode. A transaction either commits or does not. |
| `ReadShard` + `ClassifyLine` stages | `recordsql.Events` | Undecodable-line classification has no analogue: a row is a row. |

### Two channels that must NOT become plausible zeros

- **`Merged.Discarded`** tells apart a healthy re-dispatch from one seat id used for two
  sittings, where the losing shard held work that exists nowhere downstream. `capture.go`
  gates a run on it. In SQLite BOTH sittings' events are rows: nothing is discarded, so the
  loss is **unrepresentable**, not merely absent. DELETE the field and the gate, with the
  reason recorded — leaving it always-empty is exactly the shape this migration exists to
  remove, and it would read as "no loss detected" forever.
- **`Merged.Anomalies`** has two producers. The replay-time ones (torn line, undecodable
  row, missing gap) go away — the first two are unrepresentable, and `missingGap` is now a
  foreign key. The **projection-time** ones in `viewjson.go` stay and are unrelated. So the
  field survives, its replay-time producers do not, and `view.go:124`'s headline count
  changes meaning: say so in the change.

### Validation loop

Re-arms on any edit under `internal/record/`:

    (cd plugins/frank-exchange-of-views/tools && GOTOOLCHAIN=go1.25.0 go build -gcflags=-e ./... 2>&1 | grep -c '\.go:[0-9]')
    (cd plugins/frank-exchange-of-views/tools && GOTOOLCHAIN=go1.25.0 go test ./internal/record/... ./internal/cli/...)
    (cd plugins/frank-exchange-of-views/tools && UPDATE_GOLDENS=1 GOTOOLCHAIN=go1.25.0 go test ./internal/record/recordsql && git diff --stat)

Known-red and NOT a regression: `internal/difftest`, `internal/flags`
(`TestGradeValueSetRejections`), `internal/proof` — all three fail identically at HEAD,
verified by stashing. The remaining build failures are the in-flight test-fixture
conversion (capture, verify, view, fuzz, report, cli, record).

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

## Open decisions the cutover forces (operator's call)

**Re-verifying one source in one sitting is now refused.** A verify keys on its reference
(`url` is in `keyFields`), so two verifications of one source share a key. Under shards
both were written and the reader kept one — "idempotent, updates in place" was a read-time
illusion over an append-only log. `events.key` is UNIQUE now, so the second is refused.

The loss is real: a lens that re-reads a source mid-sitting and finds something different
cannot record the second reading. Three ways out, none obviously right:

1. **Leave it.** The refusal follows from append-only plus one-act-per-key, both
   deliberate, and it teaches. A seat that must revise says so in a later sitting.
2. **Drop `url` from `keyFields`** so verifies take an ordinal. Re-verification works, and
   a crash-retry of the same command now writes a SECOND event instead of being idempotent
   — which is what the key was for.
3. **Scope the key to the reading, not the source** (reference + outcome, say). Both cases
   work; the key stops being derivable from one field, which is how `keyFields` goes stale
   silently (record.go's own note on `reference`).

Pinned by `TestBoardCountsCiteEvents`, which asserts the refusal so the behaviour cannot
change without someone editing the assertion and reading this.

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

## Presence for repeated fields: measured, not reasoned

**The question.** A repeated field lands in a child table, so "does this gap supersede
anything" needs a join. Could an insert hook on the child set a boolean on the parent, so
presence comes for free with the row?

**The mechanism works.** An `AFTER INSERT` trigger on the child updating the parent does
what you would expect — measured: parent with a child reads 1, parent without reads 0.

**It cannot recover the distinction that bit us.** An explicitly-empty list and an absent
one are lost ABOVE SQL. Measured on the real messages:

    unset             Has=false  len=0  nil=true
    explicitly empty  Has=false  len=0  nil=false
    one entry         Has=true   len=1  nil=false
    wire bytes identical: true

`protoreflect` reports no presence for either, and the two marshal to the same bytes. By
the time a writer reaches the storage there is nothing left to tell apart, so no trigger
can recover it.

**Three costs, one of them fatal to the strongest use.**

1. It is a **second copy** of a fact the child table already holds — derivable by `EXISTS`,
   which is what `facts-are-fields` says to prefer generating over guarding.
2. It **cannot be used by a CHECK.** The parent row is written before any child, so a
   `CHECK (has_kids = 1)` fails at the moment the parent is inserted. Measured:
   `CHECK constraint failed: has_kids = 1`. That was the one thing a stored column could do
   that a view cannot, and it does not work.
3. It requires an **UPDATE on an append-only record.** Measured: the same guard that
   protects `events` refuses it (`constraint failed: append-only`).

**What was done instead.** The `gap` view carries `supersedes_count` and `found_by_count`.
Free, derived, cannot disagree with the rows it counts, and it answers the join several
readers were writing for themselves.

**Where the fact is a CLAIM rather than a consequence,** the schema already has the right
shape and it is a field the WRITER sets: `SpotCheck.none` — a bool with real presence,
meaning "I checked nothing, deliberately". Derived where derivable; declared where claimed.

## Where SQL earns its place, and where Go keeps it

The cutover is a good sample: nine defects, several constraints added, two of them wrong.
The pattern is sharp enough to state as a rule.

**Ask what the thing DOES, not where it could live.**

**Refusing a state that is unconditionally illegal → SQL.** This is where the whole move
paid. `gap_id` referencing `mint.gap_id` turned "a dangling reference is an ANOMALY
discovered on read, per reader, if any reader looks" into "the row cannot be written."
`events.key` UNIQUE turned a silent read-time dedup into a refusal the seat sees. The
transaction replaced torn-line healing. None of that is logic moved out of Go — it is the
DATA MODEL doing what Go was doing badly, and it holds against anything writing to the
file, which for a record that is EVIDENCE is the point.

**Refusing a state that is conditionally illegal → Go.** Both constraints that had to be
removed were this shape. `motion_rule.motion_id` referencing `motion.motion_id` is right
for two subjects and false for the third, because a direction motion has no motion row.
`Avenue.line` required is right for a proposal and wrong for a move. SQL cannot see the
condition: a CHECK holds no subquery, and a foreign key has no idea what subject the row
is. **A constraint that is wrong for a third of its cases is worse than none** — it
refuses correct work while reading as a guarantee.

**Explaining a refusal → Go, always.** `RAISE(ABORT, …)` takes a STATIC string. Every
refusal in this tool that earns its keep names the flag, the set, and what the omission
costs: "merge close: `carried` defers the gap instead of closing it, and deferring is the
BENCH decision". SQL can refuse; it cannot teach. Where a constraint and a message are both
wanted, the constraint goes in SQL as the wall and the message stays in Go as the door —
that is why `merge close`'s subset check exists twice on purpose.

**Computing a derived value → a VIEW, not a trigger.** Same language, opposite direction:
a view is SQL used as a QUERY, which is what it is good at. A trigger maintaining a
denormalised column is logic, and it buys nothing a view does not — see the presence
section above for the measurement.

**The one trigger that earns its place** is the append-only guard on `events`. It refuses
an unconditionally illegal state (editing a written event), and its static message is
adequate precisely because there is nothing conditional to explain: you cannot edit the
record, and that is the whole sentence.

### Presence for a list, if a verb ever needs it

The proto answer is a wrapper message (`optional Lineage supersedes` where `Lineage {
repeated string values = 1; }`) — message fields have always had presence, so set-but-empty
is distinguishable from absent. Confirmed: the wire bytes differ. `optional repeated` is a
syntax error in every proto version; the labels are one slot.

**We should not use the wrapper here.** It exists to work around proto's lack of
repeated-field presence at the LANGUAGE BINDING level — it buys `has_supersedes()` in every
language. This schema is DERIVED from descriptors with our own annotations, so we do not
have that constraint, and the wrapper costs us specifically: a message field becomes its own
table, so one list becomes TWO tables (a wrapper table whose only column is `event_id`, plus
the list table under it). Avoiding that needs a `presence_only` annotation AND generator
support to collapse the wrapper and re-key the list to the grandparent — real work to undo a
workaround we adopted for a limitation we do not have.

**The flat shape gets the same fact for nothing:**

    repeated string supersedes = 20;
    optional bool supersedes_stated = 21;   // "I considered lineage; there is none"

One column, no nesting, no second table, no generator change — and it is the idiom this
schema already uses well (`SpotCheck.none`).

**And it is the better model, not just the cheaper one.** A wrapper makes presence a
STRUCTURAL fact (is the message set); a sibling bool makes it a DECLARED fact (did the seat
say so). This record wants the second: a claim should be a field a writer can be REFUSED on.
An empty wrapper the CLI sets because a flag was registered is not a seat asserting
anything, and `--supersedes ""` should not become "I thought about lineage and there is
none" by accident.

**Not built.** No verb produces the distinction today, and inventing a record for a fact
nobody states is the cost facts-are-fields warns about. If a bench ever needs "red
considered lineage and found none" as distinct from "red did not address lineage", that is
the moment — and it is one field.

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


## The storage posture, measured (2026-08-23)

Four settings, each with the measurement that justifies it. Three were already right; two were
missing and one was actively expensive.

| setting | why | measured |
|---|---|---|
| `journal_mode(WAL)` | concurrent readers across processes | — |
| `busy_timeout(5000)` | SQLite resolves cross-process races internally before Go sees an error | — |
| `_txlock=immediate` | a deferred BEGIN upgrades read→write, and SQLite deliberately will NOT apply busy_timeout to an upgrade | 8 processes lost ~half their writes without it |
| `SetMaxOpenConns(1)` | **was missing.** Two goroutines on two pooled connections collide in SQLITE_BUSY; with one they queue on a Go mutex | — |
| schema in ONE transaction | **was 171 implicit transactions, each fsyncing** | **525ms → 39ms per fresh database** |

### The schema was the expensive one, and it cost nothing to fix

`Open` applied 171 DDL statements. SQLite autocommits any statement not already in a transaction
and every commit is an fsync, so creating a run directory paid 171 disk syncs — to write a schema
that is DERIVED and regenerated for free whenever `events` is absent.

Every alternative traded durability and was still slower:

    as-is (171 implicit transactions)   499ms   full durability
    synchronous=normal                   67ms   can drop recent commits
    synchronous=off                      46ms   corruption risk
    ONE transaction                      39ms   full durability, unchanged

One fsync instead of 171 is not a relaxed guarantee — it is the same guarantee, asked for once.
SQLite's own forum states the mechanism: an autocommitted statement fsyncs on its own, so
batching replaces one-fsync-per-statement with a single fsync at COMMIT.

### What is NOT there, stated so nobody assumes it is

**No retry, anywhere on the write path.** The only retry in the record layer is
`renameWithRetry`, a Windows rename workaround with LINEAR backoff and no jitter. Everything else
relies on `busy_timeout`, and SQLite's default busy handler is a FIXED schedule with no jitter —
`{1,2,5,10,15,20,25,25,25,50,50,100}`ms, 100ms thereafter — so contending waiters wake in
lockstep. Past five seconds it is a hard failure and the seat is told its write failed.

`SetMaxOpenConns(1)` removes the in-process half of that surface entirely, so what remains is
seat-vs-seat across processes, which the 8-process test exercises and currently passes. Adding
exponential backoff with full jitter (`avast/retry-go` has `FullJitterBackoffDelay`;
`cenkalti/backoff` is a port of Google's algorithm) is the standard answer IF that test starts
to show strain — but it is a dependency added on evidence, not on reasoning, and the evidence is
not there yet.

### One trade taken knowingly

`SetMaxOpenConns(1)` serializes the dashboard server's per-request renders, which run in
`http.Server`'s per-request goroutines. It is a local single-operator dashboard, not a
throughput-sensitive service, and the fix if it ever matters is a separate read-only handle for
the server rather than reverting the pool setting — the seat write path is what the setting
protects.

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
