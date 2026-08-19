# Finding-markers — the immortal audit anchor

> **Purpose.** Red's audit points must not be able to silently disappear from blue's report, and blue's
> "fix" must be verifiable against what red actually flagged. This note is the design-of-record for how
> the tool anchors a red finding in `blue/report.md` with a durable, immortal marker that carries the
> finding's identity and a snapshot of what it flagged. It is the concrete realization of slice 1b of
> [propagation-and-anchoring.md](propagation-and-anchoring.md) — reframed, after design, from "cite the
> claim" to "anchor the finding," because a finding already carries the id, the offending content, and
> the reason the anchoring needs.

## Two axes, DIFFERENT mechanisms

A finding-anchor and a citation are NOT the same mechanism — an earlier draft of this note tried to
unify them under "footnotes, one namespace," and a live whole-chain run proved that wrong (see *Why
invisible*, below). They are two axes with two forms:

- **Findings → INVISIBLE in-doc anchors** (this note). Red anchors each audit point with an invisible
  HTML-comment token; the anchor is a machine-readable locator, NOT reader-facing content; it can never
  silently leave; every content change under it is red-gated.
- **Citations → the bibliography system** (built in 0.28.0/#256 — see `record-flow.md`). A citation is
  *also* a tool-inserted **invisible immortal** anchor — `<!--cite:c-…-->`, splice-identical to a
  finding marker (it reuses `lens.InsertAnchor`) — but with the OPPOSITE assembly fate: where a finding
  is STRIPPED, a citation is **RESOLVED**. `blue cite` fetches the source **once** to a hash-addressed
  cache (`<run>/cache/<sha256>`, so red re-reads the exact bytes blue cited), then anchors the sentence;
  assembly weaves every anchor into a VISIBLE `[^N]` footnote + a composed `## Bibliography`. This is
  where the visible-footnote / resolve-at-assembly machinery belongs — NOT on findings.

## The marker

- Format `<!--fx:<finding_id>-->` — an **invisible HTML comment**. It renders to nothing, is not a
  footnote (so no seat reads it as content or audits it as an undefined reference), and it cannot touch
  `claim_count` at all — a comment is never a `[^…]`, so no namespace rule or `Count` exclusion is
  needed. It is a pure machine anchor.
- `finding_id` is the finding event's **already-minted unique id** (`f-<hex>`) — findings get one today.
  The marker CONTENT is that id, a pointer ref; construction and grep-extraction agree on it verbatim.
- The finding RECORD snapshots **{offending-content copy, reason}** at creation. That snapshot is what
  makes red's re-audit independent of blue: red re-locates and re-checks against the original, and
  re-finds the problem even if blue tried to bury it. It doubles as the robust-anchor recovery quote if
  a marker token is ever lost.
- The detector's PRESENT set is the finding ids grepped from the `<!--fx:…-->` tokens in the current
  report; EXPECTED is the `anchor` events. (A finding's markers are its sites, so the same tokens also
  support propagation — but as a grep over the comment layer, not via the `[^label]` claim-index.)

## Why invisible (the live-run lesson)

The visible `[^f-<id>]` footnote form failed a live whole-chain run: red READ the report, saw the
finding-markers as **undefined footnote references** (no `[^f-<id>]:` definition existed), and filed a
finding auditing them — the tool's own anchors became a debate defect. Worse, red's finding TEXT quoted
the markers, so they rode the record-derived findings/risk sections into the shipped report even though
assembly stripped blue's authored content. Both failure modes trace to one cause: **a machine anchor
was dressed as reader-facing content.** The fix is the invisible comment form here; the visible-footnote
+ definition-resolution model is correct only for the CITATION axis, where a reference genuinely IS
reader-facing content that resolves to a source.

## Immortality + the tampering guarantee

The marker is **immortal in the working doc** (`report.md` during the run): neither side deletes it. So
the integrity check has **no "legitimately withdrawn" exception** to reason about — **any missing
marker is blue tampering with red's record, a hard violation, full stop.** What is gated instead is
every *content* transition under a marker. (This is the property that dissolves the identity-vs-retire
collision an id-plus-retire model runs into: markers don't leave, so "did this leave legitimately?" is
never asked of the marker.)

## The handshake — layered onto the existing finding → gap → closure lifecycle

The fix-then-approve rhythm is **not new machinery**; it is the debate's existing lifecycle (red flags
a finding → merge coalesces into a gap → blue repairs → red re-audits and closes). Finding-markers add
three things to that lifecycle, nothing more:

