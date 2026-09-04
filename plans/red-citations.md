# Red's evidence reaches the reader

> STATUS 2026-09-02: shipped — historical record (carriers verified in-tree: `Verify.Label` in recordpb, `CitationLabelsOf`, `ExistingCorroborationLabel`, `LocateUniqueReplacing`, the PASS refusal over unanswered contradictions at `record/refs.go:414`; the §V.7 mutate sweep remains the one open item)

## I. Summary & goals

**No `Verify` event reaches the report.** Neither `lens verify` (a citation red audited) nor
`lens corroborate` (a source red found itself) has any reader in `internal/report`. Their only
consumer is `internal/record/evidenceview.go`, the seat-facing `evidence` projection.

So blue cites and gets a footnote; red audits those citations, finds its own sources, and a
reader of the document sees none of it. The strongest thing an adversarial process produces —
a claim confirmed by a source its author never chose — dies in a projection.

This is not new. `citationid.go` says it outright: *"Red's `lens cite` carries no label and is
EXCLUDED, so this is exactly the tool-inserted citation set."* Red's citations were deliberately
kept out, and #341 then split the shared event into `verify` / `corroborate`.

**The operator's call (2026-08-22):** a human reader only cares that the text has appropriate
references. There is no reason for the reader to see that two different teams inserted
footnotes. Red's supporting corroborations become footnotes like any other.

**Goals**

1. A supporting corroboration renders as an ordinary footnote in the assembled bibliography.
2. Its record key becomes a MINTED label — unique by construction — so one source may
   corroborate many claims. This is the collision that started the thread.
