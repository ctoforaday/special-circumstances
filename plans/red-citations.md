# Red's evidence reaches the reader

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

- `supports` / `supports_with_bridge` → mint a label, splice, render as a footnote.
- `refutes` / `absent` / `weak` → red finding the text unsupported IS a defect in the text.
  **A LENS CANNOT MINT** — that is structural, not an oversight ("A lens structurally cannot
  mint or close a board gap: no such verb exists in its namespace"). So the carrier is
  `lens finding`, which the merge raises to a gap through `found_by`. The negative
  corroboration records as evidence AND the seat owes a finding; the tool says so.
- `unreachable` is neither: red could not read the source. Record-only.

**Open sub-question, deliberately not decided here:** whether the finding should be written by
the tool as a side effect of the negative corroboration, or demanded of the seat with a refusal.
The first is a verb owning two acts; the second can be ignored. Leaning toward recording the
corroboration and reporting a negative-with-no-finding as a duty gap, which is the shape
`InquiryReviewDue` already uses.

## III. Carriers and consumers

- `recordpb.Verify` — new `label` field (reserved numbering: next free).
- `internal/cli/lens/verify.go` — `corroborate` mints, splices, and branches on outcome.
- `internal/record/citationid.go` — `CitedSources` must include labelled corroborations; its
  comment about red being excluded is the thing being changed and must be rewritten, not left.
- The `unbacked_citations` detector reads `CitedSources`' labels as its EXPECTED set — check it
  does not now report red's anchors as unbacked.
- `internal/record/evidenceview.go` — the `evidence` projection still shows both.
- Goldens: `internal/difftest` (help + error catalogue), `recordsql/testdata/schema.sql`,
  the simulator's prompt goldens if any prompt text moves.
- `internal/fuzz` — drive corroborate on both arms, or the split is unexercised.

## IV. Risks

- **Splicing into blue's document from a red seat.** The anchor is TOOL-placed for blue too, so
  this is the tool citing on red's behalf rather than red editing prose. Still the sharpest edge
  in this change: red has no edit verb by design, and this gives it a write into `report.md`.
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
   INTERRUPTED before producing results (it restored `record.go` cleanly). Not yet run.
