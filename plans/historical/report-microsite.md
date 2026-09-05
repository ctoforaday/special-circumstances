# Report microsite — split the assembled report by audience, render it as a site

## I. Summary & Goals

`bench assemble` emits ONE markdown file per run. Measured on the two archived runs, 70–76% of
that file is process record — the debate transcript, red's findings in full, a friction log
larger than the entire research argument (60–65 KB), revision history, cost. The research it was
commissioned to deliver is 24–30%. A reader gets six documents unioned into one and no way to
address any of them separately.

The report is already a PROJECTION, not a source: `run-archive/*.tar.gz` carries `records/` and
`proofs/` only, and `report.Assemble` re-derives the document from the event log. Multiple
renderings therefore cost nothing in provenance.

Goals:

1. Split the single artifact into seven markdown documents by AUDIENCE, each standalone, each
   carrying a link bar to its siblings, indexed by a run `README.md`.
2. Render the same set as one self-contained `report.html` with real tabs — no build step, no
   server, no network — from the same Go package that already builds `dashboard.html`.
3. Fix the defects the split makes visible on page one: a `Verdict:` field carrying a paragraph,
   the bench's ask rendered once per `certify` event, a boilerplate line contradicting it, an
   empty `## Cost` heading, and proof footnotes referenced but never defined (#590).

Non-goals: a static-site generator, a markdown AST round-trip (raw-slicing is load-bearing —
see the package doc), any change to what the seats write or to the record schema.

## II. Current State

- `internal/report/assemble.go:186` `Assemble(runDir)` composes ~20 sections into one string and
  writes `<run>/report.md`. Callers: `internal/cli/bench/assemble.go:34`, one test.
- `internal/report/proofs.go:56` appends `## Proofs` post-composition; `internal/capture/
  capture.go:1591` appends `## Cost` post-assembly.
- `weaveCitations` / `weaveProofs` rewrite invisible anchors to `[^N]` / `[^PN]` over the WHOLE
  document and append their definitions at the end.
- `internal/dashboard/render.go` already builds a self-contained HTML+SVG file in Go with no
  dependencies. `go.mod` has no markdown library and will not get one.

Known defects, all reproduced in `research/2026-08-23_research-loop-counterparts/report.md`:

| # | Defect | Site |
|---|---|---|
| D1 | `**Verdict:**` field carries a 400-char paragraph | `assemble.go` `verdictStamp` |
| D2 | One "The bench asks a human to re-examine" block per `certify` event — a re-certified run prints it twice | `assemble.go:459` |
| D3 | "_(no open gaps remain — nothing outstanding to re-examine)_" printed under blocks asking for re-examination | `assemble.go:472` |
| D4 | `## Cost` renders as an empty heading; the table sits under `## Per seat-round` | `capture.go:1591` |
| D5 | Proof appendix headings written as `### [^PN]`, which markdown reads as a second dangling reference; no `[^PN]:` definition is ever emitted, so every proof footnote in the body is dangling | `proofs.go` |
| D6 | H1 is the entire research question with " — research report" appended | `assemble.go` `titleOr` |

## III. Design

### Tier 1 — markdown, plural

`AssembleAll(runDir) ([]Doc, error)` composes the same section strings and groups them:

| File | Sections |
|---|---|
| `report.md` | title, verdict, how this run was conducted, read this first, TL;DR, Catechism, technical foundations, analysis, risk matrix, the three inquiry areas, open questions, blue embed |
| `docket.md` | red team findings in full (open gaps, closure index, lens findings, archive spot-checks, correctness manifest) |
| `debate.md` | the round-by-round transcript and the bench disposition |
| `judgments.md` | motions — every contested question and its ruling |
| `evidence.md` | proofs in full: script, output, sha256, red's re-run |
| `run.md` | friction, record verification, cost |
| `CHANGELOG.md` | report revision history, claims withdrawn, post-run repairs |

`Doc{File, Title, Blurb, Body}`. An empty body means the file is not written (a run with no
motions has no `judgments.md`), and the link bar omits it. `README.md` indexes the set.

Every doc opens with the same link bar, current document bolded and unlinked.

CITATION AND PROOF LAYERS ARE WOVEN PER FILE, after the split: each document numbers the anchors
it actually contains and carries the definitions for exactly those. A footnote definition cannot
cross a file boundary, so the alternative — weave globally, then split — would ship dangling
references in six of the seven documents. Numbering is therefore per-document, which is
self-consistent within each and stated in the bibliography's own line. PROOF numbering is the
exception, and deliberately: a proof number is the RUN's, so `P2` in the debate and `P2` in the
report name the same computation (`TestProofNumbersAreRunWideNotPerDocument`) — only the
DEFINITIONS are per-file, which is what the file boundary actually requires.

`Assemble` keeps its signature and its `report.md` return: it calls `AssembleAll` and writes the
set. No caller changes shape.

### Tier 2 — one self-contained HTML file

`RenderSite(docs []Doc, ...) string` → `<run>/report.html`: inline CSS, inline JS, no network.
Tabs are the docs. Requires markdown→HTML, and `go.mod` gets no new dependency, so
`internal/report/md.go` renders the SUBSET this assembler emits: ATX headings, fenced code,
GFM tables, lists, blockquotes, rules, footnote definitions and references, and the inline set
(code, strong, emphasis, links, autolinks). Raw HTML passes through.

What the site gets that markdown cannot express:

- **Cross-tab identity links.** Gap, motion and proof ids become links to their defining heading;
  the tab switches automatically. `R4-3` appears in four documents today and the reader scrolls.
- **A fact box**: verdict, rounds, gaps open/closed, adversary conduct, cost — currently spread
  across four composers and 880 lines.
- **Client-side filtering** of the debate by seat and round.

### Defect fixes

D1 verdict field is the enum alone; its explanation moves to the first line of "Read this first".
D2 render the TERMINAL certify only; earlier ones are listed as superseded in `CHANGELOG.md`.
D3 the no-open-gaps line prints only when the bench asked for nothing.
D4 `appendCostToReport` writes the sliced table under its own heading, not a second empty one.
D5 proof appendix headings become `### PN — <reason>` with a real `[^PN]:` definition pointing at
   them, and the full script/output moves to `evidence.md`.
D6 the H1 is blue's title truncated at the first sentence boundary; the full question renders as a
   `**Question:**` line beneath it.

Plus an invariant: no document ships a heading with no body.

## IV. Carriers and consumers

`report.md` is read by `capture.go:466` and `:541` (weave + citation screens), `appendCostToReport`,
`run-record-audit.md`'s generator, `skills/research-protocol/references/report_template.md`,
`agents/lead-judge.md`'s assemble verb, `commands/research.md`'s closing message,
`skills/research-protocol/scripts/debate.js`, `docs/record-flow.md`'s bibliography diagram, and
both `README.md`s. Every one is swept in this change: `report.md` continues to exist and to be the
research document, so no reader BREAKS, but each that describes the artifact as the whole record
now describes the set.

TWO CORRECTIONS TO THIS CENSUS, both found by running it rather than reading it:

- **The gray-area audit skills are NOT carriers.** `grep -rn 'report\.md' plugins/gray-area`
  returns nothing; they audit checkpoints, pull-request bodies, repetition and seat coverage, and
  never read the run's report. The line above claimed a consumer that does not exist.
- **Six sites match the string and mean something else.** `agents/blue-synthesizer.md`,
  `agents/blue-researcher.md`, `internal/cli/seat/help/edit.md`, `docs/seat-command-triggers.md`
  and `docs/finding-markers.md` all say `report.md`, and every one of them means
  `blue/report.md` — the WORKING DRAFT under the edit lockdown, a different document with a
  different lifecycle. They are deliberately untouched. `docs/record-flow.md` is the one file
  that names both: its `RPT` node is blue's draft (untouched) and its `ASM` node is the assembled
  artifact (swept). This is [[facts-are-fields]] clause 4 in the concrete: a carrier is a site
  that speaks the same CONCEPT, not one that shares a STRING.

## V. Verification Plan

```
cd plugins/frank-exchange-of-views/tools
go build ./...
go test ./internal/report/... ./internal/capture/... ./internal/cli/...
go test ./integration/fuzz/...
go vet ./...
```

Re-armed by: any change under `internal/report/`, `internal/capture/`, or the agent/skill/command
markdown that names `report.md`.

End-to-end, against a real archived run:

```
go run ./cmd/feov-record --seat-id <the run's bench seat> assemble --run <rundir>   # writes the set
ls <rundir>/*.md <rundir>/report.html
```

`--seat-id` is REQUIRED and the verb is `assemble`, not `bench assemble`: the surface is scoped to
the registered seat, and for a bench seat the bench verbs are hoisted to the top level.

**This check cannot be run against `run-archive/`.** Both archived runs predate the protobuf record
and hold `events-*.jsonl` shards with no `record.db`; this binary refuses them by design rather than
reporting an empty board. The end-to-end assertion therefore lives in the test suite, driven through
the production write path, where the record is built under the current epoch:

- `TestAssembleEndToEnd` — the seven files, the split by audience, the link bar, the index, and the
  suppression of a document with no content.
- `TestNoDocumentInTheSetShipsADanglingFootnote` — every `[^N]` and `[^PN]` reference in each
  document has a definition in that SAME document, over a run seeded with a citation that reaches
  only `docket.md` and a proof that reaches only `debate.md`. It asserts the spread as well as the
  closure: a dangling-reference scan over a set that carries no references reports a clean board in
  the same bytes it would use for a real one.
- `TestTheSiteIsSelfContained` / `TestTabsAreTheDocumentSet` — `report.html` carries its own CSS and
  JS, requests nothing, and its tabs are the document set.
