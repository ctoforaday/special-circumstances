# Derived seat identity — the last fact a seat asserts about itself

> STATUS 2026-09-07: proposed, **required pre-v2.0.0** (gblock's ruling). Blocks the three
> `--v2.0.0` tags and #785. Board: #792.

## I. Summary & Goals

`roster.go` states the defect against itself: **"register is the one call that takes a seat's
word for who it is."** Everything after register already derives — `BoundSeat` joins
`agent_id` to the register event and `ResolveSeat` refuses a `--seat-id` that disagrees with
it. The self-assertion survives at exactly one call, and every fact downstream inherits from
it.

**Goal: at `register`, the seat id is COMPUTED from facts the seat cannot supply, and the
`--seat-id` flag stops being an input to identity.**

Identity becomes `(agent_type, round-from-record, agent_id)`:

| fact | source today | source after |
|---|---|---|
| role | `agent_type`, attested (#674) | unchanged |
| **area / lane** | a substring of the typed id | **`agent_type`** — one configuration per seat |
| **round** | `RoundOf`'s `-r(\d+)` over the typed id | **`CurrentRoundOf(record)`** |
| which dispatch | `agent_id`, injected | unchanged |

### Why this is a RELEASE boundary — and what that depends on

The derivation is pre-tag because gblock ruled it. What makes that boundary **load-bearing
rather than arbitrary is III.3's ruling**, and the two have to be read together.

Identity is written into every event. With no provenance field, nothing on a row says whether
its `seat_id` was typed or derived — so a v2 record and a v2.1 record would differ on the
authorship of the same field and say nothing about it, on the artifact `CLAUDE.md` says every
audit re-reads. That is the silent-disagreement case, and the epoch is where it belongs.

**Stated plainly because it is a real trade and it cuts the other way too:** had III.3 gone the
other way and stamped `ATTESTED`/`DECLARED`, records COULD have disagreed about authorship and
said so in the row — which would have made this change safe at any boundary, tag or no tag. The
field that would make it safe is the same field that would make the boundary unnecessary. The
enum was chosen instead because it buys a REFUSAL the field does not, and the price of that
choice is exactly this: the change becomes a one-way door, and the door is the tag. Anyone
re-reading this plan and finding the boundary argument convenient should know it was bought.

(Raised by a peer session reviewing the rationale, 2026-09-07, against the version of this plan
that still recommended the field. The tension was real and it is resolved by naming the trade,
not by dropping the argument.)

### What this is NOT

**Not the deletion of `seat_id`.** It stays the key namespace: `deriveKey` scopes the per-seat
ordinal to it and every SQL projection groups on it. The column is unchanged; its AUTHOR
changes. Deleting it is a different and much larger change with no payoff.

**Not persistence.** #748/#749/#624 (a lane that survives rounds and is messaged "start round
X") would make the round arrive as a message. That is blocked on a harness primitive `agent()`
does not have — it returns a result, not a handle. This plan needs none of it: the record
already knows the round.

## II. Technical Context

Verified in-tree 2026-09-07 at `1c28e254`, by reading each cited line.

**The register path.** `seat.go:247` `Of()` → `ResolveSeat(seatID, BoundSeat(run),
record.RoundIn(run))` at `:264`. `BoundSeat` (`:234`) returns nil unless the run is valid AND
`seatenv.AgentID() != ""`; otherwise `record.SeatOfAgent(run, AgentID())`. `ResolveSeat`
(`seatenv/identity.go:120`) prefers the bound value, refuses a disagreeing flag with
`feov.Conflict`, and falls back to the flag when nothing is bound.

**The round path.** `RoundIn(run)` → `RoundOf(seatID)` → `roundRe = -r(\d+)`
(`record/round.go:8`). `RoundOf` also answers 0 for `frontier`, `blue-synthesize` and
`blue-lane-\d+` (dispatched before the round loop), and `RoundIn` derives the round for
`judge-terminal` and `assemble` **from the record** — `round.go:62` calls that "a derivation
from FIELDS rather than from a name, which is the distinction this whole exercise is about."
The pattern already exists; it was applied only where a name could not answer.

Consumers of the regex, complete: `seat.go:264`, `seat.go:519`, and
`consistency/consistency.go:473` (which is #676's subject — it cross-checks the stamped
`event.round` against the regex).

**The attestation table.** `record/agentrole.go:35` maps four `agent_type` values to role
SETS, because the mapping is not one-to-one: `red-auditor` covers lens AND merge;
`blue-researcher` covers all four lanes. `agentrole.go:22` states this as a real limit and
predicts the fix: *"It narrows to exact if red-merge is ever given its own agent
configuration."*

**What the engine dispatches** (`debate.js`): `blue-researcher` ×4, `blue-synthesizer` ×2,
`lead-judge` ×4, `red-auditor` ×2 — four configurations for the whole cast.

**Per-area content today.** `RED_AREAS` lens strings are 102–510 characters each, plus
conditional clauses composed at dispatch (`ledgerClause` + `consolidatedClause` for evidence,
`steelmanClause` for logic and dark-side). `agents/red-auditor.md` is 68 lines of shared
skepticism, gate ownership and protocol, and it already delegates via
`skills: [research-protocol, critical-stance, terse-communication]`.

## III. Proposed Changes (the spec)

### III.1 One agent configuration per seat

Seven lens areas, red-merge, four blue lanes, and the existing synthesizer and bench. The
shared 68 lines do **not** get copied twelve times: they move into the `research-protocol`
skill (or a sibling), and each agent file becomes frontmatter plus its own lens paragraph —
which is what the `skills:` key is for and what `red-auditor.md` already half does.

**Hand-written content, gated membership.** The files are prompts and prompts are tuned by
hand per area; generating their bodies would make the one thing worth editing the one thing
you must not edit. What is machine-checked is the SET:
`TestTheLensAreasMatchWhatTheEngineDeclares` extends to a third carrier, so
`agents/*.md` ≡ `RED_AREAS` ≡ `record.LensAreas` in all directions.

`agentrole.go`'s sets collapse to scalars and the ambiguous row disappears.

**This is load-bearing, not tidiness.** Without it, blue's four lanes share one configuration
inside a single round and the derivation cannot distinguish them.

**They cannot be hidden from the consumer's agent picker, and that was checked rather than
assumed.** Agent frontmatter in this repository uses five keys — `name`, `description`, `tools`,
`skills`, `memory` — and no plugin here declares anything that marks an agent internal. No such
affordance is vendored or documented in-tree, so the picker growing from 4 feov entries to ~13
is the accepted cost (gblock, 2026-09-07, on the stated fallback). If the harness gains one, it
is a frontmatter line per file and nothing else.

### III.2 Round from the record

`RoundOf` is deleted. `RoundIn` keeps its record-derived branch and becomes the only answer:
the run's current round for an ordinary seat, the last round for a terminal one, 0 before the
loop opens. `consistency.go:473`'s cross-check loses its second opinion and closes #676 — there
is no longer a name to disagree with the field.

### III.3 The declared path is REFUSABLE — decided by gblock 2026-09-07

`CheckAttestedRole` treats an absent `agent_type` as **not a violation**, deliberately: an
operator at a shell, CI, the Go tests and the simulator have no attestation, and refusing them
demands a mechanism their environment does not have. Once identity is DERIVED from
`agent_type`, absent means no identity at all — so that branch becomes either a hard refusal
that breaks every non-harness caller, or a fallback.

**RULED: fallback, and the fallback is refusable.** `--seat-id` is accepted when nothing is
attested, and it is checked against two things it can fail:

1. **A closed ENUM, not a shape.** Per-seat configurations (III.1) plus the record's round make
   the run's admissible seat ids *enumerable* — this is the change that turns the roster from a
   pattern into a list. `red-lens-r99-evidence` is refused because the run has no round 99, not
   because `r99` looks wrong. That refusal does not exist today at any strictness: the shape
   admits it and 99 is stamped into `event.round`.
2. **No conflict with registration.** A declared value that disagrees with what the record
   already binds for this `agent_id` is refused — `ResolveSeat`'s existing `feov.Conflict`
   refusal (`seatenv/identity.go:137`), which today guards every act EXCEPT the one that
   creates the binding. It now guards that one too.

This is stronger than stamping provenance and watching it, which was the alternative: a
DECLARED value that must be a member of the run's own enum and must not contradict the record
is not a weaker fact than a derived one — it is the same fact, checked at the door instead of
authored there. An `ATTESTED`/`DECLARED` field remains a cheap addition if a run ever needs to
answer "how much of this was attested", and is deliberately not built now.

### III.4 `red-merge` is renamed `chair` — decided by gblock 2026-09-07

The seat that merges the round and owns the board is the CHAIR. The rename lands here rather
than separately because it moves the same carriers this plan is already moving: the seat id
(`red-merge-r\d+`), the role in `agentrole.go`, `event.role`'s value, the roster, the agent
configuration III.1 gives it, the engine's prompts and their goldens.

**Archived runs are not rewritten.** The reader takes both, exactly as the lens rename did in
#791: a `red-merge-r1` seat id and `role: merge` in an August run stay readable forever, and the
two names cannot collide. Nothing is renumbered and no record is migrated.

## IV. Risk & Mitigation

| risk | mitigation |
|---|---|
| **An act lands after its round advances** and is stamped with the new round. Today the name pins it. | The engine `await`s each `parallel()` phase, so a straggler should be impossible — **confirm by execution, not by reading the dispatch code.** This is the failure mode the whole plan exists to remove, so it may not be assumed away. |
| **A re-dispatch inside one round** makes `(agent_type, round)` ambiguous. | `agent_id` disambiguates and the join already exists (`SeatOfAgent`). The proto comment at `record.proto:448` records the prior collision and its fix. |
| **Twelve agent files drift from `RED_AREAS`.** | III.1's three-way bind test. A carrier added without its area fails; an area added without its carrier fails. |
| **The consumer's agent list grows** from 4 feov entries to ~13. | Real and unmitigated. Named here so it is a decision rather than a discovery. |
| **The untyped population may be large.** III.3's fallback is the path taken whenever `agent_type` is absent, and if that is most acts the enum check is doing nearly all the work while the derivation does nearly none. | Measure it on the first run rather than assume. A neighbouring surface in this repo — gray-area's `SubagentStop` rows — found `agent_type` absent on 146 of 165, its single most common state, and took four investigations to stop reading that absence as a defect (peer session, 2026-09-07). **The base rate does not transfer**: that is a harness hook payload, not feov's own attestation path, and the populations are not comparable. What transfers is the shape of the mistake to avoid — treating absence as an anomaly rather than as a value the system will spend most of its time in. |
| **The simulator dispatches by `agentType`** and its goldens name seats. | 96 tests are the gate; they must be regenerated deliberately, and a diff that changes a seat id is the thing to read closely rather than accept. |

## V. Verification Plan

Each check with what re-arms it.

```bash
# The three-way carrier bind. Re-arms: agents/*.md, RED_AREAS, record.LensAreas.
(cd plugins/frank-exchange-of-views/tools && go test ./internal/record/ -run TestTheLensAreas)

# Attestation completeness — every dispatched agent_type is in the table.
# Re-arms: debate.js's agentType strings, agentrole.go.
(cd plugins/frank-exchange-of-views/tools && go test ./internal/record/ -run TestEveryDispatchedAgentType)

# The regex is GONE, not merely unused. Re-arms: any reintroduction.
! grep -rn 'roundRe\|RoundOf(' --include='*.go' plugins/frank-exchange-of-views/tools/internal/

# Full suites. Re-arms: any change under tools/ or the engine.
(cd plugins/frank-exchange-of-views/tools && go vet ./... && go test ./...)
node --test plugins/frank-exchange-of-views/tests/simulator/debate.test.mjs \
             plugins/frank-exchange-of-views/tests/simulator/prompts.test.mjs

# Gates. Re-arms: agent-facing surfaces (archaeology), protocol surfaces (rule-sweep),
# plugin manifests and binary counts (pluginparity).
(cd scripts && go run ./check)
```

**The check no suite can make**, and the one that decides III.2: a real run, reading whether
any seat's acts were stamped with a round it did not act in. `event.round` is already on every
row, so the query is available the moment a run exists — it does not need new instrumentation.
This is B4's smoke on #792 doing double duty.

## VI. Deliberately not in this plan

- **Deleting `seat_id`.** §I.
- **Persistent lanes and round-as-message** (#748, #749, #624). Blocked on a harness primitive;
  this plan is designed so that arriving later changes nothing here — the round would move from
  a record read to a message, and neither is a name.
- **Refusing an unattested seat.** III.3 makes it visible; making it fatal is a later call on
  evidence this plan produces.
- **Blue lane METHOD as a second carrier.** `LANE_METHODS[i % len]` recovers a lane's method
  from its index — the same fork #791 healed for lenses, flagged in that PR's sibling sweep and
  left for #497. Per-lane configurations make it trivial afterwards; doing it here widens the
  change without shortening the path to the tag.
