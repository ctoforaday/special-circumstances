# Resolving a scholarly URL without a per-publisher adapter

## I. Summary & Goals

`fetch` reaches a source by extracting a DOI or an arXiv identifier from its url and asking
Unpaywall, OpenAlex, Crossref or the Wayback CDX about it. Everything else is a live GET. That
covers the sources a seat happens to name with a doi.org link and fails, silently, on the rest —
and the silence is the defect, not the coverage.

The goal is not more publishers. It is a **resolution ladder** whose rungs are generic, plus an
**acceptance gate** that can say *that was not the document*, plus a **terminal state per
attempt** so that six different outcomes stop arriving as the same sentence.

Three facts decide the shape, each measured rather than assumed:

- **32.1% of the works OpenAlex indexes carry no DOI at all** — and the gap is a genre gap, not a
  historical one (dissertations 81.3%, books 59.0%, articles 38.8%; pre-1980 works 38.6%, no worse
  than the modern average). **18.4 million DOI-less works have a recorded open-access location**,
  which Unpaywall — keyed on DOI — can never answer for.
- **Of 42 urls the open-access hubs labelled as a PDF, 28 returned HTML.** A fetcher that accepts
  any 200 records a bot wall as a reading. This is the single largest correctness hole, and it is
  not about which hub you ask.
- **Unpaywall and OpenAlex are one opinion, not two.** On 100 random Crossref DOIs: 96/96 identical
  `is_oa`, 42/42 byte-identical `best_oa_location`, zero disagreements. Both are OurResearch
  products and the alignment is stated policy. `OpenAccessURL` asks both in sequence and pays
  latency for no information.

Non-goal: **translators.** Zotero maintains 748 of them, and its own documentation says the best
translator is none. What a translator buys that a resolver cannot is the *authenticated DOM* — a
browser session this tool does not have and should not simulate. Where a document is behind a
wall, the honest terminal state hands the url to a human, who is the translator of last resort.

Also a non-goal, and stated so it is not re-proposed: **solving a challenge.** Where a host places
a proof-of-work or a CAPTCHA, the answer is the route its owner sanctions, never a solver. PubMed
Central is the worked example and it is not an exception — NCBI names its approved automated
routes itself, on its own page (verified, `pmc.ncbi.nlm.nih.gov/tools/openftlist/`): *"The PMC
Cloud Service, PMC OAI-PMH Service, E-Utilities and BioC API are the only services that may be
used for automated retrieval of PMC content."* The `/articles/PMC…/pdf/` path the challenge guards
is not among them. The anonymous S3 bucket serves the same article with no key and no challenge
(verified: `pmc-oa-opendata.s3.amazonaws.com`, anonymous list returns the JATS, the plaintext, the
PDF and a licence sidecar). **The gate was never the obstacle; using the unsanctioned door was.**

## II. The ladder

Six rungs. Each is generic — no rung names a publisher.

1. **Harvest** — regex the url for *every* identifier type and emit `[{type, value}]`, not a
   string. A 404 on one identifier is not a miss: `10.48550/arXiv.1706.03762` 404s at OpenAlex
   while OpenAlex holds that paper under another DOI. Where the input is a landing page being
   fetched anyway, harvest `citation_doi`, `DC.identifier`, JSON-LD and the signposting
   `Link rel="cite-as"` from that same response — no extra call.
2. **Canonicalize** — one OpenAlex singleton lookup by path (`/works/doi:…`, `/works/pmid:…`,
   `/works/W…`). Measured: the singleton path costs **$0.00 and zero credits**, while
   `?filter=doi:` bills $0.0001 and exhausts an anonymous daily budget in 1,000 lookups. The
   returned record carries the whole identifier set, which subsumes every converter except PMCID
   (Europe PMC) and handles/ARKs (handle.net, N2T).
3. **Locate** — take `locations[]`, **not** `best_oa_location`. Rank `source.type == repository`
   above `journal`. This is the finding that replaces a hostname blocklist with a field: for
   `10.1038/nature12373` the default `best_oa_location` returns the gated Nature PDF while
   `locations[]` holds an arXiv copy and a CC-BY Harvard DASH copy. Repositories served real PDFs;
   publishers gated.
4. **Fetch through the acceptance gate** (§III). On a landing page, take one hop via
   `citation_pdf_url` or signposting `rel="item"`.
5. **Escalate only on a miss** — Semantic Scholar's *batch* endpoint (the per-DOI anonymous pool
   429'd 89 of 100 calls; the batch took 100 ids unauthenticated in one 200), Europe PMC
   (`fullTextUrlList` types each url `Free` vs `Subscription required`), then the archive. Each
   result re-enters the gate at rung 4.
6. **Record the terminal state with its evidence** (§IV).

Round trips: 2 on the good path, 3–4 when the first location is gated, ~8 capped.

## III. The acceptance gate

A fetch must be able to conclude *these bytes are not the document*. Every naive signal fails on
measured data: status lies in both directions (a 200 challenge, a 403 carrying 1.2 MB), size lies
(a 3 KB gate and a 1.2 MB gate), text volume lies (89,371 visible characters of challenge page),
and identity lies — Springer and JSTOR return byte-identical 3,038-byte pages under one cookie
name, because the gate is a *vendor*, not a site.

