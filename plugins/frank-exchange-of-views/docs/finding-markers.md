# Finding-markers — the immortal audit anchor

> **Purpose.** Red's audit points must not be able to silently disappear from blue's report, and blue's
> "fix" must be verifiable against what red actually flagged. This note is the design-of-record for how
> the tool anchors a red finding in `blue/report.md` with a durable, immortal marker that carries the
> finding's identity and a snapshot of what it flagged. It is the concrete realization of slice 1b of
> [propagation-and-anchoring.md](propagation-and-anchoring.md) — reframed, after design, from "cite the
> claim" to "anchor the finding," because a finding already carries the id, the offending content, and
> the reason the anchoring needs.

## Two axes, cleanly split

A marker is a marker — one mechanism (footnotes), two purposes distinguished by namespace:

- **Findings → immortal in-doc markers** (this note). Red anchors each audit point; the marker can
  never silently leave; every content change under it is red-gated; the footnote resolves to the
  in-doc argument.
- **Citations → the bibliography system** (a separate, later feature). URL → download **once** → local
  cache → **hash-verified-unchanging** → footnote-by-reference; the footnote resolves to an external
  source in a bibliography section. Not built here; it plugs into the same assembly resolution
  framework (below).

## The marker

- Format `[^f<finding_id>]` — a footnote in a reserved **`f` namespace**, distinct from blue's claim
  footnotes (`[^L1]`). `claimcount.Count` **excludes** the `f` namespace, so `claim_count` is untouched
  (a finding-marker is a locator, not a footnoted-claim unit).
- `finding_id` is the finding event's **already-minted unique id** — findings get one today. This is
  the real per-item identity the anchoring rests on; the marker CONTENT is that id, a pointer ref.
- The finding RECORD snapshots **{offending-content copy, reason}** at creation. That snapshot is what
  makes red's re-audit independent of blue: red re-locates and re-checks against the original, and
  re-finds the problem even if blue tried to bury it. It doubles as the robust-anchor recovery quote if
  a marker token is ever lost.
- `claim-index` enumerates `[^f…]` markers as occurrences (typed as finding-anchors, distinct from
  claim occurrences), so a finding's markers are its sites — integrity **and** propagation from one
  mechanism.

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

1. **open** — red files the finding → the tool inserts `[^f<id>]` at the offending content and the
   record snapshots {content, reason}.
2. Blue **fixes** the content (edits the marked sentence; it cannot touch the marker).
3. Red **approves** — reads the doc, compares against the snapshot, confirms the fix is real: not moved
   to a dead corner, no offending content left behind, not a fake fix. (This is red's gap closure.)
4. Blue **withdraws** the content; red **verifies**. End state, in the working doc: **the marker
   remains, the content is gone.**

Blue never resolves unilaterally and never removes a marker; "blue can't remove what red verified
without red's assent" holds because the marker is immortal and every content change is red-confirmed.

## Two-layer integrity

- **Mechanical screen (cheap):** immortal markers ⇒ the present set MUST contain the expected set; a
  missing marker = tampering. Pure id set-membership, no text matching.
- **Red's semantic read (the real check):** the screen cannot see a marker dragged to a dead zone or
  offending content left behind. Red reads the doc, grounded in the {content, reason} snapshot, to
  confirm the fix. The screen catches silent drops; the read catches gaming. Neither alone suffices.

## Assembly — every footnote resolves; there is always somewhere to point

Footnote **definitions are never authored** in blue's raw report — the inline `[^…]` refs sit
"invalid-until-assembled," and **assembly constructs every definition**. A report validator checks that
each ref resolves — no dangling footnote. There is always a target, of one of two kinds:

- **Bibliography (external):** a citation resolves to a `[^ref]: <source>` entry in a bibliography
  section built at the end of the doc (the citation axis).
- **In-doc argument (internal):** a finding's footnote resolves to an in-doc anchor to the exposed
  argument / adjudication material.

By resolution state:

- **Resolved** findings: the reference resolves away — dropped from the shipped report; no
  resolved-audit noise.
- **Unresolved** findings (including an early **abort**): the argument is **exposed** in the doc and
  the footnote links to that in-doc anchor — the reader sees what is still contested and can follow it
  to the adjudication. This is why immortal markers cost nothing at ship: an unresolved marker always
  has an argument to point at; a resolved one is dropped.
- Raw `[^f…]` marker tokens never ship as tokens — assembly resolves them into proper refs+definitions
  or drops them; no raw `[^f` leaks.

The resolution framework (construct definitions, validate no dangling) is **shared** by both axes. 1b
builds the in-doc-argument resolution for findings; the external-bibliography resolution is the
deferred citation feature, plugging into the same framework.

## Status

- **Proposed, not built:** everything in this note. It supersedes the earlier `cite`-based 1b draft
  (which failed plan-audit because `cite` events carry no unique id and `retire` collided with the
  marker lifecycle — both dissolved by anchoring *findings*, which already have the id, and by immortal
  markers, which have no retire lifecycle).
- **Prerequisites (shipped):** the claim-index + `f`-namespace-ready `claimcount` (1a, #240); the
  finding id (`findingid.go`).
- **Deferred (own tracks):** the bibliography/citation system (download-once, cached, hash-verified,
  programmatic definitions, validator); 1c shared-labels; 1d value-transclusion. `grep` remains only
  the transitional net for bare figures pending 1d.