3. The negative outcomes reach the board, not the bibliography.
4. `verify` and `corroborate` stay distinct acts on the record (#341 is not undone).

## II. Design

### The bibliography is already team-agnostic

`weaveCitations(md, sources)` scans the assembled report for `<!--cite:c-…-->` tokens, numbers
them in order of appearance, and renders `## Bibliography` from `record.Source` entries matched
BY LABEL. It does not know or care who wrote the token. **So the supports case needs no report
change at all** — it needs red's corroboration to (a) carry a label, (b) splice the token, and
(c) appear in the source list.

### The key falls out for free

`keyFields` is `["gap_id", "label", "id", "observation", "anchor", "url"]`, first match wins.

| verb | field that matches first | key | effect |
|---|---|---|---|
| `blue cite` | `label` | the minted `c-<hex>` | one url may be cited many times |
| `lens verify` | `anchor` | the citation checked | correct — one anchor sits at one sentence |
| `lens corroborate` | `url` | **the source** | **one claim per source per sitting** |

Giving `Verify` a `label` field, set only on the corroborate path, moves corroborate's key from
`url` to `label` without touching `keyFields` at all. The composite-key change considered
earlier is unnecessary — which is the better answer, since `keyFields`' own comment warns that
a key not derivable from one field is how the list goes stale silently.

Crash-retry idempotency stays where blue's is: on `--key`, a stable label the seat supplies.

### The negatives go to the board, not the bibliography

`corroborate --as` takes `supports`, `supports_with_bridge`, `weak`, `refutes`, `absent`,
`unreachable`. A source that CONTRADICTS blue's sentence is not a reference backing it; spliced
as a footnote it reads as support, and the report's own assembly check already treats a live
refuted citation as a failure.

- `supports` / `supports_with_bridge` / `weak` → mint a label, splice, render as a footnote.

  **`weak` was excluded and that was wrong** (corrected 2026-08-23, operator's call). The argument
  is symmetry, not strength: when BLUE cites a source and red grades it `weak` through
  `lens verify`, the footnote STAYS — verify adjudicates and never touches the report. Excluding a
  weak corroboration rendered the same (claim, source, grade) triple as a footnote when blue found
  the source and as NOTHING when red did, which is the one difference a reader must not be able to
  see. It also left the reading in a silent middle state: neither cited nor owing a finding,
  visible only in the evidence projection. A footnote is a POINTER, not an endorsement — red's
  judgement of how well the source bears is preserved where judgements live.
- `refutes` / `absent` / `weak` → red finding the text unsupported IS a defect in the text.
  **A LENS CANNOT MINT** — that is structural, not an oversight ("A lens structurally cannot
  mint or close a board gap: no such verb exists in its namespace"). So the carrier is
  `lens finding`, which the merge raises to a gap through `found_by`. The negative
  corroboration records as evidence AND the seat owes a finding; the tool says so.
- `unreachable` is the honest exclusion: red could not read the source, so there is nothing to
  point a reader at. Record-only.

**SETTLED, on evidence rather than preference.** `Finding` carries severity, likelihood and
impact, and `lens finding` demands all three — so the tool writing one on red's behalf would
mean INVENTING three grades nobody chose, feeding the mass calculation that decides what a gap
is worth. A fabricated grade reads exactly like a judged one. So the duty is REPORTED, not
discharged: `UnansweredContradictions` blocks a PASS over a `refutes` or `absent` reading with
no finding quoting that claim, naming the claim and naming the act that clears it. Red grades
its own finding; the merge decides whether to raise it.

The match is deliberately loose — any lens's finding quoting the claim answers it. A stricter
join (same seat, same round) would refuse a contradiction one lens found and another raised,
which is the collaboration the lens roles exist for.

## III. Carriers and consumers

- `recordpb.Verify` — new `label` field (reserved numbering: next free).
- `internal/cli/lens/verify.go` — `corroborate` mints, splices, and branches on outcome.
- `internal/record/citationid.go` — `CitedSources` must include labelled corroborations; its
  comment about red being excluded is the thing being changed and must be rewritten, not left.
- The `unbacked_citations` detector — **and this is where the carrier census earned its keep.**
  It did NOT read `CitedSources`; it built its EXPECTED set with its own inline loop over `Cite`
  events, a THIRD copy of the rule. Widening the other two left it on the old one, so blue
  dropping a RED citation anchor would be caught by the hookgate lockdown and MISSED here — two
  detectors for one protection, disagreeing, neither able to say so. Collapsed onto
  `CitationLabelsOf`.
- `hookgate.DefaultAnchorIDs` reads `CitationLabels`, so **red's citation anchors are IMMORTAL
  now**, protected by the PostToolUse lockdown exactly as blue's are. Before this they were
  spliced into the report and guarded by nothing. A gain that came with the change rather than
  one it was designed for.
- `requireCitation`'s refusal said "blue has cited %d source(s)" — true before, false after.
- `internal/record/evidenceview.go` — the `evidence` projection still shows both, and gained
  `unanswered_contradictions` plus the minted `label`. The duty is ENFORCED at the merge's PASS
  gate, which is the right place to refuse; but a duty that surfaces only when someone else is
  blocked at the end of the round is one the owing seat never sees. The empty array matters as
  much as a full one — without the field, "nothing outstanding" and "nobody checked" are the
  same absence.
- Goldens: `internal/difftest` (help + error catalogue), `recordsql/testdata/schema.sql`,
  the simulator's prompt goldens if any prompt text moves.
- `internal/fuzz` — drive corroborate on both arms, or the split is unexercised.

## IV. Risks

- ~~**Splicing into blue's document from a red seat.**~~ **NOT A RISK, and calling it "the
  sharpest edge in this change" was wrong.** `lens finding` ALREADY splices invisible anchors
  into `blue/report.md`, and `InsertAnchor` lives in the lens package, documented as "the shared
  invisible-anchor placement behind lens finding and blue cite". A lens placing a tool-managed
  marker is existing, deliberate behaviour. The boundary the system protects is red editing
  blue's PROSE, and there is still no lens edit verb.
- **Minting kills idempotency, and this one WAS real.** A fresh label every call means nothing
  collides, so a crash-retry splices a second anchor and records a second event — precisely the
  cost that made "drop `url` from keyFields" the wrong answer, reached by another route. Caught
  by an existing test that asserted the retry was refused. `ExistingCorroborationLabel` restores
  it on (source, claim), which is the act itself, rather than on a seat-supplied `--key`: a
  retry should not need the seat to have anticipated it.
- **The `--quote` span must exist in the live report**, as blue's cite requires. A corroboration
  of a claim blue has since edited away must be refused, not spliced blind.
- **Double-counting.** `citations_checked` counts verify events; the bibliography counts labels.
  A corroboration now appears in both. Confirm no gate adds them.

## V. Verification plan

All paths absolute; `export PATH=$PATH:/usr/local/go/bin`; `GOTOOLCHAIN=go1.25.0`.

1. `go build ./...` and `go vet ./...` in `tools` → clean.
2. New tests, each written before its change:
   - a supporting corroboration renders in `## Bibliography` (assemble-level).
   - the same source corroborates TWO claims and both record — the collision that started this.
   - a `refutes` corroboration does NOT splice an anchor.
   - a corroboration whose quote is absent from the live report is refused.
   - `CitedSources` includes labelled corroborations and still excludes unlabelled events.
3. `go test ./internal/record/... ./internal/cli/... ./internal/report/` → green.
4. `go test -timeout 900s ./internal/fuzz/` → 0/60, no unreached flags. **Never concurrently
   with `./...`.**
5. Goldens regenerated with `UPDATE_GOLDENS=1` (difftest, recordsql) and
   `cd scripts && go run ./golden -update` (prompts); every diff read and justified in the
   commit that regenerates it.
6. `cd tests/simulator && node --test` → 93+ pass, 0 fail.
7. Still owed from the parent plan: `cd scripts && go run ./mutate` — attempted 2026-08-22 and
   INTERRUPTED before producing results (it restored `record.go` cleanly).
   **RUN 2026-09-03, PARTLY. The result is a split, not a number.**
   - `-filter internal/claimcount` → 1 survivor / 2 behavioural. EXPLAINED and benign:
     `j >= 0` → `j > 0` on a `strings.Index` result, where a `-->` closer at index 0 cannot
     occur because content precedes it. Equivalent mutant.
   - `-filter internal/anchor` → 19 survivors / 21 behavioural. **This is a test-coverage
     finding, not a weak-assertion one, and the number would mislead without its cause:**
     the package holds only `window_test.go`, so `anchor.go` — which places the citation and
     finding anchors, 13 of the 19 survivors — has NO package-local test at all. The narrow
     stage runs a mutant against its own package only, so what survived is what nothing in
     `internal/anchor` tests; `internal/record` and `internal/cli` exercise it from outside.
     `-confirm` would settle which, at ~8 minutes per survivor (~2.5 hours here) — not paid.
     **The gap is real either way: anchor.go's own package asserts nothing about it.**
   - `internal/record` (`citationid.go`, `refs.go`) — SLOW, and **running**. A first 50-minute
     attempt produced no output, which I initially recorded as "zero mutants" and as
     INFEASIBLE-AS-BUILT. **That was wrong, and the way it was wrong is the point: a KILLED mutant
     prints nothing.** An empty log is what a working sweep and a broken one both look like, so
     silence was read as failure on no evidence — the plausible zero, committed to a plan by the
     agent citing the rule against it.
     Measured instead, 2026-09-04: `-selftest` passes (the tool mutates and observes); the sweep
     runs in a SANDBOX COPY (`mutate/sandbox.go`), so polling the real tree for writes measures
     nothing and my first probe did exactly that; polling the sandbox showed **1 mutation per
     ~150s** under concurrent load. So the earlier silent 50 minutes was roughly 20 mutants, all
     killed. Feasible, just slow — and the rate is a property of the load and of
     `internal/record`'s ~60s suite, not of the tool.
   - `internal/cli` (`blue/cite.go`, `lens/anchor.go`) — not attempted, same reason, worse: that
     package's suite runs 20+ minutes once.

**Result, 2026-08-22.** All of 1–6 green: tools 37 packages / 0 failures, fuzz 0/60 with the new
gate and its remedy both driven, simulator 93 pass / 0 fail. The difftest goldens did not move
(the help text did not change and the catalogue's probes do not reach the new refusal); the
schema golden gained exactly one line, `"label" TEXT`. `qlty` is absent on this box, so the
format/lint gate is SKIPPED rather than passed.

## VI. The anchor model, audited (2026-08-23)

Adding red's citations made a pre-existing question urgent: what happens to an anchor when the
text under it changes? The audit answered it in three parts.

**The no-loss promise HOLDS.** All five paths that mutate `blue/report.md` funnel through
`MutateBlueReport`; `AnchorsTransitUnchanged` refuses a replacement that drops an anchor and
`droppedMarker` backstops the whole edit. I initially called this a hole. It is not.

**The transit was never DELIBERATE, which is the part that was broken.** `InsertAnchor` places a
token before the terminal punctuation; `normalizeQuote` skips annotation spans and trims trailing
punctuation. So a quote that omitted the marker still matched, and the span it located ended just
short of it — meaning on the commonest edit there is, a whole-sentence rewrite, the guard never
fired at all. Measured: a citation on *"The sky is blue and the grass is green"* followed the text
to *"The sky is green and the grass is on fire"*, silently. The orphaned `now.<!--cite-->.` was the
visible residue: `tidySeam` collapses an exact `..` pair only when the two are adjacent, and the
marker sat between them.

**Why the complexity existed at all**, since the code records it: `trailingPunct` is there because
*"a quote may omit or include a terminal period the report has, or vice versa"* — seats type quotes
by reading prose and are inconsistent about terminal punctuation. Every guard downstream is
paying for one thing: **text is addressed by quoting it, and the quote is normalized, so span
boundaries stop corresponding to the raw bytes.**

**The operator's ruling: the markers ARE the mechanism.** Not block ids — positional addressing
fails an LLM the way line numbers do. Not hidden markers — that was tried and broke worse. The
markers stay real, stay visible, and stay in the edit stream, because `edit` mirrors how an agent
actually edits any document: quote what is there, write what should be there. A seat that can SEE
a token copies it like any other character.

So the fix DELETES A TOLERANCE rather than adding machinery. The quote is the evidence of intent:
contains the token → the span swallows the punctuation run and the token, the guard fires, and the
terminator travels with the replacement; omits it → refused, with the token printed to carry.
A fragment edit that does not reach the anchor stays legal, and `reopened` records that its
sentence moved.

**`sentence_hash` is burned.** No production reader, no prompt, no document — only its own test.
Its job was "so blue can match an occurrence after edits move it", and blue cannot compute FNV-1a
by hand. The anchors are the locator; that is what visible markers are FOR.

**The rule is scoped to REPLACEMENT, and finding that out cost a commit.** Baked into
`LocateUnique` it also fired on `merge mint --quote`, which names the sentence a defect LIVES AT
and rewrites nothing — so minting a gap about any already-anchored sentence was refused, with a
message about "the text you are replacing". It is `LocateUniqueReplacing` now: `blue edit`, and
`ValidateProposal` (red's proposed text is applied verbatim by blue, so it owes the same duty).

The method note matters more than the fix. It was found by reading a golden that had ALREADY BEEN
COMMITTED unread — swept in by `git add -A plugins` — where a whole round's integration output had
recorded the regression as the new expected shape. A golden records what the code does, so a
golden regenerated without reading turns a regression into a baseline. Read every diff; regenerate
in its own act.

**Not fixed, and worth stating:** `InsertAnchor` still places the token before the terminal
punctuation. With the span now swallowing both, nothing depends on the side any more — but placing
it after would let `tidySeam` see an adjacent pair unaided, which is one less thing balanced on an
offset.
