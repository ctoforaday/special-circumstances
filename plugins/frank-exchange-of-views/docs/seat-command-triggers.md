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

## Lens (`feov-record lens`)

| verb | the one trigger | competing channel | verdict |
|---|---|---|---|
| `register` | first action at the seat | — | CLEAN |
| `finding` | a defect in the report, anchored to a location | — | CLEAN — the canonical lens surface |
| `cite` | a claim verified at its source | envelope `corroboration[]` re-types claim/reference/confidence | COLLAPSE → cite events; envelope drops `corroboration[]` (blue reads `--view citation-ledger`) — **NOT EXECUTED** (#326): the envelope still declares it and three constitutions still name it |
| `observe` | a below-bar note (weaker than a finding) | `finding` — the above/below-bar line is a judgment the lens makes on *every* thing it notices | COLLAPSE → retire; findings superseded it — **RE-OPENED** (#327): #320 gave observations a reader, which was the actual cause of the disuse this rested on |
| `avenue` | a line of inquiry pursued/abandoned/declined | — | CLEAN |
| `friction` | a missing capability | — | CLEAN |
| `petition` | an ethical/safety/integrity/constitutional objection | envelope `petitions[]` is the origination channel; the `petition` event is written but consumers read `petition-rule` + envelope, never the event | RESOLVED (#315, shipped): the `petition` event is the filing, and the report renders it beside its ruling |
| `show` | read a projection | — | CLEAN (read path) |

## Merge (`feov-record merge`)

| verb | the one trigger | competing channel | verdict |
|---|---|---|---|
| `register` / `mint` / `close` | the core board acts | — | CLEAN |
| the open docket (read-back) | transcribe the board into the envelope | `RED_ENVELOPE.gaps[]` re-types every open gap's prose every round; already on the board via `mint` | COLLAPSE → envelope carries `{id, severity, likelihood, impact, supersedes}` only; the tool re-derives prose at capture/assembly (the #1 token lever) |
| `dispose` | give a lens observation a fate | bound to `observe` (being retired) | COLLAPSE → retire with observe — **RE-OPENED** with it (#327) |
| `regrade` | change a gap's grade in place | debate.js prose "apply the new grade in the ledger" never names the verb — seat may re-mint or edit instead | COLLAPSE → `regrade` is canonical; debate.js must name it — **NOT EXECUTED** (#325): zero mentions in debate.js or the constitutions; the report reads it and the fuzz is its only caller |
| `spot-check` | the round archive spot-check duty | the `spot-check` event, and only that — the envelope array is deleted | RESOLVED (#317, shipped): the verb is the single channel, and its W1.8 floor is COMPUTED from the board at verify time rather than reported by the seat |
| `dispute-respond` | answer a blue grade dispute | — | CLEAN (already a routing ref) |
| `verdict` | the terminal PASS/FAIL act | — | CLEAN |
| `existence` (field) | verified\|suspected — checked-at-leaf vs inferred | required by the envelope, read by the board, but **no `mint --existence` flag** — never on the record | HALF EXECUTED (#331): `mint --existence` shipped and is driven; no gap in the report shows verified vs suspected, which is grading v2's whole point |

## Blue (`feov-record blue`)

| verb | the one trigger | competing channel | verdict |
|---|---|---|---|
| `register` / `position` / `closing` / `dispute` / `confidence` / `avenue` / `friction` | their acts | — | CLEAN |
| `revision` | log a per-round revision | hand-written `blue/CHANGELOG.md` | COLLAPSE → `revision` event + changelog view; drop the hand-written file — **NOT EXECUTED** (#251): the event is canonical and rendered, `blue/CHANGELOG.md` is still written and still demanded |
| `retire` | remove a claim from the report (additive integrity) | the detector read a phantom envelope field — **fixed (#226)**; the event is now the sole channel | CLEAN (post-#226) |
| `manifest-row` | a self-audit manifest row | envelope `manifest[]` array (the `row` prose is genuinely envelope-only) | RESOLVED (#318, shipped): the verb is the single source, the row is on the record and in the report, and the envelope carries gap ids only |
| `tldr` / `open_questions` (envelope) | — | authored into `report.md` AND round-tripped in the envelope; never consumed by the script | CLEAN — executed: the envelope declares neither, and assembly lifts both from report.md |

## Bench (`feov-record bench`)

| verb | the one trigger | competing channel | verdict |
|---|---|---|---|
| `register` / `opinion` / `outcome` | their acts | — | CLEAN |
| `petition-rule` | rule granted\|denied\|**halt** on a petition | the ruling's `rationale`/`opinion` is re-typed in the envelope (= the opinion event) | COLLAPSE (prose) — **NOT EXECUTED** (#330): the ruling's `opinion` is still re-typed in the envelope. See the halt fork below |
| **`halt`** | end the run on a safety boundary | `petition-rule --as halt` (debate.js-driven) vs the standalone `halt` verb (tool-designed, orphaned); **the enum rejects `halt`, so the driven path fails to record** | **NOT EXECUTED** (#329) — decided (route through `bench halt`) and never done. debate.js still emits `ruling: 'halt'`, the record enum still refuses it, so a halt-on-petition FAILS TO RECORD. The safety boundary is the one path that must not have this defect |
| `certify` | the bench asking a human to re-examine something at run end | — (the fold-into-`outcome` decision is REVERSED, #328: a certification and a verdict are different speech acts) | CLEAN — unique trigger, a driver, and two readers in the report |

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
   **NOT EXECUTED — #329.** debate.js still emits `ruling: 'halt'` and the record enum still
   refuses it, so a halt-on-petition fails to record. This is the safety boundary.
2. **existence → add `mint --existence`.** Wire the write-path (`verified|suspected`) so the
   leaf-check axis lands on the record and dedups out of the envelope.
   **HALF EXECUTED — #331.** The flag shipped and is driven; no gap in the report shows which it
   is, which is grading v2's whole point.
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

## Status of the collapses

| collapse | state |
|---|---|
| **Envelope round-trip** — the dominant token lever: the envelope carries refs, the tool re-derives from the record | PARTIAL. Done for grade disputes, manifest (#318) and the docket. `corroboration[]` (#326) and the petition ruling's opinion prose (#330) still round-trip. |
| **observe/dispose → retire** | RE-OPENED (#327). The collapse rested on disuse; #320 found the cause was that the work reached no reader, and gave it one. |
| **revision / CHANGELOG.md** — the event is canonical; drop the file | NOT EXECUTED (#251). The last item of the record-tool plan's deletion list; the other four are done. |
| **regrade → canonical; debate.js must name it** | NOT EXECUTED (#325). Zero mentions in debate.js or the constitutions. |
| **tldr / open_questions → drop from the envelope** | DONE. Now `CLEAN` in the table. |
| Shipped before this section had trackers | the board `mint` duplication (#225), the retire detector (#226), the `undisposed_observations` metric (#227). |