1. the immortal marker anchor on the finding's content,
2. the {content, reason} snapshot, and
3. the tampering detector.

Red-gated at every step, as the lifecycle already is:

1. **open** — red files the finding → the tool inserts `<!--fx:<id>-->` at the offending content and
   the record snapshots {content, reason}.
2. Blue **fixes** the content (edits the marked sentence; it cannot touch the marker).
3. Red **approves** — reads the doc, compares against the snapshot, confirms the fix is real: not moved
   to a dead corner, no offending content left behind, not a fake fix. (This is red's gap closure.)
4. Blue **withdraws** the content; red **verifies**. End state, in the working doc: **the marker
   remains, the content is gone.**

Blue never resolves unilaterally and never removes a marker; "blue can't remove what red verified
without red's assent" holds because the marker is immortal and every content change is red-confirmed.

## The write-time lockdown (0.27.0) — blue cannot drop a marker in the first place

Immortality was prompt-hoped in slice 1b (blue was TOLD to grep `<!--fx:` and never delete one); a live
run showed that is not enough. It is now **enforced at the write**. `blue/report.md` is **read-only to
every seat except the allowlisted author `agent_type`** (`frank-exchange-of-views:blue-synthesizer`, the
round-0 synthesis seat). A response seat changes the report ONLY through **`feov-record edit --quote
<exact current span> --new <replacement> --reason <why> [--answers <gap-id>]`** — an exact-on-stripped span replace (reusing
`lens.LocateSpan`, the finding-marker matcher) that appends a `blue_edit` event, an append-only
diff-stack replaying onto the round-0 report.

**Anchors TRANSIT an edit; they are never created, destroyed or duplicated by one.** The rule was once
"reject any span containing a marker — edit around it", and that produced a DEADLOCK (demonstrated,
0.28.x): where a word appears twice and the only disambiguating context carries red's anchor, the minimal
quote is refused as ambiguous and the contextual quote is refused as anchor-spanning, so the anchored
occurrence — the one red actually flagged — became the only uneditable one. 71% of anchored quotes in the
2026-08-04 smoke had their anchor mid-span, so that was the common shape, not a corner. What the tool
checks now is the multiset of anchor ids in `--quote` against `--new`: blue may carry an anchor across
verbatim, and may not change WHICH anchors exist. That is strictly stronger than the old rule, which said
nothing about `--new` at all — 5 smoke edits smuggled an anchor INTO `--new` and duplicated it.

`--answers <gap-id>` records which gap the edit responds to (0.29.0, #267), validated against the board
like every other reference. It is the join key between the `required_fix` red asked for and the change
blue actually made; naming a real gap in `--reason` while `--answers` is empty is refused, so the key
cannot decay back into the 73%-reliable convention it replaced.

The READ-ONLY half is enforced by a plugin PreToolUse/PostToolUse hook (`feov-record hook
pretooluse|posttooluse`) that reads `agent_type` from the hook stdin:
- **PreToolUse (prevent):** a non-author raw write to `blue/report.md` is DENIED — Edit/Write by
  `file_path` (airtight), common Bash by target-path resolution (`cp`/`mv`/`tee`/`>`/`sed -i`).
- **PostToolUse (detect + force-repair):** after a non-author call, a dropped marker (via any mechanism,
  including exotic Bash the PreToolUse verb-set cannot enumerate) blocks and forces the seat to restore it
  — reusing the same `claimcount.MissingAnchorIDs` check as the scorecard detector.
The tool's own writes (`record.MutateBlueReport`, in-process) bypass the hook, so `lens finding` and
`blue edit` still anchor/amend under the lock.

## Integrity in depth

- **Write-time prevention (primary):** the lockdown above — a response seat has NO path that drops a
  marker; `blue edit` refuses the span, and a raw write is denied.
- **Mechanical screen (cheap):** immortal markers ⇒ the present set MUST contain the expected set; a
  missing marker = tampering. Pure id set-membership, no text matching (the scorecard's
  `dropped_finding_markers`, and the PostToolUse backstop, share `claimcount.MissingAnchorIDs`).
- **Red's semantic read (the real check):** the screen cannot see a marker dragged to a dead zone or
  offending content left behind. Red reads the doc, grounded in the {content, reason} snapshot, to
  confirm the fix. Prevention blocks the drop; the screen catches a drop that slipped an exotic path; the
  read catches gaming (a marker MOVED but present). No single layer suffices.

## Assembly — findings are STRIPPED; resolution belongs to the citation axis

Finding-markers are invisible machine anchors, so at assembly they are simply **stripped** — the tool
removes every `<!--fx:…-->` token from the **final composed report** (not just blue's lifted content:
the strip runs on the whole `out`, so a marker that a finding's location/reason text quoted into a
record-derived findings/transcript section is stripped too — that record-derived path was the leak the
live run exposed). Nothing about a finding-marker is reader-facing, so there is nothing to resolve and
nothing to link; the shipped report simply carries **zero** `<!--fx:`.

Unresolved findings remain **surfaced** to the reader the way they always were — via the record-derived
findings/risk sections — independent of the marker layer.

The "every footnote resolves to a definition (in-doc or bibliography), validate no dangling" model is a
real design, and it is the **citation axis's** concern (a citation IS reader-facing content that
resolves to a source). It is now BUILT there (0.28.0/#256): assembly composes the `## Bibliography` from
the `cite` events, weaves each `<!--cite:-->` anchor to a `[^N]`, and surfaces a dangling anchor (no
source) as an explicit unresolved line — none of which applies to the invisible, stripped
finding-anchors. *(Slice 1b ships the finding integrity core + strip; the citation weave is its
resolve-at-assembly twin.)*

## Status

- **Built (integrity core):** invisible `<!--fx:…-->` finding-markers inserted by `lens finding`
  (validate-quote-exists → reject a mis-quote; insert via the locked, atomic `record.MutateBlueReport`);
  the `anchor` event; the `dropped_finding_markers` detector; the final-output strip. Validated live
  against a whole-chain run (markers inserted, no over-rejection, blue-preserved across rounds) — which
  is what caught the visible-footnote flaw this note now records (*Why invisible*).
- **Superseded drafts:** the `cite`-based draft (no unique id; `retire` collided with the marker
  lifecycle — both dissolved by anchoring *findings*) and the visible-`[^f-]`-footnote draft (audited as
  undefined footnotes + leaked; dissolved by the invisible comment form).
- **Built (citation axis, 0.28.0/#256):** the bibliography/citation system — `fetch` (download-once,
  cached, hash-verified) + `blue cite` (invisible immortal `<!--cite:-->` anchor, class-swept into the
  blue-edit lockdown so it is a strict cite⟺anchor bijection) + the assembly weave (anchors → visible
  `[^N]` + composed `## Bibliography`) + the `unbacked_citations` detector. The claim unit is now the
  cite anchor (the hand-typed `[^label]` footnote is retired). See `record-flow.md`.
- **Deferred (own tracks):** the citation **trust-gate** (Slice 2, #247 — Iffy blocklist, Crossref
  DOI/retraction, provenance); 1c shared-labels; 1d value-transclusion. `grep` remains only the
  transitional net for bare figures pending 1d.
