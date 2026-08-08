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
| `cite` | a claim verified at its source | envelope `corroboration[]` re-types claim/reference/confidence | COLLAPSE → cite events; envelope drops `corroboration[]` (blue reads `--view citation-ledger`) |
| `observe` | a below-bar note (weaker than a finding) | `finding` — the above/below-bar line is a judgment the lens makes on *every* thing it notices | COLLAPSE → retire; findings superseded it (0 recent use; debate.js never drove it) |
| `avenue` | a line of inquiry pursued/abandoned/declined | — | CLEAN |
| `friction` | a missing capability | — | CLEAN |
| `petition` | an ethical/safety/integrity/constitutional objection | envelope `petitions[]` is the origination channel; the `petition` event is written but consumers read `petition-rule` + envelope, never the event | DECIDE — is the event or the envelope the record of a filing? |
| `show` | read a projection | — | CLEAN (read path) |

## Merge (`feov-record merge`)

| verb | the one trigger | competing channel | verdict |
|---|---|---|---|
| `register` / `mint` / `close` | the core board acts | — | CLEAN |
| the open docket (read-back) | transcribe the board into the envelope | `RED_ENVELOPE.gaps[]` re-types every open gap's prose every round; already on the board via `mint` | COLLAPSE → envelope carries `{id, severity, likelihood, impact, supersedes}` only; the tool re-derives prose at capture/assembly (the #1 token lever) |
| `dispose` | give a lens observation a fate | bound to `observe` (being retired) | COLLAPSE → retire with observe |
| `regrade` | change a gap's grade in place | debate.js prose "apply the new grade in the ledger" never names the verb — seat may re-mint or edit instead | COLLAPSE → `regrade` is canonical; debate.js must name it |
| `spot-check` | the round archive spot-check duty | the `spot-check` event, and only that — the envelope array is deleted | RESOLVED (#317, shipped): the verb is the single channel, and its W1.8 floor is COMPUTED from the board at verify time rather than reported by the seat |
| `dispute-respond` | answer a blue grade dispute | — | CLEAN (already a routing ref) |
| `verdict` | the terminal PASS/FAIL act | — | CLEAN |
| `existence` (field) | verified\|suspected — checked-at-leaf vs inferred | required by the envelope, read by the board, but **no `mint --existence` flag** — never on the record | DECIDE → add the write-path (then it dedups) or drop the board plumbing |

## Blue (`feov-record blue`)

| verb | the one trigger | competing channel | verdict |
|---|---|---|---|
| `register` / `position` / `closing` / `dispute` / `confidence` / `avenue` / `friction` | their acts | — | CLEAN |
| `revision` | log a per-round revision | hand-written `blue/CHANGELOG.md` | COLLAPSE → `revision` event + changelog view; drop the hand-written file |
| `retire` | remove a claim from the report (additive integrity) | the detector read a phantom envelope field — **fixed (#226)**; the event is now the sole channel | CLEAN (post-#226) |
| `manifest-row` | a self-audit manifest row | envelope `manifest[]` array (the `row` prose is genuinely envelope-only) | DECIDE — but note the row is NOT on the record today |
| `tldr` / `open_questions` (envelope) | — | authored into `report.md` AND round-tripped in the envelope; never consumed by the script | COLLAPSE → drop from the envelope; assembly lifts them from report.md |

## Bench (`feov-record bench`)

| verb | the one trigger | competing channel | verdict |
|---|---|---|---|
| `register` / `opinion` / `outcome` | their acts | — | CLEAN |
| `petition-rule` | rule granted\|denied\|**halt** on a petition | the ruling's `rationale`/`opinion` is re-typed in the envelope (= the opinion event) | COLLAPSE (prose) — and see the halt fork below |
| **`halt`** | end the run on a safety boundary | `petition-rule --as halt` (debate.js-driven) vs the standalone `halt` verb (tool-designed, orphaned); **the enum rejects `halt`, so the driven path fails to record** | **DECIDE** — the live fork: make `petition-rule` accept `halt` (align with the driver; retire the `halt` verb) OR route halt through the `halt` verb (change debate.js) |
| `certify` | a run-end certification | read by assembly but driven by no prompt | DECIDE — wire it or fold into `outcome` |

## Resolved decisions (the `DECIDE` rows)

1. **halt channel → the `halt` verb is canonical.** The tool's original design stands: a halt is
   the bench's own verb, not a petition ruling. debate.js changes so a halt-on-petition routes
   through `bench halt` (carrying the opinion); `petition-rule` records `granted|denied` only, and
   its enum stays as-is. The `halt` verb is KEPT and wired; it is not orphaned once debate.js names
   it. *(Decided against aligning the enum to the driven path — the design intent that halt is a
   first-class terminal act, distinct from a petition disposition, won.)*
2. **existence → add `mint --existence`.** Wire the write-path (enum `verified|suspected`) so the
   leaf-check axis lands on the record and dedups out of the envelope. The board already reads it.
3. **petition filing / spot-check / manifest-row → the record verb is the single source.** The verb
   event is canonical; the envelope carries a routing ref, not the data. (manifest's `row` prose is
   genuinely envelope-only today — the verb must store it for this to hold.)
4. **certify → fold into `outcome`.** A certification is part of the terminal verdict; retire the
   `certify` verb and let `outcome` carry it.

## Status of the collapses

- **Envelope round-trip** (gaps prose, corroboration, tldr/open_questions, opinion prose) → the (c)
  work: envelope carries refs; the tool re-derives from the record. The dominant token lever.
- **observe/dispose** → retire (findings superseded them).
- **revision / CHANGELOG.md** → the event is canonical; drop the file.
- **regrade** → canonical; debate.js must name it.
- Shipped already: the board `mint` duplication (#225), the retire detector (#226), the
  `undisposed_observations` metric (#227).
