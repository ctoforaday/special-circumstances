# Seat command triggers — the one-way contract

> **Purpose.** Every seat learns its verbs from `feov-record <role> --help` (the help *is* the
> instruction — the seats read it, debate.js tells them to "discover"). For that to make a seat
> RELIABLE, each situation it can be in must map to **exactly one** command. Where two channels
> reach the same act — a verb *and* an envelope field, a hand-written file, or a debate.js prose
> instruction — the seat has to decide "which do I run?", and choice-under-ambiguity is where a
> seat gets it wrong or inconsistent across runs. That ambiguity is the cost, even when both
> channels "work".
>
> This table is the contract: for each verb, the ONE trigger that should invoke it, and whether a
> competing channel creates an ambiguous trigger. **Kept current with the code** — update it in the
> same change as any verb/envelope/prompt move. A verb with a clear unique trigger stays regardless
> of how often it fires; a verb whose trigger overlaps another channel is a reliability liability to
> collapse. Fewer verbs is never the goal; **no ambiguous triggers** is.

## The test

For a candidate collapse, both must hold: (1) two channels reach the same act, AND (2) removing one
**reduces** a real ambiguity the seat would otherwise face. Collapsing toward the **record** (the
verb) is the default — it is the auditable single source; envelopes, files, and prose derive from it.

## Legend

`CLEAN` = unique trigger, no competing channel. `COLLAPSE` = an ambiguous trigger to resolve (the
canonical channel named). `DECIDE` = a genuine fork needing a human call before collapse.

## Every seat (`register` · `friction` · `show`)

Three acts every role carries, listed once. The full command path is in the first cell of every row
in this file — **that is what makes the table checkable** (`TestEveryVerbHasATriggerRow`), and it
replaces the role headings that used to carry the role as prose beside a bare verb name.

| command | the one trigger | competing channel | verdict |
|---|---|---|---|
| `lens register` `merge register` `blue register` `bench register` | first action at the seat | — | CLEAN |
| `lens friction` `merge friction` `blue friction` `bench friction` | a capability gap, or the explicit `--none` that says nothing blocked you | — | CLEAN. Every seat WRITES it; the READ is the operator's (`feov-record friction`), because a capability gap is a report to the human who can retool the seat, not material for the debate |
| `lens show` `merge show` `blue show` `bench show` | read a projection | — | CLEAN (read path). The projections are their own vocabulary with their own gate — see `TestEveryViewNamesTheVerbThatFillsIt` — so the `show <view>` subtree is not enumerated here |

## Lens (`feov-record lens`)

