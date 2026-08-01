# Propagation and anchoring — the engine's claim-tracking model

> **Purpose.** When blue corrects a claim, the correction must reach *every* site that states the
> same claim, or the report ships an internal contradiction. Incomplete propagation was run 3's
> dominant blue failure class (5 regressions in 5 rounds). This note is the durable model for how
> the engine tracks a claim across a mutating report — the design that governs the (c) worklist /
> claim-index work and everything after it. It describes a **destination and the gated path to it**;
> the Status section says what is built versus proposed.

## The reframe: this is a DRY problem, not a search problem

The instinct is to make blue *search* the report better — a claim-index, a report-wide grep. But
search only exists because the same claim is **copied as prose to N sites with no link between the
copies**. The report violates DRY, and every propagation sweep is the cost of that violation. The
real lever is not a faster search over duplicates; it is to **give the copies a shared identity, or
collapse them to one source**, so that "find the other sites" stops being a search and becomes
"follow the pointer." Search is removed by removing duplication, not by out-running it.

This is the repository's own governing principle — *the record is the single source; projections
derive on read* — pushed down to the claim level.

## The three tiers of restatement

A correction has to propagate across three kinds of duplicate, and they do not cost the same:

| tier | the duplicate | what catches it | cost |
|---|---|---|---|
| **T1** | the same figure/string repeated verbatim | a `grep` for the value | deterministic, ~free |
| **T2** | a footnoted claim asserted at several marker sites | the claim-index (enumerate the marker) | deterministic, cheap |
| **T3** | the claim **restated in different words**, unfootnoted | only a mind reading the report | the real cost |

T3 is the irreducible core. No marker index and no string sweep can see a claim paraphrased at a
site that shares neither its words nor its footnote. The mistake to avoid — the one the plan-audit
gate caught twice — is letting a mechanism that covers T1 or T2 **silently redefine completeness**
down to its own tier while the guarantee still reads as "all sites." A test that pins
marker-completeness passes green while the T3 guarantee quietly weakens.

## The anchoring primitive: mark at creation, never recover by search

A bookmark into prose dies under editing. Storing a claim's *quote* and re-finding it later (W3C
`TextQuoteSelector` / robust anchoring) fails precisely under blue's workload, because blue's job is
rewriting the sentences the quote is made of — the thing you match against is exactly the thing that
changes.

The mature systems (Google Docs comments, editor bookmarks, LSP anchors) never re-find an anchor by
matching text. They **stamp a first-class anchor object at the moment of creation** and maintain it
through edits. The engine's version, compatible with a **read-only red**:

- **Tool-inserted marker at citation.** When red cites a claim, the tool has identified the exact
  span *once, at that moment of certainty*. It inserts a durable, **invisible** marker there
  (an HTML-comment-style token, `<!--#c17-->`, that markdown ignores). It is the **tool** that
  writes, not red — red stays read-only; red's `cite` event is the trigger.
- **Invisible and assembly-stripped.** Blue edits *around* the token; assembly removes it from the
  final report. Blue's authored prose is never touched, so the "blue owns the visible report"
  invariant holds. The marker is a parallel anchor layer embedded in the artifact.
- **Locating is `grep`, not comprehension.** From insertion on, "where is claim c17?" is
  `grep <!--#c17-->` — O(hits), and it **survives blue's rewrites because the token rides the
  sentence** it is embedded in.
- **Death-on-deletion is correct, and it composes with the retire-detector.** If blue deletes the
  marked sentence the anchor goes with it — which is exactly what should happen: a claim leaves the
  report only through the `retire` verb, never by silently vanishing. A marker that disappears with
  no matching retire event is a detector hit, not a lost bookmark.
- **Robust anchoring is the recovery fallback, not the primary.** If a marker is lost (blue corrupts
  the raw token), re-anchor by last-known quote+context — belt and suspenders, never the main path.
- **The marker layer is a projection, not a forked source.** Each insertion is the image of a `cite`
  event on the text (`red cited span S → tool inserted marker M`). Replay the report edits plus the
  cite events and the anchored document reconstructs. It passes the render-shadow test.

**Sequencing is safe.** Blue drafts → red cites (the tool stamps markers) → blue responds (edits
around them). Turn boundaries, not concurrent writes to one file. Round-0 blue is multi-lane, but
that is *before* red cites anything.

## Two complementary anchor sources

Markers are not the whole answer, and it would be over-selling to say they are:

- **Cite-markers** anchor the claims red **actually cited** — the audit trail — and hand read-only
  red its re-location for free (`locate(c17)` next round instead of re-reading the report).
- **Shared footnote labels** anchor **blue's co-referent authoring**: when blue asserts one claim at
  several prose sites, the sites share **one** `[^label]` (today footnotes are per-sentence-unique;
  the trick is reusing the label). Propagating that claim is then `grep [^label]` — T3 converted to
  T2 **by authoring discipline at write time**.

Together they cover both directions — the claims under audit and the claims under authoring. The
residue is a genuinely new restatement blue wrote *after* cite-time and never labelled: the
irreducible T3 floor.

