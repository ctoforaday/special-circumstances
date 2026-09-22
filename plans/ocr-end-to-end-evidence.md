# One run that cites a PDF

Status: proposed 2026-09-21. Carries #1101. **The run is a spend and therefore gblock's call; this
plan says what to run, what would count, and what would fail it.**

## I. Summary & Goals

**The gap, measured by another session and not by me.** Page-anchored OCR citation landed
2026-09-15. Across every live record on disk since — today's smoke, universe-1077, and five
2026-09-19 baseline runs — there are 13 citations and **zero cited PDFs**, zero non-empty
`ocr_quote`, zero `cite_pages` rows, and no `application/pdf` fetched in either universe (#1101).

The cause is structural and is not a defect: every run since has been `--smoke` — one lane, `haiku`
on both tiers, `--mint-budget 1 --k-max 2`, topic "is 91 prime". One haiku lane on trivial
arithmetic reaches Wikipedia and stops. The runs that DID cite PDFs were three-lane
development-tier runs on topics where a lane had reason to reach for a standards document.

**Why the silence is the problem rather than the zero.** "No friction reported on the PDF path" and
"the PDF path has never been taken" are the same bytes. Everything shipped in the last two days —
#1074, #1104, #1107, #1112 — is validated by corpus goldens and unit tests, which compile code and
compare readings but **never dispatch a seat**.

**The sharpest case for doing this now.** PR #1107 fixed `lens render-page`: it rendered at a fixed
300 DPI and compared render hashes against a reading made at the page's own resolution, so
`matches_reading` was **false for every scanned page** — a verification check reporting "these
differ" about a document nobody had touched. That bug shipped, survived, and was found by a unit
test only because the same change broke the test. **A check that was silently wrong is now silently
right, and only a run that actually cites a scanned PDF distinguishes the two.**

**Goals.**

1. One debate cites at least one PDF, and at least one of those citations carries an OCR quote.
2. The citation surface downstream of the reading is exercised, not just the reading: `ocr_quote`,
   `cite_pages`, `page_render_sha`, `reading_render_sha`, the lens verify refusal that demands a
   checked page, and `render-page`.
3. Whatever it finds is written down. **A clean run is a result, not a non-event** — it is the first
   evidence this path works at all.

**Non-goal.** Changing the smoke. One haiku lane at `--mint-budget 1 --k-max 2` is correct and cheap
and should stay that way; a fetch-heavy topic in the rotation changes what it costs and what it
proves. PDF coverage needs a DELIBERATE run, not a smoke that wanders into one.

## II. What to run

- **Tier:** development — three lanes, `sonnet`. The smoke configuration cannot reach this.
- **Topic:** one whose authoritative evidence IS a PDF, and preferably a SCANNED one, since a
  born-digital PDF has a text layer and never reaches the OCR path at all. A standards question is
  the natural shape: `is-91-prime-b6` found NIST FIPS 186-5 unprompted, and NIST's legacy
  publications at `nvlpubs.nist.gov` are scans — the corpus's own NBS SP 602 came from there.
- **Sequencing, and it matters:** this runs AFTER #1104, so the OCR it exercises is the dashed-rule
  path and the 1:1 renderer. Running it before would have exercised code that no longer exists.
  Deliberate, not accidental.
- **Cost:** it can ride a development-tier run already owed for another purpose (#1079 and #1085
  both need one and are open), which makes the marginal cost of this close to the topic choice
  alone.

## III. What counts as evidence

Checked against the run's own record, not against a transcript:

| claim | where it is checked | passes when |
|---|---|---|
| a PDF was fetched | the fetch cache | an entry with `application/pdf` |
| it was a SCAN | that entry | `text_extracted: false`, and a reading exists |
| a seat cited it | `cite` rows | at least one cite whose source is that sha |
| the cite is page-anchored | `ocr_quote`, `cite_pages` | both non-empty on that cite |
| the render is re-derivable | `page_render_sha`, `reading_render_sha` | both present and EQUAL |
| red could check it | `lens render-page` | `matches_reading: true` |

The fifth and sixth rows are the ones #1107 fixed and nothing has exercised.

## IV. Accept / reject, fixed before the run

Written now so the result cannot be read generously afterwards — the discipline that caught a
fabricated table in #1104 and a non-deterministic reader in `ocr-declare-the-resolution`.

- **PASS:** every row of §III holds. Record it and close #1101.
- **FAIL, and it blocks the release tag:** a citation carries an OCR quote whose text is NOT on the
  page it names, or `matches_reading` is false on a page nobody edited. Either means a seat can
  cite something the document does not say, which is the failure this whole line exists to prevent.
- **INCONCLUSIVE, and it is NOT a pass:** no lane reaches for a PDF at all. Then the topic was
  wrong, not the code, and the run is repeated with a topic that forces one. **A run that never took
  the path proves exactly as much as the seven runs before it.**

## V. What this cannot show

- One run is one sample. It can prove the path WORKS; it cannot bound how often it fails.
- It exercises whatever the lanes happen to cite. A scanned table — the case #932, #933 and #1027
  were built for — is likelier from a standards document than a journal paper, but it is not
  guaranteed, and if no cited PDF is a scanned table then the table half remains unexercised and
  this plan says so rather than claiming otherwise.
- It says nothing about recognition accuracy. The corpus measures that; this measures whether the
  plumbing carries a citation end to end.