**Accept**, in order: body begins `%PDF-` (magic bytes, never `Content-Type` — one real PDF
arrived as `application/force-download`); JATS/TEI root element; or HTML carrying a bibliographic
meta tag *and* ≥500 visible characters, which is a **landing page** — harvest and hop, do not
cite.

**Reject as a gate**: `cf-mitigated: challenge` (`cf-ray` alone is a false positive — it appears on
legitimate pages); vendor cookies (`_fs_ch_st_*`, `_Incapsula_`, `_abck`); `POW_CHALLENGE` in the
body; a `<title>` in the observed vendor set; a `202`, or a `200` with zero bytes. **And the
backstop that generalizes: under 500 visible characters with no bibliographic meta tag** — which
caught five gates carrying no vendor marker at all.

The cheapest signal is free: **expectation mismatch.** Asking "I requested a PDF and received
HTML" flags two thirds of these before any body inspection.

We already have the shape-based half of this. `ShellReason` fires correctly on PubMed Central
today, and it was *I* who reported it did not — my survey harness printed a fixed field set that
omitted `not_renderable`, and I filed the absence of a warning I had never looked for. What is
missing is not the detection but the **attribution**: the reason says "a client-side app renders
into this" where the cause is a proof-of-work challenge, and those license different next actions.
The archive rung needs the same gate — the Wayback availability API reports `status: 200` for
captures of gated pages, so its metadata cannot tell a captured article from a captured
interstitial.

## IV. Terminal states

Ten, each carrying its evidence. Today several arrive as one sentence: *"that backend has no answer
for this url"*. Starred are the ones currently collapsed.

| State | Evidence it must carry |
|---|---|
| `READ` | identifier used, location url, `source.type`, version, which accept rule fired |
| `READ_SURROGATE` | the same, plus the surrogate class — a capture, a submitted version, an OCR reading |
| `METADATA_ONLY` | which hub answered, whether an abstract came with it |
| ★ `LOCATED_BUT_BLOCKED` | the gate vendor, the rule that fired, and the url — **actionable by a human with a browser** |
| `LOCATED_BUT_ABSENT` | url, status, whether a capture exists |
| ★ `NO_OPEN_COPY` | which hubs agreed — "cite something else" |
| `IDENTIFIER_UNRESOLVED` | which identifiers were tried |
| ★ `RESOLVES_TO_NON_DOCUMENT` | 4 of 100 random Crossref DOIs were a figure, a supplement or a peer-review report — they resolve at doi.org and 404 at every hub |
| `NO_IDENTIFIER_FOUND` / `AMBIGUOUS_MATCH` | candidates and scores |
| `SOURCE_MAY_NOT_EXIST` | the queries run — **assertable only from searches that succeeded and found nothing, never from a chain of errors** |
| `TRANSIENT` | must never collapse into any of the above |

The pair that matters most is `LOCATED_BUT_BLOCKED` against `NO_OPEN_COPY`: the difference between
*Cloudflare said no* and *this citation cannot be supported*. `Recover`'s `oa` branch currently
turns the first into the second's absence — it finds a location, the fetch errors, and it returns
nil, which the caller renders as "that backend has no answer".

## V. Verification Plan

Each check names what re-arms it.

1. `go test -count=1 ./internal/fetchcache/ ./internal/cli/` → ok. *Re-armed by any change under
   `internal/fetchcache` or the fetch summary.*
2. `go -C scripts run ./check` → 0 failed. *Re-armed by any change to a help surface or golden.*
3. **The gate against real bytes, both ways.** A corpus of captured responses — real PDFs, real
   landing pages, and the challenge pages measured here (PMC proof-of-work, F5/Shape 3,038-byte,
   Cloudflare `Just a moment`, the 1.2 MB ScienceDirect 403) — asserting accept and reject. The
   false-positive direction is the one with no evidence yet: rejecting a legitimate page is worse
   than accepting a wall, because it turns a readable source into a fabricated absence. *Re-armed
   by any change to the gate rules.*
4. **A terminal state per fixture**, asserted as a field and never by matching prose. *Re-armed by
   any change to `Recover`.*
5. **The ladder end to end, live**, on a list spanning the identifier classes: an old-scheme arXiv
   id, a PMCID, a handle, a DOI-less thesis, a paywalled article with a repository copy, and a DOI
   that resolves to a figure. Recording which rung answered. *Re-armed by a rung change.*

## VI. What this cannot show

- The numbers come from samples of 100 random Crossref DOIs and 42 open-access urls, run from one
  cloud IP on one day. Directions are large and consistent; the 28.6% fetch-success figure has an
  interval that should be widened before it is quoted as policy.
- **Block state is a property of the (address, client) pair, not of a source.** A reCAPTCHA reached
  one client while a plain `curl` from this box got a 200 in the same minute. "This host is
  blocked" must never be cached as "this source is blocked" — which is also why a gate verdict
  cache expires in hours and a hostname blocklist is not an acceptable shortcut for it.
- OpenAlex's index is not the universe of scholarship, and it over-indexes repository content. The
  DOI-coverage percentages are of what OpenAlex holds.
- No published precision/recall comparison exists for any pair of these hubs. The "ask one, escalate
  on a negative" recommendation rests on the measurement above, not on literature.
- Nothing here has been run inside a debate. Gates compile code and compare goldens; they never
  dispatch a seat, and a seat's judgement about which terminal state licenses which next act is not
  testable by any check in §V.
