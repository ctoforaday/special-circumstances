# Command groups: the seat's verb set IS the tree

> The design as built. The entity-group tree this plan originally proposed, its staged
> execution, the reversed dual-read and the measured defects that justified the work are in
> `plans/historical/command-groups.md`.

Agreed in design dialogue 2026-08-09 and shipped in seven stages, which live code still cites by
number — **1** #341 `cite` split · **2** #342 closure vocabulary · **3** #343 `reproduce` records
its verdict · **4** #344 the `motion` group · **5** #348 identity as fields · **6** #345 the
scoped tree (landed as #480) · **7** #346 the agent-facing surfaces — with #290 supplying the
detected identity underneath. This document states what binds the surface now;
the archaeology document states what changed and why, including the parts of the original
design that were never built.

> **This AMENDS `plans/feov-cli-architecture.md` (#59, Cut 1), which reads "role belongs in
> the tree — the role is the verb's parent node."** Cut 1 was about role as *constructor
> data*: a `role string` passed to every verb was a command being told its own position.
> That argument stands and is kept — no constructor takes a role here either.
>
> What changed is where position is read. The role node came off the command line entirely and
> the seat is **detected, not typed**: the PreToolUse hook injects it exactly as it already
> injects `FEOV_RUN`, and a seat sees only its own verbs. This is the mechanism half of **#290**
> ("seat identity is self-asserted — … and identity-scoped command surfaces").

## I. The tree

`feov-record` builds ONE tree per caller, from the identity it was dispatched with
(`internal/cli/root.go`, `NewRootFor`). A seat's verbs are mounted at the **root of its own
tree** (`seat.RoleVerbs`) — `mint`, not `merge mint`; `cite`, not `blue cite`.

**The verb set IS the role boundary.** A lens structurally cannot mint or close a board gap:
no such verb exists in its namespace. Blue has no board verbs at all. That is a property of the
tool rather than a rule someone has to follow, and it is why no `SeatPermissions` table was
built: a table consulted at the write would be a SECOND answer to "which seat may do this",
beside the tree that already answers it.

Four nodes across the seat surfaces hold sub-verbs, and each earned it.

| group | shape | seats | why it is a group |
|---|---|---|---|
| `motion` | `motion <subject> <verb>` | all | the subjects' contracts diverge in required flags, ruler and escalation — a flag cannot express it without the tool accepting nonsense |
| `show` | `show <projection>` | all | eleven projections that were flag VALUES behind `--view`, invisible to `--help` until you already knew the vocabulary |
| `line-of-inquiry` | `propose` · `move` | blue | proposing a line and moving its status are two contracts on one entity's lifecycle |
| `class` | `new` | merge | the gap-class registry — **it was four flags on `mint`**, and that is what a verb looks like when it is wearing another verb's clothes |

The operator tree has its own two: `show diagnostics|tiers` and `ocr pages|read`.

**Where a flag discerns a TYPE today, that flag is a subgroup candidate tomorrow** — `--view`
was the clearest case in the tool and became `show <projection>`. `class` is the same rule from
the other end: `--class-new` was a boolean whose whole meaning was "I also passed
`--definition`, `--neighbor` and `--distinguisher`", the three were optional to cobra and
policed in the mint handler, and a mint that coined a class wrote TWO events, one of which had
nothing to do with putting a gap on the board. **A cluster of flags that must travel together,
changes what else is required, and writes its own event, is a verb.** As one, cobra refuses an
incomplete coining at parse, and `--help` can say what a class IS without that competing for
room with what a gap is.

**A group earns its place when it has two or more verbs AND its name disambiguates or
teaches.** Everything else is a top-level verb. **Counter-pressure, and it is real: depth costs
tokens on every call** — the speed clause tells every seat a message costs ~20s regardless of
content. One group is worth it where it deletes hand-validation; a tier applied for tidiness is
a tax on every invocation.

### The surfaces

Verified in `internal/cli/{blue,lens,merge,bench}/command.go`. `show` is appended to every
seat by `seat.RoleVerbs`; `motion`, `fetch` and `count-claims` are added in `NewRootFor`.

| seat | verbs |
|---|---|
| **blue** | `register` `edit` `cite` `prove` `revision` `retire` `line-of-inquiry {propose,move}` `manifest-row` `claim-index` `position` `closing` `friction` · `show` `motion` `fetch` `count-claims` |
| **lens** | `register` `finding` `verify` `corroborate` `reproduce` `friction` · `show` `motion` `fetch` `count-claims` |
| **merge** | `register` `mint` `class {new}` `close` `carry` `regrade` `spot-check` `inquiry-support` `near-match` `position` `closing` `verdict` `friction` · `show` `motion` `fetch` `count-claims` |
| **bench** | `register` `opinion` `halt` `certify` `declare` `outcome` `assemble` `friction` · `show` `motion` `fetch` `count-claims` |

`show` projections (`internal/cli/seat/verbs.go`, `views`): `report` `board` `findings` `work`
`motions` `debate` `changes` `evidence` `lines-of-inquiry` `telemetry` `scorecard`.

**A SEAT IS NOT AN OPERATOR, and the two sets are disjoint by construction.** `--seat-id
operator` selects `verify` `graph` `show {diagnostics,tiers}` `count-claims` `friction` `fetch`
`ocr {pages,read}` `setup` `scorecard` `dashboard` `capture`. Two names collide across the split:
`friction` is a seat's write verb AND the operator's read of the channel; `verify` is a lens's
citation adjudication AND the operator's whole-record cross-check. One tree cannot hold both
meanings, and picking a winner would hand some seat a verb that does the other thing. `fetch`
and `count-claims` cross the split because seats genuinely run them — a lens reads the EXACT
bytes blue cited, from the run cache, so it audits the same artifact rather than a page that
may have drifted; blue's `claim_count` is defined as what `count-claims` prints — and neither
name collides.

**NO IDENTITY IS NOT A MODE.** An unidentified caller used to fall through to the operator
surface, which made "nobody said who I am" a way of selecting a tree — and a tree nobody
selects is one nobody can be refused from. There is nothing to show until the question is
answered; `--seat-id operator` is how an operator answers it.

**Rejected: a binary per seat.** It restores the structural boundary, but the prompt must then
name `feov-lens` — putting the seat back on the command line, which is the thing being removed
— or the hook routes between four binaries, which is more magic rather than less. It also
multiplies the version surface `setup` preflights against the event-schema epoch.

## II. Identity: detected, cross-checked, and stamped as fields

**The seat is DETECTED.** The PreToolUse hook exports the harness agent handle; `register`
binds it to a seat as a field on its own event, and identity is read back from the record.

**`--seat-id` and `--run` SURVIVE as CROSS-CHECKS, not as sources.** A passed value that
DISAGREES with the resolved one is REFUSED. `seatenv` is the precedent and the reason: its
point was never that the environment is a second source, it is that a `--run` disagreeing with
the injected value is refused. That exists because a seat typed `special circumstances` for
`special-circumstances`, the tool answered "names gap R1-2, which no mint event created", and
the seat believed it, filed a false bug and abandoned five manifest receipts. Detected identity
with no cross-check would make attribution a derived fact that fails silently.

**Identity arrives as FIELDS (#348).** Role and round are stamped at the append path rather
than recovered from the seat id:

- `record.PartyOf(e)` reads the stamped `role`. It falls back to `roleOfSeat(e.SeatID)` ONLY
  for records written before the field existed — a real corpus this tool still reads — and the
  fallback is guarded on the stamped value being USABLE, not merely present, because `role` is
  optional in the schema and so has three states where it had two (stamped, stamped-empty,
  absent). Before the field, `strings.HasPrefix(e.SeatID, "red-merge")` decided whether a
  position rendered as RED or BLUE, so an id that failed to match its expected prefix rendered
  as the wrong party with nothing to notice.
- `record.RoundOf` survives, and its second return is the whole point: it answers `(round,
  known)`, where a bare `0` used to mean both "round 0, which is synthesis and a real round
  with real events" and "this name says nothing about a round". That collapse produced the
  phantom-archive bug (#327) — `judge-terminal` carries no round, so it yielded 0, so a
  terminal bench closure looked like a closure before round 1. `internal/consistency` now uses
  it as a CROSS-CHECK against the stamped field rather than as the source.

**The seat id stays the SHARD KEY and the concurrency namespace.** Only the DERIVED facts
became fields. A lens index recovered from a seat name is what makes a lock-free counter safe
under parallel dispatch; collapsing it once made 39 of 60 disposals ambiguous. `record.RoleOf`
(in `findinglabel.go`) is that lens-index reader and is deliberately NOT the name given to
`PartyOf` — two different things were about to share one name.

## III. The refusal carries what the surface cannot

**The boundary went from legible to invisible, and that cost is paid rather than hidden.** A
seat knew its role because it typed it. Scoped, an out-of-role verb returns "unknown command" —
indistinguishable from the capability not existing, and a seat handed an unavailable verb logs
friction and works around it, losing the capability for the run.

The surface cannot carry the verb (it does not know which seat's), so the REFUSAL carries the
reason. `unknownCommandRefusal` (`internal/cli/root.go`) names the owner, and `whereItLives`
searches every role's tree to find it. **Help states the identity and lists only that seat's
verbs**, and says nothing about other seats: the earlier draft's polymorphic help is hard to
phrase without inviting the reader to try the thing it just named, and it puts noise in every
invocation to solve a problem that occurs once.

**MEASURED 2026-08-17.** A red-merge seat holding a work-list duty that named `inquiry-support`
typed `feov-record inquiry-support --help`, was told no such command exists, believed it —
reasonably, the message is unambiguous — went hunting, and settled on `motion inquiry rule`,
which is a DIFFERENT ACT: whether the direction is worth the run's time, not whether the report
still carries it. It did not lose a turn; it landed on the wrong verb with confidence.

**An unknown command PRINTS THE SURFACE.** Cobra's default answers `unknown command "x"` and
stops — at the one moment a seat is definitively looking for what exists, which was the one
moment the tool said least. The root takes `ArbitraryArgs` so the miss routes into `RunE`, help
goes out first, and the refusal follows it. The same argument makes `show` teach on an unknown
projection, and makes a `motion` subject that lacks a verb say which verbs it has.

**ABSENT-BECAUSE-NOT-YOURS IS NOT ABSENT-BY-DESIGN**, and the seat must be told which it met.
`motion <subject> rule` is added only to the gavel-holder's tree, and the subject group's `RunE`
answers a non-holder by naming BOTH parties — the seat that holds the gavel and the one that
asked. Either alone sends a seat to the wrong fix: "the bench rules this" without saying who
you are reads as advice, and "you are merge" without saying who does reads as a dead end.

**The friction footer hangs on the seat root's `Long`**, so a seat is told that a capability it
cannot find is a finding about the tooling rather than something to work around.

## IV. `motion` — one adjudication mechanism, subgrouped by subject

The propose→rule exchange was implemented three times with three vocabularies — `ruling` /
`ruling` / **`response`** for one concept, three renderers, and no shared id. `internal/cli/motion`
and `internal/record/motion.go` are the one mechanism, and **a motion has an ID, so the ask and
its answer join on it** (closing #312).

| subject | file | rule | required at file | ruler | appeal |
|---|---|---|---|---|---|
| `grade` | yes | yes | `--id` `--dimension` `--proposed` | **merge** | yes |
| `petition` | yes | yes | `--class` `--relief` | **bench** | **no** |
| `inquiry` | **no** | yes | — | **merge** | yes |

**The subjects are SUBGROUPS and not a `--on` flag.** Their contracts diverge, and cobra has no
"required only when `--on=grade`" — `MarkFlagsRequiredTogether` and `MarkFlagsMutuallyExclusive`
do not express conditional requirement. A flag-discerned subject would put three contracts into
hand-written `RunE` validation, which is the prose-standing-in-for-a-schema shape this suite
exists to remove. As subgroups, cobra refuses the nonsense at parse.

**This costs the collapse nothing.** The shared mechanism is the EVENT AND THE ID, not the path:
all three write one motion event into one id space, joined once and rendered once.

**Absence is a design statement, expressed by absence.** `petition` has no `appeal` — a petition
is heard BEFORE the debate continues, so there is nothing to escalate to. `inquiry` has no
`file` — the proposal IS the filing (`blue line-of-inquiry`), so a filing verb here would be a
second way to say what blue already said. Each is stated where a seat actually meets it, in the
subject group's refusal.

**IT IS NOT CALLED `docket`.** `docket` was already live and load-bearing with a DIFFERENT
referent — the bench's contested-gap adjudication list, in `debate.js`, `lead-judge.md`
("a docket you can dispose of by carrying it is a docket you have failed"), `red-auditor.md`
("docket-bound"), `record.go`, `verify.go` and `assemble.go`. `motion` is free and it is the
right word: a party FILES a motion and the ruler decides it. The verb is `file`, not `submit`,
for the same reason.

**The verdict flag is `--as` and the payload key is `ruling`.** This is not cosmetic: a collapse
that did not pick one spelling would reproduce inside the new group the exact defect it was
built to remove. `--as` is the repo's settled convention for a verdict (`close`, `opinion`,
`verdict`, `outcome`, `reproduce` all use it). `flags.Petitioner` and `flags.Ruling` are gone:
under a motion id you rule with `--id` and the filer is on the record, and a `--petitioner` flag
beside a motion id would be the second join key #312 is about.

**THE GAVEL IS NOT TYPED IN THE COMMAND TREE.** It was — `subject("petition", …, "bench")` — and
the PASS gate in `internal/record`, which cannot import the cli package, told blocked seats to
rule motions without knowing whose ruling it would be. Both readers now take it off the
`MotionSubject` enum (`recordpb.SubjectRuler`), so a subject cannot be added with a gavel in one
place and not the other. `rulerFor` PANICS at command construction on an unknown or unannotated
subject: the failure is at startup for every seat rather than at the moment one tries to rule.

**`motion` is the one group that could not be expressed the old way**, and that is why it was
built first. The other groups hung off a role node and `seat.Of` read the role from the
command's POSITION; under `motion grade file` the parent is `grade`. And it should not work that
way here, because a motion is not one seat's act: all four seats file, and one rules. These
verbs take the acting role from the SEAT'S IDENTITY, which is the model the whole tree moved to.

## V. What is NOT collapsed, and why

- **`merge close` and `bench opinion`** both end a gap's life and stay two verbs: the merge
  closes on verified repair (anchor required), the bench closes on judgement (principle
  required). Different evidence bars are a real distinction. Their VOCABULARIES merged (#342)
  into ONE enum — `recordpb.Disposition`, reached as `close --as <closure_class>` and
  `opinion --as <disposition>` — because a split vocabulary is what would re-grow the
  duplication. Before it was unified, four surfaces spelled the same three outcomes six ways and
  no mechanism could see them disagree, because every set was open. The one asymmetry left is
  `carried`, which is bench-only and does not close: "I repaired it by carrying it" is not a
  sentence a merge can say. That subset is ENFORCED rather than documented — `Close.closure_class`
  carries `subset: "closes"`, and the check is generated from the enum's own `(closes)`
  annotation.
- **`lens verify` (a source) and `lens reproduce` (a computation)** are different subjects with
  different evidence, and both record. `reproduce` used to record NOTHING, so the newest
  evidence axis was less audited than the oldest; it now takes `--as` over a soundness enum,
  having READ the script.
- **`friction` on four seats** is not duplication. Not every repetition is a defect. It is
  defined once in `seat.Friction` and shared, and its `--none` form is the EXPLICIT negative:
  across eighteen probed seat dispatches not one friction event was ever recorded, and "no
  friction on the record" is equally consistent with a clean sitting and with a seat that hit
  walls and never used the channel. Those are the same bytes.
- **`cite` is two acts and is now two event types.** Red verifying a source and blue authoring
  a citation were told apart at read time by the ABSENCE of a field
  (`e.Type == "cite" && e.Payload.Str("label") == ""`), so a blue cite without a label counted
  as red's audit volume — a number red reads as how much work it did. `IsVerifiedCite` and
  `IsAuthoredCite` are deleted; the epitaphs at `internal/record/citationid.go:215` and
  `internal/cli/lens/verify.go:25` are deliberate and should stay.

## VI. Decisions that still bind

**`show` is a group, not a flag, and `defaultFor` survived.** `--view <name>` retired: twelve
projections were hiding behind a flag, invisible to `--help` and scoped to nobody. As verbs they
are first-class, per-seat scoped and individually documented. Bare `show` RENDERS the seat's
default — its own pending work — rather than listing, and that is a DECLARED capability
(`BareIsACapability`, `internal/cli/surface.go:164`), not a silent guess; an unknown projection
still gets the listing refusal. **Honest accounting: this does not shrink the count.** Making
hidden things visible is the point; claiming compression that came from hiding them would be the
same defect one level up.

**SHORT IS THE MENU, LONG IS THE INSTRUCTION.** They were one string, so `show --help` carried
every projection's essay and each leaf page repeated one verbatim — `pagesNeverUsed` came back
at 11 of 13 for a seat that opened its whole tree, because the pages beneath held nothing new.
Each `Short` is written to be read AS A SET, and each names the verb that fills the projection.

**`register` was KEPT.** Under detection it is where the harness agent handle binds to the seat
(#290) — "You pass it ONCE, at `register`, which binds it to you on the record". `register` is
also the first and only place that can see the binding was never possible, and it WARNS rather
than refuses: refusing would wedge a run whose only fault is a hook that did not fire, and the
seat can work perfectly well by passing `--seat-id`. Measured, and it is the whole of #512: in
`research/2026-08-22_is-7-prime` all fourteen registers carry no agent id, and the resulting
"this agent has not registered" was returned 92 times and was FALSE every single time.

**`revision` survived the restructure.** #251 would have retired it — the revision summarizes
edits the record already carries with old/new spans — and #251 never landed.
`internal/cli/blue/revision.go` and its seat help page are live. It is named here so a future
restructure cannot delete it by omission, which is the failure mode this exercise exists to
prevent.

**Considered and DECLINED: entity-scoped projections (`<group> show`).** Several views span
entities and would need an arbitrary home: `telemetry` is run-level, and `debate` spans
argument, motion and closure. Forcing those into one entity is a worse lie than keeping
projections in one place — and a reader looking for a projection would then have to guess which
entity someone filed it under, where the answer is always `show`.

**Considered and DECLINED: an `audit` group** (`audit source` + `audit proof` + `audit
closure`). It groups by VERB MOOD, not by subject: the object of one is a source, of another a
closure, and nothing binds them but the mood. `reproduce` belongs beside `prove` for the same
reason `close` belongs beside `mint` — author and audit are two operations on one thing. What
survives is only the OBSERVATION that the author/audit pairing should be visible, and that
belongs in verb naming and help text.

**A verb MAY appear in more than one place.** The rule is not "once only": it must make sense
where it sits, be restricted to that context's operations, and not carry a wildly different
meaning under the same name elsewhere. `friction` (record a complaint) and `show`'s read voice
are the same concept — legitimate. `cite` meaning "author a citation" in one place and "verify a
source" in another was not.

## VII. Verification

From `plugins/frank-exchange-of-views/tools` unless noted:

1. `gofmt -l .` → empty · 2. `go vet ./...` → empty
3. `go test ./...` → exit 0 (**redirect to a file and echo `$?`; piping to `tail` masks it**)
4. `go test ./releasegate/fuzz -run TestFuzzDebate -count=1 -timeout 900s`
5. `go test ./releasegate/fuzz -run TestFuzzHaltPath -count=1`
6. `cd .. && node --test tests/simulator/*.mjs` (**the only check that catches a backtick nested
   in a template literal; `node --check` passes on it**)
7. `cd ../../scripts && go run ./golden` FIRST — this is the DETECTION step and its diff must be
   READ. Only then `go run ./golden -update`. **Reversed, the pair is self-fulfilling**:
   `-update` rewrites the goldens the compare then reads, so it can never fail and proves only
   "the suite matches itself after being updated to match the change".
8. `go run ./frontmatter` · `./validatejson` · `./mjsparity` · `./pluginparity` · `./versionguard`
   · `./decisions` · `./rulesweep` (last, after committing)

The verb-lifecycle gates must stay green — **exists**, **driven**, **discoverable**, **reaches a
reader**. They live in `releasegate/fuzz/promptverbs_test.go` (mirrored in
`integration/surface/`): `TestEveryVerbNamedInAPromptExists`,
`TestEveryRecordingVerbIsDiscoverableFromHelp`, `TestEveryVerbHasATriggerRow`,
`TestEveryViewNamedInAPromptExists`, `TestEveryEnumValueNamedInAPromptIsAccepted`, plus
`proseRenders` / `basisRenders` and the envelope-enum gate in `releasegate/fuzz/fuzz_test.go`.
`TestRoleHelpCarriesTheFrictionFooter` (`internal/cli/cli_test.go`) guards the footer that would
otherwise have been dropped when the role group went away.

`commandPathOf` in `releasegate/fuzz/trajectory_test.go` asks the COBRA TREE for the longest argv
prefix that is a real command path, at any depth. It used to take at most two tokens and require
the first to be a role — a two-level tree hard-coded into the tally — so `motion grade file`
collapsed to `motion`, matched no real path, and reported all seven motion verbs as NEVER
INVOKED while the record showed 35 motion events. A tally that cannot see a path it was given
reports a false absence, which is the plausible zero the gate exists to catch, one level up.