| command | the one trigger | competing channel | verdict |
|---|---|---|---|
| `lens finding` | a defect in the report, anchored to a location | — | CLEAN — the canonical lens surface |
| `lens verify` | you checked a citation against its source, or corroborated a claim from a source you found | envelope `corroboration[]` re-typed claim/reference/confidence | EXECUTED (#326/#382): the envelope carries no corroboration array, and blue reads what red found from `show evidence`. The verb now carries BOTH axes — `--as` what the source did (with its negative half: `refutes`, `absent`) and `--confidence` how sure you are of that — plus `--anchor` naming which citation, or `--independent` for a source blue never cited |
| `lens reproduce` | a proof anchor in the report you can re-run and read | — | CLEAN. Two acts under one verb ON PURPOSE, and both are recorded: the tool COMPUTES whether it reproduced, and `--as sound\|unsound` is red's judgement from reading the script. A seat that perceives only the mechanical half correctly calls it theatre — measured, in exactly those words — which is why the reading is required rather than optional |
| ~~`lens observe`~~ | RETIRED (#327) — a below-bar note | — | EXECUTED: the above/below-bar line is one judgement the lens makes on everything it notices, and `finding` plus a grade expresses it |
| ~~`lens cite`~~ | RENAMED to `verify` (#341) | — | EXECUTED. It shared the `cite` event type with blue's authoring act, told apart by the ABSENCE of a label, so a blue cite written without one counted as red's audit volume |

## Merge (`feov-record merge`)

| command | the one trigger | competing channel | verdict |
|---|---|---|---|
| `merge mint` | a defect worth putting on the board | `RED_ENVELOPE.gaps[]` re-typed every open gap's prose every round | EXECUTED: the envelope carries routing refs and the board is read back with `show worklist` |
| `merge close` | a gap answered, with an anchor naming who checked what with which tool | — | CLEAN |
| `merge regrade` | change a gap's grade in place | debate.js prose "apply the new grade in the ledger" never named the verb, so a seat may re-mint or edit instead | EXECUTED (#325, verified 2026-08-13): named in debate.js and in red-auditor.md. The 2026-08-12 elicitation then found a merge seat listing it unprompted — *"Regrade R1-1 without closing it — lower severity, keep it open until blue actually edits"* — which is the move from *never perceived* to *weighed* the naming was for |
| `merge near-match` | before minting: is this gap already on the board, open or closed? | the archive read the merge would otherwise do by hand | CLEAN (read-only, records nothing). Same elicitation result as `regrade` — unnamed in the prompts, unlisted by the seats |
| `merge spot-check` | the round archive spot-check duty | the envelope array is deleted | RESOLVED (#317): the verb is the single channel and its floor is computed from the board |
| `merge position` | the round's RED narrative | a hand-written `### RED` section in debate.md | EXECUTED: the transcript is rendered from `position` events |
| `merge closing` | a closing argument on a docketed item | — | CLEAN |
| `merge verdict` | the terminal PASS/FAIL act | — | CLEAN. Refused while a gap is open or a motion is unruled — the gate is enforced at the tool, not trusted to the seat |
| ~~`merge dispute-respond`~~ | COLLAPSED into `motion grade rule` (#344) | — | EXECUTED |

## Blue (`feov-record blue`)

| command | the one trigger | competing channel | verdict |
|---|---|---|---|
| `blue edit` | any change to `blue/report.md` after round 0 | a raw Write/Edit to the file | EXECUTED — the lockdown: report.md is read-only to a response seat and the tool refuses an edit that drops, duplicates or invents an anchor |
| `blue cite` | a source backing a claim you are authoring | a hand-typed `[^1]` footnote | EXECUTED: the tool fetches, caches, hashes and splices the invisible anchor; assembly weaves the bibliography. Blue never types a footnote |
| `blue prove` | a claim a program settles, and the gaps whose `check_kind` is `computation` | prose asserting the computation happened | EXECUTED (#277): a computation gap CANNOT be closed on prose, and the tool runs the script twice and records which of reproducible/observed it produced |
| `blue avenue` | a line of inquiry proposed, pursued, declined, abandoned or deferred | — | CLEAN |
| `blue retire` | a claim leaving the report | a claim quietly not being there any more | EXECUTED (#226): capture compares the claim-count fall against the retire events, and an unaccounted drop is a detector hit |
| `blue revision` | log a per-round revision | hand-written `blue/CHANGELOG.md` | NOT EXECUTED (#251). The event is canonical and the file is still authored beside it |
| `blue manifest-row` | the self-audit receipt for a repaired gap | envelope `manifest[]` array | RESOLVED (#318): the verb is the single source, the row is on the record, and a closed gap with no row is named in the report as a repair nobody audited |
| `blue position` | the round's BLUE narrative | a hand-written `### BLUE` section | EXECUTED |
| `blue closing` | a closing argument on a docketed item | — | CLEAN |
| `blue claim-index` | locating every site of a footnoted claim you are correcting | re-reading the whole report to hunt them | CLEAN (read-only, records nothing) |
| ~~`blue confidence`~~ | DELETED (0.54.0) — blue self-grading its own claims | — | EXECUTED. It set no grade, entered no matrix, and its calibration computation was specified and never built. The word returned to the field the plan meant by it: `lens verify --confidence`, red's per-pair judgement |
| ~~`blue dispute`~~ | COLLAPSED into `motion grade file` (#344) | — | EXECUTED |
| `tldr` / `open_questions` (envelope) | — | authored into the report AND round-tripped in the envelope | CLEAN — executed: the envelope declares neither |

## Bench (`feov-record bench`)

| command | the one trigger | competing channel | verdict |
|---|---|---|---|
| `bench opinion` | ruling on a docketed item, with principle, tension and review-flag | — | CLEAN |
| `bench outcome` | the run's terminal act | — | CLEAN. It carried no reasoning at all until a bench seat reached for `--reason`, found nothing, and filed the absence as friction — so `--reason` is required (#375): the verdict is derived, how the sitting ENDED is not |
| `bench halt` | end the run on a safety boundary | `motion petition rule --as halt` | EXECUTED (#329, verified 2026-08-13): debate.js emits `ruling: 'halt'` zero times; the halt is `bench halt` and the petition enum rules `granted\|denied` only |
| `bench certify` | asking a human to re-examine something at run end | — | CLEAN. The fold-into-`outcome` decision is REVERSED (#328): a certification and a verdict are different speech acts |
| `bench declare` | a holding that binds how the record is READ and moves no gap — a construction of a term, a correction of meaning, a rule offered for precedent | the petition ruling's opinion text, which is where declarations went for want of a verb | RESOLVED (#361): `opinion` demands an id and a fate, so a bench with a finding both seats needed had nowhere to put it. It renders under `### LEAD` beside the opinions |
| `bench assemble` | compose the final report from the record and blue's audited report | — | CLEAN |

## Motion (`feov-record motion <subject> <act>`)

One mechanism for the three propose→rule exchanges, each ask carrying an id its answer names.

| command | the one trigger | competing channel | verdict |
|---|---|---|---|
| `motion grade file` | you dispute a gap's severity, likelihood, impact or complexity | `blue dispute` | EXECUTED (#344) |
| `motion grade rule` | red answering a grade motion | `merge dispute-respond` | EXECUTED (#344) |
| `motion grade appeal` | pressing a rejected grade ruling to the bench | — | CLEAN |
| `motion direction rule` | red ruling on a proposed avenue | `merge avenue-rule` | EXECUTED (#344) |
| `motion direction appeal` | pressing a direction ruling to the bench | — | CLEAN |
| `motion petition file` | an ethical, safety, integrity or constitutional objection | envelope `petitions[]` | EXECUTED (#315): the event is the origination channel |
| `motion petition rule` | the bench ruling on a petition | — | CLEAN, except for the halt channel above (#329) |

> **The whole group was used on no subject in the 2026-08-12 elicitation** — `rule` fired nine
> times, `file` not once. Seats read grades and never treat them as contestable: 10 of 110 blue
> thinking blocks mention a grade and none weighs whether one is wrong. That is not the verbs'
> discoverability; it is the absence of a frame for the value judgement, and it is the open
> question behind this group's disuse.

## Decisions taken (the `DECIDE` rows)

> **A DECISION WITH NO ISSUE IS NOT RESOLVED — IT IS RECORDED, WHICH IS A DIFFERENT THING.**
>
> This section used to be headed "Resolved decisions", and of the four below, **two had never been
> executed**. Most of the collapses in the table above had not either. Nothing anywhere reported
> that, because a decision never carried out is byte-identical to one that was: the only artifact
> either produces is a paragraph saying it was decided. A full independent trace of the command
> surface on 2026-08-08 REDISCOVERED two of these from scratch, and found an issue filed eight days
> earlier that had already named three of them with the right diagnosis.
>
> So every verdict cell in the table now carries an issue number, and `go run ./scripts/decisions`
> FAILS on a `COLLAPSE`, `DECIDE` or `RESOLVED` that names none. Deciding is cheap; the tracker is
> what makes the decision a fact another party can act on rather than prose nothing can refuse
> (see `facts-are-fields`). Executed rows become `CLEAN` — a description of the present needs no
> tracker.

1. **halt channel → the `halt` verb is canonical.** A halt is the bench's own verb, not a petition
   ruling: debate.js routes a halt-on-petition through `bench halt` (carrying the opinion), and
   `petition-rule` records `granted|denied` only. *(Decided against aligning the enum to the driven
   path — halt is a first-class terminal act, distinct from a petition disposition.)*
   **EXECUTED — #329, verified 2026-08-13.** debate.js emits `ruling: 'halt'` zero times and a
   halt goes through `bench halt`. This paragraph claimed the safety boundary was broken for as
   long as the fix had been in — see the note under this list.
2. ~~**existence → add `mint --existence`.**~~ **REVERSED (#359, 2026-08-14) — the axis is
   DELETED.** It was created to fix a real ranking bug: v1's `likelihood` carried two questions,
   and since the board ranks by likelihood x impact, a `certain` typo outranked a `high` design
   flaw. **The value was in emptying `likelihood`, and that half worked.** The receptacle the
   other question went into never did any work — not required at mint, absent from `GapMass` so
   it ranked nothing, named in one prompt and no constitution, and uncontestable.

   And it was the last unanchored SELF-REPORT in a system that anchors everything else: a closure
   names who checked what with which tool, a proof is re-run, a citation is re-fetched. `existence:
   verified` was red's word about red's own diligence. The incident behind it — three gaps minted
   `verified` at report sections that did not exist — is better explained by `mint --location`
   having been unmatched, which 0.63.0 fixed.

   Kept here rather than deleted, like the `certify` reversal above: a decision that shipped, did
   nothing for six releases, and was removed is evidence about how these calls get made. This one
   was taken as the tidy other half of a real fix, and nobody asked what the half would DO.
3. **petition filing / spot-check / manifest-row → the record verb is the single source.** The verb
   event is canonical; the envelope carries a routing ref, not the data.
   **EXECUTED — #315 (petition filing), #317 (spot-check, with its W1.8 floor now computed from the
   board), #318 (manifest-row).**
4. ~~**certify → fold into `outcome`.**~~ **REVERSED (#328, 2026-08-09) — `certify` stays its own
   verb.** The original decision read a certification as part of the terminal verdict. It is not:
   a certification is *the bench asking a human to re-examine something*, and a verdict is *the
   run's terminal state*. Folding them would put an ask-a-human into a state field, and the two
   can hold independently — a VERIFIED run can still carry a certification, and a certification
   is not a verdict of any kind. `certify` has a driver, and it has TWO readers (the report
   promotes it into "Read this first" as the bench's terminal ask, and renders it again under
   Bench disposition), so nothing was orphaned; only the recorded decision was wrong.

   Kept here rather than deleted: a reversed decision is evidence about how these calls get made.
   This one was taken on a tidy-looking symmetry — "a certification is part of the verdict" — and
   the code disagreed for a year without anyone noticing, because nothing reconciled the two.

> **THREE OF THE ROWS BELOW WERE FALSE FOR WEEKS, AND EACH NAMED A CLOSED ISSUE.**
>
> Verified 2026-08-13: `#329` said the halt channel was broken ("this is the safety boundary");
> `#331` said no gap shows verified-vs-suspected; `#325` said `regrade` had zero mentions. All
> three were fixed, all three issues were CLOSED, and this file went on asserting the opposite.
>
> The mechanism was the section this note sits in. `scripts/decisions` checks that a tracked
> verdict NAMES an issue — it never checked whether the issue was still open, so a tracker
> discharged elsewhere left the claim standing with a valid-looking reference attached. A
> reference is evidence that someone filed something, not that the thing is still true.
>
> `NOT EXECUTED` and `HALF EXECUTED` are tracked verdicts now, and the gate resolves each named
> issue when `gh` is reachable — failing on a closed tracker beside an unexecuted claim. Where it
> cannot reach the API it says so out loud rather than passing: an unchecked claim and a checked
> one must not print the same line.

## Status of the collapses

| collapse | state |
|---|---|
| **Envelope round-trip** — the dominant token lever: the envelope carries refs, the tool re-derives from the record | PARTIAL. Done for grade disputes, manifest (#318) and the docket. `corroboration[]` (#326) and the petition ruling's opinion prose (#330) still round-trip. |
| **observe/dispose → retire** | EXECUTED (#327, 2026-08-09). Re-opened after #320 gave observations a reader, then decided to retire regardless: the two verbs and their events are gone, and "undisposed" became "credited by no gap". **The cost is stated, not hidden — `checked-held` (a check red RAN and confirmed) has no successor vocabulary**, so the run can no longer record a confirmed negative outside the spot-check and proof paths. |
| **revision / CHANGELOG.md** — the event is canonical; drop the file | NOT EXECUTED (#251). The last item of the record-tool plan's deletion list; the other four are done. |
| **regrade → canonical; debate.js must name it** | EXECUTED (#325, verified 2026-08-13). Named in debate.js and red-auditor.md, and a probed merge seat listed it among its options unprompted. |
| **tldr / open_questions → drop from the envelope** | DONE. Now `CLEAN` in the table. |
| Shipped before this section had trackers | the board `mint` duplication (#225), the retire detector (#226), the `undisposed_observations` metric (#227). |
| **the three propose→rule exchanges → one `motion` group** | EXECUTED (#344, 2026-08-10). `blue dispute`, `merge dispute-respond`, `<seat> petition`, `bench petition-rule` and `merge avenue-rule` are GONE, and so is the `contests_ruling` field. They are `motion <subject> file|rule|appeal`, subgrouped by subject, joined on a minted id. **The rows above that name those verbs are HISTORY and are deliberately not rewritten** — they record what was decided when it was decided, and a ledger edited to match the present tense stops being evidence about how these calls get made. Read them against this row. Three consequences worth naming: `petition-rule`'s opinion round-trip (#330) is now a motion ruling's `opinion`; the `(gap_id, dimension)` join that row 50 calls CLEAN is superseded by the id, which is the only thing that ever fixed #312; and the dual-read of all five retired types is PERMANENT, not a migration window — a record is permanent, and this plugin cannot see an installing project's records to know when the old shapes are gone. |
