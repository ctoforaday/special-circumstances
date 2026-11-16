# Anchors — one lifecycle for every kind

> **Purpose.** Red's gaps and blue's evidence must not be able to silently disappear from the
> report, and blue's repair must be verifiable against what red flagged. This note is the
> design-of-record for how the tool anchors a citation, a computation or a gap at its sentence in
> the report with a durable token that carries the act's identity. It is the concrete realization of
> slice 1b of [propagation-and-anchoring.md](propagation-and-anchoring.md).

## The kinds

The kinds are one table, `kinds` in `internal/anchor/anchor.go`. Each row says what an anchor of
the kind MEANS; none says how it is placed or carried, because every kind lives the same way.

| kind | id | token | placed by | at assembly |
|---|---|---|---|---|
| citation | `C-` + 8 hex | `<!--cite:C-…-->` | `blue cite` | **woven** to a visible `[^N]` footnote on its source |
| computation | `K-` + 8 hex | `<!--computation:K-…-->` | `blue compute` | **woven** to a visible `[^K<n>]` reference to its computation |
| gap | `G-` + 8 hex | `<!--gap:G-…-->` | `lens mint` | **stripped** |

A citation and a computation are **evidence anchors**: each stands behind its sentence and makes the
prose before it a counted claim. A gap anchor is red's locator for a defect: a machine-readable
pointer, NOT reader-facing content.

## The token

- Format `<!--TAG:ID-->` — an **invisible HTML comment**. It renders to nothing, is not a footnote
  (so no seat reads it as content or audits it as an undefined reference), and a comment is never a
  `[^…]`, so no namespace rule or `Count` exclusion is needed.
- `ID` is the act's **tool-minted id**. The kind is carried by the id's prefix: the minting verb
  chooses it, and every reader recovers the kind from it. The id and its token are minted together
  and never travel apart, and an id of no kind spells a token no reader recognises.
- The gap RECORD holds its quote as minted, with its problem, fix, check and reason. That is what
  makes red's re-audit independent of blue: the lens re-locates and re-checks against the original,
  and re-finds the problem even if blue tried to bury it.

## Why invisible

A machine anchor dressed as reader-facing content is audited as content. A visible `[^G-…]` marker
with no definition reads as an **undefined footnote reference**, so red would audit the tool's own
anchors as a defect of the report, and a gap's text that quoted one would carry it into every
record-derived section the report prints. The invisible comment is a locator and nothing else. The
visible-footnote + definition-resolution model belongs to the evidence anchors, where a reference
genuinely IS reader-facing content that resolves to a source or a computation.

## One lifecycle, every kind

Anchors of every kind live the same way. The tool places an anchor at the end of its quote — the
quote's one occurrence within one paragraph, or the act is refused. Every edit carries it; the tool
puts back one whose sentence an edit keeps word for word; an edit that drops or types one is refused.
**Retire is the one exit.** An evidence anchor never stands bare: an edit that would leave one with no
prose before it is refused, whether the edit carries it or not, and `blue retire --anchor` alone takes
either off a sentence that stays; a citation a live computation names is refused, naming the
computation. An edit that would leave a gap anchor bare without carrying it is refused too: a retire
reads what the edit carried, so no retire could take it. A claim whose sentence is edited away leaves
the other anchors it carries BARE (no prose before them in their segment) or on the prose that
replaced it; `blue retire` names the anchors that stood on the claim in that edit and exit with it —
the bare ones computed by the verb, one on surviving prose when the seat names it, each validated at
the write — and replay removes them at that event. The question "did this leave legitimately?" is
never asked of the anchor: it leaves only on a retire event that carries the claim and its reason. No
kind is held back: a gap whose anchor left reads `gone` on the board, with its minted text, and its
lens judges the report as it stands. `show report --anchor` on an anchor that left this way names the
retire event, not a stale id.

So the integrity check has **no "legitimately withdrawn" exception** to reason about: an anchor
missing without a retire is a hard violation, and what is gated is every *content* transition under
it.

## The handshake — the gap's lifecycle, anchored

The fix-then-approve rhythm is the debate's gap lifecycle: a lens mints a gap, blue repairs, the lens
re-audits and closes. The anchor adds two things to it: a durable locator on the gap's sentence, and
an edit guard that keeps the locator there.

Red-gated at every step:

1. **open** — a lens mints the gap → the tool inserts `<!--gap:<id>-->` at the end of its quote, and
   the record holds the quote as minted with the gap's problem, fix and check.
2. Blue **fixes** the content (edits the anchored sentence; the edit carries the anchor).
3. The lens **approves** — reads the report, compares against the gap as minted, confirms the fix is
   real: not moved to a dead corner, no offending content left behind, not a fake fix. Its `close` is
   the gap's closure.