## Value-transclusion: kill the figure-consistency class at the root

Recurring **values** (figures, counts, named quantities) need no anchoring at all — single-source
them. Blue defines a value once as a record-backed variable; the report references it; assembly
expands it. Correcting the value is one edit, zero search, zero propagation — the report **cannot**
carry an inconsistent figure because there is one source. Run 3's regressions were dominantly
figure/consistency; this removes that class outright. Prose cannot be transcluded without making the
report unreadable — but values can, and they are where the consistency bugs concentrate.

## The repair loop: edit from context, confirm, iterate — red is the outer net

Propagation is a repair *loop*, not a per-correction sweep:

1. The working set is read **once** at cycle start (already batched) and is in blue's context.
2. Blue mutates from that in-context copy — every site it can — in one pass. No re-ingest.
3. **Confirmation pass, in two halves:**
   - *Deterministic:* `grep` for the **stale** string (the old value/phrase just replaced) → zero
     hits proves T1 propagation. Grepping for the value being *eliminated* is a stronger check than
     grepping for the new one — it proves absence of the thing you are removing. Plus the
     claim-index / marker grep for T2.
   - *Comprehending:* blue re-reads for T3 — the paraphrased sites no scan can see. This is the only
     step that needs a mind, and it is where the budget should go.
4. Loop until blue's own review is clean.

The inner loop is best-effort convergence; it does not have to be perfect. **Red's next-round
re-audit is the guarantee**, and the round cap terminates. Blue's self-review exists to drive
residue *down* so red is not the one discovering all five regressions.

Note the read economy: the "~1 read per cycle" target counts the **working-set context ingest**, not
the propagation `grep` — the grep is a cheap scan of blue's own report, not a context read. Keeping
the report-wide stale-string sweep costs nothing against that target.

## The living document, and the gate for going there

If the report were a hosted, MCP-served entity, it could keep the bookmarks alive as live metadata —
the document becomes the authority on its own positions, every edit shifts the anchors after it
transactionally, and "where is c17" is a lookup the server already holds (the CRDT / editor-marker
model). That is a legitimate **destination**. The line that decides whether it is sound:

> **The live document must be a read-model *over* the append-only record — never a new source of
> truth.** The test is the render-shadow test verbatim: *kill the server; can the anchor map be
> rebuilt by replaying the event log?* If yes, it is a disposable cache with live indexes —
> legitimate. If the server's memory is the only copy of where the claims are, the truth has forked
> into a mutable process, and it is the materialization the render-shadow arc removed, reborn.

The clarifying consequence: **"keep the bookmarks alive" and "re-derive them on read" are the same
read-model — live-versus-recompute is only caching.** The architecture does not change; only *when
you pay for the index* does. Its real cost is behavioral: for a server to keep anchors live, every
blue edit must become a **recorded operation through the server**, ending free-hand markdown editing.
That is the price to weigh, and it is why the server is the destination, not the first step.

## The gated rollout

Each promotion is gated by **measurement** (the cheaper tier proved insufficient) and by the
**render-shadow test** (the new state stays a projection of the record):

1. **Now — stateless.** Shared labels + value-transclusion + cite-markers + the confirmation loop.
   No process, no source-of-truth fork, most of the win. The claim-index is recompute-on-read.
2. **Recorded edits.** Blue's edits become recorded operations, so the anchor map is maintained
   incrementally instead of recomputed. Promote only if measurement shows recompute is the cost.
3. **Live read-model.** The document served as a process with a maintained in-memory index (and
   read-only `locate` for red). Promote only if the maintained index must be *live* rather than
   recomputed per query — and only while it stays rebuildable by replay.

Do not cache until the derivation is proven expensive. That is the render-shadow discipline.

## Status

- **Built:** none of the anchoring model yet. Prerequisites landed — board de-dup (#225), retire
  detector (#226), undisposed metric (#227).
- **Specced (the (c) work):** the worklist view, `merge near-match`, the claim-index (recompute-on-
  read, footnote-marker enumeration reconciled with `claimcount.Count`), `mint --existence`. The (c)
  spec's Phase 3 is **tier-1 of this model** — the confirmation loop with the retained report-wide
  stale-string sweep as the T3 backstop.
- **Proposed, not specced:** cite-markers, shared-label semantics, value-transclusion, recorded
  edits, the live read-model. Each is a slice under this model; each earns its promotion by the two
  gates above.

## Prior art

- **W3C Web Annotation** robust anchoring (`TextQuoteSelector` + `TextPositionSelector` + prefix/
  suffix) — the canonical bookmark-survives-edits mechanism; here it is the *recovery fallback*.
- **CRDT relative positions** (Yjs `RelativePosition`, Automerge cursors) — anchors that auto-move
  with edits, but require tool-mediated editing (the tier-3 cost).
- **Transclusion / single-source** (BibTeX `\cite`, spreadsheet cell references, DRY) — model for
  value-transclusion.
- **Event sourcing + materialized read-model** (append-only log + maintained secondary index) — the
  shape the "living document" actually is.