4. Blue **withdraws** the content by editing it out and retiring the claim, which takes its anchors
   out; the lens **verifies**. The gap reads its own anchor: where that anchor left with a cut, the
   gap reads `gone` and the lens judges the report as it stands.

Blue never resolves unilaterally: it removes an anchor of any kind only by `retire`, and every content
change under one is red-confirmed.

## The report is the record — no seat can drop an anchor

There is no `blue/report.md` file to write. Each lane writes its part once and **records it**
(`report_part`); the report is a PROJECTION replayed from events — the parts in event order with every
change folded over them — and every seat changes it ONLY through the tool's edit path —
**`feov-record edit --quote <exact current span> --new <replacement> --reason <why> [--answers
<gap-id>]`** — an exact-on-stripped span replace (`anchortext.LocateSpan`, the anchor matcher) that
appends a `blue_edit` event to an append-only diff-stack. A raw file write cannot reach the report
because the report is not a file; an anchor cannot be dropped because dropping one is not an
expressible operation.

**Anchors TRANSIT an edit; an edit never creates, destroys or duplicates one.** The tool checks the
multiset of anchor ids in `--quote` against `--new`: blue may carry an anchor across verbatim, and may
not change WHICH anchors exist. A quote may therefore span an anchor — the only context that tells a
repeated word apart is often the anchor red placed on it — and an anchor typed into `--new` is refused
as a duplicate.

`--answers <gap-id>` records which gap the edit responds to, validated against the board like every
other reference. It is the join key between the `required_fix` red asked for and the change blue
actually made; naming a real gap in `--reason` while `--answers` is empty is refused.

The one PreToolUse hook (`feov-record hook pretooluse`) does something orthogonal — it injects
`FEOV_RUN`/`FEOV_AGENT_ID` into a seat's Bash commands. The report a seat reads and the report the
record holds are the same bytes by construction, so no write guard exists.

## Integrity in depth

- **Structural prevention (primary):** the report-as-record model above — a response seat has NO path
  that drops an anchor; the report is not a file, `blue edit` refuses a span that would drop one, and a
  raw write cannot reach the report at all. The render places only recorded anchors, so the anchors
  present and the anchors expected are equal by construction, and no detector compares them: a check
  that can never fire is a plausible zero.
- **Red's semantic read (the real check):** structure cannot see an anchor dragged to a dead zone or
  offending content left behind. The lens reads the report, grounded in the gap as minted, to confirm
  the fix. Prevention blocks the drop; the read catches gaming (an anchor MOVED but present). Neither
  layer suffices alone.

## Assembly — gap anchors are stripped; evidence anchors resolve

A gap anchor is an invisible machine anchor, so at assembly it is simply **stripped** — the tool
removes every `<!--gap:…-->` token from the **final composed report**, not just the report's own text:
the strip runs on the whole output, so a token a gap's location or reason quoted into a record-derived
section is stripped too. Nothing about a gap anchor is reader-facing, so there is nothing to resolve
and nothing to link; the shipped report carries **zero** `<!--gap:`. Open gaps reach the reader
through the record-derived sections — the risk matrix and the board — independent of the anchor
layer.

The "every footnote resolves to a definition, validate no dangling" model is the **evidence axis's**
concern. Assembly composes the `## Bibliography` from the `cite` events, weaves each `<!--cite:-->`
anchor to a `[^N]` and each `<!--computation:-->` anchor to its `[^K<n>]`, and surfaces a dangling
anchor (no source) as an explicit unresolved line. See
[citations-and-bibliography.md](citations-and-bibliography.md).

## Status

- **Built:** the three kinds above, each placed by its verb (quote validated, a mis-quote refused);
  the `anchor` event; the edit guard (carry, put back, refuse a drop or a typed token); `retire` as the
  one exit; the final-output strip of gap anchors.
- **Built (evidence axis):** `fetch` (download-once, cached, hash-verified) + `blue cite` (invisible
  `<!--cite:-->` anchor, a strict cite⟺anchor bijection) + `blue compute` + the assembly weave
  (anchors → visible `[^N]` / `[^K<n>]` + composed `## Bibliography`) + the `unbacked_citations`
  detector. The claim unit is the evidence anchor. See `record-flow.md`.
- **Deferred (own tracks):** the citation **trust-gate** (Slice 2, #247 — Iffy blocklist, Crossref
  DOI/retraction, provenance); 1c shared-labels; 1d value-transclusion. `grep` remains the net for
  bare figures pending 1d.
